package services

import (
	"database/sql"
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
	if topAuthor, ok := stats["top_author"]; ok {
		message += fmt.Sprintf("最常看UP主：%v\n", topAuthor)
	}
	if topCategory, ok := stats["top_category"]; ok {
		message += fmt.Sprintf("最常看分区：%v\n", topCategory)
	}
	if topTag, ok := stats["top_tag"]; ok {
		message += fmt.Sprintf("最常看标签：%v\n", topTag)
	}
	if peakHour, ok := stats["peak_hour"]; ok {
		message += fmt.Sprintf("观看高峰时段：%v\n", peakHour)
	}

	// 如果仍然没有内容，显示默认提示
	if message == "" {
		message = "今日暂无观看记录\n"
	}
	// 补充年度统计和最近观看信息（今日无数据时展示）
	if yearTotal, ok := stats["year_total"]; ok {
		message += fmt.Sprintf("本年累计观看：%v 条\n", yearTotal)
	}
	if lastView, ok := stats["last_view_time"]; ok {
		ago := ""
		if a, ok := stats["last_view_ago"]; ok {
			ago = fmt.Sprintf("（%v）", a)
		}
		message += fmt.Sprintf("最近一次观看：%v%v\n", lastView, ago)
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
		utils.LogWarning("每日报告: 数据库不可用")
		data["today_records"] = 0
		data["total_watching_time"] = "0分钟"
		data["report_date"] = now.Format("2006-01-02")
		return data
	}
	conn := db.GetDB()
	if conn == nil {
		utils.LogWarning("每日报告: 数据库连接不可用")
		data["today_records"] = 0
		data["total_watching_time"] = "0分钟"
		data["report_date"] = now.Format("2006-01-02")
		return data
	}

	tableName := fmt.Sprintf("bilibili_history_%s", year)
	exists, _ := db.TableExists(tableName)
	if !exists {
		data["today_records"] = 0
		data["total_watching_time"] = "0分钟"
		data["report_date"] = now.Format("2006-01-02")
		utils.LogInfo("每日报告: 表 %s 不存在", tableName)
		return data
	}

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Unix()
	todayEnd := todayStart + 86400

	// 今日观看数
	var todayCount int
	err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE view_at >= ? AND view_at < ?", tableName), todayStart, todayEnd).Scan(&todayCount)
	if err != nil {
		utils.LogWarning("每日报告: 查询今日观看数失败: %v", err)
		todayCount = 0
	}
	data["today_records"] = todayCount

	// 今日观看时长
	var todayDuration int
	err = conn.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(
			CASE
				WHEN progress = -1 THEN duration
				WHEN progress >= 0 AND progress > duration THEN duration
				WHEN progress >= 0 THEN progress
				ELSE 0
			END
		), 0)
		FROM %s WHERE view_at >= ? AND view_at < ?
	`, tableName), todayStart, todayEnd).Scan(&todayDuration)
	if err != nil {
		utils.LogWarning("每日报告: 查询今日观看时长失败: %v", err)
		todayDuration = 0
	}
	hours := todayDuration / 3600
	minutes := (todayDuration % 3600) / 60
	if hours > 0 {
		data["total_watching_time"] = fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		data["total_watching_time"] = fmt.Sprintf("%d分钟", minutes)
	}

	// 如果今日无数据，补充本年总计和最近观看信息
	if todayCount == 0 {
		var yearTotal int
		_ = conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&yearTotal)
		data["year_total"] = yearTotal

		// 最近一次观看时间
		var lastViewAt int64
		err = conn.QueryRow(fmt.Sprintf("SELECT MAX(view_at) FROM %s", tableName)).Scan(&lastViewAt)
		if err == nil && lastViewAt > 0 {
			lastTime := time.Unix(lastViewAt, 0)
			data["last_view_time"] = lastTime.Format("2006-01-02 15:04")
			daysSince := int(now.Sub(lastTime).Hours() / 24)
			if daysSince == 0 {
				data["last_view_ago"] = "今天"
			} else {
				data["last_view_ago"] = fmt.Sprintf("%d天前", daysSince)
			}
		}
	} else {
		// 今日最常看UP主 TOP3
		topAuthors := queryTopN(conn, tableName, todayStart, todayEnd, "author_mid", "author_name")
		if len(topAuthors) > 0 {
			data["top_author"] = formatTopN(topAuthors)
		}

		// 今日最常看分区 TOP3
		topCategories := queryTopN(conn, tableName, todayStart, todayEnd, "main_category", "main_category")
		if len(topCategories) > 0 {
			data["top_category"] = formatTopN(topCategories)
		}

		// 今日最常看标签
		var topTag string
		err = conn.QueryRow(fmt.Sprintf("SELECT tag_name FROM %s WHERE view_at >= ? AND view_at < ? AND tag_name IS NOT NULL AND tag_name != '' GROUP BY tag_name ORDER BY COUNT(*) DESC LIMIT 1", tableName), todayStart, todayEnd).Scan(&topTag)
		if err == nil && topTag != "" {
			data["top_tag"] = topTag
		}

		// 今日观看高峰时段
		var peakHour int
		err = conn.QueryRow(fmt.Sprintf("SELECT CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) as hour FROM %s WHERE view_at >= ? AND view_at < ? GROUP BY hour ORDER BY COUNT(*) DESC LIMIT 1", tableName), todayStart, todayEnd).Scan(&peakHour)
		if err == nil {
			data["peak_hour"] = fmt.Sprintf("%d:00-%d:59", peakHour, peakHour)
		}
	}

	data["report_date"] = now.Format("2006-01-02")
	utils.LogInfo("每日报告: 今日观看 %d 条, 时长 %d 分钟", todayCount, todayDuration)
	return data
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

// queryTopN queries top N items by count, returning a list of {group_key, display_name} pairs.
func queryTopN(conn *sql.DB, tableName string, start, end int64, groupCol, displayCol string) []string {
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE view_at >= ? AND view_at < ?
		AND %s IS NOT NULL AND %s != ''
		GROUP BY %s
		ORDER BY COUNT(*) DESC
		LIMIT 3
	`, displayCol, tableName, groupCol, groupCol, groupCol)

	rows, err := conn.Query(query, start, end)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil && name != "" {
			results = append(results, name)
		}
	}
	return results
}

func formatTopN(items []string) string {
	var result string
	for i, name := range items {
		result += fmt.Sprintf("%d.%s ", i+1, name)
	}
	return result
}
