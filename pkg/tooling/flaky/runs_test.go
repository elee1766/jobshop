package flaky

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// fakeGh serves `gh run list` from testdata/runs.json and copies
// testdata/failed-tests-<id>.json as the artifact; a missing file is a
// run without the artifact.
func fakeGh(_ context.Context, args ...string) ([]byte, error) {
	if args[1] == "list" {
		return os.ReadFile("testdata/runs.json")
	}
	id, dir := args[2], args[len(args)-1]
	b, err := os.ReadFile(filepath.Join("testdata", "failed-tests-"+id+".json"))
	if err != nil {
		return nil, errors.New("no valid artifacts found to download")
	}
	return nil, os.WriteFile(filepath.Join(dir, "failed-tests.json"), b, 0o600)
}

func TestReportFromFixture(t *testing.T) {
	rep, err := GH{Repo: "o/r", Branch: "master", Gh: fakeGh}.Report(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rep.NoArtifact, []int64{3}) {
		t.Fatalf("no_artifact %v", rep.NoArtifact)
	}
	type row struct {
		test, verdict string
		flaky, broken int
		runs          []int64
	}
	var got []row
	for _, g := range rep.Tests {
		got = append(got, row{g.Test, g.Verdict, g.Flaky, g.Broken, g.Runs})
	}
	want := []row{
		{"TestA", Broken, 1, 1, []int64{2, 1}},
		{"TestB", Flaky, 1, 0, []int64{1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}
