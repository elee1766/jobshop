package desk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// Push notifies subscribed browsers of new attention items (needs-human
// runs and tickets, quarantined pools, failed deploys). Desk keeps its
// own state file, not the database: it is an admin-socket client.
type Push struct {
	// File holds the VAPID key, subscriptions and the attention keys
	// already sent.
	File string
	// Subject is the VAPID contact (mailto: or https: URL) push services
	// see; some reject a missing one.
	Subject string
	// Interval is how often attention is read. Events also wake it.
	Interval time.Duration

	mu    sync.Mutex
	state pushState
	http  *http.Client
	wake  chan struct{}
}

type pushState struct {
	VAPID VAPID          `json:"vapid"`
	Subs  []Subscription `json:"subscriptions"`
	// Sent holds the attention keys pushed already, so an item is sent
	// once while it stays on the list. Pruned to the current list.
	Sent []string `json:"sent"`
	// Primed is set after the first read, so items already waiting when
	// push is turned on do not all fire at once.
	Primed bool `json:"primed"`
}

// PushMessage is the payload the service worker shows.
type PushMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Tag   string `json:"tag"`
	URL   string `json:"url"`
}

func (p *Push) open() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.http = &http.Client{Timeout: 30 * time.Second}
	p.wake = make(chan struct{}, 1)
	b, err := os.ReadFile(p.File)
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return err
	default:
		if err := json.Unmarshal(b, &p.state); err != nil {
			return fmt.Errorf("%s: %w", p.File, err)
		}
	}
	if p.state.VAPID.Private == "" {
		v, err := NewVAPID()
		if err != nil {
			return err
		}
		p.state.VAPID = *v
		return p.saveLocked()
	}
	_, err = p.state.VAPID.Public()
	return err
}

// saveLocked writes the state file atomically, readable only by its owner.
func (p *Push) saveLocked() error {
	b, err := json.MarshalIndent(p.state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p.File), 0o700); err != nil {
		return err
	}
	tmp := p.File + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p.File)
}

func (p *Push) publicKey() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	k, _ := p.state.VAPID.Public()
	return k
}

func (p *Push) subscribe(sub Subscription) error {
	if sub.Endpoint == "" || sub.Keys.P256dh == "" || sub.Keys.Auth == "" {
		return fmt.Errorf("%w: subscription needs endpoint, keys.p256dh and keys.auth", errBadRequest)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state.Subs = slices.DeleteFunc(p.state.Subs, func(s Subscription) bool { return s.Endpoint == sub.Endpoint })
	p.state.Subs = append(p.state.Subs, sub)
	return p.saveLocked()
}

func (p *Push) unsubscribe(endpoint string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state.Subs = slices.DeleteFunc(p.state.Subs, func(s Subscription) bool { return s.Endpoint == endpoint })
	return p.saveLocked()
}

func (p *Push) poke() {
	select {
	case p.wake <- struct{}{}:
	default:
	}
}

func attentionKey(it AttentionItem) string { return it.Kind + ":" + it.ID }

// diff returns items not sent yet and records the current list as sent.
// The first call only records.
func (p *Push) diff(items []AttentionItem) []AttentionItem {
	p.mu.Lock()
	defer p.mu.Unlock()
	keys := make([]string, 0, len(items))
	var fresh []AttentionItem
	for _, it := range items {
		k := attentionKey(it)
		keys = append(keys, k)
		if p.state.Primed && !slices.Contains(p.state.Sent, k) {
			fresh = append(fresh, it)
		}
	}
	changed := !p.state.Primed || !slices.Equal(keys, p.state.Sent)
	p.state.Sent, p.state.Primed = keys, true
	if changed {
		_ = p.saveLocked()
	}
	return fresh
}

func message(it AttentionItem) PushMessage {
	// Relative to the service worker's scope, so it works under --base-path.
	// The Attention plugin opens the item's resolving chat from ?item= (#535).
	m := PushMessage{Title: it.Title, Body: it.Why, Tag: attentionKey(it), URL: "./?item=" + url.QueryEscape(attentionKey(it))}
	return m
}

// send pushes m to every subscription, dropping those the service says
// are gone.
func (p *Push) send(ctx context.Context, m PushMessage, log func(string, ...any)) {
	p.mu.Lock()
	subs := slices.Clone(p.state.Subs)
	v := p.state.VAPID
	p.mu.Unlock()
	body, _ := json.Marshal(m)
	var gone []string
	for _, s := range subs {
		err := v.Send(ctx, p.http, s, p.Subject, body)
		switch {
		case errors.Is(err, errGone):
			gone = append(gone, s.Endpoint)
		case err != nil:
			log("desk: push", "err", err)
		}
	}
	for _, e := range gone {
		_ = p.unsubscribe(e)
	}
}

// runPush reads attention every Interval or on wake and pushes new items.
func (s *Server) runPush(ctx context.Context) {
	p := s.cfg.Push
	t := time.NewTicker(p.Interval)
	defer t.Stop()
	for {
		items, err := s.attention((&http.Request{}).WithContext(ctx))
		if err != nil {
			s.log.Warn("desk: push: attention", "err", err)
		} else {
			for _, it := range p.diff(items.([]AttentionItem)) {
				p.send(ctx, message(it), s.log.Warn)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		case <-p.wake:
			// Coalesce an event burst into one read.
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// watchEvents wakes the push loop on controller changes, reconnecting
// when the admin socket drops.
func (s *Server) watchEvents(ctx context.Context) {
	for ctx.Err() == nil {
		kinds := []string{rpc.EventKindTicket, rpc.EventKindRun, rpc.EventKindPool}
		_, _ = s.ctl.Events(ctx, rpc.EventsSubscribeParams{Kinds: kinds}, func(rpc.Event) { s.cfg.Push.poke() })
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}
	}
}

func (s *Server) pushKey(*http.Request) (any, error) {
	return map[string]string{"key": s.cfg.Push.publicKey()}, nil
}

func (s *Server) pushSubscribe(r *http.Request) (any, error) {
	var sub Subscription
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&sub); err != nil {
		return nil, fmt.Errorf("%w: %w", errBadRequest, err)
	}
	if err := s.cfg.Push.subscribe(sub); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}

func (s *Server) pushUnsubscribe(r *http.Request) (any, error) {
	var sub Subscription
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&sub); err != nil {
		return nil, fmt.Errorf("%w: %w", errBadRequest, err)
	}
	return map[string]bool{"ok": true}, s.cfg.Push.unsubscribe(sub.Endpoint)
}

// pushTest sends a test notification to every subscription.
func (s *Server) pushTest(r *http.Request) (any, error) {
	s.cfg.Push.send(r.Context(), PushMessage{Title: "Jobshop Desk", Body: "Test notification", Tag: "test", URL: "./"}, s.log.Warn)
	return map[string]bool{"ok": true}, nil
}
