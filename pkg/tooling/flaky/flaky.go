// Package flaky reads a `go test -json` stream, lists the failed tests and
// classifies each from its reruns as flaky or broken (#510). CI runs it when
// `go test` fails so a red master says whether the test is a flake.
package flaky

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Verdicts written to failed-tests.json.
const (
	Flaky  = "flaky"  // passed on at least one rerun
	Broken = "broken" // failed on every rerun
)

// Test is one failed test.
type Test struct {
	Test    string `json:"test"`
	Package string `json:"package"`
	Verdict string `json:"verdict,omitempty"`
}

type event struct {
	Action  string
	Package string
	Test    string
	Output  string
}

// Print writes the Output text of a `go test -json` stream, which is the
// plain `go test` log, and passes non-JSON lines (build errors) through.
func Print(w io.Writer, r io.Reader) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		var e event
		if len(line) == 0 || line[0] != '{' || json.Unmarshal(line, &e) != nil {
			if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
				return err
			}
			continue
		}
		if _, err := io.WriteString(w, e.Output); err != nil {
			return err
		}
	}
	return sc.Err()
}

// Failed returns the tests with a "fail" action, sorted by package then name.
// A subtest failure also fails its parent; only the top-level test is
// returned, since reruns select it with -run '^Name$'.
func Failed(r io.Reader) ([]Test, error) {
	seen := map[Test]bool{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 || line[0] != '{' {
			continue // build errors and other non-JSON output
		}
		var e event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.Action != "fail" || e.Test == "" {
			continue
		}
		name, _, _ := strings.Cut(e.Test, "/")
		seen[Test{Test: name, Package: e.Package}] = true
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read go test -json: %w", err)
	}
	out := make([]Test, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Package != out[j].Package {
			return out[i].Package < out[j].Package
		}
		return out[i].Test < out[j].Test
	})
	return out, nil
}

// Classify gives the verdict for a test from its rerun stream: broken when
// no run passed, flaky otherwise. With -count=N the stream holds N pass or
// fail actions for the test.
func Classify(t Test, rerun io.Reader) (string, error) {
	sc := bufio.NewScanner(rerun)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var e event
		if json.Unmarshal(line, &e) != nil {
			continue
		}
		if e.Action == "pass" && e.Test == t.Test && e.Package == t.Package {
			return Flaky, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("read rerun output: %w", err)
	}
	return Broken, nil
}

// Summary renders the verdicts as markdown for $GITHUB_STEP_SUMMARY.
func Summary(tests []Test) string {
	var b strings.Builder
	b.WriteString("### Failed tests\n\n")
	if len(tests) == 0 {
		b.WriteString("No failed tests in the go test output (build error or panic outside a test).\n")
		return b.String()
	}
	b.WriteString("| test | package | verdict |\n|---|---|---|\n")
	for _, t := range tests {
		fmt.Fprintf(&b, "| `%s` | `%s` | %s |\n", t.Test, t.Package, t.Verdict)
	}
	return b.String()
}
