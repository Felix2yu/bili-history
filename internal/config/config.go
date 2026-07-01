package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	SESSDATA        string            `yaml:"SESSDATA"`
	BiliJCT         string            `yaml:"bili_jct"`
	DedeUserID      string            `yaml:"DedeUserID"`
	DedeUserIDCkMd5 string            `yaml:"DedeUserID__ckMd5"`
	InputFolder     string            `yaml:"input_folder"`
	OutputFolder    string            `yaml:"output_folder"`
	DBFile          string            `yaml:"db_file"`
	LogFile         string            `yaml:"log_file"`
	CategoriesFile  string            `yaml:"categories_file"`
	DailyCountFolder string           `yaml:"daily_count_folder"`
	HeatmapTemplate string            `yaml:"heatmap_template"`
	FieldsToRemove  []string          `yaml:"fields_to_remove"`
	Email           EmailConfig       `yaml:"email"`
	Apprise         AppriseConfig     `yaml:"apprise"`
	LogFolder       string            `yaml:"log_folder"`
	Yutto           YuttoConfig       `yaml:"yutto"`
	Server          ServerConfig      `yaml:"server"`
	Heatmap         HeatmapConfig     `yaml:"heatmap"`
	Scheduler       SchedulerConfig   `yaml:"scheduler"`
}

type EmailConfig struct {
	SMTPServer string `yaml:"smtp_server"`
	SMTPPort   int    `yaml:"smtp_port"`
	Sender     string `yaml:"sender"`
	Password   string `yaml:"password"`
	Receiver   string `yaml:"receiver"`
}

type AppriseConfig struct {
	Enabled bool     `yaml:"enabled"`
	URLs    []string `yaml:"urls"`
}

type YuttoConfig struct {
	Basic    YuttoBasic    `yaml:"basic"`
	Resource YuttoResource `yaml:"resource"`
	Danmaku  YuttoDanmaku  `yaml:"danmaku"`
	Batch    YuttoBatch    `yaml:"batch"`
}

type YuttoBasic struct {
	Dir        string `yaml:"dir"`
	TmpDir     string `yaml:"tmp_dir"`
	VipStrict  bool   `yaml:"vip_strict"`
	LoginStrict bool  `yaml:"login_strict"`
}

type YuttoResource struct {
	RequireSubtitle bool `yaml:"require_subtitle"`
	OnlyAudio       bool `yaml:"only_audio"`
}

type YuttoDanmaku struct {
	FontSize             int      `yaml:"font_size"`
	BlockKeywordPatterns []string `yaml:"block_keyword_patterns"`
}

type YuttoBatch struct {
	WithSection bool `yaml:"with_section"`
}

type ServerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	SSLEnabled     bool   `yaml:"ssl_enabled"`
	SSLKeyFile     string `yaml:"ssl_keyfile"`
	SSLCertFile    string `yaml:"ssl_certfile"`
	DataIntegrity  DataIntegrityConfig `yaml:"data_integrity"`
}

type DataIntegrityConfig struct {
	CheckOnStartup bool `yaml:"check_on_startup"`
}

type HeatmapConfig struct {
	OutputDir    string        `yaml:"output_dir"`
	TemplateFile string        `yaml:"template_file"`
	Title        string        `yaml:"title"`
	Chart        HeatmapChart  `yaml:"chart"`
	Colors       HeatmapColors `yaml:"colors"`
}

type HeatmapChart struct {
	Width  string `yaml:"width"`
	Height string `yaml:"height"`
}

type HeatmapColors struct {
	Pieces []HeatmapPiece `yaml:"pieces"`
}

type HeatmapPiece struct {
	Min   int    `yaml:"min"`
	Max   int    `yaml:"max"`
	Color string `yaml:"color"`
}

type SchedulerConfig struct {
	TaskTimeout int `yaml:"task_timeout"`
	RetryDelay  int `yaml:"retry_delay"`
	MaxRetries  int `yaml:"max_retries"`
}

var (
	globalConfig *Config
	configOnce   sync.Once
	configPath   string
	mu           sync.RWMutex
)

// SetConfigPath sets the path for the config file.
func SetConfigPath(path string) {
	mu.Lock()
	defer mu.Unlock()
	configPath = path
	globalConfig = nil
	configOnce = sync.Once{}
}

// GetBasePath returns the base path of the application.
func GetBasePath() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

// GetConfigPath returns the path to the config file.
// NOTE: Does NOT acquire mu to avoid deadlock when called from LoadConfig.
func GetConfigPath() string {
	mu.RLock()
	p := configPath
	mu.RUnlock()
	if p != "" {
		return p
	}
	return filepath.Join(GetBasePath(), "config", "config.yaml")
}

// getConfigPathInternal returns config path without locking (for internal use).
func getConfigPathInternal() string {
	if configPath != "" {
		return configPath
	}
	return filepath.Join(GetBasePath(), "config", "config.yaml")
}

// GetOutputPath returns a path under the output directory.
func GetOutputPath(parts ...string) string {
	base := GetBasePath()
	args := append([]string{base, "output"}, parts...)
	return filepath.Join(args...)
}

// GetDatabasePath returns a path under the output/database directory.
func GetDatabasePath(parts ...string) string {
	base := GetBasePath()
	args := append([]string{base, "output", "database"}, parts...)
	return filepath.Join(args...)
}

// LoadConfig loads the configuration from YAML file.
func LoadConfig() (*Config, error) {
	// Fast path: already loaded
	mu.RLock()
	if globalConfig != nil {
		mu.RUnlock()
		return globalConfig, nil
	}
	mu.RUnlock()

	// Slow path: acquire write lock
	mu.Lock()
	defer mu.Unlock()

	// Double-check after acquiring write lock
	if globalConfig != nil {
		return globalConfig, nil
	}

	// Compute path WITHOUT calling GetConfigPath (which would deadlock)
	path := getConfigPathInternal()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8899
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.DBFile == "" {
		cfg.DBFile = "bilibili_history.db"
	}
	if cfg.Scheduler.TaskTimeout == 0 {
		cfg.Scheduler.TaskTimeout = 600
	}
	if cfg.Scheduler.RetryDelay == 0 {
		cfg.Scheduler.RetryDelay = 300
	}
	if cfg.Scheduler.MaxRetries == 0 {
		cfg.Scheduler.MaxRetries = 3
	}

	globalConfig = &cfg
	return globalConfig, nil
}

// SaveCookies updates the cookie fields in the config file.
func SaveCookies(cookies map[string]string) error {
	mu.Lock()
	defer mu.Unlock()

	path := getConfigPathInternal()
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	content := string(data)
	cookieFields := []string{"SESSDATA", "bili_jct", "DedeUserID", "DedeUserID__ckMd5"}

	for _, field := range cookieFields {
		val, ok := cookies[field]
		if !ok {
			continue
		}
		lines := splitLines(content)
		found := false
		for i, line := range lines {
			trimmed := trimSpace(line)
			if strings.HasPrefix(trimmed, field+":") {
				lines[i] = field + ": " + val
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, field+": "+val)
		}
		content = joinLines(lines)
	}

	return os.WriteFile(path, []byte(content), 0644)
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

// ReloadConfig forces a reload of the configuration.
func ReloadConfig() (*Config, error) {
	mu.Lock()
	globalConfig = nil
	mu.Unlock()
	return LoadConfig()
}
