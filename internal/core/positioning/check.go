package positioning

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// maxSurfaceBytes caps a guarded surface read. A README or manifest is small;
// the cap stops a device or runaway file being read unbounded.
const maxSurfaceBytes = 1 << 20

// Status is one surface's outcome.
type Status string

const (
	// StatusOK: the surface carries every block field it declares.
	StatusOK Status = "ok"
	// StatusDrifted: the surface was located but no longer says what the block
	// says.
	StatusDrifted Status = "drifted"
	// StatusAbsent: none of the surface's candidate files exists, so the
	// surface is not present in this repo. Not a finding.
	StatusAbsent Status = "absent"
	// StatusUnlocatable: the file exists but the locator found nothing. A
	// finding: a locator that stopped matching hides drift rather than proving
	// its absence.
	StatusUnlocatable Status = "unlocatable"
)

// SurfaceResult is one surface's comparison against the block.
type SurfaceResult struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
	// File is the candidate that was checked (empty when all were absent).
	File string `json:"file,omitempty"`
	// Line is the 1-based line the located text starts on.
	Line int `json:"line,omitempty"`
	// Found is the located text, verbatim — the exact drifted content.
	Found string `json:"found,omitempty"`
	// Missing names the block fields the surface no longer carries.
	Missing []string `json:"missing,omitempty"`
	// Canonical is what the block says this surface should carry.
	Canonical string `json:"canonical,omitempty"`
}

// Report is a whole positioning check: the block it was held to, the family
// severity, and one result per registered surface.
type Report struct {
	Block    Block           `json:"block"`
	Severity string          `json:"severity"`
	Surfaces []SurfaceResult `json:"surfaces"`
}

// Drifted counts the surfaces that need attention — drifted or unlocatable.
func (r Report) Drifted() int {
	n := 0
	for _, s := range r.Surfaces {
		if s.Status == StatusDrifted || s.Status == StatusUnlocatable {
			n++
		}
	}
	return n
}

// Check parses the identity block and compares every registered surface against
// it. It reads only; nothing is written and nothing is proposed here. An
// unreadable or malformed block is an error, never an empty clean report — a
// check that cannot run must not read as a pass.
func Check(root string, cfg Config) (Report, error) {
	if err := cfg.Validate(); err != nil {
		return Report{}, err
	}
	block, err := ParseBlock(root, cfg.Block)
	if err != nil {
		return Report{}, err
	}
	rep := Report{Block: block, Severity: cfg.severity()}
	for _, s := range cfg.surfaces() {
		res, err := checkSurface(root, s, block)
		if err != nil {
			return Report{}, err
		}
		rep.Surfaces = append(rep.Surfaces, res)
	}
	return rep, nil
}

// checkSurface resolves the surface's file, locates its self-description, and
// compares it to the block.
func checkSurface(root string, s Surface, b Block) (SurfaceResult, error) {
	res := SurfaceResult{ID: s.ID, Status: StatusAbsent, Canonical: s.render(b)}

	rel, data, ok, err := readFirstPresent(root, s.Files)
	if err != nil {
		return SurfaceResult{}, err
	}
	if !ok {
		return res, nil
	}
	res.File = rel

	found, line, located, err := s.locate(data)
	if err != nil {
		return SurfaceResult{}, fmt.Errorf("surface %q in %s: %w", s.ID, rel, err)
	}
	if !located {
		res.Status = StatusUnlocatable
		return res, nil
	}
	res.Found, res.Line = found, line

	hay := normalize(found)
	for _, field := range s.Requires {
		want, _ := b.Field(field)
		if want == "" {
			// An optional field the block does not carry cannot be required of
			// a surface; skip it rather than fabricating drift.
			continue
		}
		if !strings.Contains(hay, needle(want)) {
			res.Missing = append(res.Missing, field)
		}
	}
	res.Status = StatusOK
	if len(res.Missing) > 0 {
		res.Status = StatusDrifted
	}
	return res, nil
}

// readFirstPresent returns the first candidate that exists, its repo-relative
// path, and its contents. Every read is guarded: a surface file in an arbitrary
// audited repo may be a symlink, a device, or enormous.
func readFirstPresent(root string, files []string) (rel string, data []byte, ok bool, err error) {
	for _, f := range files {
		if !fsutil.ValidRelPath(f) {
			return "", nil, false, fmt.Errorf("%w: surface file %q is not a repo-relative path", ErrConfigInvalid, f)
		}
		b, rerr := fsutil.ReadGuarded(filepath.Join(root, filepath.FromSlash(f)), maxSurfaceBytes)
		if rerr != nil {
			if os.IsNotExist(rerr) {
				continue
			}
			// A present-but-unreadable surface (symlinked, oversize, a device)
			// is not an absence: report it rather than skipping it silently.
			return f, nil, false, fmt.Errorf("reading surface %s: %w", f, rerr)
		}
		return f, b, true, nil
	}
	return "", nil, false, nil
}

// locate extracts the surface's self-description and the 1-based line it starts
// on, reporting whether the locator matched at all.
func (s Surface) locate(data []byte) (found string, line int, ok bool, err error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	switch s.Kind {
	case KindJSONField:
		var obj map[string]json.RawMessage
		if uerr := json.Unmarshal([]byte(text), &obj); uerr != nil {
			// A manifest that is not a JSON object cannot be located; that is a
			// locator failure, not a parse crash for the whole audit.
			return "", 0, false, nil
		}
		raw, present := obj[s.Field]
		if !present {
			return "", 0, false, nil
		}
		var v string
		if uerr := json.Unmarshal(raw, &v); uerr != nil {
			return "", 0, false, nil // present but not a string
		}
		return v, jsonFieldLine(text, s.Field), true, nil
	case KindRegexp:
		for _, p := range s.Patterns {
			re, cerr := regexp.Compile(p)
			if cerr != nil {
				return "", 0, false, fmt.Errorf("pattern %q does not compile: %w", p, cerr)
			}
			loc := re.FindStringSubmatchIndex(text)
			if loc == nil || loc[2] < 0 {
				continue
			}
			return text[loc[2]:loc[3]], 1 + strings.Count(text[:loc[2]], "\n"), true, nil
		}
		return "", 0, false, nil
	}
	return "", 0, false, fmt.Errorf("%w: unknown surface kind %q", ErrConfigInvalid, s.Kind)
}

// jsonFieldLine finds the 1-based line a top-level JSON key sits on, so a
// finding can cite file:line. It is a display aid over the already-parsed
// document: a miss yields 0 rather than a wrong line.
func jsonFieldLine(text, field string) int {
	key := `"` + field + `"`
	for i, l := range strings.Split(text, "\n") {
		if strings.Contains(l, key) {
			return i + 1
		}
	}
	return 0
}

var (
	// emphasisRe strips inline markdown emphasis and code markers, so a tagline
	// bolded inside a sentence still reads as the tagline.
	emphasisRe = regexp.MustCompile("[*_`]+")
	// htmlTagRe strips inline HTML tags, so an anchor or span wrapped around
	// part of a strapline does not count as drift.
	htmlTagRe = regexp.MustCompile(`<[^>]*>`)
	// spaceRe collapses every whitespace run, so line wrapping is invisible to
	// the comparison.
	spaceRe = regexp.MustCompile(`\s+`)
	// dashRe folds the dash family and non-breaking spaces onto their plain
	// forms, so typographic polish is not drift.
	dashRe = strings.NewReplacer(
		"—", "-", "–", "-", "‒", "-", "‑", "-", "−", "-",
		" ", " ", "‘", "'", "’", "'", "“", `"`, "”", `"`,
	)
)

// normalize reduces text to the form the comparison is made in: HTML and
// markdown markup stripped, dashes and quotes folded, whitespace collapsed,
// lower-cased. It is deliberately lossy — the check asks whether a surface still
// SAYS the canonical line, not whether it renders it byte for byte.
func normalize(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = emphasisRe.ReplaceAllString(s, "")
	s = dashRe.Replace(s)
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.ToLower(strings.TrimSpace(s))
}

// needle is a block field normalised for containment: terminal sentence
// punctuation is dropped so a tagline that ends a sentence in one surface and
// continues into a clause in another still matches.
func needle(field string) string {
	return strings.TrimRight(normalize(field), " .!?;:,")
}
