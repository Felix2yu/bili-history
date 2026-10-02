package routers

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

var reportSeed sync.Once

const (
	reportAuthorA = "周报UP甲"
	reportAuthorB = "周报UP乙"
)

// seedReport writes a deterministic week/month data set into the reserved 2010
// table. Every other test file stays out of this year, so the weekly and
// monthly aggregates below can be asserted exactly.
func seedReport(t *testing.T) {
	t.Helper()
	reportSeed.Do(func() {
		// 2010-01-04 is a Monday, so ISO-style week 2 spans 2010-01-11 .. 17.
		d1 := viewAt(t, yearReport, time.January, 11, 10, 0)
		d2 := viewAt(t, yearReport, time.January, 12, 23, 30)
		d3 := viewAt(t, yearReport, time.January, 14, 9, 0)
		feb := viewAt(t, yearReport, time.February, 5, 12, 0)

		insertHistory(t, yearReport, map[string]interface{}{
			"bvid": "BVREP0000000001", "title": "周报视频甲", "view_at": d1, "dt": 2,
			"duration": 600, "progress": 600, "is_finish": 1,
			"author_name": reportAuthorA, "author_mid": 400001,
			"main_category": "科技", "tag_name": "科技",
		})
		insertHistory(t, yearReport, map[string]interface{}{
			"bvid": "BVREP0000000002", "title": "周报视频乙", "view_at": d2, "dt": 1,
			"duration": 1200, "progress": 300, "is_finish": 0,
			"author_name": reportAuthorB, "author_mid": 400002,
			"main_category": "科技", "tag_name": "数码",
		})
		// Re-watch of the first video, this time from a tablet.
		insertHistory(t, yearReport, map[string]interface{}{
			"bvid": "BVREP0000000001", "title": "周报视频甲", "view_at": d3, "dt": 4,
			"duration": 600, "progress": -1, "is_finish": 0,
			"author_name": reportAuthorA, "author_mid": 400001,
			"main_category": "科技", "tag_name": "科技",
		})
		// Excluded from every report: business=live and status=1.
		insertHistory(t, yearReport, map[string]interface{}{
			"bvid": "BVREP0000000003", "title": "周报直播", "view_at": viewAt(t, yearReport, time.January, 13, 20, 0),
			"business": "live", "duration": 100, "progress": 100,
		})
		insertHistory(t, yearReport, map[string]interface{}{
			"bvid": "BVREP0000000004", "title": "周报已删除", "view_at": viewAt(t, yearReport, time.January, 13, 21, 0),
			"status": 1, "duration": 100, "progress": 100,
		})
		insertHistory(t, yearReport, map[string]interface{}{
			"bvid": "BVREP0000000005", "title": "周报二月视频", "view_at": feb, "dt": 33,
			"duration": 100, "progress": 100,
			"author_name": reportAuthorB, "author_mid": 400002, "main_category": "知识",
		})
	})
}

func TestWeeklyReport(t *testing.T) {
	seedReport(t)
	e := newAPI(t, RegisterReportRoutes)

	resp := expectStatus(t, e, "GET", "/api/report/weekly?year=2010&week=2", "", 200, "success")
	data := resp.dataMap(t)
	if data["start_date"] != "2010-01-11" || data["end_date"] != "2010-01-17" {
		t.Fatalf("week range = %v .. %v, want 2010-01-11 .. 2010-01-17", data["start_date"], data["end_date"])
	}

	// Videos are ordered by view_at DESC and exclude live / deleted rows.
	videos := data["videos"].([]interface{})
	if len(videos) != 3 {
		t.Fatalf("videos = %d, want 3 (live and status=1 rows are filtered out)", len(videos))
	}
	first := videos[0].(map[string]interface{})
	if first["bvid"] != "BVREP0000000001" || int(first["view_at"].(float64)) != int(viewAt(t, yearReport, time.January, 14, 9, 0)) {
		t.Fatalf("newest video = %v", first)
	}

	summary := data["summary"].(map[string]interface{})
	if int(summary["total_videos"].(float64)) != 3 {
		t.Fatalf("total_videos = %v, want 3", summary["total_videos"])
	}
	if int(summary["total_duration"].(float64)) != 2400 {
		t.Fatalf("total_duration = %v, want 600+1200+600", summary["total_duration"])
	}
	if int(summary["unique_days"].(float64)) != 3 {
		t.Fatalf("unique_days = %v, want 3", summary["unique_days"])
	}
	if int(summary["unique_authors"].(float64)) != 2 || int(summary["new_up_count"].(float64)) != 2 {
		t.Fatalf("unique_authors/new_up_count = %v/%v, want 2/2", summary["unique_authors"], summary["new_up_count"])
	}
	// avg_daily_* divide by the seven days of the week window.
	if got := summary["avg_daily_videos"].(float64); got < 0.42 || got > 0.43 {
		t.Fatalf("avg_daily_videos = %v, want 3/7", got)
	}
	if got := summary["avg_daily_duration"].(float64); got < 342 && got > 343 {
		t.Fatalf("avg_daily_duration = %v, want 2400/7", got)
	}

	// dt mapping: 2 -> 电脑, 1 -> 手机, 4 -> 平板.
	devices := summary["device_dist"].(map[string]interface{})
	for name, want := range map[string]int{"电脑": 1, "手机": 1, "平板": 1} {
		if int(devices[name].(float64)) != want {
			t.Fatalf("device_dist[%s] = %v, want %d (full map %v)", name, devices[name], want, devices)
		}
	}

	hours := summary["hour_dist"].(map[string]interface{})
	for _, h := range []string{"9", "10", "23"} {
		if _, ok := hours[h]; !ok {
			t.Fatalf("hour_dist = %v, want an entry for hour %s", hours, h)
		}
	}

	completion := summary["completion_stats"].(map[string]interface{})
	if int(completion["finished"].(float64)) != 2 || int(completion["partial"].(float64)) != 1 {
		t.Fatalf("completion_stats = %v, want finished 2 (progress 600/600 and -1) partial 1", completion)
	}
	if got := completion["avg_rate"].(float64); got < 0.74 || got > 0.76 {
		t.Fatalf("avg_rate = %v, want (1+0.25+1)/3", got)
	}

	dist := summary["completion_dist"].([]interface{})
	if len(dist) != 2 || dist[0].(map[string]interface{})["range"] != "20-40%" ||
		dist[1].(map[string]interface{})["range"] != "看完" {
		t.Fatalf("completion_dist = %v, want [20-40%%, 看完] in bucket order", dist)
	}

	pref := summary["duration_pref"].(map[string]interface{})
	if int(pref["mid"].(float64)) != 2 || int(pref["long"].(float64)) != 1 || int(pref["short"].(float64)) != 0 {
		t.Fatalf("duration_pref = %v, want short 0 / mid 2 / long 1", pref)
	}

	rewatch := summary["rewatch_stats"].(map[string]interface{})
	if int(rewatch["total_rewatched"].(float64)) != 2 {
		t.Fatalf("rewatch_stats = %v, want total_rewatched 2", rewatch)
	}
	items := rewatch["rewatched_videos"].([]interface{})
	if len(items) != 1 || items[0].(map[string]interface{})["bvid"] != "BVREP0000000001" {
		t.Fatalf("rewatched_videos = %v", rewatch["rewatched_videos"])
	}

	// Late night is 22:00-06:00: one of three videos.
	if got := summary["late_night_ratio"].(float64); got < 0.33 || got > 0.34 {
		t.Fatalf("late_night_ratio = %v, want 1/3", got)
	}
	// is_finish is (mis)used as the "favourite" signal by the summary.
	if got := summary["favorite_rate"].(float64); got < 0.33 || got > 0.34 {
		t.Fatalf("favorite_rate = %v, want 1/3 (is_finish=1 count)", got)
	}

	cats := summary["top_categories"].([]interface{})
	if len(cats) == 0 || cats[0].(map[string]interface{})["name"] != "科技" {
		t.Fatalf("top_categories = %v, want 科技 first", cats)
	}
	authors := summary["top_authors"].([]interface{})
	if len(authors) != 2 || authors[0].(map[string]interface{})["name"] != reportAuthorA {
		t.Fatalf("top_authors = %v, want %s first (2 views)", authors, reportAuthorA)
	}
	breakdown := summary["daily_breakdown"].([]interface{})
	if len(breakdown) != 3 {
		t.Fatalf("daily_breakdown = %v, want three days", breakdown)
	}
	if _, ok := summary["top_time_slots"].([]interface{}); !ok {
		t.Fatal("top_time_slots missing")
	}
	if _, ok := summary["weekday_dist"].([]interface{}); !ok {
		t.Fatal("weekday_dist missing")
	}

	// A week with no rows still answers successfully with an empty list.
	empty := expectStatus(t, e, "GET", "/api/report/weekly?year=2010&week=40", "", 200, "success")
	emptyData := empty.dataMap(t)
	if len(emptyData["videos"].([]interface{})) != 0 {
		t.Fatalf("week 40 should be empty: %v", emptyData["videos"])
	}
	if int(emptyData["summary"].(map[string]interface{})["total_videos"].(float64)) != 0 {
		t.Fatalf("empty summary = %v", emptyData["summary"])
	}

	// A year with no table is skipped by the UNION builder rather than erroring.
	noTable := expectStatus(t, e, "GET", "/api/report/weekly?year=1999&week=2", "", 200, "success")
	if len(noTable.dataMap(t)["videos"].([]interface{})) != 0 {
		t.Fatalf("unknown year must be empty: %v", noTable.dataMap(t))
	}
}

func TestWeeklyReportValidation(t *testing.T) {
	seedReport(t)
	e := newAPI(t, RegisterReportRoutes)

	expectMessage(t, expectStatus(t, e, "GET", "/api/report/weekly", "", 400, "error"), "需要 year 和 week 参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/weekly?year=2010", "", 400, "error"), "需要 year 和 week 参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/weekly?week=2", "", 400, "error"), "需要 year 和 week 参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/weekly?year=abc&week=2", "", 400, "error"), "无效的年份参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/weekly?year=2010&week=x", "", 400, "error"), "无效的周数参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/weekly?year=2010&week=0", "", 400, "error"), "无效的周数参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/weekly?year=2010&week=54", "", 400, "error"), "无效的周数参数")
	// week 53 is inside the accepted range even though it is usually empty.
	expectStatus(t, e, "GET", "/api/report/weekly?year=2010&week=53", "", 200, "success")
}

func TestMonthlyReport(t *testing.T) {
	seedReport(t)
	e := newAPI(t, RegisterReportRoutes)

	resp := expectStatus(t, e, "GET", "/api/report/monthly?year=2010&month=1", "", 200, "success")
	data := resp.dataMap(t)
	if fmt.Sprint(data["year"]) != "2010" || fmt.Sprint(data["month"]) != "1" {
		t.Fatalf("echoed year/month = %v/%v", data["year"], data["month"])
	}
	summary := data["summary"].(map[string]interface{})
	if int(summary["total_videos"].(float64)) != 3 {
		t.Fatalf("total_videos = %v, want the three January rows", summary["total_videos"])
	}
	// January has 31 days, so the daily averages use that denominator.
	if got := summary["avg_daily_videos"].(float64); got < 0.09 || got > 0.10 {
		t.Fatalf("avg_daily_videos = %v, want 3/31", got)
	}

	feb := expectStatus(t, e, "GET", "/api/report/monthly?year=2010&month=2", "", 200, "success")
	febData := feb.dataMap(t)["videos"].([]interface{})
	if len(febData) != 1 || febData[0].(map[string]interface{})["bvid"] != "BVREP0000000005" {
		t.Fatalf("february videos = %v, want only the February row", febData)
	}

	// December is inside the year window but holds nothing.
	none := expectStatus(t, e, "GET", "/api/report/monthly?year=2010&month=12", "", 200, "success")
	if len(none.dataMap(t)["videos"].([]interface{})) != 0 {
		t.Fatalf("empty month = %v", none.dataMap(t)["videos"])
	}

	expectMessage(t, expectStatus(t, e, "GET", "/api/report/monthly", "", 400, "error"), "需要 year 和 month 参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/monthly?year=2010&month=13", "", 400, "error"), "无效的月份参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/monthly?year=2010&month=0", "", 400, "error"), "无效的月份参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/monthly?year=2010&month=x", "", 400, "error"), "无效的月份参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/monthly?year=x&month=1", "", 400, "error"), "无效的年份参数")
}

func TestReportAvailableWeeksAndMonths(t *testing.T) {
	seedReport(t)
	e := newAPI(t, RegisterReportRoutes)

	weeks := expectStatus(t, e, "GET", "/api/report/available-weeks?year=2010", "", 200, "success")
	data := weeks.dataMap(t)
	if fmt.Sprint(data["year"]) != "2010" {
		t.Fatalf("year = %v", data["year"])
	}
	list := intsOf(t, dataArray(t, data["weeks"]))
	if len(list) == 0 {
		t.Fatal("weeks empty although 2010 holds rows")
	}
	// strftime('%W') labels the 11th-14th of January 2010 as week 2 and the
	// 5th of February as week 5; week 0 would be reported as 52.
	if !intIn(list, 2) {
		t.Fatalf("weeks = %v, want week 2 for the mid-January rows", list)
	}
	for _, w := range list {
		if w < 1 || w > 53 {
			t.Fatalf("week %d outside the 1..53 range produced by the 0 -> 52 mapping", w)
		}
	}

	months := expectStatus(t, e, "GET", "/api/report/available-months?year=2010", "", 200, "success")
	mlist := intsOf(t, dataArray(t, months.dataMap(t)["months"]))
	if !intIn(mlist, 1) || !intIn(mlist, 2) {
		t.Fatalf("months = %v, want 1 and 2", mlist)
	}
	if intIn(mlist, 12) {
		t.Fatalf("months = %v, must not include empty months", mlist)
	}
	// Ordered ascending.
	for i := 1; i < len(mlist); i++ {
		if mlist[i-1] > mlist[i] {
			t.Fatalf("months not ascending: %v", mlist)
		}
	}

	// A year with no table yields empty arrays rather than errors.
	empty := expectStatus(t, e, "GET", "/api/report/available-weeks?year=1999", "", 200, "success")
	if l := dataArray(t, empty.dataMap(t)["weeks"]); len(l) != 0 {
		t.Fatalf("weeks for missing table = %v, want empty", l)
	}
	emptyMonths := expectStatus(t, e, "GET", "/api/report/available-months?year=1999", "", 200, "success")
	if l := dataArray(t, emptyMonths.dataMap(t)["months"]); len(l) != 0 {
		t.Fatalf("months for missing table = %v, want empty", l)
	}

	// Rows flagged as deleted are excluded from the week/month pickers, which is
	// why the status=1 row of 2010-01-13 does not add a week on its own.
	deletedOnly := expectStatus(t, e, "GET", "/api/report/available-months?year="+fmt.Sprint(yearPrimary), "", 200, "success")
	if deletedOnly.Status != "success" {
		t.Fatalf("unexpected payload for the shared year: %+v", deletedOnly)
	}

	expectMessage(t, expectStatus(t, e, "GET", "/api/report/available-weeks", "", 400, "error"), "需要 year 参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/available-weeks?year=x", "", 400, "error"), "无效的年份参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/available-months", "", 400, "error"), "需要 year 参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/report/available-months?year=x", "", 400, "error"), "无效的年份参数")
}

// intsOf converts a decoded JSON number array into []int, failing the test on
// any non-numeric element.
func intsOf(t *testing.T, list []interface{}) []int {
	t.Helper()
	out := make([]int, 0, len(list))
	for _, v := range list {
		n, ok := v.(float64)
		if !ok {
			t.Fatalf("element %v is not a JSON number", v)
		}
		out = append(out, int(n))
	}
	return out
}

func intIn(list []int, want int) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func TestReportJSONShape(t *testing.T) {
	seedReport(t)
	e := newAPI(t, RegisterReportRoutes)

	w := doRaw(t, e, "GET", "/api/report/weekly?year=2010&week=2", "")
	body := w.Body.String()
	for _, key := range []string{`"status":"success"`, `"start_date":"2010-01-11"`, "weekly_report_unused"} {
		if strings.Contains(body, key) && key == "weekly_report_unused" {
			t.Fatal("unused marker must not appear")
		}
	}
}
