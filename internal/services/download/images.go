package download

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bili-history/internal/config"
)

// ImageDownloader handles downloading and caching images.
type ImageDownloader struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewImageDownloader creates a new image downloader.
func NewImageDownloader() (*ImageDownloader, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return &ImageDownloader{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// DownloadImage downloads an image and saves it locally.
func (d *ImageDownloader) DownloadImage(url, imageType string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL")
	}

	// Generate local path based on URL hash
	hash := md5.Sum([]byte(url))
	filename := fmt.Sprintf("%x", hash)

	// Determine extension
	ext := ".jpg"
	if strings.Contains(url, ".png") {
		ext = ".png"
	} else if strings.Contains(url, ".gif") {
		ext = ".gif"
	}

	// Build local path
	localDir := config.GetOutputPath("images", imageType)
	os.MkdirAll(localDir, 0755)
	localPath := filepath.Join(localDir, filename+ext)

	// Check if already downloaded
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Download
	resp, err := d.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("write file failed: %w", err)
	}

	return localPath, nil
}

// DownloadCover downloads a video cover image.
func (d *ImageDownloader) DownloadCover(url string) (string, error) {
	return d.DownloadImage(url, "covers")
}

// DownloadAvatar downloads a user avatar image.
func (d *ImageDownloader) DownloadAvatar(url string) (string, error) {
	return d.DownloadImage(url, "avatars")
}

// BatchDownload downloads multiple images concurrently.
func (d *ImageDownloader) BatchDownload(urls []string, imageType string) map[string]string {
	results := make(map[string]string)
	type result struct {
		url  string
		path string
		err  error
	}

	ch := make(chan result, len(urls))
	sem := make(chan struct{}, 10) // Limit concurrency

	for _, url := range urls {
		sem <- struct{}{}
		go func(u string) {
			defer func() { <-sem }()
			path, err := d.DownloadImage(u, imageType)
			ch <- result{u, path, err}
		}(url)
	}

	for i := 0; i < len(urls); i++ {
		r := <-ch
		if r.err != nil {
			log.Printf("Warning: failed to download %s: %v", r.url, r.err)
		} else {
			results[r.url] = r.path
		}
	}

	return results
}

// GetLocalPath returns the local path for a given URL if it exists.
func (d *ImageDownloader) GetLocalPath(url, imageType string) string {
	if url == "" {
		return ""
	}
	hash := md5.Sum([]byte(url))
	filename := fmt.Sprintf("%x", hash)

	ext := ".jpg"
	if strings.Contains(url, ".png") {
		ext = ".png"
	} else if strings.Contains(url, ".gif") {
		ext = ".gif"
	}

	localPath := filepath.Join(config.GetOutputPath("images", imageType), filename+ext)
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	return ""
}
