package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/db"
	"bili-history/internal/handlers"
	"bili-history/internal/middleware"
	"bili-history/internal/services/mcp"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure output directories exist
	os.MkdirAll(config.GetOutputPath(), 0755)
	os.MkdirAll(config.GetOutputPath("database"), 0755)

	// Ensure all database schemas match Python backend
	if err := db.EnsureAllSchemas(); err != nil {
		log.Printf("Warning: schema init error: %v", err)
	}

	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Register routes
	registerRoutes(r, cfg)

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// Mount MCP server (read-only)
	if token := os.Getenv("BHF_MCP_TOKEN"); token != "" {
		mcpServer, err := mcp.NewMCPServer()
		if err == nil {
			r.Any("/mcp/*path", gin.WrapH(mcpServer.Handler()))
			r.Any("/mcp", gin.WrapH(mcpServer.Handler()))
			log.Println("MCP server mounted at /mcp")
		}
	}

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if cfg.Server.SSLEnabled && cfg.Server.SSLCertFile != "" && cfg.Server.SSLKeyFile != "" {
			log.Printf("Using HTTPS with cert=%s key=%s", cfg.Server.SSLCertFile, cfg.Server.SSLKeyFile)
			if err := srv.ListenAndServeTLS(cfg.Server.SSLCertFile, cfg.Server.SSLKeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server error: %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP server error: %v", err)
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connections
	db.CloseAll()

	log.Println("Server exited")
}

func registerRoutes(r *gin.Engine, cfg *config.Config) {
	// Login routes
	login := r.Group("/login")
	{
		login.GET("/qrcode/generate", handlers.GenerateQRCode)
		login.GET("/qrcode/poll", handlers.PollLogin)
		login.GET("/check-and-notify", handlers.CheckAndUpdateSESSDATA)
		login.GET("/check", handlers.CheckLogin)
		login.POST("/logout", handlers.Logout)
	}

	// History routes
	history := r.Group("/history")
	{
		history.GET("", handlers.GetHistory)
		history.GET("/list", handlers.GetHistory)
		history.GET("/available-years", handlers.GetAvailableYears)
		history.GET("/all", handlers.GetHistory)
		history.GET("/search", handlers.GetHistory)
		history.GET("/sqlite-version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"version": "3.x (modernc.org/sqlite)"})
		})
	}

	historySimple := r.Group("/history_simple")
	{
		historySimple.GET("", handlers.GetHistory)
	}

	// Fetch routes
	fetch := r.Group("/fetch")
	{
		fetch.GET("/bili-history", handlers.FetchBiliHistory)
		fetch.GET("/bili-history-simple", handlers.FetchBiliHistorySimple)
	}

	// Analysis routes
	analysis := r.Group("/analysis")
	{
		analysis.GET("", handlers.GetViewingAnalytics)
		analysis.POST("/analyze", handlers.GetViewingAnalytics)
	}

	// Viewing analytics
	viewing := r.Group("/viewing")
	{
		viewing.GET("", handlers.GetViewingAnalytics)
		viewing.GET("/monthly-stats", handlers.GetMonthlyStats)
		viewing.GET("/weekly-stats", handlers.GetWeeklyStats)
		viewing.GET("/time-slots", handlers.GetTimeSlots)
		viewing.GET("/continuity", handlers.GetContinuity)
		viewing.GET("/completion-rates", handlers.GetCompletionRates)
		viewing.GET("/author-completion", handlers.GetAuthorCompletion)
		viewing.GET("/tag-analysis", handlers.GetTagAnalysis)
		viewing.GET("/duration-analysis", handlers.GetDurationAnalysis)
		viewing.GET("/watch-counts", handlers.GetWatchCounts)
		viewing.GET("/annual-summary/json", handlers.GetAnnualSummary)
	}

	// Title analytics
	title := r.Group("/title")
	{
		title.GET("", handlers.GetTitleAnalytics)
		title.GET("/keyword-analysis", handlers.GetTitleAnalytics)
		title.GET("/length-analysis", handlers.GetTitleAnalytics)
		title.GET("/trend-analysis", handlers.GetTitleAnalytics)
		title.GET("/interaction-analysis", handlers.GetTitleAnalytics)
	}

	// Daily count
	daily := r.Group("/daily")
	{
		daily.GET("/count", handlers.GetDailyCount)
	}

	// Video details
	videoDetails := r.Group("/video_details")
	{
		videoDetails.GET("", handlers.GetVideoDetails)
		videoDetails.GET("/info/:bvid", handlers.GetVideoInfoFromDB)
		videoDetails.GET("/search", handlers.SearchVideoDetails)
		videoDetails.POST("/batch_fetch", handlers.BatchFetchVideoDetails)
		videoDetails.GET("/batch_fetch_from_history", handlers.FetchVideoDetailsFromHistory)
		videoDetails.GET("/stats", handlers.GetVideoStats)
		videoDetails.GET("/uploaders", handlers.GetUploaderList)
		videoDetails.GET("/tags", handlers.GetVideoTags)
		videoDetails.GET("/uploader/:mid", handlers.GetUploaderInfo)
		videoDetails.GET("/invalid-videos", handlers.GetInvalidVideos)
	}

	// Categories
	categories := r.Group("/categories")
	{
		categories.GET("", handlers.GetCategories)
		categories.GET("/categories", handlers.GetCategories)
		categories.GET("/main-categories", handlers.GetMainCategories)
		categories.GET("/sub-categories/:main_category", handlers.GetSubCategories)
	}

	// Heatmap
	heatmap := r.Group("/heatmap")
	{
		heatmap.GET("/data", handlers.GetHeatmapData)
		heatmap.GET("/generate_heatmap", handlers.GetHeatmapData)
	}

	// Export
	// Delete history
	deleteGroup := r.Group("/delete")
	{
		deleteGroup.DELETE("/:bvid", handlers.DeleteLocalHistory)
		deleteGroup.POST("/batch-delete", handlers.BatchDeleteHistory)
		deleteGroup.DELETE("/batch-delete", handlers.BatchDeleteHistory)
	}

	// Import routes
	importSqlite := r.Group("/importSqlite")
	{
		importSqlite.POST("/import_data_sqlite", handlers.ImportFromSQLite)
	}

	importMysql := r.Group("/importMysql")
	{
		importMysql.POST("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "MySQL import feature coming soon"})
		})
	}

	// Scheduler routes
	schedulerGroup := r.Group("/scheduler")
	{
		schedulerGroup.GET("/tasks", handlers.GetSchedulerTasks)
		schedulerGroup.POST("/tasks", handlers.CreateSchedulerTask)
		schedulerGroup.PUT("/tasks/:id", handlers.UpdateSchedulerTask)
		schedulerGroup.DELETE("/tasks/:id", handlers.DeleteSchedulerTask)
		schedulerGroup.POST("/tasks/:id/run", handlers.RunSchedulerTask)
		schedulerGroup.POST("/tasks/:id/execute", handlers.RunSchedulerTask)
		schedulerGroup.GET("/executions", handlers.GetSchedulerExecutions)
		schedulerGroup.GET("/tasks/history", handlers.GetTaskHistory)
		schedulerGroup.GET("/available-endpoints", handlers.GetSchedulerEndpoints)
	}

	// Config routes
	configGroup := r.Group("/config")
	{
		configGroup.GET("/email", handlers.GetEmailConfig)
		configGroup.POST("/email-config", handlers.UpdateEmailConfig)
		configGroup.POST("/test-email", handlers.TestEmail)
		configGroup.GET("/mcp-config", handlers.GetMCPConfig)
		configGroup.POST("/mcp-config", handlers.UpdateMCPConfig)
		configGroup.GET("/apprise-config", handlers.GetAppriseConfig)
		configGroup.POST("/apprise-config", handlers.UpdateAppriseConfig)
		configGroup.POST("/test-apprise", handlers.TestApprise)
		configGroup.POST("/test-ntfy", handlers.TestNtfy)
	}

	// Popular videos
	bilibili := r.Group("/bilibili")
	{
		bilibili.GET("/popular/all", handlers.GetPopularVideos)
	}

	// Favorite
	favorite := r.Group("/favorite")
	{
		favorite.GET("/list", handlers.GetFavorites)
		favorite.GET("/folder/created/list-all", handlers.GetFavorites)
		favorite.GET("/folder/collected/list", handlers.GetFavorites)
		favorite.GET("/list", handlers.GetFavoritesListDB)
		favorite.GET("/content/list", handlers.GetFavoritesContentList)
		favorite.GET("/check", handlers.CheckFavorite)
		favorite.POST("/check/batch", handlers.BatchCheckFavorite)
	}

	// Comment
	comment := r.Group("/comment")
	{
		comment.GET("/list", handlers.GetComments)
	}

	// Download
	downloadGroup := r.Group("/download")
	{
		downloadGroup.POST("/video", handlers.StartDownload)
		downloadGroup.POST("/audio", handlers.StartDownload)
		downloadGroup.POST("/cancel", handlers.CancelDownload)
		downloadGroup.GET("/active", handlers.GetActiveDownloads)
		downloadGroup.GET("/info", handlers.GetVideoInfo)
	}

	// Dynamic
	dynamic := r.Group("/dynamic")
	{
		dynamic.GET("/list", handlers.GetDynamics)
		dynamic.GET("/db/hosts", handlers.ListDynamicHosts)
		dynamic.GET("/db/space/:host_mid", handlers.ListDynamicSpace)
		dynamic.GET("/types", handlers.GetDynamicTypes)
		dynamic.GET("/detail/:dynamic_id", handlers.GetDynamicDetail)
		dynamic.DELETE("/space/:host_mid", handlers.DeleteDynamicSpace)
	}

	// Watch later
	watchlater := r.Group("/watchlater")
	{
		watchlater.GET("/list", handlers.GetWatchLater)
		watchlater.GET("/local", handlers.GetWatchLater)
	}

	// Like
	like := r.Group("/like")
	{
		like.GET("/list", handlers.GetLikes)
		like.GET("/local", handlers.GetLikes)
	}

	// Image downloader
	images := r.Group("/images")
	{
		images.POST("/download", handlers.DownloadImage)
		images.POST("/start", handlers.DownloadImage)
	}

	// Log send
	logGroup := r.Group("/log")
	{
		logGroup.POST("/send-email", handlers.SendLogEmail)
	}

	// Interactions
	interactions := r.Group("/interactions")
	{
		interactions.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Interactions feature coming soon"})
		})
	}

	// Collection download
	collection := r.Group("/collection")
	{
		collection.POST("/download", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Collection download feature coming soon"})
		})
	}

	// Bilibili history delete (remote)
	bilibiliHistory := r.Group("/bilibili/history")
	{
		bilibiliHistory.DELETE("", handlers.DeleteRemoteHistory)
	}

	// Popular analytics
	popular := r.Group("/popular")
	{
		popular.GET("/analytics", handlers.GetPopularVideos)
	}

	// Data sync
	dataSync := r.Group("/data_sync")
	{
		dataSync.POST("/sync", handlers.SyncDataFull)
		dataSync.POST("/check", handlers.CheckDataIntegrityFull)
		dataSync.GET("/check", handlers.CheckDataIntegrityFull)
		dataSync.GET("/report", handlers.GetIntegrityReport)
		dataSync.GET("/sync/result", handlers.GetSyncResult)
		dataSync.GET("/config", handlers.GetIntegrityCheckConfig)
		dataSync.POST("/config", handlers.UpdateIntegrityCheckConfig)
		dataSync.GET("/status", handlers.SyncStatus)
		dataSync.POST("/incremental", handlers.IncrementalSync)
		dataSync.POST("/force", handlers.ForceSync)
	}

	// Export
	exportGroup := r.Group("/export")
	{
		exportGroup.GET("/excel", handlers.ExportHistory)
	}

	// Clean
	cleanGroup := r.Group("/clean")
	{
		cleanGroup.POST("", handlers.CleanHistoryData)
	}
}
