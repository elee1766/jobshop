package desk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"slices"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/rpc/rpcschema"
	"github.com/elee1766/jobshop/pkg/store"
)

// Generic RPC for plugins (#517): POST /api/rpc/{method} forwards to the
// admin socket with the proxy user as actor, GET /api/rpc lists what this
// Desk allows.

// rpcSessionOnly are methods that only make sense from a session's own
// socket (they act on the caller's claim or write key) or stream. Never
// allowed, whatever desk.rpc says.
var rpcSessionOnly = []string{
	rpc.MethodClaim, rpc.MethodHeartbeat, rpc.MethodRelease, rpc.MethodHandoff,
	rpc.MethodWriteKeyClaim, rpc.MethodWriteKeyFinish, rpc.MethodIngestClaim, rpc.MethodIngestAdmit,
	rpc.MethodIngestDispatch,
	rpc.MethodGuard, rpc.MethodSessionSubscribe, rpc.MethodEventsSubscribe,
}

// rpcOffByDefault need a desk.rpc.allow entry.
var rpcOffByDefault = []string{rpc.MethodIdentityToken, rpc.MethodShutdown}

// rpcAllowed applies the policy: session-only never; deny globs next;
// off-by-default methods need an allow glob; with allow globs set, only
// matching methods.
func rpcAllowed(m string, p config.DeskRPC) bool {
	if !slices.Contains(rpc.Methods(), m) || slices.Contains(rpcSessionOnly, m) {
		return false
	}
	if globMatch(p.Deny, m) {
		return false
	}
	if slices.Contains(rpcOffByDefault, m) {
		return globMatch(p.Allow, m)
	}
	return len(p.Allow) == 0 || globMatch(p.Allow, m)
}

func globMatch(globs []string, m string) bool {
	for _, g := range globs {
		if ok, _ := path.Match(g, m); ok {
			return true
		}
	}
	return false
}

// RPCList is GET /api/rpc: the allowed methods, and a JSON schema for
// each one's params and result (#542). Schema $refs point into Defs.
type RPCList struct {
	Methods []string             `json:"methods"`
	Schemas map[string]RPCSchema `json:"schemas"`
	Defs    map[string]any       `json:"$defs"`
}

// RPCSchema is one method's params and result schemas.
type RPCSchema struct {
	Params map[string]any `json:"params"`
	Result map[string]any `json:"result"`
}

func (s *Server) rpcList(r *http.Request) (any, error) {
	dc, err := s.deskConfig(r)
	if err != nil {
		return nil, err
	}
	out := RPCList{Methods: []string{}, Schemas: map[string]RPCSchema{}, Defs: map[string]any{}}
	for _, m := range rpc.Methods() {
		if rpcAllowed(m, dc.RPC) {
			out.Methods = append(out.Methods, m)
			spec := rpc.Specs[m]
			out.Schemas[m] = RPCSchema{
				Params: rpcschema.Schema(spec.Params, out.Defs),
				Result: rpcschema.Schema(spec.Result, out.Defs),
			}
		}
	}
	slices.Sort(out.Methods)
	return out, nil
}

// errForbidden maps to 403.
var errForbidden = errors.New("forbidden")

// maxRPCBody bounds a plugin's request.
const maxRPCBody = 1 << 20

func (s *Server) rpcCall(r *http.Request) (any, error) {
	m := r.PathValue("method")
	dc, err := s.deskConfig(r)
	if err != nil {
		return nil, err
	}
	if !rpcAllowed(m, dc.RPC) {
		return nil, fmt.Errorf("%w: method %q is not allowed from Desk", errForbidden, m)
	}
	b, err := io.ReadAll(io.LimitReader(r.Body, maxRPCBody+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxRPCBody {
		return nil, fmt.Errorf("%w: body over %d bytes", errBadRequest, maxRPCBody)
	}
	var params json.RawMessage
	if len(b) == 0 {
		params = json.RawMessage("{}")
	} else {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(b, &obj); err != nil {
			return nil, fmt.Errorf("%w: params must be a JSON object", errBadRequest)
		}
		// The actor comes from the proxy user, not the body.
		delete(obj, rpc.MetaKey)
		params, _ = json.Marshal(obj)
	}
	var out json.RawMessage
	if err := s.ctl.Call(r.Context(), m, params, &out); err != nil {
		if errors.Is(err, rpc.ErrForbidden) {
			return nil, fmt.Errorf("%w: %w", errForbidden, err)
		}
		return nil, err
	}
	if out == nil {
		out = json.RawMessage("null")
	}
	return out, nil
}

func (s *Server) deskConfig(r *http.Request) (config.Desk, error) {
	var show rpc.ConfigShowResult
	if err := s.ctl.Call(r.Context(), rpc.MethodConfigShow, struct{}{}, &show); err != nil {
		return config.Desk{}, err
	}
	var d config.Desk
	if raw, ok := show.Config["desk"]; ok {
		b, err := json.Marshal(raw)
		if err != nil {
			return d, err
		}
		if err := json.Unmarshal(b, &d); err != nil {
			return d, err
		}
	}
	return d, nil
}

// rpcErrCode is the error code in an /api/rpc error body.
func rpcErrCode(err error) string {
	switch {
	case errors.Is(err, errForbidden):
		return "forbidden"
	case errors.Is(err, store.ErrNotFound):
		return "not_found"
	case errors.Is(err, errBadRequest), errors.Is(err, store.ErrInvalid), errors.Is(err, store.ErrClosed):
		return "invalid"
	case errors.Is(err, store.ErrConflict):
		return "conflict"
	}
	return "internal"
}
