package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	rawConfig  []byte
)

func getBasePath() string {
	exe, err := os.Executable()
	if err != nil {
		workDir, _ := os.Getwd()
		return workDir
	}
	return filepath.Dir(exe)
}

func GetConfigPath(configFile string) string {
	basePath := getBasePath()

	internalPath := filepath.Join(basePath, "_internal", "config", configFile)
	if _, err := os.Stat(internalPath); err == nil {
		return internalPath
	}

	configDirPath := filepath.Join(basePath, "config", configFile)
	if _, err := os.Stat(configDirPath); err == nil {
		return configDirPath
	}

	workDir, _ := os.Getwd()
	workConfigPath := filepath.Join(workDir, "config", configFile)
	if _, err := os.Stat(workConfigPath); err == nil {
		return workConfigPath
	}

	parentWorkDir := filepath.Join(workDir, "..", "config", configFile)
	if _, err := os.Stat(parentWorkDir); err == nil {
		absPath, _ := filepath.Abs(parentWorkDir)
		return absPath
	}

	return filepath.Join(basePath, "config", configFile)
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

		rawConfig = make([]byte, len(data))
		copy(rawConfig, data)

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

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
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

	originalData, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("读取原配置文件失败: %v", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(originalData, &root); err != nil {
		return fmt.Errorf("解析原配置文件失败: %v", err)
	}

	updateYamlNode(&root, cfg)

	var output strings.Builder
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		return fmt.Errorf("编码配置失败: %v", err)
	}
	encoder.Close()

	if err := os.WriteFile(cfgPath, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	newData, _ := os.ReadFile(cfgPath)
	rawConfig = make([]byte, len(newData))
	copy(rawConfig, newData)

	config = cfg
	return nil
}

func updateYamlNode(root *yaml.Node, cfg *Config) {
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	if root.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(root.Content); i += 2 {
		keyNode := root.Content[i]
		valueNode := root.Content[i+1]

		switch keyNode.Value {
		case "SESSDATA":
			valueNode.Value = cfg.SESSDATA
		case "input_folder":
			valueNode.Value = cfg.InputFolder
		case "output_folder":
			valueNode.Value = cfg.OutputFolder
		case "db_file":
			valueNode.Value = cfg.DBFile
		case "log_file":
			valueNode.Value = cfg.LogFile
		case "categories_file":
			valueNode.Value = cfg.CategoriesFile
		case "daily_count_folder":
			valueNode.Value = cfg.DailyCountFolder
		case "heatmap_template":
			valueNode.Value = cfg.HeatmapTemplate
		case "log_folder":
			valueNode.Value = cfg.LogFolder
		case "bili_jct":
			valueNode.Value = cfg.BiliJct
		case "DedeUserID":
			valueNode.Value = cfg.DedeUserID
		case "DedeUserID__ckMd5":
			valueNode.Value = cfg.DedeUserIDCkMd5
		case "email":
			updateEmailNode(valueNode, &cfg.Email)
		case "apprise":
			updateAppriseNode(valueNode, &cfg.Apprise)
		case "server":
			updateServerNode(valueNode, &cfg.Server)
		}
	}
}

func updateEmailNode(node *yaml.Node, email *EmailConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "smtp_server":
			val.Value = email.SMTPServer
		case "smtp_port":
			val.Value = fmt.Sprintf("%d", email.SMTPPort)
		case "sender":
			val.Value = email.Sender
		case "password":
			val.Value = email.Password
		case "receiver":
			val.Value = email.Receiver
		}
	}
}

func updateAppriseNode(node *yaml.Node, apprise *AppriseConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "enabled":
			val.Value = fmt.Sprintf("%t", apprise.Enabled)
		}
	}
}

func updateServerNode(node *yaml.Node, server *ServerConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "host":
			val.Value = server.Host
		case "port":
			val.Value = fmt.Sprintf("%d", server.Port)
		case "ssl_enabled":
			val.Value = fmt.Sprintf("%t", server.SSLEnabled)
		case "ssl_certfile":
			val.Value = server.SSLCertFile
		case "ssl_keyfile":
			val.Value = server.SSLKeyFile
		}
	}
}

func ReloadConfig() (*Config, error) {
	configOnce = sync.Once{}
	config = nil
	rawConfig = nil
	return LoadConfig()
}
