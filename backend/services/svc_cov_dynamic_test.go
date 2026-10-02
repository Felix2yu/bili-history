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
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"
)

// svcCov 动态抓取命名空间：hostMid 9000002033..9000002039，动态 id 前缀 svcCov-。
const (
	svcCovDynHostSave   = "9000002033"
	svcCovDynHostFilter = "9000002034"
	svcCovDynHostMedia  = "9000002035"
	svcCovDynHostIncr   = "9000002036"
	svcCovDynHostErr    = "9000002037"
	svcCovDynHostStop   = "9000002038"
	svcCovDynHostEmpty  = "9000002039"
)

// svcCovDrainProgress 清空动态进度 channel，避免影响 TestSVCStateAccessors 的容量断言。
func svcCovDrainProgress() {
	for {
		select {
		case <-dynamicProgressCh:
		default:
			return
		}
	}
}

func svcCovCleanupDynamicHost(t *testing.T, hostMid string) {
	t.Helper()
	t.Cleanup(func() {
		if err := database.DeleteDynamicSpace(hostMid); err != nil {
			t.Errorf("delete dynamic space %s: %v", hostMid, err)
		}
		svcCovDrainProgress()
		setDynamicFetchStatus(DynamicFetchStatus{})
	})
}

// svcCovStartDynamicServer 起 wbi-nav + feed/space 桩服务并按 offset 分发。
// feed 返回 (code, message, data)；code!=0 时 data 忽略。
func svcCovStartDynamicServer(t *testing.T, feed func(offset string) (int, string, any)) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/x/web-interface/nav", func(w http.ResponseWriter, r *http.Request) {
		svcCovWriteBiliJSON(t, w, 0, "OK", map[string]any{"wbi_img": map[string]any{
			"img_url": "https://wbi.invalid/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
			"sub_url": "https://wbi.invalid/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png",
		}})
	})
	mux.HandleFunc("/x/polymer/web-dynamic/v1/feed/space", func(w http.ResponseWriter, r *http.Request) {
		code, msg, data := feed(r.URL.Query().Get("offset"))
		svcCovWriteBiliJSON(t, w, code, msg, data)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	svcCovSetURL(t, &biliapi.WbiNavURL, srv.URL+"/x/web-interface/nav")
	svcCovSetURL(t, &biliapi.DynamicSpaceURL, srv.URL+"/x/polymer/web-dynamic/v1/feed/space")
	return srv
}

// svcCovDynItem 构造一条动态原始 item（modules 内含 author + dynamic）。
func svcCovDynItem(id, dtype, author string, pubTS int64, major map[string]any, descText string) map[string]any {
	modules := map[string]any{}
	if author != "" || pubTS > 0 {
		modules["module_author"] = map[string]any{"name": author, "face": "https://face.invalid/" + id + ".jpg", "pub_ts": pubTS}
	}
	dyn := map[string]any{}
	if major != nil {
		dyn["major"] = major
	}
	if descText != "" {
		dyn["desc"] = map[string]any{"text": descText}
	}
	if len(dyn) > 0 {
		modules["module_dynamic"] = dyn
	}
	return map[string]any{"id_str": id, "type": dtype, "modules": modules}
}

func svcCovDynPage(items []map[string]any, hasMore bool, offset string) biliapi.DynamicSpaceResponse {
	return biliapi.DynamicSpaceResponse{HasMore: hasMore, Offset: offset, Items: svcCovRawItems(items)}
}

func svcCovRawItems(items []map[string]any) []biliapi.DynamicRawItem {
	raw := make([]biliapi.DynamicRawItem, 0, len(items))
	for _, it := range items {
		raw = append(raw, biliapi.DynamicRawItem{
			IDStr:   it["id_str"].(string),
			Type:    it["type"].(string),
			Modules: svcCovMarshalModules(it["modules"]),
		})
	}
	return raw
}

func svcCovMarshalModules(modules any) json.RawMessage {
	b, err := json.Marshal(modules)
	if err != nil {
		panic(err)
	}
	return b
}

// svcCovWaitDynamicDone 等待抓取 goroutine 收敛（IsRunning=false）。
func svcCovWaitDynamicDone(t *testing.T) DynamicFetchStatus {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		st := GetDynamicFetchStatus("")
		if !st.IsRunning {
			return st
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("dynamic fetch did not finish within 20s")
	return DynamicFetchStatus{}
}

func TestSVCCovFetchDynamicSpace(t *testing.T) {
	svcCovDrainProgress()
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	avItem := svcCovDynItem("svcCov-av-1", "DYNAMIC_TYPE_AV", "svcCovUP", 1750000000, map[string]any{
		"type": "MAJOR_TYPE_ARCHIVE",
		"archive": map[string]any{
			"title": "svcCov 视频动态", "desc": "svcCov 简介", "bvid": "BVsvcCovDyn1",
			"cover": map[string]any{"src": "https://img.invalid/cov.jpg"},
		},
	}, "svcCov 转发文案")
	drawItem := svcCovDynItem("svcCov-draw-1", "DYNAMIC_TYPE_DRAW", "svcCovUP", 1750000001, map[string]any{
		"type": "MAJOR_TYPE_DRAW",
		"draw": map[string]any{"items": []any{map[string]any{"src": "https://img.invalid/d1.jpg", "width": 10, "height": 10}}},
	}, "svcCov 图片文案")
	liveItem := svcCovDynItem("svcCov-live-1", "DYNAMIC_TYPE_LIVE", "svcCovUP", 1750000002, nil, "svcCov 直播")
	textItem := svcCovDynItem("svcCov-text-2", "DYNAMIC_TYPE_TEXT", "svcCovUP", 1749999999, nil, "svcCov 纯文字")

	// 1) 全量抓取 + 保存 + 类型过滤（页 1 跳过 live，页 2 结束）
	t.Run("full fetch saves to db", func(t *testing.T) {
		svcCovCleanupDynamicHost(t, svcCovDynHostSave)
		p1 := svcCovDynPage([]map[string]any{avItem, drawItem, liveItem}, true, "o1")
		p2 := svcCovDynPage([]map[string]any{textItem}, false, "")
		svcCovStartDynamicServer(t, func(offset string) (int, string, any) {
			if offset == "" {
				return 0, "OK", p1
			}
			return 0, "OK", p2
		})

		FetchDynamicSpace(svcCovDynHostSave, false, true, false,
			[]string{"DYNAMIC_TYPE_AV", "DYNAMIC_TYPE_DRAW", "DYNAMIC_TYPE_TEXT"})
		st := svcCovWaitDynamicDone(t)
		if st.TotalFetched != 3 || st.TotalPages != 2 {
			t.Errorf("status = %+v, want fetched=3 pages=2", st)
		}
		if st.Message != "抓取完成，共 3 条动态" {
			t.Errorf("message = %q", st.Message)
		}
		count, err := database.GetDynamicCount(svcCovDynHostSave)
		if err != nil || count != 3 {
			t.Fatalf("db count = %d (err %v), want 3", count, err)
		}
		item, err := database.GetDynamicByID("svcCov-av-1")
		if err != nil || item == nil {
			t.Fatalf("read av: %v", err)
		}
		if item.Bvid != "BVsvcCovDyn1" || item.Title != "svcCov 视频动态" || item.Desc != "svcCov 简介" {
			t.Errorf("av fields = %+v", item)
		}
		if item.PublishTS != 1750000000 || item.AuthorName != "svcCovUP" || item.Txt != "svcCov 转发文案" {
			t.Errorf("av author/txt = %+v", item)
		}
		if len(item.MediaLocals) == 0 || item.MediaLocals[0] != "https://img.invalid/cov.jpg" {
			t.Errorf("av media locals = %v", item.MediaLocals)
		}
		// 被过滤的 live 动态不得入库
		if got, err := database.GetDynamicByID("svcCov-live-1"); err == nil && got != nil {
			t.Errorf("filtered dynamic should not be saved, got %+v", got)
		}
		hosts, err := database.GetDynamicHosts(200, 0)
		if err != nil {
			t.Fatalf("hosts: %v", err)
		}
		var found *database.DynamicHost
		for i := range hosts {
			if hosts[i].HostMid == svcCovDynHostSave {
				found = &hosts[i]
			}
		}
		if found == nil {
			t.Fatal("host stats row missing")
		}
		if found.ItemCount != 3 || found.CoreCount != 2 {
			t.Errorf("host stats = %+v, want items=3 core=2", found)
		}
		if found.LastPublishTS != 1750000001 || found.UpName != "svcCovUP" {
			t.Errorf("host ts/name = %+v", found)
		}
	})

	// 2) 整页都被类型过滤（saveToDB=false 路径的筛选提示 + 汇总跳过数）
	t.Run("all filtered skips page", func(t *testing.T) {
		svcCovCleanupDynamicHost(t, svcCovDynHostFilter)
		p1 := svcCovDynPage([]map[string]any{avItem, drawItem}, false, "")
		svcCovStartDynamicServer(t, func(offset string) (int, string, any) {
			if offset != "" {
				t.Errorf("unexpected offset %q", offset)
			}
			return 0, "OK", p1
		})
		FetchDynamicSpace(svcCovDynHostFilter, false, false, false, []string{"DYNAMIC_TYPE_TEXT"})
		st := svcCovWaitDynamicDone(t)
		if st.TotalFetched != 0 || st.TotalPages != 1 {
			t.Errorf("status = %+v", st)
		}
		if st.Message != "抓取完成，共 0 条动态，跳过 2 条" {
			t.Errorf("message = %q", st.Message)
		}
		if count, err := database.GetDynamicCount(svcCovDynHostFilter); err != nil || count != 0 {
			t.Errorf("count = %d (err %v), want 0", count, err)
		}
	})

	// 3) 不入库但下载媒体（saveMedia=true → downloadDynamicMedia 调用点 + draw 图片落盘）
	t.Run("media download without save", func(t *testing.T) {
		svcCovCleanupDynamicHost(t, svcCovDynHostMedia)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("svcCov-png-bytes"))
		}))
		t.Cleanup(srv.Close)
		mediaItem := svcCovDynItem("svcCov-media-1", "DYNAMIC_TYPE_DRAW", "svcCovUP", 1750000100, map[string]any{
			"type": "MAJOR_TYPE_DRAW",
			"draw": map[string]any{"items": []any{map[string]any{"src": srv.URL + "/pics/p1.png"}}},
		}, "svcCov 图片")
		p1 := svcCovDynPage([]map[string]any{mediaItem}, false, "")
		svcCovStartDynamicServer(t, func(offset string) (int, string, any) { return 0, "OK", p1 })

		FetchDynamicSpace(svcCovDynHostMedia, false, false, true, nil)
		st := svcCovWaitDynamicDone(t)
		if st.Message != "抓取完成，共 1 条动态" || st.TotalFetched != 1 {
			t.Errorf("status = %+v", st)
		}
		if count, err := database.GetDynamicCount(svcCovDynHostMedia); err != nil || count != 0 {
			t.Errorf("count = %d (err %v), want 0 (saveToDB=false)", count, err)
		}
		realPath := filepath.Join(utils.GetOutputPath("images"), "dynamic", "svcCov-media-1_0.png")
		if _, err := os.Stat(realPath); err != nil {
			t.Errorf("draw image should be downloaded to %s: %v", realPath, err)
		}
	})

	// 4) 增量模式：库中已有动态 → 遇到已存在 id 即停止并保存此前收集的新条目
	t.Run("incremental stops at existing", func(t *testing.T) {
		svcCovCleanupDynamicHost(t, svcCovDynHostIncr)
		if _, err := database.SaveDynamics(svcCovDynHostIncr, []database.DynamicItem{{
			ID: "svcCov-inc-old", Type: "DYNAMIC_TYPE_AV", HostMid: svcCovDynHostIncr, PublishTS: 1740000000,
		}}); err != nil {
			t.Fatalf("pre-save: %v", err)
		}
		newItem := svcCovDynItem("svcCov-inc-new", "DYNAMIC_TYPE_AV", "svcCovUP", 1760000000, map[string]any{
			"type":    "MAJOR_TYPE_ARCHIVE",
			"archive": map[string]any{"title": "svcCov 增量视频", "bvid": "BVsvcCovDyn2", "cover": map[string]any{"src": ""}},
		}, "svcCov 增量")
		oldCopy := svcCovDynItem("svcCov-inc-old", "DYNAMIC_TYPE_AV", "svcCovUP", 1740000000, nil, "svcCov 旧")
		p1 := svcCovDynPage([]map[string]any{newItem, oldCopy}, true, "o1")
		svcCovStartDynamicServer(t, func(offset string) (int, string, any) {
			if offset != "" {
				t.Errorf("incremental mode must stop before page %q", offset)
				return -101, "should not fetch", nil
			}
			return 0, "OK", p1
		})
		FetchDynamicSpace(svcCovDynHostIncr, false, true, false, nil)
		st := svcCovWaitDynamicDone(t)
		if st.TotalFetched != 1 || st.TotalPages != 1 {
			t.Errorf("status = %+v, want fetched=1 pages=1", st)
		}
		count, err := database.GetDynamicCount(svcCovDynHostIncr)
		if err != nil || count != 2 {
			t.Errorf("count = %d (err %v), want 2 (old + 1 new)", count, err)
		}
		if item, err := database.GetDynamicByID("svcCov-inc-new"); err != nil || item == nil || item.Title != "svcCov 增量视频" {
			t.Errorf("new dynamic not saved properly: %+v err=%v", item, err)
		}
	})

	// 5) API 错误码 -101（*biliapi.ApiError）→ 抓取出错并终止
	t.Run("api error -101 stops fetch", func(t *testing.T) {
		svcCovCleanupDynamicHost(t, svcCovDynHostErr)
		svcCovStartDynamicServer(t, func(offset string) (int, string, any) {
			return -101, "账号未登录", nil
		})
		FetchDynamicSpace(svcCovDynHostErr, false, true, false, nil)
		st := svcCovWaitDynamicDone(t)
		if !strings.Contains(st.Message, "抓取出错") || !strings.Contains(st.Message, "code=-101") {
			t.Errorf("message = %q", st.Message)
		}
		if st.TotalFetched != 0 || st.TotalPages != 1 {
			t.Errorf("status = %+v", st)
		}
	})

	// 6) 用户停止：第一页响应时置位 stop 标记 → 第二页前中断
	t.Run("stop flag interrupts loop", func(t *testing.T) {
		svcCovCleanupDynamicHost(t, svcCovDynHostStop)
		fetchedFirst := false
		p1 := svcCovDynPage([]map[string]any{avItem}, true, "o1")
		svcCovStartDynamicServer(t, func(offset string) (int, string, any) {
			if offset == "" {
				if !fetchedFirst {
					fetchedFirst = true
					StopDynamicFetch(svcCovDynHostStop)
				}
				return 0, "OK", p1
			}
			return -6, "page o1 must not be requested", nil
		})
		FetchDynamicSpace(svcCovDynHostStop, false, true, false, nil)
		st := svcCovWaitDynamicDone(t)
		if !strings.Contains(st.Message, fmt.Sprintf("已停止，共获取 %d 条动态", 1)) {
			t.Errorf("message = %q", st.Message)
		}
		if st.TotalPages != 1 {
			t.Errorf("TotalPages = %d, want 1", st.TotalPages)
		}
		if count, err := database.GetDynamicCount(svcCovDynHostStop); err != nil || count != 1 {
			t.Errorf("count = %d (err %v), want 1 (page 1 已入库后才发现停止)", count, err)
		}
	})

	// 7) 首页即空 → “没有更多动态”分支
	t.Run("empty first page completes", func(t *testing.T) {
		svcCovCleanupDynamicHost(t, svcCovDynHostEmpty)
		p1 := svcCovDynPage(nil, true, "o1")
		svcCovStartDynamicServer(t, func(offset string) (int, string, any) { return 0, "OK", p1 })
		FetchDynamicSpace(svcCovDynHostEmpty, false, true, false, nil)
		st := svcCovWaitDynamicDone(t)
		if st.Message != "抓取完成，共 0 条动态" || st.TotalPages != 1 {
			t.Errorf("status = %+v", st)
		}
	})

	svcCovDrainProgress()
}
