package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(r *gin.RouterGroup) {
	report := r.Group("/report")
	{
		report.GET("/weekly", getWeeklyReport)
		report.GET("/monthly", getMonthlyReport)
		report.GET("/available-weeks", getAvailableWeeks)
		report.GET("/available-months", getAvailableMonths)
	}
}

func getWeeklyReport(c *gin.Context) {
	yearStr := c.DefaultQuery("year", "")
	weekStr := c.DefaultQuery("week", "")

	if yearStr == "" || weekStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("需要 year 和 week 参数"))
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil || week < 1 || week > 53 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的周数参数"))
		return
	}

	result, err := database.GetWeeklyReport(year, week)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

func getMonthlyReport(c *gin.Context) {
	yearStr := c.DefaultQuery("year", "")
	monthStr := c.DefaultQuery("month", "")

	if yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("需要 year 和 month 参数"))
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的月份参数"))
		return
	}

	result, err := database.GetMonthlyReport(year, month)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

func getAvailableWeeks(c *gin.Context) {
	yearStr := c.DefaultQuery("year", "")
	if yearStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("需要 year 参数"))
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
		return
	}

	weeks, err := database.GetAvailableReportWeeks(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":  year,
		"weeks": weeks,
	}))
}

func getAvailableMonths(c *gin.Context) {
	yearStr := c.DefaultQuery("year", "")
	if yearStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("需要 year 参数"))
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的年份参数"))
		return
	}

	months, err := database.GetAvailableReportMonths(year)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":   year,
		"months": months,
	}))
}
