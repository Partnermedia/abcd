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

// LoadBaseline reads the committed ratchet. A repository without one is a state:
// found is false and the build carries on.
func LoadBaseline(repoRoot string) (Baseline, bool, error) {
	data, err := fsutil.ReadGuarded(joinRepo(repoRoot, BaselineRelPath), maxBaselineBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return Baseline{}, false, nil
		}
		return Baseline{}, false, err
	}
	var b Baseline
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&b); err != nil {
		return Baseline{}, false, fmt.Errorf("%w: %s: %v", ErrBaselineInvalid, BaselineRelPath, err)
	}
	if b.SchemaVersion != 1 {
		return Baseline{}, false, fmt.Errorf("%w: %s: schema_version is %d, want 1", ErrBaselineInvalid, BaselineRelPath, b.SchemaVersion)
	}
	return b, true, nil
}

// baselineCount is the size of the committed ratchet, or zero where there is
// none.
func baselineCount(repoRoot string) (int, error) {
	b, ok, err := LoadBaseline(repoRoot)
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

// handleHead is the store prefix of a record handle (adr-47 -> adr).
func handleHead(id string) string {
	if i := strings.LastIndex(id, "-"); i >= 0 {
		if _, err := strconv.Atoi(id[i+1:]); err == nil {
			return id[:i]
		}
	}
	return id
}

// handleLess orders two record handles by store, then NUMERICALLY, so adr-9
// precedes adr-10 — which a string sort gets backwards, visibly, in every
// rendered list.
func handleLess(a, b string) bool {
	pa, pb := handleHead(a), handleHead(b)
	if pa != pb {
		return pa < pb
	}
	na, nb := handleNum(a), handleNum(b)
	if na != nb {
		return na < nb
	}
	return a < b
}
