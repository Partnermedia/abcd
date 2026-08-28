package launch

import "testing"

// TestDirTreeResolveContainsEmbeddedTraversal pins that dirTree.resolve routes
// through the canonical fsutil.ValidRelPath guard: a legitimate repo-relative
// path resolves, and any escaping path — an embedded a/../../b that the old
// hand-rolled "reject absolute or a leading ../" guard let through — is refused
// (iss-2608270655490198).
func TestDirTreeResolveContainsEmbeddedTraversal(t *testing.T) {
	tr := &dirTree{root: t.TempDir()}

	legit := []string{
		"commands/ship.md",
		".claude-plugin/plugin.json",
		"AGENTS.md",
		"a/b/c.md",
	}
	for _, rel := range legit {
		if _, ok := tr.resolve(rel); !ok {
			t.Errorf("resolve(%q) = false, want a legitimate relative path to resolve", rel)
		}
	}

	escaping := []string{
		"a/../../b",      // embedded traversal the old guard missed
		"a/b/../../../c", // deeper embedded traversal
		"../outside.md",  // leading traversal
		"..",
		"/etc/passwd", // absolute
		"",
		"a/./b", // unclean
	}
	for _, rel := range escaping {
		if _, ok := tr.resolve(rel); ok {
			t.Errorf("resolve(%q) = true, want an escaping/unclean path to be refused", rel)
		}
	}
}
