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

// hookRepo is a throwaway repo with the committed hook installed, so a test can
// stage whatever shape it needs (a rename, a binary blob, a huge file, a
// .gitattributes that suppresses the textual diff) and then attempt a real commit.
type hookRepo struct {
	t    *testing.T
	dir  string
	env  []string
	name string
}

// newHookRepo initialises the repo, installs the committed hook, and writes the
// private banlist when body != "" (body == "" leaves it absent: the fresh-clone
// case).
func newHookRepo(t *testing.T, body string) *hookRepo {
	t.Helper()
	hook := locateHook(t)
	r := &hookRepo{t: t, dir: t.TempDir(), env: gittest.Env(t)}

	r.git("init")
	r.git("config", "user.name", "Alice Example")
	r.git("config", "user.email", "alice@example.com")

	if body != "" {
		r.writeBanlist(body)
	}

	hooksDir := filepath.Join(r.dir, ".git", "hooks")
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
	return r
}

// writeBanlist installs the private store's bytes.
func (r *hookRepo) writeBanlist(body string) {
	r.t.Helper()
	local := filepath.Join(r.dir, ".abcd", ".work.local")
	if err := os.MkdirAll(local, 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "private-names.txt"), []byte(body), 0o600); err != nil {
		r.t.Fatal(err)
	}
}

// git runs a git command that must succeed.
func (r *hookRepo) git(args ...string) string {
	r.t.Helper()
	out, err := r.tryGit(args...)
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return out
}

// tryGit runs a git command and returns its combined output and error.
func (r *hookRepo) tryGit(args ...string) (string, error) {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.dir}, args...)...)
	cmd.Env = r.env
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// write puts content at a repo-relative path, creating parents.
func (r *hookRepo) write(rel, content string) {
	r.t.Helper()
	p := filepath.Join(r.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// commit attempts a commit and reports whether the hook refused it, plus every
// byte the hook wrote.
func (r *hookRepo) commit() (blocked bool, output string) {
	r.t.Helper()
	out, err := r.tryGit("commit", "-m", "t")
	return err != nil, out
}

// hookRun is the common shape: stage `staged` as one file's content and attempt a
// commit against the given banlist body.
func hookRun(t *testing.T, banlist, staged string) (blocked bool, output string) {
	t.Helper()
	r := newHookRepo(t, banlist)
	r.write("note.md", staged)
	r.git("add", "note.md")
	return r.commit()
}

// corpus reads one of the shared fixture banlists — the files the Go parser reads
// too, so both readers are driven by identical bytes.
func corpus(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
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

// TestPreCommitHook_EntrylessStoreWarnsLoudly is the other half of AC4, and the
// one an emptied store hits: a store that exists but yields no entries checks
// exactly as much as an absent one, so it must be exactly as loud. A refresh that
// truncates the store must not convert the warning into silence.
func TestPreCommitHook_EntrylessStoreWarnsLoudly(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"empty file", "\n"},
		{"comments only", "# abcd-banlist: keyed\n# nothing yet\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			blocked, out := hookRun(t, tc.body, "widgetworks ships today\n")
			if blocked {
				t.Fatalf("commit blocked by an entryless store\n%s", out)
			}
			for _, want := range []string{"WARNING", "NO ENTRIES"} {
				if !strings.Contains(out, want) {
					t.Errorf("hook output does not mention %q; an entryless store checks nothing and must say so\n%s", want, out)
				}
			}
		})
	}
}

// TestPreCommitHook_RefusesByKeyOnly pins AC2: the refusal names the entry key
// and nothing else — not the matched string, not the pattern value.
func TestPreCommitHook_RefusesByKeyOnly(t *testing.T) {
	const banlist = "# abcd-banlist: keyed\nwidget-partner   widgetworks\n"
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

// TestPreCommitHook_KeyedCorpus pins AC3 on the keyed corpus: hostnames, IPv4/IPv6
// addresses, CIDR prefixes, and MAC addresses are matched exactly as name entries
// are, and a tab separates fields exactly as spaces do. Every value is reserved
// for documentation (RFC 5737/3849/2606/7042) or derived from the persona registry.
func TestPreCommitHook_KeyedCorpus(t *testing.T) {
	body := corpus(t, "parse-corpus.txt")
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

// TestPreCommitHook_LegacyStoreReadsWholeLines pins the compatibility rule at its
// exact strength: a store with no format declaration is read one WHOLE-LINE
// pattern per line, so it keeps matching precisely what it always matched, and no
// part of a line is ever read — or printed — as a key.
func TestPreCommitHook_LegacyStoreReadsWholeLines(t *testing.T) {
	body := corpus(t, "parse-corpus-legacy.txt")
	for _, tc := range legacyMustBlock {
		t.Run(tc.name, func(t *testing.T) {
			blocked, out := hookRun(t, body, tc.text)
			if !blocked {
				t.Fatalf("legacy line did not block; want a refusal naming %q\n%s", tc.key, out)
			}
			if !strings.Contains(out, tc.key) {
				t.Errorf("refusal does not name the synthetic key %q\n%s", tc.key, out)
			}
			for _, leak := range append([]string{"partnerco"}, legacyFirstFields...) {
				if strings.Contains(out, leak) {
					t.Errorf("output leaks %q: on a legacy line the first field is PART OF THE PATTERN, never a key\n%s", leak, out)
				}
			}
		})
	}
}

// TestPreCommitHook_LegacyStoreDoesNotSplitKeys is the must-pass half of the same
// rule and the detector for the key-splitting leak: if the hook split a legacy line
// on whitespace, the remainder-only patterns below would start matching and the
// first field would be printed as a key.
func TestPreCommitHook_LegacyStoreDoesNotSplitKeys(t *testing.T) {
	body := corpus(t, "parse-corpus-legacy.txt")
	for _, tc := range legacyMustPass {
		t.Run(tc.name, func(t *testing.T) {
			blocked, out := hookRun(t, body, tc.text)
			if blocked {
				t.Fatalf("commit refused for content matching no whole-line pattern\n%s", out)
			}
		})
	}
}

// TestPreCommitHook_PermittedCorpusPasses is the must-pass half of the keyed
// guard's bidirectional proof (guards-prove-themselves): content that matches no
// entry commits cleanly, so the guard is not simply refusing everything. It also
// pins that comment and blank lines — including the format declaration itself —
// are skipped rather than read as patterns.
func TestPreCommitHook_PermittedCorpusPasses(t *testing.T) {
	body := corpus(t, "parse-corpus.txt")
	for _, tc := range corpusMustPass {
		t.Run(tc.name, func(t *testing.T) {
			blocked, out := hookRun(t, body, tc.text)
			if blocked {
				t.Fatalf("commit refused for content matching no entry\n%s", out)
			}
		})
	}
}

// TestPreCommitHook_UnusableLinesFailSafe pins the malformed-entry contract on the
// shared corpus: every unusable line class refuses the commit by LINE NUMBER, none
// is silently skipped, and no line's content is echoed. A keyed store's unparseable
// line is the important case — its first field may be the secret, so it has no key
// and the hook must not invent one out of the line's bytes.
func TestPreCommitHook_UnusableLinesFailSafe(t *testing.T) {
	blocked, out := hookRun(t, corpus(t, "parse-corpus-malformed.txt"), "nothing sensitive here\n")
	if !blocked {
		t.Fatalf("unusable banlist lines did not fail safe\n%s", out)
	}
	for _, line := range malformedUnusableLines {
		want := "line " + itoa(line)
		if !strings.Contains(out, want) {
			t.Errorf("refusal does not name %q; every unusable line must be reported\n%s", want, out)
		}
	}
	for _, leak := range []string{"unclosed", "partnerco", "nbsp-key", "vt-key"} {
		if strings.Contains(out, leak) {
			t.Errorf("output echoes %q from an unusable line; the content is withheld by design\n%s", leak, out)
		}
	}
}

// itoa keeps the assertions above free of a strconv import in a file that is
// otherwise all shell-driving.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
