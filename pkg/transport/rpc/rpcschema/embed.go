package rpcschema

import _ "embed"

// Head is the hand-written part of the client's type declarations.
//
//go:embed head.d.ts
var Head string

// DTSPath is the generated RPC declarations relative to the repo root.
const DTSPath = "packages/jobshop-client/rpc.d.ts"
