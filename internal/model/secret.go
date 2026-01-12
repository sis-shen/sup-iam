package model

import "time"

// Secret represents AccessKey / SecretKey resource
type Secret struct {
	ID           uint64
	InstanceID   string
	UserID       uint64
	Username     string
	AccessKey    string
	SecretKey    string
	Expires      int64
	Description  *string
	ExtendShadow *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
