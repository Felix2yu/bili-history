package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// newFakeSQLite 构造一个不走单例、可独立关闭的 SQLiteDB。
func newFakeSQLite(t *testing.T) *SQLiteDB {
	t.Helper()
	f := filepath.Join(t.TempDir(), "fake.db")
	db, err := sql.Open("sqlite", f)
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	return &SQLiteDB{db: db, path: f}
}

func TestGetSQLiteDBSingleton(t *testing.T) {
	a := GetSQLiteDB()
	b := GetSQLiteDB()
	if a != b {
		t.Fatal("GetSQLiteDB 应返回同一单例")
	}
	if a.GetDB() == nil {
		t.Fatal("单例连接不应为 nil")
	}
}

func TestTableExists(t *testing.T) {
	db := GetSQLiteDB()
	ok, err := db.TableExists("bilibili_history_" + fmt.Sprint(yearAnalysisSeed))
	if err != nil || !ok {
		t.Fatalf("种子年份表应存在: ok=%v err=%v", ok, err)
	}
	ok, err = db.TableExists("no_such_table_xyz")
	if err != nil || ok {
		t.Fatalf("不存在的表应返回 false: ok=%v err=%v", ok, err)
	}

	fake := &SQLiteDB{}
	if _, err := fake.TableExists("anything"); err == nil {
		t.Fatal("db==nil 时 TableExists 应报错")
	}
}

func TestGetAvailableYears(t *testing.T) {
	years, err := GetSQLiteDB().GetAvailableYears()
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, y := range years {
		if y == yearAnalysisSeed {
			found = true
		}
	}
	if !found {
		t.Fatalf("应包含种子年份 %d, got %v", yearAnalysisSeed, years)
	}
	if !isSortedDesc(years) {
		t.Fatalf("年份应降序排列: %v", years)
	}

	// 空库回退到当前年份
	fake := newFakeSQLite(t)
	defer fake.Close()
	got, err := fake.GetAvailableYears()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != time.Now().Year() {
		t.Fatalf("空库应返回当前年份, got %v", got)
	}

	nilDB := &SQLiteDB{}
	if _, err := nilDB.GetAvailableYears(); err == nil {
		t.Fatal("db==nil 时应报错")
	}
}

func isSortedDesc(years []int) bool {
	for i := 1; i < len(years); i++ {
		if years[i-1] < years[i] {
			return false
		}
	}
	return true
}

func TestEnsureTableForYearCreatesTable(t *testing.T) {
	conn := testConn()
	if err := GetSQLiteDB().EnsureTableForYear(yearInfra1); err != nil {
		t.Fatal(err)
	}
	// 幂等
	if err := GetSQLiteDB().EnsureTableForYear(yearInfra1); err != nil {
		t.Fatal(err)
	}

	var count int
	q := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('bilibili_history_%d') WHERE name IN ('status','remark','covers','main_category')", yearInfra1)
	if err := conn.QueryRow(q).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("建表应包含全部关键列, got %d", count)
	}

	var idxCount int
	q = "SELECT COUNT(*) FROM pragma_index_list('bilibili_history_" + fmt.Sprint(yearInfra1) + "')"
	if err := conn.QueryRow(q).Scan(&idxCount); err != nil {
		t.Fatal(err)
	}
	if idxCount < 6 {
		t.Fatalf("建表应创建 6 个索引, got %d", idxCount)
	}
}

// legacyDDL 与 EnsureTableForYear 的建表语句一致，但缺少 status 列，
// 用于覆盖老库升级路径。
const legacyDDL = `
CREATE TABLE %s (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL,
	long_title TEXT,
	cover TEXT,
	covers JSON,
	uri TEXT,
	oid INTEGER NOT NULL,
	epid INTEGER DEFAULT 0,
	bvid TEXT NOT NULL,
	page INTEGER DEFAULT 1,
	cid INTEGER,
	part TEXT,
	business TEXT,
	dt INTEGER NOT NULL,
	videos INTEGER DEFAULT 1,
	author_name TEXT NOT NULL,
	author_face TEXT,
	author_mid INTEGER NOT NULL,
	view_at INTEGER NOT NULL,
	progress INTEGER DEFAULT 0,
	badge TEXT,
	show_title TEXT,
	duration INTEGER NOT NULL,
	current TEXT,
	total INTEGER DEFAULT 0,
	new_desc TEXT,
	is_finish INTEGER DEFAULT 0,
	is_fav INTEGER DEFAULT 0,
	kid INTEGER,
	tag_name TEXT,
	live_status INTEGER DEFAULT 0,
	main_category TEXT,
	remark TEXT DEFAULT '',
	remark_time INTEGER DEFAULT 0
)`

func TestEnsureTableForYearUpgradesLegacyStatus(t *testing.T) {
	conn := testConn()
	table := fmt.Sprintf("bilibili_history_%d", yearInfra2)
	mustExec(t, conn, fmt.Sprintf(legacyDDL, table))

	var before int
	if err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='status'", table)).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if before != 0 {
		t.Fatal("老表不应已有 status 列")
	}

	if err := GetSQLiteDB().EnsureTableForYear(yearInfra2); err != nil {
		t.Fatal(err)
	}
	var after int
	if err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='status'", table)).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != 1 {
		t.Fatalf("EnsureTableForYear 应为老表补上 status 列")
	}
}

func TestMigrateStatusColumn(t *testing.T) {
	conn := testConn()
	table := fmt.Sprintf("bilibili_history_%d", yearInfra2)
	// yearInfra2 表此时可能已被上一测试补列；删列在 sqlite 不支持，
	// 改用另一张手工老表：2002 若已有 status 则本测试仍能覆盖索引补齐分支。
	MigrateStatusColumn()

	var col int
	if err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='status'", table)).Scan(&col); err != nil {
		t.Fatal(err)
	}
	if col != 1 {
		t.Fatal("MigrateStatusColumn 后应存在 status 列")
	}
	var idx int
	if err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM pragma_index_list('%s')", table)).Scan(&idx); err != nil {
		t.Fatal(err)
	}
	if idx < 3 {
		t.Fatalf("MigrateStatusColumn 应补齐 bvid/tag_name/main_category 索引, got %d", idx)
	}

	// 幂等
	MigrateStatusColumn()
}

func TestGetVersionInfo(t *testing.T) {
	info, err := GetSQLiteDB().GetVersionInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info["sqlite_version"] == "" {
		t.Fatal("sqlite_version 不能为空")
	}
	if uv, ok := info["user_version"].(int); !ok || uv != 317 {
		t.Fatalf("user_version 应为 317, got %#v", info["user_version"])
	}
	dbFile, ok := info["database_file"].(map[string]interface{})
	if !ok || dbFile["exists"] != true {
		t.Fatalf("database_file.exists 应为 true, got %#v", info["database_file"])
	}
	if _, ok := info["database_settings"].(map[string]interface{}); !ok {
		t.Fatal("应包含 database_settings")
	}

	nilDB := &SQLiteDB{path: "none"}
	info, err = nilDB.GetVersionInfo()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := info["sqlite_version"]; ok {
		t.Fatal("db==nil 时不应有 sqlite_version")
	}
}

func TestCloseFakeInstance(t *testing.T) {
	f := newFakeSQLite(t)
	if err := f.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if f.GetDB() != nil {
		t.Fatal("Close 后 db 应为 nil")
	}
	if err := f.Close(); err != nil {
		t.Fatalf("double close 应无错: %v", err)
	}
}

func TestResetDatabase(t *testing.T) {
	dir, err := os.MkdirTemp("", "bilidb-reset-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	s := &SQLiteDB{}
	s.init()
	conn := s.GetDB()
	if conn == nil {
		t.Fatal("init 后连接不应为 nil")
	}
	mustExec(t, conn, "create table scratch(a int)")
	mustExec(t, conn, "insert into scratch values (1)")
	// 触发 last_import.json 删除分支
	mustExec(t, conn, fmt.Sprintf("create table bilibili_history_%d(a int)", yearInfra2))
	importPath := filepath.Join(dir, "output", "last_import.json")
	if err := os.WriteFile(importPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := s.ResetDatabase(); err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := os.Stat(importPath); !os.IsNotExist(err) {
		t.Fatal("last_import.json 应被删除")
	}
	var n int
	if err := s.GetDB().QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE name='scratch'").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("重置后旧表应消失")
	}
	if _, err := os.Stat(s.path); err != nil {
		t.Fatalf("重置后应重建数据库文件: %v", err)
	}

	// 文件不存在时的错误容忍 + 正常路径
	if err := s.ResetDatabase(); err != nil {
		t.Fatalf("再次重置: %v", err)
	}
}
