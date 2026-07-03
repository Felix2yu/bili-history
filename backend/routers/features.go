package routers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/services"

	"github.com/gin-gonic/gin"
)

func RegisterFavoriteRoutes(r *gin.RouterGroup) {
	favorite := r.Group("/favorite")
	{
		favorite.GET("/list", getFavoriteList)
		favorite.GET("/folder/created/list-all", getFavoriteFolderList)
		favorite.GET("/folder/collected/list", getCollectedFavoriteFolders)
		favorite.GET("/folder/resource/list", getFavoriteFolderContents)
		favorite.GET("/content/list", getLocalFavoriteContents)
		favorite.POST("/sync", syncFavorites)
		favorite.POST("/resource/deal", favoriteResource)
		favorite.POST("/resource/batch-deal", batchFavoriteResource)
		favorite.POST("/resource/local-batch-deal", localBatchFavoriteResource)
		favorite.POST("/check/batch", batchCheckFavoriteStatus)
	}

	like := r.Group("/like")
	{
		like.GET("/list", getLikeList)
		like.GET("/local", getLikeLocal)
		like.POST("/sync", syncLikes)
	}

	watchlater := r.Group("/watchlater")
	{
		watchlater.GET("/list", getWatchLaterList)
		watchlater.GET("/local", getWatchLaterLocal)
		watchlater.POST("/sync", syncWatchLater)
		watchlater.DELETE("/:bvid", deleteWatchLaterVideo)
		watchlater.POST("/batch-delete", batchDeleteWatchLaterVideos)
	}

	dynamic := r.Group("/dynamic")
	{
		dynamic.GET("/db/hosts", getDynamicDbHosts)
		dynamic.GET("/db/space/:host_mid", getDynamicDbSpace)
		dynamic.GET("/space/auto/:host_mid", startDynamicAutoFetch)
		dynamic.GET("/space/auto/:host_mid/progress", dynamicProgressSSE)
		dynamic.POST("/space/auto/:host_mid/stop", stopDynamicAutoFetch)
		dynamic.DELETE("/space/:host_mid", deleteDynamicSpace)
		// Legacy stubs
		dynamic.GET("/list", getDynamicListLegacy)
		dynamic.POST("/sync", syncDynamicLegacy)
	}
}

type BatchCheckFavoriteRequest struct {
	Oids interface{} `json:"oids"`
}

func batchCheckFavoriteStatus(c *gin.Context) {
	var req BatchCheckFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	var oids []int64

	switch v := req.Oids.(type) {
	case string:
		oidStrs := strings.Split(v, ",")
		for _, s := range oidStrs {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				oids = append(oids, id)
			}
		}
	case []interface{}:
		for _, item := range v {
			switch id := item.(type) {
			case float64:
				oids = append(oids, int64(id))
			case int:
				oids = append(oids, int64(id))
			case int64:
				oids = append(oids, id)
			case string:
				if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
					oids = append(oids, parsed)
				}
			}
		}
	}

	results := make([]map[string]interface{}, 0, len(oids))
	for _, oid := range oids {
		results = append(results, map[string]interface{}{
			"oid":              oid,
			"is_favorited":     false,
			"favorite_folders": []interface{}{},
		})
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"results": results,
	}))
}

func getCollectedFavoriteFolders(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  []interface{}{},
		"count": 0,
		"has_more": false,
	}))
}

func getFavoriteFolderContents(c *gin.Context) {
	mediaIDStr := c.Query("media_id")
	mediaID, _ := strconv.ParseInt(mediaIDStr, 10, 64)
	pn, _ := strconv.Atoi(c.DefaultQuery("pn", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("ps", "20"))

	list, total, err := database.GetFavoriteContents(mediaID, pn, ps)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("获取收藏夹内容失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":     list,
		"total":    total,
		"has_more": total > pn*ps,
	}))
}

func getLocalFavoriteContents(c *gin.Context) {
	mediaIDStr := c.Query("media_id")
	mediaID, _ := strconv.ParseInt(mediaIDStr, 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	list, total, err := database.GetFavoriteContents(mediaID, page, size)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("获取本地收藏内容失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	}))
}

func favoriteResource(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "收藏操作功能待实现",
	})
}

func batchFavoriteResource(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "批量收藏功能待实现",
	})
}

func localBatchFavoriteResource(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "本地批量收藏功能待实现",
	})
}

func getFavoriteList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	list, total, err := database.GetFavoriteFolders(true)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("获取收藏夹列表失败: "+err.Error()))
		return
	}

	start := (page - 1) * size
	end := start + size
	if start > len(list) {
		start = len(list)
	}
	if end > len(list) {
		end = len(list)
	}
	pagedList := list[start:end]

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  pagedList,
		"total": total,
		"page":  page,
		"size":  size,
	}))
}

func getFavoriteFolderList(c *gin.Context) {
	list, count, err := database.GetFavoriteFolders(true)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("获取收藏夹列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":   list,
		"count":  count,
		"season": []interface{}{},
	}))
}

func syncFavorites(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "收藏夹同步功能待实现",
	})
}

func getLikeList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sort := c.DefaultQuery("sort", "pubdate")
	order := c.DefaultQuery("order", "desc")

	list, total, err := database.GetLikedVideos(page, size, sort, order)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("获取点赞列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"size":     size,
		"has_more": total > page*size,
	}))
}

func getLikeLocal(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sort := c.DefaultQuery("sort", "fetch_time")
	order := c.DefaultQuery("order", "desc")

	list, total, err := database.GetLikedVideos(page, size, sort, order)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("获取本地点赞列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	}))
}

func syncLikes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "点赞同步功能待实现",
	})
}

func getWatchLaterList(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA，无法访问 B 站稍后再看"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	data, err := client.GetWatchLaterList()
	if err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
			c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
			return
		}
		c.JSON(http.StatusOK, models.ErrorResponse("获取稍后再看列表失败: "+err.Error()))
		return
	}

	// Convert to local schema and persist so the local cache stays fresh.
	localVideos := make([]database.WatchLaterVideo, 0, len(data.List))
	for _, item := range data.List {
		localVideos = append(localVideos, database.WatchLaterVideo{
			Bvid:       item.Bvid,
			Aid:        item.Aid,
			Title:      item.Title,
			Pic:        item.Pic,
			Desc:       item.Desc,
			Duration:   item.Duration,
			Tid:        item.Tid,
			Tname:      item.Tname,
			OwnerName:  item.Owner.Name,
			OwnerMid:   int64(item.Owner.Mid),
			OwnerFace:  item.Owner.Face,
			AddAt:      item.AddAt,
			Pubdate:    item.Pubdate,
			View:       item.Stat.View,
			Danmaku:    item.Stat.Danmaku,
			Link:       "https://www.bilibili.com/video/" + item.Bvid,
		})
	}
	if saveErr := database.SaveWatchLaterVideos(localVideos); saveErr != nil {
		// Non-fatal: we still return the remote list to the user.
		_ = saveErr
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  localVideos,
		"total": len(localVideos),
	}))
}

func getWatchLaterLocal(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sort := c.DefaultQuery("sort", "add_at")
	order := c.DefaultQuery("order", "desc")

	list, total, err := database.GetWatchLaterVideos(page, size, sort, order)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("获取本地稍后再看列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	}))
}

func syncWatchLater(c *gin.Context) {
	// /sync behaves identically to /list: pull the full remote list and refresh
	// the local cache. Keeping a separate endpoint so the frontend can call it
	// explicitly without changing the list response semantics.
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA，无法同步 B 站稍后再看"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	data, err := client.GetWatchLaterList()
	if err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
			c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
			return
		}
		c.JSON(http.StatusOK, models.ErrorResponse("同步稍后再看失败: "+err.Error()))
		return
	}

	localVideos := make([]database.WatchLaterVideo, 0, len(data.List))
	for _, item := range data.List {
		localVideos = append(localVideos, database.WatchLaterVideo{
			Bvid:       item.Bvid,
			Aid:        item.Aid,
			Title:      item.Title,
			Pic:        item.Pic,
			Desc:       item.Desc,
			Duration:   item.Duration,
			Tid:        item.Tid,
			Tname:      item.Tname,
			OwnerName:  item.Owner.Name,
			OwnerMid:   int64(item.Owner.Mid),
			OwnerFace:  item.Owner.Face,
			AddAt:      item.AddAt,
			Pubdate:    item.Pubdate,
			View:       item.Stat.View,
			Danmaku:    item.Stat.Danmaku,
			Link:       "https://www.bilibili.com/video/" + item.Bvid,
		})
	}
	if saveErr := database.SaveWatchLaterVideos(localVideos); saveErr != nil {
		_ = saveErr
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  localVideos,
		"total": len(localVideos),
	}))
}

// deleteWatchLaterVideo removes a single video from B 站稍后再看 by bvid.
// It looks up the aid from the local cache (since the B 站 delete API needs aid),
// then calls the remote delete API and removes the local row on success.
func deleteWatchLaterVideo(c *gin.Context) {
	bvid := c.Param("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 bvid 参数"))
		return
	}

	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" || cfg.BiliJct == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA / bili_jct，无法删除"))
		return
	}

	// Find the aid from local cache; if missing, we cannot call the remote API.
	local, err := database.GetWatchLaterVideoByBvid(bvid)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("查询本地记录失败: "+err.Error()))
		return
	}
	if local == nil {
		c.JSON(http.StatusOK, models.ErrorResponse("本地未找到该视频，请先同步稍后再看列表"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	if err := client.RemoveFromWatchLater(local.Aid); err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
			c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
			return
		}
		c.JSON(http.StatusOK, models.ErrorResponse("删除失败: "+err.Error()))
		return
	}

	// Remove from local cache.
	_ = database.DeleteWatchLaterVideo(bvid)

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"bvid":  bvid,
		"aid":   local.Aid,
	}))
}

// batchDeleteWatchLaterRequest is the body for POST /watchlater/batch-delete.
// bvids can be a JSON array of strings, or a comma-separated string for convenience.
type batchDeleteWatchLaterRequest struct {
	Bvids interface{} `json:"bvids"`
}

func batchDeleteWatchLaterVideos(c *gin.Context) {
	var req batchDeleteWatchLaterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	var bvids []string
	switch v := req.Bvids.(type) {
	case string:
		for _, s := range strings.Split(v, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				bvids = append(bvids, s)
			}
		}
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					bvids = append(bvids, s)
				}
			}
		}
	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse("bvids 参数格式错误"))
		return
	}

	if len(bvids) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("未提供要删除的 bvid"))
		return
	}

	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" || cfg.BiliJct == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA / bili_jct，无法删除"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)

	type delResult struct {
		Bvid    string `json:"bvid"`
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	results := make([]delResult, 0, len(bvids))
	successCount := 0
	for _, bvid := range bvids {
		local, err := database.GetWatchLaterVideoByBvid(bvid)
		if err != nil || local == nil {
			results = append(results, delResult{Bvid: bvid, Error: "本地未找到该视频"})
			continue
		}
		if err := client.RemoveFromWatchLater(local.Aid); err != nil {
			errMsg := err.Error()
			if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
				errMsg = "Cookie 已过期，请重新登录"
			}
			results = append(results, delResult{Bvid: bvid, Error: errMsg})
			continue
		}
		_ = database.DeleteWatchLaterVideo(bvid)
		results = append(results, delResult{Bvid: bvid, Success: true})
		successCount++
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"results": results,
		"total":   len(bvids),
		"success": successCount,
		"failed":  len(bvids) - successCount,
	}))
}

func getDynamicListLegacy(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"records": []interface{}{},
		"total":   0,
		"message": "请使用 /dynamic/db/hosts 和 /dynamic/db/space 端点",
	}))
}

func syncDynamicLegacy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "请使用 /dynamic/space/auto/{host_mid} 端点",
	})
}

func getDynamicDbHosts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	hosts, err := database.GetDynamicHosts(limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if hosts == nil {
		hosts = []database.DynamicHost{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": hosts})
}

func getDynamicDbSpace(c *gin.Context) {
	hostMid := c.Param("host_mid")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	total, items, err := database.GetDynamicSpace(hostMid, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if items == nil {
		items = []database.DynamicItem{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "total": total, "items": items})
}

func startDynamicAutoFetch(c *gin.Context) {
	hostMid := c.Param("host_mid")
	if hostMid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "host_mid 不能为空"})
		return
	}

	needTop := c.DefaultQuery("need_top", "false") == "true"
	saveToDB := c.DefaultQuery("save_to_db", "true") == "true"
	saveMedia := c.DefaultQuery("save_media", "true") == "true"

	status := services.GetDynamicFetchStatus(hostMid)
	if status.IsRunning {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "该 UP 的动态抓取正在进行中"})
		return
	}

	go services.FetchDynamicSpace(hostMid, needTop, saveToDB, saveMedia)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "开始抓取动态"})
}

func dynamicProgressSSE(c *gin.Context) {
	_ = c.Param("host_mid") // reserved for future per-host channels

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "不支持流式传输"})
		return
	}

	ch := services.GetDynamicProgressChan()
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "event: progress\ndata: {\"message\":\"%s\"}\n\n", strings.ReplaceAll(msg, `"`, `\"`))
			flusher.Flush()
			// Check for completion
			if strings.Contains(msg, "全部抓取完毕") || strings.Contains(msg, "已停止") {
				return
			}
		case <-timeout:
			return
		case <-c.Request.Context().Done():
			return
		}
	}
}

func stopDynamicAutoFetch(c *gin.Context) {
	hostMid := c.Param("host_mid")
	services.StopDynamicFetch(hostMid)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已发送停止信号"})
}

func deleteDynamicSpace(c *gin.Context) {
	hostMid := c.Param("host_mid")
	if err := database.DeleteDynamicSpace(hostMid); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}


