package services

import (
	"fmt"
	"net/url"

	"bilibili-history-go/config"
	"bilibili-history-go/utils"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
)

var shoutrrrRouter *router.ServiceRouter

func getShoutrrrRouter() (*router.ServiceRouter, error) {
	if shoutrrrRouter != nil {
		return shoutrrrRouter, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config error: %w", err)
	}

	if !cfg.Shoutrrr.Enabled || len(cfg.Shoutrrr.URLs) == 0 {
		return nil, fmt.Errorf("shoutrrr not configured")
	}

	validURLs := make([]string, 0, len(cfg.Shoutrrr.URLs))
	for _, raw := range cfg.Shoutrrr.URLs {
		if _, err := url.Parse(raw); err != nil {
			utils.LogWarning("跳过无效的Shoutrrr URL: %s, error: %v", raw, err)
			continue
		}
		validURLs = append(validURLs, raw)
	}

	if len(validURLs) == 0 {
		return nil, fmt.Errorf("no valid shoutrrr URLs")
	}

	r, err := shoutrrr.CreateSender(validURLs...)
	if err != nil {
		return nil, fmt.Errorf("create shoutrrr sender error: %w", err)
	}

	shoutrrrRouter = r
	return shoutrrrRouter, nil
}

func SendShoutrrrNotification(title, message string) error {
	r, err := getShoutrrrRouter()
	if err != nil {
		return err
	}

	body := title
	if message != "" {
		body = title + "\n" + message
	}

	errors := r.Send(body, nil)
	if len(errors) > 0 {
		var errMsgs []string
		for i, e := range errors {
			if e != nil {
				errMsgs = append(errMsgs, fmt.Sprintf("url[%d]: %v", i, e))
			}
		}
		if len(errMsgs) > 0 {
			return fmt.Errorf("shoutrrr send errors: %v", errMsgs)
		}
	}

	utils.LogSuccess("Shoutrrr通知发送成功: %s", title)
	return nil
}

func SendTestShoutrrr() error {
	title := "Bilibili历史记录管理 - 测试通知"
	message := "这是一条测试通知，Shoutrrr配置正确。"
	return SendShoutrrrNotification(title, message)
}

func SendDailyReport(stats map[string]interface{}) error {
	title := "📊 Bilibili历史记录每日报告"

	var message string
	if totalRecords, ok := stats["total_records"]; ok {
		message += fmt.Sprintf("总记录数：%v\n", totalRecords)
	}
	if todayRecords, ok := stats["today_records"]; ok {
		message += fmt.Sprintf("今日观看：%v 条\n", todayRecords)
	}
	if totalWatchingTime, ok := stats["total_watching_time"]; ok {
		message += fmt.Sprintf("总观看时长：%v\n", totalWatchingTime)
	}
	if mostActiveDay, ok := stats["most_active_day"]; ok {
		message += fmt.Sprintf("最活跃日期：%v\n", mostActiveDay)
	}

	return SendShoutrrrNotification(title, message)
}

func ResetShoutrrrRouter() {
	shoutrrrRouter = nil
}
