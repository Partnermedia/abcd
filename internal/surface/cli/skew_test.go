package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skewMeta writes a .binary-meta into a fresh plugin root and points the
// resolution env at it, the way the bootstrap and the harness do together.
func skewMeta(t *testing.T, body string) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".binary-meta"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ABCD_PLUGIN_ROOT", root)
	t.Setenv("CLAUDE_PLUGIN_ROOT", root)
}

// TestHookSessionStartReportsBinarySkew is itd-105's version-skew AC: an
// installed surface newer than the newest published binary is SAID, not silently
// run — otherwise a session executes old-binary logic against new-surface
// expectations and nothing anywhere admits it.
func TestHookSessionStartReportsBinarySkew(t *testing.T) {
	repo, _ := sessionEndRepo(t) // ready store: the transcripts notice stays silent
	skewMeta(t, "release_tag=v0.4.9\nrelease_sha=1111111111111111111111111111111111111111\nfetched_at=2026-08-01T00:00:00Z\nplugin_sha=2222222222222222222222222222222222222222\n")

	_, stderr, code := runSessionStart(startPayload("s-skew", repo), "hook", "session-start")

	if code == 0 {
		t.Error("the skew notice must exit non-zero so SessionStart renders it; got exit 0 (silent)")
	}
	for _, want := range []string{"222222222222", "v0.4.9", "111111111111"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("the notice must name the surface commit, the release tag, and the release commit; missing %q in %q", want, stderr)
		}
	}
	if !strings.Contains(stderr, "not in the binary") {
		t.Errorf("the notice must say newer merged fixes are not in the binary yet; stderr = %q", stderr)
	}
}

// TestHookSessionStartSilentOnUnknownOrMatchedRelease is the honesty half: the
// bootstrap records `unknown` when it could not resolve the release commit, and a
// notice built on that would be a guess. A binary built from the very commit the
// surface is at is not skewed at all.
func TestHookSessionStartSilentOnUnknownOrMatchedRelease(t *testing.T) {
	same := "3333333333333333333333333333333333333333"
	cases := []struct {
		name string
		meta string
	}{
		{"release commit unresolved", "release_tag=v0.4.9\nrelease_sha=unknown\nfetched_at=2026-08-01T00:00:00Z\nplugin_sha=" + same + "\n"},
		{"surface and binary share a commit", "release_tag=v0.4.9\nrelease_sha=" + same + "\nfetched_at=2026-08-01T00:00:00Z\nplugin_sha=" + same + "\n"},
		{"plugin commit unresolved", "release_tag=v0.4.9\nrelease_sha=" + same + "\nfetched_at=2026-08-01T00:00:00Z\nplugin_sha=unknown\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, _ := sessionEndRepo(t)
			skewMeta(t, tc.meta)

			stdout, stderr, code := runSessionStart(startPayload("s-quiet", repo), "hook", "session-start")
			if code != 0 {
				t.Errorf("must exit 0 (nothing can honestly be claimed), got %d", code)
			}
			if stdout != "" || stderr != "" {
				t.Errorf("must be silent; stdout=%q stderr=%q", stdout, stderr)
			}
		})
	}
}
