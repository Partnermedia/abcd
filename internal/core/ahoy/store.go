package ahoy

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"github.com/REPPL/abcd-cli/internal/gitutil"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core"
)

// pluginVersion is the version ahoy stamps into config.json["meta"] and
// compares against on upgrade detection. It tracks the build-stamped core.
func pluginVersion() string { return core.Version }

// ---------------------------------------------------------------------------
// git identity helpers
// ---------------------------------------------------------------------------

// rootCommitSHA returns the repo's root-commit SHA, or "" when it cannot be
// derived (no git, no commits). Total: never errors out of band.
func rootCommitSHA(cwd string) string {
	out, err := runGit(cwd, "rev-list", "--max-parents=0", "HEAD")
	if err != nil {
		return ""
	}
	// A repo may have multiple root commits; the first is canonical.
	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// originURL returns the trimmed origin remote URL, or "" on any failure.
func originURL(cwd string) string {
	out, err := runGit(cwd, "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func runGit(cwd string, args ...string) (string, error) {
	full := append([]string{"-C", cwd}, args...)
	cmd := exec.Command("git", full...)
	// Isolate: an inherited GIT_DIR/GIT_WORK_TREE overrides `-C cwd` and answers
	// for a DIFFERENT repository, so the root-commit SHA and origin URL this feeds
	// into RepoIdentity — which keys the cross-repo ~/.abcd/history registry and
	// drives install/refounding decisions — would be silently registered against
	// the wrong repo. rev-list/remote do not need global config, so full isolation
	// is safe here.
	cmd.Env = gitutil.IsolatedEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// deriveIdentity builds the deterministic RepoIdentity for cwd.
func deriveIdentity(cwd string) RepoIdentity {
	return RepoIdentity{
		Name:    filepath.Base(cwd),
		Github:  originURL(cwd),
		RootSHA: rootCommitSHA(cwd),
	}
}

// ---------------------------------------------------------------------------
// plugin root resolution
// ---------------------------------------------------------------------------

// osExecutable is os.Executable, overridable so a test can drive the
// executable-ancestor fallback without re-execing the test binary.
var osExecutable = os.Executable

// resolvePluginRoot resolves the plugin root via ABCD_PLUGIN_ROOT ->
// CLAUDE_PLUGIN_ROOT -> executable-ancestor fallback, validating the expected
// layout for each candidate. Returns ("", false) when every candidate fails.
func resolvePluginRoot() (string, bool) {
	var candidates []string
	if v := os.Getenv("ABCD_PLUGIN_ROOT"); v != "" {
		candidates = append(candidates, v)
	}
	if v := os.Getenv("CLAUDE_PLUGIN_ROOT"); v != "" {
		candidates = append(candidates, v)
	}
	if exe, err := osExecutable(); err == nil {
		// Walk up looking for a directory with the plugin layout. Canonicalise
		// first: os.Executable reports the path abcd was INVOKED through, and
		// `ahoy install` pins a PATH symlink (/usr/local/bin/abcd ->
		// <plugin-root>/abcd), so an unresolved path walks the symlink's
		// ancestors and never reaches the plugin root (iss-170). resolvePath
		// falls back to an absolutised form, then the original path, on error.
		dir := filepath.Dir(resolvePath(exe))
		for i := 0; i < 6 && dir != "/" && dir != "."; i++ {
			candidates = append(candidates, dir)
			dir = filepath.Dir(dir)
		}
	}
	for _, c := range candidates {
		if pluginRootValid(c) {
			return c, true
		}
	}
	return "", false
}

// pluginRootValid sanity-checks a candidate by verifying the expected plugin
// layout (a hooks/ directory).
func pluginRootValid(candidate string) bool {
	return isDir(filepath.Join(candidate, "hooks"))
}

// pluginBinaryPath is the binary the owned PATH symlink points at.
func pluginBinaryPath(pluginRoot string) string {
	return filepath.Join(pluginRoot, "abcd")
}

// binTarget is the PATH symlink target, overridable for tests.
func binTarget() string {
	if v := os.Getenv("ABCD_BIN_TARGET"); v != "" {
		return v
	}
	return "/usr/local/bin/abcd"
}

// ---------------------------------------------------------------------------
// dev-mode shim (abcd ahoy install --dev)
// ---------------------------------------------------------------------------

// devShimMarker is the identifying first-comment line of the dev shim. Detection
// recognises the shim by this marker (never mistaking it for a foreign occupant),
// and uninstall removes only a file that carries it.
const devShimMarker = "# abcd-dev-shim"

// shSingleQuote wraps s for safe interpolation inside single quotes in a POSIX
// shell, so a source-repo path containing a quote cannot break out of the string.
func shSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// renderDevShim returns the POSIX-sh track-latest shim body. On every call it
// rebuilds abcd from sourceRepo's tip into freshBin and execs it; a failed build
// fails loudly and NEVER execs a stale binary (loud-staging). The absolute paths
// live only inside this installed file (outside any repo), never in a committed
// artefact.
func renderDevShim(sourceRepo, freshBin string) string {
	return "#!/bin/sh\n" +
		devShimMarker + " (managed by `abcd ahoy install --dev`)\n" +
		"# Rebuilds abcd from the source tip on every call, then execs the fresh\n" +
		"# binary. A broken build fails loudly and never execs a stale binary.\n" +
		"# Do not edit; run `abcd ahoy install` to pin a built binary instead.\n" +
		"ABCD_DEV_REPO=" + shSingleQuote(sourceRepo) + "\n" +
		"ABCD_DEV_BIN=" + shSingleQuote(freshBin) + "\n" +
		"if ! go build -C \"$ABCD_DEV_REPO\" -o \"$ABCD_DEV_BIN\" ./cmd/abcd; then\n" +
		"\tprintf 'abcd dev shim: build failed in %s — refusing to run a stale binary\\n' \"$ABCD_DEV_REPO\" >&2\n" +
		"\texit 1\n" +
		"fi\n" +
		"exec \"$ABCD_DEV_BIN\" \"$@\"\n"
}

// devShimPrefix is the exact opening of a shim renderDevShim produced: the
// shebang followed immediately by the marker line. Recognition requires this
// prefix — not a substring match anywhere in the file — so a foreign file that
// merely mentions the marker in its body is never mistaken for ours.
var devShimPrefix = []byte("#!/bin/sh\n" + devShimMarker)

// isDevShimFile reports whether path is a regular file we wrote as the dev shim.
// Uses the same guarded read as the rest of ahoy, so a symlink leaf or an
// oversized/irregular file reads as "not our shim" rather than being followed.
func isDevShimFile(path string) bool {
	data, err := fsutil.ReadGuarded(path, maxAhoyFileBytes)
	if err != nil {
		return false
	}
	return bytes.HasPrefix(data, devShimPrefix)
}

// binTargetKind classifies what currently occupies the PATH target.
type binTargetKind int

const (
	binTargetAbsent       binTargetKind = iota // nothing there
	binTargetOwnedSymlink                      // our symlink -> the pinned binary
	binTargetDevShim                           // our dev-mode rebuild-then-exec shim
	binTargetForeign                           // something else; never clobber it
)

// classifyBinTarget inspects the PATH target and reports whether it is absent,
// our owned pinned symlink, our dev shim, or a foreign occupant.
func classifyBinTarget(target, pluginRoot string) binTargetKind {
	fi, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return binTargetAbsent
		}
		return binTargetForeign // present but unstattable: treat as foreign, never clobber
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		dest, err := os.Readlink(target)
		if err != nil {
			return binTargetForeign
		}
		if resolveSymlinkDest(target, dest) == resolvePath(pluginBinaryPath(pluginRoot)) {
			return binTargetOwnedSymlink
		}
		return binTargetForeign
	}
	if isDevShimFile(target) {
		return binTargetDevShim
	}
	return binTargetForeign
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// ---------------------------------------------------------------------------
// ~/.abcd/history store
// ---------------------------------------------------------------------------

// historyRoot returns ~/.abcd/history. HOME is respected so tests can redirect.
func historyRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".abcd", "history"), nil
}

// historyIndex is the ~/.abcd/history/index.json registry.
type historyIndex struct {
	Schema      int           `json:"schema"`
	Description string        `json:"description"`
	Repos       []historyRepo `json:"repos"`
}

// historyRepo is one registry entry keyed on the immutable root_commit.
type historyRepo struct {
	RootCommit   string `json:"root_commit"`
	Name         string `json:"name"`
	Github       string `json:"github"`
	Path         string `json:"path"`
	Status       string `json:"status"`
	Supersedes   string `json:"supersedes,omitempty"`
	SupersededBy string `json:"superseded_by,omitempty"`
}

const historyIndexDescription = "abcd history/lifeboat registry. Keyed on each repo's root-commit SHA (immutable under rename, GitHub-handle change, or remote move). Names, GitHub URLs, and paths are mutable labels held in each repo's entry and refreshed by ahoy."

// loadHistoryIndex reads ~/.abcd/history/index.json. Returns (nil,nil) when the
// store is not bootstrapped yet.
func loadHistoryIndex() (*historyIndex, error) {
	root, err := historyRoot()
	if err != nil {
		return nil, err
	}
	data, err := fsutil.ReadGuarded(filepath.Join(root, "index.json"), maxAhoyFileBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var idx historyIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// indexHasRoot reports whether idx registers rootSHA.
func indexHasRoot(idx *historyIndex, rootSHA string) bool {
	if idx == nil || rootSHA == "" {
		return false
	}
	for _, r := range idx.Repos {
		if r.RootCommit == rootSHA {
			return true
		}
	}
	return false
}

// indexEntry returns the entry for rootSHA, or nil.
func indexEntry(idx *historyIndex, rootSHA string) *historyRepo {
	if idx == nil {
		return nil
	}
	for i := range idx.Repos {
		if idx.Repos[i].RootCommit == rootSHA {
			return &idx.Repos[i]
		}
	}
	return nil
}

// findRefoundingCandidate returns an entry with a matching name (or non-empty
// matching github) but a different root_commit, or nil.
func findRefoundingCandidate(idx *historyIndex, id RepoIdentity) *historyRepo {
	if idx == nil {
		return nil
	}
	for i := range idx.Repos {
		e := &idx.Repos[i]
		if e.RootCommit == id.RootSHA {
			continue
		}
		nameMatch := e.Name == id.Name
		githubMatch := id.Github != "" && e.Github != "" && e.Github == id.Github
		if nameMatch || githubMatch {
			return e
		}
	}
	return nil
}

// historyLockFilename is the ~/.abcd/history lock file guarding the index.json
// load-modify-write. It sits beside index.json.
const historyLockFilename = ".index.lock"

// historyLockTimeout bounds acquisition of the history-index lock. A var (not
// const) so a test can shorten it to exercise contention without a multi-second
// wait, mirroring capture's lockTimeout.
var historyLockTimeout = 5 * time.Second

// afterHistoryReloadHook, when non-nil, fires inside withHistoryLock immediately
// after the lock is acquired and before fn runs (fn re-loads the index under the
// lock and writes it). It is a test-only seam (nil in production) used to force
// the iss-101 lost-update interleaving deterministically: the goroutine that grabs
// the lock pauses here, holding it, while another registration is admitted.
var afterHistoryReloadHook func()

// beforeHistoryIndexCreateHook, when non-nil, fires in bootstrapHistory just
// before the atomic link that publishes index.json. Test-only seam (nil in
// production) to force the iss-101 bootstrap check-then-act interleaving
// deterministically.
var beforeHistoryIndexCreateHook func()

// withHistoryLock runs fn while holding the exclusive ~/.abcd/history lock, so a
// registerRepo load-modify-write of index.json serializes across concurrent
// `abcd ahoy install` runs from different worktrees (iss-101). It routes through
// the shared fsutil.WithFileLock primitive; the lock file lives beside
// index.json. The lock is NEVER held across an interactive prompt — callers do
// any prompting before acquiring it and re-check the answer-relevant state inside
// fn after re-loading.
func withHistoryLock(fn func() error) error {
	root, err := historyRoot()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	lockPath := filepath.Join(root, historyLockFilename)
	return fsutil.WithFileLock(lockPath, historyLockTimeout, func() error {
		if afterHistoryReloadHook != nil {
			afterHistoryReloadHook()
		}
		return fn()
	})
}

// bootstrapHistory creates ~/.abcd/history/ + index.json when absent. Idempotent.
// The seed is published atomically (iss-101): a fully-formed temp file is written,
// synced, and os.Link'd into place. The link is the single-winner publish — it
// fails EEXIST if index.json already exists, so under two concurrent bootstraps
// exactly one wins and every other observes the existing file and reports no
// write. Because the file only ever appears complete (the link, never a bare
// create followed by a separate content write), a reader that races the bootstrap
// sees either no file yet or the finished index — never a 0-byte one that would
// make a concurrent loadHistoryIndex parse-fail and drop its own registration.
func bootstrapHistory() (bool, error) {
	root, err := historyRoot()
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return false, err
	}
	path := filepath.Join(root, "index.json")
	// Fast path: already seeded (the common idempotent re-run).
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	idx := historyIndex{Schema: 1, Description: historyIndexDescription, Repos: []historyRepo{}}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return false, err
	}
	data = append(data, '\n')

	// Write a complete temp file first, then link it into place as the atomic,
	// single-winner publish. The temp is always cleaned up.
	tmp, err := os.CreateTemp(root, ".index-*.tmp")
	if err != nil {
		return false, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return false, err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return false, err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return false, err
	}
	if err := tmp.Close(); err != nil {
		return false, err
	}

	if beforeHistoryIndexCreateHook != nil {
		beforeHistoryIndexCreateHook()
	}
	if err := os.Link(tmpName, path); err != nil {
		if os.IsExist(err) {
			return false, nil // already seeded (or won by a concurrent run)
		}
		return false, err
	}
	return true, nil
}

// writeHistoryIndex persists idx. It must be called from inside withHistoryLock
// (registerRepo) so a re-loaded index is not clobbered by a concurrent writer.
func writeHistoryIndex(idx *historyIndex) error {
	root, err := historyRoot()
	if err != nil {
		return err
	}
	return writeJSON(filepath.Join(root, "index.json"), *idx)
}

// ---------------------------------------------------------------------------
// repo config.json
// ---------------------------------------------------------------------------

// maxAhoyFileBytes caps ahoy's marker/config/index reads. Markers (CLAUDE.md /
// AGENTS.md), config.json, and the history index are small text/JSON; 1 MiB is
// far above any real one, so the cap never rejects a legitimate file while it
// bounds a FIFO or multi-GB file planted at a cwd-influenced path (iss-97).
const maxAhoyFileBytes = 1 << 20

// configPath is the repo-scope .abcd/config.json path.
func configPath(cwd string) string {
	return filepath.Join(cwd, ".abcd", "config.json")
}

// readConfig reads .abcd/config.json into a generic map. Returns (nil,nil) when
// absent, and an error on malformed JSON or a guarded-read refusal (a non-regular
// leaf such as a FIFO, or a file past the size cap). Callers treat any error as
// unusable config and fail safe.
func readConfig(cwd string) (map[string]any, error) {
	data, err := fsutil.ReadGuarded(configPath(cwd), maxAhoyFileBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// subMap returns cfg[name] as a map, or an empty map.
func subMap(cfg map[string]any, name string) map[string]any {
	if cfg == nil {
		return map[string]any{}
	}
	if v, ok := cfg[name].(map[string]any); ok {
		return v
	}
	return map[string]any{}
}

// writeJSON marshals v deterministically (map keys sorted) with a trailing
// newline via the atomic writer.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return fsutil.WriteFileAtomicPreserveMode(path, data)
}

// writeConfig persists a config map deterministically.
func writeConfig(cwd string, cfg map[string]any) error {
	return writeJSON(configPath(cwd), cfg)
}

// ---------------------------------------------------------------------------
// hook manifest verification (verify-only)
// ---------------------------------------------------------------------------

// requiredHookCommand is the substring each event's command must contain. These
// are the Go hook subcommands (`abcd hook prompt-router` / `prompt-router-reset`)
// as wired in hooks/hooks.json — the loader is a Go subcommand, not a script.
var requiredHookCommand = map[string]string{
	"UserPromptSubmit": "hook prompt-router",
	"SessionStart":     "hook prompt-router-reset",
	"PreCompact":       "hook prompt-router-reset",
}

// verifyHookManifest returns "" when hooks/hooks.json under pluginRoot is
// structurally sound, else a one-line reason. Read-only; never mutates.
func verifyHookManifest(pluginRoot string) string {
	path := filepath.Join(pluginRoot, "hooks", "hooks.json")
	// Keep the symlink pre-check: it yields a DISTINCT reason that ReadGuarded's
	// O_NOFOLLOW would otherwise collapse into a raw ELOOP -> "read failed".
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return "leaf is a symlink (refusing to follow)"
	}
	// ReadGuarded validates regular-file + size cap on the SAME descriptor it
	// reads, closing the Lstat->ReadFile swap window (iss-109). Map its sentinels
	// back to the existing reason strings, unchanged.
	data, err := fsutil.ReadGuarded(path, 256*1024)
	if err != nil {
		switch {
		case errors.Is(err, fsutil.ErrNotRegular):
			return "not a regular file"
		case errors.Is(err, fsutil.ErrTooBig):
			return "file size exceeds 256KB cap"
		case os.IsNotExist(err):
			return "file absent"
		default:
			return "read failed"
		}
	}
	hooks, reason := decodeHookEvents(data)
	if reason != "" {
		return reason
	}
	for _, event := range []string{"UserPromptSubmit", "SessionStart", "PreCompact"} {
		entries, ok := hooks[event].([]any)
		if !ok || len(entries) == 0 {
			return "missing or empty `hooks." + event + "` array"
		}
		if !eventHasCommand(entries, requiredHookCommand[event]) {
			return "`hooks." + event + "` does not reference " + requiredHookCommand[event]
		}
	}
	return ""
}

// decodeHookEvents decodes a hook manifest's `hooks` object, returning the
// per-event entries or a one-line reason. It is the single decoder for the
// manifest: verifyHookManifest checks the prompt-router events with it, and the
// guard-health check reads the PreToolUse event with it, so the two can never
// disagree about what a well-formed manifest is.
func decodeHookEvents(data []byte) (map[string]any, string) {
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, "JSON parse failed"
	}
	hooks, ok := parsed["hooks"].(map[string]any)
	if !ok {
		return nil, "missing or non-object `hooks` key"
	}
	return hooks, ""
}

// readHookEvents reads and decodes the plugin's hook manifest, returning the
// per-event entries. A manifest that cannot be read or decoded yields ok=false —
// the caller reports "not installed", which is the honest answer when nothing can
// be proven about the wiring.
func readHookEvents(pluginRoot string) (map[string]any, bool) {
	data, err := fsutil.ReadGuarded(filepath.Join(pluginRoot, "hooks", "hooks.json"), 256*1024)
	if err != nil {
		return nil, false
	}
	hooks, reason := decodeHookEvents(data)
	return hooks, reason == ""
}

// eventHasCommand reports whether any nested command string contains substring.
func eventHasCommand(entries []any, substring string) bool {
	for _, entry := range entries {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		inner, ok := m["hooks"].([]any)
		if !ok {
			continue
		}
		for _, h := range inner {
			hm, ok := h.(map[string]any)
			if !ok {
				continue
			}
			if cmd, ok := hm["command"].(string); ok && strings.Contains(cmd, substring) {
				return true
			}
		}
	}
	return false
}
