package cli

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// startTelemetry builds a long-running command's logger and tracer from
// the telemetry section of r's config. A missing config file exports
// nothing, so a command that does not otherwise need the config still runs.
// When only a pack fails, the last good config supplies the section.
func startTelemetry(ctx context.Context, r config.Root, name string, f LogFlags) (*telemetry.Process, error) {
	base, err := f.logger()
	if err != nil {
		return nil, err
	}
	var tc config.Telemetry
	if _, err := os.Stat(r.Config()); err == nil {
		// A failing pack does not change host settings such as telemetry, and
		// the controller serves the last good snapshot then (#227): do not
		// stop serve before it gets there (#1128). The pack error is dropped
		// here, the controller reports it.
		cfg, err := config.LoadOrLastGood(r)
		if cfg == nil {
			return nil, err
		}
		tc = cfg.Telemetry
	}
	bi, _ := debug.ReadBuildInfo()
	v, cm := versionInfo(version, commit, bi)
	// One resource for traces, logs and metrics (JOB-106).
	res, err := telemetry.NewResource(ctx, telemetry.Identity{
		Instance:   tc.Instance,
		Root:       r.Dir,
		Version:    v,
		Commit:     cm,
		Attributes: tc.ResourceAttrs(os.Getenv),
	})
	if err != nil {
		return nil, err
	}
	return telemetry.Start(ctx, name, telemetry.Export{
		TracesURL: tc.TracesURL(),
		LogsURL:   tc.LogsURL(),
		Headers:   tc.Headers(os.Getenv),
		Resource:  res,
	}, base.Handler())
}
