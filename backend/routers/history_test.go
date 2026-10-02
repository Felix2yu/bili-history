package routers

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/database"
)

// historyFixture seeds rows into the reserved test years and returns the bvids
// it created so each assertion can scope itself to this file's data.
type historyFixture struct {
	bvids map[string]string // logical name -> bvid
}

func newHistoryFixture(t *testing.T) *historyFixture {
	t.Helper()
	f := &historyFixture{bvids: map[string]string{}}

	add := func(year int, name string, extra map[string]interface{}) {
		bvid := uniqueBvid("BVHIST", len(f.bvids)+1)
		f.bvids[name] = bvid
		row := map[string]interface{}{"bvid": bvid, "id": nil}
		for k, v := range extra {
			row[k] = v
		}
		insertHistory(t, year, row)
	}

	v2016 := viewAt(t, 2016, time.March, 10, 12, 0)
	v2016b := viewAt(t, 2016, time.March, 10, 18, 30)
	v2006 := viewAt(t, 2006, time.January, 13, 9, 0)

	add(yearPrimary, "plain", map[string]interface{}{
		"title": "测试历史记录标题A", "tag_name": "科技", "main_category": "科技数码",
		"view_at": v2016, "duration": 300, "progress": 120, "cid": 900001,
		"author_name": "UP主甲", "author_mid": 200001, "cover": "http://i0.hdslb.com/bfs/cover/a.jpg",
	})
	add(yearPrimary, "remarked", map[string]interface{}{
		"title": "测试历史记录标题B", "tag_name": "科技", "main_category": "科技数码",
		"view_at": v2016b, "duration": 600, "progress": -1, "cid": 900002,
		"author_name": "UP主乙", "author_mid": 200002,
		"remark": "这条有备注", "remark_time": v2016b,
	})
	// business = live is filtered out unless explicitly requested.
	add(yearPrimary, "live", map[string]interface{}{
		"title": "测试直播记录", "business": "live", "view_at": v2016, "duration": 60,
		"cid": 900003, "author_name": "UP主丙", "author_mid": 200003,
	})
	add(yearSecondary, "old", map[string]interface{}{
		"title": "测试跨年记录", "tag_name": "音乐", "view_at": v2006, "duration": 240,
		"progress": 240, "cid": 900004, "author_name": "UP主甲", "author_mid": 200001,
	})
	return f
}

func TestHistoryAvailableYearsAndDates(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterHistoryRoutes)

	resp := expectStatus(t, e, "GET", "/api/history/available-years", "", 200, "success")
	years := resp.dataArray(t)
	if len(years) == 0 {
		t.Fatal("available-years returned no years although tables were seeded")
	}
	var found []string
	for _, y := range years {
		found = append(found, fmt.Sprint(y))
	}
	joined := strings.Join(found, ",")
	for _, want := range []string{"2016", "2006"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("available-years = %s, want it to contain %s", joined, want)
		}
	}
	// GetAvailableYears orders by table name DESC, so "bilibili_history_2016"
	// sorts ahead of "bilibili_history_2006".
	if fmt.Sprint(years[0]) != "2016" {
		t.Fatalf("first year = %v, want 2016 (name DESC ordering)", years[0])
	}

	dates := expectStatus(t, e, "GET", "/api/history/dates", "", 200, "success").dataArray(t)
	var dateStrings []string
	for _, d := range dates {
		dateStrings = append(dateStrings, fmt.Sprint(d))
	}
	if !contains(dateStrings, "20160310") || !contains(dateStrings, "20060113") {
		t.Fatalf("dates = %v, want 20160310 and 20060113", dateStrings)
	}
	// The dates are rendered in descending order.
	if dateStrings[0] < dateStrings[len(dateStrings)-1] {
		t.Fatalf("dates are not descending: %v", dateStrings)
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestHistoryPageFilters(t *testing.T) {
	f := newHistoryFixture(t)
	e := newAPI(t, RegisterHistoryRoutes)

	// Default listing excludes live records but spans every year table, so the
	// two seeded non-live rows of this fixture must show up somewhere.
	resp := expectStatus(t, e, "GET", "/api/history/all?page=1&size=50", "", 200, "success")
	data := resp.dataMap(t)
	records := data["records"].([]interface{})
	total := int(data["total"].(float64))
	if total < 3 {
		t.Fatalf("total = %d, want at least the 3 non-live fixture rows", total)
	}
	if fmt.Sprint(data["current"]) != "1" || fmt.Sprint(data["size"]) != "50" {
		t.Fatalf("paging echo wrong: current=%v size=%v", data["current"], data["size"])
	}
	if _, ok := data["available_years"]; !ok {
		t.Fatal("available_years missing from /history/all payload")
	}
	got := recordByBvid(t, records, f.bvids["plain"])
	if got["title"] != "测试历史记录标题A" {
		t.Fatalf("title = %v", got["title"])
	}
	// processRecord mirrors tag_name into tname and formats view_at.
	if got["tname"] != "科技" {
		t.Fatalf("tname = %v, want 科技", got["tname"])
	}
	if got["view_time"] == nil || got["view_time"] == "" {
		t.Fatalf("view_time not formatted: %v", got["view_time"])
	}
	if got["original_url"] == nil || !strings.Contains(fmt.Sprint(got["original_url"]), f.bvids["plain"]) {
		t.Fatalf("original_url = %v, want it to embed the bvid", got["original_url"])
	}
	if covers, ok := got["covers"].([]interface{}); !ok || len(covers) != 0 {
		t.Fatalf("covers = %v, want an empty array when the column is null", got["covers"])
	}

	// The live record is only visible when business is requested explicitly.
	liveResp := expectStatus(t, e, "GET", "/api/history/all?business=live&size=50", "", 200, "success")
	liveRecords := liveResp.dataMap(t)["records"].([]interface{})
	if _, ok := findRecord(t, liveRecords, f.bvids["live"]); !ok {
		t.Fatalf("business=live query did not return the live row: %v", liveRecords)
	}
	for _, r := range records {
		m := r.(map[string]interface{})
		if m["bvid"] == f.bvids["live"] {
			t.Fatal("default listing must exclude business=live")
		}
	}

	// tag_name filter.
	tagResp := expectStatus(t, e, "GET", "/api/history/all?tag_name="+f.bvids["plain"][:0]+"音乐", "", 200, "success")
	if _, ok := findRecord(t, tagResp.dataMap(t)["records"].([]interface{}), f.bvids["old"]); !ok {
		t.Fatal("tag_name=音乐 must return the cross-year row")
	}

	// main_category takes precedence over tag_name when both are supplied.
	catResp := expectStatus(t, e, "GET", "/api/history/all?main_category=科技数码&tag_name=音乐", "", 200, "success")
	catRecords := catResp.dataMap(t)["records"].([]interface{})
	if _, ok := findRecord(t, catRecords, f.bvids["plain"]); !ok {
		t.Fatal("main_category filter must win over tag_name")
	}
	if _, ok := findRecord(t, catRecords, f.bvids["old"]); ok {
		t.Fatal("main_category branch ignored the tag_name filter entirely")
	}

	// date_range restricts to a single day (parsed as %Y%m%d in local time).
	dr := expectStatus(t, e, "GET", "/api/history/all?date_range=20060113-20060113", "", 200, "success")
	drRecords := dr.dataMap(t)["records"].([]interface{})
	if _, ok := findRecord(t, drRecords, f.bvids["old"]); !ok {
		t.Fatalf("date_range 20060113 must return the January row: %v", drRecords)
	}
	if _, ok := findRecord(t, drRecords, f.bvids["plain"]); ok {
		t.Fatal("date_range leaked rows outside the requested day")
	}

	// An unparsable date_range is silently ignored (PRODUCTION behaviour: the
	// parse error is dropped, so the filter simply disappears).
	badRange := expectStatus(t, e, "GET", "/api/history/all?date_range=not-a-date", "", 200, "success")
	if _, ok := findRecord(t, badRange.dataMap(t)["records"].([]interface{}), f.bvids["plain"]); !ok {
		t.Fatal("invalid date_range should be ignored rather than filtering everything out")
	}

	// sort_order=1 flips to ascending: the first row is the oldest one in the
	// database. Which year that is depends on which other suites have seeded,
	// so compare against the live minimum instead of a fixture timestamp.
	asc := expectStatus(t, e, "GET", "/api/history/all?sort_order=1&size=500", "", 200, "success")
	ascRecords := asc.dataMap(t)["records"].([]interface{})
	if len(ascRecords) == 0 {
		t.Fatal("ascending query returned nothing")
	}
	// view_at arrives as a JSON number, so compare numerically.
	firstTS := int64(ascRecords[0].(map[string]interface{})["view_at"].(float64))
	if want := globalMinViewAt(t); firstTS != want {
		t.Fatalf("first ascending record is not the oldest: got %d want %d (%v)", firstTS, want, ascRecords[0])
	}
	lastTS := int64(ascRecords[len(ascRecords)-1].(map[string]interface{})["view_at"].(float64))
	if lastTS < firstTS {
		t.Fatalf("ascending order violated: first %d is newer than last %d", firstTS, lastTS)
	}

	// Page beyond the data yields an empty list, not an error.
	empty := expectStatus(t, e, "GET", "/api/history/all?page=9999&size=10", "", 200, "success")
	if recs := empty.dataMap(t)["records"].([]interface{}); len(recs) != 0 {
		t.Fatalf("page 9999 returned %d records, want 0", len(recs))
	}

	// Non-numeric page/size degrade to 0 through strconv errors, which the
	// handler ignores; LIMIT 0 then yields no records instead of an error.
	weird := expectStatus(t, e, "GET", "/api/history/all?page=abc&size=xyz", "", 200, "success")
	if _, ok := findRecord(t, weird.dataMap(t)["records"].([]interface{}), f.bvids["plain"]); ok {
		t.Fatal("page=abc must not return a normal page (LIMIT 0 behaviour documented here)")
	}
}

func TestHistorySearch(t *testing.T) {
	f := newHistoryFixture(t)
	e := newAPI(t, RegisterHistoryRoutes)

	base := "/api/history/search?search=" + urlEncode("测试历史记录")
	resp := expectStatus(t, e, "GET", base, "", 200, "success")
	data := resp.dataMap(t)
	if _, ok := findRecord(t, data["records"].([]interface{}), f.bvids["plain"]); !ok {
		t.Fatalf("title keyword search missed the row: %v", data["records"])
	}
	info := data["search_info"].(map[string]interface{})
	if info["keyword"] != urlEncode("测试历史记录") {
		// gin decodes the query, so the echoed keyword is the decoded value.
		if info["keyword"] != "测试历史记录" {
			t.Fatalf("search_info.keyword = %v", info["keyword"])
		}
	}
	if info["type"] != "all" || info["sort_by"] != "view_at" || info["exact_match"] != false {
		t.Fatalf("unexpected search_info: %v", info)
	}

	// search_type=author matches against author_name only.
	author := expectStatus(t, e, "GET", "/api/history/search?search="+urlEncode("UP主乙")+"&search_type=author", "", 200, "success")
	recs := author.dataMap(t)["records"].([]interface{})
	if _, ok := findRecord(t, recs, f.bvids["remarked"]); !ok {
		t.Fatalf("author search missed the row: %v", recs)
	}
	if _, ok := findRecord(t, recs, f.bvids["plain"]); ok {
		t.Fatal("author search must not match on title")
	}

	// An unknown search_type produces no field conditions at all, which makes
	// the query fall back to "everything except live/article" (PRODUCTION
	// behaviour: the type is validated by map lookup, and an unknown key is
	// silently treated as an unfiltered listing).
	unknown := expectStatus(t, e, "GET", "/api/history/search?search="+urlEncode("测试历史记录")+"&search_type=bogus", "", 200, "success")
	if _, ok := findRecord(t, unknown.dataMap(t)["records"].([]interface{}), f.bvids["remarked"]); !ok {
		t.Fatal("search_type=bogus is expected to ignore the keyword entirely")
	}

	// search_type=remark filters on the remark column.
	remark := expectStatus(t, e, "GET", "/api/history/search?search="+urlEncode("这条有备注")+"&search_type=remark", "", 200, "success")
	if _, ok := findRecord(t, remark.dataMap(t)["records"].([]interface{}), f.bvids["remarked"]); !ok {
		t.Fatal("remark search missed the remarked row")
	}

	// Empty keyword returns the default listing.
	none := expectStatus(t, e, "GET", "/api/history/search?search=", "", 200, "success")
	if int(none.dataMap(t)["total"].(float64)) < 3 {
		t.Fatal("empty keyword must return the unfiltered page")
	}
}

func TestHistorySQLiteVersion(t *testing.T) {
	e := newAPI(t, RegisterHistoryRoutes)
	resp := expectStatus(t, e, "GET", "/api/history/sqlite-version", "", 200, "success")
	data := resp.dataMap(t)
	if data == nil {
		t.Fatal("sqlite-version returned no object")
	}
	// The payload always reports the linked sqlite version; assert a plausible
	// value instead of pinning the driver build.
	if s, ok := data["sqlite_version"].(string); !ok || !strings.Contains(s, ".") {
		t.Fatalf("sqlite_version = %v (%T), want a dotted version string", data["sqlite_version"], data["sqlite_version"])
	}
}

func TestHistoryByCID(t *testing.T) {
	f := newHistoryFixture(t)
	_ = f
	e := newAPI(t, RegisterHistoryRoutes)

	resp := expectStatus(t, e, "GET", "/api/history/by_cid/900001", "", 200, "success")
	data := resp.dataMap(t)
	if data["bvid"] != f.bvids["plain"] {
		t.Fatalf("by_cid resolved to %v, want %v", data["bvid"], f.bvids["plain"])
	}
	if data["original_url"] == nil {
		t.Fatal("by_cid payload must run through processRecord")
	}

	// use_local_images / use_sessdata are accepted but processImageURL returns
	// the raw URL either way; the query only has to not fail.
	expectStatus(t, e, "GET", "/api/history/by_cid/900001?use_local_images=true&use_sessdata=false", "", 200, "success")

	expectStatus(t, e, "GET", "/api/history/by_cid/notanumber", "", 400, "error")
	missing := expectStatus(t, e, "GET", "/api/history/by_cid/987654321", "", 200, "error")
	expectMessageContains(t, missing, "未找到CID")
}

func TestHistoryBatchRemarks(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterHistoryRoutes)

	// Malformed JSON body.
	expectStatus(t, e, "POST", "/api/history/batch-remarks", "not-json", 400, "error")

	resp := expectStatus(t, e, "POST", "/api/history/batch-remarks",
		`{"items":[{"bvid":"BVHIST000000002","view_at":1457581200}]}`, 200, "success")
	data := resp.dataMap(t)
	entry, ok := data["BVHIST000000002"].(map[string]interface{})
	if !ok {
		t.Fatalf("remark for the remarked row missing: %v", data)
	}
	if entry["remark"] != "这条有备注" {
		t.Fatalf("remark = %v", entry["remark"])
	}

	// Empty items list yields an empty map (no query is issued).
	empty := expectStatus(t, e, "POST", "/api/history/batch-remarks", `{"items":[]}`, 200, "success")
	if m := empty.dataMap(t); len(m) != 0 {
		t.Fatalf("empty items should return an empty object, got %v", m)
	}

	// Blank bvids are skipped.
	blank := expectStatus(t, e, "POST", "/api/history/batch-remarks", `{"items":[{"bvid":"","view_at":1}]}`, 200, "success")
	if m := blank.dataMap(t); len(m) != 0 {
		t.Fatalf("blank bvid should be filtered out, got %v", m)
	}
}

func TestHistoryCheckDeletedGuards(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterHistoryRoutes)

	// No body at all.
	expectCode(t, e, "POST", "/api/history/check-deleted", "", 400)
	// Empty list.
	expectCode(t, e, "POST", "/api/history/check-deleted", `{"bvids":[]}`, 400)
	// Broken JSON.
	expectCode(t, e, "POST", "/api/history/check-deleted", `{"bvids":`, 400)

	// SESSDATA is empty in the test config, so the handler bails out before it
	// can reach api.bilibili.com. This is the guard that keeps the suite offline.
	restore := setSESSDATA(t, "")
	w := doRaw(t, e, "POST", "/api/history/check-deleted", `{"bvids":["BVHIST000000001"]}`)
	restore()
	if w.Code != 200 {
		t.Fatalf("check-deleted without SESSDATA: code = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "SESSDATA 未配置") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestHistoryDeletedStatus(t *testing.T) {
	f := newHistoryFixture(t)
	e := newAPI(t, RegisterHistoryRoutes)

	expectCode(t, e, "POST", "/api/history/deleted-status", `{bad`, 400)

	// Nothing is flagged yet. This route answers the {success,data} shape rather
	// than the models.Response envelope, so the raw body is what to assert on.
	body := `{"bvids":["` + f.bvids["plain"] + `"]}`
	w1 := doRaw(t, e, "POST", "/api/history/deleted-status", body)
	if w1.Code != 200 || !strings.Contains(w1.Body.String(), `"success":true`) {
		t.Fatalf("deleted-status = code %d, body %s", w1.Code, w1.Body.String())
	}
	if strings.Contains(w1.Body.String(), f.bvids["plain"]) {
		t.Fatalf("unmarked video reported as deleted: %s", w1.Body.String())
	}

	// Marking a video propagates status=1 across every year table.
	if err := database.MarkVideoDeleted(f.bvids["plain"]); err != nil {
		t.Fatalf("MarkVideoDeleted: %v", err)
	}
	w2 := doRaw(t, e, "POST", "/api/history/deleted-status", `{"bvids":["`+f.bvids["plain"]+`","`+f.bvids["live"]+`"]}`)
	if !strings.Contains(w2.Body.String(), f.bvids["plain"]) {
		t.Fatalf("marked video not reported as deleted: %s", w2.Body.String())
	}

	// An empty bvid list short-circuits to an empty map.
	w3 := doRaw(t, e, "POST", "/api/history/deleted-status", `{"bvids":[]}`)
	if !strings.Contains(w3.Body.String(), `"data":{}`) {
		t.Fatalf("empty bvid list body = %s, want an empty data object", w3.Body.String())
	}
}

func TestDailyCount(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterHistoryRoutes)

	// 2016-03-10 holds three rows and the query does not filter by business:
	// progress 120 + (progress == -1 -> duration 600) + live progress 0 = 720
	// watched seconds.
	resp := expectStatus(t, e, "GET", "/api/daily/daily-count?date=0310&year=2016", "", 200, "success")
	data := resp.dataMap(t)
	if got := int(data["total_count"].(float64)); got != 3 {
		t.Fatalf("total_count = %d, want 3", got)
	}
	if got := int(data["total_watch_seconds"].(float64)); got != 720 {
		t.Fatalf("total_watch_seconds = %d, want 720", got)
	}
	// These two fields are hard-coded to zero by the handler.
	if int(data["unique_authors"].(float64)) != 0 || int(data["total_duration"].(float64)) != 0 {
		t.Fatalf("handler documents these as constants: %v", data)
	}

	// A date with no rows makes SUM() return NULL, the scan fails and the
	// handler answers with an all-zero success payload.
	zero := expectStatus(t, e, "GET", "/api/daily/daily-count?date=1231&year=2016", "", 200, "success")
	zd := zero.dataMap(t)
	if int(zd["total_count"].(float64)) != 0 {
		t.Fatalf("empty day should be zero, got %v", zd)
	}

	// An unknown year is guarded by TableExists, so the injected table name can
	// never address anything that exists: the query short-circuits to zeros.
	// This assertion pins the guard for the client-controlled `year` parameter.
	unknown := expectStatus(t, e, "GET", "/api/daily/daily-count?date=0310&year=1999", "", 200, "success")
	if int(unknown.dataMap(t)["total_count"].(float64)) != 0 {
		t.Fatalf("missing year table must yield zeros: %v", unknown.dataMap(t))
	}
	injected := expectStatus(t, e, "GET", "/api/daily/daily-count?date=0310&year="+urlEncode("2016 WHERE 1=1 UNION SELECT COUNT(*),0 FROM sqlite_master--"), "", 200, "success")
	if s := injected.dataMap(t)["total_count"]; fmt.Sprint(s) != "0" {
		t.Fatalf("injected year must not read anything: %v", injected.dataMap(t))
	}

	// No year at all falls back to the newest available table (2016).
	fallback := expectStatus(t, e, "GET", "/api/daily/daily-count?date=0310", "", 200, "success")
	if int(fallback.dataMap(t)["total_count"].(float64)) != 3 {
		t.Fatalf("year-less query should use the latest table: %v", fallback.dataMap(t))
	}
}

func recordByBvid(t *testing.T, records []interface{}, bvid string) map[string]interface{} {
	t.Helper()
	r, ok := findRecord(t, records, bvid)
	if !ok {
		t.Fatalf("record %s not found among %d records", bvid, len(records))
	}
	return r
}

func findRecord(t *testing.T, records []interface{}, bvid string) (map[string]interface{}, bool) {
	t.Helper()
	for _, r := range records {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		if m["bvid"] == bvid {
			return m, true
		}
	}
	return nil, false
}

func urlEncode(s string) string {
	return strings.NewReplacer(" ", "%20").Replace(s)
}
