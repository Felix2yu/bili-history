package routers

// EndpointMeta 描述一个 API 端点的元信息。
// 用于 /scheduler/available-endpoints 接口，向调用方（如前端的任务创建页面）
// 展示每个端点的用途说明。
type EndpointMeta struct {
	Summary     string   // 简短描述
	Tags        []string // 分类标签
	OperationID string   // 操作标识符
}

// endpointRegistry 是 "METHOD path" -> EndpointMeta 的注册表。
// key 的格式为 "METHOD /path"，与 gin.RouteInfo 的 Method + Path 对应。
var endpointRegistry = map[string]EndpointMeta{}

// RegisterEndpointMeta 注册一个端点的元信息。
// 在各路由文件的 init() 中调用。
func RegisterEndpointMeta(method, path string, meta EndpointMeta) {
	key := method + " " + path
	endpointRegistry[key] = meta
}

// GetEndpointMeta 查找指定路由的元信息。
// 若未注册则返回零值（空 summary / 空 tags）。
func GetEndpointMeta(method, path string) EndpointMeta {
	key := method + " " + path
	if meta, ok := endpointRegistry[key]; ok {
		return meta
	}
	return EndpointMeta{
		Summary:     "",
		Tags:        []string{},
		OperationID: "",
	}
}

func init() {
	tagHistory := []string{"历史记录"}
	tagCategory := []string{"分类管理"}
	tagLogin := []string{"登录认证"}
	tagAnalysis := []string{"数据分析"}
	tagViewing := []string{"观看分析"}
	tagFavorite := []string{"收藏管理"}
	tagConfig := []string{"系统配置"}
	tagScheduler := []string{"计划任务"}
	tagDataSync := []string{"数据同步"}
	tagExport := []string{"数据导出"}
	tagImport := []string{"数据导入"}
	tagClean := []string{"数据清理"}
	tagLog := []string{"日志管理"}
	tagFetch := []string{"数据抓取"}
	tagDelete := []string{"删除操作"}
	tagPopular := []string{"热门视频"}
	tagVideo := []string{"视频详情"}
	tagImage := []string{"图片管理"}
	tagDownload := []string{"下载管理"}
	tagTitle := []string{"标题分析"}
	tagInteraction := []string{"互动记录"}

	// ========== 历史记录 ==========
	RegisterEndpointMeta("GET", "/history/available-years", EndpointMeta{
		Summary:     "获取可用年份列表",
		Tags:        tagHistory,
		OperationID: "get_available_years",
	})
	RegisterEndpointMeta("GET", "/history/all", EndpointMeta{
		Summary:     "分页获取历史记录列表",
		Tags:        tagHistory,
		OperationID: "get_history_all",
	})
	RegisterEndpointMeta("GET", "/history/search", EndpointMeta{
		Summary:     "搜索历史记录",
		Tags:        tagHistory,
		OperationID: "search_history",
	})
	RegisterEndpointMeta("GET", "/history/remarks", EndpointMeta{
		Summary:     "获取所有备注信息",
		Tags:        tagHistory,
		OperationID: "get_all_remarks",
	})
	RegisterEndpointMeta("POST", "/history/update-remark", EndpointMeta{
		Summary:     "更新视频备注",
		Tags:        tagHistory,
		OperationID: "update_remark",
	})
	RegisterEndpointMeta("POST", "/history/reset-database", EndpointMeta{
		Summary:     "重置历史数据库",
		Tags:        tagHistory,
		OperationID: "reset_database",
	})
	RegisterEndpointMeta("GET", "/history/sqlite-version", EndpointMeta{
		Summary:     "获取SQLite版本信息",
		Tags:        tagHistory,
		OperationID: "get_sqlite_version",
	})
	RegisterEndpointMeta("POST", "/history/batch-remarks", EndpointMeta{
		Summary:     "批量获取视频备注",
		Tags:        tagHistory,
		OperationID: "batch_get_remarks",
	})
	RegisterEndpointMeta("GET", "/history/by_cid/:cid", EndpointMeta{
		Summary:     "按CID查找视频",
		Tags:        tagHistory,
		OperationID: "get_video_by_cid",
	})
	RegisterEndpointMeta("GET", "/daily/daily-count", EndpointMeta{
		Summary:     "获取每日观看统计",
		Tags:        tagHistory,
		OperationID: "get_daily_count",
	})

	// ========== 分类管理 ==========
	RegisterEndpointMeta("POST", "/categories/init", EndpointMeta{
		Summary:     "初始化分类表",
		Tags:        tagCategory,
		OperationID: "init_categories",
	})
	RegisterEndpointMeta("GET", "/categories/categories", EndpointMeta{
		Summary:     "获取全部分类（含主分类和子分类）",
		Tags:        tagCategory,
		OperationID: "get_categories",
	})
	RegisterEndpointMeta("GET", "/categories/main-categories", EndpointMeta{
		Summary:     "获取主分类列表",
		Tags:        tagCategory,
		OperationID: "get_main_categories",
	})
	RegisterEndpointMeta("GET", "/categories/sub-categories/:main_category", EndpointMeta{
		Summary:     "获取指定主分类下的子分类",
		Tags:        tagCategory,
		OperationID: "get_sub_categories",
	})

	// ========== 登录认证 ==========
	RegisterEndpointMeta("GET", "/login/qrcode/generate", EndpointMeta{
		Summary:     "生成登录二维码",
		Tags:        tagLogin,
		OperationID: "generate_qrcode",
	})
	RegisterEndpointMeta("GET", "/login/qrcode/image", EndpointMeta{
		Summary:     "获取二维码图片",
		Tags:        tagLogin,
		OperationID: "get_qrcode_image",
	})
	RegisterEndpointMeta("GET", "/login/qrcode/poll", EndpointMeta{
		Summary:     "轮询二维码扫描状态",
		Tags:        tagLogin,
		OperationID: "poll_qrcode",
	})
	RegisterEndpointMeta("POST", "/login/logout", EndpointMeta{
		Summary:     "退出登录",
		Tags:        tagLogin,
		OperationID: "logout",
	})
	RegisterEndpointMeta("GET", "/login/check", EndpointMeta{
		Summary:     "检查登录状态",
		Tags:        tagLogin,
		OperationID: "check_login",
	})
	RegisterEndpointMeta("GET", "/login/check-and-notify", EndpointMeta{
		Summary:     "检查登录状态并发送通知",
		Tags:        tagLogin,
		OperationID: "check_and_notify",
	})

	// ========== 数据分析 ==========
	RegisterEndpointMeta("POST", "/analysis/analyze", EndpointMeta{
		Summary:     "执行年度历史数据分析",
		Tags:        tagAnalysis,
		OperationID: "analyze_history",
	})
	RegisterEndpointMeta("GET", "/daily/stats", EndpointMeta{
		Summary:     "获取每日统计数据",
		Tags:        tagAnalysis,
		OperationID: "get_daily_stats",
	})
	RegisterEndpointMeta("POST", "/heatmap/generate_heatmap", EndpointMeta{
		Summary:     "生成观看热力图",
		Tags:        tagAnalysis,
		OperationID: "generate_heatmap",
	})
	RegisterEndpointMeta("GET", "/heatmap/data", EndpointMeta{
		Summary:     "获取热力图数据",
		Tags:        tagAnalysis,
		OperationID: "get_heatmap_data",
	})
	RegisterEndpointMeta("GET", "/viewing/stats", EndpointMeta{
		Summary:     "获取观看统计总览（旧版）",
		Tags:        tagAnalysis,
		OperationID: "get_viewing_stats",
	})

	// ========== 观看分析 ==========
	RegisterEndpointMeta("GET", "/viewing/monthly-stats", EndpointMeta{
		Summary:     "月度观看统计",
		Tags:        tagViewing,
		OperationID: "get_monthly_stats",
	})
	RegisterEndpointMeta("GET", "/viewing/weekly-stats", EndpointMeta{
		Summary:     "周度观看统计",
		Tags:        tagViewing,
		OperationID: "get_weekly_stats",
	})
	RegisterEndpointMeta("GET", "/viewing/time-slots", EndpointMeta{
		Summary:     "观看时段分析",
		Tags:        tagViewing,
		OperationID: "get_time_slots",
	})
	RegisterEndpointMeta("GET", "/viewing/continuity", EndpointMeta{
		Summary:     "观看连续性分析",
		Tags:        tagViewing,
		OperationID: "get_continuity",
	})
	RegisterEndpointMeta("GET", "/viewing/", EndpointMeta{
		Summary:     "观看行为总览",
		Tags:        tagViewing,
		OperationID: "get_viewing_overview",
	})
	RegisterEndpointMeta("GET", "/viewing/watch-counts", EndpointMeta{
		Summary:     "重复观看分析",
		Tags:        tagViewing,
		OperationID: "get_watch_counts",
	})
	RegisterEndpointMeta("GET", "/viewing/completion-rates", EndpointMeta{
		Summary:     "视频完成率分析",
		Tags:        tagViewing,
		OperationID: "get_completion_rates",
	})
	RegisterEndpointMeta("GET", "/viewing/author-completion", EndpointMeta{
		Summary:     "UP主完成率分析",
		Tags:        tagViewing,
		OperationID: "get_author_completion",
	})
	RegisterEndpointMeta("GET", "/viewing/tag-analysis", EndpointMeta{
		Summary:     "标签观看分析",
		Tags:        tagViewing,
		OperationID: "get_tag_analysis",
	})
	RegisterEndpointMeta("GET", "/viewing/duration-analysis", EndpointMeta{
		Summary:     "视频时长分布分析",
		Tags:        tagViewing,
		OperationID: "get_duration_analysis",
	})

	// ========== 收藏管理 ==========
	RegisterEndpointMeta("GET", "/favorite/list", EndpointMeta{
		Summary:     "获取收藏夹列表",
		Tags:        tagFavorite,
		OperationID: "get_favorite_list",
	})
	RegisterEndpointMeta("GET", "/favorite/folder/created/list-all", EndpointMeta{
		Summary:     "获取创建的收藏夹列表",
		Tags:        tagFavorite,
		OperationID: "get_favorite_folder_list",
	})
	RegisterEndpointMeta("GET", "/favorite/folder/collected/list", EndpointMeta{
		Summary:     "获取收藏的收藏夹列表",
		Tags:        tagFavorite,
		OperationID: "get_collected_favorite_folders",
	})
	RegisterEndpointMeta("GET", "/favorite/folder/resource/list", EndpointMeta{
		Summary:     "获取收藏夹内容列表",
		Tags:        tagFavorite,
		OperationID: "get_favorite_folder_contents",
	})
	RegisterEndpointMeta("GET", "/favorite/content/list", EndpointMeta{
		Summary:     "获取本地收藏内容",
		Tags:        tagFavorite,
		OperationID: "get_local_favorite_contents",
	})
	RegisterEndpointMeta("POST", "/favorite/sync", EndpointMeta{
		Summary:     "同步收藏数据",
		Tags:        tagFavorite,
		OperationID: "sync_favorites",
	})
	RegisterEndpointMeta("POST", "/favorite/resource/deal", EndpointMeta{
		Summary:     "收藏/取消收藏单个视频",
		Tags:        tagFavorite,
		OperationID: "favorite_resource",
	})
	RegisterEndpointMeta("POST", "/favorite/resource/batch-deal", EndpointMeta{
		Summary:     "批量收藏/取消收藏视频",
		Tags:        tagFavorite,
		OperationID: "batch_favorite_resource",
	})
	RegisterEndpointMeta("POST", "/favorite/resource/local-batch-deal", EndpointMeta{
		Summary:     "本地批量处理收藏",
		Tags:        tagFavorite,
		OperationID: "local_batch_favorite_resource",
	})
	RegisterEndpointMeta("POST", "/favorite/check/batch", EndpointMeta{
		Summary:     "批量检查收藏状态",
		Tags:        tagFavorite,
		OperationID: "batch_check_favorite_status",
	})

	// ========== 点赞 ==========
	RegisterEndpointMeta("GET", "/like/list", EndpointMeta{
		Summary:     "获取点赞列表",
		Tags:        tagFavorite,
		OperationID: "get_like_list",
	})
	RegisterEndpointMeta("GET", "/like/local", EndpointMeta{
		Summary:     "获取本地点赞数据",
		Tags:        tagFavorite,
		OperationID: "get_like_local",
	})
	RegisterEndpointMeta("POST", "/like/sync", EndpointMeta{
		Summary:     "同步点赞数据",
		Tags:        tagFavorite,
		OperationID: "sync_likes",
	})

	// ========== 稍后再看 ==========
	RegisterEndpointMeta("GET", "/watchlater/list", EndpointMeta{
		Summary:     "获取稍后再看列表",
		Tags:        tagFavorite,
		OperationID: "get_watchlater_list",
	})
	RegisterEndpointMeta("GET", "/watchlater/local", EndpointMeta{
		Summary:     "获取本地稍后再看数据",
		Tags:        tagFavorite,
		OperationID: "get_watchlater_local",
	})
	RegisterEndpointMeta("POST", "/watchlater/sync", EndpointMeta{
		Summary:     "同步稍后再看数据",
		Tags:        tagFavorite,
		OperationID: "sync_watchlater",
	})
	RegisterEndpointMeta("DELETE", "/watchlater/:bvid", EndpointMeta{
		Summary:     "删除单个稍后再看视频",
		Tags:        tagFavorite,
		OperationID: "delete_watchlater_video",
	})
	RegisterEndpointMeta("POST", "/watchlater/batch-delete", EndpointMeta{
		Summary:     "批量删除稍后再看视频",
		Tags:        tagFavorite,
		OperationID: "batch_delete_watchlater_videos",
	})

	// ========== 动态 ==========
	RegisterEndpointMeta("GET", "/dynamic/list", EndpointMeta{
		Summary:     "获取动态列表",
		Tags:        tagFavorite,
		OperationID: "get_dynamic_list",
	})
	RegisterEndpointMeta("POST", "/dynamic/sync", EndpointMeta{
		Summary:     "同步动态数据",
		Tags:        tagFavorite,
		OperationID: "sync_dynamic",
	})

	// ========== 评论 ==========
	RegisterEndpointMeta("GET", "/comment/list", EndpointMeta{
		Summary:     "获取评论列表",
		Tags:        tagFavorite,
		OperationID: "get_comment_list",
	})
	RegisterEndpointMeta("POST", "/comment/sync", EndpointMeta{
		Summary:     "同步评论数据",
		Tags:        tagFavorite,
		OperationID: "sync_comments",
	})

	// ========== 系统配置 ==========
	RegisterEndpointMeta("GET", "/config/email", EndpointMeta{
		Summary:     "获取邮件配置",
		Tags:        tagConfig,
		OperationID: "get_email_config",
	})
	RegisterEndpointMeta("POST", "/config/email", EndpointMeta{
		Summary:     "保存邮件配置",
		Tags:        tagConfig,
		OperationID: "save_email_config",
	})
	RegisterEndpointMeta("GET", "/config/apprise", EndpointMeta{
		Summary:     "获取Apprise通知配置",
		Tags:        tagConfig,
		OperationID: "get_apprise_config",
	})
	RegisterEndpointMeta("POST", "/config/apprise", EndpointMeta{
		Summary:     "保存Apprise通知配置",
		Tags:        tagConfig,
		OperationID: "save_apprise_config",
	})
	RegisterEndpointMeta("GET", "/config/server", EndpointMeta{
		Summary:     "获取服务器配置",
		Tags:        tagConfig,
		OperationID: "get_server_config",
	})
	RegisterEndpointMeta("POST", "/config/server", EndpointMeta{
		Summary:     "保存服务器配置",
		Tags:        tagConfig,
		OperationID: "save_server_config",
	})
	// Python 兼容别名
	RegisterEndpointMeta("GET", "/config/email-config", EndpointMeta{
		Summary:     "获取邮件配置（Python兼容）",
		Tags:        tagConfig,
		OperationID: "get_email_config_alias",
	})
	RegisterEndpointMeta("POST", "/config/email-config", EndpointMeta{
		Summary:     "保存邮件配置（Python兼容）",
		Tags:        tagConfig,
		OperationID: "save_email_config_alias",
	})
	RegisterEndpointMeta("GET", "/config/apprise-config", EndpointMeta{
		Summary:     "获取Apprise配置（Python兼容）",
		Tags:        tagConfig,
		OperationID: "get_apprise_config_alias",
	})
	RegisterEndpointMeta("POST", "/config/apprise-config", EndpointMeta{
		Summary:     "保存Apprise配置（Python兼容）",
		Tags:        tagConfig,
		OperationID: "save_apprise_config_alias",
	})

	// ========== 计划任务 ==========
	RegisterEndpointMeta("GET", "/scheduler/tasks", EndpointMeta{
		Summary:     "获取计划任务列表",
		Tags:        tagScheduler,
		OperationID: "get_scheduler_tasks",
	})
	RegisterEndpointMeta("POST", "/scheduler/tasks", EndpointMeta{
		Summary:     "新建计划任务",
		Tags:        tagScheduler,
		OperationID: "add_scheduler_task",
	})
	RegisterEndpointMeta("PUT", "/scheduler/tasks/:id", EndpointMeta{
		Summary:     "更新计划任务配置",
		Tags:        tagScheduler,
		OperationID: "update_scheduler_task",
	})
	RegisterEndpointMeta("DELETE", "/scheduler/tasks/:id", EndpointMeta{
		Summary:     "删除计划任务",
		Tags:        tagScheduler,
		OperationID: "delete_scheduler_task",
	})
	RegisterEndpointMeta("POST", "/scheduler/tasks/:id/execute", EndpointMeta{
		Summary:     "立即执行计划任务",
		Tags:        tagScheduler,
		OperationID: "run_scheduler_task",
	})
	RegisterEndpointMeta("POST", "/scheduler/tasks/:id/enable", EndpointMeta{
		Summary:     "启用/禁用计划任务",
		Tags:        tagScheduler,
		OperationID: "enable_scheduler_task",
	})
	RegisterEndpointMeta("GET", "/scheduler/tasks/history", EndpointMeta{
		Summary:     "获取任务执行历史",
		Tags:        tagScheduler,
		OperationID: "get_task_history",
	})
	RegisterEndpointMeta("POST", "/scheduler/tasks/:id/subtasks", EndpointMeta{
		Summary:     "添加子任务",
		Tags:        tagScheduler,
		OperationID: "add_subtask",
	})
	RegisterEndpointMeta("DELETE", "/scheduler/tasks/:id/subtasks/:subId", EndpointMeta{
		Summary:     "删除子任务",
		Tags:        tagScheduler,
		OperationID: "delete_subtask",
	})
	RegisterEndpointMeta("GET", "/scheduler/status", EndpointMeta{
		Summary:     "获取调度器运行状态",
		Tags:        tagScheduler,
		OperationID: "get_scheduler_status",
	})

	// ========== 数据同步 ==========
	RegisterEndpointMeta("GET", "/data_sync/status", EndpointMeta{
		Summary:     "获取数据同步状态",
		Tags:        tagDataSync,
		OperationID: "get_data_sync_status",
	})
	RegisterEndpointMeta("GET", "/data_sync/config", EndpointMeta{
		Summary:     "获取数据同步配置",
		Tags:        tagDataSync,
		OperationID: "get_data_sync_config",
	})
	RegisterEndpointMeta("POST", "/data_sync/config", EndpointMeta{
		Summary:     "更新数据同步配置",
		Tags:        tagDataSync,
		OperationID: "update_data_sync_config",
	})
	RegisterEndpointMeta("POST", "/data_sync/check", EndpointMeta{
		Summary:     "检查数据完整性",
		Tags:        tagDataSync,
		OperationID: "check_data_integrity",
	})
	RegisterEndpointMeta("POST", "/data_sync/sync", EndpointMeta{
		Summary:     "执行数据同步",
		Tags:        tagDataSync,
		OperationID: "sync_data",
	})

	// ========== 数据导出 ==========
	RegisterEndpointMeta("POST", "/export/excel", EndpointMeta{
		Summary:     "导出历史记录到Excel",
		Tags:        tagExport,
		OperationID: "export_to_excel",
	})

	// ========== 数据导入 ==========
	RegisterEndpointMeta("POST", "/importMysql/start", EndpointMeta{
		Summary:     "从MySQL导入数据",
		Tags:        tagImport,
		OperationID: "import_from_mysql",
	})
	RegisterEndpointMeta("GET", "/importMysql/status", EndpointMeta{
		Summary:     "获取MySQL导入状态",
		Tags:        tagImport,
		OperationID: "get_import_mysql_status",
	})
	RegisterEndpointMeta("POST", "/importSqlite/start", EndpointMeta{
		Summary:     "从SQLite导入数据",
		Tags:        tagImport,
		OperationID: "import_from_sqlite",
	})
	RegisterEndpointMeta("GET", "/importSqlite/status", EndpointMeta{
		Summary:     "获取SQLite导入状态",
		Tags:        tagImport,
		OperationID: "get_import_sqlite_status",
	})
	RegisterEndpointMeta("POST", "/importSqlite/import_data_sqlite", EndpointMeta{
		Summary:     "导入SQLite数据（Python兼容）",
		Tags:        tagImport,
		OperationID: "import_from_sqlite_alias",
	})

	// ========== 数据清理 ==========
	RegisterEndpointMeta("POST", "/clean/start", EndpointMeta{
		Summary:     "开始清理无效数据",
		Tags:        tagClean,
		OperationID: "clean_data",
	})
	RegisterEndpointMeta("GET", "/clean/status", EndpointMeta{
		Summary:     "获取清理任务状态",
		Tags:        tagClean,
		OperationID: "get_clean_status",
	})
	RegisterEndpointMeta("POST", "/clean/clean_data", EndpointMeta{
		Summary:     "清理数据（Python兼容）",
		Tags:        tagClean,
		OperationID: "clean_data_alias",
	})

	// ========== 日志管理 ==========
	RegisterEndpointMeta("POST", "/log/send", EndpointMeta{
		Summary:     "发送日志邮件",
		Tags:        tagLog,
		OperationID: "send_log_email",
	})
	RegisterEndpointMeta("GET", "/log/list", EndpointMeta{
		Summary:     "获取日志列表",
		Tags:        tagLog,
		OperationID: "get_log_list",
	})
	RegisterEndpointMeta("POST", "/log/send-email", EndpointMeta{
		Summary:     "发送日志邮件（Python兼容）",
		Tags:        tagLog,
		OperationID: "send_log_email_alias",
	})

	// ========== 数据抓取 ==========
	RegisterEndpointMeta("POST", "/fetch/start", EndpointMeta{
		Summary:     "开始抓取B站历史记录",
		Tags:        tagFetch,
		OperationID: "fetch_bili_history",
	})
	RegisterEndpointMeta("GET", "/fetch/status", EndpointMeta{
		Summary:     "获取抓取任务状态",
		Tags:        tagFetch,
		OperationID: "get_fetch_status",
	})
	RegisterEndpointMeta("GET", "/fetch/bili-history-realtime", EndpointMeta{
		Summary:     "实时获取最新历史记录",
		Tags:        tagFetch,
		OperationID: "fetch_bili_history_realtime",
	})
	RegisterEndpointMeta("GET", "/fetch/bili-history", EndpointMeta{
		Summary:     "获取完整B站历史记录",
		Tags:        tagFetch,
		OperationID: "fetch_bili_history_full",
	})
	RegisterEndpointMeta("POST", "/fetch/bili-history", EndpointMeta{
		Summary:     "获取完整B站历史记录（POST）",
		Tags:        tagFetch,
		OperationID: "fetch_bili_history_full_post",
	})
	RegisterEndpointMeta("GET", "/fetch/invalid-videos", EndpointMeta{
		Summary:     "获取失效视频列表",
		Tags:        tagFetch,
		OperationID: "get_invalid_videos",
	})

	// ========== 删除操作 ==========
	RegisterEndpointMeta("POST", "/delete/history", EndpointMeta{
		Summary:     "删除历史记录",
		Tags:        tagDelete,
		OperationID: "delete_history_records",
	})
	RegisterEndpointMeta("DELETE", "/delete/batch-delete", EndpointMeta{
		Summary:     "批量删除历史记录",
		Tags:        tagDelete,
		OperationID: "batch_delete_history",
	})
	RegisterEndpointMeta("POST", "/bilibili/history/delete", EndpointMeta{
		Summary:     "删除B站历史记录",
		Tags:        tagDelete,
		OperationID: "delete_bili_history",
	})
	RegisterEndpointMeta("GET", "/bilibili/history/status", EndpointMeta{
		Summary:     "获取删除任务状态",
		Tags:        tagDelete,
		OperationID: "get_delete_bili_status",
	})
	RegisterEndpointMeta("DELETE", "/bilibili/history/single", EndpointMeta{
		Summary:     "删除单条B站历史记录",
		Tags:        tagDelete,
		OperationID: "delete_single_bili_history",
	})
	RegisterEndpointMeta("DELETE", "/bilibili/history/batch", EndpointMeta{
		Summary:     "批量删除B站历史记录",
		Tags:        tagDelete,
		OperationID: "delete_batch_bili_history",
	})

	// ========== 热门视频 ==========
	RegisterEndpointMeta("GET", "/bilibili/popular", EndpointMeta{
		Summary:     "获取热门视频列表",
		Tags:        tagPopular,
		OperationID: "get_popular_videos",
	})
	RegisterEndpointMeta("GET", "/popular/stats", EndpointMeta{
		Summary:     "热门视频统计分析",
		Tags:        tagPopular,
		OperationID: "get_popular_stats",
	})

	// ========== 视频详情 ==========
	RegisterEndpointMeta("GET", "/video_details/fetch/:bvid", EndpointMeta{
		Summary:     "抓取单个视频详情",
		Tags:        tagVideo,
		OperationID: "fetch_video_detail",
	})
	RegisterEndpointMeta("GET", "/video_details/info/:bvid", EndpointMeta{
		Summary:     "获取视频详情信息",
		Tags:        tagVideo,
		OperationID: "get_video_info",
	})
	RegisterEndpointMeta("GET", "/video_details/search", EndpointMeta{
		Summary:     "搜索视频",
		Tags:        tagVideo,
		OperationID: "search_videos",
	})
	RegisterEndpointMeta("POST", "/video_details/batch_fetch", EndpointMeta{
		Summary:     "批量抓取视频详情",
		Tags:        tagVideo,
		OperationID: "batch_fetch_video_details",
	})
	RegisterEndpointMeta("GET", "/video_details/batch_fetch_from_history", EndpointMeta{
		Summary:     "从历史记录批量抓取视频详情",
		Tags:        tagVideo,
		OperationID: "batch_fetch_from_history",
	})
	RegisterEndpointMeta("GET", "/video_details/stats", EndpointMeta{
		Summary:     "视频统计信息",
		Tags:        tagVideo,
		OperationID: "get_video_stats",
	})
	RegisterEndpointMeta("GET", "/video_details/database_stats", EndpointMeta{
		Summary:     "数据库统计信息",
		Tags:        tagVideo,
		OperationID: "get_database_stats",
	})
	RegisterEndpointMeta("GET", "/video_details/uploaders", EndpointMeta{
		Summary:     "获取UP主列表",
		Tags:        tagVideo,
		OperationID: "get_uploader_list",
	})
	RegisterEndpointMeta("GET", "/video_details/tags", EndpointMeta{
		Summary:     "获取标签列表",
		Tags:        tagVideo,
		OperationID: "get_tag_list",
	})
	RegisterEndpointMeta("GET", "/video_details/uploader/:mid", EndpointMeta{
		Summary:     "获取UP主详情",
		Tags:        tagVideo,
		OperationID: "get_uploader_detail",
	})
	RegisterEndpointMeta("POST", "/video_details/stop", EndpointMeta{
		Summary:     "停止视频详情抓取",
		Tags:        tagVideo,
		OperationID: "stop_video_detail_fetch",
	})
	RegisterEndpointMeta("POST", "/video_details/reset", EndpointMeta{
		Summary:     "重置视频详情抓取进度",
		Tags:        tagVideo,
		OperationID: "reset_video_detail_progress",
	})
	RegisterEndpointMeta("GET", "/video_details/progress", EndpointMeta{
		Summary:     "获取视频详情抓取进度",
		Tags:        tagVideo,
		OperationID: "get_video_detail_progress",
	})

	// ========== 图片管理 ==========
	RegisterEndpointMeta("POST", "/images/start", EndpointMeta{
		Summary:     "开始下载图片",
		Tags:        tagImage,
		OperationID: "start_image_download",
	})
	RegisterEndpointMeta("POST", "/images/stop", EndpointMeta{
		Summary:     "停止下载图片",
		Tags:        tagImage,
		OperationID: "stop_image_download",
	})
	RegisterEndpointMeta("GET", "/images/status", EndpointMeta{
		Summary:     "获取图片下载状态",
		Tags:        tagImage,
		OperationID: "get_image_download_status",
	})
	RegisterEndpointMeta("POST", "/images/clear", EndpointMeta{
		Summary:     "清理已下载图片",
		Tags:        tagImage,
		OperationID: "clear_images",
	})
	RegisterEndpointMeta("GET", "/images/local/:image_type/:file_hash", EndpointMeta{
		Summary:     "获取本地图片",
		Tags:        tagImage,
		OperationID: "get_local_image",
	})
	RegisterEndpointMeta("GET", "/images/proxy", EndpointMeta{
		Summary:     "图片代理（转发B站CDN）",
		Tags:        tagImage,
		OperationID: "proxy_image",
	})

	// ========== 下载管理 ==========
	RegisterEndpointMeta("GET", "/download/check_video_download", EndpointMeta{
		Summary:     "检查视频下载状态",
		Tags:        tagDownload,
		OperationID: "check_video_download",
	})
	RegisterEndpointMeta("GET", "/download/list_downloaded_videos", EndpointMeta{
		Summary:     "列出已下载视频",
		Tags:        tagDownload,
		OperationID: "list_downloaded_videos",
	})
	RegisterEndpointMeta("DELETE", "/download/delete_downloaded_video", EndpointMeta{
		Summary:     "删除已下载视频",
		Tags:        tagDownload,
		OperationID: "delete_downloaded_video",
	})

	// ========== 标题分析 ==========
	RegisterEndpointMeta("GET", "/title/stats", EndpointMeta{
		Summary:     "标题统计分析",
		Tags:        tagTitle,
		OperationID: "get_title_stats",
	})
	RegisterEndpointMeta("GET", "/title/patterns", EndpointMeta{
		Summary:     "标题模式分析",
		Tags:        tagTitle,
		OperationID: "get_title_patterns",
	})
	RegisterEndpointMeta("GET", "/title/sentiment", EndpointMeta{
		Summary:     "标题情感分析",
		Tags:        tagTitle,
		OperationID: "get_title_sentiment",
	})
	RegisterEndpointMeta("GET", "/title/length", EndpointMeta{
		Summary:     "标题长度分析",
		Tags:        tagTitle,
		OperationID: "get_title_length_analysis",
	})
	RegisterEndpointMeta("GET", "/title/trend", EndpointMeta{
		Summary:     "标题趋势分析",
		Tags:        tagTitle,
		OperationID: "get_title_trend",
	})

	// ========== 互动记录 ==========
	RegisterEndpointMeta("GET", "/interactions/list", EndpointMeta{
		Summary:     "获取互动记录列表",
		Tags:        tagInteraction,
		OperationID: "get_interaction_records",
	})
	RegisterEndpointMeta("POST", "/interactions/sync", EndpointMeta{
		Summary:     "同步互动记录",
		Tags:        tagInteraction,
		OperationID: "sync_interaction_records",
	})
}
