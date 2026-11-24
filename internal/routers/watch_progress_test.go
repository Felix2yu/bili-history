package routers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/internal/config"
)

func TestWatchProgressQuery(t *testing.T) {
	partial := uniqueBvid("BVWP", 1)
	finished := uniqueBvid("BVWP", 2)
	live := uniqueBvid("BVWP", 3)
	insertHistory(t, yearPrimary, map[string]interface{}{
		"bvid": partial, "title": "看了一半", "view_at": viewAt(t, 2016, 3, 10, 12, 0),
		"progress": 120, "duration": 300, "cid": 910001,
	})
	insertHistory(t, yearSecondary, map[string]interface{}{
		"bvid": finished, "title": "看完了", "view_at": viewAt(t, 2006, 1, 13, 9, 0),
		"progress": -1, "duration": 240, "cid": 910002,
	})
	insertHistory(t, yearPrimary, map[string]interface{}{
		"bvid": live, "title": "直播不算", "business": "live",
		"view_at": viewAt(t, 2016, 3, 10, 18, 0), "progress": 60, "cid": 910003,
	})

	e := newAPI(t, RegisterWatchProgressRoutes)
	body := `{"bvids":["` + partial + `","` + finished + `","` + live + `","BVmissing","` + partial + `"]}`
	resp := expectStatus(t, e, "POST", "/api/watch-progress/query", body, http.StatusOK, "success")
	data := resp.dataMap(t)

	got := map[string]map[string]interface{}{}
	for key, value := range data {
		entry, ok := value.(map[string]interface{})
		if !ok {
			t.Fatalf("%s = %T, want object", key, value)
		}
		got[key] = entry
	}
	if _, exists := got[live]; exists {
		t.Fatal("live record must not be reported as watched")
	}
	if _, exists := got["BVmissing"]; exists {
		t.Fatal("missing bvid must be absent instead of a zero entry")
	}
	p := got[partial]
	if p == nil || p["watched"] != true || p["finished"] == true {
		t.Fatalf("partial progress = %v, want watched but not finished", p)
	}
	if p["progress"].(float64) != 120 || p["duration"].(float64) != 300 {
		t.Fatalf("partial = %v, want progress 120 duration 300", p)
	}
	f := got[finished]
	if f == nil || f["watched"] != true || f["finished"] != true || f["progress"].(float64) != 240 {
		t.Fatalf("finished = %v, want finished with normalized progress 240", f)
	}
}

func TestWatchProgressReportAndReadBack(t *testing.T) {
	bvid := uniqueBvid("BVWR", 7)
	e := newAPI(t, RegisterWatchProgressRoutes)
	report := `{"bvid":"` + bvid + `","aid":12345,"cid":34567,"title":"同步测试","author":"UP","duration":500,"progress":90}`
	expectStatus(t, e, "POST", "/api/watch-progress/report", report, http.StatusOK, "success")

	resp := expectStatus(t, e, "POST", "/api/watch-progress/query", `{"bvids":["`+bvid+`"]}`, http.StatusOK, "success")
	entry := resp.dataMap(t)[bvid].(map[string]interface{})
	if entry["progress"].(float64) != 90 || entry["duration"].(float64) != 500 || entry["watched"] != true {
		t.Fatalf("reported progress not read back: %v", entry)
	}

	// A later position inside the same 24 h session updates that row in place.
	report2 := `{"bvid":"` + bvid + `","aid":12345,"cid":34567,"title":"同步测试","author":"UP","duration":500,"progress":200}`
	expectStatus(t, e, "POST", "/api/watch-progress/report", report2, http.StatusOK, "success")
	resp = expectStatus(t, e, "POST", "/api/watch-progress/query", `{"bvids":["`+bvid+`"]}`, http.StatusOK, "success")
	entry = resp.dataMap(t)[bvid].(map[string]interface{})
	if entry["progress"].(float64) != 200 {
		t.Fatalf("progress after second report = %v, want 200", entry)
	}
}

// AMac whose clock runs behind must still be able to advance the progress; only
// the view_at timestamp itself is protected from being rewound.
func TestWatchProgressReportWithBehindClock(t *testing.T) {
	bvid := uniqueBvid("BVWT", 11)
	future := time.Now().Add(2 * time.Hour).Unix()
	insertHistory(t, time.Now().Year(), map[string]interface{}{
		"bvid": bvid, "view_at": future, "progress": 10, "duration": 300, "cid": 910011,
	})
	e := newAPI(t, RegisterWatchProgressRoutes)
	report := `{"bvid":"` + bvid + `","aid":11,"cid":910011,"title":"时钟偏差","duration":300,"progress":200}`
	expectStatus(t, e, "POST", "/api/watch-progress/report", report, http.StatusOK, "success")

	resp := expectStatus(t, e, "POST", "/api/watch-progress/query", `{"bvids":["`+bvid+`"]}`, http.StatusOK, "success")
	entry := resp.dataMap(t)[bvid].(map[string]interface{})
	if entry["progress"].(float64) != 200 {
		t.Fatalf("progress = %v, want 200 even when the reporting clock is behind", entry)
	}
	if entry["view_at"].(float64) != float64(future) {
		t.Fatalf("view_at = %v, want the original newer timestamp %d", entry, future)
	}
}

func TestWatchProgressToken(t *testing.T) {
	cfg := config.GetConfig()
	if cfg == nil {
		t.Fatal("config unavailable")
	}
	previous := cfg.APIToken
	cfg.APIToken = "secret-token"
	t.Cleanup(func() { cfg.APIToken = previous })

	e := newAPI(t, RegisterWatchProgressRoutes)
	if code, _ := do(t, e, "POST", "/api/watch-progress/ping", ""); code != http.StatusUnauthorized {
		t.Fatalf("unguarded request = %d, want 401", code)
	}
	for _, header := range []struct{ key, value string }{
		{"Authorization", "Bearer secret-token"},
		{"X-API-Token", "secret-token"},
	} {
		req := httptest.NewRequest("POST", "/api/watch-progress/ping", nil)
		req.Header.Set(header.key, header.value)
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"ok":true`) {
			t.Fatalf("%s auth -> %d %s", header.key, w.Code, w.Body.String())
		}
	}
}
