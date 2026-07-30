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

// writeBanlist installs the private store's bytes, and gitignores the local tier
// exactly as a real abcd-managed repo does — without that, a test's `git add -A`
// stages the banlist itself and its patterns appear in the staged diff, which would
// let a shape test pass for the wrong reason.
func (r *hookRepo) writeBanlist(body string) {
	r.t.Helper()
	local := filepath.Join(r.dir, ".abcd", ".work.local")
	if err := os.MkdirAll(local, 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "private-names.txt"), []byte(body), 0o600); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.dir, ".gitignore"), []byte(".abcd/.work.local/\n"), 0o644); err != nil {
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

// keyedBanlist is the one-entry store the staged-shape tests below share.
const keyedBanlist = "# abcd-banlist: keyed\nwidget-partner   widgetworks\n"

// TestPreCommitHook_StagedShapesThatDefeatDiffText is the reason the guard reads
// staged BLOBS rather than diff text. Each shape below stages the banned name and
// produces a textual diff in which it does not appear as an added line — so a guard
// that greps `git diff --cached` output passes the commit while the name goes into
// history. A guard that reads `git show :<path>` sees the content itself and cannot
// be shaped around.
func TestPreCommitHook_StagedShapesThatDefeatDiffText(t *testing.T) {
	t.Run("content line beginning ++", func(t *testing.T) {
		// In a unified diff this line is emitted as "+++widgetworks", which every
		// header filter drops along with the real "+++ b/path" header.
		r := newHookRepo(t, keyedBanlist)
		r.write("note.md", "++widgetworks\n")
		r.git("add", "note.md")
		blocked, out := r.commit()
		if !blocked {
			t.Fatalf("commit not blocked; the name is in the staged content\n%s", out)
		}
	})

	t.Run("binary blob", func(t *testing.T) {
		// A NUL makes git call the file binary: the textual diff carries no + lines
		// at all, so there is nothing for a diff-text guard to match.
		r := newHookRepo(t, keyedBanlist)
		r.write("blob.bin", "\x00\x01\x02widgetworks\x00trailer\n")
		r.git("add", "blob.bin")
		blocked, out := r.commit()
		if !blocked {
			t.Fatalf("commit not blocked; the name is in a staged binary blob\n%s", out)
		}
	})

	t.Run("gitattributes suppressing the diff", func(t *testing.T) {
		// A committed `* -diff` suppresses the textual diff repo-wide, which turns a
		// diff-text guard off for every file in the repo at once.
		r := newHookRepo(t, keyedBanlist)
		r.write(".gitattributes", "* -diff\n")
		r.write("seed.md", "nothing sensitive here\n")
		r.git("add", ".gitattributes", "seed.md")
		if blocked, out := r.commit(); blocked {
			t.Fatalf("seed commit refused\n%s", out)
		}
		r.write("note.md", "widgetworks ships today\n")
		r.git("add", "note.md")
		blocked, out := r.commit()
		if !blocked {
			t.Fatalf("commit not blocked; `* -diff` must not switch the guard off\n%s", out)
		}
	})

	t.Run("rename plus modification", func(t *testing.T) {
		// Status R, which --diff-filter=ACM excludes outright.
		r := newHookRepo(t, keyedBanlist)
		r.write("old.md", strings.Repeat("filler line for rename detection\n", 20))
		r.git("add", "old.md")
		if blocked, out := r.commit(); blocked {
			t.Fatalf("seed commit refused\n%s", out)
		}
		r.git("mv", "old.md", "new.md")
		r.write("new.md", strings.Repeat("filler line for rename detection\n", 20)+"widgetworks\n")
		r.git("add", "-A")
		blocked, out := r.commit()
		if !blocked {
			t.Fatalf("commit not blocked; a renamed-and-modified file is staged content\n%s", out)
		}
	})
}

// TestPreCommitHook_LargeStagedFileStillBlocks is the regression pin for the
// SIGPIPE class the increment already fixed and never tested: a staged file far
// larger than a pipe buffer, with the name on its FIRST line, so a matching grep
// exits long before the writer finishes. Deleting the temp-file machinery would
// reintroduce a fail-open pass with a green suite.
func TestPreCommitHook_LargeStagedFileStillBlocks(t *testing.T) {
	r := newHookRepo(t, keyedBanlist)
	var b strings.Builder
	b.WriteString("widgetworks on the very first line\n")
	for i := 0; i < 200000; i++ {
		b.WriteString("filler line that matches no banlist entry at all\n")
	}
	r.write("big.md", b.String())
	r.git("add", "big.md")
	blocked, out := r.commit()
	if !blocked {
		t.Fatalf("commit not blocked; a large staged file must not defeat the guard\n%s", out)
	}
}

// TestPreCommitHook_CleansUpItsCandidateFile pins the hygiene half: the guard's
// scratch copy of the staged content lives in the gitignored local tier, not in a
// shared $TMPDIR, and it is gone whichever way the hook exits.
func TestPreCommitHook_CleansUpItsCandidateFile(t *testing.T) {
	r := newHookRepo(t, keyedBanlist)
	r.write("note.md", "the widgetworks deal closes friday\n")
	r.git("add", "note.md")
	if blocked, out := r.commit(); !blocked {
		t.Fatalf("commit not blocked\n%s", out)
	}
	left, err := os.ReadDir(filepath.Join(r.dir, ".abcd", ".work.local"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range left {
		if e.Name() != "private-names.txt" {
			t.Errorf("the guard left %q behind; the candidate copy of the staged content is trapped on every exit", e.Name())
		}
	}
}

// TestPreCommitHook_FailsClosedWhenItCannotReadStagedContent pins the direction the
// guard must fail in: "could not compute" is never allowed to look like "clean". A
// staged path whose blob cannot be produced (the index points at an object that is
// not in the object store) refuses the commit and names the step, not the content.
func TestPreCommitHook_FailsClosedWhenItCannotReadStagedContent(t *testing.T) {
	r := newHookRepo(t, keyedBanlist)
	r.write("note.md", "nothing sensitive here\n")
	r.git("add", "note.md")
	// Empty the object store: the index still names note.md, so listing the staged
	// paths succeeds and `git show :note.md` then cannot produce the blob. This is the
	// "could not compute" case, and the guard must say so rather than fall through.
	objects := filepath.Join(r.dir, ".git", "objects")
	if err := os.RemoveAll(objects); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(objects, 0o755); err != nil {
		t.Fatal(err)
	}
	_, out := r.commit()
	if !strings.Contains(out, "could not read the staged content") {
		t.Errorf("the guard did not report that it could not read the staged content; a check that cannot run must never look like a check that passed\n%s", out)
	}
}

// TestPreCommitHook_RefusesANonRegularStore pins tampering as tampering. A symlink
// at the store's path (or at its directory) silently swaps the guard's list for an
// empty or attacker-chosen one — and `git add -f` defeats the gitignore, so a
// checkout can materialise it. A directory or FIFO there would take the "absent"
// branch and read as "this machine has not opted in". Both are refusals.
func TestPreCommitHook_RefusesANonRegularStore(t *testing.T) {
	t.Run("symlinked store", func(t *testing.T) {
		r := newHookRepo(t, keyedBanlist)
		store := filepath.Join(r.dir, ".abcd", ".work.local", "private-names.txt")
		decoy := filepath.Join(r.dir, "decoy.txt")
		if err := os.WriteFile(decoy, []byte("# abcd-banlist: keyed\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(store); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(decoy, store); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		r.write("note.md", "the widgetworks deal closes friday\n")
		r.git("add", "note.md")
		blocked, out := r.commit()
		if !blocked {
			t.Fatalf("a symlinked banlist silently replaced the guard's list\n%s", out)
		}
		if !strings.Contains(out, "SYMLINK") {
			t.Errorf("refusal does not name the cause\n%s", out)
		}
	})

	t.Run("directory at the store path", func(t *testing.T) {
		r := newHookRepo(t, keyedBanlist)
		store := filepath.Join(r.dir, ".abcd", ".work.local", "private-names.txt")
		if err := os.Remove(store); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(store, 0o700); err != nil {
			t.Fatal(err)
		}
		r.write("note.md", "the widgetworks deal closes friday\n")
		r.git("add", "note.md")
		blocked, out := r.commit()
		if !blocked {
			t.Fatalf("a directory at the store path was read as a machine that never opted in\n%s", out)
		}
		if !strings.Contains(out, "not a regular file") {
			t.Errorf("refusal does not name the cause\n%s", out)
		}
	})
}

// TestPreCommitHook_DoesNotTraceThePatternsUnderXtrace pins the one thing an
// inherited shell option must not be able to do: turn the guard into a printer of
// the very list it protects. An exported SHELLOPTS=xtrace switches tracing on for
// the hook's own shell before its first line runs, so the guard turns it back off.
func TestPreCommitHook_DoesNotTraceThePatternsUnderXtrace(t *testing.T) {
	r := newHookRepo(t, keyedBanlist)
	r.write("note.md", "nothing sensitive here\n")
	r.git("add", "note.md")
	r.env = append(r.env, "SHELLOPTS=xtrace")
	blocked, out := r.commit()
	if blocked {
		t.Fatalf("commit refused for clean content\n%s", out)
	}
	if strings.Contains(out, "widgetworks") {
		t.Errorf("an inherited xtrace printed the banlist patterns:\n%s", out)
	}
}
