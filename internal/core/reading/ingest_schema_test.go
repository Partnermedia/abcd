package reading

// ingest_schema_test.go is itd-185's ac-1 and the structural half of ac-8: a
// malformed output is refused, the refusal names the offending field, and
// nothing durable exists anywhere for that run.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// TestPatternFieldIsTheRecordEnvelopeField is the executable form of a claim
// this contract would otherwise only assert in prose: the payload's provenance
// key is the reading RECORD's own envelope field, so the payload, the record and
// the four definitions say one word for one thing.
func TestPatternFieldIsTheRecordEnvelopeField(t *testing.T) {
	found := false
	for _, f := range issueschema.ReadingRequired {
		if f == PatternField {
			found = true
		}
	}
	if !found {
		t.Fatalf("the payload's provenance key is %q, which is not one of the reading record's envelope "+
			"fields %v; a rename between the payload and the record is a translation nobody asked for",
			PatternField, issueschema.ReadingRequired)
	}
	for _, p := range Positions() {
		for _, field := range issueschema.ReadingBodyFields[string(p)] {
			if field == PatternField {
				t.Errorf("the %s position declares %q as a BODY field; the pattern is an envelope field, "+
					"because a universal core condition must not live in a variant part", p, field)
			}
		}
	}
}

// TestMalformedPayloadWritesNothing is ac-1. Four shapes of malformed, and after
// each the run has left nothing in the reading-record family, nothing in the
// readings tree, and nothing in the stage.
func TestMalformedPayloadWritesNothing(t *testing.T) {
	cases := []struct {
		name string
		doc  func(f *ingestFixture) any
		want string
	}{
		{"not json", func(f *ingestFixture) any { return nil }, ""},
		{"wrong type tag", func(f *ingestFixture) any {
			d := f.payload(1)
			d["_type"] = "abcd.reading.output/9"
			return d
		}, "_type"},
		{"run id is not a run id", func(f *ingestFixture) any {
			d := f.payload(1)
			d["run_id"] = "rdg-../../etc"
			return d
		}, "run_id"},
		{"unknown position", func(f *ingestFixture) any {
			d := f.payload(1)
			d["position"] = "speculative"
			return d
		}, "speculative"},
		{"no items", func(f *ingestFixture) any {
			d := f.payload(0)
			return d
		}, "no items"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newIngestFixture(t, "detection")
			var err error
			if tc.name == "not json" {
				path := filepath.Join(t.TempDir(), "output.json")
				if writeErr := os.WriteFile(path, []byte("{not json"), 0o644); writeErr != nil {
					t.Fatal(writeErr)
				}
				_, err = Ingest(IngestRequest{RepoRoot: f.root, OutputPath: path})
			} else {
				_, err = f.ingest(tc.doc(f))
			}
			if err == nil {
				t.Fatal("a malformed output was accepted")
			}
			if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the refusal does not name %q: %v", tc.want, err)
			}
			f.nothingDurable(f.runID)
		})
	}
}

// TestUnknownFieldRefusedAtEveryLevel is ac-1's "names the offending field" at
// each of the three levels a payload has: the envelope, the instrument, and the
// item. The adjacent legal payload is accepted in the same test, so a verb that
// refused everything could not pass it.
func TestUnknownFieldRefusedAtEveryLevel(t *testing.T) {
	f := newIngestFixture(t, "detection")

	envelope := f.payload(1)
	envelope["reviewer_notes"] = "smuggled"
	if _, err := f.ingest(envelope); err == nil {
		t.Error("an unknown ENVELOPE field was accepted")
	} else if !strings.Contains(err.Error(), "reviewer_notes") {
		t.Errorf("the envelope refusal does not name the field: %v", err)
	}

	instrument := f.payload(1)
	instrument["instrument"].(map[string]any)["temperature"] = "0.2"
	if _, err := f.ingest(instrument); err == nil {
		t.Error("an unknown INSTRUMENT field was accepted")
	} else if !strings.Contains(err.Error(), "temperature") {
		t.Errorf("the instrument refusal does not name the field: %v", err)
	}

	// Neither of the two above reached the item level, so nothing at all is
	// durable yet: an envelope that does not decode names no run to record
	// against.
	f.nothingDurable(f.runID)

	// The item level, where the refusal is the ITEM's and the rest of the
	// payload still lands.
	item := f.payload(2)
	item["items"].([]any)[1].(map[string]any)["confidence"] = "high"
	r := f.refusedItem(item, 2, 2)
	if r.Rule != "unknown-field" {
		t.Errorf("the item refusal cites rule %q", r.Rule)
	}
	if !strings.Contains(r.Field, "confidence") {
		t.Errorf("the item refusal does not name the field: %q", r.Field)
	}
}

// TestWrongPositionBodyIsUndecodable is the second half of ac-8: an item
// carrying another position's body is refused field by field, so a licence
// cannot be smuggled in under a body that belongs somewhere else.
func TestWrongPositionBodyIsUndecodable(t *testing.T) {
	f := newIngestFixture(t, "detection")
	doc := f.payload(1)
	item := map[string]any{PatternField: "a pattern"}
	// The comparative position's body, at the detection position.
	for _, field := range issueschema.ReadingBodyFields["comparative"] {
		item[field] = "text"
	}
	doc["items"] = []any{item}

	_, err := f.ingest(doc)
	if err == nil {
		t.Fatal("an item carrying another position's body was accepted")
	}
	for _, field := range issueschema.ReadingBodyFields["comparative"] {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("the refusal does not name the foreign field %q: %v", field, err)
		}
	}
	f.nothingDurableInTheLedger(f.runID)
	f.mustIngest(f.payload(1))
}

// TestNoDurableWriteBeforeValidation is ac-1 stated as an ordering: a payload
// whose LAST item is illegal leaves nothing behind, even though its earlier
// items are perfectly legal. Validation is step one of four, and the durable
// move happens only after the whole payload has been judged.
func TestNoDurableWriteBeforeValidation(t *testing.T) {
	f := newIngestFixture(t, "detection")
	doc := f.payload(3)
	// Make every item illegal, so the run is refused at list level rather than
	// landing two of three.
	for _, raw := range doc["items"].([]any) {
		delete(raw.(map[string]any), PatternField)
	}
	if _, err := f.ingest(doc); err == nil {
		t.Fatal("a payload whose every item was illegal was accepted")
	}
	if got := f.ledgerRecords(f.runID); len(got) != 0 {
		t.Errorf("the refused run wrote %v into the reading-record family", got)
	}
	if f.exists(IngestStageDir + "/" + f.runID) {
		t.Error("the refused run left a stage behind")
	}
	f.mustIngest(f.payload(3))
}
