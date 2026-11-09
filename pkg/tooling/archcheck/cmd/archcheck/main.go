// Command archcheck checks the module's import graph against the ownership
// table and boundary rules in pkg/tooling/archcheck (JOB-165).
//
//	go run ./pkg/tooling/archcheck/cmd/archcheck           # check, exit 1 on a problem
//	go run ./pkg/tooling/archcheck/cmd/archcheck -inventory  # package table and edges
//	go run ./pkg/tooling/archcheck/cmd/archcheck -why A B    # import chain from A to B
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/elee1766/jobshop/pkg/tooling/archcheck"
)

func main() {
	err := run(os.Args[1:])
	if err == nil {
		return
	}
	if !errors.Is(err, errProblems) {
		fmt.Fprintln(os.Stderr, "archcheck:", err)
	}
	os.Exit(1)
}

// errProblems means the check ran and found problems, which it already printed.
var errProblems = errors.New("boundary problems")

func run(args []string) error {
	fs := flag.NewFlagSet("archcheck", flag.ContinueOnError)
	dir := fs.String("C", ".", "module directory")
	tags := fs.String("tags", "victorialogs", "build tags, so tagged test files are in the graph")
	inventory := fs.Bool("inventory", false, "print every package with its owner and edges")
	why := fs.Bool("why", false, "print the shortest production import chain between two packages")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	g, err := archcheck.Load(ctx, *dir, *tags)
	if err != nil {
		return err
	}
	cfg := archcheck.Jobshop()

	switch {
	case *inventory:
		return archcheck.WriteInventory(os.Stdout, g, cfg)
	case *why:
		if fs.NArg() != 2 {
			return fmt.Errorf("-why needs two packages, such as ./pkg/hooks github.com/elee1766/acphost/pkg/acphost")
		}
		from, to := expand(g, fs.Arg(0)), expand(g, fs.Arg(1))
		chain := archcheck.Why(g, from, to)
		if chain == nil {
			return fmt.Errorf("%s does not reach %s through production imports", from, to)
		}
		for i, p := range chain {
			fmt.Printf("%s%s\n", strings.Repeat("  ", i), g.Short(p))
		}
		return nil
	}

	res := archcheck.Check(g, cfg)
	if len(res.Problems) == 0 {
		fmt.Printf("archcheck: ok, %d packages, %d accepted known violations\n", len(cfg.Owners), len(res.Accepted))
		return nil
	}
	for _, p := range res.Problems {
		fmt.Fprintln(os.Stderr, p)
	}
	fmt.Fprintf(os.Stderr, "archcheck: %d problems. Rules are in pkg/tooling/archcheck/rules.go, the ownership table in jobshop.go.\n", len(res.Problems))
	return errProblems
}

// expand turns ./pkg/x into the full import path.
func expand(g *archcheck.Graph, p string) string {
	if rest, ok := strings.CutPrefix(p, "./"); ok {
		return g.Module + "/" + rest
	}
	return p
}
