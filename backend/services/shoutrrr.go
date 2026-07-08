package services

import (
	"fmt"
	"net/url"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
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
	return SendShoutrrrNotificationWithParams(title, message, nil)
}

func SendShoutrrrNotificationWithParams(title, message string, params *types.Params) error {
	r, err := getShoutrrrRouter()
	if err != nil {
		return err
	}

	body := title
	if message != "" {
		body = title + "\n" + message
	}

	errors := r.Send(body, params)
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
	// 如果 stats 为空或只有 report_date，自动从数据库获取数据
	if len(stats) <= 1 {
		stats = gatherDailyReportData()
	}

	reportDate, _ := stats["report_date"].(string)
	if reportDate == "" {
		reportDate = time.Now().Format("2006-01-02")
	}
	title := fmt.Sprintf("📊 Bilibili日报 %s", reportDate)

	var message string
	if todayRecords, ok := stats["today_records"]; ok {
		message += fmt.Sprintf("今日观看：%v 条\n", todayRecords)
	}
	if watchingTime, ok := stats["total_watching_time"]; ok {
		message += fmt.Sprintf("观看时长：%v\n", watchingTime)
	}

	utils.LogInfo("发送每日报告: title=%s, message=%q", title, message)
	err := SendShoutrrrNotification(title, message)
	if err != nil {
		utils.LogError("发送每日报告失败: %v", err)
	} else {
		utils.LogSuccess("每日报告发送成功")
	}
	return err
}

func gatherDailyReportData() map[string]interface{} {
	data := make(map[string]interface{})
	now := time.Now()
	year := now.Format("2006")

	db := database.GetSQLiteDB()
	if db == nil {
		return data
	}
	conn := db.GetDB()
	if conn == nil {
		return data
	}

	tableName := fmt.Sprintf("bilibili_history_%s", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		data["today_records"] = 0
		data["total_watching_time"] = "0分钟"
		data["report_date"] = now.Format("2006-01-02")
		return data
	}

	// 今日观看数
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Unix()
	todayEnd := todayStart + 86400
	var todayCount int
	conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE view_at >= ? AND view_at < ?", tableName), todayStart, todayEnd).Scan(&todayCount)
	data["today_records"] = todayCount

	// 今日观看时长
	var todayDuration int
	conn.QueryRow(fmt.Sprintf("SELECT COALESCE(SUM(duration), 0) FROM %s WHERE view_at >= ? AND view_at < ?", tableName), todayStart, todayEnd).Scan(&todayDuration)
	hours := todayDuration / 3600
	minutes := (todayDuration % 3600) / 60
	if hours > 0 {
		data["total_watching_time"] = fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		data["total_watching_time"] = fmt.Sprintf("%d分钟", minutes)
	}

	data["report_date"] = now.Format("2006-01-02")
	return data
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func SendSessdataExpiredNotification(username string) error {
	title := "⚠️ Bilibili历史记录管理 - SESSDATA 已过期"
	message := "您的 SESSDATA 已失效，请重新登录。\n"
	if username != "" {
		message += fmt.Sprintf("上次登录用户：%s\n", username)
	}
	message += "请在前端设置页面重新扫码登录，否则历史记录同步等功能将无法正常使用。"
	return SendShoutrrrNotification(title, message)
}

func ResetShoutrrrRouter() {
	shoutrrrRouter = nil
}
