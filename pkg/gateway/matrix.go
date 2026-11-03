package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// Matrix is the Matrix client-server adapter (mautrix-go). Unencrypted
// rooms only. A thread's root is the m.thread parent; a message outside a
// thread starts one rooted at itself, and the reply goes there.
type Matrix struct {
	cli *mautrix.Client
	log *slog.Logger
	// rooms to join on start (invites are not auto-accepted).
	rooms []string
}

// matrixRetries is how often a failed Matrix request is retried. With
// mautrix's 4s doubling backoff that is about a minute in total.
const matrixRetries = 4

// NewMatrix logs in with an access token and checks it with whoami.
func NewMatrix(ctx context.Context, homeserver, token string, rooms []string, log *slog.Logger) (*Matrix, error) {
	cli, err := mautrix.NewClient(homeserver, "", token)
	if err != nil {
		return nil, err
	}
	// mautrix defaults to no retries, so a 429 or a homeserver restart
	// during a turn end dropped the reply (#196). With retries it backs
	// off on 429 and 5xx, honoring Retry-After.
	cli.DefaultHTTPRetries = matrixRetries
	who, err := cli.Whoami(ctx)
	if err != nil {
		return nil, fmt.Errorf("matrix whoami: %w", err)
	}
	cli.UserID, cli.DeviceID = who.UserID, who.DeviceID
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Matrix{cli: cli, log: log, rooms: rooms}, nil
}

// Network implements Adapter.
func (m *Matrix) Network() string { return "matrix" }

// Run joins the allowed rooms and syncs. Events from before the first
// sync are skipped, so a restart does not replay history.
func (m *Matrix) Run(ctx context.Context, fn func(context.Context, Message)) error {
	for _, r := range m.rooms {
		if _, err := m.cli.JoinRoomByID(ctx, id.RoomID(r)); err != nil {
			m.log.Warn("matrix: join", "room", r, "err", err)
		}
	}
	syncer := mautrix.NewDefaultSyncer()
	first := true
	syncer.OnSync(func(_ context.Context, _ *mautrix.RespSync, _ string) bool {
		if first {
			first = false
			return false
		}
		return true
	})
	syncer.OnEventType(event.EventMessage, func(ctx context.Context, evt *event.Event) {
		if evt.Sender == m.cli.UserID {
			return
		}
		msg, ok := matrixMessage(evt)
		if ok {
			fn(ctx, msg)
		}
	})
	m.cli.Syncer = syncer
	return m.cli.SyncWithContext(ctx)
}

// matrixMessage maps a text event to a Message; edits and non-text are dropped.
func matrixMessage(evt *event.Event) (Message, bool) {
	c := evt.Content.AsMessage()
	if c == nil || (c.MsgType != event.MsgText && c.MsgType != event.MsgNotice) {
		return Message{}, false
	}
	root := evt.ID
	if rel := c.GetRelatesTo(); rel != nil {
		if rel.GetReplaceID() != "" {
			return Message{}, false
		}
		if p := rel.GetThreadParent(); p != "" {
			root = p
		}
	}
	return Message{
		Thread: Thread{Room: evt.RoomID.String(), Root: root.String()},
		ID:     evt.ID.String(),
		Sender: evt.Sender.String(),
		Text:   c.Body,
	}, true
}

// Post implements Adapter: an m.notice in the thread.
func (m *Matrix) Post(ctx context.Context, t Thread, text string) error {
	c := &event.MessageEventContent{MsgType: event.MsgNotice, Body: strings.TrimSpace(text)}
	root := id.EventID(t.Root)
	c.RelatesTo = (&event.RelatesTo{}).SetThread(root, root)
	_, err := m.cli.SendMessageEvent(ctx, id.RoomID(t.Room), event.EventMessage, c)
	if err != nil && threadGone(err) {
		return fmt.Errorf("%w: %w", ErrThreadGone, err)
	}
	return err
}

// React implements Adapter.
func (m *Matrix) React(ctx context.Context, t Thread, msgID, key string) error {
	_, err := m.cli.SendReaction(ctx, id.RoomID(t.Room), id.EventID(msgID), key)
	return err
}

// Open implements Adapter: a top-level m.notice whose event id roots the
// thread.
func (m *Matrix) Open(ctx context.Context, room, text string) (Thread, error) {
	c := &event.MessageEventContent{MsgType: event.MsgNotice, Body: strings.TrimSpace(text)}
	res, err := m.cli.SendMessageEvent(ctx, id.RoomID(room), event.EventMessage, c)
	if err != nil {
		return Thread{}, err
	}
	return Thread{Room: room, Root: res.EventID.String()}, nil
}

// threadGone reports a post rejected because the thread root does not
// exist in the room (#197). Synapse answers M_INVALID_PARAM "Invalid
// event_id in thread message: M_NOT_FOUND"; retrying gets the same error.
func threadGone(err error) bool {
	var re mautrix.RespError
	if he := (mautrix.HTTPError{}); errors.As(err, &he) && he.RespError != nil {
		re = *he.RespError
	} else if !errors.As(err, &re) {
		return false
	}
	switch re.ErrCode {
	case "M_NOT_FOUND":
		return true
	case "M_INVALID_PARAM":
		return strings.Contains(re.Err, "event_id") || strings.Contains(re.Err, "M_NOT_FOUND")
	}
	return false
}
