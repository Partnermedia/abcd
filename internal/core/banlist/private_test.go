package banlist

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// match is the test-side stand-in for what the shell hook does with `grep -iE`:
// it compiles each parsed pattern under the documented case-insensitive reading
// and returns the key of the first entry that matches, or "" for a clean text. An
// entry the engine refuses is an error, never a miss — the same fail-safe verdict
// the hook reaches. It lives in the test because production has no matcher: the
// hook is the enforcement point, and this is how its reading is checked.
func match(entries []rawEntry, text string) (string, error) {
	for _, e := range entries {
		if e.malformed {
			return "", fmt.Errorf("banlist line %d is not a usable pattern", e.line)
		}
		re, err := regexp.Compile("(?i)" + e.pattern)
		if err != nil {
			return "", fmt.Errorf("banlist line %d is not a usable pattern", e.line)
		}
		if re.MatchString(text) {
			return e.key, nil
		}
	}
	return "", nil
}

func writePrivate(t *testing.T, root, body string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(PrivateRelPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestParseAgreesWithTheHookOnTheSharedCorpus is the format-agreement test: the
// Go parser reads testdata/parse-corpus.txt — the same file hook_test.go feeds the
// committed shell hook — and must derive the same keys, skip the same comment and
// blank lines, and reach the same verdict on both probe halves.
func TestParseAgreesWithTheHookOnTheSharedCorpus(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "parse-corpus.txt"))
	if err != nil {
		t.Fatal(err)
	}
	entries, malformed := parse(data)
	if len(malformed) != 0 {
		t.Fatalf("malformed lines %v; the shared corpus must parse cleanly on both sides", malformed)
	}
	var keys []string
	for _, e := range entries {
		keys = append(keys, e.key)
	}
	if strings.Join(keys, ",") != strings.Join(corpusKeys, ",") {
		t.Fatalf("keys = %v, want %v", keys, corpusKeys)
	}

	for _, tc := range corpusMustBlock {
		hit, err := match(entries, tc.text)
		if err != nil {
			t.Fatalf("%s: match: %v", tc.name, err)
		}
		if hit == "" {
			t.Errorf("%s: no entry matched; the hook refuses this content", tc.name)
			continue
		}
		if hit != tc.key {
			t.Errorf("%s: matched key %q, want %q (the hook names %q)", tc.name, hit, tc.key, tc.key)
		}
	}
	for _, tc := range corpusMustPass {
		hit, err := match(entries, tc.text)
		if err != nil {
			t.Fatalf("%s: match: %v", tc.name, err)
		}
		if hit != "" {
			t.Errorf("%s: entry %q matched permitted content; the hook lets this commit through", tc.name, hit)
		}
	}
}

// TestAddPrivateCreatesTheStoreAndListsKeysOnly pins AC6 for the private layer:
// the store is created on first add (0600, under the gitignored local tier), and
// nothing a surface can render carries the pattern value. The whole report is
// marshalled and searched for the pattern — the assertion is about the DATA, so a
// later field addition cannot quietly reintroduce the leak.
func TestAddPrivateCreatesTheStoreAndListsKeysOnly(t *testing.T) {
	root := t.TempDir()
	const pattern = `widgetworks`
	res, err := AddPrivate(AddPrivateRequest{RepoRoot: root, Key: "widget-partner", Pattern: pattern})
	if err != nil {
		t.Fatalf("AddPrivate: %v", err)
	}
	if res.Key != "widget-partner" {
		t.Errorf("Key = %q", res.Key)
	}

	path := filepath.Join(root, filepath.FromSlash(PrivateRelPath))
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("store not created: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("store mode = %04o, want 0600 (it holds the private patterns)", perm)
	}

	rep, err := ListPrivate(root)
	if err != nil {
		t.Fatalf("ListPrivate: %v", err)
	}
	if !rep.Present || len(rep.Entries) != 1 || rep.Entries[0].Key != "widget-partner" {
		t.Fatalf("report = %+v", rep)
	}
	blob, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blob), pattern) {
		t.Errorf("the private report carries the pattern value: %s", blob)
	}
}

// TestListPrivateReportsAnAbsentStore pins AC4's Go-side counterpart: absent is a
// reported state, never an empty success that reads like "nothing is banned".
func TestListPrivateReportsAnAbsentStore(t *testing.T) {
	rep, err := ListPrivate(t.TempDir())
	if err != nil {
		t.Fatalf("ListPrivate: %v", err)
	}
	if rep.Present {
		t.Errorf("Present = true for an absent store")
	}
	if rep.Path != PrivateRelPath {
		t.Errorf("Path = %q, want %q", rep.Path, PrivateRelPath)
	}
}

// TestAddPrivateRefusals covers the input contract. Every refusal is checked for
// leakage too: an error message is output, so it may name the key and never the
// pattern.
func TestAddPrivateRefusals(t *testing.T) {
	root := t.TempDir()
	writePrivate(t, root, "widget-partner   widgetworks\n")

	cases := []struct {
		name    string
		key     string
		pattern string
		want    error
		secret  string
	}{
		{"duplicate key", "widget-partner", `otherthing`, ErrDuplicateKey, "otherthing"},
		{"key with a metacharacter", `widget*`, `otherthing`, ErrInvalidKey, "otherthing"},
		{"empty key", "", `otherthing`, ErrInvalidKey, "otherthing"},
		{"empty pattern", "empty", "", ErrInvalidPattern, ""},
		{"uncompilable pattern", "broken", `[unclosed`, ErrInvalidPattern, "unclosed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AddPrivate(AddPrivateRequest{RepoRoot: root, Key: tc.key, Pattern: tc.pattern})
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if tc.secret != "" && strings.Contains(err.Error(), tc.secret) {
				t.Errorf("error message echoes the pattern value: %v", err)
			}
		})
	}
}

// TestRemovePrivateDropsOneLineAndLeavesTheRestByteIdentical pins the edit
// contract: a removal is a line deletion, so hand-written comments, alignment,
// and ordering survive it.
func TestRemovePrivateDropsOneLineAndLeavesTheRestByteIdentical(t *testing.T) {
	root := t.TempDir()
	const body = "# local banlist\n\nwidget-partner   widgetworks\nlab-host         alice-laptop\\.example\\.com\n\nlab-ip           192\\.0\\.2\\.17\n"
	path := writePrivate(t, root, body)

	res, err := RemovePrivate(root, "lab-host")
	if err != nil {
		t.Fatalf("RemovePrivate: %v", err)
	}
	if res.Key != "lab-host" {
		t.Errorf("Key = %q", res.Key)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# local banlist\n\nwidget-partner   widgetworks\n\nlab-ip           192\\.0\\.2\\.17\n"
	if string(got) != want {
		t.Errorf("store after removal:\n%q\nwant:\n%q", got, want)
	}

	if _, err := RemovePrivate(root, "lab-host"); !errors.Is(err, ErrUnknownKey) {
		t.Errorf("removing a gone key: err = %v, want ErrUnknownKey", err)
	}
	if _, err := RemovePrivate(t.TempDir(), "lab-host"); !errors.Is(err, ErrNoStore) {
		t.Errorf("removing from an absent store: err = %v, want ErrNoStore", err)
	}
}

// TestListPrivateReportsMalformedLinesByNumber pins the fail-safe reporting the
// hook implements in shell: an unusable entry is surfaced, by line number only.
func TestListPrivateReportsMalformedLinesByNumber(t *testing.T) {
	root := t.TempDir()
	writePrivate(t, root, "widget-partner   widgetworks\nbad-entry        [unclosed\n")

	rep, err := ListPrivate(root)
	if err != nil {
		t.Fatalf("ListPrivate: %v", err)
	}
	if len(rep.Malformed) != 1 || rep.Malformed[0] != 2 {
		t.Fatalf("Malformed = %v, want [2]", rep.Malformed)
	}
	blob, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blob), "unclosed") {
		t.Errorf("report echoes the malformed line's content: %s", blob)
	}
	// The usable entry is still listed: one bad line does not blind the store.
	if len(rep.Entries) != 2 {
		t.Errorf("Entries = %+v, want both lines keyed (the malformed one by key too)", rep.Entries)
	}
}
