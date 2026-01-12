package model

import "time"

// PolicyAudit records policy change history
type PolicyAudit struct {
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
