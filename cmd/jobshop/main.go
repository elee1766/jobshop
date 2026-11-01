// Command jobshop is a thin entry point. All behavior lives in pkg/cli.
package main

import (
	"os"

	"github.com/elee1766/jobshop/pkg/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
