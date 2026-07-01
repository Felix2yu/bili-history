package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

func RegisterAnalysisRoutes(r *gin.RouterGroup) {
	analysis := r.Group("/analysis")
	{
		analysis.POST("/analyze", analyzeHistory)
	}

	daily := r.Group("/daily")
	{
		daily.GET("/stats", getDailyStats)
	}

	heatmap := r.Group("/heatmap")
	{
		heatmap.POST("/generate_heatmap", generateHeatmap)
		heatmap.GET("/data", getHeatmapData)
	}

	viewing := r.Group("/viewing")
	{
		viewing.GET("/stats", getViewingStats)
	}
}

func analyzeHistory(c *gin.Context) {
	yearStr := c.Query("year")
	var year int

	db := database.GetSQLiteDB()
	availableYears, err := db.GetAvailableYears()
	if err != nil || len(availableYears) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("未找到任何历史记录数据"))
		return
	}

	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
			return
		}

		found := false
		for _, y := range availableYears {
			if y == year {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusOK, models.ErrorResponse("未找到指定年份的历史记录数据"))
			return
		}
	} else {
		year = availableYears[0]
	}

	result, err := database.AnalyzeHistory(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"message":          "分析完成",
		"data":             result,
		"year":             year,
		"available_years":  availableYears,
	})
}

func getDailyStats(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message": "每日统计功能待实现",
	}))
}

func generateHeatmap(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message": "热力图生成功能待实现",
	}))
}

func getHeatmapData(c *gin.Context) {
	yearStr := c.Query("year")
	var year int

	db := database.GetSQLiteDB()
	availableYears, err := db.GetAvailableYears()
	if err != nil || len(availableYears) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("未找到任何历史记录数据"))
		return
	}

	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
			return
		}
	} else {
		year = availableYears[0]
	}

	data, err := database.GenerateHeatmapData(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

func getViewingStats(c *gin.Context) {
	yearStr := c.Query("year")
	var year int

	db := database.GetSQLiteDB()
	availableYears, err := db.GetAvailableYears()
	if err != nil || len(availableYears) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("未找到任何历史记录数据"))
		return
	}

	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
			return
		}
	} else {
		year = availableYears[0]
	}

	stats, err := database.GetViewingAnalytics(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}
