package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

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

type YuttoBasicConfig struct {
	Dir          string `yaml:"dir"`
	TmpDir       string `yaml:"tmp_dir"`
	VipStrict    bool   `yaml:"vip_strict"`
	LoginStrict  bool   `yaml:"login_strict"`
}

type YuttoResourceConfig struct {
	RequireSubtitle bool `yaml:"require_subtitle"`
	OnlyAudio       bool `yaml:"only_audio"`
}

type YuttoDanmakuConfig struct {
	FontSize           int      `yaml:"font_size"`
	BlockKeywordPatterns []string `yaml:"block_keyword_patterns"`
}

type YuttoBatchConfig struct {
	WithSection bool `yaml:"with_section"`
}

type YuttoConfig struct {
	Basic    YuttoBasicConfig    `yaml:"basic"`
	Resource YuttoResourceConfig `yaml:"resource"`
	Danmaku  YuttoDanmakuConfig  `yaml:"danmaku"`
	Batch    YuttoBatchConfig    `yaml:"batch"`
}

type ServerConfig struct {
	Host           string             `yaml:"host"`
	Port           int                `yaml:"port"`
	SSLEnabled     bool               `yaml:"ssl_enabled"`
	SSLCertFile    string             `yaml:"ssl_certfile"`
	SSLKeyFile     string             `yaml:"ssl_keyfile"`
	DataIntegrity  DataIntegrityConfig `yaml:"data_integrity"`
}

type DataIntegrityConfig struct {
	CheckOnStartup bool `yaml:"check_on_startup"`
}

type HeatmapChartConfig struct {
	Width  string `yaml:"width"`
	Height string `yaml:"height"`
}

type HeatmapPiecesConfig struct {
	Min   int    `yaml:"min"`
	Max   int    `yaml:"max"`
	Color string `yaml:"color"`
}

type HeatmapColorsConfig struct {
	Pieces []HeatmapPiecesConfig `yaml:"pieces"`
}

type HeatmapConfig struct {
	OutputDir    string             `yaml:"output_dir"`
	TemplateFile string             `yaml:"template_file"`
	Title        string             `yaml:"title"`
	Chart        HeatmapChartConfig `yaml:"chart"`
	Colors       HeatmapColorsConfig `yaml:"colors"`
}

type SchedulerConfig struct {
	TaskTimeout int `yaml:"task_timeout"`
	RetryDelay  int `yaml:"retry_delay"`
	MaxRetries  int `yaml:"max_retries"`
}

type Config struct {
	SESSDATA         string         `yaml:"SESSDATA"`
	InputFolder      string         `yaml:"input_folder"`
	OutputFolder     string         `yaml:"output_folder"`
	DBFile           string         `yaml:"db_file"`
	LogFile          string         `yaml:"log_file"`
	CategoriesFile   string         `yaml:"categories_file"`
	DailyCountFolder string         `yaml:"daily_count_folder"`
	HeatmapTemplate  string         `yaml:"heatmap_template"`
	FieldsToRemove   []string       `yaml:"fields_to_remove"`
	Email            EmailConfig    `yaml:"email"`
	Apprise          AppriseConfig  `yaml:"apprise"`
	LogFolder        string         `yaml:"log_folder"`
	Yutto            YuttoConfig    `yaml:"yutto"`
	Server           ServerConfig   `yaml:"server"`
	Heatmap          HeatmapConfig  `yaml:"heatmap"`
	Scheduler        SchedulerConfig `yaml:"scheduler"`
	BiliJct          string         `yaml:"bili_jct"`
	DedeUserID       string         `yaml:"DedeUserID"`
	DedeUserIDCkMd5  string         `yaml:"DedeUserID__ckMd5"`
}

var (
	config     *Config
	configOnce sync.Once
	configPath string
)

func getBasePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func GetConfigPath(configFile string) string {
	basePath := getBasePath()
	
	if _, err := os.Stat(filepath.Join(basePath, "config", configFile)); err == nil {
		return filepath.Join(basePath, "config", configFile)
	}
	
	workDir, _ := os.Getwd()
	return filepath.Join(workDir, "config", configFile)
}

func LoadConfig() (*Config, error) {
	var loadErr error
	configOnce.Do(func() {
		cfgPath := GetConfigPath("config.yaml")
		configPath = cfgPath
		
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			loadErr = fmt.Errorf("配置文件不存在: %s", cfgPath)
			return
		}
		
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			loadErr = fmt.Errorf("读取配置文件失败: %v", err)
			return
		}
		
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			loadErr = fmt.Errorf("解析配置文件失败: %v", err)
			return
		}
		
		if cfg.Server.Host == "" {
			cfg.Server.Host = "0.0.0.0"
		}
		if cfg.Server.Port == 0 {
			cfg.Server.Port = 8899
		}
		
		config = &cfg
	})
	
	if loadErr != nil {
		return nil, loadErr
	}
	
	return config, nil
}

func GetConfig() *Config {
	cfg, _ := LoadConfig()
	return cfg
}

func GetConfigPathValue() string {
	return configPath
}

func SaveConfig(cfg *Config) error {
	cfgPath := GetConfigPath("config.yaml")
	
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	
	config = cfg
	return nil
}

func ReloadConfig() (*Config, error) {
	configOnce = sync.Once{}
	config = nil
	return LoadConfig()
}
