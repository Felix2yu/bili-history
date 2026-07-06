package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var mcpServer *server.StreamableHTTPServer

func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetMCPServerURL(cfg *config.Config) string {
	host := "127.0.0.1"
	port := 8899
	if cfg.Server.Host == "0.0.0.0" || cfg.Server.Host == "::" || cfg.Server.Host == "" {
		host = "127.0.0.1"
	} else {
		host = cfg.Server.Host
	}
	if cfg.Server.Port > 0 {
		port = cfg.Server.Port
	}
	path := cfg.Mcp.Path
	if path == "" {
		path = "/mcp"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return fmt.Sprintf("http://%s:%d%s/", host, port, path)
}

func GetMCPSkillContent(cfg *config.Config) string {
	mcpURL := GetMCPServerURL(cfg)
	authLine := "Authorization: not required"
	if cfg.Mcp.AuthEnabled {
		authLine = fmt.Sprintf("Authorization: Bearer %s", cfg.Mcp.Token)
	}
	return fmt.Sprintf(`请通过 MCP Streamable HTTP 连接我的 BilibiliHistoryFetcher 只读服务。

MCP URL: %s
%s

连接后请先读取以下 Resources：
- bili://project/overview
- bili://project/data-status
- bili://project/tool-guide

使用规则：
- 这是只读 MCP，不要请求同步、下载、删除、登录、重置数据库或修改配置。
- 查询明细时必须分页，优先使用统计/摘要工具，再按需读取 records。
- 观看历史属于隐私数据，只在当前任务需要时读取。`, mcpURL, authLine)
}

func SetupMCPServer(cfg *config.Config) {
	mcpSrv := server.NewMCPServer(
		"BilibiliHistoryFetcher",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(true, false),
	)

	registerMCPResources(mcpSrv)
	registerMCPTools(mcpSrv, cfg)

	path := cfg.Mcp.Path
	if path == "" {
		path = "/mcp"
	}

	mcpServer = server.NewStreamableHTTPServer(mcpSrv,
		server.WithEndpointPath(path),
	)

	utils.LogSuccess("MCP 服务器已初始化，路径: %s", path)
}

func GetMCPHandler() http.Handler {
	return mcpServer
}

func authMiddleware(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.Mcp.AuthEnabled {
			next.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token != cfg.Mcp.Token {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func WrapWithAuth(cfg *config.Config, handler http.Handler) http.Handler {
	return authMiddleware(cfg, handler)
}

// --- Resources ---

func registerMCPResources(s *server.MCPServer) {
	// bili://project/overview
	overviewResource := mcp.NewResource(
		"bili://project/overview",
		"项目概览",
		mcp.WithResourceDescription("Bilibili 历史记录管理工具的功能概览和使用说明"),
		mcp.WithMIMEType("text/markdown"),
	)
	s.AddResource(overviewResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		content := `# Bilibili 历史记录管理工具

## 功能
- 自动同步 B 站观看历史记录到本地 SQLite 数据库
- 按年份分表存储，支持多年数据管理
- 视频详情批量获取（标题、封面、分区、UP主信息等）
- 数据分析：年度统计、热力图、时段分析、分类分析等
- 支持搜索、收藏夹同步、稍后再看、点赞列表
- Shoutrrr 多渠道通知（每日报告、SESSDATA 健康检查）
- MCP 只读服务，供 AI 客户端查询

## 数据结构
- 历史记录按年份存储在 bilibili_history_YYYY 表中
- 每条记录包含：bvid、标题、UP主、分类、观看时间、观看时长等
- 支持备注功能，可为视频添加个人备注

## 使用须知
- 本服务为只读 MCP，不支持写入操作
- 查询时请分页，避免一次性获取大量数据`
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "bili://project/overview",
				MIMEType: "text/markdown",
				Text:     content,
			},
		}, nil
	})

	// bili://project/data-status
	dataStatusResource := mcp.NewResource(
		"bili://project/data-status",
		"数据状态",
		mcp.WithResourceDescription("当前数据库中各年份的历史记录统计"),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(dataStatusResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		years, err := database.GetSQLiteDB().GetAvailableYears()
		if err != nil {
			return nil, fmt.Errorf("获取年份失败: %w", err)
		}

		status := map[string]interface{}{
			"available_years": years,
			"year_stats":      make(map[string]int),
		}

		for _, year := range years {
			data, err := database.GenerateHeatmapData(year)
			if err == nil {
				status["year_stats"].(map[string]int)[strconv.Itoa(year)] = data.Total
			}
		}

		stats, err := database.GetVideoDetailStats()
		if err == nil {
			status["video_details"] = stats
		}

		jsonData, _ := json.MarshalIndent(status, "", "  ")
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "bili://project/data-status",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	})

	// bili://project/tool-guide
	toolGuideResource := mcp.NewResource(
		"bili://project/tool-guide",
		"工具使用指南",
		mcp.WithResourceDescription("所有可用 MCP 工具的参数说明和使用示例"),
		mcp.WithMIMEType("text/markdown"),
	)
	s.AddResource(toolGuideResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		content := `# MCP 工具使用指南

## search_history - 搜索历史记录
- keyword: 搜索关键词（匹配标题、UP主、分区、备注）
- year: 年份（可选，默认最新年份）
- page: 页码（默认 1）
- page_size: 每页数量（默认 20，最大 100）

## get_history - 获取历史记录列表
- year: 年份（可选）
- month: 月份（可选，1-12）
- day: 日期（可选，1-31）
- category: 分类名称（可选）
- page: 页码（默认 1）
- page_size: 每页数量（默认 20）

## get_daily_stats - 获取每日统计
- year: 年份（可选，默认今年）
- month: 月份（可选）
- day: 日期（可选）

## get_yearly_analysis - 获取年度分析
- year: 年份（必填）

## get_video_info - 获取视频详情
- bvid: 视频 BV 号（必填）

## get_categories - 获取分类列表
- 无需参数

## get_overview - 获取总览统计
- year: 年份（可选，默认最新年份）

## 使用建议
1. 先用 get_overview 了解整体数据情况
2. 用 get_yearly_analysis 获取详细分析
3. 用 search_history 搜索特定内容
4. 用 get_video_info 查看单个视频详情
5. 所有分页查询请控制 page_size 不超过 50`
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "bili://project/tool-guide",
				MIMEType: "text/markdown",
				Text:     content,
			},
		}, nil
	})
}

// --- Tools ---

func registerMCPTools(s *server.MCPServer, cfg *config.Config) {
	maxPageSize := cfg.Mcp.MaxPageSize
	if maxPageSize <= 0 {
		maxPageSize = 100
	}

	// search_history
	searchTool := mcp.NewTool("search_history",
		mcp.WithDescription("搜索 B 站观看历史记录，支持按关键词搜索标题、UP主、分区、备注"),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("搜索关键词")),
		mcp.WithNumber("year", mcp.Description("年份，如 2024")),
		mcp.WithNumber("page", mcp.Description("页码，默认 1")),
		mcp.WithNumber("page_size", mcp.Description(fmt.Sprintf("每页数量，默认 20，最大 %d", maxPageSize))),
	)
	s.AddTool(searchTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		keyword, _ := req.RequireString("keyword")
		year := int(req.GetFloat("year", 0))
		page := int(req.GetFloat("page", 1))
		pageSize := int(req.GetFloat("page_size", 20))

		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
		if page < 1 {
			page = 1
		}

		params := database.HistorySearchParams{
			Page:       page,
			Size:       pageSize,
			SortOrder:  0,
			Search:     keyword,
			SearchType: "all",
		}

		result, _, err := database.SearchHistory(params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("搜索失败: %v", err)), nil
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// get_history
	historyTool := mcp.NewTool("get_history",
		mcp.WithDescription("分页获取 B 站观看历史记录列表"),
		mcp.WithNumber("year", mcp.Description("年份")),
		mcp.WithNumber("month", mcp.Description("月份 (1-12)")),
		mcp.WithNumber("day", mcp.Description("日期 (1-31)")),
		mcp.WithString("category", mcp.Description("分类名称")),
		mcp.WithNumber("page", mcp.Description("页码，默认 1")),
		mcp.WithNumber("page_size", mcp.Description(fmt.Sprintf("每页数量，默认 20，最大 %d", maxPageSize))),
	)
	s.AddTool(historyTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		year := int(req.GetFloat("year", 0))
		month := int(req.GetFloat("month", 0))
		day := int(req.GetFloat("day", 0))
		category, _ := req.GetString("category", "")
		page := int(req.GetFloat("page", 1))
		pageSize := int(req.GetFloat("page_size", 20))

		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
		if page < 1 {
			page = 1
		}

		dateRange := ""
		if year > 0 && month > 0 && day > 0 {
			dateRange = fmt.Sprintf("%d%02d%02d-%d%02d%02d", year, month, day, year, month, day)
		} else if year > 0 && month > 0 {
			dateRange = fmt.Sprintf("%d%02d01-%d%02d31", year, month, year, month)
		}

		params := database.HistoryQueryParams{
			Page:      page,
			Size:      pageSize,
			SortOrder: 0,
			DateRange: dateRange,
			Business:  category,
		}

		result, _, err := database.GetHistoryPage(params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("获取历史记录失败: %v", err)), nil
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// get_daily_stats
	dailyStatsTool := mcp.NewTool("get_daily_stats",
		mcp.WithDescription("获取指定日期的观看统计"),
		mcp.WithNumber("year", mcp.Description("年份，默认今年")),
		mcp.WithNumber("month", mcp.Description("月份 (1-12)")),
		mcp.WithNumber("day", mcp.Description("日期 (1-31)")),
	)
	s.AddTool(dailyStatsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		year := int(req.GetFloat("year", 0))
		month := int(req.GetFloat("month", 0))
		day := int(req.GetFloat("day", 0))

		if year == 0 {
			// TODO: use current year from time.Now()
			year = 2024
		}

		years, err := database.GetSQLiteDB().GetAvailableYears()
		if err != nil || len(years) == 0 {
			return mcp.NewToolResultError("无可用数据"), nil
		}

		var stats []map[string]interface{}
		if month > 0 && day > 0 {
			count, total, err := database.GetDailyStats(fmt.Sprintf("%02d%02d", month, day), strconv.Itoa(year))
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("获取统计失败: %v", err)), nil
			}
			stats = append(stats, map[string]interface{}{
				"year":  year,
				"month": month,
				"day":   day,
				"count": count,
				"total": total,
			})
		} else if month > 0 {
			for d := 1; d <= 31; d++ {
				count, total, err := database.GetDailyStats(fmt.Sprintf("%02d%02d", month, d), strconv.Itoa(year))
				if err != nil {
					continue
				}
				if count > 0 {
					stats = append(stats, map[string]interface{}{
						"month": month,
						"day":   d,
						"count": count,
						"total": total,
					})
				}
			}
		}

		jsonData, _ := json.MarshalIndent(stats, "", "  ")
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// get_yearly_analysis
	yearlyTool := mcp.NewTool("get_yearly_analysis",
		mcp.WithDescription("获取指定年份的完整观看分析，包含分类排行、UP主排行、月度/每日统计等"),
		mcp.WithNumber("year", mcp.Required(), mcp.Description("年份，如 2024")),
	)
	s.AddTool(yearlyTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		year := int(req.GetFloat("year", 0))

		result, err := database.AnalyzeHistory(year)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("分析失败: %v", err)), nil
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// get_video_info
	videoInfoTool := mcp.NewTool("get_video_info",
		mcp.WithDescription("获取单个视频的详细信息（标题、UP主、分区、时长等）"),
		mcp.WithString("bvid", mcp.Required(), mcp.Description("视频 BV 号，如 BV1xx411c7mD")),
	)
	s.AddTool(videoInfoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		bvid, _ := req.RequireString("bvid")

		result, err := database.GetVideoBaseInfoByBvid(bvid)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("获取视频信息失败: %v", err)), nil
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// get_categories
	categoriesTool := mcp.NewTool("get_categories",
		mcp.WithDescription("获取所有视频分类列表（主分类和子分类）"),
	)
	s.AddTool(categoriesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := database.GetCategories()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("获取分类失败: %v", err)), nil
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// get_overview
	overviewTool := mcp.NewTool("get_overview",
		mcp.WithDescription("获取观看历史总览统计，包含总记录数、各年份数据概要"),
		mcp.WithNumber("year", mcp.Description("年份，默认最新年份")),
	)
	s.AddTool(overviewTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		year := int(req.GetFloat("year", 0))

		years, err := database.GetSQLiteDB().GetAvailableYears()
		if err != nil || len(years) == 0 {
			return mcp.NewToolResultError("无可用数据"), nil
		}

		if year == 0 {
			year = years[0]
		}

		overview, err := database.GetViewingOverview(year)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("获取总览失败: %v", err)), nil
		}

		result := map[string]interface{}{
			"available_years": years,
			"overview":        overview,
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonData)), nil
	})
}
