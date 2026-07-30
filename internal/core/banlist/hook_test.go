package banlist

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

// locateHook finds the committed .githooks/pre-commit by walking up from the
// test's working directory. Skips when not run from a checkout (e.g. a build
// tarball) or when bash/git are unavailable.
func locateHook(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash unavailable")
	}
	topLevel := exec.Command("git", "rev-parse", "--show-toplevel")
	topLevel.Env = gittest.Env(t)
	out, err := topLevel.Output()
	if err != nil {
		t.Skip("not in a git checkout")
	}
	hook := filepath.Join(strings.TrimSpace(string(out)), ".githooks", "pre-commit")
	if _, err := os.Stat(hook); err != nil {
		t.Skipf("hook not found: %v", err)
	}
	return hook
}

// hookRun stands up a throwaway repo with the committed hook installed, stages
// `staged` as a file's content, and attempts a real commit. It returns whether
// the commit was refused plus every byte the hook wrote. banlist == "" leaves
// the private banlist file absent (the fresh-clone case).
func hookRun(t *testing.T, banlist, staged string) (blocked bool, output string) {
	t.Helper()
	hook := locateHook(t)
	env := gittest.Env(t)
	dir := t.TempDir()

	git := func(mustPass bool, args ...string) (string, error) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil && mustPass {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out), err
	}

	git(true, "init")
	git(true, "config", "user.name", "Alice Example")
	git(true, "config", "user.email", "alice@example.com")

	if banlist != "" {
		local := filepath.Join(dir, ".abcd", ".work.local")
		if err := os.MkdirAll(local, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(local, "private-names.txt"), []byte(banlist), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(hook)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"), src, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte(staged), 0o644); err != nil {
		t.Fatal(err)
	}
	git(true, "add", "note.md")
	out, err := git(false, "commit", "-m", "t")
	return err != nil, out
}

// corpus is the shared fixture banlist read by BOTH parsers: this hook test and
// the Go store's parse test (parse_test.go). One file, two readers — the only
// way the shell hook and the Go package can be shown to agree on the format.
func corpus(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "parse-corpus.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestPreCommitHook_AbsentBanlistWarnsLoudly pins AC4: a machine with no private
// banlist is UNPROTECTED, and the hook says so out loud. A silent pass looks
// exactly like a clean check, which is the failure mode this test exists for.
func TestPreCommitHook_AbsentBanlistWarnsLoudly(t *testing.T) {
	blocked, out := hookRun(t, "", "widgetworks ships today\n")
	if blocked {
		t.Fatalf("commit blocked with no banlist present; want it to proceed\n%s", out)
	}
	for _, want := range []string{"WARNING", "INACTIVE", "private-names.txt"} {
		if !strings.Contains(out, want) {
			t.Errorf("hook output does not mention %q; the inactive layer must announce itself\n%s", want, out)
		}
	}
}

// TestPreCommitHook_RefusesByKeyOnly pins AC2: the refusal names the entry key
// and nothing else — not the matched string, not the pattern value.
func TestPreCommitHook_RefusesByKeyOnly(t *testing.T) {
	const banlist = "widget-partner   widgetworks\n"
	blocked, out := hookRun(t, banlist, "the widgetworks deal closes friday\n")
	if !blocked {
		t.Fatalf("commit not blocked by a matching banlist entry\n%s", out)
	}
	if !strings.Contains(out, "widget-partner") {
		t.Errorf("refusal does not name the entry key\n%s", out)
	}
	for _, leak := range []string{"widgetworks", "WIDGETWORKS", "friday"} {
		if strings.Contains(strings.ToLower(out), strings.ToLower(leak)) {
			t.Errorf("output leaks %q; neither the pattern nor the matched line may be echoed\n%s", leak, out)
		}
	}
}

// TestPreCommitHook_MachineIdentifiers pins AC3: hostnames, IPv4/IPv6 addresses,
// CIDR prefixes, and MAC addresses are matched exactly as name entries are. Every
// value here is reserved for documentation (RFC 5737/3849/2606/7042) or derived
// from the persona registry.
func TestPreCommitHook_MachineIdentifiers(t *testing.T) {
	body := corpus(t)
	for _, tc := range corpusMustBlock {
		t.Run(tc.name, func(t *testing.T) {
			blocked, out := hookRun(t, body, tc.text)
			if !blocked {
				t.Fatalf("commit not blocked; want a refusal naming %q\n%s", tc.key, out)
			}
			if !strings.Contains(out, tc.key) {
				t.Errorf("refusal does not name key %q\n%s", tc.key, out)
			}
			if strings.Contains(out, strings.TrimSpace(tc.text)) {
				t.Errorf("output echoes the matched line\n%s", out)
			}
		})
	}
}

// TestPreCommitHook_LegacyBarePatternStillBlocks pins the compatibility rule: a
// banlist written in the old one-pattern-per-line format must keep blocking.
// Protection never weakens because the format grew a key column.
func TestPreCommitHook_LegacyBarePatternStillBlocks(t *testing.T) {
	blocked, out := hookRun(t, corpus(t), "partnerco.example signed\n")
	if !blocked {
		t.Fatalf("legacy bare-pattern line did not block\n%s", out)
	}
	if !strings.Contains(out, "entry-") {
		t.Errorf("refusal does not name the synthetic key for a bare pattern\n%s", out)
	}
	if strings.Contains(out, "partnerco") {
		t.Errorf("output leaks the pattern/matched text\n%s", out)
	}
}

// TestPreCommitHook_PermittedCorpusPasses is the must-pass half of the guard's
// bidirectional proof (guards-prove-themselves): content that matches no entry
// commits cleanly, so the guard is not simply refusing everything. It also pins
// that comment and blank lines are skipped rather than read as patterns.
func TestPreCommitHook_PermittedCorpusPasses(t *testing.T) {
	for _, tc := range corpusMustPass {
		t.Run(tc.name, func(t *testing.T) {
			blocked, out := hookRun(t, corpus(t), tc.text)
			if blocked {
				t.Fatalf("commit refused for content matching no entry\n%s", out)
			}
		})
	}
}

// TestPreCommitHook_MalformedLineFailsSafe pins the malformed-entry contract: an
// unusable pattern is never silently skipped (that would weaken the guard) and
// its content is never echoed — the refusal names the line number alone.
func TestPreCommitHook_MalformedLineFailsSafe(t *testing.T) {
	const banlist = "widget-partner   widgetworks\nbad-entry        [unclosed\n"
	blocked, out := hookRun(t, banlist, "nothing sensitive here\n")
	if !blocked {
		t.Fatalf("malformed banlist line did not fail safe\n%s", out)
	}
	if !strings.Contains(out, "line 2") {
		t.Errorf("refusal does not name the offending line number\n%s", out)
	}
	if strings.Contains(out, "unclosed") {
		t.Errorf("output echoes the malformed line's content\n%s", out)
	}
}
