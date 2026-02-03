package model

import "time"

// Policy represents permission policy
type Policy struct {
	ID           uint64
	InstanceID   string
	Name         string
	Username     string
	Description  *string
	PolicyShadow *string
	ExtendShadow *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
