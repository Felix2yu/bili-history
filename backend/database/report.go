package database

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ReportVideo struct {
	Title        string `json:"title"`
	Cover        string `json:"cover"`
	Bvid         string `json:"bvid"`
	AuthorName   string `json:"author_name"`
	AuthorFace   string `json:"author_face"`
	AuthorMid    int64  `json:"author_mid"`
	MainCategory string `json:"main_category"`
	TagName      string `json:"tag_name"`
	Duration     int    `json:"duration"`
	Progress     int    `json:"progress"`
	ViewAt       int64  `json:"view_at"`
	Dt           int    `json:"dt"`
	IsFinish     int    `json:"is_finish"`
	Business     string `json:"business"`
}

type CategoryStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type AuthorStat struct {
	Name      string `json:"name"`
	Mid       int64  `json:"mid"`
	Count     int    `json:"count"`
	Duration  int    `json:"duration"`
}

type ReportSummary struct {
	TotalVideos      int            `json:"total_videos"`
	TotalDuration    int            `json:"total_duration"`
	UniqueDays       int            `json:"unique_days"`
	UniqueAuthors    int            `json:"unique_authors"`
	AvgDailyVideos   float64        `json:"avg_daily_videos"`
	AvgDailyDuration float64        `json:"avg_daily_duration"`
	TopCategories    []CategoryStat `json:"top_categories"`
	TopAuthors       []AuthorStat   `json:"top_authors"`
	DeviceDist       map[string]int `json:"device_dist"`
	DailyBreakdown   []DailyBreakdown `json:"daily_breakdown"`
	HourDist         map[int]int    `json:"hour_dist"`
	CompletionStats  ReportCompletionStats `json:"completion_stats"`
	TopVideos        []ReportVideo  `json:"top_videos"`
}

type DailyBreakdown struct {
	Date       string `json:"date"`
	Count      int    `json:"count"`
	Duration   int    `json:"duration"`
	UniqueUp   int    `json:"unique_up"`
}

type ReportCompletionStats struct {
	Finished     int     `json:"finished"`
	Partial      int     `json:"partial"`
	AvgRate      float64 `json:"avg_rate"`
}

type WeeklyReportResponse struct {
	Year      int           `json:"year"`
	Week      int           `json:"week"`
	StartDate string        `json:"start_date"`
	EndDate   string        `json:"end_date"`
	Summary   ReportSummary `json:"summary"`
	Videos    []ReportVideo `json:"videos"`
}

type MonthlyReportResponse struct {
	Year    int            `json:"year"`
	Month   int            `json:"month"`
	Summary ReportSummary  `json:"summary"`
	Videos  []ReportVideo  `json:"videos"`
}

// GetWeeklyReport returns all videos watched in a given ISO week
func GetWeeklyReport(year, week int) (*WeeklyReportResponse, error) {
	startDate, endDate := getWeekRange(year, week)

	videos, err := queryReportVideos(startDate, endDate)
	if err != nil {
		return nil, err
	}

	dayCount := int(endDate.Sub(startDate).Hours()/24) + 1
	summary := computeSummary(videos, dayCount)

	return &WeeklyReportResponse{
		Year:      year,
		Week:      week,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Summary:   summary,
		Videos:    videos,
	}, nil
}

// GetMonthlyReport returns all videos watched in a given month
func GetMonthlyReport(year, month int) (*MonthlyReportResponse, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	videos, err := queryReportVideos(startDate, endDate)
	if err != nil {
		return nil, err
	}

	dayCount := endDate.Sub(startDate).Hours() / 24
	summary := computeSummary(videos, int(dayCount))

	return &MonthlyReportResponse{
		Year:    year,
		Month:   month,
		Summary: summary,
		Videos:  videos,
	}, nil
}

// getWeekRange calculates the start (Monday) and end (Sunday) of an ISO week.
// ISO 8601: Jan 4 is always in week 1. Monday is day 1, Sunday is day 7.
func getWeekRange(year, week int) (time.Time, time.Time) {
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.Local)
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	week1Monday := jan4.AddDate(0, 0, -weekday+1)

	startDate := week1Monday.AddDate(0, 0, (week-1)*7)
	endDate := startDate.AddDate(0, 0, 6).Add(time.Hour*24 - time.Second)

	return startDate, endDate
}

// queryReportVideos queries all videos within a time range across year tables
func queryReportVideos(startDate, endDate time.Time) ([]ReportVideo, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	startTS := startDate.Unix()
	endTS := endDate.Unix()

	availableYears, err := db.GetAvailableYears()
	if err != nil {
		return nil, err
	}

	var queries []string
	var paramsList []interface{}

	reportColumns := "title, cover, bvid, author_name, author_face, author_mid, main_category, tag_name, duration, progress, view_at, dt, is_finish, business"

	for _, year := range availableYears {
		yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
		yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

		// Skip years that don't overlap with the query range
		if yearEnd.Before(startDate) || yearStart.After(endDate) {
			continue
		}

		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		query := fmt.Sprintf(
			"SELECT %s FROM %s WHERE view_at >= ? AND view_at < ? AND status = 0 AND business NOT IN ('live', 'article', 'article-list')",
			reportColumns, tableName,
		)
		queries = append(queries, query)
		paramsList = append(paramsList, startTS, endTS)
	}

	if len(queries) == 0 {
		return []ReportVideo{}, nil
	}

	baseQuery := strings.Join(queries, " UNION ALL ")
	finalQuery := fmt.Sprintf("SELECT * FROM (%s) ORDER BY view_at DESC", baseQuery)

	rows, err := conn.Query(finalQuery, paramsList...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []ReportVideo
	for rows.Next() {
		var v ReportVideo
		if err := rows.Scan(
			&v.Title, &v.Cover, &v.Bvid, &v.AuthorName, &v.AuthorFace,
			&v.AuthorMid, &v.MainCategory, &v.TagName, &v.Duration,
			&v.Progress, &v.ViewAt, &v.Dt, &v.IsFinish, &v.Business,
		); err != nil {
			continue
		}
		videos = append(videos, v)
	}

	return videos, nil
}

// computeSummary calculates aggregate statistics from a video list
func computeSummary(videos []ReportVideo, dayCount int) ReportSummary {
	if len(videos) == 0 {
		return ReportSummary{
			DeviceDist:  make(map[string]int),
			HourDist:    make(map[int]int),
		}
	}

	summary := ReportSummary{
		TotalVideos:  len(videos),
		DeviceDist:   make(map[string]int),
		HourDist:     make(map[int]int),
	}

	authorMap := make(map[string]*AuthorStat)
	categoryMap := make(map[string]int)
	uniqueAuthorMids := make(map[int64]bool)

	// Daily breakdown tracking
	type dayInfo struct {
		count    int
		duration int
		authors  map[int64]bool
	}
	dailyMap := make(map[string]*dayInfo)

	var totalCompletionRate float64
	var completionCount int

	for _, v := range videos {
		summary.TotalDuration += v.Duration

		// Track daily breakdown
		day := time.Unix(v.ViewAt, 0).Format("2006-01-02")
		if _, ok := dailyMap[day]; !ok {
			dailyMap[day] = &dayInfo{authors: make(map[int64]bool)}
		}
		dailyMap[day].count++
		dailyMap[day].duration += v.Duration
		dailyMap[day].authors[v.AuthorMid] = true

		// Track unique authors
		uniqueAuthorMids[v.AuthorMid] = true
		if stat, ok := authorMap[v.AuthorName]; ok {
			stat.Count++
			stat.Duration += v.Duration
		} else {
			authorMap[v.AuthorName] = &AuthorStat{
				Name:     v.AuthorName,
				Mid:      v.AuthorMid,
				Count:    1,
				Duration: v.Duration,
			}
		}

		// Track categories
		cat := v.MainCategory
		if cat == "" {
			cat = v.TagName
		}
		if cat != "" {
			categoryMap[cat]++
		}

		// Device distribution
		summary.DeviceDist[getDeviceName(v.Dt)]++

		// Hourly distribution
		hour := time.Unix(v.ViewAt, 0).Hour()
		summary.HourDist[hour]++

		// Completion stats
		if v.Duration > 0 {
			rate := float64(v.Progress) / float64(v.Duration)
			if v.Progress == -1 || rate >= 0.9 {
				summary.CompletionStats.Finished++
			} else {
				summary.CompletionStats.Partial++
			}
			totalCompletionRate += rate
			completionCount++
		}
	}

	summary.UniqueDays = len(dailyMap)
	summary.UniqueAuthors = len(uniqueAuthorMids)

	if completionCount > 0 {
		summary.CompletionStats.AvgRate = totalCompletionRate / float64(completionCount)
	}

	if dayCount > 0 {
		summary.AvgDailyVideos = float64(summary.TotalVideos) / float64(dayCount)
		summary.AvgDailyDuration = float64(summary.TotalDuration) / float64(dayCount)
	}

	// Build daily breakdown sorted by date
	for date, info := range dailyMap {
		summary.DailyBreakdown = append(summary.DailyBreakdown, DailyBreakdown{
			Date:     date,
			Count:    info.count,
			Duration: info.duration,
			UniqueUp: len(info.authors),
		})
	}
	sort.Slice(summary.DailyBreakdown, func(i, j int) bool {
		return summary.DailyBreakdown[i].Date < summary.DailyBreakdown[j].Date
	})

	// Top categories (by count)
	for name, count := range categoryMap {
		summary.TopCategories = append(summary.TopCategories, CategoryStat{Name: name, Count: count})
	}
	sort.Slice(summary.TopCategories, func(i, j int) bool {
		return summary.TopCategories[i].Count > summary.TopCategories[j].Count
	})
	if len(summary.TopCategories) > 10 {
		summary.TopCategories = summary.TopCategories[:10]
	}

	// Top authors (by count)
	for _, stat := range authorMap {
		summary.TopAuthors = append(summary.TopAuthors, *stat)
	}
	sort.Slice(summary.TopAuthors, func(i, j int) bool {
		return summary.TopAuthors[i].Count > summary.TopAuthors[j].Count
	})
	if len(summary.TopAuthors) > 10 {
		summary.TopAuthors = summary.TopAuthors[:10]
	}

	// Top videos (by duration, up to 5)
	sortedVideos := make([]ReportVideo, len(videos))
	copy(sortedVideos, videos)
	sort.Slice(sortedVideos, func(i, j int) bool {
		return sortedVideos[i].Duration > sortedVideos[j].Duration
	})
	if len(sortedVideos) > 5 {
		sortedVideos = sortedVideos[:5]
	}
	summary.TopVideos = sortedVideos

	return summary
}

// getDeviceName maps dt code to Chinese device name
func getDeviceName(dt int) string {
	switch dt {
	case 1, 3, 5, 7, 33:
		return "手机"
	case 2:
		return "电脑"
	case 4, 6:
		return "平板"
	default:
		return "其他"
	}
}

// GetAvailableReportWeeks returns weeks with data for a given year
func GetAvailableReportWeeks(year int) ([]int, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		return []int{}, nil
	}

	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT DISTINCT
			CASE WHEN CAST(strftime('%%W', datetime(view_at, 'unixepoch', 'localtime')) AS INTEGER) = 0
				THEN 52
				ELSE CAST(strftime('%%W', datetime(view_at, 'unixepoch', 'localtime')) AS INTEGER)
			END as week_num
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND status = 0
		ORDER BY week_num
	`, tableName), yearStart.Unix(), yearEnd.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weeks []int
	for rows.Next() {
		var w int
		if err := rows.Scan(&w); err == nil {
			weeks = append(weeks, w)
		}
	}
	return weeks, nil
}

// GetAvailableReportMonths returns months with data for a given year
func GetAvailableReportMonths(year int) ([]int, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		return []int{}, nil
	}

	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT DISTINCT CAST(strftime('%%m', datetime(view_at, 'unixepoch', 'localtime')) AS INTEGER) as month_num
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND status = 0
		ORDER BY month_num
	`, tableName), yearStart.Unix(), yearEnd.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var months []int
	for rows.Next() {
		var m int
		if err := rows.Scan(&m); err == nil {
			months = append(months, m)
		}
	}
	return months, nil
}
