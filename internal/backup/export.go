package backup

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// 导出格式。zip 是本应用的默认：六个库打成一个归档；db 只带主库，是有损的对称选项。
const (
	FormatZip = "zip"
	FormatDB  = "db"
)

// sqliteMagic 是 SQLite 数据库文件的固定头（16 字节，末尾 NUL）。
const sqliteMagic = "SQLite format 3\x00"

// Run 是 `bili-history backup` 的入口，路径按应用自身约定解析。
// stdout 只写产物，诊断写 os.Stderr。
func Run(args []string, stdout io.Writer) error {
	return Execute(args, CurrentLayout(), stdout, os.Stderr)
}

// Execute 解析参数并导出。
//
// 与 Run 分开是为了让测试能把 Layout 指到临时目录，而 CLI 走真实的应用路径解析——
// 备份工具不需要任何额外配置。
func Execute(args []string, layout Layout, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("bili-history backup", flag.ContinueOnError)
	fs.SetOutput(stderr) // 用法与解析错误都不许进 stdout
	format := fs.String("format", FormatZip,
		"导出格式：zip（默认，数据目录里的全部 SQLite 库）或 db（只有主库，对本应用是残缺备份）")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("backup 不接受位置参数，多余的是: %s", strings.Join(fs.Args(), " "))
	}

	choice := strings.ToLower(strings.TrimSpace(*format))
	if choice != FormatZip && choice != FormatDB {
		return fmt.Errorf("未知格式 %q（可选 zip / db）", *format)
	}

	// 先发现再动磁盘：缺主库、空库、包内重名这些错误都必须发生在第一个字节写出之前，
	// 否则 stdout 上就留下一截残缺产物了。
	entries, err := layout.Discover(stderr)
	if err != nil {
		return err
	}
	if choice == FormatDB {
		var primary []Entry
		for _, e := range entries {
			if e.Required {
				primary = append(primary, e)
			}
		}
		entries = primary
		fmt.Fprintln(stderr,
			"提示: --format db 只导出主数据库，动态 / 点赞 / 稍后再看 / 收藏 / 调度器都不在包里；完整备份请用默认的 zip")
	}

	// 快照先全部落进临时目录并逐个校验，之后才往 stdout 写——
	// VACUUM INTO 要求目标不存在，且这样无论调用方读到一半断开，数据目录里都不会有半成品。
	dir, err := os.MkdirTemp("", "bili-history-backup-")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(dir)

	for i, e := range entries {
		snap := filepath.Join(dir, fmt.Sprintf("%02d-%s", i, e.Name))
		if err := snapshot(e.Path, snap); err != nil {
			return err
		}
		e.snapshot = snap
		entries[i] = e
	}

	switch choice {
	case FormatDB:
		return streamFile(stdout, entries[0].snapshot, entries[0].Name)
	default:
		return writeZip(stdout, entries)
	}
}

// snapshot 用 VACUUM INTO 做一份一致性快照。
//
// 不热拷贝 .db / .db-wal / .db-shm：活跃写入下那不是原子操作，会周期性拷出损坏副本。
// VACUUM INTO 由 SQLite 在事务里读出含 WAL 的完整逻辑视图，写出一份自洽的库，
// WAL 模式下也不阻塞主进程。
//
// 一致性是「每个库各自一致」：SQLite 保证单个库内的事务级自洽，但这些库是互不相干的
// 功能库（历史、动态、点赞…），不是一个关系型整体，跨库的同一个时间点做不到，
// 也不需要做到。
func snapshot(src, dst string) error {
	conn, err := openForExport(src)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Exec("VACUUM INTO ?", dst); err != nil {
		return fmt.Errorf("导出 %s 失败: %w", filepath.Base(src), err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("读取 %s 的快照失败: %w", filepath.Base(src), err)
	}
	if err := assertSQLite(dst, info.Size()); err != nil {
		return fmt.Errorf("%s 的快照不可用: %w", filepath.Base(src), err)
	}
	return nil
}

// assertSQLite 确认产物确实是一个 SQLite 文件而不是空文件或别的东西。
// 宁可当场失败，也不要把一份「看着像备份」的空字节流交给备份工具。
func assertSQLite(path string, size int64) error {
	if size < int64(len(sqliteMagic)) {
		return fmt.Errorf("快照只有 %d 字节，不是有效的数据库", size)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	head := make([]byte, len(sqliteMagic))
	if _, err := io.ReadFull(f, head); err != nil {
		return fmt.Errorf("读取快照头部失败: %w", err)
	}
	if string(head) != sqliteMagic {
		return errors.New("快照头不是 SQLite format 3")
	}
	return nil
}

// writeZip 把快照逐个流式写进 stdout 上的压缩包。
//
// 整包不缓冲在内存里：历史库可能上百 MB，备份工具是把这些字节当产物采集的，
// 攒一份再吐出来既费内存又要等齐才见第一个字节。zip.Writer 本身就是流式的。
func writeZip(w io.Writer, entries []Entry) error {
	zw := zip.NewWriter(w)
	for _, e := range entries {
		if err := zipOne(zw, e); err != nil {
			_ = zw.Close() // 先让半截包尽量可读，真正的失败原因由这里返回
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("完成压缩包失败: %w", err)
	}
	return nil
}

// zipOne 把一份快照作为条目写进压缩包，条目名就是库文件名（平铺在包根）。
func zipOne(zw *zip.Writer, e Entry) error {
	f, err := os.Open(e.snapshot)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", e.Name, err)
	}
	defer f.Close()

	hw, err := zw.Create(e.Name)
	if err != nil {
		return fmt.Errorf("创建压缩包条目 %s 失败: %w", e.Name, err)
	}
	if _, err := io.Copy(hw, f); err != nil {
		return fmt.Errorf("写入压缩包条目 %s 失败: %w", e.Name, err)
	}
	return nil
}

// streamFile 把一个文件原样拷进 w（--format db 走这条，产物是裸的 SQLite 文件）。
func streamFile(w io.Writer, path, name string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", name, err)
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("写出 %s 失败: %w", name, err)
	}
	return nil
}
