package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/host"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// stdout is swapped by tests.
var stdout io.Writer = os.Stdout

// call connects to the right socket and runs one request, printing the
// result as JSON.
func (g *Globals) call(method string, params, out any) error {
	sock := g.Socket
	if sock == "" {
		r, err := g.root()
		if err != nil {
			return err
		}
		sock = r.AdminSocket()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// An agent's calls join its session's trace (set by the controller).
	ctx = telemetry.WithTraceParent(ctx, os.Getenv("JOBSHOP_TRACEPARENT"))
	c, err := rpc.Dial(ctx, sock)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var res json.RawMessage
	if err := c.Call(ctx, method, params, &res); err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(res, out)
	}
	return printJSON(res)
}

func printJSON(raw json.RawMessage) error {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// exitCode maps errors to stable exit codes so agents can branch on them.
func exitCode(err error) int {
	switch {
	case errors.Is(err, store.ErrNoWork):
		return 3
	case errors.Is(err, store.ErrNotHolder):
		return 4
	case errors.Is(err, store.ErrNotFound):
		return 5
	case errors.Is(err, store.ErrConflict):
		return 6
	}
	return 1
}

type initCmd struct {
	Dir     string `arg:"" optional:"" default:"." help:"Directory to make a Jobshop root." type:"path"`
	Project string `help:"First project name." default:"js"`
}

func (c initCmd) Run() error {
	if err := os.MkdirAll(c.Dir, 0o700); err != nil {
		return err
	}
	r := config.Root{Dir: c.Dir}
	body := fmt.Sprintf("default_project: %s\nprojects:\n  %s: {}\n", c.Project, c.Project)
	if _, err := config.Parse([]byte(body)); err != nil {
		return err
	}
	f, err := os.OpenFile(r.Config(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(body)
	if err == nil {
		fmt.Fprintln(os.Stderr, "created", r.Config())
	}
	return err
}

type serveCmd struct {
	Tick time.Duration `help:"Lease sweep interval." default:"30s"`
}

func (c serveCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	// SIGHUP reloads the config, like `jobshop reload`.
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	defer signal.Stop(hup)
	reload := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-hup:
				select {
				case reload <- struct{}{}:
				default: // one already pending reads the latest file
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return control.Run(ctx, r, control.Options{Tick: c.Tick, Log: log, Reload: reload})
}

type reloadCmd struct{}

func (reloadCmd) Run(g *Globals) error {
	return g.call(rpc.MethodReload, struct{}{}, nil)
}

type drainCmd struct {
	Pool string `help:"Drain only this pool (default: every pool and session)."`
}

func (c drainCmd) Run(g *Globals) error {
	return g.call(rpc.MethodDrain, rpc.DrainParams{Pool: c.Pool}, nil)
}

type undrainCmd struct {
	Pool string `help:"Undrain only this pool (default: clear every drain)."`
}

func (c undrainCmd) Run(g *Globals) error {
	return g.call(rpc.MethodDrain, rpc.DrainParams{Pool: c.Pool, Undrain: true}, nil)
}

type sendCmd struct {
	To        string `arg:"" help:"Session id, session key, or pool:<name>."`
	Text      string `arg:"" help:"Message text. Use - to read stdin."`
	Interrupt bool   `help:"Cancel the recipient's running turn and deliver now."`
	Project   string `help:"Project (admin only; sessions use their own)."`
}

func (c sendCmd) Run(g *Globals) error {
	text, err := readBody(c.Text)
	if err != nil {
		return err
	}
	fields, err := json.Marshal(store.MessageFields{To: c.To, Interrupt: c.Interrupt})
	if err != nil {
		return err
	}
	title, body, _ := strings.Cut(text, "\n")
	return g.call(rpc.MethodTicketCreate, rpc.CreateParams{
		Project: c.Project, Kind: store.KindMessage, Title: title, Body: body, Fields: fields,
	}, nil)
}

type shutdownCmd struct {
	All  bool          `help:"Also end every session (stops agents and hosts)."`
	Wait time.Duration `default:"2m" help:"How long to wait for the controller to exit."`
}

func (c shutdownCmd) Run(g *Globals) error {
	if err := g.call(rpc.MethodShutdown, rpc.ShutdownParams{All: c.All}, &struct{}{}); err != nil {
		return err
	}
	r, err := g.root()
	if err != nil {
		return nil // talked to an explicit socket; nothing to wait on
	}
	deadline := time.Now().Add(c.Wait)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(r.AdminSocket()); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("controller still running after %s", c.Wait)
}

type triggerCmd struct {
	List triggerListCmd `cmd:"" help:"Show each trigger: whether it holds now, last fired, open ticket, next cron time."`
}

type configCmd struct {
	Show configShowCmd `cmd:"" help:"Print the merged config the controller runs with, and which pack or file set each agent, pool and trigger."`
}

type configShowCmd struct{}

func (configShowCmd) Run(g *Globals) error { return g.call(rpc.MethodConfigShow, nil, nil) }

type triggerListCmd struct{}

func (triggerListCmd) Run(g *Globals) error { return g.call(rpc.MethodTriggerList, nil, nil) }

type whoamiCmd struct{}

func (whoamiCmd) Run(g *Globals) error { return g.call(rpc.MethodWhoami, nil, nil) }

type ticketCmd struct {
	Create ticketCreateCmd `cmd:"" help:"Create a ticket."`
	Show   ticketShowCmd   `cmd:"" help:"Show a ticket with edges, comments and history."`
	List   ticketListCmd   `cmd:"" help:"List tickets."`
	Update ticketUpdateCmd `cmd:"" help:"Change a ticket."`
	Close  closeCmd        `cmd:"" help:"Close a ticket."`
	Reopen ticketReopenCmd `cmd:"" help:"Reopen a closed ticket (admin)."`
}

type ticketCreateCmd struct {
	Title    string   `arg:"" help:"Title."`
	Body     string   `short:"b" help:"Body. Use - to read stdin."`
	Kind     string   `default:"task" enum:"task,message" help:"Kind."`
	Project  string   `help:"Project (admin only; sessions use their own)."`
	Priority *int     `short:"p" help:"Priority, lower runs first (default 2)."`
	Pool     string   `help:"Pool to route to."`
	Assign   string   `help:"Pre-assign to a session (admin)."`
	Parent   string   `help:"Parent ticket id."`
	Label    []string `short:"l" help:"Labels."`
	Fields   string   `help:"Kind fields as JSON, e.g. '{\"to\":\"pool:dev\"}' for a message."`
	Meta     string   `help:"Pack data as JSON."`
}

func readBody(s string) (string, error) {
	if s != "-" {
		return s, nil
	}
	b, err := io.ReadAll(os.Stdin)
	return strings.TrimRight(string(b), "\n"), err
}

func rawJSON(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}

func (c ticketCreateCmd) Run(g *Globals) error {
	body, err := readBody(c.Body)
	if err != nil {
		return err
	}
	return g.call(rpc.MethodTicketCreate, rpc.CreateParams{
		Project: c.Project, Kind: store.Kind(c.Kind), Title: c.Title, Body: body,
		Priority: c.Priority, Pool: c.Pool, Holder: c.Assign, ParentID: c.Parent,
		Labels: c.Label, Fields: rawJSON(c.Fields), Meta: rawJSON(c.Meta),
	}, nil)
}

type ticketShowCmd struct {
	ID string `arg:"" optional:"" help:"Ticket id (default: the session's current ticket)."`
}

func (c ticketShowCmd) Run(g *Globals) error {
	return g.call(rpc.MethodTicketShow, rpc.IDParams{ID: c.ID}, nil)
}

type ticketListCmd struct {
	Project string `help:"Project."`
	Kind    string `help:"Kind."`
	Status  string `short:"s" enum:",open,in_progress,closed" default:"" help:"Status."`
	Pool    string `help:"Pool."`
	Holder  string `help:"Holder session id."`
	Parent  string `help:"Parent ticket id."`
	Label   string `short:"l" help:"Label."`
	Limit   int    `short:"n" help:"Max results (default 500)."`
}

func (c ticketListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodTicketList, rpc.ListParams{
		Project: c.Project, Kind: store.Kind(c.Kind), Status: store.Status(c.Status), Pool: c.Pool,
		Holder: c.Holder, ParentID: c.Parent, Label: c.Label, Limit: c.Limit,
	}, nil)
}

type ticketUpdateCmd struct {
	ID         string         `arg:"" help:"Ticket id."`
	Title      *string        `help:"New title."`
	Body       *string        `short:"b" help:"New body. Use - to read stdin."`
	Priority   *int           `short:"p" help:"New priority."`
	Pool       *string        `help:"New pool."`
	Defer      *time.Duration `name:"defer" help:"Hide from ready until now+duration (0 clears)."`
	Fields     string         `help:"Replace kind fields (JSON)."`
	Meta       string         `help:"Replace pack data (JSON)."`
	IfRevision int64          `help:"Fail unless the ticket is at this revision."`
}

func (c ticketUpdateCmd) Run(g *Globals) error {
	p := rpc.UpdateParams{
		ID: c.ID, Title: c.Title, Priority: c.Priority, Pool: c.Pool,
		Fields: rawJSON(c.Fields), Meta: rawJSON(c.Meta), IfRevision: c.IfRevision,
	}
	if c.Body != nil {
		b, err := readBody(*c.Body)
		if err != nil {
			return err
		}
		p.Body = &b
	}
	if c.Defer != nil {
		var t time.Time
		if *c.Defer > 0 {
			t = time.Now().Add(*c.Defer)
		}
		p.DeferUntil = &t
	}
	return g.call(rpc.MethodTicketUpdate, p, nil)
}

type closeCmd struct {
	ID         string `arg:"" optional:"" help:"Ticket id (default: the session's current ticket)."`
	Resolution string `short:"r" default:"done" help:"Resolution: done, failed, cancelled, ..."`
}

func (c closeCmd) Run(g *Globals) error {
	return g.call(rpc.MethodTicketClose, rpc.CloseParams{ID: c.ID, Resolution: c.Resolution}, nil)
}

type ticketReopenCmd struct {
	ID string `arg:"" help:"Ticket id."`
}

func (c ticketReopenCmd) Run(g *Globals) error {
	return g.call(rpc.MethodTicketReopen, rpc.IDParams{ID: c.ID}, nil)
}

type commentCmd struct {
	Body string `arg:"" help:"Comment text. Use - to read stdin."`
	ID   string `short:"t" name:"ticket" help:"Ticket id (default: the session's current ticket)."`
}

func (c commentCmd) Run(g *Globals) error {
	body, err := readBody(c.Body)
	if err != nil {
		return err
	}
	return g.call(rpc.MethodCommentAdd, rpc.CommentParams{ID: c.ID, Body: body}, nil)
}

type assignCmd struct {
	ID     string `arg:"" help:"Ticket id."`
	Holder string `arg:"" optional:"" help:"Session id to pre-assign; omit to unassign."`
	Pool   string `help:"Also route to this pool."`
}

func (c assignCmd) Run(g *Globals) error {
	p := rpc.UpdateParams{ID: c.ID, Holder: &c.Holder}
	if c.Pool != "" {
		p.Pool = &c.Pool
	}
	return g.call(rpc.MethodTicketUpdate, p, nil)
}

type claimCmd struct {
	ID      string `arg:"" optional:"" help:"Ticket id. Omit with --next."`
	Next    bool   `help:"Claim the next ticket: own in-progress work, then assigned, then the pool queue."`
	Project string `help:"Project (admin)."`
	Pool    string `help:"Pool (default: the session's pool)."`
	Holder  string `help:"Holder (admin only)."`
}

func (c claimCmd) Run(g *Globals) error {
	if (c.ID == "") == !c.Next {
		return errors.New("pass a ticket id or --next")
	}
	return g.call(rpc.MethodClaim, rpc.ClaimParams{ID: c.ID, Project: c.Project, Pool: c.Pool, Holder: c.Holder}, nil)
}

type heartbeatCmd struct {
	ID string `arg:"" optional:"" help:"Ticket id (default: the session's current ticket)."`
}

func (c heartbeatCmd) Run(g *Globals) error {
	return g.call(rpc.MethodHeartbeat, rpc.IDParams{ID: c.ID}, nil)
}

type releaseCmd struct {
	ID     string `arg:"" optional:"" help:"Ticket id (default: the session's current ticket)."`
	Reason string `short:"r" help:"Why the ticket is handed back."`
}

func (c releaseCmd) Run(g *Globals) error {
	return g.call(rpc.MethodRelease, rpc.ReleaseParams{ID: c.ID, Reason: c.Reason}, nil)
}

type readyCmd struct {
	Project string `help:"Project."`
	Pool    string `help:"Pool."`
	Limit   int    `short:"n" help:"Max results."`
}

func (c readyCmd) Run(g *Globals) error {
	return g.call(rpc.MethodReady, rpc.ReadyParams{Project: c.Project, Pool: c.Pool, Limit: c.Limit}, nil)
}

type edgeCmd struct {
	Add edgeArgs `cmd:"" help:"Add an edge: FROM is blocked by / waits for / tracks TO."`
	Rm  edgeRm   `cmd:"" help:"Remove an edge."`
}

type edgeArgs struct {
	From string `arg:"" help:"Dependent ticket."`
	To   string `arg:"" help:"Ticket it depends on."`
	Kind string `default:"blocks" enum:"blocks,waits_for,tracks" help:"Edge kind."`
}

func (c edgeArgs) Run(g *Globals) error {
	return g.call(rpc.MethodEdgeAdd, rpc.EdgeParams{From: c.From, To: c.To, Kind: store.EdgeKind(c.Kind)}, nil)
}

type edgeRm struct{ edgeArgs }

func (c edgeRm) Run(g *Globals) error {
	return g.call(rpc.MethodEdgeRemove, rpc.EdgeParams{From: c.From, To: c.To, Kind: store.EdgeKind(c.Kind)}, nil)
}

type labelCmd struct {
	Add labelArgs `cmd:"" help:"Add a label."`
	Rm  labelRm   `cmd:"" help:"Remove a label."`
}

type labelArgs struct {
	ID    string `arg:"" help:"Ticket id."`
	Label string `arg:"" help:"Label."`
}

func (c labelArgs) Run(g *Globals) error {
	return g.call(rpc.MethodLabelAdd, rpc.LabelParams{ID: c.ID, Label: c.Label}, nil)
}

type labelRm struct{ labelArgs }

func (c labelRm) Run(g *Globals) error {
	return g.call(rpc.MethodLabelRemove, rpc.LabelParams{ID: c.ID, Label: c.Label}, nil)
}

type sessionCmd struct {
	New     sessionNewCmd     `cmd:"" help:"Register a session and open its CLI socket (admin). Phase 2 starts agents automatically."`
	List    sessionListCmd    `cmd:"" help:"List sessions with live host state, plus orphaned hosts (admin)."`
	Reap    sessionReapCmd    `cmd:"" help:"Stop hosts with no open session; with --closed also delete old closed session dirs (admin)."`
	Restart sessionRestartCmd `cmd:"" help:"Restart a session's host with a fresh spec, keeping its conversation and claims (admin)."`
	Attach  sessionAttachCmd  `cmd:"" help:"Follow a session's conversation, read-only. Ctrl-C detaches."`
}

type sessionAttachCmd struct {
	ID     string `arg:"" help:"Session id."`
	Since  *int64 `help:"Start after this journal seq (default: the last 50 frames)."`
	JSON   bool   `name:"json" help:"Print raw journal frames."`
	Follow bool   `default:"true" negatable:"" help:"Keep following new frames."`
}

// Run reads the session's journal file directly rather than asking the
// controller to stream it: the host writes the journal even while the
// controller is down, and reading a file needs no new RPC (plan.md phase 3).
func (c sessionAttachCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	path := filepath.Join(r.SessionDir(c.ID), host.JournalFile)
	since := int64(-50)
	if c.Since != nil {
		since = *c.Since
	}
	frames, off, err := host.ReadJournal(path, since)
	if err != nil {
		return err
	}
	render := &host.Renderer{W: stdout}
	show := func(f host.Frame) {
		if c.JSON {
			_, _ = stdout.Write(host.RawFrame(f))
			return
		}
		render.Render(f)
	}
	for _, f := range frames {
		show(f)
	}
	if !c.Follow {
		return nil
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return host.FollowJournal(ctx, path, off, show)
}

type sessionListCmd struct {
	All   bool          `help:"Include closed sessions."`
	State string        `help:"Only closed sessions with this resolution (e.g. crashed)."`
	Since time.Duration `help:"Only sessions closed within this long (e.g. 24h)."`
}

func (c sessionListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodSessionList, rpc.SessionListParams{All: c.All, State: c.State, Since: c.Since}, nil)
}

// sessionRestartCmd: a busy session restarts after its current turn; the
// conversation is resumed and held tickets stay claimed.
type sessionRestartCmd struct {
	ID  string `arg:"" help:"Session id."`
	Now bool   `help:"Cancel the running turn instead of waiting for it to end."`
}

func (c sessionRestartCmd) Run(g *Globals) error {
	return g.call(rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: c.ID, Now: c.Now}, nil)
}

type sessionReapCmd struct {
	Closed    bool          `help:"Also delete directories (and worktrees) of closed sessions."`
	OlderThan time.Duration `default:"24h" help:"With --closed, only sessions closed longer ago than this."`
	DryRun    bool          `help:"Report what would be done without doing it."`
}

func (c sessionReapCmd) Run(g *Globals) error {
	return g.call(rpc.MethodSessionReap, rpc.SessionReapParams{
		Closed: c.Closed, OlderThan: c.OlderThan, DryRun: c.DryRun,
	}, nil)
}

type sessionNewCmd struct {
	Agent   string `required:"" help:"Agent name."`
	Pool    string `help:"Pool the session works."`
	Project string `help:"Project."`
	Key     string `help:"External key, e.g. matrix:<room>; unique while open."`
	Title   string `help:"Title."`
}

func (c sessionNewCmd) Run(g *Globals) error {
	return g.call(rpc.MethodSessionNew, rpc.SessionNewParams{
		Project: c.Project, Pool: c.Pool, Agent: c.Agent, Key: c.Key, Title: c.Title,
	}, nil)
}
