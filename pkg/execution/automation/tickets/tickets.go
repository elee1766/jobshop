// Package tickets implements the read-only store built-ins:
//
//	tickets(kind="", status=None, project="", pool="", label="", workflow="", key="", state="", parent="", holder="", meta={}, order="newest", limit=100, body=False, comments=False)
//	runs(workflow="", key="", status=None, state="", meta={}, order="newest", limit=100, comments=False)
//	sessions(pool="", status=None, limit=100)
//
// Reads are journaled and replayed like other capability calls, so a
// retried callback sees what its first attempt saw. They do not fence:
// actions are still checked when the controller applies them.
package tickets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
)

// MaxLimit bounds one read.
const MaxLimit = 500

var orders = map[string]bool{"newest": true, "oldest": true, "priority": true}

// Call runs tickets, runs or sessions.
func Call(host api.Host, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if options.Tickets == nil {
		return nil, fmt.Errorf("%s is unavailable in this host", fn.Name())
	}
	if len(args) != 0 {
		return nil, fmt.Errorf("%s: use named arguments", fn.Name())
	}
	q := api.TicketQuery{Order: "newest", Limit: 100}
	var status starlark.Value = starlark.None
	var meta *starlark.Dict
	var body bool
	var err error
	switch fn.Name() {
	case "tickets":
		err = starlark.UnpackArgs(fn.Name(), args, kwargs, "kind?", &q.Kind, "status?", &status, "project?", &q.Project, "pool?", &q.Pool, "label?", &q.Label,
			"workflow?", &q.Workflow, "key?", &q.Key, "state?", &q.State, "parent?", &q.Parent, "holder?", &q.Holder, "meta?", &meta, "order?", &q.Order, "limit?", &q.Limit, "body?", &body, "comments?", &q.Comments)
	case "runs":
		q.Kind = "run"
		err = starlark.UnpackArgs(fn.Name(), args, kwargs, "workflow?", &q.Workflow, "key?", &q.Key, "status?", &status, "state?", &q.State, "meta?", &meta, "order?", &q.Order, "limit?", &q.Limit, "comments?", &q.Comments)
	case "sessions":
		q.Kind = "session"
		status = starlark.NewList([]starlark.Value{starlark.String("open"), starlark.String("in_progress")})
		err = starlark.UnpackArgs(fn.Name(), args, kwargs, "pool?", &q.Pool, "status?", &status, "limit?", &q.Limit)
	}
	if err != nil {
		return nil, err
	}
	switch v := status.(type) {
	case starlark.NoneType:
	case starlark.String:
		q.Statuses = []string{string(v)}
	case *starlark.List, starlark.Tuple:
		raw, err := api.FromValue(v, options.Limits.DataBytes)
		if err != nil {
			return nil, err
		}
		for _, item := range raw.([]any) {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%s: status must be a string or a list of strings", fn.Name())
			}
			q.Statuses = append(q.Statuses, s)
		}
	default:
		return nil, fmt.Errorf("%s: status must be a string or a list of strings", fn.Name())
	}
	for _, s := range q.Statuses {
		if s != "open" && s != "in_progress" && s != "closed" {
			return nil, fmt.Errorf("%s: unknown status %q", fn.Name(), s)
		}
	}
	if meta != nil {
		q.Meta = map[string]string{}
		for _, pair := range meta.Items() {
			key, ok1 := starlark.AsString(pair[0])
			value, ok2 := starlark.AsString(pair[1])
			pack, name, ok3 := strings.Cut(key, ".")
			if !ok1 || !ok2 || !ok3 || pack == "" || name == "" {
				return nil, fmt.Errorf(`%s: meta maps "pack.key" to a string`, fn.Name())
			}
			q.Meta[key] = value
		}
		if len(q.Meta) > 3 {
			return nil, fmt.Errorf("%s: at most 3 meta matches", fn.Name())
		}
	}
	if !orders[q.Order] {
		return nil, fmt.Errorf("%s: order must be newest, oldest or priority", fn.Name())
	}
	if q.Limit < 1 || q.Limit > MaxLimit {
		return nil, fmt.Errorf("%s: limit must be 1 to %d", fn.Name(), MaxLimit)
	}
	out, err := options.Tickets(host.Context(), q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn.Name(), err)
	}
	if !body {
		for _, item := range out {
			if m, ok := item.(map[string]any); ok {
				delete(m, "body")
			}
		}
	}
	value, err := api.ToValue(out, options.Limits.DataBytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn.Name(), err)
	}
	return value, nil
}

// Restore rebuilds a recorded read.
func Restore(value any, options api.Options) (starlark.Value, error) {
	if options.Tickets == nil {
		return nil, errors.New("store reads are unavailable in this host")
	}
	if _, ok := value.([]any); !ok {
		return nil, errors.New("recorded capability result does not match its capability contract")
	}
	return api.ToValue(value, options.Limits.DataBytes)
}
