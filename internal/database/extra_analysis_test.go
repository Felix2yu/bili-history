package database

import (
	"database/sql"
	"testing"
	"time"
)

// extra_analysis_test.go 覆盖 extra_analysis.go。
// 三个独立库（likes / watchlater / favorites）由各 Get*DB() 单例自动建表。
// 约定：只读写 xana 前缀自有数据（bvid=BVrepx_*，media_id=91xxxx），
// 计数断言一律使用"基线 + 自有增量"或 >=，容忍其它测试文件先行写入的数据。

const (
	xanaFetchTS = 1700000000 // 固定 fetch_time，避免与外部数据混淆
)

func xanaCount(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func xanaFindCreator(stats []CreatorCount, name string) int {
	for _, s := range stats {
		if s.Name == name {
			return s.Count
		}
	}
	return -1
}

func xanaFindCategory(stats []CategoryCount, name string) int {
	for _, s := range stats {
		if s.Name == name {
			return s.Count
		}
	}
	return -1
}

func xanaFindBucket(dist []DurationBucket, label string) int {
	for _, d := range dist {
		if d.Range == label {
			return d.Count
		}
	}
	return -1
}

func xanaFindFolder(stats []FolderCount, title string) int {
	for _, s := range stats {
		if s.Title == title {
			return s.Count
		}
	}
	return -1
}

// 五个时长桶在生产 SQL 中固定出现（只要各桶至少一条 duration>0 数据）。
var xanaAllBuckets = []string{
	"短视频(<1分钟)", "短片(1-5分钟)", "中等(5-10分钟)", "较长(10-30分钟)", "长视频(30分钟+)",
}

func xanaAssertBucketsPresent(t *testing.T, dist []DurationBucket, where string) {
	t.Helper()
	for _, b := range xanaAllBuckets {
		if xanaFindBucket(dist, b) < 1 {
			t.Errorf("%s: duration bucket %s missing (dist=%v)", where, b, dist)
		}
	}
	if len(dist) != len(xanaAllBuckets) {
		t.Errorf("%s: bucket list should have exactly %d ordered entries, got %v", where, len(xanaAllBuckets), dist)
	}
}

// ---------- likes ----------

func TestXanaAnalyzeLikes(t *testing.T) {
	db := GetLikesDB()
	if db == nil {
		t.Skip("likes db unavailable")
	}

	// 空/基线路径：插入自有数据前先调用一次，仅要求不报错
	base := xanaCount(t, db, "liked_videos")
	emptyRes, err := AnalyzeLikes()
	if err != nil {
		t.Fatalf("AnalyzeLikes (baseline): %v", err)
	}
	if emptyRes == nil {
		t.Fatal("AnalyzeLikes returned nil result")
	}
	if emptyRes.TotalCount != base {
		t.Errorf("baseline TotalCount = %d, want %d", emptyRes.TotalCount, base)
	}

	rows := []struct {
		bvid, owner, tname string
		dur, view, danmaku int
	}{
		{"BVrepx_l1", "xana点赞UP", "xana知识", 30, 100, 10}, // 短视频(<1分钟)
		{"BVrepx_l2", "xana点赞UP", "xana知识", 120, 200, 0}, // 短片(1-5分钟)
		{"BVrepx_l3", "xana音乐UP", "xana音乐", 450, 0, 5},   // 中等(5-10分钟)
		{"BVrepx_l4", "xana音乐UP", "xana音乐", 900, 300, 0}, // 较长(10-30分钟)
		{"BVrepx_l5", "", "", 2000, 0, 0},                // 长视频(30分钟+)，owner/tname 为空被榜单排除
		{"BVrepx_l6", "xana点赞UP", "xana知识", 0, 400, 0},   // duration=0 不进入时长桶
	}
	for _, r := range rows {
		mustExec(t, db, `INSERT OR IGNORE INTO liked_videos
			(bvid, title, duration, tname, owner_name, view, danmaku, fetch_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			r.bvid, "xana 点赞视频", r.dur, r.tname, r.owner, r.view, r.danmaku, xanaFetchTS)
	}
	// 重复插入验证 INSERT OR IGNORE 幂等（同时覆盖 upsert 不改变总数）
	if len(rows) > 0 {
		r := rows[0]
		mustExec(t, db, `INSERT OR IGNORE INTO liked_videos
			(bvid, title, duration, tname, owner_name, view, danmaku, fetch_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			r.bvid, "xana 点赞视频", r.dur, r.tname, r.owner, r.view, r.danmaku, xanaFetchTS)
	}

	res, err := AnalyzeLikes()
	if err != nil {
		t.Fatalf("AnalyzeLikes: %v", err)
	}
	if res.TotalCount != base+len(rows) {
		t.Errorf("TotalCount = %d, want %d", res.TotalCount, base+len(rows))
	}
	if got := xanaFindCreator(res.TopCreators, "xana点赞UP"); got < 3 {
		t.Errorf("TopCreators xana点赞UP = %d, want >=3 (list=%v)", got, res.TopCreators)
	}
	if got := xanaFindCreator(res.TopCreators, ""); got != -1 {
		t.Errorf("empty owner_name must be excluded from TopCreators")
	}
	if got := xanaFindCategory(res.CategoryDist, "xana知识"); got < 3 {
		t.Errorf("CategoryDist xana知识 = %d, want >=3", got)
	}
	xanaAssertBucketsPresent(t, res.DurationDist, "likes")
	if res.AvgViewCount <= 0 {
		t.Errorf("AvgViewCount = %v, want > 0", res.AvgViewCount)
	}
	if res.AvgDanmakuCount <= 0 {
		t.Errorf("AvgDanmakuCount = %v, want > 0", res.AvgDanmakuCount)
	}

	// 清理，降低对后续 extras 测试的干扰
	for _, r := range rows {
		mustExec(t, db, `DELETE FROM liked_videos WHERE bvid = ?`, r.bvid)
	}
}

// ---------- watchlater ----------

func TestXanaAnalyzeWatchLater(t *testing.T) {
	db := GetWatchLaterDB()
	if db == nil {
		t.Skip("watchlater db unavailable")
	}

	base := xanaCount(t, db, "watchlater_videos")
	emptyRes, err := AnalyzeWatchLater()
	if err != nil {
		t.Fatalf("AnalyzeWatchLater (baseline): %v", err)
	}
	if emptyRes.TotalCount != base {
		t.Errorf("baseline TotalCount = %d, want %d", emptyRes.TotalCount, base)
	}

	now := time.Now().Unix()
	day := int64(86400)
	type wl struct {
		bvid, owner, tname, title string
		dur, view                 int
		addAt                     int64
	}
	rows := []wl{
		{"BVrepx_w1", "xana稍后UP", "xana科技", "xana最早稍后再看", 30, 500, now - 400*day},
		{"BVrepx_w2", "xana稍后UP", "xana科技", "xana稍后二", 120, 0, now - 30*day},
		{"BVrepx_w3", "xana手工UP", "xana手工", "xana稍后三", 450, 0, now - 20*day},
		{"BVrepx_w4", "xana手工UP", "xana手工", "xana稍后四", 900, 0, now - 10*day},
		{"BVrepx_w5", "", "", "xana稍后五", 2000, 0, now - 5*day},
		{"BVrepx_w6", "xana稍后UP", "xana科技", "xana零时间", 0, 0, 0}, // add_at=0 不进最早条目；duration=0 不进桶
		{"BVrepx_w7", "xana稍后UP", "xana科技", "xana负时间", 60, 0, -5},
	}
	for _, r := range rows {
		mustExec(t, db, `INSERT OR IGNORE INTO watchlater_videos
			(bvid, title, duration, tname, owner_name, add_at, view, fetch_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			r.bvid, r.title, r.dur, r.tname, r.owner, r.addAt, r.view, xanaFetchTS)
	}

	res, err := AnalyzeWatchLater()
	if err != nil {
		t.Fatalf("AnalyzeWatchLater: %v", err)
	}
	if res.TotalCount != base+len(rows) {
		t.Errorf("TotalCount = %d, want %d", res.TotalCount, base+len(rows))
	}
	if got := xanaFindCategory(res.CategoryDist, "xana科技"); got < 3 {
		t.Errorf("CategoryDist xana科技 = %d, want >=3", got)
	}
	xanaAssertBucketsPresent(t, res.DurationDist, "watchlater")
	if res.AvgViewCount <= 0 {
		t.Errorf("AvgViewCount = %v, want > 0", res.AvgViewCount)
	}

	// 最早条目：add_at>0 过滤 + 升序 + days_ago 计算
	if len(res.OldestItems) < 1 {
		t.Fatalf("OldestItems empty")
	}
	var oldest *WatchLaterItem
	zeroFound := false
	for i := range res.OldestItems {
		it := &res.OldestItems[i]
		if it.Title == "xana零时间" || it.Title == "xana负时间" {
			zeroFound = true
		}
		if it.Title == "xana最早稍后再看" {
			oldest = it
		}
	}
	if zeroFound {
		t.Errorf("items with add_at<=0 must be excluded from OldestItems: %v", res.OldestItems)
	}
	if oldest == nil {
		t.Fatalf("own oldest item missing from OldestItems (limit 5 may have been crowded out): %v", res.OldestItems)
	}
	if oldest.DaysAgo < 399 || oldest.DaysAgo > 401 {
		t.Errorf("oldest DaysAgo = %d, want ~400", oldest.DaysAgo)
	}
	if oldest.OwnerName != "xana稍后UP" {
		t.Errorf("oldest owner = %q", oldest.OwnerName)
	}

	for _, r := range rows {
		mustExec(t, db, `DELETE FROM watchlater_videos WHERE bvid = ?`, r.bvid)
	}
}

// ---------- favorites ----------

func TestXanaAnalyzeFavorites(t *testing.T) {
	db := GetFavoritesDB()
	if db == nil {
		t.Skip("favorites db unavailable")
	}

	baseContent := xanaCount(t, db, "favorites_content")
	baseFolder := xanaCount(t, db, "favorites_folder")
	emptyRes, err := AnalyzeFavorites()
	if err != nil {
		t.Fatalf("AnalyzeFavorites (baseline): %v", err)
	}
	if emptyRes.TotalCount != baseContent {
		t.Errorf("baseline TotalCount = %d, want %d", emptyRes.TotalCount, baseContent)
	}

	// 两个自有收藏夹（media_id 唯一，避免与保留 id 冲突）
	folders := []struct {
		mediaID int64
		title   string
	}{
		{910001, "xana收藏夹A"},
		{910002, "xana收藏夹B"},
	}
	for _, f := range folders {
		mustExec(t, db, `INSERT OR IGNORE INTO favorites_folder (media_id, title, fetch_time) VALUES (?, ?, ?)`,
			f.mediaID, f.title, xanaFetchTS)
	}
	contents := []struct {
		mediaID, contentID int64
		creator            string
		dur, play, collect int
	}{
		{910001, 911001, "xana收藏UP", 60, 100, 10},
		{910001, 911002, "xana收藏UP", 0, 200, 0},
		{910002, 912001, "", 120, 0, 5},
	}
	for _, c := range contents {
		mustExec(t, db, `INSERT OR IGNORE INTO favorites_content
			(media_id, content_id, title, creator_name, duration, play, collect, fetch_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			c.mediaID, c.contentID, "xana 收藏视频", c.creator, c.dur, c.play, c.collect, xanaFetchTS)
	}

	res, err := AnalyzeFavorites()
	if err != nil {
		t.Fatalf("AnalyzeFavorites: %v", err)
	}
	if res.TotalCount != baseContent+len(contents) {
		t.Errorf("TotalCount = %d, want %d", res.TotalCount, baseContent+len(contents))
	}
	if res.FolderCount < baseFolder+2 {
		t.Errorf("FolderCount = %d, want >= %d", res.FolderCount, baseFolder+2)
	}
	// media_id 为自有且唯一，JOIN 计数精确
	if got := xanaFindFolder(res.FolderDist, "xana收藏夹A"); got != 2 {
		t.Errorf("FolderDist xana收藏夹A = %d, want 2 (dist=%v)", got, res.FolderDist)
	}
	if got := xanaFindFolder(res.FolderDist, "xana收藏夹B"); got != 1 {
		t.Errorf("FolderDist xana收藏夹B = %d, want 1", got)
	}
	if got := xanaFindCreator(res.TopCreators, "xana收藏UP"); got != 2 {
		t.Errorf("TopCreators xana收藏UP = %d, want 2", got)
	}
	if got := xanaFindCreator(res.TopCreators, ""); got != -1 {
		t.Errorf("empty creator_name must be excluded")
	}
	if res.AvgPlayCount <= 0 {
		t.Errorf("AvgPlayCount = %v, want > 0", res.AvgPlayCount)
	}
	if res.AvgCollectCount <= 0 {
		t.Errorf("AvgCollectCount = %v, want > 0", res.AvgCollectCount)
	}

	// 清理
	mustExec(t, db, `DELETE FROM favorites_content WHERE media_id = ? OR media_id = ?`, folders[0].mediaID, folders[1].mediaID)
	mustExec(t, db, `DELETE FROM favorites_folder WHERE media_id = ? OR media_id = ?`, folders[0].mediaID, folders[1].mediaID)
}

// ---------- overview ----------

func TestXanaExtraStatsOverview(t *testing.T) {
	// 与同刻直接 SQL 计数比对：无论其它测试是否留有数据都成立
	ov := GetExtraStatsOverview()
	checks := []struct {
		key   string
		db    *sql.DB
		table string
	}{
		{"likes_count", GetLikesDB(), "liked_videos"},
		{"watchlater_count", GetWatchLaterDB(), "watchlater_videos"},
		{"favorites_count", GetFavoritesDB(), "favorites_content"},
		{"favorites_folder_count", GetFavoritesDB(), "favorites_folder"},
	}
	for _, c := range checks {
		if c.db == nil {
			t.Logf("db for %s unavailable, skipping key", c.key)
			continue
		}
		v, ok := ov[c.key]
		if !ok {
			t.Errorf("overview missing key %q (got %v)", c.key, ov)
			continue
		}
		n, ok := v.(int)
		if !ok {
			t.Errorf("overview[%s] type = %T, want int", c.key, v)
			continue
		}
		if want := xanaCount(t, c.db, c.table); n != want {
			t.Errorf("overview[%s] = %d, want %d", c.key, n, want)
		}
	}
}
