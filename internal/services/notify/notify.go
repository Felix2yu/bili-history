package notify

import (
	"fmt"
	"net/smtp"
	"time"

	"bili-history/internal/config"
)

// SendEmail sends an email notification.
func SendEmail(subject, body string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Email.SMTPServer == "" || cfg.Email.Sender == "" {
		return fmt.Errorf("email not configured")
	}

	auth := smtp.PlainAuth("", cfg.Email.Sender, cfg.Email.Password, cfg.Email.SMTPServer)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n",
		cfg.Email.Sender, cfg.Email.Receiver, subject, body))

	addr := fmt.Sprintf("%s:%d", cfg.Email.SMTPServer, cfg.Email.SMTPPort)
	return smtp.SendMail(addr, auth, cfg.Email.Sender, []string{cfg.Email.Receiver}, msg)
}

// SendLogEmail sends a log summary email.
func SendLogEmail() error {
	subject := fmt.Sprintf("BiliHistory Daily Report - %s", time.Now().Format("2006-01-02"))

	// Build simple report
	cfg, _ := config.LoadConfig()
	body := fmt.Sprintf("BiliHistory Daily Report\n"+
		"Generated at: %s\n"+
		"SESSDATA configured: %v\n",
		time.Now().Format("2006-01-02 15:04:05"),
		cfg != nil && cfg.SESSDATA != "" && cfg.SESSDATA != "Cookie里的SESSDATA字段值")

	return SendEmail(subject, body)
}

// SendNotification sends a push notification via Apprise.
func SendNotification(title, message string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	if !cfg.Apprise.Enabled || len(cfg.Apprise.URLs) == 0 {
		return nil // Apprise not configured
	}

	// Send to each Apprise URL
	for _, url := range cfg.Apprise.URLs {
		if url == "" {
			continue
		}
		// TODO: Implement Apprise webhook notification
		fmt.Printf("Would send notification to %s: %s - %s\n", url, title, message)
	}

	return nil
}
