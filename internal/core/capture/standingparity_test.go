package capture

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/core/lint"
)

// "Which answer is in force" is asked by two packages, because core/lint cannot
// import this one (this package's own tests import lint, so the edge back would
// be a cycle). Two WALKS are unavoidable; two ANSWERS were not — the verb would
// refuse a second disposition the board says is not needed, or promote an item
// the board still shows as held.
//
// The judgement now lives in issueschema, which both packages call, and this
// table is what holds the two walks to it. Each row was a real divergence before
// it was a row: the malformed-sibling case (capture's strict parser skipped the
// file before reading its supersession edge while lint's line scanner read the
// edge regardless), then the comment-led and blank-line-led cases (the same
// split, one preamble earlier). Every one made the board say "answered" and the
// verb say "not".
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
			// A comment before the opening delimiter is the shape a hand-edited
			// record most easily acquires, and it parted the two readers: the
			// lenient line scanner read straight past it to the supersession edge
			// while the strict parser refused the file outright.
			name: "a comment-led sibling retires nothing",
			files: []file{
				{"dsp-1.md", disp("dsp-1", "held", "")},
				{"dsp-2.md", "<!-- written by hand -->\n" + disp("dsp-2", "accepted", "dsp-1")},
			},
			want: "",
		},
		{
			name: "a blank-line-led sibling retires nothing",
			files: []file{
				{"dsp-1.md", disp("dsp-1", "held", "")},
				{"dsp-2.md", "\n" + disp("dsp-2", "accepted", "dsp-1")},
			},
			want: "",
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
			got := one(mine)
			if got != c.want {
				t.Fatalf("standing = %q (%v), want %q", got, mine, c.want)
			}

			// lint walks the same directory for its own report, and the OBSERVABLE
			// consequence must follow the SAME STANDING SET — not merely the same
			// verdict about whether exactly one record stands.
			//
			// That distinction is the whole assertion. An earlier version compared
			// against `got`, which is "" whenever the set is not a single record;
			// on precisely the three rows this table exists for, two records stand,
			// so the expectation collapsed to "no hold" and a lint walk standing
			// only the NEWEST record also reported no hold. All six rows passed
			// under a mutation that reintroduced the divergence.
			//
			// So the expectation is built from mine[0] — the first standing id,
			// which is the one lint's report renders — and it is non-trivial on
			// every row.
			writeReadingFor(t, repo, "rdi-9")
			report, err := lint.ReadReadingOutstanding(repo, LedgerRelPath)
			if err != nil {
				t.Fatalf("lint.ReadReadingOutstanding: %v", err)
			}
			wantOutstanding := len(mine) == 0
			if gotOutstanding := len(report.Undispositioned) == 1; gotOutstanding != wantOutstanding {
				t.Fatalf("lint reports outstanding=%v, capture stands %v — the two walks disagree",
					gotOutstanding, mine)
			}
			wantHeld, wantHeldID := false, ""
			if len(mine) > 0 {
				wantHeldID = mine[0]
				wantHeld = stateOf(t, itemDir, wantHeldID) == issueschema.DispositionHeld
			}
			gotHeld := len(report.OpenHolds) == 1 && report.OpenHolds[0].Disposition == wantHeldID
			if gotHeld != wantHeld {
				var holds []string
				for _, h := range report.OpenHolds {
					holds = append(holds, h.Disposition)
				}
				t.Fatalf("lint reports holds %v, want a hold on %q = %v — capture stands %v, and the two walks must render the same standing set",
					holds, wantHeldID, wantHeld, mine)
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

// writeReadingFor lays down the reading record the item answers, so lint's walk
// has something to report on: its report is keyed on reading items, not on
// disposition directories.
func writeReadingFor(t *testing.T, repo, item string) {
	t.Helper()
	dir := filepath.Join(repo, LedgerRelPath, issueschema.ReadingsDir, "rdg-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nschema_version: 1\nid: \"" + item + "\"\nrun: \"rdg-1\"\n" +
		"manifest: \"sha256:beef\"\nposition: \"detection\"\nregime: \"registrative\"\n" +
		"pattern: \"a stated constraint\"\ntension: \"t\"\nconstraint_in_play: \"c\"\nwhy_a_tension: \"w\"\n---\n\n"
	if err := os.WriteFile(filepath.Join(dir, item+".md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// stateOf reads one disposition's state through the shared reader.
func stateOf(t *testing.T, itemDir, id string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(itemDir, id+".md"))
	if err != nil {
		t.Fatal(err)
	}
	return issueschema.ParseDisposition(id, string(content)).State
}
