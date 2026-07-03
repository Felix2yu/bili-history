package services

import (
	"fmt"
	"sort"
	"strings"
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
	integrityReport string
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

func GetIntegrityReport() string {
	dataSyncMutex.Lock()
	defer dataSyncMutex.Unlock()
	return integrityReport
}

func RunIntegrityCheck(dbPath, jsonPath string, forceCheck bool) (*IntegrityCheckResult, error) {
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

	// Generate markdown report
	report := generateIntegrityReport(result)

	dataSyncMutex.Lock()
	lastCheckResult = result
	integrityReport = report
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

func generateIntegrityReport(result *IntegrityCheckResult) string {
	var sb strings.Builder

	sb.WriteString("# 数据完整性检查报告\n\n")
	sb.WriteString(fmt.Sprintf("**检查时间**: %s\n\n", result.Timestamp))

	sb.WriteString("## 统计概览\n\n")
	sb.WriteString(fmt.Sprintf("| 项目 | 数量 |\n"))
	sb.WriteString(fmt.Sprintf("|------|------|\n"))
	sb.WriteString(fmt.Sprintf("| 数据库记录数 | %d |\n\n", result.TotalDBRecords))

	sb.WriteString("**状态**: 数据正常\n\n")

	// Per-year breakdown
	sb.WriteString("## 各年份数据统计\n\n")
	sb.WriteString("| 年份 | 记录数 |\n")
	sb.WriteString("|------|--------|\n")

	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return sb.String()
	}

	years, err := db.GetAvailableYears()
	if err != nil {
		return sb.String()
	}

	// Sort years
	sort.Ints(years)

	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}
		var count int
		conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		sb.WriteString(fmt.Sprintf("| %d | %d |\n", year, count))
	}

	return sb.String()
}

func RunSyncData(dbPath, jsonPath string) (*SyncResult, error) {
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
