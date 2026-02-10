package apikey

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Service handles API key generation and validation
type Service struct {
	prefix string
	length int
}

// NewService creates a new API key service
func NewService(prefix string, length int) *Service {
	return &Service{
		prefix: prefix,
		length: length,
	}
}

// Generate creates a new API key
// Returns the plaintext key (to show to user) and the hashed key (to store in DB)
func (s *Service) Generate() (key, hashedKey string, err error) {
	// Generate random bytes
	bytes := make([]byte, s.length)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Encode to base64 and add prefix
	randomPart := base64.RawURLEncoding.EncodeToString(bytes)
	key = s.prefix + randomPart

	// Hash the key for storage
	hashedKey = s.Hash(key)

	return key, hashedKey, nil
}

// Hash creates an argon2id hash of the API key
func (s *Service) Hash(key string) string {
	// Argon2id parameters (OWASP recommendations)
	salt := []byte("telegram-bot-gateway-salt-change-me") // In production, use unique salt per key
	hash := argon2.IDKey([]byte(key), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(hash)
}

// Verify checks if a plaintext key matches a hashed key
func (s *Service) Verify(key, hashedKey string) bool {
	computedHash := s.Hash(key)
	return computedHash == hashedKey
}

// IsValid checks if a key has the correct format
func (s *Service) IsValid(key string) bool {
	if !strings.HasPrefix(key, s.prefix) {
		return false
	}
	// Remove prefix and check if remaining part is valid base64
	keyPart := strings.TrimPrefix(key, s.prefix)
	if len(keyPart) == 0 {
		return false
	}
	_, err := base64.RawURLEncoding.DecodeString(keyPart)
	return err == nil
}

// ExtractPrefix returns the prefix from a key
func (s *Service) ExtractPrefix(key string) string {
	if strings.HasPrefix(key, s.prefix) {
		return s.prefix
	}
	return ""
}
