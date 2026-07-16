package lifeboat

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/capture"
	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/core/intent"
)

// graveyard_abandoned.go — Layer 2 of the graveyard: what the project itself
// declared dead in its Tier-1/2 record. It is a pure function over the read-only
// SourceContext file surface (ctx.ReadFile/ctx.ListDir, both contained and
// capped) and frontmatter.Fields — it never touches git and never writes.
//
// Every qualifying record contributes a Finding keyed by the record's OWN id
// (an intent's itd-N, an ADR's adr-N, an issue's iss-N) so a layer-3 lesson can
// cite exactly the id a human reads in the record — no re-derivation. Findings
// are grouped in signalRank order (buildAbandoned appends signal by signal in
// that order) and, within a signal, in that signal's fixed deterministic order
// (sorted record id, sorted ADR id, or DECISIONS.md line order), so the
// assembled abandoned.json is byte-identical across re-plans of an unchanged
// repo and the pinned manifest hash stays stable.

// Record-id shapes. A finding is keyed by the record's own id only when that id
// has the expected shape; a record whose id is malformed (or whose id must be
// derived) is validated against these before it can key a finding.
var (
	gvIntentIDRe = regexp.MustCompile(`^itd-[0-9]+$`)
	gvADRIDRe    = regexp.MustCompile(`^adr-[0-9]+$`)
	gvIssueIDRe  = regexp.MustCompile(`^iss-[0-9]+$`)
)

// gvRejectionVerbs is the deliberately narrow set of verbs that mark a
// DECISIONS.md bullet as recording a rejected option. It is conservative on
// purpose: a broad list ("no", "instead", "not") would fire on ordinary prose,
// so only unambiguous abandonment verbs qualify. Matched as a substring of the
// lower-cased line, so "rejected" fires on "RAG rejected at this scale" but not
// on the unrelated word "rejection".
var gvRejectionVerbs = []string{
	"rejected", "dropped", "discarded", "abandoned", "ruled out", "deferred",
}

// maxAbandonedEvidencePerFinding bounds the evidence lines one layer-2 finding
// may carry. Only the alternatives-considered signal can produce an unbounded
// count (one line per bullet); the cap keeps a hostile or pathological ADR from
// ballooning a single finding.
const maxAbandonedEvidencePerFinding = 32

// buildAbandoned reads what the project explicitly declared dead and returns the
// deterministic, evidence-only layer-2 record. Findings are grouped in
// signalRank order; the slice is never nil, so abandoned.json always marshals
// "findings": [] rather than null when the record declares nothing dead.
func buildAbandoned(ctx *SourceContext) Abandoned {
	var fs []Finding
	fs = append(fs, gvSupersededIntents(ctx)...)
	fs = append(fs, gvSupersededADRs(ctx)...)
	fs = append(fs, gvAlternativesConsidered(ctx)...)
	fs = append(fs, gvWontfixIssues(ctx)...)
	fs = append(fs, gvRejectedOptions(ctx)...)
	if fs == nil {
		fs = []Finding{}
	}
	return Abandoned{SchemaVersion: GraveyardSchemaVersion, Findings: fs}
}

// gvSupersededIntents reports every intent in the superseded/ bucket, keyed by
// its own itd-N id, sorted numerically by id.
func gvSupersededIntents(ctx *SourceContext) []Finding {
	dir := intent.IntentsRelDir + "/" + intent.BucketSuperseded
	var out []Finding
	seen := map[string]bool{}
	for _, name := range ctx.ListDir(dir) {
		if !strings.HasPrefix(name, "itd-") || !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		path := dir + "/" + name
		fields, ok := gvFields(ctx, path)
		if !ok {
			continue
		}
		id := gvUnquote(fields["id"].Value)
		if !gvIntentIDRe.MatchString(id) || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, Finding{
			ID:       id,
			Signal:   SignalSupersededIntent,
			Summary:  "intent superseded",
			Evidence: []string{sanitize(path)},
		})
	}
	gvSortByID(out)
	return gvCapFindings(out)
}

// gvSupersededADRs reports every ADR the record marks superseded — by an
// explicit status: superseded or a non-null superseded_by — across the native
// ADR home and each conventional home. Keyed by the ADR's own adr-N id (derived
// from the NNNN- filename when the frontmatter carries none), deduped first-wins
// across homes (native wins), sorted numerically by id.
func gvSupersededADRs(ctx *SourceContext) []Finding {
	var out []Finding
	seen := map[string]bool{}
	gvEachADR(ctx, func(name, path string, fields map[string]frontmatter.Field) {
		status := strings.ToLower(gvUnquote(fields["status"].Value))
		supBy := gvUnquote(fields["superseded_by"].Value)
		if status != "superseded" && frontmatter.IsNull(supBy) {
			return
		}
		id := gvADRID(fields, name)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		var ev []string
		if !frontmatter.IsNull(supBy) {
			ev = append(ev, sanitize("superseded_by: "+supBy))
		}
		ev = append(ev, sanitize(path))
		out = append(out, Finding{
			ID:       id,
			Signal:   SignalSupersededADR,
			Summary:  "ADR superseded",
			Evidence: ev,
		})
	})
	gvSortByID(out)
	return gvCapFindings(out)
}

// gvAlternativesConsidered reports every ADR carrying an Alternatives-Considered
// (or Options-Considered) section, keyed by <adr-id>-alt, with each top-level
// bullet of the section as evidence. Deduped first-wins across homes, sorted by
// the underlying ADR id.
func gvAlternativesConsidered(ctx *SourceContext) []Finding {
	var out []Finding
	seen := map[string]bool{}
	gvEachADR(ctx, func(name, path string, fields map[string]frontmatter.Field) {
		id := gvADRID(fields, name)
		if id == "" || seen[id] {
			return
		}
		data, ok := ctx.ReadFile(path)
		if !ok {
			return
		}
		bullets, found := gvSectionBullets(data, "alternatives considered", "alternatives", "options considered")
		if !found {
			return
		}
		seen[id] = true
		if len(bullets) > maxAbandonedEvidencePerFinding {
			bullets = bullets[:maxAbandonedEvidencePerFinding]
		}
		ev := make([]string, 0, len(bullets))
		for _, b := range bullets {
			ev = append(ev, sanitize(b))
		}
		f := Finding{
			ID:      adrAltID(id),
			Signal:  SignalAlternativesConsidered,
			Summary: strings.ToUpper(id) + " weighed and rejected alternatives",
		}
		if len(ev) > 0 {
			f.Evidence = ev
		}
		out = append(out, f)
	})
	// Sort by the ADR id embedded in the <adr-id>-alt finding id.
	gvSortByID(out)
	return gvCapFindings(out)
}

// gvWontfixIssues reports every issue in the wontfix/ ledger bucket, keyed by its
// own iss-N id, evidence being the wontfix_reason (or the slug when absent),
// sorted numerically by id.
func gvWontfixIssues(ctx *SourceContext) []Finding {
	dir := capture.LedgerRelPath + "/wontfix"
	var out []Finding
	seen := map[string]bool{}
	for _, name := range ctx.ListDir(dir) {
		if !strings.HasPrefix(name, "iss-") || !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		path := dir + "/" + name
		fields, ok := gvFields(ctx, path)
		if !ok {
			continue
		}
		id := gvUnquote(fields["id"].Value)
		if !gvIssueIDRe.MatchString(id) || seen[id] {
			continue
		}
		seen[id] = true
		var ev []string
		if reason := gvUnquote(fields["wontfix_reason"].Value); reason != "" {
			ev = append(ev, sanitize("wontfix_reason: "+reason))
		} else if slug := gvUnquote(fields["slug"].Value); slug != "" {
			ev = append(ev, sanitize("slug: "+slug))
		}
		f := Finding{ID: id, Signal: SignalWontfixIssue, Summary: "issue closed wontfix"}
		if len(ev) > 0 {
			f.Evidence = ev
		}
		out = append(out, f)
	}
	gvSortByID(out)
	return gvCapFindings(out)
}

// gvRejectedOptions reports every top-level DECISIONS.md bullet whose text
// carries a conservative rejection verb, keyed by dec-L<line> (1-based line
// number, stable for an unchanged append-only file), in file (line) order.
func gvRejectedOptions(ctx *SourceContext) []Finding {
	data, ok := ctx.ReadFile(nativeDecisions)
	if !ok {
		return nil
	}
	var out []Finding
	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, "\r")
		if !strings.HasPrefix(line, "- ") { // top-level bullet only (no indentation)
			continue
		}
		low := strings.ToLower(line)
		matched := false
		for _, v := range gvRejectionVerbs {
			if strings.Contains(low, v) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		out = append(out, Finding{
			ID:       decisionID(i + 1),
			Signal:   SignalRejectedOption,
			Summary:  "decision log records a rejected option",
			Evidence: []string{sanitize(strings.TrimSpace(line))},
		})
	}
	return gvCapFindings(out)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// gvADRHomes lists the ADR directories in dedup priority order: the native home
// first (so it wins a first-writer-wins tie), then the conventional homes.
func gvADRHomes() []string {
	return append([]string{nativeADRDir}, convADRDirs...)
}

// gvEachADR calls fn for every ADR document (*.md/*.markdown) under every ADR
// home, in home order then sorted-name order, having parsed its frontmatter. A
// file that cannot be read is skipped.
func gvEachADR(ctx *SourceContext, fn func(name, path string, fields map[string]frontmatter.Field)) {
	for _, dir := range gvADRHomes() {
		for _, name := range ctx.ListDir(dir) {
			low := strings.ToLower(name)
			if !strings.HasSuffix(low, ".md") && !strings.HasSuffix(low, ".markdown") {
				continue
			}
			path := dir + "/" + name
			fields, ok := gvFields(ctx, path)
			if !ok {
				continue
			}
			fn(name, path, fields)
		}
	}
}

// gvADRID resolves an ADR's id: its frontmatter id when it is a valid adr-N,
// else the id derived from a leading NNNN- filename, else "" (skip).
func gvADRID(fields map[string]frontmatter.Field, name string) string {
	if id := gvUnquote(fields["id"].Value); gvADRIDRe.MatchString(id) {
		return id
	}
	if id := gvADRIDFromFilename(name); gvADRIDRe.MatchString(id) {
		return id
	}
	return ""
}

// gvADRIDFromFilename derives adr-N from a leading run of digits in an NNNN-slug
// ADR filename (leading zeros stripped), or "" when the name is not numbered.
func gvADRIDFromFilename(name string) string {
	i := 0
	for i < len(name) && name[i] >= '0' && name[i] <= '9' {
		i++
	}
	if i == 0 {
		return ""
	}
	n, err := strconv.Atoi(name[:i])
	if err != nil {
		return ""
	}
	return fmt.Sprintf("adr-%d", n)
}

// gvFields reads a record and parses its leading frontmatter block. ok is false
// when the file is absent/oversize/non-regular (the ReadFile guards).
func gvFields(ctx *SourceContext, path string) (map[string]frontmatter.Field, bool) {
	data, ok := ctx.ReadFile(path)
	if !ok {
		return nil, false
	}
	return frontmatter.Fields(strings.Split(string(data), "\n")), true
}

// gvSectionBullets finds the first Markdown section whose heading contains any
// keyword (case-insensitive) and returns that section's top-level bullet lines
// (each trimmed) up to the next heading. found reports whether such a section
// exists at all — a section present but bullet-free still counts as found.
func gvSectionBullets(data []byte, keywords ...string) (bullets []string, found bool) {
	inSection := false
	for _, raw := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "#") { // a heading
			if inSection {
				break // the next heading closes the section
			}
			low := strings.ToLower(trimmed)
			for _, k := range keywords {
				if strings.Contains(low, k) {
					inSection = true
					found = true
					break
				}
			}
			continue
		}
		if !inSection {
			continue
		}
		// A top-level bullet has no leading indentation.
		if len(raw) > 0 && raw[0] != ' ' && raw[0] != '\t' &&
			(strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")) {
			bullets = append(bullets, trimmed)
		}
	}
	return bullets, found
}

// gvUnquote strips one layer of surrounding matching quotes (frontmatter values
// are sometimes quoted — issues quote their id/slug — and sometimes bare —
// intents and ADRs). It trims surrounding whitespace first.
func gvUnquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// gvSortByID sorts findings numerically by the first integer in their id, then
// lexically — so itd-2 precedes itd-10 within a signal. Every finding in one
// slice shares a signal (and therefore an id prefix), so the numeric key orders
// them the way a human reads a record.
func gvSortByID(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		ki, kj := gvIDNum(fs[i].ID), gvIDNum(fs[j].ID)
		if ki != kj {
			return ki < kj
		}
		return fs[i].ID < fs[j].ID
	})
}

// gvIDNum returns the value of the first run of digits in id, or -1 if none.
func gvIDNum(id string) int {
	start := -1
	for i := 0; i < len(id); i++ {
		if id[i] >= '0' && id[i] <= '9' {
			start = i
			break
		}
	}
	if start < 0 {
		return -1
	}
	end := start
	for end < len(id) && id[end] >= '0' && id[end] <= '9' {
		end++
	}
	n, _ := strconv.Atoi(id[start:end])
	return n
}

// gvCapFindings bounds a single signal's findings so a hostile or pathological
// record cannot balloon abandoned.json.
func gvCapFindings(fs []Finding) []Finding {
	if len(fs) > maxGraveyardFindingsPerSignal {
		return fs[:maxGraveyardFindingsPerSignal]
	}
	return fs
}
