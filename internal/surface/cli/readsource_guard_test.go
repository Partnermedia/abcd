package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestReadSourceRefusesSymlinkAndOversize is the attack-input test for the
// operand read behind --pages-json/--page-json (memory ingest). The operand is
// untrusted content (host-produced DistilledPage JSON, a cross-machine
// artifact), so a symlink where the operand should be must NOT be followed, and
// an over-cap file must be refused — matching every other guarded operand read.
func TestReadSourceRefusesSymlinkAndOversize(t *testing.T) {
	dir := t.TempDir()
	cmd := &cobra.Command{}

	// A symlinked operand must be refused (O_NOFOLLOW), never followed to its
	// target's content. Before the guard, os.ReadFile followed the link and
	// returned the secret target bytes.
	secret := filepath.Join(dir, "secret.json")
	if err := os.WriteFile(secret, []byte(`{"secret":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "operand.json")
	if err := os.Symlink(secret, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if data, err := readSource(cmd, link); err == nil {
		t.Errorf("readSource followed a symlinked operand and returned %q; a symlink must be refused", string(data))
	} else if strings.Contains(err.Error(), "too many levels of symbolic links") {
		t.Errorf("symlink refusal leaked the raw ELOOP syscall detail: %v", err)
	}

	// An over-cap file must be refused (bounded read), not slurped whole.
	big := filepath.Join(dir, "big.json")
	if err := os.WriteFile(big, make([]byte, maxOperandJSONBytes+1), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readSource(cmd, big); err == nil {
		t.Errorf("readSource read an over-cap operand; it must be refused")
	}

	// A real regular file within the cap still loads unchanged.
	ok := filepath.Join(dir, "ok.json")
	if err := os.WriteFile(ok, []byte(`{"k":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if data, err := readSource(cmd, ok); err != nil || string(data) != `{"k":1}` {
		t.Fatalf("readSource on a real file: data=%q err=%v", string(data), err)
	}
}
