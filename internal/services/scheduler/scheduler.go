package scheduler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/db"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

// Task represents a scheduled task.
type Task struct {
	ID             string            `json:"id" yaml:"id"`
	Name           string            `json:"name" yaml:"name"`
	Endpoint       string            `json:"endpoint" yaml:"endpoint"`
	Method         string            `json:"method" yaml:"method"`
	Params         map[string]interface{} `json:"params" yaml:"params"`
	Requires       []string          `json:"requires" yaml:"requires"`
	ScheduleType   string            `json:"schedule_type" yaml:"type"`
	ScheduleTime   string            `json:"schedule_time" yaml:"time"`
	IntervalValue  int               `json:"interval_value" yaml:"interval_value"`
	IntervalUnit   string            `json:"interval_unit" yaml:"interval_unit"`
	Enabled        bool              `json:"enabled" yaml:"enabled"`
}

// TaskExecution represents a task execution record.
type TaskExecution struct {
	TaskID    string     `json:"task_id"`
	TaskName  string     `json:"task_name"`
	Status    string     `json:"status"` // "running", "success", "failed"
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// SchedulerConfig is the scheduler YAML config structure.
type SchedulerConfig struct {
	BaseURL      string                 `yaml:"base_url"`
	ErrorHandling ErrorHandlingConfig   `yaml:"error_handling"`
	Scheduler    SchedulerSettings      `yaml:"scheduler"`
	Tasks        map[string]TaskConfig   `yaml:"tasks"`
}

type ErrorHandlingConfig struct {
	NotifyOnFailure bool `yaml:"notify_on_failure"`
	StopOnFailure   bool `yaml:"stop_on_failure"`
}

type SchedulerSettings struct {
	LogLevel string              `yaml:"log_level"`
	Retry    RetryConfig         `yaml:"retry"`
}

type RetryConfig struct {
	Delay       int `yaml:"delay"`
	MaxAttempts int `yaml:"max_attempts"`
}

type TaskConfig struct {
	Endpoint string                 `json:"endpoint" yaml:"endpoint"`
	Method   string                 `json:"method" yaml:"method"`
	Name     string                 `json:"name" yaml:"name"`
	Params   map[string]interface{} `json:"params" yaml:"params"`
	Requires []string               `json:"requires" yaml:"requires"`
	Schedule ScheduleConfig         `json:"schedule" yaml:"schedule"`
}

type ScheduleConfig struct {
	Type          string `yaml:"type"`
	Time          string `yaml:"time"`
	IntervalValue int    `yaml:"interval_value"`
	IntervalUnit  string `yaml:"interval_unit"`
}

// Manager is the scheduler manager.
type Manager struct {
	cron        *cron.Cron
	tasks       map[string]*Task
	executions  []TaskExecution
	mu          sync.RWMutex
	baseURL     string
	httpClient  *http.Client
	isRunning   bool
	config      *SchedulerConfig
}

// NewManager creates a new scheduler manager.
func NewManager() (*Manager, error) {
	cfg, err := loadSchedulerConfig()
	if err != nil {
		log.Printf("Warning: failed to load scheduler config: %v", err)
		cfg = &SchedulerConfig{}
	}

	m := &Manager{
		cron:       cron.New(cron.WithSeconds()),
		tasks:      make(map[string]*Task),
		httpClient: &http.Client{Timeout: 10 * time.Minute},
		config:     cfg,
	}

	// Set base URL from config
	if cfg.BaseURL != "" {
		m.baseURL = cfg.BaseURL
	} else {
		appCfg, _ := config.LoadConfig()
		if appCfg != nil {
			m.baseURL = fmt.Sprintf("http://127.0.0.1:%d", appCfg.Server.Port)
		}
	}

	// Load tasks from config (fallback when DB is unavailable/empty)
	if cfg.Tasks != nil {
		for id, tc := range cfg.Tasks {
			task := &Task{
				ID:            id,
				Name:          tc.Name,
				Endpoint:      tc.Endpoint,
				Method:        tc.Method,
				Params:        tc.Params,
				Requires:      tc.Requires,
				ScheduleType:  tc.Schedule.Type,
				ScheduleTime:  tc.Schedule.Time,
				IntervalValue: tc.Schedule.IntervalValue,
				IntervalUnit:  tc.Schedule.IntervalUnit,
				Enabled:       true,
			}
			m.tasks[id] = task
		}
	}

	// Load tasks from SQLite database (overrides YAML tasks with the same ID)
	if err := m.loadTasksFromDB(); err != nil {
		log.Printf("Warning: failed to load tasks from database: %v", err)
	}

	return m, nil
}

// loadTasksFromDB loads tasks from the SQLite scheduler database
// (output/database/scheduler.db), which is the same database written to by
// the Python API. Tasks loaded from the DB override any YAML tasks that
// share the same ID. If the database is missing or empty, the existing
// YAML-loaded tasks are left untouched.
func (m *Manager) loadTasksFromDB() error {
	database, err := db.OpenSchedulerDB()
	if err != nil {
		return fmt.Errorf("open scheduler db: %w", err)
	}

	// Load main tasks
	mainRows, err := database.Query(`
		SELECT task_id, name, endpoint, method, params, schedule_type,
		       schedule_time, interval_value, interval_unit, enabled
		FROM main_tasks
	`)
	if err != nil {
		return fmt.Errorf("query main_tasks: %w", err)
	}

	mainCount := 0
	for mainRows.Next() {
		var (
			id            string
			name          string
			endpoint      string
			method        string
			paramsJSON    sql.NullString
			scheduleType  string
			scheduleTime  sql.NullString
			intervalValue sql.NullInt64
			intervalUnit  sql.NullString
			enabled       int
		)
		if err := mainRows.Scan(
			&id, &name, &endpoint, &method, &paramsJSON, &scheduleType,
			&scheduleTime, &intervalValue, &intervalUnit, &enabled,
		); err != nil {
			log.Printf("Warning: failed to scan main task row: %v", err)
			continue
		}

		task := &Task{
			ID:            id,
			Name:          name,
			Endpoint:      endpoint,
			Method:        method,
			ScheduleType:  scheduleType,
			ScheduleTime:  scheduleTime.String,
			IntervalValue: int(intervalValue.Int64),
			IntervalUnit:  intervalUnit.String,
			Enabled:       enabled != 0,
		}

		if paramsJSON.Valid && paramsJSON.String != "" {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(paramsJSON.String), &params); err == nil {
				task.Params = params
			} else {
				log.Printf("Warning: failed to unmarshal params for task %s: %v", id, err)
			}
		}

		m.tasks[id] = task
		mainCount++
	}
	if err := mainRows.Err(); err != nil {
		log.Printf("Warning: error iterating main tasks: %v", err)
	}
	mainRows.Close()

	// Load sub tasks as chain tasks linked to their parent task.
	subRows, err := database.Query(`
		SELECT task_id, parent_id, name, endpoint, method, params, enabled
		FROM sub_tasks
	`)
	if err != nil {
		return fmt.Errorf("query sub_tasks: %w", err)
	}
	defer subRows.Close()

	subCount := 0
	for subRows.Next() {
		var (
			id         string
			parentID   string
			name       string
			endpoint   string
			method     string
			paramsJSON sql.NullString
			enabled    int
		)
		if err := subRows.Scan(
			&id, &parentID, &name, &endpoint, &method, &paramsJSON, &enabled,
		); err != nil {
			log.Printf("Warning: failed to scan sub task row: %v", err)
			continue
		}

		task := &Task{
			ID:           id,
			Name:         name,
			Endpoint:     endpoint,
			Method:       method,
			ScheduleType: "chain",
			Requires:     []string{parentID},
			Enabled:      enabled != 0,
		}

		if paramsJSON.Valid && paramsJSON.String != "" {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(paramsJSON.String), &params); err == nil {
				task.Params = params
			} else {
				log.Printf("Warning: failed to unmarshal params for sub task %s: %v", id, err)
			}
		}

		m.tasks[id] = task
		subCount++
	}
	if err := subRows.Err(); err != nil {
		log.Printf("Warning: error iterating sub tasks: %v", err)
	}

	log.Printf("Loaded %d main tasks and %d sub tasks from scheduler database", mainCount, subCount)
	return nil
}

// Start starts the scheduler.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, task := range m.tasks {
		if !task.Enabled {
			continue
		}

		entryID, err := m.addTaskToCron(id, task)
		if err != nil {
			log.Printf("Failed to schedule task %s: %v", id, err)
			continue
		}
		log.Printf("Scheduled task: %s (ID: %d)", task.Name, entryID)
	}

	m.cron.Start()
	m.isRunning = true
	log.Println("Scheduler started")
	return nil
}

// Stop stops the scheduler.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		ctx := m.cron.Stop()
		<-ctx.Done()
		m.isRunning = false
		log.Println("Scheduler stopped")
	}
}

// GetTasks returns all registered tasks.
func (m *Manager) GetTasks() []Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []Task
	for _, t := range m.tasks {
		tasks = append(tasks, *t)
	}
	return tasks
}

// GetExecutions returns recent task executions.
func (m *Manager) GetExecutions() []TaskExecution {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.executions
}

// AddTask adds a new task to the scheduler.
func (m *Manager) AddTask(task *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks[task.ID] = task

	if task.Enabled && task.ScheduleType != "chain" {
		m.addTaskToCron(task.ID, task)
	}
}

// UpdateTask updates an existing task.
func (m *Manager) UpdateTask(taskID string, updates *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return
	}

	if updates.Name != "" {
		task.Name = updates.Name
	}
	if updates.Endpoint != "" {
		task.Endpoint = updates.Endpoint
	}
	if updates.Method != "" {
		task.Method = updates.Method
	}
	if updates.ScheduleType != "" {
		task.ScheduleType = updates.ScheduleType
	}
	if updates.ScheduleTime != "" {
		task.ScheduleTime = updates.ScheduleTime
	}
	if updates.IntervalValue > 0 {
		task.IntervalValue = updates.IntervalValue
	}
	if updates.IntervalUnit != "" {
		task.IntervalUnit = updates.IntervalUnit
	}
}

// DeleteTask removes a task from the scheduler.
func (m *Manager) DeleteTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tasks, taskID)
}

// RunTaskNow runs a task immediately.
func (m *Manager) RunTaskNow(taskID string) error {
	m.mu.RLock()
	task, ok := m.tasks[taskID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	go m.executeTask(taskID, task)
	return nil
}

func (m *Manager) addTaskToCron(id string, task *Task) (cron.EntryID, error) {
	switch task.ScheduleType {
	case "daily":
		// Parse "HH:MM" format
		if task.ScheduleTime == "" {
			task.ScheduleTime = "00:00"
		}
		hour, min := 0, 0
		fmt.Sscanf(task.ScheduleTime, "%d:%d", &hour, &min)
		schedule := fmt.Sprintf("0 %d %d * * *", min, hour)
		return m.cron.AddFunc(schedule, func() {
			m.executeTask(id, task)
		})

	case "interval":
		var schedule string
		switch task.IntervalUnit {
		case "minutes":
			schedule = fmt.Sprintf("0 */%d * * * *", task.IntervalValue)
		case "hours":
			schedule = fmt.Sprintf("0 0 */%d * * *", task.IntervalValue)
		default:
			schedule = fmt.Sprintf("0 */%d * * * *", task.IntervalValue)
		}
		return m.cron.AddFunc(schedule, func() {
			m.executeTask(id, task)
		})

	case "chain":
		// Chain tasks are triggered by their dependencies
		return 0, nil

	default:
		return 0, fmt.Errorf("unknown schedule type: %s", task.ScheduleType)
	}
}

func (m *Manager) executeTask(id string, task *Task) {
	execution := TaskExecution{
		TaskID:    id,
		TaskName:  task.Name,
		Status:    "running",
		StartTime: time.Now(),
	}

	log.Printf("Executing task: %s (%s)", task.Name, task.Endpoint)

	// Check dependencies
	for _, dep := range task.Requires {
		if !m.isDependencyMet(dep) {
			log.Printf("Task %s skipped: dependency %s not met", task.Name, dep)
			execution.Status = "failed"
			execution.Error = fmt.Sprintf("dependency %s not met", dep)
			m.recordExecution(execution)
			return
		}
	}

	// Execute HTTP request
	url := m.baseURL + task.Endpoint
	var body io.Reader

	if task.Method == "POST" && task.Params != nil {
		jsonData, _ := json.Marshal(task.Params)
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(task.Method, url, body)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		m.recordExecution(execution)
		return
	}

	if task.Method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		m.recordExecution(execution)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		execution.Status = "failed"
		execution.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		execution.Status = "success"
	}

	now := time.Now()
	execution.EndTime = &now
	m.recordExecution(execution)

	// Trigger chain tasks
	m.triggerChainTasks(id)
}

func (m *Manager) isDependencyMet(depID string) bool {
	// Check if the dependency task has been executed successfully
	for i := len(m.executions) - 1; i >= 0; i-- {
		if m.executions[i].TaskID == depID && m.executions[i].Status == "success" {
			return true
		}
	}
	return false
}

func (m *Manager) triggerChainTasks(completedTaskID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for id, task := range m.tasks {
		if task.ScheduleType != "chain" {
			continue
		}
		for _, dep := range task.Requires {
			if dep == completedTaskID {
				go m.executeTask(id, task)
				break
			}
		}
	}
}

func (m *Manager) recordExecution(execution TaskExecution) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.executions = append(m.executions, execution)

	// Keep only last 100 executions
	if len(m.executions) > 100 {
		m.executions = m.executions[len(m.executions)-100:]
	}
}

func loadSchedulerConfig() (*SchedulerConfig, error) {
	basePath := config.GetBasePath()
	configPath := filepath.Join(basePath, "config", "scheduler_config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg SchedulerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
