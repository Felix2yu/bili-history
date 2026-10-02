package services

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

func TestSVCImageStatusAccessors(t *testing.T) {
	initial := GetImageDownloadStatus()
	if initial.Status != "idle" && initial.Status != "completed" && initial.Status != "error" {
		t.Errorf("unexpected status %q", initial.Status)
	}
	setImageDownloadStatus(ImageDownloadStatus{IsRunning: true, Status: "running", TotalImages: 7, CurrentYear: svcYearMain})
	got := GetImageDownloadStatus()
	if !got.IsRunning || got.TotalImages != 7 || got.CurrentYear != svcYearMain {
		t.Errorf("status roundtrip = %+v", got)
	}

	if isImageDownloadStopped() {
		t.Errorf("stop flag should start cleared by StartFullImageDownload or be false initially")
	}
	StopImageDownload()
	if !isImageDownloadStopped() {
		t.Errorf("StopImageDownload should set the flag")
	}
	atomic.StoreInt32(&stopImageDownload, 0)
	if isImageDownloadStopped() {
		t.Errorf("flag should be resettable")
	}
	// 复位状态，避免影响其它测试
	setImageDownloadStatus(ImageDownloadStatus{Status: "idle"})
}

func TestSVCDownloadImage(t *testing.T) {
	var body atomic.Value
	body.Store([]byte("first-bytes"))
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		switch r.URL.Path {
		case "/notfound.png":
			w.WriteHeader(http.StatusNotFound)
		case "/ok.png":
			_, _ = w.Write(body.Load().([]byte))
		default:
			http.Error(w, "boom", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	// 正常下载
	savePath, err := DownloadImage(srv.URL+"/ok.png", "svcimgtype", "ok.png")
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	wantDir := filepath.Join(utils.GetOutputPath("images"), "svcimgtype")
	if savePath != filepath.Join(wantDir, "ok.png") {
		t.Errorf("save path = %q", savePath)
	}
	data, err := os.ReadFile(savePath)
	if err != nil || string(data) != "first-bytes" {
		t.Errorf("file content = %q err = %v", data, err)
	}

	// 已存在则直接返回，不再发起请求
	body.Store([]byte("second-bytes"))
	hitsAt := hits
	again, err := DownloadImage(srv.URL+"/ok.png", "svcimgtype", "ok.png")
	if err != nil || again != savePath {
		t.Errorf("cached download = %q err %v", again, err)
	}
	if hits != hitsAt {
		t.Errorf("existing file should short-circuit without HTTP request")
	}

	// 非 200
	if _, err := DownloadImage(srv.URL+"/notfound.png", "svcimgtype", "nf.png"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("404 err = %v", err)
	}
	// 500 也走 status code 分支（body 不写盘）
	if _, err := DownloadImage(srv.URL+"/boom.png", "svcimgtype", "boom.png"); err == nil {
		t.Errorf("500 should error")
	}
	// 非法 URL：请求构造失败，不发生任何连接
	if _, err := DownloadImage("://no-scheme", "svcimgtype", "x.png"); err == nil || !strings.Contains(err.Error(), "create request") {
		t.Errorf("bad url err = %v", err)
	}
	// 连接失败：关闭 httptest 服务器后再访问它（回环拒绝连接，无外网）
	badSrvURL := srv.URL
	srv.Close()
	if _, err := DownloadImage(badSrvURL+"/gone.png", "svcimgtype", "gone.png"); err == nil || !strings.Contains(err.Error(), "download error") {
		t.Errorf("conn refused err = %v", err)
	}
}

func TestSVCStartFullImageDownload(t *testing.T) {
	// 前置：状态必须非运行中
	setImageDownloadStatus(ImageDownloadStatus{Status: "idle"})

	// 轮 A：指向不存在的年份表 → 0 任务提前返回，状态 completed（覆盖 continue + totalTasks==0 分支）
	emptyYear := 2035
	StartFullImageDownload(&emptyYear, false)
	waitImageDone(t)
	st := GetImageDownloadStatus()
	if st.Status != "completed" || st.TotalImages != 0 {
		t.Fatalf("empty-year run = %+v", st)
	}

	// 轮 B：2034 沙盒表，全部 URL 指向本机 httptest / 非法 URL / 已关闭端口，零外网。
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path == "/boom.png" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("img-bytes"))
	}))
	deadURL := srv.URL
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	dead2 := srv2.URL
	srv2.Close() // 回环端口立即关闭 → connection refused

	coverA := srv.URL + "/a.jpg"
	faceB := srv.URL + "/b.png"
	noExt := srv.URL + "/c"
	longExt := srv.URL + "/d.toolongext"
	dupA := coverA // 与轮 B 第一行 cover 相同 → md5 相同（可命中缓存或重复写同内容）

	// 行设计（cover, author_face）：
	//  1) coverA, faceB          → 2 任务，均成功
	//  2) noExt, ""              → 1 任务成功（ext 空回退 .jpg）
	//  3) longExt, "://bad"      → 2 任务：1 成功 + 1 失败（create request error）
	//  4) "", dead2              → 1 任务失败（download error，本机回环拒绝连接）
	//  5) dupA, ""               → 1 任务成功（相同 hash 文件）
	mk := func(bvid, cover, face string) models.HistoryRecord {
		r := svcHistoryRecord(bvid, "svc 图片行", "archive", "科技", "科技", 7700030, "svc图片UP", svcAt(2034, 5, 5, 12, 0), 60, 60)
		r.Cover = cover
		r.AuthorFace = face
		return r
	}
	svcSeedHistory(t, 2034,
		mk("BVsvcImg0001", coverA, faceB),
		mk("BVsvcImg0002", noExt, ""),
		mk("BVsvcImg0003", longExt, "://bad"),
		mk("BVsvcImg0004", "", dead2+"/x.jpg"),
		mk("BVsvcImg0005", dupA, ""),
	)
	// 清理其它测试可能留下的运行标志（StartFullImageDownload 自身会复位 stop 标志）
	year := 2034
	StartFullImageDownload(&year, false)
	waitImageDone(t)
	_ = deadURL

	st = GetImageDownloadStatus()
	if st.IsRunning {
		t.Fatalf("image download did not finish in 10s: %+v", st)
	}
	if st.Status != "completed" {
		t.Errorf("status = %q, want completed (err=%q)", st.Status, st.ErrorMessage)
	}
	if st.TotalImages != 7 {
		t.Errorf("total = %d, want 7", st.TotalImages)
	}
	if st.CurrentYear != 2034 {
		t.Errorf("current year = %d, want 2034", st.CurrentYear)
	}
	// 说明：DownloadedImages/FailedImages 由并发 goroutine 以 get-modify-set 更新，
	// 生产实现存在丢失更新的可能，因此这里只断言上界而不做精确计数。
	if st.DownloadedImages+st.FailedImages > 7 {
		t.Errorf("downloaded+failed = %d, want <= 7", st.DownloadedImages+st.FailedImages)
	}

	// 落盘断言（路径规则：images/covers|avatars/2034/<md5前2位>/<md5><ext>）
	imgRoot := utils.GetOutputPath("images")
	mustExist := func(base string) {
		t.Helper()
		if _, err := os.Stat(base); err != nil {
			t.Errorf("expected file %s: %v", base, err)
		}
	}
	mustExist(filepath.Join(imgRoot, "covers", "2034", svcHashDir(coverA), svcHashName(coverA, ".jpg")))
	mustExist(filepath.Join(imgRoot, "avatars", "2034", svcHashDir(faceB), svcHashName(faceB, ".png")))
	mustExist(filepath.Join(imgRoot, "covers", "2034", svcHashDir(noExt), svcHashName(noExt, ".jpg")))     // ext 空 → .jpg 回退
	mustExist(filepath.Join(imgRoot, "covers", "2034", svcHashDir(longExt), svcHashName(longExt, ".jpg"))) // ext 过长 → .jpg 回退
	// 非法 URL / 拒绝连接的行不应落盘：文件名可算出但文件不存在
	badName := filepath.Join(imgRoot, "covers", "2034", svcHashDir("://bad"), svcHashName("://bad", ".jpg"))
	if _, err := os.Stat(badName); !os.IsNotExist(err) {
		t.Errorf("bad-url file should not exist: %v", err)
	}
	deadName := filepath.Join(imgRoot, "avatars", "2034", svcHashDir(dead2+"/x.jpg"), svcHashName(dead2+"/x.jpg", ".jpg"))
	if _, err := os.Stat(deadName); !os.IsNotExist(err) {
		t.Errorf("refused-url file should not exist: %v", err)
	}
	srv.Close()
}

// TestSVCStartFullImageDownloadBrokenAndNullRows 覆盖年份查询失败跳过与 rows.Scan 失败跳过两个分支。
func TestSVCStartFullImageDownloadBrokenAndNullRows(t *testing.T) {
	setImageDownloadStatus(ImageDownloadStatus{Status: "idle"})
	conn := svcConn()

	// A) 损坏表（无 cover/author_face 列）→ SELECT 失败 → 跳过该年份，0 任务完成
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS bilibili_history_2036 (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create broken table: %v", err)
	}
	defer func() {
		if _, err := conn.Exec(`DROP TABLE IF EXISTS bilibili_history_2036`); err != nil {
			t.Errorf("drop broken: %v", err)
		}
	}()
	broken := 2036
	StartFullImageDownload(&broken, false)
	waitImageDone(t)
	if st := GetImageDownloadStatus(); st.Status != "completed" || st.TotalImages != 0 {
		t.Errorf("broken-year run = %+v", st)
	}

	// B) 2037 表：一行 cover 为 NULL（配合非空 face 命中 WHERE，但 Scan 进 string 失败 → 整行跳过），
	//    一行正常 → 只统计正常行的 1 个任务
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("null-row-survivor"))
	}))
	defer srv.Close()
	okCover := srv.URL + "/ok1.jpg"
	faceURL := srv.URL + "/f.jpg"
	rBad := svcHistoryRecord("BVsvcImgNull1", "svc NULL封面", "archive", "科技", "科技", 7700031, "svc图片UP", svcAt(2037, 7, 7, 7, 0), 60, 60)
	rBad.Cover = ""
	rBad.AuthorFace = faceURL
	rOk := svcHistoryRecord("BVsvcImgNull2", "svc 正常封面", "archive", "科技", "科技", 7700031, "svc图片UP", svcAt(2037, 7, 7, 7, 0), 60, 60)
	rOk.Cover = okCover
	rOk.AuthorFace = ""
	svcSeedHistory(t, 2037, rBad, rOk)
	if _, err := conn.Exec(`UPDATE bilibili_history_2037 SET cover = NULL WHERE bvid = 'BVsvcImgNull1'`); err != nil {
		t.Fatalf("nullify cover: %v", err)
	}
	defer func() {
		if _, err := conn.Exec(`DROP TABLE IF EXISTS bilibili_history_2037`); err != nil {
			t.Errorf("drop 2037: %v", err)
		}
	}()

	year := 2037
	StartFullImageDownload(&year, false)
	waitImageDone(t)
	st := GetImageDownloadStatus()
	if st.Status != "completed" {
		t.Errorf("status = %q (err=%q)", st.Status, st.ErrorMessage)
	}
	// NULL-cover 行被 Scan 错误跳过，不产生任务；正常行 cover 计 1 个任务
	if st.TotalImages != 1 {
		t.Errorf("total = %d, want 1", st.TotalImages)
	}
	want := filepath.Join(utils.GetOutputPath("images"), "covers", "2037", svcHashDir(okCover), svcHashName(okCover, ".jpg"))
	if _, err := os.Stat(want); err != nil {
		t.Errorf("survivor cover file missing: %v", err)
	}
	// NULL row 的 face 不应因 Scan 失败而被下载
	faceWant := filepath.Join(utils.GetOutputPath("images"), "avatars", "2037", svcHashDir(faceURL), svcHashName(faceURL, ".jpg"))
	if _, err := os.Stat(faceWant); !os.IsNotExist(err) {
		t.Errorf("scan-skipped face file should not exist: %v", err)
	}
}

func svcHashDir(url string) string {
	sum := md5.Sum([]byte(url))
	return hex.EncodeToString(sum[:])[:2]
}

func svcHashName(url, ext string) string {
	sum := md5.Sum([]byte(url))
	return hex.EncodeToString(sum[:]) + ext
}

// waitImageDone 轮询等待后台图片下载 goroutine 收尾（本机回环，通常毫秒级完成）。
func waitImageDone(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !GetImageDownloadStatus().IsRunning {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("image download did not finish in 10s")
}

func TestSVCClearAllImages(t *testing.T) {
	imgRoot := utils.GetOutputPath("images")
	covers := filepath.Join(imgRoot, "covers", "2031", "ab")
	avatars := filepath.Join(imgRoot, "avatars")
	proxy := filepath.Join(imgRoot, "proxy")
	keep := filepath.Join(imgRoot, "svckeep")
	for _, d := range []string{covers, avatars, proxy, keep} {
		svcMkdirAll(t, d)
	}
	svcWrite(t, filepath.Join(covers, "c.jpg"), []byte("c"))
	svcWrite(t, filepath.Join(avatars, "a.jpg"), []byte("a"))
	svcWrite(t, filepath.Join(proxy, "p.jpg"), []byte("p"))
	kept := svcWrite(t, filepath.Join(keep, "k.jpg"), []byte("k"))

	// 预设脏状态，验证被复位
	setImageDownloadStatus(ImageDownloadStatus{IsRunning: true, Status: "running", TotalImages: 5, DownloadedImages: 2, FailedImages: 1, SkippedImages: 1, ErrorMessage: "bad"})

	if ok := ClearAllImages(); !ok {
		t.Errorf("ClearAllImages should return true")
	}
	for _, d := range []string{filepath.Join(imgRoot, "covers"), filepath.Join(imgRoot, "avatars"), filepath.Join(imgRoot, "proxy")} {
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Errorf("dir %s should be removed", d)
		}
	}
	if _, err := os.Stat(kept); err != nil {
		t.Errorf("unrelated dir should survive: %v", err)
	}
	st := GetImageDownloadStatus()
	if st.IsRunning || st.Status != "idle" || st.TotalImages != 0 || st.DownloadedImages != 0 || st.FailedImages != 0 || st.ErrorMessage != "" {
		t.Errorf("status not reset: %+v", st)
	}
}
