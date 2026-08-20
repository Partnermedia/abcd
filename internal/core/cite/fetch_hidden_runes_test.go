package cite

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestRedirectTargetControlRunesNotRecordedRaw pins the fetch boundary
// (iss-345): net/url preserves raw non-ASCII in a Location query, and
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

// An invalid UTF-8 byte in the final address must be percent-encoded raw, not
// rewritten to U+FFFD: net/url preserves bytes over 0x20 in a query verbatim,
// and losing the byte would contradict the recorded-losslessly contract.
func TestEncodeHiddenRunesKeepsInvalidBytes(t *testing.T) {
	in := "https://example.com/a?b=" + string([]byte{0xFF}) + "c"
	got := encodeHiddenRunes(in)
	if !strings.Contains(got, "%FF") {
		t.Errorf("raw 0xFF should percent-encode to %%FF, got %q", got)
	}
	if strings.ContainsRune(got, '�') {
		t.Errorf("invalid byte was rewritten to U+FFFD: %q", got)
	}
	// Already-encoded and clean inputs pass through untouched.
	for _, clean := range []string{"https://example.com/a?b=%E2%80%8B", "https://example.com/plain"} {
		if out := encodeHiddenRunes(clean); out != clean {
			t.Errorf("clean input changed: %q -> %q", clean, out)
		}
	}
}
