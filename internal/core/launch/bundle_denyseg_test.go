package launch

import "testing"

// TestNestedDeniedNamespaceExcluded is the GHSA-g2v7-wfmv-v28r nested-segment
// axis: a denied namespace nested UNDER an included tree (docs/.abcd/…,
// docs/.git/…) must be excluded(denied_namespace) and never descended, even
// though its FIRST path segment (docs) is a legitimate include. The old
// first-segment-only lookup shipped these.
func TestNestedDeniedNamespaceExcluded(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "docs/readme.md", "ok")
	writeFile(t, root, "docs/.abcd/planted.txt", "not-a-real-secret")
	writeFile(t, root, "docs/.git/config", "[core]\n\turl = https://user:token@host\n")
	writeFile(t, root, "README.md", "ok")

	b, err := ResolveBundle(root, []string{"docs", "commands", "README.md"})
	if err != nil {
		t.Fatal(err)
	}

	// The nested denied files must NEVER be Included.
	if included(b, "docs/.abcd/planted.txt") {
		t.Errorf("nested denied namespace shipped: docs/.abcd/planted.txt entered the bundle")
	}
	if included(b, "docs/.git/config") {
		t.Errorf("nested denied namespace shipped: docs/.git/config entered the bundle")
	}
	// No Included path may contain a denied segment anywhere.
	for _, f := range b.Included {
		if pathContainsDeniedSegment(f.LogicalPath) {
			t.Errorf("denied namespace path entered the bundle: %s", f.LogicalPath)
		}
	}
	// The nested denied dirs surface as excluded(denied_namespace) because an
	// include (docs) reaches them.
	for _, dir := range []string{"docs/.abcd", "docs/.git"} {
		if reason, ok := excludedReason(b, dir); !ok || reason != ExcludedDeniedNamespace {
			t.Errorf("%s not excluded(denied_namespace): reason=%q ok=%v", dir, reason, ok)
		}
	}
	// The legitimate content is still shipped.
	if !included(b, "docs/readme.md") || !included(b, "README.md") {
		t.Errorf("legitimate content dropped: %+v", b.Included)
	}
}

// TestCaseFoldDeniedNamespaceExcluded is the GHSA-g2v7-wfmv-v28r / #335
// case-fold axis: a top-level case variant of a denied namespace (.ABCD, .Git,
// MEMORY) must be excluded, not shipped. The old exact-case map lookup shipped
// these.
func TestCaseFoldDeniedNamespaceExcluded(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "ok")
	writeFile(t, root, ".ABCD/secret.md", "SECRET")
	writeFile(t, root, ".Git/config", "junk")
	writeFile(t, root, "MEMORY/notes.md", "junk")

	b, err := ResolveBundle(root, []string{"."})
	if err != nil {
		t.Fatal(err)
	}

	if included(b, ".ABCD/secret.md") {
		t.Errorf("case-variant denied namespace shipped: .ABCD/secret.md entered the bundle")
	}
	for _, f := range b.Included {
		if pathContainsDeniedSegment(f.LogicalPath) {
			t.Errorf("denied namespace path entered the bundle: %s", f.LogicalPath)
		}
	}
	for _, dir := range []string{".ABCD", ".Git", "MEMORY"} {
		if reason, ok := excludedReason(b, dir); !ok || reason != ExcludedDeniedNamespace {
			t.Errorf("%s not excluded(denied_namespace): reason=%q ok=%v", dir, reason, ok)
		}
	}
	if !included(b, "README.md") {
		t.Errorf("legitimate content dropped: %+v", b.Included)
	}
}
