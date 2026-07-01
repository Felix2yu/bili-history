package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	databases = make(map[string]*sql.DB)
	dbMu      sync.RWMutex
)

// Open opens (or returns a cached) SQLite database at the given path.
// Uses WAL mode for better concurrent read performance while maintaining
// full compatibility with Python sqlite3 databases.
func Open(dbPath string) (*sql.DB, error) {
	dbMu.RLock()
	if db, ok := databases[dbPath]; ok {
		dbMu.RUnlock()
		return db, nil
	}
	dbMu.RUnlock()

	dbMu.Lock()
	defer dbMu.Unlock()

	// Double-check
	if db, ok := databases[dbPath]; ok {
		return db, nil
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
	}

	// DSN compatible with Python sqlite3:
	// - WAL journal mode (better than DELETE for concurrent reads)
	// - busy_timeout to handle lock contention
	// - foreign_keys=on for integrity
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", dbPath, err)
	}

	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)

	// Apply PRAGMAs that Python sets on each connection
	applyPragmas(db)

	databases[dbPath] = db
	return db, nil
}

// applyPragmas sets PRAGMA values matching the Python backend.
func applyPragmas(db *sql.DB) {
	pragmas := []string{
		"PRAGMA legacy_file_format = ON",
		"PRAGMA synchronous = NORMAL",
	}
	for _, p := range pragmas {
		db.Exec(p)
	}
}

// OpenHistoryDB opens the main bilibili_history database.
func OpenHistoryDB() (*sql.DB, error) {
	dbPath := filepath.Join(OutputDir(), "database", "bilibili_history.db")
	return Open(dbPath)
}

// OpenSchedulerDB opens the scheduler database.
func OpenSchedulerDB() (*sql.DB, error) {
	dbPath := filepath.Join(OutputDir(), "database", "scheduler.db")
	return Open(dbPath)
}

// OpenFavoritesDB opens the favorites database.
func OpenFavoritesDB() (*sql.DB, error) {
	dbPath := filepath.Join(OutputDir(), "database", "favorites.db")
	return Open(dbPath)
}

// OpenDynamicDB opens the dynamic database.
func OpenDynamicDB() (*sql.DB, error) {
	dbPath := filepath.Join(OutputDir(), "database", "dynamic.db")
	return Open(dbPath)
}

// OpenInteractionDB opens the interaction records database.
func OpenInteractionDB() (*sql.DB, error) {
	dbPath := filepath.Join(OutputDir(), "database", "interaction.db")
	return Open(dbPath)
}

// OutputDir returns the base output directory path.
func OutputDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "output")
}

// CloseAll closes all cached database connections.
func CloseAll() {
	dbMu.Lock()
	defer dbMu.Unlock()
	for path, db := range databases {
		db.Close()
		delete(databases, path)
	}
}

// GetYearTable returns the table name for a given year.
func GetYearTable(year int) string {
	return fmt.Sprintf("bilibili_history_%d", year)
}

// EnsureHistoryTable creates the history table with the exact Python schema.
func EnsureHistoryTable(database *sql.DB, year int) error {
	tableName := GetYearTable(year)
	_, err := database.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bvid TEXT NOT NULL,
			aid INTEGER,
			title TEXT,
			desc TEXT,
			pic TEXT,
			duration INTEGER,
			owner_name TEXT,
			owner_mid INTEGER,
			tag_name TEXT,
			tid INTEGER,
			view_at INTEGER NOT NULL,
			progress INTEGER,
			business TEXT,
			view INTEGER,
			danmaku INTEGER,
			coin INTEGER,
			favorite INTEGER,
			like INTEGER,
			reply INTEGER,
			share INTEGER,
			PRIMARY KEY (bvid, view_at)
		)
	`, tableName))
	if err != nil {
		return fmt.Errorf("create table %s: %w", tableName, err)
	}

	// Create indexes matching Python
	database.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_bvid ON %s(bvid)", tableName, tableName))
	database.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_view_at ON %s(view_at)", tableName, tableName))
	database.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_owner ON %s(owner_name)", tableName, tableName))

	return nil
}

// MigrateIfNeeded adds missing columns to existing tables for forward compatibility.
func MigrateIfNeeded(database *sql.DB, year int) error {
	tableName := GetYearTable(year)

	// List of columns that may be missing in older databases
	columns := []struct {
		name string
		def  string
	}{
		{"aid", "INTEGER DEFAULT 0"},
		{"desc", "TEXT DEFAULT ''"},
		{"tag_name", "TEXT DEFAULT ''"},
		{"tid", "INTEGER DEFAULT 0"},
		{"business", "TEXT DEFAULT ''"},
		{"view", "INTEGER DEFAULT 0"},
		{"danmaku", "INTEGER DEFAULT 0"},
		{"coin", "INTEGER DEFAULT 0"},
		{"favorite", "INTEGER DEFAULT 0"},
		{"like", "INTEGER DEFAULT 0"},
		{"reply", "INTEGER DEFAULT 0"},
		{"share", "INTEGER DEFAULT 0"},
	}

	// Get existing columns
	rows, err := database.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil // Table doesn't exist yet, OK
	}
	defer rows.Close()

	existingCols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt interface{}
		var pk int
		rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
		existingCols[name] = true
	}

	// Add missing columns
	for _, col := range columns {
		if !existingCols[col.name] {
			sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, col.name, col.def)
			database.Exec(sql)
		}
	}

	return nil
}
