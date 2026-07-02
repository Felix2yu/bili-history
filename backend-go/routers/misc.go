package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/models"
	"bilibili-history-go/scheduler"
	"bilibili-history-go/services"
	"bilibili-history-go/utils"

	"github.com/gin-gonic/gin"
)

func RegisterConfigRoutes(r *gin.RouterGroup) {
	configGroup := r.Group("/config")
	{
		configGroup.GET("/shoutrrr", getShoutrrrConfig)
		configGroup.POST("/shoutrrr", saveShoutrrrConfig)
		configGroup.POST("/shoutrrr/test", testShoutrrrConfig)
		configGroup.GET("/server", getServerConfig)
		configGroup.POST("/server", saveServerConfig)
		// Python-compatible aliases
		configGroup.GET("/apprise-config", getShoutrrrConfig)
		configGroup.POST("/apprise-config", saveShoutrrrConfig)
	}
}

func RegisterSchedulerRoutes(r *gin.RouterGroup) {
	scheduler := r.Group("/scheduler")
	{
		scheduler.GET("/tasks", getSchedulerTasks)
		scheduler.POST("/tasks", addSchedulerTask)
		// task_id may contain "/" (Python uses endpoint paths like
		// "/fetch/bili-history" as IDs). The frontend URL-encodes the slash
		// as %2F and the Gin engine uses UseRawPath so ":id" matches the
		// encoded segment correctly.
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
		dataSync.POST("/check", checkDataIntegrity)
		dataSync.POST("/sync", syncData)
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
		// Python-compatible alias
		clean.POST("/clean_data", cleanData)
	}
}

func RegisterLogRoutes(r *gin.RouterGroup) {
	log := r.Group("/log")
	{
		log.POST("/send", sendDailyReport)
		log.GET("/list", getLogList)
		// Python-compatible alias
		log.POST("/send-email", sendDailyReport)
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
	syncDeleted := c.DefaultQuery("sync_deleted", "false") == "true"

	result, err := services.FetchHistory(true, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	_ = syncDeleted
	c.JSON(http.StatusOK, result)
}

func fetchBiliHistoryFull(c *gin.Context) {
	result, err := services.FetchHistory(false, false)
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
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"videos": []interface{}{},
		"total":  0,
	}))
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
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "批量删除功能待实现",
		"data":    map[string]interface{}{"deleted_count": 0},
	})
}

func deleteSingleBiliHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "删除B站历史记录功能待实现",
	})
}

func deleteBatchBiliHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "批量删除B站历史记录功能待实现",
	})
}

func RegisterPopularRoutes(r *gin.RouterGroup) {
	popular := r.Group("/bilibili")
	{
		popular.GET("/popular", getPopularVideos)
	}

	popularAnalytics := r.Group("/popular")
	{
		popularAnalytics.GET("/stats", getPopularStats)
	}
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

func getSchedulerTasks(c *gin.Context) {
	sched := scheduler.GetScheduler()
	tasks := sched.GetTasks()
	utils.LogInfo("获取任务列表: 共 %d 个主任务", len(tasks))
	// Frontend reads response.data.tasks (top-level tasks array, not nested
	// under data). Also include the Python-style status/message envelope.
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
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"status":  "idle",
		"message": "数据同步状态功能待实现",
	}))
}

func getDataSyncConfig(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"enabled":         false,
		"check_interval":  3600,
		"auto_fix":        false,
		"check_on_start":  false,
	}))
}

func updateDataSyncConfig(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message": "数据同步配置更新功能待实现",
	}))
}

func checkDataIntegrity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "数据完整性检查功能待实现",
	})
}

func syncData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "数据同步功能待实现",
	})
}

func exportToExcel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Excel导出功能待实现",
	})
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
	syncDeleted := false
	if syncStr := c.Query("sync_deleted"); syncStr == "true" {
		syncDeleted = true
	}

	go func() {
		sqliteImportStatus["status"] = "running"
		sqliteImportStatus["message"] = "正在导入数据..."

		result, err := services.ImportHistoryFiles(syncDeleted)
		if err != nil {
			sqliteImportStatus["status"] = "error"
			sqliteImportStatus["message"] = err.Error()
			return
		}

		sqliteImportStatus["status"] = "completed"
		sqliteImportStatus["inserted_count"] = result.InsertedCount
		sqliteImportStatus["total_files"] = result.TotalFiles
		sqliteImportStatus["message"] = "导入完成"
	}()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "开始导入SQLite数据",
	})
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

	processVideoDetails := false
	if processStr := c.Query("process_video_details"); processStr == "true" {
		processVideoDetails = true
	}

	result, err := services.FetchHistory(skipExists, processVideoDetails)
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

func getPopularVideos(c *gin.Context) {
	pn := 1
	if pnStr := c.Query("pn"); pnStr != "" {
		if p, err := strconv.Atoi(pnStr); err == nil {
			pn = p
		}
	}

	ps := 20
	if psStr := c.Query("ps"); psStr != "" {
		if p, err := strconv.Atoi(psStr); err == nil {
			ps = p
		}
	}

	client := biliapi.NewClient("")
	data, err := client.GetPopular(pn, ps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("获取热门视频失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

func getPopularStats(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"stats":   map[string]interface{}{},
		"message": "热门视频分析功能待实现",
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
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"stats":   map[string]interface{}{},
		"message": "标题统计功能待实现",
	}))
}

func getTitlePatterns(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"patterns": []interface{}{},
		"message": "标题模式发现功能待实现",
	}))
}

func getTitleSentiment(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"sentiment": map[string]interface{}{},
		"message": "标题情感分析功能待实现",
	}))
}

func getTitleLengthAnalysis(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"analysis": map[string]interface{}{},
		"message": "标题长度分析功能待实现",
	}))
}

func getTitleTrend(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"trend":   []interface{}{},
		"message": "标题趋势分析功能待实现",
	}))
}

func RegisterDownloadRoutes(r *gin.RouterGroup) {
	download := r.Group("/download")
	{
		download.GET("/check_video_download", checkVideoDownload)
		download.GET("/list_downloaded_videos", listDownloadedVideos)
		download.DELETE("/delete_downloaded_video", deleteDownloadedVideo)
	}
}

func checkVideoDownload(c *gin.Context) {
	cidsStr := c.Query("cids")
	results := make(map[string]interface{})

	if cidsStr != "" {
		cids := splitAndParseInts(cidsStr)
		for _, cid := range cids {
			results[cid] = map[string]interface{}{
				"downloaded": false,
				"file_path":  "",
				"file_size":  0,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "查询成功",
		"results": results,
	})
}

func listDownloadedVideos(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"videos": []interface{}{},
		"total":  0,
		"page":   1,
		"limit":  20,
	}))
}

func deleteDownloadedVideo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "删除功能待实现",
	})
}

func splitAndParseInts(s string) []string {
	result := []string{}
	if s == "" {
		return result
	}
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
