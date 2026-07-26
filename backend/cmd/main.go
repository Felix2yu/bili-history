package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/routers"
	"bilibili-history-go/scheduler"
	"bilibili-history-go/services"
	"bilibili-history-go/utils"
	"bilibili-history-go/web"

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

	// Ensure all year tables have the status column
	database.MigrateStatusColumn()

	database.InitCategories()
	utils.LogSuccess("分类表初始化完成")

	sched := scheduler.GetScheduler()
	sched.Start()
	utils.LogSuccess("调度器已启动")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	// task_id may contain "/" (e.g., "/fetch/bili-history").
	// The frontend URL-encodes slashes as %2F. With UseRawPath=true Gin
	// matches routes against the raw path, so %2F is treated as a single
	// segment and ":id" params work correctly.
	r.UseRawPath = true
	r.UnescapePathValues = true

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
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
		routers.RegisterVideoDetailsRoutes(api)
		routers.RegisterInteractionRoutes(api)
		routers.RegisterTitleAnalyticsRoutes(api)
		routers.RegisterReportRoutes(api)
		routers.RegisterImageRoutes(api)
		routers.RegisterDownloadRoutes(api)
	}

	distFS := web.GetDistFS()
	if distFS != nil {
		registerFrontendRoutes(r, distFS)
		utils.LogSuccess("前端静态资源已嵌入")
	} else {
		utils.LogWarning("前端静态资源未嵌入（开发模式）")
	}

	// MCP 服务
	if cfg.Mcp.Enabled {
		services.SetupMCPServer(cfg)
		mcpHandler := services.GetMCPHandler()
		mcpHandler = services.WrapWithAuth(cfg, mcpHandler)
		mcpPath := cfg.Mcp.Path
		if mcpPath == "" {
			mcpPath = "/mcp"
		}
		// Strip the path prefix before passing to the MCP handler
		mcpHandler = http.StripPrefix(mcpPath, mcpHandler)
		r.Any(mcpPath+"/*path", gin.WrapH(mcpHandler))
		r.Any(mcpPath, gin.WrapH(mcpHandler))
		utils.LogSuccess("MCP 服务已启用，路径: %s", mcpPath)
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

	if cfg.Server.DataIntegrity.CheckOnStartup {
		go func() {
			utils.LogInfo("启动时数据完整性检查...")
			result, err := services.RunIntegrityCheck(false)
			if err != nil {
				utils.LogWarning("启动时完整性检查失败: %v", err)
			} else {
				utils.LogInfo("完整性检查完成: JSON=%d条, DB=%d条, 差异=%d",
					result.TotalJSONRecords, result.TotalDBRecords, result.Difference)
			}
		}()
	}

	if cfg.Server.SSLEnabled && cfg.Server.SSLCertFile != "" && cfg.Server.SSLKeyFile != "" {
		utils.LogInfo("使用HTTPS启动服务")
		r.RunTLS(addr, cfg.Server.SSLCertFile, cfg.Server.SSLKeyFile)
	} else {
		r.Run(addr)
	}
}

func registerFrontendRoutes(r *gin.Engine, distFS fs.FS) {
	fileServer := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api/") || path == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		if strings.HasPrefix(path, "/mcp") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		if strings.HasPrefix(path, "/_nuxt/") ||
			strings.HasPrefix(path, "/icons/") ||
			strings.Contains(path, ".") {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
