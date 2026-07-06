package routers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/scheduler"
	"bilibili-history-go/services"
	"bilibili-history-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func RegisterConfigRoutes(r *gin.RouterGroup) {
	configGroup := r.Group("/config")
	{
		configGroup.GET("/shoutrrr", getShoutrrrConfig)
		configGroup.POST("/shoutrrr", saveShoutrrrConfig)
		configGroup.POST("/shoutrrr/test", testShoutrrrConfig)
		configGroup.GET("/server", getServerConfig)
		configGroup.POST("/server", saveServerConfig)
		configGroup.GET("/mcp-config", getMcpConfig)
		configGroup.POST("/mcp-config", saveMcpConfig)
	}
}

func RegisterSchedulerRoutes(r *gin.RouterGroup) {
	scheduler := r.Group("/scheduler")
	{
		scheduler.GET("/tasks", getSchedulerTasks)
		scheduler.POST("/tasks", addSchedulerTask)
		// task_id may contain "/" (e.g., "/fetch/bili-history").
		// Frontend URL-encodes slashes as %2F; Gin UseRawPath handles this.
		scheduler.PUT("/tasks/:id", updateSchedulerTask)
		scheduler.DELETE("/tasks/:id", deleteSchedulerTask)
		scheduler.POST("/tasks/:id/execute", runSchedulerTask)
		scheduler.POST("/tasks/:id/enable", enableSchedulerTask)
		// Frontend calls /tasks/history with task_id as a query param.
		scheduler.GET("/tasks/history", getTaskHistory)
		// Sub-task management (parent_id stored on the sub task).
		scheduler.POST("/tasks/:id/subtasks", addSubTask)
		scheduler.DELETE("/tasks/:id/subtasks/:subId", deleteSubTask)
		scheduler.GET("/status", getSchedulerStatus)
	}
}

func RegisterDataSyncRoutes(r *gin.RouterGroup) {
	dataSync := r.Group("/data_sync")
	{
		dataSync.GET("/status", getDataSyncStatus)
		dataSync.GET("/config", getDataSyncConfig)
		dataSync.POST("/config", updateDataSyncConfig)
		dataSync.GET("/sync-config", getSyncConfig)
		dataSync.POST("/sync-config", updateSyncConfig)
		dataSync.GET("/appearance-config", getAppearanceConfig)
		dataSync.POST("/appearance-config", updateAppearanceConfig)
		dataSync.POST("/check", checkDataIntegrity)
		dataSync.POST("/sync", syncData)
		dataSync.GET("/sync/result", getSyncResult)
		dataSync.GET("/report", getIntegrityReport)
	}
}

func RegisterExportRoutes(r *gin.RouterGroup) {
	export := r.Group("/export")
	{
		export.POST("/excel", exportToExcel)
	}
}

func RegisterImportRoutes(r *gin.RouterGroup) {
	importMysql := r.Group("/importMysql")
	{
		importMysql.POST("/start", importFromMysql)
		importMysql.GET("/status", getImportMysqlStatus)
	}

	importSqlite := r.Group("/importSqlite")
	{
		importSqlite.POST("/start", importFromSqlite)
		importSqlite.GET("/status", getImportSqliteStatus)
		importSqlite.POST("/import_data_sqlite", importFromSqlite)
	}
}

func RegisterCleanRoutes(r *gin.RouterGroup) {
	clean := r.Group("/clean")
	{
		clean.POST("/start", cleanData)
		clean.GET("/status", getCleanStatus)
	}
}

func RegisterLogRoutes(r *gin.RouterGroup) {
	log := r.Group("/log")
	{
		log.POST("/send", sendDailyReport)
		log.GET("/list", getLogList)
	}
}

func RegisterFetchRoutes(r *gin.RouterGroup) {
	fetch := r.Group("/fetch")
	{
		fetch.POST("/start", fetchBiliHistory)
		fetch.GET("/status", getFetchStatus)
		fetch.GET("/bili-history-realtime", fetchBiliHistoryRealtime)
		fetch.GET("/bili-history", fetchBiliHistoryFull)
		fetch.POST("/bili-history", fetchBiliHistoryFull)
		fetch.GET("/invalid-videos", getInvalidVideos)
	}
}

func fetchBiliHistoryRealtime(c *gin.Context) {
	result, err := services.FetchHistory(true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func fetchBiliHistoryFull(c *gin.Context) {
	result, err := services.FetchHistory(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func getInvalidVideos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	result, err := database.GetInvalidVideoList(page, size)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"videos": result.Records,
		"total":  result.Total,
		"page":   page,
		"size":   size,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

func RegisterDeleteRoutes(r *gin.RouterGroup) {
	delete := r.Group("/delete")
	{
		delete.POST("/history", deleteHistoryRecords)
		delete.DELETE("/batch-delete", batchDeleteHistory)
	}

	biliHistory := r.Group("/bilibili/history")
	{
		biliHistory.POST("/delete", deleteBiliHistory)
		biliHistory.GET("/status", getDeleteBiliStatus)
		biliHistory.DELETE("/single", deleteSingleBiliHistory)
		biliHistory.DELETE("/batch", deleteBatchBiliHistory)
	}
}

func batchDeleteHistory(c *gin.Context) {
	var req struct {
		Bvids []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	if len(req.Bvids) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("bvids 不能为空"))
		return
	}

	deletedCount := 0
	for _, bvid := range req.Bvids {
		if err := database.MarkVideoDeleted(bvid); err == nil {
			deletedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("已软删除 %d 条记录", deletedCount),
		"data":    map[string]interface{}{"deleted_count": deletedCount},
	})
}

func deleteSingleBiliHistory(c *gin.Context) {
	var req struct {
		Bvid string `json:"bvid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	if req.Bvid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("bvid 不能为空"))
		return
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("SESSDATA 未配置"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	if err := client.DeleteBiliHistory([]string{req.Bvid}); err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok {
			c.JSON(http.StatusOK, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("删除失败: code=%d, %s", apiErr.Code, apiErr.Message),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("删除失败: "+err.Error()))
		return
	}

	// 同步软删除本地记录
	_ = database.MarkVideoDeleted(req.Bvid)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "已删除B站历史记录",
	})
}

func deleteBatchBiliHistory(c *gin.Context) {
	var req struct {
		Bvids []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	if len(req.Bvids) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("bvids 不能为空"))
		return
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("SESSDATA 未配置"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	if err := client.DeleteBiliHistory(req.Bvids); err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok {
			c.JSON(http.StatusOK, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("删除失败: code=%d, %s", apiErr.Code, apiErr.Message),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("删除失败: "+err.Error()))
		return
	}

	// 同步软删除本地记录
	for _, bvid := range req.Bvids {
		_ = database.MarkVideoDeleted(bvid)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("已删除 %d 条B站历史记录", len(req.Bvids)),
	})
}

func RegisterInteractionRoutes(r *gin.RouterGroup) {
	interactions := r.Group("/interactions")
	{
		interactions.GET("/list", getInteractionRecords)
		interactions.POST("/sync", syncInteractionRecords)
	}
}

func RegisterTitleAnalyticsRoutes(r *gin.RouterGroup) {
	title := r.Group("/title")
	{
		title.GET("/stats", getTitleStats)
		title.GET("/patterns", getTitlePatterns)
		title.GET("/sentiment", getTitleSentiment)
		title.GET("/length", getTitleLengthAnalysis)
		title.GET("/trend", getTitleTrend)
	}
}

func getShoutrrrConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("配置加载失败"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(cfg.Shoutrrr))
}

func saveShoutrrrConfig(c *gin.Context) {
	var shoutrrrCfg config.ShoutrrrConfig
	if err := c.ShouldBindJSON(&shoutrrrCfg); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	cfg, _ := config.LoadConfig()
	cfg.Shoutrrr = shoutrrrCfg
	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("保存失败: "+err.Error()))
		return
	}

	services.ResetShoutrrrRouter()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Shoutrrr配置已保存",
	})
}

func testShoutrrrConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("配置加载失败"))
		return
	}

	if !cfg.Shoutrrr.Enabled || len(cfg.Shoutrrr.URLs) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Shoutrrr未启用或未配置URL"))
		return
	}

	if err := services.SendTestShoutrrr(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "测试通知发送失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "测试通知已发送，请检查各推送渠道",
	})
}

func getServerConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("配置加载失败"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(cfg.Server))
}

func saveServerConfig(c *gin.Context) {
	var serverCfg config.ServerConfig
	if err := c.ShouldBindJSON(&serverCfg); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	cfg, _ := config.LoadConfig()
	cfg.Server = serverCfg
	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("保存失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "服务器配置已保存",
	})
}

func getMcpConfig(c *gin.Context) {
	cfg, _ := config.LoadConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("配置加载失败"))
		return
	}

	host := "127.0.0.1"
	port := 8899
	if cfg.Server.Host == "0.0.0.0" || cfg.Server.Host == "::" || cfg.Server.Host == "" {
		host = "127.0.0.1"
	} else {
		host = cfg.Server.Host
	}
	if cfg.Server.Port > 0 {
		port = cfg.Server.Port
	}
	serverURL := fmt.Sprintf("http://%s:%d", host, port)

	mcpPath := cfg.Mcp.Path
	if mcpPath == "" {
		mcpPath = "/mcp"
	}
	mcpURL := serverURL + mcpPath + "/"

	tokenConfigured := cfg.Mcp.Token != ""

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"enabled":         cfg.Mcp.Enabled,
		"path":            mcpPath,
		"auth_enabled":    cfg.Mcp.AuthEnabled,
		"token":           cfg.Mcp.Token,
		"token_configured": tokenConfigured,
		"max_page_size":   cfg.Mcp.MaxPageSize,
		"server_url":      serverURL,
		"mcp_url":         mcpURL,
		"skill_content":   services.GetMCPSkillContent(cfg),
	})
}

func saveMcpConfig(c *gin.Context) {
	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("配置加载失败"))
		return
	}

	if body.Enabled != nil {
		cfg.Mcp.Enabled = *body.Enabled
	}

	// 首次启用时自动生成 Token
	if cfg.Mcp.Enabled && cfg.Mcp.Token == "" {
		cfg.Mcp.Token = services.GenerateToken()
	}

	// 设置默认值
	if cfg.Mcp.Path == "" {
		cfg.Mcp.Path = "/mcp"
	}
	if cfg.Mcp.MaxPageSize == 0 {
		cfg.Mcp.MaxPageSize = 100
	}

	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("保存失败: "+err.Error()))
		return
	}

	host := "127.0.0.1"
	port := 8899
	if cfg.Server.Host == "0.0.0.0" || cfg.Server.Host == "::" || cfg.Server.Host == "" {
		host = "127.0.0.1"
	} else {
		host = cfg.Server.Host
	}
	if cfg.Server.Port > 0 {
		port = cfg.Server.Port
	}
	serverURL := fmt.Sprintf("http://%s:%d", host, port)
	mcpURL := serverURL + cfg.Mcp.Path + "/"

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"message":         "MCP配置已保存",
		"enabled":         cfg.Mcp.Enabled,
		"path":            cfg.Mcp.Path,
		"auth_enabled":    cfg.Mcp.AuthEnabled,
		"token":           cfg.Mcp.Token,
		"token_configured": cfg.Mcp.Token != "",
		"max_page_size":   cfg.Mcp.MaxPageSize,
		"server_url":      serverURL,
		"mcp_url":         mcpURL,
		"skill_content":   services.GetMCPSkillContent(cfg),
		"restart_required": true,
	})
}

func getSchedulerTasks(c *gin.Context) {
	sched := scheduler.GetScheduler()
	taskID := c.Query("task_id")

	if taskID != "" {
		task, err := sched.GetTask(taskID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "任务不存在",
				"tasks":   []interface{}{},
				"total":   0,
			})
			return
		}
		utils.LogInfo("获取任务详情: %s", taskID)
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "获取任务信息成功",
			"tasks":   []interface{}{task},
			"total":   1,
		})
		return
	}

	tasks := sched.GetTasks()
	utils.LogInfo("获取任务列表: 共 %d 个主任务", len(tasks))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "获取任务信息成功",
		"tasks":   tasks,
		"total":   len(tasks),
	})
}

func addSchedulerTask(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	sched := scheduler.GetScheduler()
	taskInfo, err := sched.CreateTaskFromConfig(payload)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "创建任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "成功创建任务",
		"task_id":   taskInfo["task_id"],
		"task_info": taskInfo,
	})
}

func updateSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	sched := scheduler.GetScheduler()
	taskInfo, err := sched.UpdateTaskFromConfig(taskID, payload)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "更新任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "任务更新成功",
		"task_id":   taskID,
		"task_info": taskInfo,
	})
}

func deleteSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")

	sched := scheduler.GetScheduler()
	err := sched.DeleteTask(taskID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "删除任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务删除成功",
		"task_id": taskID,
	})
}

func runSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")
	utils.LogInfo("收到执行任务请求: task_id=%s", taskID)

	sched := scheduler.GetScheduler()
	err := sched.RunTask(taskID)
	if err != nil {
		utils.LogError("执行任务失败: task_id=%s, error=%v", taskID, err)
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "运行任务失败: " + err.Error(),
		})
		return
	}

	utils.LogSuccess("任务已启动: task_id=%s", taskID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务已启动",
		"task_id": taskID,
	})
}

func enableSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	sched := scheduler.GetScheduler()
	if err := sched.SetTaskEnabled(taskID, body.Enabled); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "切换任务状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务状态已更新",
		"task_id": taskID,
		"enabled": body.Enabled,
	})
}

func addSubTask(c *gin.Context) {
	parentID := c.Param("id")
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	payload["parent_id"] = parentID
	if tt, ok := payload["task_type"].(string); !ok || tt == "" {
		payload["task_type"] = "sub"
	}

	sched := scheduler.GetScheduler()
	taskInfo, err := sched.CreateTaskFromConfig(payload)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "创建子任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "成功创建子任务",
		"task_id":   taskInfo["task_id"],
		"task_info": taskInfo,
	})
}

func deleteSubTask(c *gin.Context) {
	subID := c.Param("subId")

	sched := scheduler.GetScheduler()
	if err := sched.DeleteTask(subID); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "删除子任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "子任务删除成功",
		"task_id": subID,
	})
}

func getTaskHistory(c *gin.Context) {
	taskID := c.Query("task_id")
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			pageSize = n
		}
	}

	sched := scheduler.GetScheduler()
	records := sched.GetTaskExecutions(taskID, pageSize)

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "获取任务执行历史成功",
		"history":     records,
		"total_count": len(records),
		"page":        1,
		"page_size":   pageSize,
	})
}

func getSchedulerStatus(c *gin.Context) {
	sched := scheduler.GetScheduler()
	status := sched.GetStatus()
	c.JSON(http.StatusOK, models.SuccessResponse(status))
}

func getDataSyncStatus(c *gin.Context) {
	status := services.GetDataSyncStatus()
	c.JSON(http.StatusOK, status)
}

func getDataSyncConfig(c *gin.Context) {
	cfg := config.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"check_on_startup": cfg.Server.DataIntegrity.CheckOnStartup,
	})
}

func updateDataSyncConfig(c *gin.Context) {
	var body struct {
		CheckOnStartup *bool `json:"check_on_startup"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	cfg, _ := config.LoadConfig()
	if body.CheckOnStartup != nil {
		cfg.Server.DataIntegrity.CheckOnStartup = *body.CheckOnStartup
	}
	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "配置已更新"})
}

func getSyncConfig(c *gin.Context) {
	cfg := config.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"success":                true,
		"sync_deleted":           cfg.Sync.SyncDeleted,
		"sync_delete_to_bilibili": cfg.Sync.SyncDeleteToBilibili,
	})
}

func updateSyncConfig(c *gin.Context) {
	var body struct {
		SyncDeleted           *bool `json:"sync_deleted"`
		SyncDeleteToBilibili  *bool `json:"sync_delete_to_bilibili"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	cfg, _ := config.LoadConfig()
	if body.SyncDeleted != nil {
		cfg.Sync.SyncDeleted = *body.SyncDeleted
	}
	if body.SyncDeleteToBilibili != nil {
		cfg.Sync.SyncDeleteToBilibili = *body.SyncDeleteToBilibili
	}
	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "同步配置已更新"})
}

func getAppearanceConfig(c *gin.Context) {
	cfg := config.GetConfig()
	darkMode := cfg.Appearance.DarkMode
	if darkMode == "" {
		darkMode = "system"
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"dark_mode": darkMode,
	})
}

func updateAppearanceConfig(c *gin.Context) {
	var body struct {
		DarkMode *string `json:"dark_mode"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	cfg, _ := config.LoadConfig()
	if body.DarkMode != nil {
		darkMode := *body.DarkMode
		if darkMode != "system" && darkMode != "light" && darkMode != "dark" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "dark_mode 必须是 system、light 或 dark"})
			return
		}
		cfg.Appearance.DarkMode = darkMode
	}
	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "外观配置已更新"})
}

func checkDataIntegrity(c *gin.Context) {
	var body struct {
		ForceCheck bool `json:"force_check"`
	}
	c.ShouldBindJSON(&body)

	result, err := services.RunIntegrityCheck(body.ForceCheck)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func syncData(c *gin.Context) {
	c.ShouldBindJSON(&struct{}{})

	result, err := services.RunSyncData()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func getSyncResult(c *gin.Context) {
	result := services.GetLastSyncResult()
	if result == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "暂无同步结果"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func getIntegrityReport(c *gin.Context) {
	data := services.GetIntegrityReportData()
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func exportToExcel(c *gin.Context) {
	var body struct {
		Years []int `json:"years"`
	}
	// Body is optional; if empty, export all years
	_ = c.ShouldBindJSON(&body)

	db := database.GetSQLiteDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("数据库不可用"))
		return
	}

	availableYears, err := db.GetAvailableYears()
	if err != nil || len(availableYears) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("未找到历史记录数据"))
		return
	}

	// Filter to requested years
	yearSet := make(map[int]bool)
	if len(body.Years) > 0 {
		for _, y := range body.Years {
			yearSet[y] = true
		}
	} else {
		for _, y := range availableYears {
			yearSet[y] = true
		}
	}

	f := excelize.NewFile()
	defer f.Close()

	firstSheet := true
	for _, year := range availableYears {
		if !yearSet[year] {
			continue
		}

		sheetName := fmt.Sprintf("%d", year)
		if firstSheet {
			f.SetSheetName("Sheet1", sheetName)
			firstSheet = false
		} else {
			f.NewSheet(sheetName)
		}

		// Header row
		headers := []string{"日期", "时间", "标题", "UP主", "分区", "时长(秒)", "BVID", "CID", "链接"}
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, h)
		}

		// Query records
		records, err := database.GetAllHistoryRecords(year)
		if err != nil {
			continue
		}

		for rowIdx, rec := range records {
			row := rowIdx + 2
			viewAt, _ := rec["view_at"].(int64)
			t := time.Unix(viewAt, 0)
			dateStr := t.Format("2006-01-02")
			timeStr := t.Format("15:04:05")

			title, _ := rec["title"].(string)
			authorName, _ := rec["author_name"].(string)
			tname, _ := rec["tname"].(string)
			duration, _ := rec["duration"].(int64)
			bvid, _ := rec["bvid"].(string)
			cid, _ := rec["cid"].(int64)
			link := "https://www.bilibili.com/video/" + bvid

			values := []interface{}{dateStr, timeStr, title, authorName, tname, duration, bvid, cid, link}
			for i, v := range values {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				f.SetCellValue(sheetName, cell, v)
			}
		}
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=bilibili_history.xlsx")
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("导出失败: "+err.Error()))
	}
}

func importFromMysql(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "MySQL导入功能待实现",
	})
}

func getImportMysqlStatus(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"status":  "idle",
		"message": "MySQL导入状态功能待实现",
	}))
}

var (
	sqliteImportStatus = map[string]interface{}{
		"status":  "idle",
		"message": "",
	}
)

func importFromSqlite(c *gin.Context) {
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
			"message": "暂无数据",
			"years":   []interface{}{},
			"total":   0,
		}))
		return
	}

	conn := db.GetDB()
	yearStats := make(map[string]int)
	total := 0
	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}
		var count int
		if err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count); err == nil {
			yearStats[fmt.Sprintf("%d", year)] = count
			total += count
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message":    "数据已直接写入数据库",
		"years":      yearStats,
		"total":      total,
		"year_count": len(years),
	}))
}

func getImportSqliteStatus(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(sqliteImportStatus))
}

func cleanData(c *gin.Context) {
	var options services.CleanOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	result, err := services.StartClean(options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("启动数据清洗失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}

func getCleanStatus(c *gin.Context) {
	status := services.GetCleanStatus()
	c.JSON(http.StatusOK, models.SuccessResponse(status))
}

func sendDailyReport(c *gin.Context) {
	stats := make(map[string]interface{})
	if err := c.ShouldBindJSON(&stats); err != nil {
		stats = make(map[string]interface{})
	}

	// 查找最新年份的热力图文件，构造附件 URL
	attachURL := findHeatmapAttachURL()

	err := services.SendDailyReport(stats, attachURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("发送每日报告失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "每日报告已发送",
	})
}

func findHeatmapAttachURL() string {
	cfg, _ := config.LoadConfig()

	outputDir := "output/heatmap"
	if cfg != nil && cfg.Heatmap.OutputDir != "" {
		outputDir = cfg.Heatmap.OutputDir
	} else if cfg != nil && cfg.OutputFolder != "" {
		outputDir = cfg.OutputFolder + "/heatmap"
	}

	// 从当前年份往前查找存在的热力图文件
	year := time.Now().Year()
	for y := year; y >= year-2; y-- {
		filePath := fmt.Sprintf("%s/heatmap_%d.png", outputDir, y)
		if _, err := os.Stat(filePath); err == nil {
			host := "127.0.0.1"
			port := 8899
			if cfg != nil {
				if cfg.Server.Host == "0.0.0.0" || cfg.Server.Host == "::" || cfg.Server.Host == "" {
					host = "127.0.0.1"
				} else {
					host = cfg.Server.Host
				}
				if cfg.Server.Port > 0 {
					port = cfg.Server.Port
				}
			}
			return fmt.Sprintf("http://%s:%d/api/heatmap/image?year=%d", host, port, y)
		}
	}

	return ""
}

func getLogList(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"logs":    []interface{}{},
		"message": "日志列表功能待实现",
	}))
}

func fetchBiliHistory(c *gin.Context) {
	skipExists := true
	if skipStr := c.Query("skip_exists"); skipStr == "false" {
		skipExists = false
	}

	result, err := services.FetchHistory(skipExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("启动历史记录获取失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}

func getFetchStatus(c *gin.Context) {
	status := services.GetFetchStatus()
	c.JSON(http.StatusOK, models.SuccessResponse(status))
}

func deleteHistoryRecords(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "删除历史记录功能待实现",
	})
}

func deleteBiliHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "删除B站历史记录功能待实现",
	})
}

func getDeleteBiliStatus(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"status":  "idle",
		"message": "删除B站历史记录状态功能待实现",
	}))
}

func getInteractionRecords(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"records": []interface{}{},
		"total":   0,
		"message": "互动记录功能待实现",
	}))
}

func syncInteractionRecords(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "互动记录同步功能待实现",
	})
}

func getTitleStats(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()
	var totalTitles int
	var totalLength int
	var minLength, maxLength int
	var minLengthTitle, maxLengthTitle string

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT title, LENGTH(title) as len
		FROM %s
		WHERE title != '' AND title IS NOT NULL
		ORDER BY len ASC
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	first := true
	for rows.Next() {
		var title string
		var length int
		if rows.Scan(&title, &length) == nil {
			totalTitles++
			totalLength += length
			if first {
				minLength = length
				minLengthTitle = title
				first = false
			}
			maxLength = length
			maxLengthTitle = title
		}
	}

	avgLength := 0.0
	if totalTitles > 0 {
		avgLength = float64(totalLength) / float64(totalTitles)
	}

	// 字数分布
分布 := map[string]int{
		"1-10":  0,
		"11-20": 0,
		"21-30": 0,
		"31-50": 0,
		"51+":   0,
	}

	rows2, err := conn.Query(fmt.Sprintf(`
		SELECT LENGTH(title) as len
		FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var length int
			if rows2.Scan(&length) == nil {
				switch {
				case length <= 10:
					分布["1-10"]++
				case length <= 20:
					分布["11-20"]++
				case length <= 30:
					分布["21-30"]++
				case length <= 50:
					分布["31-50"]++
				default:
					分布["51+"]++
				}
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":             year,
		"total_titles":     totalTitles,
		"avg_length":       avgLength,
		"min_length":       minLength,
		"min_length_title": minLengthTitle,
		"max_length":       maxLength,
		"max_length_title": maxLengthTitle,
		"length_distribution": 分布,
	}))
}

func getTitlePatterns(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()

	// 统计标题中包含的特殊字符模式
	patterns := map[string]int{
		"含问号":    0,
		"含感叹号":   0,
		"含数字":    0,
		"含英文":    0,
		"含【】":   0,
		"含「」":   0,
		"全大写英文":  0,
		"纯数字标题":  0,
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT title FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		if rows.Scan(&title) != nil {
			continue
		}
		if strings.ContainsAny(title, "?？") {
			patterns["含问号"]++
		}
		if strings.ContainsAny(title, "!！") {
			patterns["含感叹号"]++
		}
		if containsDigit(title) {
			patterns["含数字"]++
		}
		if containsAlpha(title) {
			patterns["含英文"]++
		}
		if strings.Contains(title, "【") || strings.Contains(title, "】") {
			patterns["含【】"]++
		}
		if strings.Contains(title, "「") || strings.Contains(title, "」") {
			patterns["含「」"]++
		}
		if isAllAlpha(title) {
			patterns["全大写英文"]++
		}
		if isAllDigit(title) {
			patterns["纯数字标题"]++
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":     year,
		"patterns": patterns,
	}))
}

func getTitleSentiment(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()

	positive := []string{"好看", "精彩", "有趣", "搞笑", "感动", "震撼", "推荐", "必看", "神作", "经典", "优秀", "完美", "厉害", "牛逼", "绝了", "爱了", "宝藏", "惊喜", "治愈", "温暖", "开心", "快乐", "爽", "赞", "棒"}
	negative := []string{"难看", "无聊", "垃圾", "差评", "失望", "浪费", "尴尬", "恶心", "愤怒", "讨厌", "失败", "错误", "翻车", "塌房", "退钱"}

	positiveCount := 0
	negativeCount := 0
	neutralCount := 0

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT title FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		if rows.Scan(&title) != nil {
			continue
		}
		isPositive := false
		isNegative := false
		for _, word := range positive {
			if strings.Contains(title, word) {
				isPositive = true
				break
			}
		}
		for _, word := range negative {
			if strings.Contains(title, word) {
				isNegative = true
				break
			}
		}
		switch {
		case isPositive && !isNegative:
			positiveCount++
		case isNegative && !isPositive:
			negativeCount++
		default:
			neutralCount++
		}
	}

	total := positiveCount + negativeCount + neutralCount
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":           year,
		"total":          total,
		"positive":       positiveCount,
		"negative":       negativeCount,
		"neutral":        neutralCount,
		"positive_ratio": float64(positiveCount) / float64(total),
		"negative_ratio": float64(negativeCount) / float64(total),
	}))
}

func getTitleLengthAnalysis(c *gin.Context) {
	yearStr := c.Query("year")
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	year := years[0]
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	tableName := fmt.Sprintf("bilibili_history_%d", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		c.JSON(http.StatusOK, models.ErrorResponse(fmt.Sprintf("未找到 %d 年数据", year)))
		return
	}

	conn := db.GetDB()

	buckets := []map[string]interface{}{
		{"range": "1-10", "min": 1, "max": 10, "count": 0, "avg_duration": 0, "total_duration": 0},
		{"range": "11-20", "min": 11, "max": 20, "count": 0, "avg_duration": 0, "total_duration": 0},
		{"range": "21-30", "min": 21, "max": 30, "count": 0, "avg_duration": 0, "total_duration": 0},
		{"range": "31-50", "min": 31, "max": 50, "count": 0, "avg_duration": 0, "total_duration": 0},
		{"range": "51+", "min": 51, "max": 9999, "count": 0, "avg_duration": 0, "total_duration": 0},
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT LENGTH(title) as len, COALESCE(duration, 0) as dur
		FROM %s
		WHERE title != '' AND title IS NOT NULL
	`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var length, duration int
		if rows.Scan(&length, &duration) != nil {
			continue
		}
		for _, b := range buckets {
			if length >= b["min"].(int) && length <= b["max"].(int) {
				b["count"] = b["count"].(int) + 1
				b["total_duration"] = b["total_duration"].(int) + duration
			}
		}
	}

	for _, b := range buckets {
		count := b["count"].(int)
		if count > 0 {
			b["avg_duration"] = b["total_duration"].(int) / count
		}
		delete(b, "min")
		delete(b, "max")
		delete(b, "total_duration")
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"year":     year,
		"analysis": buckets,
	}))
}

func getTitleTrend(c *gin.Context) {
	db := database.GetSQLiteDB()
	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("无可用数据"))
		return
	}

	conn := db.GetDB()
	type monthData struct {
		Month     string  `json:"month"`
		AvgLength float64 `json:"avg_length"`
		Count     int     `json:"count"`
	}

	var trend []monthData

	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		for month := 1; month <= 12; month++ {
			rows, err := conn.Query(fmt.Sprintf(`
				SELECT LENGTH(title) as len
				FROM %s
				WHERE title != '' AND title IS NOT NULL
				AND CAST(strftime('%%m', view_at, 'unixepoch', 'localtime') AS INTEGER) = ?
			`, tableName), month)
			if err != nil {
				continue
			}

			totalLen := 0
			count := 0
			for rows.Next() {
				var length int
				if rows.Scan(&length) == nil {
					totalLen += length
					count++
				}
			}
			rows.Close()

			if count > 0 {
				trend = append(trend, monthData{
					Month:     fmt.Sprintf("%d-%02d", year, month),
					AvgLength: float64(totalLen) / float64(count),
					Count:     count,
				})
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"trend": trend,
	}))
}

func containsDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsAlpha(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func isAllAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ') {
			return false
		}
	}
	return true
}

func isAllDigit(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}


