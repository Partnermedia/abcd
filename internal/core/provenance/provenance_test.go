package provenance

import "testing"

func TestParseOriginThreeForms(t *testing.T) {
	cases := []struct {
		in   string
		want Origin
	}{
		{"researcher-authored", Origin{Kind: KindResearcherAuthored}},
		{"extracted-from-record", Origin{Kind: KindExtractedFromRecord}},
		{"contributed-by-reading rdg-3/rdi-17", Origin{Kind: KindContributedByReading, Run: "rdg-3", Item: "rdi-17"}},
	}
	for _, c := range cases {
		got, err := ParseOrigin(c.in)
		if err != nil {
			t.Fatalf("ParseOrigin(%q): unexpected error %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("ParseOrigin(%q) = %+v, want %+v", c.in, got, c.want)
		}
		if got.String() != c.in {
			t.Errorf("round trip: %q rendered back as %q", c.in, got.String())
		}
	}
}

func TestParseOriginRejectsUnknownKind(t *testing.T) {
	for _, in := range []string{
		"",
		"researcher authored",
		"Researcher-Authored",
		"invented-by-nobody",
		"researcher-authored rdg-1/rdi-1",
		"extracted-from-record something",
	} {
		if got, err := ParseOrigin(in); err == nil {
			t.Errorf("ParseOrigin(%q) = %+v, want an error", in, got)
		}
	}
}

func TestParseOriginReadingPointer(t *testing.T) {
	// A pointer that is missing, half-present, or malformed is refused: the
	// value's whole job is to resolve to a reading record.
	for _, in := range []string{
		"contributed-by-reading",
		"contributed-by-reading ",
		"contributed-by-reading rdg-3",
		"contributed-by-reading rdg-3/",
		"contributed-by-reading /rdi-17",
		"contributed-by-reading rdi-17/rdg-3",
		"contributed-by-reading rdg-3/rdi-17/extra",
		"contributed-by-reading rdg-x/rdi-17",
		"contributed-by-reading rdg-3/rdi-",
		"contributed-by-reading  rdg-3/rdi-17",
	} {
		if got, err := ParseOrigin(in); err == nil {
			t.Errorf("ParseOrigin(%q) = %+v, want an error", in, got)
		}
	}
}

func TestProductionModeVocabularyIsClosed(t *testing.T) {
	for _, ok := range []string{"hand-written", "dictated-and-formatted", "scribe-transcribed"} {
		if _, err := ParseMode(ok); err != nil {
			t.Errorf("ParseMode(%q): unexpected error %v", ok, err)
		}
	}
	for _, bad := range []string{"", "Hand-Written", "handwritten", "typed", "hand-written "} {
		if got, err := ParseMode(bad); err == nil {
			t.Errorf("ParseMode(%q) = %q, want an error", bad, got)
		}
	}
	// The empty value defaults ONLY through the defaulting door, so a writer
	// cannot absorb an unset mode by accident.
	if got, err := ModeOrDefault(""); err != nil || got != DefaultMode {
		t.Errorf("ModeOrDefault(\"\") = %q, %v; want %q, nil", got, err, DefaultMode)
	}
	if _, err := ModeOrDefault("typed"); err == nil {
		t.Error("ModeOrDefault(\"typed\"): want an error")
	}
}

func TestNewStampRefusesTheUnmintableOrigin(t *testing.T) {
	s, err := NewStamp(KindResearcherAuthored, "")
	if err != nil {
		t.Fatalf("NewStamp: unexpected error %v", err)
	}
	if s.OriginValue() != "researcher-authored" || s.ModeValue() != string(DefaultMode) {
		t.Errorf("NewStamp defaulted to %q/%q", s.OriginValue(), s.ModeValue())
	}
	// contributed-by-reading carries a pointer no command in this repository can
	// supply yet, so no write path may mint it.
	if _, err := NewStamp(KindContributedByReading, "hand-written"); err == nil {
		t.Error("NewStamp(KindContributedByReading): want a refusal")
	}
	if _, err := NewStamp("invented", "hand-written"); err == nil {
		t.Error("NewStamp(invented): want a refusal")
	}
	if _, err := NewStamp(KindResearcherAuthored, "typed"); err == nil {
		t.Error("NewStamp with an out-of-vocabulary mode: want a refusal")
	}
}
