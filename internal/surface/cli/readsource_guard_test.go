package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/intentdriven/abcd/internal/core/ideate"
	"github.com/intentdriven/abcd/internal/core/lifeboat"
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

// TestStdinOperandReadersRefuseOverCapWhole pins that every "-" (stdin) operand
// reader refuses an over-cap payload whole rather than truncating it into a
// severed prefix. Before the cap+1 probe each read exactly `cap` bytes, so an
// over-cap payload was silently cut — its length-cap refusal never fired, the
// file transport refused the same bytes, and on the history-capture path a
// truncated transcript would be stored under a sha256 idempotency key computed
// over the prefix (spc-4's refuse-whole invariant).
func TestStdinOperandReadersRefuseOverCapWhole(t *testing.T) {
	readers := []struct {
		name string
		cap  int64
		read func(*cobra.Command) ([]byte, error)
	}{
		{"readSource", maxOperandJSONBytes, func(c *cobra.Command) ([]byte, error) { return readSource(c, "-") }},
		{"readIdeatePayload", ideate.MaxPayloadBytes, func(c *cobra.Command) ([]byte, error) { return readIdeatePayload(c, "-") }},
		{"readLessonsPayload", lifeboat.MaxLessonsBytes, func(c *cobra.Command) ([]byte, error) { return readLessonsPayload(c, "-") }},
		{"readSynthesisPayload", lifeboat.MaxSynthesisBytes, func(c *cobra.Command) ([]byte, error) { return readSynthesisPayload(c, "-") }},
	}
	for _, r := range readers {
		t.Run(r.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.SetIn(bytes.NewReader(make([]byte, r.cap+1)))
			data, err := r.read(cmd)
			if err == nil {
				t.Fatalf("%s accepted an over-cap stdin payload and returned %d bytes (a truncated prefix); it must be refused whole", r.name, len(data))
			}
		})
	}
}
