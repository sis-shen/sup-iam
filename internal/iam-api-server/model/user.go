package model

import "time"

// User represents IAM control plane user
type User struct {
	ID           uint64
	InstanceID   string
	Username     string
	Nickname     string
	IsEnable     uint8
	Phone        *string
	Email        *string
	IsAdmin      uint8
	ExtendShadow *string
	LoggedAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
