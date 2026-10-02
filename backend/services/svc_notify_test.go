package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"bilibili-history-go/config"
)

// svcWithNotifyConfig 临时替换全局配置（LoadConfig 的 sync.Once 已消耗，
// SaveConfig 会直接替换缓存指针），测试结束后恢复原指针。
func svcWithNotifyConfig(t *testing.T, mutate func(*config.Config)) {
	t.Helper()
	orig := config.GetConfig()
	next := *orig
	mutate(&next)
	if err := config.SaveConfig(&next); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	t.Cleanup(func() {
		restore := *orig
		if err := config.SaveConfig(&restore); err != nil {
			t.Errorf("restore config: %v", err)
		}
	})
}

func TestSVCGetNotifyURLsDisabled(t *testing.T) {
	// 全局配置 notify.enabled=false
	if urls, err := getNotifyURLs(); err == nil || urls != nil {
		t.Errorf("disabled notify: urls=%v err=%v", urls, err)
	}
	if err := SendNotification("t", "m"); err == nil || !strings.Contains(err.Error(), "通知未配置") {
		t.Errorf("SendNotification err = %v", err)
	}
	if err := SendTestNotification(); err == nil {
		t.Errorf("SendTestNotification should fail when unconfigured")
	}
	if err := SendSessdataExpiredNotification("svc-user"); err == nil {
		t.Errorf("expired notification should fail when unconfigured")
	}
	if err := SendSessdataExpiredNotification(""); err == nil {
		t.Errorf("expired notification (no username) should fail when unconfigured")
	}
	ResetNotifyRouter() // no-op 兼容函数，调用一遍
}

func TestSVCGetNotifyURLsAllInvalid(t *testing.T) {
	svcWithNotifyConfig(t, func(c *config.Config) {
		c.Notify.Enabled = true
		c.Notify.URLs = []string{"://bad-url"}
	})
	if _, err := getNotifyURLs(); err == nil || !strings.Contains(err.Error(), "无有效的通知 URL") {
		t.Errorf("err = %v, want 无有效的通知 URL", err)
	}
}

func TestSVCSendNotificationViaHTTP(t *testing.T) {
	// apprise json:// 目标只会 POST 到本机 httptest 服务，零外网
	var mu sync.Mutex
	type payload struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	var got []payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var p payload
		if json.Unmarshal(b, &p) == nil {
			mu.Lock()
			got = append(got, p)
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	appriseURL := "json://" + strings.TrimPrefix(srv.URL, "http://") + "/notify"
	svcWithNotifyConfig(t, func(c *config.Config) {
		c.Notify.Enabled = true
		c.Notify.URLs = []string{"://bad-entry", appriseURL}
	})

	// 空 message：body 只有标题
	if err := SendNotification("svc-title", ""); err != nil {
		t.Fatalf("SendNotification: %v", err)
	}
	// 非空 message：title\nmessage 拼接；无效 URL 被过滤但有效 URL 仍发送
	if err := SendNotificationWithParams("svc-title2", "svc-body", map[string]string{"ignored": "yes"}); err != nil {
		t.Fatalf("SendNotificationWithParams: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 2 {
		t.Fatalf("webhook payloads = %d, want 2 (%+v)", len(got), got)
	}
	if got[0].Title != "svc-title" || got[0].Message != "svc-title" {
		t.Errorf("empty-message body = %+v, want title only", got[0])
	}
	if got[1].Message != "svc-title2\nsvc-body" {
		t.Errorf("combined body = %q", got[1].Message)
	}
}

func TestSVCGatherDailyReportData(t *testing.T) {
	data := gatherDailyReportData()
	// TestMain 在当前年种入今日 2 条（100s progress + progress=-1→duration 3600 = 3700s）
	if data["today_records"] != 2 {
		t.Fatalf("today_records = %v, want 2", data["today_records"])
	}
	if data["total_watching_time"] != "1小时1分钟" {
		t.Errorf("watching time = %v, want 1小时1分钟", data["total_watching_time"])
	}
	if data["report_date"] != time.Now().Format("2006-01-02") {
		t.Errorf("report_date = %v", data["report_date"])
	}
	if top, ok := data["top_author"].(string); !ok || !strings.Contains(top, "svc日报UP") {
		t.Errorf("top_author = %v", data["top_author"])
	}
	if top, ok := data["top_category"].(string); !ok || !strings.Contains(top, "科技") {
		t.Errorf("top_category = %v", data["top_category"])
	}
	if tag, ok := data["top_tag"].(string); !ok || tag != "科技" {
		t.Errorf("top_tag = %v", data["top_tag"])
	}
	if peak, ok := data["peak_hour"].(string); !ok || !strings.Contains(peak, ":00-") {
		t.Errorf("peak_hour format = %v", data["peak_hour"])
	}
	// 今日有记录时不应出现 year_total / last_view_*
	if _, exists := data["year_total"]; exists {
		t.Errorf("year_total should be absent when today has records")
	}
	if _, exists := data["last_view_time"]; exists {
		t.Errorf("last_view_time should be absent when today has records")
	}
}

func TestSVCFormatTopNAndQueryTopN(t *testing.T) {
	if got := formatTopN(nil); got != "" {
		t.Errorf("formatTopN(nil) = %q", got)
	}
	if got := formatTopN([]string{"a", "b"}); got != "1.a 2.b " {
		t.Errorf("formatTopN = %q", got)
	}

	conn := svcConn()
	table := "bilibili_history_2031"
	tops := queryTopN(conn, table, 0, time.Now().AddDate(0, 0, 365*10).Unix(), "author_name", "author_name")
	if len(tops) != 3 {
		t.Errorf("top authors = %v, want 3 distinct", tops)
	}
	// 不存在的列 → Query 失败返回 nil
	if got := queryTopN(conn, table, 0, 1, "no_such_col", "no_such_col"); got != nil {
		t.Errorf("bad column should give nil, got %v", got)
	}
}

func TestSVCSendNotificationDeliveryFailure(t *testing.T) {
	// apprise AddAll 阶段：URL 语法合法但协议不受支持 → Send 报"通知发送失败"
	svcWithNotifyConfig(t, func(c *config.Config) {
		c.Notify.Enabled = true
		c.Notify.URLs = []string{"svcunknown://example.invalid"}
	})
	if err := SendNotification("t", "m"); err == nil || !strings.Contains(err.Error(), "通知发送失败") {
		t.Errorf("err = %v, want 通知发送失败 (unsupported scheme)", err)
	}

	// 发送阶段：json:// 指向本机回环 1 号端口（必然关闭），仅本地连接被拒绝，零外网
	svcWithNotifyConfig(t, func(c *config.Config) {
		c.Notify.Enabled = true
		c.Notify.URLs = []string{"json://127.0.0.1:1/notify"}
	})
	if err := SendTestNotification(); err == nil || !strings.Contains(err.Error(), "通知发送失败") {
		t.Errorf("unreachable webhook err = %v", err)
	}
	if err := SendSessdataExpiredNotification("svc-user"); err == nil {
		t.Errorf("expired notification should fail on unreachable webhook")
	}
}

func TestSVCGatherDailyReportNoTodayData(t *testing.T) {
	// 当前年沙盒行清理后再恢复，保证 gatherDailyReportData 的 todayCount==0 分支可测。
	table := fmt.Sprintf("bilibili_history_%d", svcCurrentYear)
	conn := svcConn()
	if _, err := conn.Exec("DELETE FROM " + table + " WHERE bvid LIKE 'BVsvcNotify%'"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	defer func() {
		now := time.Now()
		rec1 := svcHistoryRecord("BVsvcNotify0001", "svc 日报视频一", "archive", "科技", "科技", 7700020, "svc日报UP", svcAt(now.Year(), int(now.Month()), now.Day(), 9, 0), 200, 100)
		rec2 := svcHistoryRecord("BVsvcNotify0002", "svc 日报视频二", "archive", "科技", "科技", 7700020, "svc日报UP", svcAt(now.Year(), int(now.Month()), now.Day(), 10, 30), 3600, -1)
		rec3 := svcHistoryRecord("BVsvcNotify0003", "svc 往日视频", "archive", "游戏", "游戏", 7700021, " svc往日UP", svcAt(svcCurrentYear, 1, 2, 7, 0), 30, 30)
		if err := svcSeedHistoryRows(svcCurrentYear, rec1, rec2, rec3); err != nil {
			t.Errorf("reseed notify sandbox: %v", err)
		}
	}()

	// 阶段 A：整表无 svc 行且 MAX(view_at) 为空（若当前年表恰好还有非 svc 行则 year_total>=0，仅验证 today 为 0）
	if _, err := conn.Exec("DELETE FROM " + table); err != nil {
		t.Fatalf("delete all: %v", err)
	}
	data := gatherDailyReportData()
	if data["today_records"] != 0 {
		t.Fatalf("today_records = %v, want 0", data["today_records"])
	}
	if data["year_total"] != 0 {
		t.Errorf("year_total = %v, want 0 for empty table", data["year_total"])
	}
	if _, ok := data["last_view_time"]; ok {
		t.Errorf("last_view_time should be absent when table empty: %v", data["last_view_time"])
	}
	if data["total_watching_time"] != "0分钟" {
		t.Errorf("watching time = %v, want 0分钟", data["total_watching_time"])
	}
	// 走 SendDailyReport 自动采集 + "今日暂无观看记录" 消息路径（notify 关闭 → 报错但分支已覆盖）
	if err := SendDailyReport(map[string]interface{}{}); err == nil {
		t.Errorf("want error when notify disabled")
	}

	// 阶段 B：只留一条"今天之前最后一秒"的记录 → last_view_ago == 今天。
	// gatherDailyReportData 用 int(now.Sub(last).Hours()/24)==0 判定"今天"，
	// 所以这一行必须同时满足：落在 today 窗口 [todayStart, +86400) 之外，
	// 且距今不足 24 小时。todayStart-1 是唯一在任意钟点都成立的取值
	//（曾用 todayStart-3600，23:00 之后跑就会变成"1天前"）。
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Unix()
	rec := svcHistoryRecord("BVsvcNotifyLast1", "svc 昨夜视频", "archive", "科技", "科技", 7700022, "svc昨夜UP", todayStart-1, 60, 60)
	if err := svcSeedHistoryRows(svcCurrentYear, rec); err != nil {
		t.Fatalf("seed last-view: %v", err)
	}
	data = gatherDailyReportData()
	if data["today_records"] != 0 {
		t.Fatalf("phase B today_records = %v, want 0", data["today_records"])
	}
	if data["year_total"] != 1 {
		t.Errorf("phase B year_total = %v, want 1", data["year_total"])
	}
	if ago, ok := data["last_view_ago"].(string); !ok || ago != "今天" {
		t.Errorf("phase B last_view_ago = %v, want 今天", data["last_view_ago"])
	}
	if lt, ok := data["last_view_time"].(string); !ok || lt == "" {
		t.Errorf("phase B last_view_time = %v", data["last_view_time"])
	}

	// 阶段 C：7 天前的记录 → "N天前"
	if _, err := conn.Exec("DELETE FROM " + table); err != nil {
		t.Fatalf("delete phase B rows: %v", err)
	}
	rec2 := svcHistoryRecord("BVsvcNotifyLast2", "svc 一周前视频", "archive", "科技", "科技", 7700023, "svc旧番UP", time.Now().AddDate(0, 0, -7).Unix(), 60, 60)
	if err := svcSeedHistoryRows(svcCurrentYear, rec2); err != nil {
		t.Fatalf("seed 7d row: %v", err)
	}
	data = gatherDailyReportData()
	if ago, ok := data["last_view_ago"].(string); !ok || !strings.HasSuffix(ago, "天前") {
		t.Errorf("phase C last_view_ago = %v, want *天前", data["last_view_ago"])
	}

	// 清掉本测试残留，defer 再统一恢复 TestMain 的 3 条种子行
	if _, err := conn.Exec("DELETE FROM " + table + " WHERE bvid LIKE 'BVsvcNotify%'"); err != nil {
		t.Errorf("cleanup: %v", err)
	}
}

func TestSVCSendDailyReportDefaultDateBranch(t *testing.T) {
	// stats 非空但缺 report_date → 走 time.Now() 默认日期分支；notify 关闭 → 返回错误
	stats := map[string]interface{}{"today_records": 1, "total_watching_time": "1分钟"}
	err := SendDailyReport(stats)
	if err == nil || !strings.Contains(err.Error(), "通知未配置") {
		t.Errorf("err = %v, want 通知未配置", err)
	}
}

func TestSVCGatherDailyReportMissingTable(t *testing.T) {
	// 当前年表整体缺失 → gatherDailyReportData 走 TableExists=false 早退分支。
	// 测试结束通过 svcSeedHistoryRows(内部 EnsureTableForYear) 重建并恢复种子行。
	conn := svcConn()
	table := fmt.Sprintf("bilibili_history_%d", svcCurrentYear)
	if _, err := conn.Exec("DROP TABLE IF EXISTS " + table); err != nil {
		t.Fatalf("drop table: %v", err)
	}
	defer func() {
		now := time.Now()
		rec1 := svcHistoryRecord("BVsvcNotify0001", "svc 日报视频一", "archive", "科技", "科技", 7700020, "svc日报UP", svcAt(now.Year(), int(now.Month()), now.Day(), 9, 0), 200, 100)
		rec2 := svcHistoryRecord("BVsvcNotify0002", "svc 日报视频二", "archive", "科技", "科技", 7700020, "svc日报UP", svcAt(now.Year(), int(now.Month()), now.Day(), 10, 30), 3600, -1)
		rec3 := svcHistoryRecord("BVsvcNotify0003", "svc 往日视频", "archive", "游戏", "游戏", 7700021, " svc往日UP", svcAt(svcCurrentYear, 1, 2, 7, 0), 30, 30)
		if err := svcSeedHistoryRows(svcCurrentYear, rec1, rec2, rec3); err != nil {
			t.Errorf("restore current-year table: %v", err)
		}
	}()

	data := gatherDailyReportData()
	if data["today_records"] != 0 {
		t.Errorf("today_records = %v, want 0", data["today_records"])
	}
	if data["total_watching_time"] != "0分钟" {
		t.Errorf("watching time = %v, want 0分钟", data["total_watching_time"])
	}
	if data["report_date"] != time.Now().Format("2006-01-02") {
		t.Errorf("report_date = %v", data["report_date"])
	}
	if _, ok := data["year_total"]; ok {
		t.Errorf("year_total should be absent on missing-table early return")
	}
}

func TestSVCSendDailyReportFormatting(t *testing.T) {
	// 显式 stats（len>1）：格式化全部字段后走 SendNotification（notify 关闭 → 返回错误）
	stats := map[string]interface{}{
		"report_date":         "2031-06-02",
		"today_records":       5,
		"total_watching_time": "1小时1分钟",
		"top_author":          "1.svc日报UP ",
		"top_category":        "1.科技 ",
		"top_tag":             "科技",
		"peak_hour":           "9:00-9:59",
	}
	if err := SendDailyReport(stats); err == nil {
		t.Errorf("want error when notify disabled (documents guard behavior)")
	}

	// 空 stats 触发 gatherDailyReportData 自动采集路径
	if err := SendDailyReport(map[string]interface{}{}); err == nil {
		t.Errorf("want error when notify disabled")
	}

	// 今日无数据补充字段路径
	stats2 := map[string]interface{}{
		"report_date":    "2031-01-01",
		"year_total":     3,
		"last_view_time": "2030-12-31 10:00",
		"last_view_ago":  "1天前",
	}
	if err := SendDailyReport(stats2); err == nil {
		t.Errorf("want error when notify disabled")
	}

	// 全部消息在 notify 启用后应能送达 webhook
	var mu sync.Mutex
	var titles []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var p struct {
			Title string `json:"title"`
		}
		if json.Unmarshal(b, &p) == nil {
			mu.Lock()
			titles = append(titles, p.Title)
			mu.Unlock()
		}
	}))
	defer srv.Close()
	appriseURL := "json://" + strings.TrimPrefix(srv.URL, "http://") + "/hook"
	svcWithNotifyConfig(t, func(c *config.Config) {
		c.Notify.Enabled = true
		c.Notify.URLs = []string{appriseURL}
	})
	if err := SendDailyReport(stats); err != nil {
		t.Fatalf("SendDailyReport with webhook: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(titles) != 1 || !strings.Contains(titles[0], "Bilibili日报 2031-06-02") {
		t.Errorf("titles = %v", titles)
	}
}
