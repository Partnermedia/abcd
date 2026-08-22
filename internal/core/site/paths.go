package site

// Small shared helpers: repo-relative path joining, record-handle ordering, and
// the committed unresolved-reference baseline.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// BaselineRelPath is the committed unresolved-reference ratchet.
const BaselineRelPath = ".abcd/site-baseline.json"

const maxBaselineBytes = 256 * 1024

// ErrBaselineInvalid is returned for a baseline the build cannot read.
var ErrBaselineInvalid = errors.New("site: reference baseline is invalid")

// Baseline is `.abcd/site-baseline.json`.
type Baseline struct {
	SchemaVersion        int             `json:"schema_version"`
	UnresolvedReferences []BaselineEntry `json:"unresolved_references"`
}

// BaselineEntry is one reference the baseline admits.
type BaselineEntry struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// LoadBaseline reads the ratchet at rel, or at the default path when rel is
// empty.
//
// The distinction between "named" and "defaulted" is the whole point. A
// repository that declares no baseline simply has none, and the build carries on
// (found=false). A repository that NAMES one in its manifest and does not carry
// it is a different situation: the ratchet it thinks it armed is not being
// measured against anything, and reporting a count of zero would read as good
// news. That refuses.
func LoadBaseline(repoRoot, rel string) (Baseline, bool, error) {
	named := rel != ""
	if !named {
		rel = BaselineRelPath
	}
	data, err := fsutil.ReadGuarded(joinRepo(repoRoot, rel), maxBaselineBytes)
	if err != nil {
		if os.IsNotExist(err) {
			if named {
				return Baseline{}, false, fmt.Errorf("%w: %s is declared as checks.unresolved_reference_baseline but the repository does not carry it",
					ErrBaselineInvalid, rel)
			}
			return Baseline{}, false, nil
		}
		return Baseline{}, false, err
	}
	var b Baseline
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&b); err != nil {
		return Baseline{}, false, fmt.Errorf("%w: %s: %v", ErrBaselineInvalid, rel, err)
	}
	if b.SchemaVersion != 1 {
		return Baseline{}, false, fmt.Errorf("%w: %s: schema_version is %d, want 1", ErrBaselineInvalid, rel, b.SchemaVersion)
	}
	return b, true, nil
}

// baselineCount is the size of the ratchet the manifest names, or zero where
// the repository declares none.
func baselineCount(repoRoot, rel string) (int, error) {
	b, ok, err := LoadBaseline(repoRoot, rel)
	if err != nil || !ok {
		return 0, err
	}
	return len(b.UnresolvedReferences), nil
}

// joinRepo resolves a repo-relative, slash-separated path against a repo root.
func joinRepo(repoRoot, rel string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(rel))
}

// handleNum is the number in a record handle (adr-47 -> 47).
func handleNum(id string) int {
	i := strings.LastIndex(id, "-")
	if i < 0 {
		return 0
	}
	n, err := strconv.Atoi(id[i+1:])
	if err != nil {
		return 0
	}
	return n
}
