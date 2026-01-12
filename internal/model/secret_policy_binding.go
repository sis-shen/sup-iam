package model

import "time"

// SecretPolicyBinding represents m:n binding between Secret and Policy
type SecretPolicyBinding struct {
	ID           uint64
	SecretID     uint64
	PolicyID     uint64
	Username     string
	ExtendShadow *string
	CreatedAt    time.Time
}
