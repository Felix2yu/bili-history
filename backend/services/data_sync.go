package services

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
)

type DataSyncStatus struct {
	IsRunning    bool   `json:"is_running"`
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
	StartTime    int64  `json:"start_time,omitempty"`
	LastUpdateAt int64  `json:"last_update_at,omitempty"`
}

type IntegrityCheckResult struct {
	Success          bool   `json:"success"`
	Timestamp        string `json:"timestamp"`
	TotalJSONFiles   int    `json:"total_json_files"`
	TotalJSONRecords int    `json:"total_json_records"`
	TotalDBRecords   int    `json:"total_db_records"`
	Difference       int    `json:"difference"`
	Message          string `json:"message,omitempty"`
}

type SyncResult struct {
	Success         bool              `json:"success"`
	Timestamp       string            `json:"timestamp"`
	TotalSynced     int               `json:"total_synced"`
	JSONToDBCount   int               `json:"json_to_db_count"`
	DBToJSONCount   int               `json:"db_to_json_count"`
	TotalJSONRecords int              `json:"total_json_records"`
	TotalDBRecords  int               `json:"total_db_records"`
	SyncedDays      []SyncedDayDetail `json:"synced_days,omitempty"`
	Message         string            `json:"message,omitempty"`
}

type SyncedDayDetail struct {
	Date          string   `json:"date"`
	ImportedCount int      `json:"imported_count"`
	Source        string   `json:"source"`
	Titles        []string `json:"titles,omitempty"`
}

var (
	dataSyncStatus  = DataSyncStatus{Status: "idle"}
	dataSyncMutex   sync.Mutex
	lastCheckResult *IntegrityCheckResult
	lastSyncResult  *SyncResult
)

func GetDataSyncStatus() DataSyncStatus {
	dataSyncMutex.Lock()
	defer dataSyncMutex.Unlock()
	return dataSyncStatus
}

func setDataSyncStatus(status DataSyncStatus) {
	dataSyncMutex.Lock()
	defer dataSyncMutex.Unlock()
	dataSyncStatus = status
}

func GetLastSyncResult() *SyncResult {
	dataSyncMutex.Lock()
	defer dataSyncMutex.Unlock()
	return lastSyncResult
}

func GetIntegrityReportData() *IntegrityReportData {
	return getReportData()
}

func RunIntegrityCheck(forceCheck bool) (*IntegrityCheckResult, error) {
	cfg := config.GetConfig()
	if !forceCheck && cfg != nil && !cfg.Server.DataIntegrity.CheckOnStartup {
		return &IntegrityCheckResult{
			Success: true,
			Message: "数据完整性校验已在配置中禁用",
		}, nil
	}

	now := time.Now()
	result := &IntegrityCheckResult{
		Success:   true,
		Timestamp: now.Format(time.RFC3339),
	}

	// Count DB records
	dbRecords, err := countDBRecords()
	if err != nil {
		return nil, fmt.Errorf("统计数据库记录失败: %v", err)
	}
	result.TotalDBRecords = dbRecords
	result.TotalJSONRecords = dbRecords // no JSON layer anymore
	result.Difference = 0

	result.Message = fmt.Sprintf("数据库共 %d 条记录", dbRecords)

	dataSyncMutex.Lock()
	lastCheckResult = result
	dataSyncMutex.Unlock()

	return result, nil
}



func countDBRecords() (int, error) {
	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}

	years, err := db.GetAvailableYears()
	if err != nil {
		return 0, err
	}

	total := 0
	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}
		var count int
		conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		total += count
	}

	return total, nil
}

type IntegrityReportData struct {
	TotalRecords int            `json:"total_records"`
	Years        []YearStatData `json:"years"`
	MaxYearCount int            `json:"max_year_count"`
}

type YearStatData struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

func getReportData() *IntegrityReportData {
	data := &IntegrityReportData{}

	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return data
	}

	years, err := db.GetAvailableYears()
	if err != nil {
		return data
	}

	sort.Ints(years)

	maxCount := 0
	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}
		var count int
		conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		data.Years = append(data.Years, YearStatData{Year: year, Count: count})
		data.TotalRecords += count
		if count > maxCount {
			maxCount = count
		}
	}
	data.MaxYearCount = maxCount

	return data
}

func RunSyncData() (*SyncResult, error) {
	status := GetDataSyncStatus()
	if status.IsRunning {
		return nil, fmt.Errorf("同步正在进行中")
	}

	setDataSyncStatus(DataSyncStatus{
		IsRunning: true,
		Status:    "running",
		Message:   "正在检查数据...",
		StartTime: time.Now().Unix(),
	})

	result := syncDataInternal()

	dataSyncMutex.Lock()
	lastSyncResult = result
	dataSyncMutex.Unlock()

	setDataSyncStatus(DataSyncStatus{
		IsRunning:    false,
		Status:       "completed",
		Message:      result.Message,
		LastUpdateAt: time.Now().Unix(),
	})

	return result, nil
}

func syncDataInternal() *SyncResult {
	now := time.Now()
	result := &SyncResult{
		Success:   true,
		Timestamp: now.Format(time.RFC3339),
	}

	dbRecords, _ := countDBRecords()
	result.TotalDBRecords = dbRecords
	result.TotalSynced = 0
	result.JSONToDBCount = 0
	result.DBToJSONCount = 0

	result.Message = fmt.Sprintf("数据已直接写入数据库，共 %d 条记录", dbRecords)

	return result
}
