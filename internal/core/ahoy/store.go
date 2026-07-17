package ahoy

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

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
	if exe, err := os.Executable(); err == nil {
		// Walk up looking for a directory with the plugin layout.
		dir := filepath.Dir(exe)
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
	data, err := os.ReadFile(filepath.Join(root, "index.json"))
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

// bootstrapHistory creates ~/.abcd/history/ + index.json when absent. Idempotent.
func bootstrapHistory() (bool, error) {
	root, err := historyRoot()
	if err != nil {
		return false, err
	}
	if isDir(root) {
		if _, err := os.Stat(filepath.Join(root, "index.json")); err == nil {
			return false, nil
		}
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return false, err
	}
	idx := historyIndex{Schema: 1, Description: historyIndexDescription, Repos: []historyRepo{}}
	return true, writeJSON(filepath.Join(root, "index.json"), idx)
}

// writeHistoryIndex persists idx.
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

// configPath is the repo-scope .abcd/config.json path.
func configPath(cwd string) string {
	return filepath.Join(cwd, ".abcd", "config.json")
}

// readConfig reads .abcd/config.json into a generic map. Returns (nil,nil) when
// absent, and an error only on malformed JSON.
func readConfig(cwd string) (map[string]any, error) {
	data, err := os.ReadFile(configPath(cwd))
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
	fi, err := os.Lstat(path)
	if err != nil {
		return "file absent"
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return "leaf is a symlink (refusing to follow)"
	}
	if !fi.Mode().IsRegular() {
		return "not a regular file"
	}
	if fi.Size() > 256*1024 {
		return "file size exceeds 256KB cap"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "read failed"
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "JSON parse failed"
	}
	hooks, ok := parsed["hooks"].(map[string]any)
	if !ok {
		return "missing or non-object `hooks` key"
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
