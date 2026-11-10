// Command ci-flaky reruns the tests that failed in a `go test -json` log and
// writes flaky/broken verdicts (#510). Usage:
//
//	go run ./pkg/tooling/flaky/cmd/ci-flaky -in test.json -out failed-tests.json -summary "$GITHUB_STEP_SUMMARY"
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/gfx-labs/jobshop/pkg/tooling/flaky"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ci-flaky:", err)
		os.Exit(1)
	}
}

func run() error {
	in := flag.String("in", "test.json", "go test -json output")
	out := flag.String("out", "failed-tests.json", "verdicts file")
	summary := flag.String("summary", "", "markdown file to append to ($GITHUB_STEP_SUMMARY)")
	count := flag.Int("count", 20, "reruns per test")
	printLog := flag.Bool("print", false, "print the test output text from -in and exit")
	flag.Parse()

	if *printLog {
		f, err := os.Open(*in)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return flaky.Print(os.Stdout, f)
	}

	f, err := os.Open(*in)
	if err != nil {
		return err
	}
	tests, err := flaky.Failed(f)
	_ = f.Close()
	if err != nil {
		return err
	}
	for i, t := range tests {
		var buf bytes.Buffer
		// #nosec G204 -- names come from our own test output.
		cmd := exec.Command("go", "test", "-json", "-race",
			"-gcflags=modernc.org/...=-d=checkptr=0",
			"-gcflags=modernc.org/sqlite/lib=-race=false",
			"-run", "^"+t.Test+"$", fmt.Sprintf("-count=%d", *count), t.Package)
		cmd.Stdout = &buf
		cmd.Stderr = os.Stderr
		_ = cmd.Run() // a non-zero exit is expected for failing reruns
		v, err := flaky.Classify(t, &buf)
		if err != nil {
			return err
		}
		tests[i].Verdict = v
		fmt.Printf("%s %s: %s\n", t.Package, t.Test, v)
	}
	data, err := json.MarshalIndent(tests, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(*out, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if *summary != "" {
		sf, err := os.OpenFile(*summary, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		if _, err := sf.WriteString(flaky.Summary(tests)); err != nil {
			_ = sf.Close()
			return err
		}
		return sf.Close()
	}
	return nil
}
