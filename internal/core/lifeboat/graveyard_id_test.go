package lifeboat

import "testing"

// TestIdCleanIsInjective proves distinct id components never collide after
// cleaning. The previous implementation DELETED spaces/control chars, so
// "my file" and "myfile" (and two branch names differing only in whitespace)
// produced the same id and one finding silently shadowed the other.
func TestIdCleanIsInjective(t *testing.T) {
	inputs := []string{
		"my file", "myfile", "a b", "ab", "a  b", "tab\tsep", "tabsep",
		"pct%20", "pct 20", "trailing ", "trailing",
	}
	seen := map[string]string{}
	for _, in := range inputs {
		got := idClean(in)
		if prev, ok := seen[got]; ok && prev != in {
			t.Fatalf("idClean collision: %q and %q both -> %q", prev, in, got)
		}
		seen[got] = in
	}
}

// TestIdCleanStableForOrdinaryPaths proves the common case — a path with no
// control chars, spaces, or '%' — is unchanged, so existing ids stay stable
// across the re-plan invariant.
func TestIdCleanStableForOrdinaryPaths(t *testing.T) {
	for _, p := range []string{"src/main.go", "docs/how-to/thing.md", "iss-7-slug"} {
		if got := idClean(p); got != p {
			t.Errorf("idClean(%q) = %q; ordinary path must pass through unchanged", p, got)
		}
	}
}

// TestDeletedPathIDDistinct is the concrete graveyard symptom: two distinct
// deleted paths must key distinct findings.
func TestDeletedPathIDDistinct(t *testing.T) {
	if deletedPathID("notes draft.md") == deletedPathID("notesdraft.md") {
		t.Error("distinct deleted paths collapsed onto one graveyard finding id")
	}
}
