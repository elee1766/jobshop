package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

// discordLimit is Discord's message content limit, in characters.
const discordLimit = 2000

// discordAPI is the part of *discordgo.Session the adapter calls, so tests
// can use a fake session.
type discordAPI interface {
	Channel(channelID string, options ...discordgo.RequestOption) (*discordgo.Channel, error)
	ChannelMessageSend(channelID, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
	MessageThreadStart(channelID, messageID, name string, archiveDuration int, options ...discordgo.RequestOption) (*discordgo.Channel, error)
	ForumThreadStart(channelID, name string, archiveDuration int, content string, options ...discordgo.RequestOption) (*discordgo.Channel, error)
	MessageReactionAdd(channelID, messageID, emojiID string, options ...discordgo.RequestOption) error
}

// Discord is the Discord bot adapter (discordgo). Thread.Room is the
// text or forum channel, Thread.Root the thread channel. A message
// outside a thread starts a thread on itself (a thread started from a
// message has the message's id); forum messages are always in a post.
type Discord struct {
	s   *discordgo.Session // nil in tests
	api discordAPI
	log *slog.Logger
	me  string // bot user id
}

// NewDiscord returns an adapter for a bot token (without "Bot ").
func NewDiscord(token string, log *slog.Logger) (*Discord, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	s.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
	// discordgo reconnects and resumes the gateway session on errors.
	s.ShouldReconnectOnError = true
	// Handlers run in order, as Adapter.Run requires.
	s.SyncEvents = true
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Discord{s: s, api: s, log: log}, nil
}

// Network implements Adapter.
func (d *Discord) Network() string { return "discord" }

// Run opens the gateway connection and handles MessageCreate until ctx
// ends. Edits arrive as MessageUpdate and are not handled.
func (d *Discord) Run(ctx context.Context, fn func(context.Context, Message)) error {
	d.s.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) { d.me = r.User.ID })
	d.s.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		if msg, ok := d.message(m.Message); ok {
			fn(ctx, msg)
		}
	})
	if err := d.s.Open(); err != nil {
		return fmt.Errorf("discord open: %w", err)
	}
	<-ctx.Done()
	_ = d.s.Close()
	return ctx.Err()
}

// message maps a created message to a Message, starting a thread on it
// when it is not in one. Bot, own, empty and non-default messages are
// dropped.
func (d *Discord) message(m *discordgo.Message) (Message, bool) {
	if m == nil || m.Author == nil || m.Author.Bot || m.Author.ID == d.me {
		return Message{}, false
	}
	if m.Type != discordgo.MessageTypeDefault && m.Type != discordgo.MessageTypeReply {
		return Message{}, false
	}
	if strings.TrimSpace(m.Content) == "" {
		return Message{}, false
	}
	ch, err := d.api.Channel(m.ChannelID)
	if err != nil {
		d.log.Warn("discord: channel", "channel", m.ChannelID, "err", err)
		return Message{}, false
	}
	t := Thread{Room: m.ChannelID, Root: m.ChannelID}
	switch {
	case ch.IsThread():
		t.Room = ch.ParentID
	case ch.Type == discordgo.ChannelTypeGuildText || ch.Type == discordgo.ChannelTypeGuildNews:
		th, err := d.api.MessageThreadStart(m.ChannelID, m.ID, threadName(m.Content), 0)
		if err != nil {
			d.log.Warn("discord: start thread", "channel", m.ChannelID, "err", err)
			return Message{}, false
		}
		t.Root = th.ID
	default:
		return Message{}, false
	}
	return Message{Thread: t, ID: m.ID, Sender: m.Author.ID, Text: m.Content}, true
}

// Post implements Adapter, splitting at Discord's limit.
func (d *Discord) Post(_ context.Context, t Thread, text string) error {
	for _, part := range splitMessage(strings.TrimSpace(text), discordLimit) {
		if _, err := d.api.ChannelMessageSend(t.Root, part); err != nil {
			return err
		}
	}
	return nil
}

// React implements Adapter. A thread's starter message has the thread's
// id and lives in the parent channel.
func (d *Discord) React(_ context.Context, t Thread, msgID, key string) error {
	ch := t.Root
	if msgID == t.Root {
		ch = t.Room
	}
	return d.api.MessageReactionAdd(ch, msgID, key)
}

// Open implements Adapter: a new post in a forum channel, else a message
// in the channel with a thread started on it.
func (d *Discord) Open(ctx context.Context, room, text string) (Thread, error) {
	ch, err := d.api.Channel(room)
	if err != nil {
		return Thread{}, err
	}
	parts := splitMessage(strings.TrimSpace(text), discordLimit)
	if len(parts) == 0 {
		parts = []string{"(empty)"}
	}
	var th *discordgo.Channel
	if ch.Type == discordgo.ChannelTypeGuildForum {
		th, err = d.api.ForumThreadStart(room, threadName(parts[0]), 0, parts[0])
	} else {
		var m *discordgo.Message
		if m, err = d.api.ChannelMessageSend(room, parts[0]); err == nil {
			th, err = d.api.MessageThreadStart(room, m.ID, threadName(parts[0]), 0)
		}
	}
	if err != nil {
		return Thread{}, err
	}
	t := Thread{Room: room, Root: th.ID}
	if len(parts) > 1 {
		if err := d.Post(ctx, t, strings.Join(parts[1:], "\n")); err != nil {
			return t, err
		}
	}
	return t, nil
}

// threadName is the first line of text, cut to Discord's 100 characters.
func threadName(text string) string {
	name, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	name = strings.TrimSpace(name)
	if name == "" {
		name = "jobshop"
	}
	if r := []rune(name); len(r) > 100 {
		name = string(r[:100])
	}
	return name
}

// splitMessage cuts text into parts of at most limit characters,
// preferring a line break, then a space, in the second half of a part.
func splitMessage(text string, limit int) []string {
	var parts []string
	for text != "" {
		if utf8.RuneCountInString(text) <= limit {
			parts = append(parts, text)
			break
		}
		r := []rune(text)
		cut := limit
		head := string(r[:limit])
		if i := strings.LastIndex(head, "\n"); i > 0 && utf8.RuneCountInString(head[:i]) >= limit/2 {
			cut = utf8.RuneCountInString(head[:i])
		} else if i := strings.LastIndex(head, " "); i > 0 && utf8.RuneCountInString(head[:i]) >= limit/2 {
			cut = utf8.RuneCountInString(head[:i])
		}
		parts = append(parts, string(r[:cut]))
		text = strings.TrimLeft(string(r[cut:]), " \n")
	}
	return parts
}
