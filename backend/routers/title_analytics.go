package routers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

func RegisterTitleAnalyticsRoutes(r *gin.RouterGroup) {
	titleAnalytics := r.Group("/title-analytics")
	{
		titleAnalytics.GET("/stats", getTitleStats)
		titleAnalytics.GET("/patterns", getTitlePatterns)
		titleAnalytics.GET("/sentiment", getTitleSentiment)
		titleAnalytics.GET("/length", getTitleLengthAnalysis)
		titleAnalytics.GET("/trend", getTitleTrend)
	}
}

func getTitleStats(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()
	var totalTitles int
	var totalLength int
	var minLength, maxLength int
	var minLengthTitle, maxLengthTitle string

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT title, LENGTH(title) as len
		FROM %s
		WHERE title != '' AND title IS NOT NULL
		ORDER BY len ASC
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	first := true
	for rows.Next() {
		var title string
		var length int
		if rows.Scan(&title, &length) == nil {
			totalTitles++
			totalLength += length
			if first {
				minLength = length
				minLengthTitle = title
				first = false
			}
			maxLength = length
			maxLengthTitle = title
		}
	}

	avgLength := 0.0
	if totalTitles > 0 {
		avgLength = float64(totalLength) / float64(totalTitles)
	}

	// 字数分布
	distribution := map[string]int{
		"1-10":  0,
		"11-20": 0,
		"21-30": 0,
		"31-50": 0,
		"51+":   0,
	}

	rows2, err := conn.Query(fmt.Sprintf(`
		SELECT LENGTH(title) as len
		FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var length int
			if rows2.Scan(&length) == nil {
				switch {
				case length <= 10:
					distribution["1-10"]++
				case length <= 20:
					distribution["11-20"]++
				case length <= 30:
					distribution["21-30"]++
				case length <= 50:
					distribution["31-50"]++
				default:
					distribution["51+"]++
				}
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":             year,
		"total_titles":     totalTitles,
		"avg_length":       avgLength,
		"min_length":       minLength,
		"min_length_title": minLengthTitle,
		"max_length":       maxLength,
		"max_length_title": maxLengthTitle,
		"distribution":     distribution,
	}))
}

func getTitlePatterns(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()

	// 标题模式分析
	patterns := map[string]int{
		"含数字":     0,
		"含字母":     0,
		"纯中文":     0,
		"纯数字":     0,
		"中英混合":    0,
		"含标点符号":  0,
		"含问号":     0,
		"含感叹号":    0,
		"问句":       0,
		"感叹句":     0,
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT title FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		if rows.Scan(&title) == nil {
			if containsDigit(title) {
				patterns["含数字"]++
			}
			if containsAlpha(title) {
				patterns["含字母"]++
			}
			if isAllAlpha(title) {
				patterns["纯中文"]++
			}
			if isAllDigit(title) {
				patterns["纯数字"]++
			}
			if containsAlpha(title) && containsDigit(title) {
				patterns["中英混合"]++
			}
			if strings.ContainsAny(title, ",.!?;:\"\u201C\u201D\u2018\u2019") {
				patterns["含标点符号"]++
			}
			if strings.Contains(title, "?") || strings.Contains(title, "？") {
				patterns["含问号"]++
			}
			if strings.Contains(title, "!") || strings.Contains(title, "！") {
				patterns["含感叹号"]++
			}
			if strings.HasSuffix(title, "?") || strings.HasSuffix(title, "？") {
				patterns["问句"]++
			}
			if strings.HasSuffix(title, "!") || strings.HasSuffix(title, "！") {
				patterns["感叹句"]++
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":     year,
		"patterns": patterns,
	}))
}

func getTitleSentiment(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()

	sentiment := map[string]int{
		"积极": 0,
		"消极": 0,
		"中性": 0,
	}

	positiveWords := []string{"好", "棒", "赞", "美", "帅", "酷", "爱", "喜", "开心", "快乐", "优秀", "精彩", "好看", "有趣", "搞笑"}
	negativeWords := []string{"差", "烂", "丑", "恶", "恨", "难", "悲", "惨", "无聊", "垃圾", "难看", "无聊", "失望", "愤怒"}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT title FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		if rows.Scan(&title) == nil {
			hasPositive := false
			hasNegative := false
			for _, word := range positiveWords {
				if strings.Contains(title, word) {
					hasPositive = true
					break
				}
			}
			for _, word := range negativeWords {
				if strings.Contains(title, word) {
					hasNegative = true
					break
				}
			}
			if hasPositive && hasNegative {
				sentiment["中性"]++
			} else if hasPositive {
				sentiment["积极"]++
			} else if hasNegative {
				sentiment["消极"]++
			} else {
				sentiment["中性"]++
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":      year,
		"sentiment": sentiment,
	}))
}

func getTitleLengthAnalysis(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()

	// 按字数分组统计
	lengthGroups := map[string]int{
		"1-5":    0,
		"6-10":   0,
		"11-15":  0,
		"16-20":  0,
		"21-25":  0,
		"26-30":  0,
		"31-40":  0,
		"41-50":  0,
		"51-100": 0,
		"100+":   0,
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT LENGTH(title) as len
		FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var length int
		if rows.Scan(&length) == nil {
			switch {
			case length <= 5:
				lengthGroups["1-5"]++
			case length <= 10:
				lengthGroups["6-10"]++
			case length <= 15:
				lengthGroups["11-15"]++
			case length <= 20:
				lengthGroups["16-20"]++
			case length <= 25:
				lengthGroups["21-25"]++
			case length <= 30:
				lengthGroups["26-30"]++
			case length <= 40:
				lengthGroups["31-40"]++
			case length <= 50:
				lengthGroups["41-50"]++
			case length <= 100:
				lengthGroups["51-100"]++
			default:
				lengthGroups["100+"]++
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":          year,
		"length_groups": lengthGroups,
	}))
}

func getTitleTrend(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()
	query := fmt.Sprintf(`
		SELECT 
			CAST(strftime('%%m', view_at, 'unixepoch', 'localtime') AS INTEGER) as month,
			AVG(LENGTH(title)) as avg_len,
			MIN(LENGTH(title)) as min_len,
			MAX(LENGTH(title)) as max_len,
			COUNT(*) as count
		FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName)

	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			query += fmt.Sprintf(" AND CAST(strftime('%%m', view_at, 'unixepoch', 'localtime') AS INTEGER) = %d", m)
		}
	}

	query += " GROUP BY month ORDER BY month"

	rows, err := conn.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	type MonthTrend struct {
		Month   int     `json:"month"`
		AvgLen  float64 `json:"avg_len"`
		MinLen  int     `json:"min_len"`
		MaxLen  int     `json:"max_len"`
		Count   int     `json:"count"`
	}

	var trends []MonthTrend
	for rows.Next() {
		var t MonthTrend
		if rows.Scan(&t.Month, &t.AvgLen, &t.MinLen, &t.MaxLen, &t.Count) == nil {
			trends = append(trends, t)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":   year,
		"trends": trends,
	}))
}

func containsDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsAlpha(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func isAllAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ') {
			return false
		}
	}
	return true
}

func isAllDigit(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
