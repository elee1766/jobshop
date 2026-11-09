// Command deskgen writes packages/jobshop-client/rpc.d.ts from rpc.Specs. Run it
// from the repo root after changing an rpc type; a test fails when the
// committed file is stale.
package main

import (
	"fmt"
	"os"

	"github.com/elee1766/jobshop/pkg/transport/rpc/rpcschema"
)

func main() {
	if err := os.WriteFile(rpcschema.DTSPath, []byte(rpcschema.TypeScript(rpcschema.Head)), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "deskgen:", err)
		os.Exit(1)
	}
}
