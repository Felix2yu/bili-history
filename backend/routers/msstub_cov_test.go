package routers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"bilibili-history-go/config"
)

// ---------------------------------------------------------------------------
// Shared helpers for the misc/login/images coverage tests. Everything added
// here carries the ms/MS prefix so it cannot collide with identifiers written
// by the other coverage agents working on this package concurrently.
// ---------------------------------------------------------------------------

// msStubRT is a RoundTripper that intercepts requests aimed at a fixed set of
// hosts (or every non-loopback host when hosts is nil) and answers them from
// an in-process stub. Everything else is delegated to the previous transport,
// so the local httptest servers used by the rest of the suite keep working.
// This exists because login.go and services.DownloadImage hard-code their
// endpoint URLs as literals; the biliapi package vars cannot be repointed for
// them. Nothing ever leaves the machine: requests to real domains are served
// from msHTTPJSON and requests to closed loopback ports stay local.
type msStubRT struct {
	prev  http.RoundTripper
	hosts map[string]bool
	fn    func(*http.Request) (*http.Response, error)
}

func (rt *msStubRT) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Host
	if rt.hosts != nil {
		if !rt.hosts[host] {
			return rt.prev.RoundTrip(req)
		}
	} else if !strings.HasPrefix(host, "127.0.0.1") && !strings.HasPrefix(host, "localhost") && host != "[::1]" {
		// intercept
	} else {
		return rt.prev.RoundTrip(req)
	}
	return rt.fn(req)
}

// msUseBiliTransport intercepts only the hard-coded bilibili hosts used by
func msUseBiliTransport(t *testing.T, fn func(*http.Request) (*http.Response, error)) {
	t.Helper()
	msUseTransportOnHosts(t, map[string]bool{"passport.bilibili.com": true, "api.bilibili.com": true}, fn)
}

// msUseAllTransport intercepts every non-loopback host.
func msUseAllTransport(t *testing.T, fn func(*http.Request) (*http.Response, error)) {
	t.Helper()
	msUseTransportOnHosts(t, nil, fn)
}

func msUseTransportOnHosts(t *testing.T, hosts map[string]bool, fn func(*http.Request) (*http.Response, error)) {
	t.Helper()
	prev := http.DefaultTransport
	http.DefaultTransport = &msStubRT{prev: prev, hosts: hosts, fn: fn}
	t.Cleanup(func() { http.DefaultTransport = prev })
}

// msHTTPJSON builds a synthetic JSON response without touching the network.
func msHTTPJSON(status int, body string) *http.Response {
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.WriteString(body)
	return w.Result()
}

// msHTTPBytes builds a synthetic response with an arbitrary body, used to stub
// the image downloads performed by services.DownloadImage.
func msHTTPBytes(status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"image/png"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

// msHTTPNetworkError makes the calling http.Client.Do fail as if the dial
// itself was refused, exercising the "网络请求失败" branches offline.
func msHTTPNetworkError(msg string) func(*http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		return nil, errors.New(msg)
	}
}

// msCfgSnapshot deep-copies the config singleton and the on-disk config.yaml,
// restoring both when the test finishes. Handlers under test mutate the
// singleton in place and rewrite the file via config.SaveConfig, so every test
// that flips a config field must snapshot first.
func msCfgSnapshot(t *testing.T) {
	t.Helper()
	cfg := loadCfg(t)
	cp := *cfg
	cp.Notify.URLs = append([]string(nil), cfg.Notify.URLs...)
	cp.FieldsToRemove = append([]string(nil), cfg.FieldsToRemove...)
	path := config.GetConfigPathValue()
	var prevFile []byte
	if path != "" {
		prevFile, _ = os.ReadFile(path)
	}
	t.Cleanup(func() {
		cur := loadCfg(t)
		*cur = cp
		cur.Notify.URLs = append([]string(nil), cp.Notify.URLs...)
		cur.FieldsToRemove = append([]string(nil), cp.FieldsToRemove...)
		if path != "" && prevFile != nil {
			_ = os.WriteFile(path, prevFile, 0o644)
		}
	})
}

// msSetCreds overwrites the three session credentials on the singleton for the
// duration of the test (restoring config fields and the yaml file afterwards).
func msSetCreds(t *testing.T, sessdata, biliJct, dedeUserID string) {
	t.Helper()
	msCfgSnapshot(t)
	cfg := loadCfg(t)
	cfg.SESSDATA = sessdata
	cfg.BiliJct = biliJct
	cfg.DedeUserID = dedeUserID
}

// msRawMap decodes a whole response body into a map (for handlers that answer
// with a bare gin.H rather than the models.Response envelope).
func msRawMap(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode raw body %q: %v", w.Body.String(), err)
	}
	return m
}

// msWaitFor polls cond until it holds or the timeout elapses.
func msWaitFor(t *testing.T, timeout time.Duration, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// msNotifyCapture runs a local apprise "json://" sink. The returned URL is the
// apprise target to put into cfg.Notify.URLs, titles() lists the notification
// titles the stub received. Requests stay on the loopback interface.
func msNotifyCapture(t *testing.T) (target string, titles func() []string) {
	t.Helper()
	var mu sync.Mutex
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			Title string `json:"title"`
		}
		_ = json.NewDecoder(r.Body).Decode(&p)
		mu.Lock()
		got = append(got, p.Title)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return "json://" + strings.TrimPrefix(srv.URL, "http://") + "/notify", func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), got...)
	}
}

// msEnableNotifyForTest snapshots the config and turns the notify router on
// against a loopback sink; it returns the titles collector.
func msEnableNotifyForTest(t *testing.T) func() []string {
	t.Helper()
	target, titles := msNotifyCapture(t)
	msCfgSnapshot(t)
	cfg := loadCfg(t)
	cfg.Notify.Enabled = true
	cfg.Notify.URLs = []string{target}
	return titles
}
