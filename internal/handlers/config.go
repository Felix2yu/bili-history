package handlers

import (
	"net/http"
	"os"

	"bili-history/internal/config"

	"github.com/gin-gonic/gin"
)

// GetEmailConfig returns email configuration.
func GetEmailConfig(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"smtp_server": cfg.Email.SMTPServer,
		"smtp_port":   cfg.Email.SMTPPort,
		"sender":      cfg.Email.Sender,
		"receiver":    cfg.Email.Receiver,
		// Don't expose password
	})
}

// UpdateEmailConfig updates email configuration.
func UpdateEmailConfig(c *gin.Context) {
	var req struct {
		SMTPServer string `json:"smtp_server"`
		SMTPPort   int    `json:"smtp_port"`
		Sender     string `json:"sender"`
		Password   string `json:"password"`
		Receiver   string `json:"receiver"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update config file
	configPath := config.GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config"})
		return
	}

	content := string(data)
	// Simple field updates
	if req.SMTPServer != "" {
		content = updateYAMLField(content, "smtp_server", req.SMTPServer)
	}
	if req.SMTPPort > 0 {
		content = updateYAMLInt(content, "smtp_port", req.SMTPPort)
	}
	if req.Sender != "" {
		content = updateYAMLField(content, "sender", req.Sender)
	}
	if req.Password != "" {
		content = updateYAMLField(content, "password", req.Password)
	}
	if req.Receiver != "" {
		content = updateYAMLField(content, "receiver", req.Receiver)
	}

	os.WriteFile(configPath, []byte(content), 0644)
	config.ReloadConfig()

	c.JSON(http.StatusOK, gin.H{"message": "Email config updated"})
}

// TestEmail sends a test email.
func TestEmail(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	if cfg.Email.SMTPServer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email not configured"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent (placeholder - implement SMTP)",
		"sender":  cfg.Email.Sender,
		"receiver": cfg.Email.Receiver,
	})
}

// GetMCPConfig returns MCP configuration.
func GetMCPConfig(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	mcpEnabled := false
	mcpToken := ""
	_ = cfg // MCP config is from env var

	c.JSON(http.StatusOK, gin.H{
		"enabled": mcpEnabled,
		"token":   mcpToken,
	})
}

// UpdateMCPConfig updates MCP configuration (placeholder).
func UpdateMCPConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "MCP config update via environment variable BHF_MCP_TOKEN",
	})
}

// GetAppriseConfig returns Apprise configuration.
func GetAppriseConfig(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": cfg.Apprise.Enabled,
		"urls":    cfg.Apprise.URLs,
	})
}

// UpdateAppriseConfig updates Apprise configuration.
func UpdateAppriseConfig(c *gin.Context) {
	var req struct {
		Enabled bool     `json:"enabled"`
		URLs    []string `json:"urls"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	configPath := config.GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config"})
		return
	}

	content := string(data)
	enabledStr := "false"
	if req.Enabled {
		enabledStr = "true"
	}
	content = updateYAMLField(content, "enabled", enabledStr)

	os.WriteFile(configPath, []byte(content), 0644)
	config.ReloadConfig()

	c.JSON(http.StatusOK, gin.H{"message": "Apprise config updated"})
}

// TestApprise sends a test Apprise notification.
func TestApprise(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Test notification sent (placeholder)"})
}

// TestNtfy tests ntfy push notification.
func TestNtfy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ntfy test sent (placeholder)"})
}

// Logout logs out the user (clears cookies from config).
func Logout(c *gin.Context) {
	cookies := map[string]string{
		"SESSDATA":        "",
		"bili_jct":        "",
		"DedeUserID":      "",
		"DedeUserID__ckMd5": "",
	}
	config.SaveCookies(cookies)
	config.ReloadConfig()

	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// CheckLogin checks if the user is logged in.
func CheckLogin(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"logged_in": false})
		return
	}

	loggedIn := cfg.SESSDATA != "" && cfg.SESSDATA != "Cookie里的SESSDATA字段值"
	c.JSON(http.StatusOK, gin.H{
		"logged_in": loggedIn,
	})
}

func updateYAMLField(content, field, value string) string {
	lines := splitLines(content)
	for i, line := range lines {
		trimmed := trimSpace(line)
		if len(trimmed) > len(field)+1 && trimmed[:len(field)+1] == field+": " || trimmed == field+":" {
			lines[i] = field + ": " + value
			return joinLines(lines)
		}
	}
	// Field not found, append
	lines = append(lines, field+": "+value)
	return joinLines(lines)
}

func updateYAMLInt(content, field string, value int) string {
	return updateYAMLField(content, field, string(rune('0'+value)))
}

func splitLines(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
