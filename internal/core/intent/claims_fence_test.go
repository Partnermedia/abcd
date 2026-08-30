package intent

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// fencedShadow is the shadowing case of iss-2608300235388164: a fenced EXAMPLE
// of the section (a template, a how-to) appears above the record's real one.
// First-match-wins made the example the section, so `plan` wrote a marker inside
// the code fence and `ready` reported READY on the example while the record's
// real conditions went unread.
const fencedShadow = "---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n\n# alpha\n\n" +
	"## How to write one\n\n" +
	"```markdown\n" +
	"## Scope Conditions\n\n" +
	"- an example condition\n" +
	"```\n\n" +
	"## Scope Conditions\n\n" +
	"- holds only where a POSIX shell exists\n\n" +
	"## Acceptance Criteria\n\n- ok\n"

func TestSectionBoundIgnoresAFencedHeading(t *testing.T) {
	c := ParseClaims(fencedShadow)
	if c.ConditionsState != ClaimStated {
		t.Fatalf("ConditionsState = %q, want %q", c.ConditionsState, ClaimStated)
	}
	if len(c.Conditions) != 1 {
		t.Fatalf("got %d conditions, want the record's own one: %+v", len(c.Conditions), c.Conditions)
	}
	if c.Conditions[0].Text != "holds only where a POSIX shell exists" {
		t.Fatalf("read the fenced example instead of the record: %q", c.Conditions[0].Text)
	}
}

// A fenced heading must not terminate the section it sits inside either.
func TestSectionBoundIgnoresAFencedTerminator(t *testing.T) {
	content := "---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n\n# alpha\n\n" +
		"## Scope Conditions\n\n- holds on POSIX\n\n" +
		"~~~\n## Acceptance Criteria\n~~~\n\n" +
		"- holds below 10k records\n\n" +
		"## Acceptance Criteria\n\n- ok\n"
	conds := ParseClaims(content).Conditions
	if len(conds) != 2 {
		t.Fatalf("got %d conditions, want 2 (the fenced heading is not a terminator): %+v", len(conds), conds)
	}
}

// A fenced bullet UNDER the real heading is an example, not a condition.
func TestFencedBulletIsNotACondition(t *testing.T) {
	body := "- holds on POSIX\n\n```\n- not a condition\n```\n"
	conds := ParseClaims(claimRecord(nil, str(body))).Conditions
	if len(conds) != 1 {
		t.Fatalf("got %d conditions, want 1: %+v", len(conds), conds)
	}
}

// TestStampRefusesAFencedSection: writing a marker into a section that contains
// a fence risks landing it inside the example. The stamp refuses, writes
// nothing, and says why.
func TestStampRefusesAFencedSection(t *testing.T) {
	content := claimRecord(nil, str("- holds on POSIX\n\n```\n- example\n```\n"))
	got, n, err := stampScopeConditions(content, fixedMinter(7))
	if err == nil {
		t.Fatal("a fenced section must refuse the stamp")
	}
	if !strings.Contains(err.Error(), "fenced") {
		t.Fatalf("refusal must name the fence, got %q", err)
	}
	if n != 0 || got != "" {
		t.Fatalf("a refusal writes nothing (n=%d)", n)
	}
}

// TestStampRefusesADuplicateHeading: two `## Scope Conditions` headings make
// "the section" ambiguous, so the stamp refuses rather than picking one.
func TestStampRefusesADuplicateHeading(t *testing.T) {
	content := "---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n\n# alpha\n\n" +
		"## Scope Conditions\n\n- holds on POSIX\n\n" +
		"## Scope Conditions\n\n- and again\n\n" +
		"## Acceptance Criteria\n\n- ok\n"
	if _, n, err := stampScopeConditions(content, fixedMinter(7)); err == nil || n != 0 {
		t.Fatalf("a duplicated heading must refuse the stamp (n=%d, err=%v)", n, err)
	}
}

// TestMalformedMarkerIsNeverGluedBesideARealOne: a hand-typed marker of the
// wrong shape read as ordinary prose, so the stamp added a real marker next to
// it. The bullet is now refused and reported instead.
func TestMalformedMarkerIsNeverGluedBesideARealOne(t *testing.T) {
	body := "- holds on POSIX <!-- cond: cond-123 -->\n- properly unmarked\n"
	c := ParseClaims(claimRecord(nil, str(body)))
	if !c.Conditions[0].MalformedMarker {
		t.Fatalf("condition 1 must be flagged as carrying a malformed marker: %+v", c.Conditions[0])
	}
	if c.Conditions[1].MalformedMarker {
		t.Fatalf("condition 2 carries no comment at all: %+v", c.Conditions[1])
	}
	stamped, n, err := stampScopeConditions(claimRecord(nil, str(body)), fixedMinter(7))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("stamped %d, want only the clean bullet", n)
	}
	if strings.Contains(stamped, "cond-123 --> <!-- cond:") {
		t.Fatalf("a real marker was glued beside the malformed one:\n%s", stamped)
	}
}

// TestFenceAwareBoundLeavesEveryAuditReceiptUnchanged is the corpus proof asked
// for by iss-2608300235388164: sectionLineRange is the single bound, so making
// it fence-aware moves every section body in the tree — including the
// `## Acceptance Criteria` body whose sha256 IS a parked review receipt. A
// record with a fenced heading inside its AC section would silently invalidate
// its own receipt. This asserts no such record exists, against the real tree.
func TestFenceAwareBoundLeavesEveryAuditReceiptUnchanged(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	var checked int
	for _, bucket := range Buckets {
		dir := filepath.Join(root, IntentsRelDir, bucket)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !intentFileRe.MatchString(e.Name()) {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			for _, headRe := range []*regexp.Regexp{acHeadingRe, auditHeadingRe, mechanismHeadingRe, scopeHeadingRe} {
				if got, want := sectionBody(content, headRe), fenceBlindSectionBody(content, headRe); got != want {
					t.Errorf("%s: the fence-aware bound changed a section body\n got: %q\nwant: %q", e.Name(), got, want)
				}
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("read no intent records — the corpus proof asserted nothing")
	}
	t.Logf("checked %d intent records", checked)
}

// fenceBlindSectionBody is the pre-fence-tracking rule, kept here as the
// baseline the corpus is compared against. It exists only for this test.
func fenceBlindSectionBody(content string, headRe *regexp.Regexp) string {
	lines := strings.Split(content, "\n")
	for i, ln := range lines {
		if !headRe.MatchString(strings.TrimRight(ln, "\r")) {
			continue
		}
		var body []string
		for _, b := range lines[i+1:] {
			if headingRe.MatchString(strings.TrimRight(b, "\r")) {
				break
			}
			body = append(body, b)
		}
		return strings.Join(body, "\n")
	}
	return ""
}

// TestStampPlannedHoldsTheMintLock: stampPlanned is a read-modify-write over a
// record two sessions can reach at once, and the ids it writes come from a mint
// that reads no ledger. It takes the same advisory lock every other mint in this
// package takes, so a concurrent create cannot interleave with it.
func TestStampPlannedHoldsTheMintLock(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n# alpha\n\n"+
			"## Scope Conditions\n\n- holds on POSIX\n\n## Acceptance Criteria\n\n- ok\n")

	held := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- withIntentMintLock(root, func() error {
			close(held)
			// Hold the lock long enough that an unlocked stamp would finish inside it.
			time.Sleep(150 * time.Millisecond)
			return nil
		})
	}()
	<-held
	start := time.Now()
	if _, err := Plan(root, "itd-10"); err != nil {
		t.Fatal(err)
	}
	waited := time.Since(start)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if waited < 100*time.Millisecond {
		t.Fatalf("the stamp completed in %v while the mint lock was held — it took no lock", waited)
	}
}

// commentedSection is iss-2608300259316871: a condition parked inside an HTML
// comment. Read as live it is counted as a condition and STAMPED — and the
// marker's own `-->` then closes the outer comment early, so the rest of the
// parked block renders as prose. A rewrite that corrupts what the file renders
// is the worst thing this stamper can do, so the section is refused whole.
const commentedSection = "- holds on a POSIX shell\n\n<!--\n- parked while we decide\n-->\n"

func TestCommentedBulletIsNotACondition(t *testing.T) {
	c := ParseClaims(claimRecord(nil, str(commentedSection)))
	if len(c.Conditions) != 1 {
		t.Fatalf("got %d conditions, want 1 (the parked bullet is not live): %+v", len(c.Conditions), c.Conditions)
	}
	if !c.ConditionsCommented {
		t.Error("the section must be reported as carrying a comment span")
	}
}

// A bullet whose own line opens a comment that closes on the next line is not a
// live bullet either — the whole span is masked.
func TestBulletOpeningACommentIsMasked(t *testing.T) {
	body := "- holds on a POSIX shell\n- parked <!--\n  because we are unsure -->\n"
	c := ParseClaims(claimRecord(nil, str(body)))
	if len(c.Conditions) != 1 {
		t.Fatalf("got %d conditions, want 1: %+v", len(c.Conditions), c.Conditions)
	}
	if c.Conditions[0].Text != "holds on a POSIX shell" {
		t.Fatalf("read the parked bullet: %q", c.Conditions[0].Text)
	}
	if !c.ConditionsCommented {
		t.Error("the section must be reported as carrying a comment span")
	}
}

func TestStampRefusesACommentedSection(t *testing.T) {
	content := claimRecord(nil, str(commentedSection))
	got, n, err := stampScopeConditions(content, fixedMinter(7))
	if err == nil {
		t.Fatal("a section carrying a comment span must refuse the stamp")
	}
	if !strings.Contains(err.Error(), "comment") {
		t.Fatalf("the refusal must name the comment, got %q", err)
	}
	if n != 0 || got != "" {
		t.Fatalf("a refusal writes nothing (n=%d)", n)
	}
}

// TestPlanLeavesACommentedSectionByteIdentical is the claim that matters: the
// refusal reaches the disk, not just the function.
func TestPlanLeavesACommentedSectionByteIdentical(t *testing.T) {
	root := t.TempDir()
	before := "---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n# alpha\n\n" +
		"## Scope Conditions\n\n" + commentedSection + "\n## Acceptance Criteria\n\n- ok\n"
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", before)

	if _, err := Plan(root, "itd-10"); err == nil {
		t.Fatal("Plan must refuse a section carrying a comment span")
	}
	after, err := os.ReadFile(filepath.Join(root, plannedDir, "itd-10-alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != before {
		t.Fatalf("the record was rewritten by a refused stamp:\n%s", after)
	}
}
