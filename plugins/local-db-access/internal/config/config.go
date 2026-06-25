package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"local-db-access/internal/common"

	"gopkg.in/yaml.v3"
)

const (
	pluginRootEnv      = "CLAUDE_PLUGIN_ROOT"
	internalConfigPath = "internal/config/config.yaml"
)

var Conf Config

type Config struct {
	DefaultDatabase string                     `yaml:"default_database,omitempty" json:"default_database,omitempty"`
	Databases       map[string]*DatabaseConfig `yaml:"databases" json:"databases"`
}

type QueryLimits struct {
	MaxRows     int `json:"max_rows" yaml:"max_rows"`
	MaxTimeSec  int `json:"max_time_sec" yaml:"max_time_sec"`
	MaxRowLimit int `json:"max_row_limit" yaml:"max_row_limit"`
}

type DatabaseConfig struct {
	Type            common.DatabaseType `yaml:"type" json:"type"`
	Enabled         bool                `yaml:"enabled" json:"enabled"`
	QueryLimits     *QueryLimits        `yaml:"query_limits,omitempty" json:"query_limits,omitempty"`
	MaxOpenConn     int                 `yaml:"max_open_conn,omitempty" json:"max_open_conn,omitempty"`
	MaxIdleConn     int                 `yaml:"max_idle_conn,omitempty" json:"max_idle_conn,omitempty"`
	ConnMaxLifetime time.Duration       `yaml:"conn_max_lifetime,omitempty" json:"conn_max_lifetime,omitempty"`
	ConnMaxIdleTime time.Duration       `yaml:"conn_max_idle_time,omitempty" json:"conn_max_idle_time,omitempty"`
	Host            string              `yaml:"host" json:"host"`
	Port            int                 `yaml:"port" json:"port"`
	User            string              `yaml:"user" json:"user"`
	Password        string              `yaml:"password" json:"password"`
	Database        string              `yaml:"database" json:"database"`
	Charset         string              `yaml:"charset,omitempty" json:"charset,omitempty"`
	Timeout         int                 `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

type InitDatabasesInput struct {
	Overwrite       bool                       `json:"overwrite"`
	DefaultDatabase string                     `json:"default_database"`
	Databases       map[string]*DatabaseConfig `json:"databases"`
}

type WriteConfigResult struct {
	Success         bool   `json:"success"`
	FilePath        string `json:"file_path"`
	DatabaseCount   int    `json:"database_count"`
	DefaultDatabase string `json:"default_database"`
	Overwritten     bool   `json:"overwritten"`
}

// ConfigPath 返回插件内部 config.yaml 的绝对路径。
func ConfigPath() (string, error) {
	root := os.Getenv(pluginRootEnv)
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("获取当前工作目录失败: %w", err)
		}
		root = wd
	}
	return filepath.Join(root, internalConfigPath), nil
}

// InitConf 从插件内部 config.yaml 读取并填充全局 Conf。
func InitConf() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("解析配置 YAML 失败: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return err
	}
	Conf = cfg
	return nil
}

// WriteInternalConfig 将连接配置写回插件内部 config.yaml。
func WriteInternalConfig(input *InitDatabasesInput) (*WriteConfigResult, error) {
	if input == nil {
		return nil, fmt.Errorf("初始化输入不能为空")
	}
	cfg := &Config{
		DefaultDatabase: input.DefaultDatabase,
		Databases:       input.Databases,
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}

	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	if !input.Overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil, fmt.Errorf("配置文件已存在，且 overwrite=false: %s", path)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, fmt.Errorf("写入配置文件失败: %w", err)
	}

	Conf = *cfg
	return &WriteConfigResult{
		Success:         true,
		FilePath:        path,
		DatabaseCount:   len(cfg.Databases),
		DefaultDatabase: cfg.DefaultDatabase,
		Overwritten:     input.Overwrite,
	}, nil
}

func validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}
	if len(cfg.Databases) == 0 {
		return fmt.Errorf("配置中至少要包含一个数据库连接")
	}
	if cfg.DefaultDatabase != "" {
		if _, ok := cfg.Databases[cfg.DefaultDatabase]; !ok {
			return fmt.Errorf("default_database '%s' 不在 databases 中", cfg.DefaultDatabase)
		}
	}
	for alias, dbCfg := range cfg.Databases {
		if alias == "" {
			return fmt.Errorf("数据库别名不能为空")
		}
		if dbCfg == nil {
			return fmt.Errorf("数据库 '%s' 配置不能为空", alias)
		}
		if dbCfg.Type == "" {
			dbCfg.Type = common.DatabaseTypeMySQL
		}
		if dbCfg.Host == "" || dbCfg.Port == 0 || dbCfg.User == "" || dbCfg.Password == "" || dbCfg.Database == "" {
			return fmt.Errorf("数据库 '%s' 的 host/port/user/password/database 不能为空", alias)
		}
	}
	return nil
}
