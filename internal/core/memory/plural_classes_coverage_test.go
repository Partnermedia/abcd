package memory

import (
	"strings"
	"testing"
)

// TestExternalSourceHashesReadsPluralClasses (iss-2608270908341877) proves a
// plural-shape page — `classes: [external_pdf]` with a single top-level
// `source_hash` — contributes its hash to quotation-coverage accounting.
// externalSourceHashes read only the scalar `source.class`, so the plural
// `classes` list was invisible: the page's source_hash dropped out of coverage
// even as the lint's derivedClasses counted the external class, the coverage
// sibling of the resolved plural-classes licence gap.
func TestExternalSourceHashesReadsPluralClasses(t *testing.T) {
	hash := strings.Repeat("a", 64)

	// Plural `classes` list, no scalar `class`, single top-level source_hash.
	plural := map[string]any{
		"classes":     []any{"external_pdf"},
		"source_hash": hash,
	}
	got := externalSourceHashes(plural)
	if len(got) != 1 || got[0] != hash {
		t.Fatalf("externalSourceHashes(plural classes) = %v, want [%s]; a plural-shape external page's source_hash must reach coverage", got, hash)
	}

	// Both shapes present, only the plural one external: still counts.
	both := map[string]any{
		"class":       "session_memory",
		"classes":     []any{"session_memory", "external_pdf"},
		"source_hash": hash,
	}
	if got := externalSourceHashes(both); len(got) != 1 || got[0] != hash {
		t.Fatalf("externalSourceHashes(both shapes) = %v, want [%s]", got, hash)
	}

	// Control: a purely internal plural page contributes nothing.
	internal := map[string]any{
		"classes":     []any{"session_memory"},
		"source_hash": hash,
	}
	if got := externalSourceHashes(internal); len(got) != 0 {
		t.Fatalf("externalSourceHashes(internal plural) = %v, want none", got)
	}
}

// TestSourceClassesUnionsBothShapes (iss-2608270945468534) proves SourceClasses
// returns the UNION of the scalar `class` and the plural `classes` list when both
// are present — the same derivation the lint makes (derivedClasses). It used to
// return the scalar class alone, the opposite direction, so a both-shapes page
// rendered under only its scalar class in the generated index and write log while
// lint counted the union.
func TestSourceClassesUnionsBothShapes(t *testing.T) {
	src := map[string]any{
		"class":   "external_pdf",
		"classes": []any{"session_memory"},
	}
	got := SourceClasses(src)
	want := []string{"external_pdf", "session_memory"}
	if len(got) != len(want) {
		t.Fatalf("SourceClasses(both shapes) = %v, want %v (the union, matching the lint derivation)", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SourceClasses(both shapes) = %v, want %v", got, want)
		}
	}

	// A class named by both shapes appears once (deduplicated).
	dup := map[string]any{
		"class":   "external_pdf",
		"classes": []any{"external_pdf", "session_memory"},
	}
	if got := SourceClasses(dup); len(got) != 2 {
		t.Fatalf("SourceClasses dedup = %v, want two distinct classes", got)
	}
}
