package handlers

import (
	"net/http"

	"bili-history/internal/services/scheduler"

	"github.com/gin-gonic/gin"
)

var schedulerManager *scheduler.Manager

func init() {
	var err error
	schedulerManager, err = scheduler.NewManager()
	if err != nil {
		// Non-fatal: scheduler won't run but app starts
		return
	}
	schedulerManager.Start()
}

// GetSchedulerTasks returns all scheduler tasks.
func GetSchedulerTasks(c *gin.Context) {
	tasks := schedulerManager.GetTasks()
	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

// GetSchedulerExecutions returns recent task executions.
func GetSchedulerExecutions(c *gin.Context) {
	executions := schedulerManager.GetExecutions()
	c.JSON(http.StatusOK, gin.H{"data": executions})
}

// RunSchedulerTask runs a task immediately.
func RunSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	if err := schedulerManager.RunTaskNow(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "started",
		"task_id": taskID,
	})
}

// CreateSchedulerTask creates a new scheduler task.
func CreateSchedulerTask(c *gin.Context) {
	var req struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Endpoint       string `json:"endpoint"`
		Method         string `json:"method"`
		ScheduleType   string `json:"schedule_type"`
		ScheduleTime   string `json:"schedule_time"`
		IntervalValue  int    `json:"interval_value"`
		IntervalUnit   string `json:"interval_unit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	task := &scheduler.Task{
		ID:            req.ID,
		Name:          req.Name,
		Endpoint:      req.Endpoint,
		Method:        req.Method,
		ScheduleType:  req.ScheduleType,
		ScheduleTime:  req.ScheduleTime,
		IntervalValue: req.IntervalValue,
		IntervalUnit:  req.IntervalUnit,
		Enabled:       true,
	}

	schedulerManager.AddTask(task)

	c.JSON(http.StatusOK, gin.H{
		"status":  "created",
		"task_id": req.ID,
	})
}

// UpdateSchedulerTask updates a scheduler task.
func UpdateSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	var req struct {
		Name          string `json:"name"`
		Endpoint      string `json:"endpoint"`
		Method        string `json:"method"`
		ScheduleType  string `json:"schedule_type"`
		ScheduleTime  string `json:"schedule_time"`
		IntervalValue int    `json:"interval_value"`
		IntervalUnit  string `json:"interval_unit"`
		Enabled       *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	schedulerManager.UpdateTask(taskID, &scheduler.Task{
		Name:          req.Name,
		Endpoint:      req.Endpoint,
		Method:        req.Method,
		ScheduleType:  req.ScheduleType,
		ScheduleTime:  req.ScheduleTime,
		IntervalValue: req.IntervalValue,
		IntervalUnit:  req.IntervalUnit,
	})

	c.JSON(http.StatusOK, gin.H{
		"status":  "updated",
		"task_id": taskID,
	})
}

// DeleteSchedulerTask deletes a scheduler task.
func DeleteSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	schedulerManager.DeleteTask(taskID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "deleted",
		"task_id": taskID,
	})
}

// GetSchedulerEndpoints returns available API endpoints for scheduler.
func GetSchedulerEndpoints(c *gin.Context) {
	endpoints := []map[string]string{
		{"path": "/fetch/bili-history", "method": "GET", "name": "获取历史记录"},
		{"path": "/importSqlite/import_data_sqlite", "method": "POST", "name": "导入数据"},
		{"path": "/analysis/analyze", "method": "POST", "name": "分析数据"},
		{"path": "/heatmap/generate_heatmap", "method": "POST", "name": "生成热力图"},
		{"path": "/log/send-email", "method": "POST", "name": "发送邮件"},
		{"path": "/bilibili/popular/all", "method": "GET", "name": "获取热门视频"},
		{"path": "/like/list", "method": "GET", "name": "同步点赞"},
		{"path": "/login/check-and-notify", "method": "GET", "name": "SESSDATA检查"},
	}

	c.JSON(http.StatusOK, gin.H{"data": endpoints})
}

// GetTaskHistory returns task execution history.
func GetTaskHistory(c *gin.Context) {
	executions := schedulerManager.GetExecutions()
	c.JSON(http.StatusOK, gin.H{"data": executions})
}
