package mcp

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/services"
)

const (
	maxDefaultPageSize = 100
)

// MCPServer is a read-only MCP server for BiliHistory data.
type MCPServer struct {
	cfg    *config.Config
	token  string
	enabled bool
}

// NewMCPServer creates a new MCP server.
func NewMCPServer() (*MCPServer, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	token := os.Getenv("BHF_MCP_TOKEN")
	enabled := token != ""

	return &MCPServer{
		cfg:     cfg,
		token:   token,
		enabled: enabled,
	}, nil
}

// Handler returns an HTTP handler for MCP endpoints.
func (s *MCPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/tools", s.handleTools)
	mux.HandleFunc("/tools/call", s.handleToolCall)
	mux.HandleFunc("/resources", s.handleResources)
	mux.HandleFunc("/resources/read", s.handleResourceRead)
	return mux
}

func (s *MCPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.writeJSON(w, map[string]interface{}{
		"name":        "BilibiliHistoryFetcher",
		"description": "Read-only MCP server for Bilibili history data",
		"version":     "1.0.0",
	})
}

func (s *MCPServer) handleTools(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(w, r) {
		return
	}

	tools := []map[string]interface{}{
		{
			"name":        "get_history_page",
			"description": "Get a paginated page of watch history records",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page":      map[string]string{"type": "integer", "description": "Page number (1-based)"},
					"page_size": map[string]string{"type": "integer", "description": "Items per page"},
					"year":      map[string]string{"type": "integer", "description": "Year filter"},
					"keyword":   map[string]string{"type": "string", "description": "Search keyword in title"},
				},
			},
		},
		{
			"name":        "get_daily_summary",
			"description": "Get daily watch count summary for a year",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]string{"type": "integer", "description": "Year"},
				},
			},
		},
		{
			"name":        "get_monthly_summary",
			"description": "Get monthly watch count summary for a year",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]string{"type": "integer", "description": "Year"},
				},
			},
		},
		{
			"name":        "get_available_years",
			"description": "Get list of years with available data",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			"name":        "get_video_info",
			"description": "Get detailed video information by BVID",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bvid": map[string]string{"type": "string", "description": "Video BVID"},
				},
				"required": []string{"bvid"},
			},
		},
		{
			"name":        "get_heatmap_data",
			"description": "Get daily heatmap data for a year",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]string{"type": "integer", "description": "Year"},
				},
			},
		},
		{
			"name":        "get_viewing_analytics",
			"description": "Get comprehensive viewing behavior analytics",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]string{"type": "integer", "description": "Year"},
				},
			},
		},
	}

	s.writeJSON(w, map[string]interface{}{"tools": tools})
}

func (s *MCPServer) handleToolCall(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(w, r) {
		return
	}

	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	var result interface{}
	var err error

	switch req.Name {
	case "get_history_page":
		page := getIntArg(req.Arguments, "page", 1)
		pageSize := getIntArg(req.Arguments, "page_size", 20)
		year := getIntArg(req.Arguments, "year", time.Now().Year())
		keyword := getStringArg(req.Arguments, "keyword", "")
		result, _, err = services.QueryHistory(year, page, pageSize, keyword)

	case "get_daily_summary":
		year := getIntArg(req.Arguments, "year", time.Now().Year())
		result, err = services.GetDailyCounts(year)

	case "get_monthly_summary":
		year := getIntArg(req.Arguments, "year", time.Now().Year())
		counts, e := services.GetDailyCounts(year)
		err = e
		// Aggregate by month
		monthly := make(map[string]int)
		for _, dc := range counts {
			month := dc.Date[:7] // YYYY-MM
			monthly[month] += dc.Count
		}
		result = monthly

	case "get_available_years":
		result, err = services.GetAvailableYears()

	case "get_heatmap_data":
		year := getIntArg(req.Arguments, "year", time.Now().Year())
		result, err = services.GetHeatmapData(year)

	case "get_viewing_analytics":
		year := getIntArg(req.Arguments, "year", time.Now().Year())
		result, err = getBasicAnalytics(year)

	default:
		s.writeError(w, http.StatusBadRequest, "Unknown tool: "+req.Name)
		return
	}

	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": formatResult(result),
			},
		},
	})
}

func (s *MCPServer) handleResources(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(w, r) {
		return
	}

	resources := []map[string]interface{}{
		{
			"uri":         "bili://project/overview",
			"name":        "project_overview",
			"description": "BilibiliHistoryFetcher project capability overview",
			"mimeType":    "application/json",
		},
		{
			"uri":         "bili://project/data-status",
			"name":        "data_status",
			"description": "Current data status and available years",
			"mimeType":    "application/json",
		},
	}

	s.writeJSON(w, map[string]interface{}{"resources": resources})
}

func (s *MCPServer) handleResourceRead(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(w, r) {
		return
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	var contents interface{}

	switch req.URI {
	case "bili://project/overview":
		contents = map[string]interface{}{
			"name":    "BilibiliHistoryFetcher",
			"summary": "Bilibili history fetching, storage, analysis and visualization backend",
			"policy":  "Read-only by default",
		}
	case "bili://project/data-status":
		years, _ := services.GetAvailableYears()
		contents = map[string]interface{}{
			"available_years": years,
			"data_path":       config.GetOutputPath("database"),
		}
	default:
		s.writeError(w, http.StatusNotFound, "Resource not found")
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      req.URI,
				"mimeType": "application/json",
				"text":     formatResult(contents),
			},
		},
	})
}

func (s *MCPServer) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if s.token == "" {
		return true
	}

	auth := r.Header.Get("Authorization")
	if auth == "Bearer "+s.token {
		return true
	}

	s.writeError(w, http.StatusUnauthorized, "Unauthorized")
	return false
}

func (s *MCPServer) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *MCPServer) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"message": message},
	})
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return defaultVal
}

func getStringArg(args map[string]interface{}, key string, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func formatResult(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func getBasicAnalytics(year int) (map[string]interface{}, error) {
	counts, err := services.GetDailyCounts(year)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, dc := range counts {
		total += dc.Count
	}

	return map[string]interface{}{
		"year":        year,
		"total_videos": total,
		"daily_counts": counts,
	}, nil
}
