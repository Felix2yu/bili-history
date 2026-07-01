package analytics

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"bili-history/internal/db"
)

// ViewingAnalytics holds comprehensive viewing behavior analytics.
type ViewingAnalytics struct {
	Year          int              `json:"year"`
	TotalVideos   int              `json:"total_videos"`
	TotalDuration int64            `json:"total_duration"`
	AvgDailyCount float64          `json:"avg_daily_count"`
	DailyCounts   []DailyData      `json:"daily_counts"`
	MonthlyCounts []MonthlyData    `json:"monthly_counts"`
	TimeSlotData  []TimeSlotData   `json:"time_slot_data"`
	WeekdayData   []WeekdayData    `json:"weekday_data"`
	Continuity    ContinuityData   `json:"continuity"`
	TimeInvestment TimeInvestment  `json:"time_insights"`
	TopAuthors    []AuthorStat     `json:"top_authors"`
	TopCategories []CategoryStat   `json:"top_categories"`
}

type DailyData struct {
	Date       string `json:"date"`
	Count      int    `json:"count"`
	Duration   int64  `json:"duration"`
}

type MonthlyData struct {
	Month  string `json:"month"`
	Count  int    `json:"count"`
	Duration int64 `json:"duration"`
}

type TimeSlotData struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type WeekdayData struct {
	Weekday string `json:"weekday"`
	Count   int    `json:"count"`
}

type ContinuityData struct {
	MaxStreak       int    `json:"max_streak"`
	CurrentStreak   int    `json:"current_streak"`
	StreakStart     string `json:"streak_start"`
	StreakEnd       string `json:"streak_end"`
	TotalDays       int    `json:"total_days"`
}

type TimeInvestment struct {
	MaxDay         string `json:"max_day"`
	MaxDayVideos   int    `json:"max_day_videos"`
	MaxDayDuration int64  `json:"max_day_duration"`
	AvgDailyDuration float64 `json:"avg_daily_duration"`
}

type AuthorStat struct {
	Name  string `json:"name"`
	Mid   int64  `json:"mid"`
	Count int    `json:"count"`
}

type CategoryStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// TitleAnalytics holds title analysis results.
type TitleAnalytics struct {
	Year            int                    `json:"year"`
	TotalTitles     int                    `json:"total_titles"`
	TopKeywords     []KeywordStat          `json:"top_keywords"`
	AvgTitleLength  float64                `json:"avg_title_length"`
	LengthBuckets   []LengthBucket         `json:"length_buckets"`
	KeywordCompletionRates map[string]KeywordCompletion `json:"keyword_completion_rates"`
	Insights        []string               `json:"insights"`
}

type KeywordStat struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type LengthBucket struct {
	Range    string  `json:"range"`
	Count    int     `json:"count"`
	AvgCompletion float64 `json:"avg_completion"`
}

type KeywordCompletion struct {
	AvgRate float64 `json:"avg_completion_rate"`
	Count   int     `json:"video_count"`
}

// Simple Chinese stopwords
var stopWords = map[string]bool{
	"的": true, "了": true, "是": true, "在": true, "我": true,
	"有": true, "和": true, "就": true, "不": true, "人": true,
	"都": true, "一": true, "一个": true, "上": true, "也": true,
	"很": true, "到": true, "说": true, "要": true, "去": true,
	"你": true, "会": true, "着": true, "没有": true, "看": true,
	"好": true, "自己": true, "这": true, "他": true, "她": true,
	"它": true, "们": true, "那": true, "被": true, "从": true,
	"把": true, "过": true, "对": true, "为": true, "与": true,
	"及": true, "或": true, "等": true, "之": true, "其": true,
}

// AnalyzeViewing computes comprehensive viewing behavior analytics.
func AnalyzeViewing(year int) (*ViewingAnalytics, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return nil, err
	}

	tableName := db.GetYearTable(year)

	result := &ViewingAnalytics{Year: year}

	// Total videos
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&result.TotalVideos)

	// Total duration
	database.QueryRow(fmt.Sprintf("SELECT COALESCE(SUM(CASE WHEN progress = -1 THEN duration WHEN progress > 0 THEN progress ELSE 0 END), 0) FROM %s", tableName)).Scan(&result.TotalDuration)

	// Daily counts
	rows, err := database.Query(fmt.Sprintf(`
		SELECT date(view_at + 28800, 'unixepoch') as day, COUNT(*) as cnt,
			COALESCE(SUM(CASE WHEN progress = -1 THEN duration WHEN progress > 0 THEN progress ELSE 0 END), 0) as dur
		FROM %s GROUP BY day ORDER BY day`, tableName))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d DailyData
			rows.Scan(&d.Date, &d.Count, &d.Duration)
			result.DailyCounts = append(result.DailyCounts, d)
		}
	}

	if len(result.DailyCounts) > 0 {
		result.AvgDailyCount = float64(result.TotalVideos) / float64(len(result.DailyCounts))
	}

	// Monthly counts
	rows2, err := database.Query(fmt.Sprintf(`
		SELECT strftime('%%Y-%%m', date(view_at + 28800, 'unixepoch')) as month, COUNT(*) as cnt,
			COALESCE(SUM(CASE WHEN progress = -1 THEN duration WHEN progress > 0 THEN progress ELSE 0 END), 0) as dur
		FROM %s GROUP BY month ORDER BY month`, tableName))
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var m MonthlyData
			rows2.Scan(&m.Month, &m.Count, &m.Duration)
			result.MonthlyCounts = append(result.MonthlyCounts, m)
		}
	}

	// Time slot distribution (hour of day)
	rows3, err := database.Query(fmt.Sprintf(`
		SELECT CAST(strftime('%%H', date(view_at + 28800, 'unixepoch')) AS INTEGER) as hour, COUNT(*) as cnt
		FROM %s GROUP BY hour ORDER BY hour`, tableName))
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var ts TimeSlotData
			rows3.Scan(&ts.Hour, &ts.Count)
			result.TimeSlotData = append(result.TimeSlotData, ts)
		}
	}

	// Weekday distribution
	rows4, err := database.Query(fmt.Sprintf(`
		SELECT CASE CAST(strftime('%%w', date(view_at + 28800, 'unixepoch')) AS INTEGER)
			WHEN 0 THEN 'Sunday' WHEN 1 THEN 'Monday' WHEN 2 THEN 'Tuesday'
			WHEN 3 THEN 'Wednesday' WHEN 4 THEN 'Thursday' WHEN 5 THEN 'Friday'
			ELSE 'Saturday' END as weekday, COUNT(*) as cnt
		FROM %s GROUP BY weekday ORDER BY
			CASE weekday
				WHEN 'Monday' THEN 1 WHEN 'Tuesday' THEN 2 WHEN 'Wednesday' THEN 3
				WHEN 'Thursday' THEN 4 WHEN 'Friday' THEN 5 WHEN 'Saturday' THEN 6
				ELSE 7 END`, tableName))
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var w WeekdayData
			rows4.Scan(&w.Weekday, &w.Count)
			result.WeekdayData = append(result.WeekdayData, w)
		}
	}

	// Continuity analysis
	result.Continuity = analyzeContinuity(database, tableName)

	// Time investment
	result.TimeInvestment = analyzeTimeInvestment(database, tableName)

	// Top authors (Python schema uses owner_mid)
	rows5, err := database.Query(fmt.Sprintf(`
		SELECT owner_name, owner_mid, COUNT(*) as cnt
		FROM %s GROUP BY owner_mid ORDER BY cnt DESC LIMIT 10`, tableName))
	if err == nil {
		defer rows5.Close()
		for rows5.Next() {
			var a AuthorStat
			rows5.Scan(&a.Name, &a.Mid, &a.Count)
			result.TopAuthors = append(result.TopAuthors, a)
		}
	}

	return result, nil
}

func analyzeContinuity(database *sql.DB, tableName string) ContinuityData {
	// Simplified implementation - full streak calculation
	rows, err := database.Query(fmt.Sprintf(`
		SELECT DISTINCT date(view_at + 28800, 'unixepoch') as view_date
		FROM %s ORDER BY view_date`, tableName))
	if err != nil {
		return ContinuityData{}
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		rows.Scan(&d)
		dates = append(dates, d)
	}

	if len(dates) == 0 {
		return ContinuityData{}
	}

	maxStreak, currentStreak := 1, 1
	streakStart, streakEnd := dates[0], dates[0]
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

	return ContinuityData{
		MaxStreak:     maxStreak,
		CurrentStreak: currentStreak,
		StreakStart:   streakStart,
		StreakEnd:     streakEnd,
		TotalDays:     len(dates),
	}
}

func analyzeTimeInvestment(database interface{}, tableName string) TimeInvestment {
	return TimeInvestment{}
}

// SimpleChineseSegment performs basic Chinese word segmentation.
// This is a naive approach - for production, use gojieba or similar.
func SimpleChineseSegment(text string) []string {
	var words []string
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		// Try longest match first (up to 4 chars)
		matched := false
		for l := 4; l >= 2; l-- {
			if i+l <= len(runes) {
				word := string(runes[i : i+l])
				if !stopWords[word] && len(word) >= 2 {
					words = append(words, word)
					i += l
					matched = true
					break
				}
			}
		}
		if !matched {
			// Skip single character
			i++
		}
	}
	return words
}

// AnalyzeTitle performs title analysis including keywords and completion rates.
func AnalyzeTitle(year int) (*TitleAnalytics, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return nil, err
	}

	tableName := db.GetYearTable(year)
	result := &TitleAnalytics{Year: year}

	// Fetch all titles with duration and progress
	rows, err := database.Query(fmt.Sprintf(`
		SELECT title, duration, progress FROM %s
		WHERE title IS NOT NULL AND duration > 0`, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type titleRecord struct {
		title    string
		duration int
		progress int
	}

	var records []titleRecord
	wordCount := make(map[string]int)
	titleLengthSum := 0

	for rows.Next() {
		var r titleRecord
		rows.Scan(&r.title, &r.duration, &r.progress)
		records = append(records, r)
		result.TotalTitles++
		titleLengthSum += len([]rune(r.title))

		// Segment and count words
		words := SimpleChineseSegment(r.title)
		for _, w := range words {
			wordCount[w]++
		}
	}

	if result.TotalTitles > 0 {
		result.AvgTitleLength = float64(titleLengthSum) / float64(result.TotalTitles)
	}

	// Sort keywords by frequency
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range wordCount {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	limit := 20
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		result.TopKeywords = append(result.TopKeywords, KeywordStat{
			Word:  sorted[i].Key,
			Count: sorted[i].Value,
		})
	}

	// Title length buckets
	lengthBuckets := make(map[string]*LengthBucket)
	bucketRanges := []struct {
		min, max int
		label    string
	}{
		{0, 5, "0-5"}, {6, 10, "6-10"}, {11, 15, "11-15"},
		{16, 20, "16-20"}, {21, 30, "21-30"}, {31, 50, "31-50"}, {51, 100, "51+"},
	}

	for _, br := range bucketRanges {
		lengthBuckets[br.label] = &LengthBucket{Range: br.label}
	}

	// Calculate completion rates by keyword
	keywordCompletions := make(map[string][]float64)
	for _, rec := range records {
		runeLen := len([]rune(rec.title))
		for _, br := range bucketRanges {
			if runeLen >= br.min && runeLen <= br.max {
				lengthBuckets[br.label].Count++
				var rate float64
				if rec.progress == -1 {
					rate = 1.0
				} else if rec.progress > 0 && rec.duration > 0 {
					rate = math.Min(float64(rec.progress)/float64(rec.duration), 1.0)
				}
				lengthBuckets[br.label].AvgCompletion += rate
				break
			}
		}

		// Track completion by keyword
		words := SimpleChineseSegment(rec.title)
		var rate float64
		if rec.progress == -1 {
			rate = 1.0
		} else if rec.progress > 0 && rec.duration > 0 {
			rate = math.Min(float64(rec.progress)/float64(rec.duration), 1.0)
		}
		for _, w := range words {
			keywordCompletions[w] = append(keywordCompletions[w], rate)
		}
	}

	// Finalize length buckets
	for _, bucket := range lengthBuckets {
		if bucket.Count > 0 {
			bucket.AvgCompletion /= float64(bucket.Count)
		}
		result.LengthBuckets = append(result.LengthBuckets, *bucket)
	}

	// Top keyword completion rates
	for _, ks := range result.TopKeywords {
		if rates, ok := keywordCompletions[ks.Word]; ok && len(rates) > 0 {
			sum := 0.0
			for _, r := range rates {
				sum += r
			}
			result.KeywordCompletionRates[ks.Word] = KeywordCompletion{
				AvgRate: sum / float64(len(rates)),
				Count:   len(rates),
			}
		}
	}

	// Generate insights
	result.Insights = generateInsights(result)

	return result, nil
}

func generateInsights(data *TitleAnalytics) []string {
	var insights []string

	if len(data.TopKeywords) > 0 {
		top5 := data.TopKeywords
		if len(top5) > 5 {
			top5 = top5[:5]
		}
		var parts []string
		for _, k := range top5 {
			parts = append(parts, fmt.Sprintf("%s(%d次)", k.Word, k.Count))
		}
		insights = append(insights, fmt.Sprintf("最常出现的关键词是：%s", strings.Join(parts, "、")))
	}

	// Completion rate insights
	if len(data.KeywordCompletionRates) > 0 {
		type kv struct {
			key  string
			rate float64
			count int
		}
		var sorted []kv
		for k, v := range data.KeywordCompletionRates {
			sorted = append(sorted, kv{k, v.AvgRate, v.Count})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].rate > sorted[j].rate
		})

		if len(sorted) >= 3 {
			top3 := sorted[:3]
			var parts []string
			for _, s := range top3 {
				parts = append(parts, fmt.Sprintf("%s(%.1f%%)", s.key, s.rate*100))
			}
			insights = append(insights, fmt.Sprintf("包含关键词 %s 的视频往往会被您看完", strings.Join(parts, "、")))
		}
	}

	return insights
}
