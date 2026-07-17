package ahoy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"sort"
	"strings"
	"time"

	"github.com/REPPL/abcd-cli/internal/core/identity"
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

	// Idempotency: zero required+resolvable gaps => exact no-op. The one
	// exception is the advisory git-identity pin, which install adopts through an
	// interactive confirmation (never under --yes): fall through so a later
	// `ahoy install` on an otherwise-clean repo can still pin, as the gap's fix
	// hint advertises.
	if len(actionable(det.Gaps)) == 0 && !(!opts.Yes && pinAdoptable(det.Gaps)) {
		return InstallResult{Status: "already_up_to_date"}, nil
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
	}

	// Ordered apply steps.
	ac.stepDependencies()
	ac.stepSkeleton()
	cfg := ac.stepConfigValues()
	ac.stepVisibility(cfg)
	ac.stepHistory()
	ac.stepMarker(cfg)
	ac.stepSymlink()
	ac.stepRules()
	ac.stepVersionStamp()
	ac.stepIdentityPin()

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
		Remaining:          remaining,
		DeclinedCategories: declined,
	}, nil
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
	autoYes    bool // --yes: every category auto-approved without interaction
}

func (a *applyCtx) note(path string) { a.writes = append(a.writes, path) }

// stepIdentityPin adopts the iss-62 identity gate for an un-pinned repo: it
// writes .abcd/config/identity.json from the current git author identity (the
// proposal), gated on ConfigChange approval (the confirmation). A mismatch is
// never auto-resolved — abcd must not silently change the pin or the user's git
// identity — so it stays a guided manual fix.
//
// It does NOT auto-adopt under --yes: pinning captures whatever git identity is
// currently set, so a non-interactive run could pin a wrong/sandbox identity as
// canonical (the very value the gate exists to reject). Under --yes the
// un-pinned gap simply remains, to be adopted with an interactive confirmation.
func (a *applyCtx) stepIdentityPin() {
	if a.autoYes || !a.approved[ConfigChange] || !a.has("git_identity.unpinned") {
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
	// Nothing to collect if there are no config gaps and category not approved.
	hasConfigGap := a.has("config.visibility_missing") || a.has("config.docs_target_missing") ||
		a.has("config.oracle_backend_missing") || a.has("config.scan_deep_missing")
	cfgMap, cfgErr := readConfig(a.cwd)
	if cfgErr != nil {
		// config.json exists but is malformed JSON. Collecting values and writing
		// them back would rebuild the file from scratch, DESTROYING whatever the
		// user had. Refuse to touch a file we cannot parse and report a partial
		// install so the operator repairs it by hand.
		return nil
	}
	repo := subMap(cfgMap, "repo")
	docs := subMap(cfgMap, "docs")
	oracle := subMap(cfgMap, "oracle")
	scan := subMap(cfgMap, "scan")

	// Load any already-valid persisted values.
	ic := &InstallConfig{}
	if v, ok := stringVal(repo, "visibility"); ok && inSet(v, visibilityChoices) {
		ic.Visibility = v
	}
	if v, ok := stringVal(docs, "target"); ok && inSet(v, docsTargetChoices) {
		ic.DocsTarget = v
	}
	if v, ok := stringVal(oracle, "backend"); ok && inSet(v, oracleBackendChoices) {
		ic.OracleBackend = v
	}
	if v, ok := boolVal(scan, "deep"); ok {
		vv := v
		ic.ScanDeep = &vv
	}

	if !hasConfigGap {
		return ic // all values already valid; nothing to collect
	}
	if !a.approved[ConfigChange] {
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
	cfgMap, cfgErr = readConfig(a.cwd)
	if cfgErr != nil {
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
	if err := writeConfig(a.cwd, cfgMap); err == nil {
		a.note(configPath(a.cwd))
	}
	return ic
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

// stepVisibility rewrites the .gitignore block for the chosen visibility.
func (a *applyCtx) stepVisibility(cfg *InstallConfig) {
	if !a.approved[ConfigChange] || cfg == nil || cfg.Visibility == "" {
		return
	}
	wrote, err := applyVisibilityBlock(a.cwd, cfg.Visibility)
	if err == nil && wrote {
		a.note(filepath.Join(a.cwd, ".gitignore"))
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
func (a *applyCtx) registerRepo(sha string) {
	idx, err := loadHistoryIndex()
	if err != nil || idx == nil {
		return
	}
	id := a.det.RepoIdentity
	if e := indexEntry(idx, sha); e != nil {
		e.Name, e.Github, e.Path = id.Name, id.Github, a.cwd // refresh mutable labels
		if e.Status == "" {
			e.Status = "active"
		}
		_ = writeHistoryIndex(idx)
		if root, e2 := historyRoot(); e2 == nil {
			a.note(filepath.Join(root, "index.json"))
		}
		return
	}
	newEntry := historyRepo{RootCommit: sha, Name: id.Name, Github: id.Github, Path: a.cwd, Status: "active"}
	if cand := findRefoundingCandidate(idx, id); cand != nil {
		if a.prompter.Confirm("Re-founded from " + shortSHA(cand.RootCommit) + "? Link lineage?") {
			newEntry.Supersedes = cand.RootCommit
			cand.SupersededBy = sha
			cand.Status = "superseded"
		}
	}
	idx.Repos = append(idx.Repos, newEntry)
	_ = writeHistoryIndex(idx)
	if root, e2 := historyRoot(); e2 == nil {
		a.note(filepath.Join(root, "index.json"))
	}
}

// stepMarker plants/refreshes the block in the docs.target files.
func (a *applyCtx) stepMarker(cfg *InstallConfig) {
	if !a.approved[PluginOwned] {
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
}

// stepSymlink installs the owned PATH symlink. It refuses to clobber a foreign
// binary. Default: yes for private, no for public.
func (a *applyCtx) stepSymlink() {
	if !a.approved[ConfigChange] || !a.has("symlink.missing") {
		return
	}
	if a.det.pluginRoot == "" {
		return
	}
	target := binTarget()
	source := pluginBinaryPath(a.det.pluginRoot)
	// Refuse to clobber anything already present.
	if _, err := os.Lstat(target); err == nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return
	}
	if err := os.Symlink(source, target); err == nil {
		a.note(target)
	}
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
func Uninstall(cwd string) (UninstallReceipt, error) {
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

	// Symlink: remove only if it points at this plugin's binary.
	target := binTarget()
	receipt.Symlink.Target = target
	pluginRoot, ok := resolvePluginRoot()
	fi, lerr := os.Lstat(target)
	switch {
	case lerr != nil:
		receipt.Symlink.Note = "absent"
	case fi.Mode()&os.ModeSymlink == 0:
		receipt.Symlink.Note = "not a symlink; left untouched"
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

// pinAdoptable reports whether the advisory git-identity pin is the remaining
// work. It is the one gap install closes through an interactive confirmation
// (never under --yes), so it must not be short-circuited by the
// "already_up_to_date" early return.
func pinAdoptable(gaps []Gap) bool {
	for _, g := range gaps {
		if g.ID == "git_identity.unpinned" {
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
		for c := range present {
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
