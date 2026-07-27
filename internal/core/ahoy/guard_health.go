package ahoy

import (
	"os"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/guard"
)

// guardHookCommand is the substring the PreToolUse manifest entry must contain
// for the execution-time guard to be armed. It mirrors requiredHookCommand's
// shape, but is checked separately and reported separately: a missing guard hook
// is a guard-health fact, not evidence that the plugin install is broken.
const guardHookCommand = "guard hook"

// GuardHealth is the answer to "is the execution-time guard actually on?", split
// into the three things that can independently be false. It exists because a
// guard that fails open is indistinguishable from a guard that is working, from
// inside the session — so the state has to be visible from outside it too
// (itd-103 AC 1, spc-16 "Fail-open-loud and health").
type GuardHealth struct {
	// HookInstalled reports whether the plugin's hook manifest declares the
	// pre-tool-use entry that runs the guard.
	HookInstalled bool `json:"hook_installed"`
	// BinaryReachable reports whether the binary that entry invokes exists and
	// is executable. When it is not, the shim fails open on every command.
	BinaryReachable bool `json:"binary_reachable"`
	// RegistryLoadable reports whether the bundled defaults merged with this
	// repo's .abcd/guard.json parse and validate. When they do not, the guard
	// declines to answer and every command runs unchecked.
	RegistryLoadable bool `json:"registry_loadable"`
	// Disabled reports a deliberately switched-off registry. It is not a fault —
	// the committed kill switch is the only sanctioned escape — but a disabled
	// guard that looks armed is exactly the state this report exists to prevent.
	Disabled bool `json:"disabled"`
	// Entries is how many hazards the loaded registry holds, so "loadable" is
	// backed by a number rather than a boolean nobody can check.
	Entries int `json:"entries"`
	// Detail is a one-line human reason when something is false; empty when the
	// guard is healthy.
	Detail string `json:"detail,omitempty"`
}

// Healthy reports whether commands in a session are actually being checked: the
// hook runs, the binary it calls exists, and the registry loads.
func (g GuardHealth) Healthy() bool {
	return g.HookInstalled && g.BinaryReachable && g.RegistryLoadable
}

// detectGuardHealth answers the three questions for one repo. It never returns an
// error: an unanswerable question is a false with a reason, which is the whole
// point — a health check that could itself fail silently would report nothing
// where it matters most.
func detectGuardHealth(cwd, pluginRoot string, pluginOK bool) GuardHealth {
	var h GuardHealth
	var reasons []string
	if !pluginOK {
		reasons = append(reasons, "plugin root not resolvable, so the hook manifest cannot be read")
	} else {
		h.HookInstalled = manifestArmsGuard(pluginRoot)
		h.BinaryReachable = isExecutableFile(pluginBinaryPath(pluginRoot))
		if !h.HookInstalled {
			reasons = append(reasons, "hooks/hooks.json declares no PreToolUse entry running `abcd guard hook`")
		}
		if !h.BinaryReachable {
			reasons = append(reasons, "the plugin-root abcd binary is missing or not executable, so the hook shim fails open on every command")
		}
	}

	if reg, err := guard.Load(cwd); err != nil {
		// The raw error can name a per-repo path, and the action a human takes is
		// the same whatever the parse failure was: `abcd guard check` prints it.
		reasons = append(reasons, guardRegistryUnloadableReason)
	} else {
		h.RegistryLoadable = true
		h.Disabled = reg.Disabled
		h.Entries = len(reg.Entries)
	}

	h.Detail = strings.Join(reasons, "; ")
	return h
}

// guardRegistryUnloadableReason is the one reason string for a registry that will
// not load, shared by the health detail and the gap so a reader is never told two
// different things about one fault.
const guardRegistryUnloadableReason = guard.RepoRelPath + " does not load, so the guard declines to answer and commands run unchecked"

// detectGuardGaps turns an unhealthy guard into gaps, so a broken guard shows up
// on the same status board as every other discrepancy rather than only in a line
// a reader might skip. They are diagnostic: ahoy does not write a repo's guard
// config, and a missing hook entry means a broken or stale plugin install.
func detectGuardGaps(h GuardHealth) []Gap {
	var gaps []Gap
	if !h.HookInstalled {
		gaps = append(gaps, Gap{
			ID: "guard.hook_missing", Category: PluginOwned, Scope: "machine",
			Title:   "shell-hazard guard not armed",
			Detail:  "No PreToolUse hook runs `abcd guard hook`, so dangerous shell commands are not checked before they execute.",
			FixHint: "Update or reinstall the abcd plugin.", Required: false, Resolvable: false,
		})
	}
	if h.HookInstalled && !h.BinaryReachable {
		gaps = append(gaps, Gap{
			ID: "guard.binary_unreachable", Category: PluginOwned, Scope: "machine",
			Title:   "guard hook installed but its binary is unreachable",
			Detail:  "The guard hook is wired but the abcd binary it calls is missing or not executable, so it fails open on every command.",
			FixHint: "Reinstall the abcd plugin, or run `abcd ahoy install`.", Required: false, Resolvable: false,
		})
	}
	if !h.RegistryLoadable {
		gaps = append(gaps, Gap{
			ID: "guard.registry_unloadable", Category: ConfigChange, Scope: "repo",
			Title:    "hazard registry does not load",
			Detail:   guardRegistryUnloadableReason,
			FixHint:  "Fix or remove " + guard.RepoRelPath + "; `abcd guard check --command ls` names the parse error.",
			Required: false, Resolvable: false,
		})
	}
	return gaps
}

// manifestArmsGuard reports whether the plugin's hook manifest declares a
// PreToolUse entry running the guard. It reuses the same guarded read and the
// same command-substring test as the prompt-router manifest check, so the two
// cannot disagree about what "installed" means.
func manifestArmsGuard(pluginRoot string) bool {
	hooks, ok := readHookEvents(pluginRoot)
	if !ok {
		return false
	}
	entries, ok := hooks["PreToolUse"].([]any)
	if !ok {
		return false
	}
	return eventHasCommand(entries, guardHookCommand)
}

// isExecutableFile reports whether path is a regular file with an execute bit —
// what the hook shim needs in order to run at all.
func isExecutableFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}
	return fi.Mode().Perm()&0o111 != 0
}
