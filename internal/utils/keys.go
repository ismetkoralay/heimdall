package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

var ErrGeneratingKey = errors.New("error generating key")

func GenerateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("%w: %w", ErrGeneratingKey, err)
	}

	return "hd_" + base64.RawURLEncoding.EncodeToString(b), nil
}
