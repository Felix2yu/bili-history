package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

// extras_test.go 覆盖 extras.go：likes / watchlater / favorites 三套独立 db 文件。
// 命名空间：bvid 前缀 ext_，media_id 711xx/712xx 段，owner mid 71000+。
// 注意：本包另一测试文件 extra_analysis_test.go 先于本文件运行，可能留下 rep_ 前缀行；
// 差量同步（SaveWatchLaterVideos / SaveFavoriteFolders）删除它们是预期行为，
// 但 liked_videos / favorites_content 无裁剪逻辑，纯计数断言只针对 ext_ 自有行或自有 media_id。

// ---------- helpers (ext 前缀) ----------

func extCount(t *testing.T, db *sql.DB, query string, args ...interface{}) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

func extI64(m map[string]interface{}, col string) int64 {
	switch v := m[col].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		n, _ := strconv.ParseInt(string(v), 10, 64)
		return n
	case string:
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	}
	return 0
}

func extStr(m map[string]interface{}, col string) string {
	switch v := m[col].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return ""
}

// extAssertOrdered 校验 rows 在 col 上按 dir 单调（允许相等）。
func extAssertOrdered(t *testing.T, rows []map[string]interface{}, col, dir string) {
	t.Helper()
	for i := 1; i < len(rows); i++ {
		prev, cur := extI64(rows[i-1], col), extI64(rows[i], col)
		if dir == "asc" && cur < prev {
			t.Fatalf("rows not ASC by %s: %v then %v", col, prev, cur)
		}
		if dir != "asc" && cur > prev {
			t.Fatalf("rows not DESC by %s: %v then %v", col, prev, cur)
		}
	}
}

func extBvids(rows []map[string]interface{}) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, extStr(r, "bvid"))
	}
	return out
}

func extHas(rows []map[string]interface{}, col, want string) bool {
	for _, r := range rows {
		if extStr(r, col) == want {
			return true
		}
	}
	return false
}

// ---------- 单例与 schema ----------

func TestExtDBSingletons(t *testing.T) {
	likes := GetLikesDB()
	if likes == nil {
		t.Fatal("GetLikesDB returned nil")
	}
	if again := GetLikesDB(); again != likes {
		t.Fatal("GetLikesDB not a singleton")
	}
	wl := GetWatchLaterDB()
	if wl == nil {
		t.Fatal("GetWatchLaterDB returned nil")
	}
	if again := GetWatchLaterDB(); again != wl {
		t.Fatal("GetWatchLaterDB not a singleton")
	}
	fav := GetFavoritesDB()
	if fav == nil {
		t.Fatal("GetFavoritesDB returned nil")
	}
	if again := GetFavoritesDB(); again != fav {
		t.Fatal("GetFavoritesDB not a singleton")
	}

	for _, tc := range []struct {
		db   *sql.DB
		want []string
	}{
		{likes, []string{"liked_videos"}},
		{wl, []string{"watchlater_videos"}},
		{fav, []string{"favorites_folder", "favorites_content"}},
	} {
		for _, tbl := range tc.want {
			n := extCount(t, tc.db, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", tbl)
			if n != 1 {
				t.Fatalf("table %s missing (count=%d)", tbl, n)
			}
		}
	}
	// GetFavoritesDB 的 folder_type 迁移分支
	if n := extCount(t, fav, "SELECT COUNT(*) FROM pragma_table_info('favorites_folder') WHERE name='folder_type'"); n != 1 {
		t.Fatalf("favorites_folder should have folder_type column, got count %d", n)
	}
}

// ---------- likes ----------

func TestExtLatestLikeTimeBeforeAnyInsert(t *testing.T) {
	// rep_ 测试数据可能已存在；仅在表为空时断言空分支。
	if extCount(t, GetLikesDB(), "SELECT COUNT(*) FROM liked_videos") == 0 {
		lt, c, err := GetLatestLikeTime()
		if err != nil {
			t.Fatalf("GetLatestLikeTime: %v", err)
		}
		if lt != 0 || c != 0 {
			t.Fatalf("empty likes: want (0,0), got (%d,%d)", lt, c)
		}
	} else {
		lt, c, err := GetLatestLikeTime()
		if err != nil || c == 0 {
			t.Fatalf("populated-by-others branch: lt=%d c=%d err=%v", lt, c, err)
		}
	}
}

func TestExtSaveLikedVideos(t *testing.T) {
	db := GetLikesDB()

	err := SaveLikedVideos([]LikeVideo{
		{Bvid: "", Title: "EXT 空标题应被跳过"},
		{Bvid: "ext_lv_1", Aid: 71001, Title: "EXT 一", Pubdate: 1000, Duration: 100, View: 10, IsSeen: 0},
		{Bvid: "ext_lv_2", Aid: 71002, Title: "EXT 二", Pubdate: 2000, Duration: 200, View: 20},
		{Bvid: "ext_lv_3", Aid: 71003, Title: "EXT 三", Pubdate: 4102444800, Duration: 300, View: 30},
	})
	if err != nil {
		t.Fatalf("SaveLikedVideos: %v", err)
	}

	if n := extCount(t, db, "SELECT COUNT(*) FROM liked_videos WHERE title = ?", "EXT 空标题应被跳过"); n != 0 {
		t.Fatalf("empty bvid should be skipped, found %d rows", n)
	}
	for _, b := range []string{"ext_lv_1", "ext_lv_2", "ext_lv_3"} {
		if n := extCount(t, db, "SELECT COUNT(*) FROM liked_videos WHERE bvid = ?", b); n != 1 {
			t.Fatalf("%s: want 1 row got %d", b, n)
		}
	}
	// link 由 bvid 构建
	var link string
	if err := db.QueryRow("SELECT link FROM liked_videos WHERE bvid = 'ext_lv_1'").Scan(&link); err != nil {
		t.Fatalf("read link: %v", err)
	}
	if link != "https://www.bilibili.com/video/ext_lv_1" {
		t.Fatalf("unexpected link: %q", link)
	}

	// 再次保存 = upsert 更新；is_seen 不在 DO UPDATE SET 列表中（疑似生产 bug），
	// 现有行更新后 is_seen 仍保持旧值 0。
	err = SaveLikedVideos([]LikeVideo{
		{Bvid: "ext_lv_1", Title: "EXT 一改", Pubdate: 1000, Duration: 100, View: 10, IsSeen: 1},
	})
	if err != nil {
		t.Fatalf("SaveLikedVideos update: %v", err)
	}
	var title string
	var isSeen int
	if err := db.QueryRow("SELECT title, is_seen FROM liked_videos WHERE bvid = 'ext_lv_1'").Scan(&title, &isSeen); err != nil {
		t.Fatalf("read updated row: %v", err)
	}
	if title != "EXT 一改" {
		t.Fatalf("title not updated: %q", title)
	}
	if isSeen != 0 {
		t.Fatalf("is_seen unexpectedly %d; current impl does not update it on conflict", isSeen)
	}
	if n := extCount(t, db, "SELECT COUNT(*) FROM liked_videos WHERE bvid = 'ext_lv_1'"); n != 1 {
		t.Fatalf("upsert duplicated row: %d", n)
	}

	// 直接写入含 NULL 列的行（title/pic/desc 无默认值），供 GetLikedVideos 的
	// val == nil 跳过分支使用
	mustExec(t, db, `INSERT OR IGNORE INTO liked_videos (bvid, fetch_time) VALUES ('ext_lv_null', 100)`)
}

func TestExtGetLikedVideos(t *testing.T) {
	sorts := []struct{ sort, col string }{
		{"pubdate", "pubdate"},
		{"fetch_time", "fetch_time"},
		{"duration", "duration"},
		{"view", "view"},
		{"invalid_sort", "fetch_time"}, // 非法 sort 回落默认列
		{"", "fetch_time"},             // 空 sort 亦用默认列
	}
	for _, s := range sorts {
		for _, dir := range []string{"asc", "desc"} {
			rows, total, err := GetLikedVideos(1, 1000, s.sort, dir)
			if err != nil {
				t.Fatalf("GetLikedVideos(%s,%s): %v", s.sort, dir, err)
			}
			extAssertOrdered(t, rows, s.col, dir)
			if total < 3 || total != len(rows) {
				t.Fatalf("GetLikedVideos(%s,%s): total=%d len=%d", s.sort, dir, total, len(rows))
			}
			for _, want := range []string{"ext_lv_1", "ext_lv_2", "ext_lv_3"} {
				if !extHas(rows, "bvid", want) {
					t.Fatalf("missing ext_lv row %s in %s/%s result", want, s.sort, dir)
				}
			}
			// NULL 列应从结果 map 中跳过（val == nil 分支）
			for _, r := range rows {
				if extStr(r, "bvid") == "ext_lv_null" {
					if _, ok := r["title"]; ok {
						t.Fatal("NULL title should be omitted from result map")
					}
					if _, ok := r["pic"]; ok {
						t.Fatal("NULL pic should be omitted from result map")
					}
				}
			}
			// map 字段可读且 []byte 已转 string
			if extI64(rows[0], "id") == 0 {
				t.Fatal("id column missing or zero in result map")
			}
		}
	}

	// 分页：page1+page2 应等于全量前 4 行
	full, _, err := GetLikedVideos(1, 1000, "pubdate", "desc")
	if err != nil {
		t.Fatalf("full list: %v", err)
	}
	p1, _, err := GetLikedVideos(1, 2, "pubdate", "desc")
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	p2, _, err := GetLikedVideos(2, 2, "pubdate", "desc")
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	// rep_ 行数不确定，仅要求两页均非空且与全量前缀一致
	if len(p1) != 2 || len(p2) == 0 {
		t.Fatalf("paging sizes: %d %d", len(p1), len(p2))
	}
	joined := append(extBvids(p1), extBvids(p2)...)
	want := extBvids(full[:len(p1)+len(p2)])
	if strings.Join(joined, ",") != strings.Join(want, ",") {
		t.Fatalf("pages differ from full list: %v vs %v", joined, want)
	}
	// 越界页返回空
	if p9, _, err := GetLikedVideos(99999, 10, "pubdate", "desc"); err != nil || len(p9) != 0 {
		t.Fatalf("out-of-range page: len=%d err=%v", len(p9), err)
	}
}

func TestExtLatestLikeTimePopulated(t *testing.T) {
	lt, c, err := GetLatestLikeTime()
	if err != nil {
		t.Fatalf("GetLatestLikeTime: %v", err)
	}
	// ext_lv_3 的 pubdate 为 2100 年时间戳，必然 >= 其它数据
	if lt < 4102444800 {
		t.Fatalf("latest pubdate should dominate, got %d", lt)
	}
	if c < 3 {
		t.Fatalf("count should include our 3 rows, got %d", c)
	}
}

// ---------- watchlater ----------

func TestExtWatchLaterSync(t *testing.T) {
	db := GetWatchLaterDB()

	err := SaveWatchLaterVideos([]WatchLaterVideo{
		{Bvid: "", Title: "EXT WL 空标题应被跳过"},
		{Bvid: "ext_wl_1", Aid: 71011, Title: "EXT WL1", Link: "https://x/ext_wl_1", AddAt: 500, Pubdate: 1000, Duration: 100, View: 10},
		{Bvid: "ext_wl_2", Aid: 71012, Title: "EXT WL2", Link: "https://x/ext_wl_2", AddAt: 400, Pubdate: 2000, Duration: 200, View: 20},
		{Bvid: "ext_wl_3", Title: "EXT WL3", AddAt: 300, Pubdate: 1500, Duration: 150, View: 15},
		{Bvid: "ext_wl_4", Title: "EXT WL4", AddAt: 200, Pubdate: 1200, Duration: 120, View: 12},
	})
	if err != nil {
		t.Fatalf("SaveWatchLaterVideos initial: %v", err)
	}
	if n := extCount(t, db, "SELECT COUNT(*) FROM watchlater_videos WHERE title = ?", "EXT WL 空标题应被跳过"); n != 0 {
		t.Fatalf("empty bvid should be skipped, found %d", n)
	}
	// 差量同步语义：远端列表之外的本地行（含 rep_ 前缀）被删除 → 表内恰为我们的 4 行
	if n := extCount(t, db, "SELECT COUNT(*) FROM watchlater_videos"); n != 4 {
		t.Fatalf("after initial sync want exactly 4 rows, got %d", n)
	}

	// 单条读取（link 原样保存，不重新构建）
	v, err := GetWatchLaterVideoByBvid("ext_wl_1")
	if err != nil {
		t.Fatalf("GetWatchLaterVideoByBvid: %v", err)
	}
	if v == nil {
		t.Fatal("ext_wl_1 not found")
	}
	if v.Title != "EXT WL1" || v.Link != "https://x/ext_wl_1" || v.Aid != 71011 || v.AddAt != 500 || v.View != 10 {
		t.Fatalf("field mismatch: %+v", v)
	}
	if v.FetchTime < time.Now().Unix()-60 {
		t.Fatalf("fetch_time not recent: %d", v.FetchTime)
	}

	// 不存在的 bvid → (nil, nil)
	missing, err := GetWatchLaterVideoByBvid("ext_wl_absent")
	if err != nil || missing != nil {
		t.Fatalf("missing bvid: v=%v err=%v", missing, err)
	}

	// 二次同步：更新 wl_1，差量删除 wl_4
	err = SaveWatchLaterVideos([]WatchLaterVideo{
		{Bvid: "ext_wl_1", Aid: 71011, Title: "EXT WL1新", AddAt: 500, Pubdate: 1000, Duration: 100, View: 10},
		{Bvid: "ext_wl_2", Title: "EXT WL2", AddAt: 400, Pubdate: 2000, Duration: 200, View: 20},
		{Bvid: "ext_wl_3", Title: "EXT WL3", AddAt: 300, Pubdate: 1500, Duration: 150, View: 15},
	})
	if err != nil {
		t.Fatalf("SaveWatchLaterVideos update: %v", err)
	}
	v, _ = GetWatchLaterVideoByBvid("ext_wl_1")
	if v == nil || v.Title != "EXT WL1新" {
		t.Fatalf("watchlater not updated: %+v", v)
	}
	if n := extCount(t, db, "SELECT COUNT(*) FROM watchlater_videos WHERE bvid = 'ext_wl_4'"); n != 0 {
		t.Fatalf("ext_wl_4 should be pruned, got %d", n)
	}
	if n := extCount(t, db, "SELECT COUNT(*) FROM watchlater_videos"); n != 3 {
		t.Fatalf("after update sync want 3 rows, got %d", n)
	}

	// 单条删除
	if err := DeleteWatchLaterVideo("ext_wl_3"); err != nil {
		t.Fatalf("DeleteWatchLaterVideo: %v", err)
	}
	if v, err := GetWatchLaterVideoByBvid("ext_wl_3"); err != nil || v != nil {
		t.Fatalf("deleted row still present: %+v %v", v, err)
	}
	if n := extCount(t, db, "SELECT COUNT(*) FROM watchlater_videos"); n != 2 {
		t.Fatalf("want 2 rows after delete, got %d", n)
	}
}

func TestExtGetWatchLaterVideos(t *testing.T) {
	// 此刻表内恰有 ext_wl_1 / ext_wl_2；再直插一条仅 bvid/title/fetch_time 的
	// NULL 列行（pic/desc/tname/owner 等为 NULL），覆盖 map 循环的 val == nil 跳过分支。
	mustExec(t, GetWatchLaterDB(),
		`INSERT OR IGNORE INTO watchlater_videos (bvid, title, fetch_time) VALUES ('ext_wl_null', 'EXT 空列', 100)`)
	cases := []struct{ sort, col string }{
		{"pubdate", "pubdate"},
		{"add_at", "add_at"},
		{"fetch_time", "fetch_time"},
		{"duration", "duration"},
		{"view", "view"},
		{"bogus", "add_at"}, // 非法 sort 回落 add_at
		{"", "add_at"},
	}
	for _, c := range cases {
		// wl1: pub 1000 < wl2: 2000；add 500 > 400；duration/view wl2 大；null 行各数值列默认 0/100
		rows, total, err := GetWatchLaterVideos(1, 100, c.sort, "desc")
		if err != nil {
			t.Fatalf("GetWatchLaterVideos(%s): %v", c.sort, err)
		}
		if total != 3 || len(rows) != 3 {
			t.Fatalf("total=%d len=%d", total, len(rows))
		}
		extAssertOrdered(t, rows, c.col, "desc")

		asc, _, err := GetWatchLaterVideos(1, 100, c.sort, "asc")
		if err != nil {
			t.Fatalf("asc: %v", err)
		}
		extAssertOrdered(t, asc, c.col, "asc")
	}

	// pubdate desc 应为 wl2 在前
	rows, _, _ := GetWatchLaterVideos(1, 100, "pubdate", "desc")
	if extStr(rows[0], "bvid") != "ext_wl_2" {
		t.Fatalf("pubdate desc head = %s", extStr(rows[0], "bvid"))
	}
	for _, r := range rows {
		if extStr(r, "bvid") == "ext_wl_null" {
			if _, ok := r["pic"]; ok {
				t.Fatal("NULL pic should be omitted from result map")
			}
			if _, ok := r["owner_name"]; ok {
				t.Fatal("NULL owner_name should be omitted from result map")
			}
		}
	}
	// 分页
	p1, total, err := GetWatchLaterVideos(1, 1, "pubdate", "desc")
	if err != nil || len(p1) != 1 || total != 3 {
		t.Fatalf("page1: len=%d total=%d err=%v", len(p1), total, err)
	}
	p2, _, err := GetWatchLaterVideos(2, 1, "pubdate", "desc")
	if err != nil || len(p2) != 1 || extStr(p2[0], "bvid") != "ext_wl_1" {
		t.Fatalf("page2: %+v err=%v", p2, err)
	}
	p3, _, err := GetWatchLaterVideos(3, 1, "pubdate", "desc")
	if err != nil || len(p3) != 1 || extStr(p3[0], "bvid") != "ext_wl_null" {
		t.Fatalf("page3: %+v err=%v", p3, err)
	}
	if p4, _, err := GetWatchLaterVideos(4, 1, "pubdate", "desc"); err != nil || len(p4) != 0 {
		t.Fatalf("page4 should be empty: %v %v", p4, err)
	}
}

// ---------- favorites: folders ----------

func TestExtFavoriteFoldersSync(t *testing.T) {
	db := GetFavoritesDB()
	base := FavoriteFolder{Fid: 1, Mid: 71021, Cover: "https://cover", Intro: "in", State: 0, MediaCount: 9}

	f1 := base
	f1.MediaID, f1.Title, f1.Mtime, f1.FolderType = 71101, "EXT 收藏夹A", 300, 0
	f2 := base
	f2.MediaID, f2.Title, f2.Mtime, f2.FolderType = 71102, "EXT 收藏夹B", 200, 0
	f3 := base
	f3.MediaID, f3.Title, f3.Mtime, f3.FolderType, f3.Ctime = 71103, "EXT 他夹C", 100, 1, 55

	if err := SaveFavoriteFolders([]FavoriteFolder{f1, f2, f3}); err != nil {
		t.Fatalf("SaveFavoriteFolders: %v", err)
	}
	// 裁剪语义：favorites_folder 中只剩这三条
	if n := extCount(t, db, "SELECT COUNT(*) FROM favorites_folder"); n != 3 {
		t.Fatalf("want exactly 3 folders, got %d", n)
	}

	// created=true 实际返回全部收藏夹（不按 folder_type 过滤，疑似生产 bug）
	created, total, err := GetFavoriteFolders(true)
	if err != nil {
		t.Fatalf("GetFavoriteFolders: %v", err)
	}
	if total != 3 || len(created) != 3 {
		t.Fatalf("created folders: total=%d len=%d", total, len(created))
	}
	if extI64(created[0], "media_id") != 71101 { // mtime DESC
		t.Fatalf("first folder by mtime desc = %d", extI64(created[0], "media_id"))
	}
	if extStr(created[0], "title") != "EXT 收藏夹A" || extI64(created[0], "folder_type") != 0 {
		t.Fatalf("folder map fields wrong: %v", created[0])
	}

	// created=false 直接返回空（另一分支）
	got, total, err := GetFavoriteFolders(false)
	if err != nil || len(got) != 0 || total != 0 {
		t.Fatalf("created=false: len=%d total=%d err=%v", len(got), total, err)
	}

	// 按类型过滤 + 分页
	type0, t0, err := GetFavoriteFoldersByType(0, 1, 100)
	if err != nil || t0 != 2 || len(type0) != 2 {
		t.Fatalf("type0: total=%d len=%d err=%v", t0, len(type0), err)
	}
	for _, r := range type0 {
		if extI64(r, "folder_type") != 0 {
			t.Fatalf("type filter leaked: %+v", r)
		}
	}
	type1, t1, err := GetFavoriteFoldersByType(1, 1, 100)
	if err != nil || t1 != 1 || extI64(type1[0], "media_id") != 71103 {
		t.Fatalf("type1: total=%d err=%v rows=%v", t1, err, type1)
	}
	page2, _, err := GetFavoriteFoldersByType(0, 2, 1)
	if err != nil || len(page2) != 1 || extI64(page2[0], "media_id") != 71102 {
		t.Fatalf("type0 page2: %+v err=%v", page2, err)
	}
	if none, tn, err := GetFavoriteFoldersByType(7, 1, 10); err != nil || tn != 0 || len(none) != 0 {
		t.Fatalf("unknown type: total=%d len=%d err=%v", tn, len(none), err)
	}

	// upsert：更新 f1 标题与 mtime；未列出的 f2 被裁剪
	f1u := f1
	f1u.Title, f1u.Mtime = "EXT 收藏夹A新", 400
	if err := SaveFavoriteFolders([]FavoriteFolder{f1u, f3}); err != nil {
		t.Fatalf("SaveFavoriteFolders update: %v", err)
	}
	if n := extCount(t, db, "SELECT COUNT(*) FROM favorites_folder WHERE media_id = 71102"); n != 0 {
		t.Fatalf("f2 should be pruned, got %d", n)
	}
	if title := func() string {
		var s string
		_ = db.QueryRow("SELECT title FROM favorites_folder WHERE media_id = 71101").Scan(&s)
		return s
	}(); title != "EXT 收藏夹A新" {
		t.Fatalf("folder not updated: %q", title)
	}

	// 空列表 → 清空全部分支
	if err := SaveFavoriteFolders([]FavoriteFolder{}); err != nil {
		t.Fatalf("SaveFavoriteFolders empty: %v", err)
	}
	if n := extCount(t, db, "SELECT COUNT(*) FROM favorites_folder"); n != 0 {
		t.Fatalf("empty list should clear table, got %d", n)
	}
}

// TestExtFavoriteFolderNullColumns 直插含 NULL 文本列的收藏夹行，
// 覆盖 GetFavoriteFolders / GetFavoriteFoldersByType map 循环的 NULL 跳过分支。
// 本测试内不做任何 SaveFavoriteFolders 调用，避免裁剪掉该手工行。
func TestExtFavoriteFolderNullColumns(t *testing.T) {
	db := GetFavoritesDB()
	mustExec(t, db, `INSERT OR IGNORE INTO favorites_folder (media_id, fetch_time) VALUES (71109, 100)`)

	folders, total, err := GetFavoriteFolders(true)
	if err != nil || total != 1 || len(folders) != 1 {
		t.Fatalf("null-col folder: total=%d len=%d err=%v", total, len(folders), err)
	}
	if extI64(folders[0], "media_id") != 71109 {
		t.Fatalf("media_id wrong: %v", folders[0])
	}
	if _, ok := folders[0]["title"]; ok {
		t.Fatal("NULL title should be omitted from result map")
	}
	// folder_type 默认 0 → 类型过滤可见
	byType, t2, err := GetFavoriteFoldersByType(0, 1, 10)
	if err != nil || t2 != 1 || len(byType) != 1 {
		t.Fatalf("null-col folder by type: total=%d len=%d err=%v", t2, len(byType), err)
	}
	if _, ok := byType[0]["cover"]; ok {
		t.Fatal("NULL cover should be omitted from result map")
	}
	if _, ok := byType[0]["intro"]; ok {
		t.Fatal("NULL intro should be omitted from result map")
	}
}

// ---------- favorites: contents ----------

func TestExtFavoriteContentsSync(t *testing.T) {
	db := GetFavoritesDB()
	mk := func(cid int64, title string, fav, dur int64) FavoriteContent {
		return FavoriteContent{
			MediaID: 71101, ContentID: cid, Type: 21, Title: title,
			Cover: "https://c", Bvid: fmt.Sprintf("ext_fc_%d", cid), Intro: "i",
			Page: 1, Duration: int(dur), UpperMid: 71031, FavTime: fav,
			CreatorName: "EXT up", CreatorFace: "https://face", Collect: 1, Play: 2,
		}
	}
	if err := SaveFavoriteContents(71101, []FavoriteContent{
		mk(71201, "EXT 内容一", 100, 10),
		mk(71202, "EXT 内容二", 200, 20),
		mk(71203, "EXT 内容三", 300, 30),
	}); err != nil {
		t.Fatalf("SaveFavoriteContents: %v", err)
	}

	rows, total, err := GetFavoriteContents(71101, 1, 100)
	if err != nil || total != 3 || len(rows) != 3 {
		t.Fatalf("contents: total=%d len=%d err=%v", total, len(rows), err)
	}
	// fav_time DESC
	if extI64(rows[0], "content_id") != 71203 || extI64(rows[2], "content_id") != 71201 {
		t.Fatalf("order wrong: %v", extBvids(rows))
	}
	if extStr(rows[0], "bvid") != "ext_fc_71203" || extStr(rows[0], "creator_name") != "EXT up" {
		t.Fatalf("map fields wrong: %v", rows[0])
	}
	if extCount(t, db, "SELECT COUNT(*) FROM favorites_content WHERE media_id = 71101") != 3 {
		t.Fatal("row count mismatch in table")
	}

	// 再同步：upsert 71201（fav_time 提到 500），未列出的旧内容保留（增量合并）
	if err := SaveFavoriteContents(71101, []FavoriteContent{
		mk(71201, "EXT 内容一改", 500, 99),
	}); err != nil {
		t.Fatalf("SaveFavoriteContents upsert: %v", err)
	}
	rows, total, err = GetFavoriteContents(71101, 1, 100)
	if err != nil || total != 3 {
		t.Fatalf("after upsert: total=%d err=%v", total, err)
	}
	if extI64(rows[0], "content_id") != 71201 || extStr(rows[0], "title") != "EXT 内容一改" || extI64(rows[0], "duration") != 99 {
		t.Fatalf("upserted row wrong: %v", rows[0])
	}

	// 分页
	p, _, err := GetFavoriteContents(71101, 2, 2)
	if err != nil || len(p) != 1 || extI64(p[0], "content_id") != 71202 {
		t.Fatalf("page2: %+v err=%v", p, err)
	}
	// 未知 media_id → 空
	if e, et, err := GetFavoriteContents(79999, 1, 10); err != nil || et != 0 || len(e) != 0 {
		t.Fatalf("unknown media: total=%d len=%d err=%v", et, len(e), err)
	}

	// 直插仅 media_id/content_id/fetch_time 的 NULL 列行，覆盖 GetFavoriteContents 的
	// val == nil 跳过分支（fav_time 默认 0 → 排在最后）
	mustExec(t, db, `INSERT OR IGNORE INTO favorites_content (media_id, content_id, fetch_time) VALUES (71105, 71250, 100)`)
	nullRows, nt, err := GetFavoriteContents(71105, 1, 10)
	if err != nil || nt != 1 || len(nullRows) != 1 {
		t.Fatalf("null-col content: total=%d len=%d err=%v", nt, len(nullRows), err)
	}
	if extI64(nullRows[0], "content_id") != 71250 {
		t.Fatalf("content_id wrong: %v", nullRows[0])
	}
	if _, ok := nullRows[0]["title"]; ok {
		t.Fatal("NULL title should be omitted from result map")
	}
	if _, ok := nullRows[0]["bvid"]; ok {
		t.Fatal("NULL bvid should be omitted from result map")
	}
}

func TestExtFolderSyncState(t *testing.T) {
	st, err := GetFolderSyncState(71101)
	if err != nil {
		t.Fatalf("GetFolderSyncState: %v", err)
	}
	if st == nil {
		t.Fatal("nil state")
	}
	if st.LatestFavTime != 500 || st.Count != 3 || len(st.ContentIDs) != 3 {
		t.Fatalf("state mismatch: %+v", st)
	}
	if !st.ContentIDs[71201] || !st.ContentIDs[71203] {
		t.Fatalf("content ids missing: %v", st.ContentIDs)
	}

	// 无内容的 media_id：回归——MAX(fav_time) 曾为 NULL 导致 Scan 报错，
	// 已用 COALESCE 修复，应返回全零状态。
	st, err = GetFolderSyncState(79998)
	if err != nil {
		t.Fatalf("空文件夹应返回零状态而非报错: %v", err)
	}
	if st == nil || st.LatestFavTime != 0 || st.Count != 0 || len(st.ContentIDs) != 0 {
		t.Fatalf("零状态错误: %+v", st)
	}
}

// TestExtFavoriteContentTitlePreserveDeadlock 回归：SaveFavoriteContents
// 曾在"保留原标题"分支用 db 直接 QueryRow——事务持有唯一连接
// (SetMaxOpenConns(1)) 时必然死锁。修复后查询走 tx。
// 保留 goroutine+超时探测，防死锁回归时测试挂死整个套件。
func TestExtFavoriteContentTitlePreserveDeadlock(t *testing.T) {
	db := GetFavoritesDB()
	// 先用正常标题写入一行
	if err := SaveFavoriteContents(71104, []FavoriteContent{
		{MediaID: 71104, ContentID: 71240, Title: "EXT 原标题", Bvid: "ext_dp_1", FavTime: 1},
	}); err != nil {
		t.Fatalf("seed content: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- SaveFavoriteContents(71104, []FavoriteContent{
			{MediaID: 71104, ContentID: 71240, Title: "已失效视频", Bvid: "ext_dp_1", FavTime: 2},
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("SaveFavoriteContents with 已失效视频: %v", err)
		}
		var title string
		var favTime int64
		if err := db.QueryRow("SELECT title, fav_time FROM favorites_content WHERE media_id = 71104 AND content_id = 71240").Scan(&title, &favTime); err != nil {
			t.Fatalf("read title: %v", err)
		}
		if title != "EXT 原标题" {
			t.Fatalf("old title should be preserved, got %q", title)
		}
		if favTime != 2 {
			t.Fatalf("fav_time should be updated to 2, got %d", favTime)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SaveFavoriteContents 在保留原标题分支死锁（回归）")
	}
}
