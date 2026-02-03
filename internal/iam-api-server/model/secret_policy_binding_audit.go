package model

import "time"

// SecretPolicyBindingAudit records binding change history
type SecretPolicyBindingAudit struct {
	ID           uint64
	SecretID     uint64
	PolicyID     uint64
	Username     string
	ExtendShadow *string
	CreatedAt    time.Time
}
