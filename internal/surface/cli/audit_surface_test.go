package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// auditRepo builds a git repo at t.TempDir with the given layout knobs and
// returns its path. A conforming repo satisfies every v1 rule.
func auditRepo(t *testing.T, conforming bool) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := t.TempDir()
	runGitT(t, repo, "init", "-q")
	write := func(rel, body string) {
		p := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".gitignore", ".abcd/.work.local/\n")
	write(".abcd/development/README.md", "x\n")
	write(".abcd/.work.local/NEXT.md", "x\n")
	write("AGENTS.md", "x\n")
	if conforming {
		write(".abcd/work/DECISIONS.md", "# decisions\n")
	}
	runGitT(t, repo, "add", "-A")
	runGitT(t, repo, "-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-q", "-m", "fixture")
	return repo
}

func runGitT(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// A conforming repo: `abcd audit` exits 0, and `--json` emits {"findings": []}.
func TestAuditConformingExitsZero(t *testing.T) {
	repo := auditRepo(t, true)
	t.Chdir(repo)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"audit", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstdout:%s\nstderr:%s", code, stdout.String(), stderr.String())
	}
	var res struct {
		Findings []any `json:"findings"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, stdout.String())
	}
	if res.Findings == nil {
		t.Error(`clean repo must emit "findings": [] (present empty array), not null`)
	}
	if len(res.Findings) != 0 {
		t.Errorf("conforming repo findings = %d, want 0", len(res.Findings))
	}
}

// A repo missing the committed work tier: exit 2, and the JSON carries the
// three-tier-layout rule id at error severity.
func TestAuditMissingWorkTierExitsTwo(t *testing.T) {
	repo := auditRepo(t, false) // no .abcd/work/DECISIONS.md, so no work/ tier
	t.Chdir(repo)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"audit", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2\nstdout:%s", code, stdout.String())
	}
	var res struct {
		Findings []struct {
			RuleID   string `json:"ruleId"`
			Severity string `json:"severity"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, stdout.String())
	}
	found := false
	for _, f := range res.Findings {
		if f.RuleID == "three-tier-layout" && f.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Errorf("no three-tier-layout error in JSON:\n%s", stdout.String())
	}
}

// A cobra usage error (a stray positional argument, an unknown flag) must exit 2,
// not 1: audit documents Conftest's tri-state where exit 1 means "warnings only",
// so a mistyped invocation landing on 1 would let a CI gate record a clean-ish
// pass for an audit that never ran (B13). These fail before RunE, so they need no
// repo fixture.
func TestAuditUsageErrorsExitTwo(t *testing.T) {
	cases := [][]string{
		{"audit", "unexpected-arg"}, // stray positional under cobra.NoArgs
		{"audit", "--nosuchflag"},   // unknown flag
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(args, &stdout, &stderr)
			if code != 2 {
				t.Fatalf("usage error exit = %d, want 2 (must not collide with the audit tri-state's exit-1 'warnings only')\nstderr:%s", code, stderr.String())
			}
		})
	}
}

// `abcd audit --root <missing>` must report a usage error, not fabricate
// convention violations against a directory that is not there (B41).
func TestAuditNonexistentRootIsUsageError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"audit", "--root", missing}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2\nstderr:%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "is not a directory") {
		t.Errorf("want an 'is not a directory' diagnostic, got stderr:\n%s", stderr.String())
	}
	if strings.Contains(stdout.String(), "conventions-router") || strings.Contains(stdout.String(), "three-tier-layout") {
		t.Errorf("audit fabricated convention findings against a missing dir:\n%s", stdout.String())
	}
}

// The human render (no --json) is grouped and readable, and stdout stays free of
// JSON braces.
func TestAuditHumanRender(t *testing.T) {
	repo := auditRepo(t, false)
	t.Chdir(repo)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"audit"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	out := stdout.String()
	if strings.Contains(out, "{") {
		t.Errorf("human render leaked JSON braces:\n%s", out)
	}
	if !strings.Contains(out, "three-tier-layout") {
		t.Errorf("human render omits the failing rule id:\n%s", out)
	}
}
