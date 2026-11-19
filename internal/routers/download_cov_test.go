package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/internal/services"
)

// Reserved namespace for this file: years 2001-2003 are only touched here via
// the BVDLLIST metadata join, and every file/dir created lives under
// <cwd>/output/downloads (the TestMain temp dir) with a "dl" marker in the
// name. The success flows of ExtractVideoInfo / CheckCollection / Download
// are not exercised: they route through the bili-dl library and then fetch
// bytes from bilibili's CDN, which the offline guard below refuses on purpose.

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

	// bvid 先拼成标准观看链接；"DLYY0001" 不含 BV 号，
	// 服务层会在 bili-dl 发起任何请求之前失败。
	noBv := expectStatus(t, e, "GET", "/api/download/video_info?bvid=DLYY0001", "", 200, "error")
	expectMessageContains(t, noBv, "无法解析链接")
	expectMessageContains(t, noBv, "https://www.bilibili.com/video/DLYY0001")

	expectMessageContains(t, expectStatus(t, e, "GET", "/api/download/video_info?url=https://example.com/no-bv", "", 200, "error"), "无法解析链接")
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
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/check_video_download", "", 400, "error"), "缺少 bvids 参数")

	dlWriteDownloadFile(t, filepath.Join("dlcheck", "DL视频_BVDLCHK00000001.mp4"), []byte("x"))
	// 封面 sidecar 不算已下载成品。
	dlWriteDownloadFile(t, filepath.Join("dlcheck", "DL封面_BVDLCHK00000002.cover.jpg"), []byte("y"))

	resp := expectStatus(t, e, "GET", "/api/download/check_video_download?bvids=BVDLCHK00000001,%20,BVDLCHK00000002,BVDLABSENT99,", "", 200, "success").dataMap(t)
	if len(resp) != 3 {
		t.Fatalf("blank bvids must be dropped, got %v", resp)
	}
	if resp["BVDLCHK00000001"] != true {
		t.Fatalf("BVDLCHK00000001 = %v, want true", resp["BVDLCHK00000001"])
	}
	if resp["BVDLCHK00000002"] != false {
		t.Fatalf("cover-only BV = %v, want false", resp["BVDLCHK00000002"])
	}
	if resp["BVDLABSENT99"] != false {
		t.Fatalf("BVDLABSENT99 = %v, want false", resp["BVDLABSENT99"])
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
	if v["bvid"] != "BVDLLIST000001" || v["title"] != "带前缀视频" {
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

	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video", "", 200, "error"), "缺少 bvid 参数")

	// 文件名里的大小写与查询参数不一致，匹配仍要命中；附加内容一起删除。
	f := dlWriteDownloadFile(t, filepath.Join("dldelone", "DL视频_BVDLDEL00000001.mp4"), []byte("z"))
	sidecar := dlWriteDownloadFile(t, filepath.Join("dldelone", "DL视频_BVDLDEL00000001.danmaku.ass"), []byte("d"))
	sibling := dlWriteDownloadFile(t, filepath.Join("dldelone", "DL另一条_BVDLDEL00000002.mp4"), []byte("s"))
	expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?bvid=bvdldel00000001", "", 200, "success")
	for _, path := range []string{f, sidecar} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s should have been deleted", path)
		}
	}
	if _, err := os.Stat(sibling); err != nil {
		t.Fatalf("同 UP 其它条目必须保留: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(f)); err != nil {
		t.Fatalf("parent dir must survive a file-only delete: %v", err)
	}

	// 最后一条删除时目录被清空，随之移除。
	expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?bvid=BVDLDEL00000002&delete_directory=true", "", 200, "success")
	if _, err := os.Stat(filepath.Dir(sibling)); !os.IsNotExist(err) {
		t.Errorf("emptied directory should be removed, stat = %v", err)
	}

	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?bvid=BVDLDELNOTEXIST999", "", 200, "error"), "未找到匹配的视频文件")

	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?directory="+urlEncode("/etc")+"&delete_directory=true", "", 200, "error"), "无效的目录路径")
	// 只把 downloads 当字符串前缀的兄弟目录同样必须被拒绝。
	siblingDir := dlDownloadsDir() + "-evil"
	if err := os.MkdirAll(siblingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(siblingDir) })
	expectMessage(t, expectStatus(t, e, "DELETE", "/api/download/delete_downloaded_video?directory="+urlEncode(siblingDir)+"&delete_directory=true", "", 200, "error"), "无效的目录路径")
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

	// downloads 之外的路径直接拒绝，不再泄露本机任意文件。
	outside := filepath.Join(t.TempDir(), "nope.mp4")
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/stream_video?file_path="+urlEncode(outside), "", 403, "error"), "无权访问该文件")

	missing := filepath.Join(dlDownloadsDir(), "dlstreammissing", "nope.mp4")
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/stream_video?file_path="+urlEncode(missing), "", 404, "error"), "文件不存在")

	path := dlWriteDownloadFile(t, filepath.Join("dlstream", "dlmovie.mkv"), []byte("DL-STREAM-CONTENT"))
	w := doRaw(t, e, "GET", "/api/download/stream_video?file_path="+urlEncode(path), "")
	if w.Code != http.StatusOK {
		t.Fatalf("stream code = %d body=%q", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != "DL-STREAM-CONTENT" {
		t.Fatalf("streamed bytes = %q", got)
	}
	// 容器按实际编码决定后，MKV 也要给出正确的 MIME，否则浏览器只会下载。
	if ct := w.Header().Get("Content-Type"); ct != "video/x-matroska" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if ra := w.Header().Get("Accept-Ranges"); ra != "bytes" {
		t.Fatalf("Accept-Ranges = %q", ra)
	}
}

func TestDLStreamDanmaku(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/stream_danmaku", "", 400, "error"), "缺少 file_path 或 bvid 参数")

	// 前端传的是视频成品路径，sidecar 命名由后端解析。
	video := dlWriteDownloadFile(t, filepath.Join("dldm", "DL弹幕视频_BVDLDM00000001.mkv"), []byte("v"))
	assContent := `[Script Info]
ScriptType: v4.00+

[Events]
Dialogue: 0,0:00:01.00,0:00:05.00,Default,,0,0,0,,DL测试弹幕
`
	dlWriteDownloadFile(t, filepath.Join("dldm", "DL弹幕视频_BVDLDM00000001.danmaku.ass"), []byte(assContent))
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(dlDownloadsDir(), "dldm")) })

	w := doRaw(t, e, "GET", "/api/download/stream_danmaku?file_path="+urlEncode(video), "")
	if w.Code != http.StatusOK {
		t.Fatalf("danmaku code = %d body=%q", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("content-type = %q", ct)
	}
	if !strings.Contains(w.Body.String(), "DL测试弹幕") {
		t.Fatalf("danmaku body = %q", w.Body.String())
	}

	// 只给 BV 号也能检索到。
	byBv := doRaw(t, e, "GET", "/api/download/stream_danmaku?bvid=BVDLDM00000001", "")
	if byBv.Code != http.StatusOK || !strings.Contains(byBv.Body.String(), "DL测试弹幕") {
		t.Fatalf("by-bvid code=%d body=%q", byBv.Code, byBv.Body.String())
	}

	// 目录外路径与不存在的弹幕分别给出 403 与 404。
	outside := filepath.Join(t.TempDir(), "secret.ass")
	if err := os.WriteFile(outside, []byte("SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/stream_danmaku?file_path="+urlEncode(outside), "", 403, "error"), "无权访问该文件")

	quiet := dlWriteDownloadFile(t, filepath.Join("dldm", "DL无弹幕_BVDLDM00000002.mp4"), []byte("v"))
	expectMessage(t, expectStatus(t, e, "GET", "/api/download/stream_danmaku?file_path="+urlEncode(quiet), "", 404, "error"), "弹幕文件不存在")
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

// dlSSEPayloads 解析 SSE 帧：新契约下每条 data 都是 JSON 事件。
func dlSSEPayloads(t *testing.T, body string) []map[string]interface{} {
	t.Helper()
	var events []map[string]interface{}
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &payload); err != nil {
			t.Fatalf("SSE data 不是 JSON: %q (%v)", line, err)
		}
		events = append(events, payload)
	}
	if len(events) == 0 {
		t.Fatalf("SSE body 里没有任何事件: %q", body)
	}
	return events
}

// dlSSEText 把所有 log/done 事件的文本拼成一段，便于按关键词断言。
func dlSSEText(events []map[string]interface{}) string {
	var sb strings.Builder
	for _, ev := range events {
		if text, ok := ev["text"].(string); ok {
			sb.WriteString(text)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// dlSSELastDone 返回流末尾的 done 事件。
func dlSSELastDone(t *testing.T, events []map[string]interface{}) map[string]interface{} {
	t.Helper()
	last := events[len(events)-1]
	if last["type"] != "done" {
		t.Fatalf("流末尾应是 done 事件, got %v", last)
	}
	return last
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

	expectMessage(t, expectStatus(t, e, "POST", "/api/download/download_video", `{}`, 200, "error"), "缺少 url 参数")

	w = doRaw(t, e, "POST", "/api/download/download_video",
		`{"url":"https://example.com/DLYY0001","sessdata":"s","video_quality":116,"only_audio":true}`)
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content-type = %q", ct)
	}
	events := dlSSEPayloads(t, w.Body.String())
	texts := dlSSEText(events)
	if !strings.Contains(texts, "解析 https://example.com/DLYY0001 失败") {
		t.Fatalf("缺少解析失败日志: %q", texts)
	}
	done := dlSSELastDone(t, events)
	if done["text"] != "没有可下载的内容" {
		t.Fatalf("done = %v, want 没有可下载的内容", done)
	}
}

func TestDLBatchDownloadSSE(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	w := doRaw(t, e, "POST", "/api/download/batch_download", "[1,2")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON code = %d body=%q", w.Code, w.Body.String())
	}

	// 只有非空 bvid 会拼成下载链接，空 bvid 与缺 key 的条目直接跳过。
	w = doRaw(t, e, "POST", "/api/download/batch_download",
		`{"videos":[{"bvid":""},{"bvid":"DLYY0001"},{"title":"无bvid"}],"sessdata":"s","only_audio":false}`)
	events := dlSSEPayloads(t, w.Body.String())
	texts := dlSSEText(events)
	if !strings.Contains(texts, "解析 https://www.bilibili.com/video/DLYY0001 失败") {
		t.Fatalf("batch SSE 缺少解析失败日志: %q", texts)
	}
	if strings.Contains(texts, "video/\"\n") || strings.Contains(texts, "bilibili.com/video\n") {
		t.Fatalf("空 bvid 不应拼成链接: %q", texts)
	}
	if got := dlSSELastDone(t, events); got["text"] != "没有可下载的内容" {
		t.Fatalf("done = %v", got)
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

	fail := dlSSEPayloads(t, doRaw(t, e, "POST", "/api/download/download_user_videos", `{"user_id":"666"}`).Body.String())
	failTexts := dlSSEText(fail)
	if !strings.Contains(failTexts, "正在获取用户视频列表...") {
		t.Fatalf("list-failure stream missing stage log: %q", failTexts)
	}
	if got := dlSSELastDone(t, fail); !strings.Contains(got["text"].(string), "获取视频列表失败: API 返回错误码: -100") {
		t.Fatalf("done = %v", got)
	}
	if strings.Contains(failTexts, "共找到") {
		t.Fatalf("failed listing must not continue: %q", failTexts)
	}

	ok := dlSSEPayloads(t, doRaw(t, e, "POST", "/api/download/download_user_videos", `{"user_id":"123","only_audio":true}`).Body.String())
	okTexts := dlSSEText(ok)
	if !strings.Contains(okTexts, "共找到 9 个视频") {
		t.Fatalf("success stream = %q", okTexts)
	}
	if !strings.Contains(okTexts, "解析 https://www.bilibili.com/video/DLYY0002 失败") {
		t.Fatalf("per-video 解析日志缺失: %q", okTexts)
	}
	if strings.Count(okTexts, "解析 ") != 1 {
		t.Fatalf("空 bvid 条目应被跳过: %q", okTexts)
	}
	if got := dlSSELastDone(t, ok); got["text"] != "没有可下载的内容" {
		t.Fatalf("done = %v", got)
	}
}

func TestDLDownloadFavoritesSSE(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	w := doRaw(t, e, "POST", "/api/download/download_favorites", `{"user_id":`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON code = %d", w.Code)
	}

	expectMessage(t, expectStatus(t, e, "POST", "/api/download/download_favorites", `{}`, 200, "error"), "缺少 fid 参数")

	// fid 以数字到达同样可用；本地未同步且无 SESSDATA 时在线回退被提前拒绝，
	// 不会发出任何真实请求。
	w = doRaw(t, e, "POST", "/api/download/download_favorites", `{"fid":9007199254740991}`)
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content-type = %q body=%q", ct, w.Body.String())
	}
	events := dlSSEPayloads(t, w.Body.String())
	texts := dlSSEText(events)
	if !strings.Contains(texts, "正在获取收藏夹视频列表...") {
		t.Fatalf("missing stage log: %q", texts)
	}
	if got := dlSSELastDone(t, events); got["event"] != "error" ||
		!strings.Contains(got["text"].(string), "未同步到本地") {
		t.Fatalf("done = %v", got)
	}
}

func TestDLCollectionRoutes(t *testing.T) {
	e := newAPI(t, RegisterDownloadRoutes)

	expectMessage(t, expectStatus(t, e, "GET", "/api/collection/check_collection", "", 400, "error"), "缺少 url 参数")
	expectMessageContains(t, expectStatus(t, e, "GET", "/api/collection/check_collection?url=https%3A%2F%2Fwww.bilibili.com%2Fvideo%2FDLYY0003", "", 200, "error"), "无法解析链接")

	w := doRaw(t, e, "POST", "/api/collection/download_collection", "{oops")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "参数错误") {
		t.Fatalf("bad JSON: code=%d body=%q", w.Code, w.Body.String())
	}

	expectMessage(t, expectStatus(t, e, "POST", "/api/collection/download_collection", `{}`, 200, "error"), "缺少 url 参数")

	w = doRaw(t, e, "POST", "/api/collection/download_collection", `{"url":"https://www.bilibili.com/video/DLYY0003"}`)
	events := dlSSEPayloads(t, w.Body.String())
	texts := dlSSEText(events)
	if !strings.Contains(texts, "解析 https://www.bilibili.com/video/DLYY0003 失败") {
		t.Fatalf("collection SSE missing resolve log: %q", texts)
	}
	if got := dlSSELastDone(t, events); got["text"] != "没有可下载的内容" {
		t.Fatalf("done = %v", got)
	}
}
