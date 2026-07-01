package utils

import (
	"os"
	"path/filepath"

	"bilibili-history-go/config"
)

func GetBasePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func GetOutputPath(paths ...string) string {
	basePath := GetBasePath()
	outputDir := filepath.Join(basePath, "output")
	os.MkdirAll(outputDir, 0755)
	
	fullPath := filepath.Join(append([]string{outputDir}, paths...)...)
	os.MkdirAll(filepath.Dir(fullPath), 0755)
	
	return fullPath
}

func GetDatabasePath(paths ...string) string {
	basePath := GetBasePath()
	databaseDir := filepath.Join(basePath, "output", "database")
	os.MkdirAll(databaseDir, 0755)
	
	fullPath := filepath.Join(append([]string{databaseDir}, paths...)...)
	os.MkdirAll(filepath.Dir(fullPath), 0755)
	
	return fullPath
}

func GetLogsPath() string {
	now := Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	
	logPath := filepath.Join("output", "logs", year, month, day, day+".log")
	os.MkdirAll(filepath.Dir(logPath), 0755)
	
	return logPath
}

func GetDBFilePath() string {
	cfg := config.GetConfig()
	if cfg == nil {
		return GetOutputPath("bilibili_history.db")
	}
	return GetOutputPath(cfg.DBFile)
}
