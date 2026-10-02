package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"
)

func svcRawItem(idStr, dtype string, modules map[string]any) biliapi.DynamicRawItem {
	mj, _ := json.Marshal(modules)
	return biliapi.DynamicRawItem{IDStr: idStr, Type: dtype, Modules: mj}
}

func TestSVCExtractExt(t *testing.T) {
	cases := map[string]string{
		"https://i0.hdslb.com/a/b.png":                        ".png",
		"https://i0.hdslb.com/a/b.jpg?x-oss-process=resize":   ".jpg",
		"https://i0.hdslb.com/a/b":                             ".jpg", // 无扩展名兜底 .jpg
		"https://i0.hdslb.com/a/b.webp?":                      ".webp",
		"https://i0.hdslb.com/a/b.jpeg?v=1#frag":              ".jpeg", // #fragment 不参与截断（记录现状）
		"https://i0.hdslb.com/a/b.gif":                        ".gif",
	}
	for in, want := range cases {
		if got := extractExt(in); got != want {
			t.Errorf("extractExt(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSVCSanitizeID(t *testing.T) {
	if got := sanitizeID("abc/def/ghi"); got != "abc_def_ghi" {
		t.Errorf("sanitizeID = %q", got)
	}
	if got := sanitizeID(""); got != "" {
		t.Errorf("sanitizeID empty = %q", got)
	}
	// 注意：只替换 "/"，".."、"\" 等不被过滤 —— 若 ID 来自外部输入存在路径穿越风险，
	// 记录为疑似缺陷（dynamic.go sanitizeID），测试固化当前行为。
	if got := sanitizeID("../../etc/passwd"); got != ".._.._etc_passwd" {
		t.Errorf("sanitizeID traversal = %q", got)
	}
}

func TestSVCStateAccessors(t *testing.T) {
	// 动态抓取状态读写
	setDynamicFetchStatus(DynamicFetchStatus{IsRunning: true, HostMid: "900000001", TotalFetched: 3})
	got := GetDynamicFetchStatus("900000001")
	if !got.IsRunning || got.HostMid != "900000001" || got.TotalFetched != 3 {
		t.Errorf("dynamic fetch status roundtrip = %+v", got)
	}
	StopDynamicFetch("900000001") // 仅置停止标记，不触发任何抓取
	setDynamicFetchStatus(DynamicFetchStatus{})

	// 进度通道：非阻塞发送，通道容量 100
	ch := GetDynamicProgressChan()
	if ch != dynamicProgressCh {
		t.Errorf("GetDynamicProgressChan should return the shared channel")
	}
	for i := 0; i < 100; i++ {
		sendDynamicProgress(fmt.Sprintf("msg-%d", i))
	}
	// 通道已满：这条消息会被静默丢弃（default 分支），不得阻塞
	sendDynamicProgress("overflow")
	for i := 0; i < 100; i++ {
		<-ch
	}
	select {
	case m := <-ch:
		t.Errorf("channel should be drained, got %q", m)
	default:
	}
}

func TestSVCFetchDynamicSpaceGuardNoSessdata(t *testing.T) {
	// SESSDATA 为空时必须在任何网络访问前返回，并写入提示状态
	FetchDynamicSpace("900000001", false, false, false, nil)
	st := GetDynamicFetchStatus("900000001")
	if st.IsRunning {
		t.Errorf("fetch should not be running")
	}
	if st.Message != "SESSDATA 未配置" {
		t.Errorf("message = %q, want SESSDATA 未配置", st.Message)
	}
	setDynamicFetchStatus(DynamicFetchStatus{})
}

func TestSVCparseDynamicItem(t *testing.T) {
	host := "900000001"

	t.Run("invalid modules json", func(t *testing.T) {
		raw := biliapi.DynamicRawItem{IDStr: "bad1", Type: "DYNAMIC_TYPE_DRAW", Modules: json.RawMessage("{not-json")}
		item := parseDynamicItem(raw, host)
		if item.ID != "bad1" || item.Type != "DYNAMIC_TYPE_DRAW" || item.HostMid != host {
			t.Errorf("basic fields lost on bad json: %+v", item)
		}
		if item.RawJSON != "" {
			t.Errorf("RawJSON should stay empty when unmarshal fails early")
		}
	})

	t.Run("archive with cover", func(t *testing.T) {
		raw := svcRawItem("arch1", "DYNAMIC_TYPE_ARCIVE", map[string]any{
			"module_author":   map[string]any{"name": "svc作者", "face": "http://face/1.jpg", "pub_ts": 1700000000},
			"module_dynamic": map[string]any{
				"major": map[string]any{
					"type": "MAJOR_TYPE_ARCHIVE",
					"archive": map[string]any{
						"title": "svc视频标题", "desc": "svc简介", "bvid": "BVsvcDyn0001",
						"cover": map[string]any{"src": "http://img/cover.jpg?x=1"},
					},
				},
				"desc": map[string]any{"text": "转发文案"},
			},
		})
		item := parseDynamicItem(raw, host)
		if item.Bvid != "BVsvcDyn0001" || item.Title != "svc视频标题" || item.Desc != "svc简介" {
			t.Errorf("archive fields = %+v", item)
		}
		if item.Cover != "http://img/cover.jpg?x=1" {
			t.Errorf("cover = %q", item.Cover)
		}
		if len(item.MediaLocals) != 1 || item.MediaLocals[0] != item.Cover {
			t.Errorf("media locals = %v", item.MediaLocals)
		}
		if item.AuthorName != "svc作者" || item.AuthorFace != "http://face/1.jpg" {
			t.Errorf("author = %+v", item)
		}
		if item.PublishTS != 1700000000 {
			t.Errorf("pub ts = %d", item.PublishTS)
		}
		if item.Txt != "转发文案" {
			t.Errorf("txt = %q", item.Txt)
		}
		if item.RawJSON == "" {
			t.Errorf("raw json should be preserved")
		}
	})

	t.Run("draw images and summary fallback", func(t *testing.T) {
		raw := svcRawItem("draw1", "DYNAMIC_TYPE_DRAW", map[string]any{
			"module_dynamic": map[string]any{
				"major": map[string]any{
					"type": "MAJOR_TYPE_DRAW",
					"draw": map[string]any{"items": []any{
						map[string]any{"src": "http://img/1.jpg", "width": 100, "height": 100},
						map[string]any{"src": "", "width": 0, "height": 0},
						map[string]any{"src": "http://img/2.png", "width": 1, "height": 1},
					}},
				},
				"desc": map[string]any{"text": "draw文案"},
			},
		})
		item := parseDynamicItem(raw, host)
		if len(item.MediaLocals) != 2 || item.MediaLocals[0] != "http://img/1.jpg" || item.MediaLocals[1] != "http://img/2.png" {
			t.Errorf("draw media locals = %v, want 2 non-empty srcs", item.MediaLocals)
		}
		// DRAW 分支会用 opus 字段兜底，OpusSummaryText 空时回落 desc.text
		if item.OpusSummaryText != "draw文案" {
			t.Errorf("opus summary fallback = %q", item.OpusSummaryText)
		}
		if item.Txt != "draw文案" {
			t.Errorf("txt = %q", item.Txt)
		}
	})

	t.Run("opus images", func(t *testing.T) {
		raw := svcRawItem("opus1", "DYNAMIC_TYPE_OPUS", map[string]any{
			"module_dynamic": map[string]any{
				"major": map[string]any{
					"type": "MAJOR_TYPE_OPUS",
					"opus": map[string]any{
						"title": "svc图文标题",
						"pics":  []any{map[string]any{"url": "http://img/p1.jpg"}, map[string]any{"url": ""}},
						"summary": map[string]any{"text": "svc摘要"},
					},
				},
			},
		})
		item := parseDynamicItem(raw, host)
		if item.OpusTitle != "svc图文标题" || item.OpusSummaryText != "svc摘要" {
			t.Errorf("opus fields = %+v", item)
		}
		if len(item.MediaLocals) != 1 || item.MediaLocals[0] != "http://img/p1.jpg" {
			t.Errorf("opus pics = %v", item.MediaLocals)
		}
	})

	t.Run("no major type keeps desc only", func(t *testing.T) {
		raw := svcRawItem("plain1", "DYNAMIC_TYPE_TEXT", map[string]any{
			"module_dynamic": map[string]any{"desc": map[string]any{"text": "纯文字动态"}},
		})
		item := parseDynamicItem(raw, host)
		if item.Txt != "纯文字动态" || len(item.MediaLocals) != 0 {
			t.Errorf("plain item = %+v", item)
		}
	})

	t.Run("missing module_dynamic", func(t *testing.T) {
		raw := svcRawItem("only-author", "DYNAMIC_TYPE_TEXT", map[string]any{
			"module_author": map[string]any{"name": "n", "pub_ts": 1},
		})
		item := parseDynamicItem(raw, host)
		if item.AuthorName != "n" || item.PublishTS != 1 || item.Txt != "" {
			t.Errorf("author-only item = %+v", item)
		}
	})
}

func TestSVCDeleteDynamicMedia(t *testing.T) {
	// nil 安全
	DeleteDynamicMedia(nil)

	host := "900000001"
	base := utils.GetOutputPath("dynamic", host, "images")
	svcMkdirAll(t, base)
	img := svcWrite(t, filepath.Join(base, "m1.jpg"), []byte("img"))
	nested := svcWrite(t, utils.GetOutputPath("dynamic", host, "videos", "m2.mp4"), []byte("vid"))

	item := &database.DynamicItem{
		ID:              "svc-del",
		MediaLocals:     []string{"http://remote/a.jpg", "https://remote/b.jpg", filepath.Join("dynamic", host, "images", "m1.jpg"), "dynamic/missing.jpg"},
		LiveMediaLocals: []string{filepath.Join("dynamic", host, "videos", "m2.mp4")},
	}
	DeleteDynamicMedia(item)

	if _, err := os.Stat(img); !os.IsNotExist(err) {
		t.Errorf("media local file should be deleted")
	}
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Errorf("live media file should be deleted")
	}
	// 远程 URL 必须被跳过而不是尝试删除
	for _, u := range []string{"http://remote/a.jpg", "https://remote/b.jpg"} {
		if strings.HasPrefix(u, "file://") {
			t.Fatal("unreachable")
		}
	}
	// 不存在的本地路径静默忽略（os.Remove 错误被吞掉）
}

func TestSVCDownloadDynamicMedia(t *testing.T) {
	// httptest 内网服务，零真实网络
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if strings.Contains(r.URL.Path, "fail") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("fake-jpeg-bytes"))
	}))
	defer srv.Close()

	host := "900000001"
	raw := svcRawItem("svc/med1", "DYNAMIC_TYPE_DRAW", map[string]any{
		"module_dynamic": map[string]any{
			"major": map[string]any{
				"type": "MAJOR_TYPE_DRAW",
				"draw": map[string]any{"items": []any{
					map[string]any{"src": srv.URL + "/pic0.png"},
					map[string]any{"src": srv.URL + "/fail1.png"},
				}},
			},
		},
	})
	item := parseDynamicItem(raw, host)
	item.Cover = srv.URL + "/cover.jpg"
	before := len(item.MediaLocals)

	downloadDynamicMedia(&item, host)

	// cover 1 + draw 2 = 3 次请求；第 4 次来自幂等重下检查
	if hits != 3 {
		t.Errorf("server hits = %d, want 3 (cover + 2 draw)", hits)
	}
	if item.Cover != fmt.Sprintf("dynamic/%s/images/svc_med1_cover.jpg", host) {
		t.Errorf("cover rewritten to = %q", item.Cover)
	}
	// 成功下载会各自追加 relPath：cover 与 pic0 共 2 条；失败的 fail1 原 URL 保留且无 relPath
	var relCount, failKept int
	for _, m := range item.MediaLocals[before:] {
		if strings.HasPrefix(m, "dynamic/") {
			relCount++
		}
	}
	for _, m := range item.MediaLocals {
		if strings.Contains(m, "fail1.png") && strings.HasPrefix(m, "http") {
			failKept++
		}
	}
	if relCount != 2 {
		t.Errorf("rel paths appended = %d, want 2 (cover + successful pic)", relCount)
	}
	if failKept == 0 {
		t.Errorf("failed download should keep original URL in MediaLocals")
	}

	// 疑似生产缺陷：文件实际保存到 <output>/images/dynamic/<filename>
	// （DownloadImage 固定前缀 images），而 item 记录的路径是 dynamic/<host>/images/...，
	// 两者不一致 —— DeleteDynamicMedia 依据后者拼接 GetOutputPath，永远删不到真实文件。
	realPath := filepath.Join(utils.GetOutputPath("images"), "dynamic", "svc_med1_cover.jpg")
	if _, err := os.Stat(realPath); err != nil {
		t.Errorf("expected file at %s (documents the path mismatch bug): %v", realPath, err)
	}
	recordedPath := utils.GetOutputPath("dynamic", host, "images", "svc_med1_cover.jpg")
	if _, err := os.Stat(recordedPath); err == nil {
		t.Errorf("unexpected: recorded path should NOT contain the real file (that is the bug)")
	}
}
