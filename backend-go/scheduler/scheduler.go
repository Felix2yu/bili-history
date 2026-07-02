package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"
)

type TaskStatus string

const (
	TaskStatusIdle      TaskStatus = "idle"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusStopped   TaskStatus = "stopped"
)

// ScheduleTask is the in-memory representation of a scheduled task. It mirrors
// the main_tasks DB row plus runtime execution state from task_status.
type ScheduleTask struct {
	ID            string `json:"task_id"`
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	Method        string `json:"method"`
	Params        string `json:"params"`
	ScheduleType  string `json:"schedule_type"` // daily | interval | chain | once
	ScheduleTime  string `json:"schedule_time"` // HH:MM for daily
	IntervalValue int    `json:"interval_value"`
	IntervalUnit  string `json:"interval_unit"` // minutes | hours | days | months | years
	Enabled       bool   `json:"enabled"`
	TaskType      string `json:"task_type"` // main | sub
	ParentID      string `json:"parent_id"`
	DependsOn     string `json:"depends_on"`
	CronExpr      string `json:"cron_expr"`
	CreatedAt     string `json:"created_at"`
	LastModified  string `json:"last_modified"`

	// runtime execution state
	LastRunTime string  `json:"last_run_time"`
	NextRunTime string  `json:"next_run_time"`
	LastStatus  string  `json:"last_status"`
	LastError   string  `json:"last_error"`
	TotalRuns   int     `json:"total_runs"`
	SuccessRuns int     `json:"success_runs"`
	FailRuns    int     `json:"fail_runs"`
	AvgDuration float64 `json:"avg_duration"`
	SuccessRate float64 `json:"success_rate"`

	// in-memory running flag (not persisted)
	Running bool `json:"-"`
}

type TaskExecution struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time,omitempty"`
	Status    string `json:"status"`
	Result    string `json:"result,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Scheduler struct {
	tasks      map[string]*ScheduleTask
	mu         sync.RWMutex
	stopCh     chan struct{}
	running    bool
	wg         sync.WaitGroup
	serverPort int
}

var (
	instance *Scheduler
	once     sync.Once
)

func GetScheduler() *Scheduler {
	once.Do(func() {
		port := 8899
		if cfg := config.GetConfig(); cfg != nil && cfg.Server.Port > 0 {
			port = cfg.Server.Port
		}
		instance = &Scheduler{
			tasks:      make(map[string]*ScheduleTask),
			stopCh:     make(chan struct{}),
			serverPort: port,
		}
		instance.loadTasks()
		instance.initDefaultTasks()
	})
	return instance
}

func (s *Scheduler) loadTasks() {
	mainTasks, err := database.GetMainTasks()
	if err != nil {
		utils.LogError("Failed to load scheduler tasks: %v", err)
		return
	}

	statusMap, err := database.GetTaskStatusMap()
	if err != nil {
		utils.LogError("Failed to load scheduler task status: %v", err)
		statusMap = make(map[string]database.TaskStatus)
	}

	for _, mt := range mainTasks {
		task := convertTask(mt)
		if status, ok := statusMap[mt.TaskID]; ok {
			task.LastRunTime = status.LastRunTime
			task.NextRunTime = status.NextRunTime
			task.LastStatus = status.LastStatus
			task.LastError = status.LastError
			task.TotalRuns = status.TotalRuns
			task.SuccessRuns = status.SuccessRuns
			task.FailRuns = status.FailRuns
			task.AvgDuration = status.AvgDuration
			task.SuccessRate = status.SuccessRate
			if status.LastStatus == "running" {
				task.Running = true
			}
		}
		s.tasks[task.ID] = task
	}
}

func convertTask(mt database.MainTask) *ScheduleTask {
	task := &ScheduleTask{
		ID:            mt.TaskID,
		Name:          mt.Name,
		Endpoint:      mt.Endpoint,
		Method:        mt.Method,
		Params:        mt.Params,
		ScheduleType:  mt.ScheduleType,
		ScheduleTime:  mt.ScheduleTime,
		IntervalValue: mt.IntervalValue,
		IntervalUnit:  mt.IntervalUnit,
		Enabled:       mt.Enabled == 1,
		TaskType:      mt.TaskType,
		ParentID:      mt.ParentID,
		DependsOn:     mt.DependsOn,
		CreatedAt:     mt.CreatedAt,
		LastModified:  mt.LastModified,
		LastStatus:    "idle",
	}
	if task.Method == "" {
		task.Method = "GET"
	}
	if task.ScheduleType == "" {
		task.ScheduleType = "daily"
	}
	if task.TaskType == "" {
		task.TaskType = "main"
	}
	task.CronExpr = buildCronExpr(*task)
	return task
}

func buildCronExpr(t ScheduleTask) string {
	switch t.ScheduleType {
	case "daily":
		if t.ScheduleTime != "" {
			parts := strings.Split(t.ScheduleTime, ":")
			if len(parts) == 2 {
				return fmt.Sprintf("%s %s * * *", parts[1], parts[0])
			}
		}
		return "0 0 * * *"
	case "interval":
		if t.IntervalValue <= 0 {
			return ""
		}
		switch t.IntervalUnit {
		case "minutes":
			return fmt.Sprintf("*/%d * * * *", t.IntervalValue)
		case "hours":
			return fmt.Sprintf("0 */%d * * *", t.IntervalValue)
		case "days":
			return fmt.Sprintf("0 0 */%d * *", t.IntervalValue)
		case "months":
			return fmt.Sprintf("0 0 1 */%d *", t.IntervalValue)
		case "years":
			// cron has no year field; approximate with yearly on Jan 1
			return "0 0 1 1 *"
		}
	case "chain", "once":
		// chain tasks are triggered by their parent; once tasks have no cron.
		return ""
	}
	return ""
}

func (s *Scheduler) initDefaultTasks() {
	s.mu.Lock()
	// Only seed if there are no tasks at all (fresh DB / first run).
	if len(s.tasks) > 0 {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	// Seed the 8 default tasks from the Python scheduler_config.yaml.
	// These mirror the Python defaults so users upgrading see the same tasks.
	defaults := []database.MainTask{
		{TaskID: "fetch_popular_videos", Name: "获取热门视频", Endpoint: "/bilibili/popular", Method: "GET", ScheduleType: "interval", IntervalValue: 10, IntervalUnit: "minutes", Enabled: 0, TaskType: "main"},
		{TaskID: "sessdata_health_check", Name: "SESSDATA 健康检查", Endpoint: "/login/check-and-notify", Method: "GET", ScheduleType: "interval", IntervalValue: 10, IntervalUnit: "minutes", Enabled: 0, TaskType: "main"},
		{TaskID: "sync_likes", Name: "同步点赞列表", Endpoint: "/like/list", Method: "GET", ScheduleType: "interval", IntervalValue: 1, IntervalUnit: "hours", Enabled: 0, TaskType: "main"},
		{TaskID: "fetch_history", Name: "获取B站历史记录", Endpoint: "/fetch/bili-history", Method: "GET", ScheduleType: "daily", ScheduleTime: "00:00", Enabled: 0, TaskType: "main"},
		// The four chain tasks below depend on fetch_history and form the daily pipeline.
		{TaskID: "import_data", Name: "导入数据", Endpoint: "/importSqlite/import_data_sqlite", Method: "POST", ScheduleType: "chain", DependsOn: "fetch_history", Enabled: 0, TaskType: "main"},
		{TaskID: "analyze_data", Name: "分析数据", Endpoint: "/analysis/analyze", Method: "POST", ScheduleType: "chain", DependsOn: "import_data", Enabled: 0, TaskType: "main"},
		{TaskID: "generate_heatmap", Name: "生成热力图", Endpoint: "/heatmap/generate_heatmap", Method: "POST", ScheduleType: "chain", DependsOn: "analyze_data", Enabled: 0, TaskType: "main"},
		{TaskID: "send_daily_report", Name: "发送每日报告", Endpoint: "/log/send", Method: "POST", ScheduleType: "chain", DependsOn: "generate_heatmap", Enabled: 0, TaskType: "main"},
	}

	for _, d := range defaults {
		if err := database.UpsertMainTask(d); err != nil {
			utils.LogError("Failed to seed default task %s: %v", d.TaskID, err)
			continue
		}
		s.mu.Lock()
		s.tasks[d.TaskID] = convertTask(d)
		s.mu.Unlock()
	}
	utils.LogSuccess("调度器已种子化 %d 个默认任务", len(defaults))
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	utils.LogSuccess("调度器已启动")

	s.wg.Add(1)
	go s.run()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
	utils.LogSuccess("调度器已停止")
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	// Check every minute; cron has minute granularity.
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkAndRunTasks()
		}
	}
}

func (s *Scheduler) checkAndRunTasks() {
	s.mu.RLock()
	now := time.Now()
	candidates := make([]*ScheduleTask, 0)
	for _, task := range s.tasks {
		if !task.Enabled || task.Running {
			continue
		}
		// chain tasks are triggered by their dependency, not by the cron tick directly.
		if task.ScheduleType == "chain" {
			continue
		}
		if s.shouldRun(task, now) {
			candidates = append(candidates, task)
		}
	}
	s.mu.RUnlock()

	for _, task := range candidates {
		go s.executeTask(task.ID, true)
	}
}

// shouldRun evaluates a 5-field cron expression against the current time.
// It returns true if the cron matches AND the task has not already run this minute.
func (s *Scheduler) shouldRun(task *ScheduleTask, now time.Time) bool {
	if task.CronExpr == "" {
		return false
	}
	matched, err := cronMatch(task.CronExpr, now)
	if err != nil || !matched {
		return false
	}
	// Avoid re-running within the same minute: if last run was within the
	// current minute window, skip.
	if task.LastRunTime != "" {
		if lastT, err := time.ParseInLocation("2006-01-02 15:04:05", task.LastRunTime, time.Local); err == nil {
			if lastT.Year() == now.Year() && lastT.Month() == now.Month() && lastT.Day() == now.Day() &&
				lastT.Hour() == now.Hour() && lastT.Minute() == now.Minute() {
				return false
			}
		}
	}
	return true
}

// executeTask runs a task by calling its configured endpoint via HTTP against
// the local server. If triggerChain is true and the task has dependents
// (other tasks with DependsOn == task.ID), they are executed sequentially
// after successful completion.
func (s *Scheduler) executeTask(taskID string, triggerChain bool) {
	s.mu.Lock()
	task, ok := s.tasks[taskID]
	if !ok {
		s.mu.Unlock()
		return
	}
	if task.Running {
		s.mu.Unlock()
		return
	}
	task.Running = true
	task.LastStatus = "running"
	endpoint := task.Endpoint
	method := task.Method
	params := task.Params
	s.mu.Unlock()

	start := time.Now()
	utils.LogSuccess("开始执行任务: %s (%s) -> %s %s", task.Name, taskID, method, endpoint)

	result, err := s.callEndpoint(method, endpoint, params)
	duration := time.Since(start).Seconds()
	end := time.Now()

	s.mu.Lock()
	task.Running = false
	task.LastRunTime = start.Format("2006-01-02 15:04:05")
	success := err == nil
	if success {
		task.LastStatus = "completed"
		task.LastError = ""
		utils.LogSuccess("任务执行完成: %s (%s)", task.Name, taskID)
	} else {
		task.LastStatus = "failed"
		task.LastError = err.Error()
		utils.LogError("任务执行失败: %s (%s) - %v", task.Name, taskID, err)
	}
	dependents := s.findDependents(taskID)
	s.mu.Unlock()

	// Persist execution status and history.
	_ = database.UpdateTaskStatus(taskID, task.LastStatus, task.LastError, duration, success)
	execID := fmt.Sprintf("exec_%d", start.UnixNano())
	statusStr := "completed"
	errMsg := ""
	resultStr := ""
	if !success {
		statusStr = "failed"
		errMsg = err.Error()
	} else if result != "" {
		// Truncate very long results for storage.
		if len(result) > 500 {
			resultStr = result[:500]
		} else {
			resultStr = result
		}
	}
	_ = database.RecordExecution(execID, taskID, statusStr, resultStr, errMsg, start, end)

	// Reload runtime stats from DB into the in-memory task.
	s.mu.Lock()
	if status, err := database.GetTaskStatusMap(); err == nil {
		if st, ok := status[taskID]; ok {
			task.TotalRuns = st.TotalRuns
			task.SuccessRuns = st.SuccessRuns
			task.FailRuns = st.FailRuns
			task.AvgDuration = st.AvgDuration
			task.SuccessRate = st.SuccessRate
		}
	}
	s.mu.Unlock()

	// Trigger dependent chain tasks on success.
	if triggerChain && success {
		for _, dep := range dependents {
			utils.LogSuccess("触发链式任务: %s -> %s", taskID, dep.ID)
			s.executeTask(dep.ID, true)
		}
	}
}

// findDependents returns tasks whose DependsOn == taskID. Caller must hold s.mu.
func (s *Scheduler) findDependents(taskID string) []*ScheduleTask {
	var out []*ScheduleTask
	for _, t := range s.tasks {
		if t.DependsOn == taskID {
			out = append(out, t)
		}
	}
	return out
}

// callEndpoint performs an HTTP request to the local server at the task's endpoint.
// For GET requests, params is treated as a query string (k=v&k=v). For POST/PUT,
// params is sent as the JSON body.
func (s *Scheduler) callEndpoint(method, endpoint, params string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("任务未配置 endpoint")
	}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", s.serverPort)
	urlStr := baseURL + endpoint

	if strings.EqualFold(method, "GET") || method == "" {
		if params != "" {
			sep := "?"
			if strings.Contains(endpoint, "?") {
				sep = "&"
			}
			urlStr = baseURL + endpoint + sep + params
		}
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return "", err
		}
		return s.doRequest(req)
	}

	body := strings.NewReader(params)
	if params == "" {
		body = strings.NewReader("{}")
	}
	req, err := http.NewRequest(strings.ToUpper(method), urlStr, body)
	if err != nil {
		return "", err
	}
	if params != "" && strings.HasPrefix(strings.TrimSpace(params), "{") {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return s.doRequest(req)
}

func (s *Scheduler) doRequest(req *http.Request) (string, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	bodyStr := string(respBody)
	if resp.StatusCode >= 400 {
		return bodyStr, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(bodyStr, 200))
	}
	// Inspect the JSON envelope: treat status == "error" as a failed task.
	var envelope struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if json.Unmarshal(respBody, &envelope) == nil && envelope.Status == "error" {
		return bodyStr, fmt.Errorf("任务返回错误: %s", envelope.Message)
	}
	return bodyStr, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// GetTasks returns all main tasks (excluding sub-tasks) in the frontend's
// expected nested format: {task_id, task_type, config, execution, sub_tasks, ...}.
func (s *Scheduler) GetTasks() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(s.tasks))
	for _, task := range s.tasks {
		if task.TaskType == "sub" {
			continue
		}
		result = append(result, s.toFrontendFormat(task))
	}
	return result
}

func (s *Scheduler) toFrontendFormat(task *ScheduleTask) map[string]interface{} {
	var subTasks []map[string]interface{}
	for _, t := range s.tasks {
		if t.ParentID == task.ID {
			subTasks = append(subTasks, s.toFrontendFormat(t))
		}
	}

	out := map[string]interface{}{
		"task_id":   task.ID,
		"task_type": task.TaskType,
		"config": map[string]interface{}{
			"name":           task.Name,
			"endpoint":       task.Endpoint,
			"method":         task.Method,
			"params":         parseParams(task.Params),
			"schedule_type":  task.ScheduleType,
			"schedule_time":  task.ScheduleTime,
			"interval_value": task.IntervalValue,
			"interval_unit":  task.IntervalUnit,
			"enabled":        task.Enabled,
		},
		"execution": map[string]interface{}{
			"last_run":     task.LastRunTime,
			"next_run":     task.NextRunTime,
			"status":       task.LastStatus,
			"success_rate": task.SuccessRate,
			"avg_duration": task.AvgDuration,
			"total_runs":   task.TotalRuns,
			"success_runs": task.SuccessRuns,
			"fail_runs":    task.FailRuns,
			"last_error":   task.LastError,
		},
		"created_at":     task.CreatedAt,
		"last_modified":  task.LastModified,
	}
	if task.ParentID != "" {
		out["parent_id"] = task.ParentID
	}
	if task.DependsOn != "" {
		out["depends_on"] = task.DependsOn
	}
	if len(subTasks) > 0 {
		out["sub_tasks"] = subTasks
	} else {
		out["sub_tasks"] = []interface{}{}
	}
	return out
}

func parseParams(paramsStr string) interface{} {
	if paramsStr == "" {
		return map[string]interface{}{}
	}
	var v interface{}
	if err := json.Unmarshal([]byte(paramsStr), &v); err == nil {
		return v
	}
	return paramsStr
}

func (s *Scheduler) GetTask(taskID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found")
	}
	return s.toFrontendFormat(task), nil
}

// CreateTaskFromConfig creates a task from the frontend payload shape:
// {task_id, task_type, config:{name,endpoint,method,params,schedule_type,
//  schedule_time,interval,unit,enabled}, depends_on}
func (s *Scheduler) CreateTaskFromConfig(payload map[string]interface{}) (map[string]interface{}, error) {
	taskID, _ := payload["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id 不能为空")
	}
	taskType, _ := payload["task_type"].(string)
	if taskType == "" {
		taskType = "main"
	}

	cfgMap, ok := payload["config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少 config 字段")
	}

	mt := database.MainTask{
		TaskID: taskID,
		Name:   getString(cfgMap, "name"),
		Endpoint: getString(cfgMap, "endpoint"),
		Method:   getString(cfgMap, "method"),
		Params:   getParamsString(cfgMap),
		ScheduleType: getString(cfgMap, "schedule_type"),
		ScheduleTime: getString(cfgMap, "schedule_time"),
		IntervalValue: getInt(cfgMap, "interval_value", "interval"),
		IntervalUnit:  getString(cfgMap, "interval_unit", "unit"),
		TaskType: taskType,
	}
	if enabled, ok := cfgMap["enabled"].(bool); ok && enabled {
		mt.Enabled = 1
	}
	if parentID, ok := payload["parent_id"].(string); ok {
		mt.ParentID = parentID
	}
	if dep, ok := payload["depends_on"].(string); ok {
		mt.DependsOn = dep
	}

	if err := database.UpsertMainTask(mt); err != nil {
		return nil, err
	}

	s.mu.Lock()
	task := convertTask(mt)
	s.tasks[taskID] = task
	s.mu.Unlock()

	return s.GetTask(taskID)
}

// UpdateTaskFromConfig updates a task from the frontend PUT payload.
func (s *Scheduler) UpdateTaskFromConfig(taskID string, payload map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	existing, ok := s.tasks[taskID]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("task not found")
	}

	mt := database.MainTask{
		TaskID:   existing.ID,
		Name:     existing.Name,
		Endpoint: existing.Endpoint,
		Method:   existing.Method,
		Params:   existing.Params,
		ScheduleType: existing.ScheduleType,
		ScheduleTime: existing.ScheduleTime,
		IntervalValue: existing.IntervalValue,
		IntervalUnit:  existing.IntervalUnit,
		Enabled: 0,
		TaskType: existing.TaskType,
		ParentID: existing.ParentID,
		DependsOn: existing.DependsOn,
		CreatedAt: existing.CreatedAt,
	}
	if existing.Enabled {
		mt.Enabled = 1
	}

	if cfg, ok := payload["config"].(map[string]interface{}); ok {
		if v := getString(cfg, "name"); v != "" {
			mt.Name = v
		}
		if v := getString(cfg, "endpoint"); v != "" {
			mt.Endpoint = v
		}
		if v := getString(cfg, "method"); v != "" {
			mt.Method = v
		}
		if p, ok := cfg["params"]; ok {
			mt.Params = paramsToString(p)
		}
		if v := getString(cfg, "schedule_type"); v != "" {
			mt.ScheduleType = v
		}
		if v, ok := cfg["schedule_time"].(string); ok {
			mt.ScheduleTime = v
		}
		if v := getInt(cfg, "interval_value", "interval"); v > 0 {
			mt.IntervalValue = v
		}
		if v := getString(cfg, "interval_unit", "unit"); v != "" {
			mt.IntervalUnit = v
		}
		if enabled, ok := cfg["enabled"].(bool); ok {
			if enabled {
				mt.Enabled = 1
			} else {
				mt.Enabled = 0
			}
		}
	}
	// Top-level enabled override (used by the enable endpoint indirectly).
	if enabled, ok := payload["enabled"].(bool); ok {
		if enabled {
			mt.Enabled = 1
		} else {
			mt.Enabled = 0
		}
	}

	if err := database.UpsertMainTask(mt); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.tasks[taskID] = convertTask(mt)
	s.mu.Unlock()

	return s.GetTask(taskID)
}

func (s *Scheduler) DeleteTask(taskID string) error {
	s.mu.Lock()
	if _, ok := s.tasks[taskID]; !ok {
		s.mu.Unlock()
		return fmt.Errorf("task not found")
	}
	// Remove in-memory copy + dependents.
	delete(s.tasks, taskID)
	for id, t := range s.tasks {
		if t.ParentID == taskID {
			delete(s.tasks, id)
		}
	}
	s.mu.Unlock()

	return database.DeleteMainTask(taskID)
}

func (s *Scheduler) SetTaskEnabled(taskID string, enabled bool) error {
	s.mu.Lock()
	if task, ok := s.tasks[taskID]; ok {
		task.Enabled = enabled
	}
	s.mu.Unlock()
	return database.SetTaskEnabled(taskID, enabled)
}

func (s *Scheduler) RunTask(taskID string) error {
	s.mu.RLock()
	task, ok := s.tasks[taskID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("task not found")
	}
	if task.Running {
		return fmt.Errorf("task is already running")
	}
	go s.executeTask(taskID, true)
	return nil
}

// GetTaskExecutions returns recent execution history for a task from the DB.
func (s *Scheduler) GetTaskExecutions(taskID string, limit int) []map[string]interface{} {
	records, err := database.GetExecutionHistory(taskID, limit)
	if err != nil {
		return []map[string]interface{}{}
	}
	return records
}

func (s *Scheduler) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalTasks := len(s.tasks)
	runningTasks := 0
	enabledTasks := 0
	for _, task := range s.tasks {
		if task.Running {
			runningTasks++
		}
		if task.Enabled {
			enabledTasks++
		}
	}
	return map[string]interface{}{
		"running":         s.running,
		"total_tasks":     totalTasks,
		"running_tasks":   runningTasks,
		"enabled_tasks":   enabledTasks,
	}
}

// ---- helpers for parsing the frontend config payload ----

func getString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func getInt(m map[string]interface{}, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case string:
				if i, err := strconv.Atoi(n); err == nil {
					return i
				}
			}
		}
	}
	return 0
}

func getParamsString(cfg map[string]interface{}) string {
	if p, ok := cfg["params"]; ok {
		return paramsToString(p)
	}
	return ""
}

func paramsToString(p interface{}) string {
	if p == nil {
		return ""
	}
	switch v := p.(type) {
	case string:
		return v
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(data)
	}
}
