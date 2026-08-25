package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alcaldoalgenus/cli/internal/config"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

type ApiOptions struct {
	IOStreams  *IOStreams
	AuthConfig *config.AuthConfig
	HttpClient *http.Client
	Hostname   string
	Endpoint   string
	Method     string
	Body       io.Reader
}

func ApiRun(ctx context.Context, opts *ApiOptions) error {
	hostname := opts.Hostname
	if hostname == "" {
		hostname = "github.com"
	}

	client := opts.HttpClient
	if client == nil {
		var err error
		client, err = NewHTTPClient(opts.AuthConfig, hostname)
		if err != nil {
			if opts.IOStreams != nil && opts.IOStreams.ErrOut != nil {
				fmt.Fprintf(opts.IOStreams.ErrOut, "X %v\n", err)
			}
			return err
		}
	}

	endpoint := opts.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if hostname == "github.com" {
			endpoint = "https://api.github.com/" + strings.TrimPrefix(endpoint, "/")
		} else {
			endpoint = fmt.Sprintf("https://%s/api/v3/%s", hostname, strings.TrimPrefix(endpoint, "/"))
		}
	}

	method := opts.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, opts.Body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}

	if opts.IOStreams != nil && opts.IOStreams.Out != nil {
		_, err = io.Copy(opts.IOStreams.Out, resp.Body)
		return err
	}

	return nil
}
