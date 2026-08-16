package utils

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	assert.NoError(t, err)

	assert.True(t, strings.HasPrefix(key, "hd_"))

	raw := strings.TrimPrefix(key, "hd_")
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(decoded))
}

func TestGenerateKeyUniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for range 1000 {
		k, err := GenerateKey()
		assert.NoError(t, err)

		s, ok := seen[k]
		assert.False(t, s)
		assert.False(t, ok)

		seen[k] = true
	}
}
