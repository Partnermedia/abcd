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

// A tier path that exists but is a regular FILE (not a directory) does not
// satisfy three-tier-layout — the tiers are directories.
func TestRule_TierPresentButNotADirectory(t *testing.T) {
	b := newFixtureRepo(t).
		file(".gitignore", ".abcd/.work.local/\n").
		file(".abcd/development", "I am a file, not the durable-record tier\n"). // a file at a tier path
		file(".abcd/work/DECISIONS.md", "x\n").
		file(".abcd/.work.local/NEXT.md", "x\n").
		file("AGENTS.md", "x\n").
		commit()
	res := b.run()

	f := findingFor(res, "three-tier-layout")
	if f == nil {
		t.Fatal("no three-tier-layout finding when .abcd/development is a file, not a directory")
	}
	if f.Severity != audit.SeverityError {
		t.Errorf("severity = %q, want error", f.Severity)
	}
}

// A directory named AGENTS.md does not satisfy conventions-router — the router is
// a file.
func TestRule_ConventionsRouterIsADirectory(t *testing.T) {
	b := newFixtureRepo(t).
		file(".gitignore", ".abcd/.work.local/\n").
		file(".abcd/development/README.md", "x\n").
		file(".abcd/work/DECISIONS.md", "x\n").
		file(".abcd/.work.local/NEXT.md", "x\n").
		file("AGENTS.md/keep.txt", "AGENTS.md is a directory here\n"). // dir, not a router file
		commit()
	res := b.run()

	f := findingFor(res, "conventions-router")
	if f == nil {
		t.Fatal("no conventions-router finding when AGENTS.md is a directory")
	}
	if f.Severity != audit.SeverityError {
		t.Errorf("severity = %q, want error", f.Severity)
	}
}

// docs-currency: docs/ exists but no docs-lint config → a warn that the check
// could not run, never a silent pass.
func TestRule_DocsCurrencyNoConfigWarns(t *testing.T) {
	b := newFixtureRepo(t).conforming().
		file("docs/how-to/thing.md", "clean docs\n"). // docs/ present, but no .abcd/docs-lint.json
		commit()
	res := b.run()

	f := findingFor(res, "docs-currency")
	if f == nil {
		t.Fatal("docs-currency silently passed when docs/ exists but the config is missing")
	}
	if f.Severity != audit.SeverityWarn {
		t.Errorf("severity = %q, want warn", f.Severity)
	}
	// It must be a warn, not an error: exit 1, not 2.
	if res.ExitCode != 1 {
		t.Errorf("exit = %d, want 1", res.ExitCode)
	}
}

// AC3: a committed file with an absolute local path → privacy-hygiene error
// citing file:line — unless a waiver escape is on that line.
func TestAC_PrivacyAbsolutePath(t *testing.T) {
	const leak = "see /Users/alice/secret/notes.md for context\n" // abcd-audit:allow
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
	// Kept out of docs/ so docs-currency stays skipped and does not add a warn —
	// this test isolates the privacy waiver's effect on the exit code.
	b := newFixtureRepo(t).conforming().
		file("reference/paths.md", waived).
		commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("waiver escape did not suppress the finding: %+v", f)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0 (waived)", res.ExitCode)
	}
}

// A hostile repo cannot turn privacy-hygiene into an out-of-repo read: a tracked
// symlink pointing outside the work tree is skipped, never followed. Its target
// contains an absolute path, so following it would both leak and (if it were
// /dev/zero) hang.
func TestRule_PrivacySkipsTrackedSymlink(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("leak /Users/victim/.ssh/id_rsa\n"), 0o600); err != nil { // abcd-audit:allow
		t.Fatal(err)
	}
	b := newFixtureRepo(t).conforming()
	link := filepath.Join(b.root, "pointer.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	b.commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("privacy-hygiene followed a tracked symlink out of the repo: %+v", f)
	}
}

// A hostile working tree cannot escape the repo via a symlinked INTERMEDIATE
// directory either: sub/ is tracked, then swapped on disk for a symlink to an
// out-of-repo directory. The scan must refuse to read through it.
func TestRule_PrivacyRejectsIntermediateSymlinkEscape(t *testing.T) {
	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "data.txt"), []byte("leak /Users/victim/x/\n"), 0o600); err != nil { // abcd-audit:allow
		t.Fatal(err)
	}
	b := newFixtureRepo(t).conforming().
		file("sub/data.txt", "clean in-repo content\n"). // tracked, no leak
		commit()
	// Swap the tracked intermediate directory for a symlink pointing outside.
	if err := os.RemoveAll(filepath.Join(b.root, "sub")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(b.root, "sub")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("privacy-hygiene read through a symlinked intermediate directory: %+v", f)
	}
}

// A tracked file larger than the scan cap is skipped rather than loaded whole —
// a large committed binary must not OOM the audit.
func TestRule_PrivacySkipsOversizeFile(t *testing.T) {
	b := newFixtureRepo(t).conforming()
	big := make([]byte, audit.MaxScanBytesForTest()+1)
	for i := range big {
		big[i] = 'a'
	}
	// Put a leak on the first line so a naive scanner would flag it; the size cap
	// must skip the file before it is read.
	copy(big, []byte("/Users/alice/x/\n")) // abcd-audit:allow
	b.file("huge.txt", string(big)).commit()
	res := b.run()

	if f := findingFor(res, "privacy-hygiene"); f != nil {
		t.Fatalf("privacy-hygiene scanned an oversize file: %+v", f)
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
