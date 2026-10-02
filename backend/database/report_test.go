package database

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"bilibili-history-go/models"
)

// report_test.go 覆盖 report.go。
// 命名空间约定：本文件独占主库年份表 2021（bvid 前缀 BVrep），1997 用作"无表"分支。
// 2005 种子表只读，仅做结构性断言（其内容由 TestMain 固定，可安全断言确定性事实）。

const (
	repYear     = 2021
	repYearNone = 1997
)

var repSeedDone bool

// repV 构造一个内存中的 ReportVideo，避免依赖数据库。
func repV(bvid, title, cat, tag, author string, mid int64, viewAt int64, dur, prog, dt, fin int) ReportVideo {
	return ReportVideo{
		Title:        title,
		Bvid:         bvid,
		AuthorName:   author,
		AuthorMid:    mid,
		MainCategory: cat,
		TagName:      tag,
		Duration:     dur,
		Progress:     prog,
		ViewAt:       viewAt,
		Dt:           dt,
		IsFinish:     fin,
		Business:     "archive",
	}
}

func repApprox(t *testing.T, name string, got, want float64) {
	t.Helper()
	d := got - want
	if d < 0 {
		d = -d
	}
	if d > 1e-9 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// repFindBucket 在 CompletionDist 中查找区间计数，不存在返回 -1。
func repFindBucket(dist []CompletionDistItem, label string) int {
	for _, d := range dist {
		if d.Range == label {
			return d.Count
		}
	}
	return -1
}

func repFindCat(stats []CategoryStat, name string) int {
	for _, s := range stats {
		if s.Name == name {
			return s.Count
		}
	}
	return -1
}

func repHasKeyword(stats []KeywordStat, word string) bool {
	return repFindKeywordCount(stats, word) > 0
}

func repFindKeywordCount(stats []KeywordStat, word string) int {
	for _, k := range stats {
		if k.Word == word {
			return k.Count
		}
	}
	return 0
}

func repEqualInts(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// ---------- getWeekRange ----------

func TestRepGetWeekRange(t *testing.T) {
	cases := []struct {
		year, week         int
		wantStart, wantEnd string
	}{
		// 2021-01-04 是周一：week1Monday = 1 月 4 日
		{2021, 1, "2021-01-04", "2021-01-10"},
		{2021, 2, "2021-01-11", "2021-01-17"},
		{2021, 9, "2021-03-01", "2021-03-07"},
		{2021, 53, "2022-01-03", "2022-01-09"},
		// 2005-01-04 是周二：week1Monday = 1 月 3 日
		{2005, 1, "2005-01-03", "2005-01-09"},
		{2005, 2, "2005-01-10", "2005-01-16"},
		// 2015-01-04 是周日（weekday==0 -> 7 分支）：week1Monday = 2014-12-29
		{2015, 1, "2014-12-29", "2015-01-04"},
		// 2018-12-31 是周一：2018-01-04 周四 -> week1Monday = 1 月 1 日；第 53 周恰好 12-31 开始
		{2018, 53, "2018-12-31", "2019-01-06"},
		{2018, 52, "2018-12-24", "2018-12-30"},
	}
	for _, c := range cases {
		start, end := getWeekRange(c.year, c.week)
		if got := start.Format("2006-01-02"); got != c.wantStart {
			t.Errorf("getWeekRange(%d,%d) start = %s, want %s", c.year, c.week, got, c.wantStart)
		}
		if got := end.Format("2006-01-02"); got != c.wantEnd {
			t.Errorf("getWeekRange(%d,%d) end = %s, want %s", c.year, c.week, got, c.wantEnd)
		}
		if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
			t.Errorf("getWeekRange(%d,%d) start time = %v, want midnight", c.year, c.week, start)
		}
		if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
			t.Errorf("getWeekRange(%d,%d) end time = %v, want 23:59:59", c.year, c.week, end)
		}
	}
}

// ---------- getDeviceName ----------

func TestRepGetDeviceName(t *testing.T) {
	cases := map[int]string{
		1: "手机", 3: "手机", 5: "手机", 7: "手机", 33: "手机",
		2: "电脑",
		4: "平板", 6: "平板",
		0: "其他", 8: "其他", -1: "其他", 99: "其他",
	}
	for dt, want := range cases {
		if got := getDeviceName(dt); got != want {
			t.Errorf("getDeviceName(%d) = %q, want %q", dt, got, want)
		}
	}
}

// ---------- computeSummary：空输入 ----------

func TestRepComputeSummaryEmpty(t *testing.T) {
	s := computeSummary(nil, 7)
	if s.TotalVideos != 0 || s.TotalDuration != 0 {
		t.Errorf("empty summary: total = %d/%d", s.TotalVideos, s.TotalDuration)
	}
	if s.DeviceDist == nil || s.HourDist == nil {
		t.Error("empty summary must still allocate DeviceDist/HourDist maps")
	}
	if len(s.DeviceDist) != 0 || len(s.HourDist) != 0 {
		t.Errorf("empty summary maps should be empty: %v %v", s.DeviceDist, s.HourDist)
	}
	if s.TopVideos != nil || s.CompletionDist != nil || s.TitleKeywords != nil {
		t.Error("empty summary should have nil slices")
	}
	// dayCount=0 不应导致除零
	s2 := computeSummary([]ReportVideo{repV("BVrepX1", "标题", "", "", "up", 1, 1614550000, 0, 0, 1, 0)}, 0)
	if s2.AvgDailyVideos != 0 || s2.AvgDailyDuration != 0 {
		t.Errorf("dayCount=0 should keep avg fields at 0, got %v/%v", s2.AvgDailyVideos, s2.AvgDailyDuration)
	}
}

// ---------- computeSummary：全零时长 ----------

func TestRepComputeSummaryZeroDurations(t *testing.T) {
	vs := []ReportVideo{
		repV("BVrepZ1", "甲 标题", "分类", "", "upA", 1, atTime(repYear, 3, 1, 9, 0), 0, 0, 1, 0),
		repV("BVrepZ2", "乙 标题", "", "标签", "upB", 2, atTime(repYear, 3, 1, 9, 30), 0, -1, 2, 1),
	}
	s := computeSummary(vs, 1)
	if s.TotalDuration != 0 {
		t.Errorf("TotalDuration = %d, want 0", s.TotalDuration)
	}
	if s.CompletionStats.Finished != 0 || s.CompletionStats.Partial != 0 || s.CompletionStats.AvgRate != 0 {
		t.Errorf("duration=0 rows must be skipped by completion: %+v", s.CompletionStats)
	}
	if s.CompletionDist != nil {
		t.Errorf("CompletionDist should stay nil: %v", s.CompletionDist)
	}
	// 时长 0 < 300 归入 short
	if s.DurationPref.Short != 2 || s.DurationPref.Mid != 0 || s.DurationPref.Long != 0 {
		t.Errorf("DurationPref = %+v", s.DurationPref)
	}
	repApprox(t, "DurationPref.ShortRatio", s.DurationPref.ShortRatio, 1.0)
	// 分类回退：main 为空时用 tag；两者皆空则跳过（此处两行都有值）
	if len(s.TopCategories) != 2 {
		t.Fatalf("TopCategories = %v, want 2 entries", s.TopCategories)
	}
	if s.TopCategories[0].Count != 1 || s.TopCategories[1].Count != 1 {
		t.Errorf("TopCategories counts wrong: %v", s.TopCategories)
	}
	// is_finish=1 计入 FavoriteRate
	repApprox(t, "FavoriteRate", s.FavoriteRate, 0.5)
	// duration=0 不计弃看
	repApprox(t, "AbandonRate", s.AbandonRate, 0)
	if s.UniqueDays != 1 || s.DailyBreakdown[0].Count != 2 {
		t.Errorf("daily breakdown wrong: %+v", s.DailyBreakdown)
	}
	if s.DailyBreakdown[0].UniqueUp != 2 {
		t.Errorf("UniqueUp = %d, want 2", s.DailyBreakdown[0].UniqueUp)
	}
}

// ---------- computeSummary：完成度分桶 + 弃看率边界 ----------

func TestRepComputeSummaryCompletionBuckets(t *testing.T) {
	at := atTime(repYear, 3, 1, 10, 0)
	vs := []ReportVideo{
		repV("BVrepC1", "看完A", "c", "", "u", 1, at, 100, 100, 1, 0), // rate=1.0 看完
		repV("BVrepC2", "看完B", "c", "", "u", 1, at, 100, -1, 1, 0),  // progress=-1 -> 看完
		repV("BVrepC3", "九十五", "c", "", "u", 1, at, 100, 95, 1, 0),  // 80-100%，finished(>=0.9)
		repV("BVrepC4", "九十", "c", "", "u", 1, at, 100, 90, 1, 0),   // 恰好 0.9 -> finished / 80-100%
		repV("BVrepC5", "七十", "c", "", "u", 1, at, 100, 70, 1, 0),   // 60-80%，partial
		repV("BVrepC6", "五十", "c", "", "u", 1, at, 100, 50, 1, 0),   // 40-60%
		repV("BVrepC7", "三十", "c", "", "u", 1, at, 100, 30, 1, 0),   // 20-40%
		repV("BVrepC8", "十", "c", "", "u", 1, at, 100, 10, 1, 0),    // 0-20%
		repV("BVrepC9", "二十边界", "c", "", "u", 1, at, 100, 20, 1, 0), // 恰好 0.2 -> 20-40%，不算弃看
		repV("BVrepC10", "弃看", "c", "", "u", 1, at, 100, 19, 1, 0),  // 0.19 -> 弃看
	}
	s := computeSummary(vs, 1)
	if s.CompletionStats.Finished != 4 || s.CompletionStats.Partial != 6 {
		t.Errorf("finished/partial = %d/%d, want 4/6", s.CompletionStats.Finished, s.CompletionStats.Partial)
	}
	want := map[string]int{"看完": 2, "80-100%": 2, "60-80%": 1, "40-60%": 1, "20-40%": 2, "0-20%": 2}
	for label, cnt := range want {
		if got := repFindBucket(s.CompletionDist, label); got != cnt {
			t.Errorf("bucket %s = %d, want %d (dist=%v)", label, got, cnt, s.CompletionDist)
		}
	}
	if len(s.CompletionDist) != 6 || s.CompletionDist[0].Range != "0-20%" || s.CompletionDist[5].Range != "看完" {
		t.Errorf("CompletionDist order wrong: %v", s.CompletionDist)
	}
	// avg rate = (1+1+0.95+0.9+0.7+0.5+0.3+0.1+0.2+0.19)/10
	repApprox(t, "AvgRate", s.CompletionStats.AvgRate, 5.84/10)
	// 弃看：<0.2 的仅 C8(0.1) 与 C10(0.19)，progress=-1 不计
	repApprox(t, "AbandonRate", s.AbandonRate, 2.0/float64(len(vs)))
}

// ---------- computeSummary：深夜 / 时段 / 黄金时段（含跨零点环形）/ 星期 ----------

func TestRepComputeSummaryTimeDistribution(t *testing.T) {
	vs := []ReportVideo{
		repV("BVrepT1", "深夜一", "c", "", "u", 1, atTime(repYear, 3, 1, 23, 30), 100, 50, 1, 0),
		repV("BVrepT2", "凌晨二", "c", "", "u", 1, atTime(repYear, 3, 2, 0, 15), 100, 50, 1, 0),
		repV("BVrepT3", "凌晨三", "c", "", "u", 1, atTime(repYear, 3, 2, 5, 45), 100, 50, 1, 0),
		repV("BVrepT4", "白天四", "c", "", "u", 1, atTime(repYear, 3, 3, 12, 0), 100, 50, 1, 0),
	}
	s := computeSummary(vs, 4)
	repApprox(t, "LateNightRatio", s.LateNightRatio, 3.0/4.0)
	if s.HourDist[23] != 1 || s.HourDist[0] != 1 || s.HourDist[5] != 1 || s.HourDist[12] != 1 {
		t.Errorf("HourDist wrong: %v", s.HourDist)
	}
	if len(s.TopTimeSlots) != 4 {
		t.Errorf("TopTimeSlots should always contain 4 slots, got %v", s.TopTimeSlots)
	}
	if s.TopTimeSlots[0].Name != "深夜(22-6)" || s.TopTimeSlots[0].Count != 3 {
		t.Errorf("top slot wrong: %v", s.TopTimeSlots)
	}
	// 黄金时段：最大连续 3 小时窗口为 (23,0,1) 计 2 -> 2/4
	repApprox(t, "GoldenSlotRatio", s.GoldenSlotRatio, 2.0/4.0)

	// 环形跨零点：23/0/1 各一条 -> ratio 1.0
	vs2 := []ReportVideo{
		repV("BVrepT5", "a", "c", "", "u", 1, atTime(repYear, 3, 1, 23, 0), 100, 50, 1, 0),
		repV("BVrepT6", "a", "c", "", "u", 1, atTime(repYear, 3, 2, 0, 0), 100, 50, 1, 0),
		repV("BVrepT7", "a", "c", "", "u", 1, atTime(repYear, 3, 3, 1, 0), 100, 50, 1, 0),
	}
	s2 := computeSummary(vs2, 3)
	repApprox(t, "wrap GoldenSlotRatio", s2.GoldenSlotRatio, 1.0)

	// 星期分布：2021-03-01 周一 … 2021-03-07 周日
	wd := []ReportVideo{}
	for d := 1; d <= 7; d++ {
		wd = append(wd, repV(fmt.Sprintf("BVrepW%d", d), "w", "c", "", "u", 1, atTime(repYear, 3, d, 10, 0), 60, 60, 1, 0))
	}
	s3 := computeSummary(wd, 7)
	if len(s3.WeekdayDist) != 7 {
		t.Fatalf("WeekdayDist len = %d, want 7", len(s3.WeekdayDist))
	}
	wantNames := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	for i, st := range s3.WeekdayDist {
		if st.Name != wantNames[i] {
			t.Errorf("WeekdayDist[%d].Name = %s, want %s", i, st.Name, wantNames[i])
		}
		if st.Count != 1 {
			t.Errorf("WeekdayDist[%s] = %d, want 1", st.Name, st.Count)
		}
	}
}

// ---------- computeSummary：设备分布 / 刷片 / 榜单截断 / 多样性 ----------

func TestRepComputeSummaryAggregations(t *testing.T) {
	vs := []ReportVideo{}
	dts := []int{1, 3, 5, 7, 33, 2, 4, 6, 9}
	for i, dt := range dts {
		vs = append(vs, repV(fmt.Sprintf("BVrepD%d", i), "设备 测试", "主分类", "", "up", int64(100+i), atTime(repYear, 3, 4, 10, i), 60, 60, dt, 0))
	}
	s := computeSummary(vs, 1)
	if s.DeviceDist["手机"] != 5 || s.DeviceDist["电脑"] != 1 || s.DeviceDist["平板"] != 2 || s.DeviceDist["其他"] != 1 {
		t.Errorf("DeviceDist wrong: %v", s.DeviceDist)
	}

	// 刷片：7 个 bvid 各 2 次 + 1 个 bvid 3 次 -> TotalRewatched=17，列表截断为 5
	rw := []ReportVideo{}
	for i := 0; i < 7; i++ {
		bvid := fmt.Sprintf("BVrepR%d", i)
		rw = append(rw,
			repV(bvid, "重看 标题", "c", "", "up", 1, atTime(repYear, 4, 1, 10, i), 100, 100, 1, 0),
			repV(bvid, "重看 标题", "c", "", "up", 1, atTime(repYear, 4, 2, 10, i), 100, 100, 1, 0),
		)
	}
	for j := 0; j < 3; j++ {
		rw = append(rw, repV("BVrepR9", "三刷 标题", "c", "", "up", 1, atTime(repYear, 4, 3, 10, j), 100, 100, 1, 0))
	}
	rs := computeSummary(rw, 3)
	if rs.RewatchStats.TotalRewatched != 17 {
		t.Errorf("TotalRewatched = %d, want 17", rs.RewatchStats.TotalRewatched)
	}
	if len(rs.RewatchStats.RewatchedVideos) != 5 {
		t.Errorf("rewatch list should truncate to 5, got %d", len(rs.RewatchStats.RewatchedVideos))
	}
	if rs.RewatchStats.RewatchedVideos[0].Bvid != "BVrepR9" || rs.RewatchStats.RewatchedVideos[0].Count != 3 {
		t.Errorf("top rewatch should be the 3-play video: %+v", rs.RewatchStats.RewatchedVideos[0])
	}
	if rs.RewatchStats.RewatchedVideos[0].TotalDur != 300 {
		t.Errorf("rewatch TotalDur = %d, want 300", rs.RewatchStats.RewatchedVideos[0].TotalDur)
	}
	if rs.RewatchStats.RewatchedVideos[1].Count != 2 {
		t.Errorf("second rewatch count = %d, want 2", rs.RewatchStats.RewatchedVideos[1].Count)
	}
	// 只看过一次的 bvid 不进刷片榜
	for _, rv := range rs.RewatchStats.RewatchedVideos {
		if rv.Bvid == "BVrepD0" {
			t.Errorf("single-play video should not appear in rewatch list")
		}
	}

	// 榜单截断：12 个分类 / 12 位作者 / 14 个视频（TopVideos<=5）
	mixed := []ReportVideo{}
	for i := 0; i < 12; i++ {
		mid := int64(200 + i)
		name := fmt.Sprintf("作者%02d", i)
		mixed = append(mixed, repV(fmt.Sprintf("BVrepM%d", i), "混合 测试", fmt.Sprintf("分类%02d", i), "", name, mid, atTime(repYear, 5, 1+i%28, 8+i%12, 0), 100+i, 100, 1, 0))
	}
	mixed = append(mixed,
		repV("BVrepM90", "混合 测试 加长", "分类00", "", "作者00", 200, atTime(repYear, 5, 3, 9, 0), 5000, 5000, 2, 0),
		repV("BVrepM91", "混合 测试 加长", "分类01", "", "作者01", 201, atTime(repYear, 5, 4, 9, 0), 4000, 4000, 2, 0),
	)
	ms := computeSummary(mixed, 31)
	if len(ms.TopCategories) != 10 {
		t.Errorf("TopCategories len = %d, want 10", len(ms.TopCategories))
	}
	if !sort.SliceIsSorted(ms.TopCategories, func(i, j int) bool { return ms.TopCategories[i].Count > ms.TopCategories[j].Count }) {
		t.Errorf("TopCategories not sorted desc: %v", ms.TopCategories)
	}
	if len(ms.TopAuthors) != 10 {
		t.Errorf("TopAuthors len = %d, want 10", len(ms.TopAuthors))
	}
	if ms.TopAuthors[0].Count != 2 {
		t.Errorf("top author count = %d, want 2: %+v", ms.TopAuthors[0].Count, ms.TopAuthors[0])
	}
	if ms.TopAuthors[0].Name == "作者00" && ms.TopAuthors[0].Duration != 100+5000 {
		t.Errorf("top author duration = %d, want 5100", ms.TopAuthors[0].Duration)
	}
	if len(ms.TopVideos) != 5 {
		t.Errorf("TopVideos len = %d, want 5", len(ms.TopVideos))
	}
	if ms.TopVideos[0].Bvid != "BVrepM90" || ms.TopVideos[0].Duration != 5000 {
		t.Errorf("top video should be longest: %+v", ms.TopVideos[0])
	}
	if ms.UniqueAuthors != 12 || ms.NewUpCount != 12 {
		t.Errorf("UniqueAuthors/NewUpCount = %d/%d, want 12/12", ms.UniqueAuthors, ms.NewUpCount)
	}
	repApprox(t, "UpDiversity", ms.UpDiversity, 12.0/14.0)

	// 同名不同 mid：authorMap 以名字聚合（覆盖 Count++ 分支）
	dup := []ReportVideo{
		repV("BVrepP1", "p", "c", "", "同名UP", 701, atTime(repYear, 6, 1, 10, 0), 60, 60, 1, 0),
		repV("BVrepP2", "p", "c", "", "同名UP", 702, atTime(repYear, 6, 1, 11, 0), 60, 60, 1, 0),
	}
	ds := computeSummary(dup, 1)
	if ds.TopAuthors[0].Count != 2 || ds.TopAuthors[0].Mid != 701 {
		t.Errorf("same-name aggregation wrong: %+v", ds.TopAuthors[0])
	}
	if ds.UniqueAuthors != 2 {
		t.Errorf("UniqueAuthors by mid = %d, want 2", ds.UniqueAuthors)
	}
	if ds.DailyBreakdown[0].UniqueUp != 2 {
		t.Errorf("daily UniqueUp = %d, want 2", ds.DailyBreakdown[0].UniqueUp)
	}
	if len(ds.TopCategories) != 1 || ds.TopCategories[0].Name != "c" {
		t.Errorf("cat c should be the single entry: %v", ds.TopCategories)
	}
	// 空分类 + 空 tag 应被跳过
	none := computeSummary([]ReportVideo{repV("BVrepP3", "x", "", "", "u", 1, atTime(repYear, 6, 2, 10, 0), 60, 60, 1, 0)}, 1)
	if len(none.TopCategories) != 0 {
		t.Errorf("empty cat/tag must be skipped: %v", none.TopCategories)
	}
}

// ---------- computeSummary：标题关键词 ----------

func TestRepComputeSummaryKeywords(t *testing.T) {
	// 1) 两条同题视频 => 重复出现的中文片段计数 x2；停用词被过滤
	dup := repV("BVrepK1", "一个人和一个东西", "c", "", "u", 1, atTime(repYear, 3, 1, 10, 0), 100, 100, 1, 0)
	s := computeSummary([]ReportVideo{dup, dup}, 1)
	if len(s.TitleKeywords) == 0 {
		t.Fatal("expected repeated keywords to be extracted")
	}
	if !repHasKeyword(s.TitleKeywords, "东西") {
		t.Errorf("expected keyword 东西, got %v", s.TitleKeywords)
	}
	if repHasKeyword(s.TitleKeywords, "一个") {
		t.Errorf("stopword 一个 should be filtered out")
	}
	// 计数阈值：仅出现 1 次的短语不入选
	once := computeSummary([]ReportVideo{dup}, 1)
	for _, k := range once.TitleKeywords {
		if k.Count < 2 {
			t.Errorf("keyword %q count %d < 2 leaked", k.Word, k.Count)
		}
	}

	// 2) 截断到 top 20：25 个双字词拼成长标题，两条同题视频
	words := []string{"测试", "视频", "动画", "游戏", "音乐", "知识", "科技", "生活", "搞笑", "美食",
		"旅行", "健身", "读书", "电影", "综艺", "体育", "历史", "地理", "物理", "化学",
		"生物", "英语", "数学", "编程", "设计"}
	long := strings.Join(words, "")
	v1 := repV("BVrepK2", long, "c", "", "u", 1, atTime(repYear, 3, 1, 10, 0), 100, 100, 1, 0)
	v2 := repV("BVrepK3", long, "c", "", "u", 1, atTime(repYear, 3, 2, 10, 0), 100, 100, 1, 0)
	ls := computeSummary([]ReportVideo{v1, v2}, 2)
	if len(ls.TitleKeywords) != 20 {
		t.Errorf("TitleKeywords should truncate to 20, got %d", len(ls.TitleKeywords))
	}
	if !sort.SliceIsSorted(ls.TitleKeywords, func(i, j int) bool { return ls.TitleKeywords[i].Count > ls.TitleKeywords[j].Count }) {
		t.Error("TitleKeywords not sorted by count desc")
	}

	// 3) 非中文字符被剔除
	ascii := repV("BVrepK4", "Hello World 123", "c", "", "u", 1, atTime(repYear, 3, 1, 10, 0), 100, 100, 1, 0)
	as := computeSummary([]ReportVideo{ascii, ascii}, 1)
	if len(as.TitleKeywords) != 0 {
		t.Errorf("ascii titles should yield no keywords: %v", as.TitleKeywords)
	}
}

// ---------- 2021 精确数据集（主库，本文件独占） ----------

// repSeed2021 幂等建表并写入 BVrep* 数据（重复调用只写一次）。
func repSeed2021(t *testing.T) {
	t.Helper()
	// 测试约定禁用 t.Parallel，单线程顺序执行，bool 守卫即可保证幂等。
	if repSeedDone {
		return
	}
	repSeedDone = true
	r1 := historyRecord("BVrep0001", "测试 知识 视频", "archive", "", "知识", 6001, "UP甲", atTime(repYear, 3, 1, 10, 0), 100, 100, 2)
	r1.IsFinish = 1
	r2 := historyRecord("BVrep0002", "音乐 现场 演出", "archive", "音乐", "", 6002, "UP乙", atTime(repYear, 3, 2, 23, 0), 200, 10, 1)
	r3 := historyRecord("BVrep0002", "音乐 现场 演出", "archive", "音乐", "", 6002, "UP乙", atTime(repYear, 3, 3, 15, 0), 200, -1, 4)
	r4 := historyRecord("BVrep0003", "番剧 完结 撒花", "pgc", "番剧", "番剧", 6003, "UP丙", atTime(repYear, 3, 4, 3, 0), 3000, 2700, 33)
	r5 := historyRecord("BVrep0004", "直播 房间 测试", "live", "直播", "直播", 6004, "主播丁", atTime(repYear, 3, 5, 20, 0), 600, 600, 1)
	r6 := historyRecord("BVrep0005", "删除 视频 测试", "archive", "科技", "科技", 6005, "UP戊", atTime(repYear, 5, 10, 12, 0), 500, 500, 1)
	r6.Status = 1
	jan := make([]models.HistoryRecord, 0, 4)
	for d := 1; d <= 4; d++ {
		jan = append(jan, historyRecord(fmt.Sprintf("BVrepJan%d", d), "跨年 零点 视频", "archive", "年终", "年终", int64(6006+d), "UP己", atTime(repYear, 1, d, 12, 0), 60, 60, 2))
	}
	mustSeedHistory(t, repYear, r1, r2, r3, r4, r5, r6)
	mustSeedHistory(t, repYear, jan...)
}

func TestRepMonthlyAndWeeklyExact(t *testing.T) {
	repSeed2021(t)

	// 三月：4 条可见（live 与 status=1 被排除）
	m, err := GetMonthlyReport(repYear, 3)
	if err != nil {
		t.Fatalf("GetMonthlyReport: %v", err)
	}
	s := m.Summary
	if s.TotalVideos != 4 {
		t.Fatalf("March TotalVideos = %d, want 4 (videos=%v)", s.TotalVideos, m.Videos)
	}
	if s.TotalDuration != 3500 || s.UniqueDays != 4 || s.UniqueAuthors != 3 || s.NewUpCount != 3 {
		t.Errorf("core stats wrong: %+v", s)
	}
	if s.CompletionStats.Finished != 3 || s.CompletionStats.Partial != 1 {
		t.Errorf("finished/partial = %+v", s.CompletionStats)
	}
	repApprox(t, "AvgRate", s.CompletionStats.AvgRate, 2.95/4)
	if got := repFindBucket(s.CompletionDist, "看完"); got != 2 {
		t.Errorf("看完 bucket = %d, want 2", got)
	}
	if got := repFindBucket(s.CompletionDist, "0-20%"); got != 1 {
		t.Errorf("0-20%% bucket = %d, want 1", got)
	}
	if got := repFindBucket(s.CompletionDist, "80-100%"); got != 1 {
		t.Errorf("80-100%% bucket = %d, want 1", got)
	}
	repApprox(t, "LateNightRatio", s.LateNightRatio, 0.5)
	repApprox(t, "FavoriteRate", s.FavoriteRate, 0.25)
	repApprox(t, "AbandonRate", s.AbandonRate, 0.25)
	repApprox(t, "UpDiversity", s.UpDiversity, 0.75)
	if s.DurationPref.Short != 3 || s.DurationPref.Mid != 0 || s.DurationPref.Long != 1 {
		t.Errorf("DurationPref = %+v", s.DurationPref)
	}
	repApprox(t, "ShortRatio", s.DurationPref.ShortRatio, 0.75)
	repApprox(t, "LongRatio", s.DurationPref.LongRatio, 0.25)
	if s.DeviceDist["手机"] != 2 || s.DeviceDist["电脑"] != 1 || s.DeviceDist["平板"] != 1 {
		t.Errorf("DeviceDist = %v", s.DeviceDist)
	}
	if s.RewatchStats.TotalRewatched != 2 || len(s.RewatchStats.RewatchedVideos) != 1 {
		t.Errorf("rewatch = %+v", s.RewatchStats)
	}
	if s.RewatchStats.RewatchedVideos[0].Bvid != "BVrep0002" || s.RewatchStats.RewatchedVideos[0].TotalDur != 400 {
		t.Errorf("rewatch video = %+v", s.RewatchStats.RewatchedVideos[0])
	}
	if got := repFindCat(s.TopCategories, "音乐"); got != 2 { // tag 回退聚合两条
		t.Errorf("cat 音乐 = %d, want 2 (cats=%v)", got, s.TopCategories)
	}
	if s.TopVideos[0].Bvid != "BVrep0003" {
		t.Errorf("top video = %s, want BVrep0003", s.TopVideos[0].Bvid)
	}
	if len(s.WeekdayDist) != 7 || s.WeekdayDist[0].Count != 1 { // 2021-03-01 周一
		t.Errorf("WeekdayDist = %v", s.WeekdayDist)
	}
	if repFindKeywordCount(s.TitleKeywords, "音乐") != 2 {
		t.Errorf("TitleKeywords missing 音乐=2: %v", s.TitleKeywords)
	}
	repApprox(t, "monthly AvgDailyVideos", s.AvgDailyVideos, 4.0/31.0)
	repApprox(t, "monthly AvgDailyDuration", s.AvgDailyDuration, 3500.0/31.0)
	if len(s.DailyBreakdown) != 4 {
		t.Errorf("DailyBreakdown len = %d, want 4", len(s.DailyBreakdown))
	}
	// videos 按 view_at DESC
	if m.Videos[0].Bvid != "BVrep0003" || m.Videos[3].Bvid != "BVrep0001" {
		t.Errorf("videos order wrong: %v %v", m.Videos[0].Bvid, m.Videos[3].Bvid)
	}

	// 第九周（2021-03-01 ~ 03-07）
	w, err := GetWeeklyReport(repYear, 9)
	if err != nil {
		t.Fatalf("GetWeeklyReport: %v", err)
	}
	if w.StartDate != "2021-03-01" || w.EndDate != "2021-03-07" {
		t.Errorf("week range = %s ~ %s", w.StartDate, w.EndDate)
	}
	if w.Summary.TotalVideos != 4 {
		t.Errorf("week9 TotalVideos = %d, want 4", w.Summary.TotalVideos)
	}
	repApprox(t, "weekly AvgDailyVideos", w.Summary.AvgDailyVideos, 4.0/7.0)

	// 一月：4 条跨年视频
	janRep, err := GetMonthlyReport(repYear, 1)
	if err != nil {
		t.Fatalf("GetMonthlyReport jan: %v", err)
	}
	if janRep.Summary.TotalVideos != 4 {
		t.Errorf("Jan TotalVideos = %d, want 4", janRep.Summary.TotalVideos)
	}

	// 五月：唯一一行 status=1 被排除
	mayRep, err := GetMonthlyReport(repYear, 5)
	if err != nil {
		t.Fatalf("GetMonthlyReport may: %v", err)
	}
	if mayRep.Summary.TotalVideos != 0 {
		t.Errorf("May should exclude status=1 rows, got %d", mayRep.Summary.TotalVideos)
	}

	// 六月：表存在但无数据 -> 空摘要
	jun, err := GetMonthlyReport(repYear, 6)
	if err != nil {
		t.Fatalf("GetMonthlyReport jun: %v", err)
	}
	// Videos must be an empty slice, not nil: JSON null breaks the client-side
	// `videos.length === 0` empty state.
	if jun.Summary.TotalVideos != 0 || jun.Summary.DeviceDist == nil || jun.Videos == nil || len(jun.Videos) != 0 {
		t.Errorf("empty month wrong: total=%d videos=%v", jun.Summary.TotalVideos, jun.Videos)
	}
}

// ---------- GetAvailableReportWeeks / Months ----------

func TestRepAvailableWeeksMonths(t *testing.T) {
	repSeed2021(t)
	conn := testConn()

	// 探测一月各日的 %W，复刻生产 SQL 的 week-0 -> 52 重映射
	weekSet := map[int]bool{}
	for d := 1; d <= 4; d++ {
		var sw string
		ts := atTime(repYear, 1, d, 12, 0)
		if err := conn.QueryRow(`SELECT strftime('%W', datetime(?, 'unixepoch', 'localtime'))`, ts).Scan(&sw); err != nil {
			t.Fatalf("probe %%W for jan %d: %v", d, err)
		}
		n, err := strconv.Atoi(strings.TrimSpace(sw))
		if err != nil {
			t.Fatalf("parse %%W %q: %v", sw, err)
		}
		if n == 0 {
			weekSet[52] = true
		} else {
			weekSet[n] = true
		}
	}
	// 三月 status=0 各行（含 live 行，周查询不排除 business）都在第 9 周
	weekSet[9] = true
	want := make([]int, 0, len(weekSet))
	for w := range weekSet {
		want = append(want, w)
	}
	sort.Ints(want)

	weeks, err := GetAvailableReportWeeks(repYear)
	if err != nil {
		t.Fatalf("GetAvailableReportWeeks: %v", err)
	}
	if !repEqualInts(weeks, want) {
		t.Errorf("weeks(%d) = %v, want %v", repYear, weeks, want)
	}
	// 五月的唯一行 status=1 被排除；live 行仍计入月份
	months, err := GetAvailableReportMonths(repYear)
	if err != nil {
		t.Fatalf("GetAvailableReportMonths: %v", err)
	}
	if !repEqualInts(months, []int{1, 3}) {
		t.Errorf("months(%d) = %v, want [1 3]", repYear, months)
	}
}

func TestRepAvailableNoTable(t *testing.T) {
	weeks, err := GetAvailableReportWeeks(repYearNone)
	if err != nil {
		t.Fatalf("weeks for %d: %v", repYearNone, err)
	}
	if len(weeks) != 0 {
		t.Errorf("weeks for missing table = %v, want empty", weeks)
	}
	months, err := GetAvailableReportMonths(repYearNone)
	if err != nil {
		t.Fatalf("months for %d: %v", repYearNone, err)
	}
	if len(months) != 0 {
		t.Errorf("months for missing table = %v, want empty", months)
	}
	// 无表年份的报告应返回空结果而非错误
	rep, err := GetMonthlyReport(repYearNone, 7)
	if err != nil {
		t.Fatalf("GetMonthlyReport(%d): %v", repYearNone, err)
	}
	if rep.Summary.TotalVideos != 0 || len(rep.Videos) != 0 {
		t.Errorf("missing-table report not empty: %+v", rep)
	}
	wr, err := GetWeeklyReport(repYearNone, 1)
	if err != nil {
		t.Fatalf("GetWeeklyReport(%d): %v", repYearNone, err)
	}
	if wr.Summary.TotalVideos != 0 {
		t.Errorf("missing-table weekly report not empty")
	}
}

// ---------- 包内小工具 ----------

func TestRepPtrInt(t *testing.T) {
	p := ptrInt(42)
	if p == nil || *p != 42 {
		t.Errorf("ptrInt(42) = %v", p)
	}
}

// ---------- 2005 只读种子：结构性检查 ----------

func TestRepSeedYearStructural(t *testing.T) {
	// 六月只有 live + article 两行，均被 report 查询排除
	jun, err := GetMonthlyReport(yearAnalysisSeed, 6)
	if err != nil {
		t.Fatalf("2005 june: %v", err)
	}
	if jun.Summary.TotalVideos != 0 {
		t.Errorf("2005 June should exclude live/article, got %d", jun.Summary.TotalVideos)
	}
	for _, v := range jun.Videos {
		if v.Business == "live" || v.Business == "article" {
			t.Errorf("excluded business leaked into report: %+v", v)
		}
	}

	// 十二月唯一行 status=1 -> 空
	dec, err := GetMonthlyReport(yearAnalysisSeed, 12)
	if err != nil {
		t.Fatalf("2005 december: %v", err)
	}
	if dec.Summary.TotalVideos != 0 {
		t.Errorf("2005 December should be empty (status=1), got %d", dec.Summary.TotalVideos)
	}

	// 一月第一周（2005 week1Monday=1/3）包含 BVseed0001 的首刷
	w, err := GetWeeklyReport(yearAnalysisSeed, 1)
	if err != nil {
		t.Fatalf("2005 week1: %v", err)
	}
	if w.Summary.TotalVideos < 1 {
		t.Fatalf("2005 week1 should contain the Jan-3 seed row, got %d", w.Summary.TotalVideos)
	}
	if w.StartDate != "2005-01-03" || w.EndDate != "2005-01-09" {
		t.Errorf("2005 week1 range = %s ~ %s", w.StartDate, w.EndDate)
	}
	for _, v := range w.Videos {
		if !strings.HasPrefix(v.Bvid, analysisBvidPrefix) {
			t.Errorf("2005 range leaked non-seed bvid %s", v.Bvid)
		}
	}

	// 全年月度可用集合：status=0 行分布在 1-9、11 月，12 月被排除
	months, err := GetAvailableReportMonths(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("2005 months: %v", err)
	}
	if len(months) != 10 {
		t.Errorf("2005 months = %v, want 10 distinct months", months)
	}
	for _, mo := range months {
		if mo == 12 {
			t.Errorf("2005 month 12 must be excluded (status=1)")
		}
	}
	if !sort.IntsAreSorted(months) {
		t.Errorf("2005 months not sorted: %v", months)
	}

	// 周列表：非空、去重、升序、不含 0（week-0 已重映射）
	weeks, err := GetAvailableReportWeeks(yearAnalysisSeed)
	if err != nil {
		t.Fatalf("2005 weeks: %v", err)
	}
	if len(weeks) == 0 || !sort.IntsAreSorted(weeks) {
		t.Errorf("2005 weeks not sorted/nonempty: %v", weeks)
	}
	seen := map[int]bool{}
	for _, wk := range weeks {
		if wk < 1 || wk > 53 {
			t.Errorf("2005 week out of range: %v", weeks)
		}
		if seen[wk] {
			t.Errorf("duplicate week %d in %v", wk, weeks)
		}
		seen[wk] = true
	}
}
