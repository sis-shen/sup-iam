package model

import "time"

// User represents IAM control plane user
type User struct {
	ID           uint64     `gorm:"column:id;comment:自增主键"`
	InstanceID   string     `gorm:"column:instanceID;comment:跨域 UUID"`
	Username     string     `gorm:"column:username;comment:用户名"`
	Nickname     string     `gorm:"column:nickname;comment:昵称"`
	PasswordHash string     `gorm:"column:passwordHash;comment:密码哈希值"`
	IsEnable     uint8      `gorm:"column:isEnable;comment:是否启用：1-可用，0-不可用"`
	Phone        *string    `gorm:"column:phone;comment:手机号"`
	Email        *string    `gorm:"column:email;comment:邮箱"`
	IsAdmin      uint8      `gorm:"column:isAdmin;comment:是否管理员"`
	ExtendShadow *string    `gorm:"column:extendShadow;comment:扩展字段"`
	LoggedAt     *time.Time `gorm:"column:loggedAt;comment:最近登录时间"`
	CreatedAt    time.Time  `gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt    time.Time  `gorm:"column:updatedAt;comment:更新时间"`
}
