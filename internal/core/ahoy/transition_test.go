package ahoy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionTransitionFrom(t *testing.T) {
	cases := []struct {
		name        string
		recorded    string
		running     string
		wantChanged bool
	}{
		{"a newer running version is a transition", "v1.2.0", "v1.3.0", true},
		{"an older running version is still a transition (direction-neutral)", "v1.3.0", "v1.2.0", true},
		{"same version is no transition", "v1.2.0", "v1.2.0", false},
		{"a dev running version never reports a transition", "v1.2.0", "dev", false},
		{"an unrecorded version never reports a transition", "", "v1.3.0", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, changed := versionTransitionFrom(c.recorded, c.running)
			if changed != c.wantChanged {
				t.Fatalf("changed = %v, want %v", changed, c.wantChanged)
			}
		})
	}
}

func TestRecordedSetupVersionReadsConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	const cfg = `{"meta":{"setup_version":"v1.2.0","schema_version":1}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, ".abcd", "config.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := recordedSetupVersion(dir); got != "v1.2.0" {
		t.Fatalf("recordedSetupVersion = %q, want v1.2.0", got)
	}
	if got := recordedSetupVersion(t.TempDir()); got != "" {
		t.Fatalf("recordedSetupVersion with no config = %q, want empty", got)
	}
}
