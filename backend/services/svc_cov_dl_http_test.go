package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"bilibili-history-go/config"
)

// svcDlReq 是 RoundTripper 截获到的请求快照（http.Client 会在传输前后改写
// req 对象，因此只保存断言需要的字段）。
type svcDlReq struct {
	Method string
	URL    *url.URL
	Header http.Header
}

func (r *svcDlReq) query() url.Values {
	if r.URL == nil {
		return url.Values{}
	}
	return r.URL.Query()
}

// svcDlRecorder 记录被截获的请求，保证所有断言都不依赖真实网络。
type svcDlRecorder struct {
	Reqs []*svcDlReq
}

// svcDlStubTransport 用假 RoundTripper 替换 http.DefaultTransport。
// download.go 里 fetchVideoDetail / fetchDashStreams / GetUserVideos 都用
// &http.Client{Timeout: 30 * time.Second}（Transport 为 nil），因此会走这份
// 全局 transport —— 可以在完全离线的情况下断言 URL、Header 与解析结果。
// 原始 transport 在 t.Cleanup 中恢复，包内测试串行执行，互不影响。
func svcDlStubTransport(t *testing.T, fn func(*http.Request) (*http.Response, error)) *svcDlRecorder {
	t.Helper()
	rec := &svcDlRecorder{}
	orig := http.DefaultTransport
	http.DefaultTransport = &svcDlTransport{rec: rec, fn: fn}
	t.Cleanup(func() { http.DefaultTransport = orig })
	return rec
}

// svcDlTransport 是被替换进 http.DefaultTransport 的假传输层。
type svcDlTransport struct {
	rec *svcDlRecorder
	fn  func(*http.Request) (*http.Response, error)
}

func (tr *svcDlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tr.rec.Reqs = append(tr.rec.Reqs, &svcDlReq{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header.Clone(),
	})
	return tr.fn(req)
}

// svcDlJSON 构造一个 JSON 响应体。
func svcDlJSON(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// svcDlFailReader 让 io.ReadAll 失败，覆盖响应体读取错误分支。
type svcDlFailReader struct{}

func (svcDlFailReader) Read([]byte) (int, error) { return 0, errors.New("svcDl body read boom") }
func (svcDlFailReader) Close() error             { return nil }

func svcDlJSONBodyBroken(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{},
		Body:       svcDlFailReader{},
	}
}

// svcDlStubOK 固定返回给定 JSON，且断言只发生一次请求。
func svcDlStubOK(t *testing.T, body string) *svcDlRecorder {
	t.Helper()
	return svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return svcDlJSON(http.StatusOK, body), nil
	})
}

var svcDlErrDial = errors.New("svcDl dial refused")

// ---------------------------------------------------------------------------
// fetchVideoDetail
// ---------------------------------------------------------------------------

const svcDlDetailOK = `{"code":0,"data":{"title":"svc 投稿标题","pic":"https://cover.invalid/svcDL.jpg",` +
	`"owner":{"name":"svcDL作者","face":"https://face.invalid/svcDL.jpg","mid":880001}}}`

func TestSVCDLFetchVideoDetailSuccess(t *testing.T) {
	rec := svcDlStubOK(t, svcDlDetailOK)

	got, err := fetchVideoDetail("BVsvcDL0001", "")
	if err != nil {
		t.Fatalf("fetchVideoDetail: %v", err)
	}
	if got.Title != "svc 投稿标题" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Cover != "https://cover.invalid/svcDL.jpg" {
		t.Errorf("cover = %q", got.Cover)
	}
	if got.Owner.Name != "svcDL作者" || got.Owner.Face != "https://face.invalid/svcDL.jpg" || got.Owner.Mid != 880001 {
		t.Errorf("owner = %+v", got.Owner)
	}

	if len(rec.Reqs) != 1 {
		t.Fatalf("requests captured = %d, want 1", len(rec.Reqs))
	}
	req := rec.Reqs[0]
	if req.Method != "GET" {
		t.Errorf("method = %q, want GET", req.Method)
	}
	if req.URL.Host != "api.bilibili.com" || req.URL.Path != "/x/web-interface/view" {
		t.Errorf("url = %q", req.URL)
	}
	if q := req.query().Get("bvid"); q != "BVsvcDL0001" {
		t.Errorf("query bvid = %q", q)
	}
	if !strings.HasPrefix(req.Header.Get("User-Agent"), "Mozilla/5.0") {
		t.Errorf("user-agent = %q", req.Header.Get("User-Agent"))
	}
	if req.Header.Get("Referer") != "https://www.bilibili.com/" {
		t.Errorf("referer = %q", req.Header.Get("Referer"))
	}
	// cookie 为空时不应附带 SESSDATA
	if cookie := req.Header.Get("Cookie"); strings.Contains(cookie, "SESSDATA") {
		t.Errorf("cookie header = %q, want no SESSDATA", cookie)
	}
}

func TestSVCDLFetchVideoDetailWithCookie(t *testing.T) {
	rec := svcDlStubOK(t, svcDlDetailOK)

	if _, err := fetchVideoDetail("BVsvcDL0002", "svcDL-cookie"); err != nil {
		t.Fatalf("fetchVideoDetail: %v", err)
	}
	if len(rec.Reqs) != 1 {
		t.Fatalf("reqs = %d", len(rec.Reqs))
	}
	if cookie := rec.Reqs[0].Header.Get("Cookie"); !strings.Contains(cookie, "SESSDATA=svcDL-cookie") {
		t.Errorf("cookie header = %q, want SESSDATA=svcDL-cookie", cookie)
	}
}

func TestSVCDLFetchVideoDetailAPIErrorCode(t *testing.T) {
	svcDlStubOK(t, `{"code":-352,"data":{"title":""}}`)

	got, err := fetchVideoDetail("BVsvcDL0003", "")
	if err == nil || !strings.Contains(err.Error(), "API 返回错误码: -352") {
		t.Fatalf("err = %v, want API 错误码 -352", err)
	}
	if got != nil {
		t.Errorf("data = %+v, want nil on error code", got)
	}
}

func TestSVCDLFetchVideoDetailBadJSON(t *testing.T) {
	svcDlStubOK(t, `{"code":0,"data":`)

	if _, err := fetchVideoDetail("BVsvcDL0004", ""); err == nil || !strings.Contains(err.Error(), "unexpected end") {
		t.Fatalf("err = %v, want json parse error", err)
	}
}

func TestSVCDLFetchVideoDetailTransportError(t *testing.T) {
	svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return nil, svcDlErrDial
	})

	_, err := fetchVideoDetail("BVsvcDL0005", "")
	if !errors.Is(err, svcDlErrDial) {
		t.Fatalf("err = %v, want wrapped dial error", err)
	}
}

func TestSVCDLFetchVideoDetailBodyReadError(t *testing.T) {
	svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return svcDlJSONBodyBroken(http.StatusOK), nil
	})

	_, err := fetchVideoDetail("BVsvcDL0006", "")
	if err == nil || !strings.Contains(err.Error(), "svcDl body read boom") {
		t.Fatalf("err = %v, want body read error", err)
	}
}

func TestSVCDLFetchVideoDetailInvalidBVInURL(t *testing.T) {
	rec := svcDlStubOK(t, svcDlDetailOK)

	// 控制字符会让 url.Parse 失败，请求根本不会发出。
	if _, err := fetchVideoDetail("BVsvcDL\n0007", ""); err == nil ||
		!strings.Contains(err.Error(), "invalid control character") {
		t.Fatalf("err = %v, want invalid control character", err)
	}
	if len(rec.Reqs) != 0 {
		t.Errorf("requests = %d, want 0 (no dial on bad URL)", len(rec.Reqs))
	}
}

// ---------------------------------------------------------------------------
// fetchDashStreams
// ---------------------------------------------------------------------------

const svcDlDashOK = `{"code":0,"data":{"dash":{` +
	`"video":[{"id":116,"codecs":"avc1.640032","width":1920,"height":1080,"bandwidth":1800000,"base_url":"https://v.svc.invalid/avc.m4s"},` +
	`{"id":120,"codecs":"av01.0.05M.08","width":3840,"height":2160,"bandwidth":9000000,"base_url":"https://v.svc.invalid/av1.m4s"}],` +
	`"audio":[{"id":30280,"codecs":"mp4a.40.2","bandwidth":192000,"base_url":"https://a.svc.invalid/high.m4a"},` +
	`{"id":30216,"codecs":"mp4a.40.15","bandwidth":64000,"base_url":"https://a.svc.invalid/low.m4a"}]}}}`

func TestSVCDLFetchDashStreamsSuccess(t *testing.T) {
	rec := svcDlStubOK(t, svcDlDashOK)

	videos, audios, err := fetchDashStreams("BVsvcDL0010", "330010", "")
	if err != nil {
		t.Fatalf("fetchDashStreams: %v", err)
	}
	if len(videos) != 2 || len(audios) != 2 {
		t.Fatalf("streams = (%d video, %d audio), want (2, 2)", len(videos), len(audios))
	}
	if videos[0].ID != 116 || videos[0].Codecs != "avc1.640032" || videos[0].Width != 1920 ||
		videos[0].Height != 1080 || videos[0].Bandwidth != 1800000 ||
		videos[0].BaseURL != "https://v.svc.invalid/avc.m4s" {
		t.Errorf("video[0] = %+v", videos[0])
	}
	if videos[1].ID != 120 || videos[1].Codecs != "av01.0.05M.08" {
		t.Errorf("video[1] = %+v", videos[1])
	}
	if audios[0].ID != 30280 || audios[0].Bandwidth != 192000 ||
		audios[0].BaseURL != "https://a.svc.invalid/high.m4a" {
		t.Errorf("audio[0] = %+v", audios[0])
	}
	if audios[1].Codecs != "mp4a.40.15" {
		t.Errorf("audio[1] = %+v", audios[1])
	}

	req := rec.Reqs[0]
	if req.URL.Path != "/x/player/wbi/playurl" {
		t.Errorf("path = %q", req.URL.Path)
	}
	q := req.query()
	if q.Get("bvid") != "BVsvcDL0010" || q.Get("cid") != "330010" {
		t.Errorf("query = %v", q)
	}
	if q.Get("fnval") != "3216" || q.Get("fourk") != "1" || q.Get("qn") != "127" || q.Get("fnver") != "0" {
		t.Errorf("quality query params = %v", q)
	}
	if strings.Contains(req.Header.Get("Cookie"), "SESSDATA") {
		t.Errorf("cookie = %q, want none", req.Header.Get("Cookie"))
	}
}

func TestSVCDLFetchDashStreamsWithCookie(t *testing.T) {
	rec := svcDlStubOK(t, svcDlDashOK)

	if _, _, err := fetchDashStreams("BVsvcDL0011", "330011", "svcDL-dash-cookie"); err != nil {
		t.Fatalf("fetchDashStreams: %v", err)
	}
	if got := rec.Reqs[0].Header.Get("Cookie"); !strings.Contains(got, "SESSDATA=svcDL-dash-cookie") {
		t.Errorf("cookie = %q", got)
	}
}

func TestSVCDLFetchDashStreamsEmptyLists(t *testing.T) {
	svcDlStubOK(t, `{"code":0,"data":{"dash":{"video":[],"audio":[]}}}`)

	videos, audios, err := fetchDashStreams("BVsvcDL0012", "330012", "")
	if err != nil {
		t.Fatalf("empty dash should not error: %v", err)
	}
	if len(videos) != 0 || len(audios) != 0 {
		t.Errorf("videos=%d audios=%d, want 0/0", len(videos), len(audios))
	}
}

func TestSVCDLFetchDashStreamsAPIErrorCode(t *testing.T) {
	svcDlStubOK(t, `{"code":-10,"data":{}}`)

	videos, audios, err := fetchDashStreams("BVsvcDL0013", "330013", "")
	if err == nil || !strings.Contains(err.Error(), "API 返回错误码: -10") {
		t.Fatalf("err = %v", err)
	}
	if videos != nil || audios != nil {
		t.Errorf("want nil streams on error, got %v / %v", videos, audios)
	}
}

func TestSVCDLFetchDashStreamsBadJSON(t *testing.T) {
	svcDlStubOK(t, `not-json`)

	if _, _, err := fetchDashStreams("BVsvcDL0014", "330014", ""); err == nil {
		t.Fatal("want json unmarshal error")
	}
}

func TestSVCDLFetchDashStreamsTransportError(t *testing.T) {
	svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return nil, svcDlErrDial
	})

	if _, _, err := fetchDashStreams("BVsvcDL0015", "330015", ""); !errors.Is(err, svcDlErrDial) {
		t.Fatalf("err = %v, want dial error", err)
	}
}

func TestSVCDLFetchDashStreamsBodyReadError(t *testing.T) {
	svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return svcDlJSONBodyBroken(http.StatusOK), nil
	})

	if _, _, err := fetchDashStreams("BVsvcDL0016", "330016", ""); err == nil ||
		!strings.Contains(err.Error(), "svcDl body read boom") {
		t.Fatalf("err = %v", err)
	}
}

func TestSVCDLFetchDashStreamsControlCharInBV(t *testing.T) {
	// fetchDashStreams 先把固定常量 URL 交给 url.Parse，再用 query.Encode()
	// 注入 bvid/cid，所以控制字符只会被百分号转义，http.NewRequest 不会失败
	// （download.go:234-237 的 NewRequest 错误分支因此不可达）。
	rec := svcDlStubOK(t, svcDlDashOK)

	videos, audios, err := fetchDashStreams("BVsvcDL\n0017", "330017", "")
	if err != nil {
		t.Fatalf("fetchDashStreams = %v, want escaped request to succeed", err)
	}
	if len(videos) != 2 || len(audios) != 2 {
		t.Errorf("streams = (%d,%d), want (2,2)", len(videos), len(audios))
	}
	if len(rec.Reqs) != 1 {
		t.Fatalf("requests = %d, want 1", len(rec.Reqs))
	}
	if got := rec.Reqs[0].query().Get("bvid"); got != "BVsvcDL\n0017" {
		t.Errorf("decoded bvid = %q", got)
	}
	if !strings.Contains(rec.Reqs[0].URL.RawQuery, "%0A") {
		t.Errorf("raw query = %q, want percent-encoded control char", rec.Reqs[0].URL.RawQuery)
	}
}

// ---------------------------------------------------------------------------
// GetUserVideos
// ---------------------------------------------------------------------------

const svcDlUserVideosOK = `{"code":0,"data":{"list":{"vlist":[` +
	`{"bvid":"BVsvcDL0020","title":"svc 用户视频一","pic":"https://pic.invalid/svcDL1.jpg","aid":900001,"length":"01:23"},` +
	`{"bvid":"BVsvcDL0021","title":"svc 用户视频二","pic":"","aid":900002,"length":"02:00"}]},` +
	`"page":{"count":2}}}`

func TestSVCDLGetUserVideosSuccess(t *testing.T) {
	rec := svcDlStubOK(t, svcDlUserVideosOK)

	videos, count, err := GetUserVideos("880001", 2, 30)
	if err != nil {
		t.Fatalf("GetUserVideos: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if len(videos) != 2 {
		t.Fatalf("videos = %d, want 2", len(videos))
	}
	if videos[0]["bvid"] != "BVsvcDL0020" || videos[0]["title"] != "svc 用户视频一" ||
		videos[0]["pic"] != "https://pic.invalid/svcDL1.jpg" || videos[0]["aid"] != "900001" ||
		videos[0]["length"] != "01:23" {
		t.Errorf("videos[0] = %v", videos[0])
	}
	if videos[1]["aid"] != "900002" || videos[1]["pic"] != "" {
		t.Errorf("videos[1] = %v", videos[1])
	}

	req := rec.Reqs[0]
	if req.URL.Path != "/x/space/wbi/arc/search" {
		t.Errorf("path = %q", req.URL.Path)
	}
	q := req.query()
	if q.Get("mid") != "880001" || q.Get("pn") != "2" || q.Get("ps") != "30" || q.Get("order") != "pubdate" {
		t.Errorf("query = %v", q)
	}
	if req.Header.Get("Referer") != "https://www.bilibili.com" {
		t.Errorf("referer = %q", req.Header.Get("Referer"))
	}
	// 测试配置 SESSDATA 为空 => 不下发 Cookie 头
	if req.Header.Get("Cookie") != "" {
		t.Errorf("cookie = %q, want empty when SESSDATA unset", req.Header.Get("Cookie"))
	}
}

func TestSVCDLGetUserVideosWithConfigCookie(t *testing.T) {
	// getCookie() 走全局配置：临时注入 SESSDATA 才能覆盖 Cookie 头分支。
	svcWithNotifyConfig(t, func(c *config.Config) {
		c.SESSDATA = "svcDL-config-cookie"
	})
	rec := svcDlStubOK(t, svcDlUserVideosOK)

	if _, _, err := GetUserVideos("880002", 1, 50); err != nil {
		t.Fatalf("GetUserVideos: %v", err)
	}
	if got := rec.Reqs[0].Header.Get("Cookie"); got != "svcDL-config-cookie" {
		t.Errorf("cookie = %q, want svcDL-config-cookie", got)
	}
}

func TestSVCDLGetUserVideosEmptyList(t *testing.T) {
	svcDlStubOK(t, `{"code":0,"data":{"list":{"vlist":[]},"page":{"count":0}}}`)

	videos, count, err := GetUserVideos("880003", 1, 50)
	if err != nil || videos != nil || count != 0 {
		t.Errorf("videos=%v count=%d err=%v, want nil/0/nil", videos, count, err)
	}
}

func TestSVCDLGetUserVideosAPIErrorCode(t *testing.T) {
	svcDlStubOK(t, `{"code":-403,"data":{}}`)

	videos, count, err := GetUserVideos("880004", 1, 50)
	if err == nil || !strings.Contains(err.Error(), "API 返回错误码: -403") {
		t.Fatalf("err = %v", err)
	}
	if videos != nil || count != 0 {
		t.Errorf("videos=%v count=%d, want nil/0", videos, count)
	}
}

func TestSVCDLGetUserVideosBadJSON(t *testing.T) {
	svcDlStubOK(t, `{"code":0,`)

	if _, _, err := GetUserVideos("880005", 1, 50); err == nil {
		t.Fatal("want unmarshal error")
	} else {
		var pe *json.SyntaxError
		if !errors.As(err, &pe) {
			t.Logf("error type = %T (%v)", err, err)
		}
	}
}

func TestSVCDLGetUserVideosTransportError(t *testing.T) {
	svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return nil, svcDlErrDial
	})

	videos, count, err := GetUserVideos("880006", 1, 50)
	if !errors.Is(err, svcDlErrDial) || videos != nil || count != 0 {
		t.Fatalf("videos=%v count=%d err=%v", videos, count, err)
	}
}

func TestSVCDLGetUserVideosBodyReadError(t *testing.T) {
	svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return svcDlJSONBodyBroken(http.StatusOK), nil
	})

	if _, _, err := GetUserVideos("880007", 1, 50); err == nil ||
		!strings.Contains(err.Error(), "svcDl body read boom") {
		t.Fatalf("err = %v", err)
	}
}

func TestSVCDLGetUserVideosInvalidMidInURL(t *testing.T) {
	rec := svcDlStubOK(t, svcDlUserVideosOK)

	if _, _, err := GetUserVideos("8800\n08", 1, 50); err == nil ||
		!strings.Contains(err.Error(), "invalid control character") {
		t.Fatalf("err = %v", err)
	}
	if len(rec.Reqs) != 0 {
		t.Errorf("requests = %d, want 0", len(rec.Reqs))
	}
}

// svcDlStubTransport 是个自检：确认 download.go 用的三个 HTTP 入口
// 都走 http.DefaultTransport（否则上面的桩不会生效）。
func TestSVCDLDefaultTransportIsUsed(t *testing.T) {
	var called int
	rec := svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		called++
		return svcDlJSON(http.StatusOK, `{"code":0,"data":{"title":"svc probe"}}`), nil
	})

	detail, err := fetchVideoDetail("BVsvcDL0099", "")
	if err != nil || detail.Title != "svc probe" {
		t.Fatalf("probe via DefaultTransport: detail=%+v err=%v", detail, err)
	}
	if _, _, err := fetchDashStreams("BVsvcDL0099", "1", ""); err != nil {
		t.Fatalf("dash probe: %v", err)
	}
	if _, _, err := GetUserVideos("99", 1, 10); err != nil {
		t.Fatalf("user videos probe: %v", err)
	}
	if called != 3 || len(rec.Reqs) != 3 {
		t.Errorf("called=%d recorded=%d, want 3/3 (all three entry points use http.DefaultTransport)", called, len(rec.Reqs))
	}
}
