package routers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/services"
)

// Reserved namespace for this file: years 2001-2003 are only touched here via
// the BVDLLIST metadata join, and every file/dir created lives under
// <cwd>/output/downloads (the TestMain temp dir) with a "dl" marker in the
// name. The success flows of ExtractVideoInfo / CheckCollection /
// DownloadVideoWithProgress are not exercised: they route through the bili-dl
// library whose api.bilibili.com client cannot be repointed from tests, and
// merging needs a real ffmpeg binary.

// dlTransportStub makes every api.bilibili.com request answered by h instead
// of the internet. services.* builds its http.Clients with the default
// (nil) transport, so replacing http.DefaultTransport is enough.
func dlTransportStub(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	target := strings.TrimPrefix(srv.URL, "http://")
	base := &http.Transport{}
	prev := http.DefaultTransport
	http.DefaultTransport = dlHostRewriter{target: target, base: base}
	t.Cleanup(func() {
		http.DefaultTransport = prev
		base.CloseIdleConnections()
		srv.Close()
	})
}

type dlHostRewriter struct {
	target string
	base   *http.Transport
}

func (r dlHostRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.HasSuffix(req.URL.Host, "api.bilibili.com") {
		return nil, fmt.Errorf("dl offline guard: refusing request to %s", req.URL.Host)
	}
	req = req.Clone(req.Context())
	req.URL.Scheme = "http"
	req.URL.Host = r.target
	return r.base.RoundTrip(req)
}

func dlDownloadsDir() string { return services.GetDownloadOutputPath() }

func dlWriteDownloadFile(t *testing.T, rel string, content []byte) string {
	t.Helper()
	path := filepath.Join(dlDownloadsDir(), rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func TestDLExtractVideoInfoValidation(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	expectMessage(t, expectStatus(t, e, "GET", "/api/download/video_info", "", 400, "error"), "缺少 url 或 bvid 参数")

	// bvid is expanded into the canonical watch URL first; "DLYY0001" holds no
	// BV id, so the service fails offline before bili-dl would be called.
	noBv := expectStatus(t, e, "GET", "/api/download/video_info?bvid=DLYY0001", "", 200, "error")
	expectMessageContains(t, noBv, "无法从 URL 提取 BV 号")
	expectMessageContains(t, noBv, "https://www.bilibili.com/video/DLYY0001")

	expectMessageContains(t, expectStatus(t, e, "GET", "/api/download/video_info?url=https://b23.tv/xyz", "", 200, "error"), "无法从 URL 提取 BV 号")
}

func TestDLGetUserVideos(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/user_videos", "", 400, "error"), "缺少 mid 参数")

	dlTransportStub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/space/wbi/arc/search" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("mid") == "999" {
			fmt.Fprint(w, `{"code":-352,"message":"risk","data":null}`)
			return
		}
		fmt.Fprint(w, `{"code":0,"message":"0","data":{"list":{"vlist":[
			{"bvid":"BVDLUV00000001","title":"DL用户视频","pic":"http://p/1.jpg","aid":12,"length":"1:00"},
			{"bvid":"BVDLUV00000002","title":"DL用户视频二","pic":"http://p/2.jpg","aid":13,"length":"2:00"}]},
			"page":{"count":2}}}`)
	})

	resp := expectStatus(t, e, "GET", "/api/download/user_videos?mid=123&pn=2&ps=5", "", 200, "success").dataMap(t)
	if int(resp["total"].(float64)) != 2 || int(resp["page"].(float64)) != 2 || int(resp["size"].(float64)) != 5 {
		t.Fatalf("paging envelope = %v", resp)
	}
	list := dataArray(t, resp["list"])
	if len(list) != 2 {
		t.Fatalf("list = %v, want two videos", list)
	}
	v0 := list[0].(map[string]interface{})
	if v0["bvid"] != "BVDLUV00000001" || v0["title"] != "DL用户视频" || v0["aid"] != "12" {
		t.Fatalf("first video = %v", v0)
	}

	// pn is parsed leniently: garbage falls back to 0 and is echoed back.
	loose := expectStatus(t, e, "GET", "/api/download/user_videos?mid=123&pn=abc", "", 200, "success").dataMap(t)
	if int(loose["page"].(float64)) != 0 {
		t.Fatalf("page for pn=abc = %v, want 0", loose["page"])
	}

	errResp := expectStatus(t, e, "GET", "/api/download/user_videos?mid=999", "", 200, "error")
	expectMessageContains(t, errResp, "API 返回错误码: -352")
}

func TestDLCheckVideoDownload(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/check_video_download", "", 400, "error"), "缺少 cids 参数")

	dlWriteDownloadFile(t, filepath.Join("dlcheck", "DLVIDEO_900123_p1.mp4"), []byte("x"))

	resp := expectStatus(t, e, "GET", "/api/download/check_video_download?cids=900123,%20,9999999999,", "", 200, "success").dataMap(t)
	if len(resp) != 2 {
		t.Fatalf("blank cids must be dropped, got %v", resp)
	}
	if resp["900123"] != true {
		t.Fatalf("900123 = %v, want true (case-insensitive name match)", resp["900123"])
	}
	if resp["9999999999"] != false {
		t.Fatalf("9999999999 = %v, want false", resp["9999999999"])
	}
}

func TestDLListDownloadedVideos(t *testing.T) {
	// No history rows are seeded: this file runs before history_test.go, whose
	// cross-year sort assertions require 2006-01-13 to stay the oldest row,
	// and before the title-analytics fixtures that assert exact per-year
	// counts. The metadata join therefore runs on a miss and the author comes
	// from the bracket prefix in the file name.
	dlWriteDownloadFile(t, filepath.Join("dllist", "[DL作者甲]带前缀视频 BVDLLIST000001.mp4"), []byte("0123456789"))
	dlWriteDownloadFile(t, filepath.Join("dllist", "DL裸标题影片.mp4"), []byte("ab"))
	dlWriteDownloadFile(t, filepath.Join("dllist", "DL有声书_audio.mp4"), []byte("abc"))
	dlWriteDownloadFile(t, filepath.Join("dllist", "DL忽略.txt"), []byte("x"))
	e := newAPI(t, RegisterDownloadRoutes)

	data := expectStatus(t, e, "GET", "/api/download/list_downloaded_videos?search_term="+urlEncode("BVDLLIST000001"), "", 200, "success").dataMap(t)
	if int(data["total"].(float64)) != 1 {
		t.Fatalf("total = %v, want 1 (%v)", data["total"], data["videos"])
	}
	v := dataArray(t, data["videos"])[0].(map[string]interface{})
	if v["bvid"] != "BVDLLIST000001" || v["title"] != "带前缀视频 BVDLLIST000001" {
		t.Fatalf("parsed name = %v", v)
	}
	if v["author_name"] != "DL作者甲" {
		t.Fatalf("bracket author = %v", v["author_name"])
	}
	if v["cover"] != "" || v["author_face"] != "" || v["author_mid"] != float64(0) {
		t.Fatalf("metadata join on miss must leave empty values: %v", v)
	}
	files := dataArray(t, v["files"])
	f0 := files[0].(map[string]interface{})
	if f0["is_audio_only"] != false || f0["size_mb"].(float64) <= 0 {
		t.Fatalf("file entry = %v", f0)
	}
	if v["download_date"] != time.Now().Format("2006-01-02") {
		t.Fatalf("download_date = %v, want today (local)", v["download_date"])
	}

	// No bracket prefix and no BV id: author comes from the DB join or stays empty.
	plain := expectStatus(t, e, "GET", "/api/download/list_downloaded_videos?search_term="+urlEncode("DL裸标题"), "", 200, "success").dataMap(t)
	pv := dataArray(t, plain["videos"])[0].(map[string]interface{})
	if pv["title"] != "DL裸标题影片" || pv["author_name"] != "" || pv["bvid"] != "" {
		t.Fatalf("plain video = %v", pv)
	}

	audio := expectStatus(t, e, "GET", "/api/download/list_downloaded_videos?search_term="+urlEncode("DL有声书"), "", 200, "success").dataMap(t)
	av := dataArray(t, audio["videos"])[0].(map[string]interface{})
	if av["files"].([]interface{})[0].(map[string]interface{})["is_audio_only"] != true {
		t.Fatalf("audio flag = %v", av)
	}

	// .txt is not a video extension.
	txt := expectStatus(t, e, "GET", "/api/download/list_downloaded_videos?search_term="+urlEncode("DL忽略"), "", 200, "success").dataMap(t)
	if int(txt["total"].(float64)) != 0 {
		t.Fatalf("txt listed as video: %v", txt)
	}

	// limit normalisation and out-of-range page.
	echo := expectStatus(t, e, "GET", "/api/download/list_downloaded_videos?page=0&limit=999&search_term="+urlEncode("DL裸标题"), "", 200, "success").dataMap(t)
	if int(echo["page"].(float64)) != 1 || int(echo["limit"].(float64)) != 20 {
		t.Fatalf("normalised paging echo = %v/%v", echo["page"], echo["limit"])
	}
	empty := expectStatus(t, e, "GET", "/api/download/list_downloaded_videos?page=50&limit=5&search_term="+urlEncode("BVDLLIST000001"), "", 200, "success").dataMap(t)
	if int(empty["total"].(float64)) != 1 {
		t.Fatalf("total = %v, want 1", empty["total"])
	}
	if vs, ok := empty["videos"].([]interface{}); ok && len(vs) != 0 {
		t.Fatalf("out-of-range page = %v, want empty", vs)
	}
}

func TestDLDeleteDownloadedVideo(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video", "", 200, "error"), "缺少 cid 参数")

	f := dlWriteDownloadFile(t, filepath.Join("dldelone", "dldelfile_918273.mp4"), []byte("z"))
	expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?cid=918273", "", 200, "success")
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Fatalf("file %s should have been deleted", f)
	}
	if _, err := os.Stat(filepath.Dir(f)); err != nil {
		t.Fatalf("parent dir must survive a file-only delete: %v", err)
	}

	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?cid=dldelnotexist999", "", 200, "error"), "未找到匹配的视频文件")

	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?directory="+urlEncode("/etc")+"&delete_directory=true", "", 200, "error"), "无效的目录路径")
	missing := filepath.Join(dlDownloadsDir(), "dllmissingdir")
	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?directory="+urlEncode(missing)+"&delete_directory=true", "", 200, "error"), "目录不存在")

	keep := filepath.Join(dlDownloadsDir(), "dldelkeep")
	kf := dlWriteDownloadFile(t, filepath.Join("dldelkeep", "dlkeepme.mp4"), []byte("k"))
	expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?directory="+urlEncode(keep), "", 200, "success")
	if _, err := os.Stat(kf); err != nil {
		t.Fatalf("file must be kept without delete_directory: %v", err)
	}

	deldir := filepath.Join(dlDownloadsDir(), "dldelwhole")
	dlWriteDownloadFile(t, filepath.Join("dldelwhole", "dlvictim.mp4"), []byte("v"))
	got := expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?directory="+urlEncode(deldir)+"&delete_directory=true", "", 200, "success").dataMap(t)
	if got["message"] != "删除成功" {
		t.Fatalf("payload = %v", got)
	}
	if _, err := os.Stat(deldir); !os.IsNotExist(err) {
		t.Fatalf("directory %s should be gone", deldir)
	}
}

func TestDLStreamVideo(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/stream_video", "", 400, "error"), "缺少 file_path 参数")

	dir := t.TempDir()
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/stream_video?file_path="+urlEncode(filepath.Join(dir, "nope.mp4")), "", 404, "error"), "文件不存在")

	path := filepath.Join(dir, "dlmovie.mp4")
	if err := os.WriteFile(path, []byte("DL-STREAM-CONTENT"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := doRaw(t, e, "GET", "/api/download/stream_video?file_path="+urlEncode(path), "")
	if w.Code != http.StatusOK {
		t.Fatalf("stream code = %d body=%q", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != "DL-STREAM-CONTENT" {
		t.Fatalf("streamed bytes = %q", got)
	}
	if ct := w.Header().Get("Content-Type"); ct != "video/mp4" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if ra := w.Header().Get("Accept-Ranges"); ra != "bytes" {
		t.Fatalf("Accept-Ranges = %q", ra)
	}
}

func TestDLCheckFFmpeg(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)
	data := expectStatus(t, e, "GET", "/api/download/check_ffmpeg", "", 200, "success").dataMap(t)
	installed, ok := data["installed"].(bool)
	if !ok {
		t.Fatalf("installed = %v, want bool", data["installed"])
	}
	version, _ := data["version"].(string)
	if installed != (version != "") {
		t.Fatalf("installed=%v but version=%q (probe consistency)", installed, version)
	}
}

func TestDLDownloadVideoSSE(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	w := doRaw(t, e, "POST", "/api/download/download_video", "not-json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON code = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "参数错误") {
		t.Fatalf("bad JSON body = %q", w.Body.String())
	}

	w = doRaw(t, e, "POST", "/api/download/download_video", `{"url":"https://example.com/DLYY0001","sessdata":"s"}`)
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content-type = %q", ct)
	}
	body := w.Body.String()
	for _, want := range []string{"data: 开始下载...", "data: 下载失败: 无法从 URL 提取 BV 号", "data: close"} {
		if !strings.Contains(body, want) {
			t.Fatalf("SSE body %q missing %q", body, want)
		}
	}
}

func TestDLBatchDownloadSSE(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	w := doRaw(t, e, "POST", "/api/download/batch_download", "[1,2")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON code = %d body=%q", w.Code, w.Body.String())
	}

	// video 1 has an empty bvid and video 3 has no bvid key: both must be
	// skipped; only the second (BV-less but non-empty) starts a failing task.
	w = doRaw(t, e, "POST", "/api/download/batch_download",
		`{"videos":[{"bvid":""},{"bvid":"DLYY0001"},{"title":"无bvid"}],"sessdata":"s","only_audio":false}`)
	body := w.Body.String()
	for _, want := range []string{
		"data: [2/3] 开始下载: DLYY0001",
		"data: [2/3] 下载失败: 无法从 URL 提取 BV 号",
		"data: close",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("batch SSE body %q missing %q", body, want)
		}
	}
	if strings.Contains(body, "[1/3]") || strings.Contains(body, "[3/3]") {
		t.Fatalf("empty-bvid videos must be skipped entirely: %q", body)
	}
}

func TestDLDownloadUserVideosSSE(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	w := doRaw(t, e, "POST", "/api/download/download_user_videos", `{"user_id":`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON code = %d", w.Code)
	}

	dlTransportStub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mid") == "666" {
			fmt.Fprint(w, `{"code":-100,"message":"blocked","data":null}`)
			return
		}
		fmt.Fprint(w, `{"code":0,"message":"0","data":{"list":{"vlist":[
			{"bvid":"","title":"DL空bvid","pic":"","aid":1,"length":"0:10"},
			{"bvid":"DLYY0002","title":"DL用户视频","pic":"","aid":2,"length":"0:20"}]},
			"page":{"count":9}}}`)
	})

	fail := doRaw(t, e, "POST", "/api/download/download_user_videos", `{"user_id":"666"}`).Body.String()
	for _, want := range []string{"data: 正在获取用户视频列表...", "data: 获取视频列表失败: API 返回错误码: -100", "data: close"} {
		if !strings.Contains(fail, want) {
			t.Fatalf("list-failure stream %q missing %q", fail, want)
		}
	}
	if strings.Contains(fail, "共找到") {
		t.Fatalf("failed listing must not continue: %q", fail)
	}

	ok := doRaw(t, e, "POST", "/api/download/download_user_videos", `{"user_id":"123","only_audio":true}`).Body.String()
	if !strings.Contains(ok, "data: 共找到 9 个视频") {
		t.Fatalf("success stream = %q", ok)
	}
	if !strings.Contains(ok, "data: [2/2] 开始下载: DL用户视频") || !strings.Contains(ok, "[2/2] 下载失败: 无法从 URL 提取 BV 号") {
		t.Fatalf("per-video events missing: %q", ok)
	}
	if strings.Contains(ok, "[1/2]") {
		t.Fatalf("empty-bvid entry should be skipped: %q", ok)
	}
	if !strings.HasSuffix(ok, "data: close\n\n") {
		t.Fatalf("stream must end with close: %q", ok)
	}
}

func TestDLCollectionRoutes(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	expectMessage(t, expectStatus(t, e, "GET", "/api/collection/check_collection", "", 400, "error"), "缺少 url 参数")
	expectMessage(t, expectStatus(t, e, "GET", "/api/collection/check_collection?url=https%3A%2F%2Fwww.bilibili.com%2Fvideo%2FDLYY0003", "", 200, "error"), "无法从 URL 提取 BV 号")

	w := doRaw(t, e, "POST", "/api/collection/download_collection", "{oops")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "参数错误") {
		t.Fatalf("bad JSON: code=%d body=%q", w.Code, w.Body.String())
	}

	w = doRaw(t, e, "POST", "/api/collection/download_collection", `{"url":"https://b23.tv/DLYY0003"}`)
	body := w.Body.String()
	for _, want := range []string{"data: 正在获取合集信息...", "data: 下载失败: 无法从 URL 提取 BV 号: https://b23.tv/DLYY0003", "data: close"} {
		if !strings.Contains(body, want) {
			t.Fatalf("collection SSE %q missing %q", body, want)
		}
	}
}
