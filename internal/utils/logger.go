package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelSuccess
)

var levelNames = map[LogLevel]string{
	LevelDebug:   "DEBUG",
	LevelInfo:    "INFO",
	LevelWarning: "WARNING",
	LevelError:   "ERROR",
	LevelSuccess: "SUCCESS",
}

type Logger struct {
	mu          sync.Mutex
	logFile     *os.File
	errorFile   *os.File
	currentDate string
	logDir      string
}

var (
	loggerInstance *Logger
	loggerOnce     sync.Once
)

func GetLogger() *Logger {
	loggerOnce.Do(func() {
		loggerInstance = &Logger{}
		loggerInstance.init()
	})
	return loggerInstance
}

func (l *Logger) init() {
	now := Now()
	l.currentDate = now.Format("2006-01-02")
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	l.logDir = filepath.Join("output", "logs", year, month, day)
	// 目录建不出来时，下面的 OpenFile 会把真实原因打到 stderr，这里不重复报
	_ = os.MkdirAll(l.logDir, 0755)

	logPath := filepath.Join(l.logDir, day+".log")
	errorPath := filepath.Join(l.logDir, "error_"+day+".log")

	var err error
	l.logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
	}

	l.errorFile, err = os.OpenFile(errorPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open error log file: %v\n", err)
	}
}

func (l *Logger) checkDate() {
	now := Now()
	today := now.Format("2006-01-02")

	if today != l.currentDate {
		l.mu.Lock()
		defer l.mu.Unlock()

		if today != l.currentDate {
			if l.logFile != nil {
				l.logFile.Close()
			}
			if l.errorFile != nil {
				l.errorFile.Close()
			}
			l.init()
		}
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.checkDate()

	l.mu.Lock()
	defer l.mu.Unlock()

	now := Now()
	timestamp := now.Format("2006-01-02 15:04:05")
	levelName := levelNames[level]

	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, levelName, message)

	// 日志写入是最后一道防线：失败时既不能递归记日志，也不该往 stderr 刷屏，
	// 下一个日志行还会重试，因此显式丢弃。
	if l.logFile != nil {
		_, _ = l.logFile.WriteString(logLine)
	}

	if level >= LevelError && l.errorFile != nil {
		_, _ = l.errorFile.WriteString(logLine)
	}

	if level == LevelError || level == LevelWarning || level == LevelSuccess {
		fmt.Print(logLine)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(LevelWarning, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

func (l *Logger) Success(format string, args ...interface{}) {
	l.log(LevelSuccess, format, args...)
}

func LogInfo(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func LogWarning(format string, args ...interface{}) {
	GetLogger().Warning(format, args...)
}

func LogError(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func LogSuccess(format string, args ...interface{}) {
	GetLogger().Success(format, args...)
}

// 日志器不在包初始化时启动：GetLogger 本来就用 sync.Once 懒建，
// 保留一个 init() 反而让「任何导入本包的进程」都在当前工作目录建出 output/logs 并开文件。
// `bili-history backup` 是经 docker exec 调起来的（默认身份是 root，应用本身跑在 1000），
// 它只读地打开数据库往外写快照，不该在挂载的数据卷里留下属主错误的日志目录与空文件。
