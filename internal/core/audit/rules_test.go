package audit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/audit"
)

// --- fixture repo builder ---------------------------------------------------

type repoBuilder struct {
	t    *testing.T
	root string
}

// newFixtureRepo starts a git repo with an isolated identity. Chain With* calls
// to lay out tiers, then Commit() to make the tracked set real.
func newFixtureRepo(t *testing.T) *repoBuilder {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	root := t.TempDir()
	git(t, root, "init", "-q")
	return &repoBuilder{t: t, root: root}
}

func git(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func (b *repoBuilder) file(rel, body string) *repoBuilder {
	b.t.Helper()
	p := filepath.Join(b.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		b.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		b.t.Fatal(err)
	}
	return b
}

func (b *repoBuilder) dir(rel string) *repoBuilder {
	b.t.Helper()
	if err := os.MkdirAll(filepath.Join(b.root, filepath.FromSlash(rel)), 0o755); err != nil {
		b.t.Fatal(err)
	}
	return b
}

// conforming lays out a repo that satisfies every v1 rule: all three tiers,
// .work.local gitignored, an AGENTS.md router, and a committed DECISIONS.md.
func (b *repoBuilder) conforming() *repoBuilder {
	return b.
		file(".gitignore", ".abcd/.work.local/\n").
		file(".abcd/development/README.md", "# durable record\n").
		file(".abcd/work/DECISIONS.md", "# decisions\n- 2026-07-13: a thing.\n").
		file(".abcd/.work.local/NEXT.md", "# local handoff\n").
		file("AGENTS.md", "# conventions\n")
}

func (b *repoBuilder) commit() *repoBuilder {
	b.t.Helper()
	git(b.t, b.root, "add", "-A")
	git(b.t, b.root, "-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-q", "-m", "fixture")
	return b
}

func (b *repoBuilder) run() audit.Result {
	b.t.Helper()
	res, err := audit.Evaluate(audit.DefaultRules(), audit.Context{RepoRoot: b.root})
	if err != nil {
		b.t.Fatalf("Evaluate: %v", err)
	}
	return res
}

func findingFor(res audit.Result, ruleID string) *audit.Finding {
	for i := range res.Findings {
		if res.Findings[i].RuleID == ruleID {
			return &res.Findings[i]
		}
	}
	return nil
}

// --- acceptance criteria (itd-85) -------------------------------------------

// AC4: a conforming repo exits 0 with no findings.
func TestAC_ConformingRepoClean(t *testing.T) {
	res := newFixtureRepo(t).conforming().commit().run()
	if len(res.Findings) != 0 {
		t.Fatalf("conforming repo has findings: %+v", res.Findings)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0", res.ExitCode)
	}
}

// AC1: a repo missing .abcd/work/ → three-tier-layout error, exit 2, fix names
// the missing tier.
func TestAC_MissingWorkTier(t *testing.T) {
	b := newFixtureRepo(t).
		file(".gitignore", ".abcd/.work.local/\n").
		file(".abcd/development/README.md", "x\n").
		file(".abcd/.work.local/NEXT.md", "x\n").
		file("AGENTS.md", "x\n").
		commit()
	res := b.run()

	f := findingFor(res, "three-tier-layout")
	if f == nil {
		t.Fatal("no three-tier-layout finding for a repo missing .abcd/work/")
	}
	if f.Severity != audit.SeverityError {
		t.Errorf("severity = %q, want error", f.Severity)
	}
	if !strings.Contains(f.Message+f.Fix, "work") {
		t.Errorf("finding does not name the missing work/ tier: %q / %q", f.Message, f.Fix)
	}
	if res.ExitCode != 2 {
		t.Errorf("exit = %d, want 2", res.ExitCode)
	}
}

// AC2: decisions living only in the gitignored layer → decision-durability warn,
// exit 1 (no errors).
func TestAC_DecisionsOnlyInGitignoredLayer(t *testing.T) {
	b := newFixtureRepo(t).
		file(".gitignore", ".abcd/.work.local/\n").
		file(".abcd/development/README.md", "x\n").
		dir(".abcd/work"). // work/ tier present, but no committed DECISIONS.md
		file(".abcd/work/CONTEXT.md", "x\n").
		file(".abcd/.work.local/DECISIONS.md", "- a decision that will not survive a clone\n").
		file("AGENTS.md", "x\n").
		commit()
	res := b.run()

	f := findingFor(res, "decision-durability")
	if f == nil {
		t.Fatal("no decision-durability finding when decisions are only in the gitignored layer")
	}
	if f.Severity != audit.SeverityWarn {
		t.Errorf("severity = %q, want warn", f.Severity)
	}
	if res.Blockers != 0 {
		t.Errorf("blockers = %d, want 0", res.Blockers)
	}
	if res.ExitCode != 1 {
		t.Errorf("exit = %d, want 1", res.ExitCode)
	}
}

// conventions-router: AGENTS.md absent → error.
func TestRule_ConventionsRouterMissing(t *testing.T) {
	b := newFixtureRepo(t).
		file(".gitignore", ".abcd/.work.local/\n").
		file(".abcd/development/README.md", "x\n").
		file(".abcd/work/DECISIONS.md", "x\n").
		file(".abcd/.work.local/NEXT.md", "x\n").
		commit() // no AGENTS.md
	res := b.run()

	f := findingFor(res, "conventions-router")
	if f == nil {
		t.Fatal("no conventions-router finding when AGENTS.md is absent")
	}
	if f.Severity != audit.SeverityError {
		t.Errorf("severity = %q, want error", f.Severity)
	}
}

// three-tier-layout: .work.local present but NOT gitignored → error (the leak
// the tier exists to prevent). git is available, so this is a real violation,
// not cannot-tell.
func TestRule_WorkLocalNotGitignored(t *testing.T) {
	b := newFixtureRepo(t).
		file(".gitignore", "# .work.local deliberately NOT ignored here\n").
		file(".abcd/development/README.md", "x\n").
		file(".abcd/work/DECISIONS.md", "x\n").
		file(".abcd/.work.local/NEXT.md", "x\n").
		file("AGENTS.md", "x\n").
		commit()
	res := b.run()

	f := findingFor(res, "three-tier-layout")
	if f == nil {
		t.Fatal("no three-tier-layout finding when .work.local is not gitignored")
	}
	if f.Severity != audit.SeverityError {
		t.Errorf("severity = %q, want error", f.Severity)
	}
	if !strings.Contains(strings.ToLower(f.Message+f.Fix), "ignore") {
		t.Errorf("finding does not mention gitignore: %q / %q", f.Message, f.Fix)
	}
}

// AC3: a committed file with an absolute local path → privacy-hygiene error
// citing file:line — unless a waiver escape is on that line.
func TestAC_PrivacyAbsolutePath(t *testing.T) {
	const leak = "see /Users/alice/secret/notes.md for context\n"
	b := newFixtureRepo(t).conforming().
		file("docs/how-to/thing.md", leak).
		commit()
	res := b.run()

	f := findingFor(res, "privacy-hygiene")
	if f == nil {
		t.Fatal("no privacy-hygiene finding for a committed absolute local path")
	}
	if f.Severity != audit.SeverityError {
		t.Errorf("severity = %q, want error", f.Severity)
	}
	if f.File != "docs/how-to/thing.md" || f.Line != 1 {
		t.Errorf("citation = %s:%d, want docs/how-to/thing.md:1", f.File, f.Line)
	}
	if res.ExitCode != 2 {
		t.Errorf("exit = %d, want 2", res.ExitCode)
	}
}

func TestAC_PrivacyWaiverSuppresses(t *testing.T) {
	const waived = "example path /Users/alice/x is illustrative  abcd-audit:allow\n"
	b := newFixtureRepo(t).conforming().
		file("docs/reference/paths.md", waived).
		commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("waiver escape did not suppress the finding: %+v", f)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0 (waived)", res.ExitCode)
	}
}

// AC5: docs/ absent → docs-currency skipped via Where, not failed.
func TestAC_DocsCurrencySkippedWhenNoDocs(t *testing.T) {
	res := newFixtureRepo(t).conforming().commit().run() // conforming() creates no docs/
	for _, f := range res.Findings {
		if f.RuleID == "docs-currency" {
			t.Fatalf("docs-currency produced a finding when docs/ is absent: %+v", f)
		}
	}
	found := false
	for _, id := range res.Skipped {
		if id == "docs-currency" {
			found = true
		}
	}
	if !found {
		t.Errorf("docs-currency not in Skipped when docs/ is absent: skipped=%v", res.Skipped)
	}
}
