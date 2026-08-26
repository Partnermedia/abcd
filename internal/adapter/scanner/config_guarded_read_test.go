package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The per-repo .abcd/config/pii.json override is a trust boundary: it sits in a
// repo working tree, so untrusted content (a hostile clone, a pull request, an
// agent writing into the tree) can replace it with something that is not a
// regular file. It is also reached automatically — the SessionEnd hook's history
// capture builds a Scanner — so nothing human stands between a planted path and
// the read.
//
// These tests pin the three properties fsutil.ReadGuarded gives and a bare
// os.ReadFile does not: a FIFO cannot block the open, a symlinked leaf is never
// followed, and an oversize file is refused rather than read whole. Each one
// leaves the scanner in the existing fail-closed degraded state (iss-202).

// newWithinDeadline builds a Scanner for repoRoot and fails the test if New has
// not returned within d. A bare os.ReadFile on a FIFO with no writer blocks
// forever, so this is the assertion that separates a guarded open from a hang —
// without it the test would simply never finish.
func newWithinDeadline(t *testing.T, repoRoot string, d time.Duration) *Scanner {
	t.Helper()
	type result struct {
		s   *Scanner
		err error
	}
	done := make(chan result, 1)
	go func() {
		s, err := New(repoRoot)
		done <- result{s, err}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("New returned an error: %v", r.err)
		}
		return r.s
	case <-time.After(d):
		// The goroutine is wedged in the read and cannot be reclaimed; failing
		// here reports the defect rather than hanging the package's test binary.
		t.Fatalf("New did not return within %s — the config read blocked", d)
		return nil
	}
}

func writeRepoConfig(t *testing.T, repoRoot string) string {
	t.Helper()
	cfgPath := filepath.Join(repoRoot, repoConfigRelPath)
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	return cfgPath
}

// A FIFO planted at the config path must not block New. Opened without
// O_NONBLOCK, a reader on a writerless FIFO blocks indefinitely — and the
// SessionEnd capture path holds the history repo lock across this call, so the
// hang wedges transcript capture for the repo permanently, not just for one run.
func TestNewRefusesFIFOConfigWithoutBlocking(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("mkfifo is not available on Windows")
	}
	repoRoot := t.TempDir()
	cfgPath := writeRepoConfig(t, repoRoot)
	if err := syscall.Mkfifo(cfgPath, 0o644); err != nil {
		t.Skipf("mkfifo unavailable on this filesystem: %v", err)
	}

	s := newWithinDeadline(t, repoRoot, 3*time.Second)

	unavailable, reason := s.Unavailable()
	if !unavailable {
		t.Fatal("a FIFO config left the scanner available; it must fail closed")
	}
	if !strings.Contains(reason, repoConfigRelPath) {
		t.Errorf("reason does not name the config file: %q", reason)
	}
}

// A symlinked leaf must never be followed. Committed as git mode 120000, this
// needs no local write access to plant, and /dev/zero yields an endless read
// that grows the process toward OOM.
func TestNewRefusesSymlinkedConfigLeaf(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	repoRoot := t.TempDir()
	cfgPath := writeRepoConfig(t, repoRoot)

	// A regular file with valid content is the symlink target, so the test
	// distinguishes "the symlink was refused" from "the target was unreadable".
	target := filepath.Join(t.TempDir(), "real.json")
	if err := os.WriteFile(target, []byte(`{"skip_dirs":["vendor"]}`), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.Symlink(target, cfgPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	s := newWithinDeadline(t, repoRoot, 3*time.Second)

	if unavailable, _ := s.Unavailable(); !unavailable {
		t.Fatal("a symlinked config was followed; the leaf must not be followed")
	}
}

// A symlinked ancestor redirects the whole read out of the tree, and the leaf
// guard cannot see it — O_NOFOLLOW applies to the final component only. The
// config sits two directories down, so BOTH ancestors are attack surface: the
// hand-rolled check this replaced lstat'd .abcd alone, and a symlinked
// .abcd/config walked the read out of the tree with no guard firing. The
// os.Root containment refuses the escape wherever it sits.
func TestNewRefusesSymlinkedConfigAncestor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	for _, tc := range []struct {
		name string
		// plant maps repoRoot to the ancestor that becomes the symlink and the
		// out-of-tree layout it points at.
		plant func(t *testing.T, repoRoot string)
	}{
		{"symlinked .abcd", func(t *testing.T, repoRoot string) {
			elsewhere := t.TempDir()
			if err := os.MkdirAll(filepath.Join(elsewhere, "config"), 0o755); err != nil {
				t.Fatalf("mkdir elsewhere: %v", err)
			}
			if err := os.WriteFile(filepath.Join(elsewhere, "config", "pii.json"), []byte(`{"skip_dirs":["vendor"]}`), 0o644); err != nil {
				t.Fatalf("write elsewhere config: %v", err)
			}
			if err := os.Symlink(elsewhere, filepath.Join(repoRoot, ".abcd")); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
		{"symlinked .abcd/config", func(t *testing.T, repoRoot string) {
			elsewhere := t.TempDir()
			if err := os.WriteFile(filepath.Join(elsewhere, "pii.json"), []byte(`{"skip_dirs":["vendor"]}`), 0o644); err != nil {
				t.Fatalf("write elsewhere config: %v", err)
			}
			if err := os.MkdirAll(filepath.Join(repoRoot, ".abcd"), 0o755); err != nil {
				t.Fatalf("mkdir .abcd: %v", err)
			}
			if err := os.Symlink(elsewhere, filepath.Join(repoRoot, ".abcd", "config")); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			tc.plant(t, repoRoot)

			s := newWithinDeadline(t, repoRoot, 3*time.Second)

			if unavailable, _ := s.Unavailable(); !unavailable {
				t.Fatalf("%s was followed; the read must be refused", tc.name)
			}
		})
	}
}

// An oversize regular file is refused on the fstat rather than read whole, so a
// large planted file cannot be turned into unbounded allocation.
func TestNewRefusesOversizeConfig(t *testing.T) {
	repoRoot := t.TempDir()
	cfgPath := writeRepoConfig(t, repoRoot)

	// The oversize file must be VALID JSON that the merge accepts. Padding with
	// whitespace would fail the parse instead, so the test would pass without a
	// size cap ever running and prove nothing about the property it names.
	var b strings.Builder
	b.WriteString(`{"skip_dirs":[`)
	for b.Len() < maxScannerConfigBytes {
		b.WriteString(`"` + strings.Repeat("a", 512) + `",`)
	}
	b.WriteString(`"vendor"]}`)
	if b.Len() <= maxScannerConfigBytes {
		t.Fatalf("test fixture is not oversize: %d bytes", b.Len())
	}
	if err := os.WriteFile(cfgPath, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write oversize config: %v", err)
	}

	s := newWithinDeadline(t, repoRoot, 3*time.Second)

	if unavailable, _ := s.Unavailable(); !unavailable {
		t.Fatal("an oversize config was accepted; it must be refused at the cap")
	}
}

// The control: the guarded read must not disturb the two paths that already
// work. An absent config leaves the built-in defaults standing and the scanner
// available; a well-formed one is still merged.
func TestNewGuardedReadKeepsWorkingPaths(t *testing.T) {
	t.Run("absent config yields defaults", func(t *testing.T) {
		repoRoot := t.TempDir()
		s := newWithinDeadline(t, repoRoot, 3*time.Second)
		if unavailable, reason := s.Unavailable(); unavailable {
			t.Fatalf("absent config marked the scanner unavailable: %s", reason)
		}
	})

	t.Run("valid config is merged", func(t *testing.T) {
		repoRoot := t.TempDir()
		cfgPath := writeRepoConfig(t, repoRoot)
		if err := os.WriteFile(cfgPath, []byte(`{"skip_dirs":["vendor"]}`), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		s := newWithinDeadline(t, repoRoot, 3*time.Second)
		if unavailable, reason := s.Unavailable(); unavailable {
			t.Fatalf("valid config marked the scanner unavailable: %s", reason)
		}
	})
}
