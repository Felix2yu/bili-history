package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/models"
	"bilibili-history-go/scheduler"
	"bilibili-history-go/services"

	"github.com/gin-gonic/gin"
)

func RegisterConfigRoutes(r *gin.RouterGroup) {
	configGroup := r.Group("/config")
	{
		configGroup.GET("/email", getEmailConfig)
		configGroup.POST("/email", saveEmailConfig)
		configGroup.GET("/apprise", getAppriseConfig)
		configGroup.POST("/apprise", saveAppriseConfig)
		configGroup.GET("/server", getServerConfig)
		configGroup.POST("/server", saveServerConfig)
	}
}

func RegisterSchedulerRoutes(r *gin.RouterGroup) {
	scheduler := r.Group("/scheduler")
	{
		scheduler.GET("/tasks", getSchedulerTasks)
		scheduler.POST("/tasks", addSchedulerTask)
		scheduler.PUT("/tasks/:id", updateSchedulerTask)
		scheduler.DELETE("/tasks/:id", deleteSchedulerTask)
		scheduler.POST("/tasks/:id/run", runSchedulerTask)
		scheduler.GET("/tasks/:id/history", getTaskHistory)
		scheduler.GET("/status", getSchedulerStatus)
	}
}

func RegisterDataSyncRoutes(r *gin.RouterGroup) {
	dataSync := r.Group("/data_sync")
	{
		dataSync.GET("/status", getDataSyncStatus)
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
		log.POST("/send", sendLogEmail)
		log.GET("/list", getLogList)
	}
}

func RegisterFetchRoutes(r *gin.RouterGroup) {
	fetch := r.Group("/fetch")
	{
		fetch.POST("/start", fetchBiliHistory)
		fetch.GET("/status", getFetchStatus)
	}
}

func RegisterDeleteRoutes(r *gin.RouterGroup) {
	delete := r.Group("/delete")
	{
		delete.POST("/history", deleteHistoryRecords)
	}

	biliHistory := r.Group("/bilibili/history")
	{
		biliHistory.POST("/delete", deleteBiliHistory)
		biliHistory.GET("/status", getDeleteBiliStatus)
	}
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

func RegisterVideoDetailsRoutes(r *gin.RouterGroup) {
	videoDetails := r.Group("/video_details")
	{
		videoDetails.GET("/:bvid", getVideoDetails)
		videoDetails.POST("/sync", syncVideoDetails)
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

func getEmailConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("配置加载失败"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(cfg.Email))
}

func saveEmailConfig(c *gin.Context) {
	var emailCfg config.EmailConfig
	if err := c.ShouldBindJSON(&emailCfg); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	cfg, _ := config.LoadConfig()
	cfg.Email = emailCfg
	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("保存失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "邮件配置已保存",
	})
}

func getAppriseConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("配置加载失败"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(cfg.Apprise))
}

func saveAppriseConfig(c *gin.Context) {
	var appriseCfg config.AppriseConfig
	if err := c.ShouldBindJSON(&appriseCfg); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	cfg, _ := config.LoadConfig()
	cfg.Apprise = appriseCfg
	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("保存失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Apprise配置已保存",
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
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	}))
}

func addSchedulerTask(c *gin.Context) {
	var task scheduler.ScheduleTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	sched := scheduler.GetScheduler()
	newTask, err := sched.CreateTask(&task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("创建任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务创建成功",
		"data":    newTask,
	})
}

func updateSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	sched := scheduler.GetScheduler()
	updatedTask, err := sched.UpdateTask(taskID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("更新任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务更新成功",
		"data":    updatedTask,
	})
}

func deleteSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")

	sched := scheduler.GetScheduler()
	err := sched.DeleteTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("删除任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务删除成功",
	})
}

func runSchedulerTask(c *gin.Context) {
	taskID := c.Param("id")

	sched := scheduler.GetScheduler()
	err := sched.RunTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("运行任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务已启动",
	})
}

func getTaskHistory(c *gin.Context) {
	taskID := c.Param("id")
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	sched := scheduler.GetScheduler()
	executions := sched.GetTaskExecutions(taskID, limit)

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"records": executions,
		"total":   len(executions),
	}))
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

func sendLogEmail(c *gin.Context) {
	err := services.SendTestEmail()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("发送邮件失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "测试邮件已发送",
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

func getVideoDetails(c *gin.Context) {
	bvid := c.Param("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少bvid参数"))
		return
	}

	client := biliapi.NewClient("")
	videoInfo, err := client.GetVideoInfo(bvid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("获取视频详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(videoInfo))
}

func syncVideoDetails(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "视频详情同步功能待实现",
	})
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
