package utils

import "testing"

func TestLoggerInstance(t *testing.T) {
	l := GetLogger()
	if l == nil {
		t.Fatalf("GetLogger 返回 nil")
	}
	// 再次调用应返回同一实例
	if GetLogger() != l {
		t.Errorf("GetLogger 未返回单例实例")
	}
}

func TestLoggerLevels(t *testing.T) {
	l := GetLogger()
	l.Info("信息 %d", 1)
	l.Warning("警告 %s", "w")
	l.Error("错误 %s", "e")
	l.Success("成功 %s", "ok")
}

func TestPackageLevelLoggers(t *testing.T) {
	LogInfo("pkg info %d", 1)
	LogWarning("pkg warning")
	LogError("pkg error %s", "x")
	LogSuccess("pkg success")
}
