package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const ruleDeliveryState = "delivery_state"

var (
	// A CHANGELOG version entry opens with a bracketed version heading —
	// "## [Unreleased]", "## [0.4.1] - 2026-07-28" (Keep a Changelog).
	changelogVersionRe = regexp.MustCompile(`^##\s+\[`)
	// A change-type heading inside a version entry: "### Added", "### Fixed".
	changelogSectionRe = regexp.MustCompile(`^###\s+(.+?)\s*$`)
	// An intent handle as prose or as a code span; the whole entry is scanned, so
	// `itd-60` and itd-60 are one citation.
	changelogIntentRe = regexp.MustCompile(`(?i)\bitd-(\d+)\b`)
)

// deliveryStateSections are the change-type headings that make a DELIVERY claim:
// a capability that is new, or one whose behaviour changed. The other Keep a
// Changelog headings (Fixed, Removed, Security, Breaking, Deprecated) narrate a
// correction, and an intent id under them is normally provenance — which draft
// two branches minted at once — rather than a claim that the intent is built.
var deliveryStateSections = []string{"Added", "Changed"}

// The intents-store bucket whose meaning contradicts a delivery claim: captured,
// uncommitted, no spec (adr-34, intents/README.md § Lifecycle Directories).
const deliveryStateDraftBucket = "drafts"

// checkDeliveryState is the lifecycle half of iss-41. The record represents
// delivered capability by the intent LEAVING drafts/: an intent still sitting
// there is a captured idea nobody has committed to build, so a released entry
// crediting it for shipped capability makes the CHANGELOG and the intent tree say
// opposite things about the same work — and the tree is the one nobody re-reads.
//
// The rule holds one invariant: no delivery entry cites an intent that is still
// in drafts/. Its two remedies are the two truths it could be reporting — the
// intent shipped whole and has not been promoted, or the entry credits an intent
// for something less than it promises, in which case it describes the capability
// without the citation (an intent is delivered whole or not at all, per the
// split-the-intent doctrine).
//
// It fails closed on the three ways the gate could be armed and check nothing: a
// changelog it cannot read, a configured intents store that is not there, and one
// that is there but holds no lifecycle bucket to resolve a citation against.
func checkDeliveryState(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	changelog := cfg.Changelog
	if changelog == "" {
		changelog = "CHANGELOG.md"
	}
	intentsRoot := cfg.IntentsRoot
	if intentsRoot == "" {
		intentsRoot = filepath.Join(".abcd", "development", "intents")
	}
	// The configured list WIDENS the defaults rather than replacing them: a repo
	// naming its own delivery heading is adding a place capability is announced,
	// never revoking Added/Changed, and a config that could silently narrow the
	// gate is a config that can silently disarm it.
	sections := map[string]bool{}
	for _, s := range deliveryStateSections {
		sections[strings.ToLower(strings.TrimSpace(s))] = true
	}
	for _, s := range cfg.DeliverySections {
		sections[strings.ToLower(strings.TrimSpace(s))] = true
	}

	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(changelog)))
	if err != nil {
		return nil, &configError{ruleDeliveryState + ": reading " + changelog + ": " + err.Error()}
	}
	if err := requireIntentBuckets(repoRoot, intentsRoot); err != nil {
		return nil, err
	}
	tree, err := scanIntentTree(repoRoot, repoRoot, filepath.FromSlash(intentsRoot))
	if err != nil {
		return nil, err
	}
	bucket := map[string]string{}
	for _, r := range tree.records {
		// Key on the canonical id spelling so a zero-padded intent filename
		// (itd-047-*.md) and a canonical citation (itd-47) resolve to the same
		// bucket — else the drafts-citation gate silently fails open over exactly
		// the case it exists to catch (iss-392's canonRecordID keyspace).
		bucket[canonRecordID(strings.ToLower(intentIDRe.FindString(r.name)))] = r.bucket
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	mask := fenceMask(lines)
	var out []Finding
	inEntry, section := false, ""
	for i, line := range lines {
		if mask[i] {
			continue
		}
		if changelogVersionRe.MatchString(line) {
			inEntry, section = true, ""
			continue
		}
		if m := changelogSectionRe.FindStringSubmatch(line); m != nil {
			section = strings.ToLower(strings.Trim(m[1], "* "))
			continue
		}
		if !inEntry || !sections[section] {
			continue
		}
		for _, id := range changelogIntentCitations(line) {
			if bucket[id] != deliveryStateDraftBucket {
				continue
			}
			out = append(out, Finding{
				File: changelog, Line: i + 1, RuleID: ruleDeliveryState, Severity: cfg.Severity,
				Message: "delivery entry credits " + id + ", which is still in " + intentsRoot + "/" +
					deliveryStateDraftBucket + " — captured, not built; promote the intent if the capability shipped, " +
					"otherwise describe what shipped without citing it",
			})
		}
	}
	return out, nil
}

// requireIntentBuckets is the third way an armed gate could check nothing: a root
// that exists but holds no lifecycle bucket resolves every citation to "found
// nowhere", which reads exactly like a clean corpus. The likeliest cause is a
// root pointed one level too deep — at a bucket rather than at the store — so the
// state is reported as the misconfiguration it is rather than as a pass.
func requireIntentBuckets(repoRoot, intentsRoot string) error {
	ents, err := os.ReadDir(filepath.Join(repoRoot, filepath.FromSlash(intentsRoot)))
	if err != nil {
		return &configError{ruleDeliveryState + ": reading the intents store " + intentsRoot + ": " + err.Error()}
	}
	for _, e := range ents {
		if e.IsDir() && intentBuckets[e.Name()] {
			return nil
		}
	}
	return &configError{ruleDeliveryState + ": the intents store " + intentsRoot +
		" holds none of the lifecycle buckets " + strings.Join(intentBucketNames, "/") +
		"; every citation would resolve to nothing and the gate would pass on an empty corpus"}
}

// changelogIntentCitations returns the distinct intent ids a line cites, in the
// order they appear, so a line naming one intent twice is one finding.
func changelogIntentCitations(line string) []string {
	var ids []string
	seen := map[string]bool{}
	for _, m := range changelogIntentRe.FindAllStringSubmatch(line, -1) {
		// Canonical spelling so a padded citation (itd-047) resolves against the
		// canonical bucket key — the citation side of the same keyspace fix.
		id := canonRecordID("itd-" + m[1])
		if seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}
