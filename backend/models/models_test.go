package models

import "testing"

func TestSuccessResponse(t *testing.T) {
	r := SuccessResponse("payload")
	if r.Status != "success" {
		t.Errorf("期望 status=success, 实际 %q", r.Status)
	}
	if r.Data != "payload" {
		t.Errorf("期望 data=payload, 实际 %v", r.Data)
	}
	if r.Message != "" {
		t.Errorf("期望 message 为空, 实际 %q", r.Message)
	}
}

func TestErrorResponse(t *testing.T) {
	r := ErrorResponse("出错了")
	if r.Status != "error" {
		t.Errorf("期望 status=error, 实际 %q", r.Status)
	}
	if r.Message != "出错了" {
		t.Errorf("期望 message=出错了, 实际 %q", r.Message)
	}
	if r.Data != nil {
		t.Errorf("期望 data 为 nil, 实际 %v", r.Data)
	}
}

func TestModelFields(t *testing.T) {
	h := HistoryRecord{Title: "标题", Bvid: "BV1", ViewAt: 1700000000}
	if h.Title != "标题" || h.Bvid != "BV1" {
		t.Errorf("HistoryRecord 字段赋值异常: %+v", h)
	}

	v := VideoBaseInfo{Bvid: "BV2", Title: "视频"}
	if v.Bvid != "BV2" || v.Title != "视频" {
		t.Errorf("VideoBaseInfo 字段赋值异常: %+v", v)
	}

	p := PagedResponse{Records: []int{1, 2}, Total: 2, Size: 10, Current: 1}
	if p.Total != 2 || p.Current != 1 {
		t.Errorf("PagedResponse 字段赋值异常: %+v", p)
	}

	u := UploaderStats{Mid: 1, Name: "up", VideoCount: 3}
	if u.Name != "up" || u.VideoCount != 3 {
		t.Errorf("UploaderStats 字段赋值异常: %+v", u)
	}
}
