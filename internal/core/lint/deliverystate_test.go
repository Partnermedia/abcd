package lint

import (
	"path/filepath"
	"strings"
	"testing"
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
		"- `abcd audit` — a read-only repo-conformance check (itd-85).",
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
