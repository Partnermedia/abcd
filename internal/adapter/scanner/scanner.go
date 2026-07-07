package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// BundleFile is the minimal descriptor ScanBundle needs. It is defined here
// (rather than importing the launch package) so the scanner has NO dependency on
// launch — that would create an import cycle, since launch's dry-run orchestrator
// imports the scanner. The launch package adapts its IncludedFile slice to this.
type BundleFile struct {
	LogicalPath  string
	ResolvedPath string
}

// ScanResult is the outcome of scanning a bundle.
type ScanResult struct {
	Findings          []Finding `json:"findings"`
	FilesScanned      int       `json:"files_scanned"`
	HardFails         int       `json:"hard_fails"`
	Unavailable       bool      `json:"unavailable"`
	UnavailableReason string    `json:"unavailable_reason,omitempty"`
	// Unscanned lists bundle files that were present but classified
	// binary/unscannable by the content sniff (e.g. a leading-NUL file). They
	// are surfaced, never silently dropped, so a crafted binary cannot smuggle
	// unscanned content into a source bundle without being visible.
	Unscanned []string `json:"unscanned,omitempty"`
}

// Config is the on-disk scanner configuration (the per-repo pii.json override
// shape). Only the consumed fields are modelled.
type Config struct {
	SkipDirs           []string              `json:"skip_dirs"`
	SkipPathFragments  []string              `json:"skip_path_fragments"`
	SkipExtensions     []string              `json:"skip_extensions"`
	SkipFilenames      []string              `json:"skip_filenames"`
	Patterns           map[string]patternDef `json:"patterns"`
	IdentitySeverities map[string]Severity   `json:"identity_severities"`
}

// patternDef is one pattern definition in a config override.
type patternDef struct {
	Regex           string   `json:"regex"`
	Kind            string   `json:"kind"`
	Label           string   `json:"label"`
	Severity        Severity `json:"severity"`
	CaseInsensitive bool     `json:"case_insensitive"`
	Suggestion      string   `json:"suggestion"`
}

// Scanner holds the merged config, compiled patterns and probed identity for a
// repo. Construct it with New.
type Scanner struct {
	patterns       []Pattern
	identity       Identity
	identSev       map[string]Severity
	skipExtensions map[string]struct{}
	skipFilenames  map[string]struct{}
	skipFragments  []string
	unavailable    bool
	unavailReason  string
}

// defaultSkipExtensions / defaultSkipFilenames mirror the bundled pii.json
// binary/noise skip sets.
var (
	defaultSkipExtensions = []string{
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".pdf", ".ico", ".webp",
		".mp3", ".mp4", ".mov", ".webm", ".wav",
		".zip", ".tar", ".gz", ".tgz", ".bz2", ".xz", ".7z",
		".pyc", ".pyo", ".so", ".dylib", ".dll", ".exe",
		".sqlite", ".db", ".lock",
	}
	defaultSkipFilenames = []string{".DS_Store", "Thumbs.db", ".gitignore"}
	defaultSkipFragments = []string{".abcd/logbook/pii-scan/", ".abcd/logbook/audit-history/"}
	repoConfigRelPath    = filepath.Join(".abcd", "config", "pii.json")
)

// New builds a Scanner for repoRoot: it starts from the built-in secret set and
// identity floors, merges the per-repo .abcd/config/pii.json override when
// present (enforcing the severity floor), and probes identity. If the per-repo
// config exists but cannot be read or parsed, the scanner is marked unavailable
// (fail-closed): New still returns a usable value with no error, and ScanBundle
// surfaces Unavailable=true.
func New(repoRoot string) (*Scanner, error) {
	s := &Scanner{
		patterns:       DefaultPatterns(),
		identity:       ProbeIdentity(repoRoot),
		identSev:       DefaultIdentitySeverities(),
		skipExtensions: toSet(defaultSkipExtensions),
		skipFilenames:  toSet(defaultSkipFilenames),
		skipFragments:  append([]string(nil), defaultSkipFragments...),
	}

	cfgPath := filepath.Join(repoRoot, repoConfigRelPath)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil // no override — built-in defaults stand
		}
		s.unavailable = true
		s.unavailReason = "per-repo scanner config unreadable: " + repoConfigRelPath
		return s, nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		s.unavailable = true
		s.unavailReason = "per-repo scanner config is not valid JSON: " + repoConfigRelPath
		return s, nil
	}
	if err := s.mergeConfig(cfg); err != nil {
		s.unavailable = true
		s.unavailReason = err.Error()
		return s, nil
	}
	return s, nil
}

// mergeConfig layers a per-repo override onto the built-in defaults, enforcing
// the severity floor on both bundled patterns and identity kinds.
func (s *Scanner) mergeConfig(cfg Config) error {
	for _, e := range cfg.SkipExtensions {
		// An empty extension entry matches every extensionless file (LICENSE,
		// Makefile, …) and would silently drop them from coverage — drop the
		// entry, not the files.
		if strings.TrimSpace(e) == "" {
			continue
		}
		s.skipExtensions[strings.ToLower(e)] = struct{}{}
	}
	for _, f := range cfg.SkipFilenames {
		s.skipFilenames[f] = struct{}{}
	}
	for _, frag := range cfg.SkipPathFragments {
		// A blank or slash-only fragment is a substring of every logical path,
		// so strings.Contains would skip the whole bundle and zero the scan's
		// coverage — reject it rather than let it disable the scanner.
		if strings.Trim(frag, "/ \t\r\n") == "" {
			continue
		}
		s.skipFragments = append(s.skipFragments, frag)
	}

	floors := defaultPatternFloors()
	byName := map[string]int{}
	for i, p := range s.patterns {
		byName[p.Name] = i
	}
	// Deterministic order over override pattern names.
	names := make([]string, 0, len(cfg.Patterns))
	for name := range cfg.Patterns {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		def := cfg.Patterns[name]
		if def.Regex == "" {
			continue
		}
		expr := def.Regex
		if def.CaseInsensitive {
			expr = "(?i)" + expr
		}
		re, err := regexp.Compile(expr)
		if err != nil {
			// A malformed override regex is a config fault → fail-closed.
			return errUnreadable("pattern " + name + " has an invalid regex")
		}
		sev := def.Severity
		if sev == "" {
			sev = defaultPatternSeverity
		}
		if floor, ok := floors[name]; ok {
			sev = applyFloor(sev, floor)
		} else if !isValidSeverity(sev) {
			sev = defaultPatternSeverity
		}
		kind := def.Kind
		if kind == "" {
			kind = name
		}
		np := Pattern{
			Name: name, Kind: kind, Label: def.Label, Re: re,
			Severity: sev, Suggestion: def.Suggestion,
		}
		if i, ok := byName[name]; ok {
			// Preserve the built-in Skip predicate when overriding a bundled name.
			np.Skip = s.patterns[i].Skip
			s.patterns[i] = np
		} else {
			s.patterns = append(s.patterns, np)
			byName[name] = len(s.patterns) - 1
		}
	}

	for kind, sev := range cfg.IdentitySeverities {
		if !isValidSeverity(sev) {
			continue
		}
		if floor, ok := s.identSev[kind]; ok {
			s.identSev[kind] = applyFloor(sev, floor)
		} else {
			s.identSev[kind] = sev
		}
	}
	return nil
}

// errUnreadable is a tiny error type for config faults.
type errUnreadable string

func (e errUnreadable) Error() string { return string(e) }

// ScanText scans text for every secret pattern and identity-derived match,
// returning findings sorted deterministically. It is pure: identity, patterns
// and severities are all passed in.
func ScanText(text string, id Identity, patterns []Pattern, id2sev map[string]Severity, file string) []Finding {
	if id2sev == nil {
		id2sev = DefaultIdentitySeverities()
	}
	matchers := newIdentityMatchers(id)
	var findings []Finding
	lineno := 0
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		lineno++
		findings = append(findings, matchers.findings(line, lineno, id2sev, file)...)
		for _, cp := range patterns {
			for _, loc := range cp.Re.FindAllStringIndex(line, -1) {
				matched := line[loc[0]:loc[1]]
				if cp.Skip != nil && cp.Skip(matched) {
					continue
				}
				findings = append(findings, Finding{
					File: file, Line: lineno, Column: loc[0] + 1, Kind: cp.Kind,
					Severity: cp.Severity, Snippet: snippet(line), Matched: matched,
					Suggested: cp.Suggestion, line: line,
				})
			}
		}
	}
	sortFindings(findings)
	return findings
}

// ScanBundle scans the resolved content of every bundle file, reading
// ResolvedPath and reporting under LogicalPath. Binary/oversized/skip-listed
// files are skipped via the extension/filename sets and a null-byte + UTF-8
// sniff. If the scanner is unavailable (config unreadable), it returns
// Unavailable=true and scans nothing (fail-closed).
func (s *Scanner) ScanBundle(files []BundleFile) (ScanResult, error) {
	if s.unavailable {
		return ScanResult{Unavailable: true, UnavailableReason: s.unavailReason}, nil
	}
	var res ScanResult
	for _, f := range files {
		if s.skipByName(f.LogicalPath) || s.skipByFragment(f.LogicalPath) {
			continue
		}
		data, err := os.ReadFile(f.ResolvedPath)
		if err != nil {
			continue // unreadable individual file is skipped, not fatal
		}
		if !isText(data) {
			res.Unscanned = append(res.Unscanned, f.LogicalPath)
			continue
		}
		res.FilesScanned++
		res.Findings = append(res.Findings, ScanText(string(data), s.identity, s.patterns, s.identSev, f.LogicalPath)...)
	}
	for _, fnd := range res.Findings {
		if fnd.Severity == SeverityHardFail {
			res.HardFails++
		}
	}
	sortFindings(res.Findings)
	// Zero-coverage sentinel: a bundle with files but nothing scanned (every
	// file skip-listed or unscannable, however that came about — an over-broad
	// skip config, an all-binary tree) means the scanner effectively did not
	// run. Fail closed so the launch path refuses rather than publishing an
	// unscanned bundle while reporting "would publish".
	if len(files) > 0 && res.FilesScanned == 0 && !res.Unavailable {
		res.Unavailable = true
		res.UnavailableReason = "scanner covered zero of " + strconv.Itoa(len(files)) +
			" bundle files (all skip-listed or unscannable)"
	}
	return res, nil
}

func (s *Scanner) skipByName(logical string) bool {
	ext := strings.ToLower(filepath.Ext(logical))
	if _, ok := s.skipExtensions[ext]; ok {
		return true
	}
	base := path_base(logical)
	_, ok := s.skipFilenames[base]
	return ok
}

func (s *Scanner) skipByFragment(logical string) bool {
	l := filepath.ToSlash(logical)
	for _, frag := range s.skipFragments {
		if strings.Contains(l, frag) {
			return true
		}
	}
	return false
}

// path_base returns the final path element of a POSIX/OS logical path.
func path_base(p string) string {
	p = filepath.ToSlash(p)
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// isText sniffs the first 8KB: a null byte or invalid UTF-8 means binary.
func isText(data []byte) bool {
	const sniff = 8192
	chunk := data
	if len(chunk) > sniff {
		chunk = chunk[:sniff]
	}
	for _, b := range chunk {
		if b == 0 {
			return false
		}
	}
	return utf8.Valid(chunk)
}

func toSet(items []string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, it := range items {
		m[it] = struct{}{}
	}
	return m
}

// sortFindings orders findings deterministically: file → line → column → kind.
func sortFindings(f []Finding) {
	sort.SliceStable(f, func(i, j int) bool {
		if f[i].File != f[j].File {
			return f[i].File < f[j].File
		}
		if f[i].Line != f[j].Line {
			return f[i].Line < f[j].Line
		}
		if f[i].Column != f[j].Column {
			return f[i].Column < f[j].Column
		}
		if f[i].Kind != f[j].Kind {
			return f[i].Kind < f[j].Kind
		}
		return f[i].Matched < f[j].Matched
	})
}
