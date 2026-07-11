package identity

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// isolate points git at empty global/system config so the test only sees the
// temp repo's local config, making identity resolution hermetic regardless of
// the machine's real git identity.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
}

func gitRepo(t *testing.T, name, email string) string {
	t.Helper()
	dir := t.TempDir()
	runGitT(t, dir, "init")
	if name != "" {
		runGitT(t, dir, "config", "user.name", name)
	}
	if email != "" {
		runGitT(t, dir, "config", "user.email", email)
	}
	return dir
}

func runGitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writePin(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".abcd", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "identity.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheck_Match(t *testing.T) {
	isolate(t)
	dir := gitRepo(t, "Alex Reppel", "alex@example.com")
	writePin(t, dir, `{"name":"Alex Reppel","email":"alex@example.com"}`)
	res, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusOK {
		t.Fatalf("want StatusOK, got %v — %s", res.Status, res.Reason)
	}
}

func TestCheck_Mismatch(t *testing.T) {
	isolate(t)
	dir := gitRepo(t, "Test User", "test@example.com")
	writePin(t, dir, `{"name":"Alex Reppel","email":"alex@example.com"}`)
	res, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusMismatch {
		t.Fatalf("want StatusMismatch, got %v — %s", res.Status, res.Reason)
	}
	if res.Effective.Email != "test@example.com" || res.Pin.Email != "alex@example.com" {
		t.Fatalf("result did not carry both identities: %+v", res)
	}
}

func TestCheck_NoPin(t *testing.T) {
	isolate(t)
	dir := gitRepo(t, "Alex Reppel", "alex@example.com")
	// no identity.json written
	res, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusNoPin {
		t.Fatalf("want StatusNoPin, got %v — %s", res.Status, res.Reason)
	}
}

func TestCheck_UnsetIdentity(t *testing.T) {
	isolate(t)
	dir := gitRepo(t, "", "") // no user.name/email anywhere
	writePin(t, dir, `{"name":"Alex Reppel","email":"alex@example.com"}`)
	res, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusUnset {
		t.Fatalf("want StatusUnset, got %v — %s", res.Status, res.Reason)
	}
}

func TestLoadPin_Malformed(t *testing.T) {
	isolate(t)
	dir := t.TempDir()
	writePin(t, dir, `{"name": "no closing quote`)
	if _, _, err := LoadPin(dir); err == nil {
		t.Fatal("want error on malformed identity.json, got nil")
	}
}

func TestLoadPin_MissingFields(t *testing.T) {
	isolate(t)
	dir := t.TempDir()
	writePin(t, dir, `{"name":"Alex Reppel"}`) // no email
	if _, _, err := LoadPin(dir); err == nil {
		t.Fatal("want error when email is empty, got nil")
	}
}

func TestWritePin_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Pin{Name: "Alex Reppel", Email: "alex@example.com"}
	if err := WritePin(dir, want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := LoadPin(dir)
	if err != nil || !ok {
		t.Fatalf("LoadPin after WritePin: ok=%v err=%v", ok, err)
	}
	if got != want {
		t.Fatalf("round-trip mismatch: got %+v want %+v", got, want)
	}
}

func TestWritePin_RequiresBothFields(t *testing.T) {
	dir := t.TempDir()
	if err := WritePin(dir, Pin{Name: "Alex"}); err == nil {
		t.Fatal("want error when email is missing")
	}
}

// Blocks reports whether a pre-commit hook should refuse the commit: a mismatch
// or an unset identity blocks; OK and NoPin (opted-out) do not.
func TestBlocks(t *testing.T) {
	cases := map[Status]bool{
		StatusOK:       false,
		StatusNoPin:    false,
		StatusMismatch: true,
		StatusUnset:    true,
	}
	for s, want := range cases {
		if got := (Result{Status: s}).Blocks(); got != want {
			t.Errorf("Status %v: Blocks()=%v, want %v", s, got, want)
		}
	}
}
