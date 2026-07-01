package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/services"
	"bili-history/internal/services/analytics"
	"bili-history/internal/services/biliapi"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns the server health status.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":          "running",
		"timestamp":       time.Now().Format(time.RFC3339),
		"scheduler_status": "running",
	})
}

// GetHistory returns paginated history records from the database.
func GetHistory(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	keyword := c.Query("keyword")
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	year, _ := strconv.Atoi(yearStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := services.QueryHistory(year, page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAvailableYears returns all years with history data.
func GetAvailableYears(c *gin.Context) {
	years, err := services.GetAvailableYears()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get years"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": years})
}

// FetchBiliHistory triggers fetching history from Bilibili API.
func FetchBiliHistory(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	if cfg.SESSDATA == "" || cfg.SESSDATA == "Cookie里的SESSDATA字段值" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SESSDATA not configured"})
		return
	}

	fetcher, err := services.NewHistoryFetcher()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fetcher"})
		return
	}

	// SSE streaming response
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	entries, err := fetcher.FetchAndStore(func(page int, count int) {
		event := fmt.Sprintf("data: {\"page\":%d,\"count\":%d}\n\n", page, count)
		c.Writer.Write([]byte(event))
		flusher.Flush()
	})

	if err != nil {
		event := fmt.Sprintf("data: {\"error\":\"%s\"}\n\n", err.Error())
		c.Writer.Write([]byte(event))
		flusher.Flush()
		return
	}

	event := fmt.Sprintf("data: {\"status\":\"done\",\"total\":%d}\n\n", len(entries))
	c.Writer.Write([]byte(event))
	flusher.Flush()
}

// FetchBiliHistorySimple triggers fetching without SSE (returns JSON).
func FetchBiliHistorySimple(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	if cfg.SESSDATA == "" || cfg.SESSDATA == "Cookie里的SESSDATA字段值" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SESSDATA not configured"})
		return
	}

	fetcher, err := services.NewHistoryFetcher()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fetcher"})
		return
	}

	entries, err := fetcher.FetchAndStore(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "History fetch completed",
		"data":    entries,
	})
}

// GetDailyCount returns daily watch counts for a given year.
func GetDailyCount(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	counts, err := services.GetDailyCounts(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": counts, "year": year})
}

// GetCategories returns the Bilibili video categories.
func GetCategories(c *gin.Context) {
	data, err := os.ReadFile(config.GetOutputPath("categories.json"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": map[string]interface{}{}})
		return
	}

	var categories interface{}
	json.Unmarshal(data, &categories)
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GetHeatmapData returns heatmap data for visualization.
func GetHeatmapData(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	data, err := services.GetHeatmapData(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "year": year})
}

// DeleteLocalHistory deletes a history record from the local database.
func DeleteLocalHistory(c *gin.Context) {
	bvid := c.Param("bvid")
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	if bvid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvid is required"})
		return
	}

	affected, err := services.DeleteHistoryRecord(year, bvid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Delete successful",
		"affected": affected,
	})
}

// GetVideoDetails returns detailed video information.
func GetVideoDetails(c *gin.Context) {
	bvid := c.Query("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvid is required"})
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	info, err := client.FetchVideoInfo(bvid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": info})
}

// ImportFromSQLite imports history data from local JSON files.
func ImportFromSQLite(c *gin.Context) {
	if err := services.ImportFromJSONFiles(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Import completed",
	})
}

// GetViewingAnalytics returns viewing behavior analytics.
func GetViewingAnalytics(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	data, err := analytics.AnalyzeViewing(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analytics error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetTitleAnalytics returns title analysis results.
func GetTitleAnalytics(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	data, err := analytics.AnalyzeTitle(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Title analytics error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// ExportToExcel exports history data to Excel.
func ExportToExcel(c *gin.Context) {
	// TODO: Implement Excel export using excelize
	c.JSON(http.StatusOK, gin.H{"message": "Export feature coming soon"})
}

// CleanData removes unnecessary fields from history records.
func CleanData(c *gin.Context) {
	// TODO: Implement data cleaning
	c.JSON(http.StatusOK, gin.H{"message": "Data cleaning feature coming soon"})
}

// parseJSON is a helper to parse JSON from an io.Reader.
func parseJSON(r io.Reader, v interface{}) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
