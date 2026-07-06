package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewVersion(t *testing.T) {
	v := NewVersion()
	if v.Name != "abcd" {
		t.Fatalf("name = %q, want abcd", v.Name)
	}
	if v.Version == "" {
		t.Fatal("version is empty; ldflags default should be \"dev\"")
	}
}

func TestStatusBareDir(t *testing.T) {
	dir := t.TempDir()
	s, err := Status(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.IsGitRepo || s.HasRecord {
		t.Fatalf("bare dir reported git=%v record=%v, want both false", s.IsGitRepo, s.HasRecord)
	}
	if len(s.WorkTiers) != 0 {
		t.Fatalf("bare dir reported tiers %v, want none", s.WorkTiers)
	}
}

func TestStatusWithRecordAndGit(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustMkdir(t, filepath.Join(dir, ".abcd", "development"))
	mustMkdir(t, filepath.Join(dir, ".abcd", "work"))

	s, err := Status(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsGitRepo || !s.HasRecord {
		t.Fatalf("expected git+record, got git=%v record=%v", s.IsGitRepo, s.HasRecord)
	}
	if !contains(s.WorkTiers, "development") || !contains(s.WorkTiers, "work") {
		t.Fatalf("tiers = %v, want development+work", s.WorkTiers)
	}
	if contains(s.WorkTiers, "work.local") {
		t.Fatalf("tiers = %v, work.local should be absent", s.WorkTiers)
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
