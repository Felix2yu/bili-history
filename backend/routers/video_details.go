package routers

import (
	"net/http"
	"strconv"

	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/services"

	"github.com/gin-gonic/gin"
)

// RegisterVideoDetailsRoutes 注册视频详情相关路由
func RegisterVideoDetailsRoutes(r *gin.RouterGroup) {
	videoDetails := r.Group("/video_details")
	{
		videoDetails.GET("/fetch/:bvid", fetchVideoDetail)
		videoDetails.GET("/info/:bvid", getVideoInfo)
		videoDetails.GET("/search", searchVideos)
		videoDetails.POST("/batch_fetch", batchFetchVideoDetails)
		videoDetails.GET("/batch_fetch_from_history", batchFetchFromHistory)
		videoDetails.GET("/stats", getVideoStats)
		videoDetails.GET("/database_stats", getDatabaseStats)
		videoDetails.GET("/uploaders", getUploaderList)
		videoDetails.GET("/tags", getTagList)
		videoDetails.GET("/uploader/:mid", getUploaderDetail)
		videoDetails.POST("/stop", stopVideoDetailFetch)
		videoDetails.POST("/reset", resetVideoDetailProgress)
		videoDetails.GET("/progress", getVideoDetailProgress)
	}
}

// fetchVideoDetail 获取单个视频详情（从B站）
func fetchVideoDetail(c *gin.Context) {
	bvid := c.Param("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("bvid不能为空"))
		return
	}

	video, err := services.FetchVideoDetailFromBili(bvid)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	err = services.SaveVideoDetail(video)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("保存视频详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(video))
}

// getVideoInfo 从数据库获取视频信息
func getVideoInfo(c *gin.Context) {
	bvid := c.Param("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("bvid不能为空"))
		return
	}

	video, err := database.GetVideoBaseInfoByBvid(bvid)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	if video == nil {
		c.JSON(http.StatusOK, models.SuccessResponse(nil))
		return
	}

	result := map[string]interface{}{
		"id":              video.ID,
		"bvid":            video.Bvid,
		"aid":             video.Aid,
		"videos":          video.Videos,
		"tid":             video.Tid,
		"tname":           video.Tname,
		"copyright":       video.Copyright,
		"pic":             video.Pic,
		"title":           video.Title,
		"pubdate":         video.Pubdate,
		"ctime":           video.Ctime,
		"desc":            video.Desc,
		"duration":        video.Duration,
		"cid":             video.Cid,
		"owner_mid":       video.OwnerMid,
		"owner_name":      video.OwnerName,
		"owner_face":      video.OwnerFace,
		"stat_view":       video.StatView,
		"stat_danmaku":    video.StatDanmaku,
		"stat_reply":      video.StatReply,
		"stat_favorite":   video.StatFavorite,
		"stat_coin":       video.StatCoin,
		"stat_share":      video.StatShare,
		"stat_like":       video.StatLike,
		"fetch_time":      video.FetchTime,
		"update_time":     video.UpdateTime,
		"original_url":    "https://www.bilibili.com/video/" + video.Bvid,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// searchVideos 搜索视频
func searchVideos(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if keyword == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("搜索关键词不能为空"))
		return
	}

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	result, err := database.SearchVideos(keyword, page, size)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"records": result.Records,
		"total":   result.Total,
		"size":    result.Size,
		"current": result.Current,
		"keyword": keyword,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// batchFetchVideoDetails 批量获取视频详情
func batchFetchVideoDetails(c *gin.Context) {
	var req models.BatchFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	result, err := services.BatchFetchVideoDetails(req.Bvids)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// batchFetchFromHistory 从历史记录批量获取视频详情
func batchFetchFromHistory(c *gin.Context) {
	skipExisting := c.DefaultQuery("skip_existing", "true") == "true"

	result, err := services.BatchFetchFromHistory(skipExisting)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// getVideoStats 获取视频详情统计信息
func getVideoStats(c *gin.Context) {
	stats, err := database.GetVideoDetailStats()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	historyBvids, err := database.GetUniqueBvidsFromHistory()
	if err == nil {
		stats.TotalVideos = int64(len(historyBvids))
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}

// getDatabaseStats 获取数据库详细统计
func getDatabaseStats(c *gin.Context) {
	stats, err := database.GetDatabaseStats()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}

// getUploaderList 获取UP主列表
func getUploaderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sortBy := c.DefaultQuery("sort_by", "video_count")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	validSortBy := map[string]bool{
		"video_count": true,
		"views":       true,
		"likes":       true,
	}
	if !validSortBy[sortBy] {
		sortBy = "video_count"
	}

	result, err := database.GetUploaderList(page, size, sortBy)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"records": result.Records,
		"total":   result.Total,
		"size":    result.Size,
		"current": result.Current,
		"sort_by": sortBy,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getTagList 获取标签列表
func getTagList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	result, err := database.GetTagList(page, size)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	data := map[string]interface{}{
		"records": result.Records,
		"total":   result.Total,
		"size":    result.Size,
		"current": result.Current,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// getUploaderDetail 获取UP主详细信息
func getUploaderDetail(c *gin.Context) {
	midStr := c.Param("mid")
	mid, err := strconv.Atoi(midStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的UP主ID"))
		return
	}

	result, err := database.GetUploaderDetail(mid)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// stopVideoDetailFetch 停止获取任务
func stopVideoDetailFetch(c *gin.Context) {
	err := services.StopVideoDetailFetch()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message": "已发送停止信号",
	}))
}

// resetVideoDetailProgress 重置获取状态
func resetVideoDetailProgress(c *gin.Context) {
	services.ResetVideoDetailProgress()

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message": "已重置获取状态",
	}))
}

// getVideoDetailProgress 获取进度
func getVideoDetailProgress(c *gin.Context) {
	progress := services.GetVideoDetailProgress()
	c.JSON(http.StatusOK, models.SuccessResponse(progress))
}
