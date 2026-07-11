// Command record-lint is the deterministic drift gate for the abcd design
// record. It loads .abcd/record-lint.json, lints the markdown record under the
// resolved repo root, prints each finding as `file:line: [SEVERITY RuleID]
// message`, and exits non-zero when any blocker finding exists.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/lint"
)

func main() {
	configPath := flag.String("config", "", "path to record-lint.json (default: <root>/.abcd/record-lint.json)")
	rootPath := flag.String("root", "", "repo root to lint (default: git toplevel, or cwd)")
	releaseGate := flag.String("release-gate", "", "arm the receipt_gate rule for a release: fail closed unless a PROMOTE semantic-pass receipt exists for this commit sha (release-time only; a CI workflow supplies the sha)")
	var requireGates multiFlag
	flag.Var(&requireGates, "require-gate", "a required semantic gate name for --release-gate (repeatable); overrides the config list so the workflow, not the in-tree file, is the trust root")
	flag.Parse()

	root := *rootPath
	if root == "" {
		root = resolveRoot()
	}

	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = filepath.Join(root, ".abcd", "record-lint.json")
	}

	cfg, err := lint.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "record-lint: load config:", err)
		os.Exit(2)
	}

	// --release-gate arms the receipt_gate rule from the CLI invocation, so the
	// release-time decision to gate lives with the CI workflow, not the in-tree
	// (committer-editable) config, which keeps the rule disabled for ordinary runs.
	if *releaseGate != "" {
		cfg = lint.ArmReceiptGate(cfg, *releaseGate, requireGates)
	}

	findings, err := lint.Lint(cfg, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "record-lint:", err)
		os.Exit(2)
	}

	blockers := 0
	for _, f := range findings {
		fmt.Printf("%s:%d: [%s %s] %s\n",
			f.File, f.Line, strings.ToUpper(f.Severity), f.RuleID, f.Message)
		if f.Severity == "blocker" {
			blockers++
		}
	}

	if blockers > 0 {
		os.Exit(1)
	}
}

// multiFlag collects a repeatable string flag (each --require-gate appends).
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

// resolveRoot returns the git toplevel, falling back to the working directory.
func resolveRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		if top := strings.TrimSpace(string(out)); top != "" {
			return top
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}
