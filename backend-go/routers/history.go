package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

func RegisterHistoryRoutes(r *gin.RouterGroup) {
	history := r.Group("/history")
	{
		history.GET("/available-years", getAvailableYears)
		history.GET("/all", getHistoryPage)
		history.GET("/search", searchHistory)
		history.GET("/remarks", getAllRemarks)
		history.POST("/update-remark", updateRemark)
		history.POST("/reset-database", resetDatabase)
		history.GET("/sqlite-version", getSQLiteVersion)
		history.POST("/batch-remarks", getBatchRemarks)
		history.GET("/by_cid/:cid", getVideoByCID)
	}
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

func getAllRemarks(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse([]interface{}{}))
}

type UpdateRemarkRequest struct {
	Bvid    string `json:"bvid" binding:"required"`
	ViewAt  int64  `json:"view_at" binding:"required"`
	Remark  string `json:"remark"`
}

func updateRemark(c *gin.Context) {
	var req UpdateRemarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	result, err := database.UpdateRemark(req.Bvid, req.ViewAt, req.Remark)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "备注更新成功",
		"data":    result,
	})
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

type BatchRemarksRequest struct {
	Items []map[string]interface{} `json:"items" binding:"required"`
}

func getBatchRemarks(c *gin.Context) {
	var req BatchRemarksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	results := make(map[string]interface{})
	c.JSON(http.StatusOK, models.SuccessResponse(results))
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
