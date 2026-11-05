// Package cli implements the jobshop command line: the controller, the
// session host and every client subcommand. cmd/jobshop only calls Main.
package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"

	host "github.com/elee1766/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hostspec"
	"github.com/elee1766/jobshop/pkg/rpc"
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
	Drain       drainCmd       `cmd:"" help:"Start no new sessions or tickets; running turns finish (admin)."`
	Undrain     undrainCmd     `cmd:"" help:"Undo drain (admin)."`
	Reload      reloadCmd      `cmd:"" help:"Re-read jobshop.yaml and packs; keeps the old config if invalid (admin)."`
	Gateway     gatewayCmd     `cmd:"" help:"Run a chat gateway from jobshop.yaml (admin socket client)."`
	Desk        deskCmd        `cmd:"" help:"Serve Jobshop Desk, the read-only dashboard (admin socket client)."`
	Ingest      ingestCmd      `cmd:"" help:"Receive GitHub and Linear webhooks into the ingest spool."`

	Ticket    ticketCmd    `cmd:"" help:"Create, show, list, update and close tickets."`
	Close     closeCmd     `cmd:"" help:"Close a ticket (default: the session's current one)."`
	Comment   commentCmd   `cmd:"" help:"Comment on a ticket."`
	Guard     guardCmd     `cmd:"" help:"Ask the guardrail hooks before an outside action; exits 1 unless allowed."`
	Subscribe subscribeCmd `cmd:"" help:"Follow a GitHub issue or PR thread from a ticket; new human comments come back to it."`
	Wait      waitCmd      `cmd:"" help:"Park a ticket until a PR review, checks, PR state, a reply or a time holds (#381)."`
	Send      sendCmd      `cmd:"" help:"Send a message to a session or pool."`
	Assign    assignCmd    `cmd:"" help:"Pre-assign a ticket to a session (admin)."`
	Claim     claimCmd     `cmd:"" help:"Claim a ticket, or the next one with --next."`
	Heartbeat heartbeatCmd `cmd:"" help:"Renew the lease on a held ticket."`
	Release   releaseCmd   `cmd:"" help:"Hand a held ticket back to the queue."`
	Handoff   handoffCmd   `cmd:"" help:"Move a held ticket to another pool, release it and leave a note."`
	Ready     readyCmd     `cmd:"" help:"List claimable tickets."`
	Edge      edgeCmd      `cmd:"" help:"Manage dependencies between tickets."`
	Label     labelCmd     `cmd:"" help:"Manage labels."`
	Session   sessionCmd   `cmd:"" help:"Manage sessions."`
	Run       runCmd       `cmd:"" help:"Start and inspect workflow runs."`
	Batch     batchCmd     `cmd:"" help:"Start ordered groups of runs."`
	Resource  resourceCmd  `cmd:"" help:"Provider-owned resource slots: add, remove, unhealthy, list."`
	Trigger   triggerCmd   `cmd:"" help:"Inspect ticket triggers (admin)."`
	Pool      poolCmd      `cmd:"" help:"Inspect and unquarantine pools (admin)."`
	Attention attentionCmd `cmd:"" help:"List what waits on a human, oldest first; empty when healthy (admin)."`
	Activity  activityCmd  `cmd:"" help:"Recent GitHub and Linear activity in the repos and teams jobshop works, newest first (admin or pool grant activity)."`
	Report    reportCmd    `cmd:"" help:"Report your ticket's state (progress, waiting, ready, blocked, failed, dropped), or reliability reports (admin)."`
	Answer    answerCmd    `cmd:"" help:"Answer a worker's question; the answer goes to the asking session (admin)."`
	Learn     learnCmd     `cmd:"" help:"Record a correction, failure or workaround for the learning loop, or list learnings."`
	Config    configCmd    `cmd:"" help:"Inspect the active config (admin)."`
	Memory    memoryCmd    `cmd:"" help:"Memory files and skills from a git directory (#266)."`
	Events    eventsCmd    `cmd:"" help:"Print the change stream as JSON lines (admin, for debugging)."`
	Identity  identityCmd  `cmd:"" help:"GitHub App and Linear identities: list, check, tokens and the git credential helper."`
	Linear    linearCmd    `cmd:"" help:"Read, comment on and move Linear issues as the identity for their team."`
	Recall    recallCmd    `cmd:"" help:"Search past comments, session journals and learnings. Run it on a failure and before asking a human."`
	Whoami    whoamiCmd    `cmd:"" help:"Show the caller's scope."`
	Version   versionCmd   `cmd:"" help:"Print version information."`
	Docs      docsCmd      `cmd:"" help:"Print links to the docs for this binary (packs, readme). Needs no root."`
}

// Main parses args, runs the selected command and returns a process exit code.
func Main(args []string) int {
	if len(args) > 0 && args[0] == "gh" {
		// Before kong: gh's own flags must reach gh unparsed (#349).
		g := &Globals{Root: os.Getenv("JOBSHOP_ROOT"), Socket: os.Getenv("JOBSHOP_SOCKET")}
		return runGH(g, args[1:], os.Stdin, os.Stdout, os.Stderr)
	}
	var r root
	// kong calls Exit after --help; a no-op would fall through and run the
	// command, so record the code and stop before Run.
	exited := -1
	parser, err := kong.New(&r,
		kong.Name("jobshop"),
		kong.Description("Multi-agent work orchestration."),
		kong.UsageOnError(),
		kong.Exit(func(code int) { exited = code }),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	ctx, err := parser.Parse(args)
	if exited >= 0 {
		return exited
	}
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

type versionCmd struct {
	Server bool `help:"Print the running controller's version instead of this binary's. A session's own binary is fixed when it starts, so deploy checks use this (#180)."`
}

func (c versionCmd) Run(g *Globals) error {
	if c.Server {
		var r rpc.VersionResult
		if err := g.call(rpc.MethodVersion, nil, &r); err != nil {
			return err
		}
		_, err := fmt.Fprintf(stdout, "jobshop %s (%s)\n", r.Version, r.Commit)
		return err
	}
	bi, _ := debug.ReadBuildInfo()
	v, cm := versionInfo(version, commit, bi)
	_, err := fmt.Fprintf(stdout, "jobshop %s (%s)\n", v, cm)
	return err
}

// versionInfo fills in version and commit from Go build info when -ldflags
// left the defaults, so plain `go build` / `go install` builds identify
// themselves. Values set with -ldflags (the Nix build) win.
func versionInfo(v, c string, bi *debug.BuildInfo) (string, string) {
	if bi == nil {
		return v, c
	}
	if v == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		v = bi.Main.Version
	}
	if c == "unknown" {
		var rev string
		var dirty bool
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.modified":
				dirty = s.Value == "true"
			}
		}
		if rev != "" {
			c = rev[:min(12, len(rev))]
			if dirty {
				c += "-dirty"
			}
		}
	}
	return v, c
}

type sessionHostCmd struct {
	Dir string `arg:"" help:"Session directory containing spec.json." type:"path"`
}

func (c sessionHostCmd) Run() error { return host.Main(c.Dir, hostspec.HostOptions(c.Dir)...) }
