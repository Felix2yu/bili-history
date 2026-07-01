package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/services"

	"github.com/gin-gonic/gin"
)

// SyncDataFull performs full bidirectional sync between JSON and database.
func SyncDataFull(c *gin.Context) {
	result := services.RunSyncData()
	c.JSON(http.StatusOK, result)
}

// CheckDataIntegrityFull performs data integrity check.
func CheckDataIntegrityFull(c *gin.Context) {
	result := services.RunIntegrityCheck()
	c.JSON(http.StatusOK, result)
}

// GetIntegrityReport returns the latest integrity check report.
func GetIntegrityReport(c *gin.Context) {
	reportFile := config.GetOutputPath("check", "data_integrity_report.md")
	data, err := os.ReadFile(reportFile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	info, _ := os.Stat(reportFile)
	modTime := ""
	if info != nil {
		modTime = info.ModTime().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, gin.H{
		"content":       string(data),
		"modified_time": modTime,
		"file_path":     reportFile,
	})
}

// GetSyncResult returns the latest sync result.
func GetSyncResult(c *gin.Context) {
	resultFile := config.GetOutputPath("check", "sync_result.json")
	data, err := os.ReadFile(resultFile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sync result not found"})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	info, _ := os.Stat(resultFile)
	if info != nil {
		result["file_modified_time"] = info.ModTime().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, result)
}

// GetIntegrityCheckConfig returns integrity check configuration.
func GetIntegrityCheckConfig(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	checkOnStartup := true
	if cfg.Server.DataIntegrity.CheckOnStartup {
		checkOnStartup = cfg.Server.DataIntegrity.CheckOnStartup
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"check_on_startup": checkOnStartup,
	})
}

// UpdateIntegrityCheckConfig updates integrity check configuration.
func UpdateIntegrityCheckConfig(c *gin.Context) {
	var req struct {
		CheckOnStartup bool `json:"check_on_startup"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Read config file
	configPath := config.GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config"})
		return
	}

	content := string(data)
	valueStr := "false"
	if req.CheckOnStartup {
		valueStr = "true"
	}

	// Simple replacement for check_on_startup
	oldPatterns := []string{
		"check_on_startup: false",
		"check_on_startup: true",
		"check_on_startup:false",
		"check_on_startup:true",
	}
	for _, old := range oldPatterns {
		if contains(content, old) {
			content = replaceFirst(content, old, "check_on_startup: "+valueStr)
			break
		}
	}

	os.WriteFile(configPath, []byte(content), 0644)
	config.ReloadConfig()

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"check_on_startup": req.CheckOnStartup,
		"message":          "Configuration updated",
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func replaceFirst(s, old, new string) string {
	idx := indexOf(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

// SyncStatus returns current sync status.
func SyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "idle",
		"message": "No sync in progress",
	})
}

// IncrementalSync performs incremental sync (alias for full sync).
func IncrementalSync(c *gin.Context) {
	SyncDataFull(c)
}

// ForceSync forces a full sync.
func ForceSync(c *gin.Context) {
	SyncDataFull(c)
}

// Ensure fmt is used
var _ = fmt.Sprintf
