package ahoy

import (
	"testing"
)

// TestResolveSymlinkDestRelativeToLinkDir is the attack/behaviour test for the
// relative-symlink fix: a RELATIVE readlink target must be resolved against the
// symlink's OWN directory (as the kernel does), not the process working directory.
// Feeding a relative dest straight to resolvePath resolved it against the CWD, so a
// correct relative install (/usr/local/bin/abcd -> ../lib/abcd/abcd) was
// misread as foreign from any other directory.
func TestResolveSymlinkDestRelativeToLinkDir(t *testing.T) {
	link := "/usr/local/bin/abcd"
	rel := "../lib/abcd/abcd"
	want := resolvePath("/usr/local/lib/abcd/abcd")
	if got := resolveSymlinkDest(link, rel); got != want {
		t.Errorf("resolveSymlinkDest(%q, %q) = %q, want %q (relative target must resolve against the link's dir)", link, rel, got, want)
	}
	// An absolute target is used as-is.
	abs := "/opt/abcd/abcd"
	if got := resolveSymlinkDest(link, abs); got != resolvePath(abs) {
		t.Errorf("resolveSymlinkDest with absolute target = %q, want %q", got, resolvePath(abs))
	}
}
