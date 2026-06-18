package db

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"local-db-access/internal/common"
	"local-db-access/internal/config"
)

// DatabaseManager 以别名为 key 管理全部数据库连接
type DatabaseManager struct {
	mu        sync.RWMutex
	databases map[string]DatabaseQueryTool
	configs   map[string]*config.DatabaseConfig
}

func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{
		databases: make(map[string]DatabaseQueryTool),
		configs:   make(map[string]*config.DatabaseConfig),
	}
}

// AddDatabase 以别名注册一条数据库连接
func (dm *DatabaseManager) AddDatabase(alias string, dbConfig *config.DatabaseConfig) error {
	if alias == "" {
		return errors.New("数据库别名不能为空")
	}
	if dbConfig == nil {
		return errors.New("数据库配置不能为空")
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.databases[alias]; exists {
		return fmt.Errorf("数据库别名 '%s' 已存在", alias)
	}

	tool, err := newDatabaseTool(alias, dbConfig)
	if err != nil {
		return fmt.Errorf("创建数据库工具失败: %w", err)
	}

	dm.databases[alias] = tool
	dm.configs[alias] = dbConfig
	return nil
}

// AddConfigOnly 仅注册配置信息（用于未启用的连接，不建立实际数据库连接）
func (dm *DatabaseManager) AddConfigOnly(alias string, dbConfig *config.DatabaseConfig) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.configs[alias] = dbConfig
}

// ReloadResult 配置重载结果
type ReloadResult struct {
	Added    []string          `json:"added"`
	Removed  []string          `json:"removed"`
	Updated  []string          `json:"updated"`
	Errors   map[string]string `json:"errors,omitempty"`
}

// ReloadConfig 从磁盘重新读取配置文件并热更新数据库连接，无需重启服务。
func (dm *DatabaseManager) ReloadConfig() (*ReloadResult, error) {
	if err := config.InitConf(); err != nil {
		return nil, fmt.Errorf("重新读取配置失败: %w", err)
	}

	result := &ReloadResult{
		Removed: make([]string, 0),
		Added:   make([]string, 0),
		Updated: make([]string, 0),
		Errors:  make(map[string]string),
	}

	// 获取当前连接快照（读锁）
	dm.mu.RLock()
	existingConfigs := make(map[string]*config.DatabaseConfig, len(dm.configs))
	for k, v := range dm.configs {
		existingConfigs[k] = v
	}
	dm.mu.RUnlock()

	newConfigs := config.Conf.Databases

	// 计算变更集
	toRemove := make(map[string]bool)
	toAdd := make(map[string]*config.DatabaseConfig)
	toUpdate := make(map[string]*config.DatabaseConfig)

	for alias, cfg := range newConfigs {
		if !cfg.Enabled {
			toRemove[alias] = true
			continue
		}
		existingCfg, exists := existingConfigs[alias]
		if !exists {
			toAdd[alias] = cfg
		} else if configChanged(existingCfg, cfg) {
			toUpdate[alias] = cfg
		}
	}
	for alias := range existingConfigs {
		if _, exists := newConfigs[alias]; !exists {
			toRemove[alias] = true
		}
	}

	// 创建新连接（不持锁，避免 I/O 阻塞）
	newTools := make(map[string]DatabaseQueryTool)
	for alias, cfg := range toAdd {
		tool, err := newDatabaseTool(alias, cfg)
		if err != nil {
			result.Errors[alias] = fmt.Sprintf("添加失败: %v", err)
			continue
		}
		newTools[alias] = tool
		result.Added = append(result.Added, alias)
	}

	updateTools := make(map[string]DatabaseQueryTool)
	for alias, cfg := range toUpdate {
		tool, err := newDatabaseTool(alias, cfg)
		if err != nil {
			result.Errors[alias] = fmt.Sprintf("更新失败: %v", err)
			continue
		}
		updateTools[alias] = tool
		result.Updated = append(result.Updated, alias)
	}

	// 应用变更（写锁）
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for alias := range toRemove {
		if tool, exists := dm.databases[alias]; exists {
			_ = tool.Close()
			delete(dm.databases, alias)
		}
		// 不删除 dm.configs 中的未启用配置，保留给 list_databases 展示
		if _, kept := dm.configs[alias]; !kept {
			// 如果配置在 newConfigs 中（只是被禁用），保留它
			if cfg, ok := newConfigs[alias]; ok {
				dm.configs[alias] = cfg
			} else {
				delete(dm.configs, alias)
			}
		} else {
			// 已存在于 dm.configs 的，更新为最新配置
			if cfg, ok := newConfigs[alias]; ok {
				dm.configs[alias] = cfg
			}
		}
		result.Removed = append(result.Removed, alias)
	}

	for alias, tool := range newTools {
		dm.databases[alias] = tool
		dm.configs[alias] = newConfigs[alias]
	}

	for alias, tool := range updateTools {
		if oldTool, exists := dm.databases[alias]; exists {
			_ = oldTool.Close()
		}
		dm.databases[alias] = tool
		dm.configs[alias] = newConfigs[alias]
	}

	return result, nil
}

// configChanged 判断数据库配置是否实质变更（需要重建连接）。
func configChanged(a, b *config.DatabaseConfig) bool {
	return a.Host != b.Host ||
		a.Port != b.Port ||
		a.User != b.User ||
		a.Password != b.Password ||
		a.Database != b.Database ||
		a.Charset != b.Charset ||
		a.Type != b.Type
}

// newDatabaseTool 根据配置类型创建对应的数据库工具实例。
func newDatabaseTool(alias string, dbConfig *config.DatabaseConfig) (DatabaseQueryTool, error) {
	switch dbConfig.Type {
	case common.DatabaseTypeMySQL:
		return NewMySQLQueryTool(alias, dbConfig)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbConfig.Type)
	}
}

func (dm *DatabaseManager) GetDatabase(alias string) (DatabaseQueryTool, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	tool, exists := dm.databases[alias]
	if !exists {
		return nil, fmt.Errorf("数据库别名 '%s' 不存在", alias)
	}
	return tool, nil
}

func (dm *DatabaseManager) GetConfig(alias string) (*config.DatabaseConfig, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	cfg, ok := dm.configs[alias]
	return cfg, ok
}

// Aliases 返回已注册别名（字典序，仅启用的连接）
func (dm *DatabaseManager) Aliases() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	aliases := make([]string, 0, len(dm.databases))
	for alias := range dm.databases {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	return aliases
}

// AllAliases 返回配置中全部别名（字典序，包含未启用的连接）
func (dm *DatabaseManager) AllAliases() []string {
	// 优先从当前运行时配置获取全部（包含未启用）
	allCfgs := config.Conf.Databases
	if len(allCfgs) > 0 {
		aliases := make([]string, 0, len(allCfgs))
		for alias := range allCfgs {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)
		return aliases
	}
	// 回退到已注册的连接
	return dm.Aliases()
}

// ListTables 返回别名对应数据库的全部表名(走 SHOW TABLES)
func (dm *DatabaseManager) ListTables(ctx context.Context, alias string) ([]string, error) {
	tool, err := dm.GetDatabase(alias)
	if err != nil {
		return nil, err
	}
	result, err := tool.ExecuteQuery(ctx, "SHOW TABLES", nil, 0)
	if err != nil {
		return nil, err
	}
	tables := make([]string, 0, len(result.Data))
	for _, row := range result.Data {
		for _, v := range row {
			tables = append(tables, fmt.Sprintf("%v", v))
			break
		}
	}
	sort.Strings(tables)
	return tables, nil
}

func (dm *DatabaseManager) HealthCheckAll(ctx context.Context) map[string]error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	results := make(map[string]error, len(dm.databases))
	for alias, tool := range dm.databases {
		results[alias] = tool.HealthCheck(ctx)
	}
	return results
}
