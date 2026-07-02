package services

import (
	"fmt"
	"sync"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

var (
	videoDetailProgress = models.VideoDetailProgress{
		IsProcessing: false,
		IsComplete:   false,
		IsStopped:    false,
		Status:       "idle",
	}
	videoDetailMutex sync.Mutex
)

// GetVideoDetailProgress 获取视频详情获取进度
func GetVideoDetailProgress() models.VideoDetailProgress {
	videoDetailMutex.Lock()
	defer videoDetailMutex.Unlock()
	return videoDetailProgress
}

func setVideoDetailProgress(progress models.VideoDetailProgress) {
	videoDetailMutex.Lock()
	defer videoDetailMutex.Unlock()
	videoDetailProgress = progress
}

// StopVideoDetailFetch 停止视频详情获取
func StopVideoDetailFetch() error {
	videoDetailMutex.Lock()
	defer videoDetailMutex.Unlock()

	if !videoDetailProgress.IsProcessing {
		return fmt.Errorf("没有正在进行的获取任务")
	}

	videoDetailProgress.IsStopped = true
	videoDetailProgress.Status = "stopping"
	return nil
}

// ResetVideoDetailProgress 重置视频详情获取状态
func ResetVideoDetailProgress() {
	videoDetailMutex.Lock()
	defer videoDetailMutex.Unlock()
	videoDetailProgress = models.VideoDetailProgress{
		IsProcessing: false,
		IsComplete:   false,
		IsStopped:    false,
		Status:       "idle",
	}
}

// FetchVideoDetailFromBili 从B站获取单个视频详情
func FetchVideoDetailFromBili(bvid string) (*models.VideoBaseInfo, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	client := biliapi.NewClient(cfg.SESSDATA)
	videoInfo, err := client.GetVideoInfo(bvid)
	if err != nil {
		return nil, fmt.Errorf("获取视频详情失败: %v", err)
	}

	now := time.Now().Unix()
	video := &models.VideoBaseInfo{
		Bvid:         videoInfo.Bvid,
		Aid:          videoInfo.Aid,
		Videos:       videoInfo.Videos,
		Tid:          videoInfo.Tid,
		Tname:        videoInfo.Tname,
		Copyright:    videoInfo.Copyright,
		Pic:          videoInfo.Pic,
		Title:        videoInfo.Title,
		Pubdate:      videoInfo.Pubdate,
		Ctime:        videoInfo.Ctime,
		Desc:         videoInfo.Desc,
		Duration:     videoInfo.Duration,
		Cid:          0,
		OwnerMid:     videoInfo.Owner.Mid,
		OwnerName:    videoInfo.Owner.Name,
		OwnerFace:    videoInfo.Owner.Face,
		StatView:     videoInfo.Stat.View,
		StatDanmaku:  videoInfo.Stat.Danmaku,
		StatReply:    videoInfo.Stat.Reply,
		StatFavorite: videoInfo.Stat.Favorite,
		StatCoin:     videoInfo.Stat.Coin,
		StatShare:    videoInfo.Stat.Share,
		StatLike:     videoInfo.Stat.Like,
		FetchTime:    now,
		UpdateTime:   now,
	}

	return video, nil
}

// SaveVideoDetail 保存视频详情到数据库
func SaveVideoDetail(video *models.VideoBaseInfo) error {
	return database.UpsertVideoBaseInfo(video)
}

// BatchFetchVideoDetails 批量获取视频详情
func BatchFetchVideoDetails(bvids []string) (map[string]interface{}, error) {
	status := GetVideoDetailProgress()
	if status.IsProcessing {
		return nil, fmt.Errorf("已有获取任务正在运行")
	}

	if len(bvids) == 0 {
		return nil, fmt.Errorf("视频列表不能为空")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}
	if cfg.SESSDATA == "" {
		return nil, fmt.Errorf("SESSDATA未配置")
	}

	now := time.Now().Unix()
	newProgress := models.VideoDetailProgress{
		IsProcessing:    true,
		IsComplete:      false,
		IsStopped:       false,
		TotalVideos:     len(bvids),
		ProcessedVideos: 0,
		SuccessCount:    0,
		FailedCount:     0,
		StartTime:       now,
		LastUpdateTime:  now,
		Status:          "running",
	}
	setVideoDetailProgress(newProgress)

	go processBatchFetch(bvids, cfg.SESSDATA)

	return map[string]interface{}{
		"status":  "success",
		"message": "开始批量获取视频详情",
		"total":   len(bvids),
	}, nil
}

// BatchFetchFromHistory 从历史记录批量获取视频详情
func BatchFetchFromHistory(skipExisting bool) (map[string]interface{}, error) {
	status := GetVideoDetailProgress()
	if status.IsProcessing {
		return nil, fmt.Errorf("已有获取任务正在运行")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}
	if cfg.SESSDATA == "" {
		return nil, fmt.Errorf("SESSDATA未配置")
	}

	allBvids, err := database.GetUniqueBvidsFromHistory()
	if err != nil {
		return nil, fmt.Errorf("获取历史记录视频列表失败: %v", err)
	}

	var bvidsToFetch []string
	if skipExisting {
		fetchedBvids, err := database.GetFetchedBvids()
		if err != nil {
			return nil, fmt.Errorf("获取已获取视频列表失败: %v", err)
		}
		for _, bvid := range allBvids {
			if !fetchedBvids[bvid] {
				bvidsToFetch = append(bvidsToFetch, bvid)
			}
		}
	} else {
		bvidsToFetch = allBvids
	}

	if len(bvidsToFetch) == 0 {
		return map[string]interface{}{
			"status":  "success",
			"message": "没有需要获取的视频",
			"total":   0,
		}, nil
	}

	now := time.Now().Unix()
	newProgress := models.VideoDetailProgress{
		IsProcessing:    true,
		IsComplete:      false,
		IsStopped:       false,
		TotalVideos:     len(bvidsToFetch),
		ProcessedVideos: 0,
		SuccessCount:    0,
		FailedCount:     0,
		StartTime:       now,
		LastUpdateTime:  now,
		Status:          "running",
	}
	setVideoDetailProgress(newProgress)

	go processBatchFetch(bvidsToFetch, cfg.SESSDATA)

	return map[string]interface{}{
		"status":       "success",
		"message":      "开始从历史记录批量获取视频详情",
		"total":        len(bvidsToFetch),
		"history_total": len(allBvids),
	}, nil
}

func processBatchFetch(bvids []string, sessdata string) {
	defer func() {
		progress := GetVideoDetailProgress()
		progress.IsProcessing = false
		progress.LastUpdateTime = time.Now().Unix()
		if progress.IsStopped {
			progress.Status = "stopped"
		} else if progress.ErrorMessage != "" {
			progress.Status = "error"
		} else {
			progress.IsComplete = true
			progress.Status = "completed"
		}
		setVideoDetailProgress(progress)
	}()

	client := biliapi.NewClient(sessdata)
	successCount := 0
	failedCount := 0

	for i, bvid := range bvids {
		progress := GetVideoDetailProgress()
		if progress.IsStopped {
			break
		}

		videoInfo, err := client.GetVideoInfo(bvid)
		if err != nil {
			failedCount++
			utils.LogWarning("获取视频详情失败 %s: %v", bvid, err)
		} else {
			now := time.Now().Unix()
			video := &models.VideoBaseInfo{
				Bvid:         videoInfo.Bvid,
				Aid:          videoInfo.Aid,
				Videos:       videoInfo.Videos,
				Tid:          videoInfo.Tid,
				Tname:        videoInfo.Tname,
				Copyright:    videoInfo.Copyright,
				Pic:          videoInfo.Pic,
				Title:        videoInfo.Title,
				Pubdate:      videoInfo.Pubdate,
				Ctime:        videoInfo.Ctime,
				Desc:         videoInfo.Desc,
				Duration:     videoInfo.Duration,
				Cid:          0,
				OwnerMid:     videoInfo.Owner.Mid,
				OwnerName:    videoInfo.Owner.Name,
				OwnerFace:    videoInfo.Owner.Face,
				StatView:     videoInfo.Stat.View,
				StatDanmaku:  videoInfo.Stat.Danmaku,
				StatReply:    videoInfo.Stat.Reply,
				StatFavorite: videoInfo.Stat.Favorite,
				StatCoin:     videoInfo.Stat.Coin,
				StatShare:    videoInfo.Stat.Share,
				StatLike:     videoInfo.Stat.Like,
				FetchTime:    now,
				UpdateTime:   now,
			}

			err = database.UpsertVideoBaseInfo(video)
			if err != nil {
				failedCount++
				utils.LogWarning("保存视频详情失败 %s: %v", bvid, err)
			} else {
				successCount++
			}
		}

		progress = GetVideoDetailProgress()
		progress.ProcessedVideos = i + 1
		progress.SuccessCount = successCount
		progress.FailedCount = failedCount
		progress.LastUpdateTime = time.Now().Unix()
		setVideoDetailProgress(progress)

		time.Sleep(200 * time.Millisecond)
	}
}
