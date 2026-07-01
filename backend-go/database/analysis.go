package database

import (
	"fmt"
	"time"

	"bilibili-history-go/utils"
)

type DailyStats struct {
	Date              string             `json:"date"`
	TotalCount        int                `json:"total_count"`
	UniqueAuthors     int                `json:"unique_authors"`
	AvgDuration       float64            `json:"avg_duration"`
	AvgCompletionRate float64            `json:"avg_completion_rate"`
	CompletedVideos   int                `json:"completed_videos"`
	TagDistribution   map[string]int     `json:"tag_distribution"`
	AuthorDistribution map[string]int    `json:"author_distribution"`
	TotalDuration     int                `json:"total_duration"`
}

type MonthlyStats struct {
	Month             string  `json:"month"`
	TotalCount        int     `json:"total_count"`
	UniqueAuthors     int     `json:"unique_authors"`
	TotalDuration     int     `json:"total_duration"`
	AvgDuration       float64 `json:"avg_duration"`
	AvgCompletionRate float64 `json:"avg_completion_rate"`
	CompletedVideos   int     `json:"completed_videos"`
}

type AnalysisResult struct {
	Year           int            `json:"year"`
	DailyStats     []DailyStats   `json:"daily_stats"`
	MonthlyStats   []MonthlyStats `json:"monthly_stats"`
	TotalVideos    int            `json:"total_videos"`
	TotalDuration  int            `json:"total_duration"`
	UniqueAuthors  int            `json:"unique_authors"`
	AvgDuration    float64        `json:"avg_duration"`
	TagRanking     []TagCount     `json:"tag_ranking"`
	AuthorRanking  []AuthorCount  `json:"author_ranking"`
}

type TagCount struct {
	TagName string `json:"tag_name"`
	Count   int    `json:"count"`
}

type AuthorCount struct {
	AuthorName string `json:"author_name"`
	AuthorMid  int64  `json:"author_mid"`
	Count      int    `json:"count"`
}

func AnalyzeHistory(year int) (*AnalysisResult, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		return nil, fmt.Errorf("未找到 %d 年的历史记录数据", year)
	}

	result := &AnalysisResult{
		Year: year,
	}

	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)
	yearStartTS := yearStart.Unix()
	yearEndTS := yearEnd.Unix()

	var totalVideos int
	var totalDuration int64
	var uniqueAuthors int
	var avgDuration float64

	err := conn.QueryRow(fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			COALESCE(SUM(duration), 0) as total_duration,
			COUNT(DISTINCT author_mid) as unique_authors,
			COALESCE(AVG(duration), 0) as avg_duration
		FROM %s
		WHERE view_at >= ? AND view_at < ?
	`, tableName), yearStartTS, yearEndTS).Scan(&totalVideos, &totalDuration, &uniqueAuthors, &avgDuration)
	if err != nil {
		return nil, err
	}

	result.TotalVideos = totalVideos
	result.TotalDuration = int(totalDuration)
	result.UniqueAuthors = uniqueAuthors
	result.AvgDuration = avgDuration

	tagRows, err := conn.Query(fmt.Sprintf(`
		SELECT tag_name, COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND tag_name != ''
		GROUP BY tag_name
		ORDER BY count DESC
		LIMIT 20
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var tagName string
			var count int
			if tagRows.Scan(&tagName, &count) == nil {
				result.TagRanking = append(result.TagRanking, TagCount{
					TagName: tagName,
					Count:   count,
				})
			}
		}
	}

	authorRows, err := conn.Query(fmt.Sprintf(`
		SELECT author_name, author_mid, COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY author_mid
		ORDER BY count DESC
		LIMIT 20
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer authorRows.Close()
		for authorRows.Next() {
			var authorName string
			var authorMid int64
			var count int
			if authorRows.Scan(&authorName, &authorMid, &count) == nil {
				result.AuthorRanking = append(result.AuthorRanking, AuthorCount{
					AuthorName: authorName,
					AuthorMid:  authorMid,
					Count:      count,
				})
			}
		}
	}

	monthlyStatsMap := make(map[int]*MonthlyStats)
	for month := 1; month <= 12; month++ {
		monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		monthEnd := monthStart.AddDate(0, 1, 0)
		monthStartTS := monthStart.Unix()
		monthEndTS := monthEnd.Unix()

		var monthCount int
		var monthDuration int64
		var monthAuthors int
		var monthAvgDuration float64
		var monthAvgCompletion float64
		var monthCompleted int

		err := conn.QueryRow(fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_count,
				COALESCE(SUM(duration), 0) as total_duration,
				COUNT(DISTINCT author_mid) as unique_authors,
				COALESCE(AVG(duration), 0) as avg_duration,
				COALESCE(AVG(CAST(progress AS FLOAT) / NULLIF(duration, 0)), 0) as avg_completion_rate,
				COUNT(CASE WHEN progress >= duration * 0.9 AND duration > 0 THEN 1 END) as completed_videos
			FROM %s
			WHERE view_at >= ? AND view_at < ?
		`, tableName), monthStartTS, monthEndTS).Scan(
			&monthCount, &monthDuration, &monthAuthors,
			&monthAvgDuration, &monthAvgCompletion, &monthCompleted,
		)

		if err == nil && monthCount > 0 {
			monthlyStatsMap[month] = &MonthlyStats{
				Month:             fmt.Sprintf("%02d", month),
				TotalCount:        monthCount,
				UniqueAuthors:     monthAuthors,
				TotalDuration:     int(monthDuration),
				AvgDuration:       monthAvgDuration,
				AvgCompletionRate: monthAvgCompletion,
				CompletedVideos:   monthCompleted,
			}
		}
	}

	for month := 1; month <= 12; month++ {
		if stats, ok := monthlyStatsMap[month]; ok {
			result.MonthlyStats = append(result.MonthlyStats, *stats)
		}
	}

	dailyRows, err := conn.Query(fmt.Sprintf(`
		SELECT 
			DATE(view_at, 'unixepoch', 'localtime') as date,
			COUNT(*) as total_count,
			COUNT(DISTINCT author_mid) as unique_authors,
			COALESCE(SUM(duration), 0) as total_duration,
			COALESCE(AVG(duration), 0) as avg_duration,
			COALESCE(AVG(CAST(progress AS FLOAT) / NULLIF(duration, 0)), 0) as avg_completion_rate,
			COUNT(CASE WHEN progress >= duration * 0.9 AND duration > 0 THEN 1 END) as completed_videos
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY DATE(view_at, 'unixepoch', 'localtime')
		ORDER BY date
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer dailyRows.Close()
		for dailyRows.Next() {
			var date string
			var totalCount, uniqueAuthors, totalDuration, completedVideos int
			var avgDuration, avgCompletionRate float64

			if dailyRows.Scan(&date, &totalCount, &uniqueAuthors, &totalDuration,
				&avgDuration, &avgCompletionRate, &completedVideos) == nil {
				result.DailyStats = append(result.DailyStats, DailyStats{
					Date:              date,
					TotalCount:        totalCount,
					UniqueAuthors:     uniqueAuthors,
					TotalDuration:     totalDuration,
					AvgDuration:       avgDuration,
					AvgCompletionRate: avgCompletionRate,
					CompletedVideos:   completedVideos,
					TagDistribution:   make(map[string]int),
					AuthorDistribution: make(map[string]int),
				})
			}
		}
	}

	return result, nil
}

type HeatmapData struct {
	Year  int              `json:"year"`
	Data  map[string]int   `json:"data"`
	Total int              `json:"total"`
}

func GenerateHeatmapData(year int) (*HeatmapData, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		return nil, fmt.Errorf("未找到 %d 年的历史记录数据", year)
	}

	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT DATE(view_at, 'unixepoch', 'localtime') as date, COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY DATE(view_at, 'unixepoch', 'localtime')
		ORDER BY date
	`, tableName), yearStart.Unix(), yearEnd.Unix())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	heatmapData := &HeatmapData{
		Year: year,
		Data: make(map[string]int),
	}

	for rows.Next() {
		var date string
		var count int
		if rows.Scan(&date, &count) == nil {
			heatmapData.Data[date] = count
			heatmapData.Total += count
		}
	}

	return heatmapData, nil
}

type ViewingStats struct {
	TotalWatchTime      int            `json:"total_watch_time"`
	TotalVideos         int            `json:"total_videos"`
	AvgWatchTimePerDay  float64        `json:"avg_watch_time_per_day"`
	PeakDay             string         `json:"peak_day"`
	PeakDayCount        int            `json:"peak_day_count"`
	TimeDistribution    map[string]int `json:"time_distribution"`
	WeekdayDistribution map[string]int `json:"weekday_distribution"`
}

func GetViewingAnalytics(year int) (*ViewingStats, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		return nil, fmt.Errorf("未找到 %d 年的历史记录数据", year)
	}

	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)
	yearStartTS := yearStart.Unix()
	yearEndTS := yearEnd.Unix()

	stats := &ViewingStats{
		TimeDistribution:    make(map[string]int),
		WeekdayDistribution: make(map[string]int),
	}

	var totalDuration int64
	var totalVideos int

	err := conn.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(progress), 0) as total_watch_time, COUNT(*) as total_videos
		FROM %s
		WHERE view_at >= ? AND view_at < ?
	`, tableName), yearStartTS, yearEndTS).Scan(&totalDuration, &totalVideos)

	if err != nil {
		return nil, err
	}

	stats.TotalWatchTime = int(totalDuration)
	stats.TotalVideos = totalVideos

	daysInYear := 365
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		daysInYear = 366
	}
	stats.AvgWatchTimePerDay = float64(totalDuration) / float64(daysInYear)

	peakRows, err := conn.Query(fmt.Sprintf(`
		SELECT DATE(view_at, 'unixepoch', 'localtime') as date, COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY date
		ORDER BY count DESC
		LIMIT 1
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer peakRows.Close()
		if peakRows.Next() {
			var date string
			var count int
			if peakRows.Scan(&date, &count) == nil {
				stats.PeakDay = date
				stats.PeakDayCount = count
			}
		}
	}

	timeRows, err := conn.Query(fmt.Sprintf(`
		SELECT CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) as hour, COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY hour
		ORDER BY hour
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer timeRows.Close()
		for timeRows.Next() {
			var hour int
			var count int
			if timeRows.Scan(&hour, &count) == nil {
				stats.TimeDistribution[fmt.Sprintf("%02d", hour)] = count
			}
		}
	}

	weekdayRows, err := conn.Query(fmt.Sprintf(`
		SELECT CAST(strftime('%%w', view_at, 'unixepoch', 'localtime') AS INTEGER) as weekday, COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY weekday
		ORDER BY weekday
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer weekdayRows.Close()
		weekdayNames := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
		for weekdayRows.Next() {
			var weekday int
			var count int
			if weekdayRows.Scan(&weekday, &count) == nil {
				if weekday >= 0 && weekday < 7 {
					stats.WeekdayDistribution[weekdayNames[weekday]] = count
				}
			}
		}
	}

	return stats, nil
}

func init() {
	_ = utils.Now
}
