package cli

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Keyed GitHub writes (#415). GitHub has no idempotency key, so the key
// travels in the posted body as a hidden HTML comment. The local write key
// is a 48h cache: a done key replays its result, and a missing or pending
// one first searches the target for the marker, so a retry after the
// local row is gone finds the first post instead of posting again.

// ghKeyFlag is jobshop's own flag, stripped before gh runs. It is not
// --key so it can never shadow a gh flag.
const ghKeyFlag = "--jobshop-key"

// ghKeyMarker is the hidden marker for a project-scoped key. GitHub does
// not render HTML comments, so readers and the agent guidelines see
// nothing.
func ghKeyMarker(scoped string) string { return "<!-- jobshop-key: " + scoped + " -->" }

// cutGHKey removes --jobshop-key K (or =K) from args.
func cutGHKey(args []string) (key string, rest []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return key, append(rest, args[i:]...)
		}
		if v, ok := strings.CutPrefix(a, ghKeyFlag+"="); ok {
			key = v
			continue
		}
		if a == ghKeyFlag && i+1 < len(args) {
			key = args[i+1]
			i++
			continue
		}
		rest = append(rest, a)
	}
	return key, rest
}

// ghKeyed is a keyed write: where to look for the marker and how.
type ghKeyed struct {
	kind   string // "comment", "review" or "issue"
	repo   string
	number string
	body   string
	args   []string // gh args with the body replaced by body+marker
}

// planGHKeyed checks a keyed call and puts the marker in its body.
func planGHKeyed(c ghCall, repo string, args []string, scoped string, stdin io.Reader) (ghKeyed, error) {
	words, _ := ghWords(args)
	k := ghKeyed{repo: repo, number: c.number}
	switch {
	case len(words) >= 2 && words[1] == "comment" && (words[0] == "pr" || words[0] == "issue"):
		k.kind = "comment"
	case len(words) >= 2 && words[0] == "pr" && words[1] == "review":
		k.kind = "review"
	case len(words) >= 2 && words[0] == "issue" && words[1] == "create":
		k.kind = "issue"
	default:
		return k, fmt.Errorf("%s works with pr|issue comment, pr review and issue create", ghKeyFlag)
	}
	if repo == "" || (k.kind != "issue" && k.number == "") {
		return k, fmt.Errorf("%s needs the repo and the PR or issue number", ghKeyFlag)
	}
	body, rest, err := cutGHBody(args, stdin)
	if err != nil {
		return k, err
	}
	if strings.TrimSpace(body) == "" {
		return k, fmt.Errorf("%s needs a body (-b or -F)", ghKeyFlag)
	}
	k.body = body + "\n\n" + ghKeyMarker(scoped)
	k.args = append(slices.Clone(rest), "--body", k.body)
	return k, nil
}

// cutGHBody removes -b/--body/-F/--body-file from args and returns the
// body text, read from the file or stdin for -F.
func cutGHBody(args []string, stdin io.Reader) (string, []string, error) {
	var body string
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		name, val, hasVal := a, "", false
		if k, v, ok := strings.Cut(a, "="); ok && strings.HasPrefix(a, "--") {
			name, val, hasVal = k, v, true
		} else if len(a) > 2 && (a[:2] == "-b" || a[:2] == "-F") {
			name, val, hasVal = a[:2], a[2:], true
		}
		if !slices.Contains([]string{"-b", "--body", "-F", "--body-file"}, name) {
			rest = append(rest, a)
			continue
		}
		if !hasVal {
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("%s needs a value", name)
			}
			val = args[i+1]
			i++
		}
		if name == "-b" || name == "--body" {
			body = val
			continue
		}
		var b []byte
		var err error
		if val == "-" {
			b, err = io.ReadAll(stdin)
		} else {
			b, err = os.ReadFile(val)
		}
		if err != nil {
			return "", nil, err
		}
		body = string(b)
	}
	return body, rest, nil
}

// findMarker returns the URL of what already carries marker, "" if
// nothing does.
func (k ghKeyed) findMarker(gh string, env []string, marker string) (string, error) {
	jq := fmt.Sprintf(`[.[] | select((.body // "") | contains(%q)) | .html_url] | first // ""`, marker)
	var args []string
	switch k.kind {
	case "comment":
		args = []string{"api", "--paginate", "--slurp", fmt.Sprintf("repos/%s/issues/%s/comments?per_page=100", k.repo, k.number), "--jq", "add // [] | " + jq}
	case "review":
		args = []string{"api", "--paginate", "--slurp", fmt.Sprintf("repos/%s/pulls/%s/reviews?per_page=100", k.repo, k.number), "--jq", "add // [] | " + jq}
	case "issue":
		// Search is eventually consistent: a post seconds old may be
		// missed. A key that young still has its local row.
		args = []string{"api", "-X", "GET", "search/issues", "-f", fmt.Sprintf(`q=repo:%s in:body "%s"`, k.repo, marker), "--jq", ".items | " + jq}
	}
	var out, errb bytes.Buffer
	cmd := exec.Command(gh, args...)
	cmd.Stdout, cmd.Stderr, cmd.Env = &out, &errb, env
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("search for %s: %w: %s", marker, err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// runGHKeyed runs a keyed write after the guard allowed it. It returns
// the exit code.
func runGHKeyed(g *Globals, gh string, env []string, c ghCall, repo, key string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var w store.WriteKey
	if err := g.call(rpc.MethodWriteKeyClaim, rpc.WriteKeyParams{Key: key, Op: "gh." + c.action}, &w); err != nil {
		warn(stderr, "jobshop gh: write key: "+err.Error()+" (not run)")
		return 1
	}
	if w.Done {
		_, _ = fmt.Fprintln(stdout, w.Result)
		return 0
	}
	scoped := cmp.Or(w.Scoped, key)
	k, err := planGHKeyed(c, repo, args, scoped, stdin)
	if err != nil {
		warn(stderr, "jobshop gh: "+err.Error()+" (not run)")
		return 1
	}
	url, err := k.findMarker(gh, env, ghKeyMarker(scoped))
	if err != nil {
		warn(stderr, "jobshop gh: "+err.Error()+" (not run)")
		return 1
	}
	if url == "" {
		var out bytes.Buffer
		cmd := exec.Command(gh, k.args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.Env = nil, io.MultiWriter(stdout, &out), stderr, env
		if err := cmd.Run(); err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				return ee.ExitCode()
			}
			warn(stderr, "jobshop gh: "+err.Error())
			return 1
		}
		url = strings.TrimSpace(out.String())
	} else {
		_, _ = fmt.Fprintln(stdout, url)
	}
	if url == "" {
		url = "posted" // finish needs a result; gh pr review prints nothing
	}
	if err := g.call(rpc.MethodWriteKeyFinish, rpc.WriteKeyParams{Key: key, Result: url}, &w); err != nil {
		warn(stderr, "jobshop gh: write key finish: "+err.Error())
	}
	return 0
}
