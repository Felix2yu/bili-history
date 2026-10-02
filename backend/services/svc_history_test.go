package services

import (
	"bilibili-history-go/database"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSVCFindLatestHistoryDate(t *testing.T) {
	latest, err := FindLatestHistoryDate()
	if err != nil {
		t.Fatalf("FindLatestHistoryDate: %v", err)
	}
	// 契约是"跨所有年份表的最大 view_at"，不是某个固定种子值：其他用例会往
	// 命名空间表里追加行，所以期望值实时从库里取（-shuffle=on 下同样成立）。
	want := svcGlobalMaxViewAt(t)
	if want < svcAt(svcYearClean, 2, 3, 12, 0) {
		t.Fatalf("fixture sanity: max view_at = %d, want at least the 2032-02-03 row", want)
	}
	if latest.Unix() != want {
		t.Errorf("latest = %v (%d), want %d", latest, latest.Unix(), want)
	}
	if latest.Format("2006-01-02") != time.Unix(want, 0).In(time.Local).Format("2006-01-02") {
		t.Errorf("latest should render in local time")
	}
}

// svcGlobalMaxViewAt 复刻 FindLatestHistoryDate 的语义：遍历可用年份表取
// MAX(view_at)，缺少 view_at 列的表跳过（其查询报错，与生产实现一致）。
func svcGlobalMaxViewAt(t *testing.T) int64 {
	t.Helper()
	db := database.GetSQLiteDB()
	conn := db.GetDB()
	years, err := db.GetAvailableYears()
	if err != nil {
		t.Fatalf("available years: %v", err)
	}
	var max int64
	for _, y := range years {
		var v int64
		q := "SELECT COALESCE(MAX(view_at), 0) FROM bilibili_history_" + strconv.Itoa(y)
		if err := conn.QueryRow(q).Scan(&v); err != nil {
			continue
		}
		if v > max {
			max = v
		}
	}
	return max
}

func TestSVCFetchTaskStatusLifecycle(t *testing.T) {
	taskID := "svc-task-1"
	if GetFetchTaskStatus(taskID) != nil {
		t.Fatalf("precondition: task should not exist")
	}
	st := &FetchStatus{IsRunning: true, Status: "running", TotalPages: 2}
	setFetchTaskStatus(taskID, st)
	got := GetFetchTaskStatus(taskID)
	if got == nil || got.TaskID != taskID || !got.IsRunning {
		t.Fatalf("task status = %+v", got)
	}

	overall := GetFetchStatusOverall()
	rc, ok := overall["running_count"].(int32)
	if !ok {
		t.Fatalf("running_count type = %T", overall["running_count"])
	}
	if rc < 0 {
		t.Errorf("running_count negative: %d", rc)
	}
	tasks, ok := overall["tasks"].(map[string]*FetchStatus)
	if !ok {
		t.Fatalf("tasks type = %T", overall["tasks"])
	}
	if _, exists := tasks[taskID]; !exists {
		t.Errorf("tasks should contain %s", taskID)
	}

	removeFetchTaskStatus(taskID)
	if GetFetchTaskStatus(taskID) != nil {
		t.Errorf("task should be removed")
	}
}

func TestSVCFetchGuardsWithoutSessdata(t *testing.T) {
	// SESSDATA 为空：两个抓取入口都必须在创建任务/发起网络之前返回错误
	if res, err := FetchHistory("svc-guard-1", false); err == nil {
		t.Errorf("FetchHistory should refuse without SESSDATA, got %v", res)
	} else if !strings.Contains(err.Error(), "SESSDATA") {
		t.Errorf("FetchHistory err = %v", err)
	}
	if res, err := FetchHistorySync("svc-guard-2", false); err == nil {
		t.Errorf("FetchHistorySync should refuse without SESSDATA, got %v", res)
	} else if !strings.Contains(err.Error(), "SESSDATA") {
		t.Errorf("FetchHistorySync err = %v", err)
	}
	// 守卫先于任务注册：不应留下任务记录
	if GetFetchTaskStatus("svc-guard-1") != nil || GetFetchTaskStatus("svc-guard-2") != nil {
		t.Errorf("guards must not register fetch tasks")
	}
}

func TestSVCFindLatestHistoryDateSkipsBrokenTable(t *testing.T) {
	// 缺少 view_at 列的年份表：MAX() 查询报错应被跳过，全局最大值不变
	conn := svcConn()
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS bilibili_history_2036 (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create broken table: %v", err)
	}
	defer func() {
		if _, err := conn.Exec(`DROP TABLE IF EXISTS bilibili_history_2036`); err != nil {
			t.Errorf("drop broken table: %v", err)
		}
	}()
	latest, err := FindLatestHistoryDate()
	if err != nil {
		t.Fatalf("FindLatestHistoryDate: %v", err)
	}
	// 期望值同样实时取自库，避免依赖其它用例的执行顺序。
	if want := svcGlobalMaxViewAt(t); latest.Unix() != want {
		t.Errorf("latest = %d, want %d", latest.Unix(), want)
	}
}

func TestSVCSetFetchTaskStatusAssignsTaskID(t *testing.T) {
	st := &FetchStatus{Status: "queued"}
	setFetchTaskStatus("svc-task-id", st)
	defer removeFetchTaskStatus("svc-task-id")
	if st.TaskID != "svc-task-id" {
		t.Errorf("TaskID should be assigned by setter, got %q", st.TaskID)
	}
	// GetFetchStatusOverall 的副本包含该任务
	overall := GetFetchStatusOverall()
	tasks := overall["tasks"].(map[string]*FetchStatus)
	if _, ok := tasks["svc-task-id"]; !ok {
		t.Errorf("missing task in snapshot")
	}
}
