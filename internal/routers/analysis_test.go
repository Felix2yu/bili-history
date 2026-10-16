package routers

import (
	"fmt"
	"testing"
	"time"
)

// seedAnalysis writes a small, deterministic data set into the reserved 2012
// table: three videos in March and one in July, all authored by two UP hosts.
// It returns the timestamps so the assertions can compute expected numbers.
func seedAnalysis(t *testing.T) (marchTS, julyTS int64) {
	t.Helper()
	marchTS = viewAt(t, yearQuaternary, time.March, 10, 12, 0)
	julyTS = viewAt(t, yearQuaternary, time.July, 1, 23, 30)

	insertHistory(t, yearQuaternary, map[string]interface{}{
		"bvid": uniqueBvid("BVAN", 1), "title": "分析样本A", "tag_name": "科技",
		"author_name": "分析UP甲", "author_mid": 300001, "view_at": marchTS,
		"duration": 200, "progress": 200, "business": "video",
	})
	insertHistory(t, yearQuaternary, map[string]interface{}{
		"bvid": uniqueBvid("BVAN", 2), "title": "分析样本B", "tag_name": "科技",
		"author_name": "分析UP甲", "author_mid": 300001, "view_at": marchTS + 60,
		"duration": 100, "progress": 50, "business": "video",
	})
	insertHistory(t, yearQuaternary, map[string]interface{}{
		"bvid": uniqueBvid("BVAN", 3), "title": "分析样本C", "tag_name": "音乐",
		"author_name": "分析UP乙", "author_mid": 300002, "view_at": marchTS + 120,
		"duration": 50, "progress": -1, "business": "pgc",
	})
	insertHistory(t, yearQuaternary, map[string]interface{}{
		"bvid": uniqueBvid("BVAN", 4), "title": "分析样本D", "tag_name": "",
		"author_name": "分析UP乙", "author_mid": 300002, "view_at": julyTS,
		"duration": 10, "progress": 0, "business": "live",
	})
	return marchTS, julyTS
}

func TestAnalyzeHistory(t *testing.T) {
	seedAnalysis(t)
	e := newAPI(t, RegisterAnalysisRoutes)

	resp := expectStatus(t, e, "POST", "/api/analysis/analyze?year=2012", "", 200, "success")
	data := resp.dataMap(t)
	if int(data["total_videos"].(float64)) != 4 {
		t.Fatalf("total_videos = %v, want the four seeded rows", data["total_videos"])
	}
	// progress -1 falls back to duration: 200 + 50 + 50 + 0.
	if int(data["total_duration"].(float64)) != 300 {
		t.Fatalf("total_duration = %v, want 300", data["total_duration"])
	}
	if int(data["unique_authors"].(float64)) != 2 {
		t.Fatalf("unique_authors = %v, want 2", data["unique_authors"])
	}
	if len(data["tag_ranking"].([]interface{})) != 2 {
		t.Fatalf("tag_ranking = %v, want 两个非空标签", data["tag_ranking"])
	}
	monthly, ok := data["monthly_stats"].([]interface{})
	if !ok || len(monthly) != 2 {
		t.Fatalf("monthly_stats = %v, want only the months holding data", data["monthly_stats"])
	}
	// The handler wraps the analysis result in its own envelope fields.
	if fmt.Sprint(data["year"]) != "2012" {
		t.Fatalf("nested year = %v, want 2012", data["year"])
	}

	// Without a year the newest available table is analysed.
	noYear := expectCode(t, e, "POST", "/api/analysis/analyze", "", 200)
	if noYear.Status != "success" {
		t.Fatalf("no-year analyze = %+v", noYear)
	}

	// A non-numeric year is rejected.
	expectStatus(t, e, "POST", "/api/analysis/analyze?year=abc", "", 400, "error")
	// A numeric year that has no table is rejected by the membership check.
	missing := expectStatus(t, e, "POST", "/api/analysis/analyze?year=1999", "", 200, "error")
	expectMessageContains(t, missing, "未找到指定年份")
}

func TestDailyStatsValidation(t *testing.T) {
	seedAnalysis(t)
	e := newAPI(t, RegisterAnalysisRoutes)

	cases := []struct {
		query string
		want  string
	}{
		{"", "日期参数无效，应为MMDD格式，例如0113表示1月13日"},
		{"?date=123", "日期参数无效，应为MMDD格式，例如0113表示1月13日"},
		{"?date=12345", "日期参数无效，应为MMDD格式，例如0113表示1月13日"},
		{"?date=1301", "月份无效"},
		{"?date=ab01", "月份无效"},
		{"?date=0132", "日期无效"},
		{"?date=01ab", "日期无效"},
		{"?date=0101&year=abc", "年份参数无效"},
	}
	for _, c := range cases {
		resp := expectStatus(t, e, "GET", "/api/daily/stats"+c.query, "", 400, "error")
		expectMessage(t, resp, c.want)
	}

	// A date of the form "0000" reaches the day check: month 0 is invalid.
	expectMessage(t, expectStatus(t, e, "GET", "/api/daily/stats?date=0000", "", 400, "error"), "月份无效")

	// 2012-03-10 holds the three March rows.
	resp := expectStatus(t, e, "GET", "/api/daily/stats?date=0310&year=2012", "", 200, "success")
	data := resp.dataMap(t)
	if data["date"] != "2012-03-10" {
		t.Fatalf("date = %v, want 2012-03-10 (local rendering)", data["date"])
	}
	if int(data["total_count"].(float64)) != 3 {
		t.Fatalf("total_count = %v, want 3", data["total_count"])
	}
	types := data["type_counts"].(map[string]interface{})
	if int(types["video"].(float64)) != 2 || int(types["pgc"].(float64)) != 1 {
		t.Fatalf("type_counts = %v, want video:2 pgc:1", types)
	}
	if int(data["total_watch_seconds"].(float64)) != 300 {
		t.Fatalf("total_watch_seconds = %v, want 300", data["total_watch_seconds"])
	}
	if int(data["completed_videos"].(float64)) != 1 {
		t.Fatalf("completed_videos = %v, want 1 (200/200)", data["completed_videos"])
	}
	if data["tag_distribution"] == nil {
		t.Fatal("tag_distribution missing")
	}
	if _, ok := data["insights"].([]interface{}); !ok {
		t.Fatalf("insights = %v, want an array", data["insights"])
	}

	// A year without a table answers an error payload, not a 400.
	missing := expectStatus(t, e, "GET", "/api/daily/stats?date=0310&year=1999", "", 200, "error")
	expectMessageContains(t, missing, "未找到 1999 年的历史记录数据")

	// Without a year the handler falls back to the current calendar year, which
	// has no table in the fixture database.
	now := expectStatus(t, e, "GET", "/api/daily/stats?date=0310", "", 200, "error")
	expectMessageContains(t, now, "年的历史记录数据")
}

func TestHeatmapData(t *testing.T) {
	seedAnalysis(t)
	e := newAPI(t, RegisterAnalysisRoutes)

	resp := expectStatus(t, e, "GET", "/api/heatmap/data?year=2012", "", 200, "success")
	data := resp.dataMap(t)
	if fmt.Sprint(data["year"]) != "2012" {
		t.Fatalf("year = %v", data["year"])
	}
	entries := data["data"].(map[string]interface{})
	if int(entries["2012-03-10"].(float64)) != 3 {
		t.Fatalf("2012-03-10 = %v, want 3 (heatmap counts every business)", entries)
	}
	if int(data["total"].(float64)) != 4 {
		t.Fatalf("total = %v, want 4", data["total"])
	}

	// The year is *not* checked against the available years here; an unknown
	// year only fails inside the query helper (PRODUCTION behaviour differs from
	// /analysis/analyze, which validates membership).
	unknown := expectStatus(t, e, "GET", "/api/heatmap/data?year=1999", "", 200, "error")
	expectMessageContains(t, unknown, "未找到 1999 年的历史记录数据")

	expectStatus(t, e, "GET", "/api/heatmap/data?year=oops", "", 400, "error")

	// No year: the newest table wins.
	noYear := expectStatus(t, e, "GET", "/api/heatmap/data", "", 200, "success")
	if noYear.dataMap(t)["data"] == nil {
		t.Fatalf("default-year heatmap payload: %v", noYear.dataMap(t))
	}
}

func TestViewingStats(t *testing.T) {
	ensureYear(t, yearEmpty)
	e := newAPI(t, RegisterAnalysisRoutes)

	// An empty year table answers zeroes. (GetViewingAnalytics once deadlocked on
	// its undrained "peak day" cursor against a non-empty year, because
	// database/sqlite.go:46 caps the pool at one connection; that is fixed in the
	// data layer, and the populated case below is what guards it.)
	resp := expectStatus(t, e, "GET", "/api/viewing/stats?year=2011", "", 200, "success")
	data := resp.dataMap(t)
	if int(data["total_videos"].(float64)) != 0 {
		t.Fatalf("total_videos = %v, want 0 for the empty table", data["total_videos"])
	}
	if int(data["total_watch_time"].(float64)) != 0 {
		t.Fatalf("total_watch_time = %v, want 0", data["total_watch_time"])
	}
	if _, ok := data["time_distribution"].(map[string]interface{}); !ok {
		t.Fatalf("time_distribution missing: %v", data)
	}
	if _, ok := data["weekday_distribution"].(map[string]interface{}); !ok {
		t.Fatalf("weekday_distribution missing: %v", data)
	}

	// A year with rows must complete rather than hang: every analytics query in
	// the handler shares the single pooled connection.
	insertHistory(t, yearViewing, map[string]interface{}{
		"bvid": uniqueBvid("BVVW", 1), "title": "观看样本A", "tag_name": "科技",
		"view_at": viewAt(t, yearViewing, time.June, 5, 21, 0),
		"duration": 300, "progress": 300, "business": "video", "author_mid": 310001,
	})
	insertHistory(t, yearViewing, map[string]interface{}{
		"bvid": uniqueBvid("BVVW", 2), "title": "观看样本B", "tag_name": "音乐",
		"view_at": viewAt(t, yearViewing, time.June, 6, 8, 0),
		"duration": 120, "progress": 60, "business": "video", "author_mid": 310002,
	})
	// A second June row so the peak day has an unambiguous winner: the query
	// orders by COUNT DESC LIMIT 1 and a 1-1 tie has no defined tie-break.
	insertHistory(t, yearViewing, map[string]interface{}{
		"bvid": uniqueBvid("BVVW", 3), "title": "观看样本C", "tag_name": "科技",
		"view_at": viewAt(t, yearViewing, time.June, 5, 22, 30),
		"duration": 90, "progress": 0, "business": "video", "author_mid": 310003,
	})
	done := make(chan apiResp, 1)
	go func() {
		code, r := do(t, e, "GET", fmt.Sprintf("/api/viewing/stats?year=%d", yearViewing), "")
		if code != 200 || r.Status != "success" {
			t.Errorf("populated viewing stats = code %d, %+v", code, r)
		}
		done <- r
	}()
	select {
	case r := <-done:
		p := r.dataMap(t)
		if int(p["total_videos"].(float64)) != 3 {
			t.Fatalf("total_videos = %v, want the 3 seeded rows", p["total_videos"])
		}
		if int(p["total_watch_time"].(float64)) != 360 {
			t.Fatalf("total_watch_time = %v, want 360 (300 + 60 + 0)", p["total_watch_time"])
		}
		if p["peak_day"] != fmt.Sprintf("%d-06-05", yearViewing) {
			t.Errorf("peak_day = %v, want the two-row June 5th", p["peak_day"])
		}
		if int(p["peak_day_count"].(float64)) != 2 {
			t.Errorf("peak_day_count = %v, want 2", p["peak_day_count"])
		}
	case <-time.After(10 * time.Second):
		t.Fatal("viewing/stats blocked on a populated year (peak-day cursor deadlock回归)")
	}

	expectStatus(t, e, "GET", "/api/viewing/stats?year=oops", "", 400, "error")
	unknown := expectStatus(t, e, "GET", "/api/viewing/stats?year=1999", "", 200, "error")
	expectMessageContains(t, unknown, "未找到 1999 年的历史记录数据")
}
