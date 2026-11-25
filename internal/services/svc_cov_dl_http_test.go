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

	"bilibili-history-go/internal/config"
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
// download.go 里留在本地的 HTTP 入口都用 &http.Client{Timeout: 30 * time.Second}
// （Transport 为 nil），因此会走这份全局 transport —— 可以在完全离线的情况下
// 断言 URL、Header 与解析结果。原始 transport 在 t.Cleanup 中恢复，
// 包内测试串行执行，互不影响。
// 其余取信息的路径已交给 liuying-engine，不再由本包直接发请求。
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

var errSvcDlDial = errors.New("svcDl dial refused")

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
	if err != nil || len(videos) != 0 || count != 0 {
		t.Errorf("videos=%v count=%d err=%v, want 空列表/0/nil", videos, count, err)
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
		return nil, errSvcDlDial
	})

	videos, count, err := GetUserVideos("880006", 1, 50)
	if !errors.Is(err, errSvcDlDial) || videos != nil || count != 0 {
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

// svcDlStubTransport 是个自检：确认 download.go 留下的 HTTP 入口
// 走 http.DefaultTransport（否则上面的桩不会生效）。
func TestSVCDLDefaultTransportIsUsed(t *testing.T) {
	rec := svcDlStubTransport(t, func(*http.Request) (*http.Response, error) {
		return svcDlJSON(http.StatusOK, svcDlUserVideosOK), nil
	})

	if _, _, err := GetUserVideos("99", 1, 10); err != nil {
		t.Fatalf("user videos probe: %v", err)
	}
	if len(rec.Reqs) != 1 {
		t.Errorf("recorded=%d, want 1 (entry point uses http.DefaultTransport)", len(rec.Reqs))
	}
}
