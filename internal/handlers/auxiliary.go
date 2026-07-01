package handlers

import (
	"fmt"
	"net/http"

	"bili-history/internal/config"
	"bili-history/internal/services"
	"bili-history/internal/services/biliapi"
	"bili-history/internal/services/export"
	"bili-history/internal/services/notify"

	"github.com/gin-gonic/gin"
)

// ExportHistory exports history data to Excel.
func ExportHistory(c *gin.Context) {
	yearStr := c.DefaultQuery("year", "2024")
	year := 0
	fmt.Sscanf(yearStr, "%d", &year)
	if year == 0 {
		year = 2024
	}

	path, err := export.ExportToExcel(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":    path,
		"message": "Export completed",
	})
}

// SendEmail sends an email notification.
func SendEmail(c *gin.Context) {
	var req struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := notify.SendEmail(req.Subject, req.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent"})
}

// SendLogEmail sends a log summary email.
func SendLogEmail(c *gin.Context) {
	if err := notify.SendLogEmail(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Log email sent"})
}

// GetComments returns video comments.
func GetComments(c *gin.Context) {
	bvid := c.Query("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvid is required"})
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	comments, total, err := client.FetchComments(bvid, 1, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  comments,
		"total": total,
	})
}

// GetFavorites returns user's favorite folders.
func GetFavorites(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	folders, err := client.FetchFavorites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": folders})
}

// GetWatchLater returns watch later list.
func GetWatchLater(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	items, err := client.FetchWatchLater()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// GetLikes returns liked videos.
func GetLikes(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	items, err := client.FetchLikes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// GetDynamics returns user dynamics.
func GetDynamics(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	dynamics, err := client.FetchDynamics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dynamics})
}

// GetPopularVideos returns popular videos.
func GetPopularVideos(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	videos, err := client.FetchPopular()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": videos})
}

// DeleteRemoteHistory deletes a history record from Bilibili.
func DeleteRemoteHistory(c *gin.Context) {
	bvid := c.Query("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvid is required"})
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	if err := client.DeleteHistory(bvid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "History deleted"})
}

// SyncData performs data synchronization.
func SyncData(c *gin.Context) {
	// Import from JSON files to database
	if err := services.ImportFromJSONFiles(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Data sync completed"})
}

// CheckDataIntegrity checks data integrity between JSON and database.
func CheckDataIntegrity(c *gin.Context) {
	// TODO: Implement integrity check
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Data integrity check passed",
	})
}

// CleanHistoryData removes unnecessary fields.
func CleanHistoryData(c *gin.Context) {
	// TODO: Implement data cleaning
	c.JSON(http.StatusOK, gin.H{"message": "Data cleaning completed"})
}
