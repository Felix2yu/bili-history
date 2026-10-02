package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bilibili-history-go/config"
)

func TestSVCExtractBVFromURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://www.bilibili.com/video/BV1xx411c7mD", "BV1xx411c7mD"},
		{"https://b23.tv/abc", ""},
		{"BV1xx411c7mD", "BV1xx411c7mD"},
		{"", ""},
		{"https://example.com/?v=BV2ab3c4d5eF&p=1", "BV2ab3c4d5eF"},
		{"https://www.bilibili.com/video/BV1xx411c7mD/", "BV1xx411c7mD"},
		{"bv1xx411c7mD", ""}, // 正则区分大小写，小写 bv 不匹配（记录现状）
		{"https://www.bilibili.com/video/BV1xx411c7mD?spm_id_from=333.788", "BV1xx411c7mD"},
	}
	for _, c := range cases {
		if got := extractBVFromURL(c.in); got != c.want {
			t.Errorf("extractBVFromURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSVCCodecPriorityStr(t *testing.T) {
	cases := map[string]int{
		"av01.0.05M.08": 3,
		"hev1.1.6":      2,
		"hvc1.1.6":      2,
		"avc1.640032":   1,
		"":              0,
		"mp4a.40.2":     0,
	}
	for codec, want := range cases {
		if got := codecPriorityStr(codec); got != want {
			t.Errorf("codecPriorityStr(%q) = %d, want %d", codec, got, want)
		}
	}
}

func TestSVCFriendlyCodecName(t *testing.T) {
	cases := map[string]string{
		"av01.0.05M.08": "AV1",
		"hev1.1.6":      "HEVC",
		"hvc1.1.6":      "HEVC",
		"avc1.640032":   "AVC",
		"mp4a.40.2":     "mp4a.40.2", // 未识别时原样返回
	}
	for codec, want := range cases {
		if got := friendlyCodecName(codec); got != want {
			t.Errorf("friendlyCodecName(%q) = %q, want %q", codec, got, want)
		}
	}
}

func TestSVCContainerForCodec(t *testing.T) {
	cases := map[string]string{
		"av01.0.05M.08": "MKV",
		"hev1.1.6":      "MOV",
		"hvc1.1.6":      "MOV",
		"avc1.640032":   "MP4",
		"":              "MP4",
	}
	for codec, want := range cases {
		if got := containerForCodec(codec); got != want {
			t.Errorf("containerForCodec(%q) = %q, want %q", codec, got, want)
		}
	}
}

func TestSVCResolutionLabel(t *testing.T) {
	// 疑似生产缺陷：取的是 max(width, height) 而非短边，
	// 因此所有横屏视频档位被整体抬高一级（1920x1080 → "2K"）。
	// 这里固化实际行为，不做修复（见主会话报告）。
	cases := []struct {
		w, h int
		want string
	}{
		{3840, 2160, "4K"},
		{2160, 3840, "4K"}, // 竖屏取长边
		{2560, 1440, "4K"}, // max 边 2560 >= 2160 → 4K（实为 2K 屏）
		{1920, 1080, "2K"}, // max 边 1920 >= 1440 → 2K（实为 1080p）
		{1080, 1920, "2K"}, // 竖屏同样按长边
		{1280, 720, "1080p"},
		{854, 480, "720p"},
		{640, 360, "480p"},
		{320, 240, "320p"},
		{0, 0, "0p"},
	}
	for _, c := range cases {
		if got := resolutionLabel(c.w, c.h); got != c.want {
			t.Errorf("resolutionLabel(%d,%d) = %q, want %q", c.w, c.h, got, c.want)
		}
	}
}

func TestSVCFriendlyQualityLabel(t *testing.T) {
	// >= 1Mbps 走 Mbps 分支（分辨率标签沿用 resolutionLabel 的长边行为，见上）
	got := friendlyQualityLabel("avc1.640032", 1920, 1080, 3_000_000)
	if want := "2K · AVC · MP4 · 3.0Mbps"; got != want {
		t.Errorf("friendlyQualityLabel Mbps = %q, want %q", got, want)
	}
	// < 1Mbps 走 Kbps 分支（整除截断）
	got = friendlyQualityLabel("hev1.1.6", 2560, 1440, 320_000)
	if want := "4K · HEVC · MOV · 320Kbps"; got != want {
		t.Errorf("friendlyQualityLabel Kbps = %q, want %q", got, want)
	}
	got = friendlyQualityLabel("av01.0.05M.08", 3840, 2160, 1_000_000)
	if want := "4K · AV1 · MKV · 1.0Mbps"; got != want {
		t.Errorf("friendlyQualityLabel AV1 = %q, want %q", got, want)
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
	// 测试配置未设置 SESSDATA，getCookie 应返回空串
	if got := getCookie(); got != "" {
		t.Errorf("getCookie() = %q, want empty", got)
	}
}

func TestSVCCheckVideoDownloaded(t *testing.T) {
	out := GetDownloadOutputPath()
	svcMkdirAll(t, filepath.Join(out, "svcdir"))
	svcWrite(t, filepath.Join(out, "svcdir", "SVCVIDEO1_bvsvc_a.mp4"), []byte("x"))
	svcWrite(t, filepath.Join(out, "svcdir", "other.mp4"), []byte("y"))

	got := CheckVideoDownloaded([]string{"svcvideo1", "missingcid", "other"})
	if !got["svcvideo1"] {
		t.Errorf("expected svcvideo1 to be found (case-insensitive match)")
	}
	if got["missingcid"] {
		t.Errorf("missingcid should not be found")
	}
	if !got["other"] {
		t.Errorf("other should be found")
	}
	// 不存在的 cid 列表也应返回 false map 而非 panic
	got = CheckVideoDownloaded(nil)
	if len(got) != 0 {
		t.Errorf("nil cids should give empty map, got %v", got)
	}
}

func TestSVCListDownloadedVideos(t *testing.T) {
	out := GetDownloadOutputPath()
	sub := filepath.Join(out, "svclist")
	svcMkdirAll(t, sub)

	files := []string{
		"[svcUP主]svc带前缀标题 BVsvcMain0001.mp4",
		"svc普通标题.mp4",
		"svc音频_audio.mp4",
		"BVsvcMain0002.mkv",
		"ignored.txt",
	}
	for _, f := range files {
		svcWrite(t, filepath.Join(sub, f), []byte("0123456789"))
	}
	// 嵌套子目录也应被遍历
	deep := filepath.Join(sub, "nested")
	svcMkdirAll(t, deep)
	svcWrite(t, filepath.Join(deep, "svc嵌套视频.mp4"), []byte("abc"))

	 videos, total, err := ListDownloadedVideos("", 1, 100)
	if err != nil {
		t.Fatalf("ListDownloadedVideos: %v", err)
	}
	// 本测试创建 5 个媒体扩展名文件（ignored.txt 被排除）；
	// 更早测试遗留在同一 downloads 树中的 .mp4 也会被列出，因此只做下限断言。
	if total < 5 {
		t.Fatalf("total = %d, want >= 5 (%v)", total, videoTitles(videos))
	}

	byTitle := map[string]DownloadedVideo{}
	for _, v := range videos {
		byTitle[v.Title] = v
	}

	withPrefix, ok := byTitle["svc带前缀标题 BVsvcMain0001"]
	if !ok {
		t.Fatalf("bracket prefix title not parsed: %v", videoTitles(videos))
	}
	if withPrefix.AuthorName != "svcUP主" {
		t.Errorf("author prefix parse = %q, want svcUP主", withPrefix.AuthorName)
	}
	if withPrefix.BVID != "BVsvcMain0001" {
		t.Errorf("bvid parse = %q, want BVsvcMain0001", withPrefix.BVID)
	}
	// 与 2031 主数据集联动的元数据 join
	if withPrefix.AuthorMid != 7700001 {
		t.Errorf("author mid from DB = %d, want 7700001", withPrefix.AuthorMid)
	}
	if !strings.Contains(withPrefix.Cover, "BVsvcMain0001.jpg") {
		t.Errorf("cover from DB = %q", withPrefix.Cover)
	}

	mk, ok := byTitle["BVsvcMain0002"]
	if !ok {
		t.Fatalf("mkv video missing")
	}
	if mk.AuthorName != "svc作者乙" {
		t.Errorf("author fallback from DB = %q, want svc作者乙", mk.AuthorName)
	}
	if mk.BVID != "BVsvcMain0002" {
		t.Errorf("bvid = %q", mk.BVID)
	}

	audio := byTitle["svc音频_audio"]
	if !audio.Files[0].IsAudioOnly {
		t.Errorf("audio flag not set for _audio.mp4")
	}
	if audio.Files[0].SizeMB <= 0 {
		t.Errorf("size mb = %f, want > 0", audio.Files[0].SizeMB)
	}

	// 搜索过滤（大小写不敏感）
	filtered, ftotal, err := ListDownloadedVideos("普通", 1, 10)
	if err != nil || ftotal != 1 || filtered[0].Title != "svc普通标题" {
		t.Errorf("search filter: total=%d err=%v list=%v", ftotal, err, videoTitles(filtered))
	}

	// 分页越界：start/end 夹到 total，返回空页而非 panic
	page, ptotal, err := ListDownloadedVideos("", 99, 10)
	if err != nil || ptotal != total || len(page) != 0 {
		t.Errorf("out of range page: len=%d total=%d want total=%d err=%v", len(page), ptotal, total, err)
	}
	// limit 为 0 时返回空页
	zero, _, err := ListDownloadedVideos("", 1, 0)
	if err != nil || len(zero) != 0 {
		t.Errorf("zero limit: len=%d err=%v", len(zero), err)
	}
}

func videoTitles(vs []DownloadedVideo) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, v.Title)
	}
	return out
}

func TestSVCDeleteDownloadedVideo(t *testing.T) {
	out := GetDownloadOutputPath()

	// 空参数
	if err := DeleteDownloadedVideo("", false, ""); err == nil || !strings.Contains(err.Error(), "cid") {
		t.Errorf("empty cid/dir err = %v, want 缺少 cid", err)
	}

	// 按 cid 删除单文件
	dir := filepath.Join(out, "svcdel")
	svcMkdirAll(t, dir)
	f := svcWrite(t, filepath.Join(dir, "svccid999_part.mp4"), []byte("z"))
	if err := DeleteDownloadedVideo("svccid999", false, ""); err != nil {
		t.Fatalf("delete by cid: %v", err)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Errorf("file should be deleted")
	}

	// 未找到
	if err := DeleteDownloadedVideo("svcnotthere", false, ""); err == nil {
		t.Errorf("want 未找到匹配的视频文件 error")
	}

	// 指定目录删除整个目录
	deldir := filepath.Join(out, "svcdir2")
	svcMkdirAll(t, deldir)
	svcWrite(t, filepath.Join(deldir, "a.mp4"), []byte("a"))
	if err := DeleteDownloadedVideo("", true, deldir); err != nil {
		t.Fatalf("delete dir: %v", err)
	}
	if _, err := os.Stat(deldir); !os.IsNotExist(err) {
		t.Errorf("dir should be removed")
	}

	// 指定目录但不删目录：文件保留
	keep := filepath.Join(out, "svcdir3")
	svcMkdirAll(t, keep)
	kf := svcWrite(t, filepath.Join(keep, "b.mp4"), []byte("b"))
	if err := DeleteDownloadedVideo("", false, keep); err != nil {
		t.Fatalf("keep dir: %v", err)
	}
	if _, err := os.Stat(kf); err != nil {
		t.Errorf("file should be kept")
	}

	// 外部绝对目录被拒绝
	if err := DeleteDownloadedVideo("", true, mustGetwd(t)); err == nil {
		t.Errorf("want 无效的目录路径 for outside dir")
	}
	// 不存在的合法目录
	missing := filepath.Join(out, "svcdoesnotexist")
	if err := DeleteDownloadedVideo("", true, missing); err == nil || !strings.Contains(err.Error(), "不存在") {
		t.Errorf("missing dir err = %v, want 目录不存在", err)
	}

	// 疑似生产缺陷：目录校验仅用 strings.HasPrefix(outputDir)，
	// 兄弟目录 <output>/downloads-evil 带 outputDir 作为字符串前缀即可绕过。
	// 这里仅记录现状（发生在临时目录内，安全），不做修复。
	evil := out + "-evil"
	svcMkdirAll(t, evil)
	ef := filepath.Join(evil, "victim.txt")
	svcWrite(t, ef, []byte("v"))
	if err := DeleteDownloadedVideo("", true, evil); err != nil {
		t.Logf("sibling-prefix dir rejected: %v", err)
	} else if _, err := os.Stat(evil); os.IsNotExist(err) {
		t.Logf("NOTE: sibling-prefix dir %q passed HasPrefix validation and was deleted — suspected path-validation weakness in DeleteDownloadedVideo (download.go)", evil)
	}
}

func TestSVCCollectionGuardsWithoutBV(t *testing.T) {
	// 无 BV 号时三个入口都应在任何网络调用之前返回错误
	if _, err := ExtractVideoInfo("https://example.com/no-bv"); err == nil {
		t.Errorf("ExtractVideoInfo should fail without BV")
	}
	err := DownloadVideoWithProgress("https://example.com/no-bv", "", mustGetwd(t), false, "", func(string) {})
	if err == nil || !strings.Contains(err.Error(), "BV") {
		t.Errorf("DownloadVideoWithProgress err = %v", err)
	}
	if _, err := CheckCollection("https://example.com/no-bv"); err == nil {
		t.Errorf("CheckCollection should fail without BV")
	}
	err = DownloadCollectionWithProgress("https://example.com/no-bv", "", mustGetwd(t), false, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "BV") {
		t.Errorf("DownloadCollectionWithProgress err = %v", err)
	}
}

func TestSVCRemuxToContainerMissingInput(t *testing.T) {
	// 输入文件不存在时应提前返回错误，绝不触发 ffmpeg 子进程
	dir := filepath.Join(mustGetwd(t), "output", "svc-remux")
	svcMkdirAll(t, dir)
	err := remuxToContainer(dir, "svc不存在的视频", "BVsvcNone0001", "MOV")
	if err == nil || !strings.Contains(err.Error(), "文件不存在") {
		t.Errorf("remuxToContainer err = %v, want 文件不存在", err)
	}
}

func TestSVCDownloadVideoWithProgressMissingStreams(t *testing.T) {
	// 该函数在无 BV 之外的第一个失败点是 api.VideoFromBV（网络），
	// 因此除 BV 校验外不再深入；这里仅确认 BV 校验先行、onProgress 未被调用。
	var called int
	err := DownloadVideoWithProgress("garbage", "s", "", false, "", func(string) { called++ })
	if err == nil || called != 0 {
		t.Errorf("err=%v called=%d, want early error before progress", err, called)
	}
}

func TestSVCCookieWithSessdata(t *testing.T) {
	// 默认配置 SESSDATA 为空（TestSVCGetDownloadOutputPath 已断言空串分支）；
	// 这里临时注入 SESSDATA 覆盖返回 token 的分支。svcWithNotifyConfig 在测试结束时恢复原配置指针。
	if got := getCookie(); got != "" {
		t.Fatalf("precondition: getCookie() = %q, want empty before swap", got)
	}
	svcWithNotifyConfig(t, func(c *config.Config) {
		c.SESSDATA = "svc-cookie-token"
	})
	if got := getCookie(); got != "svc-cookie-token" {
		t.Errorf("getCookie() = %q, want svc-cookie-token", got)
	}
	// 恢复后再次为空（Cleanup 在本测试返回后执行，此处只验证 swap 生效语义）
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
