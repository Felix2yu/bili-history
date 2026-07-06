package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type ShoutrrrConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	URLs    []string `yaml:"urls" json:"urls"`
}

type ServerConfig struct {
	Host           string              `yaml:"host" json:"host"`
	Port           int                 `yaml:"port" json:"port"`
	SSLEnabled     bool                `yaml:"ssl_enabled" json:"ssl_enabled"`
	SSLCertFile    string              `yaml:"ssl_certfile" json:"ssl_certfile"`
	SSLKeyFile     string              `yaml:"ssl_keyfile" json:"ssl_keyfile"`
	DataIntegrity  DataIntegrityConfig `yaml:"data_integrity" json:"data_integrity"`
}

type DataIntegrityConfig struct {
	CheckOnStartup bool `yaml:"check_on_startup" json:"check_on_startup"`
}

type SyncConfig struct {
	SyncDeleted         bool `yaml:"sync_deleted" json:"sync_deleted"`
	SyncDeleteToBilibili bool `yaml:"sync_delete_to_bilibili" json:"sync_delete_to_bilibili"`
}

type AppearanceConfig struct {
	DarkMode string `yaml:"dark_mode" json:"dark_mode"`
}

type SchedulerConfig struct {
	TaskTimeout int `yaml:"task_timeout" json:"task_timeout"`
	RetryDelay  int `yaml:"retry_delay" json:"retry_delay"`
	MaxRetries  int `yaml:"max_retries" json:"max_retries"`
}

type ColorPiece struct {
	Min   int    `yaml:"min" json:"min"`
	Max   int    `yaml:"max" json:"max"`
	Color string `yaml:"color" json:"color"`
}

type HeatmapConfig struct {
	OutputDir    string       `yaml:"output_dir" json:"output_dir"`
	TemplateFile string       `yaml:"template_file" json:"template_file"`
	Title        string       `yaml:"title" json:"title"`
	Colors       []ColorPiece `yaml:"colors" json:"colors"`
}

type McpConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Path        string `yaml:"path" json:"path"`
	AuthEnabled bool   `yaml:"auth_enabled" json:"auth_enabled"`
	Token       string `yaml:"token" json:"token"`
	MaxPageSize int    `yaml:"max_page_size" json:"max_page_size"`
}

type Config struct {
	SESSDATA         string          `yaml:"SESSDATA" json:"SESSDATA"`
	OutputFolder     string          `yaml:"output_folder" json:"output_folder"`
	DBFile           string          `yaml:"db_file" json:"db_file"`
	LogFile          string          `yaml:"log_file" json:"log_file"`
	CategoriesFile   string          `yaml:"categories_file" json:"categories_file"`
	FieldsToRemove   []string        `yaml:"fields_to_remove" json:"fields_to_remove"`
	Shoutrrr         ShoutrrrConfig  `yaml:"shoutrrr" json:"shoutrrr"`
	LogFolder        string          `yaml:"log_folder" json:"log_folder"`
	Server           ServerConfig    `yaml:"server" json:"server"`
	Scheduler        SchedulerConfig `yaml:"scheduler" json:"scheduler"`
	Sync             SyncConfig      `yaml:"sync" json:"sync"`
	Appearance       AppearanceConfig `yaml:"appearance" json:"appearance"`
	Heatmap          HeatmapConfig   `yaml:"heatmap" json:"heatmap"`
	Mcp              McpConfig       `yaml:"mcp" json:"mcp"`
	BiliJct          string          `yaml:"bili_jct" json:"bili_jct"`
	DedeUserID       string          `yaml:"DedeUserID" json:"DedeUserID"`
	DedeUserIDCkMd5  string          `yaml:"DedeUserID__ckMd5" json:"DedeUserID__ckMd5"`
}

var (
	config     *Config
	configOnce sync.Once
	configPath string
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

		var cfg Config
		fileExists := true

		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			fileExists = false
		} else {
			data, err := os.ReadFile(cfgPath)
			if err != nil {
				loadErr = fmt.Errorf("读取配置文件失败: %v", err)
				return
			}

			if err := yaml.Unmarshal(data, &cfg); err != nil {
				loadErr = fmt.Errorf("解析配置文件失败: %v", err)
				return
			}
		}

		if cfg.Server.Host == "" {
			cfg.Server.Host = "0.0.0.0"
		}
		if cfg.Server.Port == 0 {
			cfg.Server.Port = 8899
		}

		applyEnvOverrides(&cfg)

		if !fileExists {
			loadErr = fmt.Errorf("配置文件不存在: %s，使用默认配置", cfgPath)
		}

		config = &cfg
	})

	if config == nil {
		return nil, loadErr
	}

	return config, loadErr
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

	existingKeys := make(map[string]int)
	for i := 0; i < len(root.Content); i += 2 {
		existingKeys[root.Content[i].Value] = i
	}

	getOrAddNode := func(key string) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			return root.Content[idx+1]
		}
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
			Tag:   "!!str",
		}
		valueNode := &yaml.Node{
			Kind: yaml.MappingNode,
		}
		root.Content = append(root.Content, keyNode, valueNode)
		existingKeys[key] = len(root.Content) - 2
		return valueNode
	}

	for i := 0; i < len(root.Content); i += 2 {
		keyNode := root.Content[i]
		valueNode := root.Content[i+1]

		switch keyNode.Value {
		case "SESSDATA":
			valueNode.Value = cfg.SESSDATA
		case "output_folder":
			valueNode.Value = cfg.OutputFolder
		case "db_file":
			valueNode.Value = cfg.DBFile
		case "log_file":
			valueNode.Value = cfg.LogFile
		case "categories_file":
			valueNode.Value = cfg.CategoriesFile
		case "log_folder":
			valueNode.Value = cfg.LogFolder
		case "bili_jct":
			valueNode.Value = cfg.BiliJct
		case "DedeUserID":
			valueNode.Value = cfg.DedeUserID
		case "DedeUserID__ckMd5":
			valueNode.Value = cfg.DedeUserIDCkMd5
		case "shoutrrr":
			updateShoutrrrNode(valueNode, &cfg.Shoutrrr)
		case "server":
			updateServerNode(valueNode, &cfg.Server)
		case "sync":
			updateSyncNode(valueNode, &cfg.Sync)
		case "appearance":
			updateAppearanceNode(valueNode, &cfg.Appearance)
		case "heatmap":
			updateHeatmapNode(valueNode, &cfg.Heatmap)
		case "mcp":
			updateMcpNode(valueNode, &cfg.Mcp)
		}
	}

	if _, ok := existingKeys["shoutrrr"]; !ok {
		node := getOrAddNode("shoutrrr")
		updateShoutrrrNode(node, &cfg.Shoutrrr)
	}
	if _, ok := existingKeys["server"]; !ok {
		node := getOrAddNode("server")
		updateServerNode(node, &cfg.Server)
	}
	if _, ok := existingKeys["sync"]; !ok {
		node := getOrAddNode("sync")
		updateSyncNode(node, &cfg.Sync)
	}
	if _, ok := existingKeys["appearance"]; !ok {
		node := getOrAddNode("appearance")
		updateAppearanceNode(node, &cfg.Appearance)
	}
	if _, ok := existingKeys["heatmap"]; !ok {
		node := getOrAddNode("heatmap")
		updateHeatmapNode(node, &cfg.Heatmap)
	}
	if _, ok := existingKeys["mcp"]; !ok {
		node := getOrAddNode("mcp")
		updateMcpNode(node, &cfg.Mcp)
	}
}

func updateShoutrrrNode(node *yaml.Node, shoutrrr *ShoutrrrConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	existingKeys := make(map[string]int)
	for i := 0; i < len(node.Content); i += 2 {
		existingKeys[node.Content[i].Value] = i
	}

	getOrAddScalar := func(key string) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!bool"
				n.Content = nil
			}
			return n
		}
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
			Tag:   "!!str",
		}
		valNode := &yaml.Node{
			Kind: yaml.ScalarNode,
			Tag:  "!!bool",
		}
		node.Content = append(node.Content, keyNode, valNode)
		existingKeys[key] = len(node.Content) - 2
		return valNode
	}

	getOrAddSeq := func(key string) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.SequenceNode {
				n.Kind = yaml.SequenceNode
				n.Tag = ""
				n.Content = nil
			}
			return n
		}
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
			Tag:   "!!str",
		}
		valNode := &yaml.Node{
			Kind: yaml.SequenceNode,
		}
		node.Content = append(node.Content, keyNode, valNode)
		existingKeys[key] = len(node.Content) - 2
		return valNode
	}

	enabledNode := getOrAddScalar("enabled")
	enabledNode.Value = fmt.Sprintf("%t", shoutrrr.Enabled)

	urlsNode := getOrAddSeq("urls")
	urlsNode.Content = nil
	for _, url := range shoutrrr.URLs {
		urlsNode.Content = append(urlsNode.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: url,
			Tag:   "!!str",
		})
	}
}

func updateServerNode(node *yaml.Node, server *ServerConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	existingKeys := make(map[string]int)
	for i := 0; i < len(node.Content); i += 2 {
		existingKeys[node.Content[i].Value] = i
	}

	getOrAddScalar := func(key string) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!str"
				n.Content = nil
			}
			return n
		}
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
			Tag:   "!!str",
		}
		valNode := &yaml.Node{
			Kind: yaml.ScalarNode,
			Tag:  "!!str",
		}
		node.Content = append(node.Content, keyNode, valNode)
		existingKeys[key] = len(node.Content) - 2
		return valNode
	}

	getOrAddScalar("host").Value = server.Host
	getOrAddScalar("port").Value = fmt.Sprintf("%d", server.Port)
	getOrAddScalar("ssl_enabled").Value = fmt.Sprintf("%t", server.SSLEnabled)
	getOrAddScalar("ssl_certfile").Value = server.SSLCertFile
	getOrAddScalar("ssl_keyfile").Value = server.SSLKeyFile

	// data_integrity sub-node
	diKey := "data_integrity"
	var diNode *yaml.Node
	if idx, ok := existingKeys[diKey]; ok {
		diNode = node.Content[idx+1]
	} else {
		kn := &yaml.Node{Kind: yaml.ScalarNode, Value: diKey, Tag: "!!str"}
		diNode = &yaml.Node{Kind: yaml.MappingNode}
		node.Content = append(node.Content, kn, diNode)
	}
	if diNode.Kind != yaml.MappingNode {
		diNode.Kind = yaml.MappingNode
		diNode.Content = nil
	}
	diExisting := make(map[string]int)
	for i := 0; i < len(diNode.Content); i += 2 {
		diExisting[diNode.Content[i].Value] = i
	}
	if idx, ok := diExisting["check_on_startup"]; ok {
		diNode.Content[idx+1].Value = fmt.Sprintf("%t", server.DataIntegrity.CheckOnStartup)
	} else {
		kn := &yaml.Node{Kind: yaml.ScalarNode, Value: "check_on_startup", Tag: "!!str"}
		vn := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprintf("%t", server.DataIntegrity.CheckOnStartup)}
		diNode.Content = append(diNode.Content, kn, vn)
	}
}

func updateSyncNode(node *yaml.Node, sync *SyncConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	existingKeys := make(map[string]int)
	for i := 0; i < len(node.Content); i += 2 {
		existingKeys[node.Content[i].Value] = i
	}

	getOrAddBool := func(key string, val bool) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!bool"
				n.Content = nil
			}
			n.Value = fmt.Sprintf("%t", val)
			return n
		}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprintf("%t", val)}
		node.Content = append(node.Content, keyNode, valNode)
		return valNode
	}

	getOrAddBool("sync_deleted", sync.SyncDeleted)
	getOrAddBool("sync_delete_to_bilibili", sync.SyncDeleteToBilibili)
}

func updateAppearanceNode(node *yaml.Node, appearance *AppearanceConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	existingKeys := make(map[string]int)
	for i := 0; i < len(node.Content); i += 2 {
		existingKeys[node.Content[i].Value] = i
	}

	getOrAddString := func(key string, val string) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!str"
				n.Content = nil
			}
			n.Value = val
			return n
		}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val}
		node.Content = append(node.Content, keyNode, valNode)
		return valNode
	}

	getOrAddString("dark_mode", appearance.DarkMode)
}

func updateHeatmapNode(node *yaml.Node, heatmap *HeatmapConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	existingKeys := make(map[string]int)
	for i := 0; i < len(node.Content); i += 2 {
		existingKeys[node.Content[i].Value] = i
	}

	getOrAddString := func(key string, val string) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!str"
				n.Content = nil
			}
			n.Value = val
			return n
		}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val}
		node.Content = append(node.Content, keyNode, valNode)
		return valNode
	}

	getOrAddString("output_dir", heatmap.OutputDir)
	getOrAddString("template_file", heatmap.TemplateFile)
	getOrAddString("title", heatmap.Title)
}

func updateMcpNode(node *yaml.Node, mcp *McpConfig) {
	if node.Kind != yaml.MappingNode {
		return
	}
	existingKeys := make(map[string]int)
	for i := 0; i < len(node.Content); i += 2 {
		existingKeys[node.Content[i].Value] = i
	}

	getOrAddBool := func(key string, val bool) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!bool"
				n.Content = nil
			}
			n.Value = fmt.Sprintf("%t", val)
			return n
		}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprintf("%t", val)}
		node.Content = append(node.Content, keyNode, valNode)
		return valNode
	}

	getOrAddString := func(key string, val string) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!str"
				n.Content = nil
			}
			n.Value = val
			return n
		}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val}
		node.Content = append(node.Content, keyNode, valNode)
		return valNode
	}

	getOrAddInt := func(key string, val int) *yaml.Node {
		if idx, ok := existingKeys[key]; ok {
			n := node.Content[idx+1]
			if n.Kind != yaml.ScalarNode {
				n.Kind = yaml.ScalarNode
				n.Tag = "!!int"
				n.Content = nil
			}
			n.Value = fmt.Sprintf("%d", val)
			return n
		}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", val)}
		node.Content = append(node.Content, keyNode, valNode)
		return valNode
	}

	getOrAddBool("enabled", mcp.Enabled)
	getOrAddString("path", mcp.Path)
	getOrAddBool("auth_enabled", mcp.AuthEnabled)
	getOrAddString("token", mcp.Token)
	getOrAddInt("max_page_size", mcp.MaxPageSize)
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SESSDATA"); v != "" {
		cfg.SESSDATA = v
	}
	if v := os.Getenv("BILI_JCT"); v != "" {
		cfg.BiliJct = v
	}
	if v := os.Getenv("DEDE_USER_ID"); v != "" {
		cfg.DedeUserID = v
	}
	if v := os.Getenv("DEDE_USER_ID_CKMD5"); v != "" {
		cfg.DedeUserIDCkMd5 = v
	}
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 {
			cfg.Server.Port = port
		}
	}
}
