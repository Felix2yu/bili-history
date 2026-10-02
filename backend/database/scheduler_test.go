package database

import (
	"database/sql"
	"math"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// scheduler_test.go 覆盖 scheduler.go。scheduler.db 是本包内本文件独占的独立 db 文件，
// 但为稳健起见所有断言只针对 sched_ 前缀的自有 task_id。

var schedTimeRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)

// ---------- helpers (sched 前缀) ----------

func schedStr(m map[string]interface{}, col string) string {
	switch v := m[col].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return ""
}

func schedFindTask(t *testing.T, id string) *MainTask {
	t.Helper()
	tasks, err := GetMainTasks()
	if err != nil {
		t.Fatalf("GetMainTasks: %v", err)
	}
	for i := range tasks {
		if tasks[i].TaskID == id {
			return &tasks[i]
		}
	}
	return nil
}

func schedMustFind(t *testing.T, id string) MainTask {
	t.Helper()
	task := schedFindTask(t, id)
	if task == nil {
		t.Fatalf("task %s not found in GetMainTasks", id)
	}
	return *task
}

func schedRawCount(t *testing.T, query string, args ...interface{}) int {
	t.Helper()
	var n int
	if err := GetSchedulerDB().QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

// ---------- singleton / schema / migration ----------

func TestSchedDBSingletonAndSchema(t *testing.T) {
	db := GetSchedulerDB()
	if db == nil {
		t.Fatal("GetSchedulerDB returned nil")
	}
	if again := GetSchedulerDB(); again != db {
		t.Fatal("GetSchedulerDB not a singleton")
	}
	for _, tbl := range []string{"main_tasks", "task_status", "task_execution_history"} {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", tbl).Scan(&n); err != nil {
			t.Fatalf("check table %s: %v", tbl, err)
		}
		if n != 1 {
			t.Fatalf("table %s missing", tbl)
		}
	}
	// 全量 schema 已含迁移列；再次 migrate 应为幂等（列不重复）
	for _, col := range []string{"parent_id", "depends_on", "task_type", "schedule_delay"} {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('main_tasks') WHERE name = ?", col).Scan(&n); err != nil {
			t.Fatalf("pragma %s: %v", col, err)
		}
		if n != 1 {
			t.Fatalf("column %s count=%d, want 1", col, n)
		}
	}
	migrateSchedulerDB(db) // 幂等重入
	migrateSchedulerDB(db)
	for _, col := range []string{"parent_id", "depends_on", "task_type", "schedule_delay"} {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('main_tasks') WHERE name = ?", col).Scan(&n); err != nil {
			t.Fatalf("pragma %s after migrate: %v", col, err)
		}
		if n != 1 {
			t.Fatalf("migrateSchedulerDB not idempotent for %s: count=%d", col, n)
		}
	}
}

func TestSchedMigrateLegacySchema(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "legacy.db"))
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer db.Close()

	// 旧版 schema：缺少 parent_id / depends_on / task_type / schedule_delay
	if _, err := db.Exec(`CREATE TABLE main_tasks (
		task_id TEXT PRIMARY KEY, name TEXT NOT NULL, endpoint TEXT,
		method TEXT DEFAULT 'GET', params TEXT, schedule_type TEXT DEFAULT 'daily',
		schedule_time TEXT, interval_value INTEGER DEFAULT 0, interval_unit TEXT DEFAULT '',
		enabled INTEGER DEFAULT 0, created_at TEXT, last_modified TEXT)`); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO main_tasks (task_id, name) VALUES ('legacy_1', 'old')`); err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	migrateSchedulerDB(db)

	for _, col := range []string{"parent_id", "depends_on", "task_type", "schedule_delay"} {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('main_tasks') WHERE name = ?", col).Scan(&n); err != nil {
			t.Fatalf("pragma %s: %v", col, err)
		}
		if n != 1 {
			t.Fatalf("column %s not added by migration (count=%d)", col, n)
		}
	}
	// 存量行获得默认值
	var ptype string
	var delay int
	if err := db.QueryRow("SELECT COALESCE(task_type,''), schedule_delay FROM main_tasks WHERE task_id='legacy_1'").Scan(&ptype, &delay); err != nil {
		t.Fatalf("read migrated row: %v", err)
	}
	if ptype != "main" || delay != 0 {
		t.Fatalf("legacy row defaults wrong: type=%q delay=%d", ptype, delay)
	}

	// 再跑一次：列已存在，走跳过分支
	migrateSchedulerDB(db)
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('main_tasks') WHERE name='parent_id'").Scan(&n); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if n != 1 {
		t.Fatalf("double migrate duplicated column: %d", n)
	}

	// ensureColumns 对不存在的表：PRAGMA 无行 → ALTER 失败 → 仅记录日志不 panic
	ensureColumns(db, "no_such_table", map[string]string{"x": "TEXT"})

	// ensureColumns 的 PRAGMA 查询失败分支：已关闭的连接池
	closed, err := sql.Open("sqlite", filepath.Join(dir, "closed.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := closed.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	ensureColumns(closed, "main_tasks", map[string]string{"y": "TEXT"}) // 应直接返回
}

// ---------- upsert / query / null handling ----------

func TestSchedUpsertAndNullHandling(t *testing.T) {
	db := GetSchedulerDB()

	// 插入路径：CreatedAt 为空 → 自动填充
	task := MainTask{
		TaskID: "sched_t1", Name: "sched 一", Endpoint: "/api/one", Method: "POST",
		Params: `{"a":1}`, ScheduleType: "interval", ScheduleTime: "03:00",
		ScheduleDelay: 30, IntervalValue: 10, IntervalUnit: "minutes",
		Enabled: 1, TaskType: "main", DependsOn: "",
	}
	if err := UpsertMainTask(task); err != nil {
		t.Fatalf("UpsertMainTask insert: %v", err)
	}
	got := schedMustFind(t, "sched_t1")
	if !schedTimeRe.MatchString(got.CreatedAt) {
		t.Fatalf("created_at default wrong: %q", got.CreatedAt)
	}
	if !schedTimeRe.MatchString(got.LastModified) {
		t.Fatalf("last_modified format wrong: %q", got.LastModified)
	}
	if got.Name != task.Name || got.Method != "POST" || got.ScheduleDelay != 30 || got.Params != `{"a":1}` {
		t.Fatalf("inserted fields wrong: %+v", got)
	}
	firstCreated := got.CreatedAt

	// 更新路径：created_at 不被 SET 覆盖，name 更新
	task.Name = "sched 一改"
	if err := UpsertMainTask(task); err != nil {
		t.Fatalf("UpsertMainTask update: %v", err)
	}
	got = schedMustFind(t, "sched_t1")
	if got.Name != "sched 一改" {
		t.Fatalf("name not updated: %q", got.Name)
	}
	if got.CreatedAt != firstCreated {
		t.Fatalf("created_at changed on update: %q -> %q", firstCreated, got.CreatedAt)
	}
	if n := schedRawCount(t, "SELECT COUNT(*) FROM main_tasks WHERE task_id='sched_t1'"); n != 1 {
		t.Fatalf("upsert duplicated: %d", n)
	}

	// GetMainTasks 的 NullString 分支：直接写入 NULL 列（endpoint 保持非 NULL 才可扫描）
	mustExec(t, db, `INSERT OR IGNORE INTO main_tasks (task_id, name, endpoint, method, params,
		schedule_type, schedule_time, enabled, created_at, last_modified)
		VALUES ('sched_nulls', 'null row', '', 'GET', NULL, 'daily', NULL, 0, NULL, NULL)`)
	mustExec(t, db, `UPDATE main_tasks SET parent_id = NULL, depends_on = NULL WHERE task_id = 'sched_nulls'`)
	nilTask := schedMustFind(t, "sched_nulls")
	if nilTask.Params != "" || nilTask.ScheduleTime != "" || nilTask.CreatedAt != "" || nilTask.LastModified != "" {
		t.Fatalf("NULL strings should coerce to empty: %+v", nilTask)
	}
	if nilTask.ParentID != "" || nilTask.DependsOn != "" {
		t.Fatalf("COALESCE(parent/depends) failed: %+v", nilTask)
	}

	// Scan 失败 continue 分支：enabled 为 NULL 无法扫进 int → 该行被跳过
	mustExec(t, db, `INSERT OR IGNORE INTO main_tasks (task_id, name, endpoint, enabled) VALUES ('sched_badrow', 'bad', '', NULL)`)
	if found := schedFindTask(t, "sched_badrow"); found != nil {
		t.Fatalf("row with NULL enabled should be skipped by scan-error branch")
	}
}

// ---------- 生命周期：开关 / 依赖 / 端点 / 状态 / 历史 / 级联删除 ----------

func TestSchedTaskLifecycle(t *testing.T) {
	db := GetSchedulerDB()

	parent := MainTask{TaskID: "sched_p", Name: "sched 父任务", Endpoint: "/api/p", Method: "GET",
		ScheduleType: "daily", TaskType: "main"}
	c1 := MainTask{TaskID: "sched_c1", Name: "sched 子一", TaskType: "sub", ParentID: "sched_p"}
	c2 := MainTask{TaskID: "sched_c2", Name: "sched 子二", TaskType: "sub", ParentID: "sched_p"}
	for _, tsk := range []MainTask{parent, c1, c2} {
		if err := UpsertMainTask(tsk); err != nil {
			t.Fatalf("upsert %s: %v", tsk.TaskID, err)
		}
	}

	// SetTaskEnabled 两分支
	if err := SetTaskEnabled("sched_p", true); err != nil {
		t.Fatalf("SetTaskEnabled true: %v", err)
	}
	if got := schedMustFind(t, "sched_p"); got.Enabled != 1 {
		t.Fatalf("enabled want 1 got %d", got.Enabled)
	}
	if err := SetTaskEnabled("sched_p", false); err != nil {
		t.Fatalf("SetTaskEnabled false: %v", err)
	}
	if got := schedMustFind(t, "sched_p"); got.Enabled != 0 {
		t.Fatalf("enabled want 0 got %d", got.Enabled)
	}
	// 不存在的 task：无错误、无行受影响
	if err := SetTaskEnabled("sched_absent", true); err != nil {
		t.Fatalf("SetTaskEnabled absent: %v", err)
	}

	// UpdateTaskDependsOn / UpdateTaskEndpoint
	if err := UpdateTaskDependsOn("sched_p", "sched_c1,sched_c2"); err != nil {
		t.Fatalf("UpdateTaskDependsOn: %v", err)
	}
	if err := UpdateTaskEndpoint("sched_p", "/api/p2"); err != nil {
		t.Fatalf("UpdateTaskEndpoint: %v", err)
	}
	got := schedMustFind(t, "sched_p")
	if got.DependsOn != "sched_c1,sched_c2" || got.Endpoint != "/api/p2" {
		t.Fatalf("depends/endpoint not updated: %+v", got)
	}
	if err := UpdateTaskDependsOn("sched_absent", "x"); err != nil {
		t.Fatalf("UpdateTaskDependsOn absent: %v", err)
	}
	if err := UpdateTaskEndpoint("sched_absent", "x"); err != nil {
		t.Fatalf("UpdateTaskEndpoint absent: %v", err)
	}

	// UpdateTaskStatus：首跑成功（无既有行 → 变量从 0 起）
	if err := UpdateTaskStatus("sched_p", "success", "", 10, true); err != nil {
		t.Fatalf("UpdateTaskStatus #1: %v", err)
	}
	sm, err := GetTaskStatusMap()
	if err != nil {
		t.Fatalf("GetTaskStatusMap: %v", err)
	}
	st := sm["sched_p"]
	if st.TotalRuns != 1 || st.SuccessRuns != 1 || st.FailRuns != 0 || st.LastStatus != "success" {
		t.Fatalf("first run stats: %+v", st)
	}
	if math.Abs(st.AvgDuration-10) > 1e-9 || math.Abs(st.SuccessRate-100) > 1e-9 {
		t.Fatalf("first run avg/rate: %+v", st)
	}
	if !schedTimeRe.MatchString(st.LastRunTime) {
		t.Fatalf("last_run_time format: %q", st.LastRunTime)
	}

	// 二跑失败：滑动平均 (10*1+15)/2=12.5，成功率 50
	if err := UpdateTaskStatus("sched_p", "failed", "boom", 15, false); err != nil {
		t.Fatalf("UpdateTaskStatus #2: %v", err)
	}
	sm, _ = GetTaskStatusMap()
	st = sm["sched_p"]
	if st.TotalRuns != 2 || st.SuccessRuns != 1 || st.FailRuns != 1 {
		t.Fatalf("second run stats: %+v", st)
	}
	if math.Abs(st.AvgDuration-12.5) > 1e-9 || math.Abs(st.SuccessRate-50) > 1e-9 {
		t.Fatalf("avg=%v rate=%v", st.AvgDuration, st.SuccessRate)
	}
	if st.LastError != "boom" {
		t.Fatalf("last_error: %q", st.LastError)
	}

	// 子任务状态（供级联删除验证）
	if err := UpdateTaskStatus("sched_c1", "success", "", 1, true); err != nil {
		t.Fatalf("UpdateTaskStatus c1: %v", err)
	}

	// GetTaskStatusMap 的 NullString 分支：数值列有值、文本列显式写入 NULL
	mustExec(t, db, `INSERT OR IGNORE INTO task_status (task_id, last_run_time, next_run_time, last_status,
		last_error, tags, total_runs, success_runs, fail_runs, avg_duration, success_rate)
		VALUES ('sched_nullst', NULL, NULL, NULL, NULL, NULL, 0, 0, 0, 0, 0)`)
	sm, err = GetTaskStatusMap()
	if err != nil {
		t.Fatalf("GetTaskStatusMap: %v", err)
	}
	ns := sm["sched_nullst"]
	if ns.TaskID != "sched_nullst" || ns.LastRunTime != "" || ns.LastStatus != "" || ns.LastError != "" || ns.Tags != "" {
		t.Fatalf("NULL status strings should coerce to empty: %+v", ns)
	}

	// Scan 失败 continue 分支：total_runs 为 NULL 无法扫进 int → 该行被跳过
	mustExec(t, db, `INSERT OR IGNORE INTO task_status (task_id, total_runs) VALUES ('sched_badst', NULL)`)
	sm, _ = GetTaskStatusMap()
	if _, ok := sm["sched_badst"]; ok {
		t.Fatal("row with NULL total_runs should be skipped by scan-error branch")
	}

	// RecordExecution：3 条父任务 + 1 条子任务，start_time 递增
	base := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	fmtTime := func(mins int) time.Time { return base.Add(time.Duration(mins) * time.Minute) }
	recs := []struct {
		id, task, status string
		start, end       time.Time
	}{
		{"sched_e1", "sched_p", "success", fmtTime(0), fmtTime(1)},
		{"sched_e2", "sched_p", "failed", fmtTime(10), fmtTime(11)},
		{"sched_e3", "sched_p", "success", fmtTime(20), fmtTime(21)},
		{"sched_e4", "sched_c1", "success", fmtTime(30), fmtTime(31)},
	}
	for _, r := range recs {
		if err := RecordExecution(r.id, r.task, r.status, "res", "err?", r.start, r.end); err != nil {
			t.Fatalf("RecordExecution %s: %v", r.id, err)
		}
	}
	// 主键冲突 → 返回错误
	if err := RecordExecution("sched_e1", "sched_p", "dup", "", "", base, base); err == nil {
		t.Fatal("duplicate execution id should error")
	}

	// 全量历史（limit<=0 → 50）：start_time DESC
	all, err := GetExecutionHistory("", 0)
	if err != nil {
		t.Fatalf("history all: %v", err)
	}
	if len(all) != 4 {
		t.Fatalf("want 4 history rows, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if schedStr(all[i-1], "start_time") < schedStr(all[i], "start_time") {
			t.Fatalf("history not DESC by start_time: %v", all)
		}
	}
	if schedStr(all[0], "id") != "sched_e4" || schedStr(all[0], "end_time") == "" {
		t.Fatalf("head row wrong: %v", all[0])
	}

	// 按 task 过滤 + limit 截断
	filtered, err := GetExecutionHistory("sched_p", 2)
	if err != nil || len(filtered) != 2 {
		t.Fatalf("filtered history: len=%d err=%v", len(filtered), err)
	}
	if schedStr(filtered[0], "id") != "sched_e3" || schedStr(filtered[1], "id") != "sched_e2" {
		t.Fatalf("filter/limit order wrong: %v", filtered)
	}
	// limit>200 → 钳到 50；本处 4 行全返回，与 limit=200 分支对照
	clamped, err := GetExecutionHistory("", 300)
	if err != nil || len(clamped) != 4 {
		t.Fatalf("clamped history: len=%d err=%v", len(clamped), err)
	}
	big, err := GetExecutionHistory("", 200)
	if err != nil || len(big) != 4 {
		t.Fatalf("limit=200 history: len=%d err=%v", len(big), err)
	}
	// 未知 task → 空
	if none, err := GetExecutionHistory("sched_absent", 10); err != nil || len(none) != 0 {
		t.Fatalf("unknown task history: len=%d err=%v", len(none), err)
	}

	// 级联删除：父任务 + 子任务 + task_status + 执行历史
	if err := DeleteMainTask("sched_p"); err != nil {
		t.Fatalf("DeleteMainTask: %v", err)
	}
	for _, id := range []string{"sched_p", "sched_c1", "sched_c2"} {
		if n := schedRawCount(t, "SELECT COUNT(*) FROM main_tasks WHERE task_id = ?", id); n != 0 {
			t.Fatalf("task %s not deleted", id)
		}
	}
	if n := schedRawCount(t, "SELECT COUNT(*) FROM task_status WHERE task_id IN ('sched_p','sched_c1')"); n != 0 {
		t.Fatalf("status rows not cascade-deleted: %d", n)
	}
	if n := schedRawCount(t, "SELECT COUNT(*) FROM task_execution_history WHERE task_id IN ('sched_p','sched_c1')"); n != 0 {
		t.Fatalf("history rows not cascade-deleted: %d", n)
	}
	sm, _ = GetTaskStatusMap()
	if _, ok := sm["sched_p"]; ok {
		t.Fatal("sched_p still in status map")
	}
	if hs, err := GetExecutionHistory("sched_c1", 10); err != nil || len(hs) != 0 {
		t.Fatalf("c1 history should be gone: %v %v", hs, err)
	}
	// 删除不存在的任务不应报错
	if err := DeleteMainTask("sched_absent"); err != nil {
		t.Fatalf("DeleteMainTask absent: %v", err)
	}
}
