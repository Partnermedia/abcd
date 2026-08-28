package lint

// The cross-store family (cross_store_id_claim): a record id claimed by a
// document that is not in the store that id belongs to.
//
// record_schema reasons across the stores, but only INSIDE them: a file outside
// every configured store is not a malformed record to the engine, it is not a
// record at all. So a markdown file can head itself `# ADR-23`, declare a status,
// reuse an id a real ADR already holds, and pass the whole gate at exit 0 with
// zero findings. That was not hypothetical — it was measured by probe, and the
// known instance survived a full architecture change (iss-2608230752354926).
//
// The rule fires on a PAIR of signals, never on one:
//
//   - the document's H1 OPENS with a record handle, which is a claim on that id
//     rather than a mention of it; and
//   - the body carries the decision shape (`## Status`, `Status: Accepted`); and
//   - the id is already HELD by a real record in the corresponding store.
//
// The pair is what grandfathers the undated Phase 0 notes without an allowlist of
// them. A filename that reads like an ordinal (`01-harness-interface.md`) claims
// nothing; a reading note that names a decision in prose claims nothing; a
// meeting note with a Status block claims no id. None of them fires, and none of
// them had to be enumerated to be spared — which matters, because an allowlist of
// grandfathered files is a list that grows silently.
//
// It is a STRUCTURAL rule and so does not consult contentExempt, for the reason
// record_schema does not: an id collision is not a question of how a document is
// WRITTEN. The exempt paths excuse the historical record from the
// content-authoring rules; being historical cannot excuse a document from
// claiming an id another record answers to, and the known instance lives in
// exactly such a tree — exempting it would disarm the rule over the corpus that
// motivated it.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/intentdriven/abcd/internal/fsutil"
)

const ruleCrossStoreIDClaim = "cross_store_id_claim"

var (
	// A record handle OPENING a heading: `# ADR-23: …`, `# itd-7 …`. Anchored,
	// because the anchor is what parts a CLAIM from a mention — "# How adr-23 was
	// decided" is a document about a decision, not a second copy of it.
	claimedHandleRe = regexp.MustCompile(`(?i)^(adr|itd|iss|spc)-(\d+)\b`)
	// The decision shape: a `Status` heading (the ADR template's own), or a
	// `Status:` key line in either the bold or the bare spelling. Frontmatter is
	// not read for this — a `status:` field is ordinary record metadata on plenty
	// of documents, while a Status SECTION is how a decision announces itself.
	decisionStatusHeadingRe = regexp.MustCompile(`(?i)^\s*#{1,6}\s+status\s*$`)
	decisionStatusKeyRe     = regexp.MustCompile(`(?i)^\s*\**status\**\s*:(.*)$`)
	// The status VALUE has to be a record LIFECYCLE state, not merely a line that
	// says "status". That distinction is what parts the two populations the corpus
	// actually contains. A design plan heads itself with the record it plans for
	// (`# itd-3 modular rules loader`) and carries a status of its own
	// (`**Status:** design recorded 2026-07-11`, `**Status:** SIGNED OFF`): it is a
	// document ABOUT a record, and there are dozens of them. A document DECLARING
	// itself accepted or superseded under a handle another record answers to is a
	// second copy of that record. Only the second is this rule's finding.
	decisionStatusValueRe = regexp.MustCompile(`(?i)\b(accepted|proposed|rejected|superseded|deprecated)\b`)
)

// maxCrossStoreBytes caps a candidate read. The rule reads ordinary markdown, so
// the cap is the citation walk's own page limit rather than a second number.
const maxCrossStoreBytes = citationPageSizeLimit

// checkCrossStoreIDClaim walks the markdown OUTSIDE the configured record stores
// and reports every decision-shaped document that claims an id a real record
// already holds. A repo with no stores configured has no baseline to weigh a
// claim against, so it contributes nothing and is not an error.
func checkCrossStoreIDClaim(repoRoot string, cfg Config, rc RuleConfig) ([]Finding, error) {
	stores := rc.RecordStores
	if len(stores) == 0 {
		stores = cfg.Rules[ruleRecordSchema].RecordStores
	}
	if len(stores) == 0 {
		return nil, nil
	}

	// The taken-id set comes from the record graph — the same scan record_schema
	// trusts for its own high-water marks — so this detector cannot drift from the
	// stores' own view of which ids are taken.
	graphCfg := cfg
	graphCfg.Rules = map[string]RuleConfig{ruleRecordSchema: {RecordStores: stores}}
	graph, err := LoadRecordGraph(graphCfg, repoRoot)
	if err != nil {
		return nil, err
	}
	taken := make(map[string]bool, len(graph.Nodes))
	for _, n := range graph.Nodes {
		taken[canonRecordID(strings.ToLower(n.ID))] = true
	}
	if len(taken) == 0 {
		return nil, nil
	}

	storeDirs := make([]string, 0, len(stores))
	for _, dir := range stores {
		if dir == "" {
			continue
		}
		if err := containedRepoPath(dir); err != nil {
			return nil, &configError{ruleCrossStoreIDClaim + ": record store " + quote(dir) + " " + err.Error() +
				"; the lint reads only inside the repository"}
		}
		storeDirs = append(storeDirs, filepath.ToSlash(strings.TrimSuffix(filepath.ToSlash(dir), "/"))+"/")
	}

	var out []Finding
	err = filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			// A tree that vanishes MID-walk (a checkout racing the gate) is an
			// error; the root's own absence is not reachable here, since the
			// caller resolved it.
			return walkErr
		}
		rel := filepath.ToSlash(repoRel(repoRoot, path))
		if d.IsDir() {
			// .git is machinery, never content; a record store's own files are
			// judged by record_schema, and this rule is about what sits OUTSIDE
			// them. A symlinked directory is not followed by WalkDir at all, so a
			// committed link cannot redirect the walk out of the repository.
			if d.Name() == ".git" || hasAnyPrefix(rel+"/", storeDirs) {
				return filepath.SkipDir
			}
			return nil
		}
		if !hasMarkdownExt(d.Name()) || hasAnyPrefix(rel, storeDirs) {
			return nil
		}
		// The walk reads committed files, so the leaf is contained and guarded the
		// way the roots walk contains and guards its own (a `docs/leak.md ->
		// /etc/passwd` link must not be read, and a `-> /dev/zero` one must not be
		// read unbounded).
		realPath, err := containedRealPath(repoRoot, path)
		if err != nil {
			return nil // a leaf resolving outside the repo is not this rule's finding
		}
		content, err := fsutil.ReadGuarded(realPath, maxCrossStoreBytes)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(content), "\n")
		f, ok := crossStoreClaim(rel, lines, taken, rc.Severity)
		if ok {
			out = append(out, f)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortFindings(out)
	return out, nil
}

// crossStoreClaim judges one candidate document: does its H1 claim a taken id,
// and does its body carry the decision shape.
func crossStoreClaim(rel string, lines []string, taken map[string]bool, severity string) (Finding, bool) {
	heading, line := recordH1(lines)
	m := claimedHandleRe.FindStringSubmatch(strings.TrimSpace(heading))
	if m == nil {
		return Finding{}, false
	}
	num, err := strconv.Atoi(m[2])
	if err != nil {
		return Finding{}, false
	}
	id := canonRecordID(strings.ToLower(m[1]) + "-" + strconv.Itoa(num))
	if !taken[id] {
		return Finding{}, false
	}
	if !hasDecisionShape(lines) {
		return Finding{}, false
	}
	return Finding{
		File: rel, Line: line, RuleID: ruleCrossStoreIDClaim, Severity: severity,
		Message: "document claims the record id '" + id + "' but does not sit in that record's store; " +
			id + " is already held by a real record, so this file is a second document answering to one handle — " +
			"every cross-reference and index that keys on it resolves to the other. Give the document its own " +
			"identity (a title that claims no handle), or file it as a record in the store and let the id be minted",
	}, true
}

// hasDecisionShape reports whether the body DECLARES a record lifecycle status —
// a `## Status` section or a `Status:` line whose value is one of the record's
// own lifecycle states. Fenced blocks are masked by the shared fenceMask: a page
// QUOTING the shape (which is how this rule gets documented) is an example, not a
// decision.
func hasDecisionShape(lines []string) bool {
	mask := fenceMask(lines)
	for i, line := range lines {
		if mask[i] {
			continue
		}
		if m := decisionStatusKeyRe.FindStringSubmatch(line); m != nil {
			if decisionStatusValueRe.MatchString(m[1]) {
				return true
			}
			continue
		}
		// A `## Status` heading carries its value in the section beneath it, so
		// the first non-blank line after the heading is the value.
		if decisionStatusHeadingRe.MatchString(line) {
			for j := i + 1; j < len(lines); j++ {
				if mask[j] || strings.TrimSpace(lines[j]) == "" {
					continue
				}
				if decisionStatusValueRe.MatchString(lines[j]) {
					return true
				}
				break
			}
		}
	}
	return false
}
