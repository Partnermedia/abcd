package ahoy

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"sort"

	"github.com/REPPL/abcd-cli/internal/core/identity"
)

// Enumerations for config-value validation.
var (
	visibilityChoices    = []string{"private", "public"}
	docsTargetChoices    = []string{"claude_md", "agents_md", "both", "skip"}
	oracleBackendChoices = []string{"host-delegated", "native", "cli", "api", "mcp"}
)

const (
	docsTargetDefault    = "both"
	oracleBackendDefault = "host-delegated"
)

// Detect runs the full detection pass over cwd and returns the canonical
// envelope. Total over a normal folder: broken-plugin conditions surface as
// gaps, never as a hard error. An error is returned only for malformed input.
func Detect(cwd string) (DetectionResult, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return DetectionResult{}, err
	}

	// Step 4 (pre-computed): identity, once per pass.
	identity := deriveIdentity(abs)

	// Step 0: folder classification.
	idx, _ := loadHistoryIndex()
	kind, signals := classify(abs, identity, idx)

	// Step 1: plugin root.
	pluginRoot, pluginOK := resolvePluginRoot()
	pluginStatus := "missing"
	if pluginOK {
		pluginStatus = "resolved"
	}

	res := DetectionResult{
		FolderKind:       kind,
		RootSHA:          identity.RootSHA,
		PluginRootStatus: pluginStatus,
		RepoIdentity:     identity,
		Signals:          signals,
		pluginRoot:       pluginRoot,
	}

	var gaps []Gap
	gaps = append(gaps, detectPluginRoot(pluginOK)...)
	if kind != UnmanagedFolder {
		gaps = append(gaps, detectDependencies()...)
		gaps = append(gaps, detectSkeleton(abs)...)
		gaps = append(gaps, detectIdentity(identity, idx)...)
		gaps = append(gaps, detectGitIdentity(abs)...)
		gaps = append(gaps, detectHistoryStore(identity.RootSHA)...)
		gaps = append(gaps, detectConfigValues(abs)...)
		gaps = append(gaps, detectMarkerDrift(abs)...)
		gaps = append(gaps, detectPathSymlink(pluginRoot, pluginOK)...)
		gaps = append(gaps, detectHookManifest(pluginRoot, pluginOK)...)
		gaps = append(gaps, detectVersion(abs)...)
	}

	sortGaps(gaps)
	res.Gaps = gaps
	return res, nil
}

// DryRun runs Detect and returns the envelope (Adopted=nil). Zero writes.
func DryRun(cwd string) (DetectionResult, error) { return Detect(cwd) }

// classify keys on the signal hierarchy (brief step 0 / itd-40).
func classify(cwd string, id RepoIdentity, idx *historyIndex) (FolderKind, map[string]any) {
	signals := map[string]any{}

	registered := indexHasRoot(idx, id.RootSHA)
	signals["index_registered"] = registered

	abcdDir := fsutil.IsRealDir(filepath.Join(cwd, ".abcd"))
	signals["abcd_dir"] = abcdDir

	markerFired := false
	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		if markerFileHasBlock(filepath.Join(cwd, name)) {
			markerFired = true
			break
		}
	}
	signals["marker_block"] = markerFired

	gitRepo := isDir(filepath.Join(cwd, ".git"))
	signals["git_repo"] = gitRepo

	// A bare .abcd/ directory is not a managed signal on its own (iss-88): only
	// index registration or a marker block promotes a folder to managed-repo.
	strong := registered || markerFired
	switch {
	case strong:
		return ManagedRepo, signals
	case gitRepo:
		return UnmanagedRepo, signals
	default:
		return UnmanagedFolder, signals
	}
}

func detectPluginRoot(ok bool) []Gap {
	if ok {
		return nil
	}
	return []Gap{{
		ID:       "plugin.root_missing",
		Category: PluginOwned,
		Scope:    "machine",
		Title:    "plugin root not resolvable",
		Detail:   "ABCD_PLUGIN_ROOT and CLAUDE_PLUGIN_ROOT are unset and the fallback found no plugin layout.",
		FixHint:  "Reinstall the abcd plugin, or set ABCD_PLUGIN_ROOT.",
	}}
}

func detectDependencies() []Gap {
	var gaps []Gap
	if !onPath("gitleaks") {
		gaps = append(gaps, Gap{
			ID: "deps.gitleaks_missing", Category: Dependency, Scope: "machine",
			Title: "gitleaks not on PATH", Detail: "gitleaks enables a deeper secret scan.",
			FixHint: "brew install gitleaks", Required: false, Resolvable: true,
		})
	}
	if !onPath("trufflehog") {
		gaps = append(gaps, Gap{
			ID: "deps.trufflehog_missing", Category: Dependency, Scope: "machine",
			Title: "trufflehog not on PATH", Detail: "trufflehog enables deep secret scanning when scan.deep=true.",
			FixHint: "brew install trufflehog", Required: false, Resolvable: true,
		})
	}
	return gaps
}

func onPath(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func detectSkeleton(cwd string) []Gap {
	var gaps []Gap
	if !fileExists(filepath.Join(cwd, ".abcd", "config.json")) {
		gaps = append(gaps, Gap{
			ID: "skeleton.config_missing", Category: SafeAutocreate, Scope: "repo",
			Title: ".abcd/config.json missing", Detail: "Repo requires a configuration file at .abcd/config.json.",
			FixHint: "ahoy install creates it from defaults.", Required: true, Resolvable: true,
		})
	}
	if !fileExists(filepath.Join(cwd, ".abcd", "rules.json")) {
		gaps = append(gaps, Gap{
			ID: "rules.missing", Category: SafeAutocreate, Scope: "repo",
			Title: ".abcd/rules.json missing", Detail: "Rules skeleton at .abcd/rules.json is absent.",
			FixHint: "ahoy install writes the bundled default rules.json.", Required: true, Resolvable: true,
		})
	}
	return gaps
}

func detectIdentity(id RepoIdentity, idx *historyIndex) []Gap {
	if id.RootSHA == "" {
		return nil
	}
	if idx == nil || indexHasRoot(idx, id.RootSHA) {
		return nil
	}
	gaps := []Gap{{
		ID: "identity.unregistered", Category: UserState, Scope: "repo",
		Title:   "root SHA not in history index",
		Detail:  "Root commit " + shortSHA(id.RootSHA) + " is absent from ~/.abcd/history/index.json.",
		FixHint: "ahoy install registers the repo entry.", Required: true, Resolvable: true,
	}}
	if cand := findRefoundingCandidate(idx, id); cand != nil {
		gaps = append(gaps, Gap{
			ID: "identity.refounding_candidate", Category: UserState, Scope: "repo",
			Title:   "possible re-founded repo",
			Detail:  "A sibling entry (root " + shortSHA(cand.RootCommit) + ") matches name/github. Confirm re-founding or register as new.",
			FixHint: "ahoy install asks; declining registers as new.", Required: true, Resolvable: true,
		})
	}
	return gaps
}

// detectGitIdentity compares the git author identity a commit would use against
// the committed .abcd/config/identity.json pin (iss-62). A mismatch or an unset
// identity is a required, resolvable gap; an un-pinned repo gets an advisory gap
// (adopt the gate); a match yields nothing.
func detectGitIdentity(cwd string) []Gap {
	res, err := identity.Check(cwd)
	if err != nil {
		// A present-but-unreadable pin is an adopted-but-broken gate (required);
		// an error with no pin is a git/environment issue (advisory).
		_, statErr := os.Stat(filepath.Join(cwd, identity.PinRelPath))
		pinPresent := statErr == nil
		return []Gap{{
			ID: "git_identity.uncheckable", Category: ConfigChange, Scope: "repo",
			Title:      "git identity could not be checked",
			Detail:     err.Error(),
			FixHint:    "fix the pin JSON in " + identity.PinRelPath + " (both name and email), or ensure git is available",
			Required:   pinPresent,
			Resolvable: false,
		}}
	}
	switch res.Status {
	case identity.StatusMismatch:
		return []Gap{{
			ID: "git_identity.mismatch", Category: ConfigChange, Scope: "repo",
			Title:      "git commit identity does not match the pin",
			Detail:     res.Reason,
			FixHint:    "set this repo's git user.name/user.email to match the pin in " + identity.PinRelPath + " (or update the pin if the identity changed)",
			Required:   true,
			Resolvable: true,
		}}
	case identity.StatusUnset:
		return []Gap{{
			ID: "git_identity.unset", Category: ConfigChange, Scope: "repo",
			Title:      "git author identity is not configured",
			Detail:     res.Reason,
			FixHint:    "set git user.name/user.email to the pinned identity in " + identity.PinRelPath,
			Required:   true,
			Resolvable: true,
		}}
	case identity.StatusNoPin:
		return []Gap{{
			ID: "git_identity.unpinned", Category: ConfigChange, Scope: "repo",
			Title:      "no git identity pin",
			Detail:     res.Reason,
			FixHint:    "ahoy install can pin the current git identity to " + identity.PinRelPath,
			Required:   false,
			Resolvable: true,
		}}
	}
	return nil
}

func detectHistoryStore(rootSHA string) []Gap {
	var gaps []Gap
	root, err := historyRoot()
	if err != nil {
		return nil
	}
	if !isDir(root) {
		gaps = append(gaps, Gap{
			ID: "history.bootstrap_missing", Category: UserState, Scope: "machine",
			Title: "~/.abcd/history/ not bootstrapped", Detail: "The shared history store directory is absent.",
			FixHint: "ahoy install bootstraps it.", Required: true, Resolvable: true,
		})
	}
	if rootSHA == "" {
		return gaps
	}
	repoDir := filepath.Join(root, rootSHA)
	if !fsutil.IsRealDir(filepath.Join(repoDir, "transcripts")) {
		gaps = append(gaps, Gap{
			ID: "history.transcripts_missing", Category: SafeAutocreate, Scope: "repo",
			Title:   "history transcripts/ dir missing",
			Detail:  "~/.abcd/history/" + shortSHA(rootSHA) + "/transcripts/ is absent or not a real directory.",
			FixHint: "ahoy install creates the transcript directory.", Required: true, Resolvable: true,
		})
	}
	if !fileExists(filepath.Join(repoDir, "meta.json")) {
		gaps = append(gaps, Gap{
			ID: "history.meta_missing", Category: UserState, Scope: "repo",
			Title:   "history meta.json missing",
			Detail:  "~/.abcd/history/" + shortSHA(rootSHA) + "/meta.json is absent.",
			FixHint: "ahoy install writes the per-repo meta.json.", Required: true, Resolvable: true,
		})
	}
	return gaps
}

func detectConfigValues(cwd string) []Gap {
	var gaps []Gap
	cfg, err := readConfig(cwd)
	if err != nil {
		cfg = nil // malformed config is treated as absent for value checks
	}
	repo := subMap(cfg, "repo")
	docs := subMap(cfg, "docs")
	oracle := subMap(cfg, "oracle")
	scan := subMap(cfg, "scan")

	visibility, visOK := stringVal(repo, "visibility")
	if !visOK || !inSet(visibility, visibilityChoices) {
		gaps = append(gaps, cfgGap("config.visibility_missing", "repo.visibility not set",
			"Visibility (private/public) controls .gitignore policy."))
	}
	if v, ok := stringVal(docs, "target"); !ok || !inSet(v, docsTargetChoices) {
		gaps = append(gaps, cfgGap("config.docs_target_missing", "docs.target not set",
			"Which docs file (CLAUDE.md / AGENTS.md / both / skip) hosts the marker block."))
	}
	if v, ok := stringVal(oracle, "backend"); !ok || !inSet(v, oracleBackendChoices) {
		gaps = append(gaps, cfgGap("config.oracle_backend_missing", "oracle.backend not set",
			"Oracle backend (host-delegated/native/cli/api/mcp)."))
	}

	validVis := ""
	if visOK && inSet(visibility, visibilityChoices) {
		validVis = visibility
	}
	// scan.deep is conditional: private + trufflehog present.
	if validVis == "private" && onPath("trufflehog") {
		if _, ok := boolVal(scan, "deep"); !ok {
			gaps = append(gaps, cfgGap("config.scan_deep_missing", "scan.deep not set",
				"Private repo + trufflehog present — confirm deep secret scanning."))
		}
	}
	// visibility.gitignore_drift only when a valid visibility is persisted.
	if validVis == "private" || validVis == "public" {
		if gitignoreBlockDrifts(cwd, validVis) {
			gaps = append(gaps, cfgGap("visibility.gitignore_drift",
				"abcd-managed .gitignore block drifts from visibility",
				"The .gitignore abcd block does not match the canonical rules for visibility="+validVis+"."))
		}
	}
	return gaps
}

func cfgGap(id, title, detail string) Gap {
	return Gap{ID: id, Category: ConfigChange, Scope: "repo", Title: title, Detail: detail,
		FixHint: "ahoy install prompts for the value.", Required: true, Resolvable: true}
}

func detectMarkerDrift(cwd string) []Gap {
	cfg, _ := readConfig(cwd)
	docs := subMap(cfg, "docs")
	target, _ := stringVal(docs, "target")
	files := markerTargets(target)
	var gaps []Gap
	for _, name := range files {
		switch classifyMarker(filepath.Join(cwd, name)) {
		case markerMissing:
			gaps = append(gaps, Gap{
				ID: "marker.missing", Category: PluginOwned, Scope: "repo",
				Title: name + " marker block missing", Detail: name + " has no <!-- BEGIN ABCD --> block.",
				FixHint: "ahoy install plants the canonical marker block.", Required: true, Resolvable: true,
			})
		case markerOutdated:
			gaps = append(gaps, Gap{
				ID: "marker.outdated", Category: PluginOwned, Scope: "repo",
				Title: name + " marker block outdated", Detail: name + " marker block differs from the template.",
				FixHint: "ahoy install rewrites it to canonical (silent overwrite).", Required: true, Resolvable: true,
			})
		}
	}
	return gaps
}

func detectPathSymlink(pluginRoot string, pluginOK bool) []Gap {
	if !pluginOK {
		return nil
	}
	target := binTarget()
	expected := pluginBinaryPath(pluginRoot)
	fi, err := lstat(target)
	if err != nil {
		if isNotExist(err) {
			return []Gap{{
				ID: "symlink.missing", Category: ConfigChange, Scope: "machine",
				Title: "PATH symlink not installed", Detail: target + " does not exist.",
				FixHint: "ahoy install creates the symlink (refuses to clobber).", Required: true, Resolvable: true,
			}}
		}
		return nil
	}
	if fi.Mode()&modeSymlink == 0 {
		return []Gap{{
			ID: "symlink.foreign", Category: ConfigChange, Scope: "machine",
			Title: "non-symlink at " + target, Detail: "A regular file occupies the PATH symlink target.",
			FixHint: "Resolve manually; ahoy refuses to clobber.", Required: false, Resolvable: false,
		}}
	}
	dest, err := readlink(target)
	if err != nil {
		return nil
	}
	if resolveSymlinkDest(target, dest) == resolvePath(expected) {
		return nil
	}
	return []Gap{{
		ID: "symlink.foreign", Category: ConfigChange, Scope: "machine",
		Title: "foreign symlink at " + target, Detail: target + " -> " + dest + " (expected " + expected + ").",
		FixHint: "Resolve manually; ahoy refuses to clobber.", Required: false, Resolvable: false,
	}}
}

func detectHookManifest(pluginRoot string, pluginOK bool) []Gap {
	if !pluginOK {
		return nil
	}
	reason := verifyHookManifest(pluginRoot)
	if reason == "" {
		return nil
	}
	return []Gap{{
		ID: "hooks.manifest_missing", Category: PluginOwned, Scope: "machine",
		Title: "hooks/hooks.json missing or malformed", Detail: reason,
		FixHint: "Broken plugin install — reinstall.", Required: false, Resolvable: false,
	}}
}

func detectVersion(cwd string) []Gap {
	cfg, _ := readConfig(cwd)
	meta := subMap(cfg, "meta")
	setupVersion, hasVersion := stringVal(meta, "setup_version")
	setupDate, hasDate := stringVal(meta, "setup_date")
	if !hasVersion || !hasDate || setupVersion == "" || setupDate == "" {
		return []Gap{{
			ID: "install_meta.missing", Category: SafeAutocreate, Scope: "repo",
			Title: "setup metadata absent", Detail: "setup_version / setup_date not recorded; first install.",
			FixHint: "ahoy install stamps the meta block.", Required: true, Resolvable: true,
		}}
	}
	if current := pluginVersion(); current != "" && setupVersion != current {
		return []Gap{{
			ID: "version.upgrade", Category: SafeAutocreate, Scope: "repo",
			Title:   "plugin upgrade " + setupVersion + " -> " + current,
			Detail:  "Recorded setup_version differs from the current plugin version.",
			FixHint: "ahoy install re-stamps setup_version + setup_date.", Required: true, Resolvable: true,
		}}
	}
	return nil
}

// ---------------------------------------------------------------------------
// small helpers
// ---------------------------------------------------------------------------

func sortGaps(gaps []Gap) {
	sort.SliceStable(gaps, func(i, j int) bool {
		if gaps[i].Category != gaps[j].Category {
			return gaps[i].Category < gaps[j].Category
		}
		if gaps[i].ID != gaps[j].ID {
			return gaps[i].ID < gaps[j].ID
		}
		return gaps[i].Scope < gaps[j].Scope
	})
}

func shortSHA(s string) string {
	if len(s) > 12 {
		return s[:12] + "…"
	}
	return s
}

func stringVal(m map[string]any, key string) (string, bool) {
	v, ok := m[key].(string)
	return v, ok
}

func boolVal(m map[string]any, key string) (bool, bool) {
	v, ok := m[key].(bool)
	return v, ok
}

func inSet(v string, set []string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}
