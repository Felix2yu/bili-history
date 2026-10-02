package routers

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"bilibili-history-go/database"
)

// Reserved namespace for this file: years 2001 (titles), 2005 (never seeded,
// used for the "table missing" branch) and 2003 (monthly trend).

var dlTitleSeedOnce sync.Once

func dlTitleBvid(n int) string { return uniqueBvid("BVDLTA", n) }

// dlSeedTitles writes the fixed title fixture into the 2001 table. Titles are
// chosen so every pattern / sentiment / length bucket branch is hit with an
// exactly known count.
func dlSeedTitles(t *testing.T) {
	t.Helper()
	dlTitleSeedOnce.Do(func() {
		ensureYear(t, 2001)
		titles := []string{
			"你好世界",                  // len 4, positive via 好 substring
			"Hello World",           // len 11, pure letters (buggy 纯中文 bucket)
			"1024",                  // len 4, pure digits
			"BV1测试666！这太棒了",         // len 13, digits+letters, full-width !, positive
			"好看吗？",                  // len 4, full-width ?, positive
			"难看的烂片",                 // len 5, negative
			"又好看又难看",                // len 6, positive+negative -> neutral
			"优秀,非常赞!",               // len 7, ascii , and !, positive
			"Test123",               // len 7, letters+digits
			strings.Repeat("字", 55), // len 55, 51+ bucket
		}
		for i, title := range titles {
			insertHistory(t, 2001, map[string]interface{}{
				"bvid":    dlTitleBvid(i + 1),
				"title":   title,
				"view_at": viewAt(t, 2001, time.June, 1+i%20, 10, i),
			})
		}
	})
}

func dlSeedTrend(t *testing.T) {
	t.Helper()
	ensureYear(t, 2003)
	rows := []struct {
		n     int
		title string
		vt    int64
	}{
		{1, "三月观看一二", viewAt(t, 2003, time.March, 5, 10, 0)},
		{2, "短标", viewAt(t, 2003, time.March, 20, 11, 0)},
		{3, "四月中的标题十八", viewAt(t, 2003, time.April, 10, 12, 0)},
		{4, "五月", viewAt(t, 2003, time.May, 3, 9, 0)},
	}
	for _, r := range rows {
		insertHistory(t, 2003, map[string]interface{}{
			"bvid": uniqueBvid("BVDLTB", r.n), "title": r.title, "view_at": r.vt,
		})
	}
}

func dlYearPath(path, year string) string {
	if year == "" {
		return "/api/title-analytics/" + path
	}
	return "/api/title-analytics/" + path + "?year=" + year
}

func TestDLTitleStats(t *testing.T) {
	dlSeedTitles(t)
	e := newAPI(t, RegisterTitleAnalyticsRoutes)

	resp := expectStatus(t, e, "GET", dlYearPath("stats", "2001"), "", 200, "success")
	data := resp.dataMap(t)
	if int(data["year"].(float64)) != 2001 {
		t.Fatalf("year = %v, want 2001", data["year"])
	}
	if int(data["total_titles"].(float64)) != 10 {
		t.Fatalf("total_titles = %v, want 10", data["total_titles"])
	}
	// lengths: 4+11+4+13+4+5+6+7+7+55 = 116 -> avg 11.6
	if got := data["avg_length"].(float64); got < 11.59 || got > 11.61 {
		t.Fatalf("avg_length = %v, want 11.6", got)
	}
	if int(data["min_length"].(float64)) != 4 || int(data["max_length"].(float64)) != 55 {
		t.Fatalf("min/max length = %v/%v, want 4/55", data["min_length"], data["max_length"])
	}
	if data["max_length_title"] != strings.Repeat("字", 55) {
		t.Fatalf("max_length_title = %q", data["max_length_title"])
	}
	dist := data["distribution"].(map[string]interface{})
	for bucket, want := range map[string]int{"1-10": 7, "11-20": 2, "21-30": 0, "31-50": 0, "51+": 1} {
		if int(dist[bucket].(float64)) != want {
			t.Fatalf("distribution[%s] = %v, want %d (full %v)", bucket, dist[bucket], want, dist)
		}
	}

	// Missing year table is reported, not crashed on.
	missing := expectStatus(t, e, "GET", dlYearPath("stats", "2005"), "", 200, "error")
	expectMessage(t, missing, "未找到 2005 年数据")

	// Unparseable year falls back to the newest available table.
	fallback := expectStatus(t, e, "GET", dlYearPath("stats", "abc"), "", 200, "success")
	fb := fallback.dataMap(t)
	if _, ok := fb["distribution"]; !ok {
		t.Fatalf("fallback payload missing distribution: %v", fb)
	}
	if int(fb["year"].(float64)) < 2001 {
		t.Fatalf("fallback year = %v", fb["year"])
	}
}

func TestDLTitlePatterns(t *testing.T) {
	dlSeedTitles(t)
	e := newAPI(t, RegisterTitleAnalyticsRoutes)

	data := expectStatus(t, e, "GET", dlYearPath("patterns", "2001"), "", 200, "success").dataMap(t)
	p := data["patterns"].(map[string]interface{})
	count := func(k string) int { return int(p[k].(float64)) }
	// 含数字: 1024 / BV1测试666！... / Test123
	if count("含数字") != 3 {
		t.Fatalf("含数字 = %v, want 3", p["含数字"])
	}
	if count("含字母") != 3 { // Hello World / BV1测试... / Test123
		t.Fatalf("含字母 = %v, want 3", p["含字母"])
	}
	if count("纯数字") != 1 {
		t.Fatalf("纯数字 = %v, want 1", p["纯数字"])
	}
	// 中英混合 = 拉丁字母 + 汉字：只有 "BV1测试666！这太棒了"。
	// (此前误用 containsAlpha && containsDigit，把 "Test123" 这类字母+数字标题算成混合。)
	if count("中英混合") != 1 {
		t.Fatalf("中英混合 = %v, want 1", p["中英混合"])
	}
	// 纯中文 = 含汉字且不含 ASCII 字母数字：你好世界 / 好看吗？ / 难看的烂片 /
	// 又好看又难看 / 优秀,非常赞! / 字×55。
	// (此前误用 isAllAlpha，只会命中 "Hello World" 这类纯字母标题。)
	if count("纯中文") != 6 {
		t.Fatalf("纯中文 = %v, want 6", p["纯中文"])
	}
	if count("含标点符号") != 1 { // only "优秀,非常赞!" uses ASCII punctuation
		t.Fatalf("含标点符号 = %v, want 1", p["含标点符号"])
	}
	if count("含问号") != 1 || count("问句") != 1 { // "好看吗？"
		t.Fatalf("含问号/问句 = %v/%v, want 1/1", p["含问号"], p["问句"])
	}
	if count("含感叹号") != 2 { // full-width ! and ASCII !
		t.Fatalf("含感叹号 = %v, want 2", p["含感叹号"])
	}
	if count("感叹句") != 1 { // only the ASCII-! title ends with a bang
		t.Fatalf("感叹句 = %v, want 1", p["感叹句"])
	}

	expectMessage(t, expectStatus(t, e, "GET", dlYearPath("patterns", "2005"), "", 200, "error"), "未找到 2005 年数据")
}

func TestDLTitleSentiment(t *testing.T) {
	dlSeedTitles(t)
	e := newAPI(t, RegisterTitleAnalyticsRoutes)

	data := expectStatus(t, e, "GET", dlYearPath("sentiment", "2001"), "", 200, "success").dataMap(t)
	s := data["sentiment"].(map[string]interface{})
	// 积极: 你好世界 (substring 好) / 太棒了 / 好看吗 / 优秀赞; 消极: 难看的烂片;
	// 其余 5 条中性 (含 又好看又难看 -> positive+negative cancels out).
	if int(s["积极"].(float64)) != 4 || int(s["消极"].(float64)) != 1 || int(s["中性"].(float64)) != 5 {
		t.Fatalf("sentiment = %v, want 积极4/消极1/中性5", s)
	}

	expectMessage(t, expectStatus(t, e, "GET", dlYearPath("sentiment", "2005"), "", 200, "error"), "未找到 2005 年数据")
}

func TestDLTitleLengthAnalysis(t *testing.T) {
	dlSeedTitles(t)
	e := newAPI(t, RegisterTitleAnalyticsRoutes)

	data := expectStatus(t, e, "GET", dlYearPath("length", "2001"), "", 200, "success").dataMap(t)
	g := data["length_groups"].(map[string]interface{})
	for bucket, want := range map[string]int{
		"1-5": 4, "6-10": 3, "11-15": 2, "16-20": 0, "21-25": 0,
		"26-30": 0, "31-40": 0, "41-50": 0, "51-100": 1, "100+": 0,
	} {
		if int(g[bucket].(float64)) != want {
			t.Fatalf("length_groups[%s] = %v, want %d (full %v)", bucket, g[bucket], want, g)
		}
	}

	expectMessage(t, expectStatus(t, e, "GET", dlYearPath("length", "2005"), "", 200, "error"), "未找到 2005 年数据")
}

func TestDLTitleTrend(t *testing.T) {
	dlSeedTrend(t)
	e := newAPI(t, RegisterTitleAnalyticsRoutes)

	data := expectStatus(t, e, "GET", dlYearPath("trend", "2003"), "", 200, "success").dataMap(t)
	trends := dataArray(t, data["trends"])
	if len(trends) != 3 {
		t.Fatalf("trend groups = %d, want 3 (Mar/Apr/May)", len(trends))
	}
	// Ordered by month ascending.
	var months []int
	for _, tr := range trends {
		m := tr.(map[string]interface{})
		months = append(months, int(m["month"].(float64)))
	}
	if fmt.Sprint(months) != "[3 4 5]" {
		t.Fatalf("months = %v, want [3 4 5]", months)
	}
	mar := trends[0].(map[string]interface{})
	if int(mar["count"].(float64)) != 2 || int(mar["min_len"].(float64)) != 2 || int(mar["max_len"].(float64)) != 6 {
		t.Fatalf("march group = %v, want count2 min2 max6", mar)
	}
	if avg := mar["avg_len"].(float64); avg < 3.99 || avg > 4.01 {
		t.Fatalf("march avg_len = %v, want 4", avg)
	}

	// Month filter keeps a single group.
	filtered := expectStatus(t, e, "GET", dlYearPath("trend", "2003")+"&month=4", "", 200, "success").dataMap(t)
	groups := dataArray(t, filtered["trends"])
	if len(groups) != 1 {
		t.Fatalf("month=4 groups = %v, want exactly one", groups)
	}
	g0 := groups[0].(map[string]interface{})
	if int(g0["month"].(float64)) != 4 || int(g0["count"].(float64)) != 1 || int(g0["max_len"].(float64)) != 8 {
		t.Fatalf("month=4 group = %v", g0)
	}

	// Out-of-range and non-numeric months are silently ignored (no filter).
	for _, m := range []string{"13", "0", "abc"} {
		resp := expectStatus(t, e, "GET", dlYearPath("trend", "2003")+"&month="+m, "", 200, "success")
		if l := len(dataArray(t, resp.dataMap(t)["trends"])); l != 3 {
			t.Fatalf("month=%s should not filter: groups = %d, want 3", m, l)
		}
	}

	expectMessage(t, expectStatus(t, e, "GET", dlYearPath("trend", "2005"), "", 200, "error"), "未找到 2005 年数据")
}

func TestDLTitleAnalyticsDefaultYear(t *testing.T) {
	dlSeedTitles(t)
	e := newAPI(t, RegisterTitleAnalyticsRoutes)

	// No year parameter: years[0] is the newest existing table, which belongs
	// to other fixtures, so only the envelope shape is asserted here.
	years, err := database.GetSQLiteDB().GetAvailableYears()
	if err != nil || len(years) == 0 {
		t.Fatalf("expected seeded years, got %v (%v)", years, err)
	}
	for _, path := range []string{"stats", "patterns", "sentiment", "length", "trend"} {
		resp := expectStatus(t, e, "GET", dlYearPath(path, ""), "", 200, "success")
		if int(resp.dataMap(t)["year"].(float64)) != years[0] {
			t.Fatalf("%s default year = %v, want %d", path, resp.dataMap(t)["year"], years[0])
		}
	}
}
