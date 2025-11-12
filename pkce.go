package octanox

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

// StringStateMap stores string values by key with expiry similar to StateMap.
type StringStateMap map[string]string

func (s StringStateMap) Store(key, value string, seconds int) {
	s[key] = value
	go func(k string) {
		<-time.After(time.Duration(seconds) * time.Second)
		delete(s, k)
	}(key)
}

func (s StringStateMap) Pop(key string) string {
	val, ok := s[key]
	if ok {
		delete(s, key)
		return val
	}
	return ""
}

// generatePKCE returns a (verifier, challenge) pair using S256 method.
func generatePKCE() (string, string) {
	// Generate 32 bytes (results in 43-char base64url encoded string)
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	verifier := base64.RawURLEncoding.EncodeToString(b)

	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge
}
