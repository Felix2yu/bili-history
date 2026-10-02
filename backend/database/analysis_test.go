package database

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"bilibili-history-go/models"
)

// 本文件专属于 analysis.go 的单元测试。
// 年份表沙箱：2020（主数据集，可随意读写）、2022（空表 + 单UP主边界数据集）、
// 1998（永远不建表，用于错误路径）。
// 所有 helper / 测试变量均以 ana 前缀命名，bvid 均以 BVana 前缀命名。

const (
	anaYear        = 2020
	anaEdgeYear    = 2022
	anaMissingYear = 1998
)

// ==================== 2020 数据集定义 ====================

type anaRowDef struct {
	bvid, title, business, tag, main, author string
	mid                                      int64
	month, day, hour, minute                 int
	duration, progress, dt                   int
}

func (r anaRowDef) at() int64 {
	return atTime(anaYear, r.month, r.day, r.hour, r.minute)
}

func (r anaRowDef) record() models.HistoryRecord {
	return historyRecord(r.bvid, r.title, r.business, r.tag, r.main, r.mid, r.author,
		r.at(), r.duration, r.progress, r.dt)
}

// anaRows2020 返回确定性的 34 行数据集：
//   - UP主分布（≥5 部才进入作者完成率过滤）：
//     ana大UP(7101)=10 行（含 3 组重复 bvid 各 2 次）、ana游戏UP(7102)=7、
//     ana音乐UP(7103)=7、ana星UP(7104)=5（全部看完，用于 potential 分支）、
//     ana少UP(7105)=3（被过滤）、直播UP(7106)=1、无名(7107)=1。
//   - 连续观看日：1/1-1/3（3 连）、3/10-3/14（5 连，最长）、12/29-12/30（当前连续 2）。
//   - 6 月为空档期（6/15 作为空日期查询样本）。
//   - 覆盖 progress=-1 / 0 / 超进度 / duration=0 / 空 tag / 空 business / 各种 dt 值。
func anaRows2020() []anaRowDef {
	return []anaRowDef{
		// ana大UP (7101) — tag ana知识, 10 行, bvid B001/B002/B003 各出现 2 次
		{"BVanaB001", "ana 知识 深度 一", "archive", "ana知识", "ana综合", "ana大UP", 7101, 1, 1, 2, 0, 120, 120, 1},
		{"BVanaB002", "ana 知识 深度 二", "archive", "ana知识", "ana综合", "ana大UP", 7101, 1, 2, 3, 10, 300, -1, 1},
		{"BVanaB003", "ana 知识 深度 三", "archive", "ana知识", "ana综合", "ana大UP", 7101, 1, 3, 4, 0, 1200, 600, 2},
		{"BVanaB004", "ana 知识 深度 四", "archive", "ana知识", "ana综合", "ana大UP", 7101, 1, 3, 21, 30, 1800, 1800, 33},
		{"BVanaB002", "ana 知识 深度 二", "archive", "ana知识", "ana综合", "ana大UP", 7101, 1, 5, 8, 0, 300, 0, 1},
		{"BVanaB005", "ana 知识 深度 五", "archive", "ana知识", "ana综合", "ana大UP", 7101, 2, 14, 13, 0, 240, 220, 1},
		{"BVanaB003", "ana 知识 深度 三", "archive", "ana知识", "ana综合", "ana大UP", 7101, 2, 20, 22, 0, 1200, 300, 4},
		{"BVanaB006", "ana 知识 深度 六", "archive", "ana知识", "ana综合", "ana大UP", 7101, 5, 30, 16, 45, 600, 300, 5},
		{"BVanaB007", "ana 知识 深度 七", "archive", "ana知识", "ana综合", "ana大UP", 7101, 7, 8, 1, 15, 120, 0, 1},
		{"BVanaB001", "ana 知识 深度 一", "archive", "ana知识", "ana综合", "ana大UP", 7101, 10, 10, 11, 0, 120, 240, 99},
		// ana游戏UP (7102) — tag ana游戏（M006 空 tag）, 7 行
		{"BVanaM001", "ana 游戏 实况 一", "archive", "ana游戏", "ana游戏区", "ana游戏UP", 7102, 3, 1, 9, 0, 400, 100, 1},
		{"BVanaM002", "ana 游戏 实况 二", "archive", "ana游戏", "ana游戏区", "ana游戏UP", 7102, 3, 2, 10, 0, 400, 200, 1},
		{"BVanaM003", "ana 游戏 实况 三", "archive", "ana游戏", "ana游戏区", "ana游戏UP", 7102, 3, 3, 23, 0, 400, 400, 2},
		{"BVanaM004", "ana 游戏 实况 四", "archive", "ana游戏", "ana游戏区", "ana游戏UP", 7102, 3, 8, 14, 0, 400, 380, 1},
		{"BVanaM005", "ana 游戏 实况 五", "archive", "ana游戏", "ana游戏区", "ana游戏UP", 7102, 4, 22, 19, 0, 400, 0, 1},
		{"BVanaM006", "ana 游戏 实况 六", "archive", "", "ana游戏区", "ana游戏UP", 7102, 7, 20, 6, 0, 0, 0, 1},
		{"BVanaM007", "ana 游戏 实况 七", "archive", "ana游戏", "ana游戏区", "ana游戏UP", 7102, 12, 29, 20, 0, 800, 750, 2},
		// ana音乐UP (7103) — tag ana音乐, 7 行
		{"BVanaC001", "ana 音乐 现场 一", "archive", "ana音乐", "ana音乐区", "ana音乐UP", 7103, 8, 1, 15, 0, 120, 120, 1},
		{"BVanaC002", "ana 音乐 现场 二", "archive", "ana音乐", "ana音乐区", "ana音乐UP", 7103, 8, 2, 15, 30, 100, -1, 4},
		{"BVanaC003", "ana 音乐 现场 三", "archive", "ana音乐", "ana音乐区", "ana音乐UP", 7103, 8, 3, 5, 0, 60, 60, 1},
		{"BVanaC004", "ana 音乐 现场 四", "archive", "ana音乐", "ana音乐区", "ana音乐UP", 7103, 9, 9, 12, 0, 60, 30, 1},
		{"BVanaC005", "ana 音乐 现场 五", "archive", "ana音乐", "ana音乐区", "ana音乐UP", 7103, 10, 25, 17, 0, 30, 0, 1},
		{"BVanaC006", "ana 音乐 现场 六", "archive", "ana音乐", "ana音乐区", "ana音乐UP", 7103, 11, 11, 23, 0, 2400, 1200, 6},
		{"BVanaC007", "ana 音乐 现场 七", "archive", "ana音乐", "ana音乐区", "ana音乐UP", 7103, 12, 30, 23, 30, 240, 240, 1},
		// ana星UP (7104) — tag ana科技, 5 行, 几乎全部看完 => potential 分支
		{"BVanaS001", "ana 科技 前沿 一", "archive", "ana科技", "ana科技区", "ana星UP", 7104, 3, 10, 7, 0, 300, 300, 1},
		{"BVanaS002", "ana 科技 前沿 二", "archive", "ana科技", "ana科技区", "ana星UP", 7104, 3, 11, 8, 0, 300, 299, 1},
		{"BVanaS003", "ana 科技 前沿 三", "archive", "ana科技", "ana科技区", "ana星UP", 7104, 3, 12, 9, 0, 300, -1, 1},
		{"BVanaS004", "ana 科技 前沿 四", "archive", "ana科技", "ana科技区", "ana星UP", 7104, 3, 13, 10, 0, 300, 300, 2},
		{"BVanaS005", "ana 科技 前沿 五", "archive", "ana科技", "ana科技区", "ana星UP", 7104, 3, 14, 11, 0, 300, 300, 1},
		// ana少UP (7105) — 3 行，被 >=5 过滤
		{"BVanaF001", "ana 生活 日常 一", "archive", "ana生活", "ana生活区", "ana少UP", 7105, 4, 1, 13, 0, 100, 50, 1},
		{"BVanaF002", "ana 生活 日常 二", "archive", "ana生活", "ana生活区", "ana少UP", 7105, 4, 15, 13, 30, 100, 10, 1},
		{"BVanaF003", "ana 生活 日常 三", "archive", "ana生活", "ana生活区", "ana少UP", 7105, 11, 30, 0, 5, 100, 95, 1},
		// 特殊行：直播 + 空 business（TypeCounts "other" 分支）
		{"BVanaL001", "ana 直播 房间 测试", "live", "ana直播", "ana直播区", "ana直播UP", 7106, 7, 4, 21, 0, 600, 300, 4},
		{"BVanaX001", "ana 无分类 视频", "", "ana其他", "ana其他区", "ana无名", 7107, 1, 3, 22, 15, 60, 60, 5},
	}
}

var (
	ana2020Once sync.Once
	ana2020Err  error
)

func anaEnsure2020(t *testing.T) {
	t.Helper()
	ana2020Once.Do(func() {
		recs := make([]models.HistoryRecord, 0, 32)
		for _, r := range anaRows2020() {
			recs = append(recs, r.record())
		}
		if err := GetSQLiteDB().EnsureTableForYear(anaYear); err != nil {
			ana2020Err = err
			return
		}
		ana2020Err = seedHistoryRecords(testConn(), anaYear, recs...)
	})
	if ana2020Err != nil {
		t.Fatalf("seed %d: %v", anaYear, ana2020Err)
	}
}

// ==================== 期望值推导工具（ana 前缀） ====================

func anaEff(r anaRowDef) int {
	if r.progress == -1 {
		return r.duration
	}
	return r.progress
}

func anaDate(r anaRowDef) string {
	return time.Unix(r.at(), 0).In(time.Local).Format("2006-01-02")
}

// anaCompletion 与 production 的完成率规则一致：-1 => 100；duration>0 => 比例*100；否则 0。
func anaCompletion(duration, progress int) float64 {
	if progress == -1 {
		return 100
	}
	if duration > 0 {
		return float64(progress) / float64(duration) * 100
	}
	return 0
}

func anaUniqueDates(rows []anaRowDef) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rows {
		d := anaDate(r)
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	sort.Strings(out)
	return out
}

// anaExpectStreak 独立实现连续性期望值（基于本地日期加减一天）。
func anaExpectStreak(dates []string) (maxStreak int, start, end string, curStreak int, curStart string) {
	if len(dates) == 0 {
		return 0, "", "", 0, ""
	}
	maxStreak, curStreak = 1, 1
	start, end = dates[0], dates[0]
	curStart = dates[0]
	for i := 1; i < len(dates); i++ {
		prev, _ := time.Parse("2006-01-02", dates[i-1])
		cur, _ := time.Parse("2006-01-02", dates[i])
		if cur.Equal(prev.AddDate(0, 0, 1)) {
			curStreak++
			if curStreak > maxStreak {
				maxStreak = curStreak
				start = dates[i-maxStreak+1]
				end = dates[i]
			}
		} else {
			curStreak = 1
			curStart = dates[i]
		}
	}
	return
}

func approxEq(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

// ==================== 纯函数测试 ====================

func TestAnaExtractKeywords(t *testing.T) {
	cases := []struct {
		name  string
		title string
		want  []string
	}{
		{"empty", "", nil},
		{"spaces-only", "   ", nil},
		{"single-rune-words-dropped", "a b c 1", nil},
		{"one-long-word", "hello", []string{"hello"}},
		{"cjk-run", "哔哩哔哩", []string{"哔哩哔哩"}},
		{"separators", "AI-Review|2020【测试】", []string{"AI", "Review", "2020", "测试"}},
		{"parens-brackets", "视频 (上) [下] · 合集 & 番剧—终",
			[]string{"视频", "合集", "番剧"}}, // 单字 "上"/"下"/"终" 应被丢弃
		{"quotes-and-punct", "「引用」(测试)", []string{"引用", "测试"}},
		{"pure-punct", "!!!???", nil},
		{"mixed-alpha-digit", "Go101 Rust", []string{"Go101", "Rust"}},
		{"trailing-separator", "结尾测试-", []string{"结尾测试"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractKeywords(tc.title)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v want %v", got, tc.want)
				}
			}
		})
	}
}

func TestAnaSortMapByValueDesc(t *testing.T) {
	if got := sortMapByValueDesc(map[string]int{}); len(got) != 0 {
		t.Fatalf("empty map: got %v", got)
	}
	got := sortMapByValueDesc(map[string]int{"a": 1, "b": 5, "c": -3, "d": 5})
	if len(got) != 4 {
		t.Fatalf("len %d", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].value < got[i].value {
			t.Fatalf("not desc: %+v", got)
		}
	}
	if got[0].value != 5 || got[3].value != -3 {
		t.Fatalf("edges wrong: %+v", got)
	}
	if got[3].key != "c" {
		t.Fatalf("min should be c: %+v", got)
	}
	single := sortMapByValueDesc(map[string]int{"only": 42})
	if len(single) != 1 || single[0].key != "only" || single[0].value != 42 {
		t.Fatalf("single: %+v", single)
	}
}

func TestAnaSqrt(t *testing.T) {
	cases := []struct {
		x, want float64
	}{
		{-5, 0}, {0, 0}, {1, 1}, {4, 2}, {2, math.Sqrt2}, {0.25, 0.5}, {100, 10},
	}
	for _, c := range cases {
		got := sqrt(c.x)
		tol := math.Abs(c.want) * 1e-6
		if tol < 1e-9 {
			tol = 1e-9
		}
		if !approxEq(got, c.want, tol) {
			t.Fatalf("sqrt(%v)=%v want %v", c.x, got, c.want)
		}
	}
}

func TestAnaRoundFloat(t *testing.T) {
	cases := []struct {
		val       float64
		precision int
		want      float64
	}{
		{0, 2, 0},
		{1.234, 2, 1.23},
		{1.236, 2, 1.24},
		{2.5, 0, 3},
		{1.49, 0, 1},
		{-1.234, 2, -1.22}, // Go int 转换向零截断 => 负数行为特殊
		{99.999, 2, 100},
		{0.005, 3, 0.005},
	}
	for _, c := range cases {
		if got := roundFloat(c.val, c.precision); !approxEq(got, c.want, 1e-9) {
			t.Fatalf("roundFloat(%v,%d)=%v want %v", c.val, c.precision, got, c.want)
		}
	}
}

func anaEntry(name string, videoCount int, avg, comp, quality float64) authorEntry {
	return authorEntry{name, AuthorCompletionStat{
		VideoCount: videoCount, AverageCompletionRate: avg,
		ComprehensiveScore: comp, QualityScore: quality,
	}}
}

func TestAnaAuthorSorts(t *testing.T) {
	build := func() []authorEntry {
		return []authorEntry{
			anaEntry("a", 3, 10, 5, 1),
			anaEntry("b", 7, 50, 90, 40),
			anaEntry("c", 1, 99, 20, 95),
			anaEntry("d", 7, 30, 55, 10),
		}
	}

	field := map[string]func(authorEntry) float64{
		"VideoCount":    func(e authorEntry) float64 { return float64(e.stats.VideoCount) },
		"Completion":    func(e authorEntry) float64 { return e.stats.AverageCompletionRate },
		"Comprehensive": func(e authorEntry) float64 { return e.stats.ComprehensiveScore },
		"Quality":       func(e authorEntry) float64 { return e.stats.QualityScore },
	}
	sorters := map[string]func([]authorEntry){
		"VideoCount":    sortByVideoCountDesc,
		"Completion":    sortByCompletionRateDesc,
		"Comprehensive": sortByComprehensiveScoreDesc,
		"Quality":       sortByQualityScoreDesc,
	}
	for name, fn := range sorters {
		get := field[name]
		t.Run(name, func(t *testing.T) {
			list := build()
			fn(list)
			for i := 1; i < len(list); i++ {
				if get(list[i-1]) < get(list[i]) {
					t.Fatalf("%s not desc: %+v", name, list)
				}
			}
			// 空列表与单元素
			fn([]authorEntry{})
			one := []authorEntry{anaEntry("z", 1, 1, 1, 1)}
			fn(one)
			if one[0].name != "z" {
				t.Fatalf("single element changed: %+v", one)
			}
		})
	}
}

func TestAnaSortIntsDesc(t *testing.T) {
	sortIntsDesc(nil)
	got := []int{-1, -5, 3, 0, 7, 7}
	sortIntsDesc(got)
	want := []int{7, 7, 3, 0, -1, -5}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	single := []int{42}
	sortIntsDesc(single)
	if single[0] != 42 {
		t.Fatalf("single changed")
	}
}

// ==================== 错误路径：表不存在 ====================

func TestAnaMissingTableErrors(t *testing.T) {
	if _, err := AnalyzeHistory(anaMissingYear); err == nil || !strings.Contains(err.Error(), fmt.Sprint(anaMissingYear)) {
		t.Fatalf("AnalyzeHistory: %v", err)
	}
	if _, err := GenerateHeatmapData(anaMissingYear); err == nil {
		t.Fatal("GenerateHeatmapData expected error")
	}
	if _, err := GetViewingAnalytics(anaMissingYear); err == nil {
		t.Fatal("GetViewingAnalytics expected error")
	}
	if _, err := AnalyzeContinuity(anaMissingYear); err == nil {
		t.Fatal("AnalyzeContinuity expected error")
	}
	if _, err := AnalyzeTimeSlots(anaMissingYear); err == nil {
		t.Fatal("AnalyzeTimeSlots expected error")
	}
	if _, err := AnalyzeWeekly(anaMissingYear); err == nil {
		t.Fatal("AnalyzeWeekly expected error")
	}
	if _, err := AnalyzeCompletionRates(anaMissingYear); err == nil {
		t.Fatal("AnalyzeCompletionRates expected error")
	}
	if _, err := AnalyzeAuthorCompletion(anaMissingYear); err == nil {
		t.Fatal("AnalyzeAuthorCompletion expected error")
	}
	if _, err := AnalyzeTagAnalysis(anaMissingYear); err == nil {
		t.Fatal("AnalyzeTagAnalysis expected error")
	}
	if _, err := AnalyzeDurationAnalysis(anaMissingYear); err == nil {
		t.Fatal("AnalyzeDurationAnalysis expected error")
	}
	if _, err := AnalyzeWatchCounts(anaMissingYear); err == nil {
		t.Fatal("AnalyzeWatchCounts expected error")
	}
	if _, err := GetViewingOverview(anaMissingYear); err == nil {
		t.Fatal("GetViewingOverview expected error")
	}
	if _, err := GetDailyCountStats(anaMissingYear, 1, 1); err == nil {
		t.Fatal("GetDailyCountStats expected error")
	}
}

// ==================== 2022：空表边界 + 单一高产UP主边界 ====================

func TestAnaEdgeYear2022(t *testing.T) {
	// --- 阶段 1：表存在但为空 ---
	if err := GetSQLiteDB().EnsureTableForYear(anaEdgeYear); err != nil {
		t.Fatalf("ensure 2022: %v", err)
	}
	if n := anaCountRows(t, anaEdgeYear); n != 0 {
		t.Fatalf("2022 should start empty, got %d", n)
	}

	res, err := AnalyzeHistory(anaEdgeYear)
	if err != nil {
		t.Fatalf("AnalyzeHistory empty: %v", err)
	}
	if res.TotalVideos != 0 || res.ActiveDays != 0 || len(res.MonthlyStats) != 0 || len(res.TagRanking) != 0 {
		t.Fatalf("empty AnalyzeHistory: %+v", res)
	}
	if len(res.WeekdayDistribution) != 7 {
		t.Fatalf("weekday dist len %d", len(res.WeekdayDistribution))
	}
	for _, w := range res.WeekdayDistribution {
		if w.Count != 0 {
			t.Fatalf("weekday %s nonzero", w.Name)
		}
	}
	if len(res.DeviceDistribution) != 0 || len(res.TitleKeywords) != 0 || len(res.TopVideos) != 0 {
		t.Fatalf("empty dists: %+v", res)
	}

	hm, err := GenerateHeatmapData(anaEdgeYear)
	if err != nil || hm.Total != 0 || len(hm.Data) != 0 || hm.Year != anaEdgeYear {
		t.Fatalf("heatmap empty: %+v %v", hm, err)
	}

	va, err := GetViewingAnalytics(anaEdgeYear)
	if err != nil || va.TotalVideos != 0 || va.TotalWatchTime != 0 || va.AvgWatchTimePerDay != 0 ||
		va.PeakDay != "" || len(va.TimeDistribution) != 0 || len(va.WeekdayDistribution) != 0 {
		t.Fatalf("viewing analytics empty: %+v %v", va, err)
	}

	cont, err := AnalyzeContinuity(anaEdgeYear)
	if err != nil || cont.MaxStreak != 0 || cont.LongestStreakPeriod.Start != "" || cont.CurrentStreak != 0 {
		t.Fatalf("continuity empty: %+v %v", cont, err)
	}

	ts, err := AnalyzeTimeSlots(anaEdgeYear)
	if err != nil {
		t.Fatalf("timeslots empty: %v", err)
	}
	if len(ts.DailyTimeSlots) != 0 || len(ts.PeakHours) != 0 || ts.MaxDailyRecord != nil ||
		ts.TimeInvestment.MaxDurationDay.Date != "" || ts.TimeInvestment.AvgDailyDuration != 0 {
		t.Fatalf("timeslots empty: %+v", ts)
	}

	wk, err := AnalyzeWeekly(anaEdgeYear)
	if err != nil || wk.ActiveDays != 0 || len(wk.SeasonalPatterns) != 0 {
		t.Fatalf("weekly empty: %+v %v", wk, err)
	}
	if len(wk.WeeklyStats) != 7 {
		t.Fatalf("weekly stats keys %d", len(wk.WeeklyStats))
	}
	for _, v := range wk.WeeklyStats {
		if v != 0 {
			t.Fatal("weekly nonzero")
		}
	}

	cr, err := AnalyzeCompletionRates(anaEdgeYear)
	if err != nil || cr.OverallStats.TotalVideos != 0 {
		t.Fatalf("completion empty: %+v %v", cr, err)
	}
	if len(cr.CompletionDistribution) != 6 {
		t.Fatalf("distribution keys %d", len(cr.CompletionDistribution))
	}
	for _, v := range cr.CompletionDistribution {
		if v != 0 {
			t.Fatal("distribution nonzero")
		}
	}
	for cat, s := range cr.DurationBasedStats {
		if s.VideoCount != 0 {
			t.Fatalf("cat %s not zero", cat)
		}
	}

	ac, err := AnalyzeAuthorCompletion(anaEdgeYear)
	if err != nil || len(ac.MostWatchedAuthors) != 0 || len(ac.HighestCompletionAuthors) != 0 ||
		len(ac.MostValuableAuthors) != 0 || len(ac.PotentialAuthors) != 0 {
		t.Fatalf("author completion empty: %+v %v", ac, err)
	}

	tg, err := AnalyzeTagAnalysis(anaEdgeYear)
	if err != nil || len(tg.TagDistribution) != 0 || len(tg.TagCompletionRates) != 0 {
		t.Fatalf("tag analysis empty: %+v %v", tg, err)
	}

	da, err := AnalyzeDurationAnalysis(anaEdgeYear)
	if err != nil || len(da.DurationCorrelation) != 4 {
		t.Fatalf("duration analysis empty: %+v %v", da, err)
	}
	for p, m := range da.DurationCorrelation {
		if len(m) != 3 {
			t.Fatalf("period %s types %d", p, len(m))
		}
		for dt, s := range m {
			if s.VideoCount != 0 || s.AvgDuration != 0 {
				t.Fatalf("period %s type %s nonzero", p, dt)
			}
		}
	}

	wc, err := AnalyzeWatchCounts(anaEdgeYear)
	if err != nil || wc.RewatchStats.TotalUniqueVideos != 0 || wc.RewatchStats.RewatchRate != 0 ||
		len(wc.MostWatchedVideos) != 0 || len(wc.DurationDistribution) != 0 || len(wc.TagDistribution) != 0 {
		t.Fatalf("watch counts empty: %+v %v", wc, err)
	}

	ov, err := GetViewingOverview(anaEdgeYear)
	if err != nil || len(ov.Report) != 0 || ov.Details.TotalDays != 0 || len(ov.Details.Devices) != 0 {
		t.Fatalf("overview empty: %+v %v", ov, err)
	}

	ds, err := GetDailyCountStats(anaEdgeYear, 1, 1)
	if err != nil || ds.TotalVideos != 0 || len(ds.Insights) != 6 {
		t.Fatalf("daily stats empty: %+v %v", ds, err)
	}

	// --- 阶段 2：单一UP主 22 行（>=20 => confidence 封顶；max==min => viewRange=1）---
	recs := make([]models.HistoryRecord, 0, 22)
	for i := 0; i < 21; i++ {
		recs = append(recs, historyRecord(
			fmt.Sprintf("BVanaH%03d", i+1), fmt.Sprintf("ana 边缘 高产 %02d", i+1),
			"archive", "ana标签22", "ana边缘区", 7201, "ana高产UP",
			atTime(anaEdgeYear, 1, i+1, 12, 0), 100, 100, 1))
	}
	recs = append(recs, historyRecord(
		"BVanaH001", "ana 边缘 高产 01", "archive", "ana标签22", "ana边缘区", 7201, "ana高产UP",
		atTime(anaEdgeYear, 2, 1, 12, 0), 100, 100, 1))
	if err := GetSQLiteDB().EnsureTableForYear(anaEdgeYear); err != nil {
		t.Fatalf("ensure 2022 again: %v", err)
	}
	if err := seedHistoryRecords(testConn(), anaEdgeYear, recs...); err != nil {
		t.Fatalf("seed 2022: %v", err)
	}
	if n := anaCountRows(t, anaEdgeYear); n != 22 {
		t.Fatalf("2022 rows %d", n)
	}

	ac2, err := AnalyzeAuthorCompletion(anaEdgeYear)
	if err != nil {
		t.Fatalf("author completion 2022: %v", err)
	}
	if len(ac2.MostWatchedAuthors) != 1 || len(ac2.PotentialAuthors) != 0 {
		t.Fatalf("2022 author maps: %d %d", len(ac2.MostWatchedAuthors), len(ac2.PotentialAuthors))
	}
	stat, ok := ac2.MostWatchedAuthors["ana高产UP"]
	if !ok {
		t.Fatalf("author missing: %+v", ac2.MostWatchedAuthors)
	}
	// avg=100, fully=100, normalized=0, confidence 22/20>1 封顶 1 => 0.25*0+0.5*100+0.25*100 = 75
	if !approxEq(stat.ComprehensiveScore, 75, 1e-9) {
		t.Fatalf("comprehensive %v want 75", stat.ComprehensiveScore)
	}
	if !approxEq(stat.QualityScore, 100, 1e-9) {
		t.Fatalf("quality %v want 100", stat.QualityScore)
	}
	// loyalty = 22 * sqrt(1) = 22
	if !approxEq(stat.LoyaltyScore, 22, 1e-6) {
		t.Fatalf("loyalty %v want 22", stat.LoyaltyScore)
	}
	if stat.VideoCount != 22 || stat.FullyWatched != 22 {
		t.Fatalf("counts %+v", stat)
	}

	wc2, err := AnalyzeWatchCounts(anaEdgeYear)
	if err != nil {
		t.Fatalf("watch counts 2022: %v", err)
	}
	if wc2.RewatchStats.TotalRewatchedVideos != 1 || wc2.RewatchStats.TotalUniqueVideos != 21 ||
		wc2.RewatchStats.TotalRewatchCount != 1 {
		t.Fatalf("rewatch stats 2022: %+v", wc2.RewatchStats)
	}
	if !approxEq(wc2.RewatchStats.RewatchRate, float64(1)/float64(21)*100, 1e-9) {
		t.Fatalf("rate %v", wc2.RewatchStats.RewatchRate)
	}
	if len(wc2.MostWatchedVideos) != 1 || wc2.MostWatchedVideos[0].Bvid != "BVanaH001" ||
		wc2.MostWatchedVideos[0].WatchCount != 2 {
		t.Fatalf("most watched 2022: %+v", wc2.MostWatchedVideos)
	}
}

func anaCountRows(t *testing.T, year int) int {
	t.Helper()
	var n int
	err := testConn().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM bilibili_history_%d", year)).Scan(&n)
	if err != nil {
		t.Fatalf("count %d: %v", year, err)
	}
	return n
}

// ==================== 2005 种子数据集：结构性检查（只读） ====================

func TestAna2005SeedStructure(t *testing.T) {
	const seedCount = 13

	if n := anaCountRows(t, yearAnalysisSeed); n != seedCount {
		t.Fatalf("seed rows %d", n)
	}

	res, err := AnalyzeHistory(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("AnalyzeHistory 2005: %v", err)
	}
	if res.Year != yearAnalysisSeed || res.TotalVideos != seedCount {
		t.Fatalf("header: %+v", res)
	}
	if res.UniqueAuthors != 7 {
		t.Fatalf("unique authors %d want 7", res.UniqueAuthors)
	}
	// TotalDuration 直接由 SQL 对照
	var expectDuration int64
	if err := testConn().QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(CASE WHEN progress = -1 THEN duration ELSE progress END),0)
		FROM bilibili_history_%d`, yearAnalysisSeed)).Scan(&expectDuration); err != nil {
		t.Fatal(err)
	}
	if int64(res.TotalDuration) != expectDuration {
		t.Fatalf("duration %d vs sql %d", res.TotalDuration, expectDuration)
	}
	// 刷片重复行 BVseed0001 两次、BVseed0003 两次：TopVideos 按时长降序
	if len(res.TopVideos) == 0 || len(res.TopVideos) > 10 {
		t.Fatalf("top videos %+v", res.TopVideos)
	}
	for i := 1; i < len(res.TopVideos); i++ {
		if res.TopVideos[i-1].Duration < res.TopVideos[i].Duration {
			t.Fatalf("top videos not sorted: %+v", res.TopVideos)
		}
	}
	for i := 1; i < len(res.TagRanking); i++ {
		if res.TagRanking[i-1].Count < res.TagRanking[i].Count {
			t.Fatal("tag ranking not sorted")
		}
	}
	if len(res.TagRanking) == 0 {
		t.Fatal("tag ranking empty")
	}
	for i := 1; i < len(res.AuthorRanking); i++ {
		if res.AuthorRanking[i-1].Count < res.AuthorRanking[i].Count {
			t.Fatal("author ranking not sorted")
		}
	}
	if len(res.AuthorRanking) != 7 {
		t.Fatalf("author ranking len %d", len(res.AuthorRanking))
	}
	// 月统计：10 月无数据
	if len(res.MonthlyStats) != 11 {
		t.Fatalf("monthly stats len %d want 11", len(res.MonthlyStats))
	}
	monthlySum := 0
	for _, m := range res.MonthlyStats {
		if m.Month == "10" {
			t.Fatal("month 10 should be absent")
		}
		monthlySum += m.TotalCount
	}
	if monthlySum != seedCount {
		t.Fatalf("monthly sum %d", monthlySum)
	}
	// 日统计
	if len(res.DailyStats) != seedCount {
		t.Fatalf("daily stats len %d want %d (每天一行)", len(res.DailyStats), seedCount)
	}
	dailySum := 0
	for _, d := range res.DailyStats {
		dailySum += d.TotalCount
	}
	if dailySum != seedCount || res.ActiveDays != len(res.DailyStats) {
		t.Fatalf("daily sum %d active %d", dailySum, res.ActiveDays)
	}
	// 设备分布总和 = 行数
	devSum := 0
	for _, c := range res.DeviceDistribution {
		devSum += c
	}
	if devSum != seedCount {
		t.Fatalf("device sum %d", devSum)
	}
	// 时长偏好
	durSum := res.DurationPreference.Short + res.DurationPreference.Mid + res.DurationPreference.Long
	var expectDurRows int
	if err := testConn().QueryRow(fmt.Sprintf(
		"SELECT COUNT(*) FROM bilibili_history_%d WHERE duration > 0", yearAnalysisSeed)).Scan(&expectDurRows); err != nil {
		t.Fatal(err)
	}
	if durSum != expectDurRows {
		t.Fatalf("duration pref %d vs %d", durSum, expectDurRows)
	}
	ratioSum := res.DurationPreference.ShortRatio + res.DurationPreference.MidRatio + res.DurationPreference.LongRatio
	if !approxEq(ratioSum, 1, 1e-9) {
		t.Fatalf("ratios %v", ratioSum)
	}
	// 关键词降序
	for i := 1; i < len(res.TitleKeywords); i++ {
		if res.TitleKeywords[i-1].Count < res.TitleKeywords[i].Count {
			t.Fatal("keywords not sorted")
		}
	}
	if len(res.TitleKeywords) == 0 || len(res.TitleKeywords) > 30 {
		t.Fatalf("keywords len %d", len(res.TitleKeywords))
	}
	if len(res.WeekdayDistribution) != 7 {
		t.Fatalf("weekday len")
	}
	wdSum := 0
	for _, w := range res.WeekdayDistribution {
		wdSum += w.Count
	}
	if wdSum != seedCount {
		t.Fatalf("weekday sum %d", wdSum)
	}

	// 热力图：13 行分布在 13 个不同日期
	hm, err := GenerateHeatmapData(yearAnalysisSeed)
	if err != nil || hm.Total != seedCount || len(hm.Data) != seedCount {
		t.Fatalf("heatmap 2005: %+v %v", hm, err)
	}
	if hm.Data["2005-01-03"] != 1 || hm.Data["2005-01-20"] != 1 {
		t.Fatalf("heatmap cells: %v", hm.Data)
	}

	// 回归：GetViewingAnalytics 曾在非空表上死锁——peakRows (LIMIT 1)
	// 未耗尽/关闭就发起下一次查询，占死 MaxOpenConns(1) 的连接。已修复。
	anaEnsure2020(t)
	var wantCount2020 int
	if err := testConn().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM bilibili_history_%d", anaYear)).Scan(&wantCount2020); err != nil {
		t.Fatal(err)
	}
	va, err := GetViewingAnalytics(anaYear)
	if err != nil {
		t.Fatalf("GetViewingAnalytics(%d) 回归: %v", anaYear, err)
	}
	if va.TotalVideos != wantCount2020 || va.PeakDay == "" || va.PeakDayCount <= 0 {
		t.Fatalf("viewing analytics 2020: %+v (want %d videos)", va, wantCount2020)
	}
	hourSum := 0
	for _, c := range va.TimeDistribution {
		hourSum += c
	}
	if hourSum != wantCount2020 {
		t.Fatalf("TimeDistribution 总和 %d != %d", hourSum, wantCount2020)
	}

	// 连续性：2005 种子日期互不相邻 => 最长 1
	cont, err := AnalyzeContinuity(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("continuity 2005: %v", err)
	}
	if cont.MaxStreak != 1 || cont.CurrentStreak != 1 {
		t.Fatalf("continuity 2005: %+v", cont)
	}
	if cont.LongestStreakPeriod.Start != "2005-01-03" || cont.LongestStreakPeriod.End != "2005-01-03" {
		t.Fatalf("longest period: %+v", cont.LongestStreakPeriod)
	}
	if cont.CurrentStreakStart != "2005-12-25" {
		t.Fatalf("current start %s", cont.CurrentStreakStart)
	}

	// 时段
	ts, err := AnalyzeTimeSlots(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("timeslots 2005: %v", err)
	}
	tsSum := 0
	for _, c := range ts.DailyTimeSlots {
		tsSum += c
	}
	if tsSum != seedCount {
		t.Fatalf("slot sum %d", tsSum)
	}
	if len(ts.PeakHours) == 0 || len(ts.PeakHours) > 5 {
		t.Fatalf("peak hours %+v", ts.PeakHours)
	}
	for i := 1; i < len(ts.PeakHours); i++ {
		if ts.PeakHours[i-1].ViewCount < ts.PeakHours[i].ViewCount {
			t.Fatal("peak hours not sorted")
		}
	}
	if ts.MaxDailyRecord == nil || ts.MaxDailyRecord.VideoCount != 1 {
		t.Fatalf("max daily %+v", ts.MaxDailyRecord)
	}

	// 周分析
	wk, err := AnalyzeWeekly(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("weekly 2005: %v", err)
	}
	wkSum := 0
	for _, c := range wk.WeeklyStats {
		wkSum += c
	}
	if wkSum != seedCount || wk.ActiveDays != seedCount {
		t.Fatalf("weekly sum %d active %d", wkSum, wk.ActiveDays)
	}
	seasonSum := 0
	for _, p := range wk.SeasonalPatterns {
		seasonSum += p.ViewCount
	}
	if seasonSum != seedCount {
		t.Fatalf("season sum %d", seasonSum)
	}

	// 完成率：13 行，progress=-1 一行(BVseed0003 首次)，duration=0 一行(article)
	crs, err := AnalyzeCompletionRates(yearAnalysisSeed)
	if err != nil || crs.OverallStats.TotalVideos != seedCount {
		t.Fatalf("completion 2005: %+v %v", crs, err)
	}
	distSum := 0
	for _, c := range crs.CompletionDistribution {
		distSum += c
	}
	if distSum != seedCount {
		t.Fatalf("completion dist sum %d", distSum)
	}
	if crs.OverallStats.NotStartedCount != 2 {
		// progress=0 的音乐行 + duration=0 的专栏行（completionRate=0）
		t.Fatalf("not started %d want 2", crs.OverallStats.NotStartedCount)
	}

	// 作者完成率：2005 无 >=5 部作者 => 全空
	ac, err := AnalyzeAuthorCompletion(yearAnalysisSeed)
	if err != nil || len(ac.MostWatchedAuthors) != 0 || len(ac.HighestCompletionAuthors) != 0 ||
		len(ac.MostValuableAuthors) != 0 || len(ac.PotentialAuthors) != 0 {
		t.Fatalf("author completion 2005: %+v %v", ac, err)
	}

	// 标签分析：无 >=5 部标签；科技标签应为 4
	tg, err := AnalyzeTagAnalysis(yearAnalysisSeed)
	if err != nil || len(tg.TagCompletionRates) != 0 {
		t.Fatalf("tag completion 2005: %+v %v", tg, err)
	}
	if tg.TagDistribution["科技"] != 4 {
		t.Fatalf("tag 科技 %d", tg.TagDistribution["科技"])
	}

	// 时长分析
	da, err := AnalyzeDurationAnalysis(yearAnalysisSeed)
	if err != nil || len(da.DurationCorrelation) != 4 {
		t.Fatalf("duration 2005: %+v %v", da, err)
	}

	// 重复观看：BVseed0001 与 BVseed0003 各 2 次
	wca, err := AnalyzeWatchCounts(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("watch counts 2005: %v", err)
	}
	if wca.RewatchStats.TotalRewatchedVideos != 2 || wca.RewatchStats.TotalUniqueVideos != 11 ||
		wca.RewatchStats.TotalRewatchCount != 2 {
		t.Fatalf("rewatch 2005: %+v", wca.RewatchStats)
	}
	if !approxEq(wca.RewatchStats.RewatchRate, 2.0/11.0*100, 1e-9) {
		t.Fatalf("rate %v", wca.RewatchStats.RewatchRate)
	}
	if len(wca.MostWatchedVideos) != 2 {
		t.Fatalf("most watched 2005: %+v", wca.MostWatchedVideos)
	}

	// 总览
	ov, err := GetViewingOverview(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("overview 2005: %v", err)
	}
	if ov.Details.TotalDays != seedCount {
		t.Fatalf("total days %d", ov.Details.TotalDays)
	}
	if len(ov.Report) != 5 {
		t.Fatalf("report keys %+v", ov.Report)
	}
	if len(ov.Details.TopCategories) != 8 || len(ov.Details.Devices) == 0 {
		t.Fatalf("overview details: cats %d devices %d", len(ov.Details.TopCategories), len(ov.Details.Devices))
	}

	// 单日统计：2005-01-03 有 1 行（BVseed0001 首次，progress=duration=900）
	ds, err := GetDailyCountStats(yearAnalysisSeed, 1, 3)
	if err != nil {
		t.Fatalf("daily 2005-01-03: %v", err)
	}
	if ds.TotalVideos != 1 || ds.TotalCount != 1 || ds.UniqueAuthors != 1 || ds.CompletedVideos != 1 {
		t.Fatalf("daily: %+v", ds)
	}
	if ds.TotalWatchSeconds != 900 {
		t.Fatalf("watch seconds %d", ds.TotalWatchSeconds)
	}
	if !approxEq(ds.AvgCompletionRate, 100, 1e-9) {
		t.Fatalf("avg completion %v", ds.AvgCompletionRate)
	}
	// 空日：2005-01-04
	ds2, err := GetDailyCountStats(yearAnalysisSeed, 1, 4)
	if err != nil || ds2.TotalVideos != 0 || ds2.Date != "2005-01-04" {
		t.Fatalf("daily empty: %+v %v", ds2, err)
	}
}

// ==================== 2020 主数据集：全函数验证 ====================

func TestAna2020AnalyzeHistory(t *testing.T) {
	anaEnsure2020(t)
	rows := anaRows2020()

	res, err := AnalyzeHistory(anaYear)
	if err != nil {
		t.Fatalf("AnalyzeHistory 2020: %v", err)
	}
	if res.Year != anaYear || res.TotalVideos != len(rows) {
		t.Fatalf("header: %+v", res)
	}

	var wantDuration int64
	for _, r := range rows {
		wantDuration += int64(anaEff(r))
	}
	if int64(res.TotalDuration) != wantDuration {
		t.Fatalf("duration %d want %d", res.TotalDuration, wantDuration)
	}
	if !approxEq(res.AvgDuration, float64(wantDuration)/float64(len(rows)), 1e-6) {
		t.Fatalf("avg %v", res.AvgDuration)
	}
	if res.UniqueAuthors != 7 {
		t.Fatalf("unique authors %d", res.UniqueAuthors)
	}

	// tag ranking：ana知识=10 居首
	if len(res.TagRanking) == 0 {
		t.Fatal("tag ranking empty")
	}
	for i := 1; i < len(res.TagRanking); i++ {
		if res.TagRanking[i-1].Count < res.TagRanking[i].Count {
			t.Fatal("tags not sorted")
		}
	}
	if res.TagRanking[0].TagName != "ana知识" || res.TagRanking[0].Count != 10 {
		t.Fatalf("top tag %+v", res.TagRanking[0])
	}
	for _, tc := range res.TagRanking {
		if tc.TagName == "" {
			t.Fatal("empty tag in ranking")
		}
	}

	if len(res.AuthorRanking) == 0 || res.AuthorRanking[0].AuthorName != "ana大UP" || res.AuthorRanking[0].Count != 10 {
		t.Fatalf("top author %+v", res.AuthorRanking)
	}

	// 月份：6 月无数据 => 11 条
	if len(res.MonthlyStats) != 11 {
		t.Fatalf("monthly len %d", len(res.MonthlyStats))
	}
	for _, m := range res.MonthlyStats {
		if m.Month == "06" {
			t.Fatal("June should be empty")
		}
	}

	// 日统计与活跃天数
	wantDates := anaUniqueDates(rows)
	if res.ActiveDays != len(wantDates) || len(res.DailyStats) != len(wantDates) {
		t.Fatalf("active %d dates %d want %d", res.ActiveDays, len(res.DailyStats), len(wantDates))
	}
	for i := range res.DailyStats {
		if res.DailyStats[i].TagDistribution == nil || res.DailyStats[i].AuthorDistribution == nil {
			t.Fatal("daily maps must be initialized")
		}
	}

	devSum := 0
	for _, c := range res.DeviceDistribution {
		devSum += c
	}
	if devSum != len(rows) {
		t.Fatalf("device sum %d", devSum)
	}
	// dt: 1/5/7/33 -> 手机，2 -> 电脑，4/6 -> 平板，其他 -> 其他（getDeviceName）
	for _, key := range []string{"手机", "电脑", "平板", "其他"} {
		if res.DeviceDistribution[key] == 0 {
			t.Fatalf("device %s missing", key)
		}
	}

	// 时长偏好：仅统计 duration>0（M006 duration=0 除外）
	dp := res.DurationPreference
	short, mid, long := 0, 0, 0
	total := 0
	for _, r := range rows {
		if r.duration <= 0 {
			continue
		}
		total++
		switch {
		case r.duration <= 300:
			short++
		case r.duration <= 1200:
			mid++
		default:
			long++
		}
	}
	if dp.Short != short || dp.Mid != mid || dp.Long != long {
		t.Fatalf("duration pref %+v want %d/%d/%d", dp, short, mid, long)
	}
	if !approxEq(dp.ShortRatio+dp.MidRatio+dp.LongRatio, 1, 1e-9) {
		t.Fatalf("ratios %+v", dp)
	}

	// TopVideos：duration>0 的有 33 行 => 截断 10，降序
	if len(res.TopVideos) != 10 {
		t.Fatalf("top len %d", len(res.TopVideos))
	}
	for i := 1; i < len(res.TopVideos); i++ {
		if res.TopVideos[i-1].Duration < res.TopVideos[i].Duration {
			t.Fatal("top not sorted")
		}
	}
	if res.TopVideos[0].Duration != 2400 {
		t.Fatalf("max duration %d", res.TopVideos[0].Duration)
	}

	// 关键词：全部计数为正且降序、截断 30
	for i := 1; i < len(res.TitleKeywords); i++ {
		if res.TitleKeywords[i-1].Count < res.TitleKeywords[i].Count {
			t.Fatal("keywords not sorted")
		}
	}
	if len(res.TitleKeywords) == 0 || len(res.TitleKeywords) > 30 {
		t.Fatalf("keywords len %d", len(res.TitleKeywords))
	}
	kwMap := map[string]int{}
	for _, k := range res.TitleKeywords {
		if k.Count <= 0 {
			t.Fatalf("nonpositive keyword %+v", k)
		}
		kwMap[k.Word] = k.Count
	}
	if kwMap["ana"] != len(rows) {
		t.Fatalf("keyword ana=%d want %d", kwMap["ana"], len(rows))
	}
}

func TestAna2020HeatmapAndAnalytics(t *testing.T) {
	anaEnsure2020(t)
	rows := anaRows2020()

	hm, err := GenerateHeatmapData(anaYear)
	if err != nil {
		t.Fatalf("heatmap 2020: %v", err)
	}
	if hm.Total != len(rows) {
		t.Fatalf("heatmap total %d", hm.Total)
	}
	wantDates := anaUniqueDates(rows)
	if len(hm.Data) != len(wantDates) {
		t.Fatalf("heatmap days %d want %d", len(hm.Data), len(wantDates))
	}
	if hm.Data["2020-01-03"] != 3 {
		t.Fatalf("2020-01-03 count %d want 3", hm.Data["2020-01-03"])
	}
	// GetViewingAnalytics(anaYear) 的非空回归见 TestAna2005SeedStructure。
}

func TestAna2020ContinuityAndWeekly(t *testing.T) {
	anaEnsure2020(t)
	rows := anaRows2020()
	wantDates := anaUniqueDates(rows)
	wantMax, wantStart, wantEnd, wantCur, wantCurStart := anaExpectStreak(wantDates)

	// 验证数据集本身：最长应为 3/10-3/14 的 5 连，尾部 12/29-12/30 的 2 连
	if wantMax != 5 || wantStart != "2020-03-10" || wantEnd != "2020-03-14" || wantCur != 2 || wantCurStart != "2020-12-29" {
		t.Fatalf("dataset changed: %v %s %s %v %s", wantMax, wantStart, wantEnd, wantCur, wantCurStart)
	}

	cont, err := AnalyzeContinuity(anaYear)
	if err != nil {
		t.Fatalf("continuity 2020: %v", err)
	}
	if cont.MaxStreak != wantMax || cont.LongestStreakPeriod.Start != wantStart ||
		cont.LongestStreakPeriod.End != wantEnd || cont.CurrentStreak != wantCur ||
		cont.CurrentStreakStart != wantCurStart {
		t.Fatalf("continuity %+v want %+v", cont, wantMax)
	}

	wk, err := AnalyzeWeekly(anaYear)
	if err != nil {
		t.Fatalf("weekly 2020: %v", err)
	}
	if len(wk.WeeklyStats) != 7 {
		t.Fatalf("weekly keys %d", len(wk.WeeklyStats))
	}
	total := 0
	for _, c := range wk.WeeklyStats {
		total += c
	}
	if total != len(rows) {
		t.Fatalf("weekly total %d", total)
	}
	if wk.ActiveDays != len(wantDates) {
		t.Fatalf("active days %d want %d", wk.ActiveDays, len(wantDates))
	}
	// 按 SQL 规则（1-3 春季，4-6 夏季，7-9 秋季，10-12 冬季）推导
	seasonCount := map[string]int{}
	seasonDur := map[string]int64{}
	for _, r := range rows {
		var s string
		switch {
		case r.month <= 3:
			s = "春季"
		case r.month <= 6:
			s = "夏季"
		case r.month <= 9:
			s = "秋季"
		default:
			s = "冬季"
		}
		seasonCount[s]++
		seasonDur[s] += int64(anaEff(r))
	}
	if len(wk.SeasonalPatterns) != len(seasonCount) {
		t.Fatalf("seasons %d want %d", len(wk.SeasonalPatterns), len(seasonCount))
	}
	for s, c := range seasonCount {
		p, ok := wk.SeasonalPatterns[s]
		if !ok || p.ViewCount != c {
			t.Fatalf("season %s got %+v want %d", s, p, c)
		}
		want := float64(seasonDur[s]) / float64(c)
		if !approxEq(p.AvgDuration, want, 1e-6) {
			t.Fatalf("season %s avg %v want %v", s, p.AvgDuration, want)
		}
	}

	ts, err := AnalyzeTimeSlots(anaYear)
	if err != nil {
		t.Fatalf("timeslots 2020: %v", err)
	}
	// 小时分布
	hourCount := map[int]int{}
	dayCount := map[string]int{}
	dayDur := map[string]int64{}
	for _, r := range rows {
		hourCount[r.hour]++
		d := anaDate(r)
		dayCount[d]++
		dayDur[d] += int64(anaEff(r))
	}
	sum := 0
	for k, c := range ts.DailyTimeSlots {
		sum += c
		_ = k
	}
	if sum != len(rows) {
		t.Fatalf("slot sum %d", sum)
	}
	if ts.DailyTimeSlots["15时"] != 2 {
		t.Fatalf("15时 %d want 2", ts.DailyTimeSlots["15时"])
	}
	if len(ts.PeakHours) == 0 || len(ts.PeakHours) > 5 {
		t.Fatalf("peak hours %+v", ts.PeakHours)
	}
	for i := 1; i < len(ts.PeakHours); i++ {
		if ts.PeakHours[i-1].ViewCount < ts.PeakHours[i].ViewCount {
			t.Fatal("peak not sorted")
		}
	}
	// 单日最长时长：2020-01-03（600+1800+60=2460）
	wantMaxDurDay, wantMaxDur := "", int64(-1)
	for _, d := range anaUniqueDates(rows) {
		if dayDur[d] > wantMaxDur {
			wantMaxDur, wantMaxDurDay = dayDur[d], d
		}
	}
	if ts.TimeInvestment.MaxDurationDay.Date != wantMaxDurDay ||
		ts.TimeInvestment.MaxDurationDay.TotalDuration != int(wantMaxDur) ||
		ts.TimeInvestment.MaxDurationDay.VideoCount != dayCount[wantMaxDurDay] {
		t.Fatalf("max duration day %+v want %s(%d)", ts.TimeInvestment.MaxDurationDay, wantMaxDurDay, wantMaxDur)
	}
	var avgSum int64
	for _, d := range anaUniqueDates(rows) {
		avgSum += dayDur[d]
	}
	if !approxEq(ts.TimeInvestment.AvgDailyDuration, float64(avgSum)/float64(len(dayDur)), 1e-6) {
		t.Fatalf("avg daily %v", ts.TimeInvestment.AvgDailyDuration)
	}
	// 单日最多视频：2020-01-03 的 3 个（唯一）
	if ts.MaxDailyRecord == nil || ts.MaxDailyRecord.Date != "2020-01-03" || ts.MaxDailyRecord.VideoCount != 3 {
		t.Fatalf("max daily %+v", ts.MaxDailyRecord)
	}
}

func TestAna2020CompletionAndAuthors(t *testing.T) {
	anaEnsure2020(t)
	rows := anaRows2020()

	crs, err := AnalyzeCompletionRates(anaYear)
	if err != nil {
		t.Fatalf("completion 2020: %v", err)
	}
	if crs.OverallStats.TotalVideos != len(rows) {
		t.Fatalf("total %d", crs.OverallStats.TotalVideos)
	}
	var sumRate float64
	fully, notStarted := 0, 0
	buckets := map[string]int{}
	cats := map[string]int{}
	for _, r := range rows {
		rate := anaCompletion(r.duration, r.progress)
		sumRate += rate
		if rate >= 90 {
			fully++
		} else if rate == 0 {
			notStarted++
		}
		switch {
		case rate <= 10:
			buckets["0-10%"]++
		case rate <= 30:
			buckets["10-30%"]++
		case rate <= 50:
			buckets["30-50%"]++
		case rate <= 70:
			buckets["50-70%"]++
		case rate <= 90:
			buckets["70-90%"]++
		default:
			buckets["90-100%"]++
		}
		switch {
		case r.duration <= 300:
			cats["短视频(≤5分钟)"]++
		case r.duration <= 1200:
			cats["中等视频(5-20分钟)"]++
		default:
			cats["长视频(>20分钟)"]++
		}
	}
	if crs.OverallStats.FullyWatchedCount != fully || crs.OverallStats.NotStartedCount != notStarted {
		t.Fatalf("overall %+v want fully=%d not=%d", crs.OverallStats, fully, notStarted)
	}
	if !approxEq(crs.OverallStats.AverageCompletionRate, sumRate/float64(len(rows)), 1e-6) {
		t.Fatalf("avg rate %v", crs.OverallStats.AverageCompletionRate)
	}
	if !approxEq(crs.OverallStats.FullyWatchedRate, float64(fully)/float64(len(rows))*100, 1e-6) {
		t.Fatalf("fully rate %v", crs.OverallStats.FullyWatchedRate)
	}
	for k, v := range buckets {
		if crs.CompletionDistribution[k] != v {
			t.Fatalf("bucket %s got %d want %d", k, crs.CompletionDistribution[k], v)
		}
	}
	for k, v := range cats {
		if crs.DurationBasedStats[k].VideoCount != v {
			t.Fatalf("cat %s got %d want %d", k, crs.DurationBasedStats[k].VideoCount, v)
		}
	}
	// progress=-1 => 100% 归入 90-100%；超进度 200% 同
	if buckets["90-100%"] == 0 {
		t.Fatal("expected some >=90 rows")
	}

	// 作者完成率
	ac, err := AnalyzeAuthorCompletion(anaYear)
	if err != nil {
		t.Fatalf("author completion 2020: %v", err)
	}
	// 过滤后：大UP10 / 游戏7 / 音乐7 / 星5，共 4；少UP3、直播1、无名1 被过滤
	if len(ac.MostWatchedAuthors) != 4 || len(ac.HighestCompletionAuthors) != 4 || len(ac.MostValuableAuthors) != 4 {
		t.Fatalf("map lens %d %d %d", len(ac.MostWatchedAuthors), len(ac.HighestCompletionAuthors), len(ac.MostValuableAuthors))
	}
	big, ok := ac.MostWatchedAuthors["ana大UP"]
	if !ok || big.VideoCount != 10 || big.AuthorMid != 7101 {
		t.Fatalf("big up %+v", big)
	}
	if _, ok := ac.MostWatchedAuthors["ana少UP"]; ok {
		t.Fatal("few-video author should be filtered")
	}
	// potential：threshold=7（[10,7,7,5] 第 3 名），仅 ana星UP count5<7 且 quality≈99.9>85
	if len(ac.PotentialAuthors) != 1 {
		t.Fatalf("potential %+v", ac.PotentialAuthors)
	}
	star, ok := ac.PotentialAuthors["ana星UP"]
	if !ok {
		t.Fatalf("potential missing star: %+v", ac.PotentialAuthors)
	}
	if star.VideoCount != 5 || star.QualityScore <= 85 {
		t.Fatalf("star %+v", star)
	}
	// 星UP 完成率：100, 99.667, 100, 100, 100 => avg≈99.933，fully=100%
	wantAvg := (100 + 299.0/300*100 + 100 + 100 + 100) / 5
	if !approxEq(star.AverageCompletionRate, wantAvg, 1e-6) {
		t.Fatalf("star avg %v want %v", star.AverageCompletionRate, wantAvg)
	}
	if !approxEq(star.FullyWatchedRate, 100, 1e-9) {
		t.Fatalf("star fully rate %v", star.FullyWatchedRate)
	}
	// normalized=(5-5)/(10-5)*100=0 => comp=(0*.25+avg*.5+100*.25)*(5/20)
	confidence := 5.0 / 20.0
	wantComp := (0 + wantAvg*0.5 + 100*0.25) * confidence
	if !approxEq(star.ComprehensiveScore, roundFloat(wantComp, 2), 1e-9) {
		t.Fatalf("star comp %v want %v", star.ComprehensiveScore, wantComp)
	}

	// 标签分析
	tg, err := AnalyzeTagAnalysis(anaYear)
	if err != nil {
		t.Fatalf("tag 2020: %v", err)
	}
	wantDist := map[string]int{"ana知识": 10, "ana游戏": 6, "ana音乐": 7, "ana科技": 5, "ana生活": 3, "ana直播": 1, "ana其他": 1}
	if len(tg.TagDistribution) != len(wantDist) {
		t.Fatalf("tag dist %+v want %+v", tg.TagDistribution, wantDist)
	}
	for k, v := range wantDist {
		if tg.TagDistribution[k] != v {
			t.Fatalf("tag %s got %d want %d", k, tg.TagDistribution[k], v)
		}
	}
	// >=5 部才有完成率统计：知识10/音乐7/科技5（游戏 6）
	if len(tg.TagCompletionRates) != 4 {
		t.Fatalf("completion rates %+v", tg.TagCompletionRates)
	}
	if _, ok := tg.TagCompletionRates["ana生活"]; ok {
		t.Fatal("3-video tag should be filtered")
	}
	tech := tg.TagCompletionRates["ana科技"]
	if tech.VideoCount != 5 || tech.FullyWatched != 5 {
		t.Fatalf("tech tag %+v", tech)
	}
}

func TestAna2020DurationAndWatchCounts(t *testing.T) {
	anaEnsure2020(t)
	rows := anaRows2020()

	da, err := AnalyzeDurationAnalysis(anaYear)
	if err != nil {
		t.Fatalf("duration 2020: %v", err)
	}
	if len(da.DurationCorrelation) != 4 {
		t.Fatalf("periods %d", len(da.DurationCorrelation))
	}
	// 期望推导（与 production 分桶一致：<300 短、<1200 中、否则长）
	type key struct{ p, t string }
	want := map[key]DurationTypeStat{}
	for _, r := range rows {
		if r.duration <= 0 {
			continue
		}
		var period string
		switch {
		case r.hour < 6:
			period = "凌晨"
		case r.hour < 12:
			period = "上午"
		case r.hour < 18:
			period = "下午"
		default:
			period = "晚上"
		}
		var dtype string
		switch {
		case r.duration < 300:
			dtype = "短视频"
		case r.duration < 1200:
			dtype = "中等视频"
		default:
			dtype = "长视频"
		}
		s := want[key{period, dtype}]
		s.VideoCount++
		s.TotalDuration += float64(r.duration)
		want[key{period, dtype}] = s
	}
	for k, s := range want {
		got := da.DurationCorrelation[k.p][k.t]
		if got.VideoCount != s.VideoCount || !approxEq(got.TotalDuration, s.TotalDuration, 1e-6) {
			t.Fatalf("%v/%v got %+v want %+v", k.p, k.t, got, s)
		}
		if !approxEq(got.AvgDuration, s.TotalDuration/float64(s.VideoCount), 1e-6) {
			t.Fatalf("%v/%v avg %v", k.p, k.t, got.AvgDuration)
		}
	}
	// 每个时段/类型组合初始化必须存在（计数为 0 的组合保留默认结构）
	for _, p := range []string{"凌晨", "上午", "下午", "晚上"} {
		if len(da.DurationCorrelation[p]) != 3 {
			t.Fatalf("period %s types %d", p, len(da.DurationCorrelation[p]))
		}
	}

	wca, err := AnalyzeWatchCounts(anaYear)
	if err != nil {
		t.Fatalf("watch counts 2020: %v", err)
	}
	// B001/B002/B003 各 2 次；行数 34 => 去重后 31
	rs := wca.RewatchStats
	if rs.TotalRewatchedVideos != 3 || rs.TotalUniqueVideos != 31 || rs.TotalRewatchCount != 3 {
		t.Fatalf("rewatch stats %+v", rs)
	}
	if !approxEq(rs.RewatchRate, 3.0/31.0*100, 1e-9) {
		t.Fatalf("rate %v", rs.RewatchRate)
	}
	if len(wca.MostWatchedVideos) != 3 {
		t.Fatalf("most watched %+v", wca.MostWatchedVideos)
	}
	seen := map[string]MostWatchedVideo{}
	for _, v := range wca.MostWatchedVideos {
		if v.WatchCount != 2 {
			t.Fatalf("watch count %d for %s", v.WatchCount, v.Bvid)
		}
		if v.AvgInterval <= 0 || v.LastView <= v.FirstView {
			t.Fatalf("interval %+v", v)
		}
		if v.TagName != "ana知识" || v.AuthorName != "ana大UP" {
			t.Fatalf("meta %+v", v)
		}
		seen[v.Bvid] = v
	}
	for _, bvid := range []string{"BVanaB001", "BVanaB002", "BVanaB003"} {
		if _, ok := seen[bvid]; !ok {
			t.Fatalf("missing %s in %+v", bvid, seen)
		}
	}
	if seen["BVanaB001"].FirstView != atTime(anaYear, 1, 1, 2, 0) ||
		seen["BVanaB001"].LastView != atTime(anaYear, 10, 10, 11, 0) {
		t.Fatalf("b001 views %+v", seen["BVanaB001"])
	}
	// 时长分布：B001(120) 短、B002(300) 短（<=300）、B003(1200) 中
	if wca.DurationDistribution["短视频(≤5分钟)"] != 2 ||
		wca.DurationDistribution["中等视频(5-20分钟)"] != 1 ||
		wca.DurationDistribution["长视频(>20分钟)"] != 0 {
		t.Fatalf("duration dist %+v", wca.DurationDistribution)
	}
	if wca.TagDistribution["ana知识"] != 3 {
		t.Fatalf("tag dist %+v", wca.TagDistribution)
	}
}

func TestAna2020OverviewAndDaily(t *testing.T) {
	anaEnsure2020(t)
	rows := anaRows2020()

	ov, err := GetViewingOverview(anaYear)
	if err != nil {
		t.Fatalf("overview 2020: %v", err)
	}
	var wantWatch int64
	daySet := map[string]bool{}
	slotDays := map[string]map[string]bool{}
	catCount := map[string]int{}
	for _, r := range rows {
		wantWatch += int64(anaEff(r))
		d := anaDate(r)
		daySet[d] = true
		slot := anaOverviewSlot(r.hour)
		if slotDays[slot] == nil {
			slotDays[slot] = map[string]bool{}
		}
		slotDays[slot][d] = true
		catCount[r.main]++
	}
	if !approxEq(ov.Details.TotalWatchHours, float64(wantWatch)/3600.0, 1e-6) {
		t.Fatalf("hours %v want %v", ov.Details.TotalWatchHours, float64(wantWatch)/3600.0)
	}
	if ov.Details.TotalDays != len(daySet) {
		t.Fatalf("days %d want %d", ov.Details.TotalDays, len(daySet))
	}
	if len(ov.Details.TopCategories) != len(catCount) {
		t.Fatalf("categories %+v", ov.Details.TopCategories)
	}
	if ov.Details.TopCategories[0].Category != "ana综合" || ov.Details.TopCategories[0].ViewCount != 10 {
		t.Fatalf("top category %+v", ov.Details.TopCategories[0])
	}
	for i := 1; i < len(ov.Details.TopCategories); i++ {
		if ov.Details.TopCategories[i-1].ViewCount < ov.Details.TopCategories[i].ViewCount {
			t.Fatal("categories not sorted")
		}
	}
	if len(ov.Details.FavoriteUpUsers) != 7 || ov.Details.FavoriteUpUsers[0].Name != "ana大UP" {
		t.Fatalf("ups %+v", ov.Details.FavoriteUpUsers)
	}
	// 时段活动：天数 <= 总天数，百分比之和可 >100（同一天可跨时段）
	for slot, act := range ov.Details.TimeSlotActivity {
		want, ok := slotDays[slot]
		if !ok || act.Days != len(want) {
			t.Fatalf("slot %s got %+v want %d", slot, act, len(want))
		}
		if !approxEq(act.Percentage, float64(len(want))/float64(len(daySet))*100, 1e-6) {
			t.Fatalf("slot %s pct %v", slot, act.Percentage)
		}
	}
	// 设备平台共 5 类（手机24 / 网页4 / 平板4 / 电视1 / 其他1），LIMIT 3 => 手机 + 两个并列 4
	if len(ov.Details.Devices) != 3 {
		t.Fatalf("devices %+v", ov.Details.Devices)
	}
	if ov.Details.Devices[0].Name != "手机" || ov.Details.Devices[0].Count != 24 {
		t.Fatalf("top device %+v", ov.Details.Devices[0])
	}
	for i := 1; i < len(ov.Details.Devices); i++ {
		if ov.Details.Devices[i-1].Count < ov.Details.Devices[i].Count {
			t.Fatal("devices not sorted")
		}
	}
	if ov.Details.Devices[1].Count != 4 || ov.Details.Devices[2].Count != 4 {
		t.Fatalf("devices tail %+v", ov.Details.Devices)
	}
	// 5 个 report 键都应有
	for _, k := range []string{"total_summary", "time_slot_summary", "category_summary", "up_summary", "device_summary"} {
		if _, ok := ov.Report[k]; !ok {
			t.Fatalf("report key %s missing: %+v", k, ov.Report)
		}
	}
	if !strings.Contains(ov.Report["category_summary"], "ana综合") {
		t.Fatalf("category summary %q", ov.Report["category_summary"])
	}

	// 单日：2020-01-03 三行 B003(1200/600) B004(1800/1800) X001(60/60)
	ds, err := GetDailyCountStats(anaYear, 1, 3)
	if err != nil {
		t.Fatalf("daily 2020-01-03: %v", err)
	}
	if ds.Date != "2020-01-03" || ds.TotalVideos != 3 || ds.TotalCount != 3 || ds.UniqueAuthors != 2 {
		t.Fatalf("daily %+v", ds)
	}
	if ds.TypeCounts["archive"] != 2 || ds.TypeCounts["other"] != 1 {
		t.Fatalf("type counts %+v", ds.TypeCounts)
	}
	if ds.TotalWatchSeconds != 600+1800+60 {
		t.Fatalf("seconds %d", ds.TotalWatchSeconds)
	}
	if ds.CompletedVideos != 2 {
		t.Fatalf("completed %d", ds.CompletedVideos)
	}
	if !approxEq(ds.AvgDuration, (1200+1800+60)/3.0, 1e-6) {
		t.Fatalf("avg duration %v", ds.AvgDuration)
	}
	if !approxEq(ds.AvgCompletionRate, (0.5+1.0+1.0)/3.0*100, 1e-6) {
		t.Fatalf("avg completion %v", ds.AvgCompletionRate)
	}
	if ds.TagDistribution["ana知识"] != 2 || ds.TagDistribution["ana其他"] != 1 {
		t.Fatalf("tag dist %+v", ds.TagDistribution)
	}
	if ds.AuthorDistribution["ana大UP"] != 2 || ds.AuthorDistribution["ana无名"] != 1 {
		t.Fatalf("author dist %+v", ds.AuthorDistribution)
	}
	if len(ds.Insights) != 6 {
		t.Fatalf("insights %+v", ds.Insights)
	}

	// 空日 2020-06-15
	ds2, err := GetDailyCountStats(anaYear, 6, 15)
	if err != nil {
		t.Fatalf("daily empty: %v", err)
	}
	if ds2.TotalVideos != 0 || ds2.TotalCount != 0 || ds2.TotalWatchSeconds != 0 || len(ds2.TypeCounts) != 0 {
		t.Fatalf("daily empty %+v", ds2)
	}
	if len(ds2.Insights) != 6 || !strings.Contains(ds2.Insights[0], "0 个视频") {
		t.Fatalf("insights %+v", ds2.Insights)
	}
}

// anaOverviewSlot 复刻 GetViewingOverview 的时段 SQL：5-11 上午 / 12-17 下午 / 18-22 晚上 / 其余深夜。
func anaOverviewSlot(hour int) string {
	switch {
	case hour >= 5 && hour <= 11:
		return "上午"
	case hour >= 12 && hour <= 17:
		return "下午"
	case hour >= 18 && hour <= 22:
		return "晚上"
	default:
		return "深夜"
	}
}
