package ingestqueue

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// applyJSONPatch applies an RFC 6902 patch to doc. Admission is the only
// user; a small local version avoids a dependency (and a vendor hash).
func applyJSONPatch(doc, patch []byte) ([]byte, error) {
	var ops []struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		From  string          `json:"from"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(patch, &ops); err != nil {
		return nil, fmt.Errorf("patch: %w", err)
	}
	var root any
	if err := json.Unmarshal(doc, &root); err != nil {
		return nil, fmt.Errorf("patch target: %w", err)
	}
	for i, op := range ops {
		var err error
		value := func() (any, error) {
			if op.Value == nil {
				return nil, errors.New("value is required")
			}
			var v any
			return v, json.Unmarshal(op.Value, &v)
		}
		switch op.Op {
		case "add", "replace":
			var v any
			if v, err = value(); err == nil {
				root, err = patchSet(root, op.Path, v, op.Op == "add")
			}
		case "remove":
			root, _, err = patchRemove(root, op.Path)
		case "move":
			var v any
			if root, v, err = patchRemove(root, op.From); err == nil {
				root, err = patchSet(root, op.Path, v, true)
			}
		case "copy":
			var v any
			if v, err = patchGet(root, op.From); err == nil {
				root, err = patchSet(root, op.Path, deepCopy(v), true)
			}
		case "test":
			var want, got any
			if want, err = value(); err == nil {
				if got, err = patchGet(root, op.Path); err == nil && !reflect.DeepEqual(want, got) {
					err = errors.New("test failed")
				}
			}
		default:
			err = fmt.Errorf("unknown op %q", op.Op)
		}
		if err != nil {
			return nil, fmt.Errorf("patch op %d (%s %s): %w", i, op.Op, op.Path, err)
		}
	}
	return json.Marshal(root)
}

func deepCopy(v any) any {
	b, _ := json.Marshal(v)
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}

// splitPointer splits an RFC 6901 pointer into unescaped tokens.
func splitPointer(p string) ([]string, error) {
	if p == "" {
		return nil, nil
	}
	if !strings.HasPrefix(p, "/") {
		return nil, fmt.Errorf("pointer %q must start with /", p)
	}
	toks := strings.Split(p[1:], "/")
	for i, t := range toks {
		toks[i] = strings.ReplaceAll(strings.ReplaceAll(t, "~1", "/"), "~0", "~")
	}
	return toks, nil
}

func arrayIndex(tok string, n int, appendOK bool) (int, error) {
	if appendOK && tok == "-" {
		return n, nil
	}
	i, err := strconv.Atoi(tok)
	if err != nil || i < 0 || (tok != "0" && strings.HasPrefix(tok, "0")) {
		return 0, fmt.Errorf("bad array index %q", tok)
	}
	limit := n - 1
	if appendOK {
		limit = n
	}
	if i > limit {
		return 0, fmt.Errorf("array index %d out of range", i)
	}
	return i, nil
}

func patchGet(root any, p string) (any, error) {
	toks, err := splitPointer(p)
	if err != nil {
		return nil, err
	}
	cur := root
	for _, t := range toks {
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[t]
			if !ok {
				return nil, fmt.Errorf("no member %q", t)
			}
			cur = v
		case []any:
			i, err := arrayIndex(t, len(c), false)
			if err != nil {
				return nil, err
			}
			cur = c[i]
		default:
			return nil, fmt.Errorf("cannot descend into %q", t)
		}
	}
	return cur, nil
}

// patchEdit walks to the parent of p and calls edit on it, storing back
// the container edit returns (arrays change identity when resized).
func patchEdit(root any, p string, edit func(parent any, last string) (any, error)) (any, error) {
	toks, err := splitPointer(p)
	if err != nil {
		return nil, err
	}
	if len(toks) == 0 {
		return edit(nil, "")
	}
	var walk func(cur any, toks []string) (any, error)
	walk = func(cur any, toks []string) (any, error) {
		if len(toks) == 1 {
			return edit(cur, toks[0])
		}
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[toks[0]]
			if !ok {
				return nil, fmt.Errorf("no member %q", toks[0])
			}
			nv, err := walk(v, toks[1:])
			if err != nil {
				return nil, err
			}
			c[toks[0]] = nv
			return c, nil
		case []any:
			i, err := arrayIndex(toks[0], len(c), false)
			if err != nil {
				return nil, err
			}
			nv, err := walk(c[i], toks[1:])
			if err != nil {
				return nil, err
			}
			c[i] = nv
			return c, nil
		}
		return nil, fmt.Errorf("cannot descend into %q", toks[0])
	}
	return walk(root, toks)
}

func patchSet(root any, p string, v any, add bool) (any, error) {
	return patchEdit(root, p, func(parent any, last string) (any, error) {
		switch c := parent.(type) {
		case nil:
			if p == "" {
				return v, nil
			}
		case map[string]any:
			if _, ok := c[last]; !ok && !add {
				return nil, fmt.Errorf("no member %q", last)
			}
			c[last] = v
			return c, nil
		case []any:
			i, err := arrayIndex(last, len(c), add)
			if err != nil {
				return nil, err
			}
			if !add {
				c[i] = v
				return c, nil
			}
			c = append(c, nil)
			copy(c[i+1:], c[i:])
			c[i] = v
			return c, nil
		}
		return nil, fmt.Errorf("cannot set %q", last)
	})
}

func patchRemove(root any, p string) (any, any, error) {
	var removed any
	out, err := patchEdit(root, p, func(parent any, last string) (any, error) {
		switch c := parent.(type) {
		case map[string]any:
			v, ok := c[last]
			if !ok {
				return nil, fmt.Errorf("no member %q", last)
			}
			removed = v
			delete(c, last)
			return c, nil
		case []any:
			i, err := arrayIndex(last, len(c), false)
			if err != nil {
				return nil, err
			}
			removed = c[i]
			return append(c[:i], c[i+1:]...), nil
		}
		return nil, errors.New("cannot remove the root")
	})
	return out, removed, err
}

// topLevelDiff records a full-event mutation as replace ops on the event
// fields that changed, so hook_results shows what a hook did.
func topLevelDiff(before, after []byte) json.RawMessage {
	var a, b map[string]any
	if json.Unmarshal(before, &a) != nil || json.Unmarshal(after, &b) != nil {
		return nil
	}
	type op struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value any    `json:"value,omitempty"`
	}
	var ops []op
	for _, k := range sortedKeys(a, b) {
		av, aok := a[k]
		bv, bok := b[k]
		path := "/" + strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1")
		switch {
		case aok && !bok:
			ops = append(ops, op{Op: "remove", Path: path})
		case !aok && bok:
			ops = append(ops, op{Op: "add", Path: path, Value: bv})
		case !reflect.DeepEqual(av, bv):
			ops = append(ops, op{Op: "replace", Path: path, Value: bv})
		}
	}
	out, err := json.Marshal(ops)
	if err != nil || len(ops) == 0 {
		return nil
	}
	return out
}

func sortedKeys(ms ...map[string]any) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range ms {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	slices.Sort(out)
	return out
}
