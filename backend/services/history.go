package services

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

type FetchStatus struct {
	TaskID         string `json:"task_id"`
	IsRunning      bool   `json:"is_running"`
	TotalPages     int    `json:"total_pages"`
	CurrentPage    int    `json:"current_page"`
	TotalRecords   int    `json:"total_records"`
	NewRecords     int    `json:"new_records"`
	ErrorMessage   string `json:"error_message,omitempty"`
	Status         string `json:"status"`
	StartTime      int64  `json:"start_time,omitempty"`
	LastUpdateTime int64  `json:"last_update_time,omitempty"`
}

var (
	fetchRunningCount atomic.Int32
	fetchTasks        = make(map[string]*FetchStatus)
	fetchMutex        sync.RWMutex
)

// GetFetchStatusOverall returns overall status with RunningCount
func GetFetchStatusOverall() map[string]interface{} {
	fetchMutex.RLock()
	defer fetchMutex.RUnlock()

	tasks := make(map[string]*FetchStatus)
	for id, status := range fetchTasks {
		tasks[id] = status
	}

	return map[string]interface{}{
		"running_count": fetchRunningCount.Load(),
		"is_running":    fetchRunningCount.Load() > 0,
		"tasks":         tasks,
	}
}

func setFetchTaskStatus(taskID string, status *FetchStatus) {
	fetchMutex.Lock()
	defer fetchMutex.Unlock()
	status.TaskID = taskID
	fetchTasks[taskID] = status
}

func removeFetchTaskStatus(taskID string) {
	fetchMutex.Lock()
	defer fetchMutex.Unlock()
	delete(fetchTasks, taskID)
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

func FetchHistory(taskID string, skipExists bool) (map[string]interface{}, error) {
	if taskID == "" {
		taskID = fmt.Sprintf("manual_%d", time.Now().UnixNano())
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config error: %w", err)
	}
	if cfg.SESSDATA == "" {
		return nil, fmt.Errorf("SESSDATA not configured")
	}

	fetchRunningCount.Add(1)
	setFetchTaskStatus(taskID, &FetchStatus{
		TaskID:    taskID,
		IsRunning: true,
		Status:    "running",
		StartTime: time.Now().Unix(),
	})

	go func() {
		utils.LogWarning("[%s] 后台抓取协程已启动, skipExists=%v", taskID, skipExists)
		defer func() {
			if r := recover(); r != nil {
				utils.LogError("[%s] 协程 panic: %v", taskID, r)
				status := GetFetchTaskStatus(taskID)
				if status != nil {
					status.ErrorMessage = fmt.Sprintf("panic: %v", r)
					status.Status = "error"
					status.IsRunning = false
					setFetchTaskStatus(taskID, status)
				}
			}
			fetchRunningCount.Add(-1)
			status := GetFetchTaskStatus(taskID)
			if status != nil {
				status.IsRunning = false
				status.LastUpdateTime = time.Now().Unix()
				if status.ErrorMessage != "" {
					status.Status = "error"
				} else {
					status.Status = "completed"
				}
				setFetchTaskStatus(taskID, status)
			}
		}()

		status := GetFetchTaskStatus(taskID)
		client := biliapi.NewClient(cfg.SESSDATA)

		var cutoffTimestamp int64
		if skipExists {
			latestDate, err := FindLatestHistoryDate()
			if err == nil && !latestDate.IsZero() {
				cutoffTimestamp = time.Date(latestDate.Year(), latestDate.Month(), latestDate.Day(), 0, 0, 0, 0, time.Local).Unix()
			}
			utils.LogWarning("[%s] 增量模式: latestDate=%v, cutoffTimestamp=%d", taskID, latestDate, cutoffTimestamp)
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

		utils.LogWarning("[%s] 开始抓取历史记录...", taskID)
		for {
			pageCount++

			data, err := client.GetHistory(max, viewAt, ps)
			if err != nil {
				consecutiveErrors++
				utils.LogWarning("[%s] 获取历史记录第 %d 页失败: %v (连续失败 %d/%d)", taskID, pageCount, err, consecutiveErrors, maxConsecutiveErrors)
				if consecutiveErrors >= maxConsecutiveErrors {
					status.ErrorMessage = fmt.Sprintf("连续 %d 页失败，停止抓取: %v", consecutiveErrors, err)
					setFetchTaskStatus(taskID, status)
					_ = database.SetTaskEnabled("fetch_history", false)
					alertMsg := fmt.Sprintf("⚠️ 历史记录抓取已自动暂停\n\n任务: %s\n原因：连续 %d 页请求失败\n最后错误：%s\n已抓取：%d 页，%d 条记录\n\n请检查网络或代理设置后，在任务列表中重新启用。", taskID, consecutiveErrors, err.Error(), pageCount, len(allEntries))
					_ = SendShoutrrrNotification("⚠️ B站历史抓取异常", alertMsg)
					break
				}
				time.Sleep(2 * time.Second)
				continue
			}
			consecutiveErrors = 0
			utils.LogWarning("[%s] 第 %d 页: 获取到 %d 条记录", taskID, pageCount, len(data.List))

			status.CurrentPage = pageCount
			status.LastUpdateTime = time.Now().Unix()
			setFetchTaskStatus(taskID, status)

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

			status.TotalRecords = len(allEntries)
			status.NewRecords = len(allEntries)
			setFetchTaskStatus(taskID, status)

			if !hasNew && cutoffTimestamp > 0 {
				utils.LogWarning("[%s] 没有新数据(cutoff=%d), 停止抓取", taskID, cutoffTimestamp)
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
			status.ErrorMessage = "数据库未初始化"
			setFetchTaskStatus(taskID, status)
			return
		}

		// Deduplicate
		seen := make(map[string]bool)
		var deduped []biliapi.HistoryEntry
		for _, entry := range allEntries {
			bv := entry.Bvid
			if bv == "" {
				bv = entry.History.Bvid
			}
			key := fmt.Sprintf("%s_%d", bv, entry.ViewAt)
			if !seen[key] {
				seen[key] = true
				deduped = append(deduped, entry)
			}
		}
		allEntries = deduped

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

		status.TotalPages = pageCount
		status.NewRecords = insertedCount
		setFetchTaskStatus(taskID, status)

		utils.LogSuccess("[%s] 历史记录异步获取完成: %d 页, %d 条新记录", taskID, pageCount, insertedCount)
	}()

	return map[string]interface{}{
		"status":  "success",
		"message": "开始获取历史记录",
		"task_id": taskID,
	}, nil
}

// GetFetchTaskStatus returns status for a specific task
func GetFetchTaskStatus(taskID string) *FetchStatus {
	fetchMutex.RLock()
	defer fetchMutex.RUnlock()
	if status, ok := fetchTasks[taskID]; ok {
		return status
	}
	return nil
}

// FetchHistorySync starts a fetch and waits for it to complete before returning.
func FetchHistorySync(taskID string, skipExists bool) (map[string]interface{}, error) {
	if taskID == "" {
		taskID = fmt.Sprintf("sync_%d", time.Now().UnixNano())
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config error: %w", err)
	}
	if cfg.SESSDATA == "" {
		return nil, fmt.Errorf("SESSDATA not configured")
	}

	fetchRunningCount.Add(1)
	defer fetchRunningCount.Add(-1)

	setFetchTaskStatus(taskID, &FetchStatus{
		TaskID:    taskID,
		IsRunning: true,
		Status:    "running",
		StartTime: time.Now().Unix(),
	})

	defer removeFetchTaskStatus(taskID)

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
	requestInterval := 1500 * time.Millisecond

	status := GetFetchTaskStatus(taskID)

	for {
		pageCount++

		if pageCount > 1 {
			time.Sleep(requestInterval)
		}

		data, err := client.GetHistory(max, viewAt, ps)
		if err != nil {
			consecutiveErrors++
			lastErrMsg = err.Error()
			utils.LogWarning("[%s] 获取历史记录第 %d 页失败: %v (连续失败 %d/%d)", taskID, pageCount, err, consecutiveErrors, maxConsecutiveErrors)
			if consecutiveErrors >= maxConsecutiveErrors {
				status.ErrorMessage = fmt.Sprintf("连续 %d 页失败，停止抓取: %v", consecutiveErrors, err)
				setFetchTaskStatus(taskID, status)
				_ = database.SetTaskEnabled("fetch_history", false)
				alertMsg := fmt.Sprintf("⚠️ 历史记录抓取已自动暂停\n\n任务: %s\n原因：连续 %d 页请求失败\n最后错误：%s\n已抓取：%d 页，%d 条记录\n\n请检查网络或代理设置后，在任务列表中重新启用。", taskID, consecutiveErrors, lastErrMsg, pageCount, len(allEntries))
				_ = SendShoutrrrNotification("⚠️ B站历史抓取异常", alertMsg)
				break
			}
			time.Sleep(2 * time.Second)
			continue
		}
		consecutiveErrors = 0

		status.CurrentPage = pageCount
		status.LastUpdateTime = time.Now().Unix()
		setFetchTaskStatus(taskID, status)

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

		status.TotalRecords = len(allEntries)
		status.NewRecords = len(allEntries)
		setFetchTaskStatus(taskID, status)

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
		return nil, fmt.Errorf("database not initialized")
	}

	// Deduplicate
	seen := make(map[string]bool)
	var deduped []biliapi.HistoryEntry
	for _, entry := range allEntries {
		bv := entry.Bvid
		if bv == "" {
			bv = entry.History.Bvid
		}
		key := fmt.Sprintf("%s_%d", bv, entry.ViewAt)
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, entry)
		}
	}
	allEntries = deduped

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

	utils.LogSuccess("[%s] 历史记录同步获取完成: %d 页, %d 条新记录", taskID, pageCount, insertedCount)

	return map[string]interface{}{
		"status":       "success",
		"message":      "历史记录获取完成",
		"task_id":      taskID,
		"total_pages":  pageCount,
		"total_records": len(allEntries),
		"new_records":  insertedCount,
	}, nil
}
