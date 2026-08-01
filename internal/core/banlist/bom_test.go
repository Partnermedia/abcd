package banlist

import (
	"errors"
	"strings"
	"testing"
)

// bom is the UTF-8 byte-order mark. An editor that writes one silently is common
// enough that a store's first line arriving with it is an accident, not an attack —
// which is exactly why the readers must agree about what it means.
const bom = "\xef\xbb\xbf"

// TestBOMDoesNotDowngradeAKeyedStore is the fail-open both readers shared: a
// leading BOM made the first line differ from the declaration, so a keyed store
// read as LEGACY. Every keyed line then became one whole-line pattern
// (`lab-host carol-server\.example\.net`) which matches nothing a commit contains,
// while the entry count stayed >=1 so the zero-entry warning never fired. The
// store looked healthy and checked nothing.
func TestBOMDoesNotDowngradeAKeyedStore(t *testing.T) {
	entries, keyed, err := parse([]byte(bom + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n"))
	if err != nil {
		t.Fatalf("a BOM'd declaration must be accepted, not refused: %v", err)
	}
	if !keyed {
		t.Fatal("a BOM before the declaration downgraded the store to legacy")
	}
	if len(entries) != 1 || entries[0].key != "lab-host" {
		t.Fatalf("entries = %+v; want the one keyed entry", entries)
	}
}

// TestCorruptedDeclarationIsRefused: a first line that is not the declaration but
// would be after stripping leading blanks is a DAMAGED declaration, never a legacy
// store. Reading it as legacy is the same silent downgrade the BOM caused, so it
// fails closed — the guard refuses every commit and the verbs refuse to write.
func TestCorruptedDeclarationIsRefused(t *testing.T) {
	for _, prefix := range []string{" ", "\t", "  \t"} {
		_, _, err := parse([]byte(prefix + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n"))
		if !errors.Is(err, ErrMalformedStore) {
			t.Errorf("first line %q: err = %v; want ErrMalformedStore", prefix+privateFormatDecl, err)
		}
	}
}

// TestGenuineLegacyStoreStillReadsAsLegacy keeps the compatibility the refusal
// must not cost: a first line that is nothing like the declaration is an ordinary
// legacy store, and every line in it stays a whole-line pattern.
func TestGenuineLegacyStoreStillReadsAsLegacy(t *testing.T) {
	// One identifier per SOURCE line. Two reserved fixture hosts joined by a literal
	// newline escape on a single line read as ONE token whose first label is neither
	// persona's, so the repo's own privacy-hygiene rule flags a pair of values that
	// are individually fine.
	legacy := "carol-server\\.example\\.net\n" +
		"alice-laptop\n"
	entries, keyed, err := parse([]byte(legacy))
	if err != nil {
		t.Fatalf("a legacy store must still parse: %v", err)
	}
	if keyed {
		t.Fatal("a store with no declaration read as keyed")
	}
	if len(entries) != 2 || entries[0].pattern != "carol-server\\.example\\.net" {
		t.Fatalf("entries = %+v; want two whole-line patterns", entries)
	}
}

// TestListPrivateRefusesACorruptedDeclaration: the read surface must not render a
// store the guard refuses to use as though it were merely legacy — during an
// incident the two answers send a reader in opposite directions.
func TestListPrivateRefusesACorruptedDeclaration(t *testing.T) {
	repo := t.TempDir()
	writePrivate(t, repo, " "+privateFormatDecl+"\nlab-host carol-server\\.example\\.net\n")
	_, err := ListPrivate(repo)
	if !errors.Is(err, ErrMalformedStore) {
		t.Fatalf("ListPrivate err = %v; want ErrMalformedStore", err)
	}
	if strings.Contains(err.Error(), "carol-server") {
		t.Errorf("the refusal echoes store content: %v", err)
	}
}

// TestDeclarationMustBeLineOne closes the two remaining silent downgrades. One
// inserted blank line above the declaration, or a comment moved above it, left the
// store reading as LEGACY — every keyed entry becoming a whole-line pattern that
// matches nothing — with no warning, because the entry count stayed >=1. The
// declaration is only ever read from line 1, so finding it further down the leading
// header is an edit that broke the file, not a legacy store.
func TestDeclarationMustBeLineOne(t *testing.T) {
	for name, body := range map[string]string{
		"blank line above":   "\n" + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n",
		"comment above":      "# my notes\n" + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n",
		"blank and comment":  "\n# my notes\n" + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n",
		"non-utf8 bom above": "\xff\xfe" + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n",
		"space then bom":     " " + bom + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := parse([]byte(body)); !errors.Is(err, ErrMalformedStore) {
				t.Errorf("err = %v; want ErrMalformedStore", err)
			}
		})
	}
}

// TestDeclarationTextBelowTheHeaderIsJustAComment keeps the refusal from
// over-reaching: once a real entry line has been seen the header is over, and a
// later line that happens to spell the declaration is an ordinary comment. A store
// that refused on that would be unusable for anyone who documented their own file.
func TestDeclarationTextBelowTheHeaderIsJustAComment(t *testing.T) {
	body := privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n" + privateFormatDecl + "\n"
	entries, keyed, err := parse([]byte(body))
	if err != nil {
		t.Fatalf("a declaration-shaped comment below the entries must not refuse: %v", err)
	}
	if !keyed || len(entries) != 1 {
		t.Fatalf("keyed=%v entries=%+v; want the one keyed entry", keyed, entries)
	}
}

// TestKeyedStoreRefusesADuplicatedDeclaration is the two-reader divergence. The
// header scan was gated on the store NOT already being keyed, so a keyed store with
// the declaration repeated on line 2 read as healthy to the Go parser — "present,
// keyed, 1 entry" on the status board — while the shell hook blocked every commit.
// A store one reader calls healthy and the other refuses is worse than either
// verdict alone: the board sends you to look at a file the guard has already
// rejected.
func TestKeyedStoreRefusesADuplicatedDeclaration(t *testing.T) {
	body := privateFormatDecl + "\n" + privateFormatDecl + "\nlab-host carol-server\\.example\\.net\n"
	if _, _, err := parse([]byte(body)); !errors.Is(err, ErrMalformedStore) {
		t.Fatalf("err = %v; want ErrMalformedStore", err)
	}
}

// TestParseRefusesTheBOMHeaderCorpus is the last reader split. The header scan did
// not strip a UTF-8 BOM from line 1 the way the entry loop below it does, so a BOM'd
// FIRST line stopped the scan (it looked like neither a blank nor a comment) and a
// declaration on line 2 went unseen: Go called the file a healthy legacy store while
// the shell refused every commit. Its shell half is
// TestPreCommitHook_BOMHeaderCorpusRefuses — one fixture, both readers.
func TestParseRefusesTheBOMHeaderCorpus(t *testing.T) {
	_, _, err := parse(readCorpus(t, "parse-corpus-bom-header.txt"))
	if !errors.Is(err, ErrMalformedStore) {
		t.Fatalf("err = %v; want ErrMalformedStore", err)
	}
	if strings.Contains(err.Error(), "alice-laptop") {
		t.Errorf("the refusal echoes store content: %v", err)
	}
}
