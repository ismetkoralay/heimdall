package auth

import "time"

type APIKey struct {
	ID              string
	HashedKey       string
	Name            string
	RPMLimit        int
	DailyTokenQuota int
	Revoked         bool
	CreatedAt       time.Time
}
