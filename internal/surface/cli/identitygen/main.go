// Command identitygen bakes the canonical identity block's title and tagline
// into a generated Go file (itd-112/spc-41). The released binary ships
// without .abcd/**, so the banner cannot read the block at runtime — and
// must not read the cwd's block, which in a foreign repo is someone else's
// identity. Baked at build time the text is trusted-static; the drift test
// in the cli package holds the constants to the block, and the generated
// file is a registered itd-102 positioning surface.
//
// Invoked via `go generate ./internal/surface/cli` (working directory is the
// package directory).
package main

import (
	"fmt"
	"os"

	"github.com/Partnermedia/abcd/internal/core/positioning"
	"github.com/Partnermedia/abcd/internal/surface/cli"
)

const (
	repoRoot = "../../.."
	outFile  = "identity_gen.go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "identitygen:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, _, err := positioning.LoadConfig(repoRoot)
	if err != nil {
		return err
	}
	block, err := positioning.ParseBlock(repoRoot, cfg.Block)
	if err != nil {
		return err
	}
	return os.WriteFile(outFile, cli.IdentityGenSource(block.Title, block.Tagline), 0o644)
}
