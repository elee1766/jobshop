package desk

import (
	"fmt"
	"net/http"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// poolScale sets a pool's manual target or returns it to auto (#284).
func (s *Server) poolScale(r *http.Request) (any, error) {
	var b rpc.PoolScaleParams
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if b.Pool == "" || b.Auto == (b.Target != nil) {
		return nil, fmt.Errorf("%w: give pool and one of target or auto", errBadRequest)
	}
	if b.Target != nil && *b.Target < 0 {
		return nil, fmt.Errorf("%w: target must be >= 0", errBadRequest)
	}
	var res rpc.PoolScale
	if err := s.ctl.Call(r.Context(), rpc.MethodPoolScale, b, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// sessionCordon cordons, drains or uncordons one session (#284).
func (s *Server) sessionCordon(r *http.Request) (any, error) {
	var b rpc.SessionCordonParams
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	if b.ID == "" || (b.Drain && b.Uncordon) {
		return nil, fmt.Errorf("%w: give id, and not both drain and uncordon", errBadRequest)
	}
	var res rpc.SessionInfo
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionCordon, b, &res); err != nil {
		return nil, err
	}
	return res, nil
}
