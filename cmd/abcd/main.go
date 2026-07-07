// Command abcd is the CLI front door to the abcd engine. It is a thin shell:
// all behaviour lives in internal/core, surfaced here via internal/surface/cli.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/REPPL/abcd-cli/internal/surface/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		// A command may request a specific exit code (usage errors, the memory
		// lint curator contract). An empty message means it already rendered its
		// output and only the exit code should propagate.
		var coded interface{ ExitCode() int }
		if errors.As(err, &coded) {
			if msg := err.Error(); msg != "" {
				fmt.Fprintln(os.Stderr, "abcd:", msg)
			}
			os.Exit(coded.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "abcd:", err)
		os.Exit(1)
	}
}
