package routers

import (
	"net/http"
	"strconv"
	"strings"

	"bilibili-history-go/models"

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
	}

	dynamic := r.Group("/dynamic")
	{
		dynamic.GET("/list", getDynamicList)
		dynamic.POST("/sync", syncDynamic)
	}

	comment := r.Group("/comment")
	{
		comment.GET("/list", getCommentList)
		comment.POST("/sync", syncComments)
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
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":     []interface{}{},
		"total":    0,
		"has_more": false,
	}))
}

func getLocalFavoriteContents(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":  []interface{}{},
		"total": 0,
		"page":  1,
		"size":  20,
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
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":    []interface{}{},
		"total":   0,
		"has_more": false,
		"message": "收藏夹功能待实现",
	}))
}

func getFavoriteFolderList(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":     []interface{}{},
		"count":    0,
		"season":   []interface{}{},
		"message":  "收藏夹列表功能待实现",
	}))
}

func syncFavorites(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "收藏夹同步功能待实现",
	})
}

func getLikeList(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":    []interface{}{},
		"total":   0,
		"message": "点赞列表功能待实现",
	}))
}

func getLikeLocal(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":    []interface{}{},
		"total":   0,
		"page":    1,
		"size":    50,
		"message": "本地点赞列表功能待实现",
	}))
}

func syncLikes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "点赞同步功能待实现",
	})
}

func getWatchLaterList(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":    []interface{}{},
		"total":   0,
		"message": "稍后再看功能待实现",
	}))
}

func getWatchLaterLocal(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"list":    []interface{}{},
		"total":   0,
		"page":    1,
		"size":    50,
		"message": "本地稍后再看功能待实现",
	}))
}

func syncWatchLater(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "稍后再看同步功能待实现",
	})
}

func getDynamicList(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"records": []interface{}{},
		"total":   0,
		"message": "动态列表功能待实现",
	}))
}

func syncDynamic(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "动态同步功能待实现",
	})
}

func getCommentList(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"records": []interface{}{},
		"total":   0,
		"message": "评论功能待实现",
	}))
}

func syncComments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "评论同步功能待实现",
	})
}
