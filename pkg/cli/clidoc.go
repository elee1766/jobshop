package cli

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
)

//go:generate env JOBSHOP_UPDATE_CLI_DOC=1 go test -run TestCLIDoc .

// newParser builds the kong parser for the command tree. Main and the
// llm/cli.md generator share it so the doc matches what --help prints.
func newParser(r *root, exit func(int), opts ...kong.Option) (*kong.Kong, error) {
	return kong.New(r, append([]kong.Option{
		kong.Name("jobshop"),
		kong.Description("Multi-agent work orchestration."),
		kong.UsageOnError(),
		kong.Exit(exit),
	}, opts...)...)
}

// cliDoc renders `jobshop --help` and the help of every visible
// subcommand as markdown, for llm/cli.md (#551).
func cliDoc() (string, error) {
	var r root
	p, err := newParser(&r, func(int) {})
	if err != nil {
		return "", err
	}
	var paths [][]string
	var walk func(n *kong.Node, path []string)
	walk = func(n *kong.Node, path []string) {
		paths = append(paths, path)
		for _, c := range n.Children {
			if c.Type != kong.CommandNode || c.Hidden {
				continue
			}
			walk(c, append(append([]string(nil), path...), c.Name))
		}
	}
	walk(p.Model.Node, nil)

	var b strings.Builder
	b.WriteString("# jobshop CLI\n\nGenerated from the kong help by `go generate ./pkg/cli`. Do not edit.\n")
	// Each help needs a fresh parser, and building one is the slow part:
	// render them in parallel (JOB-16), then join in walk order.
	helps := make([]string, len(paths))
	errs := make([]error, len(paths))
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	for i, path := range paths {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			helps[i], errs[i] = helpText(path)
		}()
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return "", err
	}
	for i, path := range paths {
		title := strings.Join(append([]string{"jobshop"}, path...), " ")
		fmt.Fprintf(&b, "\n## %s\n\n```\n%s```\n", title, helps[i])
	}
	return b.String(), nil
}

func helpText(path []string) (string, error) {
	var buf bytes.Buffer
	var hr root
	hp, err := newParser(&hr, func(int) {}, kong.Writers(&buf, &buf))
	if err != nil {
		return "", err
	}
	// --help prints, then calls Exit, a no-op here, so Parse may go on
	// to fail on a missing argument; the help is already written.
	_, _ = hp.Parse(append(append([]string(nil), path...), "--help"))
	return buf.String(), nil
}
