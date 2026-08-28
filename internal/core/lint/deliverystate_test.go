package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// deliveryCfg builds a one-rule delivery_state config pointed at a fixture tree.
func deliveryCfg() Config {
	return Config{Rules: map[string]RuleConfig{
		ruleDeliveryState: {Enabled: true, Severity: severityBlocker},
	}}
}

// writeIntent puts an intent file in a lifecycle bucket. Only the location
// matters to this rule — the bucket IS the lifecycle state.
func writeIntent(t *testing.T, root, bucket, name string) {
	t.Helper()
	writeFile(t, root, filepath.Join(".abcd", "development", "intents", bucket, name),
		"---\nid: "+intentIDRe.FindString(name)+"\n---\n\n# fixture\n")
}

// TestDeliveryStateCatchesDraftCitation reproduces the iss-41 corpus shape: a
// released Added entry crediting an intent that is still on the uncommitted
// bench, beside a sibling entry whose intent has left drafts/ and must not fire.
func TestDeliveryStateCatchesDraftCitation(t *testing.T) {
	root := t.TempDir()
	writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")
	writeIntent(t, root, "planned", "itd-73-derived-versioning.md")
	writeFile(t, root, "CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"## [v0.1.0] - 2026-07-07",
		"",
		"### Added",
		"",
		"- `abcd docs lint` (itd-60 layer 1) — a deterministic docs-currency gate.",
		"- Derived-versioning design record (intent itd-73 + ADR-31).",
	}, "\n")+"\n")

	fs, err := Lint(deliveryCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleDeliveryState); n != 1 {
		t.Fatalf("expected exactly the drafts-stage citation to fire, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, "CHANGELOG.md", ruleDeliveryState, 7) {
		t.Errorf("expected the finding on the citing line 7; got %+v", fs)
	}
	if !messageContains(fs, "itd-60") {
		t.Errorf("expected the message to name the cited intent; got %+v", fs)
	}
}

// TestDeliveryStatePaddedIntentFilename pins the canonical-id keyspace on the
// delivery-state gate in both directions: a zero-padded intent filename
// (itd-047-*.md) cited as itd-47, and a canonically-named intent cited padded
// (itd-048). Before the canonRecordID keyspace fix the bucket map keyed on the
// raw filename spelling and the citation on the raw digits, so the drafts
// citation resolved to "" and the gate failed open over exactly the case it
// exists to catch.
func TestDeliveryStatePaddedIntentFilename(t *testing.T) {
	root := t.TempDir()
	// Padded filename, canonical citation.
	writeFile(t, root, filepath.Join(".abcd", "development", "intents", "drafts", "itd-047-padded.md"),
		"---\nid: itd-47\n---\n\n# fixture\n")
	// Canonical filename, padded citation.
	writeFile(t, root, filepath.Join(".abcd", "development", "intents", "drafts", "itd-48-canon.md"),
		"---\nid: itd-48\n---\n\n# fixture\n")
	writeFile(t, root, "CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"## [v0.1.0] - 2026-07-07",
		"",
		"### Added",
		"",
		"- Padded intent file cited canonically (itd-47).",
		"- Canonical intent file cited padded (itd-048).",
	}, "\n")+"\n")

	fs, err := Lint(deliveryCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleDeliveryState); n != 2 {
		t.Fatalf("expected both padded-spelling drafts citations to fire, got %d: %+v", n, fs)
	}
}

// TestDeliveryStateIgnoresNonDeliverySections is the precision half: an id named
// under Fixed is provenance for a defect (which draft two branches minted at
// once), not a claim that the intent is built, and must not fire.
func TestDeliveryStateIgnoresNonDeliverySections(t *testing.T) {
	root := t.TempDir()
	writeIntent(t, root, "drafts", "itd-82-drain-ledger-triage.md")
	writeIntent(t, root, "drafts", "itd-83-review-bar-fires-itself.md")
	writeFile(t, root, "CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"## [0.2.0] - 2026-07-17",
		"",
		"### Fixed",
		"",
		"- Two intents both claimed `itd-82`; the later claimant renumbered to `itd-83`.",
	}, "\n")+"\n")

	fs, err := Lint(deliveryCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleDeliveryState); n != 0 {
		t.Fatalf("expected a non-delivery section to be silent, got %d: %+v", n, fs)
	}
}

// TestDeliveryStateSectionsWiden proves the knob only ever ADDS: a repo that
// files delivery under its own heading gates it by naming it, and the built-in
// Added/Changed keep gating alongside. A configured list that SUBSTITUTED for the
// defaults would let one added heading quietly retire the two that matter.
func TestDeliveryStateSectionsWiden(t *testing.T) {
	root := t.TempDir()
	writeIntent(t, root, "drafts", "itd-85-audit-verb.md")
	writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")
	writeFile(t, root, "CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"## [0.2.0] - 2026-07-17",
		"",
		"### Shipped",
		"",
		"- `abcd lint` — a read-only repo-conformance check (itd-85).",
		"",
		"### Added",
		"",
		"- `abcd docs lint` — a deterministic docs-currency gate (itd-60).",
	}, "\n")+"\n")

	cfg := deliveryCfg()
	rc := cfg.Rules[ruleDeliveryState]
	rc.DeliverySections = []string{"Shipped"}
	cfg.Rules[ruleDeliveryState] = rc

	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleDeliveryState); n != 2 {
		t.Fatalf("expected the configured section AND the built-in defaults to gate, got %d: %+v", n, fs)
	}
	if !messageContains(fs, "itd-85") {
		t.Errorf("expected the configured section to gate; got %+v", fs)
	}
	if !messageContains(fs, "itd-60") {
		t.Errorf("expected the built-in Added section to keep gating; got %+v", fs)
	}
}

// TestDeliveryStateSkipsPreambleAndFences pins the two places a match is not a
// delivery claim: the file preamble ahead of any version entry, and a fenced
// code block showing an entry as an example.
func TestDeliveryStateSkipsPreambleAndFences(t *testing.T) {
	root := t.TempDir()
	writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")
	writeFile(t, root, "CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"### Added",
		"",
		"- This heading precedes every version entry: itd-60.",
		"",
		"## [Unreleased]",
		"",
		"### Added",
		"",
		"```md",
		"- an example entry citing itd-60",
		"```",
	}, "\n")+"\n")

	fs, err := Lint(deliveryCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleDeliveryState); n != 0 {
		t.Fatalf("expected the preamble and the fenced example to be silent, got %d: %+v", n, fs)
	}
}

// TestDeliveryStateGatesUnreleased proves the rule is forward-looking: the
// entry that stops the NEXT stale citation is the one not yet released.
func TestDeliveryStateGatesUnreleased(t *testing.T) {
	root := t.TempDir()
	writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")
	writeFile(t, root, "CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"## [Unreleased]",
		"",
		"### Changed",
		"",
		"- The docs gate grows a second layer (itd-60).",
	}, "\n")+"\n")

	fs, err := Lint(deliveryCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleDeliveryState); n != 1 {
		t.Fatalf("expected an unreleased delivery entry to gate, got %d: %+v", n, fs)
	}
}

// TestDeliveryStateFailsClosed covers the two ways an armed gate could check
// nothing at all: no changelog to read, and no intents store to resolve against.
func TestDeliveryStateFailsClosed(t *testing.T) {
	t.Run("changelog missing", func(t *testing.T) {
		root := t.TempDir()
		writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")
		if _, err := Lint(deliveryCfg(), root); err == nil {
			t.Fatal("expected an error when the configured changelog is absent")
		}
	})

	t.Run("intents store missing", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "CHANGELOG.md", "# Changelog\n\n## [Unreleased]\n\n### Added\n\n- itd-60.\n")
		if _, err := Lint(deliveryCfg(), root); err == nil {
			t.Fatal("expected an error when the configured intents store is absent")
		}
	})

	// A root pointed one level too deep — at a bucket rather than at the store —
	// exists, reads cleanly, and resolves every citation to nothing. That is the
	// shape a silent pass takes, so it is the shape the guard must catch.
	t.Run("intents store holds no lifecycle bucket", func(t *testing.T) {
		root := t.TempDir()
		writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")
		writeFile(t, root, "CHANGELOG.md", "# Changelog\n\n## [Unreleased]\n\n### Added\n\n- itd-60.\n")

		cfg := deliveryCfg()
		rc := cfg.Rules[ruleDeliveryState]
		rc.IntentsRoot = ".abcd/development/intents/drafts"
		cfg.Rules[ruleDeliveryState] = rc

		if _, err := Lint(cfg, root); err == nil {
			t.Fatal("expected an error when the configured intents store holds no lifecycle bucket")
		}
	})
}

// The delivery_state changelog read is the SIBLING of the agent_contract one:
// same in-tree config supplying the path, same repo supplying the file, same
// `go run ./cmd/record-lint` step reading both, and (before this change) the same
// bare os.ReadFile at the end of it. Both rules are armed as blockers in this
// repository's own .abcd/record-lint.json, so one guarded read and one unguarded
// one leaves the step exposed either way.

// A changelog that is a symlink to a character device must be refused, not read.
// Unguarded, this never returned: it allocated until the CI runner was out of
// memory.
func TestDeliveryStateRefusesADeviceChangelog(t *testing.T) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		t.Skip("no /dev/zero on this platform")
	}
	root := t.TempDir()
	writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")
	if err := os.Symlink("/dev/zero", filepath.Join(root, "CHANGELOG.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := Lint(deliveryCfg(), root)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("the lint read a character device as the changelog")
		}
	case <-time.After(20 * time.Second):
		t.Fatal("the lint hung reading a character device as the changelog")
	}
}

// A configured changelog that resolves outside the repository must be refused
// before it is read. The rule already failed closed on a path it could not read,
// which hid this: point it at a file that IS there and the read succeeded, so the
// gate was judging a document outside the checkout.
func TestDeliveryStateRefusesAnEscapingChangelogPath(t *testing.T) {
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "CHANGELOG.md"),
		[]byte("# Changelog\n\n## [Unreleased]\n\n### Added\n\n- nothing.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	writeIntent(t, root, "drafts", "itd-60-doc-fidelity-anti-drift.md")

	rel, err := filepath.Rel(root, filepath.Join(outside, "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := deliveryCfg()
	rc := cfg.Rules[ruleDeliveryState]
	rc.Changelog = filepath.ToSlash(rel)
	cfg.Rules[ruleDeliveryState] = rc

	_, err = Lint(cfg, root)
	if err == nil {
		t.Fatal("a changelog outside the repository was read")
	}
	if !strings.Contains(err.Error(), "the lint reads only inside the repository") {
		t.Errorf("expected the containment refusal its sibling read gives; got %v", err)
	}
}
