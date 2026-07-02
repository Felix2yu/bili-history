package routers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"bilibili-history-go/models"
	"bilibili-history-go/services"

	"github.com/gin-gonic/gin"
)

func RegisterDownloadRoutes(r *gin.RouterGroup) {
	download := r.Group("/download")
	{
		download.GET("/video_info", getVideoInfo)
		download.GET("/user_videos", getUserVideos)
		download.GET("/check_video_download", checkVideoDownload)
		download.GET("/list_downloaded_videos", listDownloadedVideos)
		download.DELETE("/delete_downloaded_video", deleteDownloadedVideo)
		download.GET("/stream_video", streamVideo)
		download.GET("/check_ffmpeg", checkFFmpeg)
		download.POST("/download_video", downloadVideoSSE)
		download.POST("/batch_download", batchDownloadSSE)
		download.POST("/download_user_videos", downloadUserVideosSSE)
	}

	collection := r.Group("/collection")
	{
		collection.GET("/check_collection", checkCollection)
		collection.POST("/download_collection", downloadCollectionSSE)
	}
}

func getVideoInfo(c *gin.Context) {
	url := c.Query("url")
	bvid := c.Query("bvid")
	if url == "" && bvid != "" {
		url = "https://www.bilibili.com/video/" + bvid
	}
	if url == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 url 或 bvid 参数"))
		return
	}

	info, err := services.ExtractVideoInfo(url)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(info))
}

func getUserVideos(c *gin.Context) {
	mid := c.Query("mid")
	if mid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 mid 参数"))
		return
	}

	pn, _ := strconv.Atoi(c.DefaultQuery("pn", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("ps", "30"))

	videos, total, err := services.GetUserVideos(mid, pn, ps)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"list":   videos,
		"total":  total,
		"page":   pn,
		"size":   ps,
	}))
}

func checkVideoDownload(c *gin.Context) {
	cidsStr := c.Query("cids")
	if cidsStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 cids 参数"))
		return
	}

	var cids []string
	for _, cid := range strings.Split(cidsStr, ",") {
		cid = strings.TrimSpace(cid)
		if cid != "" {
			cids = append(cids, cid)
		}
	}

	result := services.CheckVideoDownloaded(cids)
	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

func listDownloadedVideos(c *gin.Context) {
	search := c.Query("search_term")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	videos, total, err := services.ListDownloadedVideos(search, page, limit)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"videos": videos,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}))
}

func deleteDownloadedVideo(c *gin.Context) {
	cid := c.Query("cid")
	deleteDir := c.Query("delete_directory") == "true"
	directory := c.Query("directory")

	err := services.DeleteDownloadedVideo(cid, deleteDir, directory)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "删除成功",
	}))
}

func streamVideo(c *gin.Context) {
	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 file_path 参数"))
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse("文件不存在"))
		return
	}

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")
	c.File(filePath)
}

func checkFFmpeg(c *gin.Context) {
	_, err := exec.LookPath("ffmpeg")
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"installed": err == nil,
		"version":   getFFmpegVersion(),
	}))
}

func getFFmpegVersion() string {
	out, err := exec.Command("ffmpeg", "-version").Output()
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func downloadVideoSSE(c *gin.Context) {
	var req struct {
		URL           string `json:"url"`
		Sessdata      string `json:"sessdata"`
		DownloadCover bool   `json:"download_cover"`
		OnlyAudio     bool   `json:"only_audio"`
		CID           int    `json:"cid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("不支持流式传输"))
		return
	}

	sendEvent := func(msg string) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
		flusher.Flush()
	}

	sendEvent("开始下载...")

	outputDir := services.GetDownloadOutputPath()
	err := services.DownloadVideoWithProgress(req.URL, req.Sessdata, outputDir, req.OnlyAudio, sendEvent)
	if err != nil {
		sendEvent(fmt.Sprintf("下载失败: %v", err))
	}

	sendEvent("close")
}

func batchDownloadSSE(c *gin.Context) {
	var req struct {
		Videos       []map[string]interface{} `json:"videos"`
		Sessdata     string                   `json:"sessdata"`
		DownloadCover bool                    `json:"download_cover"`
		OnlyAudio    bool                     `json:"only_audio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("不支持流式传输"))
		return
	}

	sendEvent := func(msg string) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
		flusher.Flush()
	}

	outputDir := services.GetDownloadOutputPath()
	total := len(req.Videos)
	for i, video := range req.Videos {
		bvid, _ := video["bvid"].(string)
		if bvid == "" {
			continue
		}
		url := "https://www.bilibili.com/video/" + bvid
		sendEvent(fmt.Sprintf("[%d/%d] 开始下载: %s", i+1, total, bvid))

		err := services.DownloadVideoWithProgress(url, req.Sessdata, outputDir, req.OnlyAudio, func(msg string) {
			sendEvent(fmt.Sprintf("[%d/%d] %s", i+1, total, msg))
		})
		if err != nil {
			sendEvent(fmt.Sprintf("[%d/%d] 下载失败: %v", i+1, total, err))
		}
	}

	sendEvent("close")
}

func downloadUserVideosSSE(c *gin.Context) {
	var req struct {
		UserID       string `json:"user_id"`
		Sessdata     string `json:"sessdata"`
		DownloadCover bool  `json:"download_cover"`
		OnlyAudio    bool   `json:"only_audio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("不支持流式传输"))
		return
	}

	sendEvent := func(msg string) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
		flusher.Flush()
	}

	sendEvent("正在获取用户视频列表...")

	videos, total, err := services.GetUserVideos(req.UserID, 1, 50)
	if err != nil {
		sendEvent(fmt.Sprintf("获取视频列表失败: %v", err))
		sendEvent("close")
		return
	}

	sendEvent(fmt.Sprintf("共找到 %d 个视频", total))

	outputDir := services.GetDownloadOutputPath()
	for i, video := range videos {
		bvid := video["bvid"]
		if bvid == "" {
			continue
		}
		url := "https://www.bilibili.com/video/" + bvid
		title := video["title"]
		sendEvent(fmt.Sprintf("[%d/%d] 开始下载: %s", i+1, len(videos), title))

		err := services.DownloadVideoWithProgress(url, req.Sessdata, outputDir, req.OnlyAudio, func(msg string) {
			sendEvent(fmt.Sprintf("[%d/%d] %s", i+1, len(videos), msg))
		})
		if err != nil {
			sendEvent(fmt.Sprintf("[%d/%d] 下载失败: %v", i+1, len(videos), err))
		}
	}

	sendEvent("close")
}

func checkCollection(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 url 参数"))
		return
	}

	result, err := services.CheckCollection(url)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

func downloadCollectionSSE(c *gin.Context) {
	var req struct {
		URL          string `json:"url"`
		CID          int    `json:"cid"`
		Sessdata     string `json:"sessdata"`
		DownloadCover bool  `json:"download_cover"`
		OnlyAudio    bool   `json:"only_audio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("不支持流式传输"))
		return
	}

	sendEvent := func(msg string) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
		flusher.Flush()
	}

	sendEvent("正在获取合集信息...")

	outputDir := services.GetDownloadOutputPath()
	err := services.DownloadCollectionWithProgress(req.URL, req.Sessdata, outputDir, req.OnlyAudio, sendEvent)
	if err != nil {
		sendEvent(fmt.Sprintf("下载失败: %v", err))
	}

	sendEvent("close")
}
