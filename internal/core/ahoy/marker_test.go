package ahoy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkerInsertIntoAbsentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	wrote, ok := installMarkerFile(path)
	if !ok || !wrote {
		t.Fatalf("install into absent file: wrote=%v ok=%v", wrote, ok)
	}
	if classifyMarker(path) != markerCurrent {
		t.Errorf("state after install = %q, want current", classifyMarker(path))
	}
}

func TestMarkerInsertAfterFrontmatterAndH1(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	original := "---\ntitle: x\n---\n# Heading\n\nbody text\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	wrote, ok := installMarkerFile(path)
	if !ok || !wrote {
		t.Fatalf("install: wrote=%v ok=%v", wrote, ok)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Block must land after the frontmatter close, before the heading, and
	// preserve the body.
	if !bytes.Contains(got, []byte("body text")) {
		t.Errorf("body text lost:\n%s", got)
	}
	fmClose := bytes.Index(got, []byte("---\n\n")) // frontmatter's closing fence
	blockAt := bytes.Index(got, markerBegin)
	headingAt := bytes.Index(got, []byte("# Heading"))
	if blockAt == -1 || fmClose == -1 || headingAt == -1 {
		t.Fatalf("missing landmark (block=%d fmClose=%d heading=%d):\n%s", blockAt, fmClose, headingAt, got)
	}
	if !(fmClose < blockAt && blockAt < headingAt) {
		t.Errorf("block not placed between frontmatter and heading:\n%s", got)
	}
	if classifyMarker(path) != markerCurrent {
		t.Errorf("state = %q, want current", classifyMarker(path))
	}
}

func TestMarkerInsertSkipsFencedH1(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	// A '# ' shell comment inside a fenced snippet precedes the real H1. The
	// block must not split the fence; it belongs after the real heading.
	original := "```bash\n# install deps\nmake build\n```\n# Real Title\n\nbody text\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := installMarkerFile(path); !ok {
		t.Fatal("install failed")
	}
	got, _ := os.ReadFile(path)
	// The fenced snippet must stay contiguous (block did not land inside it).
	if !bytes.Contains(got, []byte("# install deps\nmake build")) {
		t.Errorf("marker split the fenced snippet:\n%s", got)
	}
	blockAt := bytes.Index(got, markerBegin)
	titleAt := bytes.Index(got, []byte("# Real Title"))
	fenceCommentAt := bytes.Index(got, []byte("# install deps"))
	if blockAt == -1 || titleAt == -1 {
		t.Fatalf("missing landmark (block=%d title=%d):\n%s", blockAt, titleAt, got)
	}
	if !(fenceCommentAt < blockAt && titleAt < blockAt) {
		t.Errorf("block not placed after the real H1:\n%s", got)
	}
	if classifyMarker(path) != markerCurrent {
		t.Errorf("state = %q, want current", classifyMarker(path))
	}
}

func TestClassifySymlinkedMarkerIsNotResolvableGap(t *testing.T) {
	dir := t.TempDir()
	// docs.target=claude_md so detection checks only the symlinked CLAUDE.md.
	if err := os.MkdirAll(filepath.Join(dir, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".abcd", "config.json"),
		[]byte(`{"docs":{"target":"claude_md"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlinked CLAUDE.md whose target lacks the block: classifyMarker must
	// report it as a symlink (non-resolvable), not "missing", so detection does
	// not emit a resolvable gap that install can never close.
	real := filepath.Join(t.TempDir(), "real.md")
	if err := os.WriteFile(real, []byte("# Title\n\nno block here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "CLAUDE.md")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if got := classifyMarker(link); got != markerSymlink {
		t.Fatalf("classifyMarker on symlink = %q, want %q", got, markerSymlink)
	}
	// install refuses to write through the symlink, so the two must agree: no
	// resolvable gap paired with a silent no-op.
	if wrote, ok := installMarkerFile(link); wrote || ok {
		t.Fatalf("installMarkerFile through symlink: wrote=%v ok=%v, want false/false", wrote, ok)
	}
	// detectMarkerDrift must not emit an actionable (required+resolvable) gap.
	for _, g := range detectMarkerDrift(dir) {
		if g.Required && g.Resolvable {
			t.Errorf("symlinked marker produced an actionable gap: %+v", g)
		}
	}
}

func TestMarkerInstallIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(path, []byte("# Title\n\nprose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := installMarkerFile(path); !ok {
		t.Fatal("first install failed")
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	wrote, ok := installMarkerFile(path)
	if !ok {
		t.Fatal("second install failed")
	}
	if wrote {
		t.Errorf("second install rewrote a current block (not byte-stable)")
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("marker install not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestMarkerOutdatedBlockIsRewritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	stale := "# Title\n\n<!-- BEGIN ABCD -->\nOLD CONTENT\n<!-- END ABCD -->\n\nafter\n"
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	if classifyMarker(path) != markerOutdated {
		t.Fatalf("precondition: expected outdated, got %q", classifyMarker(path))
	}
	wrote, ok := installMarkerFile(path)
	if !ok || !wrote {
		t.Fatalf("rewrite: wrote=%v ok=%v", wrote, ok)
	}
	if classifyMarker(path) != markerCurrent {
		t.Errorf("state after rewrite = %q, want current", classifyMarker(path))
	}
	got, _ := os.ReadFile(path)
	if bytes.Contains(got, []byte("OLD CONTENT")) {
		t.Errorf("stale content survived rewrite:\n%s", got)
	}
	if !bytes.Contains(got, []byte("after")) {
		t.Errorf("trailing content lost:\n%s", got)
	}
}

func TestMarkerMultiBlockCollapsesToOne(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	dup := "# T\n\n<!-- BEGIN ABCD -->\na\n<!-- END ABCD -->\n\nmid\n\n<!-- BEGIN ABCD -->\nb\n<!-- END ABCD -->\n"
	if err := os.WriteFile(path, []byte(dup), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := installMarkerFile(path); !ok {
		t.Fatal("install failed")
	}
	got, _ := os.ReadFile(path)
	if n := bytes.Count(got, markerBegin); n != 1 {
		t.Errorf("expected exactly one block after collapse, got %d:\n%s", n, got)
	}
	if !bytes.Contains(got, []byte("mid")) {
		t.Errorf("inter-block content lost:\n%s", got)
	}
}

func TestMarkerRemoveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	original := "# Title\n\nprose here\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := installMarkerFile(path); !ok {
		t.Fatal("install failed")
	}
	if _, ok := removeMarkerFile(path); !ok {
		t.Fatal("remove failed")
	}
	got, _ := os.ReadFile(path)
	if bytes.Contains(got, markerBegin) {
		t.Errorf("block survived removal:\n%s", got)
	}
	if !bytes.Equal(got, []byte(original)) {
		t.Errorf("round-trip not byte-identical:\nwant:\n%q\ngot:\n%q", original, got)
	}
}

func TestMarkerCRLFPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	crlf := "# Title\r\n\r\nprose\r\n"
	if err := os.WriteFile(path, []byte(crlf), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := installMarkerFile(path); !ok {
		t.Fatal("install failed")
	}
	got, _ := os.ReadFile(path)
	if bytes.Contains(got, []byte("\n")) && !bytes.Contains(got, []byte("\r\n")) {
		t.Errorf("CRLF flavour lost")
	}
	// The block itself must be CRLF so the file has no mixed EOLs.
	if bytes.Contains(got, append([]byte("<!-- BEGIN ABCD -->"), '\n')) &&
		!bytes.Contains(got, append([]byte("<!-- BEGIN ABCD -->"), []byte("\r\n")...)) {
		t.Errorf("block wrapped with LF in a CRLF file:\n%q", got)
	}
	if classifyMarker(path) != markerCurrent {
		t.Errorf("state = %q, want current", classifyMarker(path))
	}
}
