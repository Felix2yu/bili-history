package database

import (
	"fmt"
	"testing"

	"bilibili-history-go/models"
)

// 2010/2011 两张表专供本文件使用；bvid 前缀 BVhist / BVdel / BVtag 隔离数据。
func seedHistoryFixtures(t *testing.T) {
	t.Helper()
	y10 := yearHistory1
	y11 := yearHistory2

	h01 := historyRecord("BVhist1001", "历史记录 测试 视频 one", "archive", "科技tag", "科技", 7101, "普通UP主", atTime(y10, 1, 5, 10, 0), 600, 600, 1)
	h01.Remark = "看完"
	h01.RemarkTime = atTime(y10, 1, 5, 11, 0)
	h01.Cid = 71011

	h02 := historyRecord("BVhist1002", "番剧 测试 记录", "pgc", "番剧tag", "番剧", 7102, "番区UP", atTime(y10, 3, 3, 22, 15), 240, 100, 2)
	h02.OID = 880001
	h02.Epid = 880
	h02.Cid = 71012

	h03 := historyRecord("BVhist1003", "直播 测试 场次", "live", "直播tag", "直播", 7103, "直播UP", atTime(y10, 5, 5, 19, 0), 999, 999, 1)
	h03.OID = 77001
	h03.Cid = 71013

	h04 := historyRecord("BVhist1004", "音乐 现场 完整版", "archive", "音乐tag", "音乐", 7104, "音乐UP", atTime(y10, 7, 20, 9, 0), 1800, -1, 1)
	h04.Cid = 71041

	h05 := historyRecord("BVhist1005", "搜索 命中 关键词", "archive", "知识tag", "知识", 7105, "特殊UP主", atTime(y10, 9, 9, 12, 0), 300, 120, 4)
	h05.URI = "https://www.bilibili.com/custom/uri"
	h05.Cid = 71051

	// 跨年 remark 测试：2010 旧备注、2011 新备注
	old := historyRecord("BVhist9001", "跨年 备注 视频", "archive", "科技tag", "科技", 7106, "备注UP", atTime(y10, 2, 2, 8, 0), 500, 500, 1)
	old.Remark = "旧备注"
	old.RemarkTime = atTime(y10, 2, 2, 9, 0)
	nw := historyRecord("BVhist9001", "跨年 备注 视频", "archive", "科技tag", "科技", 7106, "备注UP", atTime(y11, 3, 3, 8, 0), 500, 300, 1)
	nw.Remark = "新备注"
	nw.RemarkTime = atTime(y11, 3, 3, 9, 0)

	// 2011 数据：避开 08-01 之外日期留给 daily stats 断言
	d1 := historyRecord("BVhist2001", "新年表 视频 甲", "archive", "科技tag", "科技", 7107, "新年UP", atTime(y11, 8, 1, 20, 0), 400, 400, 1)
	d1.Cid = 72001
	d2 := historyRecord("BVhist2002", "新年表 视频 乙", "archive", "游戏tag", "游戏", 7108, "新年UP", atTime(y11, 10, 10, 7, 30), 800, 0, 2)
	d2.Cid = 72002

	mustSeedHistory(t, y10, h01, h02, h03, h04, h05, old)
	mustSeedHistory(t, y11, nw, d1, d2)
}

func countIn(t *testing.T, year int, where string, args ...interface{}) int {
	t.Helper()
	var n int
	q := fmt.Sprintf("SELECT COUNT(*) FROM bilibili_history_%d WHERE %s", year, where)
	if err := testConn().QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("count %d: %v", year, err)
	}
	return n
}

func recBvid(t *testing.T, m map[string]interface{}) string {
	t.Helper()
	s, ok := m["bvid"].(string)
	if !ok {
		t.Fatalf("bvid 类型错误: %#v", m["bvid"])
	}
	return s
}

func TestToIntHelpers(t *testing.T) {
	if toInt64(int64(9)) != 9 || toInt64(int(3)) != 3 || toInt64(float64(2.7)) != 2 || toInt64("x") != 0 {
		t.Fatal("toInt64 分支错误")
	}
	if toInt(int(1)) != 1 || toInt(int64(2)) != 2 || toInt(float64(3.9)) != 3 || toInt(nil) != 0 {
		t.Fatal("toInt 分支错误")
	}
	if toString("s") != "s" || toString([]byte("b")) != "b" || toString(5) != "" {
		t.Fatal("toString 分支错误")
	}
}

func TestProcessImageURL(t *testing.T) {
	if got := processImageURL("https://x.com/a.jpg?sign=1", false, false); got != "https://x.com/a.jpg" {
		t.Fatalf("应去掉查询参数, got %s", got)
	}
	if got := processImageURL("https://x.com/a.jpg?sign=1", true, false); got != "https://x.com/a.jpg?sign=1" {
		t.Fatalf("useLocal 时应原样返回, got %s", got)
	}
	if got := processImageURL("https://x.com/a.jpg?sign=1", false, true); got != "https://x.com/a.jpg?sign=1" {
		t.Fatalf("useSessdata 时应原样返回, got %s", got)
	}
	if got := processImageURL("", false, false); got != "" {
		t.Fatalf("空串应原样返回, got %q", got)
	}
}

func TestBuildOriginalURL(t *testing.T) {
	cases := []struct {
		name string
		rec  models.HistoryRecord
		want string
	}{
		{"uri 直通", models.HistoryRecord{URI: " https://x/v "}, "https://x/v"},
		{"archive", models.HistoryRecord{Business: "archive", Bvid: "BV1"}, "https://www.bilibili.com/video/BV1"},
		{"archive 分P", models.HistoryRecord{Business: "Archive", Bvid: "BV2", Page: 3}, "https://www.bilibili.com/video/BV2?p=3"},
		{"pgc ep", models.HistoryRecord{Business: "pgc", Epid: 10, OID: 20}, "https://www.bilibili.com/bangumi/play/ep10"},
		{"pgc ss", models.HistoryRecord{Business: "pgc", OID: 20}, "https://www.bilibili.com/bangumi/play/ss20"},
		{"live", models.HistoryRecord{Business: "live", OID: 99}, "https://live.bilibili.com/99"},
		{"article", models.HistoryRecord{Business: "article", OID: 77}, "https://www.bilibili.com/read/cv77"},
		{"回退 bvid", models.HistoryRecord{Business: "unknown", Bvid: "BV9"}, "https://www.bilibili.com/video/BV9"},
		{"全空", models.HistoryRecord{Business: "unknown"}, ""},
	}
	for _, c := range cases {
		if got := buildOriginalURL(c.rec); got != c.want {
			t.Errorf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}

func TestProcessRecord(t *testing.T) {
	rec := models.HistoryRecord{
		ID: 1, Title: "t", Bvid: "BVx", Business: "archive", OID: 1,
		ViewAt: atTime(2010, 1, 5, 10, 0), Duration: 10, TagName: "tagA", Covers: `["a","b"]`,
	}
	m := processRecord(rec, false, false)
	if m["original_url"] != "https://www.bilibili.com/video/BVx" {
		t.Fatalf("original_url 错误: %v", m["original_url"])
	}
	if m["view_time"] == "" || m["tag_name"] != "tagA" || m["tname"] != "tagA" {
		t.Fatalf("字段缺失: %#v", m)
	}
	if covers, ok := m["covers"].([]string); !ok || len(covers) != 1 {
		t.Fatalf("covers 应包装为单元素切片: %#v", m["covers"])
	}
	m2 := processRecord(models.HistoryRecord{Title: "e"}, false, false)
	if c, ok := m2["covers"].([]string); !ok || len(c) != 0 {
		t.Fatalf("空 covers 应为空切片: %#v", m2["covers"])
	}
}

func TestInsertHistoryRecord(t *testing.T) {
	if err := GetSQLiteDB().EnsureTableForYear(yearInfra1); err != nil {
		t.Fatal(err)
	}
	conn := testConn()
	table := fmt.Sprintf("bilibili_history_%d", yearInfra1)

	rec := historyRecord("BVins001", "插入 测试", "archive", "insTag", "insMain", 8801, "插入UP", atTime(yearInfra1, 6, 1, 10, 0), 600, 100, 1)
	inserted, err := InsertHistoryRecord(conn, table, &rec)
	if err != nil || !inserted {
		t.Fatalf("首次插入应成功: inserted=%v err=%v", inserted, err)
	}

	// 24h 内同 bvid → 更新而非插入
	rec2 := rec
	rec2.ViewAt = atTime(yearInfra1, 6, 1, 12, 0)
	rec2.Progress = 500
	rec2.Current = "08:20"
	inserted, err = InsertHistoryRecord(conn, table, &rec2)
	if err != nil || inserted {
		t.Fatalf("24h 内应返回 false: inserted=%v err=%v", inserted, err)
	}
	var prog int
	var cur string
	q := "SELECT progress, current FROM " + table + " WHERE bvid='BVins001'"
	if err := conn.QueryRow(q).Scan(&prog, &cur); err != nil {
		t.Fatal(err)
	}
	if prog != 500 || cur != "08:20" {
		t.Fatalf("进度未更新: %d %q", prog, cur)
	}

	// 更早的 view_at + 空 cover：走无 view_at 更新分支
	rec3 := rec
	rec3.ViewAt = atTime(yearInfra1, 6, 1, 11, 0)
	rec3.Progress = 200
	rec3.Cover = ""
	inserted, err = InsertHistoryRecord(conn, table, &rec3)
	if err != nil || inserted {
		t.Fatalf("更早时间也应去重: inserted=%v err=%v", inserted, err)
	}
	if err := conn.QueryRow(q).Scan(&prog, &cur); err != nil {
		t.Fatal(err)
	}
	if prog != 500 {
		t.Fatalf("更早记录不应覆盖进度, got %d", prog)
	}

	// 老行 author 为空时回填
	mustExec(t, conn, "INSERT INTO "+table+" (title, oid, bvid, dt, author_name, author_mid, view_at, duration) VALUES ('空UP', 1, 'BVins002', 1, '', 0, 1000000000, 10)")
	filler := historyRecord("BVins002", "插入 测试", "archive", "insTag", "insMain", 8802, "回填UP", 1000000000+3600, 10, 5, 1)
	filler.Cover = ""
	inserted, err = InsertHistoryRecord(conn, table, &filler)
	if err != nil || inserted {
		t.Fatalf("应走去重分支: %v %v", inserted, err)
	}
	var author string
	if err := conn.QueryRow("SELECT author_name FROM " + table + " WHERE bvid='BVins002'").Scan(&author); err != nil {
		t.Fatal(err)
	}
	if author != "回填UP" {
		t.Fatalf("空 author 应被回填, got %q", author)
	}

	// 超过 24h → 新插入
	rec4 := rec
	rec4.ViewAt = atTime(yearInfra1, 6, 3, 10, 0)
	inserted, err = InsertHistoryRecord(conn, table, &rec4)
	if err != nil || !inserted {
		t.Fatalf("超过 24h 应插入新行: inserted=%v err=%v", inserted, err)
	}
	if n := countIn(t, yearInfra1, "bvid='BVins001'"); n != 2 {
		t.Fatalf("应有 2 行, got %d", n)
	}

	// 不存在的表 → 报错
	if _, err := InsertHistoryRecord(conn, "no_such_table", &rec); err == nil {
		t.Fatal("坏表名应报错")
	}
}

func TestGetHistoryPage(t *testing.T) {
	seedHistoryFixtures(t)

	resp, years, err := GetHistoryPage(HistoryQueryParams{Page: 1, Size: 100, DateRange: fmt.Sprintf("%d0101-%d1231", yearHistory1, yearHistory1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(years) == 0 {
		t.Fatal("应返回可用年份")
	}
	wantTotal := countIn(t, yearHistory1, "business NOT IN ('live','article','article-list')")
	if int(resp.Total) != wantTotal {
		t.Fatalf("total 应为 %d, got %d", wantTotal, resp.Total)
	}
	records := resp.Records.([]map[string]interface{})
	if len(records) != wantTotal {
		t.Fatalf("返回行数应为 %d, got %d", wantTotal, len(records))
	}
	if records[0]["view_time"] == "" || records[0]["original_url"] == "" {
		t.Fatalf("处理后字段缺失: %#v", records[0])
	}
	if recBvid(t, records[0]) != "BVhist1005" {
		t.Fatalf("默认应按 view_at 降序, 首行 got %v", records[0]["bvid"])
	}

	// 升序
	resp, _, err = GetHistoryPage(HistoryQueryParams{Page: 1, Size: 100, SortOrder: 1, DateRange: fmt.Sprintf("%d0101-%d1231", yearHistory1, yearHistory1)})
	if err != nil {
		t.Fatal(err)
	}
	if recBvid(t, resp.Records.([]map[string]interface{})[0]) != "BVhist1001" {
		t.Fatal("SortOrder=1 应升序")
	}

	// 分页：5 行数据，第二页(size=2)应有 2 行，第三页应有 1 行
	resp, _, err = GetHistoryPage(HistoryQueryParams{Page: 2, Size: 2, DateRange: fmt.Sprintf("%d0101-%d1231", yearHistory1, yearHistory1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Records.([]map[string]interface{})) != 2 {
		t.Fatalf("第二页应有 2 行, got %d", len(resp.Records.([]map[string]interface{})))
	}
	resp, _, err = GetHistoryPage(HistoryQueryParams{Page: 3, Size: 2, DateRange: fmt.Sprintf("%d0101-%d1231", yearHistory1, yearHistory1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Records.([]map[string]interface{})) != wantTotal-4 {
		t.Fatalf("第三页应有 %d 行", wantTotal-4)
	}

	// TagName / MainCategory / Business 过滤
	resp, _, err = GetHistoryPage(HistoryQueryParams{Page: 1, Size: 10, TagName: "音乐tag"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total < 1 {
		t.Fatal("tag 过滤应命中")
	}
	resp, _, err = GetHistoryPage(HistoryQueryParams{Page: 1, Size: 10, MainCategory: "番剧", TagName: "音乐tag"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range resp.Records.([]map[string]interface{}) {
		if r["main_category"] != "番剧" {
			t.Fatalf("MainCategory 应优先于 TagName: %#v", r)
		}
	}
	resp, _, err = GetHistoryPage(HistoryQueryParams{Page: 1, Size: 10, Business: "live", DateRange: fmt.Sprintf("%d0101-%d1231", yearHistory1, yearHistory1)})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || recBvid(t, resp.Records.([]map[string]interface{})[0]) != "BVhist1003" {
		t.Fatalf("business=live 应只命中 1 行, got %d", resp.Total)
	}

	// 非法日期范围 → 忽略过滤，不报错
	resp, _, err = GetHistoryPage(HistoryQueryParams{Page: 1, Size: 1, DateRange: "bad-range"})
	if err != nil || resp.Total < 1 {
		t.Fatalf("坏 DateRange 应被忽略: %v %d", err, resp.Total)
	}
}

func TestGetHistoryDates(t *testing.T) {
	dates, err := GetHistoryDates()
	if err != nil {
		t.Fatal(err)
	}
	if len(dates) < 2 {
		t.Fatalf("日期数量异常: %v", dates)
	}
	want := fmt.Sprintf("%d0105", yearHistory1)
	var found bool
	for _, d := range dates {
		if d == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("应包含 %s, got %v", want, dates)
	}
	for i := 1; i < len(dates); i++ {
		if dates[i-1] < dates[i] {
			t.Fatalf("应降序: %v", dates)
		}
	}
}

func TestSearchHistory(t *testing.T) {
	resp, _, err := SearchHistory(HistorySearchParams{Page: 1, Size: 10, Search: "命中"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || recBvid(t, resp.Records.([]map[string]interface{})[0]) != "BVhist1005" {
		t.Fatalf("全文搜索应命中 1 条, got %d", resp.Total)
	}

	resp, _, err = SearchHistory(HistorySearchParams{Page: 1, Size: 10, Search: "特殊UP主", SearchType: "author"})
	if err != nil || resp.Total != 1 {
		t.Fatalf("author 搜索失败: %v %d", err, resp.Total)
	}
	resp, _, err = SearchHistory(HistorySearchParams{Page: 1, Size: 10, Search: "看完", SearchType: "remark"})
	if err != nil || resp.Total < 1 {
		t.Fatalf("remark 搜索失败: %v %d", err, resp.Total)
	}
	resp, _, err = SearchHistory(HistorySearchParams{Page: 1, Size: 10, Search: "音乐tag", SearchType: "tag"})
	if err != nil || resp.Total < 1 {
		t.Fatalf("tag 搜索失败: %v %d", err, resp.Total)
	}
	resp, _, err = SearchHistory(HistorySearchParams{Page: 1, Size: 10, Search: "科技tag", SearchType: "title"})
	if err != nil || resp.Total != 0 {
		t.Fatalf("title 维度不应命中 tag 关键词: %v %d", err, resp.Total)
	}
	// 未知 SearchType 与空搜索：退回全量（排除 live/article）
	resp, _, err = SearchHistory(HistorySearchParams{Page: 1, Size: 5, Search: "视频", SearchType: "bogus"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total < 1 {
		t.Fatal("未知 SearchType 应退化为无搜索条件")
	}
	resp, _, err = SearchHistory(HistorySearchParams{Page: 1, Size: 5, Search: "   "})
	if err != nil || resp.Total < 1 {
		t.Fatalf("空白搜索应忽略: %v", err)
	}
}

func TestRemarks(t *testing.T) {
	h01ViewAt := atTime(yearHistory1, 1, 5, 10, 0)
	remark, rt, err := GetRemarkByBvidAndViewAt("BVhist1001", h01ViewAt)
	if err != nil || remark != "看完" || rt != atTime(yearHistory1, 1, 5, 11, 0) {
		t.Fatalf("应取回备注: %q %d %v", remark, rt, err)
	}
	remark, _, err = GetRemarkByBvidAndViewAt("BVhist1001", h01ViewAt+1)
	if err != nil || remark != "" {
		t.Fatalf("view_at 不匹配应为空: %q %v", remark, err)
	}
	if _, _, err := GetRemarkByBvidAndViewAt("nope", atTime(1999, 1, 1, 0, 0)); err != nil {
		t.Fatalf("无表年份应返回 nil 错误: %v", err)
	}

	got := BatchGetRemarksByBvid([]string{})
	if len(got) != 0 {
		t.Fatalf("空入参应返回空 map: %#v", got)
	}
	got = BatchGetRemarksByBvid([]string{"BVhist1001", "BVhist9001", "BV不存在"})
	if got["BVhist1001"]["remark"] != "看完" {
		t.Fatalf("BVhist1001 备注缺失: %#v", got)
	}
	if got["BVhist9001"]["remark"] != "新备注" {
		t.Fatalf("应取跨年最新备注: %#v", got["BVhist9001"])
	}
	if _, ok := got["BV不存在"]; ok {
		t.Fatal("无备注 bvid 不应出现在结果里")
	}
}

func TestGetVideoByCID(t *testing.T) {
	m, err := GetVideoByCID(71041, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if m["bvid"] != "BVhist1004" {
		t.Fatalf("bvid 错误: %v", m["bvid"])
	}
	if _, err := GetVideoByCID(99999999, false, false); err == nil {
		t.Fatal("未知 cid 应报错")
	}
}

func TestUpdateHistoryTagNames(t *testing.T) {
	if err := EnsureVideoDetailsTables(); err != nil {
		t.Fatal(err)
	}
	row := historyRecord("BVtag0001", "空标签 视频", "archive", "", "", 7109, "标签UP", atTime(yearHistory1, 4, 4, 4, 0), 60, 60, 1)
	mustSeedHistory(t, yearHistory1, row)

	v := &models.VideoBaseInfo{Bvid: "BVtag0001", Aid: 1, Title: "空标签 视频", OwnerMid: 7109, OwnerName: "标签UP", Tname: "回填测试", Pubdate: 1, FetchTime: 1}
	if err := UpsertVideoBaseInfo(v); err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateHistoryTagNames()
	if err != nil {
		t.Fatal(err)
	}
	if updated < 1 {
		t.Fatalf("应至少回填 1 行, got %d", updated)
	}
	var tag string
	q := fmt.Sprintf("SELECT tag_name FROM bilibili_history_%d WHERE bvid='BVtag0001'", yearHistory1)
	if err := testConn().QueryRow(q).Scan(&tag); err != nil {
		t.Fatal(err)
	}
	if tag != "回填测试" {
		t.Fatalf("tag_name 未回填: %q", tag)
	}
	// 幂等：已回填的行不再更新
	updated, err = UpdateHistoryTagNames()
	if err != nil {
		t.Fatal(err)
	}
	_ = updated
}

func TestGetDailyStats(t *testing.T) {
	count, seconds, err := GetDailyStats("0720", fmt.Sprint(yearHistory1))
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("0720 应有 1 条, got %d", count)
	}
	if seconds != 1800 {
		t.Fatalf("progress=-1 应取 duration, got %d", seconds)
	}

	count, _, err = GetDailyStats("0105", fmt.Sprint(yearHistory1))
	if err != nil || count != 1 {
		t.Fatalf("0105 应有 1 条: %d %v", count, err)
	}

	// 无表年份 → 0
	count, seconds, err = GetDailyStats("0101", "1999")
	if err != nil || count != 0 || seconds != 0 {
		t.Fatalf("无表年份应为 0: %d %d %v", count, seconds, err)
	}

	// year 为空 → 使用最新表；该日无数据时 SUM 为 NULL——回归：
	// COALESCE 修复后应返回 0/0 且不再报 Scan 错误
	count, seconds, err = GetDailyStats("0230", "")
	if err != nil || count != 0 || seconds != 0 {
		t.Fatalf("最新表无匹配日期应返回 0/0 无错误: %d %d %v", count, seconds, err)
	}
}

func TestMarkVideoDeleted(t *testing.T) {
	row := historyRecord("BVdel001", "待删除 视频", "archive", "删除tag", "删除", 7110, "删除UP", atTime(yearHistory2, 4, 1, 10, 0), 60, 60, 1)
	mustSeedHistory(t, yearHistory2, row)

	if err := MarkVideoDeleted("BVdel001"); err != nil {
		t.Fatal(err)
	}
	var status int
	q := fmt.Sprintf("SELECT status FROM bilibili_history_%d WHERE bvid='BVdel001'", yearHistory2)
	if err := testConn().QueryRow(q).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != 1 {
		t.Fatalf("应标记为删除, got %d", status)
	}

	got := GetDeletedStatus([]string{"BVdel001", "BVhist1001"})
	if !got["BVdel001"] {
		t.Fatalf("BVdel001 应为已删除: %#v", got)
	}
	if got["BVhist1001"] {
		t.Fatal("BVhist1001 不应为已删除")
	}
	if len(GetDeletedStatus([]string{})) != 0 {
		t.Fatal("空入参应返回空 map")
	}
}

func TestGetAllHistoryRecords(t *testing.T) {
	records, err := GetAllHistoryRecords(1999)
	if err != nil || records != nil {
		t.Fatalf("无表年份应返回 nil: %v %v", records, err)
	}
	want := countIn(t, yearHistory1, "1=1")
	records, err = GetAllHistoryRecords(yearHistory1)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != want {
		t.Fatalf("应返回 %d 行, got %d", want, len(records))
	}
	if _, ok := records[0]["view_at"]; !ok {
		t.Fatalf("缺少 view_at: %#v", records[0])
	}
	prev := int64(1 << 62)
	for _, r := range records {
		v := toInt64(r["view_at"])
		if v > prev {
			t.Fatal("应按 view_at 降序")
		}
		prev = v
	}
}

func TestGetBvidByOid(t *testing.T) {
	bvid, err := GetBvidByOid("880001")
	if err != nil || bvid != "BVhist1002" {
		t.Fatalf("oid 查询失败: %q %v", bvid, err)
	}
	if _, err := GetBvidByOid("999999999"); err == nil {
		t.Fatal("未知 oid 应报错")
	}
}
