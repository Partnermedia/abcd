package ahoy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Partnermedia/abcd/internal/fsutil"
	"sort"
	"strings"
	"time"

	"github.com/Partnermedia/abcd/internal/core/identity"
)

// Install runs detect + apply over the approved categories. It is idempotent:
// a re-run with zero required+resolvable gaps writes nothing and reports
// "already_up_to_date".
func Install(cwd string, opts InstallOptions, p Prompter) (InstallResult, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return InstallResult{}, err
	}
	if p == nil {
		p = RefusingPrompter{}
	}

	det, err := Detect(abs)
	if err != nil {
		return InstallResult{}, err
	}

	// Unmanaged folder: nothing to act on.
	if det.FolderKind == UnmanagedFolder {
		return InstallResult{Status: "aborted"}, nil
	}

	// Adoption gate for an unmanaged repo.
	adopted := false
	if det.FolderKind == UnmanagedRepo {
		switch {
		case opts.Adopt != nil && !*opts.Adopt:
			return InstallResult{Status: "aborted"}, nil
		case opts.Adopt != nil && *opts.Adopt:
			adopted = true
		default:
			if !p.Confirm("Adopt this unmanaged repo into abcd?") {
				return InstallResult{Status: "aborted"}, nil
			}
			adopted = true
		}
	}
	_ = adopted

	// Where the PATH entry goes, decided BEFORE any write but AFTER the adoption
	// gate: an explicit --bin-dir abcd cannot write to fails the whole install
	// loudly rather than being discovered halfway through, and abcd never re-runs
	// itself with privilege. Deciding it after adoption keeps the writability
	// probe — which creates and removes a temp file — out of a run the user
	// declines.
	binTargetPath, err := resolveInstallTarget(opts, det.pluginRoot)
	if err != nil {
		return InstallResult{}, err
	}

	// Idempotency: zero required+resolvable gaps => exact no-op. Two exceptions
	// fall through: the advisory git-identity pin, which install adopts against a
	// confirmed answer (never under --yes), as the gap's fix hint
	// advertises; and an explicit value override that differs from the persisted
	// config, which forces an apply-as-update on an otherwise-clean repo (iss-107).
	// An explicit --dev (or a plain install over an existing dev shim) forces an
	// apply-as-update on an otherwise-clean repo, the same way an explicit value
	// override does (iss-107): the requested install mode differs from what is on
	// disk, so there is work to do even with zero gaps.
	modeForced := modeWouldChange(opts, det, binTargetPath)

	if len(actionable(det.Gaps)) == 0 &&
		!(!opts.Yes && pinAdoptable(det.Gaps)) &&
		!overridesWouldChange(abs, opts.ValueOverrides) &&
		!modeForced {
		return InstallResult{
			Status:          "already_up_to_date",
			OptionalSkipped: optionalSkipped(opts, det.Gaps),
		}, nil
	}

	approved, declined := resolveApproval(det.Gaps, opts, p)

	ac := &applyCtx{
		cwd:        abs,
		det:        det,
		approved:   approved,
		overrides:  opts.ValueOverrides,
		prompter:   p,
		gapPresent: gapIDSet(det.Gaps),
		autoYes:    opts.Yes,
		devMode:    opts.Dev,
		modeForced: modeForced,
		binTarget:  binTargetPath,
	}

	// itd-111 refusal: a binary that is stale against its own source tip, or
	// whose vintage cannot be determined, must not run stale install logic
	// against the machine — the trap the evidence session (iss-228) fell into.
	// The check is disk-only; an explicit --allow-stale-binary override proceeds.
	// It sits before the first apply step so nothing is written on a refusal.
	if !opts.AllowStaleBinary {
		if reason := staleBinaryRefusal(currentVintage(), abs); reason != "" {
			ac.refuse(reason)
			return InstallResult{Status: "refused", Notes: ac.notes}, nil
		}
	}

	// Ordered apply steps.
	ac.stepDependencies()
	ac.stepSkeleton()
	cfg := ac.stepConfigValues()
	ac.stepVisibility(cfg)
	// After stepVisibility, never before: the private stub is only written once the
	// .gitignore fence that keeps it untracked is on disk.
	ac.stepBanlist()
	ac.stepHistory()
	ac.stepMarker(cfg)
	ac.stepSymlink()
	ac.stepRules()
	ac.stepVersionStamp()
	ac.stepIdentityPin()
	// Last, because it describes the entry the steps above actually wrote.
	ac.noteReachability()

	// Re-detect to compute what remains.
	final, err := Detect(abs)
	if err != nil {
		return InstallResult{}, err
	}
	remaining := gapIDs(actionable(final.Gaps))

	status := "clean"
	if len(remaining) > 0 {
		status = "partial"
	}
	return InstallResult{
		Status:             status,
		Writes:             ac.writes,
		Changes:            ac.changes,
		Remaining:          remaining,
		DeclinedCategories: declined,
		Notes:              ac.notes,
		OptionalSkipped:    optionalSkipped(opts, final.Gaps),
	}, nil
}

// resolveInstallTarget decides where the PATH entry goes, before any write.
//
// The default is the field-standard single-user location (~/.local/bin), created
// when absent; an abcd-owned entry already on PATH is adopted exactly where it
// stands, so a second install is never planted beside a working one. A
// system-wide directory is reachable only through an explicit --bin-dir, and an
// unwritable one is an error: abcd never escalates privileges, so there is
// nothing to fall back to and pretending otherwise would install nothing while
// reporting success.
func resolveInstallTarget(opts InstallOptions, pluginRoot string) (string, error) {
	if opts.BinDir == "" {
		return adoptedBinTarget(pluginRoot), nil
	}
	dir, err := filepath.Abs(opts.BinDir)
	if err != nil {
		return "", fmt.Errorf("abcd ahoy install: --bin-dir %s is not a usable path: %w", opts.BinDir, err)
	}
	// Probe writability without creating anything: the directory is created by
	// the apply step, under the same approval as every other write.
	probe, existing := dir, false
	if fi, serr := os.Stat(dir); serr == nil {
		if !fi.IsDir() {
			return "", fmt.Errorf("abcd ahoy install: --bin-dir %s is not a directory", displayPath(dir))
		}
		existing = true
	} else {
		probe = nearestExistingDir(dir)
	}
	if !dirWritable(probe) {
		detail := "it could not be created — " + displayPath(probe) + " is not writable"
		if existing {
			detail = "no file can be created in it"
		}
		return "", fmt.Errorf("abcd ahoy install: --bin-dir %s is not writable (%s). abcd never escalates privileges — name a directory you own, or omit --bin-dir to install into ~/.local/bin", displayPath(dir), detail)
	}
	return filepath.Join(dir, binName), nil
}

// nearestExistingDir walks up from dir to the first directory that exists — the
// one whose writability decides whether dir can be created at all.
func nearestExistingDir(dir string) string {
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		if fi, err := os.Stat(parent); err == nil && fi.IsDir() {
			return parent
		}
		dir = parent
	}
}

// adoptedBinTarget is the PATH entry abcd owns and acts on when no --bin-dir is
// given: an existing owned entry adopted exactly where it stands, else a
// dangling one of ours repaired in place (installing elsewhere would leave it
// shadowing the new entry from earlier in PATH), else the default location.
// Empty when the home directory cannot be resolved — there is no user-scope
// location to write, and inventing a privileged one is what iss-171 removes.
func adoptedBinTarget(pluginRoot string) string {
	if pluginRoot != "" {
		if e, ok := ownedPathEntry(pluginRoot); ok {
			return e.path
		}
		if e, ok := danglingPathEntry(pluginRoot); ok {
			return e.path
		}
	}
	return binTarget()
}

// applyCtx threads the approved-category set and accumulated writes through the
// ordered apply steps.
type applyCtx struct {
	cwd        string
	det        DetectionResult
	approved   map[GapCategory]bool
	overrides  map[string]string
	prompter   Prompter
	gapPresent map[string]bool
	writes     []string
	changes    []string // human-readable value changes an explicit override forced
	notes      []string // loud refusals: what abcd deliberately did not do, and why
	autoYes    bool     // --yes: every category auto-approved without interaction
	devMode    bool     // --dev: install the track-latest shim instead of the symlink
	modeForced bool     // the requested install mode differs from the on-disk state
	binTarget  string   // the resolved PATH entry this run installs (never re-derived)

	visibilityForced bool     // an explicit --visibility override overwrote a valid value
	docsTargetForced bool     // an explicit --docs-target override overwrote a valid value
	markerRetract    []string // marker files a narrowed docs-target override de-selected
}

// note is the receipt seam; it lives in receipt.go with the path scrub it
// applies, because every apply step below reports its writes through it.

// refuse records a thing abcd deliberately did not do, and why. A refusal that
// shows up only as a still-open gap reads as a silent failure, so the reason
// travels with the install result. Callers pass text already rendered for a
// human — user-scope paths in tilde form.
func (a *applyCtx) refuse(reason string) { a.notes = append(a.notes, reason) }

// stepIdentityPin adopts the iss-62 identity gate for an un-pinned repo: it
// writes .abcd/config/identity.json from the current git author identity (the
// proposal), gated on ConfigChange approval (the confirmation). A mismatch is
// never auto-resolved — abcd must not silently change the pin or the user's git
// identity — so it stays a guided manual fix.
//
// It does NOT auto-adopt under --yes: pinning captures whatever git identity is
// currently set, so a blanket approval could pin a wrong/sandbox identity as
// canonical (the very value the gate exists to reject). Under --yes the
// un-pinned gap simply remains, to be adopted against a confirmed answer — typed
// at a terminal, or piped to a non-interactive run (iss-167). The exclusion is
// reported in InstallResult.OptionalSkipped, never left silent (iss-166).
func (a *applyCtx) stepIdentityPin() {
	if a.autoYes || !a.approved[ConfigChange] || !a.has(OptionalPinGapID) {
		return
	}
	eff, err := identity.EffectiveIdentity(a.cwd)
	if err != nil || eff.Name == "" || eff.Email == "" {
		return
	}
	if err := identity.WritePin(a.cwd, identity.Pin{Name: eff.Name, Email: eff.Email}); err == nil {
		a.note(identity.PinRelPath)
	}
}

func (a *applyCtx) has(id string) bool { return a.gapPresent[id] }

// stepDependencies re-probes PATH; surfaces the fix hint but never auto-runs a
// package manager.
func (a *applyCtx) stepDependencies() {
	if !a.approved[Dependency] {
		return
	}
	for _, g := range a.det.Gaps {
		if g.Category != Dependency {
			continue
		}
		tool := strings.TrimPrefix(strings.TrimSuffix(g.ID, "_missing"), "deps.")
		if !onPath(tool) {
			a.note("dependency: " + g.FixHint)
		}
	}
}

// stepSkeleton writes .abcd/config.json seed when the skeleton gap is present.
func (a *applyCtx) stepSkeleton() {
	if !a.approved[SafeAutocreate] || !a.has("skeleton.config_missing") {
		return
	}
	cfg := map[string]any{"meta": map[string]any{"schema_version": 1}}
	if err := writeConfig(a.cwd, cfg); err == nil {
		a.note(configPath(a.cwd))
	}
}

// stepConfigValues collects and persists the four config values. Returns nil on
// partial install (config-change declined and a required value missing).
func (a *applyCtx) stepConfigValues() *InstallConfig {
	hasConfigGap := a.has("config.visibility_missing") || a.has("config.docs_target_missing") ||
		a.has("config.oracle_backend_missing") || a.has("config.scan_deep_missing")

	// Load any already-valid persisted values. ok=false means config.json is
	// malformed JSON: collecting values and writing them back would rebuild the
	// file from scratch, DESTROYING whatever the user had. Refuse to touch a file
	// we cannot parse and report a partial install so the operator repairs it.
	ic, ok := loadPersistedInstallConfig(a.cwd)
	if !ok {
		return nil
	}

	// An explicitly-passed value override forces its slot even when the persisted
	// value is already valid, overwriting it and echoing the change (iss-107). The
	// shared applyOverride path covers all four config slots, so a re-install with
	// an explicit flag is never silently dropped; a re-install with NO override
	// leaves an already-valid value untouched (a silent no-op).
	oldDocsTarget := ic.DocsTarget
	visForced := a.applyOverride("visibility", visibilityChoices, &ic.Visibility)
	docsForced := a.applyOverride("docs_target", docsTargetChoices, &ic.DocsTarget)
	oracleForced := a.applyOverride("oracle_backend", oracleBackendChoices, &ic.OracleBackend)
	scanForced := a.applyScanDeepOverride(ic)
	forced := visForced || docsForced || oracleForced || scanForced
	a.visibilityForced = visForced  // stepVisibility must refresh .gitignore for a new visibility
	a.docsTargetForced = docsForced // stepMarker must re-plant markers for a new docs target
	if docsForced {
		// Narrowing the target set (e.g. both -> claude_md, or -> skip) leaves the
		// de-selected file's block orphaned; stepMarker retracts it so nothing is
		// left inconsistent.
		a.markerRetract = markerFilesDropped(oldDocsTarget, ic.DocsTarget)
	}

	if !hasConfigGap && !forced {
		return ic // all values already valid and no override forced a change
	}
	if hasConfigGap && !a.approved[ConfigChange] {
		// Category declined; a required value is missing.
		return nil
	}

	// Collect the missing values.
	if ic.Visibility == "" {
		ic.Visibility = a.resolveValue("visibility", visibilityChoices, "")
		if !inSet(ic.Visibility, visibilityChoices) {
			return nil // no valid visibility => partial
		}
	}
	if ic.DocsTarget == "" {
		ic.DocsTarget = a.resolveValue("docs_target", docsTargetChoices, docsTargetDefault)
		if !inSet(ic.DocsTarget, docsTargetChoices) {
			return nil // no valid docs target => partial (never persist a typo)
		}
	}
	if ic.OracleBackend == "" {
		ic.OracleBackend = a.resolveValue("oracle_backend", oracleBackendChoices, oracleBackendDefault)
		if !inSet(ic.OracleBackend, oracleBackendChoices) {
			return nil // no valid oracle backend => partial
		}
	}
	if ic.Visibility == "private" && onPath("trufflehog") && ic.ScanDeep == nil {
		v := a.resolveValue("scan_deep", []string{"true", "false"}, "false") == "true"
		ic.ScanDeep = &v
	}

	// Persist into the config map (read-modify-write). Re-read defensively; if the
	// file turned malformed since the first read, refuse rather than clobber it.
	cfgMap, cfgErr := readConfig(a.cwd)
	if cfgErr != nil {
		a.rollbackForced()
		return nil
	}
	if cfgMap == nil {
		cfgMap = map[string]any{}
	}
	setSub(cfgMap, "repo", "visibility", ic.Visibility)
	setSub(cfgMap, "docs", "target", ic.DocsTarget)
	setSub(cfgMap, "oracle", "backend", ic.OracleBackend)
	if ic.ScanDeep != nil {
		setSub(cfgMap, "scan", "deep", *ic.ScanDeep)
	}
	if err := writeConfig(a.cwd, cfgMap); err != nil {
		// The write did not land; do not echo a change or let downstream steps
		// reconcile .gitignore/markers against a config value that was not saved.
		a.rollbackForced()
		return nil
	}
	a.note(configPath(a.cwd))
	return ic
}

// rollbackForced discards the effects of a forced override whose config write did
// not land, so the echoed change and the downstream reconciliation steps never
// claim a change that was not persisted.
func (a *applyCtx) rollbackForced() {
	a.changes = nil
	a.visibilityForced = false
	a.docsTargetForced = false
	a.markerRetract = nil
}

// resolveValue picks a config value: an override wins, else the prompter.
func (a *applyCtx) resolveValue(key string, choices []string, def string) string {
	if a.overrides != nil {
		if v, ok := a.overrides[key]; ok && v != "" {
			return v
		}
	}
	return a.prompter.Prompt(key, choices, def)
}

// loadPersistedInstallConfig returns the already-valid persisted config values.
// A missing or invalid slot is left zero; a malformed config.json yields
// ok=false so callers refuse to touch a file they cannot parse.
func loadPersistedInstallConfig(cwd string) (*InstallConfig, bool) {
	cfgMap, err := readConfig(cwd)
	if err != nil {
		return nil, false
	}
	ic := &InstallConfig{}
	if v, ok := stringVal(subMap(cfgMap, "repo"), "visibility"); ok && inSet(v, visibilityChoices) {
		ic.Visibility = v
	}
	if v, ok := stringVal(subMap(cfgMap, "docs"), "target"); ok && inSet(v, docsTargetChoices) {
		ic.DocsTarget = v
	}
	if v, ok := stringVal(subMap(cfgMap, "oracle"), "backend"); ok && inSet(v, oracleBackendChoices) {
		ic.OracleBackend = v
	}
	if v, ok := boolVal(subMap(cfgMap, "scan"), "deep"); ok {
		vv := v
		ic.ScanDeep = &vv
	}
	return ic, true
}

// applyOverride force-sets *dst to an explicit, valid override for key when it
// differs from the current value, echoing the change and reporting whether it
// changed the slot. An empty slot is left for the collect-missing path (no
// overwrite, no echo); a missing/invalid override, or one already equal to the
// current value, is a no-op — so a plain re-install never clobbers (iss-107).
func (a *applyCtx) applyOverride(key string, choices []string, dst *string) bool {
	if *dst == "" {
		return false
	}
	v, ok := a.overrides[key]
	if !ok || v == "" || !inSet(v, choices) || *dst == v {
		return false
	}
	a.echoChange(key, *dst, v)
	*dst = v
	return true
}

// applyScanDeepOverride is applyOverride for the boolean scan.deep slot. It only
// forces an already-set value; an unset slot is left for the conditional
// collect-missing path.
func (a *applyCtx) applyScanDeepOverride(ic *InstallConfig) bool {
	if ic.ScanDeep == nil {
		return false
	}
	v, ok := a.overrides["scan_deep"]
	if !ok || (v != "true" && v != "false") {
		return false
	}
	want := v == "true"
	if *ic.ScanDeep == want {
		return false
	}
	from := "false"
	if *ic.ScanDeep {
		from = "true"
	}
	a.echoChange("scan_deep", from, v)
	ic.ScanDeep = &want
	return true
}

// echoChange records a human-readable "key: from -> to" line so an explicit
// override that overwrites an already-valid value is surfaced, not silent.
func (a *applyCtx) echoChange(key, from, to string) {
	a.changes = append(a.changes, fmt.Sprintf("%s: %s -> %s", key, from, to))
}

// overridesWouldChange reports whether any explicit value override differs from
// the currently-persisted config — the signal that an otherwise up-to-date repo
// still has work to do, so Install must not short-circuit as already_up_to_date
// (iss-107). A malformed config is treated as "no change": stepConfigValues
// refuses to touch it.
func overridesWouldChange(cwd string, overrides map[string]string) bool {
	if len(overrides) == 0 {
		return false
	}
	ic, ok := loadPersistedInstallConfig(cwd)
	if !ok {
		return false
	}
	differs := func(key string, choices []string, cur string) bool {
		v, ok := overrides[key]
		return ok && v != "" && cur != "" && inSet(v, choices) && cur != v
	}
	if differs("visibility", visibilityChoices, ic.Visibility) ||
		differs("docs_target", docsTargetChoices, ic.DocsTarget) ||
		differs("oracle_backend", oracleBackendChoices, ic.OracleBackend) {
		return true
	}
	if v, ok := overrides["scan_deep"]; ok && (v == "true" || v == "false") && ic.ScanDeep != nil {
		if *ic.ScanDeep != (v == "true") {
			return true
		}
	}
	return false
}

// stepVisibility rewrites the .gitignore block for the chosen visibility. It
// also runs when an explicit --visibility override forced a new value on an
// otherwise up-to-date repo (iss-107), so the .gitignore never drifts from the
// freshly-set visibility.
func (a *applyCtx) stepVisibility(cfg *InstallConfig) {
	if (!a.approved[ConfigChange] && !a.visibilityForced) || cfg == nil || cfg.Visibility == "" {
		return
	}
	wrote, err := applyVisibilityBlock(a.cwd, cfg.Visibility)
	if err == nil && wrote {
		a.note(filepath.Join(a.cwd, ".gitignore"))
	}
	// A narrowed public fence is said out loud (iss-255): the reader must learn
	// that the committed record tiers stay published, from the receipt rather
	// than from a later surprise in git status. Gated on the write succeeding —
	// a refused .gitignore holds no fence, and the note must not assert one.
	if err == nil {
		if _, narrowed := effectiveVisibilityEntries(a.cwd, cfg.Visibility); narrowed {
			a.refuse("visibility is public, but .abcd/ holds tracked files — an ignore rule cannot untrack committed records, so the .abcd/ fence covers only the local tier (.abcd/.work.local/; the memory/ snapshot fence is kept) and the committed record tiers remain published")
		}
	}
}

// stepHistory bootstraps ~/.abcd/history/, creates the per-root-sha dirs, writes
// meta.json, and registers/refreshes the repo entry.
func (a *applyCtx) stepHistory() {
	if !a.approved[UserState] && !a.approved[SafeAutocreate] {
		return
	}
	if a.approved[UserState] || a.approved[SafeAutocreate] {
		if wrote, err := bootstrapHistory(); err == nil && wrote {
			if root, e := historyRoot(); e == nil {
				a.note(filepath.Join(root, "index.json"))
			}
		}
	}
	root, err := historyRoot()
	if err != nil {
		return
	}
	sha := a.det.RepoIdentity.RootSHA
	if sha == "" {
		return
	}
	repoDir := filepath.Join(root, sha)
	transcripts := filepath.Join(repoDir, "transcripts")
	if a.approved[SafeAutocreate] && !fsutil.IsRealDir(transcripts) {
		if err := os.MkdirAll(transcripts, 0o755); err == nil {
			a.note(transcripts)
		}
	}
	metaPath := filepath.Join(repoDir, "meta.json")
	if a.approved[UserState] && !fileExists(metaPath) {
		meta := map[string]any{
			"root_commit": sha,
			"name":        a.det.RepoIdentity.Name,
			"github":      a.det.RepoIdentity.Github,
			"corpus":      map[string]any{"transcripts": "transcripts/"},
		}
		if err := writeJSON(metaPath, meta); err == nil {
			a.note(metaPath)
		}
	}
	if a.approved[UserState] {
		a.registerRepo(sha)
	}
}

// registerRepo registers or refreshes this repo's entry in index.json by its
// immutable root_commit. Re-founding lineage is only set on explicit confirm.
//
// The load-modify-write runs UNDER withHistoryLock with a re-load inside the lock
// (iss-101), so two concurrent installs cannot drop a registration by racing
// last-rename-wins. The re-founding confirmation, which can block on user input,
// is asked BEFORE the lock is acquired — the lock is never held across a prompt.
// The state that prompt validated (the candidate lineage exists and is still
// linkable) is RE-CHECKED against the re-loaded index inside the lock: if a
// concurrent install changed it, the lineage link the user approved against stale
// state is refused rather than silently applied.
func (a *applyCtx) registerRepo(sha string) {
	idx, err := loadHistoryIndex()
	if err != nil || idx == nil {
		return
	}
	id := a.det.RepoIdentity

	// Decide, OUTSIDE the lock, whether a re-founding prompt is needed. A prompt is
	// only relevant for a not-yet-registered sha with a lineage candidate; the
	// refresh path (sha already present) involves no prompt.
	linkLineage := false
	var candSHA string
	if indexEntry(idx, sha) == nil {
		if cand := findRefoundingCandidate(idx, id); cand != nil {
			candSHA = cand.RootCommit
			linkLineage = a.prompter.Confirm("Re-founded from " + shortSHA(candSHA) + "? Link lineage?")
		}
	}

	wrote := false
	lineageConflict := false
	_ = withHistoryLock(func() error {
		// Re-load inside the lock: the pre-lock read may be stale after a concurrent
		// install's write, and the mutation must build on the current index.
		locked, lerr := loadHistoryIndex()
		if lerr != nil || locked == nil {
			return lerr
		}
		if e := indexEntry(locked, sha); e != nil {
			e.Name, e.Github, e.Path = id.Name, id.Github, a.cwd // refresh mutable labels
			if e.Status == "" {
				e.Status = "active"
			}
			if werr := writeHistoryIndex(locked); werr != nil {
				return werr
			}
			wrote = true
			return nil
		}
		newEntry := historyRepo{RootCommit: sha, Name: id.Name, Github: id.Github, Path: a.cwd, Status: "active"}
		if linkLineage {
			// Re-check the answer-relevant state under the lock: the candidate we
			// prompted about must still exist by its root_commit and still be
			// linkable (not already superseded by a concurrent install). A changed
			// state is a conflict — skip the lineage link rather than write one the
			// user approved against a now-stale index.
			cand := indexEntry(locked, candSHA)
			if cand == nil || cand.SupersededBy != "" {
				linkLineage = false
				lineageConflict = true
			} else {
				newEntry.Supersedes = candSHA
				cand.SupersededBy = sha
				cand.Status = "superseded"
			}
		}
		locked.Repos = append(locked.Repos, newEntry)
		if werr := writeHistoryIndex(locked); werr != nil {
			return werr
		}
		wrote = true
		return nil
	})
	if lineageConflict {
		// Surface the refused link rather than silently proceed: the re-founding
		// candidate the user confirmed changed under the lock, so the repo is
		// registered without the lineage link they approved against stale state.
		a.changes = append(a.changes,
			"history lineage: link to "+shortSHA(candSHA)+" skipped (candidate changed under concurrent install)")
	}
	if wrote {
		if root, e2 := historyRoot(); e2 == nil {
			a.note(filepath.Join(root, "index.json"))
		}
	}
}

// stepMarker plants/refreshes the block in the docs.target files.
func (a *applyCtx) stepMarker(cfg *InstallConfig) {
	// Also runs when an explicit --docs-target override forced a new value on an
	// otherwise up-to-date repo (iss-107), so the marker block lands in the newly
	// chosen target file.
	if !a.approved[PluginOwned] && !a.docsTargetForced {
		return
	}
	target := docsTargetDefault
	if cfg != nil && cfg.DocsTarget != "" {
		target = cfg.DocsTarget
	} else {
		// fall back to persisted config
		cm, _ := readConfig(a.cwd)
		if v, ok := stringVal(subMap(cm, "docs"), "target"); ok {
			target = v
		}
	}
	for _, name := range markerTargets(target) {
		path := filepath.Join(a.cwd, name)
		if wrote, ok := installMarkerFile(path); ok && wrote {
			a.note(path)
		}
	}
	// Retract the block from files a narrowed docs-target override de-selected,
	// so a target change (e.g. both -> claude_md, or -> skip) leaves no orphan.
	for _, name := range a.markerRetract {
		path := filepath.Join(a.cwd, name)
		if wrote, ok := removeMarkerFile(path); ok && wrote {
			a.note(path)
		}
	}
}

// markerFilesDropped returns the marker files targeted by from but no longer by
// to — the blocks a docs-target narrowing orphans.
func markerFilesDropped(from, to string) []string {
	keep := map[string]bool{}
	for _, n := range markerTargets(to) {
		keep[n] = true
	}
	var dropped []string
	for _, n := range markerTargets(from) {
		if !keep[n] {
			dropped = append(dropped, n)
		}
	}
	return dropped
}

// stepSymlink installs the PATH entry: an owned symlink to the pinned binary
// (default), or the track-latest dev shim under --dev. It runs on a fresh install
// (the symlink.missing / symlink.dangling gaps) or when a mode switch was forced
// on an already-present owned entry (apply-as-update, iss-107). It refuses to
// clobber a foreign binary.
func (a *applyCtx) stepSymlink() {
	if a.det.pluginRoot == "" || a.binTarget == "" {
		return
	}
	gapDriven := a.approved[ConfigChange] && (a.has("symlink.missing") || a.has("symlink.dangling"))
	if !gapDriven && !a.modeForced {
		return
	}
	target := a.binTarget
	// A dangling entry of ours is cleared first: it resolves to nothing, so
	// removing it destroys nothing, while leaving it in place would keep a link
	// that shadows every later PATH entry — including the one being installed.
	a.clearDanglingEntry(target)
	kind := classifyBinTarget(target, a.det.pluginRoot)
	if kind == binTargetForeign {
		// Never clobber something we do not own — and never in silence. The
		// symlink.foreign gap is advisory, so it is filtered out of Remaining, and
		// with no note the run reports nothing written and no reason why; under an
		// explicit --bin-dir the detection gap does not even describe this location.
		a.refuse("refused to write the PATH entry " + displayPath(target) +
			": it is occupied by " + describeEntry(pathEntry{path: target, kind: kind}) +
			". abcd never clobbers a binary it does not own — remove it, or choose another directory with `--bin-dir`.")
		return
	}
	if a.devMode {
		a.installDevShim(target, kind)
		return
	}
	a.installPinnedSymlink(target, kind)
}

// installDevShim writes the track-latest shim, replacing an owned pinned symlink
// if one is there. An existing dev shim is left as-is (idempotent).
func (a *applyCtx) installDevShim(target string, kind binTargetKind) {
	if kind == binTargetDevShim {
		return
	}
	if kind == binTargetOwnedSymlink {
		if err := os.Remove(target); err != nil {
			return
		}
		a.echoChange("install_mode", "pinned", "dev")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return
	}
	content := renderDevShim(a.det.pluginRoot, pluginBinaryPath(a.det.pluginRoot))
	// Atomic no-follow write (rename over the target), matching the rest of the
	// store: a symlink pre-planted at the leaf is replaced, never written through,
	// and the executable bit is set via fchmod on the temp descriptor.
	if err := fsutil.WriteFileAtomic(target, []byte(content), 0o755); err != nil {
		return
	}
	a.note(target)
}

// clearDanglingEntry removes an abcd-owned symlink at target whose destination
// no longer exists. It is deliberately narrow: only a SYMLINK, only one that
// resolves to nothing, and only when the binary it would be repointed at exists.
// Nothing is destroyed (the link already answered nothing) and the alternative is
// worse — a dangling `abcd` earlier on PATH shadows the working install.
func (a *applyCtx) clearDanglingEntry(target string) {
	fi, err := os.Lstat(target)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		return
	}
	if present, serr := fsutil.Exists(target); serr != nil || present {
		return
	}
	if !fileExists(pluginBinaryPath(a.det.pluginRoot)) {
		// Nothing to repoint it at. Leaving the link is the lesser evil (removing
		// it would take abcd off PATH entirely for no gain), and installPinnedSymlink
		// states the reason — its source check runs before the owned-entry early
		// return precisely so this case is never silent.
		return
	}
	if err := os.Remove(target); err != nil {
		a.refuse("could not remove the dangling PATH entry " + displayPath(target) + ": " + errText(err))
	}
}

// installPinnedSymlink writes the owned symlink to the pinned binary, replacing a
// dev shim if one is there. An existing owned symlink is left as-is (idempotent).
// It REFUSES to create a link whose target does not exist: a dangling `abcd` on
// PATH shadows whatever else would have answered, so a broken plugin install must
// not be converted into a broken PATH (iss-171).
func (a *applyCtx) installPinnedSymlink(target string, kind binTargetKind) {
	// The source check comes FIRST, before the idempotent early return: an owned
	// entry whose binary is gone classifies as owned, so checking the kind first
	// would return silently and leave a dangling link reported as a healthy
	// install with no reason recorded anywhere.
	source := pluginBinaryPath(a.det.pluginRoot)
	if !fileExists(source) {
		a.refuse("refused to write the PATH entry " + displayPath(target) +
			": its target " + displayPath(source) + " does not exist — a dangling link would shadow any working abcd on PATH. Reinstall the plugin, then re-run `abcd ahoy install`.")
		return
	}
	if kind == binTargetOwnedSymlink {
		return
	}
	if kind == binTargetDevShim {
		if err := os.Remove(target); err != nil {
			return
		}
		a.echoChange("install_mode", "dev", "pinned")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		a.refuse("could not create the install directory " + displayPath(filepath.Dir(target)) + ": " + errText(err))
		return
	}
	if err := os.Symlink(source, target); err == nil {
		a.note(target)
	} else {
		a.refuse("could not write the PATH entry " + displayPath(target) + ": " + errText(err))
	}
}

// noteReachability describes, on the install result itself, whether the entry
// that was just written can actually be RUN.
//
// Both conditions are required gaps no apply step can close, and neither reaches
// the user any other way: a required+non-resolvable gap is excluded from
// Remaining, InstallResult carries no gaps, and `abcd ahoy doctor` — where the
// gap text lives — cannot be invoked by name on a machine where abcd is not yet
// on PATH. Install is the one place a fresh user sees output, so it says both
// things here. It also computes them against the target actually written, which
// is the only place an explicit --bin-dir is known.
func (a *applyCtx) noteReachability() {
	if a.binTarget == "" {
		return
	}
	present, err := fsutil.ExistsNoFollow(a.binTarget)
	if err != nil || !present {
		return // nothing was installed here; a refusal already said why
	}
	if dir := filepath.Dir(a.binTarget); !dirOnPath(dir) {
		a.refuse(pathReachMessage(dir))
	}
	if e, ok := shadowingEntry(a.det.pluginRoot, a.binTarget); ok {
		a.refuse(shadowMessage(e, a.binTarget))
	}
}

// modeWouldChange reports whether the requested install mode differs from the
// on-disk PATH target, so an otherwise up-to-date repo still has work to do. A
// foreign occupant is never touched, so it never counts as a change.
func modeWouldChange(opts InstallOptions, det DetectionResult, target string) bool {
	if det.pluginRoot == "" || target == "" {
		return false
	}
	kind := classifyBinTarget(target, det.pluginRoot)
	if kind == binTargetForeign {
		return false
	}
	if opts.Dev {
		// Only a switch away from the pinned symlink is a forced change. A missing
		// entry is NOT forced: it flows through the symlink.missing gap so a declined
		// ConfigChange approval is honoured (a fresh --dev must not bypass consent the
		// way a plain install cannot). An existing dev shim is already correct.
		return kind == binTargetOwnedSymlink
	}
	// Plain install: only a switch away from an existing dev shim is a forced
	// change; a missing entry is handled by the symlink.missing gap, and an owned
	// symlink is already correct.
	return kind == binTargetDevShim
}

// stepRules writes the per-repo .abcd/rules.json override skeleton when absent.
// It is deliberately the empty-domains skeleton, NOT a copy of the bundled
// defaults: the default domains live once in the abcd binary (itd-3), and this
// file only overrides them per-field (one-canonical-primitive). An empty
// domains map inherits every bundled default as-is.
func (a *applyCtx) stepRules() {
	if !a.approved[SafeAutocreate] || !a.has("rules.missing") {
		return
	}
	path := filepath.Join(a.cwd, ".abcd", "rules.json")
	rules := map[string]any{"schema_version": 1, "disabled": false, "domains": map[string]any{}}
	if err := writeJSON(path, rules); err == nil {
		a.note(path)
	}
}

// stepVersionStamp writes the meta setup block.
func (a *applyCtx) stepVersionStamp() {
	if !a.approved[SafeAutocreate] {
		return
	}
	if !a.has("install_meta.missing") && !a.has("version.upgrade") {
		return
	}
	cfgMap, _ := readConfig(a.cwd)
	if cfgMap == nil {
		cfgMap = map[string]any{}
	}
	meta := subMap(cfgMap, "meta")
	meta["schema_version"] = 1
	meta["setup_version"] = pluginVersion()
	meta["setup_date"] = time.Now().UTC().Format("2006-01-02")
	meta["project_name"] = a.det.RepoIdentity.Name
	cfgMap["meta"] = meta
	if err := writeConfig(a.cwd, cfgMap); err == nil {
		a.note(configPath(a.cwd))
	}
}

// Uninstall removes the marker block and the owned PATH symlink only. It never
// mutates hooks.json or the .abcd/ namespace.
//
// binDir names the directory to look in, for the one case detection cannot
// derive: an install placed by `--bin-dir` in a directory that is not on PATH is
// invisible to a PATH scan, so uninstall would otherwise orphan it. Empty means
// the derived location — an abcd-owned entry anywhere on PATH, else the default.
func Uninstall(cwd, binDir string) (UninstallReceipt, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return UninstallReceipt{}, err
	}
	var receipt UninstallReceipt

	// Marker: clean both surfaces regardless of the current docs.target.
	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		path := filepath.Join(abs, name)
		if wrote, ok := removeMarkerFile(path); ok && wrote {
			receipt.Marker.Removed = append(receipt.Marker.Removed, name)
		} else if !ok {
			receipt.Marker.Skipped = append(receipt.Marker.Skipped, name)
		}
	}

	// Symlink: remove only if it points at this plugin's binary. The entry is
	// found the same way detection finds it — an owned entry anywhere on PATH,
	// else the default location — so uninstall reaches the install that exists
	// rather than one blessed path.
	pluginRoot, ok := resolvePluginRoot()
	target := adoptedBinTarget(pluginRoot)
	if binDir != "" {
		if abs, aerr := filepath.Abs(binDir); aerr == nil {
			target = filepath.Join(abs, binName)
		}
	}
	if target == "" {
		receipt.Symlink.Note = "no user-scope install location; left untouched"
		return receipt, nil
	}
	// The receipt is written to be pasted into an issue, so the location is
	// rendered in tilde form and never carries the username (iss-177).
	receipt.Symlink.Target = displayPath(target)
	fi, lerr := os.Lstat(target)
	switch {
	case lerr != nil:
		receipt.Symlink.Note = "absent"
	case fi.Mode()&os.ModeSymlink == 0:
		// A regular file: our own dev shim (remove it — it is ours), else foreign.
		if isDevShimFile(target) {
			if err := os.Remove(target); err == nil {
				receipt.Symlink.Removed = true
				receipt.Symlink.Note = "removed dev shim"
			} else {
				receipt.Symlink.Note = "remove failed"
			}
		} else {
			receipt.Symlink.Note = "not a symlink; left untouched"
		}
	case !ok:
		receipt.Symlink.Note = "plugin root unresolved; left untouched"
	default:
		dest, _ := os.Readlink(target)
		if resolveSymlinkDest(target, dest) == resolvePath(pluginBinaryPath(pluginRoot)) {
			if err := os.Remove(target); err == nil {
				receipt.Symlink.Removed = true
			} else {
				receipt.Symlink.Note = "remove failed"
			}
		} else {
			receipt.Symlink.Note = "foreign symlink; left untouched"
		}
	}
	return receipt, nil
}

// Doctor runs the detection pass plus a read-only cross-machine audit. Zero
// writes.
func Doctor(cwd string) (DoctorReport, error) {
	det, err := Detect(cwd)
	if err != nil {
		return DoctorReport{}, err
	}
	report := DoctorReport{Detection: det}
	report.AuditGaps = auditGaps(cwd, det)
	return report, nil
}

// auditGaps reports read-only reconciliation issues (a stale registered path).
func auditGaps(cwd string, det DetectionResult) []Gap {
	var gaps []Gap
	if det.RootSHA == "" {
		return nil
	}
	idx, err := loadHistoryIndex()
	if err != nil || idx == nil {
		return nil
	}
	entry := indexEntry(idx, det.RootSHA)
	if entry == nil {
		return nil
	}
	abs, _ := filepath.Abs(cwd)
	if entry.Path != "" && entry.Path != abs {
		gaps = append(gaps, Gap{
			ID: "history.path_stale", Category: UserState, Scope: "repo",
			Title:   "registered path is stale",
			Detail:  "index.json records " + entry.Path + " but the repo is at " + abs + ".",
			FixHint: "ahoy install refreshes the registered path.", Required: false, Resolvable: true,
		})
	}
	return gaps
}

// Status renders the bare-command human summary. Zero writes.
func Status(cwd string) (string, error) {
	det, err := Detect(cwd)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "abcd ahoy — %s\n", det.FolderKind)
	fmt.Fprintf(&b, "plugin root: %s\n", det.PluginRootStatus)
	if det.RootSHA != "" {
		fmt.Fprintf(&b, "root sha: %s\n", shortSHA(det.RootSHA))
	}
	if mode, _ := det.Signals["install_mode"].(string); mode != "" {
		fmt.Fprintf(&b, "install: %s\n", mode)
	}
	act := actionable(det.Gaps)
	switch det.FolderKind {
	case UnmanagedFolder:
		b.WriteString("nothing to act on (not a git repo, no abcd markers)\n")
	case UnmanagedRepo:
		b.WriteString("unmanaged repo — run `abcd ahoy install` to adopt it\n")
	default:
		if len(act) == 0 {
			b.WriteString("already up to date\n")
		} else {
			fmt.Fprintf(&b, "%d actionable gap(s) — run `abcd ahoy install`\n", len(act))
		}
	}
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// approval + gap helpers
// ---------------------------------------------------------------------------

// OptionalPinGapID is the one optional gap --yes does not cover. Named here so
// the front doors can point at it without re-deriving the string.
const OptionalPinGapID = "git_identity.unpinned"

// optionalSkipped lists the optional gaps a --yes run left un-applied. --yes
// approves every resolvable category but never adopts the identity pin (see
// stepIdentityPin), so the skip is deliberate — and therefore has to be
// reported rather than left ambient (iss-166). Outside --yes the pin is offered
// as a confirmation, so nothing is skipped silently and the list stays empty.
func optionalSkipped(opts InstallOptions, gaps []Gap) []string {
	if !opts.Yes || !pinAdoptable(gaps) {
		return nil
	}
	return []string{OptionalPinGapID}
}

// pinAdoptable reports whether the advisory git-identity pin is the remaining
// work. It is the one gap install closes through an interactive confirmation
// (never under --yes), so it must not be short-circuited by the
// "already_up_to_date" early return.
func pinAdoptable(gaps []Gap) bool {
	for _, g := range gaps {
		if g.ID == OptionalPinGapID {
			return true
		}
	}
	return false
}

// actionable returns the required+resolvable gaps (the ones install must close).
func actionable(gaps []Gap) []Gap {
	var out []Gap
	for _, g := range gaps {
		if g.Required && g.Resolvable {
			out = append(out, g)
		}
	}
	return out
}

func gapIDs(gaps []Gap) []string {
	ids := make([]string, 0, len(gaps))
	for _, g := range gaps {
		ids = append(ids, g.ID)
	}
	sort.Strings(ids)
	return ids
}

func gapIDSet(gaps []Gap) map[string]bool {
	set := make(map[string]bool, len(gaps))
	for _, g := range gaps {
		set[g.ID] = true
	}
	return set
}

// categoryPromptOrder is the order the approval questions are asked in. It is
// the order the apply pass acts in — dependencies surfaced, the skeleton
// written, config values settled, user-scope state touched, and the marker
// block written last — so what the user is asked about follows what will
// happen to their repo.
//
// A fixed order is a contract, not a cosmetic. The questions are answered
// POSITIONALLY: a human reads down the list, and a non-interactive caller pipes
// answers in sequence. Ranging over the presence map instead handed out a fresh
// permutation on every run, so "answer y to the first question" approved a
// different category each time — a wrong answer that exits 0 and looks like a
// clean install (iss-167).
var categoryPromptOrder = []GapCategory{
	Dependency,
	SafeAutocreate,
	ConfigChange,
	UserState,
	PluginOwned,
}

// presentInPromptOrder returns the present categories in categoryPromptOrder,
// with any category the order does not name appended sorted. The tail matters:
// a category added to the type and forgotten here must still land in a FIXED
// place, so the order contract cannot be broken by omission — only made less
// meaningful, which the coverage test catches.
func presentInPromptOrder(present map[GapCategory]bool) []GapCategory {
	out := make([]GapCategory, 0, len(present))
	named := map[GapCategory]bool{}
	for _, c := range categoryPromptOrder {
		named[c] = true
		if present[c] {
			out = append(out, c)
		}
	}
	var rest []string
	for c := range present {
		if !named[c] {
			rest = append(rest, string(c))
		}
	}
	sort.Strings(rest)
	for _, c := range rest {
		out = append(out, GapCategory(c))
	}
	return out
}

// resolveApproval computes the approved category set once and the declined list.
func resolveApproval(gaps []Gap, opts InstallOptions, p Prompter) (map[GapCategory]bool, []string) {
	// Categories that have at least one resolvable gap can be approved.
	present := map[GapCategory]bool{}
	for _, g := range gaps {
		if g.Resolvable {
			present[g.Category] = true
		}
	}
	approved := map[GapCategory]bool{}
	switch {
	case opts.ApprovedCategories != nil:
		for c := range present {
			if opts.ApprovedCategories[c] {
				approved[c] = true
			}
		}
	case opts.Yes:
		for c := range present {
			approved[c] = true
		}
	default:
		for _, c := range presentInPromptOrder(present) {
			if p.Confirm("Apply " + string(c) + " changes?") {
				approved[c] = true
			}
		}
	}
	var declined []string
	for c := range present {
		if !approved[c] {
			declined = append(declined, string(c))
		}
	}
	sort.Strings(declined)
	return approved, declined
}

// setSub sets cfg[section][key] = value, creating the section map as needed.
func setSub(cfg map[string]any, section, key string, value any) {
	sub, ok := cfg[section].(map[string]any)
	if !ok {
		sub = map[string]any{}
	}
	sub[key] = value
	cfg[section] = sub
}
