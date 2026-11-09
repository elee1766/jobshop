// Package linearsvc is the controller's Linear orchestration, split out of
// pkg/orchestration/control (JOB-142): the agent-session sweep, linear_issue_no_run
// intake, webhook handling, question mirroring on the issue, and the
// outage backoff every Linear poller shares.
//
// It reuses pkg/integrations/linear for every request and the controller's shared
// per-identity clients (Host.Client), so credentials and token minting stay
// in the controller. Every store write goes back through Host. Slow
// requests run outside the supervisor pass: the sweeps start them through
// Host.OffGo or Host.OffTick and a later pass takes the result.
package linearsvc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/orchestration/ingest"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Host is what the service needs from the controller.
type Host interface {
	Config() *config.Config
	Now() time.Time
	Log() *slog.Logger
	// Context is the controller's, ended at shutdown.
	Context() context.Context
	// Go runs fn on the controller's WaitGroup, so shutdown joins it.
	Go(fn func())

	// Client is identity name's shared Linear client. The controller owns
	// the credentials behind it.
	Client(name string) *linear.Client
	// ClientIn is the shared client of the identity covering key's team
	// under cfg.
	ClientIn(cfg *config.Config, key string) (*linear.Client, error)
	// CanMint reports whether identities can get tokens at all.
	CanMint() bool

	// OffGo runs fn outside the supervisor pass under timeout, one in
	// flight per key. It reports whether fn started.
	OffGo(key, poller string, timeout time.Duration, fn func(ctx context.Context)) bool
	// OffTick returns run's result for a pass: pending while a pass only
	// collects the call (the replay has the result), else the value or the
	// error run returned.
	OffTick(ctx context.Context, key, poller string, timeout time.Duration, run func(ctx context.Context) (any, error)) (val any, pending bool, err error)

	// List and Get read tickets.
	List(ctx context.Context, f store.Filter) ([]domain.Ticket, error)
	Get(ctx context.Context, id string) (domain.Ticket, error)
	// OpenNeedsHuman lists open needs-human tickets of kind: questions are
	// messages, escalations tasks.
	OpenNeedsHuman(ctx context.Context, kind domain.Kind) ([]domain.Ticket, error)
	// IsQuestion and IsEscalation classify a ticket for mirroring.
	IsQuestion(t domain.Ticket) bool
	IsEscalation(t domain.Ticket) bool
	// SetMeta replaces meta[pack] on a ticket.
	SetMeta(ctx context.Context, id, pack string, v any) error
	// Comment adds a controller comment to a ticket.
	Comment(ctx context.Context, id, text string) error
	// Keys is the write-key store of project, for idempotent Linear writes.
	Keys(project string) linear.KeyStore
	// AnswerQuestion answers an open question as the controller.
	AnswerQuestion(ctx context.Context, id, text, from string) error

	// GitHubIssueOpen reads a GitHub issue's state for a "Depends on" line.
	GitHubIssueOpen(ctx context.Context, cfg *config.Config, d linear.GitHubDep) (bool, error)

	// IngestSecret resolves an endpoint's signing secret.
	IngestSecret(e config.IngestEndpoint) (ingest.Secret, error)
	// Spool opens the ingest spool the webhook writes to.
	Spool(ctx context.Context) (*ingest.Spool, error)
	// WakeDrain wakes the inbox drain after a delivery is spooled.
	WakeDrain()
	// RunTriggers runs the trigger pass now, under the tick lock.
	RunTriggers(ctx context.Context)
	// PollThread polls the subscription thread of a Linear issue.
	PollThread(ctx context.Context, issue string)
}

// Service holds the Linear state the controller used to keep itself: the
// outage backoff, the intake cache and the session sweep's throttle.
type Service struct {
	host    Host
	timeout time.Duration

	health   health
	intake   intakeCache
	sessions sessionThrottle
}

// New makes a service on h.
func New(h Host) *Service { return &Service{host: h} }

func (s *Service) log() *slog.Logger { return s.host.Log() }

// RequestTimeout is the per-request bound for Linear calls made by pollers
// and wait passes.
const RequestTimeout = 15 * time.Second

// SetRequestTimeout replaces RequestTimeout; tests lower it. Call it before
// the service is in use.
func (s *Service) SetRequestTimeout(d time.Duration) { s.timeout = d }

func (s *Service) requestTimeout() time.Duration {
	if s.timeout > 0 {
		return s.timeout
	}
	return RequestTimeout
}

// PollerCtx bounds each request made with ctx by the request timeout.
func (s *Service) PollerCtx(ctx context.Context) context.Context {
	return apicall.WithRequestTimeout(ctx, s.requestTimeout())
}

// Do runs fn with identity name's shared client, each request bounded by
// the request timeout, through Call.
func (s *Service) Do(ctx context.Context, name string, fn func(context.Context, *linear.Client) error) error {
	return s.Call(ctx, name, func(ctx context.Context) error {
		return fn(s.PollerCtx(ctx), s.host.Client(name))
	})
}

// DoKey is Do as the identity covering key's team.
func (s *Service) DoKey(ctx context.Context, key string, fn func(context.Context, *linear.Client) error) error {
	team, ok := linear.Team(key)
	if !ok {
		return fmt.Errorf("bad Linear issue key %q", key)
	}
	cfg := s.host.Config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), team)
	if !ok {
		return fmt.Errorf("no identity covers team %s", team)
	}
	return s.Do(ctx, name, fn)
}
