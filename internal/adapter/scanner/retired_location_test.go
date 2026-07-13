package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoRetiredLogbookLocationInSource is the iss-73 detector: `.abcd/logbook/`
// is a retired runtime-output location (iss-36/iss-56). A 2026-07-12 maintainer
// adjudication placed runtime artefacts in the gitignored `.abcd/.work.local/logs/`
// tier instead, so no non-test Go source under internal/ may name the retired
// `logbook` location — not memory's lint-report dir, not the scanner's skip
// fragments. It walks internal/ from this package (internal/adapter/scanner).
func TestNoRetiredLogbookLocationInSource(t *testing.T) {
	internalRoot := filepath.Join("..", "..") // internal/adapter/scanner -> internal/
	var offenders []string
	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "logbook") {
			offenders = append(offenders, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/: %v", err)
	}
	if len(offenders) > 0 {
		t.Fatalf("retired '.abcd/logbook' location named in Go source (relocate to .abcd/.work.local/logs/):\n  %s",
			strings.Join(offenders, "\n  "))
	}
}
