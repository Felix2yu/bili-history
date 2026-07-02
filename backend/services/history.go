package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

type FetchStatus struct {
	IsRunning       bool   `json:"is_running"`
	TotalPages      int    `json:"total_pages"`
	CurrentPage     int    `json:"current_page"`
	TotalRecords    int    `json:"total_records"`
	NewRecords      int    `json:"new_records"`
	ErrorMessage    string `json:"error_message,omitempty"`
	Status          string `json:"status"`
	StartTime       int64  `json:"start_time,omitempty"`
	LastUpdateTime  int64  `json:"last_update_time,omitempty"`
}

var (
	fetchStatus = FetchStatus{
		IsRunning: false,
		Status:    "idle",
	}
	fetchMutex sync.Mutex
)

func GetFetchStatus() FetchStatus {
	fetchMutex.Lock()
	defer fetchMutex.Unlock()
	return fetchStatus
}

func setFetchStatus(status FetchStatus) {
	fetchMutex.Lock()
	defer fetchMutex.Unlock()
	fetchStatus = status
}

type HistoryFileEntry struct {
	Title    string          `json:"title"`
	LongTitle string         `json:"long_title"`
	Cover    string          `json:"cover"`
	URI      string          `json:"uri"`
	History  HistoryInfoData `json:"history"`
	ViewAt   int64           `json:"view_at"`
	Progress int             `json:"progress"`
	Badge    string          `json:"badge"`
	ShowTitle string         `json:"show_title"`
	Icon     string          `json:"icon"`
	Business string          `json:"business"`
	Bvid     string          `json:"bvid"`
	Duration int             `json:"duration"`
}

type HistoryInfoData struct {
	Bvid     string `json:"bvid"`
	Page     int    `json:"page"`
	Cid      int    `json:"cid"`
	Part     string `json:"part"`
	Business string `json:"business"`
	Dt       int    `json:"dt"`
}

func FindLatestHistoryDate() (time.Time, error) {
	basePath := utils.GetOutputPath("history_by_date")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return time.Time{}, nil
	}

	latestDate := time.Time{}

	years, err := filepath.Glob(filepath.Join(basePath, "*"))
	if err != nil {
		return time.Time{}, err
	}

	for _, yearPath := range years {
		yearInfo, err := os.Stat(yearPath)
		if err != nil || !yearInfo.IsDir() {
			continue
		}
		year := yearInfo.Name()

		months, err := filepath.Glob(filepath.Join(yearPath, "*"))
		if err != nil {
			continue
		}

		for _, monthPath := range months {
			monthInfo, err := os.Stat(monthPath)
			if err != nil || !monthInfo.IsDir() {
				continue
			}
			month := monthInfo.Name()

			days, err := filepath.Glob(filepath.Join(monthPath, "*.json"))
			if err != nil {
				continue
			}

			for _, dayPath := range days {
				dayFile := filepath.Base(dayPath)
				day := dayFile[:len(dayFile)-5]

				dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)
				date, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					continue
				}

				if date.After(latestDate) {
					latestDate = date
				}
			}
		}
	}

	return latestDate, nil
}

func SaveHistoryToFile(entries []biliapi.HistoryEntry) (int, error) {
	basePath := utils.GetOutputPath("history_by_date")
	savedCount := 0

	for _, entry := range entries {
		viewAt := time.Unix(entry.ViewAt, 0)
		year := viewAt.Format("2006")
		month := viewAt.Format("01")
		day := viewAt.Format("02")

		folderPath := filepath.Join(basePath, year, month)
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			return savedCount, fmt.Errorf("create folder error: %w", err)
		}

		filePath := filepath.Join(folderPath, fmt.Sprintf("%s.json", day))

		fileEntry := HistoryFileEntry{
			Title:    entry.Title,
			LongTitle: entry.LongTitle,
			Cover:    entry.Cover,
			URI:      entry.URI,
			History: HistoryInfoData{
				Bvid:     entry.History.Bvid,
				Page:     entry.History.Page,
				Cid:      entry.History.Cid,
				Part:     entry.History.Part,
				Business: entry.History.Business,
				Dt:       entry.History.Dt,
			},
			ViewAt:   entry.ViewAt,
			Progress: entry.Progress,
			Badge:    entry.Badge,
			ShowTitle: entry.ShowTitle,
			Icon:     entry.Icon,
			Business: entry.Business,
			Bvid:     entry.Bvid,
			Duration: entry.DTotal,
		}

		existingRecords := make(map[string]bool)
		var dailyData []HistoryFileEntry

		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				json.Unmarshal(data, &dailyData)
				for _, item := range dailyData {
					key := fmt.Sprintf("%s_%d", item.History.Bvid, item.ViewAt)
					existingRecords[key] = true
				}
			}
		}

		key := fmt.Sprintf("%s_%d", entry.History.Bvid, entry.ViewAt)
		if !existingRecords[key] {
			dailyData = append(dailyData, fileEntry)
			existingRecords[key] = true
			savedCount++
		}

		data, err := json.MarshalIndent(dailyData, "", "  ")
		if err != nil {
			return savedCount, fmt.Errorf("marshal data error: %w", err)
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return savedCount, fmt.Errorf("write file error: %w", err)
		}
	}

	return savedCount, nil
}

func FetchHistory(skipExists bool, processVideoDetails bool) (map[string]interface{}, error) {
	status := GetFetchStatus()
	if status.IsRunning {
		return nil, fmt.Errorf("fetch already running")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config error: %w", err)
	}
	if cfg.SESSDATA == "" {
		return nil, fmt.Errorf("SESSDATA not configured")
	}

	newStatus := FetchStatus{
		IsRunning:      true,
		Status:         "running",
		StartTime:      time.Now().Unix(),
		LastUpdateTime: time.Now().Unix(),
	}
	setFetchStatus(newStatus)

	go func() {
		defer func() {
			status := GetFetchStatus()
			status.IsRunning = false
			status.LastUpdateTime = time.Now().Unix()
			if status.ErrorMessage != "" {
				status.Status = "error"
			} else {
				status.Status = "completed"
			}
			setFetchStatus(status)
		}()

		client := biliapi.NewClient(cfg.SESSDATA)

		var cutoffTimestamp int64
		if skipExists {
			latestDate, err := FindLatestHistoryDate()
			if err == nil && !latestDate.IsZero() {
				cutoffTimestamp = time.Date(latestDate.Year(), latestDate.Month(), latestDate.Day(), 0, 0, 0, 0, time.Local).Unix()
			}
		}

		var allEntries []biliapi.HistoryEntry
		pageCount := 0
		var max int64 = 0
		var viewAt int64 = 0
		emptyPageCount := 0
		maxEmptyPages := 3
		ps := 30

		for {
			pageCount++

			data, err := client.GetHistory(max, viewAt, ps)
			if err != nil {
				status := GetFetchStatus()
				status.ErrorMessage = err.Error()
				setFetchStatus(status)
				return
			}

			status := GetFetchStatus()
			status.CurrentPage = pageCount
			status.LastUpdateTime = time.Now().Unix()
			setFetchStatus(status)

			if len(data.List) == 0 {
				emptyPageCount++
				if emptyPageCount >= maxEmptyPages {
					break
				}
				if data.Cursor.Max == 0 || (max > 0 && data.Cursor.Max < 1000000) {
					break
				}
				max = data.Cursor.Max
				viewAt = data.Cursor.ViewAt
				continue
			}

			emptyPageCount = 0

			hasNew := false
			for _, entry := range data.List {
				if entry.ViewAt > cutoffTimestamp {
					allEntries = append(allEntries, entry)
					hasNew = true
				}
			}

			status = GetFetchStatus()
			status.TotalRecords = len(allEntries)
			status.NewRecords = len(allEntries)
			setFetchStatus(status)

			if !hasNew && cutoffTimestamp > 0 {
				break
			}

			if len(data.List) > 0 {
				viewAt = data.List[len(data.List)-1].ViewAt
			}
			max = data.Cursor.Max

			if max == 0 && len(data.List) == 0 {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		savedCount, err := SaveHistoryToFile(allEntries)
		if err != nil {
			status := GetFetchStatus()
			status.ErrorMessage = fmt.Sprintf("save history error: %v", err)
			setFetchStatus(status)
			return
		}

		status = GetFetchStatus()
		status.TotalPages = pageCount
		status.NewRecords = savedCount
		setFetchStatus(status)

		importResult, err := ImportHistoryFiles(false)
		if err != nil {
			status := GetFetchStatus()
			status.ErrorMessage = fmt.Sprintf("import error: %v", err)
			setFetchStatus(status)
			return
		}

		status = GetFetchStatus()
		status.NewRecords = importResult.InsertedCount
		setFetchStatus(status)
	}()

	return map[string]interface{}{
		"status":  "success",
		"message": "开始获取历史记录",
	}, nil
}

func ImportHistoryFiles(syncDeleted bool) (*ImportResult, error) {
	basePath := utils.GetOutputPath("history_by_date")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return &ImportResult{InsertedCount: 0}, nil
	}

	result := &ImportResult{
		InsertedCount: 0,
		TotalFiles:    0,
	}

	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	years, err := filepath.Glob(filepath.Join(basePath, "*"))
	if err != nil {
		return nil, err
	}

	for _, yearPath := range years {
		yearInfo, err := os.Stat(yearPath)
		if err != nil || !yearInfo.IsDir() {
			continue
		}
		yearStr := yearInfo.Name()
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			continue
		}

		err = db.EnsureTableForYear(year)
		if err != nil {
			return nil, fmt.Errorf("ensure year table error: %w", err)
		}

		tableName := fmt.Sprintf("bilibili_history_%d", year)

		months, err := filepath.Glob(filepath.Join(yearPath, "*"))
		if err != nil {
			continue
		}

		for _, monthPath := range months {
			monthInfo, err := os.Stat(monthPath)
			if err != nil || !monthInfo.IsDir() {
				continue
			}

			days, err := filepath.Glob(filepath.Join(monthPath, "*.json"))
			if err != nil {
				continue
			}

			for _, dayPath := range days {
				result.TotalFiles++

				data, err := os.ReadFile(dayPath)
				if err != nil {
					continue
				}

				var entries []HistoryFileEntry
				if err := json.Unmarshal(data, &entries); err != nil {
					continue
				}

				for _, entry := range entries {
					business := entry.Business
					if business == "" {
						business = entry.History.Business
					}
					if business != "archive" {
						continue
					}

					bvid := entry.Bvid
					if bvid == "" {
						bvid = entry.History.Bvid
					}
					if bvid == "" {
						continue
					}

					history := models.HistoryRecord{
						Bvid:       bvid,
						Title:      entry.Title,
						LongTitle:  entry.LongTitle,
						Cover:      entry.Cover,
						URI:        entry.URI,
						Page:       entry.History.Page,
						Cid:        int64(entry.History.Cid),
						Part:       entry.History.Part,
						Business:   business,
						Dt:         entry.History.Dt,
						ViewAt:     entry.ViewAt,
						Progress:   entry.Progress,
						Badge:      entry.Badge,
						ShowTitle:  entry.ShowTitle,
						Duration:   entry.Duration,
					}

					inserted, err := database.InsertHistoryRecord(conn, tableName, &history)
					if err != nil {
						continue
					}
					if inserted {
						result.InsertedCount++
					}
				}
			}
		}
	}

	return result, nil
}

type ImportResult struct {
	InsertedCount int `json:"inserted_count"`
	TotalFiles    int `json:"total_files"`
}
