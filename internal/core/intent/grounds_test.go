package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/grounds"
)

// mustGrounds builds a validated Grounds for a test, where the vocabulary and
// the floor are not what is under test.
func mustGrounds(t *testing.T, tok grounds.Token, text string) grounds.Grounds {
	t.Helper()
	g, err := grounds.New(tok, text)
	if err != nil {
		t.Fatalf("grounds.New(%q, %q): %v", tok, text, err)
	}
	return g
}

// readIntent reads a record back off disk after a write.
func readIntent(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("reading %s: %v", rel, err)
	}
	return string(data)
}

// TestRecordGroundsWritesEntry: the entry lands under the record's existing
// `## Grounds` heading as one top-level bullet in the shared grammar, so the
// record reads as prose and the gate reads it as data with no second parser.
func TestRecordGroundsWritesEntry(t *testing.T) {
	root := t.TempDir()
	const rel = plannedDir + "/itd-10-alpha.md"
	writeFile(t, root, rel, plannedUnlinked("itd-10", "alpha")+"\n## Grounds\n\n- pursued: an earlier conjecture already recorded here\n")

	g := mustGrounds(t, grounds.Pursued, "we expect a stamped identity to survive rewording")
	res, err := RecordGrounds(root, "itd-10", g)
	if err != nil {
		t.Fatalf("RecordGrounds: %v", err)
	}
	if res.IntentID != "itd-10" || res.Path != rel {
		t.Fatalf("result = %+v, want itd-10 at %s", res, rel)
	}
	body := readIntent(t, root, rel)
	if !strings.Contains(body, g.Bullet()) {
		t.Fatalf("record does not carry %q:\n%s", g.Bullet(), body)
	}
	if got := ParseGrounds(body); len(got) != 2 {
		t.Fatalf("ParseGrounds read %d entries, want 2:\n%s", len(got), body)
	}
	if res.Entries != 2 {
		t.Fatalf("result Entries = %d, want 2", res.Entries)
	}
}

// TestRecordGroundsCreatesSectionWhenAbsent: a record that has never carried
// grounds gains the section rather than losing the entry.
func TestRecordGroundsCreatesSectionWhenAbsent(t *testing.T) {
	root := t.TempDir()
	const rel = plannedDir + "/itd-10-alpha.md"
	writeFile(t, root, rel, plannedUnlinked("itd-10", "alpha"))

	g := mustGrounds(t, grounds.Pursued, "we expect the gate to make the reasoning survive the session")
	if _, err := RecordGrounds(root, "itd-10", g); err != nil {
		t.Fatalf("RecordGrounds: %v", err)
	}
	body := readIntent(t, root, rel)
	if !strings.Contains(body, "## "+GroundsHeading) {
		t.Fatalf("the section was not created:\n%s", body)
	}
	got := ParseGrounds(body)
	if len(got) != 1 || got[0] != g {
		t.Fatalf("ParseGrounds = %+v, want exactly %+v", got, g)
	}
}

// TestRecordGroundsAppendsSecondEntry: recording is append-only. A second gate
// decision on one record adds an entry; it never rewrites the first, because the
// earlier conjecture is what a later reader checks the outcome against.
func TestRecordGroundsAppendsSecondEntry(t *testing.T) {
	root := t.TempDir()
	const rel = plannedDir + "/itd-10-alpha.md"
	writeFile(t, root, rel, plannedUnlinked("itd-10", "alpha"))

	first := mustGrounds(t, grounds.Deferred, "deferred while the identity it keys on has no consumer")
	second := mustGrounds(t, grounds.Pursued, "pursued now that the disposition finally has something to key on")
	if _, err := RecordGrounds(root, "itd-10", first); err != nil {
		t.Fatalf("first RecordGrounds: %v", err)
	}
	res, err := RecordGrounds(root, "itd-10", second)
	if err != nil {
		t.Fatalf("second RecordGrounds: %v", err)
	}
	if res.Entries != 2 {
		t.Fatalf("Entries = %d, want 2", res.Entries)
	}
	got := ParseGrounds(readIntent(t, root, rel))
	if len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("ParseGrounds = %+v, want [%+v %+v] in that order", got, first, second)
	}
}

// TestRecordGroundsRedactsText: a grounds text is operator prose landing in a
// committed record, so it goes through the same detector every other durable
// free text does — before the write, never after. The spans below are FAKE
// shapes, matched only by shape, never real credentials.
func TestRecordGroundsRedactsText(t *testing.T) {
	root := t.TempDir()
	const rel = plannedDir + "/itd-10-alpha.md"
	writeFile(t, root, rel, plannedUnlinked("itd-10", "alpha"))

	const fakeHome = "/Users/alice/.ssh/id_rsa"
	g := grounds.Grounds{Token: grounds.Pursued,
		Text: "we expect the receipt at " + fakeHome + " to be the thing that proves it"}
	res, err := RecordGrounds(root, "itd-10", g)
	if err != nil {
		t.Fatalf("RecordGrounds: %v", err)
	}
	body := readIntent(t, root, rel)
	if strings.Contains(body, "/Users/alice") {
		t.Fatalf("the record persisted the raw home path verbatim:\n%s", body)
	}
	if res.Redacted == 0 {
		t.Fatal("Redacted = 0; a rewritten text must be reported, never silently altered")
	}
}

// TestRecordGroundsRefusesUnknownIntent: an id naming no record is a structural
// fault, and nothing is written anywhere.
func TestRecordGroundsRefusesUnknownIntent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md", plannedUnlinked("itd-10", "alpha"))

	g := mustGrounds(t, grounds.Pursued, "we expect an unknown id to be refused before any write")
	if _, err := RecordGrounds(root, "itd-999", g); err == nil {
		t.Fatal("RecordGrounds on an unknown intent = nil error, want a refusal")
	}
	if _, err := RecordGrounds(root, "nonsense", g); err == nil {
		t.Fatal("RecordGrounds on a malformed id = nil error, want a refusal")
	}
	if body := readIntent(t, root, plannedDir+"/itd-10-alpha.md"); strings.Contains(body, "## "+GroundsHeading) {
		t.Fatalf("a refused call wrote to an unrelated record:\n%s", body)
	}
}

// TestRecordGroundsRefusesDegenerateTextAfterRedaction: validation runs on the
// text that will actually be written. Redacting after validation would let a
// rewritten span reach a field the validator had already passed.
func TestRecordGroundsRefusesDegenerateTextAfterRedaction(t *testing.T) {
	root := t.TempDir()
	const rel = plannedDir + "/itd-10-alpha.md"
	writeFile(t, root, rel, plannedUnlinked("itd-10", "alpha"))

	if _, err := RecordGrounds(root, "itd-10", grounds.Grounds{Token: grounds.Pursued, Text: "because"}); err == nil {
		t.Fatal("a text below the floor = nil error, want a refusal")
	}
	if _, err := RecordGrounds(root, "itd-10", grounds.Grounds{Token: "planned", Text: "a perfectly good conjecture about identity"}); err == nil {
		t.Fatal("an out-of-vocabulary token = nil error, want a refusal")
	}
	if body := readIntent(t, root, rel); strings.Contains(body, "## "+GroundsHeading) {
		t.Fatalf("a refused call still wrote the section:\n%s", body)
	}
}

// TestParseGroundsIgnoresFencedAndCommentedBullets: an example of the grammar in
// a fenced block, or a bullet parked in an HTML comment, is not a recorded
// ground — the same rule the scope-conditions reader already holds.
func TestParseGroundsIgnoresFencedAndCommentedBullets(t *testing.T) {
	content := plannedUnlinked("itd-10", "alpha") + "\n## Grounds\n\n" +
		"```\n- pursued: this is an example of the grammar, not a recorded ground\n```\n\n" +
		"<!--\n- deferred: this one was parked and is not live either\n-->\n\n" +
		"- pursued: the only live entry on this record names a real conjecture\n"
	got := ParseGrounds(content)
	if len(got) != 1 {
		t.Fatalf("ParseGrounds read %d entries, want 1: %+v", len(got), got)
	}
	if !strings.HasPrefix(got[0].Text, "the only live entry") {
		t.Fatalf("ParseGrounds read the wrong bullet: %+v", got)
	}
}
