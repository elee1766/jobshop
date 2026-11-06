package cli

import (
	"context"
	"os"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// startTelemetry builds a long-running command's logger and tracer from
// the telemetry section of r's config. A missing config file exports
// nothing, so a command that does not otherwise need the config still runs.
func startTelemetry(ctx context.Context, r config.Root, name string) (*telemetry.Process, error) {
	var tc config.Telemetry
	if _, err := os.Stat(r.Config()); err == nil {
		cfg, err := config.Load(r)
		if err != nil {
			return nil, err
		}
		tc = cfg.Telemetry
	}
	return telemetry.Start(ctx, name, telemetry.Export{
		TracesURL: tc.TracesURL(),
		LogsURL:   tc.LogsURL(),
		Headers:   tc.Headers(os.Getenv),
	}, os.Stderr)
}
