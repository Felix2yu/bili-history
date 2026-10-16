package biliapi

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"
)

// The endpoint URL constants in client.go are package-level vars so that tests can
// repoint them at a local stub server. Every request in this file goes to an
// httptest server; nothing reaches the network.

// capturedRequest records what the stub server saw.
type capturedRequest struct {
	Method   string
	Path     string
	RawQuery string
	Query    url.Values
	Header   http.Header
	Body     string
}

type stubServer struct {
	*httptest.Server
	reqs []capturedRequest
	// responses maps a URL path to the body returned for it.
	responses map[string]string
}

const navPath = "/x/web-interface/nav"

const navFixture = `{"code":0,"message":"0","data":{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad61d4f65b841.png","sub_url":"https://i0.hdslb.com/bfs/wbi/4932caff0ff746eab6f01bf08b70ac45.png"}}}`

// newStub starts an httptest server that answers from responses (keyed by URL
// path). The nav endpoint always answers with navFixture unless a test overrides
// it, because several API methods refresh wbi keys implicitly.
func newStub(t *testing.T, responses map[string]string) *stubServer {
	t.Helper()
	st := &stubServer{responses: map[string]string{}}
	for k, v := range responses {
		st.responses[k] = v
	}
	if _, ok := st.responses[navPath]; !ok {
		st.responses[navPath] = navFixture
	}
	st.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := st.responses[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":-404,"message":"no fixture for %s","data":null}`, r.URL.Path)
			return
		}
		header := w.Header()
		// Both GET and POST responses are gzipped so that the decompression path
		// of Get, GetWithDm and PostForm is exercised by every table case.
		if r.Header.Get("Accept-Encoding") != "" {
			header.Set("Content-Encoding", "gzip")
			header.Set("Content-Type", "application/json")
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			if _, err := gz.Write([]byte(body)); err != nil {
				t.Errorf("gzip write: %v", err)
			}
			if err := gz.Close(); err != nil {
				t.Errorf("gzip close: %v", err)
			}
			_, _ = io.WriteString(w, buf.String())
		} else {
			header.Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, body)
		}
		rec := capturedRequest{
			Method:   r.Method,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
			Query:    r.URL.Query(),
			Header:   r.Header.Clone(),
		}
		if r.Body != nil {
			raw, _ := io.ReadAll(r.Body)
			rec.Body = string(raw)
		}
		st.reqs = append(st.reqs, rec)
	}))
	t.Cleanup(st.Close)
	return st
}

func (s *stubServer) lastRequest(t *testing.T) capturedRequest {
	t.Helper()
	if len(s.reqs) == 0 {
		t.Fatalf("stub server recorded no requests")
	}
	return s.reqs[len(s.reqs)-1]
}

// endpointVars lists every package-level endpoint URL so a test can repoint them
// all at a stub server. They used to be untyped consts, which made the API methods
// untestable without real network access.
func endpointVars() []*string {
	return []*string{
		&HistoryURL, &HistoryDelURL, &VideoInfoURL, &WatchLaterURL, &WatchLaterDelURL,
		&DynamicSpaceURL, &UserCardURL, &FavoriteFolderListURL, &FavoriteCollectedListURL,
		&FavoriteResourceListURL, &FavoriteSeasonListURL, &FavoriteDealURL, &LikedVideoURL,
		&LikeURL, &WbiNavURL,
	}
}

// endpointSnapshot remembers the production URLs before the first test rewrites
// them, so several stub servers can be used inside one test function and every
// value is restored exactly.
var endpointSnapshot []string

func TestMain(m *testing.M) {
	for _, v := range endpointVars() {
		endpointSnapshot = append(endpointSnapshot, *v)
	}
	os.Exit(m.Run())
}

// stubClient repoints all endpoints at srv and returns a logged-in client.
func stubClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	vars := endpointVars()
	for i, v := range vars {
		prod := endpointSnapshot[i]
		u, err := url.Parse(prod)
		if err != nil {
			t.Fatalf("endpoint %q unparseable: %v", prod, err)
		}
		if u.Scheme != "https" || u.Host != "api.bilibili.com" {
			t.Fatalf("endpoint %q unexpectedly not an api.bilibili.com https URL", prod)
		}
		*v = baseURL + u.Path
		vp, restore := v, prod
		t.Cleanup(func() { *vp = restore })
	}
	c := NewClientWithConfig("test-sessdata", "test-csrf-token", "123456")
	if c.BiliJct != "test-csrf-token" || c.DedeUserID != "123456" {
		t.Fatalf("stub client credentials not applied")
	}
	return c
}

func productionPaths(t *testing.T) map[string]string {
	t.Helper()
	paths := map[string]string{}
	for _, prod := range endpointSnapshot {
		u, err := url.Parse(prod)
		if err != nil {
			t.Fatalf("parse %q: %v", prod, err)
		}
		paths[u.Path] = prod
	}
	return paths
}

func okEnvelope(data string) string {
	return fmt.Sprintf(`{"code":0,"message":"0","ttl":1,"data":%s}`, data)
}

const (
	notLoggedInBody  = `{"code":-101,"message":"账号未登录","data":null}`
	riskControlBody  = `{"code":-352,"message":"-352","data":null}`
	malformedBody    = `<html><body>500 Internal Server Error</body></html>`
	dataShapeErrBody = `{"code":0,"message":"0","data":"not-an-object"}`
)

func requireApiErr(t *testing.T, err error, wantContains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantContains)
	}
	if !strings.Contains(err.Error(), wantContains) {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), wantContains)
	}
}

// TestTransportErrors checks that every API method surfaces a connection failure
// instead of panicking or silently returning empty data.
func TestTransportErrors(t *testing.T) {
	calls := map[string]func(*Client) error{
		"FetchWbiKeys":           func(c *Client) error { return c.FetchWbiKeys() },
		"GetHistory":             func(c *Client) error { _, err := c.GetHistory(1, 2, 3); return err },
		"GetVideoInfo":           func(c *Client) error { _, err := c.GetVideoInfo("BV1"); return err },
		"GetWatchLaterList":      func(c *Client) error { _, err := c.GetWatchLaterList(); return err },
		"GetUserCard":            func(c *Client) error { _, err := c.GetUserCard("1"); return err },
		"GetDynamicList":         func(c *Client) error { _, err := c.GetDynamicList("1", "", 30); return err },
		"GetFavoriteFolderList":  func(c *Client) error { _, err := c.GetFavoriteFolderList(); return err },
		"GetCollectedFavFolders": func(c *Client) error { _, err := c.GetCollectedFavoriteFolders("1", 1, 20); return err },
		"GetFavoriteResources":   func(c *Client) error { _, err := c.GetFavoriteResources(1, 1, 20); return err },
		"GetSeasonContents":      func(c *Client) error { _, err := c.GetSeasonContents(1, 1, 20); return err },
		"GetLikedVideos":         func(c *Client) error { _, err := c.GetLikedVideos(1); return err },
		"RemoveFromWatchLater":   func(c *Client) error { return c.RemoveFromWatchLater(1) },
		"DeleteBiliHistory":      func(c *Client) error { return c.DeleteBiliHistory([]string{"BV1"}) },
		"DeleteBiliHistoryByKid": func(c *Client) error { return c.DeleteBiliHistoryByKid("archive_1") },
		"DealFavoriteResource":   func(c *Client) error { return c.DealFavoriteResource("1", "2") },
		"LikeVideo":              func(c *Client) error { return c.LikeVideo("BV1", true) },
	}
	if len(calls) != len(endpointVars())+1 {
		t.Fatalf("transport error table covers %d methods, want %d", len(calls), len(endpointVars())+1)
	}

	// A server that is closed immediately gives a refused connection, so no real
	// network access is attempted.
	closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := closed.URL
	closed.Close()

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			c := stubClient(t, deadURL)
			c.ImgKey = "7cd084941338484aae1ad61d4f65b841"
			c.SubKey = "4932caff0ff746eab6f01bf08b70ac45"
			err := call(c)
			if err == nil {
				t.Fatalf("%s: expected request error against a closed server", name)
			}
			if !strings.Contains(err.Error(), "request error") && !strings.Contains(err.Error(), "create request error") {
				t.Fatalf("%s: error = %q, want a request/transport failure", name, err)
			}
		})
	}
}

// TestAPIMethodResponses drives every API method against recorded fixtures and
// checks the four response-envelope branches each of them handles.
func TestAPIMethodResponses(t *testing.T) {
	type testCase struct {
		name string
		path string
		// ignoresData marks the write endpoints, which decode the envelope but never
		// look at `data`, so a wrongly shaped payload cannot fail them.
		ignoresData bool
		success     string
		call        func(*Client) (any, error)
		check       func(*testing.T, any)
	}

	cases := []testCase{
		{
			name:    "GetHistory",
			path:    "/x/web-interface/history/cursor",
			success: okEnvelope(`{"cursor":{"max":1700000000,"view_at":1700000000,"business":"archive","ps":8},"list":[{"title":"视频标题","bvid":"BV1xx411c7mD","view_at":1690000000,"duration":120,"author_name":"up","author_mid":99,"history":{"bvid":"BV1xx411c7mD","cid":42,"dt":5}}]}`),
			call: func(c *Client) (any, error) {
				return c.GetHistory(55, 66, 8)
			},
			check: func(t *testing.T, got any) {
				d := got.(*HistoryCursorData)
				if d.Cursor.Max != 1700000000 || d.Cursor.Ps != 8 || d.Cursor.Business != "archive" {
					t.Fatalf("cursor = %+v", d.Cursor)
				}
				if len(d.List) != 1 || d.List[0].Bvid != "BV1xx411c7mD" || d.List[0].Title != "视频标题" || d.List[0].History.Cid != 42 {
					t.Fatalf("list = %+v", d.List)
				}
			},
		},
		{
			name:    "GetVideoInfo",
			path:    "/x/web-interface/view",
			success: okEnvelope(`{"bvid":"BV1yy","aid":777,"videos":3,"tid":138,"tname":"搞笑","pic":"p","title":"标题","pubdate":1000,"duration":200,"owner":{"mid":5,"name":"o","face":"f"},"stat":{"view":11,"like":12}}`),
			call: func(c *Client) (any, error) {
				return c.GetVideoInfo("BV1yy")
			},
			check: func(t *testing.T, got any) {
				v := got.(*VideoInfo)
				if v.Bvid != "BV1yy" || v.Aid != 777 || v.Owner.Mid != 5 || v.Stat.View != 11 || v.Stat.Like != 12 || v.Tname != "搞笑" {
					t.Fatalf("video = %+v", v)
				}
			},
		},
		{
			name:    "GetWatchLaterList",
			path:    "/x/v2/history/toview",
			success: okEnvelope(`{"count":2,"list":[{"aid":1,"bvid":"BV1a","title":"a","duration":10,"owner":{"mid":2,"name":"n"},"add_at":33},{"aid":2,"bvid":"BV1b","title":"b","duration":20,"add_at":44}]}`),
			call: func(c *Client) (any, error) {
				return c.GetWatchLaterList()
			},
			check: func(t *testing.T, got any) {
				d := got.(*WatchLaterData)
				if d.Count != 2 || len(d.List) != 2 || d.List[1].Bvid != "BV1b" || d.List[0].Owner.Name != "n" {
					t.Fatalf("data = %+v", d)
				}
			},
		},
		{
			name:    "GetUserCard",
			path:    "/x/web-interface/card",
			success: okEnvelope(`{"card":{"mid":"300","name":"up主","face":"http://face","sign":"签名","level":6,"fans":4000,"attention":12,"archive":30}}`),
			call: func(c *Client) (any, error) {
				return c.GetUserCard("300")
			},
			check: func(t *testing.T, got any) {
				u := got.(*UserCardInfo)
				if u.Mid != "300" || u.Name != "up主" || u.Level != 6 || u.Fans != 4000 || u.Archive != 30 {
					t.Fatalf("card = %+v", u)
				}
			},
		},
		{
			name:    "GetDynamicList",
			path:    "/x/polymer/web-dynamic/v1/feed/space",
			success: okEnvelope(`{"has_more":true,"offset":"dyn-offset","items":[{"id_str":"1001","type":"DRAW","modules":{"post":{"title":"动态标题"}}}]}`),
			call: func(c *Client) (any, error) {
				return c.GetDynamicList("300", "dyn-offset", 30)
			},
			check: func(t *testing.T, got any) {
				d := got.(*DynamicSpaceResponse)
				if !d.HasMore || d.Offset != "dyn-offset" || len(d.Items) != 1 {
					t.Fatalf("dynamic = %+v", d)
				}
				if d.Items[0].IDStr != "1001" || d.Items[0].Type != "DRAW" {
					t.Fatalf("item = %+v", d.Items[0])
				}
				var mods struct {
					Post struct {
						Title string `json:"title"`
					} `json:"post"`
				}
				if err := json.Unmarshal(d.Items[0].Modules, &mods); err != nil || mods.Post.Title != "动态标题" {
					t.Fatalf("modules = %s (%v)", d.Items[0].Modules, err)
				}
			},
		},
		{
			name:    "GetFavoriteFolderList",
			path:    "/x/v3/fav/folder/created/list-all",
			success: okEnvelope(`{"count":1,"list":[{"id":10,"fid":10,"mid":123456,"title":"默认收藏夹","media_count":7,"attr":16,"ctime":5,"upper":{"mid":1,"name":"u","face":"f"}}]}`),
			call: func(c *Client) (any, error) {
				return c.GetFavoriteFolderList()
			},
			check: func(t *testing.T, got any) {
				d := got.(*FavFolderListData)
				if d.Count != 1 || d.List[0].ID != 10 || d.List[0].Title != "默认收藏夹" || d.List[0].MediaCount != 7 || d.List[0].Upper.Mid != 1 {
					t.Fatalf("folders = %+v", d)
				}
			},
		},
		{
			name:    "GetCollectedFavoriteFolders",
			path:    "/x/v3/fav/folder/collected/list",
			success: okEnvelope(`{"count":1,"list":[{"id":20,"fid":21,"title":"别人的收藏夹","link":"https://b23.tv/x"}]}`),
			call: func(c *Client) (any, error) {
				return c.GetCollectedFavoriteFolders("987", 1, 20)
			},
			check: func(t *testing.T, got any) {
				d := got.(*FavFolderListData)
				if len(d.List) != 1 || d.List[0].Fid != 21 || d.List[0].Link == "" {
					t.Fatalf("collected = %+v", d)
				}
			},
		},
		{
			name:    "GetFavoriteResources",
			path:    "/x/v3/fav/resource/list",
			success: okEnvelope(`{"info":{"id":10,"fid":10,"title":"收藏夹"},"medias":[{"id":1,"type":2,"title":"m1","page":1,"duration":60,"cid":9,"ugc":{"bvid":"BV1m1"},"stat":{"view":3}},{"id":2,"title":"m2","duration":70}],"page":{"num":1,"size":20,"count":2,"total":2}}`),
			call: func(c *Client) (any, error) {
				return c.GetFavoriteResources(10, 1, 20)
			},
			check: func(t *testing.T, got any) {
				d := got.(*FavResourceData)
				if d.Info == nil || d.Info.Fid != 10 {
					t.Fatalf("info = %+v", d.Info)
				}
				if len(d.Media) != 2 || d.Media[0].UGC == nil || d.Media[0].UGC.Bvid != "BV1m1" || d.Media[1].Stat != nil {
					t.Fatalf("medias = %+v", d.Media)
				}
				if d.Page.Count != 2 || d.Page.Size != 20 || d.Page.Total != 2 {
					t.Fatalf("page = %+v", d.Page)
				}
			},
		},
		{
			name:    "GetSeasonContents",
			path:    "/x/space/fav/season/list",
			success: okEnvelope(`{"info":{"id":88,"title":"合集"},"medias":[{"id":1,"title":"s1","bvid":"BV1s1","duration":88,"upper":{"mid":3,"name":"n"},"cnt_info":{"view":4}}],"page":{"num":2,"size":20,"count":1,"total":1}}`),
			call: func(c *Client) (any, error) {
				return c.GetSeasonContents(88, 2, 20)
			},
			check: func(t *testing.T, got any) {
				d := got.(*SeasonData)
				if d.Info == nil || d.Info.ID != 88 || len(d.Media) != 1 {
					t.Fatalf("season = %+v", d)
				}
				if d.Media[0].Bvid != "BV1s1" || d.Media[0].CntInfo == nil || d.Media[0].CntInfo.View != 4 {
					t.Fatalf("media = %+v", d.Media[0])
				}
				if d.Page.Num != 2 {
					t.Fatalf("page = %+v", d.Page)
				}
			},
		},
		{
			name:    "GetLikedVideos",
			path:    "/x/space/like/video",
			success: okEnvelope(`{"list":[{"aid":9,"bvid":"BV1like","title":"点赞视频","duration":50,"pubdate":1700,"owner":{"mid":4,"name":"o"},"stat":{"like":1}}]}`),
			call: func(c *Client) (any, error) {
				return c.GetLikedVideos(123456)
			},
			check: func(t *testing.T, got any) {
				d := got.(*LikedVideoData)
				if len(d.List) != 1 || d.List[0].Bvid != "BV1like" || d.List[0].Aid != 9 || d.List[0].Pubdate != 1700 {
					t.Fatalf("liked = %+v", d)
				}
			},
		},
		{
			name:        "RemoveFromWatchLater",
			path:        "/x/v2/history/toview/del",
			ignoresData: true,
			success:     okEnvelope(`null`),
			call: func(c *Client) (any, error) {
				return nil, c.RemoveFromWatchLater(4242)
			},
			check: func(*testing.T, any) {},
		},
		{
			name:        "DeleteBiliHistory",
			path:        "/x/web-interface/history/del",
			ignoresData: true,
			success:     okEnvelope(`null`),
			call: func(c *Client) (any, error) {
				return nil, c.DeleteBiliHistory([]string{"BV1a", "BV1b"})
			},
			check: func(*testing.T, any) {},
		},
		{
			name:        "DeleteBiliHistoryByKid",
			path:        "/x/web-interface/history/del",
			ignoresData: true,
			success:     okEnvelope(`{"code":0,"message":"0"}`),
			call: func(c *Client) (any, error) {
				return nil, c.DeleteBiliHistoryByKid("archive_5")
			},
			check: func(*testing.T, any) {},
		},
		{
			name:        "DealFavoriteResource",
			path:        "/x/v3/fav/resource/deal",
			ignoresData: true,
			success:     okEnvelope(`{"code":0}`),
			call: func(c *Client) (any, error) {
				return nil, c.DealFavoriteResource("3:1:video", "10")
			},
			check: func(*testing.T, any) {},
		},
		{
			name:        "LikeVideo",
			path:        "/x/web-interface/archive/like",
			ignoresData: true,
			success:     okEnvelope(`{"code":0}`),
			call: func(c *Client) (any, error) {
				return nil, c.LikeVideo("BV1like", true)
			},
			check: func(*testing.T, any) {},
		},
	}

	// Make sure the table really covers every endpoint constant. The wbi nav
	// endpoint is exercised by TestFetchWbiKeys / TestSignWbi instead, because it
	// decodes into an anonymous struct rather than the shared BiliResponse shape.
	seen := map[string]bool{navPath: true}
	for _, tc := range cases {
		seen[tc.path] = true
	}
	for _, p := range productionPaths(t) {
		u, _ := url.Parse(p)
		if !seen[u.Path] {
			t.Errorf("endpoint path %s is not covered by the response table", u.Path)
		}
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"/success", func(t *testing.T) {
			srv := newStub(t, map[string]string{tc.path: tc.success})
			c := stubClient(t, srv.URL)
			got, err := tc.call(c)
			if err != nil {
				t.Fatalf("unexpected error: %v (body=%s)", err, srv.lastRequest(t).Body)
			}
			tc.check(t, got)
			req := srv.lastRequest(t)
			if req.Path != tc.path {
				t.Fatalf("request path = %s, want %s", req.Path, tc.path)
			}
			if req.Header.Get("User-Agent") != c.UserAgent {
				t.Fatalf("User-Agent = %q", req.Header.Get("User-Agent"))
			}
			if req.Header.Get("Referer") != "https://www.bilibili.com" {
				t.Fatalf("Referer = %q", req.Header.Get("Referer"))
			}
			if cookie := req.Header.Get("Cookie"); !strings.Contains(cookie, "SESSDATA=test-sessdata") || !strings.Contains(cookie, "bili_jct=test-csrf-token") {
				t.Fatalf("Cookie = %q", cookie)
			}
		})

		t.Run(tc.name+"/api-error-code", func(t *testing.T) {
			srv := newStub(t, map[string]string{tc.path: notLoggedInBody})
			c := stubClient(t, srv.URL)
			_, err := tc.call(c)
			requireApiErr(t, err, "code=-101")
			requireApiErr(t, err, "账号未登录")
			if apiErr, ok := err.(*ApiError); ok {
				if apiErr.Code != -101 {
					t.Fatalf("ApiError code = %d", apiErr.Code)
				}
			}
		})

		t.Run(tc.name+"/risk-control-352", func(t *testing.T) {
			srv := newStub(t, map[string]string{tc.path: riskControlBody})
			c := stubClient(t, srv.URL)
			_, err := tc.call(c)
			requireApiErr(t, err, "code=-352")
		})

		t.Run(tc.name+"/malformed-json", func(t *testing.T) {
			srv := newStub(t, map[string]string{tc.path: malformedBody})
			c := stubClient(t, srv.URL)
			_, err := tc.call(c)
			if err == nil {
				t.Fatalf("expected decode error for non-JSON body")
			}
			// POST endpoints fail differently: PostForm only checks the status, so
			// an HTML error page still arrives with HTTP 200 and fails the unmarshal.
			if !strings.Contains(err.Error(), "unmarshal") {
				t.Fatalf("error = %q, want an unmarshal failure", err)
			}
		})

		t.Run(tc.name+"/data-shape-mismatch", func(t *testing.T) {
			srv := newStub(t, map[string]string{tc.path: dataShapeErrBody})
			c := stubClient(t, srv.URL)
			_, err := tc.call(c)
			// Envelope methods that decode `data` into a struct must report it; pure
			// POST endpoints ignore `data` entirely and therefore succeed.
			if tc.ignoresData {
				if err != nil {
					t.Fatalf("%s ignores the data payload, got error: %v", tc.name, err)
				}
				return
			}
			requireApiErr(t, err, "unmarshal data error")
		})

		t.Run(tc.name+"/no-fixture-404", func(t *testing.T) {
			srv := newStub(t, nil)
			c := stubClient(t, srv.URL)
			_, err := tc.call(c)
			if err == nil {
				t.Fatalf("expected error when the endpoint is missing from the stub")
			}
			// Get/GetWithDm ignore HTTP status, so the -404 code inside the JSON
			// envelope is what surfaces; PostForm checks the status directly.
			if !strings.Contains(err.Error(), "code=-404") && !strings.Contains(err.Error(), "HTTP 404") {
				t.Fatalf("error = %q, want a -404 or HTTP 404 failure", err)
			}
		})
	}
}

func TestFetchWbiKeys(t *testing.T) {
	t.Run("extracts keys from urls", func(t *testing.T) {
		srv := newStub(t, map[string]string{navPath: navFixture})
		c := stubClient(t, srv.URL)
		if err := c.FetchWbiKeys(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ImgKey != "7cd084941338484aae1ad61d4f65b841" {
			t.Fatalf("ImgKey = %q", c.ImgKey)
		}
		if c.SubKey != "4932caff0ff746eab6f01bf08b70ac45" {
			t.Fatalf("SubKey = %q", c.SubKey)
		}
		req := srv.lastRequest(t)
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s", req.Method)
		}
		// GetWithDm must attach the dm anti-crawling parameters.
		for _, k := range []string{"dm_img_list", "dm_img_str", "dm_cover_img_str"} {
			if req.Query.Get(k) == "" {
				t.Fatalf("missing %s in query %q", k, req.RawQuery)
			}
		}
	})

	t.Run("non-zero code", func(t *testing.T) {
		srv := newStub(t, map[string]string{navPath: notLoggedInBody})
		c := stubClient(t, srv.URL)
		err := c.FetchWbiKeys()
		requireApiErr(t, err, "fetch wbi keys failed: code=-101")
		if c.ImgKey != "" || c.SubKey != "" {
			t.Fatalf("keys must stay empty on failure")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		srv := newStub(t, map[string]string{navPath: `{"code":`})
		c := stubClient(t, srv.URL)
		requireApiErr(t, c.FetchWbiKeys(), "unmarshal wbi keys error")
	})

	t.Run("transport error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		dead := srv.URL
		srv.Close()
		c := stubClient(t, dead)
		requireApiErr(t, c.FetchWbiKeys(), "fetch wbi keys error")
	})

	// Regression: FetchWbiKeys used to slice the key out of the wbi image URLs
	// unconditionally, so a URL without an extension made strings.LastIndex
	// return -1 and "slice bounds out of range [-1]" panicked the caller. An
	// unparseable URL is now an error.
	t.Run("key url without extension", func(t *testing.T) {
		srv := newStub(t, map[string]string{navPath: okEnvelope(`{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/noextension","sub_url":"https://i0.hdslb.com/bfs/wbi/sub.png"}}`)})
		c := stubClient(t, srv.URL)
		requireApiErr(t, c.FetchWbiKeys(), "无法从 wbi 地址解析 key")
		if c.ImgKey != "" || c.SubKey != "" {
			t.Errorf("keys must stay unset on a parse failure: %q / %q", c.ImgKey, c.SubKey)
		}
	})

	t.Run("empty key url", func(t *testing.T) {
		srv := newStub(t, map[string]string{navPath: okEnvelope(`{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/","sub_url":"https://i0.hdslb.com/bfs/wbi/sub.png"}}`)})
		c := stubClient(t, srv.URL)
		requireApiErr(t, c.FetchWbiKeys(), "无法从 wbi 地址解析 key")
	})
}

func TestSignWbi(t *testing.T) {
	t.Run("uses cached keys", func(t *testing.T) {
		srv := newStub(t, nil)
		c := stubClient(t, srv.URL)
		c.ImgKey = "7cd084941338484aae1ad61d4f65b841"
		c.SubKey = "4932caff0ff746eab6f01bf08b70ac45"
		params := map[string]string{"mid": "1"}
		sig := c.SignWbi(params)
		if len(sig) != 32 {
			t.Fatalf("SignWbi = %q, want a 32-char md5", sig)
		}
		want := wbiExpectedForTest(t, params, getMixinKey(c.ImgKey+c.SubKey))
		if sig != want {
			t.Fatalf("SignWbi = %q, want %q", sig, want)
		}
		if len(srv.reqs) != 0 {
			t.Fatalf("cached keys must not trigger a nav request, saw %d", len(srv.reqs))
		}
	})

	t.Run("fetches keys when missing", func(t *testing.T) {
		srv := newStub(t, nil)
		c := stubClient(t, srv.URL)
		params := map[string]string{"mid": "1"}
		sig := c.SignWbi(params)
		if len(sig) != 32 {
			t.Fatalf("SignWbi = %q after implicit key fetch", sig)
		}
		if c.ImgKey == "" || c.SubKey == "" {
			t.Fatalf("SignWbi should have populated the wbi keys")
		}
		if len(srv.reqs) != 1 || srv.reqs[0].Path != navPath {
			t.Fatalf("expected one nav request, got %+v", srv.reqs)
		}
	})

	t.Run("returns empty string when keys unavailable", func(t *testing.T) {
		srv := newStub(t, map[string]string{navPath: notLoggedInBody})
		c := stubClient(t, srv.URL)
		if got := c.SignWbi(map[string]string{"mid": "1"}); got != "" {
			t.Fatalf("SignWbi = %q, want empty when nav fails", got)
		}
	})
}

func wbiExpectedForTest(t *testing.T, params map[string]string, mixin string) string {
	t.Helper()
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := params[k]
		for _, c := range []string{"'", "(", ")", "!"} {
			v = strings.ReplaceAll(v, c, "")
		}
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	sum := md5.Sum([]byte(strings.Join(parts, "&") + mixin))
	return fmt.Sprintf("%x", sum)
}

// TestWbiSignatureOnTheWire verifies that the w_rid the client actually sends
// matches the WBI signature of the query parameters that were transmitted.
func TestWbiSignatureOnTheWire(t *testing.T) {
	srv := newStub(t, map[string]string{
		"/x/polymer/web-dynamic/v1/feed/space": okEnvelope(`{"has_more":false,"offset":"","items":[]}`),
	})
	c := stubClient(t, srv.URL)
	c.ImgKey = "7cd084941338484aae1ad61d4f65b841"
	c.SubKey = "4932caff0ff746eab6f01bf08b70ac45"

	if _, err := c.GetDynamicList("300", "off", 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := srv.lastRequest(t)
	q := req.Query

	if q.Get("host_mid") != "300" || q.Get("ps") != "30" || q.Get("offset") != "off" || q.Get("platform") != "web" {
		t.Fatalf("query = %v", q)
	}
	if !strings.Contains(q.Get("features"), "onlyfansVote") {
		t.Fatalf("features = %q", q.Get("features"))
	}
	wrid := q.Get("w_rid")
	if len(wrid) != 32 {
		t.Fatalf("w_rid = %q, want a 32-char signature (raw query %q)", wrid, req.RawQuery)
	}

	// Rebuild the signed string from the transmitted parameters: everything except
	// the signature itself and the dm_* anti-crawling additions.
	signed := map[string]string{}
	for k, vs := range q {
		switch k {
		case "w_rid", "dm_img_list", "dm_img_str", "dm_cover_img_str":
			continue
		}
		if len(vs) != 1 {
			t.Fatalf("parameter %s appeared %d times", k, len(vs))
		}
		signed[k] = vs[0]
	}
	want := wbiExpectedForTest(t, signed, getMixinKey(c.ImgKey+c.SubKey))
	if wrid != want {
		t.Fatalf("w_rid = %q, want %q for signed params %v", wrid, want, signed)
	}

	// Without wbi keys the client still calls the endpoint unsigned.
	srv2 := newStub(t, map[string]string{
		navPath:                                notLoggedInBody,
		"/x/polymer/web-dynamic/v1/feed/space": okEnvelope(`{"has_more":false,"items":[]}`),
	})
	c2 := stubClient(t, srv2.URL)
	if _, err := c2.GetDynamicList("300", "", 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req2 := srv2.lastRequest(t)
	if req2.Query.Get("w_rid") != "" {
		t.Fatalf("unexpected w_rid without keys: %q", req2.RawQuery)
	}
	if req2.Query.Has("offset") {
		t.Fatalf("empty offset should not be sent: %q", req2.RawQuery)
	}
}

func TestGetRequestShape(t *testing.T) {
	srv := newStub(t, map[string]string{
		"/x/web-interface/history/cursor": okEnvelope(`{"cursor":{"max":1,"ps":8},"list":[]}`),
	})
	c := stubClient(t, srv.URL)
	if _, err := c.GetHistory(123, 456, 8); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := srv.lastRequest(t)
	if req.Method != http.MethodGet {
		t.Fatalf("method = %s", req.Method)
	}
	if req.Query.Get("ps") != "8" || req.Query.Get("max") != "123" || req.Query.Get("view_at") != "456" {
		t.Fatalf("query = %v", req.Query)
	}
	if req.Query.Get("business") != "" {
		t.Fatalf("business = %q, want empty", req.Query.Get("business"))
	}
	if got := req.Header.Get("Accept-Encoding"); got != "gzip, deflate" {
		t.Fatalf("Accept-Encoding = %q", got)
	}
	if got := req.Header.Get("Origin"); got != "https://www.bilibili.com" {
		t.Fatalf("Origin = %q", got)
	}

	// max / view_at of 0 are sent as empty strings, not "0".
	srv2 := newStub(t, map[string]string{
		"/x/web-interface/history/cursor": okEnvelope(`{"cursor":{},"list":null}`),
	})
	c2 := stubClient(t, srv2.URL)
	data, err := c2.GetHistory(0, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req2 := srv2.lastRequest(t)
	if req2.Query.Get("max") != "" || req2.Query.Get("view_at") != "" {
		t.Fatalf("query = %v, want empty max/view_at", req2.Query)
	}
	if data.List != nil {
		t.Fatalf("list = %+v, want nil for JSON null", data.List)
	}

	// The stub always gzips when the client advertises Accept-Encoding, so the
	// gzip decoding path is exercised by every request above; assert it directly.
	t.Run("gzip decode", func(t *testing.T) {
		gzSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			_, _ = gw.Write([]byte(`{"plain":"body"}`))
			_ = gw.Close()
			w.Header().Set("Content-Encoding", "gzip")
			_, _ = w.Write(buf.Bytes())
		}))
		t.Cleanup(gzSrv.Close)
		body, err := NewClient("sd").Get(gzSrv.URL+"/plain", map[string]string{"a b": "c d"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(body) != `{"plain":"body"}` {
			t.Fatalf("body = %q", body)
		}
	})

	t.Run("corrupt gzip payload is reported", func(t *testing.T) {
		bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Encoding", "gzip")
			_, _ = io.WriteString(w, "not gzip at all")
		}))
		t.Cleanup(bad.Close)
		_, err := NewClient("").Get(bad.URL, nil)
		requireApiErr(t, err, "gzip reader error")
	})

	t.Run("corrupt gzip payload with dm", func(t *testing.T) {
		bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Encoding", "gzip")
			_, _ = io.WriteString(w, "not gzip at all")
		}))
		t.Cleanup(bad.Close)
		_, err := NewClient("").GetWithDm(bad.URL, nil)
		requireApiErr(t, err, "gzip reader error")
	})

	// A response whose body is truncated against its Content-Length makes
	// io.ReadAll fail, which is the only way to reach the read-body error
	// branches of Get, GetWithDm and PostForm.
	truncated := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Errorf("stub server is not a hijacker")
			return
		}
		conn, bw, err := hj.Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		_, _ = bw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 512\r\n\r\npartial")
		_ = bw.Flush()
		_ = conn.Close()
	}))
	t.Cleanup(truncated.Close)

	t.Run("truncated body reported by Get", func(t *testing.T) {
		_, err := NewClient("").Get(truncated.URL, map[string]string{"a": "b"})
		requireApiErr(t, err, "read body error")
	})

	t.Run("truncated body reported by GetWithDm", func(t *testing.T) {
		_, err := NewClient("").GetWithDm(truncated.URL, map[string]string{"a": "b"})
		requireApiErr(t, err, "read body error")
	})

	t.Run("truncated body reported by PostForm", func(t *testing.T) {
		_, err := NewClient("").PostForm(truncated.URL, url.Values{"a": {"b"}})
		requireApiErr(t, err, "read body error")
	})

	t.Run("http status is ignored by Get", func(t *testing.T) {
		errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"code":0,"message":"0","data":{}}`)
		}))
		t.Cleanup(errSrv.Close)
		body, err := NewClient("").Get(errSrv.URL, nil)
		if err != nil {
			t.Fatalf("Get does not check HTTP status, got error %v", err)
		}
		if !strings.Contains(string(body), `"code":0`) {
			t.Fatalf("body = %q", body)
		}
	})

	t.Run("redirects are followed", func(t *testing.T) {
		var hits int32
		var redirectSrv *httptest.Server
		redirectSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/final" {
				_, _ = io.WriteString(w, "landed")
				return
			}
			hits++
			http.Redirect(w, r, redirectSrv.URL+"/final", http.StatusFound)
		}))
		t.Cleanup(redirectSrv.Close)
		body, err := NewClient("").Get(redirectSrv.URL+"/start", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(body) != "landed" || hits != 1 {
			t.Fatalf("body = %q hits = %d", body, hits)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		_, err := NewClient("").Get("http://[::1", nil)
		requireApiErr(t, err, "parse url error")
		_, err = NewClient("").GetWithDm("http://[::1", nil)
		requireApiErr(t, err, "parse url error")
	})

	t.Run("transport failure", func(t *testing.T) {
		// Get / GetWithDm parse the URL before building the request, so their
		// "create request error" branches (client.go:344, 399) cannot be reached:
		// anything url.Parse accepts is also accepted by http.NewRequest.
		closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		dead := closed.URL
		closed.Close()
		_, err := NewClient("").Get(dead, nil)
		requireApiErr(t, err, "request error")
		_, err = NewClient("").GetWithDm(dead, map[string]string{"a": "b"})
		requireApiErr(t, err, "request error")
		_, err = NewClient("").PostForm(dead, url.Values{"a": {"b"}})
		requireApiErr(t, err, "request error")
	})
}

func TestGetWithDmRequestShape(t *testing.T) {
	srv := newStub(t, map[string]string{
		"/x/web-interface/card": okEnvelope(`{"card":{"mid":"300","name":"up"}}`),
	})
	c := stubClient(t, srv.URL)
	if _, err := c.GetUserCard("300"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := srv.lastRequest(t)
	if req.Method != http.MethodGet {
		t.Fatalf("method = %s", req.Method)
	}
	if req.Query.Get("mid") != "300" {
		t.Fatalf("query = %v", req.Query)
	}
	list := req.Query.Get("dm_img_list")
	if !strings.HasPrefix(list, `[{"x":`) {
		t.Fatalf("dm_img_list was not decoded from the query: %q", list)
	}
	if req.Query.Get("dm_img_str") != "bm8gd2ViZ2w" {
		t.Fatalf("dm_img_str = %q", req.Query.Get("dm_img_str"))
	}
	// Values with spaces are escaped with "+" by GetWithDm.
	if !strings.Contains(req.RawQuery, "mid=300") {
		t.Fatalf("raw query = %q", req.RawQuery)
	}
	// Anonymous clients must not leak a Cookie header.
	anonSrv := newStub(t, map[string]string{"/x/web-interface/card": okEnvelope(`{"card":{"mid":"1"}}`)})
	anon := stubClient(t, anonSrv.URL)
	anon.SESSDATA = ""
	if _, err := anon.GetUserCard("1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := anonSrv.lastRequest(t).Header.Get("Cookie"); got != "" {
		t.Fatalf("anonymous request carried Cookie %q", got)
	}
}

func TestPostFormRequestShape(t *testing.T) {
	srv := newStub(t, map[string]string{
		"/x/web-interface/archive/like": okEnvelope(`{"message":"0"}`),
	})
	c := stubClient(t, srv.URL)

	if err := c.LikeVideo("BV1like", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := srv.lastRequest(t)
	if req.Method != http.MethodPost {
		t.Fatalf("method = %s", req.Method)
	}
	if ct := req.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q", ct)
	}
	form, err := url.ParseQuery(req.Body)
	if err != nil {
		t.Fatalf("body %q is not a form: %v", req.Body, err)
	}
	if form.Get("bvid") != "BV1like" || form.Get("like") != "1" {
		t.Fatalf("form = %v", form)
	}
	if form.Get("csrf") != "test-csrf-token" {
		t.Fatalf("csrf = %q, want the bili_jct token", form.Get("csrf"))
	}

	// Unlike a plain GET, PostForm reports non-200 responses.
	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "forbidden")
	}))
	t.Cleanup(errSrv.Close)
	_, err = NewClient("sd").PostForm(errSrv.URL, url.Values{"a": {"b"}})
	requireApiErr(t, err, "HTTP 403: forbidden")

	// Invalid URL for the POST endpoint hits request creation.
	if _, err := NewClient("sd").PostForm("http://%zz", nil); err == nil {
		t.Fatalf("expected create request error")
	} else if !strings.Contains(err.Error(), "create request error") {
		t.Fatalf("error = %q", err)
	}
}

func TestWriteEndpointsFormPayloads(t *testing.T) {
	t.Run("DeleteBiliHistory joins bvids", func(t *testing.T) {
		srv := newStub(t, map[string]string{"/x/web-interface/history/del": okEnvelope(`null`)})
		c := stubClient(t, srv.URL)
		if err := c.DeleteBiliHistory([]string{"BV1a", "BV1b", "BV1c"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		form, err := url.ParseQuery(srv.lastRequest(t).Body)
		if err != nil {
			t.Fatalf("bad body: %v", err)
		}
		if form.Get("bvid") != "BV1a,BV1b,BV1c" || form.Get("csrf") != "test-csrf-token" {
			t.Fatalf("form = %v", form)
		}
	})

	t.Run("DeleteBiliHistoryByKid sends kid", func(t *testing.T) {
		srv := newStub(t, map[string]string{"/x/web-interface/history/del": okEnvelope(`null`)})
		c := stubClient(t, srv.URL)
		if err := c.DeleteBiliHistoryByKid("live_playback_99"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		form, _ := url.ParseQuery(srv.lastRequest(t).Body)
		if form.Get("kid") != "live_playback_99" || form.Get("csrf") != "test-csrf-token" {
			t.Fatalf("form = %v", form)
		}
	})

	t.Run("RemoveFromWatchLater sends aid and platform", func(t *testing.T) {
		srv := newStub(t, map[string]string{"/x/v2/history/toview/del": okEnvelope(`null`)})
		c := stubClient(t, srv.URL)
		if err := c.RemoveFromWatchLater(4242); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		form, _ := url.ParseQuery(srv.lastRequest(t).Body)
		if form.Get("aid") != "4242" || form.Get("platform") != "web" || form.Get("csrf") != "test-csrf-token" {
			t.Fatalf("form = %v", form)
		}
	})

	t.Run("DealFavoriteResource sends resources and media_ids", func(t *testing.T) {
		srv := newStub(t, map[string]string{"/x/v3/fav/resource/deal": okEnvelope(`{"message":"0"}`)})
		c := stubClient(t, srv.URL)
		if err := c.DealFavoriteResource("3:42:video", "10,11"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		form, _ := url.ParseQuery(srv.lastRequest(t).Body)
		if form.Get("resources") != "3:42:video" || form.Get("media_ids") != "10,11" || form.Get("csrf") != "test-csrf-token" {
			t.Fatalf("form = %v", form)
		}
	})

	t.Run("LikeVideo unlike sends 0", func(t *testing.T) {
		srv := newStub(t, map[string]string{"/x/web-interface/archive/like": okEnvelope(`{"message":"0"}`)})
		c := stubClient(t, srv.URL)
		if err := c.LikeVideo("BV1like", false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		form, _ := url.ParseQuery(srv.lastRequest(t).Body)
		if form.Get("like") != "0" {
			t.Fatalf("form = %v, want like=0", form)
		}
	})

	t.Run("csrf -101 failure", func(t *testing.T) {
		srv := newStub(t, map[string]string{
			"/x/web-interface/history/del": `{"code":-101,"message":"csrf 校验失败","data":null}`,
		})
		c := stubClient(t, srv.URL)
		err := c.DeleteBiliHistory([]string{"BV1a"})
		requireApiErr(t, err, "code=-101")
		if _, ok := err.(*ApiError); !ok {
			t.Fatalf("error type = %T, want *ApiError", err)
		}
	})

	t.Run("missing csrf guards", func(t *testing.T) {
		c := NewClient("sd")
		if got := c.BiliJct; got != "" {
			t.Fatalf("BiliJct = %q", got)
		}
		requireApiErr(t, c.DeleteBiliHistory([]string{"BV1"}), "bili_jct (csrf) is required to delete history")
		requireApiErr(t, c.DeleteBiliHistoryByKid("archive_1"), "bili_jct (csrf) is required to delete history")
		requireApiErr(t, c.RemoveFromWatchLater(1), "bili_jct (csrf) is required to remove watch later items")
		requireApiErr(t, c.DealFavoriteResource("r", "m"), "bili_jct (csrf) is required for favorite operations")
		requireApiErr(t, c.LikeVideo("BV1", true), "bili_jct (csrf) is required for like operations")

		// With csrf present, the argument validation branches are reachable.
		c2 := NewClientWithConfig("sd", "jct", "1")
		requireApiErr(t, c2.DeleteBiliHistory(nil), "bvids is empty")
		requireApiErr(t, c2.DeleteBiliHistory([]string{}), "bvids is empty")
		requireApiErr(t, c2.DeleteBiliHistoryByKid(""), "kid is empty")
	})
}

func TestFavoriteListParams(t *testing.T) {
	srv := newStub(t, map[string]string{
		"/x/v3/fav/folder/created/list-all": okEnvelope(`{"count":0,"list":[]}`),
		"/x/v3/fav/folder/collected/list":   okEnvelope(`{"count":0,"list":[]}`),
		"/x/v3/fav/resource/list":           okEnvelope(`{"medias":[],"page":{}}`),
		"/x/space/fav/season/list":          okEnvelope(`{"medias":[],"page":{}}`),
		"/x/space/like/video":               okEnvelope(`{"list":[]}`),
		"/x/v2/history/toview":              okEnvelope(`{"count":0,"list":[]}`),
		"/x/web-interface/view":             okEnvelope(`{"bvid":"BV1"}`),
	})
	c := stubClient(t, srv.URL)
	c.ImgKey = "7cd084941338484aae1ad61d4f65b841"
	c.SubKey = "4932caff0ff746eab6f01bf08b70ac45"

	if _, err := c.GetFavoriteFolderList(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := srv.lastRequest(t).Query.Get("up_mid"); got != "123456" {
		t.Fatalf("up_mid = %q, want the client's DedeUserID", got)
	}

	if _, err := c.GetCollectedFavoriteFolders("987", 2, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q := srv.lastRequest(t).Query
	if q.Get("up_mid") != "987" || q.Get("pn") != "2" || q.Get("ps") != "20" || q.Get("platform") != "web" || q.Get("web_location") != "0.0" {
		t.Fatalf("collected query = %v", q)
	}
	if len(q.Get("w_rid")) != 32 {
		t.Fatalf("collected folders should be wbi signed: %q", q.Get("w_rid"))
	}

	if _, err := c.GetFavoriteResources(7, 3, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q = srv.lastRequest(t).Query
	if q.Get("media_id") != "7" || q.Get("pn") != "3" || q.Get("ps") != "20" || q.Get("order") != "fav_time" {
		t.Fatalf("resource query = %v", q)
	}

	if _, err := c.GetSeasonContents(8, 1, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q = srv.lastRequest(t).Query
	if q.Get("season_id") != "8" || q.Get("pn") != "1" || q.Get("web_location") != "0.0" {
		t.Fatalf("season query = %v", q)
	}

	if _, err := c.GetLikedVideos(123456); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := srv.lastRequest(t).Query.Get("vmid"); got != "123456" {
		t.Fatalf("vmid = %q", got)
	}

	if _, err := c.GetWatchLaterList(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := srv.lastRequest(t).RawQuery; got != "" {
		t.Fatalf("watch later should be requested without params, raw query = %q", got)
	}

	if _, err := c.GetVideoInfo("BV1xx"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := srv.lastRequest(t).Query.Get("bvid"); got != "BV1xx" {
		t.Fatalf("bvid = %q", got)
	}
}

// TestEndpointConstantsUnchanged guards the const-to-var conversion: the values
// must stay byte-for-byte identical to the production endpoints.
func TestEndpointConstantsUnchanged(t *testing.T) {
	want := map[string]string{
		"HistoryURL":               "https://api.bilibili.com/x/web-interface/history/cursor",
		"HistoryDelURL":            "https://api.bilibili.com/x/web-interface/history/del",
		"VideoInfoURL":             "https://api.bilibili.com/x/web-interface/view",
		"WatchLaterURL":            "https://api.bilibili.com/x/v2/history/toview",
		"WatchLaterDelURL":         "https://api.bilibili.com/x/v2/history/toview/del",
		"DynamicSpaceURL":          "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space",
		"UserCardURL":              "https://api.bilibili.com/x/web-interface/card",
		"FavoriteFolderListURL":    "https://api.bilibili.com/x/v3/fav/folder/created/list-all",
		"FavoriteCollectedListURL": "https://api.bilibili.com/x/v3/fav/folder/collected/list",
		"FavoriteResourceListURL":  "https://api.bilibili.com/x/v3/fav/resource/list",
		"FavoriteSeasonListURL":    "https://api.bilibili.com/x/space/fav/season/list",
		"FavoriteDealURL":          "https://api.bilibili.com/x/v3/fav/resource/deal",
		"LikedVideoURL":            "https://api.bilibili.com/x/space/like/video",
		"LikeURL":                  "https://api.bilibili.com/x/web-interface/archive/like",
		"WbiNavURL":                "https://api.bilibili.com/x/web-interface/nav",
	}
	got := map[string]string{
		"HistoryURL":               HistoryURL,
		"HistoryDelURL":            HistoryDelURL,
		"VideoInfoURL":             VideoInfoURL,
		"WatchLaterURL":            WatchLaterURL,
		"WatchLaterDelURL":         WatchLaterDelURL,
		"DynamicSpaceURL":          DynamicSpaceURL,
		"UserCardURL":              UserCardURL,
		"FavoriteFolderListURL":    FavoriteFolderListURL,
		"FavoriteCollectedListURL": FavoriteCollectedListURL,
		"FavoriteResourceListURL":  FavoriteResourceListURL,
		"FavoriteSeasonListURL":    FavoriteSeasonListURL,
		"FavoriteDealURL":          FavoriteDealURL,
		"LikedVideoURL":            LikedVideoURL,
		"LikeURL":                  LikeURL,
		"WbiNavURL":                WbiNavURL,
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("%s = %q, want %q", name, got[name], w)
		}
	}
	if len(got) != len(endpointVars()) {
		t.Errorf("endpoint table has %d entries but client.go exposes %d vars", len(got), len(endpointVars()))
	}
	// After every stubbing test the live values must be back to the snapshot.
	for i, v := range endpointVars() {
		if *v != endpointSnapshot[i] {
			t.Errorf("endpoint var %d leaked a stub value: %q", i, *v)
		}
	}
}

// TestPostFormDecodesGzipBody covers the transport asymmetry the write path used
// to have: every request advertises "Accept-Encoding: gzip, deflate" through
// getHeaders, but PostForm used to read the body without checking
// Content-Encoding, unlike Get and GetWithDm. A server that gzipped a
// write-response body made the JSON decode fail, so a successful Bilibili write
// was reported as an error. PostForm now inflates gzip like the GET helpers.
func TestPostFormDecodesGzipBody(t *testing.T) {
	gzSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, _ = gw.Write([]byte(`{"code":0,"message":"0","data":null}`))
		_ = gw.Close()
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(buf.Bytes())
	}))
	t.Cleanup(gzSrv.Close)

	c := NewClientWithConfig("sd", "jct", "1")
	original := WatchLaterDelURL
	WatchLaterDelURL = gzSrv.URL + "/x/v2/history/toview/del"
	t.Cleanup(func() { WatchLaterDelURL = original })

	if err := c.RemoveFromWatchLater(1); err != nil {
		t.Fatalf("PostForm must decode a gzip body: %v", err)
	}

	// The same handler works fine for a GET helper, which does inflate gzip.
	body, err := c.Get(gzSrv.URL+"/x/v2/history/toview", nil)
	if err != nil {
		t.Fatalf("Get should decode gzip: %v", err)
	}
	if !strings.Contains(string(body), `"code":0`) {
		t.Fatalf("body = %q", body)
	}
}

// TestGzipBodyErrors covers the inflate failure branch shared by the three
// transport helpers: a response that announces gzip but carries no gzip stream
// must surface "gzip reader error" instead of a confusing JSON decode error.
func TestGzipBodyErrors(t *testing.T) {
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write([]byte("not gzipped at all"))
	}))
	t.Cleanup(badSrv.Close)

	c := NewClientWithConfig("sd", "jct", "1")
	t.Run("Get", func(t *testing.T) {
		_, err := c.Get(badSrv.URL+"/x", nil)
		requireApiErr(t, err, "gzip reader error")
	})
	t.Run("GetWithDm", func(t *testing.T) {
		_, err := c.GetWithDm(badSrv.URL+"/x", nil)
		requireApiErr(t, err, "gzip reader error")
	})
	t.Run("PostForm", func(t *testing.T) {
		_, err := c.PostForm(badSrv.URL+"/x", url.Values{})
		requireApiErr(t, err, "gzip reader error")
	})
}
