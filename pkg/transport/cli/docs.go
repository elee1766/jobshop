package cli

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// docsBase is where `jobshop docs` links point.
const docsBase = "https://github.com/elee1766/jobshop"

// docTopics lists `jobshop docs` topics in print order. Each link is a
// path in the repo; a trailing slash is a directory (tree/, not blob/).
var docTopics = []struct {
	name, help string
	paths      []string
}{
	{"packs", "writing a pack, and the example packs", []string{"docs/packs.md", "packs/"}},
	{"desk-plugins", "Desk API client and migration from plugins", []string{"docs/desk-plugins.md", "packages/jobshop-client/"}},
	{"readme", "overview, goals and example flows", []string{"README.md"}},
	{"llm", "entry point for agents: what to read, the full CLI", []string{"llm/README.md"}},
}

type docsCmd struct {
	Topic string `arg:"" optional:"" help:"Topic (packs, desk-plugins, readme, llm). Empty lists topics."`
}

func (c docsCmd) Run() error {
	bi, _ := debug.ReadBuildInfo()
	out, err := docsText(c.Topic, docsRev(commit, bi))
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, out)
	return err
}

// docsRev is the commit the links point at: the binary's own, so docs
// match the binary, else master. A dirty or unknown build uses master.
func docsRev(c string, bi *debug.BuildInfo) string {
	_, cm := versionInfo("dev", c, bi)
	if cm == "unknown" || cm == "" || strings.Contains(cm, "dirty") {
		return "master"
	}
	return cm
}

func docsText(topic, rev string) (string, error) {
	var b strings.Builder
	if topic == "" {
		for _, t := range docTopics {
			fmt.Fprintf(&b, "%-13s %s\n", t.name, t.help)
		}
		return b.String(), nil
	}
	for _, t := range docTopics {
		if t.name != topic {
			continue
		}
		for _, p := range t.paths {
			kind := "blob"
			if strings.HasSuffix(p, "/") {
				kind = "tree"
				p = strings.TrimSuffix(p, "/")
			}
			fmt.Fprintf(&b, "%s/%s/%s/%s\n", docsBase, kind, rev, p)
		}
		return b.String(), nil
	}
	return "", fmt.Errorf("unknown topic %q (run jobshop docs for the list)", topic)
}

// initHint is printed after `jobshop init` creates a root.
func initHint(dir string) string {
	return "next: write a pack in " + filepath.Join(dir, "packs", "<name>") +
		" and list it under packs: in pack.yaml\n" +
		"      jobshop docs packs   # guide and example packs\n"
}
