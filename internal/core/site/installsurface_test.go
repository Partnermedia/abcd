package site

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The install command exists in four committed forms: the universal one-liner in
// README.md, the script in site-src/install.sh.tmpl, and the per-OS one-liners
// under docs/how-to/install.md's macOS and Linux headings. adr-47 makes all four
// surfaces of one repository, so they are one command written four times — and a
// reader who follows any of them must end up with the same bytes in the same
// place.
//
// The per-OS forms are the universal one with OS detection resolved: same release
// URLs, same checksum step, same install destination, `uname -s` replaced by the
// OS it would have printed. This test holds them to exactly that, so a change to
// one of them fails the build rather than shipping a page that quietly disagrees
// with the script beside it.
const (
	checksumsAsset = "checksums.txt"
	installDir     = "$HOME/.local/bin"
	osDetection    = "uname -s"

	verifierGNU = "sha256sum -c"
	verifierBSD = "shasum -a 256 -c"
)

// downloadURL matches a release-asset download, strictly: the scheme, host and
// path are the assertion, and only the asset name after the final slash is free.
var downloadURL = regexp.MustCompile(`https://[^\s"']+/releases/latest/download/[^\s"']*`)

// binaryName matches the shell assignment that names the release asset, in
// either spelling the surfaces use (`b=` in a one-liner, `binary=` in the
// script). The captured platform expression is what must agree.
var binaryName = regexp.MustCompile(`\b(?:b|binary)="abcd-([^"]*)"`)

// manifestLookup matches the grep that pulls this binary's line out of
// checksums.txt. Anchored on the trailing `$` because that anchor is the check:
// without it a manifest that does not list the binary still matches something.
var manifestLookup = regexp.MustCompile(`grep\s+"\s\$[A-Za-z0-9_]+\\?\$"\s+checksums\.txt`)

// installInvocation matches `install -m 0755` only in COMMAND POSITION — at the
// start of a line or after a shell separator. A Contains check would accept
// `sudo install -m 0755`, which is the one thing the install must never be:
// abcd does not escalate privileges, and a command that asks for a password to
// write into the user's own directory is a command that has been tampered with.
// Internal whitespace is free; the anchor is not.
var installInvocation = regexp.MustCompile(`(?m)(?:^|[;&|(]|\bthen\b|\bdo\b|\belse\b)\s*install\s+-m\s+0755\b`)

// privilegeEscalation matches the two escalation commands by word, so a mention
// inside a longer word ("pseudo") does not trip it.
var privilegeEscalation = regexp.MustCompile(`\b(?:sudo|doas)\b`)

// archMapping is one `uname -m` value and the release architecture it must map
// to. The release matrix names amd64 and arm64; a machine that reports the
// kernel spellings and is not translated downloads an asset that does not exist.
type archMapping struct{ from, to string }

var (
	amd64FromX8664   = archMapping{"x86_64", "amd64"}
	arm64FromAarch64 = archMapping{"aarch64", "arm64"}
)

// installSurface is one committed form of the install command.
type installSurface struct {
	name      string        // what to call it in a failure message
	source    string        // where it lives, repo-relative
	script    string        // the command text itself
	osToken   string        // the OS the form is specialised to: "$os" when it detects
	verifiers []string      // the checksum tools this form may use
	archMap   []archMapping // the `uname -m` translations this form must carry
}

func TestInstallSurfacesAgree(t *testing.T) {
	surfaces := loadInstallSurfaces(t)

	for _, s := range surfaces {
		t.Run(s.name, func(t *testing.T) {
			assertDownloads(t, s)
			assertChecksumStep(t, s)
			assertDestination(t, s)
			assertNoPrivilegeEscalation(t, s)
			assertOSResolution(t, s)
			assertArchMapping(t, s)
		})
	}
}

// assertDownloads checks that the form fetches exactly two things — the platform
// binary and the manifest that verifies it — from one release-download prefix,
// the same prefix every other form uses. Two origins would make the check
// vacuous: whoever serves the manifest can serve the payload it "verifies".
func assertDownloads(t *testing.T, s installSurface) {
	t.Helper()

	want := releaseDownloadPrefix(t)
	prefixes := map[string]bool{}
	gotChecksums, gotBinary := false, false
	for _, u := range downloadURL.FindAllString(s.script, -1) {
		cut := strings.LastIndex(u, "/")
		prefix, asset := u[:cut+1], u[cut+1:]
		prefixes[prefix] = true
		switch {
		case asset == checksumsAsset:
			gotChecksums = true
		case strings.HasPrefix(asset, "$"):
			gotBinary = true
		default:
			t.Errorf("%s downloads a fixed asset %q; the binary is per-platform", s.source, asset)
		}
	}

	if len(prefixes) != 1 {
		t.Fatalf("%s downloads from %d release prefixes, want exactly 1: %v", s.source, len(prefixes), sortedKeys(prefixes))
	}
	if !prefixes[want] {
		t.Errorf("%s downloads from %v, want %q", s.source, sortedKeys(prefixes), want)
	}
	if !gotBinary {
		t.Errorf("%s never downloads the platform binary from %s", s.source, want)
	}
	if !gotChecksums {
		t.Errorf("%s never downloads %s from %s", s.source, checksumsAsset, want)
	}
}

// assertChecksumStep checks that the form looks its binary up in the manifest and
// verifies the digest with the tool its platform has. A form that skipped either
// half would install an unverified download.
func assertChecksumStep(t *testing.T, s installSurface) {
	t.Helper()

	if !manifestLookup.MatchString(collapse(s.script)) {
		t.Errorf("%s has no anchored %s lookup for its binary", s.source, checksumsAsset)
	}

	var got []string
	for _, v := range []string{verifierGNU, verifierBSD} {
		if strings.Contains(collapse(s.script), v) {
			got = append(got, v)
		}
	}
	if !equalStrings(got, s.verifiers) {
		t.Errorf("%s verifies with %v, want %v", s.source, got, s.verifiers)
	}
}

// assertDestination checks that the form installs the same way into the same
// single-user directory, with the install itself in command position.
func assertDestination(t *testing.T, s installSurface) {
	t.Helper()

	if !installInvocation.MatchString(s.script) {
		t.Errorf("%s has no `install -m 0755` in command position", s.source)
	}
	if !strings.Contains(s.script, installDir) {
		t.Errorf("%s does not install into %q", s.source, installDir)
	}
}

// assertNoPrivilegeEscalation checks the invariant every form shares: abcd never
// asks for administrator rights. ~/.local/bin is the user's own directory, so a
// surface that reaches for sudo is either wrong or hostile.
func assertNoPrivilegeEscalation(t *testing.T, s installSurface) {
	t.Helper()

	if found := privilegeEscalation.FindAllString(s.script, -1); len(found) > 0 {
		t.Errorf("%s escalates privileges (%v); abcd installs into a directory the user already owns", s.source, found)
	}
}

// assertOSResolution checks the one difference the forms are allowed: a per-OS
// form names its OS where the universal form detects it, and nothing else moves.
func assertOSResolution(t *testing.T, s installSurface) {
	t.Helper()

	detects := strings.Contains(s.script, osDetection)
	if want := s.osToken == "$os"; detects != want {
		t.Errorf("%s detects the OS with %q: %v, want %v", s.source, osDetection, detects, want)
	}

	m := binaryName.FindAllStringSubmatch(s.script, -1)
	if len(m) != 1 {
		t.Fatalf("%s names the release binary %d times, want exactly 1", s.source, len(m))
	}
	platform := m[0][1]
	if !strings.HasPrefix(platform, s.osToken+"-") {
		t.Fatalf("%s builds the binary name from %q, want it to start with %q", s.source, platform, s.osToken)
	}
	if rest := strings.TrimPrefix(platform, s.osToken+"-"); rest != "$arch" {
		t.Errorf("%s resolves more than the OS: binary name is abcd-%s, want the architecture left as $arch", s.source, platform)
	}
}

// assertArchMapping checks that the form translates the kernel's architecture
// spellings into the ones the release matrix publishes. The OS-detecting forms
// carry both translations; the macOS form carries only x86_64, because Apple
// silicon already reports arm64 and aarch64 never appears there.
func assertArchMapping(t *testing.T, s installSurface) {
	t.Helper()

	for _, m := range s.archMap {
		mapping := regexp.MustCompile(regexp.QuoteMeta(m.from) + `\s*\)\s*arch=` + regexp.QuoteMeta(m.to) + `\b`)
		if !mapping.MatchString(s.script) {
			t.Errorf("%s does not map %s to %s; that machine would download an asset the release does not publish", s.source, m.from, m.to)
		}
	}
}

func loadInstallSurfaces(t *testing.T) []installSurface {
	t.Helper()

	readmePath := filepath.Join(repoRoot(), "README.md")
	templatePath := filepath.Join(repoRoot(), "site-src", "install.sh.tmpl")
	guidePath := filepath.Join(repoRoot(), "docs", "how-to", "install.md")

	guide := readFile(t, guidePath)
	bothArches := []archMapping{amd64FromX8664, arm64FromAarch64}

	return []installSurface{
		{
			name:      "readme-one-liner",
			source:    "README.md",
			script:    onlyInstallBlock(t, "README.md", fencedBlocks(readFile(t, readmePath))),
			osToken:   "$os",
			verifiers: []string{verifierGNU, verifierBSD},
			archMap:   bothArches,
		},
		{
			name:      "install-sh-template",
			source:    "site-src/install.sh.tmpl",
			script:    readFile(t, templatePath),
			osToken:   "$os",
			verifiers: []string{verifierGNU, verifierBSD},
			archMap:   bothArches,
		},
		{
			name:      "how-to-macos",
			source:    "docs/how-to/install.md (### macOS)",
			script:    onlyInstallBlock(t, "docs/how-to/install.md (### macOS)", blocksUnder(fencedBlocks(guide), "macOS")),
			osToken:   "darwin",
			verifiers: []string{verifierBSD},
			archMap:   []archMapping{amd64FromX8664},
		},
		{
			name:      "how-to-linux",
			source:    "docs/how-to/install.md (### Linux)",
			script:    onlyInstallBlock(t, "docs/how-to/install.md (### Linux)", blocksUnder(fencedBlocks(guide), "Linux")),
			osToken:   "linux",
			verifiers: []string{verifierGNU},
			archMap:   bothArches,
		},
	}
}

func repoRoot() string { return filepath.Join("..", "..", "..") }

// releaseDownloadPrefix derives the release-download base from go.mod's module
// path rather than hardcoding it. go.mod is the one place this repository's
// identity is already load-bearing for the build, so a rename lands there first
// and these surfaces fail WITH the rename instead of after it — the same
// derivation rule internal/core/launch's install-guide detector applies to the
// marketplace slug.
func releaseDownloadPrefix(t *testing.T) string {
	t.Helper()

	gomod := readFile(t, filepath.Join(repoRoot(), "go.mod"))
	for _, line := range strings.Split(gomod, "\n") {
		path, ok := strings.CutPrefix(strings.TrimSpace(line), "module ")
		if !ok {
			continue
		}
		path = strings.TrimSpace(path)
		if strings.Count(path, "/") != 2 {
			t.Fatalf("module path %q is not host/owner/repo", path)
		}
		return "https://" + path + "/releases/latest/download/"
	}
	t.Fatal("go.mod declares no module path")
	return ""
}

// fencedBlock is a fenced code block and the H3 it sits under, if any.
type fencedBlock struct {
	heading string
	body    string
}

// fencedBlocks splits Markdown into its fenced code blocks. Fences are matched on
// the trimmed line so indentation inside a list does not hide one. An H1 or H2
// CLEARS the current H3: without that, blocks in a later section inherit the last
// H3 seen and a block can be attributed to a heading it does not sit under.
func fencedBlocks(md string) []fencedBlock {
	var out []fencedBlock
	var body []string
	inFence, heading := false, ""

	for _, line := range strings.Split(md, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "```"):
			if inFence {
				out = append(out, fencedBlock{heading: heading, body: strings.Join(body, "\n")})
				body = nil
			}
			inFence = !inFence
		case inFence:
			body = append(body, line)
		case strings.HasPrefix(trimmed, "### "):
			heading = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
		case strings.HasPrefix(trimmed, "## "), strings.HasPrefix(trimmed, "# "):
			heading = ""
		}
	}
	return out
}

// blocksUnder keeps the blocks sitting under one H3.
func blocksUnder(blocks []fencedBlock, heading string) []fencedBlock {
	var out []fencedBlock
	for _, b := range blocks {
		if b.heading == heading {
			out = append(out, b)
		}
	}
	return out
}

// onlyInstallBlock returns the single block that downloads from the release, and
// fails when a surface grew a second one — two install commands on one page is
// the drift this test exists to catch.
func onlyInstallBlock(t *testing.T, source string, blocks []fencedBlock) string {
	t.Helper()

	var found []string
	for _, b := range blocks {
		if downloadURL.MatchString(b.body) {
			found = append(found, b.body)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%s holds %d release-download command blocks, want exactly 1", source, len(found))
	}
	return found[0]
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// collapse folds every run of whitespace to one space, so a command split over
// lines in the script reads the same as the one-liner it agrees with.
func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
