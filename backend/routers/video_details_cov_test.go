package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/scheduler"
	"bilibili-history-go/services"
)

// Reserved namespace for this file: year 2002 and bvid prefix BVDL. The
// video_base_info / video_tags / uploader_info / invalid_videos tables are
// global; every assertion there is either scoped to a BVDL bvid or derived
// from the database at assertion time.

// dlStubView repoints biliapi.VideoInfoURL at a local stub server.
func dlStubView(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	prev := biliapi.VideoInfoURL
	biliapi.VideoInfoURL = srv.URL + "/x/web-interface/view"
	t.Cleanup(func() {
		biliapi.VideoInfoURL = prev
		srv.Close()
	})
}

func dlViewPayload(bvid, tname string, mid int) string {
	return fmt.Sprintf(`{"code":0,"message":"0","data":{"bvid":%q,"aid":4242,"videos":1,"tid":28,`+
		`"tname":%q,"copyright":1,"pic":"http://pic/dl.jpg","title":"DL测试视频","pubdate":1600000000,`+
		`"ctime":1600000001,"desc":"dl desc","duration":99,"cid":7,"owner":{"mid":%d,"name":"DLUP主","face":"http://face/dl.jpg"},`+
		`"stat":{"view":1234,"danmaku":5,"reply":6,"favorite":7,"coin":8,"share":9,"like":10}}}`, bvid, tname, mid)
}

func dlQueryInt64(t *testing.T, query string, args ...interface{}) int64 {
	t.Helper()
	var v int64
	if err := db(t).QueryRow(query, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return v
}

func dlSaveVideo(t *testing.T, v *models.VideoBaseInfo) {
	t.Helper()
	if err := services.SaveVideoDetail(v); err != nil {
		t.Fatalf("SaveVideoDetail(%s): %v", v.Bvid, err)
	}
}

// dlDrainProgressChan empties the shared buffered SSE progress channel so a
// later test never observes an earlier batch run's buffered events.
func dlDrainProgressChan() {
	ch := services.GetVideoDetailProgressChan()
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func dlWaitNotProcessing(t *testing.T, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for services.GetVideoDetailProgress().IsProcessing {
		if time.Now().After(deadline) {
			t.Fatalf("video detail fetch task did not finish within %v", timeout)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestDLFetchVideoDetailAndInfo(t *testing.T) {
	if err := database.EnsureVideoDetailsTables(); err != nil {
		t.Fatalf("ensure tables: %v", err)
	}
	dlStubView(t, func(w http.ResponseWriter, r *http.Request) {
		bvid := r.URL.Query().Get("bvid")
		switch {
		case bvid == "BVDLFT000000001":
			fmt.Fprint(w, dlViewPayload(bvid, "科技", 705001))
		case bvid == "BVDLFTBADJSON01":
			fmt.Fprint(w, "this is not json")
		default:
			fmt.Fprintf(w, `{"code":-404,"message":"啥都木有","data":null}`)
		}
	})
	e := newAPI(t, RegisterVideoDetailsRoutes)

	resp := expectStatus(t, e, "GET", "/api/video_details/fetch/BVDLFT000000001", "", 200, "success")
	data := resp.dataMap(t)
	if data["bvid"] != "BVDLFT000000001" || data["title"] != "DL测试视频" || data["tname"] != "科技" {
		t.Fatalf("fetch payload = %v", data)
	}
	if int(data["owner_mid"].(float64)) != 705001 || int(data["stat_view"].(float64)) != 1234 {
		t.Fatalf("owner/stat mapping wrong: %v", data)
	}
	if data["cid"].(float64) != 0 {
		t.Fatalf("cid = %v, want 0 (services hard-codes it)", data["cid"])
	}
	if int64(data["fetch_time"].(float64)) == 0 {
		t.Fatalf("fetch_time not set: %v", data["fetch_time"])
	}

	// The row must be persisted; re-fetching updates it in place (UPSERT).
	if got := dlQueryInt64(t, `SELECT COUNT(*) FROM video_base_info WHERE bvid = 'BVDLFT000000001'`); got != 1 {
		t.Fatalf("persisted rows = %d, want 1", got)
	}
	expectStatus(t, e, "GET", "/api/video_details/fetch/BVDLFT000000001", "", 200, "success")
	if got := dlQueryInt64(t, `SELECT COUNT(*) FROM video_base_info WHERE bvid = 'BVDLFT000000001'`); got != 1 {
		t.Fatalf("rows after re-fetch = %d, want 1 (update path, no duplicate)", got)
	}

	info := expectStatus(t, e, "GET", "/api/video_details/info/BVDLFT000000001", "", 200, "success").dataMap(t)
	if info["bvid"] != "BVDLFT000000001" || info["aid"].(float64) != 4242 {
		t.Fatalf("info payload = %v", info)
	}
	if info["original_url"] != "https://www.bilibili.com/video/BVDLFT000000001" {
		t.Fatalf("original_url = %v", info["original_url"])
	}
	if info["owner_name"] != "DLUP主" || info["duration"].(float64) != 99 {
		t.Fatalf("info mapping = %v", info)
	}

	// Unknown bvid answers with a null payload, not an error.
	miss := expectStatus(t, e, "GET", "/api/video_details/info/BVDLUNKNOWN00001", "", 200, "success")
	if m := miss.dataMap(t); m != nil {
		t.Fatalf("missing info data = %v, want null", m)
	}

	// BiliBili API error surfaces as an envelope error.
	errResp := expectStatus(t, e, "GET", "/api/video_details/fetch/BVDLMISSING000001", "", 200, "error")
	expectMessageContains(t, errResp, "获取视频详情失败")
	expectMessageContains(t, errResp, "code=-404")

	// Garbage upstream JSON also goes through the same error branch.
	expectMessageContains(t, expectStatus(t, e, "GET", "/api/video_details/fetch/BVDLFTBADJSON01", "", 200, "error"), "获取视频详情失败")
}

func TestDLVideoSearch(t *testing.T) {
	dlSaveVideo(t, &models.VideoBaseInfo{
		Bvid: "BVDLSE000000001", Title: "DL搜索标题一", Aid: 1, OwnerMid: 706001,
		OwnerName: "DL搜索UP", StatView: 5, FetchTime: time.Now().Unix(),
	})
	dlSaveVideo(t, &models.VideoBaseInfo{
		Bvid: "BVDLSE000000002", Title: "DL搜索标题二", Aid: 2, OwnerMid: 706002,
		OwnerName: "DL搜索UP", StatView: 7, FetchTime: time.Now().Unix(),
	})
	e := newAPI(t, RegisterVideoDetailsRoutes)

	expectMessage(t, expectStatus(t, e, "GET", "/api/video_details/search", "", 400, "error"), "搜索关键词不能为空")

	resp := expectStatus(t, e, "GET", "/api/video_details/search?keyword="+urlEncode("DL搜索标题"), "", 200, "success")
	data := resp.dataMap(t)
	records := dataArray(t, data["records"])
	if len(records) != 2 || int(data["total"].(float64)) != 2 {
		t.Fatalf("search hits = %d total = %v, want 2/2", len(records), data["total"])
	}
	if data["keyword"] != "DL搜索标题" || int(data["size"].(float64)) != 20 || int(data["current"].(float64)) != 1 {
		t.Fatalf("search echo = %v", data)
	}
	// Newest pubdate first; both are zero here, so only check the fields exist.
	first := records[0].(map[string]interface{})
	if !strings.HasPrefix(fmt.Sprint(first["bvid"]), "BVDLSE") || first["original_url"] == nil {
		t.Fatalf("record shape = %v", first)
	}

	// Owner-name search matches both rows too.
	byOwner := expectStatus(t, e, "GET", "/api/video_details/search?keyword="+urlEncode("DL搜索UP"), "", 200, "success").dataMap(t)
	if int(byOwner["total"].(float64)) != 2 {
		t.Fatalf("owner search total = %v, want 2", byOwner["total"])
	}

	// Out-of-range paging parameters are normalised to page=1 size=20.
	odd := expectStatus(t, e, "GET", "/api/video_details/search?keyword=DL搜索标题&page=0&size=999", "", 200, "success").dataMap(t)
	if int(odd["size"].(float64)) != 20 || int(odd["current"].(float64)) != 1 {
		t.Fatalf("normalisation = size %v current %v, want 20/1", odd["size"], odd["current"])
	}

	// No match: total 0 with a null records payload.
	none := expectStatus(t, e, "GET", "/api/video_details/search?keyword="+urlEncode("DL不存在关键词XYZ"), "", 200, "success").dataMap(t)
	if int(none["total"].(float64)) != 0 {
		t.Fatalf("no-match total = %v, want 0", none["total"])
	}
	if none["records"] != nil {
		t.Fatalf("no-match records = %v, want null", none["records"])
	}
}

func TestDLBatchFetchValidationOffline(t *testing.T) {
	dlDrainProgressChan()
	dlWaitNotProcessing(t, 5*time.Second)
	e := newAPI(t, RegisterVideoDetailsRoutes)

	// Malformed JSON body is rejected before any work happens.
	expectStatus(t, e, "POST", "/api/video_details/batch_fetch", "{invalid", 400, "error")

	// No bvids -> BatchFetchFromHistory fallback; empty SESSDATA aborts offline.
	for _, body := range []string{`{}`, `{"bvids":[]}`} {
		expectMessage(t, expectStatus(t, e, "POST", "/api/video_details/batch_fetch", body, 200, "error"), "SESSDATA未配置")
	}
	expectMessage(t, expectStatus(t, e, "GET", "/api/video_details/batch_fetch_from_history", "", 200, "error"), "SESSDATA未配置")

	// Stop without a running task errors out; reset always succeeds.
	expectMessage(t, expectStatus(t, e, "POST", "/api/video_details/stop", "", 200, "error"), "没有正在进行的获取任务")
	reset := expectStatus(t, e, "POST", "/api/video_details/reset", "", 200, "success").dataMap(t)
	if reset["message"] != "已重置获取状态" {
		t.Fatalf("reset payload = %v", reset)
	}
	if st := services.GetVideoDetailProgress(); st.Status != "idle" || st.IsProcessing {
		t.Fatalf("progress after reset = %+v", st)
	}
}

func TestDLBatchFetchProgressAndStop(t *testing.T) {
	dlDrainProgressChan()
	dlWaitNotProcessing(t, 5*time.Second)
	const slowBvid = "BVDLSLOW00000001"
	dlStubView(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("bvid") == slowBvid {
			time.Sleep(1500 * time.Millisecond)
		}
		fmt.Fprint(w, dlViewPayload(r.URL.Query().Get("bvid"), "", 707001))
	})
	restore := setSESSDATA(t, "dl-batch-token")
	defer restore()
	e := newAPI(t, RegisterVideoDetailsRoutes)

	// The loop only checks the stop flag at the top of each iteration, so the
	// stop is scheduled 600ms after start: iteration 0 (fast video, ~200ms
	// pacing sleep) has finished, iteration 1 (1.5s slow fetch) is mid-flight.
	// The in-flight slow fetch is NOT interrupted: it completes and persists,
	// taking the counters to 2/2/0, and only the iteration-2 top check
	// observes the stop flag and breaks. Third bvid never starts.
	body := fmt.Sprintf(`{"bvids":["BVDLFAST000000001",%q,"BVDLFAST000000002"]}`, slowBvid)
	started := expectStatus(t, e, "POST", "/api/video_details/batch_fetch", body, 200, "success").dataMap(t)
	if started["message"] != "开始批量获取视频详情" || int(started["total"].(float64)) != 3 {
		t.Fatalf("start payload = %v", started)
	}

	// While the task runs, both batch entry points must refuse.
	expectMessage(t, expectStatus(t, e, "POST", "/api/video_details/batch_fetch", body, 200, "error"), "已有获取任务正在运行")
	expectMessage(t, expectStatus(t, e, "GET", "/api/video_details/batch_fetch_from_history", "", 200, "error"), "已有获取任务正在运行")

	time.Sleep(600 * time.Millisecond)
	expectStatus(t, e, "POST", "/api/video_details/stop", "", 200, "success")
	dlWaitNotProcessing(t, 30*time.Second)

	st := services.GetVideoDetailProgress()
	if st.Status != "stopped" || !st.IsStopped || st.IsComplete {
		t.Fatalf("final progress = %+v, want stopped/isStopped", st)
	}
	if st.ProcessedVideos != 2 || st.SuccessCount != 2 || st.FailedCount != 0 {
		t.Fatalf("final counters = %+v, want processed/success 2/2/0", st)
	}
	if got := dlQueryInt64(t, `SELECT COUNT(*) FROM video_base_info WHERE bvid = 'BVDLFAST000000001'`); got != 1 {
		t.Fatalf("first fast bvid persisted = %d, want 1", got)
	}
	if got := dlQueryInt64(t, `SELECT COUNT(*) FROM video_base_info WHERE bvid = ?`, slowBvid); got != 1 {
		t.Fatalf("in-flight slow bvid persisted = %d, want 1 (stop only observed at the next loop-top check)", got)
	}
	if got := dlQueryInt64(t, `SELECT COUNT(*) FROM video_base_info WHERE bvid = 'BVDLFAST000000002'`); got != 0 {
		t.Fatalf("never-started bvid persisted = %d, want 0", got)
	}

	expectStatus(t, e, "POST", "/api/video_details/reset", "", 200, "success")
	dlDrainProgressChan()
}

func TestDLBatchFetchFromHistory(t *testing.T) {
	dlDrainProgressChan()
	dlWaitNotProcessing(t, 5*time.Second)
	for i, bvid := range []string{"BVDLARCH00000001", "BVDLARCH00000002", "BVDLARCH00000003"} {
		insertHistory(t, 2002, map[string]interface{}{
			"bvid": bvid, "business": "archive", "tag_name": "",
			"title": "DL归档视频", "view_at": viewAt(t, 2002, time.February, 1+i, 10, 0),
		})
	}
	allBvids, err := database.GetUniqueBvidsFromHistory()
	if err != nil {
		t.Fatalf("GetUniqueBvidsFromHistory: %v", err)
	}

	dlStubView(t, func(w http.ResponseWriter, r *http.Request) {
		// tname stays empty so the deferred tag-name sync cannot touch the
		// history rows owned by other test files.
		fmt.Fprint(w, dlViewPayload(r.URL.Query().Get("bvid"), "", 708001))
	})
	restore := setSESSDATA(t, "dl-history-token")
	defer restore()
	e := newAPI(t, RegisterVideoDetailsRoutes)

	resp := expectStatus(t, e, "GET", "/api/video_details/batch_fetch_from_history", "", 200, "success").dataMap(t)
	if resp["message"] != "开始从历史记录批量获取视频详情" {
		t.Fatalf("start message = %v", resp["message"])
	}
	if int(resp["history_total"].(float64)) != len(allBvids) {
		t.Fatalf("history_total = %v, want %d", resp["history_total"], len(allBvids))
	}
	if int(resp["total"].(float64)) < 3 {
		t.Fatalf("total = %v, want at least the three seeded archive rows", resp["total"])
	}
	dlWaitNotProcessing(t, 60*time.Second)
	for _, bvid := range []string{"BVDLARCH00000001", "BVDLARCH00000002", "BVDLARCH00000003"} {
		if got := dlQueryInt64(t, `SELECT COUNT(*) FROM video_base_info WHERE bvid = ?`, bvid); got != 1 {
			t.Fatalf("%s persisted = %d, want 1", bvid, got)
		}
	}

	// Every archive bvid is now fetched, so skip_existing=true has nothing left.
	done := expectStatus(t, e, "GET", "/api/video_details/batch_fetch_from_history", "", 200, "success").dataMap(t)
	if done["message"] != "没有需要获取的视频" || int(done["total"].(float64)) != 0 {
		t.Fatalf("second pass = %v, want the nothing-to-fetch branch", done)
	}

	// skip_existing=false re-queues every archive bvid (the else branch).
	force := expectStatus(t, e, "GET", "/api/video_details/batch_fetch_from_history?skip_existing=false", "", 200, "success").dataMap(t)
	if int(force["total"].(float64)) != len(allBvids) {
		t.Fatalf("force total = %v, want %d", force["total"], len(allBvids))
	}
	dlWaitNotProcessing(t, 60*time.Second)
	dlDrainProgressChan()
}

func TestDLVideoDetailStats(t *testing.T) {
	if err := database.EnsureVideoDetailsTables(); err != nil {
		t.Fatalf("ensure tables: %v", err)
	}
	dlSaveVideo(t, &models.VideoBaseInfo{
		Bvid: "BVDLST000000001", Title: "DL统计视频", Aid: 1, OwnerMid: 709001,
		StatView: 111, StatLike: 22, Duration: 60, FetchTime: time.Now().Unix(),
	})
	e := newAPI(t, RegisterVideoDetailsRoutes)

	hBvids, err := database.GetUniqueBvidsFromHistory()
	if err != nil {
		t.Fatalf("unique bvids: %v", err)
	}
	data := expectStatus(t, e, "GET", "/api/video_details/stats", "", 200, "success").dataMap(t)
	if int(data["total_videos"].(float64)) != len(hBvids) {
		t.Fatalf("total_videos = %v, want %d", data["total_videos"], len(hBvids))
	}
	fetched := dlQueryInt64(t, `SELECT COUNT(*) FROM video_base_info`)
	invalid := dlQueryInt64(t, `SELECT COUNT(*) FROM invalid_videos`)
	if int64(data["fetched_videos"].(float64)) != fetched {
		t.Fatalf("fetched_videos = %v, want %d", data["fetched_videos"], fetched)
	}
	if int64(data["invalid_videos_count"].(float64)) != invalid {
		t.Fatalf("invalid_videos_count = %v, want %d", data["invalid_videos_count"], invalid)
	}
	wantPending := int64(len(hBvids)) - fetched - invalid
	if wantPending < 0 {
		wantPending = 0
	}
	if int64(data["pending_videos_count"].(float64)) != wantPending {
		t.Fatalf("pending_videos_count = %v, want %d", data["pending_videos_count"], wantPending)
	}
	if int64(data["videos_without_details"].(float64)) != wantPending ||
		int64(data["videos_with_details"].(float64)) != fetched {
		t.Fatalf("alias fields = %v/%v", data["videos_without_details"], data["videos_with_details"])
	}
}

func TestDLDatabaseStats(t *testing.T) {
	// 自建依赖：video_base_info 等表由别的用例创建，-shuffle 下本用例可能先跑，
	// 那时 GetDatabaseStats 走 TableExists=false 分支，信封里根本没有 total_videos。
	if err := database.EnsureVideoDetailsTables(); err != nil {
		t.Fatalf("ensure tables: %v", err)
	}
	dlSaveVideo(t, &models.VideoBaseInfo{
		Bvid: "BVDLDBST0000001", Title: "DL库统计视频", Aid: 2, OwnerMid: 709002,
		StatView: 111, StatLike: 22, Duration: 60, FetchTime: time.Now().Unix(),
	})
	e := newAPI(t, RegisterVideoDetailsRoutes)
	data := expectStatus(t, e, "GET", "/api/video_details/database_stats", "", 200, "success").dataMap(t)

	cases := map[string]string{
		"total_videos":           `SELECT COUNT(*) FROM video_base_info`,
		"total_views":            `SELECT COALESCE(SUM(stat_view),0) FROM video_base_info`,
		"total_likes":            `SELECT COALESCE(SUM(stat_like),0) FROM video_base_info`,
		"total_duration_seconds": `SELECT COALESCE(SUM(duration),0) FROM video_base_info`,
		"total_uploaders":        `SELECT COUNT(*) FROM uploader_info`,
		"total_unique_tags":      `SELECT COUNT(DISTINCT tag_name) FROM video_tags`,
	}
	for key, q := range cases {
		if int64(data[key].(float64)) != dlQueryInt64(t, q) {
			t.Fatalf("%s = %v, want the live SQL value %q", key, data[key], q)
		}
	}
	if int64(data["total_views"].(float64)) < 111 {
		t.Fatalf("total_views = %v, want at least the DL统计视频 contribution", data["total_views"])
	}
}

func TestDLUploaderListTagsAndDetail(t *testing.T) {
	// Remove the uploader_info row this test owns (inserted below) so the
	// pre-seed "no fans" branch also holds on repeated runs against the
	// shared package database.
	if _, err := db(t).Exec(`DELETE FROM uploader_info WHERE mid = 71100001`); err != nil {
		t.Fatalf("reset uploader_info: %v", err)
	}
	now := time.Now().Unix()
	for i, stat := range []struct{ view, like int }{{100, 10}, {200, 20}, {300, 30}} {
		dlSaveVideo(t, &models.VideoBaseInfo{
			Bvid: fmt.Sprintf("BVDLUP%09d", i+1), Title: fmt.Sprintf("DL榜一视频%d", i+1),
			Aid: i + 1, OwnerMid: 71100001, OwnerName: "DL榜单UP", OwnerFace: "http://face/dl1.jpg",
			StatView: stat.view, StatLike: stat.like, StatCoin: i + 1, StatFavorite: i + 2,
			FetchTime: now,
		})
	}
	dlSaveVideo(t, &models.VideoBaseInfo{
		Bvid: "BVDLUP000000010", Title: "DL榜二视频", Aid: 10, OwnerMid: 71100002,
		OwnerName: "DL小号UP", StatView: 50, FetchTime: now,
	})
	if _, err := db(t).Exec(`INSERT OR REPLACE INTO video_tags (bvid, tag_id, tag_name) VALUES
		('BVDLUP000000001', 930001, 'DL标签甲'),
		('BVDLUP000000002', 930002, 'DL标签甲'),
		('BVDLUP000000003', 930003, 'DL标签乙')`); err != nil {
		t.Fatalf("seed tags: %v", err)
	}

	e := newAPI(t, RegisterVideoDetailsRoutes)

	ups := expectStatus(t, e, "GET", "/api/video_details/uploaders?size=100", "", 200, "success").dataMap(t)
	if ups["sort_by"] != "video_count" || int(ups["size"].(float64)) != 100 {
		t.Fatalf("uploader echo = %v", ups)
	}
	top := dlUploaderByMid(t, dataArray(t, ups["records"]), 71100001)
	if int(top["video_count"].(float64)) != 3 || int(top["total_views"].(float64)) != 600 ||
		int(top["total_likes"].(float64)) != 60 || top["name"] != "DL榜单UP" {
		t.Fatalf("top uploader record = %v", top)
	}

	// Unknown sorts fall back to video_count; views is honoured.
	for q, want := range map[string]string{"sort_by=views": "views", "sort_by=bogus": "video_count", "page=0&size=999": "video_count"} {
		resp := expectStatus(t, e, "GET", "/api/video_details/uploaders?"+q, "", 200, "success").dataMap(t)
		if resp["sort_by"] != want {
			t.Fatalf("%s -> sort_by = %v, want %v", q, resp["sort_by"], want)
		}
	}
	byViews := expectStatus(t, e, "GET", "/api/video_details/uploaders?sort_by=views&size=100", "", 200, "success").dataMap(t)
	if int(dlIntOf(t, byViews["total"])) != int(dlQueryInt64(t, `SELECT COUNT(DISTINCT owner_mid) FROM video_base_info WHERE owner_mid > 0`)) {
		t.Fatalf("uploader total = %v, want distinct owner_mid count", byViews["total"])
	}

	tags := expectStatus(t, e, "GET", "/api/video_details/tags?size=100", "", 200, "success").dataMap(t)
	if int(dlIntOf(t, tags["total"])) != int(dlQueryInt64(t, `SELECT COUNT(DISTINCT tag_name) FROM video_tags`)) {
		t.Fatalf("tag total = %v", tags["total"])
	}
	jia := dlTagByName(t, dataArray(t, tags["records"]), "DL标签甲")
	if int(jia["video_count"].(float64)) != 2 {
		t.Fatalf("DL标签甲 count = %v, want 2", jia["video_count"])
	}
	// Parameter normalisation on the tag route too.
	tune := expectStatus(t, e, "GET", "/api/video_details/tags?page=-3&size=0", "", 200, "success").dataMap(t)
	if int(tune["current"].(float64)) != 1 || int(tune["size"].(float64)) != 20 {
		t.Fatalf("tag normalisation = %v/%v, want 1/20", tune["current"], tune["size"])
	}

	expectMessage(t, expectStatus(t, e, "GET", "/api/video_details/uploader/abc", "", 400, "error"), "无效的UP主ID")

	detail := expectStatus(t, e, "GET", "/api/video_details/uploader/71100001", "", 200, "success").dataMap(t)
	if detail["name"] != "DL榜单UP" || int(detail["video_count"].(float64)) != 3 ||
		int(detail["total_views"].(float64)) != 600 || int(detail["total_likes"].(float64)) != 60 ||
		int(detail["total_coins"].(float64)) != 6 {
		t.Fatalf("uploader detail = %v", detail)
	}
	if _, ok := detail["fans"]; ok {
		t.Fatalf("fans should be absent before uploader_info row exists: %v", detail)
	}
	if _, err := db(t).Exec(`INSERT OR REPLACE INTO uploader_info (mid, name, sex, face, sign, level, fans, attention, archive_count, fetch_time, update_time)
		VALUES (71100001, 'DL榜单UP', '男', 'http://face/dl1.jpg', 'DL签名', 6, 4321, 12, 34, 1600000000, 1600000001)`); err != nil {
		t.Fatalf("seed uploader_info: %v", err)
	}
	withInfo := expectStatus(t, e, "GET", "/api/video_details/uploader/71100001", "", 200, "success").dataMap(t)
	if int(withInfo["fans"].(float64)) != 4321 || withInfo["sign"] != "DL签名" || int(withInfo["level"].(float64)) != 6 {
		t.Fatalf("uploader_info merge = %v", withInfo)
	}

	// Unknown mid still answers successfully with zeroed counters.
	unknown := expectStatus(t, e, "GET", "/api/video_details/uploader/71199999", "", 200, "success").dataMap(t)
	if int(unknown["video_count"].(float64)) != 0 || unknown["name"] != "" {
		t.Fatalf("unknown uploader = %v", unknown)
	}
}

func TestDLVideoDetailProgressSSE(t *testing.T) {
	e := newAPI(t, RegisterVideoDetailsRoutes)
	ch := services.GetVideoDetailProgressChan()

	// Completed frame: the handler must flush it and return.
	dlDrainProgressChan()
	ch <- models.VideoDetailProgress{IsComplete: true, Status: "completed", ProcessedVideos: 3, TotalVideos: 3, ProgressPercent: 100}
	w := doRaw(t, e, "GET", "/api/video_details/progress", "")
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("progress content-type = %q", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"is_complete":true`) || !strings.Contains(body, `"processed_videos":3`) {
		t.Fatalf("progress SSE body = %q", body)
	}
	if strings.Count(body, "data: ") != 1 {
		t.Fatalf("progress should stop after the terminal frame: %q", body)
	}

	// Running frame followed by an error terminal frame: two events.
	dlDrainProgressChan()
	ch <- models.VideoDetailProgress{Status: "running", ProcessedVideos: 1, TotalVideos: 2}
	ch <- models.VideoDetailProgress{Status: "error", ErrorMessage: "dl坏了"}
	w = doRaw(t, e, "GET", "/api/video_details/progress", "")
	body = w.Body.String()
	if strings.Count(body, "data: ") != 2 || !strings.Contains(body, "dl坏了") {
		t.Fatalf("error-terminal stream = %q", body)
	}

	// A stopped frame also terminates the stream.
	dlDrainProgressChan()
	ch <- models.VideoDetailProgress{Status: "stopped", IsStopped: true}
	body = doRaw(t, e, "GET", "/api/video_details/progress", "").Body.String()
	if strings.Count(body, "data: ") != 1 || !strings.Contains(body, `"is_stopped":true`) {
		t.Fatalf("stopped stream = %q", body)
	}
	dlDrainProgressChan()
}

func TestDLSyncHistoryTagNames(t *testing.T) {
	dlSaveVideo(t, &models.VideoBaseInfo{
		Bvid: "BVDLSYNC00000001", Title: "DL同步视频", Aid: 11, OwnerMid: 712001,
		Tname: "音乐", FetchTime: time.Now().Unix(),
	})
	insertHistory(t, 2002, map[string]interface{}{
		"bvid": "BVDLSYNC00000001", "business": "archive", "tag_name": "",
		"title": "DL同步视频", "view_at": viewAt(t, 2002, time.March, 1, 8, 0),
	})
	// insertHistory is idempotent (bvid+view_at), so on repeated runs the row
	// keeps the tag_name written by the previous sync; force it back to empty
	// so the update branch has work to do again.
	if _, err := db(t).Exec(`UPDATE bilibili_history_2002 SET tag_name = '' WHERE bvid = 'BVDLSYNC00000001'`); err != nil {
		t.Fatalf("reset tag_name: %v", err)
	}
	e := newAPI(t, RegisterVideoDetailsRoutes)

	resp := expectStatus(t, e, "POST", "/api/video_details/sync_tag_names", "", 200, "success").dataMap(t)
	taskID, _ := resp["task_id"].(string)
	if taskID == "" || resp["message"] != "历史分区标签同步任务已启动，正在后台执行" {
		t.Fatalf("sync payload = %v", resp)
	}

	deadline := time.Now().Add(15 * time.Second)
	var status map[string]interface{}
	for time.Now().Before(deadline) {
		status = scheduler.GetAsyncTaskStatus(taskID)
		if s, _ := status["status"].(string); s == "completed" || s == "failed" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if status == nil || status["status"] != "completed" {
		t.Fatalf("async task status = %+v, want completed", status)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(fmt.Sprint(status["result"])), &result); err != nil {
		t.Fatalf("task result %v: %v", status["result"], err)
	}
	if int(result["updated"].(float64)) < 1 {
		t.Fatalf("updated = %v, want at least 1", result["updated"])
	}

	var tagName string
	if err := db(t).QueryRow(`SELECT tag_name FROM bilibili_history_2002 WHERE bvid = 'BVDLSYNC00000001'`).Scan(&tagName); err != nil {
		t.Fatalf("read back tag_name: %v", err)
	}
	if tagName != "音乐" {
		t.Fatalf("tag_name = %q, want 音乐", tagName)
	}
}

func dlUploaderByMid(t *testing.T, records []interface{}, mid int) map[string]interface{} {
	t.Helper()
	for _, r := range records {
		m := r.(map[string]interface{})
		if int(m["mid"].(float64)) == mid {
			return m
		}
	}
	t.Fatalf("uploader mid %d missing from %v", mid, records)
	return nil
}

func dlTagByName(t *testing.T, records []interface{}, name string) map[string]interface{} {
	t.Helper()
	for _, r := range records {
		m := r.(map[string]interface{})
		if m["tag_name"] == name {
			return m
		}
	}
	t.Fatalf("tag %q missing from %v", name, records)
	return nil
}

func dlIntOf(t *testing.T, v interface{}) int {
	t.Helper()
	n, ok := v.(float64)
	if !ok {
		t.Fatalf("value %v is not a JSON number", v)
	}
	return int(n)
}
