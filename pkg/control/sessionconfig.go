package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	acp "github.com/coder/acp-go-sdk"
	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// resolutionConfigFailed ends a session whose agent refused a configured
// option (#304): a turn on the wrong model must not run silently.
const resolutionConfigFailed = "config_failed"

// advertised lists the config option ids in an ACP configOptions list and
// their current values.
func advertised(opts []acp.SessionConfigOption) ([]string, map[string]string) {
	ids := []string{}
	cur := map[string]string{}
	for _, o := range opts {
		switch {
		case o.Select != nil:
			ids = append(ids, string(o.Select.Id))
			cur[string(o.Select.Id)] = string(o.Select.CurrentValue)
		case o.Boolean != nil:
			ids = append(ids, string(o.Boolean.Id))
			cur[string(o.Boolean.Id)] = fmt.Sprint(o.Boolean.CurrentValue)
		}
	}
	slices.Sort(ids)
	return ids, cur
}

// noteAdvertised records what a session/new or session/load returned. The
// agent process is new, so every configured option is set again.
func noteAdvertised(f *store.SessionFields, opts []acp.SessionConfigOption) {
	ids, cur := advertised(opts)
	f.ConfigIDs = ids
	f.Options = nil
	f.Model, f.Effort = cur[config.OptionModel], cur[config.OptionEffort]
}

// wantOptions is what the session should run with now: the pool's and
// agent's options, then the session's override.
func (r *runner) wantOptions(f store.SessionFields) map[string]string {
	want := r.c.config().SessionOptions(r.pool, f.Agent)
	maps.Copy(want, f.Override)
	return want
}

// applyConfig sets every option whose wanted value differs from what the
// controller last set on the conversation. It runs only where no turn is
// running: after session/new or load, and between turns in work. An id the
// agent did not advertise is logged once per session and skipped. A set
// that fails twice ends the session config_failed.
func (r *runner) applyConfig(ctx context.Context, conn *acp.ClientSideConnection, acpSession string) error {
	t, err := r.c.store.Get(ctx, r.id)
	if err != nil {
		return err
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil {
		return err
	}
	want := r.wantOptions(f)
	for _, id := range slices.Sorted(maps.Keys(want)) {
		v := want[id]
		if f.Options[id] == v {
			continue
		}
		if !slices.Contains(f.ConfigIDs, id) {
			r.mu.Lock()
			if r.skipped == nil {
				r.skipped = map[string]bool{}
			}
			first := !r.skipped[id]
			r.skipped[id] = true
			r.mu.Unlock()
			if first {
				r.log.Warn("agent does not advertise config option, skipping", "option", id, "value", v, "advertised", f.ConfigIDs)
			}
			continue
		}
		req := acp.SetSessionConfigOptionRequest{ValueId: &acp.SetSessionConfigOptionValueId{
			SessionId: acp.SessionId(acpSession), ConfigId: acp.SessionConfigId(id), Value: acp.SessionConfigValueId(v),
		}}
		var res acp.SetSessionConfigOptionResponse
		for attempt := range 2 {
			res, err = conn.SetSessionConfigOption(ctx, req)
			if err == nil || ctx.Err() != nil {
				break
			}
			r.log.Warn("set config option failed", "option", id, "value", v, "attempt", attempt+1, "err", err)
		}
		if err != nil {
			if ctx.Err() != nil {
				return r.dropped(conn)
			}
			return errSessionEnd{resolution: resolutionConfigFailed, reason: fmt.Sprintf("set %s=%s: %s", id, v, agentMessage(err))}
		}
		old := f.Options[id]
		ids, cur := advertised(res.ConfigOptions)
		err = r.c.updateSessionFields(ctx, r.id, func(sf *store.SessionFields) {
			if sf.Options == nil {
				sf.Options = map[string]string{}
			}
			sf.Options[id] = v
			if len(ids) > 0 {
				sf.ConfigIDs = ids
				sf.Model, sf.Effort = cur[config.OptionModel], cur[config.OptionEffort]
			}
			f = *sf
		})
		if err != nil {
			return err
		}
		r.log.Info("session config option set", "option", id, "old", old, "new", v, "current", cur[id])
		r.lifecycle("config", attribute.String("jobshop.config.option", id),
			attribute.String("jobshop.config.old", old), attribute.String("jobshop.config.new", v))
		r.c.events.publish([]rpc.Event{{
			Kind: rpc.EventKindSession, Op: rpc.OpSessionConfig, ID: r.id, Pool: r.pool,
			Option: id, Old: old, New: v, Time: r.c.now(),
		}})
	}
	r.setModel(f.Model, f.Effort)
	return nil
}

// agentMessage is an ACP error's message and data, for the session's
// close comment.
func agentMessage(err error) string {
	var re *acp.RequestError
	if errors.As(err, &re) {
		if re.Data != nil {
			return strings.TrimSpace(fmt.Sprint(re.Message, " ", re.Data))
		}
		return re.Message
	}
	return err.Error()
}

func (r *runner) setModel(model, effort string) {
	r.mu.Lock()
	r.model, r.effort = model, effort
	r.mu.Unlock()
}

// updateSessionFields applies fn to a session's fields with a revision
// check, retrying on a concurrent change.
func (c *Controller) updateSessionFields(ctx context.Context, id string, fn func(*store.SessionFields)) error {
	for range 5 {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return err
		}
		f, err := store.DecodeFields[store.SessionFields](t)
		if err != nil {
			return err
		}
		fn(&f)
		b, err := json.Marshal(f)
		if err != nil {
			return err
		}
		_, err = c.store.Update(context.WithoutCancel(ctx), "controller", id, store.Update{Fields: b, IfRevision: t.Revision})
		if !errors.Is(err, store.ErrConflict) {
			return err
		}
	}
	return fmt.Errorf("%w: session %s kept changing", store.ErrConflict, id)
}

// sessionConfigRPC sets or resets a session's option override; the
// runner applies it between turns.
func (c *Controller) sessionConfigRPC(ctx context.Context, sc scope, p rpc.SessionConfigParams) (any, error) {
	if err := c.requireGrant(sc, "session-config", "session config"); err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != store.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, p.ID, t.Kind)
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: session %s", store.ErrClosed, p.ID)
	}
	if !p.Reset && len(p.Options) == 0 {
		return nil, fmt.Errorf("%w: give options (model=<value>) or reset", store.ErrInvalid)
	}
	for k, v := range p.Options {
		if k == "" || v == "" {
			return nil, fmt.Errorf("%w: option %q=%q: id and value are required", store.ErrInvalid, k, v)
		}
	}
	opts := map[string]string{}
	for k, v := range p.Options {
		if k == "effort" {
			k = config.OptionEffort
		}
		opts[k] = v
	}
	err = c.updateSessionFields(ctx, p.ID, func(f *store.SessionFields) {
		if p.Reset {
			// Back to the pool's values: clear what was set so applyConfig
			// sets the pool's value again.
			for k := range f.Override {
				delete(f.Options, k)
			}
			f.Override = nil
		}
		if len(opts) > 0 {
			if f.Override == nil {
				f.Override = map[string]string{}
			}
			maps.Copy(f.Override, opts)
		}
	})
	if err != nil {
		return nil, err
	}
	c.log.Info("session config override", "session", p.ID, "options", opts, "reset", p.Reset, "by", sc.actor)
	c.wakeRunner(p.ID)
	return c.sessionInfo(ctx, p.ID)
}
