// Command abcd is the CLI front door to the abcd engine. It is a thin shell:
// all behaviour lives in internal/core, surfaced here via internal/surface/cli.
package main

import (
	"os"

	"github.com/REPPL/abcd-cli/internal/surface/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
