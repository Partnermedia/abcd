package capture

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testGrounds is a conjecture-shaped operand for the tests whose subject is
// something other than the grounds argument itself.
const testGrounds = "pursued: we expect the ledger to keep the reasoning the session would otherwise lose"

// groundsScalar re-reads an issue's raw `grounds:` frontmatter line, so a test
// asserts on what was WRITTEN rather than on what a reader chose to surface.
func groundsScalar(t *testing.T, ir, issID string) string {
	t.Helper()
	path, _, err := findIssue(ir, issID)
	if err != nil {
		t.Fatalf("findIssue(%s): %v", issID, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, ln := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(ln, "grounds:") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(ln, "grounds:")), `"`)
		}
	}
	return ""
}

// TestPromoteRefusesWithoutGrounds is itd-179's first acceptance criterion: a
// capture routed to an intent draft is a conjecture being pursued, and triaging
// it without saying why is exactly the evaporation the argument closes.
func TestPromoteRefusesWithoutGrounds(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "the loader drops rules silently when the config is stale")

	if _, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID}); err == nil {
		t.Fatal("Promote without grounds = nil error, want a refusal")
	}
	for _, bad := range []string{"planned: out of vocabulary", "pursued: short", "pursued"} {
		if _, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID, Grounds: bad}); err == nil {
			t.Fatalf("Promote with grounds %q = nil error, want a refusal", bad)
		}
	}
}

// TestPromoteWithoutGroundsWritesNothing: the refusal lands BEFORE the mint, so
// a missing argument never leaves an orphan draft behind for somebody to clean
// up — the residue the promote path already works hard to avoid.
func TestPromoteWithoutGroundsWritesNothing(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "the loader drops rules silently when the config is stale")

	if _, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID}); err == nil {
		t.Fatal("want a refusal")
	}
	if n := draftCount(t, repo); n != 0 {
		t.Fatalf("a refused promote minted %d draft(s), want 0", n)
	}
	if iss := readIssue(t, ir, issID); iss.PromotedTo != "" {
		t.Fatalf("a refused promote stamped promoted_to = %q", iss.PromotedTo)
	}
	if g := groundsScalar(t, ir, issID); g != "" {
		t.Fatalf("a refused promote stamped grounds = %q", g)
	}
}

// TestPromoteStampsGrounds: the accepted value lands on the record as one plain
// YAML scalar in the shared grammar, readable by every store's frontmatter
// reader with no second parser.
func TestPromoteStampsGrounds(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "the loader drops rules silently when the config is stale")

	if _, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID, Grounds: testGrounds}); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if got := groundsScalar(t, ir, issID); got != testGrounds {
		t.Fatalf("grounds = %q, want %q", got, testGrounds)
	}
	// The stamped record still reads: an unknown key would make capture's reader
	// refuse and SKIP it, leaving it invisible to every capture surface.
	if iss := readIssue(t, ir, issID); iss.PromotedTo == "" {
		t.Fatal("the stamped record no longer reads back with its promoted_to")
	}
}

// TestResolveRefusesWithoutGrounds: resolving mints the grounds in the same
// call and has no corpus to fix, so it refuses from the first commit.
func TestResolveRefusesWithoutGrounds(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "a thing that will be fixed")

	if _, err := Resolve(ResolveRequest{
		RepoRoot: repo, IssuesRoot: ir, ID: issID, Resolution: "fixed", Impact: "fix",
	}); err == nil {
		t.Fatal("Resolve without grounds = nil error, want a refusal")
	}
	if iss := readIssue(t, ir, issID); iss.Status != StateOpen {
		t.Fatalf("a refused resolve moved the record to %s", iss.Status)
	}
}

// TestResolveStampsGrounds: the value rides the same atomic transition as the
// note and the impact.
func TestResolveStampsGrounds(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "a thing that will be fixed")

	res, err := Resolve(ResolveRequest{
		RepoRoot: repo, IssuesRoot: ir, ID: issID, Resolution: "fixed", Impact: "fix",
		Grounds: testGrounds,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.ToStatus != StateResolved {
		t.Fatalf("ToStatus = %s, want resolved", res.ToStatus)
	}
	if got := groundsScalar(t, ir, issID); got != testGrounds {
		t.Fatalf("grounds = %q, want %q", got, testGrounds)
	}
}

// TestWontfixStampsDeclinedFromReason: a wontfix can never be recorded without
// grounds — transition already refuses an empty reason — so what it lacked was
// the TYPE, not the text. It stamps `declined: <reason>` and needs no new
// required flag.
func TestWontfixStampsDeclinedFromReason(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "a thing that will not be fixed")

	const reason = "out of scope for this cycle and cheap to reopen later"
	if _, err := Wontfix(WontfixRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID, Reason: reason}); err != nil {
		t.Fatalf("Wontfix: %v", err)
	}
	if got, want := groundsScalar(t, ir, issID), "declined: "+reason; got != want {
		t.Fatalf("grounds = %q, want %q", got, want)
	}
	if iss := readIssue(t, ir, issID); iss.WontfixReason != reason {
		t.Fatalf("wontfix_reason = %q, want it unchanged", iss.WontfixReason)
	}
}

// TestWontfixGroundsOverride: the user-facing reason and the conjecture are not
// always the same sentence, so --grounds overrides the text. The token stays
// declined — a wontfix IS the non-action the value names.
func TestWontfixGroundsOverride(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "a thing that will not be fixed")

	const conjecture = "declined: we expect the successor design to make this unreachable anyway"
	if _, err := Wontfix(WontfixRequest{
		RepoRoot: repo, IssuesRoot: ir, ID: issID, Reason: "out of scope", Grounds: conjecture,
	}); err != nil {
		t.Fatalf("Wontfix: %v", err)
	}
	if got := groundsScalar(t, ir, issID); got != conjecture {
		t.Fatalf("grounds = %q, want %q", got, conjecture)
	}
	if iss := readIssue(t, ir, issID); iss.WontfixReason != "out of scope" {
		t.Fatalf("wontfix_reason = %q, want the reason untouched by the override", iss.WontfixReason)
	}
	// A non-action is declined by construction: the other two values would record
	// a route the folder contradicts.
	if _, err := Wontfix(WontfixRequest{
		RepoRoot: repo, IssuesRoot: ir, ID: issID, Reason: "out of scope", Grounds: testGrounds,
	}); err == nil {
		t.Fatal("a pursued ground on a wontfix = nil error, want a refusal")
	}
}

// TestGroundsTextIsRedacted: the grounds text is free prose bound for the
// committed ledger, so it goes through the same redactor the note already does —
// before the write, never after. The span below is a FAKE shape, matched only by
// shape, never a real path.
func TestGroundsTextIsRedacted(t *testing.T) {
	repo, ir, issID := promoteFixture(t, "a thing that will be fixed")

	const fakeHome = "/Users/alice/bin/fix.sh"
	res, err := Resolve(ResolveRequest{
		RepoRoot: repo, IssuesRoot: ir, ID: issID, Resolution: "fixed", Impact: "fix",
		Grounds: "pursued: we expect the script at " + fakeHome + " to be what proves it",
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if g := groundsScalar(t, ir, issID); strings.Contains(g, "/Users/alice") {
		t.Fatalf("the committed grounds carried the raw home path: %q", g)
	}
	if res.Redacted == 0 {
		t.Fatal("Redacted = 0; a rewritten grounds text must be reported, never silently altered")
	}
	// And the record still reads back: a redaction that broke the scalar would
	// make capture skip the record entirely.
	if _, err := os.Stat(filepath.Join(repo, res.Path)); err != nil {
		t.Fatalf("resolved record unreadable: %v", err)
	}
}

// TestReaderVerdictOnGroundsSpellings is the reader half of the parity the
// record-lint table asserts (iss-2608300927577163). The gate's job is to refuse
// exactly what this refuses — a refused record is SKIPPED, invisible to every
// capture surface while it still sits in the ledger — so the verdicts are pinned
// here rather than described in a comment over there.
func TestReaderVerdictOnGroundsSpellings(t *testing.T) {
	const head = "---\nschema_version: 1\nid: iss-1\nslug: ok\nseverity: minor\n" +
		"category: bug\nsource: user-observation\nfound_during: t\n"
	for _, tc := range []struct {
		name    string
		spell   string
		refused bool
	}{
		{"single quoted", "grounds: 'pursued: the quote survives into the value'\n", true},
		{"empty list", "grounds: []\n", true},
		{"block spelled", "grounds:\n  pursued: an indented block is a mapping\n", true},
		{"empty string", "grounds: \"\"\n", false},
		{"bare null", "grounds: null\n", false},
		{"well formed", "grounds: \"pursued: we expect the reasoning to outlive the session\"\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fm, _, err := parseFrontmatterAndBody(head + tc.spell + "---\n\nan issue\n")
			if err != nil {
				if !tc.refused {
					t.Fatalf("the reader refused at the parse: %v", err)
				}
				return
			}
			err = validateStrict(fm)
			if (err != nil) != tc.refused {
				t.Fatalf("reader refused = %v (%v), want %v", err != nil, err, tc.refused)
			}
		})
	}
}

// TestPromoteControlCharacterGroundsWriteNothing is iss-2608301206032013: the
// pre-mint gate exists so a refusal raised later cannot orphan a draft, and a
// control character walked through it — Go's `\s` excludes the vertical tab, so
// Fold kept U+000B and yamlScalar refused it afterwards, under the ledger lock,
// with the draft already minted. The refusal now lands at the argument boundary,
// where nothing has been written yet.
func TestPromoteControlCharacterGroundsWriteNothing(t *testing.T) {
	for name, bad := range map[string]string{
		"vertical tab": "pursued: we expect the gate\vto refuse a control character",
		"NUL":          "pursued: we expect the gate\x00to refuse a control character",
	} {
		repo, ir, issID := promoteFixture(t, "the loader drops rules silently when the config is stale")
		_, err := Promote(PromoteRequest{RepoRoot: repo, IssuesRoot: ir, ID: issID, Grounds: bad})
		if err == nil {
			t.Fatalf("%s: Promote = nil error, want a refusal", name)
		}
		if !errors.Is(err, ErrGroundsRefused) {
			t.Fatalf("%s: Promote = %v, want the grounds refusal (the pre-mint gate), not a later fault", name, err)
		}
		if n := draftCount(t, repo); n != 0 {
			t.Fatalf("%s: a refused promote minted %d draft(s), want 0", name, n)
		}
		if iss := readIssue(t, ir, issID); iss.PromotedTo != "" {
			t.Fatalf("%s: a refused promote stamped promoted_to = %q", name, iss.PromotedTo)
		}
	}
}
