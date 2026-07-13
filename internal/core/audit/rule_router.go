package audit

import (
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// conventionsRouter checks that an AGENTS.md conventions router is present at the
// repo root. CLAUDE.md / GEMINI.md are host-specific bridges that may accompany
// it, but they do not substitute for it: AGENTS.md is the host-agnostic canonical
// router every surface reads first.
type conventionsRouter struct{}

func (conventionsRouter) Meta() RuleMeta {
	return RuleMeta{
		ID:         "conventions-router",
		Severity:   SeverityError,
		Fix:        "add an AGENTS.md at the repo root declaring the working conventions",
		PolicyInfo: "AGENTS.md is the host-agnostic router that tells any agent how to work in this repo; a host-specific bridge (CLAUDE.md/GEMINI.md) may point to it but does not replace it",
	}
}

func (conventionsRouter) Where(Context) bool { return true }

func (conventionsRouter) Eval(ctx Context) ([]Finding, error) {
	// The router is a file: a directory named AGENTS.md does not satisfy it.
	path := filepath.Join(ctx.RepoRoot, "AGENTS.md")
	present, err := fsutil.Exists(path)
	if err != nil {
		return nil, err
	}
	if present {
		isDir, err := fsutil.IsDir(path)
		if err != nil {
			return nil, err
		}
		if !isDir {
			return nil, nil // a real router file
		}
	}
	return []Finding{{
		RuleID:   "conventions-router",
		Severity: SeverityError,
		File:     "AGENTS.md",
		Message:  "no AGENTS.md conventions router file at the repo root",
	}}, nil
}
