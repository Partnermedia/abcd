package ahoy

import "os"

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
