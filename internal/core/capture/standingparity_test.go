package capture

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/core/lint"
)

// "Which answer is in force" is asked by two packages with two readers, because
// core/lint cannot import this one (this package's own tests import lint, so the
// edge back would be a cycle). Two readers are tolerable; two ANSWERS are not —
// the verb would refuse a second disposition the board says is not needed, or
// promote an item the board still shows as held.
//
// So the two are held to one answer over the same fixtures. This is the test the
// review asked for, and the malformed-sibling row is the one that found a real
// divergence: capture's strict parser skipped such a file before reading its
// supersession edge, while lint's line scanner read the edge regardless — so a
// single malformed record made the board say "answered" and the verb say "not".
func TestStandingDispositionAgreesAcrossBothReaders(t *testing.T) {
	type file struct {
		name, body string
	}
	cases := []struct {
		name  string
		files []file
		want  string // the standing dsp id, or "" for none
	}{
		{
			name:  "a single answer stands",
			files: []file{{"dsp-1.md", disp("dsp-1", "accepted", "")}},
			want:  "dsp-1",
		},
		{
			name: "the head of a superseded chain stands",
			files: []file{
				{"dsp-1.md", disp("dsp-1", "held", "")},
				{"dsp-2.md", disp("dsp-2", "held", "dsp-1")},
				{"dsp-3.md", disp("dsp-3", "accepted", "dsp-2")},
			},
			want: "dsp-3",
		},
		{
			// A record neither reader can read cannot be trusted to retire another,
			// and it cannot vanish either: it stays present, so it stands, and
			// whoever looks is told there is something here to deal with.
			name: "a malformed sibling retires nothing",
			files: []file{
				{"dsp-1.md", disp("dsp-1", "held", "")},
				{"dsp-2.md", "---\nschema_version: 1\nid: \"dsp-2\"\nid: \"dsp-2\"\nitem: \"rdi-9\"\nstate: \"accepted\"\nsupersedes_disposition: \"dsp-1\"\n---\n\n"},
			},
			want: "", // more than one standing: neither reader may pick a winner
		},
		{
			name: "a file that is not a disposition is not an answer",
			files: []file{
				{"dsp-1.md", disp("dsp-1", "accepted", "")},
				{"README.md", "# notes\n"},
				{"notes.md", "supersedes_disposition: dsp-1\n"},
			},
			want: "dsp-1",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := t.TempDir()
			issuesRoot := filepath.Join(repo, LedgerRelPath)
			itemDir := filepath.Join(issuesRoot, issueschema.DispositionsDir, "rdi-9")
			if err := os.MkdirAll(itemDir, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, f := range c.files {
				if err := os.WriteFile(filepath.Join(itemDir, f.name), []byte(f.body), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			mine, err := standingDispositions(itemDir)
			if err != nil {
				t.Fatalf("capture.standingDispositions: %v", err)
			}
			theirs, err := lint.StandingDispositions(repo, LedgerRelPath, "rdi-9")
			if err != nil {
				t.Fatalf("lint.StandingDispositions: %v", err)
			}

			got := one(mine)
			if other := one(theirs); got != other {
				t.Fatalf("the two readers disagree: capture says %q (%v), lint says %q (%v)",
					got, mine, other, theirs)
			}
			if got != c.want {
				t.Fatalf("standing = %q (%v), want %q", got, mine, c.want)
			}
		})
	}
}

// one renders a standing set as the single answer in force, or "" when there is
// none or more than one — the two states in which no reader may pick a winner.
func one(ids []string) string {
	if len(ids) != 1 {
		return ""
	}
	return ids[0]
}

// disp builds a well-formed disposition record.
func disp(id, state, supersedes string) string {
	s := "---\nschema_version: 1\nid: \"" + id + "\"\nitem: \"rdi-9\"\nstate: \"" + state + "\"\n"
	if state == issueschema.DispositionHeld {
		s += "exit_condition: \"the closing run returns it again\"\n"
	} else {
		s += "disposition_grounds: \"worth acting on\"\n"
	}
	if supersedes != "" {
		s += "supersedes_disposition: \"" + supersedes + "\"\n"
	}
	return s + "---\n\n"
}
