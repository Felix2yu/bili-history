package download

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"bili-history/internal/config"
)

// Progress represents download progress.
type Progress struct {
	Percent    float64 `json:"percent"`
	Speed      string  `json:"speed"`
	ETA        string  `json:"eta"`
	Status     string  `json:"status"`
	Filename   string  `json:"filename"`
}

// DownloadManager manages video/audio downloads via yutto.
type DownloadManager struct {
	cfg         *config.Config
	activeDownloads map[string]context.CancelFunc
	mu          sync.RWMutex
}

// NewDownloadManager creates a new download manager.
func NewDownloadManager() (*DownloadManager, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return &DownloadManager{
		cfg:             cfg,
		activeDownloads: make(map[string]context.CancelFunc),
	}, nil
}

// DownloadVideo downloads a video using yutto with progress streaming.
func (dm *DownloadManager) DownloadVideo(ctx context.Context, url string, progressCh chan<- Progress) error {
	args := dm.buildArgs(url)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track active download
	dm.mu.Lock()
	dm.activeDownloads[url] = cancel
	dm.mu.Unlock()
	defer func() {
		dm.mu.Lock()
		delete(dm.activeDownloads, url)
		dm.mu.Unlock()
	}()

	cmd := exec.CommandContext(ctx, "yutto", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start yutto: %w", err)
	}

	// Parse progress from stdout
	go dm.parseProgress(stdout, progressCh)
	// Also read stderr for errors
	go dm.parseProgress(stderr, progressCh)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			progressCh <- Progress{Status: "cancelled"}
			return nil
		}
		return fmt.Errorf("yutto failed: %w", err)
	}

	progressCh <- Progress{Percent: 100, Status: "completed"}
	return nil
}

// DownloadAudio downloads audio only.
func (dm *DownloadManager) DownloadAudio(ctx context.Context, url string, progressCh chan<- Progress) error {
	args := dm.buildArgs(url)
	args = append(args, "--audio-only")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dm.mu.Lock()
	dm.activeDownloads[url] = cancel
	dm.mu.Unlock()
	defer func() {
		dm.mu.Lock()
		delete(dm.activeDownloads, url)
		dm.mu.Unlock()
	}()

	cmd := exec.CommandContext(ctx, "yutto", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start yutto: %w", err)
	}

	go dm.parseProgress(stdout, progressCh)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			progressCh <- Progress{Status: "cancelled"}
			return nil
		}
		return fmt.Errorf("yutto failed: %w", err)
	}

	progressCh <- Progress{Percent: 100, Status: "completed"}
	return nil
}

// CancelDownload cancels an active download.
func (dm *DownloadManager) CancelDownload(url string) bool {
	dm.mu.RLock()
	cancel, ok := dm.activeDownloads[url]
	dm.mu.RUnlock()
	if ok {
		cancel()
		return true
	}
	return false
}

// GetActiveDownloads returns the list of active download URLs.
func (dm *DownloadManager) GetActiveDownloads() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	var urls []string
	for url := range dm.activeDownloads {
		urls = append(urls, url)
	}
	return urls
}

func (dm *DownloadManager) buildArgs(url string) []string {
	cfg := dm.cfg
	args := []string{
		"--url", url,
		"-d", cfg.Yutto.Basic.Dir,
		"--tmp-dir", cfg.Yutto.Basic.TmpDir,
	}

	if cfg.Yutto.Resource.RequireSubtitle {
		args = append(args, "--subtitles")
	}
	if cfg.Yutto.Batch.WithSection {
		args = append(args, "--section")
	}

	return args
}

var progressRegex = regexp.MustCompile(`(\d+\.?\d*)%`)

func (dm *DownloadManager) parseProgress(r interface{ Read([]byte) (int, error) }, ch chan<- Progress) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse percentage
		if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
			pct, _ := strconv.ParseFloat(matches[1], 64)
			ch <- Progress{
				Percent: pct,
				Status:  "downloading",
			}
		}

		// Parse speed/ETA if present
		if strings.Contains(line, "MiB/s") || strings.Contains(line, "KiB/s") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if strings.Contains(p, "MiB/s") || strings.Contains(p, "KiB/s") {
					ch <- Progress{Speed: parts[i-1] + " " + p, Status: "downloading"}
				}
			}
		}

		// Log non-empty lines for debugging
		if line != "" {
			log.Printf("[yutto] %s", line)
		}
	}
}

// GetVideoInfo retrieves video information without downloading.
func GetVideoInfo(url string) (map[string]interface{}, error) {
	cmd := exec.Command("yutto", "--url", url, "--info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// Parse output (simplified)
	info := map[string]interface{}{
		"raw_output": string(output),
	}
	return info, nil
}
