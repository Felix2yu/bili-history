package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bilibili-history-go/internal/biliapi"
	"bilibili-history-go/internal/config"
	"bilibili-history-go/internal/database"
	"bilibili-history-go/internal/utils"

	"github.com/Felix2yu/bili-dl/api"
	"github.com/Felix2yu/bili-dl/c"
)

// 产物命名交给 bili-dl 的路径模板：UP 主子目录 + 标题带 BV 号。
// BV 留在文件名里，库列表才能按 BV 关联本地历史中的封面与作者。
const subpathTemplate = "{owner_uname}/{title}_{bv}"

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

// VideoInfo 是下载弹框需要的解析结果：可选码流、分 P、UP 主与封面。
type VideoInfo struct {
	URL          string       `json:"url"`
	Site         string       `json:"site"`
	Type         string       `json:"type"`
	Title        string       `json:"title"`
	BVID         string       `json:"bvid"`
	Owner        string       `json:"owner"`
	Cover        string       `json:"cover"`
	Streams      []StreamInfo `json:"streams"`
	AudioStreams []AudioInfo  `json:"audio_streams"`
	Pages        []PageInfo   `json:"pages"`
}

// StreamInfo 是一条可选视频码流，ID 即 DASH 画质 id（对应 bili-dl 的 -q）。
type StreamInfo struct {
	ID        int    `json:"id"`
	Codec     string `json:"codec"`
	Quality   string `json:"quality"`
	Ext       string `json:"ext"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Bandwidth int64  `json:"bandwidth"`
}

// AudioInfo 是一条可选音频码流，含 dolby 与 flac 无损轨。
type AudioInfo struct {
	ID        int    `json:"id"`
	Codec     string `json:"codec"`
	Quality   string `json:"quality"`
	Ext       string `json:"ext"`
	Bandwidth int64  `json:"bandwidth"`
}

type PageInfo struct {
	Page     int    `json:"page"`
	Cid      string `json:"cid"`
	Part     string `json:"part"`
	Duration int    `json:"duration"`
}

// DownloadOptions 一次下载的全部可选项，零值即「全自动」。
type DownloadOptions struct {
	Sessdata     string
	OnlyAudio    bool
	VideoQuality int
	VideoCodec   string
	AudioQuality int
	Page         int
	Batch        bool
	Collection   bool
	Danmaku      bool
	Subtitle     bool
	Cover        bool
	RateLimit    string
}

// DownloadEvent 推送给前端的一条下载消息。
type DownloadEvent struct {
	Type    string  `json:"type"` // log / progress
	Text    string  `json:"text,omitempty"`
	Event   string  `json:"event,omitempty"`
	BVID    string  `json:"bvid,omitempty"`
	Page    int     `json:"page,omitempty"`
	Kind    string  `json:"kind,omitempty"`
	Percent float64 `json:"percent,omitempty"`
	Speed   int64   `json:"speed_bps,omitempty"`
	Current int64   `json:"current,omitempty"`
	Total   int64   `json:"total,omitempty"`
}

type DownloadedVideoFile struct {
	FilePath    string  `json:"file_path"`
	SizeMB      float64 `json:"size_mb"`
	IsAudioOnly bool    `json:"is_audio_only"`
}

type DownloadedVideo struct {
	Title        string                `json:"title"`
	BVID         string                `json:"bvid"`
	Page         int                   `json:"page,omitempty"`
	Part         string                `json:"part,omitempty"`
	Cover        string                `json:"cover"`
	AuthorName   string                `json:"author_name"`
	AuthorFace   string                `json:"author_face"`
	AuthorMid    int64                 `json:"author_mid"`
	DownloadDate string                `json:"download_date"`
	Files        []DownloadedVideoFile `json:"files"`
	Directory    string                `json:"directory"`
}

// CollectionVideo 是合集 / 番剧 / 多分 P 投稿里的一个条目。
type CollectionVideo struct {
	BVID  string `json:"bvid"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Cid   string `json:"cid,omitempty"`
	Cover string `json:"cover,omitempty"`
	Page  int    `json:"page,omitempty"`
}

// CollectionInfo 是合集检查结果，前端据此决定是否走批量下载。
type CollectionInfo struct {
	IsCollection bool              `json:"is_collection"`
	Type         string            `json:"type"`
	Title        string            `json:"title"`
	VideoCount   int               `json:"video_count"`
	Videos       []CollectionVideo `json:"videos"`
}

// bili-dl 通过包级变量接收配置，任何两次并发调用都会互相踩参数，
// 因此所有对它的调用共用这一把锁，且全局变量只在持锁时写入。
var biliMu sync.Mutex

var (
	hookMu   sync.RWMutex
	hookSink func(api.ProgressEvent)
)

func init() {
	api.SetProgressHook(func(ev api.ProgressEvent) {
		hookMu.RLock()
		sink := hookSink
		hookMu.RUnlock()
		if sink != nil {
			sink(ev)
		}
	})
}

// useProgressSink 把 bili-dl 的进度事件接到当前任务的 emit 上。
func useProgressSink(emit func(DownloadEvent)) func() {
	hookMu.Lock()
	hookSink = func(ev api.ProgressEvent) { emit(progressEvent(ev)) }
	hookMu.Unlock()
	return func() {
		hookMu.Lock()
		hookSink = nil
		hookMu.Unlock()
	}
}

func progressEvent(ev api.ProgressEvent) DownloadEvent {
	base := DownloadEvent{Event: ev.Event, BVID: ev.BV, Page: ev.Page, Kind: ev.Kind}
	kind := kindLabel(ev.Kind)
	switch ev.Event {
	case "progress":
		base.Type = "progress"
		base.Percent = ev.Percent
		base.Speed = ev.SpeedBPS
		base.Current = ev.Current
		base.Total = ev.Total
	case "done":
		base.Type = "log"
		base.Text = fmt.Sprintf("%s已保存：%s", kind, ev.Message)
	case "error":
		base.Type = "log"
		base.Text = fmt.Sprintf("%s失败：%s", kind, ev.Message)
	default:
		base.Type = "log"
		base.Text = fmt.Sprintf("开始下载%s：%s", kind, ev.Message)
	}
	return base
}

func kindLabel(kind string) string {
	switch kind {
	case "video":
		return "视频流"
	case "audio":
		return "音频流"
	case "cover":
		return "封面"
	case "danmaku":
		return "弹幕"
	case "subtitle":
		return "字幕"
	case "":
		return "资源"
	default:
		return kind
	}
}

func GetDownloadOutputPath() string {
	return utils.GetOutputPath("downloads")
}

func getCookie() string {
	cfg := config.GetConfig()
	if cfg == nil {
		return ""
	}
	return cfg.SESSDATA
}

var (
	ffmpegOnce   sync.Once
	ffmpegPathOK bool
)

func ffmpegAvailable() bool {
	ffmpegOnce.Do(func() {
		_, err := exec.LookPath("ffmpeg")
		ffmpegPathOK = err == nil
	})
	return ffmpegPathOK
}

// shortCodecName 把 DASH 的完整 codec 串归一成 bili-dl 认识的 av1/hevc/avc。
// 传原样串会被 matchCodec 当成「不过滤」，选到用户没选的那条流。
func shortCodecName(codec string) string {
	switch {
	case codec == "":
		return ""
	case strings.HasPrefix(codec, "av01"):
		return "av1"
	case strings.HasPrefix(codec, "hev"), strings.HasPrefix(codec, "hvc"):
		return "hevc"
	case strings.HasPrefix(codec, "avc"):
		return "avc"
	default:
		return codec
	}
}

// applyGlobals 写入 bili-dl 全局参数，调用方必须持有 biliMu。
func applyGlobals(outputDir string, opts DownloadOptions, ctx context.Context) {
	c.Defaults()
	cookie := opts.Sessdata
	if cookie == "" {
		cookie = getCookie()
	}
	c.Cookie = cookie
	c.O = outputDir
	c.SubpathTemplate = subpathTemplate
	c.Delete = true
	c.AudioOnly = opts.OnlyAudio
	c.VideoQuality = opts.VideoQuality
	c.Codec = shortCodecName(opts.VideoCodec)
	c.AudioQuality = opts.AudioQuality
	c.RateBytes = c.ParseRate(opts.RateLimit)
	c.DownloadDanmaku = opts.Danmaku
	c.DownloadSubtitle = opts.Subtitle
	c.SaveCover = opts.Cover
	c.FFMPEG = ffmpegAvailable()
	// 没有 ffmpeg 就谈不上合并，DASH 分离流原样保留。
	c.Merge = c.FFMPEG
	c.Ctx = ctx
}

// resolutionLabel 按 B 站的档位语义取短边：1920x1080 是 1080p，竖屏 1080x1920 同理。
func resolutionLabel(width, height int) string {
	h := height
	if width < height {
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

func friendlyCodecName(codec string) string {
	switch {
	case strings.HasPrefix(codec, "av01"):
		return "AV1"
	case strings.HasPrefix(codec, "hev"), strings.HasPrefix(codec, "hvc"):
		return "HEVC"
	case strings.HasPrefix(codec, "avc"):
		return "AVC"
	case strings.Contains(strings.ToLower(codec), "flac"):
		return "FLAC"
	case strings.Contains(strings.ToLower(codec), "mp4a"):
		return "AAC"
	default:
		return strings.ToUpper(codec)
	}
}

// codecPriority 决定同画质下的展示与选择顺序：av1 > hevc > avc。
func codecPriority(codec string) int {
	switch {
	case strings.HasPrefix(codec, "av01"):
		return 3
	case strings.HasPrefix(codec, "hev"), strings.HasPrefix(codec, "hvc"):
		return 2
	case strings.HasPrefix(codec, "avc"):
		return 1
	default:
		return 0
	}
}

func bitrateLabel(bandwidth int64) string {
	kbps := float64(bandwidth) / 1000
	if kbps >= 1000 {
		return fmt.Sprintf("%.1fMbps", kbps/1000)
	}
	return fmt.Sprintf("%dKbps", int(kbps))
}

func streamLabel(s StreamInfo) string {
	return fmt.Sprintf("%s · %s · %s · %s",
		resolutionLabel(s.Width, s.Height), friendlyCodecName(s.Codec),
		strings.ToUpper(s.Ext), bitrateLabel(s.Bandwidth))
}

func audioLabel(a AudioInfo) string {
	return fmt.Sprintf("%s · %s · %s",
		friendlyCodecName(a.Codec), strings.ToUpper(a.Ext), bitrateLabel(a.Bandwidth))
}

func toStreamInfos(options []api.VideoOption) []StreamInfo {
	streams := make([]StreamInfo, 0, len(options))
	for _, o := range options {
		streams = append(streams, StreamInfo{
			ID: o.ID, Codec: o.Codec, Ext: o.Ext,
			Width: o.Width, Height: o.Height, Bandwidth: o.Bandwidth,
		})
	}
	sort.SliceStable(streams, func(i, j int) bool {
		if streams[i].ID != streams[j].ID {
			return streams[i].ID > streams[j].ID
		}
		return codecPriority(streams[i].Codec) > codecPriority(streams[j].Codec)
	})
	for i := range streams {
		streams[i].Quality = streamLabel(streams[i])
	}
	return streams
}

func toAudioInfos(options []api.AudioOption) []AudioInfo {
	audios := make([]AudioInfo, 0, len(options))
	for _, o := range options {
		audios = append(audios, AudioInfo{ID: o.ID, Codec: o.Codec, Ext: o.Ext, Bandwidth: o.Bandwidth})
	}
	sort.SliceStable(audios, func(i, j int) bool { return audios[i].Bandwidth > audios[j].Bandwidth })
	for i := range audios {
		audios[i].Quality = audioLabel(audios[i])
	}
	return audios
}

func pagesOf(v api.Video) []PageInfo {
	pages := make([]PageInfo, 0, len(v.Pages))
	for _, p := range v.Pages {
		pages = append(pages, PageInfo{Page: p.Page, Cid: p.Cid.String(), Part: p.Part, Duration: p.Duration})
	}
	return pages
}

// resolvePGC 把番剧输入解析成目标剧集：kind 为 ep 且未展开整季时只取该集。
func resolvePGC(kind, id string, expandAll bool) ([]api.Video, error) {
	seasonID, epID := "", id
	if kind == "ss" {
		seasonID, epID = id, ""
	}
	episodes, err := api.SeasonVideos(seasonID, epID)
	if err != nil {
		return nil, err
	}
	if kind == "ep" && !expandAll {
		for _, e := range episodes {
			if e.EpID == id {
				return []api.Video{e}, nil
			}
		}
		return nil, fmt.Errorf("番剧列表里没有 ep%s", id)
	}
	return episodes, nil
}

// resolveTargets 把一条链接展开成下载目标：番剧按剧集、投稿按分 P，
// Collection 打开时把所属合集整体展开。调用方必须持有 biliMu。
func resolveTargets(link string, opts DownloadOptions) ([]api.Video, error) {
	if kind, id := api.ParsePGCSInput(link); kind != "" {
		return resolvePGC(kind, id, opts.Batch || opts.Collection)
	}

	bv, page, err := api.ParseInput(link)
	if err != nil {
		return nil, err
	}
	if opts.Page > 0 {
		page = opts.Page
	}
	video, err := api.VideoFromBV(bv)
	if err != nil {
		return nil, err
	}

	if opts.Collection && video.UGCSeason != nil {
		if parts := api.ExpandUGCSeason(video); len(parts) > 0 {
			// 合集条目只是带着 cid 的壳，各自再解析一次才能拿到分 P。
			var out []api.Video
			for _, part := range parts {
				full, err := api.VideoFromBV(part.BV)
				if err != nil {
					return out, fmt.Errorf("合集条目 %s 解析失败: %w", part.BV, err)
				}
				out = append(out, api.ExpandPages(full, 1, opts.Batch)...)
			}
			return out, nil
		}
	}
	return api.ExpandPages(video, page, opts.Batch), nil
}

// ExtractVideoInfo 解析一条链接，返回可选码流、分 P 与基本信息。
func ExtractVideoInfo(link string) (*VideoInfo, error) {
	// 只读操作：拿到锁就按自己的参数装配，正在下载时直接沿用当前全局参数，
	// 避免弹框被一个几分钟的下载阻塞到请求超时。
	if biliMu.TryLock() {
		defer biliMu.Unlock()
		applyGlobals(GetDownloadOutputPath(), DownloadOptions{}, context.Background())
	}

	info := &VideoInfo{URL: link, Site: "bilibili"}
	var target api.Video
	if kind, id := api.ParsePGCSInput(link); kind != "" {
		targets, err := resolvePGC(kind, id, false)
		if err != nil {
			return nil, fmt.Errorf("获取番剧信息失败: %w", err)
		}
		target = targets[0]
		info.Type = "pgc"
	} else {
		bv, page, err := api.ParseInput(link)
		if err != nil {
			return nil, fmt.Errorf("无法解析链接: %w", err)
		}
		video, err := api.VideoFromBV(bv)
		if err != nil {
			return nil, fmt.Errorf("获取视频信息失败: %w", err)
		}
		targets := api.ExpandPages(video, page, false)
		if len(targets) == 0 {
			return nil, fmt.Errorf("视频 %s 没有可下载的分 P", bv)
		}
		target = targets[0]
		info.Type = "video"
	}

	info.Title = target.Title
	info.BVID = target.BV
	info.Owner = target.Owner
	info.Cover = target.Pic
	info.Pages = pagesOf(target)

	videoOptions, audioOptions, err := api.ListOptions(target)
	if err != nil {
		// 码流候选拿不到不致命：下载会走自动选流，画质栏留空即可。
		return info, nil
	}
	info.Streams = toStreamInfos(videoOptions)
	info.AudioStreams = toAudioInfos(audioOptions)
	return info, nil
}

func logEvent(emit func(DownloadEvent), format string, args ...any) {
	emit(DownloadEvent{Type: "log", Text: fmt.Sprintf(format, args...)})
}

// Download 依次下载给定链接，链接可以是投稿、番剧或多条 BV。
// ctx 取消会传导到 bili-dl 的请求与 ffmpeg 子进程。
func Download(ctx context.Context, links []string, outputDir string, opts DownloadOptions, emit func(DownloadEvent)) error {
	if len(links) == 0 {
		return fmt.Errorf("没有待下载的链接")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if outputDir == "" {
		outputDir = GetDownloadOutputPath()
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	if !biliMu.TryLock() {
		return fmt.Errorf("已有下载任务在进行中，请等待完成后再试")
	}
	defer biliMu.Unlock()

	stopHook := useProgressSink(emit)
	defer stopHook()

	applyGlobals(outputDir, opts, ctx)

	var targets []api.Video
	for _, link := range links {
		resolved, err := resolveTargets(link, opts)
		if err != nil {
			logEvent(emit, "解析 %s 失败: %v", link, err)
			continue
		}
		targets = append(targets, resolved...)
	}
	if len(targets) == 0 {
		return fmt.Errorf("没有可下载的内容")
	}
	logEvent(emit, "共 %d 个下载任务", len(targets))

	var failed int
	for i, target := range targets {
		if err := ctx.Err(); err != nil {
			logEvent(emit, "已取消，剩余 %d 个任务未执行", len(targets)-i)
			return err
		}
		logEvent(emit, "[%d/%d] %s", i+1, len(targets), target.Title)
		if err := downloadTarget(target, emit); err != nil {
			failed++
			logEvent(emit, "[%d/%d] 失败: %v", i+1, len(targets), err)
		}
	}
	if failed == len(targets) {
		return fmt.Errorf("%d 个任务全部失败", failed)
	}
	if failed > 0 {
		logEvent(emit, "完成，%d 个任务失败", failed)
	}
	logEvent(emit, "下载完成")
	return nil
}

func downloadTarget(v api.Video, emit func(DownloadEvent)) error {
	stream, err := api.GetStream(v)
	if err != nil {
		return fmt.Errorf("获取下载码流失败: %w", err)
	}
	if desc := streamDesc(stream); desc != "" {
		logEvent(emit, "画质: %s", desc)
	}
	if err := api.Dl(stream); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	switch {
	case c.Merge:
		if err := api.Merge(stream); err != nil {
			logEvent(emit, "合并失败: %v，保留单独文件", err)
		}
	case !c.AudioOnly:
		logEvent(emit, "未检测到 ffmpeg，已保留未合并的分离流")
	}
	// 弹幕 / 字幕 / 封面是 best-effort，失败不影响本条任务的结果。
	api.DownloadExtras(stream)
	return nil
}

func streamDesc(stream *api.Stream) string {
	if stream.VWidth == 0 && stream.VHeight == 0 {
		return ""
	}
	return fmt.Sprintf("%s · %s → %s",
		resolutionLabel(int(stream.VWidth), int(stream.VHeight)),
		friendlyCodecName(stream.VCodec),
		strings.ToUpper(codecExtFor(stream.VCodec)))
}

func codecExtFor(codec string) string {
	switch {
	case strings.HasPrefix(codec, "av01"):
		return "mkv"
	case strings.HasPrefix(codec, "hev"):
		return "mov"
	default:
		return "mp4"
	}
}

// CheckCollection 识别一条链接背后是否有多个可下载条目：合集、番剧或多分 P 投稿。
func CheckCollection(link string) (*CollectionInfo, error) {
	if biliMu.TryLock() {
		defer biliMu.Unlock()
		applyGlobals(GetDownloadOutputPath(), DownloadOptions{}, context.Background())
	}

	if kind, id := api.ParsePGCSInput(link); kind != "" {
		episodes, err := resolvePGC(kind, id, true)
		if err != nil {
			return nil, fmt.Errorf("获取番剧信息失败: %w", err)
		}
		info := &CollectionInfo{
			IsCollection: len(episodes) > 1,
			Type:         "pgc",
			Title:        episodes[0].Title,
			VideoCount:   len(episodes),
		}
		for _, e := range episodes {
			info.Videos = append(info.Videos, CollectionVideo{
				BVID: e.BV, Title: e.Title, Cid: e.Cid, Page: e.ActivePage,
				Cover: e.Pic,
				URL:   fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%s", e.EpID),
			})
		}
		return info, nil
	}

	bv, _, err := api.ParseInput(link)
	if err != nil {
		return nil, fmt.Errorf("无法解析链接: %w", err)
	}
	video, err := api.VideoFromBV(bv)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}

	info := &CollectionInfo{Type: "video", Title: video.Title}
	if video.UGCSeason != nil {
		if parts := api.ExpandUGCSeason(video); len(parts) > 1 {
			info.IsCollection = true
			info.VideoCount = len(parts)
			for _, part := range parts {
				title := part.Title
				if strings.TrimSpace(title) == "" {
					title = video.Title
				}
				info.Videos = append(info.Videos, CollectionVideo{
					BVID: part.BV, Title: title, Cid: part.Cid, Cover: part.Pic,
					URL: fmt.Sprintf("https://www.bilibili.com/video/%s", part.BV),
				})
			}
			return info, nil
		}
	}

	// 非合集投稿按分 P 展开，多 P 也值得让前端提示批量下载。
	info.VideoCount = len(video.Pages)
	for _, p := range video.Pages {
		info.Videos = append(info.Videos, CollectionVideo{
			BVID: bv, Title: p.Part, Cid: p.Cid.String(), Page: p.Page, Cover: video.Pic,
			URL: fmt.Sprintf("https://www.bilibili.com/video/%s?p=%d", bv, p.Page),
		})
	}
	return info, nil
}

var (
	reBVInName  = regexp.MustCompile(`BV[a-zA-Z0-9]+`)
	rePageLabel = regexp.MustCompile(` - (\d{2,}) - (.+)$`)
	rePageParen = regexp.MustCompile(`\((\d+)\)$`)
)

// mediaFile 是从落盘路径还原出的条目语义。
type mediaFile struct {
	// Author 来自文件名里的 [UP主] 前缀（历史扁平命名）。
	Author string
	// DirAuthor 来自 UP 主子目录，仅在文件名带 BV、且位于下载目录下一层时才可信，
	// 用户自己按主题分组的文件夹名不能当作者。
	DirAuthor string
	Title     string
	BVID      string
	Page      int
	Part      string
}

// parseMediaPath 解析 outputDir 下的一个文件，兼容新模板
// （UP 主子目录 + 标题_BV + 分 P 标签）与历史的扁平 [UP主]标题 命名。
func parseMediaPath(outputDir, path string) mediaFile {
	rel, err := filepath.Rel(outputDir, path)
	if err != nil {
		rel = path
	}
	var entry mediaFile
	if dir := filepath.Dir(rel); dir != "." && dir != string(filepath.Separator) {
		trimmed := strings.Trim(filepath.ToSlash(dir), "/")
		if !strings.Contains(trimmed, "/") {
			entry.DirAuthor = trimmed
		}
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if bv := reBVInName.FindString(name); bv != "" {
		entry.BVID = bv
		name = strings.Replace(name, "_"+bv, "", 1)
		name = strings.Replace(name, bv, "", 1)
	}
	if entry.BVID == "" {
		entry.DirAuthor = ""
	}
	if m := rePageLabel.FindStringSubmatch(name); len(m) == 3 {
		entry.Page, _ = strconv.Atoi(m[1])
		entry.Part = strings.TrimSpace(m[2])
		name = strings.TrimSuffix(name, m[0])
	} else if m := rePageParen.FindStringSubmatch(name); len(m) == 2 {
		entry.Page, _ = strconv.Atoi(m[1])
		name = strings.TrimSuffix(name, m[0])
	}

	if entry.Author == "" && strings.HasPrefix(name, "[") {
		if end := strings.Index(name, "]"); end > 0 {
			entry.Author = name[1:end]
			name = name[end+1:]
		}
	}
	entry.Title = strings.Join(strings.Fields(strings.Trim(name, "_- ")), " ")
	return entry
}

// isAudioOnly 判定纯音频：新产物是 m4a/flac/mp3，历史命名里带 _audio 的
// 单流文件也按纯音频处理。
func isAudioOnly(ext, name string) bool {
	return isAudioExt(ext) || strings.Contains(strings.ToLower(name), "_audio")
}

func isAudioExt(ext string) bool {
	switch ext {
	case ".m4a", ".flac", ".mp3":
		return true
	default:
		return false
	}
}

// isMediaFile 过滤成品文件：中间态 .merging 与附加内容（弹幕/字幕/封面）都不算条目。
func isMediaFile(name string) bool {
	lower := strings.ToLower(name)
	if strings.Contains(lower, ".merging") {
		return false
	}
	switch filepath.Ext(lower) {
	case ".mp4", ".mkv", ".mov", ".webm", ".avi", ".flv", ".m4a", ".flac", ".mp3":
		return true
	default:
		return false
	}
}

// insideDir 判断 abs 是否落在 base 内，含相等与父子边界，
// 单纯字符串前缀比较会把 base 的同名前兄弟目录也算进去。
func insideDir(base, abs string) bool {
	rel, err := filepath.Rel(base, abs)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

type bvMeta struct {
	cover      string
	authorName string
	authorFace string
	authorMid  int64
}

// loadBVMetadata 从历史记录数据库加载所有 BV 号对应的封面和作者信息
func loadBVMetadata() map[string]bvMeta {
	result := make(map[string]bvMeta)

	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return result
	}

	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		return result
	}

	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		rows, err := conn.Query(fmt.Sprintf(`
			SELECT bvid, cover, author_name, author_face, author_mid
			FROM %s
			WHERE bvid != '' AND cover != ''
			GROUP BY bvid
		`, tableName))
		if err != nil {
			continue
		}

		for rows.Next() {
			var bvid, cover, authorName, authorFace string
			var authorMid int64
			if rows.Scan(&bvid, &cover, &authorName, &authorFace, &authorMid) == nil {
				if _, exists := result[bvid]; !exists {
					result[bvid] = bvMeta{
						cover:      cover,
						authorName: authorName,
						authorFace: authorFace,
						authorMid:  authorMid,
					}
				}
			}
		}
		rows.Close()
	}

	return result
}

// ListDownloadedVideos 列出下载目录里的成品，按 BV 关联本地历史的封面与作者。
func ListDownloadedVideos(search string, page, limit int) ([]DownloadedVideo, int, error) {
	outputDir := GetDownloadOutputPath()
	metadata := loadBVMetadata()

	type entry struct {
		video    DownloadedVideo
		modified time.Time
	}
	var items []entry

	// 遍历错误（目录不存在等）在这里等同「没有文件」，回调已把 err 转成 nil。
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !isMediaFile(info.Name()) {
			return nil
		}

		parsed := parseMediaPath(outputDir, path)
		title := parsed.Title
		if title == "" {
			title = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		}
		if search != "" && !matchesSearch(search, parsed, title) {
			return nil
		}

		cover, face := "", ""
		var mid int64
		if meta, ok := metadata[parsed.BVID]; ok {
			cover, face, mid = meta.cover, meta.authorFace, meta.authorMid
			if parsed.Author == "" {
				parsed.Author = meta.authorName
			}
		}
		// 没看过这条投稿时历史库里没有作者记录，UP 主子目录名是唯一可靠来源。
		if parsed.Author == "" {
			parsed.Author = parsed.DirAuthor
		}

		ext := strings.ToLower(filepath.Ext(path))
		items = append(items, entry{
			modified: info.ModTime(),
			video: DownloadedVideo{
				Title:        title,
				BVID:         parsed.BVID,
				Page:         parsed.Page,
				Part:         parsed.Part,
				Cover:        cover,
				AuthorName:   parsed.Author,
				AuthorFace:   face,
				AuthorMid:    mid,
				DownloadDate: info.ModTime().Format("2006-01-02"),
				Directory:    filepath.Dir(path),
				Files: []DownloadedVideoFile{{
					FilePath:    path,
					SizeMB:      float64(info.Size()) / 1024 / 1024,
					IsAudioOnly: isAudioOnly(ext, info.Name()),
				}},
			},
		})
		return nil
	})

	sort.Slice(items, func(i, j int) bool { return items[i].modified.After(items[j].modified) })

	total := len(items)
	// 与 routers 的分页守卫保持一致：非法 page/limit 会算出越界下标，而本函数是导出的。
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	videos := make([]DownloadedVideo, 0, end-start)
	for _, item := range items[start:end] {
		videos = append(videos, item.video)
	}
	return videos, total, nil
}

func matchesSearch(search string, parsed mediaFile, title string) bool {
	needle := strings.ToLower(search)
	for _, candidate := range []string{title, parsed.Author, parsed.DirAuthor, parsed.Part, parsed.BVID} {
		if strings.Contains(strings.ToLower(candidate), needle) {
			return true
		}
	}
	return false
}

// CheckVideoDownloaded 按 BV 号判断下载目录里是否已有成品。
func CheckVideoDownloaded(bvids []string) map[string]bool {
	outputDir := GetDownloadOutputPath()
	result := make(map[string]bool, len(bvids))
	wanted := make(map[string]bool, len(bvids))
	for _, bv := range bvids {
		result[bv] = false
		wanted[bv] = true
	}

	// 目录不存在等同「都没下载过」，回调已吞掉遍历错误。
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !isMediaFile(info.Name()) {
			return nil
		}
		for _, bv := range reBVInName.FindAllString(info.Name(), -1) {
			if wanted[bv] {
				result[bv] = true
			}
		}
		return nil
	})
	return result
}

// DeleteDownloadedVideo 删除某个 BV 的全部产物（成品与弹幕/字幕/封面等附加文件）。
// directory 非空时按目录整体处理，此时 bvid 忽略。
func DeleteDownloadedVideo(bvid string, deleteDir bool, directory string) error {
	outputDir := GetDownloadOutputPath()

	if directory != "" {
		absDir, err := filepath.Abs(directory)
		if err != nil || !insideDir(outputDir, absDir) {
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

	if bvid == "" {
		return fmt.Errorf("缺少 bvid 参数")
	}

	var matched []string
	needle := strings.ToLower(bvid)
	// 同 CheckVideoDownloaded：遍历错误等同「没找到」。历史文件名里的 BV 大小写不一，
	// 匹配沿用大小写不敏感语义。
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(strings.ToLower(info.Name()), needle) {
			matched = append(matched, path)
		}
		return nil
	})
	if len(matched) == 0 {
		return fmt.Errorf("未找到匹配的视频文件")
	}

	for _, path := range matched {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	if !deleteDir {
		return nil
	}

	// UP 主子目录是整个 UP 主共用的，只能删属于本 BV 的文件，目录空了才移除。
	ownerDir := filepath.Dir(matched[0])
	if ownerDir == outputDir {
		return nil
	}
	if entries, err := os.ReadDir(ownerDir); err == nil && len(entries) == 0 {
		return os.Remove(ownerDir)
	}
	return nil
}

// favoriteOnlinePageSize 是在线拉收藏夹时每页条数，B 站接口上限为 20。
const favoriteOnlinePageSize = 20

// FavoriteLinks 返回收藏夹里可下载的投稿链接：本地收藏库优先，
// 没同步过再走在线接口，两路都拿不到时返回明确错误。
func FavoriteLinks(fid string) ([]string, error) {
	mediaID, err := strconv.ParseInt(strings.TrimSpace(fid), 10, 64)
	if err != nil || mediaID <= 0 {
		return nil, fmt.Errorf("无效收藏夹 id: %s", fid)
	}

	list, _, err := database.GetFavoriteContents(mediaID, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("读取本地收藏夹失败: %w", err)
	}
	links := make([]string, 0, len(list))
	for _, item := range list {
		if bvid, _ := item["bvid"].(string); bvid != "" {
			links = append(links, "https://www.bilibili.com/video/"+bvid)
		}
	}
	if len(links) > 0 {
		return links, nil
	}

	return favoriteLinksOnline(mediaID)
}

func favoriteLinksOnline(mediaID int64) ([]string, error) {
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		return nil, fmt.Errorf("收藏夹 %d 未同步到本地，且未配置 SESSDATA，无法在线获取", mediaID)
	}

	client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
	if client.ImgKey == "" || client.SubKey == "" {
		_ = client.FetchWbiKeys()
	}

	var links []string
	for pn := 1; ; pn++ {
		data, err := client.GetFavoriteResources(mediaID, pn, favoriteOnlinePageSize)
		if err != nil {
			if apiErr, ok := err.(*biliapi.APIError); ok && apiErr.Code == -6 {
				return nil, fmt.Errorf("Cookie 已过期，请重新登录后再试")
			}
			return nil, fmt.Errorf("获取收藏夹内容失败: %w", err)
		}
		for _, item := range data.Media {
			if item.UGC != nil && item.UGC.Bvid != "" {
				links = append(links, "https://www.bilibili.com/video/"+item.UGC.Bvid)
			}
		}
		if len(data.Media) == 0 || len(links) >= data.Page.Count {
			break
		}
	}
	return links, nil
}

// 弹幕读取的错误分支，路由据此区分 403 与 404。
var (
	ErrDanmakuForbidden = errors.New("弹幕路径不在下载目录内")
	ErrDanmakuNotFound  = errors.New("未找到弹幕文件")
)

// DanmakuFile 返回与视频成品同源的弹幕文本。入参可以是视频路径、弹幕文件路径，
// 或 BV 号；优先 bili-dl 产出的 `标题_BV.danmaku.ass`，其次是同名的任意 .ass。
// 命名规则由下载侧决定，所以推断交给这里而不是前端。
func DanmakuFile(path, bvid string) ([]byte, string, error) {
	outputDir := GetDownloadOutputPath()

	var candidates []string
	if trimmed := strings.TrimSpace(path); trimmed != "" {
		abs, err := filepath.Abs(trimmed)
		if err != nil {
			return nil, "", ErrDanmakuForbidden
		}
		if !insideDir(outputDir, abs) {
			return nil, "", ErrDanmakuForbidden
		}
		switch ext := strings.ToLower(filepath.Ext(abs)); ext {
		case ".ass", ".xml":
			candidates = append(candidates, abs)
		default:
			base := strings.TrimSuffix(abs, ext)
			candidates = append(candidates, base+".danmaku.ass", base+".ass")
		}
	}
	for _, candidate := range candidates {
		if body, err := os.ReadFile(candidate); err == nil {
			return body, filepath.Base(candidate), nil
		}
	}

	if bv := strings.TrimSpace(bvid); bv != "" {
		if found := findDanmakuByBV(outputDir, bv); found != "" {
			if body, err := os.ReadFile(found); err == nil {
				return body, filepath.Base(found), nil
			}
		}
	}
	return nil, "", ErrDanmakuNotFound
}

func findDanmakuByBV(outputDir, bv string) string {
	needle := strings.ToLower(bv)
	var hit string
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, ".ass") && strings.Contains(name, needle) {
			hit = path
			return filepath.SkipAll
		}
		return nil
	})
	return hit
}

// GetUserVideos 拉取 UP 主投稿列表；bili-dl 的 AllVideo 不带分页，这里保留自己的实现。
func GetUserVideos(mid string, pn, ps int) ([]map[string]string, int, error) {
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/space/wbi/arc/search?mid=%s&pn=%d&ps=%d&order=pubdate", mid, pn, ps)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://www.bilibili.com")
	if cookie := getCookie(); cookie != "" {
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
					BVID   string `json:"bvid"`
					Title  string `json:"title"`
					Pic    string `json:"pic"`
					AID    int64  `json:"aid"`
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

	videos := make([]map[string]string, 0, len(result.Data.List.Vlist))
	for _, v := range result.Data.List.Vlist {
		videos = append(videos, map[string]string{
			"bvid":   v.BVID,
			"title":  v.Title,
			"pic":    v.Pic,
			"aid":    strconv.FormatInt(v.AID, 10),
			"length": v.Length,
		})
	}
	return videos, result.Data.Page.Count, nil
}
