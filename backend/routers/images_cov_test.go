package routers

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bilibili-history-go/services"
	"bilibili-history-go/utils"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Coverage tests for backend/routers/images.go.
//
// Files are written only under utils.GetOutputPath("images") (which resolves to
// <testCWD>/output/images, i.e. the TestMain temp dir) or removed again in
// t.Cleanup. Remote image fetches go through services.DownloadImage, which uses
// http.DefaultTransport, so msUseAllTransport answers them from an in-process
// stub and keeps the suite offline.
// ---------------------------------------------------------------------------

func msImageAPI(t *testing.T) *gin.Engine {
	t.Helper()
	return newAPI(t, RegisterImageRoutes)
}

// msImagesRoot is the on-disk root used by every image handler.
func msImagesRoot() string { return utils.GetOutputPath("images") }

// msWriteImage writes content to path under the images root and registers a
// cleanup that removes just that file.
func msWriteImage(t *testing.T, rel, content string) string {
	t.Helper()
	full := filepath.Join(msImagesRoot(), rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", full, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
	t.Cleanup(func() { _ = os.Remove(full) })
	return full
}

func msMD5(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func TestMSImagesStatusEmptyThenPopulated(t *testing.T) {
	e := msImageAPI(t)

	// Baseline: just assert the envelope shape (counts may be zero or reflect
	// files other image tests created; we only pin the structure here).
	resp := expectStatus(t, e, http.MethodGet, "/api/images/status", "", 200, "success")
	dm := resp.dataMap(t)
	if dm["is_downloading"] != false {
		t.Fatalf("is_downloading = %v, want false", dm["is_downloading"])
	}
	if _, ok := dm["covers"].(map[string]interface{}); !ok {
		t.Fatalf("covers missing in %v", dm)
	}

	// Populate the covers dir so the "found path" + walk branch is exercised.
	msWriteImage(t, filepath.Join("covers", "2009", "ab", "msstatusab.jpg"), "x")
	w := doRaw(t, e, http.MethodGet, "/api/images/status", "")
	dm2 := msRawMap(t, w)["data"].(map[string]interface{})
	covers, _ := dm2["covers"].(map[string]interface{})
	if covers["downloaded"].(float64) < 1 {
		t.Fatalf("covers.downloaded = %v, want >= 1", covers["downloaded"])
	}
}

func TestMSImagesGetLocalImage(t *testing.T) {
	e := msImageAPI(t)

	// invalid image type -> 400
	bad := doRaw(t, e, http.MethodGet, "/api/images/local/bad/mslocala1b.jpg", "")
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("bad type code = %d, want 400", bad.Code)
	}

	// hash shorter than 2 chars -> 404 (findLocalImage returns "")
	short := doRaw(t, e, http.MethodGet, "/api/images/local/covers/x", "")
	if short.Code != http.StatusNotFound {
		t.Fatalf("short hash code = %d, want 404", short.Code)
	}

	// missing file -> 404
	miss := doRaw(t, e, http.MethodGet, "/api/images/local/covers/msmissing01", "")
	if miss.Code != http.StatusNotFound {
		t.Fatalf("missing file code = %d, want 404", miss.Code)
	}

	// present under a year dir -> 200 with content
	msWriteImage(t, filepath.Join("covers", "2009", "ms", "mslocala1b.jpg"), "YEARIMG")
	ok := doRaw(t, e, http.MethodGet, "/api/images/local/covers/mslocala1b", "")
	if ok.Code != http.StatusOK || ok.Body.String() != "YEARIMG" {
		t.Fatalf("year-dir local image code=%d body=%q", ok.Code, ok.Body.String())
	}

	// present only under the root subdir (no year match) -> fallback branch
	msWriteImage(t, filepath.Join("avatars", "ms", "msrootz9.jpg"), "ROOTIMG")
	root := doRaw(t, e, http.MethodGet, "/api/images/local/avatars/msrootz9", "")
	if root.Code != http.StatusOK || root.Body.String() != "ROOTIMG" {
		t.Fatalf("root-subdir local image code=%d body=%q", root.Code, root.Body.String())
	}
}

func TestMSImagesProxyEmptyURL(t *testing.T) {
	e := msImageAPI(t)
	w := doRaw(t, e, http.MethodGet, "/api/images/proxy", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty url code = %d, want 400", w.Code)
	}
	if m := msRawMap(t, w); m["error"] != "URL不能为空" {
		t.Fatalf("empty url resp = %v", m)
	}
}

func TestMSImagesProxyCacheHit(t *testing.T) {
	e := msImageAPI(t)
	// http:// is upgraded to https:// before hashing, so the cache key is the
	// md5 of the upgraded URL.
	upgraded := "https://msproxy.test/a.png"
	hash := msMD5(upgraded)
	rel := filepath.Join("proxy", hash[:2], hash+".png")
	msWriteImage(t, rel, "CACHEDIMG")

	w := doRaw(t, e, http.MethodGet, "/api/images/proxy?url=http://msproxy.test/a.png", "")
	if w.Code != http.StatusOK || w.Body.String() != "CACHEDIMG" {
		t.Fatalf("proxy cache-hit code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestMSImagesProxyDownloadSuccess(t *testing.T) {
	e := msImageAPI(t)
	payload := []byte("DOWNLOADEDIMG-bytes")
	// A long extension (len > 5) exercises the ".jpg" fallback in proxyImage.
	upgraded := "https://msproxy.test/toolong.averylongext"
	hash := msMD5(upgraded)

	// Make sure no stale cache exists for this key.
	_ = os.Remove(filepath.Join(msImagesRoot(), "proxy", hash[:2], hash+".jpg"))

	msUseAllTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPBytes(200, payload), nil
	})
	w := doRaw(t, e, http.MethodGet, "/api/images/proxy?url="+upgraded, "")
	if w.Code != http.StatusOK {
		t.Fatalf("proxy download code=%d body=%q", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != string(payload) {
		t.Fatalf("proxy download body=%q, want %q", got, string(payload))
	}
	// The file must now be cached on disk under the ".jpg" fallback extension.
	if _, err := os.Stat(filepath.Join(msImagesRoot(), "proxy", hash[:2], hash+".jpg")); err != nil {
		t.Fatalf("expected cached proxy file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(filepath.Join(msImagesRoot(), "proxy", hash[:2], hash+".jpg"))
	})
}

func TestMSImagesProxyDownloadError(t *testing.T) {
	e := msImageAPI(t)
	upgraded := "https://msproxy.test/fail.png"
	hash := msMD5(upgraded)
	_ = os.Remove(filepath.Join(msImagesRoot(), "proxy", hash[:2], hash+".png"))

	msUseAllTransport(t, msHTTPNetworkError("image dial refused"))
	w := doRaw(t, e, http.MethodGet, "/api/images/proxy?url="+upgraded, "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("proxy download-error code=%d, want 500", w.Code)
	}
	if m := msRawMap(t, w); m["error"] == nil {
		t.Fatalf("proxy download-error resp = %v", m)
	}
}

func TestMSImagesClear(t *testing.T) {
	e := msImageAPI(t)
	// Create a file so the clear has something to remove.
	msWriteImage(t, filepath.Join("covers", "2009", "zz", "msclearzz.jpg"), "z")

	resp := expectStatus(t, e, http.MethodPost, "/api/images/clear", "", 200, "success")
	dm := resp.dataMap(t)
	if dm["message"] != "已清空所有图片和下载状态" {
		t.Fatalf("clear resp = %v", dm)
	}
	// ClearAllImages always reports success (the error branch is unreachable).
	if _, err := os.Stat(filepath.Join(msImagesRoot(), "covers")); !os.IsNotExist(err) {
		t.Fatalf("covers dir should be gone after clear")
	}
}

func TestMSImagesStartAndStop(t *testing.T) {
	// Seed a single 2009 row whose cover/face point at an offline host that
	// msUseAllTransport answers locally.
	ensureYear(t, 2009)
	insertHistory(t, 2009, map[string]interface{}{
		"bvid":        uniqueBvid("BVMS", 901),
		"view_at":     viewAt(t, 2009, time.March, 3, 9, 0),
		"business":    "archive",
		"title":       "MS 图片行",
		"cover":       "https://msimg.test/cover.jpg",
		"author_face": "https://msimg.test/face.jpg",
	})

	e := msImageAPI(t)

	// Any non-loopback GET returns a fixed PNG payload.
	msUseAllTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPBytes(200, []byte("IMG")), nil
	})

	resp := expectStatus(t, e, http.MethodPost, "/api/images/start?year=2009", "", 200, "success")
	dm := resp.dataMap(t)
	if dm["message"] != "开始下载2009年的图片" {
		t.Fatalf("start resp message = %v", dm["message"])
	}

	// Wait until the background download actually writes both images to disk.
	// This is the reliable completion signal (the handler spawns nested
	// goroutines) and keeps msUseAllTransport installed for the whole download.
	coverHash := msMD5("https://msimg.test/cover.jpg")
	faceHash := msMD5("https://msimg.test/face.jpg")
	coverPath := filepath.Join(msImagesRoot(), "covers", "2009", coverHash[:2], coverHash+".jpg")
	facePath := filepath.Join(msImagesRoot(), "avatars", "2009", faceHash[:2], faceHash+".jpg")
	msWaitFor(t, 10*time.Second, func() bool {
		if _, err := os.Stat(coverPath); err != nil {
			return false
		}
		if _, err := os.Stat(facePath); err != nil {
			return false
		}
		return true
	}, "cover and face images to be downloaded")

	st := services.GetImageDownloadStatus()
	if st.DownloadedImages < 1 {
		t.Fatalf("expected at least one downloaded image, status = %+v", st)
	}

	// The cover and avatar files must exist on disk.
	if _, err := os.Stat(coverPath); err != nil {
		t.Fatalf("expected cover file at %s: %v", coverPath, err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(msImagesRoot(), "covers")) })

	// stop just flips the abort flag and answers.
	stop := expectStatus(t, e, http.MethodPost, "/api/images/stop", "", 200, "success")
	if stop.dataMap(t)["message"] != "已停止图片下载" {
		t.Fatalf("stop resp = %v", stop.dataMap(t))
	}
}
