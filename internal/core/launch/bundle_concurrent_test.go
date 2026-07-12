package launch

import (
	"fmt"
	"sync"
	"testing"
)

// TestGlobToRegexpConcurrent exercises the compiled-glob cache from many
// goroutines with distinct patterns, forcing concurrent cache writes. Under
// `go test -race` (part of `make preflight`) this fails on the unsynchronised
// package-level map the fix guards with globRegexpMu (iss-31).
func TestGlobToRegexpConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			pat := fmt.Sprintf("dir%d/**/*.md", n)
			re, _ := globToRegexp(pat)
			_ = re.MatchString(fmt.Sprintf("dir%d/a/b.md", n))
		}(i)
	}
	wg.Wait()
}
