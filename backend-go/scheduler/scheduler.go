package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"bilibili-history-go/database"
	"bilibili-history-go/services"
	"bilibili-history-go/utils"
)

type TaskType string

const (
	TaskTypeFetchHistory TaskType = "fetch_history"
	TaskTypeCleanHistory TaskType = "clean_history"
	TaskTypeSyncData     TaskType = "sync_data"
	TaskTypeDailyReport  TaskType = "daily_report"
)

type TaskStatus string

const (
	TaskStatusIdle      TaskStatus = "idle"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusStopped   TaskStatus = "stopped"
)

type ScheduleTask struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        TaskType    `json:"type"`
	Endpoint    string      `json:"endpoint"`
	Method      string      `json:"method"`
	Params      string      `json:"params"`
	CronExpr    string      `json:"cron_expr"`
	Enabled     bool        `json:"enabled"`
	Status      TaskStatus  `json:"status"`
	LastRunTime int64       `json:"last_run_time,omitempty"`
	NextRunTime int64       `json:"next_run_time,omitempty"`
	LastResult  string      `json:"last_result,omitempty"`
	LastError   string      `json:"last_error,omitempty"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	TotalRuns   int         `json:"total_runs"`
	SuccessRuns int         `json:"success_runs"`
	FailRuns    int         `json:"fail_runs"`
	SuccessRate float64     `json:"success_rate"`
}

type TaskExecution struct {
	ID        string      `json:"id"`
	TaskID    string      `json:"task_id"`
	StartTime int64       `json:"start_time"`
	EndTime   int64       `json:"end_time,omitempty"`
	Status    TaskStatus  `json:"status"`
	Result    string      `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type Scheduler struct {
	tasks       map[string]*ScheduleTask
	executions  []TaskExecution
	mu          sync.RWMutex
	stopCh      chan struct{}
	running     bool
	wg          sync.WaitGroup
}

var (
	instance *Scheduler
	once     sync.Once
)

func GetScheduler() *Scheduler {
	once.Do(func() {
		instance = &Scheduler{
			tasks:      make(map[string]*ScheduleTask),
			executions: make([]TaskExecution, 0),
			stopCh:     make(chan struct{}),
			running:    false,
		}
		instance.loadTasks()
		instance.initDefaultTasks()
	})
	return instance
}

func (s *Scheduler) loadTasks() {
	mainTasks, err := database.GetMainTasks()
	if err != nil || len(mainTasks) == 0 {
		s.loadTasksFromJSON()
		return
	}

	statusMap, err := database.GetTaskStatusMap()
	if err != nil {
		statusMap = make(map[string]database.TaskStatus)
	}

	for _, mt := range mainTasks {
		task := &ScheduleTask{
			ID:        mt.TaskID,
			Name:      mt.Name,
			Endpoint:  mt.Endpoint,
			Method:    mt.Method,
			Params:    mt.Params,
			Enabled:   mt.Enabled == 1,
			Status:    TaskStatusIdle,
			CreatedAt: parseTimeToUnix(mt.CreatedAt),
			UpdatedAt: parseTimeToUnix(mt.LastModified),
		}

		task.CronExpr = buildCronExpr(mt)
		task.Type = detectTaskType(mt.Endpoint)

		if status, ok := statusMap[mt.TaskID]; ok {
			task.LastRunTime = parseTimeToUnix(status.LastRunTime)
			task.LastError = status.LastError
			task.TotalRuns = status.TotalRuns
			task.SuccessRuns = status.SuccessRuns
			task.FailRuns = status.FailRuns
			task.SuccessRate = status.SuccessRate
			if status.LastStatus == "running" {
				task.Status = TaskStatusRunning
			} else if status.LastStatus == "failed" {
				task.Status = TaskStatusFailed
			} else if status.LastStatus == "completed" {
				task.Status = TaskStatusCompleted
			}
		}

		s.tasks[task.ID] = task
	}
}

func parseTimeToUnix(timeStr string) int64 {
	if timeStr == "" {
		return 0
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, timeStr, time.Local)
		if err == nil {
			return t.Unix()
		}
	}
	return 0
}

func buildCronExpr(mt database.MainTask) string {
	switch mt.ScheduleType {
	case "daily":
		if mt.ScheduleTime != "" {
			parts := splitTime(mt.ScheduleTime)
			if len(parts) == 2 {
				return fmt.Sprintf("%s %s * * *", parts[1], parts[0])
			}
		}
		return "0 0 * * *"
	case "interval":
		switch mt.IntervalUnit {
		case "minutes":
			return fmt.Sprintf("*/%d * * * *", mt.IntervalValue)
		case "hours":
			return fmt.Sprintf("0 */%d * * *", mt.IntervalValue)
		case "days":
			return fmt.Sprintf("0 0 */%d * *", mt.IntervalValue)
		}
	case "weekly":
		return "0 0 * * 0"
	case "monthly":
		return "0 0 1 * *"
	}
	return "0 0 * * *"
}

func splitTime(timeStr string) []string {
	parts := make([]string, 0, 2)
	current := ""
	for _, c := range timeStr {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func detectTaskType(endpoint string) TaskType {
	switch {
	case contains(endpoint, "history") && contains(endpoint, "fetch"):
		return TaskTypeFetchHistory
	case contains(endpoint, "clean"):
		return TaskTypeCleanHistory
	case contains(endpoint, "sync"):
		return TaskTypeSyncData
	case contains(endpoint, "report"):
		return TaskTypeDailyReport
	default:
		return TaskTypeFetchHistory
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (s *Scheduler) loadTasksFromJSON() {
	tasksFile := utils.GetOutputPath("scheduler_tasks.json")
	if _, err := os.Stat(tasksFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(tasksFile)
	if err != nil {
		utils.LogError("Failed to read scheduler tasks: %v", err)
		return
	}

	var tasks []ScheduleTask
	if err := json.Unmarshal(data, &tasks); err != nil {
		utils.LogError("Failed to parse scheduler tasks: %v", err)
		return
	}

	for i := range tasks {
		task := &tasks[i]
		s.tasks[task.ID] = task
	}
}

func (s *Scheduler) saveTasks() {
	tasksFile := utils.GetOutputPath("scheduler_tasks.json")

	tasks := make([]ScheduleTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, *task)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		utils.LogError("Failed to marshal scheduler tasks: %v", err)
		return
	}

	if err := os.WriteFile(tasksFile, data, 0644); err != nil {
		utils.LogError("Failed to save scheduler tasks: %v", err)
	}
}

func (s *Scheduler) initDefaultTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()

	if _, ok := s.tasks["default_fetch_history"]; !ok {
		s.tasks["default_fetch_history"] = &ScheduleTask{
			ID:        "default_fetch_history",
			Name:      "自动获取历史记录",
			Type:      TaskTypeFetchHistory,
			CronExpr:  "0 */6 * * *",
			Enabled:   false,
			Status:    TaskStatusIdle,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	if _, ok := s.tasks["default_daily_report"]; !ok {
		s.tasks["default_daily_report"] = &ScheduleTask{
			ID:        "default_daily_report",
			Name:      "每日数据报告",
			Type:      TaskTypeDailyReport,
			CronExpr:  "0 8 * * *",
			Enabled:   false,
			Status:    TaskStatusIdle,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	s.saveTasks()
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
	defer s.mu.RUnlock()

	now := time.Now()

	for _, task := range s.tasks {
		if !task.Enabled || task.Status == TaskStatusRunning {
			continue
		}

		if s.shouldRun(task, now) {
			go s.executeTask(task.ID)
		}
	}
}

func (s *Scheduler) shouldRun(task *ScheduleTask, now time.Time) bool {
	if task.CronExpr == "" {
		return false
	}

	return false
}

func (s *Scheduler) executeTask(taskID string) {
	s.mu.Lock()
	task, ok := s.tasks[taskID]
	if !ok {
		s.mu.Unlock()
		return
	}

	if task.Status == TaskStatusRunning {
		s.mu.Unlock()
		return
	}

	task.Status = TaskStatusRunning
	task.LastRunTime = time.Now().Unix()
	task.UpdatedAt = time.Now().Unix()

	execution := TaskExecution{
		ID:        fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		TaskID:    taskID,
		StartTime: time.Now().Unix(),
		Status:    TaskStatusRunning,
	}
	s.executions = append(s.executions, execution)
	s.mu.Unlock()

	utils.LogSuccess("开始执行任务: %s (%s)", task.Name, taskID)

	var err error
	var result string

	switch task.Type {
	case TaskTypeFetchHistory:
		_, err = services.FetchHistory(true, false)
		if err == nil {
			result = "历史记录获取任务已启动"
		}
	default:
		result = fmt.Sprintf("任务类型 %s 暂未实现", task.Type)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task.LastRunTime = time.Now().Unix()
	task.UpdatedAt = time.Now().Unix()

	if err != nil {
		task.Status = TaskStatusFailed
		task.LastError = err.Error()
		execution.Status = TaskStatusFailed
		execution.Error = err.Error()
		utils.LogError("任务执行失败: %s (%s) - %v", task.Name, taskID, err)
	} else {
		task.Status = TaskStatusCompleted
		task.LastResult = result
		execution.Status = TaskStatusCompleted
		execution.Result = result
		utils.LogSuccess("任务执行完成: %s (%s)", task.Name, taskID)
	}

	execution.EndTime = time.Now().Unix()

	for i := range s.executions {
		if s.executions[i].ID == execution.ID {
			s.executions[i] = execution
			break
		}
	}

	s.saveTasks()
}

func (s *Scheduler) GetTasks() []ScheduleTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]ScheduleTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, *task)
	}

	return tasks
}

func (s *Scheduler) GetTask(taskID string) (*ScheduleTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found")
	}

	taskCopy := *task
	return &taskCopy, nil
}

func (s *Scheduler) CreateTask(task *ScheduleTask) (*ScheduleTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	task.Status = TaskStatusIdle
	task.CreatedAt = time.Now().Unix()
	task.UpdatedAt = time.Now().Unix()

	s.tasks[task.ID] = task
	s.saveTasks()

	return task, nil
}

func (s *Scheduler) UpdateTask(taskID string, updates map[string]interface{}) (*ScheduleTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found")
	}

	if name, ok := updates["name"].(string); ok {
		task.Name = name
	}
	if cronExpr, ok := updates["cron_expr"].(string); ok {
		task.CronExpr = cronExpr
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		task.Enabled = enabled
	}

	task.UpdatedAt = time.Now().Unix()
	s.saveTasks()

	taskCopy := *task
	return &taskCopy, nil
}

func (s *Scheduler) DeleteTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[taskID]; !ok {
		return fmt.Errorf("task not found")
	}

	delete(s.tasks, taskID)
	s.saveTasks()

	return nil
}

func (s *Scheduler) RunTask(taskID string) error {
	s.mu.RLock()
	task, ok := s.tasks[taskID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("task not found")
	}

	if task.Status == TaskStatusRunning {
		return fmt.Errorf("task is already running")
	}

	go s.executeTask(taskID)
	return nil
}

func (s *Scheduler) GetTaskExecutions(taskID string, limit int) []TaskExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []TaskExecution
	count := 0

	for i := len(s.executions) - 1; i >= 0; i-- {
		if s.executions[i].TaskID == taskID {
			result = append(result, s.executions[i])
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return result
}

func (s *Scheduler) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalTasks := len(s.tasks)
	runningTasks := 0
	enabledTasks := 0

	for _, task := range s.tasks {
		if task.Status == TaskStatusRunning {
			runningTasks++
		}
		if task.Enabled {
			enabledTasks++
		}
	}

	return map[string]interface{}{
		"running":        s.running,
		"total_tasks":    totalTasks,
		"running_tasks":  runningTasks,
		"enabled_tasks":  enabledTasks,
		"total_executions": len(s.executions),
	}
}
