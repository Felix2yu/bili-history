package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bilibili-history-go/internal/config"

	"github.com/Felix2yu/bili-dl/api"
	"github.com/Felix2yu/bili-dl/c"
)

// 本文件覆盖 download.go 在改用 bili-dl 新接口之后仍然属于本包的逻辑：
// 标签渲染、码流排序、进度事件映射、全局参数装配，以及按新命名契约
// （UP 主子目录 + 标题_BV + 分 P 标签）扫描/判定/删除磁盘产物。
// 命名空间：目录 svcDL2*、bvid BVsvcMain0001/0002（关联 2031 主数据集）
// 与 BVsvcDel200x（只用于删除，绝不与别的测试重叠）。

func TestSVCCodecPriority(t *testing.T) {
	cases := []struct {
		codec string
		want  int
	}{
		{"av01.0.05M.08", 3},
		{"hev1.1.6", 2},
		{"hvc1.1.6", 2},
		{"avc1.640032", 1},
		{"mp4a.40.2", 0},
		{"", 0},
	}
	for _, c := range cases {
		if got := codecPriority(c.codec); got != c.want {
			t.Errorf("codecPriority(%q) = %d, want %d", c.codec, got, c.want)
		}
	}
}

func TestSVCFriendlyCodecName(t *testing.T) {
	cases := map[string]string{
		"av01.0.05M.08": "AV1",
		"hev1.1.6":      "HEVC",
		"hvc1.1.6":      "HEVC",
		"avc1.640032":   "AVC",
		"mp4a.40.2":     "AAC",
		"fLaC":          "FLAC",
		"unknown":       "UNKNOWN",
	}
	for codec, want := range cases {
		if got := friendlyCodecName(codec); got != want {
			t.Errorf("friendlyCodecName(%q) = %q, want %q", codec, got, want)
		}
	}
}

func TestSVCResolutionLabel(t *testing.T) {
	cases := []struct {
		w, h int
		want string
	}{
		{3840, 2160, "4K"},
		{2560, 1440, "2K"},
		{1920, 1080, "1080p"},
		{1280, 720, "720p"},
		{854, 480, "480p"},
		{640, 360, "360p"},
		{320, 240, "240p"},
		{1080, 1920, "1080p"}, // 竖屏取长边
	}
	for _, c := range cases {
		if got := resolutionLabel(c.w, c.h); got != c.want {
			t.Errorf("resolutionLabel(%d,%d) = %q, want %q", c.w, c.h, got, c.want)
		}
	}
}

func TestSVCCodecExtFor(t *testing.T) {
	cases := map[string]string{
		"av01.0.05M.08": "mkv",
		"hev1.1.6":      "mov",
		"avc1.640032":   "mp4",
		"mp4a.40.2":     "mp4",
	}
	for codec, want := range cases {
		if got := codecExtFor(codec); got != want {
			t.Errorf("codecExtFor(%q) = %q, want %q", codec, got, want)
		}
	}
}

func TestSVCStreamDesc(t *testing.T) {
	if got := streamDesc(&api.Stream{VCodec: "av01.0.05M.08", VWidth: 1920, VHeight: 1080}); got != "1080p · AV1 → MKV" {
		t.Errorf("streamDesc = %q", got)
	}
	// 仅音频时没有视频尺寸，不产出画质行。
	if got := streamDesc(&api.Stream{ACodec: "mp4a.40.2"}); got != "" {
		t.Errorf("streamDesc(audio only) = %q, want empty", got)
	}
}

func TestSVCBitrateAndStreamLabels(t *testing.T) {
	if got := bitrateLabel(3_000_000); got != "3.0Mbps" {
		t.Errorf("bitrateLabel(3M) = %q", got)
	}
	if got := bitrateLabel(320_000); got != "320Kbps" {
		t.Errorf("bitrateLabel(320K) = %q", got)
	}

	stream := StreamInfo{ID: 116, Codec: "av01.0.05M.08", Ext: "mkv", Width: 1920, Height: 1080, Bandwidth: 3_000_000}
	if got := streamLabel(stream); got != "1080p · AV1 · MKV · 3.0Mbps" {
		t.Errorf("streamLabel = %q", got)
	}
	audio := AudioInfo{ID: 30280, Codec: "mp4a.40.2", Ext: "m4a", Bandwidth: 194_560}
	if got := audioLabel(audio); got != "AAC · M4A · 194Kbps" {
		t.Errorf("audioLabel = %q", got)
	}
}

func TestSVCToStreamInfosOrdering(t *testing.T) {
	streams := toStreamInfos([]api.VideoOption{
		{ID: 32, Codec: "avc1.640032", Width: 854, Height: 480, Bandwidth: 400_000, Ext: "mp4"},
		{ID: 80, Codec: "hev1.1.6", Width: 1280, Height: 720, Bandwidth: 1_800_000, Ext: "mov"},
		{ID: 80, Codec: "avc1.640032", Width: 1280, Height: 720, Bandwidth: 1_200_000, Ext: "mp4"},
		{ID: 127, Codec: "av01.0.05M.08", Width: 3840, Height: 2160, Bandwidth: 9_000_000, Ext: "mkv"},
	})
	gotIDs := make([]string, 0, len(streams))
	for _, s := range streams {
		gotIDs = append(gotIDs, s.Quality)
	}
	if streams[0].ID != 127 {
		t.Fatalf("最高画质未排首位: %v", gotIDs)
	}
	// 同 id 内编码优先级：av01 > hev > avc。
	if streams[1].Codec != "hev1.1.6" || streams[2].Codec != "avc1.640032" {
		t.Errorf("同画质内排序错误: %v", gotIDs)
	}
	if streams[3].ID != 32 {
		t.Errorf("最低画质未排末位: %v", gotIDs)
	}
}

func TestSVCToAudioInfosByBandwidth(t *testing.T) {
	audios := toAudioInfos([]api.AudioOption{
		{ID: 30280, Codec: "mp4a.40.2", Bandwidth: 192_000, Ext: "m4a"},
		{ID: 30251, Codec: "fLaC", Bandwidth: 900_000, Ext: "flac"},
		{ID: 30250, Codec: "eac-3", Bandwidth: 640_000, Ext: "m4a"},
	})
	if audios[0].Codec != "fLaC" || audios[2].Codec != "mp4a.40.2" {
		t.Errorf("未按码率降序: %+v", audios)
	}
	if !strings.HasPrefix(audios[0].Quality, "FLAC") {
		t.Errorf("无损轨标签 = %q, want FLAC 开头", audios[0].Quality)
	}
}

func TestSVCProgressEventMapping(t *testing.T) {
	done := progressEvent(api.ProgressEvent{Event: "done", BV: "BVsvcDL2", Page: 2, Kind: "video", Message: "标题_BVsvcDL2.mp4"})
	if done.Type != "log" || !strings.Contains(done.Text, "视频流已保存") || done.BVID != "BVsvcDL2" || done.Page != 2 {
		t.Errorf("done mapping = %+v", done)
	}

	prog := progressEvent(api.ProgressEvent{Event: "progress", BV: "BVsvcDL2", Kind: "audio", Percent: 42, SpeedBPS: 1024, Current: 420, Total: 1000})
	if prog.Type != "progress" || prog.Percent != 42 || prog.Speed != 1024 || prog.Text != "" {
		t.Errorf("progress mapping = %+v", prog)
	}

	broken := progressEvent(api.ProgressEvent{Event: "error", Kind: "danmaku", Message: "HTTP 412"})
	if !strings.Contains(broken.Text, "弹幕失败") {
		t.Errorf("error mapping = %+v", broken)
	}

	start := progressEvent(api.ProgressEvent{Event: "start", Kind: "cover", Message: "cover.jpg"})
	if !strings.Contains(start.Text, "开始下载封面") {
		t.Errorf("start mapping = %+v", start)
	}
	if got := kindLabel("weird"); got != "weird" {
		t.Errorf("kindLabel(未知) = %q", got)
	}
}

func TestSVCShortCodecName(t *testing.T) {
	cases := map[string]string{
		"":                 "",
		"av1":              "av1",
		"av01.0.08M.08":    "av1",
		"hev1.1.6.L120.90": "hevc",
		"hvc1.1.6.L93.90":  "hevc",
		"avc1.64001F":      "avc",
		"mp4a.40.2":        "mp4a.40.2",
	}
	for codec, want := range cases {
		if got := shortCodecName(codec); got != want {
			t.Errorf("shortCodecName(%q) = %q, want %q", codec, got, want)
		}
	}
}

func TestSVCApplyGlobalsMapsOptions(t *testing.T) {
	t.Cleanup(func() {
		c.Defaults()
		c.Cookie, c.O, c.SubpathTemplate, c.Ctx = "", ".", "", nil
	})

	ctx := context.Background()
	applyGlobals("/tmp/svcDL2out", DownloadOptions{
		Sessdata:     "svcDL2-cookie",
		OnlyAudio:    true,
		VideoQuality: 116,
		VideoCodec:   "av01.0.08M.08",
		AudioQuality: 30280,
		Danmaku:      true,
		Subtitle:     false,
		Cover:        true,
		RateLimit:    "2m",
	}, ctx)

	if c.Cookie != "svcDL2-cookie" || c.O != "/tmp/svcDL2out" {
		t.Errorf("cookie/dir 未装配: %q %q", c.Cookie, c.O)
	}
	if c.SubpathTemplate != subpathTemplate {
		t.Errorf("模板 = %q, want %q", c.SubpathTemplate, subpathTemplate)
	}
	if !c.AudioOnly || c.VideoQuality != 116 || c.Codec != "av1" || c.AudioQuality != 30280 {
		t.Errorf("选流参数未装配（codec 需归一成 bili-dl 的短名）: %+v %d %q %d",
			c.AudioOnly, c.VideoQuality, c.Codec, c.AudioQuality)
	}
	if !c.DownloadDanmaku || !c.SaveCover || c.DownloadSubtitle {
		t.Errorf("附加内容开关未装配: danmaku=%v cover=%v subtitle=%v",
			c.DownloadDanmaku, c.SaveCover, c.DownloadSubtitle)
	}
	// 限速字符串由 bili-dl 解析成每秒字节数。
	if c.RateBytes <= 0 {
		t.Errorf("RateBytes = %d, want > 0", c.RateBytes)
	}
	if c.Ctx != ctx {
		t.Errorf("Ctx 未透传给 bili-dl")
	}
	// 合并只在 ffmpeg 可用时开启，否则保留分离流。
	if c.Merge != ffmpegAvailable() {
		t.Errorf("Merge = %v, want ffmpegAvailable() = %v", c.Merge, ffmpegAvailable())
	}
}

func TestSVCApplyGlobalsUsesConfigCookie(t *testing.T) {
	t.Cleanup(func() {
		c.Defaults()
		c.Cookie, c.O, c.SubpathTemplate, c.Ctx = "", ".", "", nil
	})
	svcWithNotifyConfig(t, func(cfg *config.Config) { cfg.SESSDATA = "svcDL2-from-config" })

	applyGlobals(".", DownloadOptions{}, context.Background())
	if c.Cookie != "svcDL2-from-config" {
		t.Errorf("未登录参数时应回落到配置 SESSDATA, got %q", c.Cookie)
	}
}

func TestSVCParseMediaPath(t *testing.T) {
	out := filepath.Join(string(filepath.Separator), "downloads")
	cases := []struct {
		name       string
		path       string
		wantAuthor string
		wantDir    string
		wantTitle  string
		wantBV     string
		wantPage   int
		wantPart   string
	}{
		{
			name:    "新模板单层UP主目录",
			path:    filepath.Join(out, "svcUP主", "svc新标题_BVsvcMain0001.mp4"),
			wantDir: "svcUP主", wantTitle: "svc新标题", wantBV: "BVsvcMain0001",
		},
		{
			name:    "多分P标签",
			path:    filepath.Join(out, "svcUP主", "svc总标题_BVsvcMain0002 - 03 - 第三集.mkv"),
			wantDir: "svcUP主", wantTitle: "svc总标题", wantBV: "BVsvcMain0002", wantPage: 3, wantPart: "第三集",
		},
		{
			name:    "分P括号退化形态",
			path:    filepath.Join(out, "svc无UP", "svc标题_BVsvcMain0003(2).mov"),
			wantDir: "svc无UP", wantTitle: "svc标题", wantBV: "BVsvcMain0003", wantPage: 2,
		},
		{
			name:       "历史扁平 [UP]标题",
			path:       filepath.Join(out, "[svc旧UP主]svc历史标题.mp4"),
			wantAuthor: "svc旧UP主", wantTitle: "svc历史标题",
		},
		{
			name:      "更深的自定义目录不猜UP主",
			path:      filepath.Join(out, "分组", "更深层", "svc标题_BVsvcMain0004.m4a"),
			wantTitle: "svc标题", wantBV: "BVsvcMain0004",
		},
		{
			// 文件名没有 BV 号时，单层目录也可能是用户的自定义分组，不能当作者。
			name:      "带分组目录但无BV号",
			path:      filepath.Join(out, "svc我的收藏夹", "svc标题.mp4"),
			wantTitle: "svc标题",
		},
		{
			name:      "无BV号的裸标题",
			path:      filepath.Join(out, "svc无号视频.mp4"),
			wantTitle: "svc无号视频",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseMediaPath(out, tc.path)
			if got.Author != tc.wantAuthor || got.DirAuthor != tc.wantDir || got.Title != tc.wantTitle ||
				got.BVID != tc.wantBV || got.Page != tc.wantPage || got.Part != tc.wantPart {
				t.Errorf("parseMediaPath(%q) = %+v", tc.path, got)
			}
		})
	}
}

func TestSVCIsMediaFile(t *testing.T) {
	cases := map[string]bool{
		"标题_BV1.mp4":         true,
		"标题_BV1.mkv":         true,
		"标题_BV1.mov":         true,
		"标题_BV1.m4a":         true,
		"标题_BV1.flac":        true,
		"标题_BV1.danmaku.ass": false,
		"标题_BV1.cn.srt":      false,
		"标题_BV1.cover.jpg":   false,
		"标题_BV1.merging.mp4": false,
		"info.txt":           false,
		"标题_BV1.xml":         false,
	}
	for name, want := range cases {
		if got := isMediaFile(name); got != want {
			t.Errorf("isMediaFile(%q) = %v, want %v", name, got, want)
		}
	}
	if !isAudioExt(".m4a") || !isAudioExt(".flac") || isAudioExt(".mkv") {
		t.Errorf("isAudioExt 判定错误")
	}
}

func TestSVCInsideDir(t *testing.T) {
	base := filepath.Join(string(filepath.Separator), "app", "downloads")
	cases := []struct {
		candidate string
		want      bool
	}{
		{base, true},
		{filepath.Join(base, "UP主"), true},
		{filepath.Join(base, "..", "UP主"), false},
		// 同名前缀的兄弟目录必须拒绝，纯字符串 HasPrefix 会漏。
		{base + "-evil", false},
		{filepath.Join(base, "nested", "deeper"), true},
	}
	for _, tc := range cases {
		if got := insideDir(base, tc.candidate); got != tc.want {
			t.Errorf("insideDir(%q) = %v, want %v", tc.candidate, got, tc.want)
		}
	}
}

// svcDL2CleanupDir 在临时命名空间内建夹具，测试结束整目录删除，
// 不污染与其它测试共享的 downloads 树。
func svcDL2CleanupDir(t *testing.T, dir string) {
	t.Helper()
	svcMkdirAll(t, dir)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
}

func TestSVCCheckVideoDownloaded(t *testing.T) {
	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDL2Check")
	svcDL2CleanupDir(t, dir)
	svcWrite(t, filepath.Join(dir, "svc标题一_BVsvcMain0001.mp4"), []byte("x"))
	svcWrite(t, filepath.Join(dir, "svc标题二_BVsvcDL2cover.cover.jpg"), []byte("y"))

	got := CheckVideoDownloaded([]string{"BVsvcMain0001", "BVsvcDL2Absent", ""})
	if !got["BVsvcMain0001"] {
		t.Errorf("BVsvcMain0001 应判定为已下载: %v", got)
	}
	if got["BVsvcDL2Absent"] {
		t.Errorf("不存在的 BV 不应命中: %v", got)
	}
	// 只有成品算已下载，封面 sidecar 不算。
	if got["BVsvcDL2cover"] {
		t.Errorf("仅有封面文件不应判定为已下载")
	}
	if empty := CheckVideoDownloaded(nil); len(empty) != 0 {
		t.Errorf("空列表应返回空 map, got %v", empty)
	}
}

func TestSVCListDownloadedVideos(t *testing.T) {
	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDL2List")
	svcDL2CleanupDir(t, dir)
	nested := filepath.Join(dir, "内层")
	svcMkdirAll(t, nested)

	svcWrite(t, filepath.Join(dir, "svc列表标题_BVsvcMain0001.mp4"), []byte("0123456789"))
	svcWrite(t, filepath.Join(dir, "svc未曾看过_BVsvcDL2nohist.mp4"), []byte("0123456789"))
	svcWrite(t, filepath.Join(dir, "svc列表标题_BVsvcMain0001.danmaku.ass"), []byte("ass"))
	svcWrite(t, filepath.Join(dir, "svc列表标题_BVsvcMain0001.cover.jpg"), []byte("jpg"))
	svcWrite(t, filepath.Join(dir, "svc列表标题_BVsvcMain0001.merging.mkv"), []byte("tmp"))
	svcWrite(t, filepath.Join(nested, "svc音频_BVsvcDL2deep.m4a"), []byte("audio"))
	svcWrite(t, filepath.Join(nested, "svc旧格式.mp4"), []byte("legacy"))
	svcWrite(t, filepath.Join(dir, "备注.txt"), []byte("note"))

	videos, total, err := ListDownloadedVideos("", 1, 1000)
	if err != nil {
		t.Fatalf("ListDownloadedVideos: %v", err)
	}
	if total != len(videos) {
		t.Errorf("total = %d, len = %d", total, len(videos))
	}

	byBV := map[string]DownloadedVideo{}
	for _, v := range videos {
		byBV[v.BVID] = v
	}

	main, ok := byBV["BVsvcMain0001"]
	if !ok {
		t.Fatalf("新命名条目未被列出: %+v", videos)
	}
	if main.Title != "svc列表标题" {
		t.Errorf("标题解析 = %q, want svc列表标题", main.Title)
	}
	if len(main.Files) != 1 || main.Files[0].IsAudioOnly {
		t.Errorf("成品文件应只有一条且非纯音频: %+v", main.Files)
	}
	// 关联 2031 主数据集里的作者与封面；库里有记录时以库为准。
	if main.AuthorMid != 7700001 || !strings.Contains(main.Cover, "BVsvcMain0001.jpg") {
		t.Errorf("历史库元数据关联失败: mid=%d cover=%q", main.AuthorMid, main.Cover)
	}
	if main.AuthorName != "svc作者甲" {
		t.Errorf("作者应以历史库为准, got %q", main.AuthorName)
	}

	// 库里没有这条投稿时，退回到 UP 主子目录名。
	fallback, ok := byBV["BVsvcDL2nohist"]
	if !ok {
		t.Fatalf("无历史记录的新命名条目未被列出: %+v", videos)
	}
	if fallback.AuthorName != "svcDL2List" {
		t.Errorf("无库记录时应取一层子目录作为 UP 主, got %q", fallback.AuthorName)
	}
	if fallback.Cover != "" || fallback.AuthorMid != 0 {
		t.Errorf("无库记录不应有封面元数据: %+v", fallback)
	}

	audio, ok := byBV["BVsvcDL2deep"]
	if !ok {
		t.Fatalf("深层目录里的音频条目缺失: %+v", videos)
	}
	if !audio.Files[0].IsAudioOnly {
		t.Errorf(".m4a 应标记为纯音频")
	}
	if audio.AuthorName != "" {
		t.Errorf("两层以上目录不应猜成 UP 主: %q", audio.AuthorName)
	}

	// 按 UP 主目录名、标题、BV 搜索都能命中。
	for _, term := range []string{"svc列表标题", "svcDL2List", "BVsvcMain0001"} {
		filtered, ftotal, err := ListDownloadedVideos(term, 1, 10)
		if err != nil || ftotal == 0 || len(filtered) != ftotal {
			t.Fatalf("搜索 %q 结果异常: total=%d len=%d err=%v", term, ftotal, len(filtered), err)
		}
	}

	// 分页越界返回空页，非法参数钳位到首页与默认页长。
	page, ptotal, err := ListDownloadedVideos("svc列表标题", 99, 10)
	if err != nil || ptotal != 1 || len(page) != 0 {
		t.Errorf("越界页: len=%d total=%d err=%v, want 0/1", len(page), ptotal, err)
	}
	if first, _, err := ListDownloadedVideos("svc列表标题", 0, 0); err != nil || len(first) != 1 {
		t.Errorf("非法分页参数应钳位到首页: len=%d err=%v", len(first), err)
	}
}

func TestSVCDeleteDownloadedVideo(t *testing.T) {
	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDL2Del")
	svcDL2CleanupDir(t, dir)

	video := svcWrite(t, filepath.Join(dir, "svc删除标题_BVsvcDel2001.mp4"), []byte("v"))
	sidecar := svcWrite(t, filepath.Join(dir, "svc删除标题_BVsvcDel2001.danmaku.ass"), []byte("d"))
	cover := svcWrite(t, filepath.Join(dir, "svc删除标题_BVsvcDel2001.cover.jpg"), []byte("c"))
	keep := svcWrite(t, filepath.Join(dir, "svc保留标题_BVsvcDel2002.mp4"), []byte("k"))

	// 空参数
	if err := DeleteDownloadedVideo("", false, ""); err == nil || !strings.Contains(err.Error(), "bvid") {
		t.Errorf("空 bvid err = %v, want 缺少 bvid", err)
	}
	// 未找到
	if err := DeleteDownloadedVideo("BVsvcDel2999", false, ""); err == nil ||
		!strings.Contains(err.Error(), "未找到匹配的视频文件") {
		t.Errorf("未命中 err = %v", err)
	}

	// 按 BV 删除：成品与附加内容一起消失，同目录其它条目保留。
	if err := DeleteDownloadedVideo("BVsvcDel2001", false, ""); err != nil {
		t.Fatalf("delete by bvid: %v", err)
	}
	for _, path := range []string{video, sidecar, cover} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s 应已删除", path)
		}
	}
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("同 UP 其它条目文件被误删: %v", err)
	}
	// 目录里还有别的条目，delete_directory 时不能整目录删除。
	if err := DeleteDownloadedVideo("BVsvcDel2002", true, ""); err != nil {
		t.Fatalf("delete last entry with directory: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("UP 目录清空后应移除: stat = %v", err)
	}
}

func TestSVCDeleteDownloadedVideoRejectsOutsideDirectory(t *testing.T) {
	out := GetDownloadOutputPath()

	// 兄弟目录带 outputDir 作为字符串前缀，早期实现会被 HasPrefix 放过。
	evil := out + "-svcDL2evil"
	svcDL2CleanupDir(t, evil)
	svcWrite(t, filepath.Join(evil, "victim.mp4"), []byte("v"))

	if err := DeleteDownloadedVideo("", true, evil); err == nil ||
		!strings.Contains(err.Error(), "无效的目录路径") {
		t.Errorf("sibling-prefix dir err = %v, want 无效的目录路径", err)
	}
	if _, err := os.Stat(evil); err != nil {
		t.Errorf("被拒绝的目录必须完好无损: %v", err)
	}

	// 合法但缺失的目录给出明确错误。
	missing := filepath.Join(out, "svcDL2NeverExists")
	if err := DeleteDownloadedVideo("", true, missing); err == nil ||
		!strings.Contains(err.Error(), "不存在") {
		t.Errorf("missing dir err = %v, want 目录不存在", err)
	}

	// 指定目录且不删目录：文件保留。
	keep := filepath.Join(out, "svcDL2Keep")
	svcDL2CleanupDir(t, keep)
	kf := svcWrite(t, filepath.Join(keep, "svc保留.mp4"), []byte("b"))
	if err := DeleteDownloadedVideo("", false, keep); err != nil {
		t.Fatalf("keep dir: %v", err)
	}
	if _, err := os.Stat(kf); err != nil {
		t.Errorf("文件应保留: %v", err)
	}
}

func TestSVCDownloadGuardsStayOffline(t *testing.T) {
	// 无法解析的链接只产生日志，最终给出「没有可下载的内容」，
	// 绝不在解析失败处悄悄走网络。
	var events []DownloadEvent
	err := Download(context.Background(), []string{"https://example.com/no-bv", "garbage"},
		filepath.Join(GetDownloadOutputPath(), "svcDL2Guard"), DownloadOptions{},
		func(ev DownloadEvent) { events = append(events, ev) })
	if err == nil || !strings.Contains(err.Error(), "没有可下载的内容") {
		t.Fatalf("err = %v, want 没有可下载的内容", err)
	}
	if len(events) == 0 || events[0].Type != "log" {
		t.Errorf("解析失败应有日志事件, got %+v", events)
	}

	if err := Download(context.Background(), nil, "", DownloadOptions{}, func(DownloadEvent) {}); err == nil ||
		!strings.Contains(err.Error(), "没有待下载的链接") {
		t.Errorf("空链接 err = %v", err)
	}
	// 已取消的 context 直接短路，不产生任何下载动作。
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Download(cancelled, []string{"BVsvcDL2cancelled"}, "", DownloadOptions{}, func(DownloadEvent) {}); err == nil {
		t.Errorf("已取消的 context 应返回错误")
	}
	_ = os.RemoveAll(filepath.Join(GetDownloadOutputPath(), "svcDL2Guard"))
}

func TestSVCDownloadRejectsWhenBusy(t *testing.T) {
	// bili-dl 靠包级变量传参，两个下载同时跑会互相踩参数，后来者必须被拒绝。
	biliMu.Lock()
	t.Cleanup(biliMu.Unlock)

	var called int
	err := Download(context.Background(), []string{"BVsvcDL2busy"}, "", DownloadOptions{},
		func(DownloadEvent) { called++ })
	if err == nil || !strings.Contains(err.Error(), "已有下载任务") {
		t.Fatalf("err = %v, want 已有下载任务在进行中", err)
	}
	if called != 0 {
		t.Errorf("被拒绝时不应推送任何事件, called=%d", called)
	}
}

func TestSVCExtractAndCheckCollectionGuardsWithoutBV(t *testing.T) {
	// 解析入口在联网之前就返回错误。
	if _, err := ExtractVideoInfo("https://example.com/no-bv"); err == nil {
		t.Errorf("ExtractVideoInfo 应在无 BV 时失败")
	}
	if _, err := CheckCollection("https://example.com/no-bv"); err == nil {
		t.Errorf("CheckCollection 应在无 BV 时失败")
	}
}

func TestSVCFavoriteLinksValidation(t *testing.T) {
	// 非法 fid 必须在任何本地查询与在线请求之前失败。
	for _, bad := range []string{"", "  ", "abc", "0", "-3"} {
		if _, err := FavoriteLinks(bad); err == nil || !strings.Contains(err.Error(), "无效收藏夹 id") {
			t.Errorf("FavoriteLinks(%q) err = %v, want 无效收藏夹 id", bad, err)
		}
	}

	// 本地收藏库没有这个收藏夹（测试环境未同步）且没配 SESSDATA 时，
	// 给出可读错误而不是静默零产出，也绝不发在线请求。
	if getCookie() != "" {
		t.Skip("测试配置带了 SESSDATA，跳过离线回退断言")
	}
	_, err := FavoriteLinks("9007199254740991")
	if err == nil || !strings.Contains(err.Error(), "未同步到本地") {
		t.Fatalf("err = %v, want 收藏夹未同步且未配置 SESSDATA", err)
	}
}

func TestSVCDanmakuFile(t *testing.T) {
	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDL2Danmaku")
	svcDL2CleanupDir(t, dir)

	video := svcWrite(t, filepath.Join(dir, "svc弹幕标题_BVsvcDL2dm01.mkv"), []byte("video"))
	ass := filepath.Join(dir, "svc弹幕标题_BVsvcDL2dm01.danmaku.ass")
	svcWrite(t, ass, []byte("[Script Info]\nDialogue: 0,0:00:01.00,0:00:05.00,Default,,0,0,0,,测试弹幕\n"))

	// 视频成品路径推导出 bili-dl 的 sidecar 命名
	body, name, err := DanmakuFile(video, "")
	if err != nil {
		t.Fatalf("DanmakuFile(video path): %v", err)
	}
	if name != filepath.Base(ass) || !strings.Contains(string(body), "测试弹幕") {
		t.Fatalf("返回内容 = %q name=%q", body, name)
	}

	// 直接传弹幕路径同样可用
	if _, _, err := DanmakuFile(ass, ""); err != nil {
		t.Errorf("DanmakuFile(ass path): %v", err)
	}

	// 只有 BV 号时按目录检索
	if _, _, err := DanmakuFile("", "BVsvcDL2dm01"); err != nil {
		t.Errorf("DanmakuFile(by bvid): %v", err)
	}

	// 下载目录之外一律拒绝，不给任意文件读取留口子
	outside := svcWrite(t, filepath.Join(t.TempDir(), "secret.ass"), []byte("SECRET"))
	if _, _, err := DanmakuFile(outside, ""); !errors.Is(err, ErrDanmakuForbidden) {
		t.Errorf("outside-dir err = %v, want ErrDanmakuForbidden", err)
	}

	// 没有弹幕文件的视频给出明确未找到
	quiet := svcWrite(t, filepath.Join(dir, "svc无弹幕_BVsvcDL2dm02.mp4"), []byte("v"))
	if _, _, err := DanmakuFile(quiet, ""); !errors.Is(err, ErrDanmakuNotFound) {
		t.Errorf("missing danmaku err = %v, want ErrDanmakuNotFound", err)
	}
}

func TestSVCGetDownloadOutputPathAndCookie(t *testing.T) {
	p := GetDownloadOutputPath()
	if !strings.HasSuffix(p, filepath.Join("output", "downloads")) {
		t.Errorf("GetDownloadOutputPath() = %q, want suffix output/downloads", p)
	}
	// 注意：GetOutputPath 只创建父目录，downloads 目录本身要等到写入时才出现
	if _, err := os.Stat(filepath.Dir(p)); err != nil {
		t.Errorf("output root should exist: %v", err)
	}
	if got := getCookie(); got != "" {
		t.Errorf("getCookie() = %q, want empty", got)
	}
}

func TestSVCCookieWithSessdata(t *testing.T) {
	// 默认配置 SESSDATA 为空（上一个测试已断言空串分支）；
	// 这里临时注入 SESSDATA 覆盖返回 token 的分支。
	if got := getCookie(); got != "" {
		t.Fatalf("precondition: getCookie() = %q, want empty before swap", got)
	}
	svcWithNotifyConfig(t, func(cfg *config.Config) { cfg.SESSDATA = "svc-cookie-token" })
	if got := getCookie(); got != "svc-cookie-token" {
		t.Errorf("getCookie() = %q, want svc-cookie-token", got)
	}
}

// --- small FS helpers ---

func svcMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func svcWrite(t *testing.T, path string, content []byte) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}
