package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"bili-history/internal/services/download"

	"github.com/gin-gonic/gin"
)

// StartDownload starts a video download with SSE progress streaming.
func StartDownload(c *gin.Context) {
	var req struct {
		URL  string `json:"url" binding:"required"`
		Type string `json:"type"` // "video" or "audio"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	manager, err := download.NewDownloadManager()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize download manager"})
		return
	}

	// SSE streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	progressCh := make(chan download.Progress, 100)
	ctx := context.Background()

	go func() {
		if req.Type == "audio" {
			manager.DownloadAudio(ctx, req.URL, progressCh)
		} else {
			manager.DownloadVideo(ctx, req.URL, progressCh)
		}
		close(progressCh)
	}()

	for progress := range progressCh {
		event := formatSSE("progress", progress)
		c.Writer.Write([]byte(event))
		flusher.Flush()
		if progress.Status == "completed" || progress.Status == "cancelled" {
			break
		}
	}
}

// CancelDownload cancels an active download.
func CancelDownload(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	manager, err := download.NewDownloadManager()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize download manager"})
		return
	}

	cancelled := manager.CancelDownload(req.URL)
	c.JSON(http.StatusOK, gin.H{
		"cancelled": cancelled,
		"message":   "Download cancelled",
	})
}

// GetActiveDownloads returns list of active downloads.
func GetActiveDownloads(c *gin.Context) {
	manager, err := download.NewDownloadManager()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize download manager"})
		return
	}

	downloads := manager.GetActiveDownloads()
	c.JSON(http.StatusOK, gin.H{"data": downloads})
}

// GetVideoInfo downloads video information without downloading.
func GetVideoInfo(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	info, err := download.GetVideoInfo(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": info})
}

// DownloadImage downloads a single image.
func DownloadImage(c *gin.Context) {
	var req struct {
		URL  string `json:"url" binding:"required"`
		Type string `json:"type"` // "covers" or "avatars"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	downloader, err := download.NewImageDownloader()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize image downloader"})
		return
	}

	if req.Type == "" {
		req.Type = "covers"
	}

	path, err := downloader.DownloadImage(req.URL, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":    path,
		"message": "Image downloaded successfully",
	})
}

func formatSSE(event string, data interface{}) string {
	return "event: " + event + "\ndata: " + toJSON(data) + "\n\n"
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
