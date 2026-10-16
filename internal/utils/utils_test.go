package utils

import (
	"strings"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	now := Now()
	if now.IsZero() {
		t.Errorf("Now() 返回零值")
	}
	if time.Since(now) > time.Minute {
		t.Errorf("Now() 与当前时间偏差过大: %v", now)
	}
}

func TestTimestampToTime(t *testing.T) {
	got := TimestampToTime(0)
	if got.Year() != 1970 || got.Month() != 1 || got.Day() != 1 {
		t.Errorf("TimestampToTime(0) 期望 1970-01-01, 实际 %v", got)
	}

	known := TimestampToTime(1700000000)
	if known.Year() != 2023 {
		t.Errorf("TimestampToTime(1700000000) 期望年份 2023, 实际 %d", known.Year())
	}
}

func TestFormatDateTime(t *testing.T) {
	tm := time.Date(2023, 11, 15, 8, 30, 45, 0, time.UTC)
	got := FormatDateTime(tm)
	want := "2023-11-15 08:30:45"
	if got != want {
		t.Errorf("FormatDateTime 期望 %q, 实际 %q", want, got)
	}
}

func TestGetYearFromTimestamp(t *testing.T) {
	if got := GetYearFromTimestamp(0); got != 1970 {
		t.Errorf("GetYearFromTimestamp(0) 期望 1970, 实际 %d", got)
	}
	if got := GetYearFromTimestamp(1700000000); got != 2023 {
		t.Errorf("GetYearFromTimestamp(1700000000) 期望 2023, 实际 %d", got)
	}
}

func TestGetBasePath(t *testing.T) {
	p := GetBasePath()
	if p == "" {
		t.Errorf("GetBasePath 返回空字符串")
	}
}

func TestGetOutputPath(t *testing.T) {
	p := GetOutputPath("sub", "file.txt")
	if !strings.Contains(p, "output") {
		t.Errorf("GetOutputPath 结果应包含 output, 实际 %q", p)
	}
	if !strings.HasSuffix(p, "file.txt") {
		t.Errorf("GetOutputPath 结果应以 file.txt 结尾, 实际 %q", p)
	}
}

func TestGetDatabasePath(t *testing.T) {
	p := GetDatabasePath("bilibili_history.db")
	if !strings.Contains(p, "database") {
		t.Errorf("GetDatabasePath 结果应包含 database, 实际 %q", p)
	}
	if !strings.HasSuffix(p, "bilibili_history.db") {
		t.Errorf("GetDatabasePath 结果应以 bilibili_history.db 结尾, 实际 %q", p)
	}
}

func TestGetDBFilePath(t *testing.T) {
	p := GetDBFilePath()
	if p == "" {
		t.Errorf("GetDBFilePath 返回空字符串")
	}
	if !strings.Contains(p, "output") {
		t.Errorf("GetDBFilePath 结果应包含 output, 实际 %q", p)
	}
}
