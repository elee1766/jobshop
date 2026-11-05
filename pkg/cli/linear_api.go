package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// stderr is swapped by tests.
var stderr io.Writer = os.Stderr

// linearAPICmd is a raw GraphQL call to Linear (#546), as `gh api` is to
// GitHub: the caller's LINEAR_API_KEY if set, else the identity token for
// the team. Mutations ask the guard first; every call is recorded.
type linearAPICmd struct {
	Team  string   `help:"Team key whose identity makes the call (default: from an issue key in the variables)."`
	Field []string `short:"f" sep:"none" help:"String variable name=value (@file reads the file). query=... sets the query."`
	Typed []string `short:"F" name:"raw-field" sep:"none" help:"Typed variable name=json (number, bool, object; @file reads the file)."`
	Input string   `help:"Request body file, - for stdin: a JSON request or bare GraphQL."`
	Query string   `arg:"" optional:"" help:"GraphQL document."`
}

// errGraphQL is a response with errors: exit 1, body already printed.
var errGraphQL = errors.New("linear api: response has errors")

func (c linearAPICmd) Run(g *Globals, l *linearCmd) error {
	query, vars, err := c.request(stdin)
	if err != nil {
		return err
	}
	op, err := parseOperation(query)
	if err != nil {
		return err
	}
	team := c.Team
	if team == "" {
		team = teamFromVars(vars)
	}
	token := os.Getenv("LINEAR_API_KEY")
	if token == "" && team == "" {
		return errors.New("linear api: no LINEAR_API_KEY and no --team (or issue key in variables) to pick an identity")
	}
	action := "linear-api-query"
	if op.kind == "mutation" {
		action = config.ActionLinearAPIMutation
	}
	command := strings.TrimSpace("linear api " + op.kind + " " + op.name)
	record := func(decision string, code int) {
		err := g.call(rpc.MethodActivityRecord, rpc.ActivityRecordParams{
			Action: action, Target: team, Decision: decision, ExitCode: code, Command: command,
		}, &struct{}{})
		if err != nil {
			warn(stderr, "jobshop linear api: activity record: "+err.Error())
		}
	}
	// The guard is about what the agent does, so it runs whoever's key
	// makes the call. No controller, no answer: a write fails closed.
	if op.kind == "mutation" {
		vb, _ := json.Marshal(vars)
		if err := checkOut(g, "linear api", l.AllowSecretFindings, outText{"query", query}, outText{"variables", string(vb)}); err != nil {
			record("secret", 1)
			return err
		}
		var r rpc.GuardResult
		detail := strings.TrimSpace(op.name + " " + strings.Join(op.fields, " "))
		if err := g.call(rpc.MethodGuard, rpc.GuardParams{Action: action, Target: team, Detail: detail}, &r); err != nil {
			return fmt.Errorf("linear api: guard %s: %w (not sent)", action, err)
		}
		if r.Decision != rpc.GuardAllow {
			record(r.Decision, 1)
			msg := fmt.Sprintf("linear api: guard %s %s: %s: %s (not sent)", action, team, r.Decision, r.Reason)
			if r.Ticket != "" {
				msg += "; waiting on " + r.Ticket
			}
			return errors.New(msg)
		}
	}
	if token == "" {
		var r rpc.IdentityTokenResult
		if err := g.call(rpc.MethodIdentityToken, rpc.IdentityParams{Target: team}, &r); err != nil {
			return fmt.Errorf("linear api: identity for %s: %w", team, err)
		}
		token = r.Token
	}
	body, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	out, err := linear.New(l.APIURL, token).Raw(ctx, body)
	if len(out) > 0 {
		_, _ = stdout.Write(out)
		if out[len(out)-1] != '\n' {
			_, _ = io.WriteString(stdout, "\n")
		}
	}
	if err != nil {
		record(rpc.GuardAllow, 1)
		return err
	}
	var r struct {
		Errors json.RawMessage `json:"errors"`
	}
	if json.Unmarshal(out, &r) == nil && len(r.Errors) > 0 && string(r.Errors) != "null" && string(r.Errors) != "[]" {
		record(rpc.GuardAllow, 1)
		warn(stderr, string(r.Errors))
		return errGraphQL
	}
	record(rpc.GuardAllow, 0)
	return nil
}

// request assembles the query and variables from --input, the argument
// and -f/-F.
func (c linearAPICmd) request(in io.Reader) (string, map[string]any, error) {
	query := c.Query
	vars := map[string]any{}
	if c.Input != "" {
		var b []byte
		var err error
		if c.Input == "-" {
			b, err = io.ReadAll(in)
		} else {
			b, err = os.ReadFile(c.Input)
		}
		if err != nil {
			return "", nil, fmt.Errorf("linear api: input: %w", err)
		}
		if isJSONRequest(strings.TrimLeftFunc(string(b), unicode.IsSpace)) {
			var req struct {
				Query     string         `json:"query"`
				Variables map[string]any `json:"variables"`
			}
			if err := json.Unmarshal(b, &req); err != nil {
				return "", nil, fmt.Errorf("linear api: input: %w", err)
			}
			if query != "" && req.Query != "" {
				return "", nil, errors.New("linear api: query given twice (argument and --input)")
			}
			if req.Query != "" {
				query = req.Query
			}
			for k, v := range req.Variables {
				vars[k] = v
			}
		} else {
			if query != "" {
				return "", nil, errors.New("linear api: query given twice (argument and --input)")
			}
			query = string(b)
		}
	}
	for _, f := range c.Field {
		k, v, err := fieldValue(f)
		if err != nil {
			return "", nil, err
		}
		if k == "query" {
			if query != "" {
				return "", nil, errors.New("linear api: query given twice")
			}
			query = v
			continue
		}
		vars[k] = v
	}
	for _, f := range c.Typed {
		k, v, err := fieldValue(f)
		if err != nil {
			return "", nil, err
		}
		var j any
		if err := json.Unmarshal([]byte(v), &j); err != nil {
			return "", nil, fmt.Errorf("linear api: -F %s: not JSON: %w", k, err)
		}
		vars[k] = j
	}
	if strings.TrimSpace(query) == "" {
		return "", nil, errors.New("linear api: no query (argument, -f query=..., or --input)")
	}
	return query, vars, nil
}

// isJSONRequest tells a JSON request from bare GraphQL: both can start
// with "{", but only JSON has a string key next.
func isJSONRequest(t string) bool {
	if !strings.HasPrefix(t, "{") {
		return false
	}
	return strings.HasPrefix(strings.TrimLeftFunc(t[1:], unicode.IsSpace), "\"")
}

// fieldValue splits name=value; @file reads value from a file.
func fieldValue(f string) (string, string, error) {
	k, v, ok := strings.Cut(f, "=")
	if !ok || k == "" {
		return "", "", fmt.Errorf("linear api: field %q is not name=value", f)
	}
	if path, ok := strings.CutPrefix(v, "@"); ok {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", "", fmt.Errorf("linear api: field %s: %w", k, err)
		}
		v = strings.TrimSuffix(string(b), "\n")
	}
	return k, v, nil
}

// teamFromVars is the team of the first issue key in the variables,
// visiting names in sorted order so the pick does not depend on map
// order.
func teamFromVars(v any) string {
	switch x := v.(type) {
	case string:
		if t, ok := linear.Team(x); ok {
			return t
		}
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			if t := teamFromVars(x[k]); t != "" {
				return t
			}
		}
	case []any:
		for _, e := range x {
			if t := teamFromVars(e); t != "" {
				return t
			}
		}
	}
	return ""
}

// operation is what the guard and the activity feed see of a document.
type operation struct {
	kind   string // query or mutation
	name   string
	fields []string // top-level fields of the operation
}

// parseOperation finds the first operation of a document with a small
// scan, not a full GraphQL parser: comments and strings are skipped, a
// leading "{" is an anonymous query, and fragments before it are passed.
func parseOperation(doc string) (operation, error) {
	toks := gqlTokens(doc)
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch t {
		case "{":
			return operation{kind: "query", fields: topFields(toks[i:])}, nil
		case "query", "mutation":
			op := operation{kind: t}
			if i+1 < len(toks) && isName(toks[i+1]) {
				op.name = toks[i+1]
			}
			for j := i + 1; j < len(toks); j++ {
				if toks[j] == "(" {
					j = skipGroup(toks, j, "(", ")")
					continue
				}
				if toks[j] == "{" {
					op.fields = topFields(toks[j:])
					break
				}
			}
			return op, nil
		case "subscription":
			return operation{}, errors.New("linear api: subscription is not supported")
		case "fragment":
			// fragment Name on Type { ... }: skip to the end of its body.
			for j := i + 1; j < len(toks); j++ {
				if toks[j] == "{" {
					i = skipGroup(toks, j, "{", "}")
					break
				}
			}
		default:
			return operation{}, fmt.Errorf("linear api: unexpected %q at the top of the document", t)
		}
	}
	return operation{}, errors.New("linear api: no operation in the document")
}

// topFields lists the field names directly in the selection set that
// starts at toks[0] ("{"): aliases give the real field, arguments and
// nested sets are skipped.
func topFields(toks []string) []string {
	var out []string
	for i := 1; i < len(toks); i++ {
		switch t := toks[i]; {
		case t == "}":
			return out
		case t == "(":
			i = skipGroup(toks, i, "(", ")")
		case t == "{":
			i = skipGroup(toks, i, "{", "}")
		case t == "...":
			i++ // fragment spread or inline fragment name
		case isName(t):
			if i+2 < len(toks) && toks[i+1] == ":" {
				i += 2
				t = toks[i]
			}
			out = append(out, t)
		}
	}
	return out
}

// skipGroup returns the index of the close matching toks[i] (open).
func skipGroup(toks []string, i int, open, end string) int {
	depth := 0
	for ; i < len(toks); i++ {
		switch toks[i] {
		case open:
			depth++
		case end:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return len(toks) - 1
}

func isName(t string) bool {
	if t == "" {
		return false
	}
	for i, r := range t {
		if r != '_' && !unicode.IsLetter(r) && (i == 0 || !unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

// gqlTokens splits a document into names and punctuation, dropping
// whitespace, commas, comments and string literals (values do not matter
// to the guard).
func gqlTokens(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '#':
			for i < len(s) && s[i] != '\n' {
				i++
			}
		case c == '"':
			if strings.HasPrefix(s[i:], `"""`) {
				end := strings.Index(s[i+3:], `"""`)
				if end < 0 {
					return out
				}
				i += end + 6
				continue
			}
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' {
					i++
				}
				i++
			}
			i++
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',':
			i++
		case strings.HasPrefix(s[i:], "..."):
			out = append(out, "...")
			i += 3
		case c == '_' || c == '$' || unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '-' || c == '.':
			j := i + 1
			for j < len(s) && (s[j] == '_' || unicode.IsLetter(rune(s[j])) || unicode.IsDigit(rune(s[j])) || s[j] == '.') {
				j++
			}
			out = append(out, s[i:j])
			i = j
		default:
			out = append(out, string(c))
			i++
		}
	}
	return out
}
