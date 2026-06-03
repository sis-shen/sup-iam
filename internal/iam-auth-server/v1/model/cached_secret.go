package model

import "time"

type CachedSecret struct {
	AccessKey string
	SecretKey string
	ExpiredAt time.Time
}
