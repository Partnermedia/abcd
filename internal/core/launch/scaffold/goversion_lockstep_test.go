package scaffold

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestWorkflowGoVersionsMatchSubstitutions couples EVERY go-version pin under
// .github/workflows/ to AbcdSubstitutions().GoVersion (iss-354).
// TestSelfScaffoldParity gates release.yml and auto-release.yml through the
// templates, but nothing read ci.yml — so the d594511 incident (CI scanning
// green on a patched toolchain while the release built on the unpatched one,
// shipping four known stdlib CVEs) could recur silently on the next bump.
func TestWorkflowGoVersionsMatchSubstitutions(t *testing.T) {
	root := repoRoot(t)
	want := AbcdSubstitutions().GoVersion
	re := regexp.MustCompile(`go-version:\s*'?"?([0-9][0-9.]*)`)

	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, e := range entries {
		if ext := filepath.Ext(e.Name()); e.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range re.FindAllStringSubmatch(string(data), -1) {
			checked++
			if m[1] != want {
				t.Errorf("%s pins go-version %s; scaffold substitutions say %s — bump them in lockstep",
					e.Name(), m[1], want)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no go-version pins found under .github/workflows — the sweep matched nothing, which cannot be right")
	}
}

// TestPullRequestTargetWorkflowsHaveNoCheckout arms the invariant external-review.yml
// states in prose ("the PR's code is never checked out, never built, never
// executed, which is what makes the trigger safe to use here"): a workflow
// triggered by pull_request_target must not check out PR code, or it becomes a
// pwn request. zizmor's dangerous-triggers audit is (necessarily) suppressed on
// that file, so nothing else catches a later checkout being added — this test does.
func TestPullRequestTargetWorkflowsHaveNoCheckout(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	triggerRe := regexp.MustCompile(`(?m)^\s*pull_request_target:`)
	checkoutRe := regexp.MustCompile(`(?m)uses:\s*['"]?[^'"\n]*actions/checkout`)
	checked := 0
	for _, e := range entries {
		if ext := filepath.Ext(e.Name()); e.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if !triggerRe.Match(data) {
			continue
		}
		checked++
		if checkoutRe.Match(data) {
			t.Errorf("%s is triggered by pull_request_target and checks out code — a pwn request; keep PR code out of this workflow", e.Name())
		}
	}
	if checked == 0 {
		t.Fatal("no pull_request_target workflow found — the sweep matched nothing; external-review.yml should trip this")
	}
}
