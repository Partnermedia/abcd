package fsutil

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// nonCanonicalPrimitiveRe matches a private redefinition of a durable-write or
// real-dir primitive — the exact names iss-32 consolidates. The canonical home
// is internal/fsutil (exported WriteFileAtomic / IsRealDir); any lowercase
// redefinition elsewhere is a divergent copy.
var nonCanonicalPrimitiveRe = regexp.MustCompile(`func\s+(writeFileAtomic|durableWrite|isRealDir)\b`)

// TestNoNonCanonicalAtomicWritePrimitives is the one-canonical-primitive
// detector: no package under internal/ (other than fsutil) may declare its own
// named atomic-write or real-dir primitive. It matches top-level function
// declarations by name (not inline temp+rename sequences), which is what the
// consolidation removes. It walks the internal/ tree from this package's
// directory; before the consolidation it flags four copies (ahoy
// marker.go/store.go, capture roots.go, memory writer.go).
func TestNoNonCanonicalAtomicWritePrimitives(t *testing.T) {
	internalRoot := filepath.Join("..") // internal/fsutil -> internal/
	var offenders []string
	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// The canonical home is exempt; it defines the real thing.
			if filepath.Base(path) == "fsutil" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range nonCanonicalPrimitiveRe.FindAllStringSubmatch(string(data), -1) {
			offenders = append(offenders, path+": func "+m[1])
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/: %v", err)
	}
	if len(offenders) > 0 {
		t.Fatalf("non-canonical atomic-write/real-dir primitives (route through internal/fsutil):\n  %s",
			strings.Join(offenders, "\n  "))
	}
}
