# Model Data Example

## MySQL Model

```go
package zybuos

import (
    "gorm.io/gorm"
    "zyb-uos-mis/models/mysql"
)

type FlowConfig struct {
    ID         int64          `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
    AppId      string         `gorm:"column:app_id;NOT NULL" json:"app_id"`
    Version    string         `gorm:"column:version;NOT NULL" json:"version"`
    Config     string         `gorm:"column:config;NOT NULL" json:"config"`
    CreateTime mysql.DateTime `gorm:"column:create_time;type:datetime" json:"create_time"`
    UpdateTime mysql.DateTime `gorm:"column:update_time;type:datetime" json:"update_time"`
}

func (t *FlowConfig) TableName() string {
    return "tbl_flow_config"
}

func (t *FlowConfig) UniqueKeys() []string {
    return []string{"app_id", "version"}
}

func (t *FlowConfig) GetID() int64 {
    return t.ID
}

func (t *FlowConfig) SoftDeleted() bool {
    return false
}

type FlowConfigModel struct {
    *mysql.BaseModel[FlowConfig, *FlowConfig]
}

func NewFlowConfigModel(db *gorm.DB) *FlowConfigModel {
    return &FlowConfigModel{
        BaseModel: mysql.NewBaseModel[FlowConfig](db),
    }
}
```

## Init Pattern

```go
package zybuos

import "gorm.io/gorm"

var (
    DB                 *gorm.DB
    FlowConfigModelIns *FlowConfigModel
)

func InitBaseZybUOS(db *gorm.DB) {
    DB = db
    FlowConfigModelIns = NewFlowConfigModel(db)
}
```

## Data Access

```go
package mysql

type FlowConfigData struct{}

var FlowConfigDataIns = &FlowConfigData{}

func (d *FlowConfigData) GetFlowConfigList(ctx *gin.Context, appId string) ([]*zybuos.FlowConfig, error) {
    wb := mysql.AND(
        mysql.EQ("app_id", appId),
        mysql.IsNotDeleted(),
    )
    return zybuos.FlowConfigModelIns.GetBy(ctx, wb)
}

func (d *FlowConfigData) UpdateFlowConfig(ctx *gin.Context, id int64, fields map[string]interface{}) error {
    return zybuos.FlowConfigModelIns.UpdateByID(ctx, id, fields)
}
```
