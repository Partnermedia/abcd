package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// linkedWorktree seeds a commit in repo and adds a linked worktree, returning its
// root. It is the shape itd-150 exists for: one repository, two working trees, and
// a gitignored local-ephemeral tier that belongs to whichever one you stand in.
func linkedWorktree(t *testing.T, repo *gittest.Repo) string {
	t.Helper()
	repo.Git("add", "-A")
	repo.Git("-c", "core.hooksPath=/dev/null", "commit", "-m", "seed")
	dir := filepath.Join(t.TempDir(), "linked")
	repo.Git("worktree", "add", "-b", "linked", dir)
	return dir
}

// TestBanlistRendersTheLayerALinkedWorktreeInherits is the CLI half of itd-150's
// parity requirement. The guard resolves the primary checkout's store and enforces
// it inside a linked worktree; a status render that showed only the (absent)
// worktree-local store would announce "INACTIVE on this machine" about a checkout
// whose every commit is being checked — the board disagreeing with the guard about
// what is protected.
func TestBanlistRendersTheLayerALinkedWorktreeInherits(t *testing.T) {
	hermeticEnv(t)
	repo := gittest.NewRepo(t)
	// The tier's fence first: the private store is only written where git already
	// ignores it, and the linked worktree inherits the fence through the commit.
	repo.Write(".gitignore", ".abcd/.work.local/\n")
	t.Chdir(repo.Root())
	if _, err := runCLIErr(t, "banlist", "add", "--private", "widget-partner", "widgetworks"); err != nil {
		t.Fatalf("banlist add: %v", err)
	}
	linked := linkedWorktree(t, repo)
	t.Chdir(linked)

	out := string(runCLI(t, "banlist"))
	if !strings.Contains(out, "inherited from the primary checkout") {
		t.Fatalf("the worktree render does not name the inherited layer:\n%s", out)
	}
	if !strings.Contains(out, "widget-partner") {
		t.Errorf("the inherited layer's key is not rendered:\n%s", out)
	}
	if strings.Contains(out, "widgetworks") {
		t.Errorf("the render leaks the pattern; private entries render by key only:\n%s", out)
	}

	// `list --private` is the same read scoped to one layer, so it must agree.
	scoped := string(runCLI(t, "banlist", "list", "--private"))
	if !strings.Contains(scoped, "inherited from the primary checkout") {
		t.Errorf("`banlist list --private` omits the inherited layer:\n%s", scoped)
	}
}

// TestBanlistJSONCarriesTheInheritedLayerOnlyInAWorktree pins both halves of the
// envelope contract: a linked worktree carries the inherited layer, and a
// standalone checkout's JSON is exactly what it always was — the field is absent,
// not present-and-null, so no consumer has to learn a new shape it will never see.
func TestBanlistJSONCarriesTheInheritedLayerOnlyInAWorktree(t *testing.T) {
	hermeticEnv(t)
	repo := gittest.NewRepo(t)
	// The tier's fence first: the private store is only written where git already
	// ignores it, and the linked worktree inherits the fence through the commit.
	repo.Write(".gitignore", ".abcd/.work.local/\n")
	t.Chdir(repo.Root())
	if _, err := runCLIErr(t, "banlist", "add", "--private", "widget-partner", "widgetworks"); err != nil {
		t.Fatalf("banlist add: %v", err)
	}

	var standalone map[string]any
	if err := json.Unmarshal(runCLI(t, "--json", "banlist"), &standalone); err != nil {
		t.Fatalf("standalone JSON: %v", err)
	}
	if _, ok := standalone["inherited"]; ok {
		t.Errorf("a standalone checkout's envelope carries an inherited layer: %v", standalone["inherited"])
	}

	linked := linkedWorktree(t, repo)
	t.Chdir(linked)
	var worktree map[string]any
	if err := json.Unmarshal(runCLI(t, "--json", "banlist"), &worktree); err != nil {
		t.Fatalf("worktree JSON: %v", err)
	}
	inh, ok := worktree["inherited"].(map[string]any)
	if !ok {
		t.Fatalf("a linked worktree's envelope carries no inherited layer: %v", worktree)
	}
	if inh["primary_root"] == "" || inh["primary_root"] == nil {
		t.Errorf("the inherited layer does not name the primary checkout: %v", inh)
	}
	if !strings.Contains(string(runCLI(t, "--json", "banlist")), "widget-partner") {
		t.Errorf("the inherited layer's key is missing from the envelope")
	}
}
