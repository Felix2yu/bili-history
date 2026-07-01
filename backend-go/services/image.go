package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bilibili-history-go/utils"
)

type ImageDownloadStatus struct {
	IsRunning        bool   `json:"is_running"`
	Status           string `json:"status"`
	TotalImages      int    `json:"total_images"`
	DownloadedImages int    `json:"downloaded_images"`
	FailedImages     int    `json:"failed_images"`
	ErrorMessage     string `json:"error_message,omitempty"`
	StartTime        int64  `json:"start_time,omitempty"`
}

var (
	imageDownloadStatus = ImageDownloadStatus{
		IsRunning: false,
		Status:    "idle",
	}
	imageDownloadMutex sync.Mutex
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
	if idx := len(filename) - 1; idx >= 0 && filename[idx] == '?' {
		filename = filename[:idx]
	}

	savePath := filepath.Join(saveDir, filename)

	if _, err := os.Stat(savePath); err == nil {
		return fmt.Sprintf("/images/%s/%s", imageType, filename)
	}

	return url
}

type ImageDownloadTask struct {
	URL       string
	ImageType string
	Filename  string
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
			wg.Add(1)
			sem <- struct{}{}

			go func(t ImageDownloadTask) {
				defer wg.Done()
				defer func() { <-sem }()

				_, err := DownloadImage(t.URL, t.ImageType, t.Filename)

				status := GetImageDownloadStatus()
				if err != nil {
					status.FailedImages++
				} else {
					status.DownloadedImages++
				}
				setImageDownloadStatus(status)
			}(task)
		}

		wg.Wait()
	}()
}
