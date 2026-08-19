package auth

import (
	"fmt"
	"time"
)

type APIKey struct {
	ID              string
	HashedKey       string
	Name            string
	RPMLimit        int
	DailyTokenQuota int
	Revoked         bool
	CreatedAt       time.Time
}

type AuthError struct {
	StatusCode int
	Err        error
}

func (ae *AuthError) Error() string {
	if ae.Err != nil {
		return fmt.Sprintf("status code: %d - error: %s", ae.StatusCode, ae.Err.Error())
	}

	return ""
}

func (ae *AuthError) Unwrap() error {
	return ae.Err
}
