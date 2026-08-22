package lint

import "testing"

// allowComment is the escape the docs-lint bans declare in allow_context.
const allowComment = "<!-- docs-lint: allow -->"

func testTokens(t *testing.T) *TokenChecker {
	t.Helper()
	tc, err := NewTokenChecker([]BannedToken{{
		ID: "present_tense/previously", Pattern: `(?i)\bpreviously\b`, Severity: "blocker",
		Successor: "present-tense phrasing", Message: "change-narration",
		AllowContext: []string{`(?i)<!--\s*docs-lint:\s*allow\b`},
	}})
	if err != nil {
		t.Fatalf("NewTokenChecker: %v", err)
	}
	return tc
}

// TestLintComposedCarriesTheEscapeAcrossTheLift is the composed-span rule: the
// rendered text cannot carry an HTML comment, so the escape is read off the
// SOURCE line the text was selected from.
func TestLintComposedCarriesTheEscapeAcrossTheLift(t *testing.T) {
	tc := testTokens(t)
	source := []string{
		"# Heading",
		"The flag was previously named --old. " + allowComment,
	}
	if got := tc.LintComposed("docs/x.md#heading", "The flag was previously named --old.", false, source); len(got) != 0 {
		t.Errorf("LintComposed findings = %+v, want none (the source line carries the escape)", got)
	}
}

// TestLintComposedWithoutTheEscapeStillFails asserts the lift grants nothing on
// its own: an escape has to be written somewhere.
func TestLintComposedWithoutTheEscapeStillFails(t *testing.T) {
	tc := testTokens(t)
	source := []string{"The flag was previously named --old."}
	got := tc.LintComposed("docs/x.md#heading", "The flag was previously named --old.", false, source)
	if len(got) != 1 {
		t.Fatalf("LintComposed findings = %d, want 1: %+v", len(got), got)
	}
	if got[0].File != "docs/x.md#heading" {
		t.Errorf("File = %q, want the span", got[0].File)
	}
}

// TestLintComposedEscapeIsPerToken asserts the escape excuses the token it was
// written for and nothing else: an allow comment on a line that does not carry
// the matched text grants nothing.
func TestLintComposedEscapeIsPerToken(t *testing.T) {
	tc := testTokens(t)
	source := []string{
		"An unrelated sentence. " + allowComment,
		"The name was previously other.",
	}
	// The escape sits on a line that does not carry the match, so it is not the
	// escape for this match.
	got := tc.LintComposed("docs/x.md#h", "The name was previously other.", false, source)
	if len(got) != 1 {
		t.Fatalf("LintComposed findings = %d, want 1: %+v", len(got), got)
	}
}

// TestLintComposedTakesTheFenceFromItsCaller asserts the code-block fact comes
// from the caller rather than from backticks in the text: composed text has no
// fence markers, and looking for them would find a backtick that is content and
// stop checking from there.
func TestLintComposedTakesTheFenceFromItsCaller(t *testing.T) {
	tc := testTokens(t)
	if got := tc.LintComposed("page", "previously", true, nil); len(got) != 0 {
		t.Errorf("LintComposed findings = %+v, want none (rendered as code)", got)
	}
	if got := tc.LintComposed("page", "previously", false, nil); len(got) != 1 {
		t.Errorf("LintComposed findings = %+v, want one (rendered as prose)", got)
	}
	// A backtick run in prose is content, and must not stop the check.
	if got := tc.LintComposed("page", "``` previously", false, nil); len(got) != 1 {
		t.Errorf("LintComposed findings = %+v, want one; a backtick in composed text is not a fence", got)
	}
}

// TestNewTokenCheckerRefusesWhatTheGateRefuses asserts the exported door
// compiles through the same path the file walk does.
func TestNewTokenCheckerRefusesWhatTheGateRefuses(t *testing.T) {
	if _, err := NewTokenChecker([]BannedToken{{ID: "bad", Pattern: "("}}); err == nil {
		t.Fatal("NewTokenChecker accepted an uncompilable pattern; want refusal")
	}
}

// TestNilTokenCheckerIsInert asserts an unarmed checker is safe to call, so a
// caller need not branch on a repository that declares no bans.
func TestNilTokenCheckerIsInert(t *testing.T) {
	var tc *TokenChecker
	if tc.Len() != 0 || tc.LintComposed("x", "previously", false, nil) != nil {
		t.Error("a nil TokenChecker reported something")
	}
}
