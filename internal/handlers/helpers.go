package handlers

import (
	"strings"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/services/biliapi"
)

func newBiliClient(cfg *config.Config) *biliapi.Client {
	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	if cfg.DedeUserIDCkMd5 != "" {
		client.SetDedeUserIDCkMd5(cfg.DedeUserIDCkMd5)
	}
	return client
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func containsPath(path, substr string) bool {
	return strings.Contains(path, substr)
}

func indexOf(s, substr string) int {
	return strings.Index(s, substr)
}
