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
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	host "github.com/elee1766/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
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
	bi, _ := debug.ReadBuildInfo()
	v, cm := versionInfo(version, commit, bi)
	return control.Run(ctx, r, control.Options{Tick: c.Tick, Log: log, Reload: reload, Version: v, Commit: cm})
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
	// Key with pool:<name> sends by pool and key (session.send), starting
	// the chat pool's session for the key, e.g. improve:<slug> (#249).
	Key string `help:"With pool:<chat pool>: the session key; starts the session if none is open (admin)."`
}

func (c sendCmd) Run(g *Globals) error {
	text, err := readBody(c.Text)
	if err != nil {
		return err
	}
	if c.Key != "" {
		pool, ok := strings.CutPrefix(c.To, "pool:")
		if !ok {
			return errors.New("--key needs pool:<name> as the recipient")
		}
		var res rpc.SessionSendResult
		if err := g.call(rpc.MethodSessionSend, rpc.SessionSendParams{Pool: pool, Key: c.Key, Text: text, Interrupt: c.Interrupt}, &res); err != nil {
			return err
		}
		fmt.Println(res.Session)
		return nil
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
	Show  configShowCmd  `cmd:"" help:"Print the merged config the controller runs with, and which pack or file set each agent, pool and trigger."`
	Check configCheckCmd `cmd:"" help:"Load and validate a config file with its packs, without a controller. Exit 1 on error."`
}

type configShowCmd struct{}

func (configShowCmd) Run(g *Globals) error { return g.call(rpc.MethodConfigShow, nil, nil) }

// configCheckCmd is what self-deploy runs from the new build before it
// switches (#261): a pack var the host config no longer satisfies fails
// here instead of after the switch.
type configCheckCmd struct {
	File    string   `help:"Config file. Default: the root's jobshop.yaml." type:"path"`
	Replace []string `help:"OLD=NEW: read packs whose path starts with OLD from NEW instead, e.g. the new commit's checkout." placeholder:"OLD=NEW"`
}

func (c configCheckCmd) Run(g *Globals) error {
	file := c.File
	if file == "" {
		r, err := g.root()
		if err != nil {
			return err
		}
		file = r.Config()
	}
	rewrite, err := pathRewrite(c.Replace)
	if err != nil {
		return err
	}
	cfg, err := config.LoadFileWith(file, rewrite)
	if err != nil {
		return err
	}
	for _, w := range cfg.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	fmt.Printf("ok: %d pools, %d agents, %d triggers, %d workflows\n", len(cfg.Pools), len(cfg.Agents), len(cfg.Triggers), len(cfg.Workflows))
	return nil
}

// pathRewrite builds a prefix rewrite from OLD=NEW pairs.
func pathRewrite(pairs []string) (func(string) string, error) {
	type rep struct{ old, new string }
	reps := make([]rep, 0, len(pairs))
	for _, p := range pairs {
		o, n, ok := strings.Cut(p, "=")
		if !ok || o == "" || n == "" {
			return nil, fmt.Errorf("--replace %q: want OLD=NEW", p)
		}
		reps = append(reps, rep{filepath.Clean(o), filepath.Clean(n)})
	}
	return func(path string) string {
		c := filepath.Clean(path)
		for _, r := range reps {
			if c == r.old {
				return r.new
			}
			if rest, ok := strings.CutPrefix(c, r.old+string(filepath.Separator)); ok {
				return filepath.Join(r.new, rest)
			}
		}
		return path
	}, nil
}

type poolCmd struct {
	List         poolListCmd         `cmd:"" help:"Show each pool's min, max, desired, open, working, idle and cordoned sessions, plus quarantined and paused pools."`
	Scale        poolScaleCmd        `cmd:"" help:"Set a pool's desired session count, or 'auto' to follow ready work again. Persisted (admin or pool-scale grant)."`
	Unquarantine poolUnquarantineCmd `cmd:"" help:"Let a quarantined or paused pool start sessions again."`
}

type poolScaleCmd struct {
	Pool   string `arg:"" help:"Pool name."`
	Target string `arg:"" help:"Session count, or auto."`
}

func (c poolScaleCmd) Run(g *Globals) error {
	p := rpc.PoolScaleParams{Pool: c.Pool}
	if c.Target == "auto" {
		p.Auto = true
	} else {
		n, err := strconv.Atoi(c.Target)
		if err != nil || n < 0 {
			return fmt.Errorf("target %q: want a count >= 0 or auto", c.Target)
		}
		p.Target = &n
	}
	return g.call(rpc.MethodPoolScale, p, nil)
}

type poolListCmd struct{}

func (poolListCmd) Run(g *Globals) error { return g.call(rpc.MethodPoolList, nil, nil) }

type poolUnquarantineCmd struct {
	Pool string `arg:"" help:"Pool name."`
}

func (c poolUnquarantineCmd) Run(g *Globals) error {
	return g.call(rpc.MethodPoolUnquarantine, rpc.PoolParams{Pool: c.Pool}, nil)
}

type triggerListCmd struct{}

func (triggerListCmd) Run(g *Globals) error { return g.call(rpc.MethodTriggerList, nil, nil) }

type recallCmd struct {
	Query   []string `arg:"" help:"Words to search for."`
	Limit   int      `short:"n" help:"Most hits to return (default 10)."`
	Scope   string   `enum:",comment,journal,learning" default:"" help:"Only one source kind: comment, journal or learning."`
	Ticket  string   `short:"t" help:"Only hits on this ticket."`
	Project string   `help:"Project (admin only; sessions search their own)."`
}

func (c recallCmd) Run(g *Globals) error {
	return g.call(rpc.MethodRecall, rpc.RecallParams{
		Query: strings.Join(c.Query, " "), Limit: c.Limit, Kind: c.Scope, TicketID: c.Ticket, Project: c.Project,
	}, nil)
}

type whoamiCmd struct{}

func (whoamiCmd) Run(g *Globals) error { return g.call(rpc.MethodWhoami, nil, nil) }

type ticketCmd struct {
	Create ticketCreateCmd `cmd:"" help:"Create a ticket."`
	Show   ticketShowCmd   `cmd:"" help:"Show a ticket with edges, comments and history."`
	Trace  ticketTraceCmd  `cmd:"" help:"Show the chain from a ticket: its trigger, holders, GitHub links and the tickets it led to."`
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

type ticketTraceCmd struct {
	ID    string `arg:"" optional:"" help:"Ticket id (default: the session's current ticket)."`
	Depth int    `short:"d" help:"Levels of children to follow (default 5)."`
}

func (c ticketTraceCmd) Run(g *Globals) error {
	return g.call(rpc.MethodTicketTrace, rpc.TraceParams{ID: c.ID, Depth: c.Depth}, nil)
}

type ticketListCmd struct {
	Project string `help:"Project."`
	Kind    string `help:"Kind."`
	Status  string `short:"s" enum:",open,in_progress,closed" default:"" help:"Status."`
	Pool    string `help:"Pool."`
	Holder  string `help:"Holder session id."`
	Parent  string `help:"Parent ticket id."`
	Label   string `short:"l" help:"Label."`
	Report  string `help:"Last worker report state (progress, waiting, ready, blocked, failed, dropped)."`
	Limit   int    `short:"n" help:"Max results (default 500)."`

	Orphaned bool `help:"Only open, unclaimed tickets with no pool or a pool not in the config."`
}

func (c ticketListCmd) params() rpc.ListParams {
	return rpc.ListParams{
		Project: c.Project, Kind: store.Kind(c.Kind), Status: store.Status(c.Status), Pool: c.Pool,
		Holder: c.Holder, ParentID: c.Parent, Label: c.Label, Report: c.Report, Limit: c.Limit, Orphaned: c.Orphaned,
	}
}

func (c ticketListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodTicketList, c.params(), nil)
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
	Force  bool   `help:"Assign even if the session is cordoned."`
}

func (c assignCmd) Run(g *Globals) error {
	p := rpc.UpdateParams{ID: c.ID, Holder: &c.Holder, Force: c.Force}
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
	Until  string `help:"Defer the ticket: a duration (2h) or RFC3339 time. Not counted as a step failure."`
}

func (c releaseCmd) Run(g *Globals) error {
	p := rpc.ReleaseParams{ID: c.ID, Reason: c.Reason}
	if c.Until != "" {
		t, err := parseUntil(c.Until, time.Now())
		if err != nil {
			return err
		}
		p.Until = &t
	}
	return g.call(rpc.MethodRelease, p, nil)
}

// parseUntil reads a duration from now or an RFC3339 time.
func parseUntil(s string, now time.Time) (time.Time, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return now.Add(d), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("--until %q: want a duration or RFC3339 time", s)
	}
	return t, nil
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
	New      sessionNewCmd      `cmd:"" help:"Register a session and open its CLI socket (admin). Phase 2 starts agents automatically."`
	List     sessionListCmd     `cmd:"" help:"List sessions with live host state, plus orphaned hosts (admin)."`
	Reap     sessionReapCmd     `cmd:"" help:"Stop hosts with no open session; with --closed also delete old closed session dirs (admin)."`
	Restart  sessionRestartCmd  `cmd:"" help:"Restart a session's host with a fresh spec, keeping its conversation and claims (admin)."`
	Attach   sessionAttachCmd   `cmd:"" help:"Follow a session's conversation; --rw also sends stdin lines as messages. Ctrl-C detaches."`
	Prompt   sessionPromptCmd   `cmd:"" help:"Print the prompt a session's next turn would get, each part labeled by source (admin)."`
	Cordon   sessionCordonCmd   `cmd:"" help:"Take a session out of rotation: it finishes its ticket and claims nothing new, but stays open (admin or session-cordon grant)."`
	Uncordon sessionUncordonCmd `cmd:"" help:"Return a cordoned session to rotation."`
	Drain    sessionDrainCmd    `cmd:"" help:"Cordon a session and end it once it holds nothing."`
}

type sessionCordonCmd struct {
	ID string `arg:"" help:"Session id."`
}

func (c sessionCordonCmd) Run(g *Globals) error {
	return g.call(rpc.MethodSessionCordon, rpc.SessionCordonParams{ID: c.ID}, nil)
}

type sessionUncordonCmd struct {
	ID string `arg:"" help:"Session id."`
}

func (c sessionUncordonCmd) Run(g *Globals) error {
	return g.call(rpc.MethodSessionCordon, rpc.SessionCordonParams{ID: c.ID, Uncordon: true}, nil)
}

type sessionDrainCmd struct {
	ID string `arg:"" help:"Session id."`
}

func (c sessionDrainCmd) Run(g *Globals) error {
	return g.call(rpc.MethodSessionCordon, rpc.SessionCordonParams{ID: c.ID, Drain: true}, nil)
}

// sessionPromptCmd shows pool prompt, prompt rules and ticket as the
// session's next turn would get them. Rule commands run.
type sessionPromptCmd struct {
	ID   string `arg:"" help:"Session id."`
	JSON bool   `name:"json" help:"Print the parts as JSON."`
}

func (c sessionPromptCmd) Run(g *Globals) error {
	if c.JSON {
		return g.call(rpc.MethodSessionPrompt, rpc.SessionPromptParams{ID: c.ID}, nil)
	}
	var res rpc.SessionPromptResult
	if err := g.call(rpc.MethodSessionPrompt, rpc.SessionPromptParams{ID: c.ID}, &res); err != nil {
		return err
	}
	_, err := fmt.Print(res.Text)
	return err
}

type sessionAttachCmd struct {
	ID     string `arg:"" help:"Session id."`
	Since  *int64 `help:"Start after this journal seq (default: the last 50 frames)."`
	JSON   bool   `name:"json" help:"Print raw journal frames."`
	Follow bool   `default:"true" negatable:"" help:"Keep following new frames."`
	RW     bool   `name:"rw" help:"Send each stdin line as a message (admin; the pool must set attach: rw). A line starting with ! interrupts the turn."`
}

// Run reads the session's journal file directly rather than asking the
// controller to stream it: the host writes the journal even while the
// controller is down, and reading a file needs no new RPC (plan.md phase 3).
func (c sessionAttachCmd) Run(g *Globals) error {
	if c.RW {
		return c.runRW(g)
	}
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

type reportCmd struct {
	Incidents reportIncidentsCmd `cmd:"" help:"Group controller WARN/ERROR lines, crashed or stuck sessions and failed run steps."`

	// Worker reports on the held ticket (#186).
	Progress reportProgressCmd `cmd:"" help:"Report progress on your ticket."`
	Waiting  reportWaitingCmd  `cmd:"" help:"Report waiting (on a question, a time, another change)."`
	Ready    reportReadyCmd    `cmd:"" help:"Report the work ready for the next step."`
	Blocked  reportBlockedCmd  `cmd:"" help:"Report blocked; labels the ticket blocked for stall-check."`
	Failed   reportFailedCmd   `cmd:"" help:"Report failure and why (then close -r failed)."`
	Dropped  reportDroppedCmd  `cmd:"" help:"Report the work dropped and why (then close)."`
}

type reportIncidentsCmd struct {
	Since time.Duration `default:"24h" help:"How far back to look. Log lines go back to the controller start at most."`
}

func (c reportIncidentsCmd) Run(g *Globals) error {
	return g.call(rpc.MethodReportIncidents, rpc.IncidentsParams{Since: c.Since}, nil)
}

type attentionCmd struct{}

func (attentionCmd) Run(g *Globals) error { return g.call(rpc.MethodAttention, nil, nil) }

type activityCmd struct {
	Since  time.Duration `help:"Only items newer than this, e.g. 1h."`
	Source string        `enum:",github,linear" default:"" help:"github or linear."`
	Repo   string        `help:"owner/name, or a Linear team key."`
	Kind   string        `help:"push, pr, merge, review, review_comment, comment, issue, branch, release, ci, state, assignee."`
	Actor  string        `help:"Login or Linear name."`
	Bots   string        `enum:",all,humans,bots" default:"" help:"humans, bots or all (default)."`
	Ticket string        `short:"t" help:"Only items linked to this run or ticket."`
	Limit  int           `default:"50" help:"Most items to print, newest first."`
}

func (c activityCmd) Run(g *Globals) error {
	return g.call(rpc.MethodActivityList, rpc.ActivityListParams{
		Since: c.Since, Source: c.Source, Repo: c.Repo, Kind: c.Kind, Actor: c.Actor,
		Bots: c.Bots, Ticket: c.Ticket, Limit: c.Limit,
	}, nil)
}

type guardCmd struct {
	Action string `arg:"" help:"Action about to happen: push, pr-create, merge, comment, message, or any name a hook knows."`
	Target string `help:"What it acts on, such as a PR url or branch."`
	Detail string `help:"Free text for the hook, such as the message body."`
}

// Run prints the answer and exits 0 only on allow, so an agent can write
// `jobshop guard push --target x && git push`.
func (c guardCmd) Run(g *Globals) error {
	var r rpc.GuardResult
	if err := g.call(rpc.MethodGuard, rpc.GuardParams{Action: c.Action, Target: c.Target, Detail: c.Detail}, &r); err != nil {
		return err
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if err := printJSON(b); err != nil {
		return err
	}
	if r.Decision != rpc.GuardAllow {
		return fmt.Errorf("guard %s: %s: %s", c.Action, r.Decision, r.Reason)
	}
	return nil
}

type subscribeCmd struct {
	Thread string `arg:"" help:"owner/name#<n>, or an issue or pull request URL."`
	PR     bool   `help:"The thread is a pull request (also read reviews). Implied by a /pull/ URL."`
	ID     string `short:"t" name:"ticket" help:"Ticket id (default: the session's current ticket)."`
}

func (c subscribeCmd) Run(g *Globals) error {
	repo, n, pr, err := parseThread(c.Thread)
	if err != nil {
		return err
	}
	return g.call(rpc.MethodSubscribe, rpc.SubscribeParams{ID: c.ID, Repo: repo, Number: n, PR: pr || c.PR}, nil)
}

// parseThread reads owner/name#n or a github.com issue or pull URL.
func parseThread(s string) (repo string, n int, pr bool, err error) {
	s = strings.TrimSpace(s)
	if rest, ok := strings.CutPrefix(s, "https://github.com/"); ok {
		parts := strings.Split(strings.Trim(rest, "/"), "/")
		if len(parts) >= 4 && (parts[2] == "issues" || parts[2] == "pull") {
			n, err = strconv.Atoi(parts[3])
			if err == nil && n > 0 {
				return parts[0] + "/" + parts[1], n, parts[2] == "pull", nil
			}
		}
		return "", 0, false, fmt.Errorf("not an issue or pull request URL: %s", s)
	}
	repo, num, ok := strings.Cut(s, "#")
	if n, err = strconv.Atoi(num); !ok || err != nil || n <= 0 || !strings.Contains(repo, "/") {
		return "", 0, false, fmt.Errorf("thread must be owner/name#<n> or a URL: %s", s)
	}
	return repo, n, false, nil
}
