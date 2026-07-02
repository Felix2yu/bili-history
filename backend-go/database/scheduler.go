package database

import (
	"database/sql"
	"path/filepath"
	"sync"

	"bilibili-history-go/utils"
	_ "github.com/mattn/go-sqlite3"
)

var (
	schedulerDB     *sql.DB
	schedulerDBOnce sync.Once
)

func GetSchedulerDB() *sql.DB {
	schedulerDBOnce.Do(func() {
		schedulerDBPath := filepath.Join(utils.GetOutputPath("database"), "scheduler.db")
		var err error
		schedulerDB, err = sql.Open("sqlite3", schedulerDBPath)
		if err != nil {
			utils.LogError("Failed to open scheduler database: %v", err)
			schedulerDB = nil
			return
		}
	})
	return schedulerDB
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

	rows, err := db.Query("SELECT task_id, name, endpoint, method, params, schedule_type, schedule_time, schedule_delay, interval_value, interval_unit, enabled, task_type, created_at, last_modified FROM main_tasks ORDER BY created_at DESC")
	if err != nil {
		return []MainTask{}, err
	}
	defer rows.Close()

	var tasks []MainTask
	for rows.Next() {
		var task MainTask
		var scheduleTime sql.NullString
		var params sql.NullString
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
			&task.CreatedAt,
			&task.LastModified,
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
