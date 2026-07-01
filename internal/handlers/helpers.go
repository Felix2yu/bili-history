package handlers

import (
	"bili-history/internal/config"
	"bili-history/internal/services/biliapi"
)

func newBiliClient(cfg *config.Config) *biliapi.Client {
	return biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
}
