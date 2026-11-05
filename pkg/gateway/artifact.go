package gateway

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// ArtifactLabel marks artifact tickets (#508).
const ArtifactLabel = "artifact"

// runArtifacts posts one line per delivered artifact to ArtifactRoom
// until ctx ends. It resumes from the last seen event after a reconnect,
// so a controller restart does not drop or repeat lines.
func (g *Gateway) runArtifacts(ctx context.Context) {
	var since int64
	for ctx.Err() == nil {
		res, err := g.ctl.Events(ctx, rpc.EventsSubscribeParams{SinceSeq: since, Labels: []string{ArtifactLabel}}, func(ev rpc.Event) {
			g.artifactEvent(ctx, ev)
		})
		if ctx.Err() != nil {
			return
		}
		if res.Seq > since {
			since = res.Seq
		}
		if err != nil {
			g.log.Warn("gateway: artifact events", "err", err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(g.retry):
		}
	}
}

func (g *Gateway) artifactEvent(ctx context.Context, ev rpc.Event) {
	if ev.Kind != rpc.EventKindTicket || ev.Op != "create" || ev.TicketKind != store.KindMessage {
		return
	}
	room := g.settings().ArtifactRoom
	if room == "" {
		return
	}
	r, err := g.ctl.Show(ctx, ev.ID)
	if err != nil {
		g.log.Warn("gateway: artifact", "ticket", ev.ID, "err", err)
		return
	}
	if _, err := g.ad.Open(ctx, room, artifactLine(r.Ticket, g.settings().DeskURL)); err != nil {
		g.log.Warn("gateway: artifact post", "ticket", ev.ID, "err", err)
	}
}

// artifactLine is "artifact <id>: <title> (from <source>) <desk link>".
func artifactLine(t store.Ticket, desk string) string {
	var m struct {
		Artifact rpc.ArtifactMeta `json:"artifact"`
	}
	_ = json.Unmarshal(t.Meta, &m)
	src := m.Artifact.Trigger
	if src != "" {
		src = "trigger " + src
	} else if m.Artifact.Pool != "" {
		src = "pool " + m.Artifact.Pool
	}
	s := "artifact " + t.ID + ": " + t.Title
	if src != "" {
		s += " (from " + src + ")"
	}
	if desk != "" {
		s += " " + strings.TrimRight(desk, "/") + "/?view=artifacts&artifact=" + url.QueryEscape(t.ID)
	}
	return s
}
