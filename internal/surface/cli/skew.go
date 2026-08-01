package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// binaryMetaFile is the provenance record hooks/bootstrap.sh writes into the
// plugin root when it installs a binary: which release it fetched, and which
// plugin commit it fetched that release for.
const binaryMetaFile = ".binary-meta"

// maxBinaryMetaBytes caps the read. Four key=value lines; anything larger is not
// the file the bootstrap wrote.
const maxBinaryMetaBytes = 4 << 10

// binarySkewNotice is the one-line session-start notice for a plugin surface
// that is newer than the binary serving it, or "" when nothing can be said.
//
// The plugin surface tracks the repository tip while the newest binary is the
// last tagged release, so a fix can merge without a release cut and leave the
// two apart (itd-105). The bootstrap records `unknown` for anything it could not
// resolve, and an unknown commit yields no notice: visibility that guesses is
// worse than none.
func binarySkewNotice() string {
	root := os.Getenv("ABCD_PLUGIN_ROOT")
	if root == "" {
		root = os.Getenv("CLAUDE_PLUGIN_ROOT")
	}
	if root == "" {
		return ""
	}
	meta := readBinaryMeta(filepath.Join(root, binaryMetaFile))
	pluginSHA, releaseSHA := meta["plugin_sha"], meta["release_sha"]
	if !resolvedSHA(pluginSHA) || !resolvedSHA(releaseSHA) || pluginSHA == releaseSHA {
		return ""
	}
	tag := meta["release_tag"]
	if tag == "" {
		tag = "unknown"
	}
	return fmt.Sprintf("abcd: the plugin surface is at commit %s, but the installed binary is release %s (commit %s) — fixes merged since that release are not in the binary yet.",
		shortSHA(pluginSHA), tag, shortSHA(releaseSHA))
}

// readBinaryMeta parses the bootstrap's key=value file. An unreadable or absent
// file is an empty map, which every caller reads as "nothing to say".
func readBinaryMeta(path string) map[string]string {
	data, err := fsutil.ReadGuarded(path, maxBinaryMetaBytes)
	if err != nil {
		return nil
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), "="); ok {
			out[k] = v
		}
	}
	return out
}

// resolvedSHA reports whether the bootstrap actually resolved this commit, as
// opposed to recording that it could not.
func resolvedSHA(s string) bool { return s != "" && s != "unknown" }

// shortSHA abbreviates a commit for a one-line notice, leaving anything shorter
// than the abbreviation alone.
func shortSHA(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:12]
}
