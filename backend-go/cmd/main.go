package main

import (
	"fmt"
	"net/http"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/routers"
	"bilibili-history-go/scheduler"
	"bilibili-history-go/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	utils.LogSuccess("=== 正在启动应用... ===")

	cfg, err := config.LoadConfig()
	if err != nil {
		utils.LogWarning("配置加载警告: %v", err)
	}
	utils.LogInfo("配置文件路径: %s", config.GetConfigPathValue())
	if cfg.SESSDATA != "" {
		utils.LogInfo("SESSDATA 已配置 (长度: %d)", len(cfg.SESSDATA))
	} else {
		utils.LogWarning("SESSDATA 未配置，请通过环境变量或配置文件设置")
	}

	db := database.GetSQLiteDB()
	if db == nil {
		utils.LogError("数据库初始化失败")
		return
	}
	utils.LogSuccess("数据库初始化完成")

	database.InitCategories()
	utils.LogSuccess("分类表初始化完成")

	sched := scheduler.GetScheduler()
	sched.Start()
	utils.LogSuccess("调度器已启动")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	// task_id may contain "/" (Python uses endpoint paths like
	// "/fetch/bili-history" as IDs). The frontend URL-encodes these slashes
	// as %2F. With UseRawPath=true Gin matches routes against the raw
	// (still-encoded) path, so %2F is treated as a single path segment and
	// ":id" route params work correctly. UnescapePathValues=true (default)
	// then gives handlers the decoded value (e.g. "/fetch/bili-history").
	r.UseRawPath = true
	r.UnescapePathValues = true

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("")
	{
		routers.RegisterHistoryRoutes(api)
		routers.RegisterCategoryRoutes(api)
		routers.RegisterLoginRoutes(api)
		routers.RegisterAnalysisRoutes(api)
		routers.RegisterViewingRoutes(api)
		routers.RegisterFavoriteRoutes(api)
		routers.RegisterConfigRoutes(api)
		routers.RegisterSchedulerRoutes(api)
		routers.RegisterDataSyncRoutes(api)
		routers.RegisterExportRoutes(api)
		routers.RegisterImportRoutes(api)
		routers.RegisterCleanRoutes(api)
		routers.RegisterLogRoutes(api)
		routers.RegisterFetchRoutes(api)
		routers.RegisterDeleteRoutes(api)
		routers.RegisterPopularRoutes(api)
		routers.RegisterVideoDetailsRoutes(api)
		routers.RegisterInteractionRoutes(api)
		routers.RegisterTitleAnalyticsRoutes(api)
		routers.RegisterImageRoutes(api)
		routers.RegisterDownloadRoutes(api)
	}

	r.GET("/health", func(c *gin.Context) {
		schedStatus := sched.GetStatus()
		c.JSON(200, gin.H{
			"status":           "running",
			"timestamp":        time.Now().Format(time.RFC3339),
			"scheduler_status": schedStatus["running"],
		})
	})

	r.GET("/routes", func(c *gin.Context) {
		routes := r.Routes()
		var routeList []map[string]interface{}
		for _, route := range routes {
			routeList = append(routeList, map[string]interface{}{
				"method": route.Method,
				"path":   route.Path,
			})
		}
		c.JSON(200, gin.H{
			"total":  len(routeList),
			"routes": routeList,
		})
	})

	r.GET("/scheduler/available-endpoints", func(c *gin.Context) {
		routes := r.Routes()
		endpoints := make([]map[string]interface{}, 0)

		skipPaths := map[string]bool{
			"/health": true,
			"/routes": true,
		}

		for _, route := range routes {
			if skipPaths[route.Path] {
				continue
			}
			if route.Method == "HEAD" || route.Method == "OPTIONS" {
				continue
			}

			meta := routers.GetEndpointMeta(route.Method, route.Path)
			tags := meta.Tags
			if tags == nil {
				tags = []string{}
			}

			endpoints = append(endpoints, map[string]interface{}{
				"path":        route.Path,
				"method":      route.Method,
				"summary":     meta.Summary,
				"tags":        tags,
				"operationId": meta.OperationID,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"message":   fmt.Sprintf("获取API端点列表成功，共 %d 个端点", len(endpoints)),
			"total":     len(endpoints),
			"endpoints": endpoints,
		})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	utils.LogSuccess("服务启动成功，监听地址: %s", addr)
	utils.LogSuccess("=== 应用启动完成 ===")

	if cfg.Server.SSLEnabled && cfg.Server.SSLCertFile != "" && cfg.Server.SSLKeyFile != "" {
		utils.LogInfo("使用HTTPS启动服务")
		r.RunTLS(addr, cfg.Server.SSLCertFile, cfg.Server.SSLKeyFile)
	} else {
		r.Run(addr)
	}
}
