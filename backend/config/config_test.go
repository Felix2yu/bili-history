package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// rawConfigData 保存仓库原始 config.yaml 内容，供测试清理/恢复使用。
var rawConfigData []byte

// TestMain 将工作目录切换到临时目录，并把仓库中的 config.yaml 复制过去，
// 既保证 LoadConfig 能成功加载（覆盖正常分支），又避免测试写入污染仓库。
// 注意：go test 运行时工作目录为包目录（backend/config），故相对路径直接使用 "config.yaml"。
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "bili-config-test")
	if err != nil {
		panic(err)
	}
	cfgDir := filepath.Join(tmp, "config")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		panic(err)
	}
	if data, err := os.ReadFile("config.yaml"); err == nil {
		rawConfigData = data
		_ = os.WriteFile(filepath.Join(cfgDir, "config.yaml"), data, 0644)
	}

	old, _ := os.Getwd()
	_ = os.Chdir(tmp)

	// 预热 LoadConfig（仅执行一次），确保后续测试拿到已加载的配置
	LoadConfig()

	code := m.Run()

	_ = os.Chdir(old)
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

func fullConfig() *Config {
	return &Config{
		SESSDATA:         "sess",
		OutputFolder:     "of",
		DBFile:           "db.sqlite",
		LogFile:          "lf.log",
		CategoriesFile:   "cf.json",
		LogFolder:        "logs",
		BiliJct:          "jct",
		DedeUserID:       "uid",
		DedeUserIDCkMd5:  "md5",
		Shoutrrr:         ShoutrrrConfig{Enabled: true, URLs: []string{"url1", "url2"}},
		Server:           ServerConfig{Host: "127.0.0.1", Port: 9000, SSLEnabled: true, SSLCertFile: "c.pem", SSLKeyFile: "k.pem", DataIntegrity: DataIntegrityConfig{CheckOnStartup: true}},
		Sync:             SyncConfig{SyncDeleted: true, SyncDeleteToBilibili: true},
		Appearance:       AppearanceConfig{DarkMode: "dark"},
		Heatmap:          HeatmapConfig{OutputDir: "o", TemplateFile: "t", Title: "ti"},
		Mcp:              McpConfig{Enabled: true, Path: "/m", AuthEnabled: true, Token: "tok", MaxPageSize: 50},
	}
}

func TestGetBasePath(t *testing.T) {
	if p := getBasePath(); p == "" {
		t.Errorf("getBasePath 返回空字符串")
	}
}

func TestGetConfigPath(t *testing.T) {
	// 存在的配置文件：走 cwd/config 分支
	if p := GetConfigPath("config.yaml"); !strings.HasSuffix(p, "config/config.yaml") {
		t.Errorf("GetConfigPath 结果异常: %s", p)
	}
	// 不存在的配置文件：走默认返回分支
	if p := GetConfigPath("nonexistent.yaml"); !strings.HasSuffix(p, "config/nonexistent.yaml") {
		t.Errorf("GetConfigPath 默认分支异常: %s", p)
	}
}

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig()
	if cfg == nil {
		t.Fatal("LoadConfig 返回 nil 配置")
	}
	if err != nil {
		t.Errorf("LoadConfig 返回非预期错误: %v", err)
	}
	if cfg.Server.Port != 8899 {
		t.Errorf("期望 Server.Port=8899, 实际 %d", cfg.Server.Port)
	}
	if !cfg.Mcp.Enabled {
		t.Errorf("期望 Mcp.Enabled=true, 实际 false")
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("期望 Server.Host=0.0.0.0, 实际 %q", cfg.Server.Host)
	}
}

func TestGetConfig(t *testing.T) {
	if cfg := GetConfig(); cfg == nil {
		t.Error("GetConfig 返回 nil")
	}
}

func TestGetConfigPathValue(t *testing.T) {
	if v := GetConfigPathValue(); v == "" {
		t.Error("GetConfigPathValue 返回空")
	}
}

func TestSaveConfigUpdateExisting(t *testing.T) {
	cfg := fullConfig()
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig 失败: %v", err)
	}
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		t.Fatalf("读取保存后的配置失败: %v", err)
	}
	content := string(data)
	for _, want := range []string{"9000", "tok", "url1", "dark", "ti", "jct"} {
		if !strings.Contains(content, want) {
			t.Errorf("保存的配置缺少 %q\n配置内容:\n%s", want, content)
		}
	}
}

func TestSaveConfigCreateNew(t *testing.T) {
	// 删除已有文件，触发「文件不存在则创建新文件」分支
	_ = os.Remove("config/config.yaml")
	// GetConfigPath 在未找到文件时返回默认路径，确保其父目录存在以便写入
	_ = os.MkdirAll(filepath.Dir(GetConfigPath("config.yaml")), 0755)
	cfg := &Config{Server: ServerConfig{Host: "h", Port: 1234}}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig 新建失败: %v", err)
	}
	if _, err := os.Stat(GetConfigPath("config.yaml")); err != nil {
		t.Errorf("SaveConfig 未创建配置文件: %v", err)
	}
	// 恢复文件，避免影响其它测试
	_ = os.WriteFile("config/config.yaml", rawConfigData, 0644)
}

func TestApplyEnvOverrides(t *testing.T) {
	envs := map[string]string{
		"SESSDATA":            "e_sess",
		"BILI_JCT":            "e_jct",
		"DEDE_USER_ID":        "e_uid",
		"DEDE_USER_ID_CKMD5":  "e_md5",
		"SERVER_HOST":         "1.2.3.4",
		"SERVER_PORT":         "7777",
	}
	for k, v := range envs {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}
	cfg := &Config{}
	applyEnvOverrides(cfg)
	if cfg.SESSDATA != "e_sess" {
		t.Errorf("SESSDATA 未被环境变量覆盖: %q", cfg.SESSDATA)
	}
	if cfg.BiliJct != "e_jct" {
		t.Errorf("BiliJct 未被环境变量覆盖: %q", cfg.BiliJct)
	}
	if cfg.DedeUserID != "e_uid" {
		t.Errorf("DedeUserID 未被环境变量覆盖: %q", cfg.DedeUserID)
	}
	if cfg.DedeUserIDCkMd5 != "e_md5" {
		t.Errorf("DedeUserIDCkMd5 未被环境变量覆盖: %q", cfg.DedeUserIDCkMd5)
	}
	if cfg.Server.Host != "1.2.3.4" {
		t.Errorf("Server.Host 未被环境变量覆盖: %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("Server.Port 未被环境变量覆盖: %d", cfg.Server.Port)
	}
}

func TestApplyEnvOverridesInvalidPort(t *testing.T) {
	os.Setenv("SERVER_PORT", "not-a-number")
	defer os.Unsetenv("SERVER_PORT")
	cfg := &Config{Server: ServerConfig{Port: 8899}}
	applyEnvOverrides(cfg)
	if cfg.Server.Port != 8899 {
		t.Errorf("非法 SERVER_PORT 不应修改端口, 实际 %d", cfg.Server.Port)
	}
}

func TestApplyEnvOverridesEmpty(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Host: "x", Port: 1}}
	applyEnvOverrides(cfg)
	if cfg.Server.Host != "x" || cfg.Server.Port != 1 {
		t.Errorf("无环境变量时配置不应被修改: %+v", cfg.Server)
	}
}

func TestUpdateYamlNodeExisting(t *testing.T) {
	doc := &yaml.Node{}
	src := "SESSDATA: a\noutput_folder: b\ndb_file: old.db\nlog_file: old.log\n" +
		"categories_file: old.json\nlog_folder: oldlogs\nbili_jct: oldjct\n" +
		"DedeUserID: olduid\nDedeUserID__ckMd5: oldmd5\n"
	if err := yaml.Unmarshal([]byte(src), doc); err != nil {
		t.Fatal(err)
	}
	cfg := fullConfig()
	updateYamlNode(doc, cfg)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		t.Fatal(err)
	}
	enc.Close()

	out := buf.String()
	// 已存在的顶层标量字段应被更新为 cfg 中的值
	for _, want := range []string{
		cfg.SESSDATA, cfg.OutputFolder, cfg.DBFile, cfg.LogFile,
		cfg.CategoriesFile, cfg.LogFolder, cfg.BiliJct, cfg.DedeUserID, cfg.DedeUserIDCkMd5,
		"9000", "tok", "url1", "dark", "ti",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("更新后的配置缺少 %q\n%s", want, out)
		}
	}
}

func TestUpdateYamlNodeNonMapping(t *testing.T) {
	scalar := &yaml.Node{Kind: yaml.ScalarNode, Value: "x"}
	updateYamlNode(scalar, &Config{}) // 非 MappingNode 应提前返回
}

func TestUpdateNodesAddMissing(t *testing.T) {
	empty := &yaml.Node{Kind: yaml.MappingNode}
	cfg := fullConfig()
	updateShoutrrrNode(empty, &cfg.Shoutrrr)
	updateServerNode(empty, &cfg.Server)
	updateSyncNode(empty, &cfg.Sync)
	updateAppearanceNode(empty, &cfg.Appearance)
	updateHeatmapNode(empty, &cfg.Heatmap)
	updateMcpNode(empty, &cfg.Mcp)
	// 走 getOrAddNode 添加缺失节点的分支
	updateYamlNode(empty, cfg)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	_ = enc.Encode(empty)
	enc.Close()
	if !strings.Contains(buf.String(), "shoutrrr") || !strings.Contains(buf.String(), "mcp") {
		t.Errorf("缺失节点未被添加:\n%s", buf.String())
	}
}

func TestUpdateNodesNonMapping(t *testing.T) {
	s := &yaml.Node{Kind: yaml.ScalarNode, Value: "x"}
	cfg := &Config{}
	updateShoutrrrNode(s, &cfg.Shoutrrr)
	updateServerNode(s, &cfg.Server)
	updateSyncNode(s, &cfg.Sync)
	updateAppearanceNode(s, &cfg.Appearance)
	updateHeatmapNode(s, &cfg.Heatmap)
	updateMcpNode(s, &cfg.Mcp)
}

func TestUpdateServerNodeWithDataIntegrity(t *testing.T) {
	doc := &yaml.Node{}
	if err := yaml.Unmarshal([]byte("server:\n  host: h\n  data_integrity:\n    check_on_startup: true\n"), doc); err != nil {
		t.Fatal(err)
	}
	root := doc.Content[0]
	for i := 0; i < len(root.Content); i += 2 {
		if root.Content[i].Value == "server" {
			updateServerNode(root.Content[i+1], &ServerConfig{Host: "nh", Port: 5, DataIntegrity: DataIntegrityConfig{CheckOnStartup: false}})
		}
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	_ = enc.Encode(doc)
	enc.Close()
	if !strings.Contains(buf.String(), "nh") || !strings.Contains(buf.String(), "5") {
		t.Errorf("server 节点更新异常:\n%s", buf.String())
	}
}
