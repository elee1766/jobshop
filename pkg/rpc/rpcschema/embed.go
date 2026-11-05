package rpcschema

import _ "embed"

// Head is the hand-written part of desk-plugin.d.ts.
//
//go:embed head.d.ts
var Head string

// DTSPath is desk-plugin.d.ts relative to the repo root.
const DTSPath = "ui/desk-plugin.d.ts"
