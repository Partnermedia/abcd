// Command abcd is the CLI front door to the abcd engine. It is a thin shell:
// all behaviour lives in internal/core, surfaced here via internal/surface/cli.
package main

import (
	"fmt"
	"os"

	"github.com/REPPL/abcd-cli/internal/surface/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "abcd:", err)
		os.Exit(1)
	}
}
