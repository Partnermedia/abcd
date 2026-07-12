package spec

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	specsOpen   = ".abcd/development/specs/open"
	specsClosed = ".abcd/development/specs/closed"
	intentsBase = ".abcd/development/intents"
)

func TestNextIDEmptyRepo(t *testing.T) {
	root := t.TempDir()
	got, err := NextID(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != "spc-1" {
		t.Fatalf("NextID(empty) = %q, want spc-1", got)
	}
}

// itd-3 is shipped with spec_id: spc-1 but has no spec-store file. NextID must
// still skip spc-1 so a freshly minted spec never collides with that reservation.
func TestNextIDReservedByIntent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, intentsBase+"/shipped/itd-3-rules-loader.md",
		"---\nid: itd-3\nslug: rules-loader\nspec_id: spc-1\nkind: standalone\n---\n# ok\n")

	got, err := NextID(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != "spc-2" {
		t.Fatalf("NextID = %q, want spc-2 (spc-1 reserved by itd-3)", got)
	}
}

func TestNextIDMaxAcrossSpecsAndIntents(t *testing.T) {
	root := t.TempDir()
	// A spec-store file at spc-5 (higher than any intent reservation).
	writeFile(t, root, specsOpen+"/spc-5-existing.md",
		"---\nid: spc-5\nslug: existing\nintent: itd-9\n---\n# ok\n")
	// An intent reserving spc-2 via the tolerated spc-N-<slug> form (lower; must
	// not lower the max). record-lint accepts this shape, so the store must too.
	writeFile(t, root, intentsBase+"/planned/itd-20-x.md",
		"---\nid: itd-20\nslug: x\nspec_id: spc-2-thing\nkind: standalone\n---\n# ok\n")

	got, err := NextID(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != "spc-6" {
		t.Fatalf("NextID = %q, want spc-6", got)
	}
}

func TestCreateRoundTrip(t *testing.T) {
	root := t.TempDir()
	sp, err := Create(root, "itd-9", "my-feature")
	if err != nil {
		t.Fatal(err)
	}
	if sp.ID != "spc-1" || sp.Intent != "itd-9" || sp.Status != StatusOpen {
		t.Fatalf("Create returned %+v", sp)
	}

	abs := filepath.Join(root, specsOpen, "spc-1-my-feature.md")
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("expected spec file on disk: %v", err)
	}
	if !strings.Contains(string(data), "intent: itd-9") {
		t.Fatalf("spec file missing intent link:\n%s", data)
	}

	// Round-trips through Load.
	store, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := store.Lookup("spc-1"); !ok || s.Intent != "itd-9" {
		t.Fatalf("Load/Lookup after Create = %+v, %v", s, ok)
	}
	if s, ok := store.ByIntent("itd-9"); !ok || s.ID != "spc-1" {
		t.Fatalf("Load/ByIntent after Create = %+v, %v", s, ok)
	}
}

func TestCreateRejectsBadIntent(t *testing.T) {
	root := t.TempDir()
	if _, err := Create(root, "itd-../../etc", "slug"); err == nil {
		t.Fatal("Create with traversal intent id must fail")
	}
	if _, err := Create(root, "spc-1", "slug"); err == nil {
		t.Fatal("Create with non-itd intent id must fail")
	}
}

func TestCreateRejectsBadSlug(t *testing.T) {
	root := t.TempDir()
	if _, err := Create(root, "itd-9", "../../etc"); err == nil {
		t.Fatal("Create with traversal slug must fail")
	}
	if _, err := Create(root, "itd-9", "Bad Slug"); err == nil {
		t.Fatal("Create with non-kebab slug must fail")
	}
}

func TestLoadMissingDirIsEmpty(t *testing.T) {
	root := t.TempDir()
	store, err := Load(root)
	if err != nil {
		t.Fatalf("Load on missing specs dir must be soft: %v", err)
	}
	if len(store.Specs) != 0 {
		t.Fatalf("expected empty store, got %+v", store.Specs)
	}
}

func TestLoadMalformedIsHardError(t *testing.T) {
	root := t.TempDir()
	// No frontmatter at all -> no id -> hard error.
	writeFile(t, root, specsOpen+"/spc-1-broken.md", "# just a title, no frontmatter\n")
	if _, err := Load(root); err == nil {
		t.Fatal("Load must hard-error on a malformed spec file")
	}
}

func TestLoadRejectsTraversalID(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, specsOpen+"/spc-1-evil.md",
		"---\nid: spc-../../etc\nslug: evil\nintent: itd-9\n---\n# evil\n")
	if _, err := Load(root); err == nil {
		t.Fatal("Load must reject a path-traversal id in frontmatter")
	}
}

func TestCloseMovesOpenToClosed(t *testing.T) {
	root := t.TempDir()
	if _, err := Create(root, "itd-9", "my-feature"); err != nil {
		t.Fatal(err)
	}

	sp, err := Close(root, "spc-1")
	if err != nil {
		t.Fatal(err)
	}
	if sp.Status != StatusClosed || sp.Intent != "itd-9" {
		t.Fatalf("Close returned %+v", sp)
	}

	if _, err := os.Stat(filepath.Join(root, specsOpen, "spc-1-my-feature.md")); !os.IsNotExist(err) {
		t.Fatal("open file should be gone after Close")
	}
	if _, err := os.Stat(filepath.Join(root, specsClosed, "spc-1-my-feature.md")); err != nil {
		t.Fatalf("closed file should exist after Close: %v", err)
	}

	// The store now reports it closed.
	store, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := store.Lookup("spc-1"); !ok || s.Status != StatusClosed {
		t.Fatalf("after Close, Lookup = %+v, %v", s, ok)
	}
}

// TestCloseRefusesWhenClosedTargetExists proves Close fails closed rather than
// clobbering a same-name spec already sitting in closed/.
func TestCloseRefusesWhenClosedTargetExists(t *testing.T) {
	root := t.TempDir()
	if _, err := Create(root, "itd-9", "my-feature"); err != nil {
		t.Fatal(err)
	}
	// A same-name spec already occupies closed/.
	writeFile(t, root, specsClosed+"/spc-1-my-feature.md",
		"---\nid: spc-1\nslug: my-feature\nintent: itd-9\n---\n# pre-existing\n")

	if _, err := Close(root, "spc-1"); err == nil {
		t.Fatal("Close must refuse to overwrite an existing closed target")
	}
	// The open file is untouched (still there), the closed one not clobbered.
	if _, err := os.Stat(filepath.Join(root, specsOpen, "spc-1-my-feature.md")); err != nil {
		t.Fatalf("open file must remain after refusal: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(root, specsClosed, "spc-1-my-feature.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "pre-existing") {
		t.Fatalf("closed target was clobbered:\n%s", body)
	}
}

func TestCloseMissingFails(t *testing.T) {
	root := t.TempDir()
	if _, err := Close(root, "spc-99"); err == nil {
		t.Fatal("Close on a missing spec must fail")
	}
}

func TestCloseAlreadyClosedFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, specsClosed+"/spc-1-done.md",
		"---\nid: spc-1\nslug: done\nintent: itd-9\n---\n# done\n")
	if _, err := Close(root, "spc-1"); err == nil {
		t.Fatal("Close on an already-closed spec must fail")
	}
}

// TestNextIDRejectsUnreservableSpecID (iss-68 P5) proves a non-null spec_id with
// no parseable reservation number ("spc-oops") fails NextID closed rather than
// being silently dropped from the reservation scan (which could hand out a
// colliding id). A well-formed "spc-N" / "spc-N-<slug>" is accepted (see
// TestNextIDMaxAcrossSpecsAndIntents); only a numberless one is rejected.
func TestNextIDRejectsUnreservableSpecID(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, intentsBase+"/planned/itd-20-x.md",
		"---\nid: itd-20\nslug: x\nspec_id: spc-oops\nkind: standalone\n---\n# ok\n")
	if _, err := NextID(root); err == nil {
		t.Fatal("NextID must fail closed on a spec_id with no reservable number, not silently drop it")
	}
}

// TestLoadRejectsFifoSpecFile (iss-68 P7) proves a FIFO at a spec path is rejected
// promptly, not hung on. The read opens with O_NOFOLLOW|O_NONBLOCK and validates
// the fd, so a FIFO returns a not-regular error instead of blocking os.ReadFile.
func TestLoadRejectsFifoSpecFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, specsOpen), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(root, specsOpen, "spc-1-x.md"), 0o644); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := Load(root)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("a FIFO spec file must be refused, not read")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Load hung on a FIFO spec file (open must not block)")
	}
}
