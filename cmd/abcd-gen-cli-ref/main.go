// Command abcd-gen-cli-ref writes the generated CLI reference page from the abcd
// command tree. It is the write half of the drift-checked reference: `go generate
// ./internal/surface/cli` runs it to refresh docs/reference/cli/commands.md, and a
// test (internal/surface/cli/reference_test.go) fails the build if the committed
// page ever diverges from the tree. It holds no rendering logic of its own — the
// deterministic walk lives in cli.GenerateReference, so the generator and the
// drift test render byte-for-byte identically.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"github.com/REPPL/abcd-cli/internal/surface/cli"
)

func main() {
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "abcd-gen-cli-ref:", err)
		os.Exit(1)
	}
	dest := filepath.Join(root, filepath.FromSlash(cli.ReferencePagePath))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "abcd-gen-cli-ref:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(dest, []byte(cli.GenerateReference()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "abcd-gen-cli-ref:", err)
		os.Exit(1)
	}
	fmt.Println("wrote", cli.ReferencePagePath)
}

// moduleRoot resolves the module root from the working directory, so the
// generator writes to the same page whether it is invoked via `go generate` (cwd
// is the package dir) or directly from the repo root. The walk itself lives in
// fsutil, shared with abcd-gen-surface: two generators anchoring on go.mod by
// two separate walks is one walk too many, and a divergence between them would
// send the two drift-checked artefacts to different directories.
func moduleRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fsutil.ModuleRoot(cwd)
}
