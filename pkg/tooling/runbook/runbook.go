// Package runbook parses docs/verify runbooks and runs their read-only
// steps (#635). A step is a fenced code block. Blocks whose info string
// has the word "readonly" (```sh readonly) are safe to run unattended;
// every other block writes, deploys or needs the operator and is left
// for the operator.
package runbook

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Marker is the info-string word that marks a block read-only.
const Marker = "readonly"

// Step is one fenced code block of a runbook.
type Step struct {
	Line     int    `json:"line"` // 1-based line of the opening fence
	Lang     string `json:"lang,omitempty"`
	ReadOnly bool   `json:"readonly"`
	Script   string `json:"script"`
}

// Parse returns the fenced code blocks of a markdown runbook in order.
func Parse(src []byte) []Step {
	var steps []Step
	var cur *Step
	var fence string
	var body strings.Builder
	sc := bufio.NewScanner(bytes.NewReader(src))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	n := 0
	for sc.Scan() {
		n++
		line := sc.Text()
		trim := strings.TrimLeft(line, " ")
		if cur == nil {
			f := fenceOf(trim)
			if f == "" {
				continue
			}
			fence = f
			info := strings.Fields(strings.TrimPrefix(trim, f))
			cur = &Step{Line: n}
			if len(info) > 0 {
				cur.Lang = info[0]
			}
			for _, w := range info {
				if w == Marker {
					cur.ReadOnly = true
				}
			}
			body.Reset()
			continue
		}
		if strings.HasPrefix(trim, fence) && strings.TrimSpace(strings.TrimLeft(trim, fence[:1])) == "" {
			cur.Script = body.String()
			steps = append(steps, *cur)
			cur = nil
			continue
		}
		body.WriteString(line)
		body.WriteByte('\n')
	}
	return steps
}

func fenceOf(s string) string {
	for _, c := range []byte{'`', '~'} {
		i := 0
		for i < len(s) && s[i] == c {
			i++
		}
		if i >= 3 {
			return s[:i]
		}
	}
	return ""
}

// Result is the outcome of one read-only step.
type Result struct {
	Step
	OK       bool   `json:"ok"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// Report is the outcome of a runbook.
type Report struct {
	File    string   `json:"file"`
	Results []Result `json:"results"`
	Skipped []Step   `json:"skipped"` // left for the operator
}

// OK reports whether every run step passed.
func (r Report) OK() bool {
	for _, s := range r.Results {
		if !s.OK {
			return false
		}
	}
	return true
}

// Options control Run.
type Options struct {
	Dir      string        // working directory for the steps
	Timeout  time.Duration // per step; 0 means 2m
	MaxBytes int           // output kept per step; 0 means 8KiB
}

// Run runs the read-only steps with sh -e in order. A failing step does
// not stop later ones: each is a separate check.
func Run(ctx context.Context, file string, steps []Step, o Options) Report {
	if o.Timeout <= 0 {
		o.Timeout = 2 * time.Minute
	}
	if o.MaxBytes <= 0 {
		o.MaxBytes = 8 << 10
	}
	rep := Report{File: file, Results: []Result{}, Skipped: []Step{}}
	for _, s := range steps {
		if !s.ReadOnly {
			rep.Skipped = append(rep.Skipped, s)
			continue
		}
		rep.Results = append(rep.Results, runStep(ctx, s, o))
	}
	return rep
}

func runStep(ctx context.Context, s Step, o Options) Result {
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-ec", s.Script)
	cmd.Dir = o.Dir
	out, err := cmd.CombinedOutput()
	r := Result{Step: s, OK: err == nil, Output: tail(string(out), o.MaxBytes)}
	if err != nil {
		r.ExitCode = -1
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			r.ExitCode = ee.ExitCode()
		}
		if ctx.Err() != nil {
			r.Error = fmt.Sprintf("timed out after %s", o.Timeout)
		} else {
			r.Error = err.Error()
		}
	}
	return r
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "...\n" + s[len(s)-n:]
}

// Markdown renders the report as an issue comment.
func (r Report) Markdown() string {
	var b strings.Builder
	pass := 0
	for _, s := range r.Results {
		if s.OK {
			pass++
		}
	}
	verdict := "passed"
	if !r.OK() {
		verdict = "FAILED"
	}
	fmt.Fprintf(&b, "live-check `%s`: %s (%d/%d read-only steps passed, %d left for the operator)\n",
		r.File, verdict, pass, len(r.Results), len(r.Skipped))
	for _, s := range r.Results {
		mark := "pass"
		if !s.OK {
			mark = fmt.Sprintf("FAIL (exit %d)", s.ExitCode)
		}
		fmt.Fprintf(&b, "\n### line %d: %s\n\n```sh\n%s```\n\n<details><summary>output</summary>\n\n```\n%s\n```\n</details>\n",
			s.Line, mark, s.Script, strings.ReplaceAll(strings.TrimRight(s.Output, "\n"), "```", "'''"))
	}
	if len(r.Skipped) > 0 {
		b.WriteString("\n### Left for the operator\n")
		for _, s := range r.Skipped {
			first, _, _ := strings.Cut(strings.TrimSpace(s.Script), "\n")
			fmt.Fprintf(&b, "- line %d: `%s`\n", s.Line, strings.ReplaceAll(first, "`", "'"))
		}
	}
	return b.String()
}
