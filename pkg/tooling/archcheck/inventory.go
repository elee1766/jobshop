package archcheck

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// WriteInventory prints one block per package: its row in the ownership
// table, its file counts and its edges inside the module and to third-party
// modules. Production and test-only edges are listed apart. The output is
// sorted, so two runs on the same tree are identical and a diff is a change.
func WriteInventory(w io.Writer, g *Graph, cfg Config) error {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	inv := cfg.Inventory(g)
	table := "PACKAGE\tFAMILY\tCHILD\tFILES\tTESTS\tTARGET\n"
	for _, i := range inv {
		table += fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%s\n", i.Pkg, i.Family, i.Child, i.Files, i.TestFiles, i.Target)
	}
	if _, err := io.WriteString(tw, table); err != nil {
		return err
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	b.WriteString("\n")
	for _, i := range inv {
		fmt.Fprintf(&b, "%s\n  prod:      %s\n  test-only: %s\n  external:  %s\n", i.Pkg, list(shorten(i.Prod, "pkg/")), list(shorten(i.TestOnly, "pkg/")), list(i.Third))
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func shorten(in []string, prefix string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.TrimPrefix(s, prefix)
	}
	return out
}

func list(in []string) string {
	if len(in) == 0 {
		return "-"
	}
	return strings.Join(in, " ")
}
