package model

import "time"

// SecretPolicyBindingAudit records binding change history
type SecretPolicyBindingAudit struct {
	ID           uint64    `gorm:"column:id;comment:自增主键"`
	SecretID     uint64    `gorm:"column:secretID;comment:密钥 ID"`
	PolicyID     uint64    `gorm:"column:policyID;comment:策略 ID"`
	Username     string    `gorm:"column:username;comment:所属用户名"`
	ExtendShadow *string   `gorm:"column:extendShadow;comment:扩展字段"`
	CreatedAt    time.Time `gorm:"column:createdAt;comment:创建时间"`
}
