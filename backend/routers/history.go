package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

func RegisterHistoryRoutes(r *gin.RouterGroup) {
	history := r.Group("/history")
	{
		history.GET("/available-years", getAvailableYears)
		history.GET("/dates", getHistoryDates)
		history.GET("/all", getHistoryPage)
		history.GET("/search", searchHistory)
		history.POST("/reset-database", resetDatabase)
		history.GET("/sqlite-version", getSQLiteVersion)
		history.GET("/by_cid/:cid", getVideoByCID)
		history.POST("/batch-remarks", batchGetRemarks)
		history.POST("/check-deleted", checkDeletedVideos)
		history.POST("/deleted-status", getDeletedStatus)
	}

	daily := r.Group("/daily")
	{
		daily.GET("/daily-count", getDailyCount)
	}
}

func getDailyCount(c *gin.Context) {
	date := c.Query("date")
	year := c.Query("year")

	count, totalSeconds, err := database.GetDailyStats(date, year)
	if err != nil {
		c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
			"total_count":        0,
			"total_watch_seconds": 0,
			"unique_authors":     0,
			"total_duration":     0,
		}))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"total_count":        count,
		"total_watch_seconds": totalSeconds,
		"unique_authors":     0,
		"total_duration":     0,
	}))
}

func getAvailableYears(c *gin.Context) {
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(years))
}

func getHistoryDates(c *gin.Context) {
	dates, err := database.GetHistoryDates()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(dates))
}

func getHistoryPage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	sortOrder, _ := strconv.Atoi(c.DefaultQuery("sort_order", "0"))
	tagName := c.Query("tag_name")
	mainCategory := c.Query("main_category")
	dateRange := c.Query("date_range")
	useLocalImages := c.DefaultQuery("use_local_images", "false") == "true"
	useSessdata := c.DefaultQuery("use_sessdata", "true") == "true"
	business := c.Query("business")

	params := database.HistoryQueryParams{
		Page:           page,
		Size:           size,
		SortOrder:      sortOrder,
		TagName:        tagName,
		MainCategory:   mainCategory,
		DateRange:      dateRange,
		Business:       business,
		UseLocalImages: useLocalImages,
		UseSessdata:    useSessdata,
	}

	result, availableYears, err := database.GetHistoryPage(params)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"records":        result.Records,
		"total":          result.Total,
		"size":           result.Size,
		"current":        result.Current,
		"available_years": availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

func searchHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "30"))
	sortOrder, _ := strconv.Atoi(c.DefaultQuery("sortOrder", "0"))
	search := c.Query("search")
	searchType := c.DefaultQuery("search_type", "all")
	useLocalImages := c.DefaultQuery("use_local_images", "false") == "true"
	useSessdata := c.DefaultQuery("use_sessdata", "true") == "true"

	params := database.HistorySearchParams{
		Page:           page,
		Size:           size,
		SortOrder:      sortOrder,
		Search:         search,
		SearchType:     searchType,
		UseLocalImages: useLocalImages,
		UseSessdata:    useSessdata,
	}

	result, availableYears, err := database.SearchHistory(params)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"records":        result.Records,
		"total":          result.Total,
		"size":           result.Size,
		"current":        result.Current,
		"available_years": availableYears,
		"search_info": map[string]interface{}{
			"keyword":     search,
			"type":        searchType,
			"exact_match": false,
			"sort_by":     "view_at",
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

func resetDatabase(c *gin.Context) {
	db := database.GetSQLiteDB()
	err := db.ResetDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "数据库已重置",
	})
}

func getSQLiteVersion(c *gin.Context) {
	db := database.GetSQLiteDB()
	versionInfo, err := db.GetVersionInfo()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(versionInfo))
}

func getVideoByCID(c *gin.Context) {
	cidStr := c.Param("cid")
	cid, err := strconv.ParseInt(cidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的CID"))
		return
	}

	useLocalImages := c.DefaultQuery("use_local_images", "false") == "true"
	useSessdata := c.DefaultQuery("use_sessdata", "true") == "true"

	result, err := database.GetVideoByCID(cid, useLocalImages, useSessdata)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

func batchGetRemarks(c *gin.Context) {
	var req struct {
		Items []struct {
			Bvid   string `json:"bvid"`
			ViewAt int64  `json:"view_at"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	results := make(map[string]interface{})
	for _, item := range req.Items {
		key := item.Bvid
		if key == "" {
			continue
		}
		remark, remarkTime, err := database.GetRemarkByBvidAndViewAt(item.Bvid, item.ViewAt)
		if err == nil && remark != "" {
			results[key] = map[string]interface{}{
				"remark":      remark,
				"remark_time": remarkTime,
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(results))
}

func checkDeletedVideos(c *gin.Context) {
	var body struct {
		Bvids []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Bvids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "SESSDATA 未配置"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA)
	deleted := []string{}

	for _, bvid := range body.Bvids {
		_, err := client.GetVideoInfo(bvid)
		if err != nil {
			// Video not found or deleted
			database.MarkVideoDeleted(bvid)
			deleted = append(deleted, bvid)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"deleted": deleted,
		"checked": len(body.Bvids),
	})
}

func getDeletedStatus(c *gin.Context) {
	var body struct {
		Bvids []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	status := database.GetDeletedStatus(body.Bvids)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": status})
}
