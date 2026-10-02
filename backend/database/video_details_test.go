package database

import (
	"testing"
	"time"

	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

// video_details_test.go 命名空间：
//   - bvid 前缀 BVvd（video_base_info / video_tags / invalid_videos / 2015 历史表）
//   - owner mid 61001+（uploader 聚合可区分）
//   - 年份表只允许 2015（yearDetails）
// 全局表（video_base_info 等）由其他测试共享，断言一律按自己的 bvid/mid/关键词作用域。

const (
	vdBvidA     = "BVvdAAA001"
	vdBvidB     = "BVvdBBB002"
	vdBvidC     = "BVvdCCC003"
	vdBvidU     = "BVvdUPD004"
	vdBvidNoTag = "BVvdNOTAG5"
	vdMidA      = 61001
	vdMidB      = 61002
	vdMidC      = 61003
	vdOwnerA    = "VD专属UP主甲6101"
	vdOwnerB    = "VD专属UP主乙6102"
	vdOwnerC    = "VD专属UP主丙6103"
)

func vdVideo(bvid string, aid, pubdate int64, title, owner string, mid, statView, statLike int) *models.VideoBaseInfo {
	return &models.VideoBaseInfo{
		Bvid:         bvid,
		Aid:          int(aid),
		Videos:       1,
		Tid:          20,
		Tname:        "知识",
		Copyright:    1,
		Pic:          "https://img/" + bvid + ".jpg",
		Title:        title,
		Pubdate:      pubdate,
		Ctime:        pubdate,
		Desc:         "vd desc",
		Duration:     120,
		Cid:          int(aid),
		OwnerMid:     mid,
		OwnerName:    owner,
		OwnerFace:    "https://face/" + owner,
		StatView:     statView,
		StatDanmaku:  5,
		StatReply:    6,
		StatFavorite: 7,
		StatCoin:     8,
		StatShare:    9,
		StatLike:     statLike,
		FetchTime:    atTime(yearDetails, 6, 1, 0, 0),
		UpdateTime:   0,
	}
}

func vdExec(t *testing.T, q string, args ...interface{}) {
	t.Helper()
	mustExec(t, testConn(), q, args...)
}

func vdQueryRowInt64(t *testing.T, q string, args ...interface{}) int64 {
	t.Helper()
	var v int64
	if err := testConn().QueryRow(q, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", q, err)
	}
	return v
}

func vdFetch(t *testing.T, bvid string) *models.VideoBaseInfo {
	t.Helper()
	v, err := GetVideoBaseInfoByBvid(bvid)
	if err != nil {
		t.Fatalf("GetVideoBaseInfoByBvid(%s): %v", bvid, err)
	}
	return v
}

// vdRecList PagedResponse.Records 是 interface{}：非空路径为 []map[string]interface{}，
// 空表快捷路径为 []interface{}{}，这里统一取出。
func vdRecList(t *testing.T, resp *models.PagedResponse) []map[string]interface{} {
	t.Helper()
	switch r := resp.Records.(type) {
	case []map[string]interface{}:
		return r
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(r))
		for _, e := range r {
			m, ok := e.(map[string]interface{})
			if !ok {
				t.Fatalf("record element type %T", e)
			}
			out = append(out, m)
		}
		return out
	default:
		t.Fatalf("unexpected Records type %T", resp.Records)
		return nil
	}
}

// TestVdEnsureTables 建表入口 + 幂等。
func TestVdEnsureTables(t *testing.T) {
	if err := EnsureVideoDetailsTables(); err != nil {
		t.Fatalf("EnsureVideoDetailsTables: %v", err)
	}
	if err := EnsureVideoDetailsTables(); err != nil {
		t.Fatalf("EnsureVideoDetailsTables second call: %v", err)
	}
	for _, tbl := range []string{videoBaseInfoTable, videoTagsTable, uploaderInfoTable, "invalid_videos"} {
		exists, err := GetSQLiteDB().TableExists(tbl)
		if err != nil {
			t.Fatalf("TableExists(%s): %v", tbl, err)
		}
		if !exists {
			t.Fatalf("table %s missing after ensure", tbl)
		}
	}
}

// TestVdReaderErrorsWhenDBNil 说明：conn==nil 的分支不在此覆盖，
// 关闭共享单例会污染整个测试进程里的其他测试文件。

// TestVdUpsertVideoBaseInfo 新增、更新、读取、缺失返回 nil。
func TestVdUpsertVideoBaseInfo(t *testing.T) {
	if err := UpsertVideoBaseInfo(vdVideo(vdBvidA, 710001, atTime(yearDetails, 1, 10, 8, 0), "VD主题 更新前", vdOwnerA, vdMidA, 1000, 100)); err != nil {
		t.Fatalf("insert: %v", err)
	}
	got := vdFetch(t, vdBvidA)
	if got == nil {
		t.Fatal("first read returned nil")
	}
	if got.Title != "VD主题 更新前" || got.StatView != 1000 || got.OwnerMid != vdMidA {
		t.Fatalf("unexpected row: %+v", got)
	}
	if got.Bvid != vdBvidA || got.Aid != 710001 || got.Duration != 120 || got.Tname != "知识" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	firstID := got.ID
	fetchTime := got.FetchTime

	// 更新路径：不应新增行，fetch_time 保持不变
	if err := UpsertVideoBaseInfo(vdVideo(vdBvidA, 710001, atTime(yearDetails, 1, 10, 8, 0), "VD主题 更新后", vdOwnerA, vdMidA, 2000, 300)); err != nil {
		t.Fatalf("update: %v", err)
	}
	got = vdFetch(t, vdBvidA)
	if got.Title != "VD主题 更新后" || got.StatView != 2000 || got.StatLike != 300 {
		t.Fatalf("update not applied: %+v", got)
	}
	if got.ID != firstID {
		t.Fatalf("update inserted a new row: id %d vs %d", got.ID, firstID)
	}
	if got.FetchTime != fetchTime {
		t.Fatalf("fetch_time should be preserved on update: %d vs %d", got.FetchTime, fetchTime)
	}

	// 未找到时返回 (nil, nil)
	if v, err := GetVideoBaseInfoByBvid("BVvdMISSING"); err != nil || v != nil {
		t.Fatalf("missing bvid: got %v, err %v", v, err)
	}
}

// TestVdSearchVideos 标题/UP主模糊搜索 + 分页 + 映射键。
func TestVdSearchVideos(t *testing.T) {
	base := atTime(yearDetails, 2, 1, 0, 0)
	var videos []*models.VideoBaseInfo
	videos = append(videos,
		vdVideo(vdBvidB, 710002, base, "VD搜索关键词 子串B", vdOwnerB, vdMidB, 10, 1),
		vdVideo(vdBvidC, 710003, base+100, "VD搜索关键词 子串C", vdOwnerC, vdMidC, 50, 5),
		vdVideo(vdBvidU, 710004, base+200, "VD无关标题", vdOwnerB, vdMidB, 20, 2),
	)
	for _, v := range videos {
		if err := UpsertVideoBaseInfo(v); err != nil {
			t.Fatalf("seed %s: %v", v.Bvid, err)
		}
	}

	resp, err := SearchVideos("VD搜索关键词", 1, 10)
	if err != nil {
		t.Fatalf("SearchVideos: %v", err)
	}
	if resp.Current != 1 || resp.Size != 10 {
		t.Fatalf("paging echo wrong: %+v", resp)
	}
	if resp.Total < 2 {
		t.Fatalf("total for own keyword should be >= 2, got %d", resp.Total)
	}
	records := vdRecList(t, resp)
	if len(records) != 2 {
		t.Fatalf("want 2 records, got %d", len(records))
	}
	// ORDER BY pubdate DESC：子串C (base+100) 在 子串B (base) 之前
	first := records[0]
	second := records[1]
	if first["bvid"] != vdBvidC || second["bvid"] != vdBvidB {
		t.Fatalf("unexpected order: %v / %v", first["bvid"], second["bvid"])
	}
	// videoBaseInfoToMap 派生键
	if first["original_url"] != "https://www.bilibili.com/video/"+vdBvidC {
		t.Fatalf("original_url wrong: %v", first["original_url"])
	}
	if first["duration_str"] != formatDuration(120) {
		t.Fatalf("duration_str wrong: %v", first["duration_str"])
	}
	if first["pubdate_str"] != utils.FormatDateTime(time.Unix(base+100, 0)) {
		t.Fatalf("pubdate_str wrong: %v", first["pubdate_str"])
	}

	// 按 UP 主名搜索（owner_name LIKE）
	resp, err = SearchVideos(vdOwnerC, 1, 10)
	if err != nil {
		t.Fatalf("SearchVideos by owner: %v", err)
	}
	if resp.Total < 1 || len(vdRecList(t, resp)) < 1 {
		t.Fatalf("owner search missed own row: %+v", resp)
	}

	// 分页：第 2 页应为空
	resp, err = SearchVideos("VD搜索关键词", 2, 10)
	if err != nil {
		t.Fatalf("SearchVideos page 2: %v", err)
	}
	if len(vdRecList(t, resp)) != 0 || resp.Total != 2 {
		t.Fatalf("page 2 wrong: total=%d len=%d", resp.Total, len(vdRecList(t, resp)))
	}

	// 无匹配
	resp, err = SearchVideos("VD不存在的关键词9999", 1, 10)
	if err != nil {
		t.Fatalf("SearchVideos miss: %v", err)
	}
	if resp.Total != 0 || len(vdRecList(t, resp)) != 0 {
		t.Fatalf("expected empty search result, got %+v", resp)
	}
}

// TestVdStats 两个统计入口：只断言本作用域数量下界与键存在。
func TestVdStats(t *testing.T) {
	stats, err := GetVideoDetailStats()
	if err != nil {
		t.Fatalf("GetVideoDetailStats: %v", err)
	}
	own := vdQueryRowInt64(t, "SELECT COUNT(*) FROM video_base_info WHERE bvid LIKE 'BVvd%'")
	if stats.FetchedVideos < own {
		t.Fatalf("FetchedVideos %d < own rows %d", stats.FetchedVideos, own)
	}

	dbStats, err := GetDatabaseStats()
	if err != nil {
		t.Fatalf("GetDatabaseStats: %v", err)
	}
	for _, key := range []string{"total_videos", "total_views", "total_danmaku", "total_likes",
		"total_coins", "total_favorites", "total_shares", "total_duration_seconds",
		"total_uploaders", "total_unique_tags"} {
		if _, ok := dbStats[key]; !ok {
			t.Fatalf("GetDatabaseStats missing key %q", key)
		}
	}
	if tv, ok := dbStats["total_videos"].(int64); !ok || tv < own {
		t.Fatalf("total_videos %v < own rows %d", dbStats["total_videos"], own)
	}
}

// TestVdUploaderAggregations GetUploaderList / GetUploaderDetail（按自己的 mid 作用域）。
func TestVdUploaderAggregations(t *testing.T) {
	// uploader_info 行（全局表，按唯一 mid 作用域）
	vdExec(t, `INSERT OR REPLACE INTO uploader_info
		(mid, name, sex, face, sign, level, fans, attention, archive_count, fetch_time, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		vdMidA, vdOwnerA, "男", "face-a", "vd签名", 5, 1234, 20, 99, atTime(yearDetails, 3, 1, 0, 0), 0)

	for sortBy, wantFirst := range map[string]int{
		"views":       vdMidA, // 自己的行里 views 最大
		"video_count": vdMidB, // vdMidB 有 2 个视频
		"likes":       vdMidA,
		"unknown":     vdMidA, // default -> video_count 排序，但 A/B 谁前取决于平局？见下
	} {
		resp, err := GetUploaderList(1, 1000000, sortBy)
		if err != nil {
			t.Fatalf("GetUploaderList(%s): %v", sortBy, err)
		}
		found := false
		for _, m := range vdRecList(t, resp) {
			if m["mid"] == wantFirst {
				found = true
				// uploaderStatsToMap 的键
				for _, k := range []string{"mid", "name", "face", "video_count",
					"total_views", "total_likes", "total_coins", "total_favorites"} {
					if _, ok := m[k]; !ok {
						t.Fatalf("record missing key %q", k)
					}
				}
			}
		}
		if !found {
			t.Fatalf("sortBy=%s: mid %d not in page (maybe other suites' rows crowd it out)", sortBy, wantFirst)
		}
	}

	// 精确聚合：mid 61001（A+C 两行，views 1000+2000+50? 注意 A 已被更新为 2000）
	detail, err := GetUploaderDetail(vdMidA)
	if err != nil {
		t.Fatalf("GetUploaderDetail A: %v", err)
	}
	// 用 SQL 按同一 mid 计算期望值，不依赖全局状态
	wantCount := vdQueryRowInt64(t, "SELECT COUNT(*) FROM video_base_info WHERE owner_mid = ?", vdMidA)
	if detail["video_count"] != wantCount {
		t.Fatalf("video_count %v want %d", detail["video_count"], wantCount)
	}
	wantViews := vdQueryRowInt64(t, "SELECT COALESCE(SUM(stat_view),0) FROM video_base_info WHERE owner_mid = ?", vdMidA)
	if detail["total_views"] != wantViews {
		t.Fatalf("total_views %v want %d", detail["total_views"], wantViews)
	}
	if detail["name"] != vdOwnerA {
		t.Fatalf("name %v", detail["name"])
	}
	// uploader_info 合并分支
	if detail["sign"] != "vd签名" || detail["level"] != 5 || detail["fans"] != 1234 ||
		detail["attention"] != 20 || detail["archive_count"] != 99 {
		t.Fatalf("uploader_info merge wrong: %+v", detail)
	}

	// 无任何视频的 mid：ErrNoRows 分支 + 零聚合
	detail, err = GetUploaderDetail(61999)
	if err != nil {
		t.Fatalf("GetUploaderDetail unknown mid: %v", err)
	}
	if detail["video_count"] != int64(0) || detail["name"] != "" || detail["face"] != "" {
		t.Fatalf("unknown mid detail wrong: %+v", detail)
	}
	if _, ok := detail["sign"]; ok {
		t.Fatalf("sign should be absent when uploader_info has no row: %+v", detail)
	}
}

// TestVdTagList 标签聚合按唯一标签名作用域。
func TestVdTagList(t *testing.T) {
	tags := []struct {
		bvid, name string
		tagID      int64
	}{
		{vdBvidA, "vd标签甲71001", 830001},
		{vdBvidA, "vd标签乙71002", 830002},
		{vdBvidB, "vd标签甲71001", 830001}, // 同名标签两行
	}
	for _, tg := range tags {
		vdExec(t, `INSERT OR IGNORE INTO video_tags (bvid, tag_id, tag_name) VALUES (?, ?, ?)`,
			tg.bvid, tg.tagID, tg.name)
	}

	resp, err := GetTagList(1, 1000000)
	if err != nil {
		t.Fatalf("GetTagList: %v", err)
	}
	counts := map[string]int64{}
	for _, m := range vdRecList(t, resp) {
		counts[m["tag_name"].(string)] = m["video_count"].(int64)
	}
	if counts["vd标签甲71001"] != 2 {
		t.Fatalf("vd标签甲71001 count=%d want 2", counts["vd标签甲71001"])
	}
	if counts["vd标签乙71002"] != 1 {
		t.Fatalf("vd标签乙71002 count=%d want 1", counts["vd标签乙71002"])
	}
	ownDistinct := vdQueryRowInt64(t, "SELECT COUNT(DISTINCT tag_name) FROM video_tags WHERE tag_name LIKE 'vd标签%'")
	if resp.Total < ownDistinct {
		t.Fatalf("total %d < own distinct %d", resp.Total, ownDistinct)
	}

	// 空页
	resp, err = GetTagList(9999, 500)
	if err != nil {
		t.Fatalf("GetTagList far page: %v", err)
	}
	if len(vdRecList(t, resp)) != 0 {
		t.Fatalf("far page should be empty")
	}
}

// TestVdUniqueBvidsFromHistory 2015 历史表成员检查（archive 收录，其它 business 排除）。
func TestVdUniqueBvidsFromHistory(t *testing.T) {
	mustSeedHistory(t, yearDetails,
		historyRecord("BVvdHistArc1", "VD历史 归档", "archive", "知识", "知识", vdMidA, vdOwnerA, atTime(yearDetails, 4, 1, 10, 0), 100, 50, 1),
		historyRecord("BVvdHistArc1", "VD历史 归档", "archive", "知识", "知识", vdMidA, vdOwnerA, atTime(yearDetails, 4, 2, 10, 0), 100, 100, 2), // 去重
		historyRecord("BVvdHistLive", "VD历史 直播", "live", "直播", "直播", 61900, "主播", atTime(yearDetails, 4, 3, 10, 0), 100, 0, 1),
		historyRecord("", "VD历史 空bvid", "archive", "知识", "知识", vdMidB, vdOwnerB, atTime(yearDetails, 4, 4, 10, 0), 100, 0, 1),
	)

	bvids, err := GetUniqueBvidsFromHistory()
	if err != nil {
		t.Fatalf("GetUniqueBvidsFromHistory: %v", err)
	}
	set := map[string]bool{}
	for _, b := range bvids {
		set[b] = true
	}
	if !set["BVvdHistArc1"] {
		t.Fatalf("own archive bvid missing")
	}
	if !set["BVseed0001"] {
		t.Fatalf("TestMain seed archive bvid missing (2005 table should be scanned)")
	}
	if set["BVvdHistLive"] {
		t.Fatalf("live business bvid must be excluded")
	}
	if set[""] {
		t.Fatalf("empty bvid must be excluded")
	}
	// 去重：同一 bvid 只出现一次
	seen := map[string]int{}
	for _, b := range bvids {
		seen[b]++
	}
	if seen["BVvdHistArc1"] != 1 {
		t.Fatalf("bvid not deduplicated: %d occurrences", seen["BVvdHistArc1"])
	}
}

// TestVdFetchedBvids 已抓取集合。
func TestVdFetchedBvids(t *testing.T) {
	set, err := GetFetchedBvids()
	if err != nil {
		t.Fatalf("GetFetchedBvids: %v", err)
	}
	for _, b := range []string{vdBvidA, vdBvidB, vdBvidC, vdBvidU} {
		if !set[b] {
			t.Fatalf("%s should be fetched", b)
		}
	}
	if set["BVvdNoSuchVideo"] {
		t.Fatalf("unknown bvid present")
	}
}

// TestVdInvalidVideos 记录、去重、集合、分页列表、计数。
func TestVdInvalidVideos(t *testing.T) {
	if err := RecordInvalidVideo("BVvdInv001", "视频已失效", -404); err != nil {
		t.Fatalf("record 1: %v", err)
	}
	if err := RecordInvalidVideo("BVvdInv002", "仅自己可见", 62003); err != nil {
		t.Fatalf("record 2: %v", err)
	}
	// INSERT OR IGNORE：重复记录不更新也不报错
	if err := RecordInvalidVideo("BVvdInv001", "换一条消息", -500); err != nil {
		t.Fatalf("record dup: %v", err)
	}
	msg := ""
	if err := testConn().QueryRow("SELECT error_message FROM invalid_videos WHERE bvid = ?", "BVvdInv001").Scan(&msg); err != nil {
		t.Fatalf("read dup: %v", err)
	}
	if msg != "视频已失效" {
		t.Fatalf("INSERT OR IGNORE overwrote first message: %q", msg)
	}

	set, err := GetInvalidBvids()
	if err != nil {
		t.Fatalf("GetInvalidBvids: %v", err)
	}
	if !set["BVvdInv001"] || !set["BVvdInv002"] || set["BVvdInvMissing"] {
		t.Fatalf("set wrong: %v", set)
	}

	resp, err := GetInvalidVideoList(1, 1000000)
	if err != nil {
		t.Fatalf("GetInvalidVideoList: %v", err)
	}
	var inv1 map[string]interface{}
	for _, m := range vdRecList(t, resp) {
		if m["bvid"] == "BVvdInv001" {
			inv1 = m
		}
	}
	if inv1 == nil {
		t.Fatal("BVvdInv001 not in list")
	}
	if inv1["error_code"] != -404 || inv1["error_message"] != "视频已失效" {
		t.Fatalf("list row wrong: %v", inv1)
	}
	if _, ok := inv1["fetch_time"].(int64); !ok {
		t.Fatalf("fetch_time type wrong: %T", inv1["fetch_time"])
	}
	if resp.Total < 2 {
		t.Fatalf("total %d < 2", resp.Total)
	}

	// 空页（LIMIT/OFFSET 分支）
	resp, err = GetInvalidVideoList(99999, 500)
	if err != nil {
		t.Fatalf("GetInvalidVideoList far page: %v", err)
	}
	if len(vdRecList(t, resp)) != 0 {
		t.Fatalf("far page should have no records")
	}

	if cnt := GetInvalidVideoCount(); cnt < 2 {
		t.Fatalf("GetInvalidVideoCount %d < 2", cnt)
	}
	own := vdQueryRowInt64(t, "SELECT COUNT(*) FROM invalid_videos WHERE bvid LIKE 'BVvdInv%'")
	if own != 2 {
		t.Fatalf("own invalid rows should be exactly 2, got %d", own)
	}
}

// TestVdFormatDuration 纯函数边界。
func TestVdFormatDuration(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "00:00"},
		{-10, "00:00"},
		{5, "00:05"},
		{65, "01:05"},
		{59*60 + 59, "59:59"},
		{3600, "01:00:00"},
		{3661, "01:01:01"},
		{7322, "02:02:02"},
		{360000, "100:00:00"},
	}
	for _, c := range cases {
		if got := formatDuration(c.in); got != c.want {
			t.Fatalf("formatDuration(%d)=%q want %q", c.in, got, c.want)
		}
	}
}

// TestVdMapHelpers 纯映射函数键值。
func TestVdMapHelpers(t *testing.T) {
	pub := atTime(yearDetails, 5, 5, 5, 5)
	v := vdVideo("BVvdMAP001", 710999, pub, "map title", "map owner", vdMidC, 11, 22)
	m := videoBaseInfoToMap(v)
	if m["pubdate_str"] != utils.FormatDateTime(time.Unix(pub, 0)) {
		t.Fatalf("pubdate_str %v", m["pubdate_str"])
	}
	if m["duration_str"] != "02:00" {
		t.Fatalf("duration_str %v want 02:00 (120s)", m["duration_str"])
	}
	if m["original_url"] != "https://www.bilibili.com/video/BVvdMAP001" {
		t.Fatalf("original_url %v", m["original_url"])
	}
	if m["bvid"] != "BVvdMAP001" || m["stat_view"] != 11 || m["owner_name"] != "map owner" {
		t.Fatalf("base keys wrong: %v", m)
	}
	if len(m) != 29 {
		t.Fatalf("videoBaseInfoToMap key count = %d, want 29", len(m))
	}

	stat := &models.UploaderStats{Mid: vdMidC, Name: "n", Face: "f", VideoCount: 1,
		TotalViews: 2, TotalLikes: 3, TotalCoins: 4, TotalFavorites: 5}
	sm := uploaderStatsToMap(stat)
	if sm["mid"] != vdMidC || sm["video_count"] != int64(1) || sm["total_views"] != int64(2) ||
		sm["total_likes"] != int64(3) || sm["total_coins"] != int64(4) || sm["total_favorites"] != int64(5) ||
		sm["name"] != "n" || sm["face"] != "f" {
		t.Fatalf("uploaderStatsToMap wrong: %v", sm)
	}
	if len(sm) != 8 {
		t.Fatalf("uploaderStatsToMap key count = %d, want 8", len(sm))
	}
}

// TestVdGetVideoBaseInfoByBvidUnknown 单独再走一次 nil,nil 分支（防止上面顺序耦合）。
func TestVdGetVideoBaseInfoByBvidUnknown(t *testing.T) {
	v, err := GetVideoBaseInfoByBvid("BVvdZZZnone")
	if v != nil || err != nil {
		t.Fatalf("want nil,nil got %v,%v", v, err)
	}
}
