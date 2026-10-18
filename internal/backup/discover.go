// Package backup 实现 `bili-history backup` 子命令：把应用的各个 SQLite 库逐一
// 做一致性快照（VACUUM INTO）后流式写进 stdout，交给外部备份工具采集。
//
// 为什么产物是 zip 而不是单个 .db：拾帧集不是「一个应用一个库」，它把历史、动态、
// 点赞、稍后再看、收藏、调度器分别落在多个 .db 文件里（见 Discover）。
// 备份工具的 full 模式是「一个容器一个文件」，所以多库容器必须把全部库打进
// 同一个归档（docker-db-auto-backup/docs/sqlite-backup-design.md §9）。
//
// 三条不变量：
//   - stdout 只承载产物字节，所有诊断信息一律走 stderr；
//   - 只读打开既有库：不建目录、不跑迁移、不写种子，绝不产出空库；
//   - 第一个字节写出之前能发现的错误（缺库、空库、包内重名）必须先失败。
package backup

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bilibili-history-go/internal/utils"

	// 纯 Go SQLite 驱动，注册名 "sqlite"；备份进程同样不依赖 cgo。
	_ "modernc.org/sqlite"
)

// optionalDBFiles 是主库之外已知的库文件名，与 internal/database 里的
// getExtraDB / GetSchedulerDB 调用一一对应（extras.go、dynamic.go、scheduler.go）。
// 这些名字写死在代码里、不受配置影响；受配置影响的只有主库名（config db_file）。
//
// 列出来只为了「少了一个要出声」：一个部署没用到收藏夹是正常的，
// 但收藏夹库凭空消失不该被静默忽略。
var optionalDBFiles = []string{
	"bilibili_dynamic.db",
	"bilibili_likes.db",
	"bilibili_watchlater.db",
	"bilibili_favorites.db",
	"scheduler.db",
}

// Entry 是归档里的一份数据库快照。
type Entry struct {
	// Name 是包内条目名，同时也是磁盘上的文件名——包内一律平铺，
	// 恢复时把条目原名放回数据目录即可，不需要任何翻译表。
	Name string
	// Path 是数据目录里的实际路径。
	Path string
	// Required 为 true 时缺文件即整体失败（只有主库）。
	Required bool

	// snapshot 是 VACUUM INTO 落到临时目录里的那份快照，仅导出过程中使用。
	snapshot string
}

// Layout 是本次导出要扫描的目录与主库名。
//
// 单独抽出这个结构，是为了让 CLI（用应用自身的路径解析）和测试（用临时目录）
// 走完全相同的发现与导出逻辑。
type Layout struct {
	OutputDir   string
	DatabaseDir string
	PrimaryName string
}

// CurrentLayout 按应用自己的路径约定解析出本次导出的 Layout。
//
// 刻意复用 internal/utils 的纯路径版本（OutputDir/DatabaseDir/DBFileName）而不是
// GetOutputPath/GetDatabasePath：后者会顺手 MkdirAll，一个只做导出的进程不该在
// 数据目录里凭空造出结构来。
func CurrentLayout() Layout {
	return Layout{
		OutputDir:   utils.OutputDir(),
		DatabaseDir: utils.DatabaseDir(),
		PrimaryName: utils.DBFileName(),
	}
}

// Discover 列出要打包的库。
//
// 主库由 config 的 db_file 决定（在 output 根目录），其余库固定落在 output/database 下。
// 两边都做运行时发现：扫目录能捞出代码清单还没来得及收录的新库，已知清单则保证
// 「本该存在的库不见了」会告警而不是被漏掉。这里只读目录，不创建任何东西。
func (l Layout) Discover(stderr io.Writer) ([]Entry, error) {
	name := strings.TrimSpace(l.PrimaryName)
	if name == "" {
		return nil, errors.New("主库文件名为空，无法确定要备份哪个数据库")
	}
	// 包内是平铺的，条目名只能取裸文件名；db_file 配成带目录的路径就没法还原了。
	if name != filepath.Base(name) {
		return nil, fmt.Errorf("主库名 %q 必须是文件名，不能带目录", name)
	}

	// 按解析后的路径去重：同一目录被扫了两遍、或 db_file 正好等于某个已知库名时，
	// 不能让一个库在包里出现两次。
	seen := make(map[string]bool)
	var entries []Entry

	add := func(path string, required bool) {
		clean := filepath.Clean(path)
		if seen[clean] {
			return
		}
		seen[clean] = true
		entries = append(entries, Entry{Name: filepath.Base(clean), Path: clean, Required: required})
	}

	// 主库优先：它排在包内第一位，且缺了就是路径配错，不是「这个部署没用这个功能」。
	add(filepath.Join(l.OutputDir, name), true)

	// 运行时发现：这两个目录下的其它 .db 也一并带上（只扫一层，不递归）。
	// -wal / -shm 不会被匹配到：glob 只收 *.db，而它们的名字是 .db-wal / .db-shm。
	for _, dir := range []string{l.OutputDir, l.DatabaseDir} {
		matches, err := filepath.Glob(filepath.Join(dir, "*.db"))
		if err != nil {
			return nil, fmt.Errorf("扫描 %s 失败: %w", dir, err)
		}
		sort.Strings(matches)
		for _, m := range matches {
			add(m, false)
		}
	}

	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		info, err := os.Stat(e.Path)
		switch {
		case err != nil && e.Required:
			// 主库缺席：报错，绝不 sql.Open 出一个空库来充当备份。
			return nil, fmt.Errorf("主数据库不可用: %w", err)
		case err != nil:
			// 可选库缺席：这个部署可能压根没用过收藏夹 / 稍后再看。出声，但不失败。
			fmt.Fprintf(stderr, "跳过 %s：%v\n", e.Name, err)
			continue
		case info.IsDir():
			if e.Required {
				return nil, fmt.Errorf("主数据库 %s 是目录，不是数据库文件", e.Path)
			}
			fmt.Fprintf(stderr, "跳过 %s：是目录，不是数据库文件\n", e.Name)
			continue
		case info.Size() == 0:
			if e.Required {
				return nil, fmt.Errorf("主数据库 %s 是空文件", e.Path)
			}
			fmt.Fprintf(stderr, "跳过 %s：空文件，不导出空库\n", e.Name)
			continue
		}
		if !e.Required && !isKnown(e.Name) {
			// 代码清单之外的库：宁可多带一个，不能漏带一个。
			fmt.Fprintf(stderr, "发现清单之外的数据库，一并打包: %s\n", e.Name)
		}
		out = append(out, e)
	}

	// 包内平铺，条目名必须唯一，否则恢复时不知道往哪放回。
	byName := make(map[string]bool, len(out))
	for _, e := range out {
		if byName[e.Name] {
			return nil, fmt.Errorf("包内条目名 %s 重复，无法确定恢复位置", e.Name)
		}
		byName[e.Name] = true
	}

	// 已知清单里没被扫到的，也要报一句「不见了」，别让它悄悄蒸发。
	// 缺席的可选库在上面那轮扫描里根本不会出现，所以这里才是它们的声音。
	for _, n := range optionalDBFiles {
		found := false
		for _, e := range out {
			if e.Name == n {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(stderr, "已知数据库 %s 不在数据目录里，本次未备份\n", n)
		}
	}

	if len(out) == 0 {
		return nil, errors.New("没有找到任何可备份的数据库")
	}
	return out, nil
}

func isKnown(name string) bool {
	for _, n := range optionalDBFiles {
		if n == name {
			return true
		}
	}
	return false
}

// openForExport 以「只做导出」的方式打开既有数据库：不建目录、不跑迁移、不写种子。
//
// 备份进程与应用主进程并发：表结构升级只应由应用自己做，一个「顺手打开数据库」的
// 进程不该把 schema 改了，更不该和它抢写。文件在不在由 Discover 判过才走到这里。
//
// DSN 刻意不设 journal_mode——导出连接不得改写库头部（WAL 标志是持久写在库头里的）；
// busy_timeout 兜住与主进程偶发的读锁冲突。已有 -wal 里的数据照常读得到：
// VACUUM INTO 读的是含 WAL 的完整逻辑视图。
func openForExport(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)", path))
	if err != nil {
		return nil, fmt.Errorf("打开 %s 失败: %w", path, err)
	}
	// 导出是顺序读，一条连接足够，也避免多余连接各自开 -shm 句柄。
	conn.SetMaxOpenConns(1)
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("连接 %s 失败: %w", path, err)
	}
	return conn, nil
}
