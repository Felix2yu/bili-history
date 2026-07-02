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

// ==================== 观看分析数据结构 ====================

// LongestStreakPeriod 最长连续观看时间段
type LongestStreakPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ContinuityAnalysis 观看连续性分析结果
type ContinuityAnalysis struct {
	MaxStreak             int                 `json:"max_streak"`
	LongestStreakPeriod   LongestStreakPeriod `json:"longest_streak_period"`
	CurrentStreak         int                 `json:"current_streak"`
	CurrentStreakStart    string              `json:"current_streak_start"`
}

// TimeInvestment 时间投入分析
type TimeInvestment struct {
	MaxDurationDay   MaxDurationDay `json:"max_duration_day"`
	AvgDailyDuration float64        `json:"avg_daily_duration"`
}

// MaxDurationDay 单日最大时长
type MaxDurationDay struct {
	Date        string `json:"date"`
	VideoCount  int    `json:"video_count"`
	TotalDuration int  `json:"total_duration"`
}

// PeakHour 高峰时段
type PeakHour struct {
	Hour      string `json:"hour"`
	ViewCount int    `json:"view_count"`
}

// MaxDailyRecord 单日观看记录
type MaxDailyRecord struct {
	Date       string `json:"date"`
	VideoCount int    `json:"video_count"`
}

// TimeSlotAnalysis 时段分析结果
type TimeSlotAnalysis struct {
	DailyTimeSlots map[string]int    `json:"daily_time_slots"`
	PeakHours      []PeakHour        `json:"peak_hours"`
	TimeInvestment TimeInvestment    `json:"time_investment"`
	MaxDailyRecord *MaxDailyRecord   `json:"max_daily_record"`
}

// SeasonalPattern 季节模式
type SeasonalPattern struct {
	ViewCount   int     `json:"view_count"`
	AvgDuration float64 `json:"avg_duration"`
}

// WeeklyAnalysis 星期分析结果
type WeeklyAnalysis struct {
	WeeklyStats     map[string]int          `json:"weekly_stats"`
	SeasonalPatterns map[string]SeasonalPattern `json:"seasonal_patterns"`
	ActiveDays      int                     `json:"active_days"`
}

// OverallCompletionStats 总体完成率统计
type OverallCompletionStats struct {
	TotalVideos           int     `json:"total_videos"`
	AverageCompletionRate float64 `json:"average_completion_rate"`
	FullyWatchedCount     int     `json:"fully_watched_count"`
	NotStartedCount       int     `json:"not_started_count"`
	FullyWatchedRate      float64 `json:"fully_watched_rate"`
	NotStartedRate        float64 `json:"not_started_rate"`
}

// DurationBasedStats 基于时长的统计
type DurationBasedStats struct {
	VideoCount          int     `json:"video_count"`
	TotalCompletion     float64 `json:"total_completion"`
	FullyWatched        int     `json:"fully_watched"`
	AverageCompletionRate float64 `json:"average_completion_rate"`
	FullyWatchedRate    float64 `json:"fully_watched_rate"`
}

// CompletionStats 完成率分析结果
type CompletionStats struct {
	OverallStats       OverallCompletionStats       `json:"overall_stats"`
	DurationBasedStats map[string]DurationBasedStats `json:"duration_based_stats"`
	CompletionDistribution map[string]int            `json:"completion_distribution"`
}

// AuthorCompletionStat UP主完成率统计
type AuthorCompletionStat struct {
	AuthorMid           int64   `json:"author_mid"`
	VideoCount          int     `json:"video_count"`
	TotalCompletion     float64 `json:"total_completion"`
	FullyWatched        int     `json:"fully_watched"`
	AverageCompletionRate float64 `json:"average_completion_rate"`
	FullyWatchedRate    float64 `json:"fully_watched_rate"`
	ComprehensiveScore  float64 `json:"comprehensive_score"`
	LoyaltyScore        float64 `json:"loyalty_score"`
	QualityScore        float64 `json:"quality_score"`
}

// AuthorCompletionAnalysis UP主完成率分析结果
type AuthorCompletionAnalysis struct {
	MostWatchedAuthors        map[string]AuthorCompletionStat `json:"most_watched_authors"`
	HighestCompletionAuthors  map[string]AuthorCompletionStat `json:"highest_completion_authors"`
	MostValuableAuthors       map[string]AuthorCompletionStat `json:"most_valuable_authors"`
	PotentialAuthors          map[string]AuthorCompletionStat `json:"potential_authors"`
}

// TagCompletionStat 标签完成率统计
type TagCompletionStat struct {
	VideoCount          int     `json:"video_count"`
	TotalCompletion     float64 `json:"total_completion"`
	FullyWatched        int     `json:"fully_watched"`
	AverageCompletionRate float64 `json:"average_completion_rate"`
	FullyWatchedRate    float64 `json:"fully_watched_rate"`
}

// TagAnalysis 标签分析结果
type TagAnalysis struct {
	TagDistribution    map[string]int              `json:"tag_distribution"`
	TagCompletionRates map[string]TagCompletionStat `json:"tag_completion_rates"`
}

// DurationTypeStat 时长类型统计
type DurationTypeStat struct {
	VideoCount    int     `json:"video_count"`
	TotalDuration float64 `json:"total_duration"`
	AvgDuration   float64 `json:"avg_duration"`
}

// DurationAnalysis 时长分析结果
type DurationAnalysis struct {
	DurationCorrelation map[string]map[string]DurationTypeStat `json:"duration_correlation"`
}

// RewatchStats 重复观看统计
type RewatchStats struct {
	TotalRewatchedVideos int     `json:"total_rewatched_videos"`
	TotalUniqueVideos    int     `json:"total_unique_videos"`
	RewatchRate          float64 `json:"rewatch_rate"`
	TotalRewatchCount    int     `json:"total_rewatch_count"`
}

// MostWatchedVideo 重复观看最多的视频
type MostWatchedVideo struct {
	Title      string  `json:"title"`
	Bvid       string  `json:"bvid"`
	Duration   int     `json:"duration"`
	TagName    string  `json:"tag_name"`
	AuthorName string  `json:"author_name"`
	WatchCount int     `json:"watch_count"`
	FirstView  int64   `json:"first_view"`
	LastView   int64   `json:"last_view"`
	AvgInterval float64 `json:"avg_interval"`
}

// WatchCountAnalysis 重复观看分析结果
type WatchCountAnalysis struct {
	RewatchStats         RewatchStats       `json:"rewatch_stats"`
	MostWatchedVideos    []MostWatchedVideo `json:"most_watched_videos"`
	DurationDistribution map[string]int     `json:"duration_distribution"`
	TagDistribution      map[string]int     `json:"tag_distribution"`
}

// CategoryStats 分区统计
type CategoryStats struct {
	Category   string  `json:"category"`
	ViewCount  int     `json:"view_count"`
	WatchHours float64 `json:"watch_hours"`
}

// FavoriteUpStats 年度挚爱UP主统计
type FavoriteUpStats struct {
	Mid        int64   `json:"mid"`
	Name       string  `json:"name"`
	ViewCount  int     `json:"view_count"`
	WatchHours float64 `json:"watch_hours"`
}

// LateNightView 深夜观看记录
type LateNightView struct {
	Date   string `json:"date"`
	Time   string `json:"time"`
	Author string `json:"author"`
	Title  string `json:"title"`
}

// TimeSlotActivity 时段活跃天数
type TimeSlotActivity struct {
	Days       int     `json:"days"`
	Percentage float64 `json:"percentage"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ViewingDetails 观看行为详情
type ViewingDetails struct {
	TotalWatchHours    float64                   `json:"total_watch_hours"`
	TotalDays          int                       `json:"total_days"`
	TopCategories      []CategoryStats           `json:"top_categories"`
	FavoriteUpUsers    []FavoriteUpStats         `json:"favorite_up_users"`
	LateNightViews     []LateNightView           `json:"late_night_views"`
	TimeSlotActivity   map[string]TimeSlotActivity `json:"time_slot_activity"`
	Devices            []DeviceInfo              `json:"devices"`
}

// ViewingOverview 观看行为总览
type ViewingOverview struct {
	Details ViewingDetails    `json:"details"`
	Report  map[string]string `json:"report"`
}

// WeeklyStatsData 周度统计数据
type WeeklyStatsData struct {
	WeeklyStats     map[string]int          `json:"weekly_stats"`
	SeasonalPatterns map[string]SeasonalPattern `json:"seasonal_patterns"`
	ActiveDays      int                     `json:"active_days"`
}

// ==================== 观看分析函数 ====================

// AnalyzeContinuity 分析观看连续性
func AnalyzeContinuity(year int) (*ContinuityAnalysis, error) {
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
		SELECT DISTINCT DATE(view_at, 'unixepoch', 'localtime') as view_date
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		ORDER BY view_date
	`, tableName), yearStart.Unix(), yearEnd.Unix())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date string
		if rows.Scan(&date) == nil {
			dates = append(dates, date)
		}
	}

	if len(dates) == 0 {
		return &ContinuityAnalysis{
			MaxStreak: 0,
			LongestStreakPeriod: LongestStreakPeriod{
				Start: "",
				End:   "",
			},
			CurrentStreak:      0,
			CurrentStreakStart: "",
		}, nil
	}

	maxStreak := 1
	currentStreak := 1
	longestStreakStart := dates[0]
	longestStreakEnd := dates[0]
	currentStreakStart := dates[0]

	for i := 1; i < len(dates); i++ {
		date1, _ := time.Parse("2006-01-02", dates[i-1])
		date2, _ := time.Parse("2006-01-02", dates[i])
		if date2.Sub(date1).Hours()/24 == 1 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
				longestStreakStart = dates[i-maxStreak+1]
				longestStreakEnd = dates[i]
			}
		} else {
			currentStreak = 1
			currentStreakStart = dates[i]
		}
	}

	return &ContinuityAnalysis{
		MaxStreak: maxStreak,
		LongestStreakPeriod: LongestStreakPeriod{
			Start: longestStreakStart,
			End:   longestStreakEnd,
		},
		CurrentStreak:      currentStreak,
		CurrentStreakStart: currentStreakStart,
	}, nil
}

// AnalyzeTimeSlots 分析时段观看
func AnalyzeTimeSlots(year int) (*TimeSlotAnalysis, error) {
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

	result := &TimeSlotAnalysis{
		DailyTimeSlots: make(map[string]int),
		PeakHours:      []PeakHour{},
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
				result.DailyTimeSlots[fmt.Sprintf("%d时", hour)] = count
			}
		}
	}

	peakRows, err := conn.Query(fmt.Sprintf(`
		SELECT CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) as hour, COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY hour
		ORDER BY count DESC
		LIMIT 5
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer peakRows.Close()
		for peakRows.Next() {
			var hour int
			var count int
			if peakRows.Scan(&hour, &count) == nil {
				result.PeakHours = append(result.PeakHours, PeakHour{
					Hour:      fmt.Sprintf("%d时", hour),
					ViewCount: count,
				})
			}
		}
	}

	var maxDate string
	var maxVideoCount int
	var maxTotalDuration int64
	err = conn.QueryRow(fmt.Sprintf(`
		SELECT 
			DATE(view_at, 'unixepoch', 'localtime') as view_date,
			COUNT(*) as video_count,
			SUM(CASE WHEN progress = -1 THEN duration ELSE progress END) as total_duration
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY view_date
		ORDER BY total_duration DESC
		LIMIT 1
	`, tableName), yearStartTS, yearEndTS).Scan(&maxDate, &maxVideoCount, &maxTotalDuration)
	if err == nil {
		result.TimeInvestment.MaxDurationDay = MaxDurationDay{
			Date:          maxDate,
			VideoCount:    maxVideoCount,
			TotalDuration: int(maxTotalDuration),
		}
	}

	var avgDailyDuration float64
	err = conn.QueryRow(fmt.Sprintf(`
		SELECT AVG(daily_duration)
		FROM (
			SELECT 
				SUM(CASE WHEN progress = -1 THEN duration ELSE progress END) as daily_duration
			FROM %s
			WHERE view_at >= ? AND view_at < ?
			GROUP BY DATE(view_at, 'unixepoch', 'localtime')
		)
	`, tableName), yearStartTS, yearEndTS).Scan(&avgDailyDuration)
	if err == nil {
		result.TimeInvestment.AvgDailyDuration = avgDailyDuration
	}

	var maxDayDate string
	var maxDayCount int
	err = conn.QueryRow(fmt.Sprintf(`
		SELECT DATE(view_at, 'unixepoch', 'localtime') as view_date, COUNT(*) as video_count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY view_date
		ORDER BY video_count DESC
		LIMIT 1
	`, tableName), yearStartTS, yearEndTS).Scan(&maxDayDate, &maxDayCount)
	if err == nil {
		result.MaxDailyRecord = &MaxDailyRecord{
			Date:       maxDayDate,
			VideoCount: maxDayCount,
		}
	}

	return result, nil
}

// AnalyzeWeekly 分析星期观看分布
func AnalyzeWeekly(year int) (*WeeklyAnalysis, error) {
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

	result := &WeeklyAnalysis{
		WeeklyStats:     make(map[string]int),
		SeasonalPatterns: make(map[string]SeasonalPattern),
	}

	weekdayNames := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	for _, name := range weekdayNames {
		result.WeeklyStats[name] = 0
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
		for weekdayRows.Next() {
			var weekday int
			var count int
			if weekdayRows.Scan(&weekday, &count) == nil {
				if weekday >= 0 && weekday < 7 {
					result.WeeklyStats[weekdayNames[weekday]] = count
				}
			}
		}
	}

	seasonRows, err := conn.Query(fmt.Sprintf(`
		SELECT
			CASE
				WHEN CAST(strftime('%%m', view_at, 'unixepoch', 'localtime') AS INTEGER) IN (1,2,3) THEN '春季'
				WHEN CAST(strftime('%%m', view_at, 'unixepoch', 'localtime') AS INTEGER) IN (4,5,6) THEN '夏季'
				WHEN CAST(strftime('%%m', view_at, 'unixepoch', 'localtime') AS INTEGER) IN (7,8,9) THEN '秋季'
				WHEN CAST(strftime('%%m', view_at, 'unixepoch', 'localtime') AS INTEGER) IN (10,11,12) THEN '冬季'
			END as season,
			COUNT(*) as view_count,
			AVG(CASE WHEN progress = -1 THEN duration ELSE progress END) as avg_duration
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY season
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer seasonRows.Close()
		for seasonRows.Next() {
			var season string
			var viewCount int
			var avgDuration float64
			if seasonRows.Scan(&season, &viewCount, &avgDuration) == nil {
				result.SeasonalPatterns[season] = SeasonalPattern{
					ViewCount:   viewCount,
					AvgDuration: avgDuration,
				}
			}
		}
	}

	var activeDays int
	err = conn.QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT DATE(view_at, 'unixepoch', 'localtime'))
		FROM %s
		WHERE view_at >= ? AND view_at < ?
	`, tableName), yearStartTS, yearEndTS).Scan(&activeDays)
	if err == nil {
		result.ActiveDays = activeDays
	}

	return result, nil
}

// AnalyzeCompletionRates 分析视频完成率
func AnalyzeCompletionRates(year int) (*CompletionStats, error) {
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

	result := &CompletionStats{
		DurationBasedStats: make(map[string]DurationBasedStats),
		CompletionDistribution: map[string]int{
			"0-10%":   0,
			"10-30%":  0,
			"30-50%":  0,
			"50-70%":  0,
			"70-90%":  0,
			"90-100%": 0,
		},
	}

	durationCategories := []string{"短视频(≤5分钟)", "中等视频(5-20分钟)", "长视频(>20分钟)"}
	for _, cat := range durationCategories {
		result.DurationBasedStats[cat] = DurationBasedStats{}
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT duration, progress
		FROM %s
		WHERE view_at >= ? AND view_at < ?
	`, tableName), yearStartTS, yearEndTS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalVideos := 0
	totalCompletion := 0.0
	fullyWatched := 0
	notStarted := 0

	for rows.Next() {
		var duration int
		var progress int
		if rows.Scan(&duration, &progress) != nil {
			continue
		}

		var completionRate float64
		if progress == -1 {
			completionRate = 100
		} else if duration > 0 {
			completionRate = float64(progress) / float64(duration) * 100
		} else {
			completionRate = 0
		}

		totalVideos++
		totalCompletion += completionRate

		if completionRate >= 90 {
			fullyWatched++
		} else if completionRate == 0 {
			notStarted++
		}

		if completionRate <= 10 {
			result.CompletionDistribution["0-10%"]++
		} else if completionRate <= 30 {
			result.CompletionDistribution["10-30%"]++
		} else if completionRate <= 50 {
			result.CompletionDistribution["30-50%"]++
		} else if completionRate <= 70 {
			result.CompletionDistribution["50-70%"]++
		} else if completionRate <= 90 {
			result.CompletionDistribution["70-90%"]++
		} else {
			result.CompletionDistribution["90-100%"]++
		}

		var category string
		if duration <= 300 {
			category = "短视频(≤5分钟)"
		} else if duration <= 1200 {
			category = "中等视频(5-20分钟)"
		} else {
			category = "长视频(>20分钟)"
		}

		stats := result.DurationBasedStats[category]
		stats.VideoCount++
		stats.TotalCompletion += completionRate
		if completionRate >= 90 {
			stats.FullyWatched++
		}
		result.DurationBasedStats[category] = stats
	}

	result.OverallStats = OverallCompletionStats{
		TotalVideos:           totalVideos,
		AverageCompletionRate: 0,
		FullyWatchedCount:     fullyWatched,
		NotStartedCount:       notStarted,
		FullyWatchedRate:      0,
		NotStartedRate:        0,
	}

	if totalVideos > 0 {
		result.OverallStats.AverageCompletionRate = totalCompletion / float64(totalVideos)
		result.OverallStats.FullyWatchedRate = float64(fullyWatched) / float64(totalVideos) * 100
		result.OverallStats.NotStartedRate = float64(notStarted) / float64(totalVideos) * 100
	}

	for cat, stats := range result.DurationBasedStats {
		if stats.VideoCount > 0 {
			stats.AverageCompletionRate = stats.TotalCompletion / float64(stats.VideoCount)
			stats.FullyWatchedRate = float64(stats.FullyWatched) / float64(stats.VideoCount) * 100
			result.DurationBasedStats[cat] = stats
		}
	}

	return result, nil
}

// AnalyzeAuthorCompletion 分析UP主完成率
func AnalyzeAuthorCompletion(year int) (*AuthorCompletionAnalysis, error) {
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

	authorStats := make(map[string]*AuthorCompletionStat)

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT duration, progress, author_name, author_mid
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND author_name != ''
	`, tableName), yearStartTS, yearEndTS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var duration int
		var progress int
		var authorName string
		var authorMid int64
		if rows.Scan(&duration, &progress, &authorName, &authorMid) != nil {
			continue
		}

		var completionRate float64
		if progress == -1 {
			completionRate = 100
		} else if duration > 0 {
			completionRate = float64(progress) / float64(duration) * 100
		} else {
			completionRate = 0
		}

		if _, ok := authorStats[authorName]; !ok {
			authorStats[authorName] = &AuthorCompletionStat{
				AuthorMid: authorMid,
			}
		}
		stats := authorStats[authorName]
		stats.VideoCount++
		stats.TotalCompletion += completionRate
		if completionRate >= 90 {
			stats.FullyWatched++
		}
	}

	filteredAuthors := make(map[string]*AuthorCompletionStat)
	for name, stats := range authorStats {
		if stats.VideoCount >= 5 {
			stats.AverageCompletionRate = stats.TotalCompletion / float64(stats.VideoCount)
			stats.FullyWatchedRate = float64(stats.FullyWatched) / float64(stats.VideoCount) * 100
			filteredAuthors[name] = stats
		}
	}

	viewCounts := make([]int, 0, len(filteredAuthors))
	for _, stats := range filteredAuthors {
		viewCounts = append(viewCounts, stats.VideoCount)
	}

	var maxViews, minViews int
	if len(viewCounts) > 0 {
		maxViews = viewCounts[0]
		minViews = viewCounts[0]
		for _, v := range viewCounts {
			if v > maxViews {
				maxViews = v
			}
			if v < minViews {
				minViews = v
			}
		}
	}
	viewRange := maxViews - minViews
	if viewRange <= 0 {
		viewRange = 1
	}

	scoredAuthors := make(map[string]AuthorCompletionStat)
	for name, stats := range filteredAuthors {
		normalizedViews := float64(stats.VideoCount-minViews) / float64(viewRange) * 100
		confidence := float64(stats.VideoCount) / 20.0
		if confidence > 1.0 {
			confidence = 1.0
		}

		comprehensiveScore := (normalizedViews*0.25 + stats.AverageCompletionRate*0.5 + stats.FullyWatchedRate*0.25) * confidence

		loyaltyScore := float64(stats.VideoCount) * sqrt(stats.AverageCompletionRate/100.0)
		qualityScore := stats.AverageCompletionRate*0.7 + stats.FullyWatchedRate*0.3

		scoredStats := *stats
		scoredStats.ComprehensiveScore = roundFloat(comprehensiveScore, 2)
		scoredStats.LoyaltyScore = roundFloat(loyaltyScore, 2)
		scoredStats.QualityScore = roundFloat(qualityScore, 2)
		scoredAuthors[name] = scoredStats
	}

	mostWatched := make(map[string]AuthorCompletionStat)
	highestCompletion := make(map[string]AuthorCompletionStat)
	mostValuable := make(map[string]AuthorCompletionStat)
	potential := make(map[string]AuthorCompletionStat)

	mostWatchedList := make([]authorEntry, 0, len(scoredAuthors))
	for name, stats := range scoredAuthors {
		mostWatchedList = append(mostWatchedList, authorEntry{name, stats})
	}
	sortByVideoCountDesc(mostWatchedList)
	for i := 0; i < len(mostWatchedList) && i < 10; i++ {
		mostWatched[mostWatchedList[i].name] = mostWatchedList[i].stats
	}

	highestCompletionList := make([]authorEntry, 0, len(scoredAuthors))
	for name, stats := range scoredAuthors {
		highestCompletionList = append(highestCompletionList, authorEntry{name, stats})
	}
	sortByCompletionRateDesc(highestCompletionList)
	for i := 0; i < len(highestCompletionList) && i < 10; i++ {
		highestCompletion[highestCompletionList[i].name] = highestCompletionList[i].stats
	}

	mostValuableList := make([]authorEntry, 0, len(scoredAuthors))
	for name, stats := range scoredAuthors {
		mostValuableList = append(mostValuableList, authorEntry{name, stats})
	}
	sortByComprehensiveScoreDesc(mostValuableList)
	for i := 0; i < len(mostValuableList) && i < 10; i++ {
		mostValuable[mostValuableList[i].name] = mostValuableList[i].stats
	}

	sortedViewCounts := make([]int, 0, len(scoredAuthors))
	for _, stats := range scoredAuthors {
		sortedViewCounts = append(sortedViewCounts, stats.VideoCount)
	}
	sortIntsDesc(sortedViewCounts)

	threshold := 0
	if len(sortedViewCounts) >= 3 {
		threshold = sortedViewCounts[2]
	} else if len(sortedViewCounts) > 0 {
		threshold = sortedViewCounts[len(sortedViewCounts)-1]
	}

	for name, stats := range scoredAuthors {
		if stats.QualityScore > 85 && stats.VideoCount < threshold {
			potential[name] = stats
		}
	}

	potentialList := make([]authorEntry, 0, len(potential))
	for name, stats := range potential {
		potentialList = append(potentialList, authorEntry{name, stats})
	}
	sortByQualityScoreDesc(potentialList)
	potential = make(map[string]AuthorCompletionStat)
	for i := 0; i < len(potentialList) && i < 5; i++ {
		potential[potentialList[i].name] = potentialList[i].stats
	}

	return &AuthorCompletionAnalysis{
		MostWatchedAuthors:       mostWatched,
		HighestCompletionAuthors: highestCompletion,
		MostValuableAuthors:      mostValuable,
		PotentialAuthors:         potential,
	}, nil
}

// AnalyzeTagAnalysis 分析标签
func AnalyzeTagAnalysis(year int) (*TagAnalysis, error) {
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

	result := &TagAnalysis{
		TagDistribution:    make(map[string]int),
		TagCompletionRates: make(map[string]TagCompletionStat),
	}

	tagStats := make(map[string]*TagCompletionStat)

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT duration, progress, tag_name
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND tag_name != ''
	`, tableName), yearStartTS, yearEndTS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var duration int
		var progress int
		var tagName string
		if rows.Scan(&duration, &progress, &tagName) != nil {
			continue
		}

		var completionRate float64
		if progress == -1 {
			completionRate = 100
		} else if duration > 0 {
			completionRate = float64(progress) / float64(duration) * 100
		} else {
			completionRate = 0
		}

		result.TagDistribution[tagName]++

		if _, ok := tagStats[tagName]; !ok {
			tagStats[tagName] = &TagCompletionStat{}
		}
		stats := tagStats[tagName]
		stats.VideoCount++
		stats.TotalCompletion += completionRate
		if completionRate >= 90 {
			stats.FullyWatched++
		}
	}

	for tag, stats := range tagStats {
		if stats.VideoCount >= 5 {
			stats.AverageCompletionRate = stats.TotalCompletion / float64(stats.VideoCount)
			stats.FullyWatchedRate = float64(stats.FullyWatched) / float64(stats.VideoCount) * 100
			result.TagCompletionRates[tag] = *stats
		}
	}

	return result, nil
}

// AnalyzeDurationAnalysis 分析视频时长
func AnalyzeDurationAnalysis(year int) (*DurationAnalysis, error) {
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

	timePeriods := map[string]struct {
		start int
		end   int
	}{
		"凌晨": {0, 6},
		"上午": {6, 12},
		"下午": {12, 18},
		"晚上": {18, 24},
	}

	result := &DurationAnalysis{
		DurationCorrelation: make(map[string]map[string]DurationTypeStat),
	}

	for period := range timePeriods {
		result.DurationCorrelation[period] = map[string]DurationTypeStat{
			"短视频":   {},
			"中等视频": {},
			"长视频":   {},
		}
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT duration, view_at
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND duration > 0
	`, tableName), yearStartTS, yearEndTS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var duration int
		var viewAt int64
		if rows.Scan(&duration, &viewAt) != nil {
			continue
		}

		viewTime := time.Unix(viewAt, 0)
		hour := viewTime.Hour()

		var period string
		for p, timeRange := range timePeriods {
			if hour >= timeRange.start && hour < timeRange.end {
				period = p
				break
			}
		}
		if period == "" {
			continue
		}

		var durationType string
		if duration < 300 {
			durationType = "短视频"
		} else if duration < 1200 {
			durationType = "中等视频"
		} else {
			durationType = "长视频"
		}

		stats := result.DurationCorrelation[period][durationType]
		stats.VideoCount++
		stats.TotalDuration += float64(duration)
		result.DurationCorrelation[period][durationType] = stats
	}

	for period := range result.DurationCorrelation {
		for durationType := range result.DurationCorrelation[period] {
			stats := result.DurationCorrelation[period][durationType]
			if stats.VideoCount > 0 {
				stats.AvgDuration = stats.TotalDuration / float64(stats.VideoCount)
				result.DurationCorrelation[period][durationType] = stats
			}
		}
	}

	return result, nil
}

// AnalyzeWatchCounts 分析重复观看
func AnalyzeWatchCounts(year int) (*WatchCountAnalysis, error) {
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

	result := &WatchCountAnalysis{
		MostWatchedVideos:    []MostWatchedVideo{},
		DurationDistribution: map[string]int{},
		TagDistribution:      map[string]int{},
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT 
			title,
			bvid,
			duration,
			tag_name,
			author_name,
			COUNT(*) as watch_count,
			MIN(view_at) as first_view,
			MAX(view_at) as last_view
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND bvid != ''
		GROUP BY bvid
		HAVING COUNT(*) > 1
		ORDER BY watch_count DESC
	`, tableName), yearStartTS, yearEndTS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalRewatchedVideos := 0
	totalRewatchCount := 0
	tagDist := make(map[string]int)

	for rows.Next() {
		var title string
		var bvid string
		var duration int
		var tagName string
		var authorName string
		var watchCount int
		var firstView int64
		var lastView int64

		if rows.Scan(&title, &bvid, &duration, &tagName, &authorName, &watchCount, &firstView, &lastView) != nil {
			continue
		}

		totalRewatchedVideos++
		totalRewatchCount += watchCount - 1

		if duration <= 300 {
			result.DurationDistribution["短视频(≤5分钟)"]++
		} else if duration <= 1200 {
			result.DurationDistribution["中等视频(5-20分钟)"]++
		} else {
			result.DurationDistribution["长视频(>20分钟)"]++
		}

		if tagName != "" {
			tagDist[tagName]++
		}

		if len(result.MostWatchedVideos) < 10 {
			var avgInterval float64
			if watchCount > 1 {
				avgInterval = float64(lastView-firstView) / float64(watchCount-1)
			}
			result.MostWatchedVideos = append(result.MostWatchedVideos, MostWatchedVideo{
				Title:       title,
				Bvid:        bvid,
				Duration:    duration,
				TagName:     tagName,
				AuthorName:  authorName,
				WatchCount:  watchCount,
				FirstView:   firstView,
				LastView:    lastView,
				AvgInterval: avgInterval,
			})
		}
	}

	var totalUniqueVideos int
	err = conn.QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT bvid)
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND bvid != ''
	`, tableName), yearStartTS, yearEndTS).Scan(&totalUniqueVideos)
	if err != nil {
		return nil, err
	}

	var rewatchRate float64
	if totalUniqueVideos > 0 {
		rewatchRate = float64(totalRewatchedVideos) / float64(totalUniqueVideos) * 100
	}

	result.RewatchStats = RewatchStats{
		TotalRewatchedVideos: totalRewatchedVideos,
		TotalUniqueVideos:    totalUniqueVideos,
		RewatchRate:          rewatchRate,
		TotalRewatchCount:    totalRewatchCount,
	}

	sortedTags := sortMapByValueDesc(tagDist)
	for i, tag := range sortedTags {
		if i >= 10 {
			break
		}
		result.TagDistribution[tag.key] = tag.value
	}

	return result, nil
}

// GetViewingOverview 获取观看行为总览
func GetViewingOverview(year int) (*ViewingOverview, error) {
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

	details := ViewingDetails{
		TopCategories:    []CategoryStats{},
		FavoriteUpUsers:  []FavoriteUpStats{},
		LateNightViews:   []LateNightView{},
		TimeSlotActivity: make(map[string]TimeSlotActivity),
		Devices:          []DeviceInfo{},
	}

	var totalSeconds int64
	err := conn.QueryRow(fmt.Sprintf(`
		SELECT SUM(CASE WHEN progress = -1 THEN duration ELSE progress END)
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND progress IS NOT NULL
	`, tableName), yearStartTS, yearEndTS).Scan(&totalSeconds)
	if err == nil {
		details.TotalWatchHours = float64(totalSeconds) / 3600.0
	}

	var totalDays int
	err = conn.QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT DATE(view_at, 'unixepoch', 'localtime'))
		FROM %s
		WHERE view_at >= ? AND view_at < ?
	`, tableName), yearStartTS, yearEndTS).Scan(&totalDays)
	if err == nil {
		details.TotalDays = totalDays
	}

	categoryRows, err := conn.Query(fmt.Sprintf(`
		SELECT 
			main_category, 
			COUNT(*) as view_count,
			SUM(CASE WHEN progress = -1 THEN duration ELSE progress END) as total_progress
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND main_category IS NOT NULL AND main_category != ''
		GROUP BY main_category
		ORDER BY view_count DESC
		LIMIT 10
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer categoryRows.Close()
		for categoryRows.Next() {
			var category string
			var viewCount int
			var totalProgress int64
			if categoryRows.Scan(&category, &viewCount, &totalProgress) == nil {
				details.TopCategories = append(details.TopCategories, CategoryStats{
					Category:   category,
					ViewCount:  viewCount,
					WatchHours: float64(totalProgress) / 3600.0,
				})
			}
		}
	}

	upRows, err := conn.Query(fmt.Sprintf(`
		SELECT 
			author_mid, 
			author_name,
			COUNT(*) as view_count,
			SUM(CASE WHEN progress = -1 THEN duration ELSE progress END) as total_progress
		FROM %s
		WHERE view_at >= ? AND view_at < ? AND author_mid IS NOT NULL
		GROUP BY author_mid
		ORDER BY view_count DESC
		LIMIT 10
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer upRows.Close()
		for upRows.Next() {
			var mid int64
			var name string
			var viewCount int
			var totalProgress int64
			if upRows.Scan(&mid, &name, &viewCount, &totalProgress) == nil {
				details.FavoriteUpUsers = append(details.FavoriteUpUsers, FavoriteUpStats{
					Mid:        mid,
					Name:       name,
					ViewCount:  viewCount,
					WatchHours: float64(totalProgress) / 3600.0,
				})
			}
		}
	}

	timeSlotRows, err := conn.Query(fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) BETWEEN 5 AND 11 THEN '上午'
				WHEN CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) BETWEEN 12 AND 17 THEN '下午'
				WHEN CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) BETWEEN 18 AND 22 THEN '晚上'
				ELSE '深夜'
			END as time_slot,
			COUNT(DISTINCT DATE(view_at, 'unixepoch', 'localtime')) as active_days
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY time_slot
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer timeSlotRows.Close()
		for timeSlotRows.Next() {
			var slot string
			var days int
			if timeSlotRows.Scan(&slot, &days) == nil {
				var percentage float64
				if totalDays > 0 {
					percentage = float64(days) / float64(totalDays) * 100
				}
				details.TimeSlotActivity[slot] = TimeSlotActivity{
					Days:       days,
					Percentage: percentage,
				}
			}
		}
	}

	deviceRows, err := conn.Query(fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN dt IN (1, 3, 5, 7) THEN '手机'
				WHEN dt = 2 THEN '网页'
				WHEN dt IN (4, 6) THEN '平板'
				WHEN dt = 33 THEN '电视'
				ELSE '其他'
			END as platform,
			COUNT(*) as count
		FROM %s
		WHERE view_at >= ? AND view_at < ?
		GROUP BY platform
		ORDER BY count DESC
		LIMIT 3
	`, tableName), yearStartTS, yearEndTS)
	if err == nil {
		defer deviceRows.Close()
		for deviceRows.Next() {
			var name string
			var count int
			if deviceRows.Scan(&name, &count) == nil {
				details.Devices = append(details.Devices, DeviceInfo{
					Name:  name,
					Count: count,
				})
			}
		}
	}

	report := make(map[string]string)
	if details.TotalDays > 0 {
		report["total_summary"] = fmt.Sprintf("和B站共度的%d天里，你观看超过%.1f小时的内容", details.TotalDays, details.TotalWatchHours)
	}

	if len(details.TimeSlotActivity) > 0 {
		maxSlot := ""
		maxPercentage := 0.0
		for slot, activity := range details.TimeSlotActivity {
			if activity.Percentage > maxPercentage {
				maxPercentage = activity.Percentage
				maxSlot = slot
			}
		}
		if maxSlot != "" {
			report["time_slot_summary"] = fmt.Sprintf("%s时段访问B站天数超过%.1f%%", maxSlot, maxPercentage)
		}
	}

	if len(details.TopCategories) > 0 {
		topCat := details.TopCategories[0]
		report["category_summary"] = fmt.Sprintf("你最喜欢的分区是%s，共观看%d个视频，时长%.1f小时", topCat.Category, topCat.ViewCount, topCat.WatchHours)
	}

	if len(details.FavoriteUpUsers) > 0 {
		topUp := details.FavoriteUpUsers[0]
		report["up_summary"] = fmt.Sprintf("年度挚爱UP主是%s，共观看%d个视频，时长%.1f小时", topUp.Name, topUp.ViewCount, topUp.WatchHours)
	}

	if len(details.Devices) > 0 {
		topDevice := details.Devices[0]
		report["device_summary"] = fmt.Sprintf("你最常用的观看设备是%s，共使用%d次", topDevice.Name, topDevice.Count)
	}

	return &ViewingOverview{
		Details: details,
		Report:  report,
	}, nil
}

// ==================== 辅助函数 ====================

type kv struct {
	key   string
	value int
}

type authorEntry struct {
	name  string
	stats AuthorCompletionStat
}

func sortMapByValueDesc(m map[string]int) []kv {
	pairs := make([]kv, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kv{k, v})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].value > pairs[i].value {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	return pairs
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func roundFloat(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}

func sortByVideoCountDesc(list []authorEntry) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].stats.VideoCount > list[i].stats.VideoCount {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortByCompletionRateDesc(list []authorEntry) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].stats.AverageCompletionRate > list[i].stats.AverageCompletionRate {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortByComprehensiveScoreDesc(list []authorEntry) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].stats.ComprehensiveScore > list[i].stats.ComprehensiveScore {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortByQualityScoreDesc(list []authorEntry) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].stats.QualityScore > list[i].stats.QualityScore {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortIntsDesc(arr []int) {
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[j] > arr[i] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
}

func init() {
	_ = utils.Now
}
