package launch

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/gittest"
)

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func included(b Bundle, logical string) bool {
	for _, f := range b.Included {
		if f.LogicalPath == logical {
			return true
		}
	}
	return false
}

func excludedReason(b Bundle, logical string) (ExcludedReason, bool) {
	for _, f := range b.Excluded {
		if f.LogicalPath == logical {
			return f.Reason, true
		}
	}
	return "", false
}

// TestDefaultDenyNewTopLevel proves the include list is a default-deny
// allowlist: a new top-level path not named by any include never enters the
// bundle, and no .abcd/** path is ever Included.
func TestDefaultDenyNewTopLevel(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "commands/a.md", "ok")
	writeFile(t, root, "docs/b.md", "ok")
	writeFile(t, root, "README.md", "ok")
	writeFile(t, root, ".abcd/secret.md", "SECRET")
	writeFile(t, root, ".work/scratch.txt", "junk")
	writeFile(t, root, "surprise/c.md", "new top-level dir")

	b, err := ResolveBundle(root, []string{"commands", "docs", "README.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !included(b, "commands/a.md") || !included(b, "docs/b.md") || !included(b, "README.md") {
		t.Fatalf("expected allowlisted files included: %+v", b.Included)
	}
	// Default-deny: a new top-level dir not in the includes must not be included.
	if included(b, "surprise/c.md") {
		t.Errorf("default-deny violated: unlisted top-level dir was included")
	}
	// The .abcd namespace must NEVER appear in Included under any circumstances.
	for _, f := range b.Included {
		if firstSegment(f.LogicalPath) == ".abcd" {
			t.Errorf(".abcd path entered the bundle: %s", f.LogicalPath)
		}
	}
}

// TestAbcdNamespaceStructurallyExcluded proves that even a broad include that
// reaches everything classifies .abcd/** as excluded(denied_namespace) and never
// Included — the load-bearing structural deny (adr-18/adr-28).
func TestAbcdNamespaceStructurallyExcluded(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "commands/a.md", "ok")
	writeFile(t, root, "README.md", "ok")
	writeFile(t, root, ".abcd/development/brief.md", "DESIGN RECORD")
	writeFile(t, root, ".work/x", "junk")
	writeFile(t, root, ".flow/y", "junk")

	b, err := ResolveBundle(root, []string{"."})
	if err != nil {
		t.Fatal(err)
	}
	// Broad include still ships the real content.
	if !included(b, "commands/a.md") || !included(b, "README.md") {
		t.Fatalf("broad include dropped real content: %+v", b.Included)
	}
	// Every denied namespace is excluded(denied_namespace), never Included.
	for _, ns := range []string{".abcd", ".work", ".flow"} {
		if reason, ok := excludedReason(b, ns); !ok || reason != ExcludedDeniedNamespace {
			t.Errorf("%s not excluded(denied_namespace): reason=%q ok=%v", ns, reason, ok)
		}
	}
	for _, f := range b.Included {
		if _, denied := DenyNamespaces[firstSegment(f.LogicalPath)]; denied {
			t.Errorf("denied namespace path entered the bundle: %s", f.LogicalPath)
		}
	}
}

// TestMissingLiteralRejected proves a literal include with no on-disk entry is a
// rejection.
func TestMissingLiteralRejected(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "ok")
	b, err := ResolveBundle(root, []string{"README.md", "LICENSE"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range b.Rejected {
		if r.LogicalPath == "LICENSE" && r.Reason == RejectedMissingLiteral {
			found = true
		}
	}
	if !found {
		t.Errorf("missing literal LICENSE not rejected: %+v", b.Rejected)
	}
	if b.HasViolation() != true {
		t.Errorf("HasViolation should be true with a rejection")
	}
}

// TestSymlinkEscapeRejected proves a symlink whose target escapes the repo is
// rejected, and a symlink into a denied namespace is rejected(deny).
func TestSymlinkEscapeRejected(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeFile(t, outside, "leak.txt", "external")
	writeFile(t, root, ".abcd/secret.txt", "SECRET")
	writeFile(t, root, "commands/keep.md", "ok")

	// commands/escape -> outside/leak.txt (escape)
	if err := os.Symlink(filepath.Join(outside, "leak.txt"), filepath.Join(root, "commands", "escape")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	// commands/smuggle -> ../.abcd/secret.txt (denied target)
	if err := os.Symlink(filepath.Join(root, ".abcd", "secret.txt"), filepath.Join(root, "commands", "smuggle")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	b, err := ResolveBundle(root, []string{"commands"})
	if err != nil {
		t.Fatal(err)
	}
	var escape, deny bool
	for _, r := range b.Rejected {
		if r.LogicalPath == "commands/escape" && r.Reason == RejectedSymlinkEscape {
			escape = true
		}
		if r.LogicalPath == "commands/smuggle" && r.Reason == RejectedDeny {
			deny = true
		}
	}
	if !escape {
		t.Errorf("symlink escape not rejected: %+v", b.Rejected)
	}
	if !deny {
		t.Errorf("symlink into denied namespace not rejected(deny): %+v", b.Rejected)
	}
	// The real file still ships; no denied content leaked in.
	if !included(b, "commands/keep.md") {
		t.Errorf("real file dropped: %+v", b.Included)
	}
}

// TestSymlinkToRepoRootDoesNotLeakDenied proves that a symlink to (or into) the
// repo root cannot smuggle a denied namespace out through a dereferenced walk:
// the structural deny is re-applied to the REAL path at every level, so
// .abcd/** reached via docs/all -> <root> is rejected(deny), not shipped.
func TestSymlinkToRepoRootDoesNotLeakDenied(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/secret.txt", "SECRET")
	writeFile(t, root, "commands/keep.md", "ok")
	writeFile(t, root, "docs/readme.md", "ok")

	if err := os.Symlink(root, filepath.Join(root, "docs", "all")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	b, err := ResolveBundle(root, []string{"docs"})
	if err != nil {
		t.Fatal(err)
	}
	for _, inc := range b.Included {
		if strings.Contains(inc.ResolvedPath, filepath.Join(root, ".abcd")+string(filepath.Separator)) {
			t.Errorf("denied .abcd content leaked via symlink walk: %+v", inc)
		}
		if strings.Contains(inc.LogicalPath, "/.abcd/") {
			t.Errorf("denied logical path shipped: %+v", inc)
		}
	}
	var denied bool
	for _, r := range b.Rejected {
		if r.Reason == RejectedDeny && strings.Contains(r.LogicalPath, ".abcd") {
			denied = true
		}
	}
	if !denied {
		t.Errorf("expected rejected(deny) for the .abcd namespace reached via symlink: %+v", b.Rejected)
	}
}

// TestGitignoredSymlinkTargetExcluded proves a symlink whose TARGET is gitignored
// is excluded even though the symlink's own (logical) name is not ignored — the
// target is the content that would actually ship.
func TestGitignoredSymlinkTargetExcluded(t *testing.T) {
	root := t.TempDir()
	gitInit := exec.Command("git", "-C", root, "init")
	gitInit.Env = gittest.Env(t)
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Skipf("git init unavailable: %v (%s)", err, out)
	}
	writeFile(t, root, ".gitignore", ".env\n")
	writeFile(t, root, ".env", "SECRET=1")
	writeFile(t, root, "docs/readme.md", "ok")
	if err := os.Symlink(filepath.Join(root, ".env"), filepath.Join(root, "docs", "alias")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	b, err := ResolveBundle(root, []string{"docs"})
	if err != nil {
		t.Fatal(err)
	}
	for _, inc := range b.Included {
		if inc.LogicalPath == "docs/alias" {
			t.Errorf("gitignored symlink target shipped: %+v", inc)
		}
	}
	var excluded bool
	for _, e := range b.Excluded {
		if e.LogicalPath == "docs/alias" && e.Reason == ExcludedGitignored {
			excluded = true
		}
	}
	if !excluded {
		t.Errorf("expected docs/alias excluded(gitignored): excluded=%+v included=%+v", b.Excluded, b.Included)
	}
}

// TestCheckIgnoredStrictIgnoresInheritedWorkTree is the attack-input test for the
// os.Environ()->gitutil.IsolatedEnv() scrub: an inherited GIT_WORK_TREE must not
// redirect the gitignore probe at a different tree. check-ignore resolves its
// .gitignore against the WORKING TREE, so an inherited GIT_WORK_TREE pointing at a
// tree with no matching rule (verified empirically to override `-C root`) makes a
// gitignored secret read as "not ignored" and promoted into the release. The scrub
// strips GIT_WORK_TREE, so the probe answers about `root`.
func TestCheckIgnoredStrictIgnoresInheritedWorkTree(t *testing.T) {
	root := t.TempDir()
	gitInit := exec.Command("git", "-C", root, "init")
	gitInit.Env = gittest.Env(t)
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Skipf("git init unavailable: %v (%s)", err, out)
	}
	writeFile(t, root, ".gitignore", ".env\n")
	writeFile(t, root, ".env", "SECRET=1")

	// A second tree with no .gitignore, pointed at via an inherited GIT_WORK_TREE.
	other := t.TempDir()
	t.Setenv("GIT_WORK_TREE", other)

	ignored, err := checkIgnoredStrict(root, []string{".env"})
	if err != nil {
		t.Fatalf("checkIgnoredStrict: %v", err)
	}
	if _, ok := ignored[".env"]; !ok {
		t.Error(".env read as NOT ignored under an inherited GIT_WORK_TREE — the probe was redirected at another tree; a gitignored secret would ship")
	}
}

// TestGlobMatchSegmentAware proves ** crosses separators but a single * stays
// in-segment.
func TestGlobMatchSegmentAware(t *testing.T) {
	cases := []struct {
		rel, pattern string
		want         bool
	}{
		{"docs/a.md", "docs/**", true},
		{"docs/sub/a.md", "docs/**", true},
		{"docs", "docs/**", true},
		{"a.md", "*.md", true},
		{"sub/a.md", "*.md", false}, // single * must not cross /
		{"sub/a.md", "**/*.md", true},
		{"README.md", "**/*.md", true},
		{"x.txt", "*.md", false},
	}
	for _, c := range cases {
		if got := matchesInclude(c.rel, c.pattern); got != c.want {
			t.Errorf("matchesInclude(%q,%q)=%v want %v", c.rel, c.pattern, got, c.want)
		}
	}
}

// TestGitignoreProbeFailsClosed guards B18: when the gitignore probe cannot
// answer (git present, root is a repo, but check-ignore itself errors), the gate
// must NOT promote unproven survivors to Included — it fails closed, rejecting
// them, mirroring the uncertain-inode-map gate. The probe is stubbed to simulate
// the real git failure (otherwise hard to trigger deterministically).
func TestGitignoreProbeFailsClosed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "docs/readme.md", "ok")

	orig := ignoreChecker
	ignoreChecker = func(string, []string) (map[string]struct{}, error) {
		return nil, fmt.Errorf("simulated git check-ignore failure")
	}
	defer func() { ignoreChecker = orig }()

	b, err := ResolveBundle(root, []string{"docs"})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Included) != 0 {
		t.Errorf("fail-open: files shipped despite an unanswerable gitignore probe: %+v", b.Included)
	}
	var rejected bool
	for _, r := range b.Rejected {
		if r.LogicalPath == "docs/readme.md" && r.Reason == RejectedFSError && r.Details["kind"] == "gitignore_check_failed" {
			rejected = true
		}
	}
	if !rejected {
		t.Errorf("expected docs/readme.md rejected(fs_error, gitignore_check_failed): %+v", b.Rejected)
	}
}

// TestCheckIgnoredStrictNonRepoIsBenign guards B18's other half: a directory that
// is simply not a git repo (or git absent) carries no gitignore semantics and
// must still resolve — an empty set with no error, never a fail-closed rejection.
func TestCheckIgnoredStrictNonRepoIsBenign(t *testing.T) {
	root := t.TempDir() // a plain dir, deliberately not `git init`ed
	got, err := checkIgnoredStrict(root, []string{"docs/readme.md"})
	if err != nil {
		t.Fatalf("non-repo dir must carry no gitignore semantics, got error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("non-repo dir reported ignored paths: %+v", got)
	}
	// And a full resolution over a non-repo dir still ships normally.
	writeFile(t, root, "docs/readme.md", "ok")
	b, err := ResolveBundle(root, []string{"docs"})
	if err != nil {
		t.Fatal(err)
	}
	var shipped bool
	for _, inc := range b.Included {
		if inc.LogicalPath == "docs/readme.md" {
			shipped = true
		}
	}
	if !shipped {
		t.Errorf("non-repo dir failed to ship docs/readme.md: included=%+v rejected=%+v", b.Included, b.Rejected)
	}
}

// TestMalformedGlobIncludeIsPreflightError guards B19: a malformed char-class
// glob (invalid range [z-a]) in the include config used to panic ResolveBundle
// via regexp.MustCompile; it must now surface as a graceful PreflightError.
func TestMalformedGlobIncludeIsPreflightError(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "docs/readme.md", "ok")

	_, err := ResolveBundle(root, []string{"docs/[z-a]*.md"})
	if err == nil {
		t.Fatal("expected a preflight error for a malformed char-class glob, got nil")
	}
	var pf *PreflightError
	if !errors.As(err, &pf) {
		t.Fatalf("expected *PreflightError, got %T: %v", err, err)
	}
}
