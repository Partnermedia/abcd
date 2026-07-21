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

	"github.com/REPPL/abcd-cli/internal/surface/cli"
)

func main() {
	root, err := repoRoot()
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

// repoRoot walks up from the working directory to the module root (the directory
// holding go.mod), so the generator writes to the same page whether it is invoked
// via `go generate` (cwd is the package dir) or directly from the repo root.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}
