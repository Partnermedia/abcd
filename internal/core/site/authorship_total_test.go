package site

import (
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// TestAuthorshipBarSumEqualsAssistedTotal pins the data invariant behind
// iss-2608231008315498: the Assisted-by panel's total (a.Assisted, the number it
// renders as the panel figure) MUST equal the sum of the per-model bars (ByModel),
// and the human-only `Assisted-by: None` declaration must never appear as a bar.
// The defect the issue reported was a note reading one total while the bars summed
// to another, because a `None` row was counted in one place and excluded in the
// other; the chart total and its bars have to agree. `None` is instead surfaced as
// its own DeclaredNone figure.
func TestAuthorshipBarSumEqualsAssistedTotal(t *testing.T) {
	r := gittest.NewRepo(t)
	// A controlled history: two single-model assisted commits, one commit naming two
	// models (two occurrences), two human-only `None` declarations, and one commit
	// with no trailer at all.
	r.Commit("feat: one\n\nAssisted-by: model-a")
	r.Commit("feat: two\n\nAssisted-by: model-b")
	r.Commit("feat: three\n\nAssisted-by: model-a\nAssisted-by: model-b")
	r.Commit("chore: human one\n\nAssisted-by: None")
	r.Commit("chore: human two\n\nAssisted-by: None")
	r.Commit("chore: undeclared")

	a, err := LoadAuthorship(r.Root())
	if err != nil {
		t.Fatalf("LoadAuthorship: %v", err)
	}

	barSum := 0
	for _, m := range a.ByModel {
		if m.Model == noneDeclaration {
			t.Fatalf("the None declaration must never be charted as a model bar: %+v", a.ByModel)
		}
		barSum += m.Commits
	}
	// The rendered panel total is a.Assisted; it must equal what the bars sum to.
	if barSum != a.Assisted {
		t.Fatalf("panel total (a.Assisted=%d) disagrees with the bar sum (%d); the note and the bars must match", a.Assisted, barSum)
	}
	// Sanity on the shape: 4 assisted occurrences (a, b, a, b), 2 None-only commits,
	// 1 undeclared.
	if a.Assisted != 4 {
		t.Errorf("assisted occurrences = %d, want 4", a.Assisted)
	}
	if a.DeclaredNone != 2 {
		t.Errorf("declared-none commits = %d, want 2", a.DeclaredNone)
	}
	if a.Undeclared != 1 {
		t.Errorf("undeclared commits = %d, want 1", a.Undeclared)
	}
}
