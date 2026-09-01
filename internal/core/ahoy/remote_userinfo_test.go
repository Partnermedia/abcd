package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// placeholderToken is a clearly fake credential, distinctive enough that no
// other fixture bytes can pass for it. Never a real one.
const placeholderToken = "ghp_EXAMPLE0000000000000000000000000000"

// TestScrubRemoteUserinfo pins the one derivation rule for the origin URL
// (GHSA-qc3w-8pv5-crc3): a credential in the userinfo never survives, while an
// SSH login name — which is a route, not a secret — does.
func TestScrubRemoteUserinfo(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"https://ci-bot:" + placeholderToken + "@github.com/owner/repo.git", "https://github.com/owner/repo.git"},
		{"https://" + placeholderToken + "@github.com/owner/repo.git", "https://github.com/owner/repo.git"},
		{"http://user:secret@example.com/o/r", "http://example.com/o/r"},
		{"https://github.com/owner/repo.git", "https://github.com/owner/repo.git"},
		{"ssh://git@github.com/owner/repo.git", "ssh://git@github.com/owner/repo.git"},
		{"ssh://git:secret@github.com/owner/repo.git", "ssh://github.com/owner/repo.git"},
		{"git@github.com:owner/repo.git", "git@github.com:owner/repo.git"},
		{"user:secret@example.com:owner/repo.git", "example.com:owner/repo.git"},
		{"/srv/git/repo.git", "/srv/git/repo.git"},
		{"", ""},
	} {
		if got := scrubRemoteUserinfo(tc.in); got != tc.want {
			t.Errorf("scrubRemoteUserinfo(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestDeriveIdentityStripsRemoteUserinfo: the identity every sink reads is
// derived once, so the scrub at that site is the whole fix.
func TestDeriveIdentityStripsRemoteUserinfo(t *testing.T) {
	setupHermetic(t)
	r := gittest.NewRepo(t)
	r.Git("remote", "add", "origin", "https://ci-bot:"+placeholderToken+"@github.com/owner/repo.git")
	id := deriveIdentity(r.Root())
	if id.Github != "https://github.com/owner/repo.git" {
		t.Fatalf("Github = %q, want the origin URL without its userinfo", id.Github)
	}
}

// TestInstallPersistsNoRemoteUserinfo drives a credentialed origin through a
// real install into an isolated home and reads back the two files the registry
// writes: neither index.json nor the per-repo meta.json may carry the credential.
func TestInstallPersistsNoRemoteUserinfo(t *testing.T) {
	home, _ := setupHermetic(t)
	r := gittest.NewRepo(t)
	r.Commit("root")
	r.Git("remote", "add", "origin", "https://ci-bot:"+placeholderToken+"@github.com/owner/repo.git")

	res, err := Install(r.Root(), installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status == "refused" {
		t.Fatalf("install refused: %v", res.Notes)
	}
	sha := deriveIdentity(r.Root()).RootSHA
	if sha == "" {
		t.Fatal("fixture has no root commit")
	}
	for _, rel := range []string{
		filepath.Join(".abcd", "history", "index.json"),
		filepath.Join(".abcd", "history", sha, "meta.json"),
	} {
		data, err := os.ReadFile(filepath.Join(home, rel))
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		if strings.Contains(string(data), placeholderToken) {
			t.Fatalf("%s persisted the origin credential:\n%s", rel, data)
		}
		if !strings.Contains(string(data), "https://github.com/owner/repo.git") {
			t.Fatalf("%s does not record the scrubbed origin URL:\n%s", rel, data)
		}
	}
}
