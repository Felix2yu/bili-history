package routers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/services"

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
		heatmap.GET("/image", serveHeatmapImage)
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
	dateStr := c.Query("date")
	yearStr := c.Query("year")

	if dateStr == "" || len(dateStr) != 4 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("日期参数无效，应为MMDD格式，例如0113表示1月13日"))
		return
	}

	month, err := strconv.Atoi(dateStr[:2])
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("月份无效"))
		return
	}

	day, err := strconv.Atoi(dateStr[2:])
	if err != nil || day < 1 || day > 31 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("日期无效"))
		return
	}

	var year int
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("年份参数无效"))
			return
		}
	} else {
		year = time.Now().Year()
	}

	stats, err := database.GetDailyCountStats(year, month, day)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}

func generateHeatmap(c *gin.Context) {
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

	outputPath, err := services.GenerateHeatmapPNG(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("生成热力图失败: "+err.Error()))
		return
	}

	cfg, _ := config.LoadConfig()
	host := "127.0.0.1"
	port := 8899
	if cfg != nil {
		if cfg.Server.Host == "0.0.0.0" || cfg.Server.Host == "::" || cfg.Server.Host == "" {
			host = "127.0.0.1"
		} else {
			host = cfg.Server.Host
		}
		if cfg.Server.Port > 0 {
			port = cfg.Server.Port
		}
	}

	imageURL := fmt.Sprintf("http://%s:%d/api/heatmap/image?year=%d", host, port, year)

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "热力图生成成功",
		"year":       year,
		"file_path":  outputPath,
		"image_url":  imageURL,
	})
}

func serveHeatmapImage(c *gin.Context) {
	yearStr := c.Query("year")
	var year int

	if yearStr != "" {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
			return
		}
	} else {
		year = time.Now().Year()
	}

	outputDir := "output/heatmap"
	cfg, _ := config.LoadConfig()
	if cfg != nil && cfg.Heatmap.OutputDir != "" {
		outputDir = cfg.Heatmap.OutputDir
	} else if cfg != nil && cfg.OutputFolder != "" {
		outputDir = cfg.OutputFolder + "/heatmap"
	}

	filePath := fmt.Sprintf("%s/heatmap_%d.png", outputDir, year)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse("热力图不存在，请先生成"))
		return
	}

	c.File(filePath)
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
