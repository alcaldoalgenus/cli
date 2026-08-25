package config

import (
	"errors"
	"os"
	"testing"
)

type mockKeyring struct {
	storage map[string]string
	err     error
}

func (m *mockKeyring) Get(service, key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	val, ok := m.storage[service]
	if !ok {
		return "", ErrNotFound
	}
	return val, nil
}

func (m *mockKeyring) Set(service, key, value string) error {
	if m.storage == nil {
		m.storage = make(map[string]string)
	}
	m.storage[service] = value
	return nil
}

func (m *mockKeyring) Delete(service, key string) error {
	delete(m.storage, service)
	return nil
}

type mockConfig struct {
	storage map[string]map[string]string
}

func (m *mockConfig) Get(host, key string) (string, error) {
	if hostMap, ok := m.storage[host]; ok {
		if val, ok := hostMap[key]; ok {
			return val, nil
		}
	}
	return "", errors.New("not found")
}

func (m *mockConfig) Set(host, key, value string) error {
	if m.storage == nil {
		m.storage = make(map[string]map[string]string)
	}
	if m.storage[host] == nil {
		m.storage[host] = make(map[string]string)
	}
	m.storage[host][key] = value
	return nil
}

func TestAuthTokenForHost_KeyringOperationalError(t *testing.T) {
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")

	dbusErr := errors.New("dbus: connection closed")
	keyring := &mockKeyring{err: dbusErr}
	authConfig := NewAuthConfig(keyring, nil)

	_, _, err := authConfig.AuthTokenForHost("github.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var keyringErr *KeyringError
	if !errors.As(err, &keyringErr) {
		t.Fatalf("expected *KeyringError, got %T (%v)", err, err)
	}

	if keyringErr.Host != "github.com" {
		t.Errorf("expected host github.com, got %s", keyringErr.Host)
	}

	if !errors.Is(err, dbusErr) {
		t.Errorf("expected unwrapped error to be dbusErr, got %v", err)
	}
}

func TestAuthTokenForHost_KeyringNotFoundFallback(t *testing.T) {
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")

	keyring := &mockKeyring{storage: map[string]string{}}
	cfg := &mockConfig{
		storage: map[string]map[string]string{
			"github.com": {"oauth_token": "config-token"},
		},
	}
	authConfig := NewAuthConfig(keyring, cfg)

	token, source, err := authConfig.AuthTokenForHost("github.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "config-token" || source != "config" {
		t.Errorf("expected token 'config-token' from 'config', got '%s' from '%s'", token, source)
	}
}

func TestAuthTokenForHost_EnvPrecedenceOverKeyringError(t *testing.T) {
	t.Setenv("GH_TOKEN", "env-token")

	dbusErr := errors.New("dbus: connection closed")
	keyring := &mockKeyring{err: dbusErr}
	authConfig := NewAuthConfig(keyring, nil)

	token, source, err := authConfig.AuthTokenForHost("github.com")
	if err != nil {
		t.Fatalf("unexpected error when env var is present: %v", err)
	}
	if token != "env-token" || source != "GH_TOKEN" {
		t.Errorf("expected token 'env-token' from 'GH_TOKEN', got '%s' from '%s'", token, source)
	}
}
