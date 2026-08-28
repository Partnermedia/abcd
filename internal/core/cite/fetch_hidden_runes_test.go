package cite

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestRedirectTargetControlRunesNotRecordedRaw pins the fetch boundary
// (iss-359): net/url preserves raw non-ASCII in a Location query, and
// encoding/json escapes none of DEL/C1/bidi/zero-width, so a hostile redirect
// could land a C1 escape or a bidi override raw in `docs cite refresh --json`
// and in the committed citations baseline. The runes are assembled numerically
// — a raw \x9b in a fixture decodes to U+FFFD and tests nothing (the repo's
// own vacuous-test trap).
func TestRedirectTargetControlRunesNotRecordedRaw(t *testing.T) {
	hazard := string(rune(0x009B)) + "2K" + string(rune(0x202E)) + "x" + string(rune(0x200B))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old" {
			w.Header().Set("Location", "/new?ref="+hazard)
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	got := testChecker(5 * time.Second).Check(srv.URL + "/old")
	if got.Status != StatusOK {
		t.Fatalf("status = %q (%s), want %q", got.Status, got.Detail, StatusOK)
	}
	for _, r := range []rune{0x009B, 0x202E, 0x200B} {
		if strings.ContainsRune(got.FinalURL, r) {
			t.Errorf("FinalURL carries U+%04X raw: %q", r, got.FinalURL)
		}
	}
	if !strings.Contains(got.FinalURL, "%C2%9B") {
		t.Errorf("FinalURL should percent-encode the hazard losslessly, got %q", got.FinalURL)
	}
}

// The direct unit contract for the percent-encoder — invalid bytes kept raw,
// clean inputs untouched — now lives with the primitive itself in
// internal/termsafe (TestEncodeHiddenRunesContract), which both this package and
// internal/core/memory route through. TestRedirectTargetControlRunesNotRecordedRaw
// above pins the fetch boundary that calls it.
