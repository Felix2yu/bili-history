package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/internal/database"
)

// yearClient is this file's private history namespace, so the archive and
// summary queries have a deterministic scope. The 继续观看 fixtures also live
// here but carry 2026 view_at values: that query reads every year table and
// ranks by view_at, so they must sit above the other suites' older rows. The
// archive test scopes itself with a keyword instead of a second table, because
// adding a newer table would change what "the latest year" means for every
// other suite.
const yearClient = 2009

func clientBvid(suffix string) string { return "BVclient" + suffix }

// insertClientRow seeds one archive row. bvid must be unique per
// (bvid, view_at) pair because insertHistory is idempotent on that key.
func insertClientRow(t *testing.T, bvid string, at int64, progress, duration int, business string) {
	t.Helper()
	insertHistory(t, yearClient, map[string]interface{}{
		"bvid":        bvid,
		"title":       "客户端样本 " + bvid,
		"author_name": "样本UP",
		"author_mid":  int64(900001),
		"view_at":     at,
		"progress":    progress,
		"duration":    duration,
		"business":    business,
	})
}

func TestClientRecentSkipsFinishedAndDedupes(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterClientRoutes)

	newest := time.Date(2026, time.November, 20, 21, 0, 0, 0, time.Local).Unix()
	older := newest - 3600
	oldest := newest - 7200
	// Two sessions of the same video: the queue must offer it once, at the
	// latest position, not twice at two positions.
	insertClientRow(t, clientBvid("A"), newest, 120, 600, "video")
	insertClientRow(t, clientBvid("A"), oldest, 60, 600, "video")
	// Finished: progress equal to the duration.
	insertClientRow(t, clientBvid("B"), older, 600, 600, "video")
	// -1 is Bilibili's "watched to the end" marker.
	insertClientRow(t, clientBvid("C"), older, -1, 600, "video")
	// A live room has no playable position and must never enter the queue.
	insertClientRow(t, clientBvid("L"), older, 30, 0, "live")

	resp := expectStatus(t, e, "POST", "/api/client/recent", `{"limit":50}`, 200, "success")
	records := dataArray(t, resp.dataMap(t)["records"])

	var found map[string]interface{}
	for _, raw := range records {
		item, ok := raw.(map[string]interface{})
		if !ok {
			t.Fatalf("record = %T, want object", raw)
		}
		if item["bvid"] == clientBvid("A") {
			if found != nil {
				t.Fatalf("继续观看 listed the same video twice: %s", resp.Data)
			}
			found = item
		}
		if item["bvid"] == clientBvid("B") || item["bvid"] == clientBvid("C") || item["bvid"] == clientBvid("L") {
			t.Fatalf("finished or live video %v entered the queue", item["bvid"])
		}
	}
	if found == nil {
		t.Fatalf("unfinished video missing from 继续观看: %s", resp.Data)
	}
	if got := int(found["progress"].(float64)); got != 120 {
		t.Fatalf("progress = %d, want the newest session's 120", got)
	}
	if found["finished"] != false {
		t.Fatalf("finished = %v, want false", found["finished"])
	}
}

func TestClientHistoryCursorPages(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterClientRoutes)

	base := time.Date(2009, time.May, 12, 20, 0, 0, 0, time.Local).Unix()
	insertClientRow(t, clientBvid("H1"), base, 10, 600, "video")
	insertClientRow(t, clientBvid("H2"), base-100, 20, 600, "video")
	insertClientRow(t, clientBvid("H3"), base-200, 30, 600, "video")

	// The keyword scopes the page to this suite's rows: the table is shared with
	// the 继续观看 fixtures above (2026 timestamps) and with other suites.
	first := expectStatus(t, e, "POST", "/api/client/history", `{"limit":2,"year":2009,"keyword":"BVclientH"}`, 200, "success")
	firstData := first.dataMap(t)
	firstRecords := dataArray(t, firstData["records"])
	if len(firstRecords) != 2 {
		t.Fatalf("first page = %d records, want 2: %s", len(firstRecords), first.Data)
	}
	firstTop := firstRecords[0].(map[string]interface{})
	next := int64(firstData["next"].(float64))
	if next == 0 {
		t.Fatal("next = 0 on a full page; the client would stop early")
	}
	if firstTop["bvid"] != clientBvid("H1") {
		t.Fatalf("first record = %v, want the newest bvid", firstTop["bvid"])
	}

	// The cursor is the previous page's last view_at, so the page boundary does
	// not move even if a new row arrives at the top meanwhile.
	body := fmt.Sprintf(`{"limit":2,"year":2009,"keyword":"BVclientH","cursor":%d}`, next)
	second := expectStatus(t, e, "POST", "/api/client/history", body, 200, "success")
	secondData := second.dataMap(t)
	secondRecords := dataArray(t, secondData["records"])
	if len(secondRecords) != 1 {
		t.Fatalf("second page = %d records, want 1: %s", len(secondRecords), second.Data)
	}
	if bvid := secondRecords[0].(map[string]interface{})["bvid"]; bvid != clientBvid("H3") {
		t.Fatalf("second page record = %v, want the oldest bvid", bvid)
	}
	if got := int64(secondData["next"].(float64)); got != 0 {
		t.Fatalf("next = %d on a partial page, want 0", got)
	}
}

func TestClientHistoryKeywordAndValidation(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterClientRoutes)

	at := time.Date(2009, time.June, 1, 9, 0, 0, 0, time.Local).Unix()
	insertClientRow(t, clientBvid("K1"), at, 10, 600, "video")

	resp := expectStatus(t, e, "POST", "/api/client/history", `{"keyword":"BVclientK1"}`, 200, "success")
	if n := len(dataArray(t, resp.dataMap(t)["records"])); n != 1 {
		t.Fatalf("keyword search = %d records, want 1", n)
	}
	resp = expectStatus(t, e, "POST", "/api/client/history", `{"keyword":"没有这个标题"}`, 200, "success")
	if n := len(dataArray(t, resp.dataMap(t)["records"])); n != 0 {
		t.Fatalf("unmatched keyword = %d records, want 0", n)
	}

	expectCode(t, e, "POST", "/api/client/history", `{"cursor":-1}`, 400)
	expectCode(t, e, "POST", "/api/client/history", `{"year":1600}`, 400)
	expectCode(t, e, "POST", "/api/client/history", `{bad`, 400)
}

func TestClientVideoStatusReadsCacheOnly(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterClientRoutes)

	// A private row: MarkVideoDeleted flips status across every year table, and
	// the shared fixture's bvids are asserted "not deleted" by history_test.go.
	gone := clientBvid("D1")
	alive := clientBvid("D2")
	at := time.Date(2009, time.August, 1, 12, 0, 0, 0, time.Local).Unix()
	insertClientRow(t, gone, at, 10, 600, "video")
	insertClientRow(t, alive, at-60, 10, 600, "video")
	if err := database.MarkVideoDeleted(gone); err != nil {
		t.Fatalf("MarkVideoDeleted: %v", err)
	}

	body := fmt.Sprintf(`{"bvids":["%s","%s","%s"]}`, gone, alive, gone)
	resp := expectStatus(t, e, "POST", "/api/client/video-status", body, 200, "success")
	list := dataArray(t, resp.dataMap(t)["deleted"])
	if len(list) != 1 || list[0] != gone {
		t.Fatalf("deleted = %v, want exactly the marked bvid (deduplicated)", list)
	}

	// The batch is capped at the same size the watch-progress channel allows.
	tooMany := make([]string, 501)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf("BVcap%04d", i)
	}
	payload, _ := json.Marshal(map[string]interface{}{"bvids": tooMany})
	expectCode(t, e, "POST", "/api/client/video-status", string(payload), 400)
}

func TestClientSummaryMatchesReportDefinitions(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterClientRoutes)

	// Two videos in the same month: one finished at 23:00 (深夜), one partial at
	// 10:00. duration is the video's own length, not the seconds watched.
	night := time.Date(2009, time.July, 3, 23, 10, 0, 0, time.Local).Unix()
	morning := time.Date(2009, time.July, 4, 10, 0, 0, 0, time.Local).Unix()
	insertClientRow(t, clientBvid("S1"), night, 600, 600, "video")
	insertClientRow(t, clientBvid("S2"), morning, 100, 600, "video")

	resp := expectStatus(t, e, "GET", "/api/client/summary?year=2009&month=7", "", 200, "success")
	data := resp.dataMap(t)
	if got := int(data["videos"].(float64)); got != 2 {
		t.Fatalf("videos = %d, want 2", got)
	}
	if got := int64(data["seconds"].(float64)); got != 1200 {
		t.Fatalf("seconds = %d, want 1200 (sum of durations)", got)
	}
	if got := int(data["finished"].(float64)); got != 1 {
		t.Fatalf("finished = %d, want 1", got)
	}
	if got := data["finished_rate"].(float64); got < 0.49 || got > 0.51 {
		t.Fatalf("finished_rate = %v, want 0.5", got)
	}
	if got := data["late_night_ratio"].(float64); got < 0.49 || got > 0.51 {
		t.Fatalf("late_night_ratio = %v, want 0.5 (22:00–06:00)", got)
	}
	if got := int(data["unique_days"].(float64)); got != 2 {
		t.Fatalf("unique_days = %d, want 2", got)
	}

	// A month with no data is a zero summary, not an error.
	resp = expectStatus(t, e, "GET", "/api/client/summary?year=2009&month=1", "", 200, "success")
	if got := int(resp.dataMap(t)["videos"].(float64)); got != 0 {
		t.Fatalf("empty month videos = %d, want 0", got)
	}
	expectCode(t, e, "GET", "/api/client/summary?year=2009&month=13", "", 400)
	expectCode(t, e, "GET", "/api/client/summary?year=1600", "", 400)
}

func TestClientMirrorsAreReadOnly(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterClientRoutes)

	// Watch later: the archive copy is what survives a deletion on Bilibili.
	wl := database.GetWatchLaterDB()
	if wl == nil {
		t.Skip("watch later database unavailable in this environment")
	}
	if _, err := wl.Exec(`INSERT OR REPLACE INTO watchlater_videos
		(bvid, aid, title, pic, duration, owner_name, add_at, view, fetch_time)
		VALUES (?, 1, '稍后再看样本', 'https://example.invalid/cover.jpg', 300, '样本UP', 9000000000, 42, 1)`,
		clientBvid("WL")); err != nil {
		t.Fatalf("seed watch later: %v", err)
	}
	resp := expectStatus(t, e, "GET", "/api/client/watchlater?page=1&size=100", "", 200, "success")
	records := dataArray(t, resp.dataMap(t)["records"])
	found := false
	for _, raw := range records {
		if raw.(map[string]interface{})["bvid"] == clientBvid("WL") {
			found = true
		}
	}
	if !found {
		t.Fatalf("seeded watch later entry missing: %s", resp.Data)
	}

	// Favorites: a folder and one item inside it.
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable in this environment")
	}
	const mediaID int64 = 20099001
	if _, err := fav.Exec(`INSERT OR REPLACE INTO favorites_folder
		(media_id, fid, mid, title, cover, mtime, media_count, folder_type, fetch_time)
		VALUES (?, 1, 1, '归档收藏夹', '', 9000000000, 1, 0, 1)`, mediaID); err != nil {
		t.Fatalf("seed favorite folder: %v", err)
	}
	if _, err := fav.Exec(`INSERT OR REPLACE INTO favorites_content
		(media_id, content_id, title, cover, bvid, creator_name, upper_mid, duration, fav_time, fetch_time)
		VALUES (?, 1, '收藏样本', '', ?, '样本UP', 900001, 420, 9000000000, 1)`,
		mediaID, clientBvid("F1")); err != nil {
		t.Fatalf("seed favorite content: %v", err)
	}

	resp = expectStatus(t, e, "GET", "/api/client/favorites", "", 200, "success")
	folders := dataArray(t, resp.dataMap(t)["folders"])
	found = false
	for _, raw := range folders {
		folder := raw.(map[string]interface{})
		if int64(folder["media_id"].(float64)) == mediaID {
			found = true
		}
	}
	if !found {
		t.Fatalf("seeded favorite folder missing: %s", resp.Data)
	}

	resp = expectStatus(t, e, "GET", fmt.Sprintf("/api/client/favorites/contents?media_id=%d", mediaID), "", 200, "success")
	contents := dataArray(t, resp.dataMap(t)["records"])
	if len(contents) != 1 || contents[0].(map[string]interface{})["bvid"] != clientBvid("F1") {
		t.Fatalf("favorite contents = %s, want the seeded bvid", resp.Data)
	}

	// media_id is required.
	expectCode(t, e, "GET", "/api/client/favorites/contents", "", 400)

	// The channel is read-only: no mutating verb is registered.
	for _, route := range fullAPIRoutes() {
		if strings.HasPrefix(route, "DELETE /api/client") || strings.HasPrefix(route, "PUT /api/client") {
			t.Fatalf("client channel exposes a mutating route: %s", route)
		}
	}
}

func TestClientTokenGuard(t *testing.T) {
	newHistoryFixture(t)
	e := newAPI(t, RegisterClientRoutes)

	cfg := loadCfg(t)
	previous := cfg.APIToken
	cfg.APIToken = "client-secret"
	defer func() { cfg.APIToken = previous }()

	expectCode(t, e, "POST", "/api/client/recent", `{"limit":1}`, 401)
	expectCode(t, e, "GET", "/api/client/summary", "", 401)

	// Both header forms are accepted, matching /watch-progress.
	resp, err := doWithHeader(t, e, "POST", "/api/client/recent", `{"limit":1}`, "Authorization", "Bearer client-secret")
	if err != nil {
		t.Fatalf("authorized request: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("authorized status = %q, message = %q", resp.Status, resp.Message)
	}
	resp, err = doWithHeader(t, e, "GET", "/api/client/summary", "", "X-API-Token", "client-secret")
	if err != nil {
		t.Fatalf("authorized request: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("authorized status = %q, message = %q", resp.Status, resp.Message)
	}
}

// doWithHeader is `do` plus one extra request header, for the token guard tests.
func doWithHeader(t *testing.T, h http.Handler, method, path, body, header, value string) (apiResp, error) {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set(header, value)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	var resp apiResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
