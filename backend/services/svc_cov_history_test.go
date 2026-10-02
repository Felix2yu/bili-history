package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
)

// 命名空间：年份表 2033 / 2034 + bvid 前缀 BVsvcCov（见任务约定，勿用 2031/2032/当前年）。
const (
	svcCovYearA = 2033
	svcCovYearB = 2034
)

const svcCovSessdata = "svc-cov-sessdata"

// svcCovSetURL 重定向 biliapi 的包级端点变量，测试结束自动恢复。
func svcCovSetURL(t *testing.T, target *string, url string) {
	t.Helper()
	orig := *target
	*target = url
	t.Cleanup(func() { *target = orig })
}

// svcCovWriteBiliJSON 以 B站响应信封写出 data。
func svcCovWriteBiliJSON(t *testing.T, w http.ResponseWriter, code int, message string, data any) {
	t.Helper()
	resp := map[string]any{"code": code, "message": message}
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			t.Errorf("marshal bili response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp["data"] = json.RawMessage(raw)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Errorf("write bili response: %v", err)
	}
}

func svcCovDropYearTables(t *testing.T, years ...int) {
	t.Helper()
	conn := svcConn()
	t.Cleanup(func() {
		for _, y := range years {
			if _, err := conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS bilibili_history_%d", y)); err != nil {
				t.Errorf("drop table %d: %v", y, err)
			}
		}
	})
}

// svcCovHistoryServer 起一个 history cursor 桩服务：按请求的 max 参数返回预设页。
func svcCovHistoryServer(t *testing.T, pages map[string]biliapi.HistoryCursorData, alwaysErr *biliapi.ApiError) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/x/web-interface/history/cursor", func(w http.ResponseWriter, r *http.Request) {
		if alwaysErr != nil {
			svcCovWriteBiliJSON(t, w, alwaysErr.Code, alwaysErr.Message, nil)
			return
		}
		max := r.URL.Query().Get("max")
		data, ok := pages[max]
		if !ok {
			// 未知游标：返回收敛页（空列表 + cursor.Max==0），避免生产循环无限翻页。
			t.Logf("unexpected history request max=%q, replying empty terminal page", max)
			svcCovWriteBiliJSON(t, w, 0, "OK", svcCovPage(biliapi.HistoryCursor{Max: 0}))
			return
		}
		svcCovWriteBiliJSON(t, w, 0, "OK", data)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	svcCovSetURL(t, &biliapi.HistoryURL, srv.URL+"/x/web-interface/history/cursor")
	return srv
}

func svcCovPage(cursor biliapi.HistoryCursor, list ...biliapi.HistoryEntry) biliapi.HistoryCursorData {
	return biliapi.HistoryCursorData{Cursor: cursor, List: list}
}

func svcCovEntry(bvid, histBvid, business, histBusiness, title string, viewAt int64, page, cid, dt int, part string, progress, duration int) biliapi.HistoryEntry {
	return biliapi.HistoryEntry{
		Title:      title,
		Cover:      "https://cover.svccov.invalid/" + bvid + histBvid + ".jpg",
		URI:        "https://www.bilibili.com/video/" + bvid,
		Business:   business,
		Bvid:       bvid,
		ViewAt:     viewAt,
		Progress:   progress,
		DTotal:     duration,
		ShowTitle:  title,
		AuthorName: "svcCov作者",
		AuthorFace: "https://face.svccov.invalid/a.jpg",
		AuthorMid:  8800001,
		History: biliapi.HistoryInfo{
			Bvid:     histBvid,
			Page:     page,
			Cid:      cid,
			Part:     part,
			Business: histBusiness,
			Dt:       dt,
		},
	}
}

// svcCovWaitFetchDone 轮询异步抓取任务直至 IsRunning=false。
func svcCovWaitFetchDone(t *testing.T, taskID string) *FetchStatus {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		st := GetFetchTaskStatus(taskID)
		if st != nil && !st.IsRunning {
			return st
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fetch task %s did not finish within 30s", taskID)
	return nil
}

func svcCovRunningCount() int32 {
	overall := GetFetchStatusOverall()
	rc, _ := overall["running_count"].(int32)
	return rc
}

// TestSVCCovFetchHistoryFullFlow 覆盖异步抓取的空页游标续拉、条目转换、
// 顶层/history 双层回退、去重键以及按 view_at 路由年份表。
func TestSVCCovFetchHistoryFullFlow(t *testing.T) {
	svcCovDropYearTables(t, svcCovYearA)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	v1 := svcAt(svcCovYearA, 5, 1, 10, 0)
	v2 := svcAt(svcCovYearA, 5, 1, 9, 0)
	v3 := svcAt(svcCovYearA, 5, 1, 8, 0)

	itemPage := svcCovPage(
		biliapi.HistoryCursor{Max: 1700000200, ViewAt: v3},
		svcCovEntry("BVsvcCovA1", "", "archive", "", "svcCov 标题一", v1, 2, 111, 7, "P2", 55, 300),
		svcCovEntry("BVsvcCovA1", "", "archive", "", "svcCov 重复键", v1, 2, 111, 7, "P2", 55, 300), // dedup by bvid_viewAt
		svcCovEntry("", "BVsvcCovA2", "", "archive", "svcCov 回退二", v2, 1, 112, 5, "P1", 10, 90),
		svcCovEntry("BVsvcCovA3", "", "live", "", "svcCov 直播", v3, 1, 113, 5, "P1", 0, 0),
		svcCovEntry("", "", "archive", "", "svcCov 无bvid", v3, 1, 114, 5, "P1", 0, 0),
	)
	pages := map[string]biliapi.HistoryCursorData{
		"":           svcCovPage(biliapi.HistoryCursor{Max: 1700000100, ViewAt: v2}), // 空页但游标可用 → 续拉
		"1700000100": itemPage,
		"1700000200": svcCovPage(biliapi.HistoryCursor{Max: 0}), // 空页 + cursor.Max==0 → 停止
	}
	svcCovHistoryServer(t, pages, nil)

	res, err := FetchHistory("", false) // 空 taskID → 自动生成 manual_*
	if err != nil {
		t.Fatalf("FetchHistory: %v", err)
	}
	taskID, _ := res["task_id"].(string)
	if !strings.HasPrefix(taskID, "manual_") {
		t.Fatalf("generated task_id = %q, want manual_ prefix", taskID)
	}
	st := svcCovWaitFetchDone(t, taskID)

	if st.Status != "completed" {
		t.Fatalf("status = %+v", st)
	}
	if st.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", st.TotalPages)
	}
	if st.TotalRecords != 5 {
		t.Errorf("TotalRecords = %d, want 5 (all collected pre-dedup)", st.TotalRecords)
	}
	if st.NewRecords != 2 {
		t.Errorf("NewRecords = %d, want 2 (dedup + business/bvid 过滤后)", st.NewRecords)
	}

	conn := svcConn()
	var cnt int
	if err := conn.QueryRow("SELECT COUNT(*) FROM bilibili_history_2033 WHERE bvid='BVsvcCovA1'").Scan(&cnt); err != nil || cnt != 1 {
		t.Errorf("A1 rows = %d (err %v), want 1", cnt, err)
	}
	var rec struct {
		title    string
		viewAt   int64
		progress int
		duration int
		author   string
		mid      int64
		business string
		page     int
		cid      int
		part     string
		dt       int
	}
	if err := conn.QueryRow(`SELECT title, view_at, progress, duration, author_name, author_mid, business, page, cid, part, dt
		FROM bilibili_history_2033 WHERE bvid='BVsvcCovA1'`).Scan(
		&rec.title, &rec.viewAt, &rec.progress, &rec.duration, &rec.author, &rec.mid,
		&rec.business, &rec.page, &rec.cid, &rec.part, &rec.dt); err != nil {
		t.Fatalf("read A1 row: %v", err)
	}
	if rec.title != "svcCov 标题一" || rec.viewAt != v1 || rec.progress != 55 || rec.duration != 300 {
		t.Errorf("A1 transform = %+v", rec)
	}
	if rec.author != "svcCov作者" || rec.mid != 8800001 || rec.business != "archive" {
		t.Errorf("A1 author/business = %+v", rec)
	}
	if rec.page != 2 || rec.cid != 111 || rec.part != "P2" || rec.dt != 7 {
		t.Errorf("A1 history fields = %+v", rec)
	}
	// history.bvid / history.business 回退
	var a2Bvid, a2Business string
	if err := conn.QueryRow("SELECT bvid, business FROM bilibili_history_2033 WHERE title='svcCov 回退二'").Scan(&a2Bvid, &a2Business); err != nil {
		t.Fatalf("read A2 row: %v", err)
	}
	if a2Bvid != "BVsvcCovA2" || a2Business != "archive" {
		t.Errorf("A2 fallback = (%q,%q), want (BVsvcCovA2, archive)", a2Bvid, a2Business)
	}
	// 非 archive 与空 bvid 都不落库
	if err := conn.QueryRow("SELECT COUNT(*) FROM bilibili_history_2033 WHERE bvid='BVsvcCovA3'").Scan(&cnt); err != nil || cnt != 0 {
		t.Errorf("live entry rows = %d (err %v), want 0", cnt, err)
	}
}

// TestSVCCovFetchHistoryWithin24hUpdates 覆盖 InsertHistoryRecord 返回
// inserted=false 的同会话分支：已存在 bvid 且 <24h 时只更新 view_at/cover。
func TestSVCCovFetchHistoryWithin24hUpdates(t *testing.T) {
	svcCovDropYearTables(t, svcCovYearA)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	base := svcAt(svcCovYearA, 8, 1, 12, 0)
	existing := svcHistoryRecord("BVsvcCovW24", "svcCov 既有行", "archive", "科技", "科技", 8800024, "svcCov旧作者", base, 100, 20)
	existing.Cover = "https://cover.svccov.invalid/old.jpg"
	if err := svcSeedHistoryRows(svcCovYearA, existing); err != nil {
		t.Fatalf("seed: %v", err)
	}

	pages := map[string]biliapi.HistoryCursorData{
		"": svcCovPage(biliapi.HistoryCursor{Max: 1700000100, ViewAt: base + 3600},
			svcCovEntry("BVsvcCovW24", "", "archive", "", "svcCov 既有行", base+3600, 1, 124, 5, "P1", 88, 100)),
		"1700000100": svcCovPage(biliapi.HistoryCursor{Max: 0}), // 终止页
	}
	svcCovHistoryServer(t, pages, nil)

	res, err := FetchHistory("svcCov-w24", false)
	if err != nil {
		t.Fatalf("FetchHistory: %v", err)
	}
	st := svcCovWaitFetchDone(t, "svcCov-w24")
	if res == nil || st.Status != "completed" {
		t.Fatalf("res=%v status=%+v", res, st)
	}
	if st.NewRecords != 0 {
		t.Errorf("NewRecords = %d, want 0 (24h 内同会话只更新不新增)", st.NewRecords)
	}

	conn := svcConn()
	var cnt int
	var viewAt int64
	var cover string
	if err := conn.QueryRow("SELECT COUNT(*) FROM bilibili_history_2033 WHERE bvid='BVsvcCovW24'").Scan(&cnt); err != nil || cnt != 1 {
		t.Fatalf("rows = %d (err %v), want 1", cnt, err)
	}
	if err := conn.QueryRow("SELECT view_at, cover FROM bilibili_history_2033 WHERE bvid='BVsvcCovW24'").Scan(&viewAt, &cover); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if viewAt != base+3600 {
		t.Errorf("view_at = %d, want updated to %d", viewAt, base+3600)
	}
	if !strings.Contains(cover, "BVsvcCovW24") {
		t.Errorf("cover = %q, want replaced by fetched entry", cover)
	}
}

// TestSVCCovFetchHistoryIncrementalCutoff 覆盖 skipExists 增量模式：
// cutoff 取全库最新 view_at 的当日零点，更早条目不落库并立即停止。
func TestSVCCovFetchHistoryIncrementalCutoff(t *testing.T) {
	svcCovDropYearTables(t, svcCovYearA, svcCovYearB)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	globalMax := svcGlobalMaxViewAt(t)
	if globalMax < svcAt(svcYearClean, 2, 3, 12, 0) {
		t.Fatalf("fixture sanity: global max view_at = %d", globalMax)
	}
	latest := time.Unix(globalMax, 0)
	cutoff := time.Date(latest.Year(), latest.Month(), latest.Day(), 0, 0, 0, 0, time.Local).Unix()
	oldViewAt := cutoff - 30*24*3600 // 早于 cutoff 30 天，任何顺序下都被过滤

	pages := map[string]biliapi.HistoryCursorData{
		"": svcCovPage(biliapi.HistoryCursor{Max: 1700000100, ViewAt: oldViewAt},
			svcCovEntry("BVsvcCovSkip01", "", "archive", "", "svcCov 旧数据", oldViewAt, 1, 131, 5, "P1", 0, 60)),
	}
	svcCovHistoryServer(t, pages, nil)

	res, err := FetchHistory("svcCov-inc", true)
	if err != nil {
		t.Fatalf("FetchHistory: %v", err)
	}
	if res["task_id"] != "svcCov-inc" {
		t.Errorf("task_id = %v, want svcCov-inc", res["task_id"])
	}
	st := svcCovWaitFetchDone(t, "svcCov-inc")
	if st.Status != "completed" {
		t.Fatalf("status = %+v", st)
	}
	if st.TotalPages != 1 {
		t.Errorf("TotalPages = %d, want 1 (cutoff 命中后应立即停止)", st.TotalPages)
	}
	if st.NewRecords != 0 || st.TotalRecords != 0 {
		t.Errorf("records = (%d,%d), want (0,0)", st.TotalRecords, st.NewRecords)
	}

	conn := svcConn()
	var cnt, exists int
	for _, tb := range []string{"bilibili_history_2031", "bilibili_history_2032", "bilibili_history_2033", "bilibili_history_2034"} {
		if err := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tb).Scan(&exists); err != nil {
			t.Fatalf("check table %s: %v", tb, err)
		}
		if exists == 0 {
			continue
		}
		if err := conn.QueryRow("SELECT COUNT(*) FROM " + tb + " WHERE bvid='BVsvcCovSkip01'").Scan(&cnt); err != nil || cnt != 0 {
			t.Errorf("%s rows for skipped bvid = %d (err %v), want 0", tb, cnt, err)
		}
	}
}

// TestSVCCovFetchHistoryConsecutiveErrors 覆盖连续 API 错误自动暂停路径：
// 每页 2s 退避，连续 3 页失败后写入 error 状态并停用 fetch_history 任务。
func TestSVCCovFetchHistoryConsecutiveErrors(t *testing.T) {
	svcCovDropYearTables(t, svcCovYearA)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })
	t.Cleanup(func() {
		if err := database.SetTaskEnabled("fetch_history", true); err != nil {
			t.Errorf("re-enable fetch_history: %v", err)
		}
	})

	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		svcCovWriteBiliJSON(t, w, -352, "风控校验失败", nil)
	}))
	t.Cleanup(srv.Close)
	svcCovSetURL(t, &biliapi.HistoryURL, srv.URL+"/x/web-interface/history/cursor")

	rcBefore := svcCovRunningCount()
	if _, err := FetchHistory("svcCov-err", false); err != nil {
		t.Fatalf("FetchHistory: %v", err)
	}
	st := svcCovWaitFetchDone(t, "svcCov-err")
	if st.Status != "error" {
		t.Errorf("status = %q, want error", st.Status)
	}
	if !strings.Contains(st.ErrorMessage, "连续 3 页失败") || !strings.Contains(st.ErrorMessage, "code=-352") {
		t.Errorf("ErrorMessage = %q", st.ErrorMessage)
	}
	if hits != 3 {
		t.Errorf("api hits = %d, want 3", hits)
	}
	if st.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", st.TotalPages)
	}
	if rc := svcCovRunningCount(); rc != rcBefore {
		t.Errorf("running_count leaked: before=%d after=%d", rcBefore, rc)
	}
	removeFetchTaskStatus("svcCov-err")
}

// TestSVCCovFetchHistoryEmptyCursorBreaks 覆盖空页的另两条退出路径：
// 游标回退 (<1000000) 与连续 3 空页。
func TestSVCCovFetchHistoryEmptyCursorBreaks(t *testing.T) {
	svcCovDropYearTables(t, svcCovYearA)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	t.Run("cursor rollback breaks", func(t *testing.T) {
		pages := map[string]biliapi.HistoryCursorData{
			"":           svcCovPage(biliapi.HistoryCursor{Max: 1700000100}),
			"1700000100": svcCovPage(biliapi.HistoryCursor{Max: 999999}),     // max>0 且回退到 <1e6 → break
			"1700000999": svcCovPage(biliapi.HistoryCursor{Max: 1700000999}), // 不应被请求
		}
		svcCovHistoryServer(t, pages, nil)
		if _, err := FetchHistory("svcCov-empty-rollback", false); err != nil {
			t.Fatalf("FetchHistory: %v", err)
		}
		st := svcCovWaitFetchDone(t, "svcCov-empty-rollback")
		if st.TotalPages != 2 || st.Status != "completed" {
			t.Errorf("status = %+v, want completed after 2 pages", st)
		}
		removeFetchTaskStatus("svcCov-empty-rollback")
	})

	t.Run("three empty pages break", func(t *testing.T) {
		pages := map[string]biliapi.HistoryCursorData{
			"":           svcCovPage(biliapi.HistoryCursor{Max: 1700000100}),
			"1700000100": svcCovPage(biliapi.HistoryCursor{Max: 1700000200}),
			"1700000200": svcCovPage(biliapi.HistoryCursor{Max: 1700000300}), // 第 3 空页 → emptyPageCount>=3 break
			"1700000300": svcCovPage(biliapi.HistoryCursor{Max: 1700000400}), // 不应被请求
		}
		svcCovHistoryServer(t, pages, nil)
		if _, err := FetchHistory("svcCov-empty-three", false); err != nil {
			t.Fatalf("FetchHistory: %v", err)
		}
		st := svcCovWaitFetchDone(t, "svcCov-empty-three")
		if st.TotalPages != 3 || st.Status != "completed" {
			t.Errorf("status = %+v, want completed after 3 pages", st)
		}
		removeFetchTaskStatus("svcCov-empty-three")
	})
}

// TestSVCCovFetchHistorySyncFullFlow 覆盖同步抓取的返回统计、任务注销与
// 2034 年份表路由。
func TestSVCCovFetchHistorySyncFullFlow(t *testing.T) {
	svcCovDropYearTables(t, svcCovYearA, svcCovYearB)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	v1 := svcAt(svcCovYearB, 6, 1, 10, 0)
	v2 := svcAt(svcCovYearB, 6, 1, 9, 0)
	pages := map[string]biliapi.HistoryCursorData{
		"": svcCovPage(biliapi.HistoryCursor{Max: 1700000100, ViewAt: v2},
			svcCovEntry("BVsvcCovS1", "", "archive", "", "svcCov 同步一", v1, 3, 201, 6, "P3", 40, 240),
			svcCovEntry("BVsvcCovS2", "", "live", "", "svcCov 同步直播", v2, 1, 202, 6, "P1", 0, 0),
		),
		"1700000100": svcCovPage(biliapi.HistoryCursor{Max: 0}),
	}
	svcCovHistoryServer(t, pages, nil)

	res, err := FetchHistorySync("", false)
	if err != nil {
		t.Fatalf("FetchHistorySync: %v", err)
	}
	if res["status"] != "success" {
		t.Fatalf("res = %v", res)
	}
	taskID, _ := res["task_id"].(string)
	if !strings.HasPrefix(taskID, "sync_") {
		t.Errorf("task_id = %q, want sync_ prefix", taskID)
	}
	if res["total_pages"] != 2 || res["total_records"] != 2 || res["new_records"] != 1 {
		t.Errorf("res = %v, want pages=2 records=2 new=1", res)
	}
	if GetFetchTaskStatus(taskID) != nil {
		t.Errorf("sync task should be removed after completion")
	}

	conn := svcConn()
	var title string
	var viewAt int64
	if err := conn.QueryRow("SELECT title, view_at FROM bilibili_history_2034 WHERE bvid='BVsvcCovS1'").Scan(&title, &viewAt); err != nil {
		t.Fatalf("read synced row from 2034 table: %v", err)
	}
	if title != "svcCov 同步一" || viewAt != v1 {
		t.Errorf("synced row = (%q,%d)", title, viewAt)
	}
	var cnt int
	if err := conn.QueryRow("SELECT COUNT(*) FROM bilibili_history_2034 WHERE bvid='BVsvcCovS2'").Scan(&cnt); err != nil || cnt != 0 {
		t.Errorf("live entry rows = %d (err %v), want 0", cnt, err)
	}
	// 2033 表不应被创建或写入（按 view_at 路由到 2034）
	var tables int
	if err := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='bilibili_history_2033'").Scan(&tables); err != nil {
		t.Fatalf("check 2033 table: %v", err)
	}
	if tables == 1 {
		if err := conn.QueryRow("SELECT COUNT(*) FROM bilibili_history_2033 WHERE bvid='BVsvcCovS1'").Scan(&cnt); err != nil || cnt != 0 {
			t.Errorf("2033 rows for synced bvid = %d (err %v), want 0 (year routing)", cnt, err)
		}
	}
}

// TestSVCCovFetchHistorySyncIncremental 覆盖同步抓取的 cutoff 停止分支。
func TestSVCCovFetchHistorySyncIncremental(t *testing.T) {
	svcCovDropYearTables(t, svcCovYearA, svcCovYearB)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	latest := time.Unix(svcGlobalMaxViewAt(t), 0)
	cutoff := time.Date(latest.Year(), latest.Month(), latest.Day(), 0, 0, 0, 0, time.Local).Unix()
	oldViewAt := cutoff - 30*24*3600

	pages := map[string]biliapi.HistoryCursorData{
		"": svcCovPage(biliapi.HistoryCursor{Max: 1700000100, ViewAt: oldViewAt},
			svcCovEntry("BVsvcCovSkip02", "", "archive", "", "svcCov 同步旧数据", oldViewAt, 1, 211, 5, "P1", 0, 60)),
	}
	svcCovHistoryServer(t, pages, nil)

	res, err := FetchHistorySync("svcCov-sync-inc", true)
	if err != nil {
		t.Fatalf("FetchHistorySync: %v", err)
	}
	if res["total_pages"] != 1 || res["total_records"] != 0 || res["new_records"] != 0 {
		t.Errorf("res = %v, want 1/0/0", res)
	}
	if GetFetchTaskStatus("svcCov-sync-inc") != nil {
		t.Errorf("task should be removed after sync completion")
	}
}

// TestSVCCovFetchHistorySyncApiError 回归 services/history.go:407 的缺陷修复：
// GetHistory 返回错误时 data 为 nil，抓取循环不得在 err 检查之前解引用它。
// 修复前这里会 nil 指针 panic；现在应走完错误分支（退避重试），第二页拿到
// 收敛空页后正常返回。
func TestSVCCovFetchHistorySyncApiError(t *testing.T) {
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) == 1 {
			svcCovWriteBiliJSON(t, w, -6, "视频不存在", nil)
			return
		}
		svcCovWriteBiliJSON(t, w, 0, "OK", svcCovPage(biliapi.HistoryCursor{Max: 0}))
	}))
	t.Cleanup(srv.Close)
	svcCovSetURL(t, &biliapi.HistoryURL, srv.URL+"/x/web-interface/history/cursor")

	rcBefore := svcCovRunningCount()
	taskID := "svcCov-sync-err"

	res, err := func() (res map[string]interface{}, err error) {
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("FetchHistorySync panicked on API error (history.go:407 nil deref): %v", r)
					res, err = nil, fmt.Errorf("panic: %v", r)
				}
			}()
			res, err = FetchHistorySync(taskID, false)
		}()
		<-done
		return
	}()

	if err != nil {
		t.Fatalf("FetchHistorySync: %v", err)
	}
	if n := atomic.LoadInt32(&hits); n != 2 {
		t.Errorf("api hits = %d, want 2 (one failure then terminal empty page)", n)
	}
	if res["status"] != "success" || res["total_records"] != 0 {
		t.Errorf("result = %+v", res)
	}
	if GetFetchTaskStatus(taskID) != nil {
		t.Errorf("task status should be removed by deferred cleanup")
	}
	if rc := svcCovRunningCount(); rc != rcBefore {
		t.Errorf("running_count leaked: before=%d after=%d", rcBefore, rc)
	}
}
