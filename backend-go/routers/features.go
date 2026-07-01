package routers

import (
	"net/http"

	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

func RegisterFavoriteRoutes(r *gin.RouterGroup) {
	favorite := r.Group("/favorite")
	{
		favorite.GET("/list", getFavoriteList)
		favorite.POST("/sync", syncFavorites)
	}

	like := r.Group("/like")
	{
		like.GET("/list", getLikeList)
		like.POST("/sync", syncLikes)
	}

	watchlater := r.Group("/watchlater")
	{
		watchlater.GET("/list", getWatchLaterList)
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

func getFavoriteList(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"records": []interface{}{},
		"total":   0,
		"message": "收藏夹功能待实现",
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
		"records": []interface{}{},
		"total":   0,
		"message": "点赞列表功能待实现",
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
		"records": []interface{}{},
		"total":   0,
		"message": "稍后再看功能待实现",
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
