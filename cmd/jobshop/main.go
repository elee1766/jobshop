// Command jobshop is a thin entry point. Commands live in pkg/transport/cli;
// this file only hands it the controller, so the CLI adapter does not import
// pkg/orchestration/control.
package main

import (
	"context"
	"os"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/execution/session"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/transport/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], runController))
}

func runController(ctx context.Context, r config.Root, o cli.ServeOptions) error {
	return control.Run(ctx, r, control.Options{
		Tick:       o.Tick,
		Log:        o.Log,
		Tracer:     o.Telemetry.Tracer,
		Resource:   o.Telemetry.Resource,
		Reload:     o.Reload,
		Version:    o.Version,
		Commit:     o.Commit,
		ShellProbe: session.DefaultShellProbe,
	})
}
