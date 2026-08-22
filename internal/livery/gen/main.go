// Command gen writes the committed SVG identity assets from the canonical
// livery grids. It is invoked via `go generate ./internal/livery/...`, which
// runs it with the package directory as the working directory; the drift gate
// (TestSVGAssetsInSync) fails the build when its output and the grids
// disagree, so this is the only sanctioned way to touch the assets.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Partnermedia/abcd/internal/livery"
)

const outDir = "../../docs/assets/img/livery"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "livery gen:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, a := range livery.Assets() {
		files := []struct {
			name string
			data []byte
		}{
			{a.Name + ".svg", livery.RenderSVG(a, true)},
			{a.Name + "-transparent.svg", livery.RenderSVG(a, false)},
			{a.Name + "-square.svg", livery.RenderSVGSquare(a, true)},
			{a.Name + "-square-transparent.svg", livery.RenderSVGSquare(a, false)},
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(outDir, f.name), f.data, 0o644); err != nil {
				return err
			}
			fmt.Println("wrote", filepath.Join(outDir, f.name))
		}
	}
	return nil
}
