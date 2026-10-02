package routers

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"bilibili-history-go/database"
)

// The viewing fixtures live in the reserved year 2013 and every seeded bvid is
// prefixed with BVFTVW, so no other test file can move these aggregate counts.
const (
	ftViewYear        = 2013
	ftViewMissingYear = 1998
)

// ftViewEndpoints are the RegisterViewingRoutes (viewing.go:14) handlers that
// resolve their year through getYearFromQuery (viewing.go:36).
var ftViewEndpoints = []string{
	"/api/viewing/monthly-stats",
	"/api/viewing/weekly-stats",
	"/api/viewing/time-slots",
	"/api/viewing/continuity",
	"/api/viewing/",
	"/api/viewing/watch-counts",
	"/api/viewing/completion-rates",
	"/api/viewing/author-completion",
	"/api/viewing/tag-analysis",
	"/api/viewing/duration-analysis",
}

// ftSeedViewingHistory writes six deterministic rows into year 2013:
//
//	FTV1 2013-01-05 12:00 科技 FTUP甲(400001) duration 200 progress 200  video
//	FTV2 2013-01-05 23:30 科技 FTUP甲(400001) duration 100 progress  50  video
//	FTV3 2013-01-05 15:00 科技 FTUP甲(400001) duration  50 progress   0  video
//	FTV4 2013-01-06 08:00 音乐 FTUP乙(400002) duration 300 progress  -1  pgc
//	FTV4 2013-01-06 20:00 音乐 FTUP乙(400002) duration 300 progress 150  pgc (repeat)
//	FTV5 2013-02-10 03:00      FTUP丙(400003) duration  60 progress  30  live
//
// That gives 6 records / 730 watched seconds / 3 authors / 3 active days, one
// rewatched bvid and an unambiguous busiest day (Jan 5th, three records).
func ftSeedViewingHistory(t *testing.T) map[string]int64 {
	t.Helper()
	ts := map[string]int64{
		"FTV1": viewAt(t, ftViewYear, time.January, 5, 12, 0),
		"FTV2": viewAt(t, ftViewYear, time.January, 5, 23, 30),
		"FTV3": viewAt(t, ftViewYear, time.January, 5, 15, 0),
		"FTV4": viewAt(t, ftViewYear, time.January, 6, 8, 0),
		"FTV5": viewAt(t, ftViewYear, time.February, 10, 3, 0),
	}
	rows := []map[string]interface{}{
		{"bvid": "BVFTVW00000001", "title": "FT观看样本一", "tag_name": "科技", "main_category": "科技数码",
			"author_name": "FTUP甲", "author_mid": 400001, "view_at": ts["FTV1"], "duration": 200, "progress": 200, "business": "video"},
		{"bvid": "BVFTVW00000002", "title": "FT观看样本二", "tag_name": "科技", "main_category": "科技数码",
			"author_name": "FTUP甲", "author_mid": 400001, "view_at": ts["FTV2"], "duration": 100, "progress": 50, "business": "video"},
		{"bvid": "BVFTVW00000003", "title": "FT观看样本三", "tag_name": "科技", "main_category": "科技数码",
			"author_name": "FTUP甲", "author_mid": 400001, "view_at": ts["FTV3"], "duration": 50, "progress": 0, "business": "video"},
		{"bvid": "BVFTVW00000004", "title": "FT观看样本四", "tag_name": "音乐", "main_category": "音乐",
			"author_name": "FTUP乙", "author_mid": 400002, "view_at": ts["FTV4"], "duration": 300, "progress": -1, "business": "pgc"},
		// A second view of the same bvid drives the rewatch analysis.
		{"bvid": "BVFTVW00000004", "title": "FT观看样本四", "tag_name": "音乐", "main_category": "音乐",
			"author_name": "FTUP乙", "author_mid": 400002, "view_at": viewAt(t, ftViewYear, time.January, 6, 20, 0), "duration": 300, "progress": 150, "business": "pgc"},
		{"bvid": "BVFTVW00000005", "title": "FT直播样本", "tag_name": "", "main_category": "",
			"author_name": "FTUP丙", "author_mid": 400003, "view_at": ts["FTV5"], "duration": 60, "progress": 30, "business": "live"},
	}
	for _, row := range rows {
		insertHistory(t, ftViewYear, row)
	}
	return ts
}

func TestFTViewingYearGuards(t *testing.T) {
	ftSeedViewingHistory(t)
	e := newAPI(t, RegisterViewingRoutes)

	// A non-numeric year is rejected with 400 by every year-driven endpoint
	// (viewing.go:47-51).
	for _, path := range ftViewEndpoints {
		resp := expectStatus(t, e, "GET", path+"?year=abc", "", 400, "error")
		expectMessage(t, resp, "无效的年份参数")
	}

	// A numeric year with no history table fails the membership check
	// (viewing.go:60-63) with a 200 error envelope.
	for _, path := range ftViewEndpoints {
		resp := expectStatus(t, e, "GET", path+"?year=1888", "", 200, "error")
		expectMessage(t, resp, "未找到指定年份的历史记录数据")
	}

	// No year at all selects the newest available table, and every payload
	// echoes both the resolved year and the full year list (viewing.go:67).
	latest := expectStatus(t, e, "GET", "/api/viewing/monthly-stats", "", 200, "success").dataMap(t)
	years := dataArray(t, latest["available_years"])
	if len(years) == 0 {
		t.Fatalf("available_years empty: %v", latest)
	}
	if fmt.Sprint(latest["year"]) != fmt.Sprint(years[0]) {
		t.Fatalf("default year = %v, want the first available year %v", latest["year"], years[0])
	}
	if !ftHasYear(years, ftViewYear) {
		t.Fatalf("available_years %v missing the seeded %d table", years, ftViewYear)
	}

	// getYearFromQuery's "未找到任何历史记录数据" branch (viewing.go:41-44) is
	// unreachable here: SQLiteDB.GetAvailableYears (database/sqlite.go:116)
	// always falls back to the current year, so it never returns an empty list.
}

func TestFTViewingAnalyzerErrors(t *testing.T) {
	ftSeedViewingHistory(t)
	e := newAPI(t, RegisterViewingRoutes)

	// bilibili_history_1998zz parses as the year 1998 in GetAvailableYears
	// (fmt.Sscanf stops at the trailing letters), but TableExists looks up the
	// exact name bilibili_history_1998 - which does not exist. That mismatch is
	// what drives the `err != nil` branch of every handler.
	name := fmt.Sprintf("bilibili_history_%dzz", ftViewMissingYear)
	if _, err := db(t).Exec("CREATE TABLE IF NOT EXISTS " + name + " (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create sentinel table: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db(t).Exec("DROP TABLE IF EXISTS " + name); err != nil {
			t.Errorf("drop sentinel table: %v", err)
		}
	})

	for _, path := range ftViewEndpoints {
		resp := expectStatus(t, e, "GET", path+"?year=1998", "", 200, "error")
		expectMessageContains(t, resp, "未找到 1998 年的历史记录数据")
	}
}

func TestFTViewingMonthlyStats(t *testing.T) {
	ftSeedViewingHistory(t)
	e := newAPI(t, RegisterViewingRoutes)

	data := expectStatus(t, e, "GET", fmt.Sprintf("/api/viewing/monthly-stats?year=%d", ftViewYear), "", 200, "success").dataMap(t)
	if fmt.Sprint(data["year"]) != fmt.Sprint(ftViewYear) {
		t.Fatalf("year = %v, want %d", data["year"], ftViewYear)
	}
	num := func(key string) int { return int(data[key].(float64)) }
	if num("total_videos") != 6 {
		t.Fatalf("total_videos = %d, want the six seeded rows", num("total_videos"))
	}
	// progress = -1 falls back to duration: 200+50+0+300+150+30.
	if num("total_duration") != 730 {
		t.Fatalf("total_duration = %d, want 730", num("total_duration"))
	}
	if num("unique_authors") != 3 {
		t.Fatalf("unique_authors = %d, want 3", num("unique_authors"))
	}
	if num("active_days") != 3 {
		t.Fatalf("active_days = %d, want 3", num("active_days"))
	}

	monthly := dataArray(t, data["monthly_stats"])
	if len(monthly) != 2 {
		t.Fatalf("monthly_stats = %v, want only January and February", monthly)
	}
	jan := monthly[0].(map[string]interface{})
	if jan["month"] != "01" || int(jan["total_count"].(float64)) != 5 {
		t.Fatalf("January bucket = %v, want month 01 with five rows", jan)
	}
	// Only FTV1 (200/200) reaches the >=90% completion threshold.
	if int(jan["completed_videos"].(float64)) != 1 {
		t.Fatalf("January completed_videos = %v, want 1", jan["completed_videos"])
	}
	if feb := monthly[1].(map[string]interface{}); feb["month"] != "02" || int(feb["total_count"].(float64)) != 1 {
		t.Fatalf("February bucket = %v, want month 02 with one row", monthly[1])
	}

	tags := dataArray(t, data["tag_ranking"])
	if len(tags) != 2 {
		t.Fatalf("tag_ranking = %v, want 科技 and 音乐", tags)
	}
	if first := tags[0].(map[string]interface{}); first["tag_name"] != "科技" || int(first["count"].(float64)) != 3 {
		t.Fatalf("top tag = %v, want 科技 with three rows", first)
	}
	authors := dataArray(t, data["author_ranking"])
	if len(authors) != 3 {
		t.Fatalf("author_ranking = %v, want the three seeded UP hosts", authors)
	}
	if top := authors[0].(map[string]interface{}); top["author_name"] != "FTUP甲" || int(top["count"].(float64)) != 3 {
		t.Fatalf("top author = %v, want FTUP甲 with three rows", top)
	}
	if len(dataArray(t, data["top_videos"])) == 0 {
		t.Fatal("top_videos should list the busiest rows")
	}
	if len(dataArray(t, data["weekday_distribution"])) == 0 {
		t.Fatal("weekday_distribution should be populated")
	}
	if _, ok := data["duration_preference"].(map[string]interface{}); !ok {
		t.Fatalf("duration_preference = %v, want an object", data["duration_preference"])
	}
	if _, ok := data["device_distribution"].(map[string]interface{}); !ok {
		t.Fatalf("device_distribution = %v, want an object", data["device_distribution"])
	}
	if len(dataArray(t, data["title_keywords"])) == 0 {
		t.Fatal("title_keywords should extract the repeated FT token")
	}
}

func TestFTViewingWeeklySlotsAndContinuity(t *testing.T) {
	ftSeedViewingHistory(t)
	e := newAPI(t, RegisterViewingRoutes)

	// getWeeklyStats -> database.AnalyzeWeekly
	weekly := expectStatus(t, e, "GET", fmt.Sprintf("/api/viewing/weekly-stats?year=%d", ftViewYear), "", 200, "success").dataMap(t)
	stats := weekly["weekly_stats"].(map[string]interface{})
	if len(stats) != 7 {
		t.Fatalf("weekly_stats = %v, want all seven weekdays", stats)
	}
	// 2013-01-05 is a Saturday, 01-06 and 02-10 are Sundays.
	saturday := time.Date(ftViewYear, time.January, 5, 12, 0, 0, 0, time.Local).Weekday()
	sunday := time.Date(ftViewYear, time.January, 6, 8, 0, 0, 0, time.Local).Weekday()
	if int(stats[ftWeekdayName(saturday)].(float64)) != 3 {
		t.Fatalf("weekly_stats[%s] = %v, want 3", ftWeekdayName(saturday), stats[ftWeekdayName(saturday)])
	}
	if int(stats[ftWeekdayName(sunday)].(float64)) != 3 {
		t.Fatalf("weekly_stats[%s] = %v, want 3", ftWeekdayName(sunday), stats[ftWeekdayName(sunday)])
	}
	if int(weekly["active_days"].(float64)) != 3 {
		t.Fatalf("active_days = %v, want 3", weekly["active_days"])
	}
	seasons := weekly["seasonal_patterns"].(map[string]interface{})
	spring, ok := seasons["春季"].(map[string]interface{})
	if !ok {
		t.Fatalf("seasonal_patterns = %v, want the 春季 bucket", seasons)
	}
	if int(spring["view_count"].(float64)) != 6 {
		t.Fatalf("春季 view_count = %v, want all six rows (January and February)", spring["view_count"])
	}
	if got := spring["avg_duration"].(float64); got < 121 || got > 122 {
		t.Fatalf("春季 avg_duration = %v, want 730/6", got)
	}

	// getTimeSlots -> database.AnalyzeTimeSlots
	slots := expectStatus(t, e, "GET", fmt.Sprintf("/api/viewing/time-slots?year=%d", ftViewYear), "", 200, "success").dataMap(t)
	daily := slots["daily_time_slots"].(map[string]interface{})
	if len(daily) != 6 {
		t.Fatalf("daily_time_slots = %v, want six distinct hours", daily)
	}
	if int(daily["12时"].(float64)) != 1 || int(daily["23时"].(float64)) != 1 {
		t.Fatalf("daily_time_slots = %v, want the seeded hours counted once each", daily)
	}
	peaks := dataArray(t, slots["peak_hours"])
	// The peak query is LIMIT 5, so five of the six tied hours are reported.
	if len(peaks) != 5 {
		t.Fatalf("peak_hours = %v, want five entries", peaks)
	}
	for _, p := range peaks {
		entry := p.(map[string]interface{})
		if int(entry["view_count"].(float64)) != 1 {
			t.Fatalf("peak hour %v, want every seeded hour counted once", entry)
		}
	}
	investment := slots["time_investment"].(map[string]interface{})
	maxDay := investment["max_duration_day"].(map[string]interface{})
	if maxDay["date"] != "2013-01-06" || int(maxDay["video_count"].(float64)) != 2 || int(maxDay["total_duration"].(float64)) != 450 {
		t.Fatalf("max_duration_day = %v, want 2013-01-06 with 2 rows and 450 seconds", maxDay)
	}
	if got := investment["avg_daily_duration"].(float64); got < 243 || got > 244 {
		t.Fatalf("avg_daily_duration = %v, want 730/3", got)
	}
	record := slots["max_daily_record"].(map[string]interface{})
	if record["date"] != "2013-01-05" || int(record["video_count"].(float64)) != 3 {
		t.Fatalf("max_daily_record = %v, want 2013-01-05 with three rows", record)
	}

	// An empty year still answers successfully with a zeroed streak.
	ensureYear(t, yearEmpty)
	empty := expectStatus(t, e, "GET", fmt.Sprintf("/api/viewing/continuity?year=%d", yearEmpty), "", 200, "success").dataMap(t)
	zero := empty["viewing_continuity"].(map[string]interface{})
	if int(zero["max_streak"].(float64)) != 0 {
		t.Fatalf("empty-year continuity = %v, want max_streak 0", zero)
	}

	// getContinuity -> database.AnalyzeContinuity over the seeded days.
	cont := expectStatus(t, e, "GET", fmt.Sprintf("/api/viewing/continuity?year=%d", ftViewYear), "", 200, "success").dataMap(t)
	streak := cont["viewing_continuity"].(map[string]interface{})
	if int(streak["max_streak"].(float64)) != 2 {
		t.Fatalf("max_streak = %v, want the 5th-6th of January pair", streak["max_streak"])
	}
	period := streak["longest_streak_period"].(map[string]interface{})
	if period["start"] != "2013-01-05" || period["end"] != "2013-01-06" {
		t.Fatalf("longest_streak_period = %v", period)
	}
	if int(streak["current_streak"].(float64)) != 1 || streak["current_streak_start"] != "2013-02-10" {
		t.Fatalf("current streak = %v, want 1 starting on the isolated February day", streak)
	}
}

func TestFTViewingBehaviourReports(t *testing.T) {
	ftSeedViewingHistory(t)
	e := newAPI(t, RegisterViewingRoutes)
	year := fmt.Sprintf("?year=%d", ftViewYear)

	// getViewingOverview -> database.GetViewingOverview
	over := expectStatus(t, e, "GET", "/api/viewing/"+year, "", 200, "success").dataMap(t)
	details := over["details"].(map[string]interface{})
	// 730 watched seconds over the year.
	if got := details["total_watch_hours"].(float64); got < 0.2 || got > 0.21 {
		t.Fatalf("total_watch_hours = %v, want 730/3600", got)
	}
	if int(details["total_days"].(float64)) != 3 {
		t.Fatalf("total_days = %v, want 3", details["total_days"])
	}
	if len(dataArray(t, details["top_categories"])) == 0 {
		t.Fatalf("top_categories = %v, want the seeded categories", details["top_categories"])
	}
	ups := dataArray(t, details["favorite_up_users"])
	if len(ups) == 0 {
		t.Fatal("favorite_up_users should list the seeded UP hosts")
	}
	// late_night_views is dead payload: GetViewingOverview seeds it with an
	// empty slice and never runs a query for it, so the seeded 23:30 and 03:00
	// rows are never reported.
	// PRODUCTION BUG: database/analysis.go:1779 initialises LateNightViews and
	// analysis.go:1776-1905 fills every other detail except it; a query over the
	// late hours (e.g. strftime('%H') >= 22 OR < 5) would populate the field the
	// frontend reads as details.late_night_views.
	if got := dataArray(t, details["late_night_views"]); len(got) != 0 {
		t.Fatalf("late_night_views = %v, want the always-empty slice", got)
	}
	if _, ok := details["time_slot_activity"].(map[string]interface{}); !ok {
		t.Fatalf("time_slot_activity = %v, want an object", details["time_slot_activity"])
	}
	if len(dataArray(t, details["devices"])) == 0 {
		t.Fatalf("devices = %v, want the seeded device buckets", details["devices"])
	}
	if _, ok := over["report"].(map[string]interface{}); !ok {
		t.Fatalf("report = %v, want the generated text summary", over["report"])
	}

	// getWatchCounts -> database.AnalyzeWatchCounts (one rewatched bvid).
	counts := expectStatus(t, e, "GET", "/api/viewing/watch-counts"+year, "", 200, "success").dataMap(t)
	rewatch := counts["watch_counts"].(map[string]interface{})["rewatch_stats"].(map[string]interface{})
	if int(rewatch["total_rewatched_videos"].(float64)) != 1 {
		t.Fatalf("total_rewatched_videos = %v, want BVFTVW00000004 only", rewatch["total_rewatched_videos"])
	}
	if int(rewatch["total_unique_videos"].(float64)) != 5 {
		t.Fatalf("total_unique_videos = %v, want five distinct bvids", rewatch["total_unique_videos"])
	}
	most := dataArray(t, counts["watch_counts"].(map[string]interface{})["most_watched_videos"])
	if len(most) == 0 || most[0].(map[string]interface{})["bvid"] != "BVFTVW00000004" {
		t.Fatalf("most_watched_videos = %v, want the twice watched row first", most)
	}
	if int(most[0].(map[string]interface{})["watch_count"].(float64)) != 2 {
		t.Fatalf("watch_count = %v, want 2", most[0].(map[string]interface{})["watch_count"])
	}

	// getCompletionRates -> database.AnalyzeCompletionRates
	rates := expectStatus(t, e, "GET", "/api/viewing/completion-rates"+year, "", 200, "success").dataMap(t)
	overall := rates["completion_rates"].(map[string]interface{})["overall_stats"].(map[string]interface{})
	if int(overall["total_videos"].(float64)) != 6 {
		t.Fatalf("overall total_videos = %v, want 6", overall["total_videos"])
	}
	// progress == duration (FTV1) and progress == -1 (the first FTV4 view) both
	// count as fully watched.
	if int(overall["fully_watched_count"].(float64)) != 2 {
		t.Fatalf("fully_watched_count = %v, want the 200/200 row plus the progress=-1 row", overall["fully_watched_count"])
	}
	if int(overall["not_started_count"].(float64)) != 1 {
		t.Fatalf("not_started_count = %v, want the progress=0 row", overall["not_started_count"])
	}
	dist := rates["completion_rates"].(map[string]interface{})["completion_distribution"].(map[string]interface{})
	if len(dist) != 6 {
		t.Fatalf("completion_distribution = %v, want the six fixed buckets", dist)
	}

	// getAuthorCompletion -> database.AnalyzeAuthorCompletion. Every author in
	// this fixture has at most three records, and the analyzer only scores
	// authors with >= 5 records (database/analysis.go:1326), so all four buckets
	// come back empty. TestFTViewingAuthorCompletion covers the scored path.
	author := expectStatus(t, e, "GET", "/api/viewing/author-completion"+year, "", 200, "success").dataMap(t)
	block := author["completion_rates"].(map[string]interface{})
	for _, key := range []string{"most_watched_authors", "highest_completion_authors", "most_valuable_authors", "potential_authors"} {
		entries, ok := block[key].(map[string]interface{})
		if !ok {
			t.Fatalf("author-completion payload missing %q: %v", key, block)
		}
		if len(entries) != 0 {
			t.Fatalf("%s = %v, want no author below the five-record gate", key, entries)
		}
	}

	// getTagAnalysis -> database.AnalyzeTagAnalysis
	tag := expectStatus(t, e, "GET", "/api/viewing/tag-analysis"+year, "", 200, "success").dataMap(t)
	tagDist := tag["watch_counts"].(map[string]interface{})["tag_distribution"].(map[string]interface{})
	if int(tagDist["科技"].(float64)) != 3 || int(tagDist["音乐"].(float64)) != 2 {
		t.Fatalf("tag_distribution = %v, want 科技:3 音乐:2", tagDist)
	}
	// Tags below the five-record gate are counted but never scored
	// (database/analysis.go:1517).
	tagRates := tag["completion_rates"].(map[string]interface{})["tag_completion_rates"].(map[string]interface{})
	if len(tagRates) != 0 {
		t.Fatalf("tag_completion_rates = %v, want no tag below five records", tagRates)
	}

	// getDurationAnalysis -> database.AnalyzeDurationAnalysis
	dur := expectStatus(t, e, "GET", "/api/viewing/duration-analysis"+year, "", 200, "success").dataMap(t)
	corr := dur["duration_correlation"].(map[string]interface{})
	if len(corr) == 0 {
		t.Fatal("duration_correlation should bucket the seeded durations")
	}
}

func TestFTViewingExtraAnalysis(t *testing.T) {
	e := newAPI(t, RegisterViewingRoutes)
	ftSeedExtraForViewing(t)

	// getLikesAnalysis -> database.AnalyzeLikes over the shared likes DB.
	likes := expectStatus(t, e, "GET", "/api/viewing/likes-analysis", "", 200, "success").dataMap(t)
	if int(likes["total_count"].(float64)) < 1 {
		t.Fatalf("likes total_count = %v, want the seeded row counted", likes["total_count"])
	}
	if !ftHasNamedEntry(t, likes["top_creators"], "FTUP观看") {
		t.Fatalf("likes top_creators = %v, want FTUP观看", likes["top_creators"])
	}
	if !ftHasNamedEntry(t, likes["category_dist"], "FT观看分区") {
		t.Fatalf("likes category_dist = %v", likes["category_dist"])
	}
	if len(dataArray(t, likes["duration_dist"])) == 0 {
		t.Fatalf("likes duration_dist = %v, want the seeded bucket", likes["duration_dist"])
	}

	// getFavoritesAnalysis -> database.AnalyzeFavorites.
	favs := expectStatus(t, e, "GET", "/api/viewing/favorites-analysis", "", 200, "success").dataMap(t)
	if int(favs["total_count"].(float64)) < 2 {
		t.Fatalf("favorites total_count = %v, want the two seeded contents", favs["total_count"])
	}
	if int(favs["folder_count"].(float64)) < 1 {
		t.Fatalf("favorites folder_count = %v, want the seeded folder", favs["folder_count"])
	}
	found := false
	for _, f := range dataArray(t, favs["folder_dist"]) {
		entry := f.(map[string]interface{})
		if entry["title"] == "FT观看收藏夹" {
			found = true
		}
	}
	if !found {
		t.Fatalf("favorites folder_dist = %v, want FT观看收藏夹", favs["folder_dist"])
	}
	if !ftHasNamedEntry(t, favs["top_creators"], "FTUP收藏") {
		t.Fatalf("favorites top_creators = %v, want FTUP收藏", favs["top_creators"])
	}

	// getWatchLaterAnalysis -> database.AnalyzeWatchLater.
	wl := expectStatus(t, e, "GET", "/api/viewing/watchlater-analysis", "", 200, "success").dataMap(t)
	if int(wl["total_count"].(float64)) < 2 {
		t.Fatalf("watchlater total_count = %v, want the two seeded rows", wl["total_count"])
	}
	if !ftHasNamedEntry(t, wl["category_dist"], "FT稍后再看") {
		t.Fatalf("watchlater category_dist = %v, want FT稍后再看", wl["category_dist"])
	}
	oldest := dataArray(t, wl["oldest_items"])
	if len(oldest) == 0 {
		t.Fatal("oldest_items should list the seeded rows")
	}
	first := oldest[0].(map[string]interface{})
	if first["title"] != "FT最早稍后再看" {
		t.Fatalf("oldest_items[0] = %v, want the row with the smallest add_at", first)
	}
	// days_ago is derived from time.Now(), so compute the expectation instead of
	// hard-coding a date.
	wantDays := int((time.Now().Unix() - viewAt(t, 2013, time.January, 1, 0, 0)) / 86400)
	if got := int(first["days_ago"].(float64)); got < wantDays-1 || got > wantDays+1 {
		t.Fatalf("oldest_items[0].days_ago = %d, want about %d", got, wantDays)
	}

	// getExtraOverview -> database.GetExtraStatsOverview (never fails).
	over := expectStatus(t, e, "GET", "/api/viewing/extra-overview", "", 200, "success").dataMap(t)
	for _, key := range []string{"likes_count", "watchlater_count", "favorites_count", "favorites_folder_count"} {
		if _, ok := over[key]; !ok {
			t.Fatalf("extra-overview payload missing %q: %v", key, over)
		}
		if int(over[key].(float64)) < 1 {
			t.Fatalf("extra-overview %s = %v, want the seeded rows", key, over[key])
		}
	}
}

// ftSeedExtraForViewing fills the likes / watchlater / favorites side DBs with
// rows that are unique to this file so the analysis payloads must mention them.
func ftSeedExtraForViewing(t *testing.T) {
	t.Helper()
	likes := database.GetLikesDB()
	if likes == nil {
		t.Skip("likes database unavailable")
	}
	ftExec(t, likes, `INSERT OR REPLACE INTO liked_videos
		(bvid, aid, title, owner_name, owner_mid, tname, duration, pubdate, view, danmaku, like_count, fetch_time)
		VALUES ('BVFTVWLK0000001', 20130001, 'FT观看点赞样本', 'FTUP观看', 410001, 'FT观看分区', 120, 1357000000, 1000, 10, 5, 1357000000)`)

	wl := database.GetWatchLaterDB()
	if wl == nil {
		t.Skip("watchlater database unavailable")
	}
	ftExec(t, wl, `INSERT OR REPLACE INTO watchlater_videos
		(bvid, aid, title, owner_name, owner_mid, tname, duration, add_at, view, danmaku, link, fetch_time)
		VALUES ('BVFTVWWL0000001', 20130002, 'FT最早稍后再看', 'FTUP稍后', 410002, 'FT稍后再看', 90, ?, 500, 5, 'https://www.bilibili.com/video/BVFTVWWL0000001', 1357000000)`,
		viewAt(t, 2013, time.January, 1, 0, 0))
	ftExec(t, wl, `INSERT OR REPLACE INTO watchlater_videos
		(bvid, aid, title, owner_name, owner_mid, tname, duration, add_at, view, danmaku, link, fetch_time)
		VALUES ('BVFTVWWL0000002', 20130003, 'FT较晚稍后再看', 'FTUP稍后', 410002, 'FT稍后再看', 400, ?, 700, 7, 'https://www.bilibili.com/video/BVFTVWWL0000002', 1357000001)`,
		viewAt(t, 2013, time.March, 1, 0, 0))

	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}
	ftExec(t, fav, `INSERT OR REPLACE INTO favorites_folder (media_id, fid, mid, title, cover, folder_type, fetch_time)
		VALUES (20130004, 20130004, 410003, 'FT观看收藏夹', '', 0, 1357000000)`)
	ftExec(t, fav, `INSERT OR REPLACE INTO favorites_content
		(media_id, content_id, type, title, bvid, duration, upper_mid, fav_time, creator_name, play, collect, fetch_time)
		VALUES (20130004, 20130011, 2, 'FT收藏内容一', 'BVFTVWFV0000001', 300, 410003, 1357000000, 'FTUP收藏', 2000, 20, 1357000000)`)
	ftExec(t, fav, `INSERT OR REPLACE INTO favorites_content
		(media_id, content_id, type, title, bvid, duration, upper_mid, fav_time, creator_name, play, collect, fetch_time)
		VALUES (20130004, 20130012, 2, 'FT收藏内容二', 'BVFTVWFV0000002', 60, 410003, 1357000001, 'FTUP收藏', 4000, 40, 1357000001)`)
}

func ftExec(t *testing.T, handle *sql.DB, query string, args ...interface{}) {
	t.Helper()
	if _, err := handle.Exec(query, args...); err != nil {
		t.Fatalf("seed %q: %v", query, err)
	}
}

func ftWeekdayName(day time.Weekday) string {
	return [...]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}[day]
}

func ftHasYear(years []interface{}, want int) bool {
	for _, y := range years {
		if fmt.Sprint(y) == fmt.Sprint(want) {
			return true
		}
	}
	return false
}

func ftHasNamedEntry(t *testing.T, entries interface{}, name string) bool {
	t.Helper()
	for _, e := range dataArray(t, entries) {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if entry["name"] == name {
			return true
		}
	}
	return false
}

// TestFTViewingAuthorAndTagScoring seeds five extra records for one author and
// one tag: AnalyzeAuthorCompletion (database/analysis.go:1326) and
// AnalyzeTagAnalysis (database/analysis.go:1517) only score entries with at
// least five records, so the six base rows alone leave both buckets empty.
// Those extra rows move the aggregate counters asserted by the other tests in
// this file, so they are deleted again on cleanup and the file passes in any
// order.
func TestFTViewingAuthorAndTagScoring(t *testing.T) {
	ftSeedViewingHistory(t)
	ensureYear(t, ftViewYear)
	// The five rows below inflate every aggregate the other tests in this file
	// assert on (monthly totals, weekday buckets, watch hours), so they have to
	// be removed again for the file to pass in any order.
	handle := db(t)
	t.Cleanup(func() {
		if _, err := handle.Exec(fmt.Sprintf(`DELETE FROM bilibili_history_%d WHERE bvid LIKE 'BVFTVWSC%%'`, ftViewYear)); err != nil {
			t.Errorf("drop the scored viewing rows: %v", err)
		}
	})
	for i := 0; i < 5; i++ {
		insertHistory(t, ftViewYear, map[string]interface{}{
			"bvid":          fmt.Sprintf("BVFTVWSC000000%d", i+1),
			"title":         fmt.Sprintf("FT评分样本%d", i+1),
			"tag_name":      "科技",
			"main_category": "科技数码",
			"author_name":   "FTUP丁",
			"author_mid":    400004,
			"view_at":       viewAt(t, ftViewYear, time.February, 20+i, 9, 30),
			"duration":      100,
			"progress":      100,
			"business":      "video",
		})
	}

	e := newAPI(t, RegisterViewingRoutes)
	year := fmt.Sprintf("?year=%d", ftViewYear)

	// getAuthorCompletion scores FTUP丁: five fully watched records give an
	// average completion of 100% and a confidence factor of 5/20 = 0.25, so the
	// comprehensive score is (0*0.25 + 100*0.5 + 100*0.25) * 0.25 = 18.75.
	block := expectStatus(t, e, "GET", "/api/viewing/author-completion"+year, "", 200, "success").
		dataMap(t)["completion_rates"].(map[string]interface{})
	lead := block["most_watched_authors"].(map[string]interface{})["FTUP丁"].(map[string]interface{})
	if int(lead["video_count"].(float64)) != 5 || int(lead["fully_watched"].(float64)) != 5 {
		t.Fatalf("FTUP丁 counts = %v, want five records all fully watched", lead)
	}
	if int64(lead["author_mid"].(float64)) != 400004 {
		t.Fatalf("FTUP丁 author_mid = %v, want 400004", lead["author_mid"])
	}
	for key, want := range map[string]float64{
		"average_completion_rate": 100,
		"fully_watched_rate":      100,
		"comprehensive_score":     18.75,
		"loyalty_score":           5,
		"quality_score":           100,
	} {
		if got := lead[key].(float64); got != want {
			t.Fatalf("FTUP丁 %s = %v, want %v", key, got, want)
		}
	}
	for _, key := range []string{"highest_completion_authors", "most_valuable_authors"} {
		if _, ok := block[key].(map[string]interface{})["FTUP丁"]; !ok {
			t.Fatalf("%s = %v, want the scored FTUP丁 entry", key, block[key])
		}
	}
	// potential_authors needs quality > 85 *and* a record count below the third
	// highest count; a single scored author can never satisfy it.
	if got := block["potential_authors"].(map[string]interface{}); len(got) != 0 {
		t.Fatalf("potential_authors = %v, want empty for a lone scored author", got)
	}

	// getTagAnalysis: the three base 科技 rows (100%, 50%, 0%) plus the five new
	// ones give eight records, an average completion of 650/8 = 81.25% and a
	// fully-watched rate of 6/8 = 75%.
	tag := expectStatus(t, e, "GET", "/api/viewing/tag-analysis"+year, "", 200, "success").dataMap(t)
	dist := tag["watch_counts"].(map[string]interface{})["tag_distribution"].(map[string]interface{})
	if int(dist["科技"].(float64)) != 8 {
		t.Fatalf("tag_distribution[科技] = %v, want 3 base rows plus 5 scored rows", dist["科技"])
	}
	rates := tag["completion_rates"].(map[string]interface{})["tag_completion_rates"].(map[string]interface{})
	scoped, ok := rates["科技"].(map[string]interface{})
	if !ok {
		t.Fatalf("tag_completion_rates = %v, want 科技 past the five-record gate", rates)
	}
	stats := scoped
	if int(stats["video_count"].(float64)) != 8 {
		t.Fatalf("科技 video_count = %v, want 8", stats["video_count"])
	}
	if got := stats["average_completion_rate"].(float64); got < 81.2 || got > 81.3 {
		t.Fatalf("科技 average_completion_rate = %v, want 650/8", got)
	}
	if got := stats["fully_watched_rate"].(float64); got < 74.9 || got > 75.1 {
		t.Fatalf("科技 fully_watched_rate = %v, want 6/8", got)
	}
	if _, ok := rates["音乐"]; ok {
		t.Fatalf("tag_completion_rates = %v, want 音乐 below the five-record gate", rates)
	}
}
