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
		favorite.GET("/collected/list", getCollectedFavoriteFolders)
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
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA，无法操作收藏"))
		return
	}

	var body struct {
		Resources    string `json:"resources"`
		MediaIDs     string `json:"media_ids"`
		Type         int    `json:"type"`
		Rid          int64  `json:"rid"`
		AddMediaIDs  string `json:"add_media_ids"`
		DelMediaIDs  string `json:"del_media_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)

	// Support both formats: resources/media_ids or rid/type/add_media_ids/del_media_ids
	if body.Resources != "" && body.MediaIDs != "" {
		if err := client.DealFavoriteResource(body.Resources, body.MediaIDs); err != nil {
			if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
				c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
				return
			}
			c.JSON(http.StatusOK, models.ErrorResponse("收藏操作失败: "+err.Error()))
			return
		}
	} else if body.Rid > 0 && (body.AddMediaIDs != "" || body.DelMediaIDs != "") {
		resources := fmt.Sprintf("%d:%d", body.Rid, body.Type)
		mediaIDs := body.AddMediaIDs
		if mediaIDs == "" {
			mediaIDs = body.DelMediaIDs
		}
		if err := client.DealFavoriteResource(resources, mediaIDs); err != nil {
			if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
				c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
				return
			}
			c.JSON(http.StatusOK, models.ErrorResponse("收藏操作失败: "+err.Error()))
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少必要参数"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message": "操作成功",
	}))
}

func batchFavoriteResource(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA，无法操作收藏"))
		return
	}

	var body struct {
		Resources string `json:"resources"`
		MediaIDs  string `json:"media_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	if body.Resources == "" || body.MediaIDs == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("resources 和 media_ids 不能为空"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	if err := client.DealFavoriteResource(body.Resources, body.MediaIDs); err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
			c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
			return
		}
		c.JSON(http.StatusOK, models.ErrorResponse("批量收藏操作失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"message": "操作成功",
	}))
}

func localBatchFavoriteResource(c *gin.Context) {
	db := database.GetFavoritesDB()
	if db == nil {
		c.JSON(http.StatusOK, models.ErrorResponse("收藏数据库不可用"))
		return
	}

	var body struct {
		MediaID int64    `json:"media_id"`
		Bvids   []string `json:"bvids"`
		Action  string   `json:"action"` // "add" or "remove"
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	if body.MediaID == 0 || len(body.Bvids) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("media_id 和 bvids 不能为空"))
		return
	}

	affected := 0
	if body.Action == "remove" {
		for _, bvid := range body.Bvids {
			result, err := db.Exec("DELETE FROM favorites_content WHERE media_id = ? AND bvid = ?", body.MediaID, bvid)
			if err != nil {
				continue
			}
			n, _ := result.RowsAffected()
			affected += int(n)
		}
	} else {
		// add: insert placeholder entries (user can sync later to fill details)
		for _, bvid := range body.Bvids {
			_, err := db.Exec(`INSERT OR IGNORE INTO favorites_content
				(media_id, content_id, bvid, type, fetch_time)
				VALUES (?, 0, ?, 2, ?)`,
				body.MediaID, bvid, time.Now().Unix())
			if err != nil {
				continue
			}
			affected++
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"affected": affected,
		"message":  "操作成功",
	}))
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

func syncFavorites(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA，无法同步收藏夹"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)

	// Fetch folder list
	folderData, err := client.GetFavoriteFolderList()
	if err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
			c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
			return
		}
		c.JSON(http.StatusOK, models.ErrorResponse("获取收藏夹列表失败: "+err.Error()))
		return
	}

	// Save folders locally
	folders := make([]database.FavoriteFolder, 0, len(folderData.List))
	for _, f := range folderData.List {
		folders = append(folders, database.FavoriteFolder{
			MediaID:    f.ID,
			Fid:        f.Fid,
			Mid:        f.Mid,
			Title:      f.Title,
			Cover:      f.Cover,
			Attr:       f.Attr,
			Intro:      f.Intro,
			Ctime:      f.Ctime,
			Mtime:      f.Mtime,
			State:      f.State,
			MediaCount: f.MediaCount,
			FavState:   f.FavState,
			LikeState:  f.LikeState,
		})
	}
	if err := database.SaveFavoriteFolders(folders); err != nil {
		_ = err // non-fatal
	}

	// Fetch contents for each folder (max 500 items per folder)
	totalContents := 0
	for _, folder := range folders {
		if folder.MediaCount == 0 {
			continue
		}
		maxItems := 500
		for pn := 1; pn*20 < maxItems; pn++ {
			res, err := client.GetFavoriteResources(folder.MediaID, pn, 20)
			if err != nil {
				break
			}
			if len(res.Media) == 0 {
				break
			}
			contents := make([]database.FavoriteContent, 0, len(res.Media))
			for _, item := range res.Media {
			bvid := ""
			if item.UGC != nil {
				bvid = item.UGC.Bvid
			}
			var statView, statDanmaku, statReply, statFavorite int
			if item.Stat != nil {
				statView = item.Stat.View
				statDanmaku = item.Stat.Danmaku
				statReply = item.Stat.Reply
				statFavorite = item.Stat.Favorite
			}
				contents = append(contents, database.FavoriteContent{
					MediaID:     folder.MediaID,
					ContentID:   item.ID,
					Type:        item.Type,
					Title:       item.Title,
					Cover:       item.Cover,
					Bvid:        bvid,
					Intro:       item.Intro,
					Page:        item.Page,
					Duration:    item.Duration,
					UpperMid:    item.Upper.Mid,
					Attr:        item.Attr,
					Ctime:       item.Ctime,
					Pubtime:     item.Pubtime,
					FavTime:     item.FavTime,
					CreatorName: item.Upper.Name,
					CreatorFace: item.Upper.Face,
					BvID:        bvid,
					Collect:     statFavorite,
					Play:        statView,
					Danmaku:     statDanmaku,
					Reply:       statReply,
					FirstCid:    item.Cid,
				})
			}
			if err := database.SaveFavoriteContents(folder.MediaID, contents); err != nil {
				_ = err
			}
			totalContents += len(contents)
			if res.Page.Total <= pn*res.Page.Size {
				break
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"folders_count":  len(folders),
		"contents_total": totalContents,
	}))
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
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 SESSDATA，无法同步点赞"))
		return
	}

	vmid, _ := strconv.ParseInt(cfg.DedeUserID, 10, 64)
	if vmid == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("未配置 DedeUserID"))
		return
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)

	res, err := client.GetLikedVideos(vmid)
	if err != nil {
		if apiErr, ok := err.(*biliapi.ApiError); ok && apiErr.Code == -6 {
			c.JSON(http.StatusOK, models.ErrorResponse("Cookie 已过期，请重新登录"))
			return
		}
		c.JSON(http.StatusOK, models.ErrorResponse("获取点赞列表失败: "+err.Error()))
		return
	}

	allVideos := make([]database.LikeVideo, 0, len(res.List))
	for _, item := range res.List {
		allVideos = append(allVideos, database.LikeVideo{
			Bvid:      item.Bvid,
			Aid:       item.Aid,
			Title:     item.Title,
			Pic:       item.Pic,
			Desc:      item.Desc,
			Duration:  item.Duration,
			Tid:       item.Tid,
			Tname:     item.Tname,
			OwnerName: item.Owner.Name,
			OwnerMid:  int64(item.Owner.Mid),
			OwnerFace: item.Owner.Face,
			Pubdate:   item.Pubdate,
			View:      item.Stat.View,
			Danmaku:   item.Stat.Danmaku,
			LikeCount: item.Stat.Like,
		})
	}

	if err := database.SaveLikedVideos(allVideos); err != nil {
		_ = err // non-fatal
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"total": len(allVideos),
	}))
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


