package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"bilibili-history-go/database"
)

// 本文件覆盖 download.go 的数据层/文件系统分支：
//   - loadBVMetadata：年份表选择、TableExists=false 跳过、查询失败跳过、
//     空 cover / 空 bvid 过滤、Scan 失败跳过
//   - ListDownloadedVideos：不可读子目录、.m4a、缺失 "]" 的文件名、分页越界
//   - DeleteDownloadedVideo：deleteDir 与按 cid 删除、不可读目录
//   - remuxToContainer：ffmpeg 参数构造（MOV/MKV）、成功、失败
//   - resolutionLabel 的 360p 档位
//
// 命名空间：年份表 2035/2036、bvid 前缀 BVsvcDL、文件名前缀 svcDl。
// 2031/2032/当前年属于其它测试，绝不落数据、也绝不据其断言
// （cleanOldHistory 会把 2031 的固定日期行删掉）。
//
// 另外 svc_mcp/svc_clean/svc_history/svc_image 把 bilibili_history_2036 当作
// “缺少业务列的损坏表”夹具（CREATE IF NOT EXISTS + DROP），svc_image 还把
// 2035 当作“没有年份表”的空夹具。所以本文件创建的两张表都只在测试期间存在，
// 结束时一律 DROP，绝不把正表留给别的测试。
const (
	svcDlYearMeta  = 2035 // 正表，承载本文件的元数据行（用完即删）
	svcDlYearGhost = 2036 // 只创建 bilibili_history_2036_ghost，不创建正表
)

// svcDlInsertMetaRow 直接 INSERT，便于写入 models.HistoryRecord 无法表达的
// 边界值（空 cover、空 bvid、author_mid 为非法文本）。mid 传 any。
func svcDlInsertMetaRow(t *testing.T, bvid, cover, name, face string, mid any, viewAt int64) {
	t.Helper()
	table := fmt.Sprintf("bilibili_history_%d", svcDlYearMeta)
	_, err := svcConn().Exec(`
		INSERT INTO `+table+` (
			title, oid, bvid, dt, author_name, author_face, author_mid,
			view_at, duration, cover, business
		) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		"svcDl 行 "+bvid, viewAt%1000+1, bvid, 1, name, face, mid, viewAt, 100, cover, "archive")
	if err != nil {
		t.Fatalf("insert svcDl meta row %s: %v", bvid, err)
	}
}

func svcDlEnsureMetaTable(t *testing.T) {
	t.Helper()
	table := fmt.Sprintf("bilibili_history_%d", svcDlYearMeta)
	if err := database.GetSQLiteDB().EnsureTableForYear(svcDlYearMeta); err != nil {
		t.Fatalf("ensure table %d: %v", svcDlYearMeta, err)
	}
	t.Cleanup(func() {
		// 先清数据再判断这张表是不是本文件建的（有 author_mid 列），
		// 是就整张删掉：别的测试把 2035 当无表年份、把 2036 当损坏表夹具。
		_, _ = svcConn().Exec("DELETE FROM " + table + " WHERE bvid LIKE 'BVsvcDL%' OR bvid = ''")
		var mine int
		if err := svcConn().QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info(?) WHERE name='author_mid'", table).Scan(&mine); err != nil {
			t.Errorf("probe %s schema: %v", table, err)
			return
		}
		if mine == 0 {
			return
		}
		if _, err := svcConn().Exec("DROP TABLE IF EXISTS " + table); err != nil {
			t.Errorf("drop %s: %v", table, err)
		}
	})
}

// svcDlGhostTable 创建一张名字能被 GetAvailableYears 解析成 2036、
// 但正表 bilibili_history_2036 不存在的表，用来覆盖 TableExists=false 的 continue。
func svcDlGhostTable(t *testing.T) {
	t.Helper()
	ghost := fmt.Sprintf("bilibili_history_%d_ghost", svcDlYearGhost)
	if _, err := svcConn().Exec("CREATE TABLE IF NOT EXISTS " + ghost + " (id INTEGER PRIMARY KEY, bvid TEXT)"); err != nil {
		t.Fatalf("create ghost table: %v", err)
	}
	if _, err := svcConn().Exec("INSERT INTO " + ghost + " (bvid) VALUES ('BVsvcDLghost')"); err != nil {
		t.Fatalf("insert ghost row: %v", err)
	}
	t.Cleanup(func() {
		if _, err := svcConn().Exec("DROP TABLE IF EXISTS " + ghost); err != nil {
			t.Errorf("drop ghost table: %v", err)
		}
	})
}

func TestSVCDLLoadBVMetadata(t *testing.T) {
	svcDlEnsureMetaTable(t)
	svcDlGhostTable(t)

	// 前置：ghost 年份被 GetAvailableYears 报出来，但同名正表不存在。
	found := false
	years, err := database.GetSQLiteDB().GetAvailableYears()
	if err != nil {
		t.Fatalf("GetAvailableYears: %v", err)
	}
	for _, y := range years {
		if y == svcDlYearGhost {
			found = true
		}
	}
	if !found {
		t.Fatalf("years %v should contain %d (derived from the ghost table)", years, svcDlYearGhost)
	}
	ghostPlain := fmt.Sprintf("bilibili_history_%d", svcDlYearGhost)
	if ok, err := database.GetSQLiteDB().TableExists(ghostPlain); err != nil || ok {
		t.Fatalf("TableExists(%s) = (%v,%v), want (false,nil)", ghostPlain, ok, err)
	}

	const (
		goodBV   = "BVsvcDLmetaGood"
		noCovBV  = "BVsvcDLmetaNoCover"
		emptyBV  = ""
		badMidBV = "BVsvcDLmetaBadMid"
	)
	ts := svcAt(svcDlYearMeta, 4, 1, 9, 0)
	svcDlInsertMetaRow(t, goodBV, "https://cover.invalid/svcDL-good.jpg", "svcDl作者甲",
		"https://face.invalid/svcDL-good.jpg", 910001, ts)
	// 同一 bvid 的第二行：GROUP BY bvid 应只产出一条。
	svcDlInsertMetaRow(t, goodBV, "https://cover.invalid/svcDL-good2.jpg", "svcDl作者甲",
		"https://face.invalid/svcDL-good2.jpg", 910001, ts+60)
	svcDlInsertMetaRow(t, noCovBV, "", "svcDl无封面", "https://face.invalid/svcDL-nocover.jpg", 910002, ts)
	svcDlInsertMetaRow(t, emptyBV, "https://cover.invalid/svcDL-empty.jpg", "svcDl空bvid",
		"https://face.invalid/svcDL-empty.jpg", 910003, ts)
	// INTEGER 亲和列塞进无法转换的文本 => 以 TEXT 存储，rows.Scan(&int64) 失败。
	svcDlInsertMetaRow(t, badMidBV, "https://cover.invalid/svcDL-badmid.jpg", "svcDl坏mid",
		"https://face.invalid/svcDL-badmid.jpg", "svcDL不是数字", ts)

	meta := loadBVMetadata()

	got, ok := meta[goodBV]
	if !ok {
		t.Fatalf("meta missing %s: keys=%d", goodBV, len(meta))
	}
	if got.authorName != "svcDl作者甲" || got.authorFace == "" || got.cover == "" {
		t.Errorf("meta[%s] = %+v", goodBV, got)
	}
	if got.authorMid != 910001 {
		t.Errorf("authorMid = %d, want 910001", got.authorMid)
	}
	if !strings.HasPrefix(got.cover, "https://cover.invalid/svcDL-good") {
		t.Errorf("cover = %q", got.cover)
	}
	if _, ok := meta[noCovBV]; ok {
		t.Errorf("empty-cover row %s must be filtered out", noCovBV)
	}
	if _, ok := meta[emptyBV]; ok {
		t.Errorf("empty-bvid row must be filtered out")
	}
	if _, ok := meta[badMidBV]; ok {
		t.Errorf("row with non-numeric author_mid must be skipped on Scan error")
	}

	// 查询失败分支：临时把 author_mid 改名，SELECT 报 "no such column"，
	// loadBVMetadata 应对该年份 continue 而不 panic（其余年份照常返回）。
	table := fmt.Sprintf("bilibili_history_%d", svcDlYearMeta)
	if _, err := svcConn().Exec("ALTER TABLE " + table + " RENAME COLUMN author_mid TO svc_dl_hidden_mid"); err != nil {
		t.Fatalf("rename author_mid: %v", err)
	}
	defer svcDlRestoreAuthorMid(t, table)
	if _, err := svcConn().Exec("SELECT author_mid FROM " + table + " LIMIT 1"); err == nil {
		t.Errorf("precondition: author_mid should be unreadable after rename")
	}
	broken := loadBVMetadata()
	if _, ok := broken[goodBV]; ok {
		t.Errorf("meta[%s] present although the year query fails", goodBV)
	}
	if _, err := svcConn().Exec("ALTER TABLE " + table + " RENAME COLUMN svc_dl_hidden_mid TO author_mid"); err != nil {
		t.Fatalf("restore author_mid: %v", err)
	}

	// 恢复后重新可见（证明上面的缺失不是数据被删）。
	meta2 := loadBVMetadata()
	if got2, ok := meta2[goodBV]; !ok || got2.authorMid != 910001 {
		t.Errorf("after restore meta[%s] = %+v ok=%v", goodBV, got2, ok)
	}
}

// svcDlRestoreAuthorMid 只有在 author_mid 仍被改名藏起来时才改名回来，
// 因此可以安全地被显式调用与 defer 各执行一次。
func svcDlRestoreAuthorMid(t *testing.T, table string) {
	t.Helper()
	var hidden int
	if err := svcConn().QueryRow("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name='svc_dl_hidden_mid'", table).Scan(&hidden); err != nil {
		t.Errorf("probe hidden column: %v", err)
		return
	}
	if hidden == 0 {
		return
	}
	if _, err := svcConn().Exec("ALTER TABLE " + table + " RENAME COLUMN svc_dl_hidden_mid TO author_mid"); err != nil {
		t.Errorf("restore author_mid: %v", err)
	}
}

func TestSVCDLListDownloadedVideosMetadataJoin(t *testing.T) {
	svcDlEnsureMetaTable(t)

	bvid := "BVsvcDLjoin001"
	ts := svcAt(svcDlYearMeta, 5, 1, 12, 0)
	svcDlInsertMetaRow(t, bvid, "https://cover.invalid/svcDL-join.jpg", "svcDl联合作者",
		"https://face.invalid/svcDL-join.jpg", 920001, ts)

	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDlJoin")
	svcMkdirAll(t, dir)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	// 无 [作者] 前缀，作者信息必须来自 2036 行。
	svcWrite(t, filepath.Join(dir, "svcDl联合作用视频 "+bvid+".mp4"), []byte("0123456789ABCD"))
	// PRODUCTION BUG: services/download.go:557-560 —— 扩展名白名单
	// (.mp4/.flv/.mkv/.webm/.avi) 不含 .m4a，所以下面 download.go:589 的
	// `strings.HasSuffix(name, ".m4a")` 纯音频判定永远不可能命中（.m4a 早在
	// 白名单处就被 return nil 跳过）。这里固化“音频文件不出现在列表中”的现状。
	m4aPath := svcWrite(t, filepath.Join(dir, "svcDl纯音频 "+bvid+".m4a"), []byte("AUDIO"))
	// 以 "[" 开头但缺少 "]"：作者名保持为空，标题保留原始文件名。
	svcWrite(t, filepath.Join(dir, "[svcDl残缺前缀标题.mp4"), []byte("X"))

	// "_audio" 关键词命中 IsAudioOnly 的字符串分支。
	svcWrite(t, filepath.Join(dir, "svcDl纯音频 "+bvid+"_audio.mp4"), []byte("AUDIO"))

	videos, total, err := ListDownloadedVideos("svcDl", 1, 50)
	if err != nil {
		t.Fatalf("ListDownloadedVideos: %v", err)
	}
	if total != len(videos) {
		t.Errorf("total = %d, len = %d", total, len(videos))
	}
	pathSet := map[string]bool{}
	for _, v := range videos {
		for _, f := range v.Files {
			pathSet[f.FilePath] = true
		}
	}
	byTitle := map[string]DownloadedVideo{}
	for _, v := range videos {
		byTitle[v.Title] = v
	}

	// 回归 download.go 的扩展名白名单：.m4a 过去不在白名单里，第 589 行的
	// 纯音频后缀判定因此永远命中不了；现在音频文件必须出现在列表中。
	if !pathSet[m4aPath] {
		t.Errorf(".m4a file %s should be listed", m4aPath)
	}
	m4a, ok := byTitle["svcDl纯音频 "+bvid]
	if !ok {
		t.Fatalf(".m4a entry missing, titles=%v", svcDlVideoTitles(videos))
	}
	if !m4a.Files[0].IsAudioOnly {
		t.Errorf(".m4a must be flagged audio only: %+v", m4a.Files[0])
	}

	mp4, ok := byTitle["svcDl联合作用视频 "+bvid]
	if !ok {
		t.Fatalf("mp4 entry missing, titles=%v", svcDlVideoTitles(videos))
	}
	if mp4.BVID != bvid {
		t.Errorf("bvid = %q, want %q", mp4.BVID, bvid)
	}
	if mp4.Cover != "https://cover.invalid/svcDL-join.jpg" {
		t.Errorf("cover = %q", mp4.Cover)
	}
	if mp4.AuthorName != "svcDl联合作者" {
		t.Errorf("author name = %q", mp4.AuthorName)
	}
	if mp4.AuthorFace != "https://face.invalid/svcDL-join.jpg" {
		t.Errorf("author face = %q", mp4.AuthorFace)
	}
	if mp4.AuthorMid != 920001 {
		t.Errorf("author mid = %d, want 920001", mp4.AuthorMid)
	}
	if mp4.Directory != dir {
		t.Errorf("directory = %q, want %q", mp4.Directory, dir)
	}
	if len(mp4.Files) != 1 || mp4.Files[0].FilePath == "" {
		t.Fatalf("files = %+v", mp4.Files)
	}
	if mp4.Files[0].IsAudioOnly {
		t.Errorf("mp4 must not be audio only")
	}
	if want := 14.0 / 1024 / 1024; mp4.Files[0].SizeMB != want {
		t.Errorf("sizeMB = %v, want %v", mp4.Files[0].SizeMB, want)
	}
	if mp4.DownloadDate == "" || len(mp4.DownloadDate) != 10 ||
		!strings.Contains(mp4.DownloadDate, "-") {
		t.Errorf("download date = %q, want YYYY-MM-DD", mp4.DownloadDate)
	}

	audio, ok := byTitle["svcDl纯音频 "+bvid+"_audio"]
	if !ok {
		t.Fatalf("_audio entry missing, titles=%v", svcDlVideoTitles(videos))
	}
	if !audio.Files[0].IsAudioOnly {
		t.Errorf("_audio.mp4 must be flagged audio only: %+v", audio.Files[0])
	}
	if audio.Files[0].SizeMB != 5.0/1024/1024 {
		t.Errorf("audio sizeMB = %v", audio.Files[0].SizeMB)
	}

	broken, ok := byTitle["[svcDl残缺前缀标题"]
	if !ok {
		t.Fatalf("bracket-less entry missing, titles=%v", svcDlVideoTitles(videos))
	}
	if broken.AuthorName != "" || broken.BVID != "" {
		t.Errorf("broken prefix parsed as %+v", broken)
	}
	if broken.Cover != "" || broken.AuthorMid != 0 {
		t.Errorf("without bvid there must be no metadata: %+v", broken)
	}

	// 搜索命中不到任何文件时返回空切片，total=0。
	none, noneTotal, err := ListDownloadedVideos("svcDl绝对不存在的名字", 1, 20)
	if err != nil || noneTotal != 0 || len(none) != 0 {
		t.Errorf("no-match search = (%v,%d,%v)", none, noneTotal, err)
	}
}

// 回归 services/download.go 的分页缺陷修复：ListDownloadedVideos 过去只把
// start/end 向 total 的上界夹紧，负数（page=0、limit<0）会算出负 start 或
// end<start，`videos[start:end]` 直接 slice bounds out of range panic。现在
// 与 HTTP 层（routers/download.go:108-113）一致地把 page/limit 钳到 >=1。
func TestSVCDLListDownloadedVideosNegativeOffset(t *testing.T) {
	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDlPanic")
	svcMkdirAll(t, dir)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	svcWrite(t, filepath.Join(dir, "svcDl分页视频.mp4"), []byte("p"))

	for _, tc := range []struct{ page, limit int }{{0, 10}, {1, -5}, {0, 0}, {-3, -2}} {
		videos, total, err := ListDownloadedVideos("svcDl分页", tc.page, tc.limit)
		if err != nil {
			t.Errorf("page=%d limit=%d: %v", tc.page, tc.limit, err)
			continue
		}
		if total != 1 || len(videos) != 1 {
			t.Errorf("page=%d limit=%d = (%d rows, total %d), want the single row", tc.page, tc.limit, len(videos), total)
		}
	}

	// 正常边界：page 越界返回空页，total 仍为 1。
	page, total, err := ListDownloadedVideos("svcDl分页", 999, 10)
	if err != nil || total != 1 || len(page) != 0 {
		t.Errorf("out-of-range page = (%d rows, total %d, err %v)", len(page), total, err)
	}
}

func TestSVCDLListDownloadedVideosSkipsUnreadableDir(t *testing.T) {
	out := GetDownloadOutputPath()
	hidden := filepath.Join(out, "svcDlHidden")
	visible := filepath.Join(out, "svcDlVisible")
	svcMkdirAll(t, hidden)
	svcMkdirAll(t, visible)
	t.Cleanup(func() {
		_ = os.Chmod(hidden, 0755)
		_ = os.RemoveAll(hidden)
		_ = os.RemoveAll(visible)
	})
	svcWrite(t, filepath.Join(hidden, "svcDl不可见视频.mp4"), []byte("h"))
	svcWrite(t, filepath.Join(visible, "svcDl可见视频.mp4"), []byte("v"))

	if err := os.Chmod(hidden, 0000); err != nil {
		t.Fatalf("chmod 000: %v", err)
	}
	if _, err := os.ReadDir(hidden); err == nil {
		// root（uid==0）下 chmod 000 依然可读，该分支无法触发，直接跳过。
		t.Skip("directory still readable after chmod 000 (running as root?)")
	}

	videos, total, err := ListDownloadedVideos("svcDl", 1, 100)
	if err != nil {
		t.Fatalf("ListDownloadedVideos: %v", err)
	}
	titles := svcDlVideoTitles(videos)
	for _, want := range []string{"svcDl可见视频"} {
		if !strings.Contains(strings.Join(titles, "|"), want) {
			t.Errorf("titles %v missing %q", titles, want)
		}
	}
	for _, unwanted := range []string{"svcDl不可见视频"} {
		if strings.Contains(strings.Join(titles, "|"), unwanted) {
			t.Errorf("titles %v should not contain %q (unreadable dir must be skipped)", titles, unwanted)
		}
	}
	if total != len(videos) {
		t.Errorf("total = %d, len(videos) = %d", total, len(videos))
	}

	if err := os.Chmod(hidden, 0755); err != nil {
		t.Fatalf("restore chmod: %v", err)
	}
	after, afterTotal, err := ListDownloadedVideos("svcDl", 1, 100)
	if err != nil {
		t.Fatalf("ListDownloadedVideos after restore: %v", err)
	}
	if afterTotal != total+1 {
		t.Errorf("total after restore = %d, want %d", afterTotal, total+1)
	}
	if !strings.Contains(strings.Join(svcDlVideoTitles(after), "|"), "svcDl不可见视频") {
		t.Errorf("hidden video still missing after chmod restore: %v", svcDlVideoTitles(after))
	}
}

func TestSVCDLDeleteDownloadedVideoByCidRemovesDirectory(t *testing.T) {
	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDlDelDir")
	svcMkdirAll(t, dir)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	matched := svcWrite(t, filepath.Join(dir, "svcDlCidDel-视频.mp4"), []byte("m"))
	sibling := svcWrite(t, filepath.Join(dir, "svcDlCidDel-音频.m4a"), []byte("s"))
	if _, err := os.Stat(matched); err != nil {
		t.Fatalf("seed files: %v", err)
	}

	if err := DeleteDownloadedVideo("svcdlciddel", true, ""); err != nil {
		t.Fatalf("delete by cid with deleteDir: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("directory %s should be removed, stat err = %v", dir, err)
	}
	for _, f := range []string{matched, sibling} {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("file %s should be gone with its directory", f)
		}
	}

	// 目录已消失后再删同一个 cid：未找到错误。
	if err := DeleteDownloadedVideo("svcdlciddel", true, ""); err == nil ||
		!strings.Contains(err.Error(), "未找到匹配的视频文件") {
		t.Errorf("second delete err = %v, want 未找到", err)
	}
}

func TestSVCDLDeleteDownloadedVideoUnreadableDir(t *testing.T) {
	out := GetDownloadOutputPath()
	dir := filepath.Join(out, "svcDlUnreadable")
	svcMkdirAll(t, dir)
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0755)
		_ = os.RemoveAll(dir)
	})
	svcWrite(t, filepath.Join(dir, "svcDlCidHidden-视频.mp4"), []byte("x"))

	if err := os.Chmod(dir, 0000); err != nil {
		t.Fatalf("chmod 000: %v", err)
	}
	if _, err := os.ReadDir(dir); err == nil {
		t.Skip("directory still readable after chmod 000 (running as root?)")
	}

	// filepath.Walk 读到不可读目录时把 err 交给回调，回调 return nil 继续，
	// 最终因为没匹配到文件而报“未找到”。
	if err := DeleteDownloadedVideo("svcdlcidhidden", false, ""); err == nil ||
		!strings.Contains(err.Error(), "未找到匹配的视频文件") {
		t.Errorf("err = %v, want 未找到匹配的视频文件", err)
	}

	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatalf("restore chmod: %v", err)
	}
	// 权限恢复后文件仍在，证明不可读期间确实一个都没删。
	hiddenFile := filepath.Join(dir, "svcDlCidHidden-视频.mp4")
	if _, err := os.Stat(hiddenFile); err != nil {
		t.Errorf("file inside unreadable dir must survive: %v", err)
	}
	if err := DeleteDownloadedVideo("svcdlcidhidden", false, ""); err != nil {
		t.Errorf("delete after restore: %v", err)
	}
	if _, err := os.Stat(hiddenFile); !os.IsNotExist(err) {
		t.Errorf("file should be deleted once readable, stat err = %v", err)
	}
}

func TestSVCDLDeleteDownloadedVideoRejectsUnresolvableDirectory(t *testing.T) {
	out := GetDownloadOutputPath()
	// directory 以 outputDir 为字符串前缀但根本不可能被 filepath.Abs 解析成
	// 真实路径时（这里用带 ~ 的相对路径验证分支不 panic），要么拒绝要么成功，
	// 不能删掉 outputDir 之外的任何东西。
	before, err := os.ReadDir(out)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if err := DeleteDownloadedVideo("", true, filepath.Join(out, "..", "svcDlNeverExists")); err == nil {
		t.Errorf("want 无效的目录路径 for path outside downloads")
	}
	after, err := os.ReadDir(out)
	if err != nil {
		t.Fatalf("read output dir after: %v", err)
	}
	if len(after) != len(before) {
		t.Errorf("output dir entries changed: %d -> %d", len(before), len(after))
	}
}

// svcDlFakeFFmpeg 把一个假 ffmpeg 放进 PATH（其余路径清空，确保绝不会调用到
// 真实二进制），脚本内容自行决定参数记录与产物。
func svcDlFakeFFmpeg(t *testing.T, script string) string {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no /bin/sh available for the fake ffmpeg stub")
	}
	bin := t.TempDir()
	wrapped := "#!/bin/sh\n" + script + "\n"
	path := filepath.Join(bin, "ffmpeg")
	if err := os.WriteFile(path, []byte(wrapped), 0755); err != nil {
		t.Fatalf("write fake ffmpeg: %v", err)
	}
	t.Setenv("PATH", bin)
	return bin
}

func svcDlReadArgsLog(t *testing.T, logPath string) []string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read ffmpeg args log: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for i := range lines {
		lines[i] = strings.Trim(lines[i], "\"")
	}
	return lines
}

func TestSVCDLRemuxToContainerMOVSuccess(t *testing.T) {
	dir := t.TempDir()
	title := "svcDlRemuxMOV"
	input := filepath.Join(dir, title+".mp4")
	output := filepath.Join(dir, title+".mov")
	logPath := filepath.Join(dir, "ffmpeg-args.txt")
	svcWrite(t, input, []byte("MP4BYTES"))

	svcDlFakeFFmpeg(t, fmt.Sprintf(`printf '%%s\n' "$@" > %q
printf 'MOVBYTES' > %q`, logPath, output))

	if err := remuxToContainer(dir, title, "BVsvcDL0200", "MOV"); err != nil {
		t.Fatalf("remuxToContainer: %v", err)
	}

	args := svcDlReadArgsLog(t, logPath)
	want := []string{"-i", input, "-c", "copy", "-y", "-tag:v", "hvc1", output}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("ffmpeg args =\n%q\nwant\n%q", args, want)
	}
	if _, err := os.Stat(input); !os.IsNotExist(err) {
		t.Errorf("source mp4 should be removed, stat = %v", err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	if string(got) != "MOVBYTES" {
		t.Errorf("output content = %q", got)
	}
}

func TestSVCDLRemuxToContainerMKVSuccess(t *testing.T) {
	dir := t.TempDir()
	title := "svcDlRemuxMKV"
	input := filepath.Join(dir, title+".mp4")
	output := filepath.Join(dir, title+".mkv")
	logPath := filepath.Join(dir, "ffmpeg-args.txt")
	svcWrite(t, input, []byte("MP4BYTES"))

	svcDlFakeFFmpeg(t, fmt.Sprintf(`printf '%%s\n' "$@" > %q
printf 'MKVBYTES' > %q`, logPath, output))

	if err := remuxToContainer(dir, title, "BVsvcDL0201", "MKV"); err != nil {
		t.Fatalf("remuxToContainer: %v", err)
	}
	args := svcDlReadArgsLog(t, logPath)
	want := []string{"-i", input, "-c", "copy", "-y", output}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("ffmpeg args = %q, want %q", args, want)
	}
	for _, a := range args {
		if a == "-tag:v" {
			t.Errorf("MKV must not carry the hvc1 tag arg")
		}
	}
	if _, err := os.Stat(output); err != nil {
		t.Errorf("mkv output missing: %v", err)
	}
	if _, err := os.Stat(input); !os.IsNotExist(err) {
		t.Errorf("source mp4 should be removed")
	}
}

func TestSVCDLRemuxToContainerFFmpegFailure(t *testing.T) {
	dir := t.TempDir()
	title := "svcDlRemuxFail"
	input := svcWrite(t, filepath.Join(dir, title+".mp4"), []byte("MP4BYTES"))
	output := filepath.Join(dir, title+".mkv")
	logPath := filepath.Join(dir, "ffmpeg-args.txt")

	svcDlFakeFFmpeg(t, fmt.Sprintf(`printf '%%s\n' "$@" > %q
echo 'svcDl ffmpeg exploded' 1>&2
exit 3`, logPath))

	err := remuxToContainer(dir, title, "BVsvcDL0202", "MKV")
	if err == nil || !strings.Contains(err.Error(), "ffmpeg remux failed") {
		t.Fatalf("err = %v, want ffmpeg remux failed", err)
	}
	if !strings.Contains(err.Error(), "svcDl ffmpeg exploded") {
		t.Errorf("error should carry combined output: %v", err)
	}
	if _, err := os.Stat(input); err != nil {
		t.Errorf("source file must be kept on failure: %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Errorf("no output file should exist on failure")
	}
}

func TestSVCDLRemuxToContainerFFmpegMissing(t *testing.T) {
	dir := t.TempDir()
	title := "svcDlRemuxNoBin"
	input := svcWrite(t, filepath.Join(dir, title+".mp4"), []byte("MP4BYTES"))

	// PATH 指向一个空目录：exec 找不到 ffmpeg，走 cmd.CombinedOutput 错误分支。
	t.Setenv("PATH", t.TempDir())

	err := remuxToContainer(dir, title, "BVsvcDL0203", "MKV")
	if err == nil || !strings.Contains(err.Error(), "ffmpeg remux failed") {
		t.Fatalf("err = %v, want remux failure when ffmpeg is absent", err)
	}
	if _, err := os.Stat(input); err != nil {
		t.Errorf("source file must be kept when ffmpeg is missing: %v", err)
	}
}

func TestSVCDLResolutionLabel360pBucket(t *testing.T) {
	// 现有用例缺 360 <= max(w,h) < 480 这一档（download.go:151）。
	cases := []struct {
		w, h int
		want string
	}{
		{380, 216, "360p"},
		{640, 360, "480p"}, // 长边 640 -> 480 档（沿用长边语义）
		{479, 270, "360p"}, // 长边 479 落在 360 <= h < 480
		{853, 480, "720p"},
		{1279, 720, "1080p"},
	}
	for _, c := range cases {
		if got := resolutionLabel(c.w, c.h); got != c.want {
			t.Errorf("resolutionLabel(%d,%d) = %q, want %q", c.w, c.h, got, c.want)
		}
	}
}

func svcDlVideoTitles(vs []DownloadedVideo) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, v.Title)
	}
	return out
}
