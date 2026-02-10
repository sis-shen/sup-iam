package model

import "time"

// Secret represents AccessKey / SecretKey resource
type Secret struct {
	ID           uint64    `gorm:"column:id;comment:自增主键"`
	InstanceID   string    `gorm:"column:instanceID;comment:跨域 UUID"`
	UserID       uint64    `gorm:"column:userID;comment:所属用户 ID"`
	Username     string    `gorm:"column:username;comment:用户名（冗余字段）"`
	AccessKey    string    `gorm:"column:accessKey;comment:AccessKey"`
	SecretKey    string    `gorm:"column:secretKey;comment:SecretKey（明文存储）"`
	Expires      int64     `gorm:"column:expires;comment:过期时间（Unix 秒）"`
	Description  *string   `gorm:"column:description;comment:密钥描述"`
	ExtendShadow *string   `gorm:"column:extendShadow;comment:扩展字段"`
	CreatedAt    time.Time `gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"column:updatedAt;comment:更新时间"`
}
