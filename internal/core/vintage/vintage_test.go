package vintage

import (
	"runtime/debug"
	"testing"
)

// stubProvider answers Expected from a fixed value, so the comparator's outcomes
// can be exercised without a disk or a network reference.
type stubProvider struct{ exp Expected }

func (s stubProvider) Expected() Expected { return s.exp }

func TestCompareOutcomes(t *testing.T) {
	cases := []struct {
		name string
		cur  Current
		exp  Expected
		want Outcome
	}{
		{
			name: "equal revisions are fresh",
			cur:  Current{Revision: "abc123", Known: true},
			exp:  Expected{Revision: "abc123", Known: true},
			want: Fresh,
		},
		{
			name: "differing revisions are stale",
			cur:  Current{Revision: "abc123", Known: true},
			exp:  Expected{Revision: "def456", Known: true},
			want: Stale,
		},
		{
			name: "unknown current is never fresh even against a matching reference",
			cur:  Current{Revision: "abc123", Known: false},
			exp:  Expected{Revision: "abc123", Known: true},
			want: Unknown,
		},
		{
			name: "unresolved reference is unknown, not fresh",
			cur:  Current{Revision: "abc123", Known: true},
			exp:  Expected{Known: false},
			want: Unknown,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Compare(c.cur, stubProvider{c.exp})
			if got.Outcome != c.want {
				t.Fatalf("Compare outcome = %q, want %q", got.Outcome, c.want)
			}
		})
	}
}

func TestCompareReportCarriesValues(t *testing.T) {
	rep := Compare(Current{Revision: "aaa", Known: true}, stubProvider{Expected{Revision: "bbb", Known: true}})
	if rep.Current != "aaa" || rep.Expected != "bbb" {
		t.Fatalf("report = %+v, want current=aaa expected=bbb", rep)
	}
}

// TestCurrentFromSettings exercises all four unknown triggers plus the fresh
// case, over synthetic build settings so no rebuild of the test binary is needed.
func TestCurrentFromSettings(t *testing.T) {
	rev := "1111111111111111111111111111111111111111"
	cases := []struct {
		name      string
		ok        bool
		settings  []debug.BuildSetting
		wantKnown bool
		wantRev   string
	}{
		{
			name:      "no build info at all is unknown",
			ok:        false,
			wantKnown: false,
		},
		{
			name:      "no vcs.revision is unknown (go run / -buildvcs=false / .git-less / go install)",
			ok:        true,
			settings:  []debug.BuildSetting{{Key: "GOOS", Value: "darwin"}},
			wantKnown: false,
		},
		{
			name:      "dirty rebuild (vcs.modified=true) is unknown",
			ok:        true,
			settings:  []debug.BuildSetting{{Key: "vcs.revision", Value: rev}, {Key: "vcs.modified", Value: "true"}},
			wantKnown: false,
			wantRev:   rev,
		},
		{
			name:      "clean stamped revision is known",
			ok:        true,
			settings:  []debug.BuildSetting{{Key: "vcs.revision", Value: rev}, {Key: "vcs.modified", Value: "false"}},
			wantKnown: true,
			wantRev:   rev,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := currentFromSettings(c.ok, c.settings)
			if got.Known != c.wantKnown {
				t.Fatalf("Known = %v, want %v", got.Known, c.wantKnown)
			}
			if got.Revision != c.wantRev {
				t.Fatalf("Revision = %q, want %q", got.Revision, c.wantRev)
			}
		})
	}
}
