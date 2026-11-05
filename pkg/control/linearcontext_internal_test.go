package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestLinearCommentContext (#595) reads testdata/linear_comments.json: two
// users named "Ann Lee", a reply, an edit of an older comment and a bot.
// The allowlist matches by id and displayName, so the other Ann is not
// delivered; the mixed batch is authored "linear:several"; the reply
// quotes its parent; the header has title and state; the edit arrives
// once.
func TestLinearCommentContext(t *testing.T) {
	t.Parallel()
	c, clk, _ := subController(t)
	ctx := t.Context()
	data, err := os.ReadFile(filepath.Join("testdata", "linear_comments.json"))
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(data) }))
	t.Cleanup(srv.Close)
	tokFile := filepath.Join(t.TempDir(), "tok")
	require.NoError(t, os.WriteFile(tokFile, []byte("lin_api_x\n"), 0o600))
	cfg := *c.config()
	cfg.Identities = map[string]config.Identity{"lin": {Kind: config.IdentityLinear, TokenFile: config.SecretPath(tokFile), Workspace: "acme", Teams: []string{"ENG"}}}
	subs := *cfg.Subscriptions
	subs.Authors = []string{"u-ann", "bob"}
	subs.Bots = []string{"b-ci"}
	cfg.Subscriptions = &subs
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter, c.linearURL = identity.NewMinter(), srv.URL

	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"linear":{"issue":"ENG-5"}}`)})
	_, _, err = c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)

	src := muxSource{gh: &fakeThreads{}, linear: linearSource{c: c, url: c.linearURL}}
	c.pollSubscriptions(ctx, src) // follows ENG-5 at 08:00
	clk.advance(10 * time.Minute)
	c.pollSubscriptions(ctx, src)
	c.pollSubscriptions(ctx, src) // nothing new: the edit is not sent again

	cs, err := c.store.Comments(ctx, tk.ID)
	require.NoError(t, err)
	var got []store.Comment
	for _, cm := range cs {
		if strings.HasPrefix(cm.Author, linearAuthorPrefix) {
			got = append(got, cm)
		}
	}
	require.Len(t, got, 1, "one delivery")
	assert.Equal(t, "linear:several", got[0].Author)
	body := got[0].Body
	assert.True(t, strings.HasPrefix(body, `New on Linear issue ENG-5 "Fix login" (In Progress):`), body)
	assert.Contains(t, body, "ann (Ann Lee) comment https://linear.app/acme/issue/ENG-5/fix-login#comment-c4\n"+`in reply to bob (Bob Ray): "Should the old login page stay?"`+"\nYes, keep")
	assert.Contains(t, body, "bob (Bob Ray) comment ")
	assert.Contains(t, body, "Deploy Bot comment ")
	assert.NotContains(t, body, "same name", "the other Ann Lee is not on authors")
	assert.Equal(t, 1, strings.Count(body, "ann (Ann Lee) comment (edited) "), body)
	assert.Contains(t, body, "also Firefox")
	assert.Len(t, messagesTo(t, c, sess.ID), 1)
}

// TestLinearDeliveryAuthor: one author gets their displayName, a reply to
// a comment outside the fetched set links its URL.
func TestLinearDeliveryAuthor(t *testing.T) {
	t.Parallel()
	th := thread{Linear: "ENG-5"}
	one := []ghComment{{Linear: true, AuthorID: "u1", Handle: "ann", Author: "ann (Ann Lee)"}, {Linear: true, AuthorID: "u1", Handle: "ann", Author: "ann (Ann Lee)"}}
	assert.Equal(t, "linear:ann", deliveryAuthor(th, one))
	two := append(one, ghComment{Linear: true, AuthorID: "u2", Handle: "ann", Author: "ann (Ann Lee)"})
	assert.Equal(t, "linear:several", deliveryAuthor(th, two), "same handle, different id")

	byID := map[string]linear.Comment{}
	assert.Equal(t, "in reply to https://linear.app/c/c0", replyLine(linear.Comment{ParentID: "c0", ParentURL: "https://linear.app/c/c0"}, byID))
	byID["c0"] = linear.Comment{Author: "Bob", Body: strings.Repeat("x", 130) + "\nmore"}
	assert.Equal(t, `in reply to Bob: "`+strings.Repeat("x", 120)+`..."`, replyLine(linear.Comment{ParentID: "c0"}, byID))
}
