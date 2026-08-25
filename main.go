package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alcaldoalgenus/cli/internal/config"
	"github.com/alcaldoalgenus/cli/pkg/cmd/api"
)

func main() {
	authConfig := config.NewAuthConfig(nil, nil)
	opts := &api.ApiOptions{
		IOStreams: &api.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		AuthConfig: authConfig,
		Hostname:   "github.com",
		Endpoint:   "user",
	}

	if err := api.ApiRun(context.Background(), opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
