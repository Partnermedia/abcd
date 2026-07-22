package launch

import (
	"path/filepath"
	"testing"
)

// TestCommittedTreeSatisfiesDevPolarity is the adr-19 detector for THIS
// repository, not a synthetic fixture: the committed manifests must carry no
// version key, and the committed version-location contract must be readable
// enough to say where that key would be.
//
// It is pinned here rather than left to the synthetic lockstep tests because the
// two halves fail in opposite directions and only the real tree proves both at
// once: an unreadable contract (no version-location.json) makes the whole gate
// inert, and a version key added to a working-tree manifest silently breaks the
// premise that the version is an output of the release cut.
func TestCommittedTreeSatisfiesDevPolarity(t *testing.T) {
	root := repoRootForTest(t)
	res := CheckLockstep(TreeDev, root, filepath.Join(root, versionLocationRelPath))
	if res.Unreadable {
		t.Fatalf("the committed version-location contract must be readable, got %s", res.Detail)
	}
	if !res.OK || res.ExitCode != 0 {
		t.Fatalf("adr-19: the committed tree must be version-ABSENT, got drifts %v", res.Drifts)
	}
}

// TestCommittedContractSelectsPluginManifest pins the version location the
// render writes to. A silent relocation would leave the render stamping a
// version somewhere the harness never reads, and the lockstep check would still
// pass because it reads the same moved pointer.
func TestCommittedContractSelectsPluginManifest(t *testing.T) {
	root := repoRootForTest(t)
	decision, err := loadJSON(filepath.Join(root, versionLocationRelPath))
	if err != nil {
		t.Fatalf("read the committed version-location contract: %v", err)
	}
	path, ptr, verr := validateVersionLocation(decision)
	if verr != "" {
		t.Fatalf("the committed contract must validate, got %s", verr)
	}
	if path != ".claude-plugin/plugin.json" || ptr != "/version" {
		t.Errorf("expected adr-19's ACCEPT outcome (.claude-plugin/plugin.json /version), got %q %q", path, ptr)
	}
}
