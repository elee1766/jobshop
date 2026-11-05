package runbook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixture = "# Fixture\n\n" +
	"```sh readonly\necho hello > /dev/null; echo ok\n```\n\n" +
	"```sh\ntouch written\n```\n\n" +
	"  ~~~bash readonly\n  exit 3\n  ~~~\n\n" +
	"````sh readonly\necho '```'\n````\n"

func TestRunOnlyReadOnly(t *testing.T) {
	dir := t.TempDir()
	steps := Parse([]byte(fixture))
	if len(steps) != 4 {
		t.Fatalf("parsed %d steps, want 4: %+v", len(steps), steps)
	}
	rep := Run(context.Background(), "fixture.md", steps, Options{Dir: dir})
	if _, err := os.Stat(filepath.Join(dir, "written")); err == nil {
		t.Fatal("unmarked step ran")
	}
	if len(rep.Results) != 3 || len(rep.Skipped) != 1 || rep.Skipped[0].Line != 7 {
		t.Fatalf("results %d skipped %+v", len(rep.Results), rep.Skipped)
	}
	if !rep.Results[0].OK || rep.Results[0].Output != "ok\n" {
		t.Fatalf("step 1: %+v", rep.Results[0])
	}
	if rep.Results[1].OK || rep.Results[1].ExitCode != 3 {
		t.Fatalf("step 2 should fail with 3: %+v", rep.Results[1])
	}
	if !rep.Results[2].OK || rep.Results[2].Output != "```\n" {
		t.Fatalf("step 3: %+v", rep.Results[2])
	}
	if rep.OK() {
		t.Fatal("report OK with a failing step")
	}
	md := rep.Markdown()
	for _, want := range []string{"FAILED (2/3", "FAIL (exit 3)", "Left for the operator", "line 7: `touch written`"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q:\n%s", want, md)
		}
	}
}

func TestAllPass(t *testing.T) {
	rep := Run(context.Background(), "f", Parse([]byte("```sh readonly\ntrue\n```\n")), Options{Dir: t.TempDir()})
	if !rep.OK() || len(rep.Skipped) != 0 {
		t.Fatalf("%+v", rep)
	}
}

// TestVerifyDocsParse checks every runbook parses with closed fences and
// marked blocks are shell.
func TestVerifyDocsParse(t *testing.T) {
	files, _ := filepath.Glob("../../docs/verify/*.md")
	if len(files) == 0 {
		t.Skip("no docs/verify")
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range Parse(b) {
			if s.ReadOnly && s.Lang != "sh" && s.Lang != "bash" {
				t.Errorf("%s:%d: readonly block must be sh or bash, got %q", f, s.Line, s.Lang)
			}
		}
	}
}
