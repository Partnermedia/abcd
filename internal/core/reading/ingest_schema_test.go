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

// TestMissingBodyFieldRefusesTheItem is the body schema's other direction, and
// it exists because removing the guard left every other case green: an item
// carrying a FOREIGN body trips the unknown-key check first, so nothing reached
// the missing-field branch (iss-2608311206480573). What it catches is a reading
// that returned a partial item at its OWN position — a tension named with no
// account of why it is one.
//
// Both the empty and the absent form are tried; only one of them is an obviously
// missing field in the payload bytes.
func TestMissingBodyFieldRefusesTheItem(t *testing.T) {
	fields := issueschema.ReadingBodyFields["detection"]
	if len(fields) < 2 {
		t.Fatalf("the detection body declares %d field(s); this case needs one to remove", len(fields))
	}
	for _, field := range fields {
		field := field
		for _, form := range []string{"empty", "absent"} {
			t.Run(field+"/"+form, func(t *testing.T) {
				f := newIngestFixture(t, "detection")
				doc := f.payload(2)
				item := doc["items"].([]any)[1].(map[string]any)
				if form == "empty" {
					item[field] = ""
				} else {
					delete(item, field)
				}

				r := f.refusedItem(doc, 2, 2)
				if r.Rule != "missing-body-field" {
					t.Errorf("the refusal cites rule %q", r.Rule)
				}
				if !strings.Contains(r.Field, field) {
					t.Errorf("the refusal names field %q, want %q", r.Field, field)
				}
				if !strings.Contains(r.Detail, "item 2") {
					t.Errorf("the refusal does not name the item's ordinal: %q", r.Detail)
				}
			})
		}
	}
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

// hostileRune reports whether r is one of the classes termsafe.Sanitize masks:
// a C0 control (ESC among them), DEL, the C1 range, a bidi override or isolate,
// or a zero-width character. The list is stated here rather than imported so
// this case fails if the sanitiser stops masking one of them.
func hostileRune(r rune) bool {
	switch {
	case r < 0x20 && r != '\n':
		return true
	case r == 0x7f, r >= 0x80 && r <= 0x9f:
		return true
	case r >= 0x202a && r <= 0x202e, r >= 0x2066 && r <= 0x2069:
		return true
	case r == 0x200b, r == 0x200c, r == 0x200d, r == 0xfeff:
		return true
	}
	return false
}

// hostileText carries one rune from each class the sanitiser masks: an ANSI
// escape that could recolour or overprint the message reporting it, a bell, a
// right-to-left override, and a zero-width space.
const hostileText = "\x1b[31m\x07\u202ereversed\u200b\x1b[2K"

// TestARefusalNeverEchoesRawPayloadBytes is the trust boundary at the OUTPUT
// end. A refusal quotes model-produced text back to a terminal and writes it
// into a durable record, so a payload carrying an escape sequence could rewrite
// the very message that reports it, and a payload carrying megabytes in one
// field could drown it.
//
// Both guards were mutation-vacuous when they were written — neutralising the
// sanitiser and neutralising the cap each left every test green
// (iss-2608311211235195) — so this case walks the fields an untrusted value
// actually reaches a message through, one at a time.
func TestARefusalNeverEchoesRawPayloadBytes(t *testing.T) {
	long := strings.Repeat("z", 3000)

	for _, field := range []string{"_type", "run_id", "position", "regime", "manifest_sha256"} {
		field := field
		for _, oversized := range []bool{false, true} {
			name, value := field+"/hostile", hostileText
			if oversized {
				name, value = field+"/oversized", long
			}
			t.Run(name, func(t *testing.T) {
				f := newIngestFixture(t, "detection")
				doc := f.payload(1)
				doc[field] = value

				_, err := f.ingest(doc)
				if err == nil {
					t.Fatalf("%s carrying untrusted text was accepted", field)
				}
				assertSafeEcho(t, err.Error(), oversized)
			})
		}
	}

	// An item's KEY is payload text too, and it reaches a message through the
	// unknown-field refusal.
	t.Run("item key", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		doc := f.payload(1)
		doc["items"].([]any)[0].(map[string]any)[hostileText] = "x"
		_, err := f.ingest(doc)
		if err == nil {
			t.Fatal("an item key carrying terminal-attack runes was accepted")
		}
		assertSafeEcho(t, err.Error(), false)
	})

	// And the DURABLE record: a list-level refusal writes the instrument's model
	// into refusal.json, where the same rule has to hold.
	t.Run("durable refusal record", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		doc := f.payload(1)
		doc["regime"] = RegimeEvaluative
		doc["instrument"].(map[string]any)["model"] = hostileText + long

		if _, err := f.ingest(doc); err == nil {
			t.Fatal("a regime mismatch was accepted")
		}
		rec := f.readRefusalRecord(f.runID)
		assertSafeEcho(t, rec.Instrument.Model, true)
		assertSafeEcho(t, rec.Reason, false)
	})
}

// assertSafeEcho holds one rendered message to both halves of the rule: no
// terminal-display attack rune survives, and a payload value that arrived
// oversized is capped rather than reproduced.
func assertSafeEcho(t *testing.T, msg string, expectCapped bool) {
	t.Helper()
	for i, r := range msg {
		if hostileRune(r) {
			t.Errorf("the message carries U+%04X at byte %d, which the sanitiser masks: %q", r, i, msg)
			return
		}
	}
	if expectCapped && strings.Contains(msg, strings.Repeat("z", maxEchoedRunes+1)) {
		t.Errorf("an oversized payload value reached the message uncapped (%d bytes): %.200q", len(msg), msg)
	}
}
