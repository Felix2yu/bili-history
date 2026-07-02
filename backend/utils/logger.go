package utils

import (
	"fmt"
	"io"
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
	os.MkdirAll(l.logDir, 0755)
	
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
	
	if l.logFile != nil {
		l.logFile.WriteString(logLine)
	}
	
	if level >= LevelError && l.errorFile != nil {
		l.errorFile.WriteString(logLine)
	}
	
	if level == LevelError || level == LevelWarning || level == LevelSuccess {
		fmt.Print(logLine)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
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

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}
	if l.errorFile != nil {
		l.errorFile.Close()
		l.errorFile = nil
	}
}

func LogInfo(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func LogDebug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
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

func init() {
	GetLogger()
}

var _ io.Writer = (*logWriter)(nil)

type logWriter struct {
	level LogLevel
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	GetLogger().log(w.level, "%s", string(p))
	return len(p), nil
}
