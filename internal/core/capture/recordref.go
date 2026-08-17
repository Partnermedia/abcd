package capture

import (
	"os"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/core/intent"
	"github.com/REPPL/abcd-cli/internal/core/spec"
)

// This file is the shared record-id existence probe (spc-25): both
// capture promote's link mode and capture resolve's provenance flags verify a
// target record exists before anything is written, through one probe.

// intentStoreRelDirs enumerates the intent store's bucket directories,
// repo-relative.
func intentStoreRelDirs() []string {
	dirs := make([]string, 0, len(intent.Buckets))
	for _, b := range intent.Buckets {
		dirs = append(dirs, filepath.Join(intent.IntentsRelDir, b))
	}
	return dirs
}

// specStoreRelDirs enumerates the spec store's status directories,
// repo-relative.
func specStoreRelDirs() []string {
	return []string{
		filepath.Join(spec.SpecsRelDir, spec.StatusOpen),
		filepath.Join(spec.SpecsRelDir, spec.StatusClosed),
	}
}

// findRecordFile probes relDirs (repo-relative) for a file belonging to id,
// using the same match rule findIssue applies to the ledger: an exact
// `<id>.md`, or an id-prefixed `<id>-<slug>.md`. Purely a filename existence
// probe — no content is read (a sha's worth of validation lives with the
// record's own store); the returned path is repo-relative. The caller is
// expected to have regex-validated id before it reaches a path.
func findRecordFile(repoRoot string, relDirs []string, id string) (string, bool) {
	exact := id + ".md"
	prefix := id + "-"
	for _, rel := range relDirs {
		entries, err := os.ReadDir(filepath.Join(repoRoot, rel))
		if err != nil {
			continue // absent store/bucket is soft, like the ledger scan
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			n := e.Name()
			if n == exact || (len(n) > len(prefix) && n[:len(prefix)] == prefix && filepath.Ext(n) == ".md") {
				return filepath.Join(rel, n), true
			}
		}
	}
	return "", false
}
