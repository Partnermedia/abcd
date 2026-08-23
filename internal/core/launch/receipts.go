package launch

// The semantic-receipts row of the release preview (iss-2608231226342272).
//
// `release.yml` arms `receipt_gate` fail-closed against the release content
// commit, and refuses unless every required gate has a PROMOTE receipt naming
// it. The preview used to omit that gate entirely, so it answered cleanly —
// bundle sized, scan clean, smoke ok — while the one gate that would actually
// refuse went unmentioned. An absent row reads as "no such gate", which is the
// one thing it is not: the gate is fully implemented and armed at release time.
// That is worse than the two `not_implemented` rows in the same list, which at
// least announce that they did not run.
//
// This row reports what is RECORDED and nothing more. It deliberately does not
// assert which gates are required, because `release.yml` owns that list on
// purpose: the workflow, not the committer-editable in-tree config, is the trust
// root for the decision to gate. A preview that published its own required-gates
// list would be a second, drifting copy of a security decision — so this reports
// presence and points at the runbook, and never reports a pass.
//
// The measurement is NOT taken here, matching the citation gate above it: it
// needs a git rev-parse and a directory read, and the front door that already
// holds both hands it in as data.

import "strconv"

// ReceiptPreflight is the state of the semantic-pass receipts, measured by the
// caller. A nil pointer means the measurement was not taken, and the gate says
// so rather than reporting an absence it never looked for.
type ReceiptPreflight struct {
	// Unreadable, when non-empty, means the measurement could not be taken —
	// the candidate commit would not resolve, or the receipts directory could
	// not be read. A broken measurement is a distinct state from "none found",
	// because reporting the first as the second is a false statement about a
	// real requirement.
	Unreadable string
	// Commit is the candidate commit the receipts would have to name. At preview
	// time this is the working tree's HEAD; the commit release.yml actually arms
	// against is derived from the merge (`<merge>^2^`), so this is indicative,
	// not authoritative — the detail text says so.
	Commit string
	// Recorded names the detectors that have a receipt for Commit, in the order
	// found. It counts receipts present, not receipts valid: validity (PROMOTE,
	// judge model, detector binding, manifest hash) is receipt_gate's to judge,
	// and duplicating that verdict here would be the second trust root this row
	// exists to avoid.
	Recorded []string
}

// receiptGate renders the semantic-receipts row. Its status is always
// "host-run": the passes spawn LLM agents, so neither CI nor this preview can
// run them, and this row never reports a pass — only what is on disk.
func receiptGate(p *ReceiptPreflight) GateSummary {
	const name = "semantic-receipts"
	runbook := "see .abcd/development/release-gate/README.md"
	if p == nil {
		return GateSummary{Name: name, Status: "host-run",
			Detail: "not measured; release.yml requires PROMOTE receipts naming the release-content commit (" + runbook + ")"}
	}
	if p.Unreadable != "" {
		return GateSummary{Name: name, Status: "host-run",
			Detail: "cannot read the recorded receipts: " + p.Unreadable + " (" + runbook + ")"}
	}
	at := "HEAD"
	if len(p.Commit) >= 7 {
		at = p.Commit[:7]
	}
	if len(p.Recorded) == 0 {
		return GateSummary{Name: name, Status: "host-run",
			Detail: "no receipts recorded for " + at + "; the release branch is two commits — the CHANGELOG roll, then the receipts naming it (" + runbook + ")"}
	}
	detail := strconv.Itoa(len(p.Recorded)) + " receipt(s) recorded for " + at + ": "
	for i, d := range p.Recorded {
		if i > 0 {
			detail += ", "
		}
		detail += d
	}
	// Never "all required gates satisfied": this row does not know the required
	// list, and release.yml judges validity. Presence is all it may claim.
	return GateSummary{Name: name, Status: "host-run",
		Detail: detail + "; release.yml judges their validity against the content commit (" + runbook + ")"}
}
