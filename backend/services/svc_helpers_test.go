package services

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
	"bilibili-history-go/models"
)

// 命名空间约定（本包所有测试共享同一组 SQLite 单例，sync.Once 无法重置）：
//   - bilibili_history_2031: 只读主历史数据集（FindLatestHistoryDate / loadBVMetadata / data_sync）
//   - bilibili_history_2032: clean 测试专用（会被 cleanDuplicates / cleanOldHistory 删行）
//   - bilibili_history_<当前年>: notify 日报测试专用（gatherDailyReportData 只查当前年）
//   - bvid 前缀 BVsvc*、动态 hostMid 900000001、文件名前缀 svc*
const (
	svcYearMain  = 2031
	svcYearClean = 2032
)

var svcCurrentYear = time.Now().Year()

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "bilisvc-test-")
	if err != nil {
		panic(err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}

	// 最小配置：db_file 必须存在否则 GetDBFilePath 返回目录导致 sqlite 打开失败。
	// SESSDATA 保持为空，让所有服务层的 "SESSDATA 未配置" 守卫分支生效（零网络）。
	if err := os.Mkdir(filepath.Join(dir, "config"), 0755); err != nil {
		panic(err)
	}
	yaml := strings.Join([]string{
		`db_file: "bilibili_history.svctest.db"`,
		`output_folder: "output"`,
		`log_folder: "output/logs"`,
		`notify:`,
		`  enabled: false`,
		`  urls: []`,
		`server:`,
		`  data_integrity:`,
		`    check_on_startup: false`,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "config", "config.yaml"), []byte(yaml), 0644); err != nil {
		panic(err)
	}

	// 触发一次配置加载（sync.Once），后续测试读到的都是这份指针。
	if cfg := config.GetConfig(); cfg == nil {
		panic("config failed to load")
	}
	db := database.GetSQLiteDB()
	if db.GetDB() == nil {
		panic("main database failed to initialize")
	}

	if err := svcSeedDataset(); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = os.Chdir(orig)
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func svcConn() *sql.DB {
	return database.GetSQLiteDB().GetDB()
}

func svcAt(y, mo, d, hh, mm int) int64 {
	return time.Date(y, time.Month(mo), d, hh, mm, 0, 0, time.Local).Unix()
}

// svcHistoryRecord 构造一条命名空间内的历史记录。
func svcHistoryRecord(bvid, title, business, tag, main string, mid int64, author string, viewAt int64, duration, progress int) models.HistoryRecord {
	return models.HistoryRecord{
		Title:        title,
		Bvid:         bvid,
		OID:          int64(len(bvid)) + viewAt%1000,
		Business:     business,
		TagName:      tag,
		MainCategory: main,
		AuthorMid:    mid,
		AuthorName:   author,
		AuthorFace:   "https://face.invalid/" + bvid + ".jpg",
		ViewAt:       viewAt,
		Duration:     duration,
		Progress:     progress,
		Dt:           1,
		Page:         1,
		Videos:       1,
		Cover:        "https://cover.invalid/svc/" + bvid + ".jpg",
	}
}

func svcSeedHistory(t *testing.T, year int, records ...models.HistoryRecord) {
	t.Helper()
	if err := svcSeedHistoryRows(year, records...); err != nil {
		t.Fatalf("seed history for %d: %v", year, err)
	}
}

func svcSeedHistoryRows(year int, records ...models.HistoryRecord) error {
	db := database.GetSQLiteDB()
	if err := db.EnsureTableForYear(year); err != nil {
		return err
	}
	conn := db.GetDB()
	table := fmt.Sprintf("bilibili_history_%d", year)
	placeholders := "(" + strings.TrimSuffix(strings.Repeat("?,", 34), ",") + ")"
	for i := range records {
		r := records[i]
		if _, err := conn.Exec(fmt.Sprintf(`
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
		); err != nil {
			return fmt.Errorf("insert %s: %w", r.Bvid, err)
		}
	}
	return nil
}

// svcSeedDataset 写入跨测试共享的确定性数据集（TestMain 调用，任何测试顺序下均可用）。
func svcSeedDataset() error {
	// 2031: 只读主数据集。max view_at = 2031-06-02 23:30。
	main := []models.HistoryRecord{
		svcHistoryRecord("BVsvcMain0001", "svc 主标题一", "archive", "科技", "科技", 7700001, "svc作者甲", svcAt(svcYearMain, 3, 1, 10, 0), 600, 300),
		svcHistoryRecord("BVsvcMain0002", "svc 主标题二", "archive", "游戏", "游戏", 7700002, "svc作者乙", svcAt(svcYearMain, 6, 2, 23, 30), 1200, -1),
		svcHistoryRecord("BVsvcMain0003", "svc 主标题三", "live", "直播", "直播", 7700003, "svc作者丙", svcAt(svcYearMain, 1, 15, 8, 0), 0, 0),
	}
	for i := range main {
		main[i].Cid = int64(770000 + i)
	}
	if err := svcSeedHistoryRows(svcYearMain, main...); err != nil {
		return err
	}

	// 2032: clean 测试沙盒。5 行：
	//  重复对 (BVsvcDup0001, t1) x2 —— cleanDuplicates 应删 1
	//  BVsvcOld0001 view_at=now-40d —— cleanOldHistory(10) 应删 1
	//  BVsvcNew0001 view_at=now-1d  —— 保留
	//  BVsvcDup0001 view_at=t2      —— 保留（唯一）
	now := time.Now()
	old := now.AddDate(0, 0, -40)
	recent := now.AddDate(0, 0, -1)
	clean := []models.HistoryRecord{
		svcHistoryRecord("BVsvcDup0001", "svc 重复动态", "archive", "知识", "知识", 7700010, "svc作者丁", svcAt(svcYearClean, 2, 2, 12, 0), 100, 50),
		svcHistoryRecord("BVsvcDup0001", "svc 重复动态", "archive", "知识", "知识", 7700010, "svc作者丁", svcAt(svcYearClean, 2, 2, 12, 0), 100, 50),
		svcHistoryRecord("BVsvcDup0001", "svc 唯一时间", "archive", "知识", "知识", 7700010, "svc作者丁", svcAt(svcYearClean, 2, 3, 12, 0), 100, 50),
		svcHistoryRecord("BVsvcOld0001", "svc 过期记录", "archive", "音乐", "音乐", 7700011, "svc作者丁", old.Unix(), 200, 200),
		svcHistoryRecord("BVsvcNew0001", "svc 新鲜记录", "archive", "音乐", "音乐", 7700011, "svc作者丁", recent.Unix(), 300, -1),
	}
	if err := svcSeedHistoryRows(svcYearClean, clean...); err != nil {
		return err
	}

	// 当前年: notify 日报沙盒（today 2 条 + 更早 1 条）。
	today := now
	rec1 := svcHistoryRecord("BVsvcNotify0001", "svc 日报视频一", "archive", "科技", "科技", 7700020, "svc日报UP", svcAt(today.Year(), int(today.Month()), today.Day(), 9, 0), 200, 100)
	rec2 := svcHistoryRecord("BVsvcNotify0002", "svc 日报视频二", "archive", "科技", "科技", 7700020, "svc日报UP", svcAt(today.Year(), int(today.Month()), today.Day(), 10, 30), 3600, -1)
	rec3 := svcHistoryRecord("BVsvcNotify0003", "svc 往日视频", "archive", "游戏", "游戏", 7700021, " svc往日UP", svcAt(svcCurrentYear, 1, 2, 7, 0), 30, 30)
	if err := svcSeedHistoryRows(svcCurrentYear, rec1, rec2, rec3); err != nil {
		return err
	}
	return nil
}

// svcWriteFile 在 rel（相对 cwd）路径写入内容，返回绝对路径。
func svcWriteFile(t *testing.T, rel string, content []byte) string {
	t.Helper()
	abs := filepath.Join(mustGetwd(t), rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(abs, content, 0644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	return abs
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}
