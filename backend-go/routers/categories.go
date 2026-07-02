package routers

import (
	"net/http"

	"bilibili-history-go/database"
	"bilibili-history-go/models"

	"github.com/gin-gonic/gin"
)

func RegisterCategoryRoutes(r *gin.RouterGroup) {
	categories := r.Group("/categories")
	{
		categories.POST("/init", initCategories)
		categories.GET("/categories", getCategories)
		categories.GET("/main-categories", getMainCategories)
		categories.GET("/sub-categories/:main_category", getSubCategories)
	}
}

func initCategories(c *gin.Context) {
	err := database.InitCategories()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse("初始化失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "视频分类表初始化成功",
	})
}

func getCategories(c *gin.Context) {
	categories, err := database.GetCategories()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(categories))
}

func getMainCategories(c *gin.Context) {
	categories, err := database.GetMainCategories()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(categories))
}

func getSubCategories(c *gin.Context) {
	mainCategory := c.Param("main_category")

	categories, err := database.GetSubCategories(mainCategory)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(categories))
}
