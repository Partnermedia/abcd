// Command abcd-gen-surface writes the committed compatibility snapshot from the
// abcd command tree and the plugin manifests. It is the write half of the
// drift-checked surface: `go generate ./internal/surface/cli` runs it to refresh
// .abcd/development/release/surface.json, and a test
// (internal/surface/cli/surface_test.go) fails the build if the committed
// snapshot ever diverges from the tree. It holds no rendering logic of its own —
// the deterministic walk and encoding live in cli.GenerateSurface, so the
// generator and the drift test render byte-for-byte identically.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"github.com/REPPL/abcd-cli/internal/surface/cli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "abcd-gen-surface:", err)
		os.Exit(1)
	}
	fmt.Println("wrote", cli.SurfaceSnapshotPath)
}

// run does the whole job so every failure path returns an error to one reporter,
// rather than each step repeating the exit dance.
func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := fsutil.ModuleRoot(cwd)
	if err != nil {
		return err
	}
	data, err := cli.GenerateSurface(root)
	if err != nil {
		return err
	}
	dest := filepath.Join(root, filepath.FromSlash(cli.SurfaceSnapshotPath))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}
