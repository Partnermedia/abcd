package launch

import (
	"path/filepath"
	"testing"
)

// writeLockstepTree writes a version-location contract, a primary plugin.json
// and a marketplace.json with the given version values (empty string == omit the
// key entirely, to exercise the absent-vs-null distinction).
func writeLockstepTree(t *testing.T, root, primaryVer, marketVer, changelogVer string) string {
	t.Helper()
	writeFile(t, root, ".abcd/config/version-location.json",
		`{"manifest_path": ".claude-plugin/plugin.json", "json_pointer": "/version"}`)
	if primaryVer == "" {
		writeFile(t, root, ".claude-plugin/plugin.json", `{"name": "abcd"}`)
	} else {
		writeFile(t, root, ".claude-plugin/plugin.json", `{"name": "abcd", "version": "`+primaryVer+`"}`)
	}
	mk := `{"plugins": [{"name": "abcd"`
	if marketVer != "" {
		mk += `, "version": "` + marketVer + `"`
	}
	if changelogVer != "" {
		mk += `, "changelog": {"version": "` + changelogVer + `"}`
	}
	mk += `}]}`
	writeFile(t, root, ".claude-plugin/marketplace.json", mk)
	return filepath.Join(root, ".abcd/config/version-location.json")
}

func TestLockstepAgreement(t *testing.T) {
	root := t.TempDir()
	vl := writeLockstepTree(t, root, "1.2.3", "1.2.3", "1.2.3")
	res := CheckLockstep(TreePublic, root, vl)
	if !res.OK || res.ExitCode != 0 {
		t.Errorf("expected agreement OK/0, got %+v", res)
	}
}

func TestLockstepDrift(t *testing.T) {
	root := t.TempDir()
	vl := writeLockstepTree(t, root, "1.2.3", "1.2.4", "1.2.3")
	res := CheckLockstep(TreePublic, root, vl)
	if res.OK || res.ExitCode != 1 || len(res.Drifts) == 0 {
		t.Errorf("expected drift/1, got %+v", res)
	}
}

func TestLockstepNonSemverIsDrift(t *testing.T) {
	root := t.TempDir()
	vl := writeLockstepTree(t, root, "1.2", "1.2", "1.2")
	res := CheckLockstep(TreePublic, root, vl)
	if res.OK || res.ExitCode != 1 {
		t.Errorf("expected non-semver primary to drift, got %+v", res)
	}
}

func TestLockstepBlockedContractUnreadable(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/config/version-location.json", `{"blocked": true}`)
	writeFile(t, root, ".claude-plugin/plugin.json", `{}`)
	writeFile(t, root, ".claude-plugin/marketplace.json", `{"plugins":[{}]}`)
	res := CheckLockstep(TreePublic, root, filepath.Join(root, ".abcd/config/version-location.json"))
	if !res.Unreadable || res.ExitCode != 2 {
		t.Errorf("expected blocked contract → unreadable/2, got %+v", res)
	}
}

func TestLockstepDevKeysAbsent(t *testing.T) {
	root := t.TempDir()
	// Dev tree: no version keys anywhere.
	vl := writeLockstepTree(t, root, "", "", "")
	res := CheckLockstep(TreeDev, root, vl)
	if !res.OK || res.ExitCode != 0 {
		t.Errorf("dev tree with absent keys must be OK, got %+v", res)
	}
	// A present version key in the dev tree is drift.
	vl2 := writeLockstepTree(t, root, "1.2.3", "", "")
	res2 := CheckLockstep(TreeDev, root, vl2)
	if res2.OK || res2.ExitCode != 1 {
		t.Errorf("dev tree with a present version key must drift, got %+v", res2)
	}
}
