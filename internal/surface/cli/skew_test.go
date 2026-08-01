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
	// The comparison establishes "different", never "which is ahead": a plugin
	// root can be an OLDER commit than the release the binary was cut from (a
	// pinned or downgraded install), and a notice claiming the binary is behind
	// would then be the exact opposite of the truth. Say different; say the two
	// directions it could be; assert neither.
	if !strings.Contains(stderr, "different commits") {
		t.Errorf("the notice must say the two are at different commits, not which one is ahead; stderr = %q", stderr)
	}
	for _, forbidden := range []string{"not in the binary", "since that release are not"} {
		if strings.Contains(stderr, forbidden) {
			t.Errorf("the notice must not assert a direction the comparison never establishes (%q); stderr = %q", forbidden, stderr)
		}
	}
}

// TestHookSessionStartSilentOnMalformedMeta: `.binary-meta` is written by a shell
// script that a crash can interrupt, so a value that is merely "not the literal
// string unknown" is not evidence of anything. A truncated commit renders a
// notice off garbage — shortSHA would happily abbreviate half a hash — so the
// shape is required: forty lowercase hex characters, or nothing is said.
func TestHookSessionStartSilentOnMalformedMeta(t *testing.T) {
	good := "4a4b4c4d4e4f4a4b4c4d4e4f4a4b4c4d4e4f4a4b"
	cases := []struct {
		name string
		meta string
	}{
		{"release commit truncated by a crash", "release_tag=v0.4.9\nrelease_sha=44444444\nplugin_sha=" + good + "\n"},
		{"plugin commit truncated by a crash", "release_tag=v0.4.9\nrelease_sha=" + good + "\nplugin_sha=5555\n"},
		{"release commit is not hex", "release_tag=v0.4.9\nrelease_sha=zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz\nplugin_sha=" + good + "\n"},
		{"release commit is upper-case", "release_tag=v0.4.9\nrelease_sha=" + strings.ToUpper(good) + "\nplugin_sha=" + good + "\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, _ := sessionEndRepo(t)
			skewMeta(t, tc.meta)

			stdout, stderr, code := runSessionStart(startPayload("s-malformed", repo), "hook", "session-start")
			if code != 0 {
				t.Errorf("a malformed commit must be treated as unresolved (exit 0), got %d", code)
			}
			if stdout != "" || stderr != "" {
				t.Errorf("must be silent; stdout=%q stderr=%q", stdout, stderr)
			}
		})
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
