// Package launch is abcd's transport-agnostic launch engine: it assembles the
// release bundle under a default-deny taxonomy, runs the native secret+PII scan,
// checks manifest lockstep, and previews newest-per-line retention — all as a
// dry-run that renders decisions without writing an artefact or touching the
// network. It performs no printing and no os.Exit, so it is fully testable and
// reusable across surfaces.
//
// The load-bearing invariant (adr-18/adr-28): the .abcd/** namespace and every
// other denied namespace can NEVER enter the bundle. This is a STRUCTURAL deny,
// not an allowlist toggle — no include pattern can promote a denied path.
package launch

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// DenyNamespaces are first-path-segment names that never ship. Structural deny
// (adr-18): NOT overridable by any allowlist. Mirrors launch_resolve
// DENY_NAMESPACES.
var DenyNamespaces = map[string]struct{}{
	".git": {}, ".abcd": {}, ".flow": {}, ".work": {}, ".specstory": {}, "memory": {},
}

// ExcludedReason is why a candidate was benignly excluded (never fails a ship).
type ExcludedReason string

const (
	ExcludedGitignored      ExcludedReason = "gitignored"
	ExcludedUnmatchedGlob   ExcludedReason = "unmatched_glob"
	ExcludedDeniedNamespace ExcludedReason = "denied_namespace"
)

// RejectedReason is why a candidate was rejected (any entry fails a ship).
type RejectedReason string

const (
	RejectedDeny            RejectedReason = "deny"
	RejectedSymlinkEscape   RejectedReason = "symlink_escape"
	RejectedSymlinkCycle    RejectedReason = "symlink_cycle"
	RejectedHardlinkDenied  RejectedReason = "hardlink_denied"
	RejectedHardlinkOffrepo RejectedReason = "hardlink_offrepo"
	RejectedDuplicate       RejectedReason = "duplicate"
	RejectedControlChar     RejectedReason = "control_char"
	RejectedMissingLiteral  RejectedReason = "missing_literal"
	RejectedFSError         RejectedReason = "fs_error"
)

// IncludedFile is a resolved payload file. Paths are repo-relative POSIX;
// ResolvedPath is the absolute on-disk (dereferenced) path.
type IncludedFile struct {
	LogicalPath  string `json:"logical_path"`
	ResolvedPath string `json:"resolved_path"`
	GitMode      string `json:"git_mode"` // "100644" | "100755"
}

// ExcludedFile is a benign exclusion.
type ExcludedFile struct {
	LogicalPath string         `json:"logical_path"`
	Reason      ExcludedReason `json:"reason"`
}

// RejectedFile is a violation.
type RejectedFile struct {
	LogicalPath string            `json:"logical_path"`
	Reason      RejectedReason    `json:"reason"`
	Details     map[string]string `json:"details,omitempty"`
}

// Bundle is the classified resolution outcome.
type Bundle struct {
	Included []IncludedFile `json:"files"`
	Excluded []ExcludedFile `json:"excluded"`
	Rejected []RejectedFile `json:"rejected"`
	Warnings []string       `json:"warnings"`
}

// HasViolation reports whether any rejected[] entry exists. ship hard-fails on
// true; dry-run reports it but still exits 0.
func (b Bundle) HasViolation() bool { return len(b.Rejected) > 0 }

// ScriptsClosureDenyDirs / ScriptsClosureDenySuffixes are the closure's own
// default-deny (dev-only names / compiled suffixes); a scripts/ path matching
// them is a benign excluded(denied_namespace) prune, not a resolution error.
var (
	scriptsDenyDirs     = map[string]struct{}{"__pycache__": {}, ".git": {}, ".mypy_cache": {}, ".pytest_cache": {}, "ralph": {}, "_intent_lint": {}}
	scriptsDenySuffixes = []string{".pyc", ".pyo"}
)

// ResolveBundle walks repoRoot, matches candidates against includes, and
// classifies each into Included / Excluded / Rejected under the ordered
// algorithm. includes==nil loads the committed config via LoadIncludes (a
// preflight fault is returned as an error).
func ResolveBundle(repoRoot string, includes []string) (Bundle, error) {
	return resolveBundle(repoRoot, includes, defaultClosureFn)
}

// resolveBundle is the injectable-closure implementation (ClosureFn is the one
// open dependency, spec §1 step 10).
func resolveBundle(repoRoot string, includes []string, closureFn ClosureFn) (Bundle, error) {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return Bundle{}, err
	}
	if real, err := filepath.EvalSymlinks(absRoot); err == nil {
		absRoot = real
	}

	if includes == nil {
		includes, err = LoadIncludes(absRoot)
		if err != nil {
			return Bundle{}, err
		}
	}

	var closure map[string]struct{}
	if anyReachesScripts(includes) && closureFn != nil {
		closure, err = closureFn(absRoot)
		if err != nil {
			return Bundle{}, preflight("scripts closure unreadable: %v", err)
		}
	}

	r := &resolver{
		root:         absRoot,
		includes:     includes,
		closure:      closure,
		matchedGlobs: map[string]struct{}{},
		inode:        buildInodeMap(absRoot),
	}

	// Missing-literal check: a literal include with no on-disk entry (Lstat, so a
	// dangling symlink counts as present) → rejected(missing_literal).
	for _, inc := range includes {
		if isGlob(inc) || inc == "." || inc == "" {
			continue
		}
		if _, err := os.Lstat(filepath.Join(absRoot, filepath.FromSlash(inc))); err != nil {
			r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: inc, Reason: RejectedMissingLiteral})
		}
	}

	// Walk + structural passes, collecting survivors pending the ignore pass.
	r.walkDir("", absRoot, map[string]struct{}{})

	// Unmatched globs → excluded(unmatched_glob).
	for _, inc := range includes {
		if isGlob(inc) {
			if _, ok := r.matchedGlobs[inc]; !ok {
				r.result.Excluded = append(r.result.Excluded, ExcludedFile{LogicalPath: inc, Reason: ExcludedUnmatchedGlob})
			}
		}
	}

	// Ignore pass (batched) + duplicate resolution → Included.
	r.finalize()

	sortBundle(&r.result)
	return r.result, nil
}

// resolver carries mutable state through the walk.
type resolver struct {
	root         string
	includes     []string
	closure      map[string]struct{}
	inode        *inodeMap
	matchedGlobs map[string]struct{}
	survivors    []candidate
	result       Bundle
}

// candidate is a survivor of the structural passes, pending the ignore pass.
type candidate struct {
	logical  string
	resolved string
	dev, ino uint64
	gitMode  string
	deref    bool
}

func (r *resolver) walkDir(rel, absDir string, ancestors map[string]struct{}) {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return // an unwalkable subtree already flagged the inode map uncertain
	}
	for _, e := range entries {
		name := e.Name()
		childRel := name
		if rel != "" {
			childRel = rel + "/" + name
		}
		childAbs := filepath.Join(absDir, name)

		if hasControlChar(childRel) {
			r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: childRel, Reason: RejectedControlChar})
			continue
		}
		info, err := os.Lstat(childAbs)
		if err != nil {
			continue
		}
		mode := info.Mode()
		switch {
		case mode&os.ModeSymlink != 0:
			r.handleSymlink(childRel, childAbs, ancestors)
		case mode.IsDir():
			if _, denied := DenyNamespaces[firstSegment(childRel)]; denied {
				// Structural deny BEFORE ignore: a denied dir reached by a broad
				// include is excluded(denied_namespace); otherwise silently pruned.
				// Either way it is never descended — .abcd/** cannot enter here.
				if r.anyIncludeMatches(childRel) {
					r.result.Excluded = append(r.result.Excluded, ExcludedFile{LogicalPath: childRel, Reason: ExcludedDeniedNamespace})
				}
				continue
			}
			r.walkDir(childRel, childAbs, ancestors)
		case mode.IsRegular():
			r.classifyRegular(childRel, childAbs, info, false)
		default:
			if r.firstMatchAndMark(childRel) != "" {
				r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: childRel, Reason: RejectedFSError})
			}
		}
	}
}

// classifyRegular applies include-match, denied, scripts-closure and hardlink
// passes to one regular file, appending a survivor when it passes.
func (r *resolver) classifyRegular(rel, abs string, info os.FileInfo, deref bool) {
	source := r.firstMatchAndMark(rel)
	if source == "" {
		return // default-deny: not requested at all
	}
	if _, denied := DenyNamespaces[firstSegment(rel)]; denied {
		r.result.Excluded = append(r.result.Excluded, ExcludedFile{LogicalPath: rel, Reason: ExcludedDeniedNamespace})
		return
	}
	if firstSegment(rel) == "scripts" && r.closure != nil {
		if _, in := r.closure[rel]; !in {
			if r.anyLiteralFileInclude(rel) && !scriptsDenied(rel) {
				r.result.Rejected = append(r.result.Rejected, RejectedFile{
					LogicalPath: rel, Reason: RejectedFSError,
					Details: map[string]string{"kind": "scripts_not_in_runtime_closure"},
				})
				return
			}
			r.result.Excluded = append(r.result.Excluded, ExcludedFile{LogicalPath: rel, Reason: ExcludedDeniedNamespace})
			return
		}
	}

	dev, ino := inodeOf(info)
	// Hardlink alias map, fail-closed: any uncertainty rejects fs_error.
	if r.inode.uncertain {
		r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: rel, Reason: RejectedFSError})
		return
	}
	if r.inode.aliasDenied(dev, ino) {
		r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: rel, Reason: RejectedHardlinkDenied})
		return
	}

	gitMode := "100644"
	if info.Mode()&0o111 != 0 {
		gitMode = "100755"
	}
	r.survivors = append(r.survivors, candidate{
		logical: rel, resolved: abs, dev: dev, ino: ino, gitMode: gitMode, deref: deref,
	})
}

// handleSymlink resolves a symlink structurally (escape/cycle/deny) and, when
// accepted, dereferences it: a file is classified under its logical path; a
// directory is walked with its contents emitted under the symlink's prefix. A
// symlink is only classified when an include could reach it (default-deny).
func (r *resolver) handleSymlink(rel, abs string, ancestors map[string]struct{}) {
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if !r.anyIncludeMatches(rel) && !r.includeMayReachDir(rel) {
			return
		}
		reason := RejectedFSError
		if strings.Contains(err.Error(), "too many links") {
			reason = RejectedSymlinkCycle
		}
		r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: rel, Reason: reason})
		return
	}
	relToRoot, err := filepath.Rel(r.root, real)
	if err != nil || relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		if r.anyIncludeMatches(rel) || r.includeMayReachDir(rel) {
			r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: rel, Reason: RejectedSymlinkEscape})
		}
		return
	}
	realRel := filepath.ToSlash(relToRoot)
	if _, denied := DenyNamespaces[firstSegment(realRel)]; denied {
		// A symlink whose target realpath is under a denied namespace is a
		// smuggling attempt — reject(deny) even if the logical path is benign.
		if r.anyIncludeMatches(rel) || r.includeMayReachDir(rel) {
			r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: rel, Reason: RejectedDeny})
		}
		return
	}
	tinfo, err := os.Stat(real)
	if err != nil {
		return
	}
	if tinfo.IsDir() {
		if !r.includeMayReachDir(rel) {
			return
		}
		if _, seen := ancestors[real]; seen {
			r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: rel, Reason: RejectedSymlinkCycle})
			return
		}
		next := map[string]struct{}{real: {}}
		for k := range ancestors {
			next[k] = struct{}{}
		}
		r.walkSymlinkTarget(rel, real, next)
		return
	}
	if tinfo.Mode().IsRegular() {
		r.classifyRegular(rel, real, tinfo, true)
	}
}

// walkSymlinkTarget walks a dereferenced symlink-target directory, emitting each
// regular file under logicalPrefix and recursing (with a cycle guard) into
// nested symlink dirs so a nested target under a denied namespace is still
// rejected rather than silently skipped.
func (r *resolver) walkSymlinkTarget(logicalPrefix, realDir string, ancestors map[string]struct{}) {
	entries, err := os.ReadDir(realDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		childLogical := logicalPrefix + "/" + name
		childAbs := filepath.Join(realDir, name)
		if hasControlChar(childLogical) {
			r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: childLogical, Reason: RejectedControlChar})
			continue
		}
		info, err := os.Lstat(childAbs)
		if err != nil {
			continue
		}
		mode := info.Mode()
		switch {
		case mode&os.ModeSymlink != 0:
			r.handleSymlink(childLogical, childAbs, ancestors)
		case mode.IsDir():
			r.walkSymlinkTarget(childLogical, childAbs, ancestors)
		case mode.IsRegular():
			r.classifyRegular(childLogical, childAbs, info, true)
		}
	}
}

// finalize applies the batched ignore pass and duplicate resolution, promoting
// surviving candidates to Included.
func (r *resolver) finalize() {
	// Batch the ignore probe over every survivor's logical path.
	paths := make([]string, 0, len(r.survivors))
	for _, c := range r.survivors {
		paths = append(paths, c.logical)
	}
	ignored := gitutil.CheckIgnored(r.root, paths)

	// Group by logical path for duplicate-by-provenance resolution.
	byLogical := map[string][]candidate{}
	var order []string
	for _, c := range r.survivors {
		if _, ok := ignored[c.logical]; ok {
			r.result.Excluded = append(r.result.Excluded, ExcludedFile{LogicalPath: c.logical, Reason: ExcludedGitignored})
			continue
		}
		if _, seen := byLogical[c.logical]; !seen {
			order = append(order, c.logical)
		}
		byLogical[c.logical] = append(byLogical[c.logical], c)
	}

	for _, logical := range order {
		group := byLogical[logical]
		if len(group) == 1 {
			c := group[0]
			r.result.Included = append(r.result.Included, IncludedFile{LogicalPath: c.logical, ResolvedPath: c.resolved, GitMode: c.gitMode})
			continue
		}
		// Same logical path from multiple sources: same inode → dedup with a
		// warning; distinct inode → rejected(duplicate).
		first := group[0]
		sameInode := true
		for _, c := range group[1:] {
			if c.dev != first.dev || c.ino != first.ino {
				sameInode = false
				break
			}
		}
		if sameInode {
			r.result.Warnings = append(r.result.Warnings, "duplicate provenance for "+logical+" (same inode); kept one")
			r.result.Included = append(r.result.Included, IncludedFile{LogicalPath: first.logical, ResolvedPath: first.resolved, GitMode: first.gitMode})
		} else {
			r.result.Rejected = append(r.result.Rejected, RejectedFile{LogicalPath: logical, Reason: RejectedDuplicate})
		}
	}
}

// ---------------------------------------------------------------------------
// Include matching (segment-aware glob, RE2-safe)
// ---------------------------------------------------------------------------

var globMeta = "*?["

func hasGlobMeta(s string) bool { return strings.ContainsAny(s, globMeta) }

func isGlob(pattern string) bool { return hasGlobMeta(pattern) }

func firstSegment(rel string) string {
	if i := strings.Index(rel, "/"); i >= 0 {
		return rel[:i]
	}
	return rel
}

// firstMatchAndMark returns the first include covering rel and marks every glob
// that covers it as matched (so a file also covered by an earlier include does
// not leave a later glob falsely reported unmatched).
func (r *resolver) firstMatchAndMark(rel string) string {
	first := ""
	for _, inc := range r.includes {
		if matchesInclude(rel, inc) {
			if first == "" {
				first = inc
			}
			if isGlob(inc) {
				r.matchedGlobs[inc] = struct{}{}
			}
		}
	}
	return first
}

func (r *resolver) anyIncludeMatches(rel string) bool {
	for _, inc := range r.includes {
		if matchesInclude(rel, inc) {
			return true
		}
	}
	return false
}

// anyLiteralFileInclude reports whether any include is a literal file path equal
// to rel (an explicit ship request for that exact file).
func (r *resolver) anyLiteralFileInclude(rel string) bool {
	for _, inc := range r.includes {
		if !isGlob(inc) && inc == rel {
			return true
		}
	}
	return false
}

// includeMayReachDir reports whether any include could match something at or
// under dir (used to decide whether to deref a symlink dir).
func (r *resolver) includeMayReachDir(dir string) bool {
	for _, inc := range r.includes {
		if matchesInclude(dir, inc) {
			return true
		}
		if !isGlob(inc) && strings.HasPrefix(inc, dir+"/") {
			return true
		}
		if isBroad(inc) {
			return true
		}
		if isGlob(inc) && strings.HasPrefix(inc, dir+"/") {
			return true
		}
	}
	return false
}

// isBroad reports whether an include may reach a denied namespace during the
// walk (a root wildcard, or a first-segment glob).
func isBroad(pattern string) bool {
	switch pattern {
	case ".", "", "**", "*":
		return true
	}
	return isGlob(firstSegment(pattern))
}

func anyReachesScripts(includes []string) bool {
	for _, inc := range includes {
		first := firstSegment(inc)
		if first == "scripts" {
			return true
		}
		switch inc {
		case ".", "", "**", "*":
			return true
		}
		if isGlob(first) {
			if ok, _ := filepath.Match(first, "scripts"); ok {
				return true
			}
		}
	}
	return false
}

// matchesInclude reports whether repo-relative rel is covered by pattern. A
// literal dir pattern covers its whole subtree; a literal file matches exactly.
// Globs are segment-aware: ** spans separators, a single * / ? / [...] never
// crosses /.
func matchesInclude(rel, pattern string) bool {
	if pattern == "." || pattern == "" {
		return true
	}
	if isGlob(pattern) {
		return globMatch(rel, pattern)
	}
	return rel == pattern || strings.HasPrefix(rel, pattern+"/")
}

// globMatch does a segment-aware glob match. A trailing /** also matches the
// base dir and its whole subtree.
func globMatch(rel, pattern string) bool {
	if strings.HasSuffix(pattern, "/**") {
		base := pattern[:len(pattern)-3]
		if base != "" && !isGlob(base) {
			return rel == base || strings.HasPrefix(rel, base+"/")
		}
	}
	re, guards := globToRegexp(pattern)
	m := re.FindStringSubmatch(rel)
	if m == nil {
		return false
	}
	// Guarded groups are positive char classes that must never cross a separator.
	for _, g := range guards {
		if g < len(m) && strings.Contains(m[g], "/") {
			return false
		}
	}
	return true
}

// globRegexpCache memoises compiled globs. It is guarded by globRegexpMu so the
// transport-agnostic core can resolve bundles concurrently without a data race
// (iss-31). A concurrent miss may compile the same pattern twice; that is benign
// (both results are equivalent), so the fast path takes only a read lock.
var (
	globRegexpMu    sync.RWMutex
	globRegexpCache = map[string]compiledGlob{}
)

type compiledGlob struct {
	re     *regexp.Regexp
	guards []int
}

// globToRegexp compiles a glob to an anchored RE2 regex where only ** crosses /.
// RE2 has no lookahead, so the single-segment separator guard is expressed via
// [^/] for * and ?, and (for a positive char class that could otherwise match /)
// via a capturing group whose captured text is post-checked for a / — the guard
// group indices are returned.
func globToRegexp(pattern string) (*regexp.Regexp, []int) {
	globRegexpMu.RLock()
	c, ok := globRegexpCache[pattern]
	globRegexpMu.RUnlock()
	if ok {
		return c.re, c.guards
	}
	var b strings.Builder
	var guards []int
	group := 0
	b.WriteString(`^`)
	i, n := 0, len(pattern)
	for i < n {
		ch := pattern[i]
		switch {
		case ch == '*':
			if i+1 < n && pattern[i+1] == '*' {
				if i+2 < n && pattern[i+2] == '/' {
					b.WriteString(`(?:.*/)?`)
					i += 3
				} else {
					b.WriteString(`.*`)
					i += 2
				}
				continue
			}
			b.WriteString(`[^/]*`)
			i++
		case ch == '?':
			b.WriteString(`[^/]`)
			i++
		case ch == '[':
			cls, negated, adv, ok := parseCharClass(pattern, i)
			if !ok {
				b.WriteString(`\[`)
				i++
				continue
			}
			if negated {
				// A negated class must also exclude /; add it to the negation.
				b.WriteString(`[^` + cls + `/]`)
			} else {
				// A positive class is wrapped in a capturing group so a match that
				// somehow spans / (e.g. a range crossing 0x2F) can be rejected.
				group++
				guards = append(guards, group)
				b.WriteString(`(` + `[` + cls + `]` + `)`)
			}
			i = adv
		case strings.IndexByte(`.^$+{}()|\`, ch) >= 0:
			b.WriteByte('\\')
			b.WriteByte(ch)
			i++
		default:
			b.WriteByte(ch)
			i++
		}
	}
	b.WriteString(`$`)
	re := regexp.MustCompile(b.String())
	globRegexpMu.Lock()
	globRegexpCache[pattern] = compiledGlob{re: re, guards: guards}
	globRegexpMu.Unlock()
	return re, guards
}

// parseCharClass parses a [...] class starting at i, returning its body (without
// the leading ^/! for a negated class), whether it is negated, the index after
// the closing ], and ok=false when unterminated.
func parseCharClass(pattern string, i int) (body string, negated bool, adv int, ok bool) {
	j := i + 1
	n := len(pattern)
	if j < n && (pattern[j] == '!' || pattern[j] == '^') {
		negated = true
		j++
	}
	if j < n && pattern[j] == ']' { // literal ] as first member
		j++
	}
	for j < n && pattern[j] != ']' {
		j++
	}
	if j >= n {
		return "", false, i + 1, false
	}
	start := i + 1
	if negated {
		start = i + 2
	}
	return pattern[start:j], negated, j + 1, true
}

// scriptsDenied reports whether a scripts/-tree path matches the closure's own
// default-deny (dev-only dir names / compiled suffixes).
func scriptsDenied(rel string) bool {
	for _, seg := range strings.Split(rel, "/") {
		if _, ok := scriptsDenyDirs[seg]; ok {
			return true
		}
	}
	for _, sfx := range scriptsDenySuffixes {
		if strings.HasSuffix(rel, sfx) {
			return true
		}
	}
	return false
}

func hasControlChar(text string) bool {
	for _, r := range text {
		if r <= 0x1F || r == 0x7F {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Hardlink inode map (fail-closed)
// ---------------------------------------------------------------------------

type inodeMap struct {
	aliases      map[[2]uint64][]string
	deniedInodes map[[2]uint64]struct{}
	uncertain    bool
}

func (m *inodeMap) aliasDenied(dev, ino uint64) bool {
	_, ok := m.deniedInodes[[2]uint64{dev, ino}]
	return ok
}

// buildInodeMap builds the (st_dev, st_ino) alias map over the repo's
// regular-file tree (excluding .git/objects), fail-closed: any Lstat error or
// unwalkable subtree marks the map uncertain so no candidate is admitted on
// unproven alias-safety.
func buildInodeMap(root string) *inodeMap {
	m := &inodeMap{aliases: map[[2]uint64][]string{}, deniedInodes: map[[2]uint64]struct{}{}}
	gitObjects := filepath.Join(root, ".git", "objects")
	var walk func(dir string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			m.uncertain = true
			return
		}
		for _, e := range entries {
			abs := filepath.Join(dir, e.Name())
			if abs == gitObjects {
				continue // content-addressed blobs, never alias-bearing
			}
			info, err := os.Lstat(abs)
			if err != nil {
				m.uncertain = true
				continue
			}
			if info.IsDir() {
				walk(abs)
				continue
			}
			if !info.Mode().IsRegular() {
				continue
			}
			dev, ino := inodeOf(info)
			key := [2]uint64{dev, ino}
			rel, _ := filepath.Rel(root, abs)
			relSlash := filepath.ToSlash(rel)
			m.aliases[key] = append(m.aliases[key], relSlash)
			if _, denied := DenyNamespaces[firstSegment(relSlash)]; denied {
				m.deniedInodes[key] = struct{}{}
			}
		}
	}
	walk(root)
	return m
}

func inodeOf(info os.FileInfo) (dev, ino uint64) {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return uint64(st.Dev), uint64(st.Ino)
	}
	return 0, 0
}

func sortBundle(b *Bundle) {
	sort.SliceStable(b.Included, func(i, j int) bool { return b.Included[i].LogicalPath < b.Included[j].LogicalPath })
	sort.SliceStable(b.Excluded, func(i, j int) bool {
		if b.Excluded[i].LogicalPath != b.Excluded[j].LogicalPath {
			return b.Excluded[i].LogicalPath < b.Excluded[j].LogicalPath
		}
		return b.Excluded[i].Reason < b.Excluded[j].Reason
	})
	sort.SliceStable(b.Rejected, func(i, j int) bool {
		if b.Rejected[i].LogicalPath != b.Rejected[j].LogicalPath {
			return b.Rejected[i].LogicalPath < b.Rejected[j].LogicalPath
		}
		return b.Rejected[i].Reason < b.Rejected[j].Reason
	})
}
