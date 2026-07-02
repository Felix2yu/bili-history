package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"bilibili-history-go/config"
	"bilibili-history-go/utils"
)

func SendEmail(to, subject, body string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config error: %w", err)
	}

	if cfg.Email.SMTPServer == "" || cfg.Email.Sender == "" {
		return fmt.Errorf("email not configured")
	}

	from := cfg.Email.Sender
	password := cfg.Email.Password
	smtpServer := cfg.Email.SMTPServer
	smtpPort := cfg.Email.SMTPPort

	if to == "" {
		to = cfg.Email.Receiver
	}
	if to == "" {
		return fmt.Errorf("no recipient specified")
	}

	recipients := strings.Split(to, ",")
	for i, r := range recipients {
		recipients[i] = strings.TrimSpace(r)
	}

	mimeHeaders := make(map[string]string)
	mimeHeaders["From"] = from
	mimeHeaders["To"] = strings.Join(recipients, ", ")
	mimeHeaders["Subject"] = subject
	mimeHeaders["MIME-Version"] = "1.0"
	mimeHeaders["Content-Type"] = "text/html; charset=UTF-8"

	var message strings.Builder
	for key, value := range mimeHeaders {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	auth := smtp.PlainAuth("", from, password, smtpServer)

	addr := fmt.Sprintf("%s:%d", smtpServer, smtpPort)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial error: %w", err)
	}

	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		return fmt.Errorf("smtp client error: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth error: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail error: %w", err)
	}

	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt error for %s: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data error: %w", err)
	}

	_, err = w.Write([]byte(message.String()))
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close error: %w", err)
	}

	utils.LogSuccess("邮件发送成功: %s -> %s", subject, to)
	return nil
}

func SendTestEmail() error {
	subject := "Bilibili历史记录管理 - 测试邮件"
	body := `
	<html>
		<body style="font-family: Arial, sans-serif;">
			<h2>测试邮件</h2>
			<p>这是一封来自Bilibili历史记录管理系统的测试邮件。</p>
			<p>如果您收到了这封邮件，说明邮件配置正确。</p>
			<hr>
			<p style="color: #888; font-size: 12px;">此邮件由系统自动发送，请勿直接回复。</p>
		</body>
	</html>
	`
	return SendEmail("", subject, body)
}

func SendDailyReport(stats map[string]interface{}) error {
	subject := "Bilibili历史记录 - 每日报告"

	var body strings.Builder
	body.WriteString(`
	<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2 style="color: #FB7299;">📊 Bilibili历史记录每日报告</h2>
	`)

	if totalRecords, ok := stats["total_records"]; ok {
		body.WriteString(fmt.Sprintf("<p><strong>总记录数：</strong>%v</p>", totalRecords))
	}
	if todayRecords, ok := stats["today_records"]; ok {
		body.WriteString(fmt.Sprintf("<p><strong>今日观看：</strong>%v 条</p>", todayRecords))
	}
	if totalWatchingTime, ok := stats["total_watching_time"]; ok {
		body.WriteString(fmt.Sprintf("<p><strong>总观看时长：</strong>%v</p>", totalWatchingTime))
	}
	if mostActiveDay, ok := stats["most_active_day"]; ok {
		body.WriteString(fmt.Sprintf("<p><strong>最活跃日期：</strong>%v</p>", mostActiveDay))
	}

	body.WriteString(`
			<hr>
			<p style="color: #888; font-size: 12px;">此邮件由Bilibili历史记录管理系统自动发送，请勿直接回复。</p>
		</body>
	</html>
	`)

	return SendEmail("", subject, body.String())
}
