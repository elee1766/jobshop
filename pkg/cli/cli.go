// Package cli implements the jobshop command line: the controller, the
// session host and every client subcommand. cmd/jobshop only calls Main.
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// Set at build time with -ldflags "-X github.com/elee1766/jobshop/pkg/cli.version=...".
var (
	version = "dev"
	commit  = "unknown"
)

// Globals are flags shared by every subcommand.
type Globals struct {
	Config string `help:"Path to jobshop.yaml." env:"JOBSHOP_CONFIG" type:"path"`
	Socket string `help:"Socket to talk to. Defaults to the session CLI socket, else the admin socket." env:"JOBSHOP_SOCKET" type:"path"`
}

// root is the kong command tree.
type root struct {
	Globals

	Serve       serveCmd       `cmd:"" help:"Run the controller."`
	SessionHost sessionHostCmd `cmd:"" name:"session-host" hidden:"" help:"Run a session host (spawned by the controller)."`
	Shutdown    shutdownCmd    `cmd:"" help:"End every session and stop the controller."`
	Version     versionCmd     `cmd:"" help:"Print version information."`
}

// Main parses args, runs the selected command and returns a process exit code.
func Main(args []string) int {
	var r root
	parser, err := kong.New(&r,
		kong.Name("jobshop"),
		kong.Description("Multi-agent work orchestration."),
		kong.UsageOnError(),
		kong.Exit(func(int) {}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	ctx, err := parser.Parse(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := ctx.Run(&r.Globals); err != nil {
		fmt.Fprintln(os.Stderr, "jobshop:", err)
		return 1
	}
	return 0
}

type versionCmd struct{}

func (versionCmd) Run() error {
	fmt.Printf("jobshop %s (%s)\n", version, commit)
	return nil
}

var errNotImplemented = errors.New("not implemented yet")

type serveCmd struct{}

func (serveCmd) Run(*Globals) error { return errNotImplemented }

type sessionHostCmd struct {
	ID string `arg:"" help:"Session id."`
}

func (sessionHostCmd) Run(*Globals) error { return errNotImplemented }

type shutdownCmd struct{}

func (shutdownCmd) Run(*Globals) error { return errNotImplemented }
