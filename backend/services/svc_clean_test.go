package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bilibili-history-go/utils"
)

// TestSVCCleanEmptyDirsEarlyReturn 作为本文件第一个测试运行：
// 清掉 output/images 与 output/logs 后，两个目录型清理器应直接返回 (0,0,nil)。
// 后续测试都会通过 svcMkdirAll / DownloadImage 自行重建目录，不受影响。
func TestSVCCleanEmptyDirsEarlyReturn(t *testing.T) {
	imgRoot := utils.GetOutputPath("images")
	logRoot := utils.GetOutputPath("logs")
	if err := os.RemoveAll(imgRoot); err != nil {
		t.Fatalf("remove images: %v", err)
	}
	if err := os.RemoveAll(logRoot); err != nil {
		t.Fatalf("remove logs: %v", err)
	}
	if c, s, err := cleanImageCache(); err != nil || c != 0 || s != 0 {
		t.Errorf("cleanImageCache on missing dir = (%d,%d,%v), want zeros", c, s, err)
	}
	if c, s, err := cleanOldLogs(7); err != nil || c != 0 || s != 0 {
		t.Errorf("cleanOldLogs on missing dir = (%d,%d,%v), want zeros", c, s, err)
	}
}

func TestSVCCleanDuplicates(t *testing.T) {
	// 自带前置：StartClean 用例可能已把 2032 的重复对去掉，这里再补一行，
	// 使本用例在任何执行顺序下（-shuffle=on）都有至少一对重复可删。
	dup := svcHistoryRecord("BVsvcDup0001", "svc 重复动态", "archive", "知识", "知识", 7700010, "svc作者丁", svcAt(svcYearClean, 2, 2, 12, 0), 100, 50)
	if err := svcSeedHistoryRows(svcYearClean, dup); err != nil {
		t.Fatalf("re-seed duplicate row: %v", err)
	}

	deleted, err := cleanDuplicates()
	if err != nil {
		t.Fatalf("cleanDuplicates: %v", err)
	}
	if deleted < 1 {
		t.Errorf("deleted = %d, want >= 1 (the seeded pair)", deleted)
	}
	// 去重后该 (bvid, view_at) 组只保留 MIN(id) 一行。
	var left int
	if err := svcConn().QueryRow("SELECT COUNT(*) FROM bilibili_history_2032 WHERE bvid='BVsvcDup0001' AND view_at=?",
		svcAt(svcYearClean, 2, 2, 12, 0)).Scan(&left); err != nil {
		t.Fatalf("count dup group: %v", err)
	}
	if left != 1 {
		t.Errorf("dup group rows after clean = %d, want 1", left)
	}
	deleted2, err := cleanDuplicates()
	if err != nil || deleted2 != 0 {
		t.Errorf("second run deleted=%d err=%v, want idempotent 0", deleted2, err)
	}
}

func TestSVCCleanOldHistory(t *testing.T) {
	// 用与实现相同的 cutoff 语义预先统计应删行数，避免依赖运行日期
	daysToKeep := 10
	cutoff := time.Now().AddDate(0, 0, -daysToKeep).Unix()
	tables := []string{"bilibili_history_2031", "bilibili_history_2032",
		fmt.Sprintf("bilibili_history_%d", svcCurrentYear)}
	var expected int
	conn := svcConn()
	for _, tb := range tables {
		var n int
		if err := conn.QueryRow("SELECT COUNT(*) FROM "+tb+" WHERE view_at < ?", cutoff).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", tb, err)
		}
		expected += n
	}

	deleted, err := cleanOldHistory(daysToKeep)
	if err != nil {
		t.Fatalf("cleanOldHistory: %v", err)
	}
	if deleted != expected {
		t.Errorf("deleted = %d, want %d", deleted, expected)
	}
	// 幂等：再跑一次应为 0
	if again, err := cleanOldHistory(daysToKeep); err != nil || again != 0 {
		t.Errorf("second run deleted=%d err=%v", again, err)
	}

	// 2032 沙盒：初始 5 行，cleanDuplicates 若先跑删 1 重复，本函数删 BVsvcOld0001（40 天前）
	var cnt int
	if err := conn.QueryRow("SELECT COUNT(*) FROM bilibili_history_2032").Scan(&cnt); err != nil {
		t.Fatalf("count: %v", err)
	}
	remaining := svcExpectedCleanRowsRemaining()
	if cnt != remaining {
		t.Errorf("2032 rows left = %d, want %d", cnt, remaining)
	}
}

// svcExpectedCleanRowsRemaining 依据当前库中重复对是否仍在，推导 2032 表剩余行数：
// 初始 5 行 - duplicates 去重(可能 1 或 0 行) - cleanOldHistory 删 BVsvcOld0001。
func svcExpectedCleanRowsRemaining() int {
	conn := svcConn()
	var dup int
	_ = conn.QueryRow("SELECT COUNT(*) FROM bilibili_history_2032 WHERE bvid='BVsvcDup0001' AND view_at=?",
		svcAt(svcYearClean, 2, 2, 12, 0)).Scan(&dup)
	if dup == 2 {
		return 4 // duplicates 尚未运行（乱序执行时的正常分支）
	}
	return 3
}

func TestSVCCleanImageCache(t *testing.T) {
	imgRoot := utils.GetOutputPath("images")
	sub := filepath.Join(imgRoot, "svccache")
	svcMkdirAll(t, sub)
	f1 := svcWrite(t, filepath.Join(sub, "a.jpg"), []byte("12345"))
	f2 := svcWrite(t, filepath.Join(sub, "b.jpg"), []byte("678"))

	// 统计 images 根下当前所有文件（cleanImageCache 会全部删除）
	before, err := countFiles(imgRoot)
	if err != nil {
		// 目录可能已被其它测试清空，不存在则视为 0
		if !os.IsNotExist(err) {
			t.Fatalf("count: %v", err)
		}
		before = 0
	}
	count, size, err := cleanImageCache()
	if err != nil {
		t.Fatalf("cleanImageCache: %v", err)
	}
	if count != before {
		t.Errorf("count = %d, want %d", count, before)
	}
	if size < 8 {
		t.Errorf("size = %d, want >= 8 bytes from our two files", size)
	}
	for _, f := range []string{f1, f2} {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("%s should be deleted", f)
		}
	}
	// 目录仍在（只删文件），再跑一次应为 0
	again, againSize, err := cleanImageCache()
	if err != nil || again != 0 || againSize != 0 {
		t.Errorf("second run = (%d,%d,%v), want zeros", again, againSize, err)
	}
}

func countFiles(root string) (int, error) {
	n := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			n++
		}
		return nil
	})
	return n, err
}

func TestSVCCleanOldLogs(t *testing.T) {
	logRoot := utils.GetOutputPath("logs")
	svcMkdirAll(t, logRoot)
	oldFile := svcWrite(t, filepath.Join(logRoot, "svc_old.log"), []byte("old-logs"))
	freshFile := svcWrite(t, filepath.Join(logRoot, "svc_fresh.log"), []byte("fresh"))
	stamp := time.Now().AddDate(0, 0, -30)
	if err := os.Chtimes(oldFile, stamp, stamp); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	count, size, err := cleanOldLogs(7)
	if err != nil {
		t.Fatalf("cleanOldLogs: %v", err)
	}
	if count < 1 || size < int64(len("old-logs")) {
		t.Errorf("count=%d size=%d, want at least our stale file", count, size)
	}
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("stale log should be deleted")
	}
	if _, err := os.Stat(freshFile); err != nil {
		t.Errorf("fresh log should survive: %v", err)
	}
	os.Remove(freshFile)
}

func TestSVCStartCleanGuardsAndRun(t *testing.T) {
	// 复位到 idle，然后伪造运行中状态验证互斥守卫
	setCleanStatus(CleanStatus{IsRunning: true, Status: "running"})
	if _, err := StartClean(CleanOptions{}); err == nil || !strings.Contains(err.Error(), "already running") {
		t.Errorf("guard err = %v, want already running", err)
	}

	// 真实一轮：只开图片缓存，等待后台 goroutine 收尾
	imgRoot := utils.GetOutputPath("images")
	svcMkdirAll(t, filepath.Join(imgRoot, "svcrun"))
	svcWrite(t, filepath.Join(imgRoot, "svcrun", "x.jpg"), []byte("xyz"))
	setCleanStatus(CleanStatus{IsRunning: false, Status: "idle"})

	res, err := StartClean(CleanOptions{CleanImageCache: true})
	if err != nil {
		t.Fatalf("StartClean: %v", err)
	}
	if res["status"] != "success" {
		t.Errorf("StartClean result = %v", res)
	}
	waitCleanDone(t)

	st := GetCleanStatus()
	if st.IsRunning {
		t.Fatalf("clean should have finished")
	}
	if st.Status != "completed" {
		t.Errorf("status = %q, want completed (err=%q)", st.Status, st.ErrorMessage)
	}
	if st.DeletedFiles < 1 {
		t.Errorf("deleted files = %d", st.DeletedFiles)
	}
	if st.StartTime == 0 || st.EndTime < st.StartTime {
		t.Errorf("timestamps invalid: start=%d end=%d", st.StartTime, st.EndTime)
	}

	// 全选项关闭的一轮也应正常完成
	setCleanStatus(CleanStatus{IsRunning: false, Status: "idle"})
	if _, err := StartClean(CleanOptions{CleanOldHistory: true, DaysToKeep: 0}); err != nil {
		t.Fatalf("StartClean noop: %v", err)
	}
	waitCleanDone(t)
	if st := GetCleanStatus(); st.Status != "completed" || st.CleanedRecords != 0 {
		t.Errorf("noop run status = %+v", st)
	}
}

func TestSVCStartCleanErrorBranch(t *testing.T) {
	// filepath.Walk 对普通文件根目录不会报错（反而会把它当文件删除），
	// 因此用 mode-000 子目录制造真实遍历错误。
	imgPath := utils.GetOutputPath("images")
	svcMkdirAll(t, imgPath)
	blocker := filepath.Join(imgPath, "svc_locked")
	svcMkdirAll(t, blocker)
	if err := os.Chmod(blocker, 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer func() {
		_ = os.Chmod(blocker, 0755)
		_ = os.RemoveAll(blocker)
	}()

	// cleanImageCache: Walk 回调收到 err 即中止并返回 error
	if _, _, err := cleanImageCache(); err == nil {
		t.Skip("running with privileges that bypass directory permissions; cannot produce walk error")
	}

	setCleanStatus(CleanStatus{IsRunning: false, Status: "idle"})
	if _, err := StartClean(CleanOptions{CleanImageCache: true}); err != nil {
		t.Fatalf("StartClean: %v", err)
	}
	waitCleanDone(t)
	st := GetCleanStatus()
	if st.Status != "error" || !strings.Contains(st.ErrorMessage, "图片缓存") {
		t.Errorf("expected error status, got %+v", st)
	}
	setCleanStatus(CleanStatus{IsRunning: false, Status: "idle"})
}

// TestSVCStartCleanAllOptions 一轮全开启的 StartClean，覆盖四个子任务的进度回写块：
// 去重 / 旧历史（DaysToKeep=3650 对命名空间的 2026+ 行全部保留）/ 图片缓存 / 过期日志。
func TestSVCStartCleanAllOptions(t *testing.T) {
	imgRoot := utils.GetOutputPath("images")
	sub := filepath.Join(imgRoot, "svcall")
	svcMkdirAll(t, sub)
	imgFile := svcWrite(t, filepath.Join(sub, "s.jpg"), []byte("1234567"))

	logRoot := utils.GetOutputPath("logs")
	svcMkdirAll(t, logRoot)
	oldLog := svcWrite(t, filepath.Join(logRoot, "svc_all_old.log"), []byte("0123456789"))
	stamp := time.Now().AddDate(0, 0, -30)
	if err := os.Chtimes(oldLog, stamp, stamp); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	// 顺序无关：重复行可能已被 TestSVCCleanDuplicates 清掉，也可能还在。先归零，
	// 下面的 CleanedRecords == 0 才能在任意执行顺序下成立。
	if _, err := cleanDuplicates(); err != nil {
		t.Fatalf("pre-clean duplicates: %v", err)
	}

	setCleanStatus(CleanStatus{IsRunning: false, Status: "idle"})
	if _, err := StartClean(CleanOptions{
		CleanDuplicates: true,
		CleanOldHistory: true,
		DaysToKeep:      3650,
		CleanImageCache: true,
		CleanLogs:       true,
		LogDaysToKeep:   7,
	}); err != nil {
		t.Fatalf("StartClean all: %v", err)
	}
	waitCleanDone(t)

	st := GetCleanStatus()
	if st.Status != "completed" {
		t.Errorf("status = %q, want completed (err=%q)", st.Status, st.ErrorMessage)
	}
	if st.DeletedFiles < 2 {
		t.Errorf("deleted files = %d, want >= 2 (1 image + 1 stale log)", st.DeletedFiles)
	}
	if st.FreedSpace < int64(len("1234567")+len("0123456789")) {
		t.Errorf("freed space = %d, want >= 17", st.FreedSpace)
	}
	if _, err := os.Stat(imgFile); !os.IsNotExist(err) {
		t.Errorf("image cache file should be deleted")
	}
	if _, err := os.Stat(oldLog); !os.IsNotExist(err) {
		t.Errorf("stale log should be deleted")
	}
	// 命名空间数据都在保留窗口内（2026+ vs 3650 天前），重复也已在本文件先行测试中清掉
	if st.CleanedRecords != 0 {
		t.Errorf("cleaned records = %d, want 0", st.CleanedRecords)
	}
	setCleanStatus(CleanStatus{IsRunning: false, Status: "idle"})
}

// TestSVCCleanSkipsBrokenTable 构造缺少业务列的年份表：
// cleanDuplicates / cleanOldHistory 对该表的 DELETE 会失败，应跳过并继续（返回 nil error）。
func TestSVCCleanSkipsBrokenTable(t *testing.T) {
	conn := svcConn()
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS bilibili_history_2036 (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create broken table: %v", err)
	}
	defer func() {
		if _, err := conn.Exec(`DROP TABLE IF EXISTS bilibili_history_2036`); err != nil {
			t.Errorf("drop broken table: %v", err)
		}
	}()

	// 顺序无关：先把全库重复行归零，本用例只验证损坏表的 DELETE 被跳过（计 0）
	if _, err := cleanDuplicates(); err != nil {
		t.Fatalf("pre-clean duplicates: %v", err)
	}
	if n, err := cleanDuplicates(); err != nil || n != 0 {
		t.Errorf("cleanDuplicates = (%d,%v), want (0,nil)", n, err)
	}
	// 3650 天保留窗口覆盖所有命名空间行（2026+），且损坏表被跳过
	if n, err := cleanOldHistory(3650); err != nil || n != 0 {
		t.Errorf("cleanOldHistory = (%d,%v), want (0,nil)", n, err)
	}
}

func waitCleanDone(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !GetCleanStatus().IsRunning {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("clean did not finish in 5s")
}
