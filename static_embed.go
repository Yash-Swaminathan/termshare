package main

import "embed"

// staticFS holds the browser UI so the binary is self-contained and does not
// depend on a sibling static/ directory or the process working directory.
//
//go:embed static
var staticFS embed.FS
