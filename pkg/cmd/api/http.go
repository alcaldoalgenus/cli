package api

import (
	"fmt"
	"net/http"

	"github.com/alcaldoalgenus/cli/internal/config"
)

type roundTripper struct {
	token string
	base  http.RoundTripper
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", rt.token))
	}
	base := rt.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// NewHTTPClient returns an authenticated HTTP client for the host using AuthConfig.
// If keyring retrieval results in an operational error, that error is returned immediately.
func NewHTTPClient(authCfg *config.AuthConfig, hostname string) (*http.Client, error) {
	var token string
	if authCfg != nil {
		t, _, err := authCfg.AuthTokenForHost(hostname)
		if err != nil {
			return nil, err
		}
		token = t
	}

	client := &http.Client{
		Transport: &roundTripper{
			token: token,
		},
	}
	return client, nil
}
