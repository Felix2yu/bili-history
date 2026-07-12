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
	Name     string `json:"name"`
	Mid      int64  `json:"mid"`
	Count    int    `json:"count"`
	Duration int    `json:"duration"`
}

type ReportSummary struct {
	TotalVideos      int                   `json:"total_videos"`
	TotalDuration    int                   `json:"total_duration"`
	UniqueDays       int                   `json:"unique_days"`
	UniqueAuthors    int                   `json:"unique_authors"`
	AvgDailyVideos   float64               `json:"avg_daily_videos"`
	AvgDailyDuration float64               `json:"avg_daily_duration"`
	TopCategories    []CategoryStat        `json:"top_categories"`
	TopAuthors       []AuthorStat          `json:"top_authors"`
	DeviceDist       map[string]int        `json:"device_dist"`
	DailyBreakdown   []DailyBreakdown      `json:"daily_breakdown"`
	HourDist         map[int]int           `json:"hour_dist"`
	CompletionStats  ReportCompletionStats `json:"completion_stats"`
	TopVideos        []ReportVideo         `json:"top_videos"`
	RewatchStats     ReportRewatchStats    `json:"rewatch_stats"`
	CompletionDist   []CompletionDistItem  `json:"completion_dist"`
	DurationPref     DurationPref          `json:"duration_pref"`
	LateNightRatio   float64               `json:"late_night_ratio"`
	FavoriteRate     float64               `json:"favorite_rate"`
	NewUpCount       int                   `json:"new_up_count"`
	TopTimeSlots     []TimeSlotStat        `json:"top_time_slots"`
	TitleKeywords    []KeywordStat         `json:"title_keywords"`
	AbandonRate      float64               `json:"abandon_rate"`
	GoldenSlotRatio  float64               `json:"golden_slot_ratio"`
	UpDiversity      float64               `json:"up_diversity"`
	WeekdayDist      []WeekdayStat         `json:"weekday_dist"`
}

type DailyBreakdown struct {
	Date     string `json:"date"`
	Count    int    `json:"count"`
	Duration int    `json:"duration"`
	UniqueUp int    `json:"unique_up"`
}

type ReportCompletionStats struct {
	Finished int     `json:"finished"`
	Partial  int     `json:"partial"`
	AvgRate  float64 `json:"avg_rate"`
}

type ReportRewatchStats struct {
	TotalRewatched  int            `json:"total_rewatched"`
	RewatchedVideos []RewatchVideo `json:"rewatched_videos"`
}

type RewatchVideo struct {
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Bvid       string `json:"bvid"`
	AuthorName string `json:"author_name"`
	Count      int    `json:"count"`
	TotalDur   int    `json:"total_duration"`
}

type CompletionDistItem struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

type DurationPref struct {
	Short      int     `json:"short"`
	Mid        int     `json:"mid"`
	Long       int     `json:"long"`
	ShortRatio float64 `json:"short_ratio"`
	MidRatio   float64 `json:"mid_ratio"`
	LongRatio  float64 `json:"long_ratio"`
}

type TimeSlotStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type KeywordStat struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type WeekdayStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
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

func GetMonthlyReport(year, month int) (*MonthlyReportResponse, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)
	videos, err := queryReportVideos(startDate, endDate)
	if err != nil {
		return nil, err
	}
	dayCount := int(endDate.Sub(startDate).Hours() / 24)
	summary := computeSummary(videos, dayCount)
	return &MonthlyReportResponse{
		Year:    year,
		Month:   month,
		Summary: summary,
		Videos:  videos,
	}, nil
}

func getWeekRange(year, week int) (time.Time, time.Time) {
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.Local)
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	week1Monday := jan4.AddDate(0, 0, -weekday+1)
	startDate := week1Monday.AddDate(0, 0, (week-1)*7)
	endDate := startDate.AddDate(0, 0, 6).Add(time.Hour*24 - time.Second)
	return startDate, endDate
}

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

func ptrInt(v int) *int { return &v }

func computeSummary(videos []ReportVideo, dayCount int) ReportSummary {
	if len(videos) == 0 {
		return ReportSummary{
			DeviceDist: make(map[string]int),
			HourDist:   make(map[int]int),
		}
	}

	summary := ReportSummary{
		TotalVideos: len(videos),
		DeviceDist:  make(map[string]int),
		HourDist:    make(map[int]int),
	}

	authorMap := make(map[string]*AuthorStat)
	categoryMap := make(map[string]int)
	uniqueAuthorMids := make(map[int64]bool)
	type dayInfo struct {
		count    int
		duration int
		authors  map[int64]bool
	}
	dailyMap := make(map[string]*dayInfo)
	rewatchMap := make(map[string][]ReportVideo)
	compBuckets := map[string]int{"0-20%": 0, "20-40%": 0, "40-60%": 0, "60-80%": 0, "80-100%": 0, "看完": 0}

	var totalCompletionRate float64
	var completionCount int
	var lateNightCount int
	var favCount int

	for _, v := range videos {
		summary.TotalDuration += v.Duration
		day := time.Unix(v.ViewAt, 0).Format("2006-01-02")
		if _, ok := dailyMap[day]; !ok {
			dailyMap[day] = &dayInfo{authors: make(map[int64]bool)}
		}
		dailyMap[day].count++
		dailyMap[day].duration += v.Duration
		dailyMap[day].authors[v.AuthorMid] = true

		uniqueAuthorMids[v.AuthorMid] = true
		if stat, ok := authorMap[v.AuthorName]; ok {
			stat.Count++
			stat.Duration += v.Duration
		} else {
			authorMap[v.AuthorName] = &AuthorStat{Name: v.AuthorName, Mid: v.AuthorMid, Count: 1, Duration: v.Duration}
		}

		cat := v.MainCategory
		if cat == "" {
			cat = v.TagName
		}
		if cat != "" {
			categoryMap[cat]++
		}

		summary.DeviceDist[getDeviceName(v.Dt)]++
		hour := time.Unix(v.ViewAt, 0).Hour()
		summary.HourDist[hour]++

		if hour >= 22 || hour < 6 {
			lateNightCount++
		}
		if v.IsFinish == 1 {
			favCount++
		}

		rewatchMap[v.Bvid] = append(rewatchMap[v.Bvid], v)

		if v.Duration > 0 {
			var rate float64
			if v.Progress == -1 {
				rate = 1.0
			} else {
				rate = float64(v.Progress) / float64(v.Duration)
			}
			if rate >= 0.9 {
				summary.CompletionStats.Finished++
			} else {
				summary.CompletionStats.Partial++
			}
			totalCompletionRate += rate
			completionCount++

			pct := rate * 100
			switch {
			case rate >= 1.0:
				compBuckets["看完"]++
			case pct >= 80:
				compBuckets["80-100%"]++
			case pct >= 60:
				compBuckets["60-80%"]++
			case pct >= 40:
				compBuckets["40-60%"]++
			case pct >= 20:
				compBuckets["20-40%"]++
			default:
				compBuckets["0-20%"]++
			}
		}

		if v.Duration < 300 {
			summary.DurationPref.Short++
		} else if v.Duration < 1200 {
			summary.DurationPref.Mid++
		} else {
			summary.DurationPref.Long++
		}
	}

	summary.UniqueDays = len(dailyMap)
	summary.UniqueAuthors = len(uniqueAuthorMids)
	summary.NewUpCount = len(uniqueAuthorMids)

	if completionCount > 0 {
		summary.CompletionStats.AvgRate = totalCompletionRate / float64(completionCount)
	}
	if summary.TotalVideos > 0 {
		summary.LateNightRatio = float64(lateNightCount) / float64(summary.TotalVideos)
		summary.FavoriteRate = float64(favCount) / float64(summary.TotalVideos)
	}
	if dayCount > 0 {
		summary.AvgDailyVideos = float64(summary.TotalVideos) / float64(dayCount)
		summary.AvgDailyDuration = float64(summary.TotalDuration) / float64(dayCount)
	}

	total := summary.DurationPref.Short + summary.DurationPref.Mid + summary.DurationPref.Long
	if total > 0 {
		summary.DurationPref.ShortRatio = float64(summary.DurationPref.Short) / float64(total)
		summary.DurationPref.MidRatio = float64(summary.DurationPref.Mid) / float64(total)
		summary.DurationPref.LongRatio = float64(summary.DurationPref.Long) / float64(total)
	}

	compOrder := map[string]int{"0-20%": 0, "20-40%": 1, "40-60%": 2, "60-80%": 3, "80-100%": 4, "看完": 5}
	for label, count := range compBuckets {
		if count > 0 {
			summary.CompletionDist = append(summary.CompletionDist, CompletionDistItem{Range: label, Count: count})
		}
	}
	sort.Slice(summary.CompletionDist, func(i, j int) bool {
		return compOrder[summary.CompletionDist[i].Range] < compOrder[summary.CompletionDist[j].Range]
	})

	for bvid, vids := range rewatchMap {
		if len(vids) > 1 {
			totalDur := 0
			for _, v := range vids {
				totalDur += v.Duration
			}
			summary.RewatchStats.RewatchedVideos = append(summary.RewatchStats.RewatchedVideos, RewatchVideo{
				Title: vids[0].Title, Cover: vids[0].Cover, Bvid: bvid,
				AuthorName: vids[0].AuthorName, Count: len(vids), TotalDur: totalDur,
			})
			summary.RewatchStats.TotalRewatched += len(vids)
		}
	}
	sort.Slice(summary.RewatchStats.RewatchedVideos, func(i, j int) bool {
		return summary.RewatchStats.RewatchedVideos[i].Count > summary.RewatchStats.RewatchedVideos[j].Count
	})
	if len(summary.RewatchStats.RewatchedVideos) > 5 {
		summary.RewatchStats.RewatchedVideos = summary.RewatchStats.RewatchedVideos[:5]
	}

	slotCounts := map[string]int{"早晨(6-12)": 0, "下午(12-18)": 0, "晚上(18-22)": 0, "深夜(22-6)": 0}
	for _, v := range videos {
		hour := time.Unix(v.ViewAt, 0).Hour()
		switch {
		case hour >= 6 && hour < 12:
			slotCounts["早晨(6-12)"]++
		case hour >= 12 && hour < 18:
			slotCounts["下午(12-18)"]++
		case hour >= 18 && hour < 22:
			slotCounts["晚上(18-22)"]++
		default:
			slotCounts["深夜(22-6)"]++
		}
	}
	for name, count := range slotCounts {
		summary.TopTimeSlots = append(summary.TopTimeSlots, TimeSlotStat{Name: name, Count: count})
	}
	sort.Slice(summary.TopTimeSlots, func(i, j int) bool {
		return summary.TopTimeSlots[i].Count > summary.TopTimeSlots[j].Count
	})

	for date, info := range dailyMap {
		summary.DailyBreakdown = append(summary.DailyBreakdown, DailyBreakdown{
			Date: date, Count: info.count, Duration: info.duration, UniqueUp: len(info.authors),
		})
	}
	sort.Slice(summary.DailyBreakdown, func(i, j int) bool {
		return summary.DailyBreakdown[i].Date < summary.DailyBreakdown[j].Date
	})

	for name, count := range categoryMap {
		summary.TopCategories = append(summary.TopCategories, CategoryStat{Name: name, Count: count})
	}
	sort.Slice(summary.TopCategories, func(i, j int) bool {
		return summary.TopCategories[i].Count > summary.TopCategories[j].Count
	})
	if len(summary.TopCategories) > 10 {
		summary.TopCategories = summary.TopCategories[:10]
	}

	for _, stat := range authorMap {
		summary.TopAuthors = append(summary.TopAuthors, *stat)
	}
	sort.Slice(summary.TopAuthors, func(i, j int) bool {
		return summary.TopAuthors[i].Count > summary.TopAuthors[j].Count
	})
	if len(summary.TopAuthors) > 10 {
		summary.TopAuthors = summary.TopAuthors[:10]
	}

	sortedVideos := make([]ReportVideo, len(videos))
	copy(sortedVideos, videos)
	sort.Slice(sortedVideos, func(i, j int) bool {
		return sortedVideos[i].Duration > sortedVideos[j].Duration
	})
	if len(sortedVideos) > 5 {
		sortedVideos = sortedVideos[:5]
	}
	summary.TopVideos = sortedVideos

	// Title keywords: simple Chinese word extraction
	keywordMap := make(map[string]int)
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true, "有": true, "和": true,
		"就": true, "不": true, "人": true, "都": true, "一": true, "一个": true, "上": true,
		"也": true, "很": true, "到": true, "说": true, "要": true, "去": true, "你": true,
		"会": true, "着": true, "没有": true, "看": true, "好": true, "自己": true, "这": true,
		"他": true, "她": true, "它": true, "么": true, "被": true, "从": true, "把": true,
		"那": true, "吗": true, "吧": true, "啊": true, "呢": true, "还": true, "但": true,
		"与": true, "及": true, "或": true, "等": true, "之": true, "对": true, "为": true,
		"中": true, "大": true, "小": true, "新": true, "全": true, "最": true, "更": true,
		"【": true, "】": true, "|": true, "-": true, " ": true,
	}
	for _, v := range videos {
		title := v.Title
		// Extract 2-4 char segments as keywords
		runes := []rune(title)
		for i := 0; i < len(runes); i++ {
			for l := 2; l <= 4 && i+l <= len(runes); l++ {
				word := string(runes[i : i+l])
				// Skip if contains non-Chinese chars or is stop word
				allChinese := true
				for _, r := range word {
					if r < 0x4e00 || r > 0x9fff {
						allChinese = false
						break
					}
				}
				if allChinese && !stopWords[word] && l >= 2 {
					keywordMap[word]++
				}
			}
		}
	}
	for word, count := range keywordMap {
		if count >= 2 {
			summary.TitleKeywords = append(summary.TitleKeywords, KeywordStat{Word: word, Count: count})
		}
	}
	sort.Slice(summary.TitleKeywords, func(i, j int) bool {
		return summary.TitleKeywords[i].Count > summary.TitleKeywords[j].Count
	})
	if len(summary.TitleKeywords) > 20 {
		summary.TitleKeywords = summary.TitleKeywords[:20]
	}

	// Abandon rate: videos with progress < 20% of duration
	var abandonCount int
	for _, v := range videos {
		if v.Duration > 0 && v.Progress >= 0 {
			rate := float64(v.Progress) / float64(v.Duration)
			if rate < 0.2 {
				abandonCount++
			}
		}
	}
	if summary.TotalVideos > 0 {
		summary.AbandonRate = float64(abandonCount) / float64(summary.TotalVideos)
	}

	// Golden slot concentration: top 3 consecutive hours ratio
	hourTotals := make([]int, 24)
	for _, v := range videos {
		hour := time.Unix(v.ViewAt, 0).Hour()
		hourTotals[hour]++
	}
	maxConsecutive := 0
	for start := 0; start < 24; start++ {
		sum := 0
		for h := 0; h < 3; h++ {
			sum += hourTotals[(start+h)%24]
		}
		if sum > maxConsecutive {
			maxConsecutive = sum
		}
	}
	if summary.TotalVideos > 0 {
		summary.GoldenSlotRatio = float64(maxConsecutive) / float64(summary.TotalVideos)
	}

	// UP主 diversity: unique authors / total videos
	if summary.TotalVideos > 0 {
		summary.UpDiversity = float64(summary.UniqueAuthors) / float64(summary.TotalVideos)
	}

	// Weekday distribution (Monday to Sunday)
	weekdayNames := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	weekdayIndices := []int{1, 2, 3, 4, 5, 6, 0} // Go: Mon=1..Sat=6, Sun=0
	weekdayCounts := make([]int, 7)
	for _, v := range videos {
		wd := time.Unix(v.ViewAt, 0).Weekday()
		weekdayCounts[wd]++
	}
	for i, name := range weekdayNames {
		summary.WeekdayDist = append(summary.WeekdayDist, WeekdayStat{Name: name, Count: weekdayCounts[weekdayIndices[i]]})
	}

	return summary
}

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
