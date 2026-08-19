package lint

import (
	"path/filepath"
	"testing"

	"github.com/Partnermedia/abcd/internal/gittest"
)

// manifestReceipt builds a manifest-era receipt: a PROMOTE for commit, bound to
// detector, carrying the two iss-122 fields (tier, manifestHash) and a raw
// findings array (failingJSON, e.g. `[]` or `[{"disposition":"accepted"}]`).
func manifestReceipt(commit, detector, tier, manifestHash, failingJSON string) string {
	return `{
  "subject": {"digest": {"gitCommit": "` + commit + `"}},
  "verifier": {"id": "maintainer"},
  "verificationResult": "PROMOTE",
  "judgeModel": "claude-opus-4-8",
  "policy": {"detector": "` + detector + `", "version": "1"},
  "tier": "` + tier + `",
  "manifestHash": "` + manifestHash + `",
  "failing": ` + failingJSON + `
}`
}

func runReceiptGate(t *testing.T, root string, cfg RuleConfig) []Finding {
	t.Helper()
	fs, err := Lint(Config{Rules: map[string]RuleConfig{"receipt_gate": cfg}}, root)
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

// The manifest-era refusals fire only when the committed manifest is present in
// the content tree, and none of them judges a finding's content. Run in a plain
// temp dir (no git), so the release impact cannot be classified and the required
// tier fails closed to "full".
func TestReceiptGateManifestEra(t *testing.T) {
	root := t.TempDir()
	const sha = "0123456789abcdef0123456789abcdef01234567"
	const gate = "iss35-brief-surface-crosscheck"
	reviews := filepath.Join(".abcd", "work", "reviews")

	manifestBody := `{"schemaVersion":1,"detector":"` + gate + `","checkerCount":22}` + "\n"
	writeFile(t, root, releaseGateManifestPath, manifestBody)
	goodHash := hashManifest([]byte(manifestBody))

	cfg := RuleConfig{Enabled: true, Severity: severityBlocker, ReceiptsDir: reviews, Commit: sha, RequiredGates: []string{gate}}
	put := func(body string) { writeFile(t, root, filepath.Join(reviews, sha, gate+".json"), body) }

	// (d) A fully conforming full-tier receipt with the right manifest hash and no
	// findings → clean.
	put(manifestReceipt(sha, gate, releaseTierFull, goodHash, `[]`))
	if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 0 {
		t.Fatalf("(d) conforming full receipt should be clean, got %d", n)
	}

	// (a) A receipt whose manifestHash does not match the committed manifest → BLOCK.
	put(manifestReceipt(sha, gate, releaseTierFull, "sha256:deadbeef", `[]`))
	if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 1 {
		t.Fatalf("(a) manifestHash mismatch should fire once, got %d", n)
	}

	// (b, refuse half) A shallow receipt on a release whose required tier is full
	// (impact unclassifiable here → strict full) → BLOCK.
	put(manifestReceipt(sha, gate, releaseTierShallow, goodHash, `[]`))
	if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 1 {
		t.Fatalf("(b) shallow receipt on a full-tier release should fire once, got %d", n)
	}

	// A blank/unknown tier is never sufficient → BLOCK.
	put(manifestReceipt(sha, gate, "", goodHash, `[]`))
	if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 1 {
		t.Fatalf("empty tier should fire once, got %d", n)
	}

	// (c) A finding with NO recorded disposition → BLOCK.
	put(manifestReceipt(sha, gate, releaseTierFull, goodHash, `[{"where":"04-abcd.md:10"}]`))
	if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 1 {
		t.Fatalf("(c) undispositioned finding should fire once, got %d", n)
	}

	// (c) The SAME finding WITH a recorded disposition (any content) → clean.
	put(manifestReceipt(sha, gate, releaseTierFull, goodHash, `[{"where":"04-abcd.md:10","disposition":"accepted: staged row"}]`))
	if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 0 {
		t.Fatalf("(c) dispositioned finding should be clean, got %d", n)
	}
}

// (e) Era gating: with NO committed manifest, a receipt that carries no tier, no
// manifestHash, and an undispositioned finding is judged by the pre-manifest
// rules only — none of the new refusals fire. This is what keeps v0.3.0's and
// v0.4.0's historical receipts valid for their own (pre-manifest) commits.
func TestReceiptGatePreManifestEra(t *testing.T) {
	root := t.TempDir()
	const sha = "0123456789abcdef0123456789abcdef01234567"
	const gate = "iss35-brief-surface-crosscheck"
	reviews := filepath.Join(".abcd", "work", "reviews")

	// No manifest file written → pre-manifest era.
	body := `{
  "subject": {"digest": {"gitCommit": "` + sha + `"}},
  "verificationResult": "PROMOTE",
  "judgeModel": "claude-opus-4-8",
  "policy": {"detector": "` + gate + `"},
  "failing": [{"where":"x"}]
}`
	writeFile(t, root, filepath.Join(reviews, sha, gate+".json"), body)
	cfg := RuleConfig{Enabled: true, Severity: severityBlocker, ReceiptsDir: reviews, Commit: sha, RequiredGates: []string{gate}}
	if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 0 {
		t.Fatalf("(e) a pre-manifest receipt must pass the old rules with no new refusals, got %d", n)
	}
}

// (b, both halves) The required tier is derived from the release's impact class,
// read from the shipped records. A fix-only release (patch) accepts a shallow
// receipt; an additive release (feature) refuses it and needs a full one. This
// is the pre-1.0 case the version number alone cannot decide.
func TestReceiptGateTierByImpact(t *testing.T) {
	const gate = "iss35-brief-surface-crosscheck"
	reviews := ".abcd/work/reviews"
	manifestBody := `{"schemaVersion":1,"detector":"` + gate + `","checkerCount":22}` + "\n"
	goodHash := hashManifest([]byte(manifestBody))

	// build a released repo whose cut since v0.3.0 carries exactly one record of
	// the given impact, then arm the gate against HEAD.
	build := func(t *testing.T, impact string) (root, sha string) {
		r := gittest.NewRepo(t)
		r.Write("CHANGELOG.md", "# Changelog\n\n## [0.3.0] - 2026-01-01\n\n### Added\n\n- base\n")
		r.Commit("base 0.3.0")
		r.Git("tag", "v0.3.0")
		// The release cut: a resolved issue carrying the impact, a newer CHANGELOG
		// heading, the pinned manifest, and the receipt.
		r.Record(".abcd/work/issues/resolved/iss-900-x.md", "iss-900", impact)
		r.Write("CHANGELOG.md", "# Changelog\n\n## [0.3.1] - 2026-02-01\n\n### Fixed\n\n- x\n\n## [0.3.0] - 2026-01-01\n\n### Added\n\n- base\n")
		r.Write(releaseGateManifestPath, manifestBody)
		r.Commit("release 0.3.1")
		return r.Root(), r.Git("rev-parse", "HEAD")
	}

	t.Run("fix release accepts shallow", func(t *testing.T) {
		root, sha := build(t, "fix")
		cfg := RuleConfig{Enabled: true, Severity: severityBlocker, ReceiptsDir: reviews, Commit: sha, RequiredGates: []string{gate}}
		writeFile(t, root, filepath.Join(reviews, sha, gate+".json"), manifestReceipt(sha, gate, releaseTierShallow, goodHash, `[]`))
		if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 0 {
			t.Fatalf("a shallow receipt on a fix (patch) release should be clean, got %d", n)
		}
	})

	t.Run("additive release refuses shallow, accepts full", func(t *testing.T) {
		root, sha := build(t, "additive")
		cfg := RuleConfig{Enabled: true, Severity: severityBlocker, ReceiptsDir: reviews, Commit: sha, RequiredGates: []string{gate}}
		path := filepath.Join(reviews, sha, gate+".json")

		writeFile(t, root, path, manifestReceipt(sha, gate, releaseTierShallow, goodHash, `[]`))
		if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 1 {
			t.Fatalf("a shallow receipt on an additive (feature) release should fire once, got %d", n)
		}

		writeFile(t, root, path, manifestReceipt(sha, gate, releaseTierFull, goodHash, `[]`))
		if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 0 {
			t.Fatalf("a full receipt on an additive release should be clean, got %d", n)
		}
	})

	// An added record whose impact cannot be read must NOT down-tier the release to
	// a patch: the cut is unclassifiable, so the gate falls back to the strict full
	// tier and a shallow receipt is refused (fail-closed within the gate).
	t.Run("unlabelled added record fails closed to full", func(t *testing.T) {
		root, sha := build(t, "notavalidimpact")
		cfg := RuleConfig{Enabled: true, Severity: severityBlocker, ReceiptsDir: reviews, Commit: sha, RequiredGates: []string{gate}}
		writeFile(t, root, filepath.Join(reviews, sha, gate+".json"), manifestReceipt(sha, gate, releaseTierShallow, goodHash, `[]`))
		if n := countRule(runReceiptGate(t, root, cfg), "receipt_gate"); n != 1 {
			t.Fatalf("a shallow receipt on an unclassifiable cut should fire once, got %d", n)
		}
	})
}
