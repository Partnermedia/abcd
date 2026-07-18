package intent

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/core/spec"
)

// Check names, in the fixed order Ready always reports them.
const (
	CheckBucket             = "bucket"
	CheckAcceptanceCriteria = "acceptance_criteria"
	CheckSpecLink           = "spec_link"
	CheckSpecBody           = "spec_body"
)

// ReadyCheck is one finding of the implement-readiness gate.
type ReadyCheck struct {
	Name   string `json:"name"` // bucket | acceptance_criteria | spec_link | spec_body
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`           // why it passed or failed
	Remedy string `json:"remedy,omitempty"` // the exact next command/action when !OK
}

// ReadyResult is the structured readiness verdict for one intent: may this
// intent be implemented now? Every check is always evaluated and reported, so a
// surface presents the full picture rather than the first failure.
type ReadyResult struct {
	IntentID string       `json:"intent_id"`
	Path     string       `json:"path"`   // repo-relative intent path
	Bucket   string       `json:"bucket"` // directory-as-truth state
	SpecID   string       `json:"spec_id"`
	Ready    bool         `json:"ready"`
	Checks   []ReadyCheck `json:"checks"` // always exactly 4, fixed order
}

// Ready reports whether an intent is ready to implement: planned
// (directory-as-truth), carrying enumerable Acceptance Criteria, linked
// bidirectionally to a spec, and that spec's body written past its minted stub.
// It is a read-only reporter — the machine-checkable form of the run protocol's
// "is this item ready?" question — and never mutates the store.
//
// "Not ready" is a result, never an error: error is reserved for structural
// faults (malformed id, unknown intent, unreadable record, store load failure),
// so a surface can map result vs error to distinct exit codes.
func Ready(repoRoot, intentID string) (ReadyResult, error) {
	if !intentIDRe.MatchString(intentID) {
		return ReadyResult{}, fmt.Errorf("intent: id %q must match ^itd-[0-9]+$", intentID)
	}
	corpus, err := Load(repoRoot)
	if err != nil {
		return ReadyResult{}, err
	}
	it, ok := corpus.Lookup(intentID)
	if !ok {
		return ReadyResult{}, fmt.Errorf("intent: %s not found in any bucket", intentID)
	}
	store, err := spec.Load(repoRoot)
	if err != nil {
		return ReadyResult{}, err
	}
	data, err := readRepoFile(filepath.Join(repoRoot, it.Path), it.Path)
	if err != nil {
		return ReadyResult{}, err
	}
	content := string(data)
	acCount := countAcceptanceCriteria(content)

	res := ReadyResult{
		IntentID: it.ID,
		Path:     it.Path,
		Bucket:   it.Bucket,
		SpecID:   it.SpecID,
	}
	res.Checks = append(res.Checks, bucketCheck(it, acCount, content))
	res.Checks = append(res.Checks, acCheck(acCount))
	linkOK, linked := specLinkCheck(it, store)
	res.Checks = append(res.Checks, linkOK)
	bodyCheck, err := specBodyCheck(repoRoot, it, linked, linkOK.OK)
	if err != nil {
		return ReadyResult{}, err
	}
	res.Checks = append(res.Checks, bodyCheck)

	res.Ready = true
	for _, c := range res.Checks {
		if !c.OK {
			res.Ready = false
			break
		}
	}
	return res, nil
}

// bucketCheck reports the lifecycle-state gate: only a planned intent may be
// implemented. For a draft the remedy names the exact route to planned —
// through the maintainer's sign-off, via the planning interview when the
// Acceptance Criteria are not yet enumerable. Terminal buckets carry no remedy:
// there is nothing to fix, the answer is simply no.
func bucketCheck(it Intent, acCount int, content string) ReadyCheck {
	c := ReadyCheck{Name: CheckBucket}
	switch it.Bucket {
	case BucketPlanned:
		c.OK = true
		c.Detail = fmt.Sprintf("%s is planned", it.ID)
	case BucketDrafts:
		c.Detail = fmt.Sprintf("%s is a draft — an intent that is not planned cannot be implemented", it.ID)
		if acCount > 0 {
			c.Remedy = fmt.Sprintf("confirm the Acceptance Criteria with the maintainer, then run `abcd intent plan %s`", it.ID)
		} else {
			c.Remedy = fmt.Sprintf("run the planning interview (/abcd:intent) to write and confirm Acceptance Criteria, then run `abcd intent plan %s`", it.ID)
		}
	case BucketShipped:
		c.Detail = fmt.Sprintf("%s is already shipped — nothing left to implement", it.ID)
	case BucketDisciplines:
		c.Detail = fmt.Sprintf("%s is a discipline — it imposes gates on other work and is never implemented via a spec of its own", it.ID)
	case BucketSuperseded:
		c.Detail = fmt.Sprintf("%s is superseded — implement the successor instead", it.ID)
		if by := frontmatter.Fields(strings.Split(content, "\n"))["superseded_by"].Value; !frontmatter.IsNull(by) {
			c.Detail += " (superseded_by: " + by + ")"
		}
	default:
		c.Detail = fmt.Sprintf("%s is in unknown bucket %q", it.ID, it.Bucket)
	}
	return c
}

// acCheck reports the itd-1 discipline through the same parser Plan and the
// fidelity review use (countAcceptanceCriteria), so the three gates can never
// disagree about what counts as a criterion.
func acCheck(acCount int) ReadyCheck {
	c := ReadyCheck{Name: CheckAcceptanceCriteria}
	if acCount > 0 {
		c.OK = true
		c.Detail = fmt.Sprintf("%d top-level bullet(s) in '## Acceptance Criteria'", acCount)
	} else {
		c.Detail = "no top-level bullets in '## Acceptance Criteria' (itd-1 discipline)"
		c.Remedy = "add at least one Given-When-Then bullet — the planning interview walks this with the maintainer"
	}
	return c
}

// specLinkCheck reports the bidirectional intent↔spec link (the same agreement
// Reconcile enforces, here as a report instead of a refusal). It returns the
// linked spec when, and only when, the link holds, so specBodyCheck never reads
// a file the link did not vouch for.
func specLinkCheck(it Intent, store spec.Store) (ReadyCheck, spec.Spec) {
	c := ReadyCheck{Name: CheckSpecLink}
	if frontmatter.IsNull(it.SpecID) {
		if claimer, ok := store.ByIntent(it.ID); ok {
			c.Detail = fmt.Sprintf("spec_id is null but %s claims %s (one-sided link)", claimer.ID, it.ID)
			c.Remedy = fmt.Sprintf("run `abcd intent link %s %s`", it.ID, claimer.ID)
			return c, spec.Spec{}
		}
		c.Detail = "spec_id is null — no spec realises this intent"
		if it.Bucket == BucketDrafts {
			c.Remedy = fmt.Sprintf("planning (`abcd intent plan %s`) mints and links the spec", it.ID)
		} else {
			c.Remedy = fmt.Sprintf("hand-author %s/open/spc-N-<slug>.md with `intent: %s`, then run `abcd intent link`", spec.SpecsRelDir, it.ID)
		}
		return c, spec.Spec{}
	}
	sp, ok := store.Lookup(it.SpecID)
	if !ok {
		c.Detail = fmt.Sprintf("spec_id is %s but no such spec exists in the store", it.SpecID)
		c.Remedy = fmt.Sprintf("restore %s or correct spec_id via `abcd intent link`", it.SpecID)
		return c, spec.Spec{}
	}
	if sp.Intent != it.ID {
		c.Detail = fmt.Sprintf("bidirectional link disagrees: %s names %s, but %s claims %s", it.ID, it.SpecID, sp.ID, sp.Intent)
		c.Remedy = "correct the spec's `intent:` field or the intent's spec_id so both sides agree"
		return c, spec.Spec{}
	}
	c.OK = true
	c.Detail = fmt.Sprintf("linked to %s (bidirectional)", sp.ID)
	if sp.Status == spec.StatusClosed && it.Bucket == BucketPlanned {
		c.Detail += "; note: the spec is closed while the intent is still planned (drift)"
	}
	return c, sp
}

// specBodyCheck reports whether the linked spec's body has been written past
// the minted stub — the spec is the design record implementation builds
// against, so an untouched placeholder means there is nothing to build from. A
// read failure on the linked spec is a structural fault, not a finding.
func specBodyCheck(repoRoot string, it Intent, sp spec.Spec, linkOK bool) (ReadyCheck, error) {
	c := ReadyCheck{Name: CheckSpecBody}
	if !linkOK {
		c.Detail = "unchecked — no linked spec"
		return c, nil
	}
	data, err := readRepoFile(filepath.Join(repoRoot, sp.Path), sp.Path)
	if err != nil {
		return ReadyCheck{}, err
	}
	if spec.BodyIsStub(string(data)) {
		c.Detail = fmt.Sprintf("%s still carries the minted _Draft: placeholder — the design record is unwritten", sp.ID)
		c.Remedy = fmt.Sprintf("write the spec body at %s, then re-run `abcd intent ready %s`", sp.Path, it.ID)
		return c, nil
	}
	c.OK = true
	c.Detail = fmt.Sprintf("spec body at %s is written", sp.Path)
	return c, nil
}
