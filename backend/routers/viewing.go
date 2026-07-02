package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

// RegisterViewingRoutes 注册观看分析相关路由
func RegisterViewingRoutes(r *gin.RouterGroup) {
	viewing := r.Group("/viewing")
	{
		viewing.GET("/monthly-stats", getMonthlyStats)
		viewing.GET("/weekly-stats", getWeeklyStats)
		viewing.GET("/time-slots", getTimeSlots)
		viewing.GET("/continuity", getContinuity)
		viewing.GET("/", getViewingOverview)
		viewing.GET("/watch-counts", getWatchCounts)
		viewing.GET("/completion-rates", getCompletionRates)
		viewing.GET("/author-completion", getAuthorCompletion)
		viewing.GET("/tag-analysis", getTagAnalysis)
		viewing.GET("/duration-analysis", getDurationAnalysis)
	}
}

// getYearFromQuery 从查询参数获取年份，不指定则使用最新年份
func getYearFromQuery(c *gin.Context) (int, []int, bool) {
	yearStr := c.Query("year")

	db := database.GetSQLiteDB()
	availableYears, err := db.GetAvailableYears()
	if err != nil || len(availableYears) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("未找到任何历史记录数据"))
		return 0, nil, false
	}

	if yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
			return 0, nil, false
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
			return 0, nil, false
		}
		return year, availableYears, true
	}

	return availableYears[0], availableYears, true
}

// getMonthlyStats 获取月度统计
func getMonthlyStats(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeHistory(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"monthly_stats":    result.MonthlyStats,
		"total_videos":     result.TotalVideos,
		"year":             year,
		"available_years":  availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getWeeklyStats 获取周度统计
func getWeeklyStats(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeWeekly(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"weekly_stats":      result.WeeklyStats,
		"seasonal_patterns": result.SeasonalPatterns,
		"active_days":       result.ActiveDays,
		"year":              year,
		"available_years":   availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getTimeSlots 获取时段分析
func getTimeSlots(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeTimeSlots(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"daily_time_slots":  result.DailyTimeSlots,
		"peak_hours":        result.PeakHours,
		"time_investment":   result.TimeInvestment,
		"max_daily_record":  result.MaxDailyRecord,
		"year":              year,
		"available_years":   availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getContinuity 获取连续性分析
func getContinuity(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeContinuity(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"viewing_continuity": result,
		"year":               year,
		"available_years":    availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getViewingOverview 获取观看行为总览
func getViewingOverview(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.GetViewingOverview(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"details":          result.Details,
		"report":           result.Report,
		"year":             year,
		"available_years":  availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getWatchCounts 获取重复观看分析
func getWatchCounts(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeWatchCounts(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"watch_counts":       result,
		"year":               year,
		"available_years":    availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getCompletionRates 获取完成率分析
func getCompletionRates(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeCompletionRates(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"completion_rates":  result,
		"year":              year,
		"available_years":   availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getAuthorCompletion 获取UP主完成率分析
func getAuthorCompletion(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeAuthorCompletion(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"completion_rates":  result,
		"year":              year,
		"available_years":   availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getTagAnalysis 获取标签分析
func getTagAnalysis(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeTagAnalysis(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"watch_counts": map[string]interface{}{
			"tag_distribution": result.TagDistribution,
		},
		"completion_rates": map[string]interface{}{
			"tag_completion_rates": result.TagCompletionRates,
		},
		"year":             year,
		"available_years":  availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getDurationAnalysis 获取时长分析
func getDurationAnalysis(c *gin.Context) {
	year, availableYears, ok := getYearFromQuery(c)
	if !ok {
		return
	}

	result, err := database.AnalyzeDurationAnalysis(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"duration_correlation": result.DurationCorrelation,
		"year":                 year,
		"available_years":      availableYears,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}
