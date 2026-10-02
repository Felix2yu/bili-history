package services

import (
	"strconv"
	"strings"
	"testing"
)

func TestSVCDataSyncStatusAccessors(t *testing.T) {
	setDataSyncStatus(DataSyncStatus{Status: "idle"})
	st := GetDataSyncStatus()
	if st.IsRunning || st.Status != "idle" {
		t.Errorf("status = %+v", st)
	}
}

func TestSVCRunIntegrityCheckDisabled(t *testing.T) {
	// config.yaml 中 check_on_startup=false（默认），forceCheck=false 时应直接短路
	res, err := RunIntegrityCheck(false)
	if err != nil {
		t.Fatalf("RunIntegrityCheck: %v", err)
	}
	if !res.Success || !strings.Contains(res.Message, "禁用") {
		t.Errorf("disabled branch = %+v", res)
	}
	// 该分支不会更新 lastCheckResult
}

func TestSVCRunIntegrityCheckForced(t *testing.T) {
	res, err := RunIntegrityCheck(true)
	if err != nil {
		t.Fatalf("RunIntegrityCheck: %v", err)
	}
	if !res.Success || res.Timestamp == "" {
		t.Fatalf("result = %+v", res)
	}
	// 至少包含 2031 主数据集 3 行
	if res.TotalDBRecords < 3 {
		t.Errorf("TotalDBRecords = %d, want >= 3", res.TotalDBRecords)
	}
	if res.Difference != 0 || res.TotalJSONRecords != res.TotalDBRecords {
		t.Errorf("json/db mismatch: %+v", res)
	}
	if !strings.Contains(res.Message, "数据库共") {
		t.Errorf("message = %q", res.Message)
	}
	dataSyncMutex.Lock()
	lc := lastCheckResult
	dataSyncMutex.Unlock()
	if lc == nil || lc.TotalDBRecords != res.TotalDBRecords {
		t.Errorf("lastCheckResult not stored: %+v", lc)
	}
}

func TestSVCCountDBRecordsAndReport(t *testing.T) {
	total, err := countDBRecords()
	if err != nil {
		t.Fatalf("countDBRecords: %v", err)
	}
	if total < 3 {
		t.Errorf("total = %d, want >= 3", total)
	}

	report := GetIntegrityReportData()
	if report == nil {
		t.Fatal("report nil")
	}
	// 2031 是只读主数据集，但 svc_video_details 的标签同步用例会往里补一行，
	// 所以期望值实时 COUNT，报告对得上真实行数即可（与执行顺序无关）。
	var want2031 int
	if err := svcConn().QueryRow("SELECT COUNT(*) FROM bilibili_history_" + strconv.Itoa(svcYearMain)).Scan(&want2031); err != nil {
		t.Fatalf("count %d: %v", svcYearMain, err)
	}
	if want2031 < 3 {
		t.Fatalf("fixture sanity: %d has %d rows, want >= 3", svcYearMain, want2031)
	}
	var found2031 bool
	for _, ys := range report.Years {
		if ys.Year == svcYearMain {
			found2031 = true
			if ys.Count != want2031 {
				t.Errorf("year %d count = %d, want %d (live COUNT)", ys.Year, ys.Count, want2031)
			}
		}
	}
	if !found2031 {
		t.Errorf("years missing %d: %+v", svcYearMain, report.Years)
	}
	if report.TotalRecords < total {
		t.Errorf("report total %d < countDBRecords %d (same data expected)", report.TotalRecords, total)
	}
	if report.MaxYearCount < 1 {
		t.Errorf("max year count = %d", report.MaxYearCount)
	}
}

func TestSVCRunSyncData(t *testing.T) {
	// 互斥守卫
	setDataSyncStatus(DataSyncStatus{IsRunning: true, Status: "running"})
	if _, err := RunSyncData(); err == nil || !strings.Contains(err.Error(), "正在进行中") {
		t.Errorf("guard err = %v", err)
	}
	setDataSyncStatus(DataSyncStatus{Status: "idle"})

	res, err := RunSyncData()
	if err != nil {
		t.Fatalf("RunSyncData: %v", err)
	}
	if !res.Success || res.TotalSynced != 0 || res.JSONToDBCount != 0 || res.DBToJSONCount != 0 {
		t.Errorf("sync result = %+v", res)
	}
	if res.TotalDBRecords < 3 || !strings.Contains(res.Message, "数据库") {
		t.Errorf("result = %+v", res)
	}
	if st := GetDataSyncStatus(); st.IsRunning || st.Status != "completed" || st.LastUpdateAt == 0 {
		t.Errorf("status after sync = %+v", st)
	}
	stored := GetLastSyncResult()
	if stored == nil || stored.Timestamp != res.Timestamp {
		t.Errorf("lastSyncResult not stored: %+v", stored)
	}
	setDataSyncStatus(DataSyncStatus{Status: "idle"})
}

func TestSVCSyncDataInternalMessage(t *testing.T) {
	r := syncDataInternal()
	if r == nil || !r.Success {
		t.Fatalf("syncDataInternal = %+v", r)
	}
	if r.SyncedDays != nil {
		t.Errorf("no JSON layer anymore, synced days should be nil")
	}
}
