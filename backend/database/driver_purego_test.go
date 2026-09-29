package database

import (
	"database/sql"
	"testing"
)

// 守卫：约束本仓库使用纯 Go SQLite 驱动（modernc.org/sqlite，注册名 "sqlite"）。
//
// 背景：mattn/go-sqlite3 在 CGO_ENABLED=0 时会被 build constraints 静默排除——
// 编译照常通过、CI 全绿，但驱动从未注册，容器一启动就
// "sql: unknown driver \"sqlite3\""。这个测试把该失败提前到 CI。
//
// 若将来换回 mattn，请把这里改为 "sqlite3" 并同步 database/sqlite.go 的打开调用。
func TestSQLiteDriverRegistered(t *testing.T) {
	want := "sqlite"
	var found bool
	for _, d := range sql.Drivers() {
		if d == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("driver %q not registered; registered drivers: %v", want, sql.Drivers())
	}

	// 与 extras.go 使用的 DSN 参数保持一致，确认 _busy_timeout / _journal_mode 被接受
	db, err := sql.Open(want, ":memory:?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("create table drv_check(a int)"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := db.Exec("insert into drv_check values (1)"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var n int
	if err := db.QueryRow("select count(*) from drv_check").Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 row, got %d", n)
	}
	t.Logf("pure-Go sqlite OK (rows=%d, drivers=%v)", n, sql.Drivers())
}
