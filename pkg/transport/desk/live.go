package desk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// live delivers state, not invalidations. A ready event closes the gap between
// subscribing and the initial snapshot. Reconnect always takes a fresh snapshot.
func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	stream, ok := newSSE(w)
	if !ok {
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	dirty := make(chan struct{}, 1)
	direct := make(chan rpc.Event, 64)
	done := make(chan error, 1)
	go func() {
		_, err := s.ctl.Events(ctx, rpc.EventsSubscribeParams{}, func(ev rpc.Event) {
			if ev.Kind == rpc.EventKindDesk || ev.Kind == rpc.EventKindSystem {
				select {
				case direct <- ev:
				case <-ctx.Done():
				}
				return
			}
			select {
			case dirty <- struct{}{}:
			default:
			}
		})
		done <- err
	}()
	last := map[string][]byte{}
	sendState := func() error {
		queryCtx, stop := context.WithTimeout(ctx, 30*time.Second)
		defer stop()
		request := r.WithContext(queryCtx)
		type section struct {
			name string
			read func(*http.Request) (any, error)
		}
		readers := []section{
			{"attention", func(r *http.Request) (any, error) {
				var result rpc.AttentionResult
				err := s.ctl.Call(r.Context(), rpc.MethodAttention, struct{}{}, &result)
				return result, err
			}},
			{"overview", s.overview},
			{"system", s.systemView},
			{"chats", s.chatList},
			{"improvements", s.improvements},
			{"embeds", s.embedList},
			// The scope and its revision (JOB-163): no provider call. An
			// activity event moves the revision and this section is sent
			// again, which is what tells an open Linear or GitHub view to read.
			{"browse", func(r *http.Request) (any, error) {
				var result rpc.BrowseScopeResult
				err := s.ctl.Call(r.Context(), rpc.MethodBrowseScope, struct{}{}, &result)
				return result, err
			}},
		}
		if ref := r.URL.Query().Get("trace"); ref != "" {
			readers = append(readers, section{"trace", func(request *http.Request) (any, error) {
				request.SetPathValue("ref", ref)
				value, err := s.traceView(request)
				if err != nil {
					// A missing trace must not stop the other live windows.
					return map[string]string{"error": err.Error()}, nil
				}
				return value, nil
			}})
		}
		if r.URL.Query().Get("audit") != "" {
			readers = append(readers, section{"automation", s.automationView})
		}
		for _, reader := range readers {
			value, err := reader.read(request)
			if err != nil {
				return err
			}
			data, err := json.Marshal(value)
			if err != nil {
				return err
			}
			if !bytes.Equal(last[reader.name], data) {
				stream.send(reader.name, 0, value)
				last[reader.name] = data
			}
		}
		return nil
	}
	// The controller's ready event requests the initial state, after registration.
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepalive.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
			stream.f.Flush()
		case err := <-done:
			if err != nil && ctx.Err() == nil {
				stream.send("error", 0, map[string]string{"error": err.Error()})
			}
			return
		case ev := <-direct:
			kind := ev.Kind
			if kind == rpc.EventKindSystem {
				kind = "scheduler"
			}
			stream.send(kind, 0, ev)
		case <-dirty:
			// Bound work during a burst, without a periodic polling loop.
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
			select {
			case <-dirty:
			default:
			}
			if err := sendState(); err != nil {
				stream.send("error", 0, map[string]string{"error": err.Error()})
				return
			}
		}
	}
}
