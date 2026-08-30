package capture

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// One item in, one record out — each under the run's own directory, each
// carrying the run identifier in its envelope. The run-scoped identifier is what
// lets the ledger say WHICH visible world a finding was returned under.
func TestIngestWritesOneRecordPerItem(t *testing.T) {
	repo, ir := ledger(t)
	run := "rdg-2608300000000001"
	res, err := IngestReading(IngestReadingRequest{
		RepoRoot: repo, IssuesRoot: ir,
		Run: run, Manifest: "sha256:beef",
		Position: "detection", Regime: "supplied",
		Items: []ReadingItem{
			{Pattern: "constraint one", Body: bodyFor("detection")},
			{Pattern: "constraint two", Body: bodyFor("detection")},
			{Pattern: "constraint three", Body: bodyFor("detection")},
		},
	})
	if err != nil {
		t.Fatalf("IngestReading: %v", err)
	}
	if len(res.Records) != 3 {
		t.Fatalf("wrote %d records, want one per item (3)", len(res.Records))
	}

	runDir := filepath.Join(ir, issueschema.ReadingsDir, run)
	entries, err := os.ReadDir(runDir)
	if err != nil {
		t.Fatalf("read run dir: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("run dir holds %d files, want 3", len(entries))
	}
	seen := map[string]bool{}
	for _, rec := range res.Records {
		if seen[rec.ID] {
			t.Fatalf("id %s written twice", rec.ID)
		}
		seen[rec.ID] = true
		content, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(rec.Path)))
		if err != nil {
			t.Fatalf("read %s: %v", rec.Path, err)
		}
		if !strings.Contains(string(content), "run: \""+run+"\"") {
			t.Fatalf("%s does not carry its run identifier:\n%s", rec.ID, content)
		}
	}
}

// Two runs returning the same tension carry different ids, because an id is a
// mint (adr-45) and never content-derived. Without that, a re-raise is
// indistinguishable from its first appearance and the recurrence signal dies.
func TestTwoRunsSameTensionMintDistinctIDs(t *testing.T) {
	repo, ir := ledger(t)
	same := ReadingItem{Pattern: "the same stated constraint", Body: bodyFor("detection")}

	var ids []string
	for _, run := range []string{"rdg-2608300000000001", "rdg-2608300000000002"} {
		res, err := IngestReading(IngestReadingRequest{
			RepoRoot: repo, IssuesRoot: ir,
			Run: run, Manifest: "sha256:beef",
			Position: "detection", Regime: "supplied",
			Items: []ReadingItem{same},
		})
		if err != nil {
			t.Fatalf("IngestReading(%s): %v", run, err)
		}
		ids = append(ids, res.Records[0].ID)
	}
	if ids[0] == ids[1] {
		t.Fatalf("two runs returning the same tension minted one id (%s); a re-raise must stay distinguishable", ids[0])
	}
}

// A second disposition for one item must say which one it replaces. The standing
// disposition of an item is the one no sibling supersedes, and a second record
// that cites nothing leaves two standing answers with no way to tell which is in
// force — while a hold that vanished when it was answered would take its own
// exit condition with it, so the superseded record stays in place.
func TestSecondDispositionForOneItemRequiresSupersedes(t *testing.T) {
	repo, ir, item := readingFixture(t, "detection")
	first, err := Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: item,
		State: issueschema.DispositionHeld, ExitCondition: "the closing run returns it again",
	})
	if err != nil {
		t.Fatalf("first disposition: %v", err)
	}

	_, err = Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: item,
		State: issueschema.DispositionAccepted, Grounds: "the closing run returned it",
	})
	if !errors.Is(err, ErrInvariantViolation) {
		t.Fatalf("second disposition without --supersedes: err = %v, want ErrInvariantViolation", err)
	}
	if !strings.Contains(err.Error(), first.ID) {
		t.Fatalf("the refusal must name the standing disposition to supersede; got %v", err)
	}

	second, err := Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: item,
		State: issueschema.DispositionAccepted, Grounds: "the closing run returned it",
		Supersedes: first.ID,
	})
	if err != nil {
		t.Fatalf("second disposition citing %s: %v", first.ID, err)
	}
	if second.ID == first.ID {
		t.Fatal("the superseding disposition must have its own id")
	}
	if _, err := os.Stat(filepath.Join(repo, filepath.FromSlash(first.Path))); err != nil {
		t.Fatalf("the superseded record must stay in place: %v", err)
	}
}
