package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/models"
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
	Success       bool              `json:"success"`
	Timestamp     string            `json:"timestamp"`
	TotalSynced   int               `json:"total_synced"`
	JSONToDBCount int               `json:"json_to_db_count"`
	DBToJSONCount int               `json:"db_to_json_count"`
	SyncedDays    []SyncedDayDetail `json:"synced_days,omitempty"`
	Message       string            `json:"message,omitempty"`
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

	// Count JSON records
	jsonFiles, jsonRecords, err := countJSONRecords(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("统计JSON文件失败: %v", err)
	}
	result.TotalJSONFiles = jsonFiles
	result.TotalJSONRecords = jsonRecords

	// Count DB records
	dbRecords, err := countDBRecords()
	if err != nil {
		return nil, fmt.Errorf("统计数据库记录失败: %v", err)
	}
	result.TotalDBRecords = dbRecords

	result.Difference = jsonRecords - dbRecords
	if result.Difference == 0 {
		result.Message = fmt.Sprintf("数据一致：共 %d 条记录", jsonRecords)
	} else if result.Difference > 0 {
		result.Message = fmt.Sprintf("数据库缺少 %d 条记录（JSON: %d, DB: %d）", result.Difference, jsonRecords, dbRecords)
	} else {
		result.Message = fmt.Sprintf("数据库多出 %d 条记录（JSON: %d, DB: %d）", -result.Difference, jsonRecords, dbRecords)
	}

	// Generate markdown report
	report := generateIntegrityReport(result, jsonPath)

	dataSyncMutex.Lock()
	lastCheckResult = result
	integrityReport = report
	dataSyncMutex.Unlock()

	return result, nil
}

func countJSONRecords(jsonPath string) (int, int, error) {
	totalFiles := 0
	totalRecords := 0

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return 0, 0, nil
	}

	years, err := filepath.Glob(filepath.Join(jsonPath, "*"))
	if err != nil {
		return 0, 0, err
	}

	for _, yearPath := range years {
		yearInfo, err := os.Stat(yearPath)
		if err != nil || !yearInfo.IsDir() {
			continue
		}

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
				totalFiles++
				data, err := os.ReadFile(dayPath)
				if err != nil {
					continue
				}

				var entries []HistoryFileEntry
				if err := json.Unmarshal(data, &entries); err != nil {
					continue
				}
				totalRecords += len(entries)
			}
		}
	}

	return totalFiles, totalRecords, nil
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

func generateIntegrityReport(result *IntegrityCheckResult, jsonPath string) string {
	var sb strings.Builder

	sb.WriteString("# 数据完整性检查报告\n\n")
	sb.WriteString(fmt.Sprintf("**检查时间**: %s\n\n", result.Timestamp))

	sb.WriteString("## 统计概览\n\n")
	sb.WriteString(fmt.Sprintf("| 项目 | 数量 |\n"))
	sb.WriteString(fmt.Sprintf("|------|------|\n"))
	sb.WriteString(fmt.Sprintf("| JSON 文件数 | %d |\n", result.TotalJSONFiles))
	sb.WriteString(fmt.Sprintf("| JSON 记录数 | %d |\n", result.TotalJSONRecords))
	sb.WriteString(fmt.Sprintf("| 数据库记录数 | %d |\n", result.TotalDBRecords))
	sb.WriteString(fmt.Sprintf("| 数据差异 | %d |\n\n", result.Difference))

	if result.Difference == 0 {
		sb.WriteString("**状态**: 数据一致 ✓\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("**状态**: %s\n\n", result.Message))
	}

	// Per-year breakdown
	sb.WriteString("## 各年份数据统计\n\n")
	sb.WriteString("| 年份 | JSON 记录 | 数据库记录 | 差异 |\n")
	sb.WriteString("|------|-----------|------------|------|\n")

	yearStats := collectYearStats(jsonPath)
	years := make([]int, 0, len(yearStats))
	for y := range yearStats {
		years = append(years, y)
	}
	sort.Ints(years)

	db := database.GetSQLiteDB()
	conn := db.GetDB()

	for _, year := range years {
		jsonCount := yearStats[year]
		dbCount := 0
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		if conn != nil {
			exists, _ := db.TableExists(tableName)
			if exists {
				conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&dbCount)
			}
		}
		diff := jsonCount - dbCount
		diffStr := fmt.Sprintf("%d", diff)
		if diff == 0 {
			diffStr = "0"
		}
		sb.WriteString(fmt.Sprintf("| %d | %d | %d | %s |\n", year, jsonCount, dbCount, diffStr))
	}

	return sb.String()
}

func collectYearStats(jsonPath string) map[int]int {
	yearStats := make(map[int]int)

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return yearStats
	}

	years, err := filepath.Glob(filepath.Join(jsonPath, "*"))
	if err != nil {
		return yearStats
	}

	for _, yearPath := range years {
		yearInfo, err := os.Stat(yearPath)
		if err != nil || !yearInfo.IsDir() {
			continue
		}
		year, err := strconv.Atoi(yearInfo.Name())
		if err != nil {
			continue
		}

		months, err := filepath.Glob(filepath.Join(yearPath, "*"))
		if err != nil {
			continue
		}

		count := 0
		for _, monthPath := range months {
			days, err := filepath.Glob(filepath.Join(monthPath, "*.json"))
			if err != nil {
				continue
			}
			for _, dayPath := range days {
				data, err := os.ReadFile(dayPath)
				if err != nil {
					continue
				}
				var entries []HistoryFileEntry
				if err := json.Unmarshal(data, &entries); err != nil {
					continue
				}
				count += len(entries)
			}
		}
		yearStats[year] = count
	}

	return yearStats
}

func RunSyncData(dbPath, jsonPath string) (*SyncResult, error) {
	status := GetDataSyncStatus()
	if status.IsRunning {
		return nil, fmt.Errorf("同步正在进行中")
	}

	setDataSyncStatus(DataSyncStatus{
		IsRunning: true,
		Status:    "running",
		Message:   "正在同步数据...",
		StartTime: time.Now().Unix(),
	})

	result := syncDataInternal(jsonPath)

	dataSyncMutex.Lock()
	lastSyncResult = result
	dataSyncMutex.Unlock()

	setDataSyncStatus(DataSyncStatus{
		IsRunning:    false,
		Status:       "completed",
		Message:      fmt.Sprintf("同步完成，共导入 %d 条记录", result.JSONToDBCount),
		LastUpdateAt: time.Now().Unix(),
	})

	return result, nil
}

func syncDataInternal(jsonPath string) *SyncResult {
	now := time.Now()
	result := &SyncResult{
		Success:   true,
		Timestamp: now.Format(time.RFC3339),
	}

	db := database.GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		result.Success = false
		result.Message = "数据库未初始化"
		return result
	}

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		result.Message = "JSON目录不存在"
		return result
	}

	years, err := filepath.Glob(filepath.Join(jsonPath, "*"))
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("遍历JSON目录失败: %v", err)
		return result
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

		if err := db.EnsureTableForYear(year); err != nil {
			continue
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
				dayFile := filepath.Base(dayPath)
				dayName := dayFile[:len(dayFile)-5]
				dateStr := fmt.Sprintf("%s-%s-%s", yearStr, monthInfo.Name(), dayName)

				data, err := os.ReadFile(dayPath)
				if err != nil {
					continue
				}

				var entries []HistoryFileEntry
				if err := json.Unmarshal(data, &entries); err != nil {
					continue
				}

				dayImported := 0
				var dayTitles []string

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
						Duration:   entry.Duration,
						AuthorName: entry.AuthorName,
						AuthorFace: entry.AuthorFace,
						AuthorMid:  entry.AuthorMid,
					}

					inserted, err := database.InsertHistoryRecord(conn, tableName, &record)
					if err != nil {
						continue
					}
					if inserted {
						dayImported++
						if len(dayTitles) < 5 {
							dayTitles = append(dayTitles, entry.Title)
						}
					}
				}

				if dayImported > 0 {
					result.SyncedDays = append(result.SyncedDays, SyncedDayDetail{
						Date:          dateStr,
						ImportedCount: dayImported,
						Source:        "json_to_db",
						Titles:        dayTitles,
					})
					result.JSONToDBCount += dayImported
				}
			}
		}
	}

	result.TotalSynced = result.JSONToDBCount
	result.DBToJSONCount = 0

	if result.JSONToDBCount == 0 {
		result.Message = "数据已是最新，无需同步"
	} else {
		result.Message = fmt.Sprintf("同步完成，共导入 %d 条记录", result.JSONToDBCount)
	}

	return result
}
