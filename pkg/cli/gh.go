package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// `jobshop gh` (#349) runs the real gh with the identity token for the
// repo it acts on, and asks the guardrail hooks before any command that
// writes to GitHub. It replaces the shell gh wrapper from #216. Main runs
// it before kong, so every gh flag (including --help) reaches gh as is.

// ghWrite is one mutating gh command and the guard action it maps to.
type ghWrite struct {
	cmd    []string // gh subcommand words, e.g. {"pr", "create"}
	action string
}

// ghWrites lists the commands that write to GitHub. Anything else is
// read-only and runs without a guard. Action names match the on: action
// names hooks use (pkg/config/hook.go).
var ghWrites = []ghWrite{
	{[]string{"pr", "create"}, "pr-create"},
	{[]string{"pr", "merge"}, "merge"},
	{[]string{"pr", "comment"}, "comment"},
	{[]string{"pr", "review"}, "review"},
	{[]string{"pr", "edit"}, "pr-edit"},
	{[]string{"pr", "close"}, "pr-close"},
	{[]string{"pr", "reopen"}, "pr-reopen"},
	{[]string{"pr", "ready"}, "pr-edit"},
	{[]string{"issue", "create"}, "issue-create"},
	{[]string{"issue", "edit"}, "issue-edit"},
	{[]string{"issue", "comment"}, "comment"},
	{[]string{"issue", "close"}, "issue-close"},
	{[]string{"issue", "reopen"}, "issue-reopen"},
	{[]string{"release", "create"}, "release-create"},
	{[]string{"repo", "create"}, "repo-write"},
	{[]string{"repo", "edit"}, "repo-write"},
	{[]string{"repo", "delete"}, "repo-write"},
	{[]string{"repo", "rename"}, "repo-write"},
	{[]string{"repo", "archive"}, "repo-write"},
	{[]string{"repo", "unarchive"}, "repo-write"},
	{[]string{"repo", "fork"}, "repo-write"},
	{[]string{"repo", "sync"}, "repo-write"},
	{[]string{"repo", "set-default"}, "repo-write"},
	{[]string{"repo", "deploy-key", "add"}, "repo-write"},
	{[]string{"repo", "deploy-key", "delete"}, "repo-write"},
}

// ghFlagsWithValue are gh flags that take a separate value, so the
// word after them is not a subcommand or positional argument. Short
// flags that are booleans in some subcommand (pr merge -m/-r, pr review
// -a/-c, pr create -f) are left out: the cost of missing one is a missed
// PR number, while eating a word could hide the subcommand.
var ghFlagsWithValue = map[string]bool{
	"-R": true, "--repo": true, "-b": true, "--body": true, "-F": true, "--body-file": true,
	"-t": true, "--title": true, "-B": true, "--base": true, "-H": true, "--head": true,
	"-X": true, "--method": true, "--raw-field": true, "--field": true,
	"--input": true, "--header": true, "--assignee": true,
	"-l": true, "--label": true, "--milestone": true, "--reviewer": true,
	"--project": true, "-q": true, "--jq": true, "--template": true,
	"--hostname": true, "--subject": true, "--match-head-commit": true, "--author-email": true,
	"--add-label": true, "--remove-label": true, "--add-assignee": true, "--remove-assignee": true,
	"--add-reviewer": true, "--remove-reviewer": true, "--add-project": true, "--remove-project": true,
	"--notes": true, "-n": true, "--notes-file": true, "--target": true, "--cache": true,
	"--preview": true, "--comment": true,
}

// ghCall is what jobshop gh learned from the arguments.
type ghCall struct {
	action   string // guard action; "" for read-only
	repo     string // owner/name, "" if not given on the command line
	number   string // PR or issue number, "" if none
	body     string // -b/--body text, for the guard's detail
	bodyFile string // -F/--body-file for pr/issue, read for detail
	method   string // gh api HTTP method, upper case
	apiPath  string // gh api endpoint
}

// words splits args into positional words, skipping flags and their
// values, and collects the flag values jobshop gh needs.
func ghWords(args []string) (words []string, flags map[string][]string) {
	flags = map[string][]string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			words = append(words, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			words = append(words, a)
			continue
		}
		if k, v, ok := strings.Cut(a, "="); ok && strings.HasPrefix(a, "--") {
			flags[k] = append(flags[k], v)
			continue
		}
		// -Rowner/name and -XPOST: a short flag with its value attached.
		if len(a) > 2 && a[1] != '-' && ghFlagsWithValue[a[:2]] {
			flags[a[:2]] = append(flags[a[:2]], a[2:])
			continue
		}
		if ghFlagsWithValue[a] && i+1 < len(args) {
			flags[a] = append(flags[a], args[i+1])
			i++
			continue
		}
		flags[a] = append(flags[a], "")
	}
	return words, flags
}

func flagValue(flags map[string][]string, names ...string) (string, bool) {
	for _, n := range names {
		if v, ok := flags[n]; ok && len(v) > 0 {
			return v[len(v)-1], true
		}
	}
	return "", false
}

var ghNumber = regexp.MustCompile(`^(?:#?(\d+)|https://github\.com/[^/]+/[^/]+/(?:pull|issues)/(\d+)\S*)$`)

// classifyGH reads gh arguments. A command is mutating when it is in
// ghWrites, or is `api` with a method other than GET (explicit, or POST
// implied by -f/-F/--field/--raw-field/--input).
func classifyGH(args []string) ghCall {
	words, flags := ghWords(args)
	var c ghCall
	c.repo, _ = flagValue(flags, "-R", "--repo")
	c.repo = normRepo(c.repo)
	if len(words) == 0 {
		return c
	}
	if words[0] == "api" {
		method, ok := flagValue(flags, "-X", "--method")
		if !ok {
			method = "GET"
			for _, f := range []string{"-f", "-F", "--field", "--raw-field", "--input"} {
				if _, has := flags[f]; has {
					method = "POST"
				}
			}
		}
		c.method = strings.ToUpper(method)
		if aw := apiWords(args); len(aw) > 1 {
			c.apiPath = aw[1]
		}
		// The repo picks the identity token, so a GET needs it too: the
		// controller's waits run gh api outside any clone (#686).
		if c.repo == "" {
			c.repo = apiRepo(c.apiPath)
		}
		if !strings.EqualFold(method, "GET") {
			c.action = "api-" + strings.ToLower(method)
		}
		return c
	}
	for _, w := range ghWrites {
		if len(words) >= len(w.cmd) && slices.Equal(words[:len(w.cmd)], w.cmd) {
			c.action = w.action
			rest := words[len(w.cmd):]
			if len(rest) > 0 {
				if m := ghNumber.FindStringSubmatch(rest[0]); m != nil {
					c.number = m[1] + m[2]
					if c.repo == "" && strings.HasPrefix(rest[0], "https://") {
						c.repo = normRepo(rest[0])
					}
				}
			}
			c.body, _ = flagValue(flags, "-b", "--body")
			if words[0] == "pr" || words[0] == "issue" {
				// -F is --body-file for pr and issue (it is a field for api).
				c.bodyFile, _ = flagValue(flags, "-F", "--body-file")
			}
			return c
		}
	}
	return c
}

// apiRepo is owner/name from an api path like repos/o/n/issues/1.
func apiRepo(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "repos" {
		return parts[1] + "/" + parts[2]
	}
	return ""
}

var repoURL = regexp.MustCompile(`github\.com[:/]+([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(?:\.git)?(?:/.*)?$`)

// normRepo turns owner/name, a github.com URL or HOST/owner/name into
// owner/name; "" if it is none of those.
func normRepo(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if m := repoURL.FindStringSubmatch(s); m != nil {
		return m[1] + "/" + m[2]
	}
	parts := strings.Split(s, "/")
	switch len(parts) {
	case 2:
		return s
	case 3: // HOST/owner/name
		return parts[1] + "/" + parts[2]
	}
	return ""
}

// originRepo is owner/name of the cwd's git origin, "" if it is not a
// GitHub remote.
func originRepo() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return normRepo(string(out))
}

// realGH finds the gh to run: JOBSHOP_GH, else the first gh on PATH that
// is not this binary, so a gh symlink to jobshop does not recurse.
func realGH() (string, error) {
	if p := os.Getenv("JOBSHOP_GH"); p != "" {
		return p, nil
	}
	self, _ := os.Executable()
	self, _ = filepath.EvalSymlinks(self)
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		p := filepath.Join(dir, "gh")
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() || fi.Mode()&0o111 == 0 {
			continue
		}
		if r, err := filepath.EvalSymlinks(p); err == nil && r == self {
			continue
		}
		return p, nil
	}
	return "", errors.New("jobshop gh: no gh found: set JOBSHOP_GH or put gh on PATH")
}

// runGH is `jobshop gh args...`; it returns the process exit code.
func runGH(g *Globals, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	gh, err := realGH()
	if err != nil {
		warn(stderr, err.Error())
		return 1
	}
	key, args := cutGHKey(args)
	allow, args := cutAllowSecrets(args)
	c := classifyGH(args)
	var extra json.RawMessage // parsed write arguments for hooks (JOB-113)
	if key != "" && c.action == "" {
		warn(stderr, "jobshop gh: "+ghKeyFlag+" is only for writes (not run)")
		return 1
	}
	if c.action != "" {
		// Scan what the write sends before the guard and gh see it (#574).
		texts, in, err := ghOutText(args, stdin)
		if err != nil {
			warn(stderr, "jobshop gh: "+err.Error()+" (not run)")
			return 1
		}
		if in != nil {
			stdin = bytes.NewReader(in)
		}
		extra = buildGHExtra(c, args, in)
		found, err := secretFindings(g, texts)
		if err != nil {
			warn(stderr, "jobshop gh: "+err.Error()+" (not run)")
			return 1
		}
		if found != "" {
			if !allow {
				warn(stderr, "jobshop gh: refused, the text looks like it holds a secret; remove it and retry (not run):\n"+strings.TrimRight(found, "\n"))
				return 1
			}
			if err := allowSecrets(g); err != nil {
				warn(stderr, "jobshop gh: "+err.Error()+" (not run)")
				return 1
			}
		}
	} else if allow {
		warn(stderr, "jobshop gh: "+allowSecretsFlag+" is only for writes (not run)")
		return 1
	}
	repo := c.repo
	if repo == "" {
		repo = originRepo()
	}
	env := os.Environ()
	if os.Getenv("GH_TOKEN") == "" && repo != "" {
		var r rpc.IdentityTokenResult
		if err := g.call(rpc.MethodIdentityToken, rpc.IdentityParams{Target: repo}, &r); err == nil && r.Token != "" {
			env = append(env, "GH_TOKEN="+r.Token)
		}
	}
	target := repo
	if c.number != "" {
		target += "#" + c.number
	}
	if c.action != "" {
		detail := c.body
		if detail == "" && c.bodyFile != "" && c.bodyFile != "-" {
			if b, err := os.ReadFile(c.bodyFile); err == nil {
				detail = string(b)
			}
		}
		var r rpc.GuardResult
		// No controller, no answer: a write fails closed.
		if err := g.call(rpc.MethodGuard, rpc.GuardParams{Action: c.action, Target: target, Detail: detail, Extra: extra}, &r); err != nil {
			warn(stderr, fmt.Sprintf("jobshop gh: guard %s: %v (not run)", c.action, err))
			return 1
		}
		if r.Decision != domain.GuardAllow {
			msg := fmt.Sprintf("jobshop gh: guard %s %s: %s: %s (not run)", c.action, target, r.Decision, r.Reason)
			if r.Ticket != "" {
				msg += "; waiting on " + r.Ticket
			}
			warn(stderr, msg)
			recordGH(g, c.action, repo, target, r.Decision, 1, args, stderr)
			return 1
		}
	}
	if key != "" {
		code := runGHKeyed(g, gh, env, c, repo, key, args, stdin, stdout, stderr)
		recordGH(g, c.action, repo, target, domain.GuardAllow, code, args, stderr)
		return code
	}
	cmd := exec.Command(gh, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.Env = stdin, stdout, stderr, env
	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			warn(stderr, "jobshop gh: "+err.Error())
			return 1
		}
		code = ee.ExitCode()
	}
	if c.action != "" {
		recordGH(g, c.action, repo, target, domain.GuardAllow, code, args, stderr)
	}
	return code
}

// recordGH adds the call to the activity feed. A failure only warns: the
// call already happened (or was refused) and must report its own result.
// Only the command words are kept, not flag values, so a body or token
// passed as a flag does not land in the feed.
func recordGH(g *Globals, action, repo, target, decision string, code int, args []string, stderr io.Writer) {
	words, _ := ghWords(args)
	err := g.call(rpc.MethodActivityRecord, rpc.ActivityRecordParams{
		Action: action, Repo: repo, Target: target, Decision: decision, ExitCode: code,
		Command: "gh " + strings.Join(words, " "),
	}, &struct{}{})
	if err != nil {
		warn(stderr, "jobshop gh: activity record: "+err.Error())
	}
}

// warn writes a line to stderr; there is nowhere to report a failed write.
func warn(w io.Writer, msg string) { _, _ = fmt.Fprintln(w, msg) }
