package capture

// The reading-record and disposition write paths (itd-180, spc-58).
//
// A reading record is what an instrument returned under a recorded visible
// world; a disposition is the researcher's answer to one such item, written
// SEPARATELY and keyed to it. The two are never one write, so the ledger can
// always show that a finding existed before it was answered.
//
// Every refusal below refuses rather than accepting-and-flagging, and names the
// rule it enforces. A record the writer would refuse must never reach the
// committed tree, because the reader refuses it too — and a record no reader can
// read is not a lax record, it is a lost one.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/fsutil"
)

// The three families' id grammars. Each is checked BEFORE the value is used to
// build a path, so a traversal id can never touch the filesystem.
var (
	reRunID         = regexp.MustCompile(`^` + issueschema.ReadingRunFamily + `-[0-9]+$`)
	reReadingItemID = regexp.MustCompile(`^` + issueschema.ReadingItemFamily + `-[0-9]+$`)
	reDispositionID = regexp.MustCompile(`^` + issueschema.DispositionFamily + `-[0-9]+$`)
)

// ReadingItem is one thing the instrument returned: the pattern it named (an
// envelope field, because a universal core condition must not live in a variant
// part) plus the position-typed body.
type ReadingItem struct {
	// Pattern is the pattern named — required at every position and in every
	// supply regime.
	Pattern string
	// Body carries exactly the fields the run's position declares
	// (issueschema.ReadingBodyFields). A field from another position's body is
	// refused: one record type, four bodies, and an item belongs to one of them.
	Body map[string]string
	// OccasionedBy is RESERVED and dormant — the join key of the surprise entry,
	// populated in Iteration 2. A populated value is refused until the shape is
	// ruled, so the reservation is a behaviour rather than a comment.
	OccasionedBy string
}

// IngestReadingRequest writes one run's items into the ledger. Position and
// Regime come from the reading's DEFINITION, never from operator input, and are
// therefore run-level rather than per-item.
type IngestReadingRequest struct {
	RepoRoot   string
	IssuesRoot string
	// Run is the run identifier (rdg-N) minted by the assembler.
	Run string
	// Manifest is the content-hash reference to the run's committed manifest,
	// which is what makes the assembly re-runnable and diffable.
	Manifest string
	Position string
	Regime   string
	Items    []ReadingItem
}

// ReadingRecordRef is one written reading record: its minted id and its
// repo-relative path.
type ReadingRecordRef struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// IngestReadingResult is the outcome of a successful ingest.
type IngestReadingResult struct {
	Run     string             `json:"run"`
	Records []ReadingRecordRef `json:"records"`
	// Redacted counts the spans the ledger redactor rewrote across every item
	// written, and Degraded is non-empty when it ran with a weakened pattern set.
	// Both exist so a surface can SAY the text was altered.
	Redacted int    `json:"redacted,omitempty"`
	Degraded string `json:"redaction_degraded,omitempty"`
}

// DispositionRequest writes one disposition, keyed to one reading item.
type DispositionRequest struct {
	RepoRoot   string
	IssuesRoot string
	// Item is the reading item this answers (rdi-N).
	Item string
	// State is one of issueschema.DispositionStates, subject to the per-position
	// availability rule.
	State string
	// Grounds is disposition_grounds: free text, required on every state except
	// held.
	Grounds string
	// ExitCondition is required on held and refused elsewhere.
	ExitCondition string
	// Supersedes names the standing disposition this one replaces (dsp-N) — the
	// only exit from a hold.
	Supersedes string
	// Recurs cites prior item ids: the recorded form of the researcher's warm
	// recognition of a persistence. Never a mechanical join, and never a state.
	Recurs []string
	// HoldFrameLocation / HoldMoscow are the RESERVED two-axis hold field. Both
	// are refused while populated; the grammars are stated and dormant.
	HoldFrameLocation string
	HoldMoscow        string
}

// DispositionResult is the outcome of a successful Disposition.
type DispositionResult struct {
	ID       string `json:"id"`
	Item     string `json:"item"`
	State    string `json:"state"`
	Position string `json:"position"`
	Path     string `json:"path"`
	Redacted int    `json:"redacted,omitempty"`
	Degraded string `json:"redaction_degraded,omitempty"`
}

// IngestReading writes one reading record per item under
// readings/<run-id>/rdi-<N>.md. Every id is minted (adr-45) and never derived
// from the item's content, so two runs returning the same tension carry
// different ids for free — a re-raise stays distinguishable from its first
// appearance, and the recurrence signal survives.
//
// It is the ledger's record WRITER, not a payload validator: the output contract
// that "validated" refers to belongs to the cold-reading output contract, which
// calls this and adds no second validation path. What is enforced here is the
// RECORD schema — the envelope, the position's own body, and the reserved
// fields — which is this package's to refuse.
func IngestReading(req IngestReadingRequest) (IngestReadingResult, error) {
	repoRoot, issuesRoot, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return IngestReadingResult{}, err
	}
	if !reRunID.MatchString(req.Run) {
		return IngestReadingResult{}, fmt.Errorf("%w: run %q does not match ^%s-[0-9]+$",
			ErrMalformedFrontmatter, req.Run, issueschema.ReadingRunFamily)
	}
	if len(req.Items) == 0 {
		return IngestReadingResult{}, fmt.Errorf("%w: a run with no items is recorded as a run with an empty item set, not as an ingest",
			ErrMissingRequiredField)
	}
	if err := mutationPreamble(issuesRoot); err != nil {
		return IngestReadingResult{}, err
	}

	// Assemble and validate EVERY item before anything is written: a run that is
	// half-written is a visible world nobody can reconstruct.
	type staged struct {
		id      string
		content string
	}
	var pending []staged
	result := IngestReadingResult{Run: req.Run}
	for i, item := range req.Items {
		id, err := minter.Mint(issueschema.ReadingItemFamily)
		if err != nil {
			return IngestReadingResult{}, err
		}
		fields, fm, redacted, degraded, err := readingFields(repoRoot, id, req, item)
		if err != nil {
			return IngestReadingResult{}, fmt.Errorf("item %d: %w", i+1, err)
		}
		if err := validateReadingStrict(fm); err != nil {
			return IngestReadingResult{}, fmt.Errorf("item %d: %w", i+1, err)
		}
		content, err := buildIssueText(fields, "")
		if err != nil {
			return IngestReadingResult{}, fmt.Errorf("item %d: %w", i+1, err)
		}
		pending = append(pending, staged{id: id, content: content})
		result.Redacted += redacted
		if degraded != "" {
			result.Degraded = degraded
		}
	}

	runDir := filepath.Join(issuesRoot, issueschema.ReadingsDir, req.Run)
	err = withLedgerLock(issuesRoot, func() error {
		if err := ensureFamilyDir(issuesRoot, issueschema.ReadingsDir, req.Run); err != nil {
			return err
		}
		for _, p := range pending {
			path := filepath.Join(runDir, p.id+".md")
			if err := refuseExistingRecord(path, p.id); err != nil {
				return err
			}
			if err := fsutil.WriteFileAtomic(path, []byte(p.content), 0o644); err != nil {
				return err
			}
			result.Records = append(result.Records, ReadingRecordRef{
				ID: p.id, Path: fsutil.RepoRel(repoRoot, path),
			})
		}
		return nil
	})
	if err != nil {
		return IngestReadingResult{}, err
	}
	return result, nil
}

// Disposition writes the researcher's answer to one reading item, into a
// directory keyed by the ITEM and a file keyed by the disposition's own id.
//
// The two keys settle requirements that pull against each other. Status is the
// presence of the keyed directory — one probe, never a folder-membership
// question. And a disposition still has an id of its own, so the only exit from
// a held state (a superseding disposition that CITES the one it replaces) has
// something to cite. The superseded record stays in place: a hold that vanished
// when it was answered would take its own exit condition with it (adr-3 — the
// location is the state, and git is the history).
func Disposition(req DispositionRequest) (DispositionResult, error) {
	repoRoot, issuesRoot, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return DispositionResult{}, err
	}
	if !reReadingItemID.MatchString(req.Item) {
		return DispositionResult{}, fmt.Errorf("%w: item %q does not match ^%s-[0-9]+$",
			ErrMalformedFrontmatter, req.Item, issueschema.ReadingItemFamily)
	}
	if err := mutationPreamble(issuesRoot); err != nil {
		return DispositionResult{}, err
	}

	// The position comes off the KEYED reading record, never from the caller: the
	// availability rule is a coupling the schema carries, and a caller-supplied
	// position would let a disposition assert the very rule it must satisfy. An
	// orphan disposition (no such item) is refused by this same path, because the
	// check has no position to reason with.
	position, err := readingItemPosition(issuesRoot, req.Item)
	if err != nil {
		return DispositionResult{}, err
	}

	id, err := minter.Mint(issueschema.DispositionFamily)
	if err != nil {
		return DispositionResult{}, err
	}
	fields, fm, redacted, degraded, err := dispositionFields(repoRoot, id, req)
	if err != nil {
		return DispositionResult{}, err
	}
	if err := validateDispositionStrict(fm, position); err != nil {
		return DispositionResult{}, err
	}
	content, err := buildIssueText(fields, "")
	if err != nil {
		return DispositionResult{}, err
	}

	itemDir := filepath.Join(issuesRoot, issueschema.DispositionsDir, req.Item)
	path := filepath.Join(itemDir, id+".md")
	err = withLedgerLock(issuesRoot, func() error {
		if err := ensureFamilyDir(issuesRoot, issueschema.DispositionsDir, req.Item); err != nil {
			return err
		}
		// A second answer to one item must say which one it replaces, and it must
		// be checked HERE, under the lock: the standing disposition is a property of
		// the directory as it is at the moment of the write.
		standing, err := standingDispositions(itemDir)
		if err != nil {
			return err
		}
		if req.Supersedes != "" && !containsString(standing, req.Supersedes) {
			return fmt.Errorf("%w: supersedes_disposition names %q, which is not a standing disposition of %s (standing: %s)",
				ErrInvariantViolation, req.Supersedes, req.Item, renderList(standing))
		}
		if req.Supersedes == "" && len(standing) > 0 {
			return fmt.Errorf("%w: %s already carries a standing disposition (%s); a second answer must cite the one it replaces (supersedes_disposition), so the record can say which is in force",
				ErrInvariantViolation, req.Item, renderList(standing))
		}
		if err := refuseExistingRecord(path, id); err != nil {
			return err
		}
		return fsutil.WriteFileAtomic(path, []byte(content), 0o644)
	})
	if err != nil {
		return DispositionResult{}, err
	}
	return DispositionResult{
		ID: id, Item: req.Item, State: req.State, Position: position,
		Path: fsutil.RepoRel(repoRoot, path), Redacted: redacted, Degraded: degraded,
	}, nil
}

// readingFields assembles one reading record's ordered frontmatter and the map
// its validator reads, redacting every free-text value on the way — a reading's
// text lands in the committed ledger exactly as a capture's does.
func readingFields(repoRoot, id string, req IngestReadingRequest, item ReadingItem) ([]kv, map[string]any, int, string, error) {
	redacted := 0
	degraded := ""
	scrub := func(s string) string {
		out, n, d := redactLedgerText(repoRoot, s)
		redacted += n
		if d != "" {
			degraded = d
		}
		return out
	}

	fields := []kv{
		{"schema_version", 1},
		{"id", id},
		{"run", req.Run},
		{"manifest", req.Manifest},
		{"position", req.Position},
		{"regime", req.Regime},
		{"pattern", scrub(item.Pattern)},
	}
	fm := map[string]any{
		"schema_version": 1,
		"id":             id,
		"run":            req.Run,
		"manifest":       req.Manifest,
		"position":       req.Position,
		"regime":         req.Regime,
		"pattern":        fields[len(fields)-1].val,
	}

	// Body fields are written in the position's DECLARED order, so two records at
	// one position are byte-comparable. A field the position does not declare is
	// still carried into the map, so the validator reports it rather than dropping
	// it silently.
	for _, f := range issueschema.ReadingBodyFields[req.Position] {
		v, ok := item.Body[f]
		if !ok {
			continue
		}
		s := scrub(v)
		fields = append(fields, kv{f, s})
		fm[f] = s
	}
	for f, v := range item.Body {
		if _, already := fm[f]; already {
			continue
		}
		fm[f] = v
	}
	if item.OccasionedBy != "" {
		fm["occasioned_by"] = item.OccasionedBy
	}
	return fields, fm, redacted, degraded, nil
}

// dispositionFields assembles one disposition's ordered frontmatter and map.
func dispositionFields(repoRoot, id string, req DispositionRequest) ([]kv, map[string]any, int, string, error) {
	redacted := 0
	degraded := ""
	scrub := func(s string) string {
		out, n, d := redactLedgerText(repoRoot, s)
		redacted += n
		if d != "" {
			degraded = d
		}
		return out
	}

	fields := []kv{
		{"schema_version", 1},
		{"id", id},
		{"item", req.Item},
		{"state", req.State},
	}
	fm := map[string]any{
		"schema_version": 1,
		"id":             id,
		"item":           req.Item,
		"state":          req.State,
	}
	add := func(key, value string) {
		if value == "" {
			return
		}
		s := scrub(value)
		fields = append(fields, kv{key, s})
		fm[key] = s
	}
	add("disposition_grounds", req.Grounds)
	add("exit_condition", req.ExitCondition)
	if req.Supersedes != "" {
		fields = append(fields, kv{"supersedes_disposition", req.Supersedes})
		fm["supersedes_disposition"] = req.Supersedes
	}
	if len(req.Recurs) > 0 {
		fields = append(fields, kv{"recurs", req.Recurs})
		fm["recurs"] = req.Recurs
	}
	// The reserved axes are carried into the map UNWRITTEN: the validator refuses
	// a populated value, so they never reach a field list.
	if req.HoldFrameLocation != "" {
		fm["hold_frame_location"] = req.HoldFrameLocation
	}
	if req.HoldMoscow != "" {
		fm["hold_moscow"] = req.HoldMoscow
	}
	return fields, fm, redacted, degraded, nil
}

// validateReadingStrict validates a reading record's frontmatter against the
// schema held in core/issueschema — the same data the committed-tree gate reads.
func validateReadingStrict(fm map[string]any) error {
	if err := requireSchemaVersion(fm); err != nil {
		return err
	}
	for k := range fm {
		if !issueschema.ReadingKnown[k] {
			return fmt.Errorf("%w: unknown property %q on a reading record", ErrMalformedFrontmatter, k)
		}
	}
	for _, req := range issueschema.ReadingRequired[1:] {
		if err := requireNonBlankString(fm, req); err != nil {
			return err
		}
	}
	id := fm["id"].(string)
	if !reReadingItemID.MatchString(id) {
		return fmt.Errorf("%w: id %q does not match ^%s-[0-9]+$", ErrMalformedFrontmatter, id, issueschema.ReadingItemFamily)
	}
	if run := fm["run"].(string); !reRunID.MatchString(run) {
		return fmt.Errorf("%w: run %q does not match ^%s-[0-9]+$", ErrMalformedFrontmatter, run, issueschema.ReadingRunFamily)
	}
	position := fm["position"].(string)
	want, known := issueschema.ReadingBodyFields[position]
	if !known {
		return fmt.Errorf("%w: position %q is not one of {%s}",
			ErrMalformedFrontmatter, position, strings.Join(issueschema.Positions, ", "))
	}

	// One record type, four bodies: the item carries the body its own position
	// declares, and no other. A field belonging to a different position is a known
	// property in the WRONG record, which is exactly the drift a single untyped
	// body has to be gated against.
	declared := map[string]bool{}
	for _, f := range want {
		declared[f] = true
		if err := requireNonBlankString(fm, f); err != nil {
			return fmt.Errorf("%w (the %s body, which the %s position returns, is %s)",
				err, issueschema.ReadingBodyName(position), position, strings.Join(want, " · "))
		}
	}
	for k := range fm {
		if declared[k] || isReadingEnvelopeField(k) {
			continue
		}
		if isReadingBodyField(k) {
			return fmt.Errorf("%w: %q belongs to another position's body; an item at the %s position carries %s",
				ErrMalformedFrontmatter, k, position, strings.Join(want, " · "))
		}
	}

	for _, f := range issueschema.ReservedSurpriseFields {
		if v, present := fm[f]; present && strings.TrimSpace(asString(v)) != "" {
			return fmt.Errorf("%w: %q is reserved and dormant — the surprise entry is a distinct record shape, populated once its shape is ruled; a populated value is refused rather than silently accepted",
				ErrInvariantViolation, f)
		}
	}
	if v, present := fm["promoted_to"]; present {
		if !reItdID.MatchString(asString(v)) {
			return fmt.Errorf("%w: promoted_to %q does not match ^itd-[0-9]+$", ErrMalformedFrontmatter, v)
		}
	}
	return nil
}

// validateDispositionStrict validates a disposition against the schema and
// against the position of the item it answers.
func validateDispositionStrict(fm map[string]any, position string) error {
	if err := requireSchemaVersion(fm); err != nil {
		return err
	}
	for k := range fm {
		if !issueschema.DispositionKnown[k] {
			return fmt.Errorf("%w: unknown property %q on a disposition", ErrMalformedFrontmatter, k)
		}
	}
	for _, req := range issueschema.DispositionRequired[1:] {
		if err := requireNonBlankString(fm, req); err != nil {
			return err
		}
	}
	if id := fm["id"].(string); !reDispositionID.MatchString(id) {
		return fmt.Errorf("%w: id %q does not match ^%s-[0-9]+$", ErrMalformedFrontmatter, id, issueschema.DispositionFamily)
	}
	if item := fm["item"].(string); !reReadingItemID.MatchString(item) {
		return fmt.Errorf("%w: item %q does not match ^%s-[0-9]+$", ErrMalformedFrontmatter, item, issueschema.ReadingItemFamily)
	}
	state := fm["state"].(string)
	if !containsString(issueschema.DispositionStates, state) {
		return fmt.Errorf("%w: state %q is not one of {%s}; nothing meaning \"already covered\" exists at any position — an undispositioned item is reported as outstanding, never named as a state",
			ErrMalformedFrontmatter, state, strings.Join(issueschema.DispositionStates, ", "))
	}

	// The grounds are what make a disposition a judgement. They are required on
	// every state except held, whose exit condition carries the same weight — and
	// an exit condition on any other state names an exit from a state that has
	// none, so it is refused rather than written and ignored.
	if state == issueschema.DispositionHeld {
		if err := requireNonBlankString(fm, "exit_condition"); err != nil {
			return fmt.Errorf("%w — a held disposition is directional, and a hold with no exit condition is the parking space `open` already is", err)
		}
	} else {
		if err := requireNonBlankString(fm, "disposition_grounds"); err != nil {
			return err
		}
		if v, present := fm["exit_condition"]; present && strings.TrimSpace(asString(v)) != "" {
			return fmt.Errorf("%w: exit_condition belongs to a held disposition; %q has no exit to name",
				ErrInvariantViolation, state)
		}
	}

	// The availability rule, read against the position the keyed reading record
	// carries. An UNRULED pair is permitted quietly: `held` at the widening
	// position is deferred, and refusing it here would settle by implementation
	// what the facilitator deferred.
	available, ruled := issueschema.DispositionStateAvailable(position, state)
	if ruled && !available {
		return fmt.Errorf("%w: %q is not available at the %s position (available there: %s); the disposition record validates its state against the envelope's position",
			ErrInvariantViolation, state, position, strings.Join(availableStates(position), ", "))
	}

	for _, f := range issueschema.ReservedHoldFields {
		if v, present := fm[f]; present && strings.TrimSpace(asString(v)) != "" {
			return fmt.Errorf("%w: %q is reserved and dormant — the two-axis hold field is present in the schema (frame-location: free text naming the frame element; MoSCoW: %s) and a populated value is refused until activation is ruled",
				ErrInvariantViolation, f, strings.Join(issueschema.HoldMoscowValues, " / "))
		}
	}
	if v, present := fm["supersedes_disposition"]; present {
		if !reDispositionID.MatchString(asString(v)) {
			return fmt.Errorf("%w: supersedes_disposition %q does not match ^%s-[0-9]+$",
				ErrMalformedFrontmatter, v, issueschema.DispositionFamily)
		}
	}
	if v, present := fm["recurs"]; present {
		items, isList := v.([]string)
		if !isList {
			return fmt.Errorf("%w: recurs must be a list of prior item ids", ErrMalformedFrontmatter)
		}
		for _, it := range items {
			if !reReadingItemID.MatchString(it) {
				return fmt.Errorf("%w: recurs item %q does not match ^%s-[0-9]+$",
					ErrMalformedFrontmatter, it, issueschema.ReadingItemFamily)
			}
		}
	}
	return nil
}

// readingItemPosition reads the position off a reading record, located by its id
// across every run directory. An id no run returned is an unknown item — the
// same sentinel an unknown issue id raises, because the fault is the same shape.
func readingItemPosition(issuesRoot, item string) (string, error) {
	path, err := findReadingItem(issuesRoot, item)
	if err != nil {
		return "", err
	}
	content, _, err := readWithChecksum(path)
	if err != nil {
		return "", err
	}
	fm, _, err := parseFrontmatterAndBody(content)
	if err != nil {
		return "", err
	}
	if err := validateReadingStrict(fm); err != nil {
		return "", err
	}
	return asString(fm["position"]), nil
}

// findReadingItem locates a reading record by id across the run directories. The
// id is run-scoped only in the sense that it was minted DURING a run: two runs
// can never mint one id, so the search needs no run argument and a caller
// dispositioning an item does not have to know which run returned it.
func findReadingItem(issuesRoot, item string) (string, error) {
	if !reReadingItemID.MatchString(item) {
		return "", fmt.Errorf("invalid %s-N identifier: %q", issueschema.ReadingItemFamily, item)
	}
	readingsRoot := filepath.Join(issuesRoot, issueschema.ReadingsDir)
	runs, err := os.ReadDir(readingsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s is not a reading item this ledger holds", ErrUnknownIssueID, item)
		}
		return "", err
	}
	var matches []string
	for _, run := range runs {
		if !run.IsDir() || !reRunID.MatchString(run.Name()) {
			continue
		}
		cand := filepath.Join(readingsRoot, run.Name(), item+".md")
		if fi, err := os.Lstat(cand); err == nil && fi.Mode().IsRegular() {
			matches = append(matches, cand)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("%w: %s is not a reading item this ledger holds", ErrUnknownIssueID, item)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%w: %s is present in more than one run directory", ErrDuplicateIssueID, item)
	}
}

// standingDispositions lists the dispositions of one item that no sibling
// supersedes — the answers currently in force. An empty result means the item is
// undispositioned, which the outstanding report says out loud and no state names.
func standingDispositions(itemDir string) ([]string, error) {
	entries, err := os.ReadDir(itemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	present := map[string]bool{}
	superseded := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		if !reDispositionID.MatchString(id) {
			continue
		}
		present[id] = true
		content, err := os.ReadFile(filepath.Join(itemDir, e.Name()))
		if err != nil {
			return nil, err
		}
		fm, _, err := parseFrontmatterAndBody(string(content))
		if err != nil {
			// A malformed sibling cannot be read as superseding anything, and it
			// must not silently license a second standing answer either — so it
			// counts as present and the caller is told to supersede it.
			continue
		}
		if s := asString(fm["supersedes_disposition"]); s != "" {
			superseded[s] = true
		}
	}
	var out []string
	for id := range present {
		if !superseded[id] {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out, nil
}

// refuseExistingRecord fails a write whose target is already taken. The id space
// is a UTC second plus four random digits, so two same-second draws CAN coincide
// — rare, and rare is exactly why it must refuse rather than overwrite: a
// committed record silently replaced by another leaves no trace that either
// happened, which is the one outcome a ledger must not produce. The check runs
// under the ledger lock, so between it and the write no other abcd process can
// claim the path.
func refuseExistingRecord(path, id string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%w: %s already exists in this ledger; the mint collided and the existing record is left untouched",
			ErrDuplicateIssueID, id)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ensureFamilyDir provisions <issuesRoot>/<family>/<key>, refusing a symlinked
// leaf at either level exactly as the status directories are provisioned.
func ensureFamilyDir(issuesRoot, family, key string) error {
	if err := safeMkdirLeaf(filepath.Join(issuesRoot, family)); err != nil {
		return err
	}
	return safeMkdirLeaf(filepath.Join(issuesRoot, family, key))
}

// requireSchemaVersion is the version check every record in this ledger opens
// with: a reader that cannot name the version it is reading cannot claim to have
// validated anything.
func requireSchemaVersion(fm map[string]any) error {
	sv, ok := fm["schema_version"]
	if !ok {
		return fmt.Errorf("%w: missing required property 'schema_version'", ErrMissingRequiredField)
	}
	if n, isInt := sv.(int); !isInt || n != 1 {
		return fmt.Errorf("%w: unsupported schema_version %v (this reader only handles 1)", ErrMissingRequiredField, sv)
	}
	return nil
}

// requireNonBlankString demands a present, string-typed, non-blank property. A
// property present but blank counts as missing for the same reason it does on an
// issue: the reader cannot make a value out of it either.
func requireNonBlankString(fm map[string]any, key string) error {
	v, present := fm[key]
	if !present {
		return fmt.Errorf("%w: missing required property %q", ErrMissingRequiredField, key)
	}
	s, isStr := v.(string)
	if !isStr {
		return fmt.Errorf("%w: %q must be a string", ErrMalformedFrontmatter, key)
	}
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%w: %q must be non-empty", ErrMissingRequiredField, key)
	}
	return nil
}

// isReadingEnvelopeField reports whether key is part of the envelope every
// reading record carries whatever its position.
func isReadingEnvelopeField(key string) bool {
	if containsString(issueschema.ReadingRequired, key) {
		return true
	}
	if containsString(issueschema.ReservedSurpriseFields, key) {
		return true
	}
	return key == "promoted_to"
}

// isReadingBodyField reports whether key belongs to SOME position's body.
func isReadingBodyField(key string) bool {
	for _, p := range issueschema.ReadingPositions {
		if containsString(p.Fields, key) {
			return true
		}
	}
	return false
}

// availableStates renders the states available at a position, for a refusal that
// names the rule rather than merely refusing.
func availableStates(position string) []string {
	var out []string
	for _, s := range issueschema.DispositionStates {
		if available, ruled := issueschema.DispositionStateAvailable(position, s); available && ruled {
			out = append(out, s)
		}
	}
	return out
}

func containsString(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

func renderList(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ", ")
}
