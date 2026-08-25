package config

import (
	"errors"
	"fmt"
	"os"
)

// ErrNotFound indicates that the requested credential was not found in the keyring.
var ErrNotFound = errors.New("secret not found in keyring")

// Keyring provides an interface for interacting with system credential storage.
type Keyring interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
}

// KeyringError represents an operational failure while accessing the keyring.
type KeyringError struct {
	Host string
	Err  error
}

func (e *KeyringError) Error() string {
	return fmt.Sprintf("failed to read credentials from keyring for %s: %v\nTo bypass the keyring, set the GH_TOKEN environment variable.", e.Host, e.Err)
}

func (e *KeyringError) Unwrap() error {
	return e.Err
}

// Config represents the CLI host configuration interface.
type Config interface {
	Get(host, key string) (string, error)
	Set(host, key, value string) error
}

// AuthConfig handles credential resolution across environment variables, keyring, and config files.
type AuthConfig struct {
	Keyring Keyring
	Config  Config
}

func NewAuthConfig(k Keyring, c Config) *AuthConfig {
	return &AuthConfig{
		Keyring: k,
		Config:  c,
	}
}

// AuthTokenForHost returns the authentication token and token source for a host.
// It returns an error if keyring access encounters an operational failure.
func (a *AuthConfig) AuthTokenForHost(hostname string) (string, string, error) {
	// 1. Precedence: GH_TOKEN env var
	if val := os.Getenv("GH_TOKEN"); val != "" {
		return val, "GH_TOKEN", nil
	}

	// 2. Precedence: GITHUB_TOKEN env var
	if val := os.Getenv("GITHUB_TOKEN"); val != "" {
		return val, "GITHUB_TOKEN", nil
	}

	// 3. Precedence: Keyring
	if a.Keyring != nil {
		token, err := a.TokenFromKeyring(hostname)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return "", "", &KeyringError{Host: hostname, Err: err}
			}
		} else if token != "" {
			return token, "keyring", nil
		}
	}

	// 4. Precedence: Configuration file
	if a.Config != nil {
		token, err := a.Config.Get(hostname, "oauth_token")
		if err == nil && token != "" {
			return token, "config", nil
		}
	}

	return "", "none", nil
}

// TokenFromKeyring retrieves the token for the given hostname from the keyring.
// Operational errors (e.g. DBus, locked keychain) are returned directly.
func (a *AuthConfig) TokenFromKeyring(hostname string) (string, error) {
	if a.Keyring == nil {
		return "", ErrNotFound
	}
	val, err := a.Keyring.Get("gh:"+hostname, "")
	if err != nil {
		return "", err
	}
	return val, nil
}
