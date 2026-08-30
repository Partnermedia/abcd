package capture

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/core/recordid"
)

// readingFixture ingests one run of one detection item into a fresh ledger and
// returns the roots plus the minted item id.
func readingFixture(t *testing.T, position string) (repo, ir, item string) {
	t.Helper()
	repo, ir = ledger(t)
	res, err := IngestReading(IngestReadingRequest{
		RepoRoot: repo, IssuesRoot: ir,
		Run: "rdg-2608300000000001", Manifest: "sha256:" + strings.Repeat("a", 64),
		Position: position, Regime: "supplied",
		Items: []ReadingItem{{Pattern: "a stated constraint", Body: bodyFor(position)}},
	})
	if err != nil {
		t.Fatalf("IngestReading: %v", err)
	}
	if len(res.Records) != 1 {
		t.Fatalf("IngestReading wrote %d records, want 1", len(res.Records))
	}
	return repo, ir, res.Records[0].ID
}

// bodyFor fills every body field the position declares with a placeholder, so a
// test that is not about the body never fails on it.
func bodyFor(position string) map[string]string {
	body := map[string]string{}
	for _, f := range issueschema.ReadingBodyFields[position] {
		body[f] = "text for " + f
	}
	return body
}

// A reading record is an additionalProperties:false record like every other
// record in this ledger: a key outside the schema is refused at the boundary,
// never accepted and flagged. Accepting it would put a property in the committed
// tree that no reader reads and no gate judges.
func TestReadingRecordRefusesUnknownProperty(t *testing.T) {
	repo, ir := ledger(t)
	_, err := IngestReading(IngestReadingRequest{
		RepoRoot: repo, IssuesRoot: ir,
		Run: "rdg-2608300000000001", Manifest: "sha256:beef",
		Position: "detection", Regime: "supplied",
		Items: []ReadingItem{{
			Pattern: "a stated constraint",
			Body:    map[string]string{"tension": "t", "constraint_in_play": "c", "why_a_tension": "w", "vibes": "no"},
		}},
	})
	if !errors.Is(err, ErrMalformedFrontmatter) {
		t.Fatalf("IngestReading with an unknown body property err = %v, want ErrMalformedFrontmatter", err)
	}
	if !strings.Contains(err.Error(), "vibes") {
		t.Fatalf("the refusal must name the offending property; got %v", err)
	}
}

// One record type, four bodies, held as data: an item at a position must carry
// exactly the body that position declares. A missing field is refused and the
// refusal names the position's set, so the caller is told the rule rather than
// left to infer it.
func TestReadingRecordRequiresPositionBodyFields(t *testing.T) {
	for _, p := range issueschema.ReadingPositions {
		t.Run(p.Position, func(t *testing.T) {
			for _, omit := range p.Fields {
				body := bodyFor(p.Position)
				delete(body, omit)
				repo, ir := ledger(t)
				_, err := IngestReading(IngestReadingRequest{
					RepoRoot: repo, IssuesRoot: ir,
					Run: "rdg-2608300000000001", Manifest: "sha256:beef",
					Position: p.Position, Regime: "supplied",
					Items: []ReadingItem{{Pattern: "a stated constraint", Body: body}},
				})
				if !errors.Is(err, ErrMissingRequiredField) {
					t.Fatalf("omitting %q at %s: err = %v, want ErrMissingRequiredField", omit, p.Position, err)
				}
				if !strings.Contains(err.Error(), omit) || !strings.Contains(err.Error(), p.Position) {
					t.Fatalf("the refusal must name the field and the position; got %v", err)
				}
			}
		})
	}
}

// The grounds are what make a disposition a judgement rather than a vote. They
// are required on every state except `held`, whose exit condition carries the
// same weight.
func TestDispositionRefusesEmptyGrounds(t *testing.T) {
	repo, ir, item := readingFixture(t, "detection")
	for _, state := range []string{
		issueschema.DispositionAccepted, issueschema.DispositionRejected,
	} {
		_, err := Disposition(DispositionRequest{
			RepoRoot: repo, IssuesRoot: ir, Item: item, State: state, Grounds: "   ",
		})
		if !errors.Is(err, ErrMissingRequiredField) {
			t.Fatalf("%s with empty grounds: err = %v, want ErrMissingRequiredField", state, err)
		}
		if !strings.Contains(err.Error(), "disposition_grounds") {
			t.Fatalf("the refusal must name disposition_grounds; got %v", err)
		}
	}
}

// A hold with no exit condition is a parking space, which is the thing `open`
// already is and the thing `held` exists not to be.
func TestHeldDispositionRefusesEmptyExitCondition(t *testing.T) {
	repo, ir, item := readingFixture(t, "detection")
	_, err := Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: item,
		State: issueschema.DispositionHeld, ExitCondition: "",
	})
	if !errors.Is(err, ErrMissingRequiredField) {
		t.Fatalf("held with no exit condition: err = %v, want ErrMissingRequiredField", err)
	}
	if !strings.Contains(err.Error(), "exit_condition") {
		t.Fatalf("the refusal must name exit_condition; got %v", err)
	}
}

// The availability rule is a coupling the schema carries: the disposition reads
// `position` off the KEYED reading record and checks the table. `declined` at a
// detection would manufacture a principle never at stake; `rejected` at the
// widening position would assert a purpose nobody proposed.
func TestDispositionRefusesStateUnavailableAtPosition(t *testing.T) {
	for _, tc := range []struct{ position, state string }{
		{"detection", issueschema.DispositionDeclined},
		{"widening", issueschema.DispositionRejected},
	} {
		repo, ir, item := readingFixture(t, tc.position)
		_, err := Disposition(DispositionRequest{
			RepoRoot: repo, IssuesRoot: ir, Item: item, State: tc.state, Grounds: "because",
		})
		if !errors.Is(err, ErrInvariantViolation) {
			t.Fatalf("%s at %s: err = %v, want ErrInvariantViolation", tc.state, tc.position, err)
		}
		if !strings.Contains(err.Error(), tc.state) || !strings.Contains(err.Error(), tc.position) {
			t.Fatalf("the refusal must name the availability rule it enforces; got %v", err)
		}
	}
}

// An orphan disposition — one keyed to an item no run returned — is refused by
// the SAME path that checks availability, because the check needs the item's
// position and cannot invent one.
func TestDispositionRefusesOrphanItem(t *testing.T) {
	repo, ir, _ := readingFixture(t, "detection")
	_, err := Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: "rdi-2608300000009999",
		State: issueschema.DispositionAccepted, Grounds: "because",
	})
	if !errors.Is(err, ErrUnknownIssueID) {
		t.Fatalf("orphan disposition: err = %v, want ErrUnknownIssueID", err)
	}
}

// The two-axis hold field is reserved and DORMANT. A populated value is refused
// until activation is ruled, which is what makes the reservation a behaviour
// rather than a comment: silently accepting one would populate a field whose
// meaning nobody has settled.
func TestPopulatedHoldAxisRefused(t *testing.T) {
	repo, ir, item := readingFixture(t, "detection")
	for _, tc := range []struct{ field, value string }{
		{"hold_frame_location", "the frame's cost element"},
		{"hold_moscow", "must"},
	} {
		req := DispositionRequest{
			RepoRoot: repo, IssuesRoot: ir, Item: item,
			State: issueschema.DispositionHeld, ExitCondition: "the closing run returns it again",
		}
		if tc.field == "hold_frame_location" {
			req.HoldFrameLocation = tc.value
		} else {
			req.HoldMoscow = tc.value
		}
		_, err := Disposition(req)
		if !errors.Is(err, ErrInvariantViolation) {
			t.Fatalf("populated %s: err = %v, want ErrInvariantViolation", tc.field, err)
		}
		if !strings.Contains(err.Error(), tc.field) {
			t.Fatalf("the refusal must name the reserved field; got %v", err)
		}
	}
}

// The reserved surprise key gets the same posture as the hold axes: reserved in
// the family now, populated in Iteration 2, and refused until then.
func TestPopulatedSurpriseKeyRefused(t *testing.T) {
	repo, ir := ledger(t)
	_, err := IngestReading(IngestReadingRequest{
		RepoRoot: repo, IssuesRoot: ir,
		Run: "rdg-2608300000000001", Manifest: "sha256:beef",
		Position: "detection", Regime: "supplied",
		Items: []ReadingItem{{
			Pattern: "a stated constraint", Body: bodyFor("detection"),
			OccasionedBy: "iss-1",
		}},
	})
	if !errors.Is(err, ErrInvariantViolation) {
		t.Fatalf("populated occasioned_by: err = %v, want ErrInvariantViolation", err)
	}
	if !strings.Contains(err.Error(), "occasioned_by") {
		t.Fatalf("the refusal must name the reserved field; got %v", err)
	}
}

// A disposition lands under a directory keyed by the ITEM, and its own file is
// keyed by its own id — the shape that lets a superseding disposition cite the
// one it replaces.
func TestDispositionLandsInTheItemKeyedDirectory(t *testing.T) {
	repo, ir, item := readingFixture(t, "detection")
	res, err := Disposition(DispositionRequest{
		RepoRoot: repo, IssuesRoot: ir, Item: item,
		State: issueschema.DispositionAccepted, Grounds: "the tension is real and worth acting on",
	})
	if err != nil {
		t.Fatalf("Disposition: %v", err)
	}
	wantDir := filepath.ToSlash(filepath.Join(LedgerRelPath, issueschema.DispositionsDir, item))
	if !strings.HasPrefix(filepath.ToSlash(res.Path), wantDir+"/") {
		t.Fatalf("disposition path = %q, want it under %q", res.Path, wantDir)
	}
	if !strings.HasPrefix(filepath.Base(res.Path), issueschema.DispositionFamily+"-") {
		t.Fatalf("disposition filename = %q, want a dsp-N record", filepath.Base(res.Path))
	}
}

// A mint that collides must refuse, never overwrite. The id space is a UTC
// second plus four random digits, so two same-second draws CAN coincide — rare,
// but a silent overwrite is a committed record replaced by another record with
// no trace that either happened, which is the one outcome a ledger must not
// produce. The pinned minter here makes the coincidence certain rather than rare.
func TestReadingRecordRefusesToOverwriteAColludingID(t *testing.T) {
	repo, ir := ledger(t)
	setMinter(t, recordid.Minter{
		Now:     func() time.Time { return time.Date(2026, 8, 30, 11, 0, 0, 0, time.UTC) },
		Entropy: bytes.NewReader([]byte{0x00, 0x07, 0x00, 0x07}),
	})
	req := IngestReadingRequest{
		RepoRoot: repo, IssuesRoot: ir,
		Run: "rdg-2608300000000001", Manifest: "sha256:beef",
		Position: "detection", Regime: "supplied",
		Items: []ReadingItem{{Pattern: "the first item", Body: bodyFor("detection")}},
	}
	first, err := IngestReading(req)
	if err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	req.Items = []ReadingItem{{Pattern: "a different item", Body: bodyFor("detection")}}
	if _, err := IngestReading(req); !errors.Is(err, ErrDuplicateIssueID) {
		t.Fatalf("a colliding mint: err = %v, want ErrDuplicateIssueID", err)
	}
	content, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(first.Records[0].Path)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "the first item") {
		t.Fatalf("the first record must survive a colliding mint:\n%s", content)
	}
}
