package model

import "time"

// PolicyAudit records policy change history
type PolicyAudit struct {
	ID           uint64    `gorm:"column:id;comment:自增主键"`
	InstanceID   string    `gorm:"column:instanceID;comment:跨域 UUID"`
	Name         string    `gorm:"column:name;comment:策略名称"`
	Username     string    `gorm:"column:username;comment:所属用户名"`
	Description  *string   `gorm:"column:description;comment:策略描述"`
	PolicyShadow *string   `gorm:"column:policyShadow;comment:策略 DSL 描述"`
	ExtendShadow *string   `gorm:"column:extendShadow;comment:扩展字段"`
	CreatedAt    time.Time `gorm:"column:createdAt;comment:创建时间"`
}
