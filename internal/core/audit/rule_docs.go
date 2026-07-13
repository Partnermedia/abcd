package audit

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/core/lint"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// docsCurrency reuses the docs-lint engine to surface documentation drift —
// change-narration tense, broken relative links, stray root markdown. It is
// Where-gated on docs/ existing (a repo with no user-facing docs cannot drift),
// so an absent docs/ skips the rule rather than failing it.
//
// Every finding is emitted at warn severity regardless of the underlying
// docs-lint severity: audit is an advisory conformance surface, and the
// authoritative docs gate is `abcd docs lint` itself (which still exits 2 on a
// blocker). Re-raising a docs blocker as an audit error would double-gate the
// same check. (Recorded in DECISIONS.md.)
type docsCurrency struct{}

func (docsCurrency) Meta() RuleMeta {
	return RuleMeta{
		ID:         "docs-currency",
		Severity:   SeverityWarn,
		Fix:        "run `abcd docs lint` and resolve the drift it reports",
		PolicyInfo: "docs describe what IS; change-narration, broken links, and stray root markdown are the drift signals the docs-lint engine already checks",
	}
}

// Where: only when docs/ exists.
func (docsCurrency) Where(ctx Context) bool {
	isDir, err := fsutil.IsDir(filepath.Join(ctx.RepoRoot, "docs"))
	return err == nil && isDir
}

func (docsCurrency) Eval(ctx Context) ([]Finding, error) {
	cfgPath := filepath.Join(ctx.RepoRoot, ".abcd", "docs-lint.json")
	// The engine reuse needs the repo's docs-lint config. A repo with docs/ but
	// no committed config cannot be linted; degrade to no findings rather than
	// fail — the config is what defines the checks, and its absence is a
	// prepare-this-repo gap, not a docs-drift violation.
	if present, err := fsutil.Exists(cfgPath); err != nil {
		return nil, err
	} else if !present {
		return nil, nil
	}

	cfg, err := lint.LoadConfig(cfgPath)
	if err != nil {
		// A malformed config is a real problem, but it is the docs-lint surface's
		// to report, not audit's — surface it as a single warn pointer without
		// leaking the underlying path error.
		return []Finding{{
			RuleID:   "docs-currency",
			Severity: SeverityWarn,
			File:     ".abcd/docs-lint.json",
			Message:  "docs-lint config could not be loaded: " + cleanErr(err),
		}}, nil
	}

	findings, err := lint.Lint(cfg, ctx.RepoRoot)
	if err != nil {
		return nil, err
	}
	out := make([]Finding, 0, len(findings))
	for _, f := range findings {
		out = append(out, Finding{
			RuleID:   "docs-currency",
			Severity: SeverityWarn, // advisory; the docs gate is `abcd docs lint`
			File:     f.File,
			Line:     f.Line,
			Message:  f.Message,
		})
	}
	return out, nil
}

// cleanErr returns an error's message with any *os.PathError path stripped, so a
// config-load failure never leaks an absolute path into a finding.
func cleanErr(err error) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
}
