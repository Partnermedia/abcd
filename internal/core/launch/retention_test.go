package launch

import (
	"os/exec"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

func mustParse(t *testing.T, v string) Semver {
	t.Helper()
	s, err := ParseSemver(v)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func contains(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func TestRetentionPrunesOlderInLine(t *testing.T) {
	plan := ComputeRetention(mustParse(t, "0.1.2"), []Semver{mustParse(t, "0.1.1")})
	if plan.Refused {
		t.Fatalf("unexpected refusal: %+v", plan)
	}
	if !contains(plan.Pruned, "v0.1.1") {
		t.Errorf("expected v0.1.1 pruned: %+v", plan.Pruned)
	}
	if !contains(plan.Kept, "v0.1.2") {
		t.Errorf("published must be kept: %+v", plan.Kept)
	}
	if contains(plan.Pruned, "v0.1.2") {
		t.Errorf("published must never be pruned: %+v", plan.Pruned)
	}
}

func TestRetentionKeepsOtherLines(t *testing.T) {
	plan := ComputeRetention(mustParse(t, "0.2.0"), []Semver{mustParse(t, "0.1.9")})
	if !contains(plan.Kept, "v0.1.9") || !contains(plan.Kept, "v0.2.0") {
		t.Errorf("both lines' newest must be kept: %+v", plan.Kept)
	}
	if len(plan.Pruned) != 0 {
		t.Errorf("nothing to prune across distinct lines: %+v", plan.Pruned)
	}
}

func TestRetentionRefusesOnNewerExisting(t *testing.T) {
	plan := ComputeRetention(mustParse(t, "0.1.2"), []Semver{mustParse(t, "0.1.3")})
	if !plan.Refused || plan.RefusalReason == "" {
		t.Errorf("expected refusal on newer existing release: %+v", plan)
	}
	if len(plan.Pruned) != 0 {
		t.Errorf("a refusal must prune nothing: %+v", plan.Pruned)
	}
}

// TestGitExistingTagsExcludesPrerelease is the attack/behaviour test for the
// prerelease/build filter: a real tag like v1.2.3-rc1 renders its core "v1.2.3"
// via Tag()/String(), so admitting it would surface a PHANTOM v1.2.3 in the
// retention plan and collapse against a real v1.2.3. Only release cores survive.
func TestGitExistingTagsExcludesPrerelease(t *testing.T) {
	root := t.TempDir()
	env := gittest.Env(t)
	gitInit := exec.Command("git", "-C", root, "init")
	gitInit.Env = env
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Skipf("git init unavailable: %v (%s)", err, out)
	}
	mustGit := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	mustGit("-c", "user.email=t@e.x", "-c", "user.name=T", "commit", "--allow-empty", "-m", "c", "--author", "T <t@e.x>")
	mustGit("-c", "user.email=t@e.x", "-c", "user.name=T", "tag", "v1.0.0")
	mustGit("-c", "user.email=t@e.x", "-c", "user.name=T", "tag", "v1.2.3-rc1")

	vers, err := GitExistingTags(root)
	if err != nil {
		t.Fatalf("GitExistingTags: %v", err)
	}
	for _, v := range vers {
		if v.Prerelease != "" || v.Build != "" {
			t.Errorf("prerelease/build tag leaked into the release set: %+v", v)
		}
		if v.Major == 1 && v.Minor == 2 && v.Patch == 3 {
			t.Errorf("v1.2.3-rc1 surfaced as a phantom release core v1.2.3")
		}
	}
	var has100 bool
	for _, v := range vers {
		if v.Major == 1 && v.Minor == 0 && v.Patch == 0 {
			has100 = true
		}
	}
	if !has100 {
		t.Errorf("real release v1.0.0 missing from %+v", vers)
	}
}
