package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bili-history/internal/db"
	"bili-history/internal/services/analytics"

	"github.com/gin-gonic/gin"
)

// GetMonthlyStats returns monthly viewing statistics.
func GetMonthlyStats(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	// Monthly stats
	rows, err := database.Query(fmt.Sprintf(`
		SELECT strftime('%%Y-%%m', datetime(view_at + 28800, 'unixepoch')) as month,
			COUNT(*) as cnt
		FROM %s GROUP BY month ORDER BY month`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	monthlyStats := make(map[string]int)
	for rows.Next() {
		var month string
		var count int
		rows.Scan(&month, &count)
		monthlyStats[month] = count
	}

	// Total videos
	var totalVideos int
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&totalVideos)

	// Active days
	var activeDays int
	database.QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT strftime('%%Y-%%m-%d', datetime(view_at + 28800, 'unixepoch')))
		FROM %s`, tableName)).Scan(&activeDays)

	// Peak month
	peakMonth := ""
	peakCount := 0
	for m, c := range monthlyStats {
		if c > peakCount {
			peakMonth = m
			peakCount = c
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"monthly_stats": monthlyStats,
			"total_videos":  totalVideos,
			"active_days":   activeDays,
			"peak_month":    peakMonth,
			"peak_count":    peakCount,
			"year":          year,
		},
	})
}

// GetWeeklyStats returns weekly viewing statistics.
func GetWeeklyStats(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	weekdayMap := map[string]string{
		"0": "周日", "1": "周一", "2": "周二", "3": "周三",
		"4": "周四", "5": "周五", "6": "周六",
	}
	weeklyStats := make(map[string]int)
	for _, v := range weekdayMap {
		weeklyStats[v] = 0
	}

	rows, err := database.Query(fmt.Sprintf(`
		SELECT strftime('%%w', datetime(view_at + 28800, 'unixepoch')) as weekday,
			COUNT(*) as cnt
		FROM %s GROUP BY weekday`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var wd string
		var count int
		rows.Scan(&wd, &count)
		if name, ok := weekdayMap[wd]; ok {
			weeklyStats[name] = count
		}
	}

	// Weekend vs weekday
	weekendCount := weeklyStats["周六"] + weeklyStats["周日"]
	weekdayCount := 0
	for _, name := range []string{"周一", "周二", "周三", "周四", "周五"} {
		weekdayCount += weeklyStats[name]
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"weekly_stats":   weeklyStats,
			"weekend_count":  weekendCount,
			"weekday_count":  weekdayCount,
			"year":           year,
		},
	})
}

// GetTimeSlots returns time-of-day viewing distribution.
func GetTimeSlots(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	rows, err := database.Query(fmt.Sprintf(`
		SELECT strftime('%%H', datetime(view_at + 28800, 'unixepoch')) as hour,
			COUNT(*) as cnt
		FROM %s GROUP BY hour ORDER BY hour`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	dailyTimeSlots := make(map[string]int)
	for rows.Next() {
		var hour string
		var count int
		rows.Scan(&hour, &count)
		dailyTimeSlots[fmt.Sprintf("%s时", hour)] = count
	}

	// Peak hour
	peakHour := ""
	peakCount := 0
	for h, c := range dailyTimeSlots {
		if c > peakCount {
			peakHour = h
			peakCount = c
		}
	}

	// Time period distribution
	morning := 0   // 6-12
	afternoon := 0 // 12-18
	evening := 0   // 18-24
	night := 0     // 0-6

	for h, c := range dailyTimeSlots {
		var hour int
		fmt.Sscanf(h, "%d时", &hour)
		switch {
		case hour >= 6 && hour < 12:
			morning += c
		case hour >= 12 && hour < 18:
			afternoon += c
		case hour >= 18 && hour < 24:
			evening += c
		default:
			night += c
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"daily_time_slots": dailyTimeSlots,
			"peak_hour":        peakHour,
			"peak_count":       peakCount,
			"periods": gin.H{
				"morning":   morning,
				"afternoon": afternoon,
				"evening":   evening,
				"night":     night,
			},
			"year": year,
		},
	})
}

// GetContinuity returns viewing continuity analysis.
func GetContinuity(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	rows, err := database.Query(fmt.Sprintf(`
		SELECT DISTINCT date(view_at + 28800, 'unixepoch') as view_date
		FROM %s ORDER BY view_date`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		rows.Scan(&d)
		dates = append(dates, d)
	}

	if len(dates) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"max_streak": 0, "total_days": 0}})
		return
	}

	// Calculate streaks
	maxStreak := 1
	currentStreak := 1
	streakStart := dates[0]
	streakEnd := dates[0]
	currentStart := dates[0]

	for i := 1; i < len(dates); i++ {
		t1, _ := time.Parse("2006-01-02", dates[i-1])
		t2, _ := time.Parse("2006-01-02", dates[i])
		if t2.Sub(t1).Hours() <= 24 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
				streakStart = currentStart
				streakEnd = dates[i]
			}
		} else {
			currentStreak = 1
			currentStart = dates[i]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"max_streak":     maxStreak,
			"streak_start":   streakStart,
			"streak_end":     streakEnd,
			"current_streak": currentStreak,
			"total_days":     len(dates),
			"year":           year,
		},
	})
}

// GetCompletionRates returns video completion rate analysis.
func GetCompletionRates(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	rows, err := database.Query(fmt.Sprintf(`
		SELECT duration, progress FROM %s
		WHERE duration > 0`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	totalVideos := 0
	fullWatched := 0
	notStarted := 0
	totalCompletion := 0.0

	for rows.Next() {
		var duration, progress int
		rows.Scan(&duration, &progress)
		totalVideos++

		var rate float64
		if progress == -1 {
			rate = 1.0
			fullWatched++
		} else if progress == 0 {
			notStarted++
			rate = 0.0
		} else if duration > 0 {
			rate = float64(progress) / float64(duration)
			if rate > 1.0 {
				rate = 1.0
			}
		}
		totalCompletion += rate
	}

	avgCompletion := 0.0
	if totalVideos > 0 {
		avgCompletion = totalCompletion / float64(totalVideos)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_videos":    totalVideos,
			"full_watched":    fullWatched,
			"not_started":     notStarted,
			"avg_completion":  avgCompletion,
			"year":            year,
		},
	})
}

// GetAuthorCompletion returns per-author completion rate analysis.
func GetAuthorCompletion(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	rows, err := database.Query(fmt.Sprintf(`
		SELECT owner_name, duration, progress FROM %s
		WHERE duration > 0 AND owner_name IS NOT NULL AND owner_name != ''`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type authorData struct {
		totalVideos   int
		totalDuration float64
		totalProgress float64
	}
	authors := make(map[string]*authorData)

	for rows.Next() {
		var ownerName string
		var duration, progress int
		rows.Scan(&ownerName, &duration, &progress)

		if _, ok := authors[ownerName]; !ok {
			authors[ownerName] = &authorData{}
		}
		a := authors[ownerName]
		a.totalVideos++
		a.totalDuration += float64(duration)

		if progress == -1 {
			a.totalProgress += float64(duration)
		} else if progress > 0 {
			a.totalProgress += float64(progress)
		}
	}

	// Build result
	type authorResult struct {
		Name       string  `json:"name"`
		VideoCount int     `json:"video_count"`
		AvgRate    float64 `json:"avg_completion_rate"`
	}
	var results []authorResult
	for name, data := range authors {
		avgRate := 0.0
		if data.totalDuration > 0 {
			avgRate = data.totalProgress / data.totalDuration
		}
		results = append(results, authorResult{
			Name:       name,
			VideoCount: data.totalVideos,
			AvgRate:    avgRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"year":  year,
	})
}

// GetTagAnalysis returns tag-based viewing analysis.
func GetTagAnalysis(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	rows, err := database.Query(fmt.Sprintf(`
		SELECT tag_name, COUNT(*) as cnt
		FROM %s
		WHERE tag_name IS NOT NULL AND tag_name != ''
		GROUP BY tag_name ORDER BY cnt DESC LIMIT 20`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type tagStat struct {
		TagName string `json:"tag_name"`
		Count   int    `json:"count"`
	}
	var tags []tagStat
	for rows.Next() {
		var t tagStat
		rows.Scan(&t.TagName, &t.Count)
		tags = append(tags, t)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tags,
		"year": year,
	})
}

// GetDurationAnalysis returns video duration distribution analysis.
func GetDurationAnalysis(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	rows, err := database.Query(fmt.Sprintf(`
		SELECT duration FROM %s WHERE duration > 0`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	buckets := map[string]int{
		"0-1分钟":   0,
		"1-5分钟":   0,
		"5-10分钟":  0,
		"10-30分钟": 0,
		"30-60分钟": 0,
		"60分钟以上": 0,
	}
	totalDuration := 0
	count := 0

	for rows.Next() {
		var dur int
		rows.Scan(&dur)
		count++
		totalDuration += dur

		switch {
		case dur < 60:
			buckets["0-1分钟"]++
		case dur < 300:
			buckets["1-5分钟"]++
		case dur < 600:
			buckets["5-10分钟"]++
		case dur < 1800:
			buckets["10-30分钟"]++
		case dur < 3600:
			buckets["30-60分钟"]++
		default:
			buckets["60分钟以上"]++
		}
	}

	avgDuration := 0.0
	if count > 0 {
		avgDuration = float64(totalDuration) / float64(count)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"duration_buckets": buckets,
			"avg_duration":     avgDuration,
			"total_videos":     count,
			"total_duration":   totalDuration,
			"year":             year,
		},
	})
}

// GetWatchCounts returns repeated viewing analysis.
func GetWatchCounts(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	tableName := db.GetYearTable(year)

	rows, err := database.Query(fmt.Sprintf(`
		SELECT bvid, title, COUNT(*) as view_count
		FROM %s
		GROUP BY bvid HAVING view_count > 1
		ORDER BY view_count DESC LIMIT 20`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type watchCount struct {
		BVID      string `json:"bvid"`
		Title     string `json:"title"`
		ViewCount int    `json:"view_count"`
	}
	var results []watchCount
	for rows.Next() {
		var w watchCount
		rows.Scan(&w.BVID, &w.Title, &w.ViewCount)
		results = append(results, w)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"year": year,
	})
}

// GetAnnualSummary returns annual summary data.
func GetAnnualSummary(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	data, err := analytics.AnalyzeViewing(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"year": year,
	})
}
