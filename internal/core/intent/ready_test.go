package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	disciplinesDir = ".abcd/development/intents/disciplines"
	supersededDir  = ".abcd/development/intents/superseded"
)

// draftSeeded is a draft in the exact shape CreateFromText seeds: the
// Acceptance Criteria section holds only the placeholder blockquote, no bullets.
func draftSeeded(id, slug string) string {
	return "---\nid: " + id + "\nslug: " + slug + "\nspec_id: null\nkind: null\n---\n" +
		"# " + slug + "\n\n## Acceptance Criteria\n\n" +
		"> _Required: add at least one Given-When-Then bullet before planning._\n"
}

// plannedUnlinked is a planned intent whose spec_id is still null (a shape Plan
// never leaves, but the gate must report it rather than trust it).
func plannedUnlinked(id, slug string) string {
	return "---\nid: " + id + "\nslug: " + slug + "\nspec_id: null\nkind: standalone\n---\n" +
		"# " + slug + "\n\n## Scope Conditions\n\n" + NullityToken + "\n\n## Acceptance Criteria\n\n- ok\n"
}

// specStub is an open spec still carrying the minted _Draft: placeholder.
func specStub(id, slug, intentID string) string {
	return "---\nid: " + id + "\nslug: " + slug + "\nintent: " + intentID + "\n---\n" +
		"# " + slug + "\n\n## Summary\n\n" +
		"_Draft: describe what " + id + " delivers for " + intentID + " — scope, approach._\n"
}

// checkByName finds one check row; the shape contract makes absence a test bug.
func checkByName(t *testing.T, res ReadyResult, name string) ReadyCheck {
	t.Helper()
	for _, c := range res.Checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("check %q missing from %+v", name, res.Checks)
	return ReadyCheck{}
}

// assertShape enforces the machine-shape contract: always exactly six rows in
// fixed order, whatever the intent's state.
func assertShape(t *testing.T, res ReadyResult) {
	t.Helper()
	want := []string{"bucket", "acceptance_criteria", "mechanism_claim", "scope_conditions", "spec_link", "spec_body"}
	if len(res.Checks) != len(want) {
		t.Fatalf("expected %d checks, got %d: %+v", len(want), len(res.Checks), res.Checks)
	}
	for i, name := range want {
		if res.Checks[i].Name != name {
			t.Fatalf("check[%d] = %q, want %q", i, res.Checks[i].Name, name)
		}
	}
}

func TestReadyDraftSeededPlaceholder(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftSeeded("itd-10", "alpha"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, res)
	if res.Ready {
		t.Fatal("a seeded draft must not be ready")
	}
	bucket := checkByName(t, res, "bucket")
	if bucket.OK {
		t.Fatal("bucket check must fail for drafts")
	}
	if !strings.Contains(bucket.Remedy, "planning interview") {
		t.Fatalf("bucket remedy for a draft without AC must offer the interview, got %q", bucket.Remedy)
	}
	if ac := checkByName(t, res, "acceptance_criteria"); ac.OK {
		t.Fatal("acceptance_criteria must fail on the seeded placeholder (no bullets)")
	}
}

// TestReadyDraftWithRealAC is the motivating incident (itd-93's shape): seeded
// AC bullets look real, but the intent is unplanned — NOT READY, with the
// confirm-then-plan remedy.
func TestReadyDraftWithRealAC(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, res)
	if res.Ready {
		t.Fatal("a draft must never be ready, however complete its AC look")
	}
	bucket := checkByName(t, res, "bucket")
	if bucket.OK || !strings.Contains(bucket.Remedy, "abcd intent plan itd-10") {
		t.Fatalf("bucket check = %+v, want fail with plan remedy", bucket)
	}
	if ac := checkByName(t, res, "acceptance_criteria"); !ac.OK {
		t.Fatalf("acceptance_criteria = %+v, want OK (bullets present)", ac)
	}
	if sb := checkByName(t, res, "spec_body"); sb.OK {
		t.Fatal("spec_body must not pass with no linked spec")
	}
}

func TestReadyPlannedNullSpecNoClaimer(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedUnlinked("itd-10", "alpha"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, res)
	if res.Ready {
		t.Fatal("planned with no spec must not be ready")
	}
	link := checkByName(t, res, "spec_link")
	if link.OK || !strings.Contains(link.Detail, "no spec realises") {
		t.Fatalf("spec_link = %+v, want fail naming the missing spec", link)
	}
}

func TestReadyPlannedNullSpecOneSidedClaimer(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedUnlinked("itd-10", "alpha"))
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-10"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	link := checkByName(t, res, "spec_link")
	if link.OK || !strings.Contains(link.Remedy, "abcd intent link itd-10 spc-1") {
		t.Fatalf("spec_link = %+v, want the link remedy for a one-sided claimer", link)
	}
}

func TestReadyPlannedStubSpecBody(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedLinked("itd-10", "alpha", "spc-1"))
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specStub("spc-1", "alpha", "itd-10"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, res)
	if res.Ready {
		t.Fatal("a stub spec body must block readiness")
	}
	if link := checkByName(t, res, "spec_link"); !link.OK {
		t.Fatalf("spec_link = %+v, want OK", link)
	}
	body := checkByName(t, res, "spec_body")
	if body.OK || !strings.Contains(body.Remedy, "write the spec body") {
		t.Fatalf("spec_body = %+v, want fail with write-the-body remedy", body)
	}
}

func TestReadyGreen(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedLinked("itd-10", "alpha", "spc-1"))
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-10"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, res)
	if !res.Ready {
		t.Fatalf("planned+linked+written must be ready: %+v", res.Checks)
	}
	for _, c := range res.Checks {
		if !c.OK {
			t.Fatalf("check %s failed on the green path: %+v", c.Name, c)
		}
	}
	if res.Bucket != BucketPlanned || res.SpecID != "spc-1" {
		t.Fatalf("result header = %+v", res)
	}
}

func TestReadyBidirectionalDrift(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedLinked("itd-10", "alpha", "spc-1"))
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-99"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatalf("drift is a report, not a fault: %v", err)
	}
	link := checkByName(t, res, "spec_link")
	if link.OK || !strings.Contains(link.Detail, "itd-99") {
		t.Fatalf("spec_link = %+v, want fail naming the disagreeing claim", link)
	}
}

func TestReadyTerminalBuckets(t *testing.T) {
	tests := []struct {
		dir, wantDetail string
	}{
		{shippedDir, "shipped"},
		{disciplinesDir, "discipline"},
		{supersededDir, "superseded"},
	}
	for _, tt := range tests {
		t.Run(tt.wantDetail, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, tt.dir+"/itd-10-alpha.md", plannedLinked("itd-10", "alpha", "spc-1"))

			res, err := Ready(root, "itd-10")
			if err != nil {
				t.Fatal(err)
			}
			assertShape(t, res)
			if res.Ready {
				t.Fatalf("%s must not be ready", tt.wantDetail)
			}
			bucket := checkByName(t, res, "bucket")
			if bucket.OK || !strings.Contains(bucket.Detail, tt.wantDetail) {
				t.Fatalf("bucket = %+v, want fail mentioning %q", bucket, tt.wantDetail)
			}
			if bucket.Remedy != "" {
				t.Fatalf("a terminal bucket has no remedy, got %q", bucket.Remedy)
			}
		})
	}
}

func TestReadyFaults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))

	if _, err := Ready(root, "itd-../etc"); err == nil {
		t.Fatal("malformed id must be a fault")
	}
	if _, err := Ready(root, "itd-999"); err == nil {
		t.Fatal("unknown intent must be a fault")
	}

	// A symlinked intent record violates the store trust boundary.
	linkRoot := t.TempDir()
	target := filepath.Join(linkRoot, "outside.md")
	if err := os.WriteFile(target, []byte(draftWithAC("itd-7", "gamma")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(linkRoot, draftsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(linkRoot, draftsDir, "itd-7-gamma.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Ready(linkRoot, "itd-7"); err == nil {
		t.Fatal("a symlinked intent record must be a fault")
	}
}

// TestReadySpecLinkToleratesSpecIDSpelling mirrors Reconcile's tolerance in the
// report: the spec_link check reads the same stored spec_id, so a lint-green
// slug-suffixed or zero-padded value must report the link as held, not missing.
func TestReadySpecLinkToleratesSpecIDSpelling(t *testing.T) {
	for _, specID := range []string{"spc-1-alpha", "spc-01"} {
		t.Run(specID, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedLinked("itd-10", "alpha", specID))
			writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-10"))

			res, err := Ready(root, "itd-10")
			if err != nil {
				t.Fatal(err)
			}
			if link := checkByName(t, res, "spec_link"); !link.OK {
				t.Fatalf("spec_link = %+v, want OK for the lint-green spec_id %q", link, specID)
			}
		})
	}
}

// TestReadyChecksOrderAndCount pins the machine-shape contract itself: the two
// claim checks report between the criterion claim and the spec link, and every
// row is present whatever the record says.
func TestReadyChecksOrderAndCount(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedLinked("itd-10", "alpha", "spc-1"))
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-10"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	assertShape(t, res)
}

// plannedWithClaims is a planned, linked intent whose claim sections carry the
// given bodies verbatim — the fixture the five gradient cases vary.
func plannedWithClaims(mechanism, conditions string) string {
	return "---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n# alpha\n\n" +
		mechanism + conditions + "## Acceptance Criteria\n\n- ok\n"
}

// readyWithClaims runs the gate over a planned record carrying the given claim
// sections, against a written spec, so only the claim checks can fail.
func readyWithClaims(t *testing.T, mechanism, conditions string) ReadyResult {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedWithClaims(mechanism, conditions))
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-10"))
	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// TestReadyScopeConditionsAbsent is itd-177's first criterion: no conditions and
// no explicit nullity exits the gate non-zero, naming the missing field.
func TestReadyScopeConditionsAbsent(t *testing.T) {
	res := readyWithClaims(t, "", "")
	if res.Ready {
		t.Fatal("an intent with no context claim must not be ready")
	}
	c := checkByName(t, res, "scope_conditions")
	if c.OK || !strings.Contains(c.Detail, "## Scope Conditions") {
		t.Fatalf("scope_conditions = %+v, want fail naming the section", c)
	}
	if !strings.Contains(c.Remedy, NullityToken) {
		t.Fatalf("remedy must name the nullity token, got %q", c.Remedy)
	}
}

// TestReadyScopeConditionsEmptyFails: a heading with nothing under it is the
// gate fault the gradient distinguishes from an absent section.
func TestReadyScopeConditionsEmptyFails(t *testing.T) {
	res := readyWithClaims(t, "", "## Scope Conditions\n\n")
	if c := checkByName(t, res, "scope_conditions"); c.OK || !strings.Contains(c.Detail, "empty") {
		t.Fatalf("scope_conditions = %+v, want fail on the empty section", c)
	}
}

// TestReadyScopeConditionsNullityPasses: the recorded decline is a pass.
func TestReadyScopeConditionsNullityPasses(t *testing.T) {
	res := readyWithClaims(t, "", "## Scope Conditions\n\n"+NullityToken+"\n\n")
	c := checkByName(t, res, "scope_conditions")
	if !c.OK || !strings.Contains(c.Detail, "nullity recorded") {
		t.Fatalf("scope_conditions = %+v, want OK reporting the recorded nullity", c)
	}
	if !res.Ready {
		t.Fatalf("a recorded nullity must leave the intent ready: %+v", res.Checks)
	}
}

// TestReadyMechanismNullityPasses is itd-177's third criterion.
func TestReadyMechanismNullityPasses(t *testing.T) {
	res := readyWithClaims(t, "## Mechanism\n\n"+NullityToken+"\n\n", "## Scope Conditions\n\n"+NullityToken+"\n\n")
	c := checkByName(t, res, "mechanism_claim")
	if !c.OK || !strings.Contains(c.Detail, "declined (nullity recorded)") {
		t.Fatalf("mechanism_claim = %+v, want OK reporting the recorded nullity", c)
	}
	if !res.Ready {
		t.Fatalf("a declined mechanism must leave the intent ready: %+v", res.Checks)
	}
}

// TestReadyMechanismEmptyFails is itd-177's fourth criterion: present but empty
// exits non-zero and names the section — write the claim or the token.
func TestReadyMechanismEmptyFails(t *testing.T) {
	res := readyWithClaims(t, "## Mechanism\n\n", "## Scope Conditions\n\n"+NullityToken+"\n\n")
	if res.Ready {
		t.Fatal("an empty mechanism section must not be ready")
	}
	c := checkByName(t, res, "mechanism_claim")
	if c.OK || !strings.Contains(c.Detail, "## Mechanism") {
		t.Fatalf("mechanism_claim = %+v, want fail naming the section", c)
	}
	if !strings.Contains(c.Remedy, NullityToken) {
		t.Fatalf("remedy must offer the claim or the token, got %q", c.Remedy)
	}
}

// TestReadyMechanismAbsentPasses: prompted is not required — an absent section
// is a claim not carried, which the gradient never conflates with a fault.
func TestReadyMechanismAbsentPasses(t *testing.T) {
	res := readyWithClaims(t, "", "## Scope Conditions\n\n"+NullityToken+"\n\n")
	if c := checkByName(t, res, "mechanism_claim"); !c.OK {
		t.Fatalf("mechanism_claim = %+v, want OK for an absent prompted claim", c)
	}
}

// TestReadyConditionMarkerMissing is itd-177's fifth criterion, missing half.
func TestReadyConditionMarkerMissing(t *testing.T) {
	res := readyWithClaims(t, "", "## Scope Conditions\n\n"+
		"- stamped <!-- cond: cond-2608300102030405 -->\n- unstamped\n\n")
	if res.Ready {
		t.Fatal("an unidentified condition must not be ready")
	}
	c := checkByName(t, res, "scope_conditions")
	if c.OK || !strings.Contains(c.Detail, "2") {
		t.Fatalf("scope_conditions = %+v, want fail naming condition 2", c)
	}
	if !strings.Contains(c.Remedy, "abcd intent plan itd-10") {
		t.Fatalf("remedy must name the write-capable verb, got %q", c.Remedy)
	}
}

// TestReadyConditionMarkerDuplicated is the fifth criterion's duplicate half:
// two conditions sharing an identity is named by the repeated cond- id.
func TestReadyConditionMarkerDuplicated(t *testing.T) {
	const id = "cond-2608300102030405"
	res := readyWithClaims(t, "", "## Scope Conditions\n\n"+
		"- one <!-- cond: "+id+" -->\n- two <!-- cond: "+id+" -->\n\n")
	if res.Ready {
		t.Fatal("a duplicated identity must not be ready")
	}
	c := checkByName(t, res, "scope_conditions")
	if c.OK || !strings.Contains(c.Detail, id) {
		t.Fatalf("scope_conditions = %+v, want fail naming %s", c, id)
	}
}

// TestReadyScopeConditionsProseWithoutBulletsFails: a condition that is not a
// top-level bullet has nothing to carry an identity, so prose alone — the
// create-path scaffold's own prompt line included — is not a recorded claim.
func TestReadyScopeConditionsProseWithoutBulletsFails(t *testing.T) {
	res := readyWithClaims(t, "", "## Scope Conditions\n\nIt holds wherever a POSIX shell exists.\n\n")
	if c := checkByName(t, res, "scope_conditions"); c.OK || !strings.Contains(c.Detail, "bullet") {
		t.Fatalf("scope_conditions = %+v, want fail on prose without bullets", c)
	}
}

// TestReadyDisciplineExemptFromClaimChecks: a discipline record's template
// carries no claim sections, so the gradient exempts it and both checks say so.
func TestReadyDisciplineExemptFromClaimChecks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, disciplinesDir+"/itd-10-alpha.md", plannedLinked("itd-10", "alpha", "spc-1"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"mechanism_claim", "scope_conditions"} {
		c := checkByName(t, res, name)
		if !c.OK || !strings.Contains(c.Detail, "discipline records carry no claim sections") {
			t.Fatalf("%s = %+v, want the exemption", name, c)
		}
	}
}

// TestReadyReportsConditionIdentities: ReadyResult carries the parsed conditions,
// which is the payload the identity criteria assert against.
func TestReadyReportsConditionIdentities(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n"+
			"# alpha\n\n## Scope Conditions\n\n"+
			"- holds on a POSIX shell <!-- cond: cond-2608300102030405 -->\n\n"+
			"## Acceptance Criteria\n\n- ok\n")
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-10"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Conditions) != 1 {
		t.Fatalf("Conditions = %+v, want one", res.Conditions)
	}
	if res.Conditions[0].ID != "cond-2608300102030405" {
		t.Fatalf("condition id = %q", res.Conditions[0].ID)
	}
	if cond := checkByName(t, res, "scope_conditions"); !cond.OK {
		t.Fatalf("scope_conditions = %+v, want OK for an identified condition", cond)
	}
}

// TestReadyClaimChecksNotApplicableInTerminalBuckets is half of
// iss-2608300210588414: spc-55 rules retro-fitting claims onto shipped/ and
// superseded/ records out of scope — an absent stamp is information — so the
// gate must not print a backfill remedy at a record nobody may backfill.
func TestReadyClaimChecksNotApplicableInTerminalBuckets(t *testing.T) {
	for _, dir := range []string{shippedDir, supersededDir} {
		t.Run(dir, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, dir+"/itd-10-alpha.md",
				"---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n"+
					"# alpha\n\n## Mechanism\n\n## Acceptance Criteria\n\n- ok\n")

			res, err := Ready(root, "itd-10")
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"mechanism_claim", "scope_conditions"} {
				c := checkByName(t, res, name)
				if !c.OK || !strings.Contains(c.Detail, "not applicable") {
					t.Errorf("%s = %+v, want OK reporting the check as not applicable", name, c)
				}
				if c.Remedy != "" {
					t.Errorf("%s must carry no remedy at a record nobody may backfill, got %q", name, c.Remedy)
				}
			}
		})
	}
}

// TestReadyScaffoldPromptIsNotAClaim is the other half: the create-path prompt
// is bytes this package wrote asking for a claim, never a claim somebody made,
// and the gate must not report it as one.
func TestReadyScaffoldPromptIsNotAClaim(t *testing.T) {
	root := t.TempDir()
	seeded := seedDraft("itd-10", DraftOptions{Slug: "alpha", Title: "alpha", SeedBody: "why it matters"})
	// Plan the criteria only, so the two claim sections stay as seeded.
	seeded = strings.Replace(seeded,
		"> _Required (the itd-1 discipline)", "- **Given** x, **when** y, **then** z.\n\n> _was: ", 1)
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		strings.Replace(seeded, "spec_id: null", "spec_id: spc-1", 1))
	writeFile(t, root, specsOpen+"/spc-1-alpha.md", specNaming("spc-1", "alpha", "itd-10"))

	res, err := Ready(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	mech := checkByName(t, res, "mechanism_claim")
	if !mech.OK {
		t.Fatalf("an unanswered prompt is not a fault (mechanism is nullable): %+v", mech)
	}
	if strings.Contains(mech.Detail, "mechanism claim stated") {
		t.Fatalf("the scaffold prompt must not read as a stated claim: %+v", mech)
	}
	if !strings.Contains(mech.Detail, "unanswered") {
		t.Fatalf("mechanism_claim = %+v, want the detail to name the unanswered prompt", mech)
	}
	cond := checkByName(t, res, "scope_conditions")
	if cond.OK || !strings.Contains(cond.Detail, "unanswered") {
		t.Fatalf("scope_conditions = %+v, want fail naming the unanswered prompt", cond)
	}
}

// TestReadyConditionCarryingTwoMarkers: a bullet holding two identities is
// named by the gate, because nothing downstream can choose between them.
func TestReadyConditionCarryingTwoMarkers(t *testing.T) {
	res := readyWithClaims(t, "", "## Scope Conditions\n\n"+
		"- one condition <!-- cond: cond-2608300102030405 --> <!-- cond: cond-2608300102030406 -->\n\n")
	if res.Ready {
		t.Fatal("a bullet with two identities must not be ready")
	}
	c := checkByName(t, res, "scope_conditions")
	if c.OK || !strings.Contains(c.Detail, "1") || !strings.Contains(c.Detail, "more than one identity") {
		t.Fatalf("scope_conditions = %+v, want fail naming the ambiguous bullet", c)
	}
	if c.Remedy == "" {
		t.Fatal("the fault must carry a remedy")
	}
}
