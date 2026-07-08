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

// FindLatestHistoryDate queries the DB for the latest view_at timestamp
func FindLatestHistoryDate() (time.Time, error) {
	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return time.Time{}, nil
	}

	years, err := db.GetAvailableYears()
	if err != nil || len(years) == 0 {
		return time.Time{}, nil
	}

	var latestViewAt int64
	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}
		var maxViewAt int64
		err := conn.QueryRow(fmt.Sprintf("SELECT COALESCE(MAX(view_at), 0) FROM %s", tableName)).Scan(&maxViewAt)
		if err != nil {
			continue
		}
		if maxViewAt > latestViewAt {
			latestViewAt = maxViewAt
		}
	}

	if latestViewAt == 0 {
		return time.Time{}, nil
	}
	return time.Unix(latestViewAt, 0), nil
}

func FetchHistory(skipExists bool) (map[string]interface{}, error) {
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

		maxConsecutiveErrors := 3
		consecutiveErrors := 0

		for {
			pageCount++

			data, err := client.GetHistory(max, viewAt, ps)
			if err != nil {
				consecutiveErrors++
				utils.LogWarning("获取历史记录第 %d 页失败: %v (连续失败 %d/%d)", pageCount, err, consecutiveErrors, maxConsecutiveErrors)
				if consecutiveErrors >= maxConsecutiveErrors {
					status := GetFetchStatus()
					status.ErrorMessage = fmt.Sprintf("连续 %d 页失败，停止抓取: %v", consecutiveErrors, err)
					setFetchStatus(status)
					// 自动暂停任务并发送告警通知
					_ = database.SetTaskEnabled("fetch_history", false)
					alertMsg := fmt.Sprintf("⚠️ 历史记录抓取已自动暂停\n\n原因：连续 %d 页请求失败\n最后错误：%s\n已抓取：%d 页，%d 条记录\n\n请检查网络或代理设置后，在任务列表中重新启用。", consecutiveErrors, err.Error(), pageCount, len(allEntries))
					_ = SendShoutrrrNotification("⚠️ B站历史抓取异常", alertMsg)
					break
				}
				time.Sleep(2 * time.Second)
				continue
			}
			consecutiveErrors = 0

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

		// Directly insert to DB
		db := database.GetSQLiteDB()
		conn := db.GetDB()
		if conn == nil {
			status := GetFetchStatus()
			status.ErrorMessage = "数据库未初始化"
			setFetchStatus(status)
			return
		}

		insertedCount := 0
		for _, entry := range allEntries {
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

			year := utils.GetYearFromTimestamp(entry.ViewAt)
			if err := db.EnsureTableForYear(year); err != nil {
				continue
			}
			tableName := fmt.Sprintf("bilibili_history_%d", year)

			record := models.HistoryRecord{
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
				Duration:   entry.DTotal,
				AuthorName: entry.AuthorName,
				AuthorFace: entry.AuthorFace,
				AuthorMid:  entry.AuthorMid,
			}

			inserted, err := database.InsertHistoryRecord(conn, tableName, &record)
			if err != nil {
				continue
			}
			if inserted {
				insertedCount++
			}
		}

		status = GetFetchStatus()
		status.TotalPages = pageCount
		status.NewRecords = insertedCount
		setFetchStatus(status)
	}()

	return map[string]interface{}{
		"status":  "success",
		"message": "开始获取历史记录",
	}, nil
}

// FetchHistorySync starts a fetch and waits for it to complete before returning.
// Used by the scheduler chain so that subsequent tasks (import, analyze, report)
// can operate on the freshly fetched data.
func FetchHistorySync(skipExists bool) (map[string]interface{}, error) {
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

	maxConsecutiveErrors := 3
	consecutiveErrors := 0
	var lastErrMsg string

	for {
		pageCount++

		data, err := client.GetHistory(max, viewAt, ps)
		if err != nil {
			consecutiveErrors++
			lastErrMsg = err.Error()
			utils.LogWarning("获取历史记录第 %d 页失败: %v (连续失败 %d/%d)", pageCount, err, consecutiveErrors, maxConsecutiveErrors)
			if consecutiveErrors >= maxConsecutiveErrors {
				s := GetFetchStatus()
				s.ErrorMessage = fmt.Sprintf("连续 %d 页失败，停止抓取: %v", consecutiveErrors, err)
				setFetchStatus(s)
				// 自动暂停任务并发送告警通知
				_ = database.SetTaskEnabled("fetch_history", false)
				alertMsg := fmt.Sprintf("⚠️ 历史记录抓取已自动暂停\n\n原因：连续 %d 页请求失败\n最后错误：%s\n已抓取：%d 页，%d 条记录\n\n请检查网络或代理设置后，在任务列表中重新启用。", consecutiveErrors, lastErrMsg, pageCount, len(allEntries))
				_ = SendShoutrrrNotification("⚠️ B站历史抓取异常", alertMsg)
				break
			}
			time.Sleep(2 * time.Second)
			continue
		}
		consecutiveErrors = 0

		s := GetFetchStatus()
		s.CurrentPage = pageCount
		s.LastUpdateTime = time.Now().Unix()
		setFetchStatus(s)

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

		s = GetFetchStatus()
		s.TotalRecords = len(allEntries)
		s.NewRecords = len(allEntries)
		setFetchStatus(s)

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

	// Directly insert to DB
	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		s := GetFetchStatus()
		s.ErrorMessage = "数据库未初始化"
		s.IsRunning = false
		s.Status = "error"
		setFetchStatus(s)
		return nil, fmt.Errorf("database not initialized")
	}

	insertedCount := 0
	for _, entry := range allEntries {
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

		year := utils.GetYearFromTimestamp(entry.ViewAt)
		if err := db.EnsureTableForYear(year); err != nil {
			continue
		}
		tableName := fmt.Sprintf("bilibili_history_%d", year)

		record := models.HistoryRecord{
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
			Duration:   entry.DTotal,
			AuthorName: entry.AuthorName,
			AuthorFace: entry.AuthorFace,
			AuthorMid:  entry.AuthorMid,
		}

		inserted, err := database.InsertHistoryRecord(conn, tableName, &record)
		if err != nil {
			continue
		}
		if inserted {
			insertedCount++
		}
	}

	finalStatus := GetFetchStatus()
	finalStatus.TotalPages = pageCount
	finalStatus.NewRecords = insertedCount
	finalStatus.IsRunning = false
	finalStatus.Status = "completed"
	finalStatus.LastUpdateTime = time.Now().Unix()
	setFetchStatus(finalStatus)

	utils.LogSuccess("历史记录同步获取完成: %d 页, %d 条新记录", pageCount, insertedCount)

	return map[string]interface{}{
		"status":       "success",
		"message":      "历史记录获取完成",
		"total_pages":  pageCount,
		"total_records": len(allEntries),
		"new_records":  insertedCount,
	}, nil
}
