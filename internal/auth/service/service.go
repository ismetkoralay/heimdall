package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ismetkoralay/heimdall/internal/auth"
)

var (
	ErrSavingKey = errors.New("error saving key to db")
)

type Repository interface {
	CreateAPIKey(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error)
	GetAPIKeyByHashedKey(ctx context.Context, hashedKey string) (auth.APIKey, error)
}

type Service struct {
	repository    Repository
	generateKeyFn func() (string, error)
	nowFn         func() time.Time
}

func New(
	repository Repository,
	nowFn func() time.Time,
	generateKeyFn func() (string, error),
) *Service {
	return &Service{
		repository:    repository,
		nowFn:         nowFn,
		generateKeyFn: generateKeyFn,
	}
}

func (s *Service) GenerateAndSaveAPIKey(
	ctx context.Context,
	name string,
	rpmLimit int,
	dailyTokenQuota int,
) (string, error) {
	key, err := s.generateKeyFn()
	if err != nil {
		return "", err
	}

	_, err = s.repository.CreateAPIKey(ctx, auth.APIKey{
		HashedKey:       hashKey(key),
		Name:            name,
		RPMLimit:        rpmLimit,
		DailyTokenQuota: dailyTokenQuota,
		Revoked:         false,
		CreatedAt:       s.nowFn(),
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSavingKey, err)
	}

	return key, nil
}

func (s *Service) ValidateAPIKey(ctx context.Context, key string) (auth.APIKey, error) {
	res, err := s.repository.GetAPIKeyByHashedKey(ctx, hashKey(key))
	if err != nil {
		return auth.APIKey{}, err
	}

	if res.Revoked {
		return auth.APIKey{}, &auth.AuthError{
			StatusCode: http.StatusUnauthorized,
			Err:        errors.New("invalid token"),
		}
	}

	return res, nil
}

func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
