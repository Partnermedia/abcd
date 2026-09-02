package ahoy

import (
	"os"
	"path/filepath"
)

// pluginDataDir is the harness's persistent per-plugin data directory
// (CLAUDE_PLUGIN_DATA): it survives plugin updates and is deleted only on full
// uninstall, which makes it the platform-sanctioned home for the bootstrap's
// download cache and for abcd's PATH-entry provenance record (spc-35). Empty
// when the harness exports none — every caller degrades loudly to the
// pre-cache behaviour rather than deriving the documented path shape, because
// a wrong guess would plant a trusted artefact in an untracked location.
func pluginDataDir() string {
	return os.Getenv("CLAUDE_PLUGIN_DATA")
}

// dataDirHazard reports why dataDir cannot be trusted as the harness's
// persistent data directory, or "" when it has the shape that directory always
// has: an absolute path, outside the repository being installed, not
// world-writable. CLAUDE_PLUGIN_DATA is read from the environment unexamined,
// and the owned-copy promotion re-verifies the cache only against the record
// beside it, so a value of any other shape would let whoever chose it — a
// relative value resolves against the checkout the verb runs in, an
// in-checkout value is committed bytes, a world-writable cache is any local
// user's — bless their own bytes as the owned PATH binary (sub-finding of
// GHSA-4q78-ccfv-f374). The harness never produces these shapes, so refusing
// them costs a real install nothing; binding the cache to an attestation the
// env cannot supply is the parent record's open decision and is not attempted
// here.
func dataDirHazard(dataDir, cwd string) string {
	if dataDir == "" {
		return ""
	}
	if !filepath.IsAbs(dataDir) {
		return "it is a relative path, which resolves against whatever directory the verb happens to run in"
	}
	if abs, err := filepath.Abs(cwd); err == nil && under(resolvePath(abs), resolvePath(dataDir)) {
		return "it lies inside the repository being installed, so its cache would be committed bytes"
	}
	for _, dir := range []string{dataDir, filepath.Join(dataDir, "cache")} {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() && fi.Mode().Perm()&0o002 != 0 {
			return "it is world-writable (" + displayPath(dir) + "), so any local user could replace both the artefact and its recorded hash"
		}
	}
	return ""
}
