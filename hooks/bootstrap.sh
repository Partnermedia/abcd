#!/bin/sh
# abcd plugin bootstrap: provision $CLAUDE_PLUGIN_ROOT/abcd from the latest
# release whenever it is missing (itd-105 / spc-21).
#
# POSIX sh and the base system only: the abcd binary is exactly what is missing
# on a fresh plugin install and after every plugin update (the harness re-clones
# into a fresh commit-stamped cache directory), so nothing here may depend on it.

set -u

plugin_root="${CLAUDE_PLUGIN_ROOT:-}"
[ -n "$plugin_root" ] || exit 0

binary="$plugin_root/abcd"
lock="$plugin_root/.bootstrap.lock"
tmp=""

repo_url="https://github.com/REPPL/abcd-cli"
releases_url="$repo_url/releases"
api_url="https://api.github.com/repos/REPPL/abcd-cli"
# Production fetches are HTTPS-only, redirects included. Without the pin a
# redirect could downgrade the transport (or point at file://) and hand both the
# payload and the manifest that verifies it to whoever can rewrite a response.
# The loopback overrides below clear it: the test fixture speaks plain http.
proto_opts='--proto =https --proto-redir =https'

cleanup() {
	[ -n "$tmp" ] && rm -rf "$tmp"
	rm -rf "$lock"
	return 0
}

# refuse is the single failure message every failing path shares: what is
# missing, what it costs, and the three ways out. A raw shell error ("No such
# file or directory") must never be the whole story a user gets.
refuse() {
	printf 'abcd bootstrap: %s\n\nThe abcd binary is not installed in the plugin root, so the abcd hooks cannot run and the shell-hazard guard is inactive — shell commands run UNGUARDED until it is.\n\nAny one of these fixes it:\n  - start a session with network access, and this script retries by itself;\n  - install the release binary by hand (%s#install) and copy it to %s;\n  - build from source for full trust: go build ./cmd/abcd, then copy the binary to %s.\n' \
		"$1" "$repo_url" "$binary" "$binary" >&2
	exit 1
}

# notice is a reported condition rather than a fault: the install succeeded, or
# the platform has no released binary. It exits 2 because a SessionStart hook's
# stdout becomes model context while only a NON-ZERO exit puts its stderr in
# front of the human (the same reason `abcd hook session-start` returns 2 for its
# own notices), and because SessionStart treats a non-zero exit as non-blocking:
# the later hooks in this event still run.
notice() {
	printf '%s\n' "$1" >&2
	exit 2
}

# 1. Fast path: a steady-state session pays one file test and no network. A
#    regular file, not merely something with an execute bit — a directory is
#    executable too (matching internal/core/ahoy's isExecutableFile).
[ -f "$binary" ] && [ -x "$binary" ] && exit 0

# 2. Platform gate. An unsupported platform is a reported condition, not a hook
#    fault, so it changes nothing and never blocks — but it is still said out
#    loud, because the shell guard stays inactive for as long as it holds.
os=$(uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')
arch=$(uname -m 2>/dev/null)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) arch='' ;;
esac
case "$os" in
	darwin | linux) ;;
	*) os='' ;;
esac
if [ -z "$os" ] || [ -z "$arch" ]; then
	notice "$(printf 'abcd bootstrap: no abcd binary is released for this platform (%s %s). Released binaries cover darwin and linux on amd64 and arm64 only, so nothing was downloaded and nothing was changed. To run abcd here, build from source: go build ./cmd/abcd, then copy the binary to %s.' \
		"$(uname -s 2>/dev/null)" "$(uname -m 2>/dev/null)" "$binary")"
fi

# 3. TEST-ONLY overrides (internal/surface/cli/bootstrap_test.go), mirroring
#    ABCD_BIN_TARGET's pattern — honoured for LOOPBACK origins only. Unrestricted,
#    either one would be a code-execution primitive rather than a test seam: the
#    binary and the checksums.txt that verifies it come from the same base, so
#    whoever sets the variable supplies both the payload and the manifest that
#    "verifies" it, and the binary installed here is then run unattended as the
#    Bash shell guard on every tool call. Refusing beats ignoring — an override
#    that was silently dropped would look like a passing test.
case "${ABCD_BOOTSTRAP_BASE_URL:-}" in
	'') ;;
	http://127.0.0.1:* | http://localhost:*)
		releases_url="$ABCD_BOOTSTRAP_BASE_URL/releases"
		proto_opts=''
		;;
	*) refuse "ABCD_BOOTSTRAP_BASE_URL is a loopback-only test override and was refused: ${ABCD_BOOTSTRAP_BASE_URL}" ;;
esac
case "${ABCD_BOOTSTRAP_API_URL:-}" in
	'') ;;
	http://127.0.0.1:* | http://localhost:*)
		api_url="$ABCD_BOOTSTRAP_API_URL"
		proto_opts=''
		;;
	*) refuse "ABCD_BOOTSTRAP_API_URL is a loopback-only test override and was refused: ${ABCD_BOOTSTRAP_API_URL}" ;;
esac

# 4. Concurrency lock. mkdir is atomic on POSIX, so the loser of the race is the
#    process whose mkdir fails; it exits quietly rather than racing the winner
#    into the same temp dir. A lock older than ten minutes belongs to a run that
#    was killed — without breaking it the plugin root stays unprovisionable.
if ! mkdir "$lock" 2>/dev/null; then
	if [ -n "$(find "$lock" -maxdepth 0 -mmin +10 2>/dev/null)" ]; then
		rm -rf "$lock"
	fi
	mkdir "$lock" 2>/dev/null || exit 0
fi
# A signal trap that RETURNS resumes the script (POSIX), which would carry on
# against directories cleanup just deleted and report a checksum mismatch that
# never happened. Terminate explicitly; cleanup is idempotent, so the EXIT trap
# firing again after it is a no-op.
trap cleanup EXIT
trap 'cleanup; exit 1' HUP INT TERM
# SIGKILL runs no trap, so a killed run leaves its PID-stamped temp directory
# behind holding a partially downloaded, UNVERIFIED binary. The lock is held from
# here on, so sweeping them cannot touch a live run's directory.
rm -rf "$plugin_root"/.bootstrap.tmp.* 2>/dev/null

command -v curl >/dev/null 2>&1 ||
	refuse 'curl is not available, so the release binary cannot be downloaded'

# 5. Resolve the release TAG before fetching anything from that release, so the
#    binary and the manifest that verifies it are pinned to ONE release: a
#    release cut between the two downloads would otherwise check a new
#    checksums.txt against an old binary and refuse a perfectly good artefact.
#    /releases/latest answers 302 -> /releases/tag/<tag>, and %{redirect_url} on
#    an UNFOLLOWED request is the only URL shape that survives: %{url_effective}
#    after -L lands on the asset CDN (release-assets.githubusercontent.com/...),
#    which carries no tag segment at all.
redirect=$(curl -fsS $proto_opts --max-time 15 -o /dev/null -w '%{redirect_url}' "$releases_url/latest" 2>/dev/null) || redirect=''
release_tag=$(printf '%s\n' "$redirect" | sed -n 's|.*/releases/tag/\([^/?#]*\).*|\1|p')
[ -n "$release_tag" ] ||
	refuse 'the latest release tag could not be resolved, so the download cannot be pinned to a single release — there may be no network'

# 6. Download into a temp dir under the plugin root — same filesystem, so the
#    install below is a rename and never a half-written binary.
asset="abcd-$os-$arch"
download_url="$releases_url/download/$release_tag"
tmp="$plugin_root/.bootstrap.tmp.$$"
rm -rf "$tmp"
mkdir -p "$tmp" 2>/dev/null ||
	refuse "a temporary directory cannot be created in the plugin root ($plugin_root)"

curl -fsSL $proto_opts --max-time 120 -o "$tmp/$asset" "$download_url/$asset" 2>/dev/null ||
	refuse "downloading $asset from release $release_tag failed — there may be no network, or that release may carry no asset for this platform"
curl -fsSL $proto_opts --max-time 30 -o "$tmp/checksums.txt" "$download_url/checksums.txt" 2>/dev/null ||
	refuse "downloading checksums.txt from release $release_tag failed, so the download cannot be verified and is not installed"

# 7. Verification against the same-origin manifest.
line=$(grep " $asset\$" "$tmp/checksums.txt" 2>/dev/null | head -n 1)
[ -n "$line" ] ||
	refuse "the release checksums.txt lists no entry for $asset, so the download cannot be verified and is not installed"
printf '%s\n' "$line" > "$tmp/manifest.txt"
if command -v shasum >/dev/null 2>&1; then
	verify='shasum -a 256 -c manifest.txt'
elif command -v sha256sum >/dev/null 2>&1; then
	verify='sha256sum -c manifest.txt'
else
	refuse 'neither shasum nor sha256sum is available, so the download cannot be verified and is not installed'
fi
(cd "$tmp" && $verify) > /dev/null 2>&1 ||
	refuse "the downloaded $asset does not match its SHA-256 checksum in the release checksums.txt — the artefact is corrupted or is not the published one, so nothing was installed"

# The hash just verified is recorded below, so a later check can tell "the binary
# at the guard path is the one that was verified" from "something replaced it".
binary_sha256=$(printf '%s\n' "$line" | sed -n 's/^\([0-9a-fA-F]\{64\}\).*/\1/p' | tr 'ABCDEF' 'abcdef')
[ -n "$binary_sha256" ] || binary_sha256=unknown

# The release commit is read from the API when it answers, and left unknown
# otherwise: the meta file is the skew notice's only evidence, so it never
# records a value it did not resolve.
release_sha=unknown
body=$(curl -fsSL $proto_opts --max-time 15 -H 'Accept: application/vnd.github+json' "$api_url/commits/$release_tag" 2>/dev/null) || body=''
candidate=$(printf '%s\n' "$body" | tr ',' '\n' |
	sed -n 's/.*"sha"[[:space:]]*:[[:space:]]*"\([0-9a-f]\{40\}\)".*/\1/p' | head -n 1)
[ -n "$candidate" ] && release_sha="$candidate"

# 8. Install, then record provenance.
chmod 0755 "$tmp/$asset" 2>/dev/null ||
	refuse "the downloaded $asset cannot be made executable"
mv -f "$tmp/$asset" "$binary" 2>/dev/null ||
	refuse "the verified $asset cannot be installed at $binary"

# plugin_sha is the harness's commit stamp: the plugin cache directory is named
# for the source commit it was cloned from. Anything else is not a commit and is
# recorded as such.
plugin_sha=$(basename "$plugin_root")
case "$plugin_sha" in
	*[!0-9a-f]*) plugin_sha=unknown ;;
esac
[ "${#plugin_sha}" -eq 40 ] || plugin_sha=unknown

# Written into the temp dir and renamed in, on the same filesystem, for the same
# reason the binary is: a crash mid-write would otherwise leave a truncated value
# that still parses, and a skew notice rendered off a truncated commit is a lie.
meta_note=''
{
	printf 'release_tag=%s\n' "$release_tag"
	printf 'release_sha=%s\n' "$release_sha"
	printf 'binary_sha256=%s\n' "$binary_sha256"
	printf 'fetched_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
	printf 'plugin_sha=%s\n' "$plugin_sha"
} > "$tmp/binary-meta" 2>/dev/null &&
	mv -f "$tmp/binary-meta" "$plugin_root/.binary-meta" 2>/dev/null ||
	meta_note=' (the .binary-meta provenance record could not be written, so version-skew reporting stays silent for this plugin root)'

# The one place PATH setup is suggested; the symlink itself stays owned by ahoy.
# This prints once per plugin root, because every later session takes the fast
# path above.
notice "$(printf 'abcd bootstrap: installed the checksum-verified abcd binary (release %s) into the plugin root, so the abcd hooks are live for this session. For the abcd command in your own terminal, run `abcd ahoy install` once.%s' \
	"$release_tag" "$meta_note")"
