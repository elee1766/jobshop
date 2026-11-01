// Package cli implements the jobshop command line: the controller, the
// session host and every client subcommand. cmd/jobshop only calls Main.
package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/host"
)

// Set at build time with -ldflags "-X github.com/elee1766/jobshop/pkg/cli.version=...".
var (
	version = "dev"
	commit  = "unknown"
)

// Globals are flags shared by every subcommand.
type Globals struct {
	Root   string `help:"Jobshop root directory (holds jobshop.yaml, the database, sockets and sessions). Default: nearest ancestor of the cwd containing jobshop.yaml." env:"JOBSHOP_ROOT" type:"path"`
	Socket string `help:"Socket to talk to. Defaults to the session CLI socket, else the root's admin socket." env:"JOBSHOP_SOCKET"`
}

// root resolves the Jobshop root directory.
func (g *Globals) root() (config.Root, error) {
	wd, err := os.Getwd()
	if err != nil {
		return config.Root{}, err
	}
	return config.FindRoot(g.Root, wd)
}

// root is the kong command tree.
type root struct {
	Globals

	Init        initCmd        `cmd:"" help:"Create a Jobshop root (jobshop.yaml) in a directory."`
	Serve       serveCmd       `cmd:"" help:"Run the controller for the root."`
	SessionHost sessionHostCmd `cmd:"" name:"session-host" hidden:"" help:"Run a session host (spawned by the controller)."`
	Shutdown    shutdownCmd    `cmd:"" help:"Stop the controller. Sessions keep running unless --all."`

	Ticket    ticketCmd    `cmd:"" help:"Create, show, list, update and close tickets."`
	Close     closeCmd     `cmd:"" help:"Close a ticket (default: the session's current one)."`
	Comment   commentCmd   `cmd:"" help:"Comment on a ticket."`
	Assign    assignCmd    `cmd:"" help:"Pre-assign a ticket to a session (admin)."`
	Claim     claimCmd     `cmd:"" help:"Claim a ticket, or the next one with --next."`
	Heartbeat heartbeatCmd `cmd:"" help:"Renew the lease on a held ticket."`
	Release   releaseCmd   `cmd:"" help:"Hand a held ticket back to the queue."`
	Ready     readyCmd     `cmd:"" help:"List claimable tickets."`
	Edge      edgeCmd      `cmd:"" help:"Manage dependencies between tickets."`
	Label     labelCmd     `cmd:"" help:"Manage labels."`
	Session   sessionCmd   `cmd:"" help:"Manage sessions."`
	Whoami    whoamiCmd    `cmd:"" help:"Show the caller's scope."`
	Version   versionCmd   `cmd:"" help:"Print version information."`
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
		return exitCode(err)
	}
	return 0
}

type versionCmd struct{}

func (versionCmd) Run() error {
	fmt.Printf("jobshop %s (%s)\n", version, commit)
	return nil
}

type sessionHostCmd struct {
	Dir string `arg:"" help:"Session directory containing spec.json." type:"path"`
}

func (c sessionHostCmd) Run() error { return host.Main(c.Dir) }
