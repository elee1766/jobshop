package flaky

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestFailedFromFixture(t *testing.T) {
	f, err := os.Open("testdata/fail.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	got, err := Failed(f)
	if err != nil {
		t.Fatal(err)
	}
	want := []Test{
		{Test: "TestA", Package: "example.com/a"},
		{Test: "TestSub", Package: "example.com/a"},
		{Test: "TestC", Package: "example.com/b"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestClassify(t *testing.T) {
	tc := Test{Test: "TestA", Package: "example.com/a"}
	flakyRun := `{"Action":"fail","Package":"example.com/a","Test":"TestA"}
{"Action":"pass","Package":"example.com/a","Test":"TestA"}
{"Action":"fail","Package":"example.com/a"}
`
	brokenRun := `{"Action":"fail","Package":"example.com/a","Test":"TestA"}
{"Action":"pass","Package":"example.com/a","Test":"TestA/sub"}
{"Action":"fail","Package":"example.com/a","Test":"TestA"}
`
	for in, want := range map[string]string{flakyRun: Flaky, brokenRun: Broken, "": Broken} {
		got, err := Classify(tc, strings.NewReader(in))
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("Classify(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestSummary(t *testing.T) {
	s := Summary([]Test{{Test: "TestA", Package: "p", Verdict: Flaky}})
	if !strings.Contains(s, "| `TestA` | `p` | flaky |") {
		t.Fatal(s)
	}
}
