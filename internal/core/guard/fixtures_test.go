package guard

import (
	"sort"
	"strings"
	"testing"
)

// TestBundledEntriesPassAdmissionGate is the admission gate of spc-16: a
// registry entry is admitted to the bundled defaults only when it ships the
// fixtures that prove it. The gate is a test over the embedded corpus, so an
// entry cannot ship without its proof and a regression in the matcher fails the
// build rather than silently un-guarding a session.
//
//   - every known-bad fixture fires ITS entry at the entry's declared tier;
//   - every known-good fixture fires NOTHING (a 100% true-negative rate — one
//     known-good that fires blocks admission);
//   - known-good is at least 40% of the entry's corpus.
func TestBundledEntriesPassAdmissionGate(t *testing.T) {
	r := Defaults()
	ids := make([]string, 0, len(r.Entries))
	for id := range r.Entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		t.Fatal("the bundled registry is empty")
	}

	for _, id := range ids {
		e := r.Entries[id]
		t.Run(id, func(t *testing.T) {
			bad, good := len(e.Fixtures.KnownBad), len(e.Fixtures.KnownGood)
			if bad == 0 {
				t.Fatalf("entry %s ships no known-bad fixtures", id)
			}
			if good == 0 {
				t.Fatalf("entry %s ships no known-good fixtures", id)
			}
			// good/(good+bad) >= 0.4, in integer arithmetic.
			if 10*good < 4*(good+bad) {
				t.Fatalf("entry %s: known-good is %d of %d fixtures, below the 40%% floor", id, good, good+bad)
			}

			want := VerdictBlock
			if e.Tier == TierWarn {
				want = VerdictWarn
			}
			for _, cmd := range e.Fixtures.KnownBad {
				d, err := r.Check(cmd)
				if err != nil {
					t.Fatalf("known-bad %q: unexpected error: %v", cmd, err)
				}
				if !contains(d.Matches, id) {
					t.Fatalf("known-bad %q did not fire entry %s (matches: %v)", cmd, id, d.Matches)
				}
				if d.Verdict != want {
					t.Fatalf("known-bad %q verdict = %q, want %q for a %s entry", cmd, d.Verdict, want, e.Tier)
				}
			}
			for _, cmd := range e.Fixtures.KnownGood {
				d, err := r.Check(cmd)
				if err != nil {
					t.Fatalf("known-good %q: unexpected error: %v", cmd, err)
				}
				if d.Verdict != VerdictAllow {
					t.Fatalf("known-good %q fired %v (verdict %q); the true-negative floor is 100%%", cmd, d.Matches, d.Verdict)
				}
			}
		})
	}
}

// TestBundledEntriesAreExplainable holds the teaching plane's end of the
// bargain: every entry carries a safe successor and a plain-language why, since
// the block message IS the lesson.
func TestBundledEntriesAreExplainable(t *testing.T) {
	for id, e := range Defaults().Entries {
		if e.Successor == "" {
			t.Errorf("entry %s has no successor; a refusal must say what to run instead", id)
		}
		if e.Why == "" {
			t.Errorf("entry %s has no why; a refusal must explain itself in plain language", id)
		}
		if e.ID != id {
			t.Errorf("entry keyed %s carries ID %q; the map key is the id", id, e.ID)
		}
	}
}

// TestIncidentCaptureIsKnownGoodFixtureOne pins the corpus case the whole
// design turns on: this repo's own incident-capture command, whose quoted text
// names a hazard, must be the first known-good fixture of the rm entry and must
// not fire. A raw-regex guard would have blocked it.
func TestIncidentCaptureIsKnownGoodFixtureOne(t *testing.T) {
	e, ok := Defaults().Entries["rm-rf-after-cd-chain"]
	if !ok {
		t.Fatal("rm-rf-after-cd-chain is missing from the bundled registry")
	}
	if len(e.Fixtures.KnownGood) == 0 {
		t.Fatal("rm-rf-after-cd-chain ships no known-good fixtures")
	}
	first := e.Fixtures.KnownGood[0]
	if !containsAll(first, "capture", "cd scratch && rm -rf *") {
		t.Fatalf("known-good fixture #1 = %q, want the incident-capture shape", first)
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
