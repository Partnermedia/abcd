package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/frontmatter"
)

// TestZWNBSPBodyLineDoesNotDesyncReaderAndWriter (iss-2608270926036966) pins the
// reader and the writer of an intent record to ONE frontmatter boundary.
//
// U+FEFF is a byte-order mark only at file position 0; mid-file it is ZERO WIDTH
// NO-BREAK SPACE, an ordinary character. When the shared delimiter predicate
// trimmed it on every line, frontmatter.Fields — the reader behind intent.Load
// and record-lint — closed the block at a body line spelled "\ufeff---", while
// setFrontmatterFields (trimming only " \t\r") read on to the next bare `---` and
// inserted spec_id THERE, in the body. The write reported success, the record
// stayed lint-green, and the value was invisible on reload.
//
// The assertion is the agreement, not either half alone: the key the writer sets
// must be the key the loader reads back.
func TestZWNBSPBodyLineDoesNotDesyncReaderAndWriter(t *testing.T) {
	const zwnbsp = "\ufeff"
	// A record whose frontmatter is closed by a "\ufeff---" line, with a
	// genuine thematic break further down the body. To the buggy reader the block
	// ends at the first; to the writer it ends at the second.
	content := "---\n" +
		"id: itd-9\n" +
		"kind: intent\n" +
		"title: Sample intent\n" +
		"status: draft\n" +
		zwnbsp + "---\n" +
		"\n" +
		"# Sample intent\n" +
		"\n" +
		"---\n" +
		"\n" +
		"Trailing prose.\n"

	// Reader and writer must agree on where the block ends.
	fields := frontmatter.Fields(strings.Split(content, "\n"))
	if _, ok := fields["status"]; !ok {
		t.Fatalf("reader lost a frontmatter key: %v", fields)
	}

	updated, err := setFrontmatterFields(content, map[string]string{"spec_id": "spc-4"})
	if err != nil {
		t.Fatalf("setFrontmatterFields: %v", err)
	}

	// The inserted key must land INSIDE the leading block — i.e. the reader must
	// see it. A writer that inserts past the reader's close writes into the body.
	after := frontmatter.Fields(strings.Split(updated, "\n"))
	if got := after["spec_id"].Value; got != "spc-4" {
		t.Fatalf("spec_id written by setFrontmatterFields reads back as %q, want %q — writer inserted past the reader's block close", got, "spc-4")
	}
	// Nothing pre-existing may have been truncated out of the reader's view
	// either: the writer inserting at ITS close while the reader stops at an
	// earlier one is the same desync seen from the other end.
	for key, want := range map[string]string{
		"id": "itd-9", "kind": "intent", "title": "Sample intent", "status": "draft",
	} {
		if got := after[key].Value; got != want {
			t.Fatalf("after the write, %s = %q, want %q", key, got, want)
		}
	}
	// The ZWNBSP line itself is body content and must survive untouched.
	if !strings.Contains(updated, zwnbsp+"---") {
		t.Fatal("the ZWNBSP line was rewritten or dropped; it is ordinary content")
	}
}

// TestZWNBSPBodyLineSurvivesLoad is the same invariant through the public loader:
// a record whose body carries a "\ufeff---" line loads with its frontmatter
// fields intact.
func TestZWNBSPBodyLineSurvivesLoad(t *testing.T) {
	root := t.TempDir()
	draftsDir := filepath.Join(root, IntentsRelDir, BucketDrafts)
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\n" +
		"id: itd-9\n" +
		"kind: intent\n" +
		"title: Sample intent\n" +
		"status: draft\n" +
		"\ufeff---\n" + // a stray ZWNBSP line INSIDE the block
		"spec_id: spc-4\n" +
		"---\n" +
		"\n" +
		"# Sample intent\n" +
		"\n" +
		"Trailing prose.\n"
	if err := os.WriteFile(filepath.Join(draftsDir, "itd-9-sample-intent.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	var found bool
	for _, it := range c.Intents {
		if it.ID != "itd-9" {
			continue
		}
		found = true
		if it.SpecID != "spc-4" {
			t.Fatalf("SpecID = %q, want %q — a body ZWNBSP line truncated the block", it.SpecID, "spc-4")
		}
	}
	if !found {
		t.Fatalf("itd-9 not loaded from %v", c.Intents)
	}
}
