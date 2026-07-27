package routers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

var asyncExportFiles sync.Map

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
		scheduler.PUT("/tasks/:id", updateSchedulerTask)
		scheduler.DELETE("/tasks/:id", deleteSchedulerTask)
		scheduler.POST("/tasks/:id/execute", runSchedulerTask)
		scheduler.POST("/tasks/:id/enable", enableSchedulerTask)
		scheduler.GET("/tasks/history", getTaskHistory)
		scheduler.POST("/tasks/:id/subtasks", addSubTask)
		scheduler.DELETE("/tasks/:id/subtasks/:subId", deleteSubTask)
		scheduler.GET("/status", getSchedulerStatus)
	}

	r.GET("/task/:id/status", getAsyncTaskStatus)
}

func getAsyncTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("task_id 不能为空"))
		return
	}

	status := scheduler.GetAsyncTaskStatus(taskID)
	if status == nil {
		c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
			"task_id": taskID,
			"status":  "not_found",
			"message": "任务不存在或已过期",
		}))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(status))
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
		export.GET("/download/:task_id", downloadExportFile)
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
	taskID := c.Query("task_id")
	result, err := services.FetchHistorySync(taskID, true)
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
	taskID := c.Query("task_id")
	result, err := services.FetchHistorySync(taskID, false)
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
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("读取请求体失败"))
		return
	}

	var bvids []string

	// 尝试格式1: {"bvids": ["bvid1", "bvid2"]}
	var reqWithKey struct {
		Bvids []string `json:"bvids"`
	}
	if json.Unmarshal(body, &reqWithKey) == nil && len(reqWithKey.Bvids) > 0 {
		bvids = reqWithKey.Bvids
	} else {
		// 尝试格式2: [{"bvid": "bvid1", "view_at": 123}, ...]
		var items []struct {
			Bvid   string `json:"bvid"`
			ViewAt int64  `json:"view_at"`
		}
		if err := json.Unmarshal(body, &items); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
			return
		}
		for _, item := range items {
			if item.Bvid != "" {
				bvids = append(bvids, item.Bvid)
			}
		}
	}

	if len(bvids) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("bvids 不能为空"))
		return
	}

	deletedCount := 0
	for _, bvid := range bvids {
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

// extractOidFromKid 从 kid 格式中提取 oid
// kid 格式：business_oid，例如 archive_123456
func extractOidFromKid(kid string) string {
	parts := strings.SplitN(kid, "_", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func deleteSingleBiliHistory(c *gin.Context) {
	kid := c.Query("kid")
	if kid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("kid 不能为空"))
		return
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("SESSDATA 未配置"))
		return
	}

	if cfg.BiliJct == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("BiliJct (csrf token) 未配置，无法删除B站历史记录"))
		return
	}

	// 从 kid 中提取 oid，再查询 bvid
	oid := extractOidFromKid(kid)
	if oid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的 kid 格式"))
		return
	}

	bvid, err := database.GetBvidByOid(oid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("未找到 oid=%s 对应的视频记录", oid),
		})
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	if err := client.DeleteBiliHistory([]string{bvid}); err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok {
			c.JSON(http.StatusOK, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("删除失败: code=%d, %s", apiErr.Code, apiErr.Message),
			})
			return
		}
		// 返回更详细的错误信息帮助排查
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("删除B站历史记录失败: %s, bvid=%s, kid=%s", err.Error(), bvid, kid),
		})
		return
	}

	// 同步软删除本地记录
	_ = database.MarkVideoDeleted(bvid)

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

	taskID, _ := scheduler.StartAsyncTask("数据同步")

	go doSyncData(taskID)

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"task_id": taskID,
		"message": "数据同步任务已启动，正在后台执行",
	}))
}

func doSyncData(taskID string) {
	success := false
	resultMsg := ""
	errMsg := ""

	defer func() {
		scheduler.CompleteAsyncTask(taskID, success, resultMsg, errMsg)
	}()

	result, err := services.RunSyncData()
	if err != nil {
		errMsg = err.Error()
		return
	}

	success = true
	resultBytes, _ := json.Marshal(result)
	resultMsg = string(resultBytes)
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
	_ = c.ShouldBindJSON(&body)

	db := database.GetSQLiteDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("数据库不可用"))
		return
	}

	taskID, _ := scheduler.StartAsyncTask("导出Excel")

	go doExportToExcel(taskID, body.Years)

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"task_id": taskID,
		"message": "Excel导出任务已启动，正在后台执行",
	}))
}

func doExportToExcel(taskID string, years []int) {
	success := false
	resultMsg := ""
	errMsg := ""

	defer func() {
		scheduler.CompleteAsyncTask(taskID, success, resultMsg, errMsg)
	}()

	db := database.GetSQLiteDB()
	if db == nil {
		errMsg = "数据库不可用"
		return
	}

	availableYears, err := db.GetAvailableYears()
	if err != nil || len(availableYears) == 0 {
		errMsg = "未找到历史记录数据"
		return
	}

	yearSet := make(map[int]bool)
	if len(years) > 0 {
		for _, y := range years {
			yearSet[y] = true
		}
	} else {
		for _, y := range availableYears {
			yearSet[y] = true
		}
	}

	f := excelize.NewFile()
	defer f.Close()

	totalRows := 0
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

		headers := []string{"日期", "时间", "标题", "UP主", "分区", "时长(秒)", "BVID", "CID", "链接"}
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, h)
		}

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
			totalRows++
		}
	}

	tmpDir := os.TempDir()
	fileName := fmt.Sprintf("bilibili_history_%s.xlsx", taskID)
	filePath := filepath.Join(tmpDir, fileName)

	if err := f.SaveAs(filePath); err != nil {
		errMsg = "保存Excel失败: " + err.Error()
		return
	}

	success = true
	resultBytes, _ := json.Marshal(map[string]interface{}{
		"total_rows": totalRows,
		"download_url": fmt.Sprintf("/api/export/download/%s", taskID),
		"file_name":    "bilibili_history.xlsx",
	})
	resultMsg = string(resultBytes)

	asyncExportFiles.Store(taskID, filePath)
}

func downloadExportFile(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 task_id 参数"))
		return
	}

	filePathVal, ok := asyncExportFiles.Load(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, models.ErrorResponse("文件不存在或已过期"))
		return
	}

	filePath, _ := filePathVal.(string)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		asyncExportFiles.Delete(taskID)
		c.JSON(http.StatusNotFound, models.ErrorResponse("文件不存在或已过期"))
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=bilibili_history.xlsx")
	c.File(filePath)
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
	// 读取原始 body 用于调试
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	utils.LogInfo("收到 /log/send 请求: body=%s, content-type=%s", string(bodyBytes), c.GetHeader("Content-Type"))

	stats := make(map[string]interface{})
	if err := c.ShouldBindJSON(&stats); err != nil {
		utils.LogWarning("ShouldBindJSON 失败: %v, 使用空 stats", err)
		stats = make(map[string]interface{})
	}
	utils.LogInfo("解析后的 stats: %v", stats)

	err := services.SendDailyReport(stats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("发送每日报告失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "每日报告已发送",
	})
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
	taskID := c.Query("task_id")

	result, err := services.FetchHistory(taskID, skipExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("启动历史记录获取失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}

func getFetchStatus(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID != "" {
		status := services.GetFetchTaskStatus(taskID)
		if status == nil {
			c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
				"task_id":   taskID,
				"status":    "not_found",
				"message":   "任务不存在或已完成",
			}))
			return
		}
		c.JSON(http.StatusOK, models.SuccessResponse(status))
		return
	}

	// Return overall status
	status := services.GetFetchStatusOverall()
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
