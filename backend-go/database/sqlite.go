package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"bilibili-history-go/utils"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

var (
	sqliteInstance *SQLiteDB
	sqliteOnce     sync.Once
)

func GetSQLiteDB() *SQLiteDB {
	sqliteOnce.Do(func() {
		sqliteInstance = &SQLiteDB{}
		sqliteInstance.init()
	})
	return sqliteInstance
}

func (s *SQLiteDB) init() {
	dbPath := utils.GetDBFilePath()
	s.path = dbPath

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		utils.LogError("Failed to open SQLite database: %v", err)
		return
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	_, err = db.Exec(`
		PRAGMA journal_mode = DELETE;
		PRAGMA synchronous = NORMAL;
		PRAGMA legacy_file_format = 1;
		PRAGMA user_version = 317;
	`)
	if err != nil {
		utils.LogWarning("Failed to set SQLite pragmas: %v", err)
	}

	s.db = db
	utils.LogSuccess("SQLite database initialized: %s", dbPath)
}

func (s *SQLiteDB) GetDB() *sql.DB {
	return s.db
}

func (s *SQLiteDB) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}

func (s *SQLiteDB) GetAvailableYears() ([]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := s.db.Query(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name LIKE 'bilibili_history_%'
		ORDER BY name DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		var year int
		_, err := fmt.Sscanf(tableName, "bilibili_history_%d", &year)
		if err != nil {
			continue
		}
		years = append(years, year)
	}

	if len(years) == 0 {
		currentYear := time.Now().Year()
		years = append(years, currentYear)
	}

	return years, nil
}

func (s *SQLiteDB) TableExists(tableName string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return false, fmt.Errorf("database not initialized")
	}

	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM sqlite_master 
			WHERE type='table' AND name=?
		)
	`, tableName).Scan(&exists)

	return exists, err
}

func (s *SQLiteDB) EnsureTableForYear(year int) error {
	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, err := s.TableExists(tableName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	createSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
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
	);
	`, tableName)

	_, err = s.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}

	indexSQLs := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_author_mid ON %s (author_mid);", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_view_at ON %s (view_at);", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_remark_time ON %s (remark_time);", tableName, tableName),
	}

	for _, idxSQL := range indexSQLs {
		_, err = s.db.Exec(idxSQL)
		if err != nil {
			utils.LogWarning("Failed to create index: %v", err)
		}
	}

	return nil
}

func (s *SQLiteDB) ResetDatabase() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		s.db.Close()
		s.db = nil
	}

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete database file: %v", err)
	}

	lastImportPath := utils.GetOutputPath("last_import.json")
	if err := os.Remove(lastImportPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete last_import.json: %v", err)
	}

	s.init()
	return nil
}

func (s *SQLiteDB) GetVersionInfo() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	result["database_file"] = map[string]interface{}{
		"exists": true,
		"path":   s.path,
	}

	if s.db == nil {
		return result, nil
	}

	var sqliteVersion string
	s.db.QueryRow("SELECT sqlite_version()").Scan(&sqliteVersion)
	result["sqlite_version"] = sqliteVersion

	var userVersion int
	s.db.QueryRow("PRAGMA user_version").Scan(&userVersion)
	result["user_version"] = userVersion

	dbSettings := make(map[string]interface{})

	var journalMode string
	s.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	dbSettings["journal_mode"] = journalMode

	var synchronous string
	s.db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	dbSettings["synchronous"] = synchronous

	var pageSize int
	s.db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	dbSettings["page_size"] = pageSize

	var cacheSize int
	s.db.QueryRow("PRAGMA cache_size").Scan(&cacheSize)
	dbSettings["cache_size"] = cacheSize

	var encoding string
	s.db.QueryRow("PRAGMA encoding").Scan(&encoding)
	dbSettings["encoding"] = encoding

	result["database_settings"] = dbSettings

	fileInfo, err := os.Stat(s.path)
	if err == nil {
		size := fileInfo.Size()
		dbFile := result["database_file"].(map[string]interface{})
		dbFile["size_bytes"] = size
		dbFile["size_mb"] = float64(size) / (1024 * 1024)
	}

	return result, nil
}
