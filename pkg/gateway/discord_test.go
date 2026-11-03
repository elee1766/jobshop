package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

func discardLog() *slog.Logger { return slog.New(slog.DiscardHandler) }

// fakeDiscord is a fake Discord REST session. Channels: "text" (text),
// "forum" (forum), "th1" (a thread in text).
type fakeDiscord struct {
	mu       sync.Mutex
	channels map[string]*discordgo.Channel
	sent     []string // channel: content
	threads  []string // started from message: channel/msg name
	forum    []string // forum posts: channel name|content
	reacts   []string // channel/msg key
	n        int
}

func newFakeDiscord() *fakeDiscord {
	return &fakeDiscord{channels: map[string]*discordgo.Channel{
		"text":  {ID: "text", Type: discordgo.ChannelTypeGuildText},
		"forum": {ID: "forum", Type: discordgo.ChannelTypeGuildForum},
		"th1":   {ID: "th1", Type: discordgo.ChannelTypeGuildPublicThread, ParentID: "text"},
		"dm":    {ID: "dm", Type: discordgo.ChannelTypeDM},
	}}
}

func (f *fakeDiscord) Channel(id string, _ ...discordgo.RequestOption) (*discordgo.Channel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if c, ok := f.channels[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("unknown channel %s", id)
}

func (f *fakeDiscord) ChannelMessageSend(ch, content string, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	f.sent = append(f.sent, ch+": "+content)
	return &discordgo.Message{ID: fmt.Sprintf("m%d", f.n), ChannelID: ch, Content: content}, nil
}

func (f *fakeDiscord) MessageThreadStart(ch, msg, name string, _ int, _ ...discordgo.RequestOption) (*discordgo.Channel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.threads = append(f.threads, ch+"/"+msg+" "+name)
	// Discord gives a thread started from a message the message's id.
	c := &discordgo.Channel{ID: msg, Type: discordgo.ChannelTypeGuildPublicThread, ParentID: ch}
	f.channels[msg] = c
	return c, nil
}

func (f *fakeDiscord) ForumThreadStart(ch, name string, _ int, content string, _ ...discordgo.RequestOption) (*discordgo.Channel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	f.forum = append(f.forum, ch+" "+name+"|"+content)
	id := fmt.Sprintf("post%d", f.n)
	c := &discordgo.Channel{ID: id, Type: discordgo.ChannelTypeGuildPublicThread, ParentID: ch}
	f.channels[id] = c
	return c, nil
}

func (f *fakeDiscord) MessageReactionAdd(ch, msg, key string, _ ...discordgo.RequestOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reacts = append(f.reacts, ch+"/"+msg+" "+key)
	return nil
}

func discordMsg(id, ch, author, text string) *discordgo.Message {
	return &discordgo.Message{ID: id, ChannelID: ch, Content: text, Author: &discordgo.User{ID: author}}
}

func TestDiscordMessage(t *testing.T) {
	f := newFakeDiscord()
	d := &Discord{api: f, log: discardLog(), me: "bot"}

	// Outside a thread: a thread is started on the message.
	m, ok := d.message(discordMsg("100", "text", "u1", "hello\nmore"))
	require.True(t, ok)
	assert.Equal(t, Thread{Room: "text", Root: "100"}, m.Thread)
	assert.Equal(t, []string{"text/100 hello"}, f.threads)

	// In a thread: room is the parent channel, root the thread.
	m, ok = d.message(discordMsg("101", "th1", "u1", "again"))
	require.True(t, ok)
	assert.Equal(t, Thread{Room: "text", Root: "th1"}, m.Thread)
	assert.Len(t, f.threads, 1)

	// Bot, own, empty, system and DM messages are dropped.
	bot := discordMsg("102", "text", "u2", "x")
	bot.Author.Bot = true
	sys := discordMsg("103", "text", "u1", "x")
	sys.Type = discordgo.MessageTypeThreadCreated
	for name, msg := range map[string]*discordgo.Message{
		"bot": bot, "own": discordMsg("104", "text", "bot", "x"), "empty": discordMsg("105", "text", "u1", " "),
		"system": sys, "dm": discordMsg("106", "dm", "u1", "x"),
	} {
		_, ok := d.message(msg)
		assert.False(t, ok, name)
	}
	assert.Len(t, f.threads, 1)
}

func TestDiscordRouting(t *testing.T) {
	f := newFakeDiscord()
	d := &Discord{api: f, log: discardLog(), me: "bot"}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16)}
	g := New(Config{Pool: "chat", Rooms: []string{"text", "forum"}, Senders: []string{"u1"}}, d, ctl, nil)

	m, ok := d.message(discordMsg("100", "text", "u1", "hi"))
	require.True(t, ok)
	assert.Equal(t, "discord:text:100", g.Key(m.Thread))
	assert.True(t, g.Allowed(m))
	m2, _ := d.message(discordMsg("101", "th1", "u9", "hi"))
	assert.False(t, g.Allowed(m2), "sender not allowed")

	// Reacting to the thread's starter message goes to the parent channel.
	require.NoError(t, d.React(context.Background(), m2.Thread, "101", ReactQueued))
	assert.Equal(t, []string{"th1/101 " + ReactQueued}, f.reacts)
	f.reacts = nil

	// End to end: the reply is posted in the thread, keyed by it.
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); g.wg.Wait() }()
	g.Handle(ctx, m)
	g.Handle(ctx, m2)
	require.Eventually(t, func() bool {
		f.mu.Lock()
		defer f.mu.Unlock()
		return len(f.sent) == 1
	}, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, "100: echo: hi", f.sent[0])
	assert.Equal(t, []string{"text/100 " + ReactQueued}, f.reacts)
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	require.Len(t, ctl.sends, 1)
	assert.Equal(t, "discord:text:100", ctl.sends[0].Key)
}

func TestDiscordOpen(t *testing.T) {
	ctx := context.Background()
	f := newFakeDiscord()
	d := &Discord{api: f, log: discardLog()}

	// Forum escalation room: one new post per escalation.
	th, err := d.Open(ctx, "forum", "escalation js-1\ndetails")
	require.NoError(t, err)
	assert.Equal(t, "forum", th.Room)
	assert.Equal(t, []string{"forum escalation js-1|escalation js-1\ndetails"}, f.forum)
	th2, err := d.Open(ctx, "forum", "escalation js-2")
	require.NoError(t, err)
	assert.NotEqual(t, th.Root, th2.Root)

	// Text channel: a message with a thread on it.
	th, err = d.Open(ctx, "text", "escalation js-3")
	require.NoError(t, err)
	assert.Equal(t, Thread{Room: "text", Root: th.Root}, th)
	assert.Equal(t, []string{"text/" + th.Root + " escalation js-3"}, f.threads)

	// A long forum post continues in the post.
	f.sent = nil
	th, err = d.Open(ctx, "forum", strings.Repeat("a ", 1500))
	require.NoError(t, err)
	require.Len(t, f.sent, 1)
	assert.True(t, strings.HasPrefix(f.sent[0], th.Root+": "))
}

func TestSplitMessage(t *testing.T) {
	assert.Nil(t, splitMessage("", 10))
	assert.Equal(t, []string{"short"}, splitMessage("short", 10))
	assert.Equal(t, []string{"aaaa bbbb", "cccc"}, splitMessage("aaaa bbbb cccc", 10))
	assert.Equal(t, []string{"aaaaa", "bb cccc"}, splitMessage("aaaaa\nbb cccc", 10))
	assert.Equal(t, []string{"aaaaaaaaaa", "aaa"}, splitMessage(strings.Repeat("a", 13), 10))

	long := strings.Repeat("é line of text\n", 400)
	parts := splitMessage(long, discordLimit)
	require.Greater(t, len(parts), 1)
	for _, p := range parts {
		assert.LessOrEqual(t, utf8.RuneCountInString(p), discordLimit)
	}
	assert.Equal(t, strings.ReplaceAll(strings.TrimSpace(long), "\n", ""), strings.ReplaceAll(strings.Join(parts, ""), "\n", ""))

	f := newFakeDiscord()
	d := &Discord{api: f, log: discardLog()}
	require.NoError(t, d.Post(context.Background(), Thread{Room: "text", Root: "th1"}, long))
	assert.Len(t, f.sent, len(parts))
}
