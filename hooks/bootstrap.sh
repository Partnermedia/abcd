#!/bin/sh
# abcd plugin bootstrap: provision $CLAUDE_PLUGIN_ROOT/abcd from the latest
# release whenever it is missing (itd-105 / spc-21).
#
# POSIX sh and the base system only: the abcd binary is exactly what is missing
# on a fresh plugin install and after every plugin update (the harness re-clones
# into a fresh commit-stamped cache directory), so nothing here may depend on it.
#
# ABCD_BOOTSTRAP_BASE_URL and ABCD_BOOTSTRAP_API_URL are TEST-ONLY overrides
# (internal/surface/cli/bootstrap_test.go), mirroring ABCD_BIN_TARGET's pattern.

set -u

plugin_root="${CLAUDE_PLUGIN_ROOT:-}"
[ -n "$plugin_root" ] || exit 0

binary="$plugin_root/abcd"
lock="$plugin_root/.bootstrap.lock"
tmp=""

repo_url="https://github.com/REPPL/abcd-cli"
base_url="${ABCD_BOOTSTRAP_BASE_URL:-$repo_url/releases/latest/download}"
api_url="${ABCD_BOOTSTRAP_API_URL:-https://api.github.com/repos/REPPL/abcd-cli}"

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

# 1. Fast path: a steady-state session pays one file test and no network.
[ -x "$binary" ] && exit 0

# 2. Platform gate. Exit 0, not an error: an unsupported platform is a reported
#    condition, not a hook fault to retry every session.
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
	printf 'abcd bootstrap: no abcd binary is released for this platform (%s %s). Released binaries cover darwin and linux on amd64 and arm64 only, so nothing was downloaded and nothing was changed. To run abcd here, build from source: go build ./cmd/abcd, then copy the binary to %s.\n' \
		"$(uname -s 2>/dev/null)" "$(uname -m 2>/dev/null)" "$binary"
	exit 0
fi

# 3. Concurrency lock. mkdir is atomic on POSIX, so the loser of the race is the
#    process whose mkdir fails; it exits quietly rather than racing the winner
#    into the same temp dir. A lock older than ten minutes belongs to a run that
#    was killed — without breaking it the plugin root stays unprovisionable.
if ! mkdir "$lock" 2>/dev/null; then
	if [ -n "$(find "$lock" -maxdepth 0 -mmin +10 2>/dev/null)" ]; then
		rm -rf "$lock"
	fi
	mkdir "$lock" 2>/dev/null || exit 0
fi
trap cleanup EXIT HUP INT TERM

command -v curl >/dev/null 2>&1 ||
	refuse 'curl is not available, so the release binary cannot be downloaded'

# 4. Download into a temp dir under the plugin root — same filesystem, so the
#    install below is a rename and never a half-written binary.
asset="abcd-$os-$arch"
tmp="$plugin_root/.bootstrap.tmp.$$"
rm -rf "$tmp"
mkdir -p "$tmp" 2>/dev/null ||
	refuse "a temporary directory cannot be created in the plugin root ($plugin_root)"

effective=$(curl -fsSL --max-time 120 -w '%{url_effective}' -o "$tmp/$asset" "$base_url/$asset" 2>/dev/null) ||
	refuse "downloading $asset from the latest release failed — there may be no network, or the latest release may carry no asset for this platform"
curl -fsSL --max-time 120 -o "$tmp/checksums.txt" "$base_url/checksums.txt" 2>/dev/null ||
	refuse 'downloading checksums.txt from the latest release failed, so the download cannot be verified and is not installed'

# The tag is whatever the /latest/ redirect resolved to; never guessed.
release_tag=$(printf '%s\n' "$effective" | sed -n 's|.*/releases/download/\([^/]*\)/.*|\1|p')
[ -n "$release_tag" ] || release_tag=unknown

# 5. Verification against the same-origin manifest.
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

# The release commit is read from the API when it answers, and left unknown
# otherwise: the meta file is the skew notice's only evidence, so it never
# records a value it did not resolve.
release_sha=unknown
if [ "$release_tag" != unknown ]; then
	body=$(curl -fsSL --max-time 30 -H 'Accept: application/vnd.github+json' "$api_url/commits/$release_tag" 2>/dev/null) || body=''
	candidate=$(printf '%s\n' "$body" | tr ',' '\n' |
		sed -n 's/.*"sha"[[:space:]]*:[[:space:]]*"\([0-9a-f]\{40\}\)".*/\1/p' | head -n 1)
	[ -n "$candidate" ] && release_sha="$candidate"
fi

# 6. Install, then record provenance.
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

{
	printf 'release_tag=%s\n' "$release_tag"
	printf 'release_sha=%s\n' "$release_sha"
	printf 'fetched_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
	printf 'plugin_sha=%s\n' "$plugin_sha"
} > "$plugin_root/.binary-meta" 2>/dev/null

# The one place PATH setup is suggested; the symlink itself stays owned by ahoy.
# This prints once per plugin root, because every later session takes the fast
# path above.
printf 'abcd bootstrap: installed the checksum-verified abcd binary (release %s) into the plugin root, so the abcd hooks are live for this session. For the abcd command in your own terminal, run `abcd ahoy install` once.\n' "$release_tag"
exit 0
