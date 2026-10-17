package utils

import (
	"os"
	"path/filepath"

	"bilibili-history-go/internal/config"
)

func GetBasePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

// GetOutputPath 只负责拼路径并顺手建目录。目录建不出来的情况会紧接着
// 暴露在真正读写的调用点上（那里才有错误出口），所以这里显式丢弃。
func GetOutputPath(paths ...string) string {
	basePath := GetBasePath()
	outputDir := filepath.Join(basePath, "output")
	_ = os.MkdirAll(outputDir, 0755)

	fullPath := filepath.Join(append([]string{outputDir}, paths...)...)
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

	return fullPath
}

func GetDatabasePath(paths ...string) string {
	basePath := GetBasePath()
	databaseDir := filepath.Join(basePath, "output", "database")
	_ = os.MkdirAll(databaseDir, 0755)

	fullPath := filepath.Join(append([]string{databaseDir}, paths...)...)
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

	return fullPath
}

func GetDBFilePath() string {
	cfg := config.GetConfig()
	if cfg == nil {
		return GetOutputPath("bilibili_history.db")
	}
	return GetOutputPath(cfg.DBFile)
}
