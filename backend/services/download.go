package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/utils"

	"github.com/Felix2yu/bili-dl/C"
	"github.com/Felix2yu/bili-dl/api"
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

type DownloadedVideoFile struct {
	FilePath    string  `json:"file_path"`
	SizeMB      float64 `json:"size_mb"`
	IsAudioOnly bool    `json:"is_audio_only"`
}

type DownloadedVideo struct {
	Title       string              `json:"title"`
	BVID        string              `json:"bvid"`
	CID         string              `json:"cid"`
	Cover       string              `json:"cover"`
	AuthorName  string              `json:"author_name"`
	AuthorFace  string              `json:"author_face"`
	AuthorMid   int64               `json:"author_mid"`
	DownloadDate string             `json:"download_date"`
	Files       []DownloadedVideoFile `json:"files"`
	Directory   string              `json:"directory"`
}

func GetDownloadOutputPath() string {
	return utils.GetOutputPath("downloads")
}

func getCookie() string {
	cfg := config.GetConfig()
	if cfg == nil {
		return ""
	}
	if cfg.SESSDATA != "" {
		return cfg.SESSDATA
	}
	return ""
}

var bvRegexp = regexp.MustCompile(`BV[a-zA-Z0-9]+`)

func extractBVFromURL(url string) string {
	if bv := bvRegexp.FindString(url); bv != "" {
		return bv
	}
	return ""
}

// codecPriority returns sort priority for bilibili codec strings
// av1 > hevc > avc
func codecPriorityStr(codec string) int {
	if strings.HasPrefix(codec, "av01") {
		return 3
	} else if strings.HasPrefix(codec, "hev") || strings.HasPrefix(codec, "hvc") {
		return 2
	} else if strings.HasPrefix(codec, "avc") {
		return 1
	}
	return 0
}

// friendlyCodecName returns a short human-readable codec name
func friendlyCodecName(codec string) string {
	if strings.HasPrefix(codec, "av01") {
		return "AV1"
	} else if strings.HasPrefix(codec, "hev") || strings.HasPrefix(codec, "hvc") {
		return "HEVC"
	} else if strings.HasPrefix(codec, "avc") {
		return "AVC"
	}
	return codec
}

// friendlyQualityLabel returns a readable quality label like "4K · HEVC · MOV · 3.0Mbps" or "480p · AVC · MP4 · 320Kbps"
func friendlyQualityLabel(codec string, width, height int, bandwidth float64) string {
	res := resolutionLabel(width, height)
	codecName := friendlyCodecName(codec)
	ext := containerForCodec(codec)
	bitrate := bandwidth / 1000
	if bitrate >= 1000 {
		return fmt.Sprintf("%s · %s · %s · %.1fMbps", res, codecName, ext, bitrate/1000)
	}
	return fmt.Sprintf("%s · %s · %s · %dKbps", res, codecName, ext, int(bitrate))
}

// containerForCodec returns the target container extension for a codec
func containerForCodec(codec string) string {
	if strings.HasPrefix(codec, "av01") {
		return "MKV"
	} else if strings.HasPrefix(codec, "hev") || strings.HasPrefix(codec, "hvc") {
		return "MOV"
	}
	return "MP4"
}

// resolutionLabel returns standard quality name from resolution
func resolutionLabel(width, height int) string {
	h := height
	if width > height {
		h = width
	}
	switch {
	case h >= 2160:
		return "4K"
	case h >= 1440:
		return "2K"
	case h >= 1080:
		return "1080p"
	case h >= 720:
		return "720p"
	case h >= 480:
		return "480p"
	case h >= 360:
		return "360p"
	default:
		return fmt.Sprintf("%dp", h)
	}
}

// dashVideoStream represents a single video stream from DASH response
type dashVideoStream struct {
	ID        float64 `json:"id"`
	Codecs    string  `json:"codecs"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Bandwidth float64 `json:"bandwidth"`
	BaseURL   string  `json:"base_url"`
}

// dashAudioStream represents a single audio stream from DASH response
type dashAudioStream struct {
	ID        float64 `json:"id"`
	Codecs    string  `json:"codecs"`
	Bandwidth float64 `json:"bandwidth"`
	BaseURL   string  `json:"base_url"`
}

// videoDetail holds extra info from bilibili view API
type videoDetail struct {
	Title string `json:"title"`
	Owner struct {
		Name string `json:"name"`
	} `json:"owner"`
}

// fetchVideoDetail fetches video detail including author name
func fetchVideoDetail(bv, cookie string) (*videoDetail, error) {
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bv)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "SESSDATA", Value: cookie})
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int         `json:"code"`
		Data videoDetail `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("API 返回错误码: %d", result.Code)
	}
	return &result.Data, nil
}

// fetchDashStreams fetches DASH stream info from bilibili API
func fetchDashStreams(bv, cid, cookie string) ([]dashVideoStream, []dashAudioStream, error) {
	apiURL := "https://api.bilibili.com/x/player/wbi/playurl?fnver=0&fnval=3216&fourk=1&qn=127"
	parse, _ := url.Parse(apiURL)
	query := parse.Query()
	query.Add("bvid", bv)
	query.Add("cid", cid)
	parse.RawQuery = query.Encode()
	apiURL = parse.String()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "SESSDATA", Value: cookie})
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Dash struct {
				Video []dashVideoStream `json:"video"`
				Audio []dashAudioStream `json:"audio"`
			} `json:"dash"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, err
	}

	if result.Code != 0 {
		return nil, nil, fmt.Errorf("API 返回错误码: %d", result.Code)
	}

	return result.Data.Dash.Video, result.Data.Dash.Audio, nil
}

func ExtractVideoInfo(url string) (*VideoInfo, error) {
	bv := extractBVFromURL(url)
	if bv == "" {
		return nil, fmt.Errorf("无法从 URL 提取 BV 号: %s", url)
	}

	cookie := getCookie()

	video, err := api.VideoFromBV(bv)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}

	_, err = api.ResolveVideo(video)
	if err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %w", err)
	}

	videoStreams, _, err := fetchDashStreams(video.BV, video.Cid, cookie)
	if err != nil {
		return nil, fmt.Errorf("获取视频流信息失败: %w", err)
	}

	var streams []StreamInfo
	for _, vs := range videoStreams {
		codec := vs.Codecs
		qualityID := int(vs.ID)
		streams = append(streams, StreamInfo{
			ID:      fmt.Sprintf("%d-%s", qualityID, codec),
			Quality: friendlyQualityLabel(codec, vs.Width, vs.Height, vs.Bandwidth),
			Size:    int64(vs.Bandwidth),
			Ext:     "mp4",
		})
	}

	sort.Slice(streams, func(i, j int) bool {
		qi, _ := strconv.Atoi(strings.SplitN(streams[i].ID, "-", 2)[0])
		qj, _ := strconv.Atoi(strings.SplitN(streams[j].ID, "-", 2)[0])
		if qi != qj {
			return qi > qj
		}
		ci := codecPriorityStr(strings.SplitN(streams[i].ID, "-", 2)[1])
		cj := codecPriorityStr(strings.SplitN(streams[j].ID, "-", 2)[1])
		return ci > cj
	})

	info := &VideoInfo{
		URL:     url,
		Site:    "bilibili",
		Title:   video.Title,
		Type:    "video",
		Streams: streams,
	}

	return info, nil
}

func DownloadVideoWithProgress(url, sessdata, outputDir string, onlyAudio bool, streamID string, onProgress func(string)) error {
	bv := extractBVFromURL(url)
	if bv == "" {
		return fmt.Errorf("无法从 URL 提取 BV 号: %s", url)
	}

	cookie := sessdata
	if cookie == "" {
		cookie = getCookie()
	}
	C.Cookie = cookie

	onProgress("正在提取视频信息...")

	video, err := api.VideoFromBV(bv)
	if err != nil {
		return fmt.Errorf("获取视频信息失败: %w", err)
	}

	_, err = api.ResolveVideo(video)
	if err != nil {
		return fmt.Errorf("解析视频信息失败: %w", err)
	}

	// 获取UP主名字，构建 [UP主名]投稿名 格式的文件名
	detail, err := fetchVideoDetail(bv, cookie)
	if err == nil && detail.Owner.Name != "" {
		video.Title = fmt.Sprintf("[%s]%s", detail.Owner.Name, video.Title)
	}

	onProgress(fmt.Sprintf("标题: %s", video.Title))

	videoStreams, audioStreams, err := fetchDashStreams(video.BV, video.Cid, cookie)
	if err != nil {
		return fmt.Errorf("获取视频流信息失败: %w", err)
	}

	if len(videoStreams) == 0 {
		return fmt.Errorf("未找到可用的视频流")
	}
	if len(audioStreams) == 0 {
		return fmt.Errorf("未找到可用的音频流")
	}

	// Select video stream based on streamID or auto-select best
	var selectedVideo *dashVideoStream
	if streamID != "" {
		// Parse streamID to find matching stream
		parts := strings.SplitN(streamID, "-", 2)
		if len(parts) == 2 {
			targetID, _ := strconv.ParseFloat(parts[0], 64)
			targetCodec := parts[1]
			for i, vs := range videoStreams {
				if vs.ID == targetID && vs.Codecs == targetCodec {
					selectedVideo = &videoStreams[i]
					break
				}
			}
		}
	}

	if selectedVideo == nil {
		// Auto-select: best codec, then highest quality
		bestIdx := 0
		for i := 1; i < len(videoStreams); i++ {
			ci, cj := codecPriorityStr(videoStreams[i].Codecs), codecPriorityStr(videoStreams[bestIdx].Codecs)
			if ci > cj || (ci == cj && videoStreams[i].ID > videoStreams[bestIdx].ID) {
				bestIdx = i
			}
		}
		selectedVideo = &videoStreams[bestIdx]
	}

	// Select best audio stream (highest bandwidth)
	bestAudioIdx := 0
	for i := 1; i < len(audioStreams); i++ {
		if audioStreams[i].Bandwidth > audioStreams[bestAudioIdx].Bandwidth {
			bestAudioIdx = i
		}
	}
	selectedAudio := &audioStreams[bestAudioIdx]

	sizeMB := float64(selectedVideo.Bandwidth) / 1024 / 1024
	onProgress(fmt.Sprintf("画质: %s · %s, 大小: %.1fMB", resolutionLabel(int(selectedVideo.Width), int(selectedVideo.Height)), friendlyCodecName(selectedVideo.Codecs), sizeMB))

	os.MkdirAll(outputDir, 0755)

	// Set bili-dl global config
	C.O = outputDir
	C.Merge = true
	C.Delete = true
	C.AddBVSuffix = false

	stream := &api.Stream{
		V:     selectedVideo.BaseURL,
		A:     selectedAudio.BaseURL,
		Video: api.Video{Title: video.Title, BV: video.BV, Cid: video.Cid},
	}

	onProgress("开始下载视频...")
	if err := api.Dl(stream); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	if !onlyAudio {
		onProgress("正在合并音视频...")
		if err := api.Merge(stream); err != nil {
			onProgress(fmt.Sprintf("合并失败: %v，保留单独文件", err))
		}

		// Remux to target container based on codec: AV1→MKV, HEVC→MOV, AVC→MP4
		targetExt := containerForCodec(selectedVideo.Codecs)
		if targetExt != "MP4" {
			onProgress(fmt.Sprintf("正在转封装为 %s ...", targetExt))
			if err := remuxToContainer(outputDir, video.Title, video.BV, targetExt); err != nil {
				onProgress(fmt.Sprintf("转封装失败: %v，保留 MP4 文件", err))
			}
		}
	}

	onProgress("下载完成")
	return nil
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

		// 解析文件名，提取元数据
		// 文件名格式：[UP主名]投稿名.mp4 或 投稿名-BVxxxxxx.mp4
		fullName := strings.TrimSuffix(name, filepath.Ext(name))
		authorName := ""
		title := fullName

		// 尝试匹配 [UP主名]标题 格式
		if strings.HasPrefix(fullName, "[") {
			endIdx := strings.Index(fullName, "]")
			if endIdx > 0 {
				authorName = fullName[1:endIdx]
				title = fullName[endIdx+1:]
			}
		}

		// 尝试从文件名中提取 BV 号
		bvid := ""
		bvMatch := bvRegexp.FindString(fullName)
		if bvMatch != "" {
			bvid = bvMatch
		}

		// 检查是否为纯音频
		isAudioOnly := strings.Contains(strings.ToLower(name), "audio") || strings.HasSuffix(strings.ToLower(name), ".m4a")

		videos = append(videos, DownloadedVideo{
			Title:        title,
			BVID:         bvid,
			AuthorName:   authorName,
			DownloadDate: info.ModTime().Format("2006-01-02"),
			Files: []DownloadedVideoFile{
				{
					FilePath:    path,
					SizeMB:      float64(info.Size()) / 1024 / 1024,
					IsAudioOnly: isAudioOnly,
				},
			},
			Directory: filepath.Dir(path),
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
	// Extract BV from URL and check if it's a collection page
	// Bilibili collections use the space/collection API
	bv := extractBVFromURL(url)
	if bv == "" {
		return nil, fmt.Errorf("无法从 URL 提取 BV 号")
	}

	cookie := getCookie()
	C.Cookie = cookie

	video, err := api.VideoFromBV(bv)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}

	_, err = api.ResolveVideo(video)
	if err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %w", err)
	}

	// For now, treat single video as non-collection
	// Collection detection would need additional API calls
	return map[string]interface{}{
		"is_collection": false,
		"video_count":   1,
		"videos": []map[string]string{
			{
				"title": video.Title,
				"url":   fmt.Sprintf("https://www.bilibili.com/video/%s", video.BV),
			},
		},
	}, nil
}

func DownloadCollectionWithProgress(url, sessdata, outputDir string, onlyAudio bool, onProgress func(string)) error {
	bv := extractBVFromURL(url)
	if bv == "" {
		return fmt.Errorf("无法从 URL 提取 BV 号: %s", url)
	}

	cookie := sessdata
	if cookie == "" {
		cookie = getCookie()
	}
	C.Cookie = cookie
	C.O = outputDir
	C.Merge = true
	C.Delete = true
	C.AddBVSuffix = true

	onProgress("正在获取视频信息...")

	video, err := api.VideoFromBV(bv)
	if err != nil {
		return fmt.Errorf("获取视频信息失败: %w", err)
	}

	_, err = api.ResolveVideo(video)
	if err != nil {
		return fmt.Errorf("解析视频信息失败: %w", err)
	}

	onProgress(fmt.Sprintf("下载: %s", video.Title))

	videoStreams, audioStreams, err := fetchDashStreams(video.BV, video.Cid, cookie)
	if err != nil {
		return fmt.Errorf("获取视频流信息失败: %w", err)
	}

	if len(videoStreams) == 0 {
		return fmt.Errorf("未找到可用的视频流")
	}
	if len(audioStreams) == 0 {
		return fmt.Errorf("未找到可用的音频流")
	}

	// Auto-select best video stream
	bestIdx := 0
	for i := 1; i < len(videoStreams); i++ {
		ci, cj := codecPriorityStr(videoStreams[i].Codecs), codecPriorityStr(videoStreams[bestIdx].Codecs)
		if ci > cj || (ci == cj && videoStreams[i].ID > videoStreams[bestIdx].ID) {
			bestIdx = i
		}
	}

	bestAudioIdx := 0
	for i := 1; i < len(audioStreams); i++ {
		if audioStreams[i].Bandwidth > audioStreams[bestAudioIdx].Bandwidth {
			bestAudioIdx = i
		}
	}

	stream := &api.Stream{
		V:     videoStreams[bestIdx].BaseURL,
		A:     audioStreams[bestAudioIdx].BaseURL,
		Video: api.Video{Title: video.Title, BV: video.BV, Cid: video.Cid},
	}

	onProgress("开始下载...")
	if err := api.Dl(stream); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	if !onlyAudio {
		onProgress("正在合并音视频...")
		if err := api.Merge(stream); err != nil {
			onProgress(fmt.Sprintf("合并失败: %v，保留单独文件", err))
		}
	}

	onProgress("下载完成")
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

func remuxToContainer(outputDir, title, bv, targetExt string) error {
	inputPath := filepath.Join(outputDir, title+".mp4")
	outputPath := filepath.Join(outputDir, title+"."+strings.ToLower(targetExt))

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", inputPath)
	}

	args := []string{"-i", inputPath, "-c", "copy", "-y"}
	// HEVC in MOV needs hvc1 tag for Apple compatibility
	if targetExt == "MOV" {
		args = append(args, "-tag:v", "hvc1")
	}
	args = append(args, outputPath)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg remux failed: %s: %w", string(output), err)
	}

	os.Remove(inputPath)
	return nil
}
