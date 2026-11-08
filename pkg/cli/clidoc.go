package cli

import (
	"bytes"
	"fmt"
	"strings"

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
	// One parser for every help: trace the command's path and print its
	// help, as --help does. A parser per help (JOB-16) cost 3-8s under
	// -race, most of TestCLIDoc (JOB-117).
	for _, path := range paths {
		help, err := traceHelp(p, path)
		if err != nil {
			return "", err
		}
		title := strings.Join(append([]string{"jobshop"}, path...), " ")
		fmt.Fprintf(&b, "\n## %s\n\n```\n%s```\n", title, help)
	}
	return b.String(), nil
}

// helpText is the help of a command path, the text `<path> --help`
// prints.
func helpText(path []string) (string, error) {
	var r root
	p, err := newParser(&r, func(int) {})
	if err != nil {
		return "", err
	}
	return traceHelp(p, path)
}

// traceHelp renders path's help on p without parsing or running it.
func traceHelp(p *kong.Kong, path []string) (string, error) {
	var buf bytes.Buffer
	stdout := p.Stdout
	p.Stdout = &buf
	defer func() { p.Stdout = stdout }()
	ctx, err := kong.Trace(p, path)
	if err != nil {
		return "", err
	}
	if ctx.Error != nil {
		return "", ctx.Error
	}
	if err := ctx.PrintUsage(false); err != nil {
		return "", err
	}
	return buf.String(), nil
}
