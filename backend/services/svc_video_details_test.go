package services

import (
	"strings"
	"testing"

	"bilibili-history-go/models"

	"bilibili-history-go/database"
)

func TestSVCVideoDetailProgressChan(t *testing.T) {
	ch1 := GetVideoDetailProgressChan()
	ch2 := GetVideoDetailProgressChan()
	if ch1 == nil {
		t.Fatal("chan nil")
	}
	if ch1 != ch2 {
		t.Errorf("chan should be memoized")
	}
}

func TestSVCVideoDetailProgressLifecycle(t *testing.T) {
	ResetVideoDetailProgress()
	p := GetVideoDetailProgress()
	if p.IsProcessing || p.IsComplete || p.IsStopped || p.Status != "idle" {
		t.Fatalf("initial reset state = %+v", p)
	}

	// 未在处理中时 Stop 报错
	if err := StopVideoDetailFetch(); err == nil || !strings.Contains(err.Error(), "没有正在进行") {
		t.Errorf("stop idle err = %v", err)
	}

	// 模拟处理中：Stop 置位
	setVideoDetailProgress(modelsProgressProcessing())
	if err := StopVideoDetailFetch(); err != nil {
		t.Fatalf("stop while processing: %v", err)
	}
	p = GetVideoDetailProgress()
	if !p.IsStopped || p.Status != "stopping" {
		t.Errorf("after stop = %+v", p)
	}

	ResetVideoDetailProgress()
	p = GetVideoDetailProgress()
	if p.IsStopped || p.Status != "idle" {
		t.Errorf("reset = %+v", p)
	}
}

func modelsProgressProcessing() models.VideoDetailProgress {
	return models.VideoDetailProgress{IsProcessing: true, Status: "running"}
}

func TestSVCSaveVideoDetail(t *testing.T) {
	v := svcVideoBaseInfo("BVsvcVD0001", "svc 详情保存", "svc分区")
	if err := SaveVideoDetail(v); err != nil {
		t.Fatalf("SaveVideoDetail: %v", err)
	}
	got, err := database.GetVideoBaseInfoByBvid("BVsvcVD0001")
	if err != nil || got == nil || got.Title != "svc 详情保存" {
		t.Fatalf("read back: %+v err=%v", got, err)
	}
	// 再次保存走 upsert 更新路径
	v.Title = "svc 详情已更新"
	if err := SaveVideoDetail(v); err != nil {
		t.Fatalf("upsert update: %v", err)
	}
	got, _ = database.GetVideoBaseInfoByBvid("BVsvcVD0001")
	if got.Title != "svc 详情已更新" {
		t.Errorf("upsert did not update: %q", got.Title)
	}
}

func TestSVCSyncHistoryTagNames(t *testing.T) {
	// 种一条 bvid 与 video_base_info 匹配、但 tag_name 为空的历史记录（2031 表专用行）
	v := svcVideoBaseInfo("BVsvcTag0001", "svc 标签同步", "svc回填分区")
	if err := database.UpsertVideoBaseInfo(v); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	rec := svcHistoryRecord("BVsvcTag0001", "svc 标签同步历史", "archive", "", "科技", 7700999, "svc标签UP", svcAt(svcYearMain, 11, 11, 1, 0), 60, 60)
	rec.TagName = ""
	if err := svcSeedHistoryRows(svcYearMain, rec); err != nil {
		t.Fatalf("seed: %v", err)
	}
	updated, err := SyncHistoryTagNames()
	if err != nil {
		t.Fatalf("SyncHistoryTagNames: %v", err)
	}
	if updated < 1 {
		t.Errorf("updated = %d, want >= 1", updated)
	}
	var tag string
	if err := svcConn().QueryRow("SELECT tag_name FROM bilibili_history_2031 WHERE bvid='BVsvcTag0001'").Scan(&tag); err != nil {
		t.Fatalf("query: %v", err)
	}
	if tag != "svc回填分区" {
		t.Errorf("tag_name after sync = %q", tag)
	}
}

func TestSVCBatchFetchGuards(t *testing.T) {
	ResetVideoDetailProgress()
	// SESSDATA 未配置：两个批量入口在启动 goroutine 前返回（零网络）
	if _, err := BatchFetchVideoDetails([]string{"BVsvcNone0001"}); err == nil || !strings.Contains(err.Error(), "SESSDATA") {
		t.Errorf("BatchFetchVideoDetails err = %v", err)
	}
	if _, err := BatchFetchFromHistory(false); err == nil || !strings.Contains(err.Error(), "SESSDATA") {
		t.Errorf("BatchFetchFromHistory err = %v", err)
	}
	// 空列表先于配置检查（第一个守卫）
	if _, err := BatchFetchVideoDetails(nil); err == nil || !strings.Contains(err.Error(), "不能为空") {
		t.Errorf("empty bvids err = %v", err)
	}
	// 处理中守卫
	setVideoDetailProgress(modelsProgressProcessing())
	defer ResetVideoDetailProgress()
	if _, err := BatchFetchVideoDetails([]string{"x"}); err == nil || !strings.Contains(err.Error(), "正在运行") {
		t.Errorf("running guard err = %v", err)
	}
	if _, err := BatchFetchFromHistory(true); err == nil || !strings.Contains(err.Error(), "正在运行") {
		t.Errorf("running guard 2 err = %v", err)
	}
	ResetVideoDetailProgress()
}
