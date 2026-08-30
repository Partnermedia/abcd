package intent

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/intentdriven/abcd/internal/core/recordid"
)

// claimRecord renders an intent body carrying the given `## Mechanism` and
// `## Scope Conditions` section bodies verbatim; an empty section argument means
// the heading is written with nothing under it, and a nil one means the heading
// is absent entirely.
func claimRecord(mechanism, conditions *string) string {
	var b strings.Builder
	b.WriteString("---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n\n# alpha\n\n")
	if mechanism != nil {
		b.WriteString("## Mechanism\n\n" + *mechanism + "\n")
	}
	if conditions != nil {
		b.WriteString("## Scope Conditions\n\n" + *conditions + "\n")
	}
	b.WriteString("\n## Acceptance Criteria\n\n- ok\n")
	return b.String()
}

func str(s string) *string { return &s }

func TestParseClaimsThreeByteStates(t *testing.T) {
	tests := []struct {
		name                 string
		mechanism            *string
		conditions           *string
		wantMech, wantCondSt ClaimState
	}{
		{"both absent", nil, nil, ClaimAbsent, ClaimAbsent},
		{"both empty", str(""), str(""), ClaimEmpty, ClaimEmpty},
		{"both nullity", str("None stated."), str("None stated."), ClaimNullity, ClaimNullity},
		{"both stated", str("We expect X because Y."), str("- holds on POSIX"), ClaimStated, ClaimStated},
		{"mechanism absent, conditions stated", nil, str("- holds on POSIX"), ClaimAbsent, ClaimStated},
		{"mechanism nullity, conditions empty", str("None stated."), str(""), ClaimNullity, ClaimEmpty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ParseClaims(claimRecord(tt.mechanism, tt.conditions))
			if c.Mechanism != tt.wantMech {
				t.Errorf("Mechanism = %q, want %q", c.Mechanism, tt.wantMech)
			}
			if c.ConditionsState != tt.wantCondSt {
				t.Errorf("ConditionsState = %q, want %q", c.ConditionsState, tt.wantCondSt)
			}
		})
	}
}

// TestNullityTokenIsExact pins the one accepted spelling: prose that merely
// contains the word "none" is a stated claim, never a recorded nullity.
func TestNullityTokenIsExact(t *testing.T) {
	stated := []string{
		"None stated",                     // no full stop
		"None stated. It never applies.",  // trailing prose on the line
		"none stated.",                    // lower case
		"_None stated._",                  // emphasised
		"None stated.\n\nBut see itd-1.",  // a second non-blank line
		"There are none stated for this.", // the word inside prose
		"- None stated.",                  // a bullet, not the bare token
		"None  stated.",                   // internal whitespace differs
		"  None stated.",                  // indented, not alone at column 0
		"None stated. ",                   // trailing space after the token
	}
	for _, body := range stated {
		t.Run(body, func(t *testing.T) {
			if got := ParseClaims(claimRecord(str(body), nil)).Mechanism; got != ClaimStated {
				t.Fatalf("Mechanism for %q = %q, want %q", body, got, ClaimStated)
			}
		})
	}
	if got := ParseClaims(claimRecord(str("None stated."), nil)).Mechanism; got != ClaimNullity {
		t.Fatalf("the exact token = %q, want %q", got, ClaimNullity)
	}
}

func TestParseConditionsMarkerExtraction(t *testing.T) {
	body := "- unmarked condition\n" +
		"- marked condition <!-- cond: cond-2608300102030405 -->\n" +
		"  - an indented sub-bullet is detail, not a condition\n" +
		"- a wrapped condition whose sentence <!-- cond: cond-2608300102030406 -->\n" +
		"  continues on the next line\n"
	c := ParseClaims(claimRecord(nil, str(body)))
	if c.ConditionsState != ClaimStated {
		t.Fatalf("ConditionsState = %q, want %q", c.ConditionsState, ClaimStated)
	}
	if len(c.Conditions) != 3 {
		t.Fatalf("got %d conditions, want 3: %+v", len(c.Conditions), c.Conditions)
	}
	if c.Conditions[0].ID != "" {
		t.Errorf("condition 1 ID = %q, want empty (unmarked)", c.Conditions[0].ID)
	}
	if c.Conditions[0].Ordinal != 1 || c.Conditions[2].Ordinal != 3 {
		t.Errorf("ordinals = %d,%d, want 1,3", c.Conditions[0].Ordinal, c.Conditions[2].Ordinal)
	}
	if c.Conditions[1].ID != "cond-2608300102030405" {
		t.Errorf("condition 2 ID = %q", c.Conditions[1].ID)
	}
	if c.Conditions[1].Text != "marked condition" {
		t.Errorf("condition 2 Text = %q, want the prose without the marker", c.Conditions[1].Text)
	}
	if want := "a wrapped condition whose sentence continues on the next line"; c.Conditions[2].Text != want {
		t.Errorf("condition 3 Text = %q, want %q", c.Conditions[2].Text, want)
	}
}

// TestConditionIdentitySurvivesEdit is itd-177's identity criterion: the marker
// is bytes inside the bullet, so rewriting the prose around it changes nothing.
func TestConditionIdentitySurvivesEdit(t *testing.T) {
	const id = "cond-2608300102030405"
	before := ParseClaims(claimRecord(nil, str("- holds on POSIX <!-- cond: "+id+" -->")))
	after := ParseClaims(claimRecord(nil, str("- holds only where a POSIX shell exists <!-- cond: "+id+" -->")))
	if len(before.Conditions) != 1 || len(after.Conditions) != 1 {
		t.Fatalf("expected one condition either side: %+v / %+v", before.Conditions, after.Conditions)
	}
	if before.Conditions[0].ID != after.Conditions[0].ID {
		t.Fatalf("identity changed across an edit: %q -> %q", before.Conditions[0].ID, after.Conditions[0].ID)
	}
	if before.Conditions[0].Text == after.Conditions[0].Text {
		t.Fatal("the test edited nothing — Text must differ either side")
	}
}

// fixedMinter is a Minter whose clock and entropy are pinned, so a stamped
// marker is byte-predictable. suffixes are consumed two bytes at a time.
func fixedMinter(suffixes ...uint16) recordid.Minter {
	var b []byte
	for _, s := range suffixes {
		b = append(b, byte(s>>8), byte(s))
	}
	return recordid.Minter{
		Now:     func() time.Time { return time.Date(2026, 8, 30, 1, 2, 3, 0, time.UTC) },
		Entropy: bytes.NewReader(b),
	}
}

func TestStampScopeConditionsMarksOnlyUnmarkedBullets(t *testing.T) {
	const kept = "cond-2608300102030405"
	body := "- first, already stamped <!-- cond: " + kept + " -->\n" +
		"- second, split out of the first and unmarked\n" +
		"  - a sub-bullet is not a condition\n"
	content := claimRecord(nil, str(body))

	stamped, n, err := stampScopeConditions(content, fixedMinter(7))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("stamped %d bullets, want 1", n)
	}
	conds := ParseClaims(stamped).Conditions
	if len(conds) != 2 {
		t.Fatalf("got %d conditions, want 2: %+v", len(conds), conds)
	}
	if conds[0].ID != kept {
		t.Errorf("the existing marker was rewritten: %q", conds[0].ID)
	}
	if conds[1].ID != "cond-2608300102030007" {
		t.Errorf("condition 2 ID = %q, want the minted marker", conds[1].ID)
	}
	if !strings.Contains(stamped, "- a sub-bullet is not a condition\n") {
		t.Error("an indented sub-bullet must be left untouched")
	}
}

func TestStampScopeConditionsIsIdempotent(t *testing.T) {
	content := claimRecord(nil, str("- holds on POSIX\n- holds below 10k records\n"))

	once, n, err := stampScopeConditions(content, fixedMinter(7, 8))
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("first pass stamped %d, want 2", n)
	}
	twice, n, err := stampScopeConditions(once, fixedMinter(9, 10))
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("second pass stamped %d, want 0", n)
	}
	if twice != once {
		t.Fatalf("a second pass rewrote the record:\n%q\n%q", once, twice)
	}
}

// TestStampScopeConditionsRedrawsOnCollision pins that one stamping pass never
// hands two bullets the same identity: the mint reads no ledger, so a
// same-second suffix coincidence is the caller's to resolve.
func TestStampScopeConditionsRedrawsOnCollision(t *testing.T) {
	content := claimRecord(nil, str("- one\n- two\n"))

	stamped, n, err := stampScopeConditions(content, fixedMinter(7, 7, 8))
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("stamped %d, want 2", n)
	}
	conds := ParseClaims(stamped).Conditions
	if conds[0].ID == conds[1].ID {
		t.Fatalf("both conditions carry %q", conds[0].ID)
	}
}

func TestStampScopeConditionsWithoutSection(t *testing.T) {
	content := claimRecord(nil, nil)
	stamped, n, err := stampScopeConditions(content, fixedMinter())
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 || stamped != content {
		t.Fatalf("an absent section must be left verbatim (n=%d)", n)
	}
}

// TestPlanStampsConditionIdentities is itd-177's mint criterion end to end: the
// identities are written by `intent plan`, the lifecycle's write-capable verb,
// never hand-typed and never minted by the read-only gate.
func TestPlanStampsConditionIdentities(t *testing.T) {
	root := t.TempDir()
	draft := "---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: null\n---\n" +
		"# alpha\n\n## Scope Conditions\n\n" +
		"- holds on a POSIX shell\n- holds below 10k records\n\n" +
		"## Acceptance Criteria\n\n- ok\n"
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draft)

	res, err := Plan(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	if res.ConditionsStamped != 2 {
		t.Fatalf("ConditionsStamped = %d, want 2", res.ConditionsStamped)
	}
	body, err := os.ReadFile(filepath.Join(root, res.Intent.Path))
	if err != nil {
		t.Fatal(err)
	}
	conds := ParseClaims(string(body)).Conditions
	if len(conds) != 2 {
		t.Fatalf("got %d conditions, want 2: %+v", len(conds), conds)
	}
	for _, c := range conds {
		if !regexp.MustCompile(`^cond-[0-9]{16}$`).MatchString(c.ID) {
			t.Errorf("condition %d id = %q, want the cond-<16 digits> grammar", c.Ordinal, c.ID)
		}
	}
	if conds[0].ID == conds[1].ID {
		t.Fatalf("both conditions were stamped %q", conds[0].ID)
	}
	if conds[0].Text != "holds on a POSIX shell" {
		t.Errorf("the stamp rewrote the prose: %q", conds[0].Text)
	}
}

// TestPlanLeavesTheNullityTokenAlone: a recorded nullity has no bullet to stamp,
// and the stamp must not invent one.
func TestPlanLeavesTheNullityTokenAlone(t *testing.T) {
	root := t.TempDir()
	draft := "---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: null\n---\n" +
		"# alpha\n\n## Scope Conditions\n\nNone stated.\n\n## Acceptance Criteria\n\n- ok\n"
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draft)

	res, err := Plan(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	if res.ConditionsStamped != 0 {
		t.Fatalf("ConditionsStamped = %d, want 0", res.ConditionsStamped)
	}
	body, err := os.ReadFile(filepath.Join(root, res.Intent.Path))
	if err != nil {
		t.Fatal(err)
	}
	if got := ParseClaims(string(body)).ConditionsState; got != ClaimNullity {
		t.Fatalf("ConditionsState = %q, want %q", got, ClaimNullity)
	}
}

// TestSeedDraftCarriesClaimSections: "prompted" has a named surface — the
// create-path scaffold, so a drafted intent arrives carrying the prompt.
func TestSeedDraftCarriesClaimSections(t *testing.T) {
	root := t.TempDir()
	it, _, err := CreateFromText(root, "the card respects the reader's time", "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, it.Path))
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, want := range []string{"## Mechanism", "## Scope Conditions", NullityToken} {
		if !strings.Contains(body, want) {
			t.Errorf("seeded draft is missing %q:\n%s", want, body)
		}
	}
	// Both scaffolded sections must read as a prompt, never as a recorded claim:
	// a seeded nullity would be a decline nobody made.
	c := ParseClaims(body)
	if c.Mechanism != ClaimStated || c.ConditionsState != ClaimStated {
		t.Fatalf("scaffold states = (%q, %q), want both %q (the one-line contract)", c.Mechanism, c.ConditionsState, ClaimStated)
	}
	// The scaffold sits above the criteria, matching the record template.
	if strings.Index(body, "## Scope Conditions") > strings.Index(body, "## Acceptance Criteria") {
		t.Error("the claim sections belong above '## Acceptance Criteria'")
	}
}
