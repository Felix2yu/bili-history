package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bilibili-history-go/database"
	"bilibili-history-go/utils"
)

type CleanStatus struct {
	IsRunning      bool   `json:"is_running"`
	Status         string `json:"status"`
	TotalRecords   int    `json:"total_records"`
	CleanedRecords int    `json:"cleaned_records"`
	DeletedFiles   int    `json:"deleted_files"`
	FreedSpace     int64  `json:"freed_space"`
	ErrorMessage   string `json:"error_message,omitempty"`
	StartTime      int64  `json:"start_time,omitempty"`
	EndTime        int64  `json:"end_time,omitempty"`
}

var (
	cleanStatus = CleanStatus{
		IsRunning: false,
		Status:    "idle",
	}
	cleanMutex sync.Mutex
)

func GetCleanStatus() CleanStatus {
	cleanMutex.Lock()
	defer cleanMutex.Unlock()
	return cleanStatus
}

func setCleanStatus(status CleanStatus) {
	cleanMutex.Lock()
	defer cleanMutex.Unlock()
	cleanStatus = status
}

type CleanOptions struct {
	CleanDuplicates    bool `json:"clean_duplicates"`
	CleanInvalidVideos bool `json:"clean_invalid_videos"`
	CleanOldHistory    bool `json:"clean_old_history"`
	DaysToKeep         int  `json:"days_to_keep"`
	CleanImageCache    bool `json:"clean_image_cache"`
	CleanLogs          bool `json:"clean_logs"`
	LogDaysToKeep      int  `json:"log_days_to_keep"`
}

func StartClean(options CleanOptions) (map[string]interface{}, error) {
	status := GetCleanStatus()
	if status.IsRunning {
		return nil, fmt.Errorf("clean already running")
	}

	newStatus := CleanStatus{
		IsRunning: true,
		Status:    "running",
		StartTime: time.Now().Unix(),
	}
	setCleanStatus(newStatus)

	go func() {
		defer func() {
			status := GetCleanStatus()
			status.IsRunning = false
			status.EndTime = time.Now().Unix()
			if status.ErrorMessage != "" {
				status.Status = "error"
			} else {
				status.Status = "completed"
			}
			setCleanStatus(status)
		}()

		var cleanedRecords int
		var deletedFiles int
		var freedSpace int64

		if options.CleanDuplicates {
			utils.LogInfo("开始清理重复记录...")
			count, err := cleanDuplicates()
			if err != nil {
				status := GetCleanStatus()
				status.ErrorMessage = fmt.Sprintf("清理重复记录失败: %v", err)
				setCleanStatus(status)
				return
			}
			cleanedRecords += count
			utils.LogSuccess("清理重复记录完成，清理了 %d 条记录", count)

			status := GetCleanStatus()
			status.CleanedRecords = cleanedRecords
			setCleanStatus(status)
		}

		if options.CleanOldHistory && options.DaysToKeep > 0 {
			utils.LogInfo("开始清理旧历史记录 (保留 %d 天)...", options.DaysToKeep)
			count, err := cleanOldHistory(options.DaysToKeep)
			if err != nil {
				status := GetCleanStatus()
				status.ErrorMessage = fmt.Sprintf("清理旧历史记录失败: %v", err)
				setCleanStatus(status)
				return
			}
			cleanedRecords += count
			utils.LogSuccess("清理旧历史记录完成，清理了 %d 条记录", count)

			status := GetCleanStatus()
			status.CleanedRecords = cleanedRecords
			setCleanStatus(status)
		}

		if options.CleanImageCache {
			utils.LogInfo("开始清理图片缓存...")
			count, space, err := cleanImageCache()
			if err != nil {
				status := GetCleanStatus()
				status.ErrorMessage = fmt.Sprintf("清理图片缓存失败: %v", err)
				setCleanStatus(status)
				return
			}
			deletedFiles += count
			freedSpace += space
			utils.LogSuccess("清理图片缓存完成，删除了 %d 个文件，释放 %d 字节空间", count, space)

			status := GetCleanStatus()
			status.DeletedFiles = deletedFiles
			status.FreedSpace = freedSpace
			setCleanStatus(status)
		}

		if options.CleanLogs && options.LogDaysToKeep > 0 {
			utils.LogInfo("开始清理旧日志 (保留 %d 天)...", options.LogDaysToKeep)
			count, space, err := cleanOldLogs(options.LogDaysToKeep)
			if err != nil {
				status := GetCleanStatus()
				status.ErrorMessage = fmt.Sprintf("清理旧日志失败: %v", err)
				setCleanStatus(status)
				return
			}
			deletedFiles += count
			freedSpace += space
			utils.LogSuccess("清理旧日志完成，删除了 %d 个文件，释放 %d 字节空间", count, space)

			status := GetCleanStatus()
			status.DeletedFiles = deletedFiles
			status.FreedSpace = freedSpace
			setCleanStatus(status)
		}

	}()

	return map[string]interface{}{
		"status":  "success",
		"message": "开始数据清洗",
	}, nil
}

func cleanDuplicates() (int, error) {
	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	availableYears, err := db.GetAvailableYears()
	if err != nil {
		return 0, err
	}

	totalDeleted := 0

	for _, year := range availableYears {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		result, err := conn.Exec(fmt.Sprintf(`
			DELETE FROM %s
			WHERE id NOT IN (
				SELECT MIN(id)
				FROM %s
				GROUP BY bvid, view_at
			)
		`, tableName, tableName))
		if err != nil {
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		totalDeleted += int(rowsAffected)
	}

	return totalDeleted, nil
}

func cleanOldHistory(daysToKeep int) (int, error) {
	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	cutoff := time.Now().AddDate(0, 0, -daysToKeep).Unix()

	availableYears, err := db.GetAvailableYears()
	if err != nil {
		return 0, err
	}

	totalDeleted := 0

	for _, year := range availableYears {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		result, err := conn.Exec(fmt.Sprintf(`
			DELETE FROM %s
			WHERE view_at < ?
		`, tableName), cutoff)
		if err != nil {
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		totalDeleted += int(rowsAffected)
	}

	return totalDeleted, nil
}

func cleanImageCache() (int, int64, error) {
	imagePath := utils.GetOutputPath("images")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return 0, 0, nil
	}

	var count int
	var totalSize int64

	err := filepath.Walk(imagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			count++
			totalSize += info.Size()
			if err := os.Remove(path); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return count, totalSize, err
	}

	return count, totalSize, nil
}

func cleanOldLogs(daysToKeep int) (int, int64, error) {
	logPath := utils.GetOutputPath("logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return 0, 0, nil
	}

	cutoff := time.Now().AddDate(0, 0, -daysToKeep)
	var count int
	var totalSize int64

	err := filepath.Walk(logPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.ModTime().Before(cutoff) {
			count++
			totalSize += info.Size()
			if err := os.Remove(path); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return count, totalSize, err
	}

	return count, totalSize, nil
}
