package api

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alcaldoalgenus/cli/internal/config"
)

type testKeyring struct {
	err error
}

func (k *testKeyring) Get(service, key string) (string, error) {
	if k.err != nil {
		return "", k.err
	}
	return "", config.ErrNotFound
}

func (k *testKeyring) Set(service, key, value string) error {
	return nil
}

func (k *testKeyring) Delete(service, key string) error {
	return nil
}

func TestApiRun_KeyringOperationalFailure(t *testing.T)	{
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	errOut := &bytes.Buffer{}
	iostreams := &IOStreams{
		ErrOut: errOut,
		Out:    &bytes.Buffer{},
	}

	authCfg := config.NewAuthConfig(&testKeyring{err: errors.New("dbus: connection closed")}, nil)

	opts := &ApiOptions{
		IOStreams:  iostreams,
		AuthConfig: authCfg,
		Hostname:   "github.com",
		Endpoint:   "user",
	}

	err := ApiRun(context.Background(), opts)
	if err == nil {
		t.Fatalf("expected ApiRun to return error, got nil")
	}

	var keyringErr *config.KeyringError
	if !errors.As(err, &keyringErr) {
		t.Fatalf("expected KeyringError, got %T: %v", err, err)
	}

	errOutput := errOut.String()
	if !strings.Contains(errOutput, "failed to read credentials from keyring for github.com: dbus: connection closed") {
		t.Errorf("expected descriptive error output, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "GH_TOKEN") {
		t.Errorf("expected workaround suggestion with GH_TOKEN, got: %s", errOutput)
	}
}
