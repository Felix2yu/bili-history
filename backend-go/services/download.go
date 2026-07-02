package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/utils"

	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"

	_ "github.com/iawia002/lux/extractors/bilibili"
)

type VideoInfo struct {
	URL        string       `json:"url"`
	Site       string       `json:"site"`
	Title      string       `json:"title"`
	Type       string       `json:"type"`
	Streams    []StreamInfo `json:"streams"`
	Cover      string       `json:"cover"`
	DanmakuURL string       `json:"danmaku_url,omitempty"`
}

type StreamInfo struct {
	ID      string `json:"id"`
	Quality string `json:"quality"`
	Size    int64  `json:"size"`
	Ext     string `json:"ext"`
}

type DownloadedVideo struct {
	Title      string `json:"title"`
	FilePath   string `json:"file_path"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	DownloadAt int64  `json:"download_at"`
}

func GetDownloadOutputPath() string {
	return utils.GetOutputPath("downloads")
}

func getCookie() string {
	cfg := config.GetConfig()
	if cfg == nil {
		return ""
	}
	var parts []string
	if cfg.SESSDATA != "" {
		parts = append(parts, "SESSDATA="+cfg.SESSDATA)
	}
	if cfg.BiliJct != "" {
		parts = append(parts, "bili_jct="+cfg.BiliJct)
	}
	if cfg.DedeUserID != "" {
		parts = append(parts, "DedeUserID="+cfg.DedeUserID)
	}
	return strings.Join(parts, "; ")
}

func initLuxRequest(cookie string) {
	request.SetOptions(request.Options{
		RetryTimes: 10,
		Cookie:     cookie,
	})
}

// codecPriority returns sort priority for bilibili codec IDs
// av1(13) > hevc(12) > avc(7)
func codecPriority(codecid int) int {
	switch codecid {
	case 13: // AV1
		return 3
	case 12: // HEVC
		return 2
	case 7: // AVC
		return 1
	default:
		return 0
	}
}

func ExtractVideoInfo(url string) (*VideoInfo, error) {
	cookie := getCookie()
	initLuxRequest(cookie)

	dataList, err := extractors.Extract(url, extractors.Options{
		Playlist: false,
		Cookie:   cookie,
	})
	if err != nil {
		return nil, fmt.Errorf("提取视频信息失败: %w", err)
	}

	if len(dataList) == 0 {
		return nil, fmt.Errorf("未找到视频信息")
	}

	data := dataList[0]
	if data.Err != nil {
		return nil, fmt.Errorf("提取视频信息出错: %w", data.Err)
	}

	var streams []StreamInfo
	for id, stream := range data.Streams {
		streams = append(streams, StreamInfo{
			ID:      id,
			Quality: stream.Quality,
			Size:    stream.Size,
			Ext:     stream.Ext,
		})
	}
	// Sort streams: higher quality first, then by codec priority (av1 > hevc > avc)
	sort.Slice(streams, func(i, j int) bool {
		qi, _ := strconv.Atoi(strings.Split(streams[i].ID, "-")[0])
		qj, _ := strconv.Atoi(strings.Split(streams[j].ID, "-")[0])
		if qi != qj {
			return qi > qj
		}
		// Codec priority: av1(13) > hevc(12) > avc(7)
		ci, _ := strconv.Atoi(strings.Split(streams[i].ID, "-")[1])
		cj, _ := strconv.Atoi(strings.Split(streams[j].ID, "-")[1])
		return codecPriority(ci) > codecPriority(cj)
	})

	info := &VideoInfo{
		URL:     data.URL,
		Site:    data.Site,
		Title:   data.Title,
		Type:    string(data.Type),
		Streams: streams,
	}

	if captions, ok := data.Captions["danmaku"]; ok {
		info.DanmakuURL = captions.URL
	}

	return info, nil
}

func DownloadVideoWithProgress(url, sessdata, outputDir string, onlyAudio bool, streamID string, onProgress func(string)) error {
	cookie := sessdata
	if cookie == "" {
		cookie = getCookie()
	}
	initLuxRequest(cookie)

	onProgress("正在提取视频信息...")

	dataList, err := extractors.Extract(url, extractors.Options{
		Playlist: false,
		Cookie:   cookie,
	})
	if err != nil {
		return fmt.Errorf("提取视频信息失败: %w", err)
	}

	if len(dataList) == 0 {
		return fmt.Errorf("未找到视频信息")
	}

	data := dataList[0]
	if data.Err != nil {
		return fmt.Errorf("提取视频出错: %w", data.Err)
	}

	onProgress(fmt.Sprintf("标题: %s", data.Title))

	// Auto-select best stream:
	// 1. Find the HIGHEST quality level available
	// 2. Among streams at that quality, prefer av1 > hevc > avc
	// This ensures we always get the clearest resolution first,
	// then pick the most efficient codec within that resolution.
	if streamID == "" {
		var bestID string
		var highestQ int
		for id := range data.Streams {
			parts := strings.Split(id, "-")
			if len(parts) != 2 {
				continue
			}
			q, _ := strconv.Atoi(parts[0])
			if q > highestQ {
				highestQ = q
			}
		}
		// Among streams at the highest quality, pick best codec
		var bestC int
		for id := range data.Streams {
			parts := strings.Split(id, "-")
			if len(parts) != 2 {
				continue
			}
			q, _ := strconv.Atoi(parts[0])
			c, _ := strconv.Atoi(parts[1])
			if q == highestQ && codecPriority(c) > bestC {
				bestC = codecPriority(c)
				bestID = id
			}
		}
		streamID = bestID
	}

	if streamID != "" {
		if stream, ok := data.Streams[streamID]; ok {
			sizeMB := float64(stream.Size) / 1024 / 1024
			onProgress(fmt.Sprintf("画质: %s, 大小: %.1fMB", stream.Quality, sizeMB))
		}
	}

	os.MkdirAll(outputDir, 0755)

	dl := downloader.New(downloader.Options{
		Silent:         true,
		Stream:         streamID,
		OutputPath:     outputDir,
		FileNameLength: 255,
		MultiThread:    true,
		ThreadNumber:   10,
		RetryTimes:     10,
		ChunkSizeMB:    1,
		AudioOnly:      onlyAudio,
	})

	onProgress("开始下载...")
	if err := dl.Download(data); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// Remux to target container based on codec:
	// av1(13) → mkv, hevc(12) → mov, avc(7) → mp4
	codecid := extractCodecid(streamID)
	targetExt := containerForCodec(codecid)
	if targetExt != "" {
		downloadedFile := findDownloadedFile(outputDir, data.Title, targetExt)
		if downloadedFile != "" {
			onProgress(fmt.Sprintf("正在转封装为 %s ...", targetExt))
			if err := remuxFile(downloadedFile, targetExt); err != nil {
				onProgress(fmt.Sprintf("转封装失败: %v，保留原文件", err))
			}
		}
	}

	onProgress("下载完成")
	return nil
}

func extractCodecid(streamID string) int {
	parts := strings.Split(streamID, "-")
	if len(parts) == 2 {
		c, _ := strconv.Atoi(parts[1])
		return c
	}
	return 0
}

// containerForCodec returns the target container extension for a codec.
// Empty string means no remux needed (already correct).
func containerForCodec(codecid int) string {
	switch codecid {
	case 13: // AV1 → mkv
		return "mkv"
	case 12: // HEVC → mov
		return "mov"
	case 7: // AVC → mp4 (lux already outputs mp4)
		return ""
	default:
		return ""
	}
}

func findDownloadedFile(outputDir, title, targetExt string) string {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		if strings.Contains(base, title) || strings.Contains(title, base) {
			// Skip if already target format
			if ext == "."+targetExt {
				return ""
			}
			if ext == ".mp4" || ext == ".webm" || ext == ".flv" {
				return filepath.Join(outputDir, name)
			}
		}
	}
	return ""
}

func remuxFile(inputPath, targetExt string) error {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "." + targetExt
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-c", "copy", "-y", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg remux failed: %s: %w", string(output), err)
	}
	os.Remove(inputPath)
	return nil
}

func GetUserVideos(mid string, pn, ps int) ([]map[string]string, int, error) {
	cookie := getCookie()

	url := fmt.Sprintf("https://api.bilibili.com/x/space/wbi/arc/search?mid=%s&pn=%d&ps=%d&order=pubdate", mid, pn, ps)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			List struct {
				Vlist []struct {
					BVID  string `json:"bvid"`
					Title string `json:"title"`
					Pic   string `json:"pic"`
					AID   int64  `json:"aid"`
					Length string `json:"length"`
				} `json:"vlist"`
			} `json:"list"`
			Page struct {
				Count int `json:"count"`
			} `json:"page"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	if result.Code != 0 {
		return nil, 0, fmt.Errorf("API 返回错误码: %d", result.Code)
	}

	var videos []map[string]string
	for _, v := range result.Data.List.Vlist {
		videos = append(videos, map[string]string{
			"bvid":  v.BVID,
			"title": v.Title,
			"pic":   v.Pic,
			"aid":   strconv.FormatInt(v.AID, 10),
			"length": v.Length,
		})
	}

	return videos, result.Data.Page.Count, nil
}

func CheckVideoDownloaded(cids []string) map[string]bool {
	outputDir := GetDownloadOutputPath()
	result := make(map[string]bool)

	for _, cid := range cids {
		result[cid] = false
	}

	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		for _, cid := range cids {
			if strings.Contains(name, cid) {
				result[cid] = true
			}
		}
		return nil
	})

	return result
}

func ListDownloadedVideos(search string, page, limit int) ([]DownloadedVideo, int, error) {
	outputDir := GetDownloadOutputPath()

	var videos []DownloadedVideo

	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mp4" && ext != ".flv" && ext != ".mkv" && ext != ".webm" && ext != ".avi" {
			return nil
		}

		name := info.Name()
		if search != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(search)) {
			return nil
		}

		videos = append(videos, DownloadedVideo{
			Title:      strings.TrimSuffix(name, filepath.Ext(name)),
			FilePath:   path,
			FileName:   name,
			FileSize:   info.Size(),
			DownloadAt: info.ModTime().Unix(),
		})
		return nil
	})

	total := len(videos)

	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return videos[start:end], total, nil
}

func DeleteDownloadedVideo(cid string, deleteDir bool, directory string) error {
	outputDir := GetDownloadOutputPath()

	if directory != "" {
		absDir, err := filepath.Abs(directory)
		if err != nil || !strings.HasPrefix(absDir, outputDir) {
			return fmt.Errorf("无效的目录路径")
		}
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			return fmt.Errorf("目录不存在")
		}
		if deleteDir {
			return os.RemoveAll(absDir)
		}
		return nil
	}

	if cid == "" {
		return fmt.Errorf("缺少 cid 参数")
	}

	// Find and delete files matching CID
	var found bool
	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		if strings.Contains(name, cid) {
			found = true
			if deleteDir {
				dir := filepath.Dir(path)
				return os.RemoveAll(dir)
			}
			return os.Remove(path)
		}
		return nil
	})

	if !found {
		return fmt.Errorf("未找到匹配的视频文件")
	}

	return nil
}

func CheckCollection(url string) (map[string]interface{}, error) {
	cookie := getCookie()
	initLuxRequest(cookie)

	dataList, err := extractors.Extract(url, extractors.Options{
		Playlist: true,
		Cookie:   cookie,
	})
	if err != nil {
		return nil, fmt.Errorf("检查合集失败: %w", err)
	}

	isCollection := len(dataList) > 1
	return map[string]interface{}{
		"is_collection": isCollection,
		"video_count":   len(dataList),
		"videos": func() []map[string]string {
			var videos []map[string]string
			for _, d := range dataList {
				if d.Err == nil {
					videos = append(videos, map[string]string{
						"title": d.Title,
						"url":   d.URL,
					})
				}
			}
			return videos
		}(),
	}, nil
}

func DownloadCollectionWithProgress(url, sessdata, outputDir string, onlyAudio bool, onProgress func(string)) error {
	cookie := sessdata
	if cookie == "" {
		cookie = getCookie()
	}
	initLuxRequest(cookie)

	onProgress("正在获取合集信息...")

	dataList, err := extractors.Extract(url, extractors.Options{
		Playlist: true,
		Cookie:   cookie,
	})
	if err != nil {
		return fmt.Errorf("获取合集信息失败: %w", err)
	}

	total := len(dataList)
	onProgress(fmt.Sprintf("共找到 %d 个视频", total))

	os.MkdirAll(outputDir, 0755)

	dl := downloader.New(downloader.Options{
		Silent:         true,
		OutputPath:     outputDir,
		FileNameLength: 255,
		MultiThread:    true,
		ThreadNumber:   10,
		RetryTimes:     10,
		ChunkSizeMB:    1,
		AudioOnly:      onlyAudio,
	})

	for i, data := range dataList {
		if data.Err != nil {
			onProgress(fmt.Sprintf("[%d/%d] 跳过: %v", i+1, total, data.Err))
			continue
		}
		onProgress(fmt.Sprintf("[%d/%d] 下载: %s", i+1, total, data.Title))
		if err := dl.Download(data); err != nil {
			onProgress(fmt.Sprintf("[%d/%d] 失败: %v", i+1, total, err))
		}
	}

	onProgress("合集下载完成")
	return nil
}
