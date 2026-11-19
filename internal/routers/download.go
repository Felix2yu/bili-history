package routers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"bilibili-history-go/internal/models"
	"bilibili-history-go/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterDownloadRoutes(r *gin.RouterGroup) {
	download := r.Group("/download")
	{
		download.GET("/video_info", extractVideoInfo)
		download.GET("/user_videos", getUserVideos)
		download.GET("/check_video_download", checkVideoDownload)
		download.GET("/list_downloaded_videos", listDownloadedVideos)
		download.DELETE("/delete_downloaded_video", deleteDownloadedVideo)
		download.GET("/stream_video", streamVideo)
		download.GET("/stream_danmaku", streamDanmaku)
		download.GET("/check_ffmpeg", checkFFmpeg)
		download.POST("/download_video", downloadVideoSSE)
		download.POST("/batch_download", batchDownloadSSE)
		download.POST("/download_user_videos", downloadUserVideosSSE)
		download.POST("/download_favorites", downloadFavoritesSSE)
	}

	collection := r.Group("/collection")
	{
		collection.GET("/check_collection", checkCollection)
		collection.POST("/download_collection", downloadCollectionSSE)
	}
}

func extractVideoInfo(c *gin.Context) {
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
		"list":  videos,
		"total": total,
		"page":  pn,
		"size":  ps,
	}))
}

func checkVideoDownload(c *gin.Context) {
	bvidsStr := c.Query("bvids")
	if bvidsStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 bvids 参数"))
		return
	}

	var bvids []string
	for _, bvid := range strings.Split(bvidsStr, ",") {
		bvid = strings.TrimSpace(bvid)
		if bvid != "" {
			bvids = append(bvids, bvid)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(services.CheckVideoDownloaded(bvids)))
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
	bvid := c.Query("bvid")
	deleteDir := c.Query("delete_directory") == "true"
	directory := c.Query("directory")

	if err := services.DeleteDownloadedVideo(bvid, deleteDir, directory); err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "删除成功",
	}))
}

var streamContentTypes = map[string]string{
	".mp4":  "video/mp4",
	".m4v":  "video/mp4",
	".mkv":  "video/x-matroska",
	".mov":  "video/quicktime",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
	".flv":  "video/x-flv",
	".m4a":  "audio/mp4",
	".flac": "audio/flac",
	".mp3":  "audio/mpeg",
}

func streamVideo(c *gin.Context) {
	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 file_path 参数"))
		return
	}

	// 播放地址直接来自请求，必须收敛在下载目录内，否则任意本地文件都会被读走。
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("无效的文件路径"))
		return
	}
	outputDir := services.GetDownloadOutputPath()
	if !strings.HasPrefix(absPath, outputDir+string(filepath.Separator)) {
		c.JSON(http.StatusForbidden, models.ErrorResponse("无权访问该文件"))
		return
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse("文件不存在"))
		return
	}

	contentType := streamContentTypes[strings.ToLower(filepath.Ext(absPath))]
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.File(absPath)
}

// streamDanmaku 返回下载目录里的弹幕文件文本（ASS）。
// file_path 可以是视频成品路径或弹幕路径；bvid 用于只知 BV 号时按目录检索。
func streamDanmaku(c *gin.Context) {
	filePath := c.Query("file_path")
	bvid := c.Query("bvid")
	if filePath == "" && bvid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少 file_path 或 bvid 参数"))
		return
	}

	body, name, err := services.DanmakuFile(filePath, bvid)
	switch {
	case errors.Is(err, services.ErrDanmakuForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse("无权访问该文件"))
		return
	case errors.Is(err, services.ErrDanmakuNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse("弹幕文件不存在"))
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.Header("Content-Disposition", "inline; filename="+strconv.Quote(name))
	c.Data(http.StatusOK, "text/plain; charset=utf-8", body)
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

// downloadRequest 是前端发起一次下载的参数，零值即「全自动」。
type downloadRequest struct {
	URL           string                   `json:"url"`
	URLs          []string                 `json:"urls"`
	Videos        []map[string]interface{} `json:"videos"`
	UserID        string                   `json:"user_id"`
	FID           interface{}              `json:"fid"`
	Sessdata      string                   `json:"sessdata"`
	OnlyAudio     bool                     `json:"only_audio"`
	VideoQuality  int                      `json:"video_quality"`
	VideoCodec    string                   `json:"video_codec"`
	AudioQuality  int                      `json:"audio_quality"`
	Page          int                      `json:"page"`
	Batch         bool                     `json:"batch"`
	Collection    bool                     `json:"collection"`
	Danmaku       bool                     `json:"danmaku"`
	Subtitle      bool                     `json:"subtitle"`
	DownloadCover *bool                    `json:"download_cover"`
	RateLimit     string                   `json:"rate_limit"`
}

func (r *downloadRequest) options() services.DownloadOptions {
	// 封面默认跟随既有 download_cover 语义：显式传 false 才关闭。
	cover := true
	if r.DownloadCover != nil {
		cover = *r.DownloadCover
	}
	return services.DownloadOptions{
		Sessdata:     r.Sessdata,
		OnlyAudio:    r.OnlyAudio,
		VideoQuality: r.VideoQuality,
		VideoCodec:   r.VideoCodec,
		AudioQuality: r.AudioQuality,
		Page:         r.Page,
		Batch:        r.Batch,
		Collection:   r.Collection,
		Danmaku:      r.Danmaku,
		Subtitle:     r.Subtitle,
		Cover:        cover,
		RateLimit:    r.RateLimit,
	}
}

// favoriteID 容忍 fid 以 JSON 数字或字符串两种形态到达。
func (r *downloadRequest) favoriteID() string {
	switch v := r.FID.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	default:
		return ""
	}
}

func bvLink(bvid string) string {
	if strings.Contains(bvid, "://") {
		return bvid
	}
	return "https://www.bilibili.com/video/" + bvid
}

func (r *downloadRequest) links() []string {
	var links []string
	if r.URL != "" {
		links = append(links, r.URL)
	}
	for _, link := range r.URLs {
		if link != "" {
			links = append(links, bvLink(link))
		}
	}
	for _, video := range r.Videos {
		if bvid, _ := video["bvid"].(string); bvid != "" {
			links = append(links, bvLink(bvid))
		}
	}
	return links
}

// sseEmitter 把下载事件写成 SSE 帧；客户端断开时取消底层 context，
// 让 bili-dl 的请求与 ffmpeg 子进程一起收尾。
type sseEmitter struct {
	mu     sync.Mutex
	w      http.ResponseWriter
	flush  http.Flusher
	ctx    context.Context
	cancel context.CancelFunc
	prefix string
}

func newSSEEmitter(c *gin.Context, prefix string) (*sseEmitter, bool) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil, false
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx, cancel := context.WithCancel(c.Request.Context())
	return &sseEmitter{w: c.Writer, flush: flusher, ctx: ctx, cancel: cancel, prefix: prefix}, true
}

func (e *sseEmitter) Event(ev services.DownloadEvent) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.prefix != "" && ev.Type == "log" {
		ev.Text = e.prefix + ev.Text
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return
	}
	if _, err := fmt.Fprintf(e.w, "data: %s\n\n", body); err != nil {
		e.cancel()
		return
	}
	e.flush.Flush()
}

func (e *sseEmitter) Finish(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	done := services.DownloadEvent{Type: "done", Event: "ok"}
	switch {
	case err == nil:
		done.Text = "下载完成"
	case errors.Is(err, context.Canceled):
		done.Event = "cancel"
		done.Text = "下载已取消"
	default:
		done.Event = "error"
		done.Text = err.Error()
	}
	body, _ := json.Marshal(done)
	fmt.Fprintf(e.w, "data: %s\n\n", body)
	e.flush.Flush()
}

// runDownloadSSE 是所有下载入口共用的一条链路：绑定参数、开流、执行、收尾。
func runDownloadSSE(c *gin.Context, req *downloadRequest, links []string, prefix string) {
	if len(links) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("没有可下载的链接"))
		return
	}

	emitter, ok := newSSEEmitter(c, prefix)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("不支持流式传输"))
		return
	}

	err := services.Download(emitter.ctx, links, services.GetDownloadOutputPath(), req.options(), emitter.Event)
	emitter.Finish(err)
}

func downloadVideoSSE(c *gin.Context) {
	var req downloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	if req.URL == "" && len(req.URLs) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse("缺少 url 参数"))
		return
	}
	runDownloadSSE(c, &req, req.links(), "")
}

func batchDownloadSSE(c *gin.Context) {
	var req downloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	runDownloadSSE(c, &req, req.links(), "")
}

func downloadUserVideosSSE(c *gin.Context) {
	var req downloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	if req.UserID == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("缺少 user_id 参数"))
		return
	}

	emitter, ok := newSSEEmitter(c, "")
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("不支持流式传输"))
		return
	}
	emitter.Event(services.DownloadEvent{Type: "log", Text: "正在获取用户视频列表..."})

	videos, total, err := services.GetUserVideos(req.UserID, 1, 50)
	if err != nil {
		emitter.Finish(fmt.Errorf("获取视频列表失败: %w", err))
		return
	}
	emitter.Event(services.DownloadEvent{Type: "log", Text: fmt.Sprintf("共找到 %d 个视频", total)})

	links := make([]string, 0, len(videos))
	for _, video := range videos {
		if bvid := video["bvid"]; bvid != "" {
			links = append(links, bvLink(bvid))
		}
	}
	err = services.Download(emitter.ctx, links, services.GetDownloadOutputPath(), req.options(), emitter.Event)
	emitter.Finish(err)
}

// downloadFavoritesSSE 下载整个收藏夹：本地收藏库优先，未同步时在线拉取。
func downloadFavoritesSSE(c *gin.Context) {
	var req downloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	fid := req.favoriteID()
	if fid == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("缺少 fid 参数"))
		return
	}

	emitter, ok := newSSEEmitter(c, "")
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("不支持流式传输"))
		return
	}
	emitter.Event(services.DownloadEvent{Type: "log", Text: "正在获取收藏夹视频列表..."})

	links, err := services.FavoriteLinks(fid)
	if err != nil {
		emitter.Finish(err)
		return
	}
	emitter.Event(services.DownloadEvent{Type: "log", Text: fmt.Sprintf("共找到 %d 个收藏视频", len(links))})

	err = services.Download(emitter.ctx, links, services.GetDownloadOutputPath(), req.options(), emitter.Event)
	emitter.Finish(err)
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
	var req downloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	if req.URL == "" {
		c.JSON(http.StatusOK, models.ErrorResponse("缺少 url 参数"))
		return
	}
	// 合集入口的语义就是整集展开，缺省即 collection+batch。
	if !req.Collection {
		req.Collection = true
		req.Batch = true
	}
	runDownloadSSE(c, &req, []string{req.URL}, "")
}
