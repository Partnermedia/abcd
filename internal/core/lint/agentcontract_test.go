package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gitutil"
)

// agentCfg builds a one-rule agent_contract config pointed at a fixture tree.
func agentCfg() Config {
	return Config{Rules: map[string]RuleConfig{
		ruleAgentContract: {Enabled: true, Severity: severityBlocker},
	}}
}

// writeAgent puts one agent prompt in the agents tree. frontmatter is the body of
// the leading `---` block, so a test can drop exactly the field under test.
func writeAgent(t *testing.T, root, name, frontmatter string) {
	t.Helper()
	writeFile(t, root, filepath.Join("agents", name+".md"),
		"---\nname: "+name+"\n"+frontmatter+"---\n\n# "+name+"\n\nPrompt body.\n")
}

// writeCanary puts the injection-canary fixture an untrusted-input agent must ship.
func writeCanary(t *testing.T, root, name string) {
	t.Helper()
	writeFile(t, root, filepath.Join("agents", name, "fixtures", "injection-canary.json"),
		"{\"expected\": {\"obeys\": false}}\n")
}

// conformingAgent is the itd-5 trust contract in full: the version, the explicit
// untrusted-input declaration, and the capability scope.
const conformingAgent = "prompt_version: 0.2.0\n" +
	"reads_untrusted_input: true\n" +
	"capability_scope:\n" +
	"  task_classes: [oracle_review]\n" +
	"  designed_for: \"Family-1 change judgement\"\n"

// TestAgentContractWalksAgentsTree is the "record-lint sees agents/ at all"
// criterion: the tree is in neither lint root, so before this rule nothing under
// it was ever read. A defective agent there must now produce a finding even
// though cfg.Roots names no tree at all.
func TestAgentContractWalksAgentsTree(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "ruthless-reviewer", conformingAgent)
	// No canary fixture: the tree must be walked for this to be noticed.

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if countRule(fs, ruleAgentContract) == 0 {
		t.Fatalf("expected agents/ to be walked and the defect reported; got %+v", fs)
	}
	if !hasFinding(fs, filepath.Join("agents", "ruthless-reviewer.md"), ruleAgentContract, 1) {
		t.Errorf("expected the finding against the agent prompt; got %+v", fs)
	}
}

// TestAgentContractCompleteAgentPasses is the clean case: frontmatter, canary and
// changelog entry all present, no finding raised.
func TestAgentContractCompleteAgentPasses(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "ruthless-reviewer", conformingAgent)
	writeCanary(t, root, "ruthless-reviewer")
	writeFile(t, root, filepath.Join("agents", "CHANGELOG.md"),
		"# Agent prompt changelog\n\n### ruthless-reviewer 0.2.0\n\nFirst entry.\n")
	// README and CHANGELOG are prose, never agent prompts.
	writeFile(t, root, filepath.Join("agents", "README.md"), "# Agents\n")

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleAgentContract); n != 0 {
		t.Fatalf("expected a conforming agent to lint clean; got %d: %+v", n, fs)
	}
}

// TestAgentContractMissingTrustFields is the itd-5 frontmatter criterion: an
// agent that declares it reads attacker-influenceable input but carries none of
// the contract fields is reported, and the message NAMES the missing field.
func TestAgentContractMissingTrustFields(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "sota-researcher", "reads_untrusted_input: true\n")
	writeCanary(t, root, "sota-researcher")

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !messageContains(fs, "prompt_version") {
		t.Errorf("expected the missing prompt_version to be named; got %+v", fs)
	}
	if !messageContains(fs, "capability_scope.task_classes") {
		t.Errorf("expected the missing task_classes to be named; got %+v", fs)
	}
	if !messageContains(fs, "capability_scope.designed_for") {
		t.Errorf("expected the missing designed_for to be named; got %+v", fs)
	}
}

// TestAgentContractRejectsMalformedPromptVersion pins the semver half: a
// prompt_version that is not a semver cannot be reconciled with a changelog
// entry, so it is reported rather than accepted as present.
func TestAgentContractRejectsMalformedPromptVersion(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "sota-researcher", strings.Replace(conformingAgent,
		"prompt_version: 0.2.0", "prompt_version: v2", 1))
	writeCanary(t, root, "sota-researcher")

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !messageContains(fs, "semver") {
		t.Errorf("expected the malformed prompt_version to be reported; got %+v", fs)
	}
}

// TestAgentContractUndeclaredUntrustedInput closes the omission route: a prompt
// that simply never declares reads_untrusted_input would otherwise escape every
// sub-check by deleting one line, which is a detector any new agent can opt out
// of. The declaration itself is required.
func TestAgentContractUndeclaredUntrustedInput(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "sota-researcher", "prompt_version: 0.2.0\n")

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !messageContains(fs, "reads_untrusted_input") {
		t.Errorf("expected the undeclared trust posture to be reported; got %+v", fs)
	}
}

// TestAgentContractDeclaredTrustedAgentNeedsNoCanary pins the other side of that
// declaration: an agent that declares it reads no untrusted input carries no
// canary obligation.
func TestAgentContractDeclaredTrustedAgentNeedsNoCanary(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "composer", "prompt_version: 0.1.0\nreads_untrusted_input: false\n")

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleAgentContract); n != 0 {
		t.Fatalf("expected no canary obligation for a declared-trusted agent; got %d: %+v", n, fs)
	}
}

// TestAgentContractMissingCanary is the injection-canary criterion: the contract
// fields are all present, the fixture is not, and the finding names it.
func TestAgentContractMissingCanary(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "security-reviewer", conformingAgent)

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !messageContains(fs, "injection-canary.json") {
		t.Errorf("expected the missing canary fixture to be named; got %+v", fs)
	}
}

// TestAgentContractChangelogEntryRequiredOverDiff is the per-agent changelog
// criterion. It fires only over a diff: an agent whose prompt changed in the
// range under lint must have gained its CHANGELOG entry in the same change.
func TestAgentContractChangelogEntryRequiredOverDiff(t *testing.T) {
	root := newAgentRepo(t)

	// Bump the prompt without touching the changelog: the shape the rule catches.
	writeAgent(t, root, "ruthless-reviewer", strings.Replace(conformingAgent,
		"prompt_version: 0.2.0", "prompt_version: 0.3.0", 1))

	fs, err := Lint(ArmAgentDiff(agentCfg(), "HEAD"), root)
	if err != nil {
		t.Fatal(err)
	}
	if !messageContains(fs, "CHANGELOG") {
		t.Fatalf("expected the missing changelog entry to be reported; got %+v", fs)
	}
	if !messageContains(fs, "0.3.0") {
		t.Errorf("expected the message to name the version it wants an entry for; got %+v", fs)
	}

	// The same diff with the entry present is clean.
	writeFile(t, root, filepath.Join("agents", "CHANGELOG.md"),
		"# Agent prompt changelog\n\n### ruthless-reviewer 0.3.0\n\nBumped.\n\n### ruthless-reviewer 0.2.0\n\nFirst entry.\n")
	fs, err = Lint(ArmAgentDiff(agentCfg(), "HEAD"), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleAgentContract); n != 0 {
		t.Fatalf("expected the changed agent with its entry to lint clean; got %d: %+v", n, fs)
	}
}

// TestAgentContractChangelogSkippedWithoutDiffRange pins the no-op: a full-tree
// lint outside a git range checks the frontmatter and the canary but says
// nothing about the changelog, because it has no diff to say it about.
func TestAgentContractChangelogSkippedWithoutDiffRange(t *testing.T) {
	root := newAgentRepo(t)
	writeAgent(t, root, "ruthless-reviewer", strings.Replace(conformingAgent,
		"prompt_version: 0.2.0", "prompt_version: 0.3.0", 1))

	fs, err := Lint(agentCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleAgentContract); n != 0 {
		t.Fatalf("expected no changelog finding without a diff range; got %d: %+v", n, fs)
	}
}

// TestAgentContractRefusesHostileDiffRange pins the argument guard: the range is
// caller-supplied and reaches a git subprocess, so a value that could be read as
// an option (or as anything but a revision) is refused rather than executed.
func TestAgentContractRefusesHostileDiffRange(t *testing.T) {
	root := newAgentRepo(t)

	if _, err := Lint(ArmAgentDiff(agentCfg(), "--output=/tmp/pwned"), root); err == nil {
		t.Fatal("expected an option-shaped diff range to be refused")
	}
}

// newAgentRepo builds a git repository holding one conforming agent, its canary
// and its changelog entry, all committed, so a test can dirty exactly one file
// and lint the resulting diff.
func newAgentRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeAgent(t, root, "ruthless-reviewer", conformingAgent)
	writeCanary(t, root, "ruthless-reviewer")
	writeFile(t, root, filepath.Join("agents", "CHANGELOG.md"),
		"# Agent prompt changelog\n\n### ruthless-reviewer 0.2.0\n\nFirst entry.\n")

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(gitutil.IsolatedEnv(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("add", "-A")
	run("commit", "-qm", "seed")
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	return root
}
