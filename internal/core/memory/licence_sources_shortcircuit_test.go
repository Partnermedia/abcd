package memory

import (
	"strings"
	"testing"
)

// TestLicenceGateSurvivesAnEmptySourcesList pins ML001 against the second way
// the blocker was silently defeated: checkLicence treated "a sources list is
// present" as proof that no page-level class needed checking and returned
// unconditionally out of that branch. The two shapes are mutually exclusive only
// on the write path (validateSourceBlock), which the lint path never runs — so a
// page carrying BOTH a scalar `class: external_pdf` and an empty or junk
// `sources` list entered the loop, matched nothing, and returned past the
// page-level licence check.
func TestLicenceGateSurvivesAnEmptySourcesList(t *testing.T) {
	hash := strings.Repeat("a", 64)

	cases := []struct {
		name string
		page string
	}{
		{
			name: "empty_sources_list",
			page: "---\nsource:\n  class: external_pdf\n  sources: []\n  source_hash: " + hash + "\n---\n# Body\n",
		},
		{
			name: "non_map_sources_entries",
			page: "---\nsource:\n  class: external_pdf\n  sources: [foo]\n  source_hash: " + hash + "\n---\n# Body\n",
		},
		{
			name: "plural_classes_with_empty_sources",
			page: "---\nsource:\n  classes: [external_pdf]\n  sources: []\n  source_hash: " + hash + "\n---\n# Body\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lr := lintPage(t, "note_privacy_gdpr.md", tc.page)
			if !hasBlocker(lr, "ML001") || lr.ExitCode != 1 {
				t.Fatalf("unlicensed external page passed lint behind a %s: exit=%d findings=%+v", tc.name, lr.ExitCode, lr.Findings)
			}
		})
	}

	// Control: a well-formed multi-source page licences its sources per entry and
	// carries no page-level `licence` — it must NOT collect a page-level ML001
	// for the classes those entries already account for.
	t.Run("multi_source_covered_by_entries", func(t *testing.T) {
		page := "---\nsource:\n  classes: [external_pdf, external_transcript]\n  weighting_note: pdf outweighs transcript\n  sources:\n" +
			"    -\n      class: external_pdf\n      citation: { type: knowledge }\n      licence: MIT\n      source_hash: " + strings.Repeat("c", 64) + "\n      ingested_at: 2026-07-06\n" +
			"    -\n      class: external_transcript\n      citation: { type: knowledge }\n      licence: MIT\n      source_hash: " + strings.Repeat("d", 64) + "\n      ingested_at: 2026-07-06\n" +
			"---\n# Body\n"
		if lr := lintPage(t, "note_privacy_gdpr.md", page); hasBlocker(lr, "ML001") {
			t.Fatalf("a fully licensed multi-source page collected a page-level ML001: %+v", lr.Findings)
		}
	})

	// Control: an entry missing its licence is still caught, and exactly once.
	t.Run("entry_missing_licence_still_caught", func(t *testing.T) {
		page := "---\nsource:\n  classes: [external_pdf]\n  sources:\n" +
			"    -\n      class: external_pdf\n      citation: { type: knowledge }\n      source_hash: " + strings.Repeat("c", 64) + "\n      ingested_at: 2026-07-06\n" +
			"---\n# Body\n"
		lr := lintPage(t, "note_privacy_gdpr.md", page)
		count := 0
		for _, f := range lr.Findings {
			if f.Code == "ML001" {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("want exactly one ML001 for the unlicensed entry, got %d: %+v", count, lr.Findings)
		}
	})
}
