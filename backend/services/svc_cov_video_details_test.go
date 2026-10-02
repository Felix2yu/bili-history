package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
)

// svcCov 视频详情命名空间：bvid 前缀 BVsvcCovV / BVsvcCovSt / BVsvcCovD，年份表 2033。

// svcCovVideoInfoPayload 返回与 biliapi.VideoInfo 对齐的 data 载荷。
func svcCovVideoInfoPayload(bvid string) map[string]any {
	return map[string]any{
		"bvid": bvid, "aid": 4200001, "videos": 2, "tid": 188, "tname": "svcCov分区",
		"copyright": 1, "pic": "https://pic.invalid/x.jpg", "title": "svcCov 视频 " + bvid,
		"pubdate": 1700000000, "ctime": 1699999000, "desc": "svcCov 简介", "duration": 233,
		"owner": map[string]any{"mid": 8800777, "name": "svcCovUP主", "face": "https://face.invalid/own.jpg"},
		"stat": map[string]any{"view": 12345, "danmaku": 66, "reply": 77, "favorite": 88,
			"coin": 99, "share": 11, "like": 222},
	}
}

// svcCovVideoServer 起 /x/web-interface/view 桩：按 bvid 参数分发成功/错误码。
func svcCovVideoServer(t *testing.T, tname string, delay time.Duration) *httptest.Server {
	t.Helper()
	failCodes := map[string]int{
		"BVsvcCovVbad1":  -412,
		"BVsvcCovErr6":   -6,
		"BVsvcCovErr101": -101,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/x/web-interface/view", func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}
		bvid := r.URL.Query().Get("bvid")
		if code, bad := failCodes[bvid]; bad {
			msg := "svcCov 失败"
			if code == -412 {
				msg = "请求被拦截"
			}
			svcCovWriteBiliJSON(t, w, code, msg, nil)
			return
		}
		payload := svcCovVideoInfoPayload(bvid)
		if tname != "" {
			payload["tname"] = tname
		}
		svcCovWriteBiliJSON(t, w, 0, "OK", payload)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	svcCovSetURL(t, &biliapi.VideoInfoURL, srv.URL+"/x/web-interface/view")
	return srv
}

func svcCovWaitVideoDetailDone(t *testing.T) models.VideoDetailProgress {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		p := GetVideoDetailProgress()
		if !p.IsProcessing {
			return p
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("video detail batch did not finish within 30s")
	return models.VideoDetailProgress{}
}

// svcCovDeleteCovVideoRows 清理本命名空间写入的 video_base_info / invalid_videos。
func svcCovDeleteCovVideoRows(t *testing.T) {
	t.Helper()
	conn := svcConn()
	t.Cleanup(func() {
		for _, q := range []string{
			"DELETE FROM video_base_info WHERE bvid LIKE 'BVsvcCovV%'",
			"DELETE FROM video_base_info WHERE bvid LIKE 'BVsvcCovSt%'",
			"DELETE FROM video_base_info WHERE bvid LIKE 'BVsvcCovD%'",
			"DELETE FROM invalid_videos WHERE bvid LIKE 'BVsvcCov%'",
		} {
			if _, err := conn.Exec(q); err != nil {
				t.Errorf("cleanup %q: %v", q, err)
			}
		}
		var exists int
		if err := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='bilibili_history_2033'").Scan(&exists); err != nil {
			t.Errorf("check 2033 table: %v", err)
		} else if exists == 1 {
			if _, err := conn.Exec("DELETE FROM bilibili_history_2033 WHERE bvid LIKE 'BVsvcCovD%'"); err != nil {
				t.Errorf("cleanup history D rows: %v", err)
			}
		}
		ResetVideoDetailProgress()
	})
}

func TestSVCCovFetchVideoDetailFromBili(t *testing.T) {
	svcCovVideoServer(t, "", 0)

	before := time.Now().Unix()
	video, err := FetchVideoDetailFromBili("BVsvcCovVFull")
	if err != nil {
		t.Fatalf("FetchVideoDetailFromBili: %v", err)
	}
	after := time.Now().Unix()
	if video.Bvid != "BVsvcCovVFull" || video.Aid != 4200001 || video.Videos != 2 {
		t.Errorf("video base = %+v", video)
	}
	if video.Tid != 188 || video.Tname != "svcCov分区" || video.Copyright != 1 || video.Duration != 233 {
		t.Errorf("video classify = %+v", video)
	}
	if video.Title != "svcCov 视频 BVsvcCovVFull" || video.Desc != "svcCov 简介" || video.Pic != "https://pic.invalid/x.jpg" {
		t.Errorf("video text = %+v", video)
	}
	if video.Pubdate != 1700000000 || video.Ctime != 1699999000 || video.Cid != 0 {
		t.Errorf("video times = %+v", video)
	}
	if video.OwnerMid != 8800777 || video.OwnerName != "svcCovUP主" || video.OwnerFace != "https://face.invalid/own.jpg" {
		t.Errorf("video owner = %+v", video)
	}
	if video.StatView != 12345 || video.StatDanmaku != 66 || video.StatReply != 77 || video.StatFavorite != 88 ||
		video.StatCoin != 99 || video.StatShare != 11 || video.StatLike != 222 {
		t.Errorf("video stats = %+v", video)
	}
	if video.FetchTime < before || video.FetchTime > after || video.UpdateTime != video.FetchTime {
		t.Errorf("fetch/update time = (%d,%d), want within [%d,%d]", video.FetchTime, video.UpdateTime, before, after)
	}

	// API 错误码分支（GetVideoInfo 返回普通错误而非 *ApiError）
	if _, err := FetchVideoDetailFromBili("BVsvcCovErr6"); err == nil ||
		!strings.Contains(err.Error(), "获取视频详情失败") || !strings.Contains(err.Error(), "code=-6") {
		t.Errorf("err6 = %v", err)
	}
	if _, err := FetchVideoDetailFromBili("BVsvcCovErr101"); err == nil || !strings.Contains(err.Error(), "code=-101") {
		t.Errorf("err101 = %v", err)
	}
}

// TestSVCCovBatchFetchVideoDetailsMixed 覆盖批量抓取的成功/失败计数、
// 进度持久化（percent/elapsed/status）与失效视频落库路径。
func TestSVCCovBatchFetchVideoDetailsMixed(t *testing.T) {
	ResetVideoDetailProgress()
	svcCovDeleteCovVideoRows(t)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })
	svcCovVideoServer(t, "", 0)

	bvids := []string{"BVsvcCovVok1", "BVsvcCovVbad1", "BVsvcCovVok2"}
	res, err := BatchFetchVideoDetails(bvids)
	if err != nil {
		t.Fatalf("BatchFetchVideoDetails: %v", err)
	}
	if res["status"] != "success" || res["total"] != 3 {
		t.Fatalf("res = %v", res)
	}
	p := svcCovWaitVideoDetailDone(t)
	if p.Status != "completed" || !p.IsComplete || p.IsStopped || p.IsProcessing {
		t.Errorf("final progress = %+v", p)
	}
	if p.TotalVideos != 3 || p.ProcessedVideos != 3 || p.SuccessCount != 2 || p.FailedCount != 1 {
		t.Errorf("counts = %+v", p)
	}
	if p.ProgressPercent != 100 {
		t.Errorf("ProgressPercent = %v, want 100", p.ProgressPercent)
	}
	if p.StartTime == 0 || p.LastUpdateTime < p.StartTime {
		t.Errorf("timestamps start=%d last=%d", p.StartTime, p.LastUpdateTime)
	}

	got, err := database.GetVideoBaseInfoByBvid("BVsvcCovVok1")
	if err != nil || got == nil {
		t.Fatalf("read ok1: %v", err)
	}
	if got.Title != "svcCov 视频 BVsvcCovVok1" || got.Tname != "svcCov分区" || got.OwnerName != "svcCovUP主" || got.StatView != 12345 {
		t.Errorf("ok1 row = %+v", got)
	}
	if missing, err := database.GetVideoBaseInfoByBvid("BVsvcCovVbad1"); err != nil || missing != nil {
		t.Errorf("failed bvid should not be in video_base_info, got %+v err=%v", missing, err)
	}
	var errMsg string
	var errCode int
	if err := svcConn().QueryRow("SELECT error_message, error_code FROM invalid_videos WHERE bvid='BVsvcCovVbad1'").Scan(&errMsg, &errCode); err != nil {
		t.Fatalf("read invalid row: %v", err)
	}
	if !strings.Contains(errMsg, "code=-412") {
		t.Errorf("invalid error_message = %q, want code=-412", errMsg)
	}
	// processBatchFetch 以 RecordInvalidVideo(bvid, err, 0) 记录，错误码固定 0
	if errCode != 0 {
		t.Errorf("invalid error_code = %d, want 0", errCode)
	}
}

// TestSVCCovBatchFetchStopped 覆盖 StopVideoDetailFetch 的中断路径：
// 已处理计数停在第一项，最终状态为 stopped 而非 completed。
func TestSVCCovBatchFetchStopped(t *testing.T) {
	ResetVideoDetailProgress()
	svcCovDeleteCovVideoRows(t)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })
	svcCovVideoServer(t, "", 40*time.Millisecond)

	if _, err := BatchFetchVideoDetails([]string{"BVsvcCovSt01", "BVsvcCovSt02", "BVsvcCovSt03"}); err != nil {
		t.Fatalf("BatchFetchVideoDetails: %v", err)
	}
	// 等第一项处理完（每项至少 40ms + 200ms sleep，轮询窗口足够）
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if GetVideoDetailProgress().ProcessedVideos >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := StopVideoDetailFetch(); err != nil {
		t.Fatalf("StopVideoDetailFetch: %v", err)
	}
	if p := GetVideoDetailProgress(); !p.IsStopped || p.Status != "stopping" {
		t.Fatalf("stopping state = %+v", p)
	}
	p := svcCovWaitVideoDetailDone(t)
	if p.Status != "stopped" || p.IsComplete {
		t.Errorf("stopped final = %+v", p)
	}
	if p.ProcessedVideos != 1 || p.SuccessCount != 1 {
		t.Errorf("processed = %+v, want 1/1", p)
	}
	if got, err := database.GetVideoBaseInfoByBvid("BVsvcCovSt01"); err != nil || got == nil {
		t.Errorf("first item should have been fetched: %+v err=%v", got, err)
	}
	if got, err := database.GetVideoBaseInfoByBvid("BVsvcCovSt02"); err != nil || got != nil {
		t.Errorf("second item should be skipped after stop, got %+v", got)
	}
}

// TestSVCCovBatchFetchFromHistory 覆盖从历史库派生 bvid 的增量抓取：
// skipExisting 过滤已抓取/失效项、defer 中 SyncHistoryTagNames 回填 tag_name、
// 以及“没有需要获取的视频”提前返回分支。
func TestSVCCovBatchFetchFromHistory(t *testing.T) {
	ResetVideoDetailProgress()
	svcCovDeleteCovVideoRows(t)
	svcWithNotifyConfig(t, func(c *config.Config) { c.SESSDATA = svcCovSessdata })

	d1 := svcAt(svcCovYearA, 7, 1, 10, 0)
	d2 := svcAt(svcCovYearA, 7, 2, 10, 0)
	d3 := svcAt(svcCovYearA, 7, 3, 10, 0)
	seed := []models.HistoryRecord{
		svcHistoryRecord("BVsvcCovD1", "svcCov 历史一", "archive", "", "科技", 8800901, "svcCov历史UP", d1, 100, 10),
		svcHistoryRecord("BVsvcCovD2", "svcCov 历史二", "archive", "", "科技", 8800902, "svcCov历史UP", d2, 100, 10),
		svcHistoryRecord("BVsvcCovD3", "svcCov 历史三", "archive", "", "科技", 8800903, "svcCov历史UP", d3, 100, 10),
	}
	seed[0].TagName = ""
	seed[1].TagName = ""
	seed[2].TagName = ""
	if err := svcSeedHistoryRows(svcCovYearA, seed...); err != nil {
		t.Fatalf("seed history: %v", err)
	}
	// D3 预置为失效视频 → skipExisting 过滤；且详情接口永不被调用
	if err := database.RecordInvalidVideo("BVsvcCovD3", "svcCov 固化失效", 0); err != nil {
		t.Fatalf("record invalid: %v", err)
	}

	allBvids, err := database.GetUniqueBvidsFromHistory()
	if err != nil {
		t.Fatalf("unique bvids: %v", err)
	}
	fetched, err := database.GetFetchedBvids()
	if err != nil {
		t.Fatalf("fetched bvids: %v", err)
	}
	skip := map[string]bool{"BVsvcCovD1": true, "BVsvcCovD2": true, "BVsvcCovD3": true}
	var added []string
	for _, b := range allBvids {
		if fetched[b] || skip[b] {
			continue
		}
		if err := database.UpsertVideoBaseInfo(svcVideoBaseInfo(b, "svcCov 占位", "svcCov占位分区")); err != nil {
			t.Fatalf("pre-fetch %s: %v", b, err)
		}
		added = append(added, b)
	}
	// 占位行不属于 BVsvcCovV/St/D 前缀，追加精确清理
	conn := svcConn()
	for _, b := range added {
		b := b
		t.Cleanup(func() {
			if _, err := conn.Exec("DELETE FROM video_base_info WHERE bvid = ?", b); err != nil {
				t.Errorf("cleanup placeholder %s: %v", b, err)
			}
		})
	}

	svcCovVideoServer(t, "svcCov回填分区", 0)
	res, err := BatchFetchFromHistory(true)
	if err != nil {
		t.Fatalf("BatchFetchFromHistory: %v", err)
	}
	if res["total"] != 2 {
		t.Errorf("total = %v, want 2 (only D1/D2 pending)", res["total"])
	}
	if res["history_total"] != len(allBvids) {
		t.Errorf("history_total = %v, want %d", res["history_total"], len(allBvids))
	}
	p := svcCovWaitVideoDetailDone(t)
	if p.Status != "completed" || p.SuccessCount != 2 || p.FailedCount != 0 {
		t.Fatalf("progress = %+v", p)
	}

	got, err := database.GetVideoBaseInfoByBvid("BVsvcCovD1")
	if err != nil || got == nil || got.Title != "svcCov 视频 BVsvcCovD1" {
		t.Fatalf("D1 detail = %+v err=%v", got, err)
	}
	if _, err := database.GetVideoBaseInfoByBvid("BVsvcCovD3"); err != nil {
		t.Errorf("read D3: %v", err)
	}

	// defer 中的 SyncHistoryTagNames 应从 video_base_info.tname 回填 D1/D2，
	// 而失效未抓取的 D3 保持空标签。
	var tag1, tag2, tag3 string
	if err := conn.QueryRow("SELECT tag_name FROM bilibili_history_2033 WHERE bvid='BVsvcCovD1'").Scan(&tag1); err != nil {
		t.Fatalf("read tag D1: %v", err)
	}
	if err := conn.QueryRow("SELECT tag_name FROM bilibili_history_2033 WHERE bvid='BVsvcCovD2'").Scan(&tag2); err != nil {
		t.Fatalf("read tag D2: %v", err)
	}
	if err := conn.QueryRow("SELECT tag_name FROM bilibili_history_2033 WHERE bvid='BVsvcCovD3'").Scan(&tag3); err != nil {
		t.Fatalf("read tag D3: %v", err)
	}
	if tag1 != "svcCov回填分区" || tag2 != "svcCov回填分区" {
		t.Errorf("backfilled tags = (%q,%q), want svcCov回填分区", tag1, tag2)
	}
	if tag3 != "" {
		t.Errorf("invalid video D3 tag should stay empty, got %q", tag3)
	}

	// 再跑一次：全部已抓取/失效 → 提前返回“没有需要获取的视频”
	res2, err := BatchFetchFromHistory(true)
	if err != nil {
		t.Fatalf("second BatchFetchFromHistory: %v", err)
	}
	if res2["total"] != 0 || !strings.Contains(res2["message"].(string), "没有需要获取的视频") {
		t.Errorf("res2 = %v", res2)
	}
}
