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

# The fetch origins are constants of this script, and there is deliberately no
# way to redirect them from the environment. An override here is not a
# convenience, it is a code-execution primitive: the binary AND the
# checksums.txt that verifies it are fetched from the same base, so whoever
# names the origin supplies both the payload and the manifest that "verifies"
# it — verification becomes vacuous — and the binary installed below is then run
# unattended as the Bash shell guard on every tool call. Two successive attempts
# to keep the seam behind a loopback allowlist were both defeated in adversarial
# review (URL userinfo of the form http://127.0.0.1:1@example.invalid slipped
# through the glob, and clearing the transport pin on the override path reopened
# redirect downgrade), so the seam is REMOVED rather than narrowed a third time.
# The tests rewrite these two literals in a throwaway COPY of this file; nothing
# in the shipped script consults the environment to decide where to fetch from.
repo_url="https://github.com/REPPL/abcd-cli"
releases_url="$repo_url/releases"
api_url="https://api.github.com/repos/REPPL/abcd-cli"

# Every fetch below is HTTPS-only, redirects included, unconditionally: there is
# no toggle and no code path that clears it. Without the pin a redirect could
# downgrade the transport (or point at file://) and hand both the payload and the
# manifest that verifies it to whoever can rewrite a response.
#
# The pin is not enough on its own, because curl reads a configuration surface
# this script's argv knows nothing about. Unless -q is its FIRST argument, curl
# loads $CURL_HOME/.curlrc (falling back to $HOME/.curlrc), and one `connect-to`
# or `resolve` line there re-points the connection while the URL above still
# literally reads https://github.com/… — the transport pin still holds, and the
# checksum check becomes vacuous, because the same config supplies the binary AND
# the checksums.txt that "verifies" it. Independent review reproduced exactly
# that. So: -q first on every call below, and the other names curl reads without
# being told to — the proxy variables, which can route both fetches through a
# server of the setter's choosing, and the CA overrides, which are what make such
# a route succeed on TLS — are removed here, before any fetch happens. The cost
# is deliberate and accepted: a machine that can only reach the network through a
# proxy no longer bootstraps automatically, and refuse() already names the
# manual install and build-from-source ways out.
unset HTTPS_PROXY https_proxy HTTP_PROXY http_proxy ALL_PROXY all_proxy CURL_HOME
unset CURL_CA_BUNDLE SSL_CERT_FILE SSL_CERT_DIR

cleanup() {
	[ -n "$tmp" ] && rm -rf "$tmp"
	rm -rf "$lock"
	return 0
}

# safe strips the control characters a message must never carry to a terminal.
# The values interpolated into the messages below are not this script's own text
# — a plugin-root path, a release tag read off a redirect — and a raw escape
# sequence in one of them can recolour, reposition, or visually rewrite what the
# reader is shown. Tab and newline survive; the report is made of them.
safe() {
	printf '%s' "$1" | tr -d '\000-\010\013-\037\177'
}

# refuse is the single failure message every failing path shares: what is
# missing, what it costs, and the three ways out. A raw shell error ("No such
# file or directory") must never be the whole story a user gets.
refuse() {
	printf 'abcd bootstrap: %s\n\nThe abcd binary is not installed in the plugin root, so the abcd hooks cannot run and the shell-hazard guard is inactive — shell commands run UNGUARDED until it is.\n\nAny one of these fixes it:\n  - start a session with network access, and this script retries by itself;\n  - install the release binary by hand (%s#install) and copy it to %s;\n  - build from source for full trust: go build ./cmd/abcd, then copy the binary to %s.\n' \
		"$(safe "$1")" "$repo_url" "$(safe "$binary")" "$(safe "$binary")" >&2
	exit 1
}

# notice is a reported condition rather than a fault: the install succeeded, or
# the platform has no released binary. It exits 2 because a SessionStart hook's
# stdout becomes model context while only a NON-ZERO exit puts its stderr in
# front of the human (the same reason `abcd hook session-start` returns 2 for its
# own notices), and because SessionStart treats a non-zero exit as non-blocking:
# the later hooks in this event still run.
notice() {
	printf '%s\n' "$(safe "$1")" >&2
	exit 2
}

# 1. Fast path: a steady-state session pays one file test and no network. A
#    regular file, not merely something with an execute bit — a directory is
#    executable too (matching internal/core/ahoy's isExecutableFile).
[ -f "$binary" ] && [ -x "$binary" ] && exit 0

# `mv -f` onto an existing DIRECTORY moves the file INTO it rather than
# replacing it, so a stray directory at $binary would otherwise "succeed"
# every session while the hooks stay broken and .binary-meta lies about it.
if [ -e "$binary" ] && [ ! -f "$binary" ]; then
	refuse "$binary exists and is not a regular file, so the release binary cannot be installed there"
fi

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

# 3. Concurrency lock. mkdir is atomic on POSIX, so the loser of the race is the
#    process whose mkdir fails; it exits quietly rather than racing the winner
#    into the same temp dir. A lock older than ten minutes belongs to a run that
#    was killed — without breaking it the plugin root stays unprovisionable.
if ! mkdir "$lock" 2>/dev/null; then
	if [ -n "$(find "$lock" -maxdepth 0 -mmin +10 2>/dev/null)" ]; then
		rm -rf "$lock"
	fi
	if ! mkdir "$lock" 2>/dev/null; then
		# A lock DIRECTORY that exists is the race: another run holds it, and
		# staying quiet is right. mkdir failing with no lock directory there is
		# something else — a read-only or unwritable plugin root, or a
		# non-directory squatting the lock path — and that is a permanent,
		# every-session failure the two cases must not share a silent exit with.
		[ -d "$lock" ] && exit 0
		refuse "the bootstrap lock ($lock) cannot be created and no lock directory is there, so this is not a concurrent run — the plugin root ($plugin_root) is not writable, or something that is not a directory occupies the lock path"
	fi
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

# 4. Resolve the release TAG before fetching anything from that release, so the
#    binary and the manifest that verifies it are pinned to ONE release: a
#    release cut between the two downloads would otherwise check a new
#    checksums.txt against an old binary and refuse a perfectly good artefact.
#    /releases/latest answers 302 -> /releases/tag/<tag>, and %{redirect_url} on
#    an UNFOLLOWED request is the only URL shape that survives: %{url_effective}
#    after -L lands on the asset CDN (release-assets.githubusercontent.com/...),
#    which carries no tag segment at all.
redirect=$(curl -q -fsS --proto '=https' --proto-redir '=https' --max-time 15 -o /dev/null -w '%{redirect_url}' "$releases_url/latest" 2>/dev/null) || redirect=''
release_tag=$(printf '%s\n' "$redirect" | sed -n 's|.*/releases/tag/\([^/?#]*\).*|\1|p')
[ -n "$release_tag" ] ||
	refuse 'the latest release tag could not be resolved, so the download cannot be pinned to a single release — there may be no network'

# 5. Download into a temp dir under the plugin root — same filesystem, so the
#    install below is a rename and never a half-written binary.
asset="abcd-$os-$arch"
download_url="$releases_url/download/$release_tag"
tmp="$plugin_root/.bootstrap.tmp.$$"
rm -rf "$tmp"
# `mkdir -p` succeeds on a directory that already exists — including one that
# reappeared, symlinked, in the window since the rm -rf above. Plain `mkdir`
# fails on that, turning a same-name race into a refusal instead of a curl
# write through a planted symlink.
mkdir "$tmp" 2>/dev/null ||
	refuse "a temporary directory cannot be created in the plugin root ($plugin_root)"

curl -q -fsSL --proto '=https' --proto-redir '=https' --max-time 120 -o "$tmp/$asset" "$download_url/$asset" 2>/dev/null ||
	refuse "downloading $asset from release $release_tag failed — there may be no network, or that release may carry no asset for this platform"
curl -q -fsSL --proto '=https' --proto-redir '=https' --max-time 30 -o "$tmp/checksums.txt" "$download_url/checksums.txt" 2>/dev/null ||
	refuse "downloading checksums.txt from release $release_tag failed, so the download cannot be verified and is not installed"

# 6. Verification against the same-origin manifest.
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

# The hash just verified is RECORDED below, not re-checked anywhere: nothing in
# this repository reads binary_sha256 today. It is provenance for a human (or a
# later verb) answering "is the binary at the guard path the one that was
# verified, or did something replace it" — the fast path deliberately does not
# recompute it, because its whole contract is that a steady-state session pays
# one file test and no more. Do not read this line as an automatic check.
binary_sha256=$(printf '%s\n' "$line" | sed -n 's/^\([0-9a-fA-F]\{64\}\).*/\1/p' | tr 'ABCDEF' 'abcdef')
[ -n "$binary_sha256" ] || binary_sha256=unknown

# The release commit is read from the API when it answers, and left unknown
# otherwise: the meta file is the skew notice's only evidence, so it never
# records a value it did not resolve.
release_sha=unknown
body=$(curl -q -fsSL --proto '=https' --proto-redir '=https' --max-time 15 -H 'Accept: application/vnd.github+json' "$api_url/commits/$release_tag" 2>/dev/null) || body=''
candidate=$(printf '%s\n' "$body" | tr ',' '\n' |
	sed -n 's/.*"sha"[[:space:]]*:[[:space:]]*"\([0-9a-f]\{40\}\)".*/\1/p' | head -n 1)
[ -n "$candidate" ] && release_sha="$candidate"

# 7. Install, then record provenance.
chmod 0755 "$tmp/$asset" 2>/dev/null ||
	refuse "the downloaded $asset cannot be made executable"
mv -f "$tmp/$asset" "$binary" 2>/dev/null ||
	refuse "the verified $asset cannot be installed at $binary"

# plugin_sha is the harness's commit stamp: the plugin cache directory is named
# for the source commit it was cloned from. Anything else is not a commit and is
# recorded as such.
#
# That naming is a WARRANT this repository takes from itd-105 and cannot verify
# against the real harness. If it ever stops holding, every basename fails the
# 40-hex gate, plugin_sha is permanently `unknown`, and the version-skew notice
# goes silent forever with nothing anywhere to look at. So the RAW basename is
# recorded beside the gated value: it is never compared and never rendered, it
# exists so that "why has the skew notice never fired" has an answer in the file
# rather than only in this comment. Control characters are stripped (a directory
# name may contain a newline, which would forge a key=value line) and the value
# is capped, so a pathological name cannot push .binary-meta past the guarded
# read size and silence the notice by a different route.
plugin_root_basename=$(basename "$plugin_root" | tr -d '\000-\037' | cut -c1-120)
plugin_sha="$plugin_root_basename"
case "$plugin_sha" in
	*[!0-9a-f]*) plugin_sha=unknown ;;
esac
[ "${#plugin_sha}" -eq 40 ] || plugin_sha=unknown

# Written into the temp dir and renamed in, on the same filesystem, for the same
# reason the binary is: a crash mid-write would otherwise leave a truncated value
# that still parses, and a skew notice rendered off a truncated commit is a lie.
meta_path="$plugin_root/.binary-meta"
meta_note=''
if [ -e "$meta_path" ] && [ ! -f "$meta_path" ]; then
	# The same `mv -f` hazard the binary path refuses above, and quieter here: a
	# DIRECTORY at this path swallows the record as .binary-meta/binary-meta,
	# where nothing reads it, while the notice below would claim provenance was
	# recorded. The install itself is genuine and is not undone for this — it is
	# reported instead, in the words the other write failure uses.
	meta_note=' (the .binary-meta provenance record could not be written because that path exists and is not a regular file, so version-skew reporting stays silent for this plugin root)'
else
	{
		printf 'release_tag=%s\n' "$release_tag"
		printf 'release_sha=%s\n' "$release_sha"
		printf 'binary_sha256=%s\n' "$binary_sha256"
		printf 'fetched_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
		printf 'plugin_sha=%s\n' "$plugin_sha"
		printf 'plugin_root_basename=%s\n' "$plugin_root_basename"
	} > "$tmp/binary-meta" 2>/dev/null &&
		mv -f "$tmp/binary-meta" "$meta_path" 2>/dev/null ||
		meta_note=' (the .binary-meta provenance record could not be written, so version-skew reporting stays silent for this plugin root)'
fi

# The one place PATH setup is suggested; the symlink itself stays owned by ahoy.
# This prints once per plugin root, because every later session takes the fast
# path above.
#
# The instruction names the binary by the ABSOLUTE path this script already
# holds, and does so for a reason worth stating: the sentence is addressed to a
# reader for whom `abcd` is not a name the shell can resolve — that is the whole
# condition it exists to fix — so an instruction reading `abcd ahoy install`
# fails with "command not found" for everyone who needs it. On the first manual
# install the agent that read it invented a `go run` incantation into the
# harness's plugin cache instead (iss-207). $binary is resolvable right now, by
# anyone, with no toolchain. Keep the invocation LAST on the line so it stays
# copy-pasteable, and keep the success leading the first word: only the first
# line of a hook's stderr reaches the transcript.
notice "$(printf 'abcd bootstrap: installed the checksum-verified abcd binary (release %s) into the plugin root, so the abcd hooks are live for this session.%s For the abcd command in your own terminal, run this once — the path is absolute because abcd is not on your PATH yet, which is exactly what the command fixes: "%s" ahoy install' \
	"$release_tag" "$meta_note" "$binary")"
