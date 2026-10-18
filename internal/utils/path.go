package utils

import (
	"os"
	"path/filepath"
	"strings"

	"bilibili-history-go/internal/config"
)

// defaultDBFile 是 config db_file 留空时的主库文件名，与 Docker/本地默认布局一致。
const defaultDBFile = "bilibili_history.db"

// GetBasePath 返回应用的工作目录，全部数据路径都以它为根（容器里即 WORKDIR /app）。
func GetBasePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

// OutputDir 返回数据输出目录，DatabaseDir 返回它下面的 database 子目录。
//
// 这两个是「只问在哪、不动磁盘」的版本：GetOutputPath / GetDatabasePath 会顺手建目录，
// 备份导出这类只读进程用它们就可能在数据目录里凭空造出结构来，所以路径约定只写在这里，
// 两边共用同一份定义。
func OutputDir() string {
	return filepath.Join(GetBasePath(), "output")
}

func DatabaseDir() string {
	return filepath.Join(OutputDir(), "database")
}

// DBFileName 返回主库文件名（config 的 db_file）。
//
// 配置读不出来或没填时回退默认名：早期实现返回空串，会让 GetDBFilePath 退化成
// output 目录本身，打开的其实是个目录——一个都不该发生的失败模式。
func DBFileName() string {
	cfg := config.GetConfig()
	if cfg == nil || strings.TrimSpace(cfg.DBFile) == "" {
		return defaultDBFile
	}
	return cfg.DBFile
}

// DBFilePath 返回主库的完整路径，不创建任何目录。
func DBFilePath() string {
	return filepath.Join(OutputDir(), DBFileName())
}

// GetOutputPath 只负责拼路径并顺手建目录。目录建不出来的情况会紧接着
// 暴露在真正读写的调用点上（那里才有错误出口），所以这里显式丢弃。
func GetOutputPath(paths ...string) string {
	outputDir := OutputDir()
	_ = os.MkdirAll(outputDir, 0755)

	fullPath := filepath.Join(append([]string{outputDir}, paths...)...)
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

	return fullPath
}

func GetDatabasePath(paths ...string) string {
	databaseDir := DatabaseDir()
	_ = os.MkdirAll(databaseDir, 0755)

	fullPath := filepath.Join(append([]string{databaseDir}, paths...)...)
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

	return fullPath
}

func GetDBFilePath() string {
	return GetOutputPath(DBFileName())
}
