package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"
)

type ImageDownloadStatus struct {
	IsRunning        bool   `json:"is_running"`
	Status           string `json:"status"`
	TotalImages      int    `json:"total_images"`
	DownloadedImages int    `json:"downloaded_images"`
	FailedImages     int    `json:"failed_images"`
	SkippedImages    int    `json:"skipped_images"`
	ErrorMessage     string `json:"error_message,omitempty"`
	StartTime        int64  `json:"start_time,omitempty"`
	CurrentYear      int    `json:"current_year,omitempty"`
}

var (
	imageDownloadStatus = ImageDownloadStatus{
		IsRunning: false,
		Status:    "idle",
	}
	imageDownloadMutex sync.Mutex
	stopImageDownload  int32
)

func GetImageDownloadStatus() ImageDownloadStatus {
	imageDownloadMutex.Lock()
	defer imageDownloadMutex.Unlock()
	return imageDownloadStatus
}

func setImageDownloadStatus(status ImageDownloadStatus) {
	imageDownloadMutex.Lock()
	defer imageDownloadMutex.Unlock()
	imageDownloadStatus = status
}

func StopImageDownload() {
	atomic.StoreInt32(&stopImageDownload, 1)
}

func isImageDownloadStopped() bool {
	return atomic.LoadInt32(&stopImageDownload) == 1
}

func DownloadImage(url, imageType, filename string) (string, error) {
	outputPath := utils.GetOutputPath("images")
	saveDir := filepath.Join(outputPath, imageType)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("create directory error: %w", err)
	}

	savePath := filepath.Join(saveDir, filename)

	if _, err := os.Stat(savePath); err == nil {
		return savePath, nil
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com")

	cfg := config.GetConfig()
	if cfg != nil && cfg.SESSDATA != "" {
		req.AddCookie(&http.Cookie{Name: "SESSDATA", Value: cfg.SESSDATA})
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body error: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "", fmt.Errorf("write file error: %w", err)
	}

	return savePath, nil
}

func GetLocalImagePath(url, imageType string) string {
	if url == "" {
		return ""
	}

	outputPath := utils.GetOutputPath("images")
	saveDir := filepath.Join(outputPath, imageType)

	filename := filepath.Base(url)
	if idx := stringsIndex(filename, '?'); idx >= 0 {
		filename = filename[:idx]
	}

	savePath := filepath.Join(saveDir, filename)

	if _, err := os.Stat(savePath); err == nil {
		return fmt.Sprintf("/images/%s/%s", imageType, filename)
	}

	return url
}

func stringsIndex(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

type ImageDownloadTask struct {
	URL       string
	ImageType string
	Filename  string
	Year      int
}

func BatchDownloadImages(tasks []ImageDownloadTask, concurrency int) {
	status := GetImageDownloadStatus()
	if status.IsRunning {
		return
	}

	newStatus := ImageDownloadStatus{
		IsRunning:   true,
		Status:      "running",
		TotalImages: len(tasks),
		StartTime:   time.Now().Unix(),
	}
	setImageDownloadStatus(newStatus)
	atomic.StoreInt32(&stopImageDownload, 0)

	go func() {
		defer func() {
			status := GetImageDownloadStatus()
			status.IsRunning = false
			if status.ErrorMessage != "" {
				status.Status = "error"
			} else {
				status.Status = "completed"
			}
			setImageDownloadStatus(status)
		}()

		if concurrency <= 0 {
			concurrency = 5
		}

		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, task := range tasks {
			if isImageDownloadStopped() {
				break
			}

			wg.Add(1)
			sem <- struct{}{}

			go func(t ImageDownloadTask) {
				defer wg.Done()
				defer func() { <-sem }()

				if isImageDownloadStopped() {
					return
				}

				_, err := DownloadImage(t.URL, t.ImageType, t.Filename)

				status := GetImageDownloadStatus()
				if err != nil {
					status.FailedImages++
				} else {
					status.DownloadedImages++
				}
				if t.Year > 0 {
					status.CurrentYear = t.Year
				}
				setImageDownloadStatus(status)
			}(task)
		}

		wg.Wait()
	}()
}

func StartFullImageDownload(year *int, useSessdata bool) {
	status := GetImageDownloadStatus()
	if status.IsRunning {
		return
	}

	atomic.StoreInt32(&stopImageDownload, 0)

	newStatus := ImageDownloadStatus{
		IsRunning: true,
		Status:    "running",
		StartTime: time.Now().Unix(),
	}
	setImageDownloadStatus(newStatus)

	go func() {
		defer func() {
			status := GetImageDownloadStatus()
			status.IsRunning = false
			if status.ErrorMessage != "" {
				status.Status = "error"
			} else {
				status.Status = "completed"
			}
			setImageDownloadStatus(status)
		}()

		db := database.GetSQLiteDB()
		if db == nil {
			status := GetImageDownloadStatus()
			status.ErrorMessage = "数据库未初始化"
			setImageDownloadStatus(status)
			return
		}

		years := []int{}
		if year != nil {
			years = append(years, *year)
		} else {
			allYears, err := db.GetAvailableYears()
			if err != nil {
				status := GetImageDownloadStatus()
				status.ErrorMessage = err.Error()
				setImageDownloadStatus(status)
				return
			}
			years = allYears
		}

		totalTasks := 0
		allTasks := []ImageDownloadTask{}

		for _, y := range years {
			if isImageDownloadStopped() {
				break
			}

			tableName := fmt.Sprintf("bilibili_history_%d", y)
			exists, _ := db.TableExists(tableName)
			if !exists {
				continue
			}

			params := database.HistoryQueryParams{
				Page:      1,
				Size:      99999,
				SortOrder: 0,
			}
			pagedResponse, _, err := database.GetHistoryPage(params)
			if err != nil || pagedResponse == nil {
				continue
			}
			records, ok := pagedResponse.Records.([]map[string]interface{})
			if !ok {
				continue
			}

			status := GetImageDownloadStatus()
			status.CurrentYear = y
			setImageDownloadStatus(status)

			for _, record := range records {
				if isImageDownloadStopped() {
					break
				}

				cover, _ := record["cover"].(string)
				authorFace, _ := record["author_face"].(string)

				if cover != "" {
					hash := md5.Sum([]byte(cover))
					hashStr := hex.EncodeToString(hash[:])
					ext := filepath.Ext(cover)
					if ext == "" || len(ext) > 5 {
						ext = ".jpg"
					}
					subDir := hashStr[:2]
					saveDir := fmt.Sprintf("covers/%d/%s", y, subDir)
					allTasks = append(allTasks, ImageDownloadTask{
						URL:       cover,
						ImageType: saveDir,
						Filename:  hashStr + ext,
						Year:      y,
					})
					totalTasks++
				}

				if authorFace != "" {
					hash := md5.Sum([]byte(authorFace))
					hashStr := hex.EncodeToString(hash[:])
					ext := filepath.Ext(authorFace)
					if ext == "" || len(ext) > 5 {
						ext = ".jpg"
					}
					subDir := hashStr[:2]
					saveDir := fmt.Sprintf("avatars/%d/%s", y, subDir)
					allTasks = append(allTasks, ImageDownloadTask{
						URL:       authorFace,
						ImageType: saveDir,
						Filename:  hashStr + ext,
						Year:      y,
					})
					totalTasks++
				}
			}
		}

		status = GetImageDownloadStatus()
		status.TotalImages = totalTasks
		setImageDownloadStatus(status)

		concurrency := 5
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, task := range allTasks {
			if isImageDownloadStopped() {
				break
			}

			wg.Add(1)
			sem <- struct{}{}

			go func(t ImageDownloadTask) {
				defer wg.Done()
				defer func() { <-sem }()

				if isImageDownloadStopped() {
					return
				}

				_, err := DownloadImage(t.URL, t.ImageType, t.Filename)

				s := GetImageDownloadStatus()
				if err != nil {
					s.FailedImages++
				} else {
					s.DownloadedImages++
				}
				setImageDownloadStatus(s)
			}(task)
		}

		wg.Wait()
	}()
}

func ClearAllImages() bool {
	outputPath := utils.GetOutputPath("images")

	dirs := []string{
		filepath.Join(outputPath, "covers"),
		filepath.Join(outputPath, "avatars"),
		filepath.Join(outputPath, "proxy"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			os.RemoveAll(dir)
		}
	}

	status := GetImageDownloadStatus()
	status.IsRunning = false
	status.Status = "idle"
	status.TotalImages = 0
	status.DownloadedImages = 0
	status.FailedImages = 0
	status.SkippedImages = 0
	status.ErrorMessage = ""
	setImageDownloadStatus(status)

	return true
}

