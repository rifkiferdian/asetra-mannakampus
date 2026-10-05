package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

const minimumSessionSecretLength = 32

// SessionSecret returns the cookie authentication key. Production must supply
// a stable secret; development gets an ephemeral key so no secret is embedded
// in source code.
func SessionSecret() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if len(secret) >= minimumSessionSecretLength {
		return []byte(secret), nil
	}

	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		return nil, fmt.Errorf("SESSION_SECRET must contain at least %d characters in production", minimumSessionSecretLength)
	}

	generated := make([]byte, 48)
	if _, err := rand.Read(generated); err != nil {
		return nil, errors.New("generate development session secret: " + err.Error())
	}
	log.Printf("WARNING: SESSION_SECRET is missing or too short; using an ephemeral development secret")
	return []byte(base64.RawURLEncoding.EncodeToString(generated)), nil
}
