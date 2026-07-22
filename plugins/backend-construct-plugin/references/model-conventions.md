# Model Conventions

## Scope

本文件整理 model 相关的细规则，用于补充 `knowledge/data.md` 与 `examples/model-data-example.md`，避免主 knowledge 过重。

## File Layout

若项目采用现有约定，优先沿用：

- `models/mysql/<数据库名>/<表名>.go`
- `models/mysql/<数据库名>/a_init.go`
- `models/redis/<file>.go`

若目标项目没有这套结构，以现有项目目录为准，不强推迁移。

## MySQL Model Shape

常见模型应包含：

- GORM tag
- JSON tag
- `TableName()`
- `UniqueKeys()`
- `GetID()`
- `SoftDeleted()`

若项目存在统一 BaseModel 封装，优先复用，不在当前业务修改中重新设计 ORM 基类。

## Type Conventions

- 主键 ID 优先 `int64`
- 时间字段若项目已有封装，优先沿用如 `mysql.DateTime`
- 唯一键、逻辑删除、分表键等能力优先复用现有接口约定

## Init Pattern

若项目使用初始化文件，常见形态为：

- `DB *gorm.DB`
- `XxxModelIns *XxxModel`
- `InitBaseXxx(db *gorm.DB)`

修改 model/data 时，应检查是否需要同步更新初始化入口。

## Redis Model

Redis 结构体与 MySQL model 不同：

- 纯缓存对象更偏序列化结构
- 不要强行补 GORM tag
- 优先围绕 `json.Marshal / Unmarshal` 的使用方式设计

## Sharding / Special Cases

若项目存在分表模型：

- 构造函数通常需要接收分表键
- 不要把普通表与分表模型混成一类
- 必须单独说明分表键来源和影响范围
