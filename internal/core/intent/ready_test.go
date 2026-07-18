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
		"# " + slug + "\n\n## Acceptance Criteria\n\n- ok\n"
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

// assertShape enforces the machine-shape contract: always exactly four rows in
// fixed order, whatever the intent's state.
func assertShape(t *testing.T, res ReadyResult) {
	t.Helper()
	want := []string{"bucket", "acceptance_criteria", "spec_link", "spec_body"}
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
