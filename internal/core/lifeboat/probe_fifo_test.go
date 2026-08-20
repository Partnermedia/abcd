//go:build unix

package lifeboat

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestListDirDoesNotBlockOnFifo proves SourceContext.ListDir opens directory
// entries non-blocking: a FIFO planted at a path a probe adapter lists (a
// statically-planted trap, no race) must not hang the probe. Before the fix
// ListDir opened with a plain O_RDONLY, so opening the FIFO blocked in the
// kernel until a writer appeared and abcd disembark probe/plan/pack hung.
func TestListDirDoesNotBlockOnFifo(t *testing.T) {
	dir := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(dir, "docs"), 0o644); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}
	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatalf("newSourceContext: %v", err)
	}

	done := make(chan []string, 1)
	go func() { done <- ctx.ListDir("docs") }()
	select {
	case names := <-done:
		// A FIFO is not a directory, so the listing is empty — but promptly.
		if len(names) != 0 {
			t.Fatalf("ListDir on a FIFO returned entries: %v", names)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("ListDir hung on a FIFO (open must not block)")
	}
}

// TestReadLifeboatFileDoesNotBlockOnFifo proves the guarded lifeboat read refuses
// a FIFO promptly rather than blocking the open. A vetted regular file in an
// untrusted lifeboat can be swapped for a FIFO between the manifest walk and the
// read; the read must return an error, not hang embark.
func TestReadLifeboatFileDoesNotBlockOnFifo(t *testing.T) {
	dir := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(dir, "trap.md"), 0o644); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer root.Close()

	done := make(chan error, 1)
	go func() {
		_, e := readLifeboatFile(root, "trap.md")
		done <- e
	}()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("readLifeboatFile must refuse a FIFO, not read it")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("readLifeboatFile hung on a FIFO (open must not block)")
	}
}
