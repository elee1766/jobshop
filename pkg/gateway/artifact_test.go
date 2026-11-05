package gateway

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestArtifactLine: a created artifact posts one line with title, source
// and Desk link; other events and an unset room post nothing (#508).
func TestArtifactLine(t *testing.T) {
	meta, _ := json.Marshal(map[string]rpc.ArtifactMeta{"artifact": {Trigger: "nightly", Pool: "ops"}})
	ad := &fakeAdapter{}
	ctl := &fakeController{ticket: map[string]store.Ticket{
		"js-a": {ID: "js-a", Kind: store.KindMessage, Title: "Nightly report", Meta: meta, Labels: []string{ArtifactLabel}},
	}}
	g := New(Config{Name: "ops", Pool: "chat"}, ad, ctl, nil)
	ev := rpc.Event{Kind: rpc.EventKindTicket, Op: "create", ID: "js-a", TicketKind: store.KindMessage, Labels: []string{ArtifactLabel}}

	g.artifactEvent(t.Context(), ev)
	assert.Zero(t, ad.opened, "no room: nothing posted")

	g.Apply(t.Context(), Config{ArtifactRoom: "!art", DeskURL: "https://desk.example/"})
	g.artifactEvent(t.Context(), rpc.Event{Kind: rpc.EventKindTicket, Op: "close", ID: "js-a", TicketKind: store.KindMessage})
	assert.Zero(t, ad.opened, "only create posts")
	g.artifactEvent(t.Context(), ev)
	require.Equal(t, 1, ad.opened)
	assert.Equal(t, "!art/$open1: artifact js-a: Nightly report (from trigger nightly) https://desk.example/?view=artifacts&artifact=js-a", ad.got()[0])
}
