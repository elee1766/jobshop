package control

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/store"
)

// dispatchController is a subController with admission hooks, a working
// session holding a ticket that follows acme/widget#7, and the thread
// registered.
func dispatchController(t *testing.T, hooks ...config.AdmissionHook) (*Controller, *scanClock, *fakeThreads, domain.Ticket) {
	t.Helper()
	c, clk, src := subController(t)
	cfg := *c.config()
	cfg.Ingest.Admission = hooks
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"github":{"repo":"acme/widget","issue":7}}`)})
	_, _, err = c.store.Claim(t.Context(), tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	c.pollSubscriptions(t.Context(), src)
	return c, clk, src, sess
}

func comment(clk *scanClock, id, body string) threads.Comment {
	return threads.Comment{Comment: github.Comment{ID: id, Author: "bob", Kind: "comment", Body: body, At: clk.now().Add(time.Minute).Truncate(time.Second)}}
}

func inboxStates(t *testing.T, c *Controller) []string {
	t.Helper()
	its, err := ingest.InboxList(t.Context(), c.store, store.InboxFilter{})
	require.NoError(t, err)
	var out []string
	for _, it := range its {
		out = append(out, it.State)
	}
	return out
}

// TestDispatchMutatedPayload: what the session receives is the mutating
// hook's rewrite, not the raw poll.
func TestDispatchMutatedPayload(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t,
		mut("redact", `printf '{"patch":[{"op":"replace","path":"/payload/comment/body","value":"[redacted]"}]}'`))
	src.add("acme/widget#7", comment(clk, "c1", "the password is hunter2").Comment)
	c.pollSubscriptions(t.Context(), src)
	msgs := messagesTo(t, c, sess.ID)
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].Body, "[redacted]")
	assert.NotContains(t, msgs[0].Body, "hunter2")
	assert.Equal(t, []string{domain.InboxDispatched}, inboxStates(t, c))
}

// TestDispatchDenied: a validating deny means no delivery, and the read
// position still moves on.
func TestDispatchDenied(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t, val("nope", `echo '{"decision":"deny","reason":"spam"}'`))
	src.add("acme/widget#7", comment(clk, "c1", "buy now").Comment)
	c.pollSubscriptions(t.Context(), src)
	c.pollSubscriptions(t.Context(), src)
	assert.Empty(t, messagesTo(t, c, sess.ID))
	assert.Equal(t, []string{domain.InboxDenied}, inboxStates(t, c))
}

// TestDispatchHeldThenReleased: a held row is not delivered; once the
// human answers it is delivered, once.
func TestDispatchHeldThenReleased(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t, val("ask", `echo '{"decision":"ask_human","reason":"check"}'`))
	src.add("acme/widget#7", comment(clk, "c1", "hello").Comment)
	c.pollSubscriptions(t.Context(), src)
	assert.Empty(t, messagesTo(t, c, sess.ID))
	assert.Equal(t, []string{domain.InboxHeld}, inboxStates(t, c))

	asks, err := c.store.List(t.Context(), store.Filter{Label: needsHumanLabel})
	require.NoError(t, err)
	require.Len(t, asks, 1)
	_, err = store.Close(t.Context(), c.store, "admin", asks[0].ID, nil, "done")
	require.NoError(t, err)
	c.ingestPass(t.Context())
	c.ingestPass(t.Context())
	require.Len(t, messagesTo(t, c, sess.ID), 1)
	assert.Equal(t, []string{domain.InboxDispatched}, inboxStates(t, c))
}

// TestDispatchCrashAfterAdmit: a dispatcher that dies holding admitted
// rows, or after delivering but before marking them, leaves them to be
// claimed again; the write key makes the second pass deliver nothing new.
func TestDispatchCrashAfterAdmit(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t)
	src.add("acme/widget#7", comment(clk, "c1", "hello").Comment)
	// Write and admit only: the crash happens before dispatch.
	th := threads.Thread{Repo: "acme/widget", Number: 7}
	require.NoError(t, c.inboxComment(t.Context(), th, src.cs["acme/widget#7"][0]))
	ok, err := c.admitNext(t.Context())
	require.NoError(t, err)
	require.True(t, ok)
	// A dispatcher claims it and dies.
	it, err := ingest.InboxClaim(t.Context(), c.store, []string{domain.InboxAdmitted}, dispatchHolder, dispatchLease)
	require.NoError(t, err)
	c.ingestPass(t.Context())
	assert.Empty(t, messagesTo(t, c, sess.ID), "leased by the dead dispatcher")

	clk.advance(dispatchLease + time.Second)
	c.ingestPass(t.Context())
	require.Len(t, messagesTo(t, c, sess.ID), 1, "redelivered after the lease expires")

	// The same row admitted again (a lost dispatched write): no repeat.
	_, err = ingest.InboxRequeue(t.Context(), c.store, it.ID, domain.InboxDispatched, domain.InboxAdmitted)
	require.NoError(t, err)
	c.ingestPass(t.Context())
	assert.Len(t, messagesTo(t, c, sess.ID), 1, "the write key stops a second delivery")
}

// TestDispatchWebhookFirst: when a webhook copy of a comment is in the
// inbox first, the poller still gets its own row delivered.
func TestDispatchWebhookFirst(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t)
	cm := comment(clk, "c42", "hi")
	putInbox(t, c, domain.InboxItem{DeliveryID: "github:w1", NaturalKey: "github:comment:acme/widget:42:" + cm.At.UTC().Format(time.RFC3339)})
	src.add("acme/widget#7", cm.Comment)
	c.pollSubscriptions(t.Context(), src)
	require.Len(t, messagesTo(t, c, sess.ID), 1)
	assert.True(t, strings.Contains(messagesTo(t, c, sess.ID)[0].Body, "hi"))
}
