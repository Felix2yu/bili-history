package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/models"
)

// 测试数据命名空间：本包所有测试共享同一组 SQLite 单例文件（sync.Once，无法重置），
// 因此每个测试文件只允许读写分配给它的年份表 / bvid 前缀，保证互不干扰且结果确定。
const (
	yearInfra1 = 2001 // sqlite_test.go 沙箱（可随意建表/删列）
	yearInfra2 = 2002 // sqlite_test.go 第二个沙箱表
	// yearAnalysisSeed 由 TestMain 预置确定性数据集，analysis/report 测试只读。
	yearAnalysisSeed = 2005
	yearHistory1     = 2010 // history_test.go 主表
	yearHistory2     = 2011 // history_test.go 第二张表（UNION 测试）
	yearDetails      = 2015 // video_details_test.go 需要的历史表
)

// analysisBvidPrefix 限定 TestMain 种子数据的 bvid 范围，供只读断言使用。
const analysisBvidPrefix = "BVseed"

var testWorkDir string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "bilidb-test-")
	if err != nil {
		panic(err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	testWorkDir = dir

	// 测试进程里没有可用的 config.yaml，LoadConfig 会得到空 db_file，
	// GetDBFilePath 将退化为 output 目录本身导致打开失败。补一个最小配置。
	if err := os.Mkdir(filepath.Join(dir, "config"), 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config", "config.yaml"),
		[]byte("db_file: \"bilibili_history.test.db\"\n"), 0644); err != nil {
		panic(err)
	}

	db := GetSQLiteDB()
	if db.GetDB() == nil {
		panic("main database failed to initialize")
	}
	if err := db.EnsureTableForYear(yearAnalysisSeed); err != nil {
		panic(err)
	}
	if err := seedAnalysisRows(db.GetDB(), yearAnalysisSeed); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = os.Chdir(orig)
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func atTime(y, mo, d, hh, mm int) int64 {
	return time.Date(y, time.Month(mo), d, hh, mm, 0, 0, time.Local).Unix()
}

func mustExec(t *testing.T, conn *sql.DB, query string, args ...interface{}) {
	t.Helper()
	if _, err := conn.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

// mustSeedHistory 建表（若不存在）并写入历史记录，包含 status 列。
func mustSeedHistory(t *testing.T, year int, records ...models.HistoryRecord) {
	t.Helper()
	if err := GetSQLiteDB().EnsureTableForYear(year); err != nil {
		t.Fatalf("EnsureTableForYear(%d): %v", year, err)
	}
	if err := seedHistoryRecords(testConn(), year, records...); err != nil {
		t.Fatalf("seed history for %d: %v", year, err)
	}
}

func testConn() *sql.DB {
	return GetSQLiteDB().GetDB()
}

func seedHistoryRecords(conn *sql.DB, year int, records ...models.HistoryRecord) error {
	table := fmt.Sprintf("bilibili_history_%d", year)
	placeholders := "(" + strings.TrimSuffix(strings.Repeat("?,", 34), ",") + ")"
	for i := range records {
		r := records[i]
		_, err := conn.Exec(fmt.Sprintf(`
			INSERT INTO %s (
				title, long_title, cover, covers, uri, oid, epid, bvid, page,
				cid, part, business, dt, videos, author_name, author_face, author_mid,
				view_at, progress, badge, show_title, duration, current, total,
				new_desc, is_finish, is_fav, kid, tag_name, live_status, main_category,
				remark, remark_time, status
			) VALUES %s
		`, table, placeholders),
			r.Title, r.LongTitle, r.Cover, r.Covers, r.URI, r.OID, r.Epid, r.Bvid, r.Page,
			r.Cid, r.Part, r.Business, r.Dt, r.Videos, r.AuthorName, r.AuthorFace, r.AuthorMid,
			r.ViewAt, r.Progress, r.Badge, r.ShowTitle, r.Duration, r.Current, r.Total,
			r.NewDesc, r.IsFinish, r.IsFav, r.Kid, r.TagName, r.LiveStatus, r.MainCategory,
			r.Remark, r.RemarkTime, r.Status,
		)
		if err != nil {
			return fmt.Errorf("insert %s: %w", r.Bvid, err)
		}
	}
	return nil
}

func historyRecord(bvid, title, business, tag, main string, mid int64, author string, viewAt int64, duration, progress, dt int) models.HistoryRecord {
	return models.HistoryRecord{
		Title:        title,
		Bvid:         bvid,
		OID:          int64(len(bvid)) + viewAt%1000,
		Business:     business,
		TagName:      tag,
		MainCategory: main,
		AuthorMid:    mid,
		AuthorName:   author,
		ViewAt:       viewAt,
		Duration:     duration,
		Progress:     progress,
		Dt:           dt,
		Page:         1,
		Videos:       1,
		Cover:        "https://i0.hdslb.com/bfs/cover/" + bvid + ".jpg",
	}
}

// seedAnalysisRows 写入 year 表（默认 2005）的确定性数据集。
// 组成：13 行，其中 2 行重复 bvid（刷片）、1 行 live、1 行 article（被大多数查询排除）、
// 1 行 status=1（已删除）、1 行 progress=-1、若干整行/半行/零进度，覆盖各聚合分支。
func seedAnalysisRows(conn *sql.DB, year int) error {
	a := "UP主一号"
	b := "游戏区UP"
	c := "番剧菌"
	d := "音乐人"
	e := "生活家"
	recs := []models.HistoryRecord{
		historyRecord("BVseed0001", "科技 年度 盘点", "archive", "科技", "科技", 5001, a, atTime(year, 1, 3, 3, 15), 900, 900, 1),
		historyRecord("BVseed0001", "科技 年度 盘点", "archive", "科技", "科技", 5001, a, atTime(year, 1, 20, 23, 40), 900, 450, 2),
		historyRecord("BVseed0003", "游戏 评测 新作 上线", "archive", "游戏", "游戏", 5002, b, atTime(year, 2, 5, 12, 0), 1800, -1, 1),
		historyRecord("BVseed0004", "番剧 第一 季 完结", "pgc", "番剧", "番剧", 5003, c, atTime(year, 3, 10, 20, 0), 240, 200, 33),
		historyRecord("BVseed0005", "音乐 现场 精选", "archive", "音乐", "音乐", 5004, d, atTime(year, 4, 15, 8, 30), 300, 0, 4),
		historyRecord("BVseed0006", "科技 改变 生活 方式", "archive", "科技", "生活", 5005, e, atTime(year, 5, 20, 18, 45), 3600, 1800, 1),
		historyRecord("BVseed0007", "直播 测试 房间", "live", "直播", "直播", 5006, "直播主", atTime(year, 6, 1, 21, 0), 600, 600, 1),
		historyRecord("BVseed0008", "知识 分享 长视频 讲座", "archive", "知识", "知识", 5001, a, atTime(year, 7, 4, 2, 0), 7200, 7200, 2),
		historyRecord("BVseed0009", "专栏 文章 测试", "article", "专栏", "专栏", 5007, "专栏作者", atTime(year, 6, 15, 10, 0), 0, 0, 1),
		historyRecord("BVseed0010", "游戏 实况 合集 精选", "archive", "游戏", "游戏", 5002, b, atTime(year, 8, 8, 21, 30), 150, 140, 1),
		historyRecord("BVseed0003", "游戏 评测 新作 上线", "archive", "游戏", "游戏", 5002, b, atTime(year, 9, 9, 15, 0), 1800, 300, 1),
		historyRecord("BVseed0012", "科技 年度 趋势 解读", "archive", "科技", "科技", 5005, e, atTime(year, 12, 25, 6, 0), 1200, 100, 2),
		historyRecord("BVseed0013", "音乐 电台 深夜 陪伴", "archive", "音乐", "音乐", 5004, d, atTime(year, 11, 11, 23, 30), 2400, 2400, 1),
	}
	// 第 12 行标记为已删除（status=1），其余为 0
	recs[11].Status = 1
	// 为 cid 分配唯一值，GetVideoByCID 测试依赖
	for i := range recs {
		recs[i].Cid = int64(90000 + i)
	}
	recs[3].Epid = 100404
	recs[3].OID = 200200
	recs[3].Remark = "神作"
	recs[3].RemarkTime = atTime(year, 3, 11, 9, 0)
	recs[0].Page = 2
	recs[6].OID = 5006
	recs[8].OID = 9001
	return seedHistoryRecords(conn, year, recs...)
}
