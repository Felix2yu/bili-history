package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"bilibili-history-go/biliapi"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Coverage tests for backend/routers/misc.go.
//
// All rows live in the reserved year table (2009) with the reserved BVMS bvid
// prefix so they never collide with the fixtures the other coverage agents
// seed in 2006/2007/2010/2011/2012/2014/2016. Bilibili traffic is intercepted
// with an httptest server (biliapi URL vars repointed for fetch/delete, and a
// DefaultTransport stub for the apprise notifications) so nothing leaves the
// machine. config mutations are always wrapped in msCfgSnapshot / msSetCreds so
// the process-wide singleton and config.yaml are restored afterwards.
// ---------------------------------------------------------------------------

func msMiscAPI(t *testing.T) *gin.Engine {
	t.Helper()
	return newAPI(t,
		RegisterConfigRoutes,
		RegisterSchedulerRoutes,
		RegisterDataSyncRoutes,
		RegisterExportRoutes,
		RegisterImportRoutes,
		RegisterCleanRoutes,
		RegisterLogRoutes,
		RegisterFetchRoutes,
		RegisterDeleteRoutes,
		RegisterInteractionRoutes,
	)
}

// msSeed2009 writes two deterministic history rows into year 2009. The first
// carries oid=2009011 so GetBvidByOid can resolve it for the single-delete
// handler. insertHistory is idempotent on (bvid, view_at).
func msSeed2009(t *testing.T) (string, string) {
	t.Helper()
	ensureYear(t, 2009)
	bvidA := uniqueBvid("BVMS", 11)
	bvidB := uniqueBvid("BVMS", 12)
	insertHistory(t, 2009, map[string]interface{}{
		"bvid":     bvidA,
		"oid":      2009011,
		"view_at":  viewAt(t, 2009, time.January, 1, 10, 0),
		"title":    "MS 视频A",
		"business": "archive",
	})
	insertHistory(t, 2009, map[string]interface{}{
		"bvid":     bvidB,
		"oid":      2009012,
		"view_at":  viewAt(t, 2009, time.January, 2, 11, 0),
		"title":    "MS 视频B",
		"business": "archive",
	})
	return bvidA, bvidB
}

// msWithDelStub repoints biliapi.HistoryDelURL at a local httptest server for
// the duration of fn and restores the original value afterwards.
func msWithDelStub(t *testing.T, handler http.HandlerFunc, fn func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	defer srv.Close()
	prev := biliapi.HistoryDelURL
	biliapi.HistoryDelURL = srv.URL
	defer func() { biliapi.HistoryDelURL = prev }()
	fn()
}

// msWithHistoryStub repoints biliapi.HistoryURL at a local httptest server.
func msWithHistoryStub(t *testing.T, handler http.HandlerFunc, fn func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	defer srv.Close()
	prev := biliapi.HistoryURL
	biliapi.HistoryURL = srv.URL
	defer func() { biliapi.HistoryURL = prev }()
	fn()
}

// msEmptyHistoryHandler answers the cursor endpoint with an empty page so the
// synchronous fetch loop terminates after a single request.
func msEmptyHistoryHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"code":0,"message":"0","data":{"cursor":{"max":0,"view_at":0,"ps":30},"list":[]}}`))
}

// ===========================================================================
// Config routes
// ===========================================================================

func TestMSConfigNotifyRoundTrip(t *testing.T) {
	msCfgSnapshot(t)
	e := msMiscAPI(t)

	resp := expectStatus(t, e, http.MethodGet, "/api/config/notify", "", 200, "success")
	if resp.dataMap(t) == nil && resp.Data == nil {
		t.Fatalf("get notify returned no data")
	}

	// invalid JSON -> 400
	bad := expectStatus(t, e, http.MethodPost, "/api/config/notify", "{oops", 400, "error")
	expectMessageContains(t, bad, "参数错误")

	// valid save -> 200 (bare gin.H)
	w := doRaw(t, e, http.MethodPost, "/api/config/notify", `{"enabled":false,"urls":[]}`)
	if m := msRawMap(t, w); m["message"] != "通知配置已保存" {
		t.Fatalf("save notify resp = %v", m)
	}
}

func TestMSConfigTestNotify(t *testing.T) {
	msCfgSnapshot(t)
	cfg := loadCfg(t)
	cfg.Notify.Enabled = false
	cfg.Notify.URLs = nil
	e := msMiscAPI(t)

	// disabled -> 400
	resp := expectStatus(t, e, http.MethodPost, "/api/config/notify/test", "", 400, "error")
	expectMessageContains(t, resp, "通知未启用或未配置URL")
}

func TestMSConfigTestNotifySuccess(t *testing.T) {
	titles := msEnableNotifyForTest(t)
	e := msMiscAPI(t)
	w := doRaw(t, e, http.MethodPost, "/api/config/notify/test", "")
	m := msRawMap(t, w)
	if m["status"] != "success" {
		t.Fatalf("test notify resp = %v (body=%s)", m, w.Body.String())
	}
	if len(titles()) == 0 {
		t.Fatalf("expected a test notification to be delivered")
	}
}

func TestMSConfigServerRoundTrip(t *testing.T) {
	msCfgSnapshot(t)
	e := msMiscAPI(t)

	resp := expectStatus(t, e, http.MethodGet, "/api/config/server", "", 200, "success")
	dm := resp.dataMap(t)
	if dm == nil {
		t.Fatalf("server config data missing")
	}

	bad := expectStatus(t, e, http.MethodPost, "/api/config/server", "{oops", 400, "error")
	expectMessageContains(t, bad, "参数错误")

	w := doRaw(t, e, http.MethodPost, "/api/config/server", `{"host":"127.0.0.1","port":8899,"ssl_enabled":false}`)
	if m := msRawMap(t, w); m["message"] != "服务器配置已保存" {
		t.Fatalf("save server resp = %v", m)
	}
}

func TestMSConfigMCPGet(t *testing.T) {
	msCfgSnapshot(t)
	cfg := loadCfg(t)
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = 8899
	cfg.Mcp.Path = ""
	cfg.Mcp.Token = ""
	e := msMiscAPI(t)

	w := doRaw(t, e, http.MethodGet, "/api/config/mcp-config", "")
	m := msRawMap(t, w)
	if m["status"] != "success" {
		t.Fatalf("get mcp resp = %v", m)
	}
	if m["path"] != "/mcp" {
		t.Fatalf("default path = %v, want /mcp", m["path"])
	}
	if m["mcp_url"] != "http://127.0.0.1:8899/mcp/" {
		t.Fatalf("mcp_url = %v", m["mcp_url"])
	}
	if m["token_configured"] != false {
		t.Fatalf("token_configured = %v, want false", m["token_configured"])
	}
}

func TestMSConfigMCPSaveEnabledGeneratesToken(t *testing.T) {
	msCfgSnapshot(t)
	cfg := loadCfg(t)
	cfg.Mcp.Token = ""
	cfg.Mcp.Path = ""
	cfg.Mcp.MaxPageSize = 0
	e := msMiscAPI(t)

	w := doRaw(t, e, http.MethodPost, "/api/config/mcp-config", `{"enabled":true}`)
	m := msRawMap(t, w)
	if m["message"] != "MCP配置已保存" {
		t.Fatalf("save mcp resp = %v", m)
	}
	if m["token_configured"] != true {
		t.Fatalf("token_configured = %v, want true after auto-generating a token", m["token_configured"])
	}
	if m["path"] != "/mcp" || m["max_page_size"].(float64) != 100 {
		t.Fatalf("defaults not applied: path=%v max_page_size=%v", m["path"], m["max_page_size"])
	}
	if m["restart_required"] != true {
		t.Fatalf("restart_required = %v, want true", m["restart_required"])
	}
}

func TestMSConfigMCPSaveBadJSON(t *testing.T) {
	msCfgSnapshot(t)
	e := msMiscAPI(t)
	resp := expectStatus(t, e, http.MethodPost, "/api/config/mcp-config", "not-json", 400, "error")
	expectMessageContains(t, resp, "参数错误")
}

// ===========================================================================
// Data-sync routes
// ===========================================================================

func TestMSDataSyncStatusAndConfigs(t *testing.T) {
	msCfgSnapshot(t)
	e := msMiscAPI(t)

	w := doRaw(t, e, http.MethodGet, "/api/data_sync/status", "")
	// dataSyncStatus is a process-wide singleton that other tests may have
	// already driven to "completed", so only assert the response shape here.
	sStatus, _ := msRawMap(t, w)["status"].(string)
	if sStatus != "idle" && sStatus != "completed" {
		t.Fatalf("data sync status = %q, want idle or completed", sStatus)
	}

	if m := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/data_sync/config", "")); m["success"] != true {
		t.Fatalf("data sync config resp = %v", m)
	}
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/data_sync/config", `{"check_on_startup":true}`)); m["message"] != "配置已更新" {
		t.Fatalf("update data sync config resp = %v", m)
	}

	if m := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/data_sync/sync-config", "")); m["success"] != true {
		t.Fatalf("sync config resp = %v", m)
	}
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/data_sync/sync-config", `{"sync_deleted":true,"sync_delete_to_bilibili":false}`)); m["message"] != "同步配置已更新" {
		t.Fatalf("update sync config resp = %v", m)
	}

	if m := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/data_sync/appearance-config", "")); m["success"] != true {
		t.Fatalf("appearance config resp = %v", m)
	}
	bad := doRaw(t, e, http.MethodPost, "/api/data_sync/appearance-config", `{"dark_mode":"neon"}`)
	if m := msRawMap(t, bad); m["success"] != false {
		t.Fatalf("invalid dark_mode resp = %v, want success=false", m)
	}
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/data_sync/appearance-config", `{"dark_mode":"dark"}`)); m["message"] != "外观配置已更新" {
		t.Fatalf("update appearance config resp = %v", m)
	}

	// check (force) counts DB records across all years
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/data_sync/check", `{"force_check":true}`)); m["success"] != true {
		t.Fatalf("integrity check resp = %v", m)
	}

	// report
	if m := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/data_sync/report", "")); m["data"] == nil {
		t.Fatalf("integrity report resp = %v", m)
	}
}

func TestMSDataSyncConfigBadJSON(t *testing.T) {
	msCfgSnapshot(t)
	e := msMiscAPI(t)
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/data_sync/config", "{oops")); m["success"] != false {
		t.Fatalf("bad data sync config resp = %v", m)
	}
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/data_sync/sync-config", "{oops")); m["success"] != false {
		t.Fatalf("bad sync config resp = %v", m)
	}
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/data_sync/appearance-config", "{oops")); m["success"] != false {
		t.Fatalf("bad appearance config resp = %v", m)
	}
}

func TestMSDataSyncSyncThenResult(t *testing.T) {
	e := msMiscAPI(t)

	// /sync runs RunSyncData (always succeeds) and stores the last result.
	w := doRaw(t, e, http.MethodPost, "/api/data_sync/sync", "")
	m := msRawMap(t, w)
	if m["success"] != true {
		t.Fatalf("sync resp = %v", m)
	}

	// /sync/result now returns the stored result (non-nil branch).
	w2 := doRaw(t, e, http.MethodGet, "/api/data_sync/sync/result", "")
	m2 := msRawMap(t, w2)
	if m2["success"] != true {
		t.Fatalf("sync result after sync resp = %v", m2)
	}
}

// ===========================================================================
// Import / clean / log / interactions stubs
// ===========================================================================

func TestMSImportRoutes(t *testing.T) {
	e := msMiscAPI(t)
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/importMysql/start", "")); m["message"] != "MySQL导入功能待实现" {
		t.Fatalf("import mysql start resp = %v", m)
	}
	expectStatus(t, e, http.MethodGet, "/api/importMysql/status", "", 200, "success")
	expectStatus(t, e, http.MethodGet, "/api/importSqlite/status", "", 200, "success")
	// importFromSqlite aggregates every year table; only assert the envelope.
	expectStatus(t, e, http.MethodPost, "/api/importSqlite/start", "", 200, "success")
	expectStatus(t, e, http.MethodPost, "/api/importSqlite/import_data_sqlite", "", 200, "success")
}

func TestMSCleanRoutes(t *testing.T) {
	e := msMiscAPI(t)
	bad := expectStatus(t, e, http.MethodPost, "/api/clean/start", "{oops", 400, "error")
	expectMessageContains(t, bad, "参数错误")

	// all-false options -> no-op goroutine, safe (touches nothing).
	w := doRaw(t, e, http.MethodPost, "/api/clean/start", `{"clean_duplicates":false,"clean_invalid_videos":false,"clean_old_history":false,"clean_image_cache":false,"clean_logs":false}`)
	if m := msRawMap(t, w); m["message"] != "开始数据清洗" {
		t.Fatalf("clean start resp = %v", m)
	}
	expectStatus(t, e, http.MethodGet, "/api/clean/status", "", 200, "success")
}

func TestMSLogRoutes(t *testing.T) {
	e := msMiscAPI(t)
	expectStatus(t, e, http.MethodGet, "/api/log/list", "", 200, "success")
}

func TestMSLogSendDisabled(t *testing.T) {
	msCfgSnapshot(t)
	cfg := loadCfg(t)
	cfg.Notify.Enabled = false
	cfg.Notify.URLs = nil
	e := msMiscAPI(t)
	// stats with >1 key skips auto-collect; notify disabled -> 500.
	resp := expectStatus(t, e, http.MethodPost, "/api/log/send", `{"report_date":"2009-01-01","today_records":2}`, 500, "error")
	expectMessageContains(t, resp, "发送每日报告失败")
}

func TestMSLogSendEnabled(t *testing.T) {
	titles := msEnableNotifyForTest(t)
	e := msMiscAPI(t)
	w := doRaw(t, e, http.MethodPost, "/api/log/send", `{"report_date":"2009-01-01","today_records":2}`)
	m := msRawMap(t, w)
	if m["message"] != "每日报告已发送" {
		t.Fatalf("log send resp = %v (body=%s)", m, w.Body.String())
	}
	if len(titles()) == 0 {
		t.Fatalf("expected daily report notification")
	}
}

func TestMSInteractionRoutes(t *testing.T) {
	e := msMiscAPI(t)
	expectStatus(t, e, http.MethodGet, "/api/interactions/list", "", 200, "success")
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/interactions/sync", "")); m["message"] != "互动记录同步功能待实现" {
		t.Fatalf("interaction sync resp = %v", m)
	}
}

// ===========================================================================
// Scheduler routes
// ===========================================================================

func TestMSSchedulerCRUD(t *testing.T) {
	e := msMiscAPI(t)
	const taskID = "mstask1"
	const subID = "mssub1"

	// list (all main tasks)
	lm := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/scheduler/tasks", ""))
	if lm["status"] != "success" {
		t.Fatalf("list tasks resp = %v", lm)
	}

	// unknown task_id -> 任务不存在
	um := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/scheduler/tasks?task_id=msnonexistent", ""))
	if um["message"] != "任务不存在" {
		t.Fatalf("unknown task resp = %v", um)
	}

	// create via nested config payload
	create := `{"task_id":"` + taskID + `","task_type":"main","config":{"name":"MS 测试任务","endpoint":"/fetch/status","method":"GET","schedule_type":"once"}}`
	w := doRaw(t, e, http.MethodPost, "/api/scheduler/tasks", create)
	cm := msRawMap(t, w)
	if cm["message"] != "成功创建任务" || cm["task_id"] != taskID {
		t.Fatalf("create task resp = %v (body=%s)", cm, w.Body.String())
	}

	// create with missing task_id -> error branch
	em := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/scheduler/tasks", `{"task_type":"main"}`))
	if em["status"] != "error" {
		t.Fatalf("create missing id resp = %v", em)
	}

	// create bad JSON -> 400
	doRaw(t, e, http.MethodPost, "/api/scheduler/tasks", "{oops")

	// fetch the specific task -> found
	fm := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/scheduler/tasks?task_id="+taskID, ""))
	if fm["message"] != "获取任务信息成功" || fm["total"].(float64) != 1 {
		t.Fatalf("get specific task resp = %v", fm)
	}

	// update
	wu := doRaw(t, e, http.MethodPut, "/api/scheduler/tasks/"+taskID, `{"config":{"name":"MS 更新任务"}}`)
	if um2 := msRawMap(t, wu); um2["message"] != "任务更新成功" {
		t.Fatalf("update task resp = %v (body=%s)", um2, wu.Body.String())
	}

	// enable
	we := doRaw(t, e, http.MethodPost, "/api/scheduler/tasks/"+taskID+"/enable", `{"enabled":true}`)
	if me := msRawMap(t, we); me["message"] != "任务状态已更新" || me["enabled"] != true {
		t.Fatalf("enable task resp = %v", me)
	}
	doRaw(t, e, http.MethodPost, "/api/scheduler/tasks/"+taskID+"/enable", "{oops") // 400

	// add sub-task
	ws := doRaw(t, e, http.MethodPost, "/api/scheduler/tasks/"+taskID+"/subtasks", `{"task_id":"`+subID+`","config":{"name":"MS 子任务","endpoint":"/fetch/status","method":"GET"}}`)
	msm := msRawMap(t, ws)
	if msm["message"] != "成功创建子任务" || msm["task_id"] != subID {
		t.Fatalf("add subtask resp = %v (body=%s)", msm, ws.Body.String())
	}

	// execution history for the task (may be empty but envelope ok)
	hm := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/scheduler/tasks/history?task_id="+taskID, ""))
	if hm["status"] != "success" {
		t.Fatalf("task history resp = %v", hm)
	}

	// scheduler status
	expectStatus(t, e, http.MethodGet, "/api/scheduler/status", "", 200, "success")

	// delete sub-task
	wd := doRaw(t, e, http.MethodDelete, "/api/scheduler/tasks/"+taskID+"/subtasks/"+subID, "")
	if dm := msRawMap(t, wd); dm["message"] != "子任务删除成功" {
		t.Fatalf("delete subtask resp = %v", dm)
	}

	// delete the main task
	wtd := doRaw(t, e, http.MethodDelete, "/api/scheduler/tasks/"+taskID, "")
	if dm := msRawMap(t, wtd); dm["message"] != "任务删除成功" {
		t.Fatalf("delete task resp = %v", dm)
	}

	// delete unknown -> error branch
	if dm := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/scheduler/tasks/msnonexistent", "")); dm["status"] != "error" {
		t.Fatalf("delete unknown resp = %v", dm)
	}
}

func TestMSSchedulerRunAndAsync(t *testing.T) {
	e := msMiscAPI(t)

	// run unknown -> 运行任务失败
	rm := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/scheduler/tasks/msnope/execute", ""))
	if rm["status"] != "error" {
		t.Fatalf("run unknown resp = %v", rm)
	}

	// async task status for a missing id -> not_found
	am := expectStatus(t, e, http.MethodGet, "/api/task/msmissing/status", "", 200, "success")
	if am.dataMap(t)["status"] != "not_found" {
		t.Fatalf("async status data = %v", am.dataMap(t))
	}
}

// ===========================================================================
// Fetch routes
// ===========================================================================

func TestMSFetchNotConfigured(t *testing.T) {
	msSetCreds(t, "", "", "")
	e := msMiscAPI(t)

	// realtime sync returns error envelope without touching the network.
	rm := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/fetch/bili-history-realtime", ""))
	if rm["status"] != "error" || rm["message"] != "SESSDATA not configured" {
		t.Fatalf("realtime not-configured resp = %v", rm)
	}

	fm := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/fetch/bili-history", ""))
	if fm["status"] != "error" {
		t.Fatalf("full not-configured resp = %v", fm)
	}

	// async start -> 500 because FetchHistory errors before spawning.
	start := expectStatus(t, e, http.MethodPost, "/api/fetch/start", "", 500, "error")
	expectMessageContains(t, start, "启动历史记录获取失败")
}

func TestMSFetchSyncSuccess(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	msWithHistoryStub(t, msEmptyHistoryHandler, func() {
		e := msMiscAPI(t)
		rm := msRawMap(t, doRaw(t, e, http.MethodGet, "/api/fetch/bili-history-realtime", ""))
		if rm["status"] != "success" {
			t.Fatalf("realtime success resp = %v", rm)
		}
		// also cover the POST variant of the full fetch
		fm := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/fetch/bili-history", ""))
		if fm["status"] != "success" {
			t.Fatalf("full success resp = %v", fm)
		}
	})
}

func TestMSFetchStatus(t *testing.T) {
	e := msMiscAPI(t)
	// unknown task -> not_found
	nm := expectStatus(t, e, http.MethodGet, "/api/fetch/status?task_id=msnope", "", 200, "success")
	if nm.dataMap(t)["status"] != "not_found" {
		t.Fatalf("fetch status data = %v", nm.dataMap(t))
	}
	// overall status
	om := expectStatus(t, e, http.MethodGet, "/api/fetch/status", "", 200, "success")
	if om.dataMap(t) == nil {
		t.Fatalf("overall fetch status missing data")
	}
}

func TestMSFetchInvalidVideos(t *testing.T) {
	e := msMiscAPI(t)
	resp := expectStatus(t, e, http.MethodGet, "/api/fetch/invalid-videos?page=1&size=10", "", 200, "success")
	dm := resp.dataMap(t)
	if dm == nil {
		t.Fatalf("invalid videos missing data")
	}
	// out-of-range size falls back to 20 internally; assert envelope shape.
	if _, ok := dm["videos"]; !ok {
		t.Fatalf("invalid videos data = %v", dm)
	}
}

// 回归 services/history.go 的缺陷修复：调试日志曾在 err 检查之前解引用
// data.List / data.Cursor.Max，GetHistory 失败（data==nil）时 nil 指针 panic
// 会顺着同步抓取处理器冲出 ServeHTTP。现在错误分支退避重试，下一张收敛空页
// 让循环正常结束。
func TestMSFetchSyncAPIErrorRetries(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	var hits int32
	failingThenEmpty := func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":-101,"message":"账号未登录","data":null}`))
			return
		}
		msEmptyHistoryHandler(w, r)
	}
	msWithHistoryStub(t, failingThenEmpty, func() {
		e := msMiscAPI(t)
		w := doRaw(t, e, http.MethodGet, "/api/fetch/bili-history-realtime", "")
		if n := atomic.LoadInt32(&hits); n != 2 {
			t.Fatalf("history hits = %d, want 2 (one failure then empty page)", n)
		}
		// 同步抓取把结果字段直接铺在信封顶层，没有 data 字段。
		m := msRawMap(t, w)
		if m["status"] != "success" {
			t.Fatalf("status = %v, want success (body %s)", m["status"], w.Body.String())
		}
		if m["total_records"] != float64(0) || m["total_pages"] != float64(2) {
			t.Fatalf("result = %v, want 0 records over 2 pages", m)
		}
	})
}

// ===========================================================================
// Delete routes (Bilibili + local soft-delete)
// ===========================================================================

func TestMSDeleteStubs(t *testing.T) {
	e := msMiscAPI(t)
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/delete/history", "")); m["message"] != "删除历史记录功能待实现" {
		t.Fatalf("delete history resp = %v", m)
	}
	if m := msRawMap(t, doRaw(t, e, http.MethodPost, "/api/bilibili/history/delete", "")); m["message"] != "删除B站历史记录功能待实现" {
		t.Fatalf("bili delete resp = %v", m)
	}
	expectStatus(t, e, http.MethodGet, "/api/bilibili/history/status", "", 200, "success")
}

func TestMSBatchDeleteLocal(t *testing.T) {
	_, bvidB := msSeed2009(t)
	e := msMiscAPI(t)

	// format 1: {"bvids": [...]}
	w1 := doRaw(t, e, http.MethodDelete, "/api/delete/batch-delete", `{"bvids":["`+bvidB+`","BVMS000000999"]}`)
	m1 := msRawMap(t, w1)
	data1, _ := m1["data"].(map[string]interface{})
	if data1["deleted_count"].(float64) != 2 {
		t.Fatalf("format1 deleted_count = %v, want 2 (resp=%v)", data1["deleted_count"], m1)
	}

	// format 2: array of {bvid, view_at}
	w2 := doRaw(t, e, http.MethodDelete, "/api/delete/batch-delete", `[{"bvid":"`+bvidB+`","view_at":123}]`)
	if m2 := msRawMap(t, w2); m2["status"] != "success" {
		t.Fatalf("format2 resp = %v", m2)
	}

	// empty -> 400 (an empty JSON array parses as format-2 with no bvids)
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/delete/batch-delete", `[]`)); m["message"] != "bvids 不能为空" {
		t.Fatalf("empty bvids resp = %v", m)
	}

	// wrong shape -> 400 参数错误
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/delete/batch-delete", `{"foo":123}`)); m["status"] != "error" {
		t.Fatalf("bad shape resp = %v", m)
	}
}

func TestMSDeleteSingleGuards(t *testing.T) {
	e := msMiscAPI(t)

	// empty kid
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single", "")); m["message"] != "kid 不能为空" {
		t.Fatalf("empty kid resp = %v", m)
	}

	// no SESSDATA
	msSetCreds(t, "", "", "")
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single?kid=archive_1", "")); m["message"] != "SESSDATA 未配置" {
		t.Fatalf("no sessdata resp = %v", m)
	}

	// no BiliJct
	msSetCreds(t, "sess", "", "")
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single?kid=archive_1", "")); m["status"] != "error" {
		t.Fatalf("no jct resp = %v", m)
	}

	// invalid kid format (no separator)
	msSetCreds(t, "sess", "jct", "dede")
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single?kid=nounderscore", "")); m["message"] != "无效的 kid 格式" {
		t.Fatalf("bad kid format resp = %v", m)
	}

	// oid not found -> error envelope (uses a fresh request per call)
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single?kid=archive_999999", "")); m["status"] != "error" {
		t.Fatalf("oid not found resp = %v", m)
	}
}

func TestMSDeleteSingleSuccess(t *testing.T) {
	bvidA, _ := msSeed2009(t)
	msSetCreds(t, "sess", "jct", "dede")
	e := msMiscAPI(t)
	msWithDelStub(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"OK"}`))
	}, func() {
		m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single?kid=archive_2009011", ""))
		if m["message"] != "已删除B站历史记录" {
			t.Fatalf("delete single success resp = %v", m)
		}
	})
	// local record should now be soft-deleted (status=1)
	n := countRows(t, "SELECT COUNT(*) FROM bilibili_history_2009 WHERE bvid = ? AND status = 1", bvidA)
	if n != 1 {
		t.Fatalf("expected bvid %s to be soft-deleted, got %d rows", bvidA, n)
	}
}

func TestMSDeleteSingleApiErrorAndHTTPError(t *testing.T) {
	msSeed2009(t)
	msSetCreds(t, "sess", "jct", "dede")
	e := msMiscAPI(t)

	// ApiError branch: code=-101 -> 删除失败: code=-101, 账号未登录
	msWithDelStub(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":-101,"message":"账号未登录"}`))
	}, func() {
		m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single?kid=archive_2009011", ""))
		if m["status"] != "error" || m["message"] != "删除失败: code=-101, 账号未登录" {
			t.Fatalf("single -101 resp = %v", m)
		}
	})

	// HTTP error branch: non-200 -> not an ApiError -> detailed message
	msWithDelStub(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}, func() {
		m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/single?kid=archive_2009011", ""))
		msg, _ := m["message"].(string)
		if m["status"] != "error" || !bytes.Contains([]byte(msg), []byte("删除B站历史记录失败")) {
			t.Fatalf("single http-error resp = %v", m)
		}
	})
}

func TestMSDeleteBatchBili(t *testing.T) {
	_, bvidB := msSeed2009(t)
	e := msMiscAPI(t)

	// empty bvids -> 400
	msSetCreds(t, "sess", "jct", "dede")
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/batch", `{"bvids":[]}`)); m["message"] != "bvids 不能为空" {
		t.Fatalf("batch empty resp = %v", m)
	}

	// no SESSDATA -> 400
	msSetCreds(t, "", "", "")
	if m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/batch", `{"bvids":["x"]}`)); m["message"] != "SESSDATA 未配置" {
		t.Fatalf("batch no sessdata resp = %v", m)
	}

	msSetCreds(t, "sess", "jct", "dede")

	// success
	msWithDelStub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"message":"OK"}`))
	}, func() {
		m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/batch", `{"bvids":["`+bvidB+`"]}`))
		if m["message"] != "已删除 1 条B站历史记录" {
			t.Fatalf("batch success resp = %v", m)
		}
	})

	// ApiError -352
	msWithDelStub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":-352,"message":"风控"}`))
	}, func() {
		m := msRawMap(t, doRaw(t, e, http.MethodDelete, "/api/bilibili/history/batch", `{"bvids":["`+bvidB+`"]}`))
		if m["status"] != "error" || m["message"] != "删除失败: code=-352, 风控" {
			t.Fatalf("batch -352 resp = %v", m)
		}
	})

	// HTTP 500 -> non-ApiError -> 500 envelope
	msWithDelStub(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("gw error"))
	}, func() {
		w := doRaw(t, e, http.MethodDelete, "/api/bilibili/history/batch", `{"bvids":["`+bvidB+`"]}`)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("batch http-error code = %d, want 500 (body=%s)", w.Code, w.Body.String())
		}
	})
}

func TestMSDeleteBatchBadJSON(t *testing.T) {
	msSetCreds(t, "sess", "jct", "dede")
	e := msMiscAPI(t)
	w := doRaw(t, e, http.MethodDelete, "/api/bilibili/history/batch", "not-json")
	if m := msRawMap(t, w); m["status"] != "error" {
		t.Fatalf("batch bad json resp = %v", m)
	}
}

// ===========================================================================
// Export to Excel (async) + download
// ===========================================================================

func TestMSExportExcelAndDownload(t *testing.T) {
	msSeed2009(t)
	want := countRows(t, "SELECT COUNT(*) FROM bilibili_history_2009")

	// keep the xlsx inside the test's own temp dir
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)

	e := msMiscAPI(t)
	resp := expectStatus(t, e, http.MethodPost, "/api/export/excel", `{"years":[2009]}`, 200, "success")
	dm := resp.dataMap(t)
	taskID, _ := dm["task_id"].(string)
	if taskID == "" {
		t.Fatalf("export returned no task_id: %v", dm)
	}

	// poll the async status until the background export finishes
	var resultObj map[string]interface{}
	msWaitFor(t, 5*time.Second, func() bool {
		st := expectStatus(t, e, http.MethodGet, "/api/task/"+taskID+"/status", "", 200, "success")
		data := st.dataMap(t)
		if data["status"] == "completed" {
			if raw, ok := data["result"].(string); ok && raw != "" {
				_ = json.Unmarshal([]byte(raw), &resultObj)
			}
			return true
		}
		return data["status"] == "failed"
	}, "excel export to complete")

	if resultObj == nil {
		t.Fatalf("export result missing (task %s)", taskID)
	}
	if got, _ := resultObj["total_rows"].(float64); int(got) != want {
		t.Fatalf("total_rows = %v, want %d", resultObj["total_rows"], want)
	}

	// the xlsx must have landed under the test temp dir
	if !msExportFileExists(dir, taskID) {
		t.Fatalf("expected xlsx under %s for task %s", dir, taskID)
	}

	// download served from the stored file path
	w := doRaw(t, e, http.MethodGet, "/api/export/download/"+taskID, "")
	if w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Fatalf("download code = %d, len = %d", w.Code, w.Body.Len())
	}

	// cleanup the process-wide registry entry
	asyncExportFiles.Delete(taskID)

	// download unknown -> 404
	unk := doRaw(t, e, http.MethodGet, "/api/export/download/msnever", "")
	if unk.Code != http.StatusNotFound {
		t.Fatalf("unknown download code = %d, want 404", unk.Code)
	}
}

// msExportFileExists reports whether doExportToExcel wrote the xlsx for taskID
// into the given directory (which was exported as $TMPDIR for the test).
func msExportFileExists(dir, taskID string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, "bilibili_history_"+taskID+".xlsx"))
	return len(matches) > 0
}
