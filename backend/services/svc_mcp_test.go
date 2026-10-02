package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func svcMcpCfg() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 8899},
		Mcp:    config.McpConfig{Path: "/mcp", AuthEnabled: true, Token: "svc-token-123", MaxPageSize: 100},
	}
}

func TestSVCGenerateToken(t *testing.T) {
	a, b := GenerateToken(), GenerateToken()
	if len(a) != 32 || len(b) != 32 {
		t.Errorf("token lengths = %d/%d, want 32", len(a), len(b))
	}
	if a == b {
		t.Errorf("tokens must differ")
	}
	for _, c := range a {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Fatalf("non-hex char %q", c)
		}
	}
}

func TestSVCGetMCPServerURL(t *testing.T) {
	cases := []struct {
		host, path string
		port       int
		want       string
	}{
		{"127.0.0.1", "/mcp", 9000, "http://127.0.0.1:9000/mcp/"},
		{"0.0.0.0", "/mcp", 8899, "http://127.0.0.1:8899/mcp/"},
		{"::", "", 0, "http://127.0.0.1:8899/mcp/"},
		{"", "/custom", 1234, "http://127.0.0.1:1234/custom/"},
		{"192.168.1.5", "nopath", 80, "http://192.168.1.5:80/nopath/"},
		{"localhost", "/mcp/", 81, "http://localhost:81/mcp//"},
	}
	for _, c := range cases {
		cfg := &config.Config{}
		cfg.Server.Host = c.host
		cfg.Server.Port = c.port
		cfg.Mcp.Path = c.path
		if got := GetMCPServerURL(cfg); got != c.want {
			t.Errorf("GetMCPServerURL(%q,%q,%d) = %q, want %q", c.host, c.path, c.port, got, c.want)
		}
	}
}

func TestSVCGetMCPSkillContent(t *testing.T) {
	cfg := svcMcpCfg()
	withAuth := GetMCPSkillContent(cfg)
	if !strings.Contains(withAuth, "Authorization: Bearer svc-token-123") {
		t.Errorf("auth line missing token: %q", withAuth)
	}
	cfg.Mcp.AuthEnabled = false
	noAuth := GetMCPSkillContent(cfg)
	if !strings.Contains(noAuth, "Authorization: not required") {
		t.Errorf("auth line = %q", noAuth)
	}
	for _, s := range []string{"bili://project/overview", "bili://project/data-status", "bili://project/tool-guide", GetMCPServerURL(cfg)} {
		if !strings.Contains(noAuth, s) {
			t.Errorf("skill content missing %q", s)
		}
	}
}

func TestSVCCredentialLeakCheckInSkillContent(t *testing.T) {
	// 该函数面向 AI 客户端输出 Token——这是设计如此（用户自己的 MCP token），
	// 但确认它不会泄漏 SESSDATA / bili_jct 等 B 站凭据。
	cfg := svcMcpCfg()
	cfg.SESSDATA = "super-secret-sessdata"
	cfg.BiliJct = "super-secret-jct"
	out := GetMCPSkillContent(cfg)
	if strings.Contains(out, "super-secret-sessdata") || strings.Contains(out, "super-secret-jct") {
		t.Errorf("bili credentials leaked into skill content")
	}
}

func TestSVCWrapWithAuth(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("inner-ok"))
	})

	call := func(cfg *config.Config, header string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/mcp", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		WrapWithAuth(cfg, inner).ServeHTTP(rec, req)
		return rec
	}

	// 未启用鉴权：直接放行
	cfg := svcMcpCfg()
	cfg.Mcp.AuthEnabled = false
	if rec := call(cfg, ""); rec.Code != 200 || rec.Body.String() != "inner-ok" {
		t.Errorf("disabled auth: code=%d body=%q", rec.Code, rec.Body.String())
	}

	// 启用鉴权
	cfg = svcMcpCfg()
	if rec := call(cfg, ""); rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "missing authorization") {
		t.Errorf("missing header: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec := call(cfg, "Bearer wrong"); rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "invalid token") {
		t.Errorf("wrong token: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec := call(cfg, "Bearer svc-token-123"); rec.Code != 200 || rec.Body.String() != "inner-ok" {
		t.Errorf("good token: code=%d body=%q", rec.Code, rec.Body.String())
	}
	// 现状记录：无 "Bearer " 前缀时 TrimPrefix 为空操作，裸 token 也可通过
	if rec := call(cfg, "svc-token-123"); rec.Code != 200 {
		t.Errorf("bare token unexpectedly rejected: code=%d", rec.Code)
	}
	// "Basic xxx" 会把整串当 token 比较，必然失败
	if rec := call(cfg, "Basic svc-token-123"); rec.Code != http.StatusUnauthorized {
		t.Errorf("basic auth should fail token compare, code=%d", rec.Code)
	}
}

func TestSVCSetupMCPServerAndHandler(t *testing.T) {
	cfg := svcMcpCfg()
	SetupMCPServer(cfg)
	if h := GetMCPHandler(); h == nil {
		t.Fatalf("GetMCPHandler returned nil after SetupMCPServer")
	}
}

func svcCallTool(t *testing.T, s *server.MCPServer, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	tool := s.GetTool(name)
	if tool == nil || tool.Handler == nil {
		t.Fatalf("tool %s not registered", name)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	res, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler %s error: %v", name, err)
	}
	return res
}

func svcToolText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatalf("empty content, IsError=%v", res.IsError)
	}
	text, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type %T", res.Content[0])
	}
	return text.Text
}

func TestSVCMCPTools(t *testing.T) {
	s := server.NewMCPServer("svc-test", "1.0.0", server.WithToolCapabilities(false))
	registerMCPTools(s, svcMcpCfg())

	// search_history：命中 2031 数据集
	res := svcCallTool(t, s, "search_history", map[string]any{"keyword": "svc 主标题一", "year": float64(svcYearMain)})
	if res.IsError {
		t.Fatalf("search_history error: %s", svcToolText(t, res))
	}
	if !strings.Contains(svcToolText(t, res), "BVsvcMain0001") {
		t.Errorf("search result missing bvid: %.200s", svcToolText(t, res))
	}

	// page_size 超过 MaxPageSize=100 时钳制，page<1 归 1：仍应成功返回
	res = svcCallTool(t, s, "search_history", map[string]any{"keyword": "svc", "page": float64(0), "page_size": float64(1000)})
	if res.IsError {
		t.Errorf("clamped paging failed: %s", svcToolText(t, res))
	}

	// get_history：精确日期 + 分类
	res = svcCallTool(t, s, "get_history", map[string]any{
		"year": float64(svcYearMain), "month": float64(6), "day": float64(2),
	})
	if res.IsError || !strings.Contains(svcToolText(t, res), "BVsvcMain0002") {
		t.Errorf("get_history date range: %.200s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_history", map[string]any{"year": float64(svcYearMain), "month": float64(6)})
	if res.IsError {
		t.Errorf("get_history month range: %s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_history", map[string]any{"category": "科技", "page_size": float64(5)})
	if res.IsError {
		t.Errorf("get_history category: %s", svcToolText(t, res))
	}

	// get_daily_stats：单日、整月、无数据三种调用形态
	res = svcCallTool(t, s, "get_daily_stats", map[string]any{"year": float64(svcYearMain), "month": float64(6), "day": float64(2)})
	if res.IsError {
		t.Errorf("daily stats day: %s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_daily_stats", map[string]any{"year": float64(svcYearMain), "month": float64(6)})
	if res.IsError || !strings.Contains(svcToolText(t, res), "\"count\"") {
		t.Errorf("daily stats month: %.200s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_daily_stats", map[string]any{"year": float64(svcYearMain), "month": float64(12), "day": float64(31)})
	// Regression: on a day with no rows database.GetDailyStats used to Scan
	// SUM()'s NULL into an int and fail, which the handler surfaced as
	// "获取统计失败". The COALESCE fix landed first, so the tool now answers zeroes.
	if res.IsError {
		t.Errorf("daily stats empty day: %s", svcToolText(t, res))
	}
	if !strings.Contains(svcToolText(t, res), `"count"`) {
		t.Errorf("daily stats empty day = %.200s, want a zeroed count", svcToolText(t, res))
	}

	// get_yearly_analysis
	res = svcCallTool(t, s, "get_yearly_analysis", map[string]any{"year": float64(svcYearMain)})
	if res.IsError {
		t.Errorf("yearly analysis: %.300s", svcToolText(t, res))
	}

	// get_video_info：先种一条视频详情（bvid 命名空间 BVsvcMCP*）
	if err := database.UpsertVideoBaseInfo(svcVideoBaseInfo("BVsvcMCP0001", "svc MCP 视频", "svc分区")); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	res = svcCallTool(t, s, "get_video_info", map[string]any{"bvid": "BVsvcMCP0001"})
	if res.IsError || !strings.Contains(svcToolText(t, res), "svc MCP 视频") {
		t.Errorf("video info: %.200s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_video_info", map[string]any{"bvid": "BVsvcNotExist9999"})
	// 现状：GetVideoBaseInfoByBvid 未找到返回 (nil, nil)，工具把 "null" 文本原样返回
	if res.IsError || svcToolText(t, res) != "null" {
		t.Errorf("missing bvid handling: IsError=%v text=%.200s", res.IsError, svcToolText(t, res))
	}

	// get_categories
	res = svcCallTool(t, s, "get_categories", nil)
	if res.IsError {
		t.Errorf("categories: %s", svcToolText(t, res))
	}

	// get_overview：默认年份与指定年份
	res = svcCallTool(t, s, "get_overview", nil)
	if res.IsError {
		t.Errorf("overview default: %.300s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_overview", map[string]any{"year": float64(svcYearMain)})
	if res.IsError {
		t.Errorf("overview year: %.300s", svcToolText(t, res))
	}

	// 点赞 / 稍后再看 / 收藏夹（空库也应返回结构化结果，不是 error）
	res = svcCallTool(t, s, "get_liked_videos", map[string]any{"page": float64(1), "page_size": float64(20)})
	if res.IsError {
		t.Errorf("liked: %s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_watch_later", map[string]any{"sort": "duration", "order": "asc"})
	if res.IsError {
		t.Errorf("watch later: %s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_favorite_folders", nil)
	if res.IsError {
		t.Errorf("folders: %s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_favorite_contents", map[string]any{"media_id": float64(770000001)})
	if res.IsError {
		t.Errorf("contents: %s", svcToolText(t, res))
	}
	// media_id=0 明确报错
	res = svcCallTool(t, s, "get_favorite_contents", map[string]any{"media_id": float64(0)})
	if !res.IsError || !strings.Contains(svcToolText(t, res), "media_id") {
		t.Errorf("contents zero id: %s", svcToolText(t, res))
	}

	// MaxPageSize<=0 时回落到 100 的配置分支
	s2 := server.NewMCPServer("svc-test2", "1.0.0", server.WithToolCapabilities(false))
	cfg2 := svcMcpCfg()
	cfg2.Mcp.MaxPageSize = 0
	registerMCPTools(s2, cfg2)
	res = svcCallTool(t, s2, "get_liked_videos", map[string]any{"page_size": float64(100000)})
	if res.IsError {
		t.Errorf("default max page size: %s", svcToolText(t, res))
	}
}

func svcVideoBaseInfo(bvid, title, tname string) *models.VideoBaseInfo {
	now := time.Now().Unix()
	return &models.VideoBaseInfo{
		Bvid:       bvid,
		Aid:        123456,
		Tid:        7700,
		Tname:      tname,
		Title:      title,
		Desc:       "svc desc",
		Duration:   100,
		Pubdate:    now,
		Ctime:      now,
		FetchTime:  now,
		UpdateTime: now,
		OwnerMid:   7700099,
		OwnerName:  "svcUP",
	}
}

// TestSVCMCPToolClampsAndDBErrors 覆盖分页钳制、默认年份以及损坏年份表带来的 DB 错误分支。
func TestSVCMCPToolClampsAndDBErrors(t *testing.T) {
	s := server.NewMCPServer("svc-clamp", "1.0.0", server.WithToolCapabilities(false))
	registerMCPTools(s, svcMcpCfg())

	// get_history：page_size>MaxPageSize=100 钳制、page<1 归 1
	res := svcCallTool(t, s, "get_history", map[string]any{"page": float64(0), "page_size": float64(1000)})
	if res.IsError {
		t.Errorf("get_history clamp: %s", svcToolText(t, res))
	}

	// get_liked_videos / get_watch_later / get_favorite_contents 的 page/page_size 钳制分支
	res = svcCallTool(t, s, "get_liked_videos", map[string]any{"page": float64(0), "page_size": float64(1000)})
	if res.IsError {
		t.Errorf("liked clamp: %s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_watch_later", map[string]any{"page": float64(-1), "page_size": float64(1000)})
	if res.IsError {
		t.Errorf("watchlater clamp: %s", svcToolText(t, res))
	}
	res = svcCallTool(t, s, "get_favorite_contents", map[string]any{"media_id": float64(770000002), "page": float64(0), "page_size": float64(1000)})
	if res.IsError {
		t.Errorf("favcontents clamp: %s", svcToolText(t, res))
	}

	// get_daily_stats 未传 year 时默认今年（month/day 皆空 → 返回空统计而非错误）
	res = svcCallTool(t, s, "get_daily_stats", nil)
	if res.IsError {
		t.Errorf("daily stats default year: %s", svcToolText(t, res))
	}

	// get_overview 指定不存在的年份 → 获取总览失败分支
	res = svcCallTool(t, s, "get_overview", map[string]any{"year": float64(2035)})
	if !res.IsError || !strings.Contains(svcToolText(t, res), "获取总览失败") {
		t.Errorf("overview missing year: IsError=%v text=%.200s", res.IsError, svcToolText(t, res))
	}

	// get_overview 指定不存在的年份 → 获取总览失败分支
	res = svcCallTool(t, s, "get_overview", map[string]any{"year": float64(2035)})
	if !res.IsError || !strings.Contains(svcToolText(t, res), "获取总览失败") {
		t.Errorf("overview missing year: IsError=%v text=%.200s", res.IsError, svcToolText(t, res))
	}

	// 损坏年份表：bilibili_history_2036 缺少业务列 → 依赖 UNION/列查询的 DB 层报错，
	// 各工具应把错误转成工具错误而不是 panic。
	conn := svcConn()
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS bilibili_history_2036 (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create broken table: %v", err)
	}
	defer func() {
		if _, err := conn.Exec(`DROP TABLE IF EXISTS bilibili_history_2036`); err != nil {
			t.Errorf("drop broken table: %v", err)
		}
	}()

	for _, tc := range []struct{ tool string; args map[string]any; want string }{
		{"search_history", map[string]any{"keyword": "svc"}, "搜索失败"},
		{"get_history", map[string]any{"page": float64(1)}, "获取历史记录失败"},
		{"get_yearly_analysis", map[string]any{"year": float64(2036)}, "分析失败"},
	} {
		res = svcCallTool(t, s, tc.tool, tc.args)
		if !res.IsError || !strings.Contains(svcToolText(t, res), tc.want) {
			t.Errorf("%s with broken table: IsError=%v text=%.200s, want contains %q", tc.tool, res.IsError, svcToolText(t, res), tc.want)
		}
	}
}

func TestSVCMCPResources(t *testing.T) {
	s := server.NewMCPServer("svc-res", "1.0.0", server.WithResourceCapabilities(true, false))
	registerMCPResources(s)

	res := s.ListResources()
	for _, uri := range []string{"bili://project/overview", "bili://project/data-status", "bili://project/tool-guide"} {
		r, ok := res[uri]
		if !ok {
			t.Fatalf("resource %s missing", uri)
		}
		contents, err := r.Handler(context.Background(), mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{URI: uri}})
		if err != nil {
			t.Fatalf("resource %s handler: %v", uri, err)
		}
		if len(contents) != 1 {
			t.Fatalf("resource %s contents = %d", uri, len(contents))
		}
		text := contents[0].(mcp.TextResourceContents)
		if text.MIMEType == "" {
			t.Errorf("resource %s empty mime", uri)
		}
		switch uri {
		case "bili://project/overview":
			if !strings.Contains(text.Text, "Bilibili 历史记录管理工具") {
				t.Errorf("overview content = %.100s", text.Text)
			}
		case "bili://project/data-status":
			var status map[string]any
			if err := json.Unmarshal([]byte(text.Text), &status); err != nil {
				t.Fatalf("data-status not json: %v", err)
			}
			if _, ok := status["available_years"]; !ok {
				t.Errorf("data-status missing years: %s", text.Text)
			}
			if _, ok := status["year_stats"]; !ok {
				t.Errorf("data-status missing stats")
			}
		case "bili://project/tool-guide":
			if !strings.Contains(text.Text, "search_history") {
				t.Errorf("tool guide missing tool docs")
			}
		}
	}
}
