package model

import "time"

type CachedSecret struct {
	ID        string
	AccessKey string
	SecretKey string
	ExpiredAt time.Time
}
