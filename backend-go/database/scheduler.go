package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"bilibili-history-go/utils"
	_ "github.com/mattn/go-sqlite3"
)

var (
	schedulerDB     *sql.DB
	schedulerDBOnce sync.Once
)

// schedulerSchema creates the main_tasks and task_status tables. Keeping the
// column list in sync with GetMainTasks / GetTaskStatusMap.
const schedulerSchema = `
CREATE TABLE IF NOT EXISTS main_tasks (
    task_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    endpoint TEXT,
    method TEXT DEFAULT 'GET',
    params TEXT,
    schedule_type TEXT DEFAULT 'daily',
    schedule_time TEXT,
    schedule_delay INTEGER DEFAULT 0,
    interval_value INTEGER DEFAULT 0,
    interval_unit TEXT DEFAULT '',
    enabled INTEGER DEFAULT 0,
    task_type TEXT DEFAULT 'main',
    parent_id TEXT DEFAULT '',
    depends_on TEXT DEFAULT '',
    created_at TEXT,
    last_modified TEXT
);
CREATE INDEX IF NOT EXISTS idx_main_tasks_parent ON main_tasks(parent_id);

CREATE TABLE IF NOT EXISTS task_status (
    task_id TEXT PRIMARY KEY,
    last_run_time TEXT,
    next_run_time TEXT,
    last_status TEXT DEFAULT 'idle',
    total_runs INTEGER DEFAULT 0,
    success_runs INTEGER DEFAULT 0,
    fail_runs INTEGER DEFAULT 0,
    avg_duration REAL DEFAULT 0,
    last_error TEXT,
    tags TEXT,
    success_rate REAL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS task_execution_history (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    start_time TEXT,
    end_time TEXT,
    status TEXT,
    result TEXT,
    error TEXT
);
CREATE INDEX IF NOT EXISTS idx_exec_history_task ON task_execution_history(task_id);
CREATE INDEX IF NOT EXISTS idx_exec_history_start ON task_execution_history(start_time);
`

func GetSchedulerDB() *sql.DB {
	schedulerDBOnce.Do(func() {
		schedulerDBPath := filepath.Join(utils.GetOutputPath("database"), "scheduler.db")
		utils.LogInfo("调度器数据库路径: %s", schedulerDBPath)
		var err error
		schedulerDB, err = sql.Open("sqlite3", schedulerDBPath)
		if err != nil {
			utils.LogError("Failed to open scheduler database: %v", err)
			schedulerDB = nil
			return
		}
		if _, err := schedulerDB.Exec(schedulerSchema); err != nil {
			utils.LogError("Failed to ensure scheduler schema: %v", err)
		}
		migrateSchedulerDB(schedulerDB)
	})
	return schedulerDB
}

// migrateSchedulerDB adds columns that were introduced after the initial schema.
// CREATE TABLE IF NOT EXISTS won't add new columns to an existing table, so we
// must ALTER TABLE explicitly for upgrades from older database versions.
func migrateSchedulerDB(db *sql.DB) {
	// Columns to ensure exist in main_tasks: column_name -> DDL fragment
	mainTaskMigrations := map[string]string{
		"parent_id":     "TEXT DEFAULT ''",
		"depends_on":    "TEXT DEFAULT ''",
		"task_type":     "TEXT DEFAULT 'main'",
		"schedule_delay": "INTEGER DEFAULT 0",
	}
	ensureColumns(db, "main_tasks", mainTaskMigrations)
}

// ensureColumns checks whether each column exists in the given table and adds
// it via ALTER TABLE ADD COLUMN if missing.
func ensureColumns(db *sql.DB, table string, columns map[string]string) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		utils.LogError("Failed to inspect %s columns: %v", table, err)
		return
	}
	defer rows.Close()

	existing := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			continue
		}
		existing[name] = true
	}

	for col, ddl := range columns {
		if !existing[col] {
			stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, ddl)
			if _, err := db.Exec(stmt); err != nil {
				utils.LogError("Failed to add column %s to %s: %v", col, table, err)
			} else {
				utils.LogSuccess("Migrated %s: added column %s", table, col)
			}
		}
	}
}

type MainTask struct {
	TaskID         string `json:"task_id"`
	Name           string `json:"name"`
	Endpoint       string `json:"endpoint"`
	Method         string `json:"method"`
	Params         string `json:"params"`
	ScheduleType   string `json:"schedule_type"`
	ScheduleTime   string `json:"schedule_time"`
	ScheduleDelay  int    `json:"schedule_delay"`
	IntervalValue  int    `json:"interval_value"`
	IntervalUnit   string `json:"interval_unit"`
	Enabled        int    `json:"enabled"`
	TaskType       string `json:"task_type"`
	ParentID       string `json:"parent_id"`
	DependsOn      string `json:"depends_on"`
	CreatedAt      string `json:"created_at"`
	LastModified   string `json:"last_modified"`
}

type TaskStatus struct {
	TaskID        string  `json:"task_id"`
	LastRunTime   string  `json:"last_run_time"`
	NextRunTime   string  `json:"next_run_time"`
	LastStatus    string  `json:"last_status"`
	TotalRuns     int     `json:"total_runs"`
	SuccessRuns   int     `json:"success_runs"`
	FailRuns      int     `json:"fail_runs"`
	AvgDuration   float64 `json:"avg_duration"`
	LastError     string  `json:"last_error"`
	Tags          string  `json:"tags"`
	SuccessRate   float64 `json:"success_rate"`
}

func GetMainTasks() ([]MainTask, error) {
	db := GetSchedulerDB()
	if db == nil {
		return []MainTask{}, nil
	}

	rows, err := db.Query(`SELECT task_id, name, endpoint, method, params, schedule_type,
		schedule_time, schedule_delay, interval_value, interval_unit, enabled, task_type,
		COALESCE(parent_id, ''), COALESCE(depends_on, ''), created_at, last_modified
		FROM main_tasks ORDER BY created_at ASC`)
	if err != nil {
		return []MainTask{}, err
	}
	defer rows.Close()

	var tasks []MainTask
	for rows.Next() {
		var task MainTask
		var scheduleTime sql.NullString
		var params sql.NullString
		var createdAt sql.NullString
		var lastModified sql.NullString
		err := rows.Scan(
			&task.TaskID,
			&task.Name,
			&task.Endpoint,
			&task.Method,
			&params,
			&task.ScheduleType,
			&scheduleTime,
			&task.ScheduleDelay,
			&task.IntervalValue,
			&task.IntervalUnit,
			&task.Enabled,
			&task.TaskType,
			&task.ParentID,
			&task.DependsOn,
			&createdAt,
			&lastModified,
		)
		if err != nil {
			continue
		}
		if scheduleTime.Valid {
			task.ScheduleTime = scheduleTime.String
		}
		if params.Valid {
			task.Params = params.String
		}
		if createdAt.Valid {
			task.CreatedAt = createdAt.String
		}
		if lastModified.Valid {
			task.LastModified = lastModified.String
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func GetSubTasks(parentID string) ([]MainTask, error) {
	db := GetSchedulerDB()
	if db == nil {
		return []MainTask{}, nil
	}
	rows, err := db.Query(`SELECT task_id, name, endpoint, method, params, schedule_type,
		schedule_time, schedule_delay, interval_value, interval_unit, enabled, task_type,
		COALESCE(parent_id, ''), COALESCE(depends_on, ''), created_at, last_modified
		FROM main_tasks WHERE parent_id = ? ORDER BY created_at ASC`, parentID)
	if err != nil {
		return []MainTask{}, err
	}
	defer rows.Close()

	var tasks []MainTask
	for rows.Next() {
		var task MainTask
		var scheduleTime sql.NullString
		var params sql.NullString
		var createdAt sql.NullString
		var lastModified sql.NullString
		err := rows.Scan(
			&task.TaskID, &task.Name, &task.Endpoint, &task.Method, &params,
			&task.ScheduleType, &scheduleTime, &task.ScheduleDelay, &task.IntervalValue,
			&task.IntervalUnit, &task.Enabled, &task.TaskType, &task.ParentID,
			&task.DependsOn, &createdAt, &lastModified,
		)
		if err != nil {
			continue
		}
		if scheduleTime.Valid {
			task.ScheduleTime = scheduleTime.String
		}
		if params.Valid {
			task.Params = params.String
		}
		if createdAt.Valid {
			task.CreatedAt = createdAt.String
		}
		if lastModified.Valid {
			task.LastModified = lastModified.String
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func GetTaskStatusMap() (map[string]TaskStatus, error) {
	db := GetSchedulerDB()
	if db == nil {
		return map[string]TaskStatus{}, nil
	}

	rows, err := db.Query("SELECT task_id, last_run_time, next_run_time, last_status, total_runs, success_runs, fail_runs, avg_duration, last_error, tags, success_rate FROM task_status")
	if err != nil {
		return map[string]TaskStatus{}, err
	}
	defer rows.Close()

	statusMap := make(map[string]TaskStatus)
	for rows.Next() {
		var status TaskStatus
		var lastRunTime sql.NullString
		var nextRunTime sql.NullString
		var lastStatus sql.NullString
		var lastError sql.NullString
		var tags sql.NullString
		err := rows.Scan(
			&status.TaskID,
			&lastRunTime,
			&nextRunTime,
			&lastStatus,
			&status.TotalRuns,
			&status.SuccessRuns,
			&status.FailRuns,
			&status.AvgDuration,
			&lastError,
			&tags,
			&status.SuccessRate,
		)
		if err != nil {
			continue
		}
		if lastRunTime.Valid {
			status.LastRunTime = lastRunTime.String
		}
		if nextRunTime.Valid {
			status.NextRunTime = nextRunTime.String
		}
		if lastStatus.Valid {
			status.LastStatus = lastStatus.String
		}
		if lastError.Valid {
			status.LastError = lastError.String
		}
		if tags.Valid {
			status.Tags = tags.String
		}
		statusMap[status.TaskID] = status
	}

	return statusMap, nil
}

// ---- Python 数据库兼容层 ----
// Python 版本使用独立的 sub_tasks / sub_task_status / task_dependencies /
// task_executions 表，Go 版本将其合并到 main_tasks / task_status /
// task_execution_history 中。以下函数用于从 Python 表中读取已有数据。

// tableExists 检查表是否存在。
func tableExists(db *sql.DB, table string) bool {
	var count int
	err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// GetPythonSubTasks 从 Python 的 sub_tasks 表读取子任务，转换为 MainTask 格式。
// 如果表不存在则返回空切片。
func GetPythonSubTasks() ([]MainTask, error) {
	db := GetSchedulerDB()
	if db == nil {
		return []MainTask{}, nil
	}
	if !tableExists(db, "sub_tasks") {
		return []MainTask{}, nil
	}
	rows, err := db.Query(`SELECT task_id, parent_id, name, endpoint, method, params,
		schedule_type, enabled, created_at, last_modified
		FROM sub_tasks ORDER BY parent_id, created_at ASC`)
	if err != nil {
		return []MainTask{}, err
	}
	defer rows.Close()

	var tasks []MainTask
	for rows.Next() {
		var t MainTask
		var params, createdAt, lastModified sql.NullString
		if err := rows.Scan(&t.TaskID, &t.ParentID, &t.Name, &t.Endpoint, &t.Method,
			&params, &t.ScheduleType, &t.Enabled, &createdAt, &lastModified); err != nil {
			continue
		}
		t.TaskType = "sub"
		if t.ScheduleType == "" {
			t.ScheduleType = "chain"
		}
		if params.Valid {
			t.Params = params.String
		}
		if createdAt.Valid {
			t.CreatedAt = createdAt.String
		}
		if lastModified.Valid {
			t.LastModified = lastModified.String
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// GetPythonSubTaskStatusMap 从 Python 的 sub_task_status 表读取子任务状态。
func GetPythonSubTaskStatusMap() (map[string]TaskStatus, error) {
	db := GetSchedulerDB()
	if db == nil {
		return map[string]TaskStatus{}, nil
	}
	if !tableExists(db, "sub_task_status") {
		return map[string]TaskStatus{}, nil
	}
	rows, err := db.Query(`SELECT task_id, last_run_time, next_run_time, last_status,
		total_runs, success_runs, fail_runs, avg_duration, last_error, tags, success_rate
		FROM sub_task_status`)
	if err != nil {
		return map[string]TaskStatus{}, err
	}
	defer rows.Close()

	statusMap := make(map[string]TaskStatus)
	for rows.Next() {
		var s TaskStatus
		var lastRunTime, nextRunTime, lastStatus, lastError, tags sql.NullString
		if err := rows.Scan(&s.TaskID, &lastRunTime, &nextRunTime, &lastStatus,
			&s.TotalRuns, &s.SuccessRuns, &s.FailRuns, &s.AvgDuration,
			&lastError, &tags, &s.SuccessRate); err != nil {
			continue
		}
		if lastRunTime.Valid {
			s.LastRunTime = lastRunTime.String
		}
		if nextRunTime.Valid {
			s.NextRunTime = nextRunTime.String
		}
		if lastStatus.Valid {
			s.LastStatus = lastStatus.String
		}
		if lastError.Valid {
			s.LastError = lastError.String
		}
		if tags.Valid {
			s.Tags = tags.String
		}
		statusMap[s.TaskID] = s
	}
	return statusMap, nil
}

// GetPythonDependencies 从 Python 的 task_dependencies 表读取依赖关系。
// 返回 task_id -> depends_on 的映射。
func GetPythonDependencies() (map[string]string, error) {
	db := GetSchedulerDB()
	if db == nil {
		return map[string]string{}, nil
	}
	if !tableExists(db, "task_dependencies") {
		return map[string]string{}, nil
	}
	rows, err := db.Query("SELECT task_id, depends_on FROM task_dependencies")
	if err != nil {
		return map[string]string{}, err
	}
	defer rows.Close()

	depMap := make(map[string]string)
	for rows.Next() {
		var taskID, dependsOn string
		if err := rows.Scan(&taskID, &dependsOn); err != nil {
			continue
		}
		depMap[taskID] = dependsOn
	}
	return depMap, nil
}

// GetPythonExecutions 从 Python 的 task_executions 表读取执行历史，兼容 Go 的返回格式。
func GetPythonExecutions(taskID string, limit int) ([]map[string]interface{}, error) {
	db := GetSchedulerDB()
	if db == nil {
		return []map[string]interface{}{}, nil
	}
	if !tableExists(db, "task_executions") {
		return []map[string]interface{}{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows *sql.Rows
	var err error
	if taskID != "" {
		rows, err = db.Query(`SELECT id, task_id, start_time, end_time, status, output, error_message
			FROM task_executions WHERE task_id = ? ORDER BY start_time DESC LIMIT ?`, taskID, limit)
	} else {
		rows, err = db.Query(`SELECT id, task_id, start_time, end_time, status, output, error_message
			FROM task_executions ORDER BY start_time DESC LIMIT ?`, limit)
	}
	if err != nil {
		return []map[string]interface{}{}, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, tid, startTime, status interface{}
		var endTime, output, errorMsg sql.NullString
		if err := rows.Scan(&id, &tid, &startTime, &endTime, &status, &output, &errorMsg); err != nil {
			continue
		}
		m := map[string]interface{}{
			"id":         id,
			"task_id":    tid,
			"start_time": startTime,
			"status":     status,
		}
		if endTime.Valid {
			m["end_time"] = endTime.String
		}
		if output.Valid {
			m["result"] = output.String
		}
		if errorMsg.Valid {
			m["error"] = errorMsg.String
		}
		results = append(results, m)
	}
	return results, nil
}

// UpsertMainTask inserts or updates a main task row. If the task already exists
// (matched by task_id), it updates all mutable fields.
func UpsertMainTask(t MainTask) error {
	db := GetSchedulerDB()
	if db == nil {
		return fmt.Errorf("scheduler database not available")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	if t.CreatedAt == "" {
		t.CreatedAt = now
	}
	t.LastModified = now
	_, err := db.Exec(`INSERT INTO main_tasks
		(task_id, name, endpoint, method, params, schedule_type, schedule_time,
		 schedule_delay, interval_value, interval_unit, enabled, task_type,
		 parent_id, depends_on, created_at, last_modified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(task_id) DO UPDATE SET
			name=excluded.name, endpoint=excluded.endpoint, method=excluded.method,
			params=excluded.params, schedule_type=excluded.schedule_type,
			schedule_time=excluded.schedule_time, schedule_delay=excluded.schedule_delay,
			interval_value=excluded.interval_value, interval_unit=excluded.interval_unit,
			enabled=excluded.enabled, task_type=excluded.task_type,
			parent_id=excluded.parent_id, depends_on=excluded.depends_on,
			last_modified=excluded.last_modified`,
		t.TaskID, t.Name, t.Endpoint, t.Method, t.Params, t.ScheduleType, t.ScheduleTime,
		t.ScheduleDelay, t.IntervalValue, t.IntervalUnit, t.Enabled, t.TaskType,
		t.ParentID, t.DependsOn, t.CreatedAt, t.LastModified)
	return err
}

// DeleteMainTask removes a task and its status row. Sub-tasks (parent_id = task_id)
// are also removed to keep referential integrity.
func DeleteMainTask(taskID string) error {
	db := GetSchedulerDB()
	if db == nil {
		return fmt.Errorf("scheduler database not available")
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM main_tasks WHERE task_id = ? OR parent_id = ?", taskID, taskID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM task_status WHERE task_id = ? OR task_id IN (SELECT task_id FROM main_tasks WHERE parent_id = ?)", taskID, taskID); err != nil {
		// best effort
		_ = err
	}
	return tx.Commit()
}

// SetTaskEnabled toggles the enabled flag of a task.
func SetTaskEnabled(taskID string, enabled bool) error {
	db := GetSchedulerDB()
	if db == nil {
		return fmt.Errorf("scheduler database not available")
	}
	val := 0
	if enabled {
		val = 1
	}
	_, err := db.Exec("UPDATE main_tasks SET enabled = ?, last_modified = ? WHERE task_id = ?",
		val, time.Now().Format("2006-01-02 15:04:05"), taskID)
	return err
}

// UpdateTaskStatus records the result of a task execution.
func UpdateTaskStatus(taskID, status, lastError string, durationSec float64, success bool) error {
	db := GetSchedulerDB()
	if db == nil {
		return fmt.Errorf("scheduler database not available")
	}
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var totalRuns, successRuns, failRuns int
	var avgDuration float64
	_ = tx.QueryRow("SELECT total_runs, success_runs, fail_runs, avg_duration FROM task_status WHERE task_id = ?", taskID).
		Scan(&totalRuns, &successRuns, &failRuns, &avgDuration)

	totalRuns++
	if success {
		successRuns++
	} else {
		failRuns++
	}
	// running average
	if totalRuns > 0 {
		avgDuration = (avgDuration*float64(totalRuns-1) + durationSec) / float64(totalRuns)
	}
	successRate := 0.0
	if totalRuns > 0 {
		successRate = float64(successRuns) / float64(totalRuns) * 100
	}

	_, err = tx.Exec(`INSERT INTO task_status
		(task_id, last_run_time, next_run_time, last_status, total_runs, success_runs,
		 fail_runs, avg_duration, last_error, tags, success_rate)
		VALUES (?, ?, '', ?, ?, ?, ?, ?, ?, '', ?)
		ON CONFLICT(task_id) DO UPDATE SET
			last_run_time=excluded.last_run_time, last_status=excluded.last_status,
			total_runs=excluded.total_runs, success_runs=excluded.success_runs,
			fail_runs=excluded.fail_runs, avg_duration=excluded.avg_duration,
			last_error=excluded.last_error, success_rate=excluded.success_rate`,
		taskID, now, status, totalRuns, successRuns, failRuns, avgDuration, lastError, successRate)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// RecordExecution inserts a row into task_execution_history for the history endpoint.
func RecordExecution(id, taskID, status, result, errMsg string, start, end time.Time) error {
	db := GetSchedulerDB()
	if db == nil {
		return nil
	}
	_, err := db.Exec(`INSERT INTO task_execution_history
		(id, task_id, start_time, end_time, status, result, error)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, taskID, start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"),
		status, result, errMsg)
	return err
}

// GetExecutionHistory returns recent execution records, optionally filtered by task_id.
func GetExecutionHistory(taskID string, limit int) ([]map[string]interface{}, error) {
	db := GetSchedulerDB()
	if db == nil {
		return []map[string]interface{}{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows *sql.Rows
	var err error
	if taskID != "" {
		rows, err = db.Query(`SELECT id, task_id, start_time, end_time, status, result, error
			FROM task_execution_history WHERE task_id = ? ORDER BY start_time DESC LIMIT ?`, taskID, limit)
	} else {
		rows, err = db.Query(`SELECT id, task_id, start_time, end_time, status, result, error
			FROM task_execution_history ORDER BY start_time DESC LIMIT ?`, limit)
	}
	if err != nil {
		return []map[string]interface{}{}, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range columns {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		m := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]
			if v == nil {
				continue
			}
			switch vv := v.(type) {
			case []byte:
				m[col] = string(vv)
			default:
				m[col] = vv
			}
		}
		results = append(results, m)
	}
	return results, nil
}
