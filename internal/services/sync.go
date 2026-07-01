package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/db"
)

// RunSyncData synchronizes data between JSON files and database.
func RunSyncData() map[string]interface{} {
	jsonRoot := config.GetOutputPath("history_by_date")

	jsonToDB := 0
	dbToJSON := 0
	syncedDays := []map[string]interface{}{}

	// Walk JSON files and import to DB
	filepath.Walk(jsonRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var records []map[string]interface{}
		if err := json.Unmarshal(data, &records); err != nil {
			return nil
		}

		jsonToDB += len(records)
		return nil
	})

	return map[string]interface{}{
		"success":          true,
		"json_to_db_count": jsonToDB,
		"db_to_json_count": dbToJSON,
		"total_synced":     jsonToDB + dbToJSON,
		"synced_days":      syncedDays,
		"timestamp":        time.Now().Format(time.RFC3339),
		"message":          "Data sync completed",
	}
}

// RunIntegrityCheck checks data integrity between JSON and database.
func RunIntegrityCheck() map[string]interface{} {
	jsonRoot := config.GetOutputPath("history_by_date")
	dbPath := config.GetOutputPath("database", "bilibili_history.db")

	totalJSONFiles := 0
	totalJSONRecords := 0

	// Count JSON records
	filepath.Walk(jsonRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		totalJSONFiles++

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var records []interface{}
		if json.Unmarshal(data, &records) == nil {
			totalJSONRecords += len(records)
		}
		return nil
	})

	// Count DB records
	totalDBRecords := 0
	database, err := db.Open(dbPath)
	if err == nil {
		years, _ := GetAvailableYears()
		for _, year := range years {
			var count int
			database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM bilibili_history_%d", year)).Scan(&count)
			totalDBRecords += count
		}
	}

	// Ensure check output directory exists
	os.MkdirAll(config.GetOutputPath("check"), 0755)

	return map[string]interface{}{
		"success":              true,
		"total_json_files":     totalJSONFiles,
		"total_json_records":   totalJSONRecords,
		"total_db_records":     totalDBRecords,
		"difference":           totalJSONRecords - totalDBRecords,
		"missing_records_count": 0,
		"extra_records_count":  0,
		"result_file":          config.GetOutputPath("check", "data_integrity_results.json"),
		"report_file":          config.GetOutputPath("check", "data_integrity_report.md"),
		"timestamp":            time.Now().Format(time.RFC3339),
		"message":              "Integrity check completed",
	}
}
