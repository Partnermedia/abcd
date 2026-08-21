package lint

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeBaselineFile plants a raw baseline document so a test can exercise the
// LOADER against bytes a human (or a hostile branch) could commit, rather than
// against a struct the loader would have produced itself.
func writeBaselineFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "citations-baseline.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestBaselineRoundTrip pins the save/load round-trip: what SaveBaseline writes
// is exactly what LoadBaseline returns, so the refresh verb and the gate agree
// on the record without a lossy hop.
func TestBaselineRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "citations-baseline.json")
	want := Baseline{
		SchemaVersion: BaselineSchemaVersion,
		Entries: map[string]BaselineEntry{
			"https://example.org/a": {
				FinalURL:     "https://example.org/a",
				LastChecked:  "2026-07-01",
				Outcome:      OutcomeAlive,
				Verification: VerificationAutomatic,
				VerifiedOn:   "2026-07-01",
			},
			"https://example.org/b": {
				URL:          "https://example.org/b",
				FinalURL:     "https://example.org/b/final",
				LastChecked:  "2026-06-02",
				Outcome:      OutcomeAlive,
				Verification: VerificationManual,
				VerifiedOn:   "2026-06-02",
			},
		},
	}
	if err := SaveBaseline(path, want); err != nil {
		t.Fatalf("SaveBaseline: %v", err)
	}
	got, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	if got.SchemaVersion != want.SchemaVersion {
		t.Errorf("schema_version = %d, want %d", got.SchemaVersion, want.SchemaVersion)
	}
	if len(got.Entries) != len(want.Entries) {
		t.Fatalf("loaded %d entries, want %d", len(got.Entries), len(want.Entries))
	}
	for k, w := range want.Entries {
		g, ok := got.Entries[k]
		if !ok {
			t.Fatalf("entry %q missing after round-trip", k)
		}
		if g.FinalURL != w.FinalURL || g.LastChecked != w.LastChecked ||
			g.Outcome != w.Outcome || g.Verification != w.Verification || g.VerifiedOn != w.VerifiedOn {
			t.Errorf("entry %q = %+v, want %+v", k, g, w)
		}
	}
}

// TestBaselineRejectsMethodField is AC 2's teeth: the schema must be
// STRUCTURALLY incapable of carrying how a human verified a link. There is no
// method field to populate, and DisallowUnknownFields means a smuggled one is a
// load-time refusal rather than a silently-ignored key.
func TestBaselineRejectsMethodField(t *testing.T) {
	for _, smuggled := range []string{"method", "how", "transcript", "headers", "notes", "evidence"} {
		body := `{"schema_version":1,"entries":{"https://example.org/a":{
			"final_url":"https://example.org/a","last_checked":"2026-07-01",
			"outcome":"alive","verification":"manual","verified_on":"2026-07-01",
			"` + smuggled + `":"clicked it in a browser"}}}`
		path := writeBaselineFile(t, body)
		_, err := LoadBaseline(path)
		if err == nil {
			t.Fatalf("LoadBaseline accepted a smuggled %q field; the schema must be incapable of recording how (AC 2)", smuggled)
		}
		var be *BaselineError
		if !errors.As(err, &be) {
			t.Fatalf("smuggled %q yielded %T (%v), want a typed *BaselineError", smuggled, err, err)
		}
	}
}

// TestBaselineRejectsUnknownTopLevelField extends the refusal to the document
// root: a stray top-level key is the same smuggling vector one level up.
func TestBaselineRejectsUnknownTopLevelField(t *testing.T) {
	path := writeBaselineFile(t, `{"schema_version":1,"entries":{},"method":"curl"}`)
	if _, err := LoadBaseline(path); err == nil {
		t.Fatal("LoadBaseline accepted an unknown top-level field")
	}
}

// TestBaselineDeclaredKeyTolerance pins the guard-reviewer lesson: an entry MAY
// restate its own map key in a url field (a human-readable, diffable record),
// and that is accepted; a MISMATCH between the declared url and the map key is
// refused, because then two different URLs claim one receipt.
func TestBaselineDeclaredKeyTolerance(t *testing.T) {
	agreeing := writeBaselineFile(t, `{"schema_version":1,"entries":{"https://example.org/a":{
		"url":"https://example.org/a","final_url":"https://example.org/a","last_checked":"2026-07-01",
		"outcome":"alive","verification":"automatic","verified_on":"2026-07-01"}}}`)
	if _, err := LoadBaseline(agreeing); err != nil {
		t.Fatalf("LoadBaseline refused an entry whose declared url matches its key: %v", err)
	}

	mismatched := writeBaselineFile(t, `{"schema_version":1,"entries":{"https://example.org/a":{
		"url":"https://example.org/OTHER","final_url":"https://example.org/a","last_checked":"2026-07-01",
		"outcome":"alive","verification":"automatic","verified_on":"2026-07-01"}}}`)
	err := LoadBaseline2Err(t, mismatched)
	if !strings.Contains(err.Error(), "https://example.org/OTHER") {
		t.Errorf("mismatch error should name the disagreeing url; got %v", err)
	}
}

// LoadBaseline2Err loads a baseline that MUST fail and returns the typed error.
func LoadBaseline2Err(t *testing.T, path string) error {
	t.Helper()
	_, err := LoadBaseline(path)
	if err == nil {
		t.Fatal("LoadBaseline succeeded where a refusal was required")
	}
	var be *BaselineError
	if !errors.As(err, &be) {
		t.Fatalf("got %T (%v), want a typed *BaselineError", err, err)
	}
	return err
}

// TestBaselineValidatesFields pins every field-level refusal. Each case is a
// baseline a committer could write by hand; the loader must fail closed rather
// than let the gate reason about a half-formed record.
func TestBaselineValidatesFields(t *testing.T) {
	entry := func(fields string) string {
		return `{"schema_version":1,"entries":{"https://example.org/a":{` + fields + `}}}`
	}
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "unsupported schema version",
			body: `{"schema_version":99,"entries":{}}`,
			want: "schema_version",
		},
		{
			name: "missing schema version",
			body: `{"entries":{}}`,
			want: "schema_version",
		},
		{
			name: "empty url key",
			body: `{"schema_version":1,"entries":{"":{"final_url":"https://example.org/a","last_checked":"2026-07-01","outcome":"alive","verification":"automatic","verified_on":"2026-07-01"}}}`,
			want: "empty url key",
		},
		{
			name: "missing final address",
			body: entry(`"last_checked":"2026-07-01","outcome":"alive","verification":"automatic","verified_on":"2026-07-01"`),
			want: "final_url",
		},
		{
			name: "unparsable last_checked",
			body: entry(`"final_url":"https://example.org/a","last_checked":"yesterday","outcome":"alive","verification":"automatic","verified_on":"2026-07-01"`),
			want: "last_checked",
		},
		{
			name: "unknown outcome",
			body: entry(`"final_url":"https://example.org/a","last_checked":"2026-07-01","outcome":"probably-fine","verification":"automatic","verified_on":"2026-07-01"`),
			want: "outcome",
		},
		{
			name: "unknown verification label",
			body: entry(`"final_url":"https://example.org/a","last_checked":"2026-07-01","outcome":"alive","verification":"semi-automatic","verified_on":"2026-07-01"`),
			want: "verification",
		},
		{
			name: "missing verified_on",
			body: entry(`"final_url":"https://example.org/a","last_checked":"2026-07-01","outcome":"alive","verification":"manual"`),
			want: "verified_on",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := LoadBaseline2Err(t, writeBaselineFile(t, tc.body))
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not name %q", err, tc.want)
			}
		})
	}
}

// TestBaselineHasNoMethodFieldByConstruction asserts the type itself, not just
// the decoder: reflection over BaselineEntry must find no field whose json tag
// could record HOW a link was verified. A future edit that adds one fails here.
func TestBaselineHasNoMethodFieldByConstruction(t *testing.T) {
	banned := map[string]bool{
		"method": true, "how": true, "transcript": true, "headers": true,
		"notes": true, "evidence": true, "tool": true, "user_agent": true, "proof": true,
	}
	for _, tag := range baselineEntryJSONTags() {
		if banned[tag] {
			t.Errorf("BaselineEntry declares a %q field; the schema must be incapable of recording how a human verified (AC 2)", tag)
		}
	}
}

// TestLoadBaselineMissingFile reports a missing baseline as os.ErrNotExist, so
// the gate can tell "no baseline yet" from "a corrupt baseline".
func TestLoadBaselineMissingFile(t *testing.T) {
	_, err := LoadBaseline(filepath.Join(t.TempDir(), "absent.json"))
	if !os.IsNotExist(err) {
		t.Fatalf("LoadBaseline on a missing file = %v, want an os.ErrNotExist", err)
	}
}

// TestBaselineRefusesControlRunesInFinalURL pins the defence-in-depth rung
// behind the fetch boundary (iss-359): whatever produced the entry, a
// final_url carrying a bidi override or C1 escape must not round-trip through
// the committed record. The rune is assembled numerically — a raw byte in a
// fixture decodes to U+FFFD and tests nothing.
func TestBaselineRefusesControlRunesInFinalURL(t *testing.T) {
	b := Baseline{Entries: map[string]BaselineEntry{
		"https://example.com/a": {
			FinalURL:     "https://example.com/a?x=" + string(rune(0x202E)),
			LastChecked:  "2026-08-20",
			Outcome:      OutcomeAlive,
			Verification: VerificationAutomatic,
		},
	}}
	err := SaveBaseline(filepath.Join(t.TempDir(), "citations-baseline.json"), b)
	var be *BaselineError
	if !errors.As(err, &be) {
		t.Fatalf("SaveBaseline accepted a final_url with U+202E: err = %v", err)
	}
	if !strings.Contains(err.Error(), "control or invisible") {
		t.Errorf("refusal should name the rune class, got: %v", err)
	}
}

// TestBaselineErrorSanitisesEchoedFields is iss-384: validate() echoes the
// entry key and the offending value into a *BaselineError that reaches the
// terminal through the CLI error surface, which applies no sanitiser. A
// malformed field carrying a bidi/C1/zero-width rune must not replay raw — the
// iss-359 rung guarded final_url; its sibling fields need the same.
func TestBaselineErrorSanitisesEchoedFields(t *testing.T) {
	rlo := string(rune(0x202E))
	body := `{"schema_version":1,"entries":{"https://example.org/a` + rlo +
		`":{"final_url":"https://example.org/a","last_checked":"2026-07-01","outcome":"alive` + rlo +
		`","verification":"automatic","verified_on":"2026-07-01"}}}`
	err := LoadBaseline2Err(t, writeBaselineFile(t, body))
	if err == nil {
		t.Fatal("a malformed outcome with a bidi rune was accepted")
	}
	if strings.ContainsRune(err.Error(), 0x202E) {
		t.Errorf("BaselineError replayed a raw U+202E to the terminal: %q", err.Error())
	}
}
