package cli

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// iss-104 acceptance corpus for extending the unrecognized-input-never-writes
// guard (iss-29) to the intent quoted-text create path. A mistyped intent
// subverb must error with a did-you-mean and file no draft; a genuine multi-word
// draft title whose first word merely resembles a subverb still files. The guard
// is id-aware for intent's itd/spc ids (the faithful-mirror + id-aware choice).

var reIntentDraftFile = regexp.MustCompile(`^itd-\d+-.*\.md$`)

// intentDraftCount walks a repo tree and counts written intent draft files.
func intentDraftCount(t *testing.T, root string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && reIntentDraftFile.MatchString(d.Name()) {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return n
}

// TestIntentTypoSubcommandNeverWrites is the headline: `intent lnk itd-5` (a
// typo for `link` followed by an itd id) must be refused with a did-you-mean and
// must not file a draft. Before the fix it was swallowed as create text.
func TestIntentTypoSubcommandNeverWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out, err := runCLIErr(t, "intent", "lnk", "itd-5")
	if err == nil {
		t.Fatalf("expected an error for the mistyped subcommand, got success:\n%s", out)
	}
	if !strings.Contains(err.Error(), "link") {
		t.Fatalf("expected a did-you-mean pointing at %q, got: %v", "link", err)
	}
	if n := intentDraftCount(t, repo); n != 0 {
		t.Fatalf("a mistyped subcommand filed %d draft(s); it must write nothing", n)
	}
}

// TestIntentTypoLoneTokenNeverWrites covers the lone-token shape: `intent paln`
// (a typo for `plan`, no trailing arg) must be refused, not filed.
func TestIntentTypoLoneTokenNeverWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out, err := runCLIErr(t, "intent", "paln")
	if err == nil {
		t.Fatalf("expected an error for the lone mistyped subcommand, got success:\n%s", out)
	}
	if !strings.Contains(err.Error(), "plan") {
		t.Fatalf("expected a did-you-mean pointing at %q, got: %v", "plan", err)
	}
	if n := intentDraftCount(t, repo); n != 0 {
		t.Fatalf("a mistyped subcommand filed %d draft(s); it must write nothing", n)
	}
}

// TestIntentIdShapeDistinguishesTypoFromProse pins the id-aware widening (the
// maintainer's chosen behaviour): with the SAME near-miss first token, a record
// id (itd/spc) as the second token trips the guard, while a prose second token
// does not — so a regression narrowing recordIDRe back to iss-only would be
// caught here, not silently.
func TestIntentIdShapeDistinguishesTypoFromProse(t *testing.T) {
	// itd/spc second token -> shaped like a subcommand call -> refused.
	for _, id := range []string{"itd-5", "spc-2"} {
		repo := t.TempDir()
		t.Chdir(repo)
		out, err := runCLIErr(t, "intent", "lnk", id)
		if err == nil {
			t.Fatalf("intent lnk %s must be refused as a typoed subcommand, got:\n%s", id, out)
		}
		if n := intentDraftCount(t, repo); n != 0 {
			t.Fatalf("intent lnk %s filed %d draft(s); must write nothing", id, n)
		}
	}
	// Same first token, prose second token -> a genuine title -> still files.
	repo := t.TempDir()
	t.Chdir(repo)
	runCLI(t, "intent", "lnk", "widen", "the", "public", "api")
	if n := intentDraftCount(t, repo); n != 1 {
		t.Fatalf("a prose title after a verb-ish first word wrote %d draft(s), want 1", n)
	}
}

// TestIntentFreeTextTitleStillWrites guards the high-precision contract: a
// genuine multi-word draft title whose first word resembles a subverb but is
// followed by prose (not a record id) still files.
func TestIntentFreeTextTitleStillWrites(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out := runCLI(t, "intent", "plans", "the", "release", "cadence", "for", "next", "quarter", "--json")
	var r struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("intent output not JSON: %v\n%s", err, out)
	}
	if r.ID != "itd-1" {
		t.Fatalf("free-text draft id = %q, want itd-1", r.ID)
	}
	if n := intentDraftCount(t, repo); n != 1 {
		t.Fatalf("free-text draft wrote %d draft(s), want 1", n)
	}
}
