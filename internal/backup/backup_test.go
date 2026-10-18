package backup

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 备份子命令的契约测试。
//
// 一律用临时目录搭出「output 根目录 + output/database」的真实布局，
// 这样发现逻辑与 CLI 走的是同一条代码路径；数据全是合成行，不含真实账号数据。

// newLayout 造一个只有主库的导出布局。
func newLayout(t *testing.T) Layout {
	t.Helper()
	base := t.TempDir()
	return Layout{
		OutputDir:   filepath.Join(base, "output"),
		DatabaseDir: filepath.Join(base, "output", "database"),
		PrimaryName: "bilibili_history.db",
	}
}

// createDB 在 path 建一个带一张表的 SQLite 库（WAL 模式，与应用运行时的形态一致）。
// 返回打开的写连接：调用方不关掉它，已提交的行就留在 -wal 里——
// 正是备份工具要在应用活跃写入时采集的那种状态。
func createDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`
		PRAGMA journal_mode = WAL;
		CREATE TABLE IF NOT EXISTS probe (id INTEGER PRIMARY KEY, note TEXT NOT NULL);
	`); err != nil {
		t.Fatalf("schema %s: %v", path, err)
	}
	return conn
}

// insert 写一行并留在 WAL 里（不做 checkpoint，不关连接）。
func insert(t *testing.T, conn *sql.DB, note string) {
	t.Helper()
	if _, err := conn.Exec("INSERT INTO probe(note) VALUES (?)", note); err != nil {
		t.Fatalf("insert: %v", err)
	}
}

// runBackup 跑一次导出，返回 stdout 字节与 stderr 字节。
func runBackup(t *testing.T, layout Layout, args ...string) (stdout, stderr []byte) {
	t.Helper()
	var out, errOut bytes.Buffer
	if err := Execute(args, layout, &out, &errOut); err != nil {
		t.Fatalf("Execute(%v) 失败: %v", args, err)
	}
	return out.Bytes(), errOut.Bytes()
}

// openZip 把产物当压缩包打开。
func openZip(t *testing.T, data []byte) *zip.Reader {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("产物不是合法压缩包: %v", err)
	}
	return zr
}

// readEntry 取出一个条目的全部字节。
func readEntry(t *testing.T, zr *zip.Reader, name string) []byte {
	t.Helper()
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open entry %s: %v", name, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read entry %s: %v", name, err)
		}
		return data
	}
	t.Fatalf("压缩包里缺条目 %s", name)
	return nil
}

// requireIntact 把快照落成临时文件，确认魔数正确且 PRAGMA integrity_check 为 ok。
// 用本仓库自带的纯 Go 驱动复核，与真实进程用 /usr/bin/sqlite3 的验证同口径。
func requireIntact(t *testing.T, data []byte) string {
	t.Helper()
	if len(data) < len(sqliteMagic) || string(data[:len(sqliteMagic)]) != sqliteMagic {
		t.Fatalf("不是 SQLite 文件，头部 %q", data[:min(len(data), 16)])
	}
	path := filepath.Join(t.TempDir(), "check.db")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	var verdict string
	if err := conn.QueryRow("PRAGMA integrity_check").Scan(&verdict); err != nil {
		t.Fatalf("integrity_check: %v", err)
	}
	if verdict != "ok" {
		t.Fatalf("integrity_check = %q，想要 ok", verdict)
	}
	return path
}

func hasNote(t *testing.T, dbPath, note string) bool {
	t.Helper()
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	var n int
	if err := conn.QueryRow("SELECT count(*) FROM probe WHERE note = ?", note).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	return n == 1
}

func sha256Of(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("hash: %v", err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ---------- (a) 默认 zip：每个库一个条目，各自自洽 ----------

func TestDefaultFormatIsZipWithEveryDB(t *testing.T) {
	layout := newLayout(t)
	writers := map[string]*sql.DB{}
	names := []string{
		"bilibili_history.db",
		"bilibili_dynamic.db",
		"bilibili_likes.db",
		"bilibili_watchlater.db",
		"bilibili_favorites.db",
		"scheduler.db",
	}
	for _, n := range names {
		dir := layout.OutputDir
		if n != layout.PrimaryName {
			dir = layout.DatabaseDir
		}
		p := filepath.Join(dir, n)
		writers[n] = createDB(t, p)
		insert(t, writers[n], "seed-"+n)
	}

	stdout, _ := runBackup(t, layout) // 不传 --format，默认必须是 zip
	if len(stdout) < 4 || string(stdout[:2]) != "PK" {
		t.Fatalf("默认产物不是 zip（前两字节 %q）", stdout[:2])
	}

	zr := openZip(t, stdout)
	if len(zr.File) != len(names) {
		t.Fatalf("包内条目 %d 个，想要 %d 个：%s", len(zr.File), len(names), zipNames(zr))
	}
	for _, n := range names {
		f := findEntry(t, zr, n)
		if strings.Contains(f.Name, string(filepath.Separator)) {
			t.Errorf("条目 %q 不该带目录成分", f.Name)
		}
		data := readEntry(t, zr, n)
		requireIntact(t, data)
	}
}

// lookupEntry 找条目但不判定，让调用方能自己决定是告警还是失败。
func lookupEntry(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func findEntry(t *testing.T, zr *zip.Reader, name string) *zip.File {
	t.Helper()
	f := lookupEntry(zr, name)
	if f == nil {
		t.Fatalf("压缩包里缺条目 %s（实有 %s）", name, zipNames(zr))
	}
	return f
}

func zipNames(zr *zip.Reader) string {
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return strings.Join(names, ", ")
}

// ---------- (b) 刚提交、还在 -wal 里的行必须被带上 ----------

func TestRowStillInWALIsExported(t *testing.T) {
	layout := newLayout(t)
	primary := filepath.Join(layout.OutputDir, layout.PrimaryName)
	extra := filepath.Join(layout.DatabaseDir, "bilibili_favorites.db")

	hPrimary := createDB(t, primary)
	hExtra := createDB(t, extra)

	notePrimary := "此刻就在 wal 里的历史行"
	noteExtra := "此刻就在 wal 里的收藏行"
	insert(t, hPrimary, notePrimary)
	insert(t, hExtra, noteExtra)

	// 两个库都得有非空的 -wal：证明了数据确实只在 WAL 里，而不是已经并回主文件。
	for _, p := range []string{primary, extra} {
		wal := p + "-wal"
		info, err := os.Stat(wal)
		if err != nil {
			t.Fatalf("%s 应该有 -wal 文件: %v", filepath.Base(p), err)
		}
		if info.Size() == 0 {
			t.Fatalf("%s 的 -wal 是空的，这个用例没有覆盖到 WAL 中的行", filepath.Base(p))
		}
	}

	stdout, _ := runBackup(t, layout)
	zr := openZip(t, stdout)

	for entry, tc := range map[string]struct {
		note string
		want string
	}{
		"bilibili_history.db":   {notePrimary, "主库"},
		"bilibili_favorites.db": {noteExtra, "收藏库"},
	} {
		path := requireIntact(t, readEntry(t, zr, entry))
		if !hasNote(t, path, tc.note) {
			t.Errorf("%s（%s）里没有那条 WAL 中的行", tc.want, entry)
		}
	}
}

// ---------- (c) --format db：裸的主库快照 ----------

func TestFormatDBEmitsBarePrimarySnapshot(t *testing.T) {
	layout := newLayout(t)
	primary := filepath.Join(layout.OutputDir, layout.PrimaryName)
	h := createDB(t, primary)
	note := "db 形态下也要在的行"
	insert(t, h, note)

	// db 形态是有损的：其它库照旧存在，但不该进产物。
	createAndFill(t, layout, "bilibili_dynamic.db", "不该出现在 db 产物里")

	stdout, stderr := runBackup(t, layout, "--format", "db")
	if len(stdout) < len(sqliteMagic) || string(stdout[:len(sqliteMagic)]) != sqliteMagic {
		t.Fatalf("--format db 的产物不是裸 SQLite 文件")
	}
	if _, err := zip.NewReader(bytes.NewReader(stdout), int64(len(stdout))); err == nil {
		t.Fatal("--format db 不该是个压缩包")
	}
	if !strings.Contains(string(stderr), "db 只导出主数据库") {
		t.Errorf("db 形态是有损备份，stderr 上该有一句提醒: %q", stderr)
	}
	path := requireIntact(t, stdout)
	if !hasNote(t, path, note) {
		t.Error("db 快照里缺那条记录")
	}
}

func createAndFill(t *testing.T, layout Layout, name, note string) {
	t.Helper()
	h := createDB(t, filepath.Join(layout.DatabaseDir, name))
	insert(t, h, note)
}

// ---------- (d) 主库缺席：非零失败，stdout 一个字节都不能有 ----------

func TestMissingPrimaryFailsWithoutWritingStdout(t *testing.T) {
	layout := newLayout(t)
	// 只建可选库，主库缺席——路径配错的样子。
	createAndFill(t, layout, "bilibili_likes.db", "有库但没有主库")

	var out, errOut bytes.Buffer
	err := Execute(nil, layout, &out, &errOut)
	if err == nil {
		t.Fatal("主库缺席时 backup 必须失败")
	}
	if out.Len() != 0 {
		t.Fatalf("失败时 stdout 必须是 0 字节，实际 %d 字节", out.Len())
	}
	if !strings.Contains(err.Error(), "主数据库") {
		t.Errorf("失败原因要点明是主库: %v", err)
	}

	// 也别在数据目录里留下半成品的 db 文件。
	matches, _ := filepath.Glob(filepath.Join(layout.OutputDir, "*.db"))
	for _, m := range matches {
		if filepath.Base(m) == layout.PrimaryName {
			t.Fatalf("导出失败后不该存在 %s", m)
		}
	}
}

func TestEmptyPrimaryFileFails(t *testing.T) {
	// 空文件是 sql.Open 造出来的空库最常见的形态，绝不能当成合法备份。
	layout := newLayout(t)
	if err := os.MkdirAll(layout.OutputDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(layout.OutputDir, layout.PrimaryName), nil, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := os.Stat(layout.DatabaseDir); err == nil {
		t.Fatal("发现路径不该建目录")
	}

	var out bytes.Buffer
	if err := Execute(nil, layout, &out, os.Stderr); err == nil {
		t.Fatal("空主库必须失败")
	}
	if out.Len() != 0 {
		t.Fatalf("失败时 stdout 必须是 0 字节，实际 %d 字节", out.Len())
	}
}

// ---------- (e) 可选库缺席：告警、跳过，归档照样有效 ----------

func TestMissingOptionalDBIsSkippedWithWarning(t *testing.T) {
	layout := newLayout(t)
	primary := filepath.Join(layout.OutputDir, layout.PrimaryName)
	h := createDB(t, primary)
	note := "只有主库的部署"
	insert(t, h, note)

	// 这个部署没用过收藏夹，也不跑调度器：两个库就是不存在。
	stdout, stderr := runBackup(t, layout)

	warned := string(stderr)
	for _, missing := range []string{"bilibili_favorites.db", "scheduler.db"} {
		if !strings.Contains(warned, missing) || !strings.Contains(warned, "未备份") {
			t.Errorf("缺席的可选库 %s 应该在 stderr 上出声，实际: %q", missing, warned)
		}
	}

	zr := openZip(t, stdout)
	if len(zr.File) != 1 {
		t.Fatalf("包里应该只有主库一个条目，实际 %s", zipNames(zr))
	}
	requireIntact(t, readEntry(t, zr, layout.PrimaryName))
}

// ---------- (f) 导出不改源库：反复跑，源文件哈希不动 ----------

func TestExportDoesNotModifySources(t *testing.T) {
	layout := newLayout(t)
	primary := filepath.Join(layout.OutputDir, layout.PrimaryName)
	dynamic := filepath.Join(layout.DatabaseDir, "bilibili_dynamic.db")
	h1 := createDB(t, primary)
	h2 := createDB(t, dynamic)
	insert(t, h1, "跑一遍也不该被动")
	insert(t, h2, "跑两遍也不该被动")

	// 源库里先放一张应用自己都没建过的表：导出要是顺手跑了迁移，schema 就会变。
	if _, err := h1.Exec("CREATE TABLE IF NOT EXISTS marker (v TEXT)"); err != nil {
		t.Fatalf("marker: %v", err)
	}
	if _, err := h1.Exec("INSERT INTO marker VALUES ('原样')"); err != nil {
		t.Fatalf("marker insert: %v", err)
	}

	before := map[string]string{
		primary: sha256Of(t, primary),
		dynamic: sha256Of(t, dynamic),
	}

	for i := 0; i < 2; i++ {
		runBackup(t, layout)
	}

	for path, want := range before {
		if got := sha256Of(t, path); got != want {
			t.Errorf("%s 被导出进程改写了：sha256 %s → %s", filepath.Base(path), want[:12], got[:12])
		}
	}

	// 快照读得到 WAL 之外的表，说明读的是完整逻辑视图，而不是只读了主文件那一截。
	stdout, _ := runBackup(t, layout)
	path := requireIntact(t, openZipReadEntry(t, stdout, layout.PrimaryName))
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	var v string
	if err := conn.QueryRow("SELECT v FROM marker").Scan(&v); err != nil {
		t.Fatalf("快照里读不到 marker 表: %v", err)
	}
	if v != "原样" {
		t.Errorf("marker = %q", v)
	}
}

func openZipReadEntry(t *testing.T, data []byte, name string) []byte {
	t.Helper()
	return readEntry(t, openZip(t, data), name)
}

// ---------- (g) 机密配置不进包 ----------

func TestConfigAndCookiesAreNotPacked(t *testing.T) {
	layout := newLayout(t)
	h := createDB(t, filepath.Join(layout.OutputDir, layout.PrimaryName))
	insert(t, h, "带 Cookie 的部署")

	// 容器里 config.yaml 是挂载进 /app/config 的，本来就在工作目录的另一侧；
	// 这里把它同时摆到 config 目录和 output 目录里，两侧都不许进包。
	secret := "SESSDATA=超级机密的会话令牌"
	base := filepath.Dir(layout.OutputDir)
	for _, p := range []string{
		filepath.Join(base, "config", "config.yaml"),
		filepath.Join(layout.OutputDir, "config.yaml"),
	} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(p, []byte(secret+"\n"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	stdout, _ := runBackup(t, layout)
	zr := openZip(t, stdout)
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, ".yaml") || strings.HasSuffix(f.Name, ".yml") {
			t.Errorf("配置文件不该进备份包: %s", f.Name)
		}
		if !strings.HasSuffix(f.Name, ".db") {
			t.Errorf("包内只该有数据库，实际有条目 %s", f.Name)
		}
	}
	if bytes.Contains(stdout, []byte("SESSDATA")) {
		t.Error("产物里出现了 SESSDATA")
	}
	// 日志、下载物同样不该混进来。
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "logs/") || strings.HasPrefix(f.Name, "downloads/") {
			t.Errorf("非数据库内容混进了备份包: %s", f.Name)
		}
	}
}

// ---------- 参数与格式校验 ----------

func TestUnknownFormatAndExtraArgs(t *testing.T) {
	layout := newLayout(t)
	createDB(t, filepath.Join(layout.OutputDir, layout.PrimaryName))

	cases := map[string][]string{
		"未知格式":   {"--format", "tar"},
		"多余位置参数": {"extra"},
	}
	for name, args := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(args, layout, &out, &errOut); err == nil {
			t.Errorf("%s: 期望失败", name)
		}
		if out.Len() != 0 {
			t.Errorf("%s: 失败时 stdout 必须是 0 字节", name)
		}
	}
}

func TestDiscoveryDoesNotCreateDirectories(t *testing.T) {
	// 空数据目录：只问「在哪」，不许动磁盘。
	base := t.TempDir()
	layout := Layout{
		OutputDir:   filepath.Join(base, "output"),
		DatabaseDir: filepath.Join(base, "output", "database"),
		PrimaryName: "bilibili_history.db",
	}
	var errOut bytes.Buffer
	if _, err := layout.Discover(&errOut); err == nil {
		t.Fatal("主库不存在时 Discover 必须报错")
	}
	if _, err := os.Stat(layout.OutputDir); !os.IsNotExist(err) {
		t.Fatal("Discover 创建了输出目录")
	}
}

func TestCustomDBFileNameIsHonoured(t *testing.T) {
	// 主库名来自 config db_file，发现逻辑必须跟着走，不能写死 bilibili_history.db。
	layout := newLayout(t)
	layout.PrimaryName = "renamed_history.db"
	h := createDB(t, filepath.Join(layout.OutputDir, layout.PrimaryName))
	note := "改过名的主库"
	insert(t, h, note)

	stdout, _ := runBackup(t, layout)
	path := requireIntact(t, openZipReadEntry(t, stdout, "renamed_history.db"))
	if !hasNote(t, path, note) {
		t.Error("改名后的主库没被正确导出")
	}
}

func TestUnexpectedDBFileIsPackedWithWarning(t *testing.T) {
	// 代码清单之外的库（将来新增的、或手工放进去的）宁可多带不能漏带。
	layout := newLayout(t)
	createDB(t, filepath.Join(layout.OutputDir, layout.PrimaryName))
	stray := filepath.Join(layout.DatabaseDir, "bilibili_something_new.db")
	h := createDB(t, stray)
	note := "清单之外的库"
	insert(t, h, note)

	stdout, stderr := runBackup(t, layout)
	zr := openZip(t, stdout)
	if lookupEntry(zr, "bilibili_something_new.db") == nil {
		t.Fatalf("清单之外的库没进包（实有 %s）", zipNames(zr))
	}
	if !strings.Contains(string(stderr), "清单之外") {
		t.Errorf("清单之外的库该在 stderr 上说明一句: %q", stderr)
	}
	path := requireIntact(t, readEntry(t, zr, "bilibili_something_new.db"))
	if !hasNote(t, path, note) {
		t.Error("清单之外的库里的那条记录没带上")
	}
}

func TestDuplicateEntryNameFails(t *testing.T) {
	// 包内平铺，同名条目会把恢复位置弄丢——必须当场失败而不是覆盖。
	layout := newLayout(t)
	createDB(t, filepath.Join(layout.OutputDir, layout.PrimaryName))
	// 把 scheduler.db 同时在两个目录各放一份。
	createDB(t, filepath.Join(layout.OutputDir, "scheduler.db"))
	createDB(t, filepath.Join(layout.DatabaseDir, "scheduler.db"))

	var out bytes.Buffer
	err := Execute(nil, layout, &out, os.Stderr)
	if err == nil || !strings.Contains(err.Error(), "重复") {
		t.Fatalf("同名条目必须报错，实际: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("失败时 stdout 必须是 0 字节，实际 %d 字节", out.Len())
	}
}

func TestLayoutRejectsPrimaryNameWithDirectory(t *testing.T) {
	layout := newLayout(t)
	layout.PrimaryName = "database/bilibili_history.db"
	var out bytes.Buffer
	if err := Execute(nil, layout, &out, os.Stderr); err == nil {
		t.Fatal("带目录的 db_file 必须报错")
	}
}

func TestDiagnosticLogsNeverReachStdout(t *testing.T) {
	// 契约的核心一条：stdout 只有产物字节。
	// 这里让所有可选库都缺席，制造最大量的告警。
	layout := newLayout(t)
	h := createDB(t, filepath.Join(layout.OutputDir, layout.PrimaryName))
	insert(t, h, "只看告警的一轮")

	stdout, stderr := runBackup(t, layout)
	if !strings.Contains(string(stderr), "本次未备份") {
		t.Fatal("这轮本来应该产生告警，否则这个用例什么都没覆盖")
	}
	zr := openZip(t, stdout)
	for _, f := range zr.File {
		data := readEntry(t, zr, f.Name)
		requireIntact(t, data) // 告警没有污染快照的头 16 字节
	}
	// 产物末尾紧贴 zip 的中央目录，中间插进任何文本都会让 zip 解析失败——
	// 上面 openZip 成功即证明 stdout 里没有日志文本。
	if bytes.Contains(stdout, []byte("WARNING")) || bytes.Contains(stdout, []byte("跳过")) {
		t.Error("stdout 里出现了日志文本")
	}
}

// ---------- 并发写：应用正在写库时导出 ----------

func TestConcurrentWritesKeepSnapshotConsistent(t *testing.T) {
	layout := newLayout(t)
	primary := filepath.Join(layout.OutputDir, layout.PrimaryName)
	writer := createDB(t, primary)

	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			if _, err := writer.Exec("INSERT INTO probe(note) VALUES (?)", fmt.Sprintf("并发写入 %d", i)); err != nil {
				done <- err
				return
			}
			if i >= 60 {
				done <- nil
				return
			}
		}
	}()

	var last string
	for i := 0; i < 3; i++ {
		stdout, _ := runBackup(t, layout)
		path := requireIntact(t, openZipReadEntry(t, stdout, layout.PrimaryName))
		conn, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		var n int
		if err := conn.QueryRow("SELECT count(*) FROM probe").Scan(&n); err != nil {
			t.Fatalf("count: %v", err)
		}
		_ = conn.Close()
		last = fmt.Sprintf("%d", n)
		if n == 0 {
			t.Fatalf("第 %d 轮快照里没有已提交的行", i)
		}
	}
	if err := <-done; err != nil {
		t.Fatalf("并发写入出错: %v", err)
	}
	close(stop)
	t.Logf("并发写入下的快照行数: %s", last)
}
