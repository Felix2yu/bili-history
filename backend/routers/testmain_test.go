package routers

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
)

// testCWD is the temporary working directory created by TestMain. Every test in
// this package runs against the local sqlite databases rooted there; nothing
// touches the network.
var testCWD string

const (
	// Reserved table years used to namespace rows created by the routers test
	// suite. 2006/2007/2012/2016 are in the past, so the production queries
	// that look at "recent" windows (last 7/30 days, current year) never mix
	// seeded rows with rows written by a different test file.
	yearPrimary   = 2016
	yearSecondary = 2006
	yearTertiary  = 2007
	yearQuaternary = 2012
	// yearEmpty holds a table with no rows (the viewing/stats zero path);
	// yearReport is reserved for the report fixtures, whose routes hardcode 2010.
	yearEmpty  = 2011
	yearReport = 2010
	// yearViewing belongs to the viewing/analytics fixtures in analysis_test.go.
	yearViewing = 2014
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "routers-test-")
	if err != nil {
		panic(err)
	}
	testCWD = dir
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}

	// The suite must not inherit credentials or server overrides from the
	// developer environment: config.LoadConfig applies them to the singleton.
	for _, key := range []string{"SESSDATA", "BILI_JCT", "DedeUserID", "DedeUserID__ckMd5", "SERVER_HOST", "SERVER_PORT"} {
		_ = os.Unsetenv(key)
	}

	if err := os.MkdirAll(filepath.Join(dir, "config"), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config", "config.yaml"), []byte(minimalConfigYAML), 0o644); err != nil {
		panic(err)
	}

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

const minimalConfigYAML = `SESSDATA: ""
bili_jct: ""
DedeUserID: ""
DedeUserID__ckMd5: ""
output_folder: ""
db_file: "test_history.db"
log_file: "test.log"
log_folder: ""
categories_file: "video_categories.json"
fields_to_remove:
  - rank
  - pt_data
server:
  host: "127.0.0.1"
  port: 8899
  ssl_enabled: false
appearance:
  dark_mode: false
notify:
  enabled: false
  urls: []
`

// loadCfg returns the process-wide config singleton. It is cached by
// config.LoadConfig (sync.Once) so every test shares it.
func loadCfg(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.GetConfig()
	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
	}
	if cfg == nil {
		t.Fatal("config singleton is nil")
	}
	return cfg
}

func sqliteDB(t *testing.T) *database.SQLiteDB {
	t.Helper()
	d := database.GetSQLiteDB()
	if d == nil {
		t.Fatal("sqlite db is nil")
	}
	return d
}

func db(t *testing.T) *sql.DB {
	t.Helper()
	d := sqliteDB(t).GetDB()
	if d == nil {
		t.Fatal("sqlite handle is nil")
	}
	return d
}

// ensureYear creates the per-year history table when missing.
func ensureYear(t *testing.T, year int) {
	t.Helper()
	if err := sqliteDB(t).EnsureTableForYear(year); err != nil {
		t.Fatalf("ensure table for %d: %v", year, err)
	}
}

// insertHistory builds a single history row inside the reserved test years.
// All rows are inserted through raw SQL so that the suite controls every
// column (view_at, business, duration, ...) precisely. Columns the table
// declares NOT NULL fall back to deterministic defaults when the caller omits
// them, so a partial fixture never trips the schema.
func insertHistory(t *testing.T, year int, h map[string]interface{}) {
	t.Helper()
	ensureYear(t, year)

	viewAtCol := int64(0)
	if v, ok := h["view_at"]; ok {
		if i, ok := v.(int64); ok {
			viewAtCol = i
		}
	}
	defaults := map[string]interface{}{
		"title":         "未命名视频",
		"long_title":    "未命名视频",
		"cover":         "",
		"covers":        "",
		"uri":           "",
		"oid":           0,
		"epid":          0,
		"bvid":          uniqueBvid("BVt", int(viewAtCol)%1000000),
		"page":          1,
		"cid":           0,
		"part":          "PART",
		"business":      "video",
		"dt":            viewAtCol,
		"videos":        1,
		"author_name":   "测试UP主",
		"author_face":   "",
		"author_mid":    100001,
		"progress":      0,
		"badge":         "",
		"show_title":    "未命名视频",
		"duration":      100,
		"current":       "1",
		"total":         1,
		"new_desc":      "",
		"is_finish":     0,
		"is_fav":        0,
		"kid":           0,
		"tag_name":      "",
		"live_status":   0,
		"main_category": "",
		"remark":        "",
		"remark_time":   0,
		"status":        0,
	}
	for k, v := range defaults {
		if _, ok := h[k]; !ok {
			h[k] = v
		}
	}

	// Fixture seeding must be idempotent: the sqlite tables are process-wide
	// singletons, so a second test that seeds the same logical row would
	// otherwise double every aggregate count the first test asserts on. The id
	// column is autoincrement, so INSERT OR REPLACE alone still appends.
	if bvid, ok := h["bvid"].(string); ok && bvid != "" {
		var existing int
		q := fmt.Sprintf("SELECT COUNT(*) FROM bilibili_history_%d WHERE bvid = ? AND view_at = ?", year)
		if err := db(t).QueryRow(q, bvid, viewAtCol).Scan(&existing); err == nil && existing > 0 {
			return
		}
	}

	cols := []string{"id", "title", "long_title", "cover", "covers", "uri", "oid", "epid", "bvid", "page",
		"cid", "part", "business", "dt", "videos", "author_name", "author_face", "author_mid", "view_at",
		"progress", "badge", "show_title", "duration", "current", "total", "new_desc", "is_finish", "is_fav",
		"kid", "tag_name", "live_status", "main_category", "remark", "remark_time", "status"}
	vals := make([]interface{}, 0, len(cols))
	for _, c := range cols {
		if v, ok := h[c]; ok {
			vals = append(vals, v)
		} else {
			vals = append(vals, nil)
		}
	}
	placeholders := strings.Repeat("?,", len(cols))
	placeholders = placeholders[:len(placeholders)-1]
	q := fmt.Sprintf("INSERT OR REPLACE INTO bilibili_history_%d (%s) VALUES (%s)", year, strings.Join(cols, ","), placeholders)
	if _, err := db(t).Exec(q, vals...); err != nil {
		t.Fatalf("insert history into %d: %v", year, err)
	}
}

// viewAt builds a unix timestamp in the machine local timezone, matching the
// way the production SQL renders dates (datetime(view_at,'unixepoch','localtime')).
func viewAt(t *testing.T, y int, mo time.Month, day, hour, min int) int64 {
	t.Helper()
	return time.Date(y, mo, day, hour, min, 0, 0, time.Local).Unix()
}

// uniqueBvid returns a bvid that cannot collide with real B站 ids used in
// fixtures elsewhere; the tests never send them to the network.
func uniqueBvid(prefix string, n int) string {
	return fmt.Sprintf("%s%09d", prefix, n)
}

func countRows(t *testing.T, query string, args ...interface{}) int {
	t.Helper()
	var n int
	if err := db(t).QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

func tableExists(t *testing.T, name string) bool {
	t.Helper()
	var got string
	err := db(t).QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&got)
	return err == nil
}

// setSESSDATA temporarily writes the cookie fields into the config singleton so
// that guarded branches can be exercised. Pass "" to simulate a logged-out
// state. The caller must restore the previous values with the returned func.
func setSESSDATA(t *testing.T, value string) func() {
	t.Helper()
	cfg := loadCfg(t)
	prev := cfg.SESSDATA
	cfg.SESSDATA = value
	return func() { cfg.SESSDATA = prev }
}

func hasData(resp map[string]interface{}, key string) bool {
	_, ok := resp[key]
	return ok
}
