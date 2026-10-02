package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
)

// dynamic_test.go 命名空间：动态库是独立文件（bilibili_dynamic.db），只有本文件写入，
// 因此可以做精确断言；每个测试开头用 dynReset 清空两张表，保证结果与文件内顺序无关。
// host_mid 一律 dyn_ 前缀，动态 id 一律 dynid_ 前缀。

func dynDB(t *testing.T) *sql.DB {
	t.Helper()
	db := GetDynamicDB()
	if db == nil {
		t.Fatal("GetDynamicDB returned nil")
	}
	return db
}

func dynReset(t *testing.T) {
	t.Helper()
	db := dynDB(t)
	if _, err := db.Exec("DELETE FROM dynamics"); err != nil {
		t.Fatalf("reset dynamics: %v", err)
	}
	if _, err := db.Exec("DELETE FROM dynamic_hosts"); err != nil {
		t.Fatalf("reset hosts: %v", err)
	}
}

func dynHostRow(t *testing.T, hostMid string) (upName, facePath, lastDynamicID string, itemCount, coreCount int, lastPublishTS, lastFetchTS int64) {
	t.Helper()
	err := dynDB(t).QueryRow(`
		SELECT up_name, face_path, last_dynamic_id, item_count, core_count, last_publish_ts, last_fetch_time
		FROM dynamic_hosts WHERE host_mid = ?`, hostMid).
		Scan(&upName, &facePath, &lastDynamicID, &itemCount, &coreCount, &lastPublishTS, &lastFetchTS)
	if err != nil {
		t.Fatalf("read host %s: %v", hostMid, err)
	}
	return
}

func dynItem(id, typ, face, author string, publishTS int64) DynamicItem {
	return DynamicItem{
		ID:          id,
		Type:        typ,
		AuthorName:  author,
		AuthorFace:  face,
		Txt:         "txt-" + id,
		Bvid:        "BVdyn" + id,
		Title:       "title-" + id,
		Desc:        "desc-" + id,
		Cover:       "cover-" + id,
		PublishTS:   publishTS,
		MediaLocals: []string{"media/a.jpg", "media/b.png"},
	}
}

// TestDynSchemaGetDynamicDB 单例初始化：两张表存在且迁移列可用。
func TestDynSchemaGetDynamicDB(t *testing.T) {
	db := dynDB(t)
	if db != GetDynamicDB() {
		t.Fatal("GetDynamicDB should return the same handle twice")
	}
	var cnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM dynamic_hosts").Scan(&cnt); err != nil {
		t.Fatalf("dynamic_hosts missing: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM dynamics").Scan(&cnt); err != nil {
		t.Fatalf("dynamics missing: %v", err)
	}
	// author_face / last_dynamic_id 迁移列
	if err := db.QueryRow("SELECT author_face FROM dynamics LIMIT 0").Scan(new(string)); err != nil && err != sql.ErrNoRows {
		t.Fatalf("author_face column: %v", err)
	}
	if err := db.QueryRow("SELECT last_dynamic_id FROM dynamic_hosts LIMIT 0").Scan(new(string)); err != nil && err != sql.ErrNoRows {
		t.Fatalf("last_dynamic_id column: %v", err)
	}
}

// TestDynSaveDynamicHostUpsert 插入 + 冲突更新（face/last_dynamic_id 仅非空才覆盖）。
func TestDynSaveDynamicHostUpsert(t *testing.T) {
	dynReset(t)

	h := DynamicHost{HostMid: "dyn_host_a", UpName: "动UP甲", FacePath: "face1",
		ItemCount: 5, CoreCount: 2, LastPublishTS: 100, LastFetchTime: 200, LastDynamicID: "ld1"}
	if err := SaveDynamicHost(h); err != nil {
		t.Fatalf("insert host: %v", err)
	}
	name, face, ldid, items, core, ts, fetch := dynHostRow(t, "dyn_host_a")
	if name != "动UP甲" || face != "face1" || ldid != "ld1" || items != 5 || core != 2 || ts != 100 || fetch != 200 {
		t.Fatalf("host roundtrip wrong: %v %v %v %v %v %v %v", name, face, ldid, items, core, ts, fetch)
	}

	// 二次 upsert：face / last_dynamic_id 传空串应保留旧值，其余覆盖
	h2 := DynamicHost{HostMid: "dyn_host_a", UpName: "动UP甲改名", FacePath: "",
		ItemCount: 9, CoreCount: 3, LastPublishTS: 150, LastFetchTime: 300, LastDynamicID: ""}
	if err := SaveDynamicHost(h2); err != nil {
		t.Fatalf("upsert host: %v", err)
	}
	name, face, ldid, items, core, ts, fetch = dynHostRow(t, "dyn_host_a")
	if name != "动UP甲改名" || items != 9 || core != 3 || ts != 150 || fetch != 300 {
		t.Fatalf("overwrite fields wrong: name=%v items=%v core=%v ts=%v fetch=%v", name, items, core, ts, fetch)
	}
	if face != "face1" || ldid != "ld1" {
		t.Fatalf("empty face/last_dynamic_id must be preserved: face=%v ldid=%v", face, ldid)
	}

	// 非空 face/ldid 才覆盖
	h3 := h2
	h3.FacePath = "face2"
	h3.LastDynamicID = "ld2"
	if err := SaveDynamicHost(h3); err != nil {
		t.Fatalf("upsert host 3: %v", err)
	}
	_, face, ldid, _, _, _, _ = dynHostRow(t, "dyn_host_a")
	if face != "face2" || ldid != "ld2" {
		t.Fatalf("non-empty face/ldid must overwrite: face=%v ldid=%v", face, ldid)
	}

	// 只剩一行
	hosts, err := GetDynamicHosts(10, 0)
	if err != nil {
		t.Fatalf("GetDynamicHosts: %v", err)
	}
	if len(hosts) != 1 || hosts[0].HostMid != "dyn_host_a" {
		t.Fatalf("want exactly 1 host, got %v", hosts)
	}
	// GetDynamicHosts 不查询 last_dynamic_id，Scan 不填该字段
	if hosts[0].LastDynamicID != "" {
		t.Fatalf("LastDynamicID should stay zero in list view: %q", hosts[0].LastDynamicID)
	}
}

// TestDynGetDynamicHostsLimit limit/offset 钳制与排序。
func TestDynGetDynamicHostsLimit(t *testing.T) {
	dynReset(t)
	// 60 个 host，last_fetch_time = i，倒序应从 59 开始
	for i := 0; i < 60; i++ {
		if err := SaveDynamicHost(DynamicHost{
			HostMid:       fmt.Sprintf("dyn_host_%02d", i),
			UpName:        "h",
			LastFetchTime: int64(i),
		}); err != nil {
			t.Fatalf("seed host %d: %v", i, err)
		}
	}

	// 正常 limit
	hosts, err := GetDynamicHosts(10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 10 || hosts[0].HostMid != "dyn_host_59" {
		t.Fatalf("limit 10 wrong: len=%d first=%v", len(hosts), hosts)
	}

	// limit<=0 / >200 都钳到 50
	for _, bad := range []int{0, -1, 201, 10000} {
		hosts, err = GetDynamicHosts(bad, 0)
		if err != nil {
			t.Fatalf("limit %d: %v", bad, err)
		}
		if len(hosts) != 50 {
			t.Fatalf("limit %d should clamp to 50, got %d", bad, len(hosts))
		}
	}
	// 合法上限 200 不钳制：60 行全量返回
	hosts, err = GetDynamicHosts(200, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 60 || hosts[0].HostMid != "dyn_host_59" || hosts[59].HostMid != "dyn_host_00" {
		t.Fatalf("limit 200 should return all 60 sorted desc, got %d", len(hosts))
	}

	// 负 offset 归 0
	hosts, err = GetDynamicHosts(10, -5)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 10 || hosts[0].HostMid != "dyn_host_59" {
		t.Fatalf("negative offset should behave as 0, got %v", hosts)
	}
	// offset 越界
	hosts, err = GetDynamicHosts(10, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 0 {
		t.Fatalf("offset 100 of 60 rows should be empty")
	}
}

// TestDynSaveDynamics insert/update 计数、存在性、按 ID 读取、media JSON 解码、face 保留。
func TestDynSaveDynamics(t *testing.T) {
	dynReset(t)
	host := "dyn_host_sd"

	items := []DynamicItem{
		dynItem("dynid_1", "DYNAMIC_TYPE_AV", "faceX", "动UP甲", 100),
		dynItem("dynid_2", "DYNAMIC_TYPE_DRAW", "faceX", "动UP甲", 300),
		dynItem("dynid_3", "DYNAMIC_TYPE_LIVING", "faceX", "动UP甲", 200),
	}
	inserted, err := SaveDynamics(host, items)
	if err != nil {
		t.Fatalf("SaveDynamics: %v", err)
	}
	if inserted != 3 {
		t.Fatalf("inserted=%d want 3", inserted)
	}

	// 首次批量保存后内部自动建 host：upName/face_path 来自 items；本批统计与全库一致。
	name, face, _, ic, cc, ts, _ := dynHostRow(t, host)
	if name != "动UP甲" || face != "faceX" {
		t.Fatalf("auto host wrong: name=%v face=%v", name, face)
	}
	if ic != 3 || cc != 2 || ts != 300 {
		t.Fatalf("auto host stats wrong: items=%d core=%d ts=%d (want 3/2/300)", ic, cc, ts)
	}

	if !IsDynamicExists("dynid_1") || !IsDynamicExists("dynid_3") {
		t.Fatal("saved dynamics should exist")
	}
	if IsDynamicExists("dynid_missing") {
		t.Fatal("unknown id must not exist")
	}

	item, err := GetDynamicByID("dynid_1")
	if err != nil {
		t.Fatalf("GetDynamicByID: %v", err)
	}
	if item.ID != "dynid_1" || item.Type != "DYNAMIC_TYPE_AV" || item.HostMid != host ||
		item.PublishTS != 100 || item.Txt != "txt-dynid_1" || item.Bvid != "BVdyndynid_1" ||
		item.Title != "title-dynid_1" || item.Desc != "desc-dynid_1" || item.Cover != "cover-dynid_1" ||
		item.AuthorName != "动UP甲" || item.AuthorFace != "faceX" {
		t.Fatalf("item roundtrip wrong: %+v", item)
	}
	if len(item.MediaLocals) != 2 || item.MediaLocals[0] != "media/a.jpg" || item.MediaLocals[1] != "media/b.png" {
		t.Fatalf("media_locals decode wrong: %v", item.MediaLocals)
	}
	if len(item.LiveMediaLocals) != 0 {
		t.Fatalf("live_media_locals should be empty, got %v", item.LiveMediaLocals)
	}

	// 重新保存：txt 更新；author_face 传空应保留旧值（CASE 分支）
	re := dynItem("dynid_1", "DYNAMIC_TYPE_AV", "", "动UP甲", 100)
	re.Txt = "txt-updated"
	inserted, err = SaveDynamics(host, []DynamicItem{re})
	if err != nil {
		t.Fatalf("SaveDynamics update: %v", err)
	}
	if inserted != 1 {
		t.Fatalf("update should also count as affected row, got %d", inserted)
	}
	item, _ = GetDynamicByID("dynid_1")
	if item.Txt != "txt-updated" {
		t.Fatalf("txt not updated: %q", item.Txt)
	}
	if item.AuthorFace != "faceX" {
		t.Fatalf("empty author_face must preserve old value, got %q", item.AuthorFace)
	}
	// 非空 author_face 覆盖
	re2 := dynItem("dynid_1", "DYNAMIC_TYPE_AV", "faceY", "动UP甲", 100)
	if _, err := SaveDynamics(host, []DynamicItem{re2}); err != nil {
		t.Fatal(err)
	}
	item, _ = GetDynamicByID("dynid_1")
	if item.AuthorFace != "faceY" {
		t.Fatalf("non-empty author_face must overwrite, got %q", item.AuthorFace)
	}

	// 回归：单条重存后 host 统计按全库重算（旧实现把 core_count/last_publish_ts
	// 缩小到本批次，增量保存会让统计越刷越小）。face 仍保留库内旧值。
	name, face, _, ic, cc, ts, _ = dynHostRow(t, host)
	if name != "动UP甲" || face != "faceX" {
		t.Fatalf("host after partial resave wrong: name=%v face=%v", name, face)
	}
	if ic != 3 || cc != 2 || ts != 300 {
		t.Fatalf("host whole-db stats wrong: items=%d core=%d ts=%d (want 3/2/300)", ic, cc, ts)
	}

	// 空 items：0 且不建 host
	emptyHost := "dyn_host_empty"
	inserted, err = SaveDynamics(emptyHost, nil)
	if err != nil {
		t.Fatalf("SaveDynamics empty: %v", err)
	}
	if inserted != 0 {
		t.Fatalf("empty insert count %d", inserted)
	}
	var n int
	if err := dynDB(t).QueryRow("SELECT COUNT(*) FROM dynamic_hosts WHERE host_mid = ?", emptyHost).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("empty SaveDynamics must not create host row")
	}

	// GetDynamicByID 缺失 → error
	if it, err := GetDynamicByID("dynid_none"); err == nil || it != nil {
		t.Fatalf("want error for missing id, got item=%v err=%v", it, err)
	}
}

// TestDynInvalidMediaJSON 非法规格 JSON 解码失败应被吞掉、字段为空。
func TestDynInvalidMediaJSON(t *testing.T) {
	dynReset(t)
	db := dynDB(t)
	if _, err := db.Exec(`INSERT INTO dynamics (id_str, host_mid, publish_ts, media_locals, live_media_locals)
		VALUES ('dynid_badjson', 'dyn_host_bad', 10, 'not-json', '[1,2]')`); err != nil {
		t.Fatalf("insert raw: %v", err)
	}
	item, err := GetDynamicByID("dynid_badjson")
	if err != nil {
		t.Fatalf("GetDynamicByID bad json: %v", err)
	}
	// "not-json" 直接解码失败 → nil；"[1,2]" 解码到 []string 部分成功后报错，错误被吞掉。
	if len(item.MediaLocals) != 0 {
		t.Fatalf("invalid JSON must decode to empty: %v", item.MediaLocals)
	}
	// GetDynamicSpace 同一解码路径
	total, items, err := GetDynamicSpace("dyn_host_bad", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || len(items[0].MediaLocals) != 0 {
		t.Fatalf("space decode wrong: total=%d items=%+v", total, items)
	}
}

// TestDynDeleteDynamic 删除单条。
func TestDynDeleteDynamic(t *testing.T) {
	dynReset(t)
	host := "dyn_host_del"
	if _, err := SaveDynamics(host, []DynamicItem{dynItem("dynid_d1", "DYNAMIC_TYPE_AV", "f", "u", 1)}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteDynamic("dynid_d1"); err != nil {
		t.Fatalf("DeleteDynamic: %v", err)
	}
	if IsDynamicExists("dynid_d1") {
		t.Fatal("deleted dynamic still exists")
	}
	// 删除不存在的 ID 不报错
	if err := DeleteDynamic("dynid_never"); err != nil {
		t.Fatalf("delete missing should be nil: %v", err)
	}
	cnt, err := GetDynamicCount(host)
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 0 {
		t.Fatalf("count after delete = %d", cnt)
	}
}

// TestDynUpdateDynamicHostStats 统计重算（AV+DRAW 为 core）。
func TestDynUpdateDynamicHostStats(t *testing.T) {
	dynReset(t)
	host := "dyn_host_stats"
	items := []DynamicItem{
		dynItem("dynid_s1", "DYNAMIC_TYPE_AV", "f", "u", 10),
		dynItem("dynid_s2", "DYNAMIC_TYPE_DRAW", "f", "u", 999),
		dynItem("dynid_s3", "DYNAMIC_TYPE_LIVING", "f", "u", 50),
		dynItem("dynid_s4", "DYNAMIC_TYPE_ARTICLE", "f", "u", 20),
	}
	if _, err := SaveDynamics(host, items); err != nil {
		t.Fatal(err)
	}
	// 人为制造不一致：删掉一条，再让 UpdateDynamicHostStats 重算
	if err := DeleteDynamic("dynid_s2"); err != nil {
		t.Fatal(err)
	}
	if _, err := dynDB(t).Exec("UPDATE dynamic_hosts SET item_count=99, core_count=99, last_publish_ts=99 WHERE host_mid=?", host); err != nil {
		t.Fatal(err)
	}
	if err := UpdateDynamicHostStats(host); err != nil {
		t.Fatalf("UpdateDynamicHostStats: %v", err)
	}
	_, _, _, ic, cc, ts, _ := dynHostRow(t, host)
	if ic != 3 || cc != 1 || ts != 50 {
		t.Fatalf("recomputed stats wrong: items=%d core=%d ts=%d want 3/1/50", ic, cc, ts)
	}

	// host 不存在：UPDATE 0 行，返回 nil
	if err := UpdateDynamicHostStats("dyn_host_nonexistent"); err != nil {
		t.Fatalf("update for missing host should be nil: %v", err)
	}

	// 没有任何动态的 host：清零
	if err := SaveDynamicHost(DynamicHost{HostMid: host, UpName: "u", ItemCount: 42, LastFetchTime: 7}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteDynamic("dynid_s1"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteDynamic("dynid_s3"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteDynamic("dynid_s4"); err != nil {
		t.Fatal(err)
	}
	if err := UpdateDynamicHostStats(host); err != nil {
		t.Fatal(err)
	}
	_, _, _, ic, cc, ts, fetch := dynHostRow(t, host)
	if ic != 0 || cc != 0 || ts != 0 {
		t.Fatalf("empty host stats wrong: %d/%d/%d", ic, cc, ts)
	}
	// UpdateDynamicHostStats 不应覆盖 last_fetch_time
	if fetch != 7 {
		t.Fatalf("last_fetch_time must be untouched by stats update, got %d", fetch)
	}
}

// TestDynGetDynamicSpace 分页、排序、limit 钳制、总数。
func TestDynGetDynamicSpace(t *testing.T) {
	dynReset(t)
	host := "dyn_host_space"
	var items []DynamicItem
	for i := 1; i <= 25; i++ {
		it := dynItem(fmt.Sprintf("dynid_p%02d", i), "DYNAMIC_TYPE_NATIVE", "f", "u", int64(i*10))
		items = append(items, it)
	}
	if _, err := SaveDynamics(host, items); err != nil {
		t.Fatal(err)
	}
	// 干扰 host 不应计入
	if _, err := SaveDynamics("dyn_host_other", []DynamicItem{dynItem("dynid_o1", "DYNAMIC_TYPE_AV", "f", "o", 1000)}); err != nil {
		t.Fatal(err)
	}

	total, page, err := GetDynamicSpace(host, 10, 0)
	if err != nil {
		t.Fatalf("GetDynamicSpace: %v", err)
	}
	if total != 25 || len(page) != 10 {
		t.Fatalf("total=%d len=%d want 25/10", total, len(page))
	}
	if page[0].PublishTS != 250 || page[9].PublishTS != 160 {
		t.Fatalf("order desc wrong: first=%d last=%d", page[0].PublishTS, page[9].PublishTS)
	}

	// 第二页
	total, page, err = GetDynamicSpace(host, 10, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 25 || len(page) != 10 || page[0].PublishTS != 150 {
		t.Fatalf("page2 wrong: total=%d len=%d first=%d", total, len(page), page[0].PublishTS)
	}

	// limit 钳制到 20（0 / 负 / >200）
	for _, bad := range []int{0, -7, 201, 5000} {
		_, p, err := GetDynamicSpace(host, bad, 0)
		if err != nil {
			t.Fatalf("limit %d: %v", bad, err)
		}
		if len(p) != 20 {
			t.Fatalf("limit %d should clamp to 20, got %d", bad, len(p))
		}
	}
	// 合法 200 不钳制
	_, p200, err := GetDynamicSpace(host, 200, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(p200) != 25 {
		t.Fatalf("limit 200 should return all 25, got %d", len(p200))
	}
	// 负 offset 归 0
	_, pOff, err := GetDynamicSpace(host, 5, -3)
	if err != nil {
		t.Fatal(err)
	}
	if len(pOff) != 5 || pOff[0].PublishTS != 250 {
		t.Fatalf("negative offset should act as 0, got %+v", pOff)
	}
	// 空 host
	total, page, err = GetDynamicSpace("dyn_host_empty_x", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(page) != 0 {
		t.Fatalf("empty host: total=%d len=%d", total, len(page))
	}

	// GetDynamicCount 精确
	cnt, err := GetDynamicCount(host)
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 25 {
		t.Fatalf("GetDynamicCount=%d want 25", cnt)
	}
	cnt, err = GetDynamicCount("dyn_host_none_y")
	if err != nil || cnt != 0 {
		t.Fatalf("count unknown host: %d %v", cnt, err)
	}
}

// TestDynDeleteDynamicSpace 整站删除且不影响他人。
func TestDynDeleteDynamicSpace(t *testing.T) {
	dynReset(t)
	delHost := "dyn_host_wipe"
	keepHost := "dyn_host_keep"
	if _, err := SaveDynamics(delHost, []DynamicItem{
		dynItem("dynid_w1", "DYNAMIC_TYPE_AV", "f", "u", 1),
		dynItem("dynid_w2", "DYNAMIC_TYPE_NATIVE", "f", "u", 2),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveDynamics(keepHost, []DynamicItem{dynItem("dynid_k1", "DYNAMIC_TYPE_AV", "f", "u", 3)}); err != nil {
		t.Fatal(err)
	}

	if err := DeleteDynamicSpace(delHost); err != nil {
		t.Fatalf("DeleteDynamicSpace: %v", err)
	}
	var n int
	if err := dynDB(t).QueryRow("SELECT COUNT(*) FROM dynamics WHERE host_mid = ?", delHost).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("dynamics of wiped host remain: %d", n)
	}
	if err := dynDB(t).QueryRow("SELECT COUNT(*) FROM dynamic_hosts WHERE host_mid = ?", delHost).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("host row of wiped host remains")
	}
	// 保留 host 不受影响
	if !IsDynamicExists("dynid_k1") {
		t.Fatal("other host's dynamic was wiped!")
	}
	// 删除不存在 host：无错
	if err := DeleteDynamicSpace("dyn_host_never_existed"); err != nil {
		t.Fatalf("delete unknown host should be nil: %v", err)
	}
}

// TestDynMediaLocalsJSONRoundTrip 直接验证写库的 JSON 形状（media_locals 默认值语义）。
func TestDynMediaLocalsJSONRoundTrip(t *testing.T) {
	dynReset(t)
	it := dynItem("dynid_j1", "DYNAMIC_TYPE_FORWARD", "f", "u", 5)
	it.MediaLocals = []string{"x/1.jpg"}
	it.LiveMediaLocals = []string{"live/1.mp4"}
	if _, err := SaveDynamics("dyn_host_json", []DynamicItem{it}); err != nil {
		t.Fatal(err)
	}
	var mediaRaw, liveRaw string
	if err := dynDB(t).QueryRow("SELECT media_locals, live_media_locals FROM dynamics WHERE id_str = ?", "dynid_j1").
		Scan(&mediaRaw, &liveRaw); err != nil {
		t.Fatal(err)
	}
	var m, l []string
	if err := json.Unmarshal([]byte(mediaRaw), &m); err != nil {
		t.Fatalf("stored media_locals not valid json: %q", mediaRaw)
	}
	if err := json.Unmarshal([]byte(liveRaw), &l); err != nil {
		t.Fatalf("stored live_media_locals not valid json: %q", liveRaw)
	}
	if len(m) != 1 || m[0] != "x/1.jpg" || len(l) != 1 || l[0] != "live/1.mp4" {
		t.Fatalf("json roundtrip wrong: %v / %v", m, l)
	}
	// nil slice 序列化为 "null"，解码回 nil 而非报错
	it2 := dynItem("dynid_j2", "DYNAMIC_TYPE_FORWARD", "f", "u", 6)
	it2.MediaLocals = nil
	if _, err := SaveDynamics("dyn_host_json", []DynamicItem{it2}); err != nil {
		t.Fatal(err)
	}
	got, err := GetDynamicByID("dynid_j2")
	if err != nil {
		t.Fatal(err)
	}
	if got.MediaLocals != nil {
		t.Fatalf("nil slice should decode back to nil, got %v", got.MediaLocals)
	}
}

// TestDynHostsAfterDeleteSpace 清场：确认 reset 后 host 列表为空。
func TestDynHostsAfterDeleteSpace(t *testing.T) {
	dynReset(t)
	hosts, err := GetDynamicHosts(50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 0 {
		t.Fatalf("reset should leave zero hosts: %v", hosts)
	}
}
