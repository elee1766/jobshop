package desk

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func TestPoolScaleAndCordon(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{}}
	h := New(Config{}, ctl, nil).Handler()

	w := postJSON(h, "/api/pool/scale", `{"pool":"impl","target":2}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, rpc.MethodPoolScale+` {"pool":"impl","target":2}`, ctl.calls[len(ctl.calls)-1])
	w = postJSON(h, "/api/pool/scale", `{"pool":"impl","auto":true}`, nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, rpc.MethodPoolScale+` {"pool":"impl","auto":true}`, ctl.calls[len(ctl.calls)-1])

	w = postJSON(h, "/api/session/cordon", `{"id":"s1","uncordon":true}`, nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, rpc.MethodSessionCordon+` {"id":"s1","uncordon":true}`, ctl.calls[len(ctl.calls)-1])

	n := len(ctl.calls)
	for name, c := range map[string][2]string{
		"no target": {"/api/pool/scale", `{"pool":"impl"}`},
		"negative":  {"/api/pool/scale", `{"pool":"impl","target":-1}`},
		"both":      {"/api/pool/scale", `{"pool":"p","target":1,"auto":true}`},
		"no id":     {"/api/session/cordon", `{}`},
		"conflict":  {"/api/session/cordon", `{"id":"s1","drain":true,"uncordon":true}`},
	} {
		assert.Equal(t, http.StatusBadRequest, postJSON(h, c[0], c[1], nil).Code, name)
	}
	assert.Len(t, ctl.calls, n)
}
