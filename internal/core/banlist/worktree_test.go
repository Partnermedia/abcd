package banlist

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// worktreePair initialises a repo, seeds a commit, and adds a linked worktree,
// returning both roots. It is the shape itd-150 exists for: one repository, two
// working trees, and a local-ephemeral tier that belongs to whichever one you are
// standing in.
func worktreePair(t *testing.T) (primary, linked string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	primary = t.TempDir()
	env := gittest.Env(t)
	git := func(dir string, args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	for _, args := range [][]string{
		{"init"},
		{"config", "user.name", "Alice Example"},
		{"config", "user.email", "alice@example.com"},
	} {
		if out, err := git(primary, args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(primary, ".gitignore"), []byte(PrivateDirRelPath+"/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := git(primary, "add", "-A"); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	if out, err := git(primary, "-c", "core.hooksPath=/dev/null", "commit", "-m", "seed"); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	linked = filepath.Join(t.TempDir(), "linked")
	if out, err := git(primary, "worktree", "add", "-b", "linked", linked); err != nil {
		t.Skipf("git worktree add unavailable: %v\n%s", err, out)
	}
	return primary, linked
}

// TestPrimaryWorktreeRootResolvesOnlyFromALinkedWorktree pins both directions of
// the resolution the guard makes: a linked worktree finds its primary checkout, and
// a standalone checkout finds nothing — the second is what keeps the fallback a
// fallback rather than a second store every repo suddenly acquires.
func TestPrimaryWorktreeRootResolvesOnlyFromALinkedWorktree(t *testing.T) {
	primary, linked := worktreePair(t)

	got, ok := PrimaryWorktreeRoot(linked)
	if !ok {
		t.Fatal("a linked worktree did not resolve its primary checkout")
	}
	// Compared through EvalSymlinks: git answers with the resolved path, and a temp
	// dir on macOS sits behind /private.
	if resolve(t, got) != resolve(t, primary) {
		t.Errorf("primary root = %q, want %q", got, primary)
	}

	if got, ok := PrimaryWorktreeRoot(primary); ok {
		t.Errorf("the primary checkout resolved a primary of its own (%q); it must inherit nothing", got)
	}
	if got, ok := PrimaryWorktreeRoot(t.TempDir()); ok {
		t.Errorf("a directory that is not a repository resolved a primary (%q)", got)
	}
}

func resolve(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p
	}
	return r
}

// TestInheritedPrivateReportsThePrimaryStore pins the read side of the parity the
// spec requires: `abcd banlist` in a linked worktree must render the layer the
// guard actually enforces there. Without it the board reports the worktree
// INACTIVE while every commit made in it is checked against the primary's list —
// a status surface that disagrees with the guard about what is protected.
func TestInheritedPrivateReportsThePrimaryStore(t *testing.T) {
	primary, linked := worktreePair(t)

	// Nothing to inherit yet: the primary has no store either.
	inh, err := InheritedPrivate(linked)
	if err != nil {
		t.Fatal(err)
	}
	if inh != nil && inh.Private.Present {
		t.Errorf("an absent primary store was reported present")
	}

	if _, err := AddPrivate(AddPrivateRequest{RepoRoot: primary, Key: "widget-partner", Pattern: "widgetworks"}); err != nil {
		t.Fatal(err)
	}
	inh, err = InheritedPrivate(linked)
	if err != nil {
		t.Fatal(err)
	}
	if inh == nil {
		t.Fatal("a linked worktree inherited nothing from a primary checkout that has a store")
	}
	if !inh.Private.Present {
		t.Error("the inherited layer is not reported present")
	}
	if len(inh.Private.Entries) != 1 || inh.Private.Entries[0].Key != "widget-partner" {
		t.Errorf("inherited entries = %+v, want the primary's one key", inh.Private.Entries)
	}
	if resolve(t, inh.PrimaryRoot) != resolve(t, primary) {
		t.Errorf("PrimaryRoot = %q, want %q", inh.PrimaryRoot, primary)
	}

	// A standalone checkout inherits nothing, whatever its own store holds.
	if inh, err := InheritedPrivate(primary); err != nil || inh != nil {
		t.Errorf("the primary checkout inherited %+v (err=%v); it must inherit nothing", inh, err)
	}
}

// TestListCarriesTheInheritedLayer pins that the one read every status surface
// makes carries the inherited layer, so no surface has to remember to ask a second
// question — the way a surface ends up disagreeing with the guard.
func TestListCarriesTheInheritedLayer(t *testing.T) {
	primary, linked := worktreePair(t)
	if _, err := AddPrivate(AddPrivateRequest{RepoRoot: primary, Key: "widget-partner", Pattern: "widgetworks"}); err != nil {
		t.Fatal(err)
	}
	rep, err := List(linked)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Inherited == nil {
		t.Fatal("List did not carry the inherited layer in a linked worktree")
	}
	if len(rep.Inherited.Private.Entries) != 1 {
		t.Errorf("inherited entries = %+v, want one", rep.Inherited.Private.Entries)
	}
	rep, err = List(primary)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Inherited != nil {
		t.Errorf("List carried an inherited layer in a standalone checkout: %+v", rep.Inherited)
	}
}
