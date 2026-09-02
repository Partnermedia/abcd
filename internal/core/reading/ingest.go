package reading

// ingest.go is the cold-reading OUTPUT contract (itd-185, spc-63): a reading
// emits JSON, this verb validates it, and this verb writes the record. It is the
// output-contract idiom this repository already carries — `intent audit ingest
// --verdict-json` and `launch ship --changelog-json` — applied to a reading, and
// it adds no second one.
//
// Three properties make it more than a schema check.
//
// Ids are MINTED here. The payload carries no item identifier at any level, so a
// payload that supplies one is refused as an unknown field. A reading holds no
// mint, and an id it chose could collide with, or impersonate, one the ledger
// already holds. The mint itself is capture.IngestReading's, under the ledger
// lock, because that is where the collision probe can see the tree it is about
// to write into.
//
// The supply REGIME is the definition's. It is resolved from the run's position
// through LoadDefinition, and the payload's self-declared regime is compared
// against that rather than trusted. There is no --regime flag and no
// configuration key, and internal/surface/cli/regime_surface_test.go is the
// standing guard that no operator surface grows one.
//
// PROVENANCE is enforced at every regime without exception: an item whose
// envelope pattern field is empty or absent is refused. The definitions instruct
// it, this contract enforces it, and nothing else checks it.
//
// This is a trust boundary. The payload is read behind fsutil.ReadGuarded with a
// byte cap, decoded strictly, and no payload string is joined into a path before
// its grammar has been checked: the run id is matched against
// recordid.ValidReadingRunID first, and it is the ONLY payload value any path is
// built from. Every payload-derived string that reaches a terminal or a durable
// record passes through termsafe.Sanitize and a length cap; no item body text
// reaches either, because the bodies belong in the ledger records, which
// capture.IngestReading redacts.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/intentdriven/abcd/internal/core/capture"
	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/core/recordid"
	"github.com/intentdriven/abcd/internal/fsutil"
	"github.com/intentdriven/abcd/internal/termsafe"
)

// The three artefact type tags this file reads and writes. They are carried in
// the documents themselves, exactly as the bundle's and the manifest's are, so a
// reader of a loose file can tell them apart without their filenames.
const (
	// OutputType is the only _type the ingest accepts.
	OutputType = "abcd.reading.output/1"
	// RunType tags the run-metadata record, which is the commit marker.
	RunType = "abcd.reading.run/1"
	// RefusalType tags the record a list-level refusal leaves behind.
	RefusalType = "abcd.reading.refusal/1"
	// StageType tags the write-aside marker an in-flight ingest parks.
	StageType = "abcd.reading.ingest-stage/1"
)

// IngestStageDir is the local-tier write-aside area one ingest stages into. It
// is deliberately in the ephemeral tier: a stage is evidence of an ingest in
// flight, never a record, and a crash must leave it somewhere no commit can pick
// it up.
const IngestStageDir = ".abcd/.work.local/scratch/reading-ingest"

// ReadingsRecordDir is the durable home of a run's own artefacts: the promoted
// manifest, the run metadata, and a refusal.
const ReadingsRecordDir = ".abcd/development/readings"

// The three filenames under a run's durable directory, and the stage's lock.
const (
	RunFileName     = "run.json"
	RefusalFileName = "refusal.json"
	stageFileName   = "stage.json"
	stageLockName   = ".lock"
)

// ingestLockTimeout bounds the wait for a peer ingest to finish. An ingest is a
// validation pass and one batch of small writes, so a wait past this is a stuck
// process rather than a busy one, and reporting contention beats hanging.
const ingestLockTimeout = 30 * time.Second

// PatternField is the envelope's provenance field: the pattern the reading read
// under. It is the reading RECORD's own field name (issueschema.ReadingRequired)
// and the key the four definitions instruct, so the payload, the record and the
// instruction all say one word. itd-185 and spc-63 call the same field the
// pattern named; there is one field, and this is its wire name.
const PatternField = "pattern"

// maxEchoedBytes caps a payload-derived string echoed into a message or a
// durable record. A refusal that quoted a four-megabyte field would be a denial
// of service against the reader of the refusal.
//
// maxQuotedNames caps how MANY payload-chosen names one refusal quotes, which is
// the same denial in the other dimension: a per-name cap bounds nothing when the
// payload chooses the number of names.
const (
	maxEchoedBytes = 120
	maxQuotedNames = 8
	// maxReportedRefusals caps how many item refusals reach a message and a
	// durable record, for the same reason in a third dimension: the number of
	// ITEMS is payload-chosen too.
	maxReportedRefusals = 20
)

// isReadingItemFile reports whether name is a reading record's filename.
//
// A rollback deletes what an ingest wrote and nothing else, so it matches file
// NAMES rather than clearing a directory — and the id inside the name is judged
// by the shared predicate, not by a second copy of the grammar here. A delete
// bounded by a pattern is only as bounded as the pattern.
func isReadingItemFile(name string) bool {
	id, ok := strings.CutSuffix(name, ".md")
	return ok && recordid.ValidReadingItemID(id)
}

// Instrument is the identity of the thing that read: the model, the content hash
// of the definition it read under, and the version of the assembler that built
// its input. All three are required, and all three are carried into the run
// metadata, so two runs claiming the same instrument are provably the same —
// which is what the closing-run comparison rests on (ruling (12)).
type Instrument struct {
	Model            string `json:"model"`
	DefinitionSHA256 string `json:"definition_sha256"`
	AssemblerVersion string `json:"assembler_version"`
}

// Output is one reading's whole return, as it arrives.
//
// Items are decoded as raw key/value maps rather than as a typed struct on
// purpose. A Go struct would need four shapes for the four position bodies, and
// json's DisallowUnknownFields would then report a licence breach as a bare
// unknown field. The key set is closed HERE instead, against the position's own
// body fields, which is strictly stronger: an unknown key is still refused, and
// a key on the position's reserved-name table is refused with the licence named.
type Output struct {
	Type           string                       `json:"_type"`
	RunID          string                       `json:"run_id"`
	Position       Position                     `json:"position"`
	Regime         string                       `json:"regime"`
	ManifestSHA256 string                       `json:"manifest_sha256"`
	Instrument     Instrument                   `json:"instrument"`
	Items          []map[string]json.RawMessage `json:"items"`
}

// ItemRefusal is one item the run refused, and why. It carries no item body
// text: a refusal names the ordinal, the rule and the offending field, which is
// everything a reader needs and nothing a redactor would have to clean.
//
// Ordinal is omitted when it is zero, because zero is not an ordinal: items are
// numbered from one, and the one entry carrying none is the elision entry a
// bounded list ends with (boundedRefusals). A durable record that wrote
// `"ordinal": 0` named an item that does not exist as surely as a terminal
// printing "item 0" would (iss-2608311518250688).
type ItemRefusal struct {
	Ordinal int    `json:"ordinal,omitempty"`
	Rule    string `json:"rule"`
	Field   string `json:"field,omitempty"`
	Detail  string `json:"detail"`
}

// refusalsElidedRule names the elision entry a bounded refusal list ends with.
const refusalsElidedRule = "refusals-elided"

// IsElision reports whether r is the list's elision entry rather than an item.
// An item has an ordinal from one; the elision entry has none, and nothing else
// is ever built without one. This is the ONE rule both surfaces render by.
func (r ItemRefusal) IsElision() bool {
	return r.Ordinal == 0
}

// Render is the one-line form of a refusal, shared by every surface that prints
// the list — the refusal record's reason and the terminal render. There is no
// item 0: the elision entry renders under its rule alone, and a reader is never
// sent looking for an item that does not exist. The terminal once carried this
// branch and the record writer did not, and the durable record was the one that
// named item 0 (iss-2608311518250688).
func (r ItemRefusal) Render() string {
	if r.IsElision() {
		return fmt.Sprintf("(%s) %s", r.Rule, r.Detail)
	}
	return fmt.Sprintf("item %d (%s): %s", r.Ordinal, r.Rule, r.Detail)
}

// ReviewFlag is a signature hit that did not refuse: the generative regime's
// path, where the licence is the widest and the constraint falls at admission
// instead.
type ReviewFlag struct {
	Ordinal     int    `json:"ordinal"`
	SignatureID string `json:"signature_id"`
	Detail      string `json:"detail"`
}

// RunRecord is the run metadata, written LAST as the commit marker: a run
// without one never happened.
type RunRecord struct {
	Type           string                     `json:"_type"`
	SchemaVersion  int                        `json:"schema_version"`
	RunID          string                     `json:"run_id"`
	Position       Position                   `json:"position"`
	Regime         string                     `json:"regime"`
	TargetCommit   string                     `json:"target_commit"`
	ManifestSHA256 string                     `json:"manifest_sha256"`
	Instrument     Instrument                 `json:"instrument"`
	Records        []capture.ReadingRecordRef `json:"records"`
	RefusedItems   []ItemRefusal              `json:"refused_items"`
	RefusedCount   int                        `json:"refused_count"`
	ReviewFlags    []ReviewFlag               `json:"review_flags"`
}

// RefusalRecord is what a list-level refusal leaves behind: the run metadata and
// the named reason, and no items. The event is durable, and a rerun is a new run
// with a new run id, never an amendment.
//
// Regime is omitted when the run has none: a run refused because its position's
// definition did not resolve — absent, malformed, or stating another position's
// licence — read under no regime the verb could name, and the record says
// nothing rather than something the verb never read. The reason names the
// definition instead.
type RefusalRecord struct {
	Type           string     `json:"_type"`
	SchemaVersion  int        `json:"schema_version"`
	RunID          string     `json:"run_id"`
	Position       Position   `json:"position"`
	Regime         string     `json:"regime,omitempty"`
	TargetCommit   string     `json:"target_commit"`
	ManifestSHA256 string     `json:"manifest_sha256"`
	Instrument     Instrument `json:"instrument"`
	Reason         string     `json:"reason"`
}

// stageMarker is the write-aside marker. It names the run in flight and, once
// the ledger write has returned, the records that landed — so a rollback removes
// exactly what this ingest wrote.
type stageMarker struct {
	Type    string   `json:"_type"`
	RunID   string   `json:"run_id"`
	Records []string `json:"records"`
}

// IngestRequest is one ingest: a repository and one payload file.
//
// There is no position operand and no regime operand. The position is the
// payload's claim, checked against the manifest of the run it names; the regime
// is the definition's, and nothing an operator types can reach it.
type IngestRequest struct {
	RepoRoot string
	// OutputPath is the reading's returned JSON. The front door resolves a
	// relative path against the working directory before it arrives here.
	OutputPath string
}

// IngestResult is what an ingest did.
type IngestResult struct {
	RunID        string                     `json:"run_id"`
	Position     Position                   `json:"position"`
	Regime       string                     `json:"regime"`
	Records      []capture.ReadingRecordRef `json:"records"`
	RefusedItems []ItemRefusal              `json:"refused_items,omitempty"`
	// RefusedCount is how many items were refused in total. RefusedItems is
	// capped — the item count is payload-chosen — so the two differ when a run
	// refused more than the cap, and the count is what nothing truncates.
	RefusedCount  int          `json:"refused_count,omitempty"`
	ReviewFlags   []ReviewFlag `json:"review_flags,omitempty"`
	RunRecordPath string       `json:"run_record,omitempty"`
	RefusalPath   string       `json:"refusal_record,omitempty"`
	ClearedStages []string     `json:"cleared_stages,omitempty"`
	// RolledBack names the reading records the sweep REMOVED from the committed
	// ledger, because their run never reached its commit marker. A delete in the
	// committed tier is reported by id: "cleared an orphaned stage" does not tell
	// an operator that records left the ledger with it.
	RolledBack []string `json:"rolled_back_records,omitempty"`
	// PendingStages names the orphaned stages this invocation found and LEFT
	// IN PLACE. The sweep is a delete in the committed tier, and every such
	// delete is fenced behind the whole payload validating, so an invocation
	// that is refused — or that fails before the sweep — reports the orphans
	// it saw and leaves them for the next one that validates. A sweep that
	// skipped something says so.
	//
	// LEFT IN PLACE is the whole of it: a stage this invocation cleared is not
	// on this list, whichever path cleared it — the sweep's, or the single
	// rollback a refusal makes on its own run id.
	PendingStages []string `json:"pending_stages,omitempty"`
	Redacted      int      `json:"redacted,omitempty"`
	Degraded      string   `json:"redaction_degraded,omitempty"`
}

// HasDisclosure reports whether the result carries something the operator has
// to be told even when the verb returned an error: a refusal record written, a
// stage cleared, a record removed from the ledger, an orphan seen and left in
// place, or a degraded outcome. The front doors render the result on the error
// path exactly when this is true — one predicate, so a field added here cannot
// be one the surface silently drops.
func (r IngestResult) HasDisclosure() bool {
	return r.RefusalPath != "" || len(r.ClearedStages) > 0 || len(r.RolledBack) > 0 ||
		len(r.PendingStages) > 0 || r.Degraded != ""
}

// The three points the test seam can be entered at: the two windows of the
// staged-write protocol, and the sweep's unlink of a rolled-back run's records.
const (
	faultAfterStage     = "after-stage"
	faultAfterLedger    = "after-ledger"
	faultDuringRollback = "during-rollback"
)

// ingestFault is the staged-write protocol's test seam, nil in production.
//
// The protocol's whole claim is about what a CRASH leaves behind, and a crash
// window cannot be observed from outside the process. A protocol whose crash
// behaviour is never executed is a protocol nobody has checked, so the two
// windows are reachable from a test rather than argued about in a comment.
var ingestFault func(step string) error

func fireFault(step string) error {
	if ingestFault == nil {
		return nil
	}
	return ingestFault(step)
}

// Ingest validates one reading's output and writes its records.
//
// The order is the protocol, and it is the order for a reason: no OTHER run's
// durable state is mutated by a run that is refused — nothing is written and
// nothing is deleted until the whole payload validates, save the refusal record
// that a refusal after step 3 exists to leave and the rollback of this run's
// own earlier crashed attempt that ac-10 obliges it to make — the ledger
// records land as one batch, and the run metadata lands last as the commit
// marker.
//
//  1. Find any orphaned stage a previous invocation left, read-only, so a
//     refusal below can name what it saw. Nothing is cleared yet.
//  2. Read and decode the payload strictly.
//  3. Check the envelope, and resolve the parked run's manifest by content hash.
//     Until this passes, the run has no proven identity, so nothing — not even a
//     refusal — can be recorded against it.
//  4. Resolve the definition and check the regime and the instrument against it.
//     A refusal from here on writes a refusal record.
//  5. Validate every item; an item-level violation refuses that item and lands
//     the rest.
//  6. Sweep the orphans found at step 1, naming and clearing each and rolling
//     its never-committed run back. This is the first delete in the committed
//     tier, and the payload it runs under has validated in full.
//  7. Stage, write the ledger records, promote the manifest, and write the run
//     metadata last.
//
// The sweep sat at step 1 once, and with an orphan present an ingest refused at
// the type check deleted that orphan's ledger records and reported the type
// error alone (iss-2608311517509690). A refused run now leaves the orphans it
// found in place and reports them as pending — and because it leaves them, the
// bare `reading` verb names them as `orphaned_ingests`, which is the only
// surface an operator has for the state.
func Ingest(req IngestRequest) (IngestResult, error) {
	res := IngestResult{}
	repoRoot := strings.TrimSpace(req.RepoRoot)
	if repoRoot == "" {
		return res, errors.New("reading: ingesting an output needs a repository root")
	}
	if strings.TrimSpace(req.OutputPath) == "" {
		return res, errors.New("reading: ingest needs the reading's output JSON")
	}

	// Every path this verb reads, writes or DELETES inside the repository is
	// resolved through this root.
	//
	// Containment cannot rest on the run-id grammar. That check makes the run id
	// a single safe path COMPONENT and says nothing about the components above
	// it, and this verb walks and removes files under two directories a hostile
	// clone can commit a symlink at — git mode 120000 on the ledger's run
	// directory or on the readings tree lands a write, or a DELETE, outside the
	// repository, and the orphan sweep is reached by any payload that validates.
	// os.Root resolves every component in the kernel and refuses the traversal,
	// which is the containment fsutil.ReadGuardedInRoot and WriteFileAtomicInRoot
	// already exist to give, and the stance core/capture takes on the same ledger
	// directory.
	root, err := os.OpenRoot(repoRoot)
	if err != nil {
		return res, fmt.Errorf("reading: opening the repository root: %w", err)
	}
	defer root.Close()

	// One ingest at a time in one checkout, from the orphan probe through the
	// commit marker.
	//
	// The sweep DELETES committed reading records, and its only test for an
	// orphan is that a stage exists with no commit marker beside it — which is
	// exactly what a live ingest looks like between its ledger write and its
	// marker. Without this lock a second invocation rolls the first one back
	// mid-flight, and the first then writes a run record naming records that no
	// longer exist and exits 0. capture's ledger lock cannot serve:
	// IngestReading re-takes it internally, and the sweep sits outside it.
	if err := ensureStageRoot(root); err != nil {
		return res, err
	}
	lock := filepath.Join(repoRoot, filepath.FromSlash(IngestStageDir), stageLockName)
	err = fsutil.WithFileLock(lock, ingestLockTimeout, func() error {
		return ingestUnderLock(root, repoRoot, req, &res)
	})
	if errors.Is(err, fsutil.ErrLockContention) {
		return res, fmt.Errorf("reading: another ingest is running in this checkout and did not finish "+
			"within %s; the sweep removes a run whose commit marker is missing, so two at once would "+
			"roll each other back", ingestLockTimeout)
	}
	return res, err
}

// ensureStageRoot creates the stage root through the containment root and
// refuses a symlinked one, so the lock file and every sweep after it act on a
// real directory inside this repository.
func ensureStageRoot(root *os.Root) error {
	if err := root.MkdirAll(IngestStageDir, 0o755); err != nil {
		return fmt.Errorf("reading: preparing the ingest stage: %w", err)
	}
	fi, err := root.Lstat(IngestStageDir)
	if err != nil {
		return fmt.Errorf("reading: preparing the ingest stage: %w", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("reading: %s is not a real directory; the ingest stage is a directory in this "+
			"repository, never a pointer at another one", IngestStageDir)
	}
	return nil
}

// ingestUnderLock is Ingest's body, everything the stage lock has to cover.
func ingestUnderLock(root *os.Root, repoRoot string, req IngestRequest, res *IngestResult) error {
	// Read-only. The orphans are named now so that every path out of this
	// function — a refusal before the run is proven, a recorded refusal, a
	// failed write — can say what it saw; they are cleared only below, once the
	// payload has validated in full. Until then every one of them is pending.
	orphans, err := findOrphanStages(root)
	if err != nil {
		return err
	}
	res.PendingStages = orphans

	raw, err := readOutputFile(req.OutputPath)
	if err != nil {
		return err
	}
	out, err := decodeOutput(raw)
	if err != nil {
		return err
	}

	pos, err := checkEnvelope(out)
	if err != nil {
		return err
	}
	res.RunID = out.RunID
	res.Position = pos

	manifest, err := resolveParkedManifest(root, out)
	if err != nil {
		return err
	}

	// The run's identity is proven from here: the payload names a parked run and
	// cites its manifest by the manifest's own content hash. A refusal below is
	// therefore recordable against that run — which is exactly why a run that
	// already HAS an outcome is refused before anything can overwrite one.
	if err := refuseARerun(root, out.RunID); err != nil {
		return err
	}

	// A definition that does not resolve refuses the run, and the refusal is
	// RECORDED like every other one from this point on: the identity is proven
	// above, so the run happened. The record states no regime, because the
	// regime is the definition's and this definition did not resolve — an empty
	// field is the honest value, and a substituted one would be the verb
	// asserting a licence it refused to read. Returning bare instead left a
	// refused run with nothing durable to find it by (iss-2608311518250688).
	def, err := LoadDefinition(repoRoot, pos)
	if err != nil {
		return refuse(root, res, out, manifest, Definition{Position: pos}, err)
	}
	res.Regime = def.Regime

	if err := checkRegime(out, def); err != nil {
		return refuse(root, res, out, manifest, def, err)
	}
	if err := checkInstrument(out, def, manifest); err != nil {
		return refuse(root, res, out, manifest, def, err)
	}

	items, refusals, flags, refusedCount, err := validateItems(out, def)
	res.RefusedItems = refusals
	res.RefusedCount = refusedCount
	res.ReviewFlags = flags
	if err != nil {
		return refuse(root, res, out, manifest, def, err)
	}

	// The whole payload has validated: this is the first point at which the
	// committed tier may be deleted from. The sweep reports what it cleared
	// and what it rolled back, and whatever it did not reach stays pending —
	// set before the error check, so a sweep that stopped short is reported
	// as far as it got rather than not at all.
	//
	// The fence is around SOMEBODY ELSE'S records: an ingest refused at the
	// `_type` check once deleted a committed reading record and reported the
	// `_type` error alone (iss-2608311517509690). It is not around this run's
	// OWN half-landed records — refuse() rolls those back, because ac-10 says a
	// refused run leaves no reading records and they are not somebody else's.
	//
	// So an orphan of another run outlives an invocation that does not validate.
	// It is reported as `pending_stages` here, and the bare `reading` verb
	// reports it as `orphaned_ingests`; `staged_runs` is a different directory
	// and does not show it.
	cleared, rolledBack, err := sweepOrphanStages(root, orphans)
	res.ClearedStages = cleared
	res.RolledBack = rolledBack
	res.PendingStages = leftPending(orphans, cleared)
	if err != nil {
		return err
	}

	return write(root, repoRoot, res, out, manifest, def, items)
}

// leftPending is the orphans found minus the stages that were cleared: what a
// later invocation will find again. Both lists are sorted where the sweep is the
// caller, and there the cleared list is a prefix of the found one in sweep
// order, but the subtraction rests on neither — the refusal path subtracts a
// single id from a list it does not otherwise touch.
func leftPending(found, cleared []string) []string {
	if len(cleared) == 0 {
		return found
	}
	done := make(map[string]bool, len(cleared))
	for _, c := range cleared {
		done[c] = true
	}
	var pending []string
	for _, f := range found {
		if !done[f] {
			pending = append(pending, f)
		}
	}
	return pending
}

// readOutputFile reads the untrusted payload behind fsutil.ReadGuarded
// (O_NOFOLLOW plus a regular-file check on the open fd plus a size cap, in one
// call). The single guarded open is the only race-free form: an Lstat-then-read
// pair leaves a window in which a symlink swapped in afterwards is followed.
func readOutputFile(path string) ([]byte, error) {
	data, err := fsutil.ReadGuarded(path, MaxFileBytes)
	if err != nil {
		switch {
		case errors.Is(err, fsutil.ErrNotRegular) || errors.Is(err, syscall.ELOOP):
			return nil, fmt.Errorf("reading: the output at %s is not a regular file "+
				"(a symlink or non-regular operand is refused)", path)
		case errors.Is(err, fsutil.ErrTooBig):
			return nil, fmt.Errorf("reading: the output at %s exceeds the %d-byte cap", path, MaxFileBytes)
		default:
			return nil, fmt.Errorf("reading: the output at %s: %w", path, err)
		}
	}
	return data, nil
}

// decodeOutput decodes the payload strictly. Unknown fields are refused at every
// declared level, and trailing content after the document is refused too: a
// second document appended to the first is a payload nobody has read.
func decodeOutput(raw []byte) (Output, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	var out Output
	if err := dec.Decode(&out); err != nil {
		return Output{}, fmt.Errorf("reading: the output is malformed: %s", echo(err.Error()))
	}
	if dec.More() {
		return Output{}, errors.New("reading: the output carries trailing content after the document")
	}
	return out, nil
}

// checkEnvelope validates the run-level fields that must hold before any path is
// built or any file is opened. The run id is checked FIRST among the values a
// path is built from, because it is the only payload value that ever becomes one.
func checkEnvelope(out Output) (Position, error) {
	if out.Type != OutputType {
		return "", fmt.Errorf("reading: the output states _type %q, want %q", echo(out.Type), OutputType)
	}
	if !recordid.ValidReadingRunID(out.RunID) {
		return "", fmt.Errorf("reading: run_id %q is not a run identifier (%s-N); "+
			"an ingest names the run an assembly parked, and a run id becomes a directory name",
			echo(out.RunID), RunIDFamily)
	}
	// The parser quotes the token it refused, and that token is payload text, so
	// the whole message goes through echo rather than the value alone.
	pos, err := ParsePosition(string(out.Position))
	if err != nil {
		return "", fmt.Errorf("reading: %s", echo(err.Error()))
	}
	if !sha256HexRe.MatchString(out.ManifestSHA256) {
		return "", fmt.Errorf("reading: manifest_sha256 %q is not a sha-256 digest", echo(out.ManifestSHA256))
	}
	if isBlank(out.Instrument.Model) || isBlank(out.Instrument.DefinitionSHA256) ||
		isBlank(out.Instrument.AssemblerVersion) {
		return "", errors.New("reading: instrument needs all three of model, definition_sha256 and " +
			"assembler_version; two runs claiming one instrument are provably the same only if the claim is whole")
	}
	if len(out.Items) == 0 {
		return "", errors.New("reading: the output carries no items; a run that returned nothing is " +
			"recorded as a run with an empty item set, which is not an ingest")
	}
	return pos, nil
}

// sha256HexRe is the shape of a content hash as this contract cites one.
var sha256HexRe = regexp.MustCompile(`^[0-9a-f]{64}$`)

// resolveParkedManifest finds the run the payload names, and proves the citation.
//
// manifest_sha256 is the content hash of the assembler's manifest — the one
// unforgeable reference, because it cannot be asserted without the bytes. A
// reference that resolves to nothing, or to a manifest whose hash disagrees,
// refuses the run.
func resolveParkedManifest(root *os.Root, out Output) (Manifest, error) {
	// out.RunID has already been matched against the run-id grammar, which makes
	// it a single safe path COMPONENT: it holds no separator and no dot. That
	// says nothing about the components above it, so the read is resolved through
	// the repository root as well — a manifest served from outside this
	// repository through a symlinked ancestor is not this repository's run.
	rel := DefaultRunDir + "/" + out.RunID + "/" + ManifestFileName
	raw, err := fsutil.ReadGuardedInRoot(root, rel, MaxFileBytes)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Manifest{}, fmt.Errorf("reading: %s cites run %s, whose manifest is not parked at %s; "+
				"an ingest resolves the run an assembly staged, and a citation that resolves to nothing "+
				"refuses the run", OutputType, out.RunID, rel)
		}
		return Manifest{}, fmt.Errorf("reading: the manifest of run %s at %s: %w", out.RunID, rel, err)
	}
	m, err := DecodeManifest(raw)
	if err != nil {
		return Manifest{}, fmt.Errorf("reading: the manifest of run %s: %w", out.RunID, err)
	}
	if got := sha256Hex(raw); got != out.ManifestSHA256 {
		return Manifest{}, fmt.Errorf("reading: manifest_sha256 is %s, and the manifest parked at %s hashes "+
			"to %s; the reference is the manifest's own content hash, and a disagreement refuses the run",
			echo(out.ManifestSHA256), rel, got)
	}
	// The manifest is a file on disk rather than the payload, but it is no more
	// trusted for that: it is read back at the operator's word and its values
	// reach the same terminal, so they are echoed under the same rule.
	if m.RunID != out.RunID {
		return Manifest{}, fmt.Errorf("reading: the manifest at %s names run %s, not %s",
			rel, echo(m.RunID), out.RunID)
	}
	if m.Position != out.Position {
		return Manifest{}, fmt.Errorf("reading: the output reads at the %s position and the manifest of run %s "+
			"names the %s position", echo(string(out.Position)), out.RunID, echo(string(m.Position)))
	}
	return m, nil
}

// write is the staged-write protocol. Nothing durable exists for the run until
// step 1 has already validated everything, and the run metadata is written last.
func write(root *os.Root, repoRoot string, res *IngestResult, out Output, m Manifest, def Definition, items []capture.ReadingItem) error {
	stageRel := IngestStageDir + "/" + out.RunID
	marker := stageMarker{Type: StageType, RunID: out.RunID, Records: []string{}}
	if err := writeJSONIn(root, stageRel+"/"+stageFileName, marker); err != nil {
		return err
	}
	if err := fireFault(faultAfterStage); err != nil {
		return err
	}

	written, err := capture.IngestReading(capture.IngestReadingRequest{
		RepoRoot: repoRoot,
		Run:      out.RunID,
		Manifest: out.ManifestSHA256,
		Position: string(def.Position),
		Regime:   def.Regime,
		Items:    items,
	})
	res.Records = written.Records
	res.Redacted = written.Redacted
	res.Degraded = written.Degraded
	if err != nil {
		return fmt.Errorf("reading: writing the records of run %s: %w", out.RunID, err)
	}

	// The marker names what landed. The rollback does NOT read it — it removes by
	// the item-id filename grammar over the run's own directory, which also
	// catches a batch that failed before this second write — so what the marker
	// buys is evidence a person can read: which run was in flight, and which ids
	// it had reached. Saying that plainly is the point; a comment claiming the
	// rollback consults it would be describing code that does not exist.
	marker.Records = recordIDs(written.Records)
	if err := writeJSONIn(root, stageRel+"/"+stageFileName, marker); err != nil {
		return err
	}
	if err := fireFault(faultAfterLedger); err != nil {
		return err
	}

	runRel := ReadingsRecordDir + "/" + out.RunID
	if err := writeJSONIn(root, runRel+"/"+ManifestFileName, m); err != nil {
		return err
	}
	run := RunRecord{
		Type: RunType, SchemaVersion: SchemaVersion, RunID: out.RunID,
		Position: def.Position, Regime: def.Regime, TargetCommit: m.TargetCommit,
		ManifestSHA256: out.ManifestSHA256, Instrument: sanitizeInstrument(out.Instrument),
		Records: written.Records, RefusedItems: res.RefusedItems, RefusedCount: res.RefusedCount,
		ReviewFlags: res.ReviewFlags,
	}
	if run.RefusedItems == nil {
		run.RefusedItems = []ItemRefusal{}
	}
	if run.ReviewFlags == nil {
		run.ReviewFlags = []ReviewFlag{}
	}
	if err := writeJSONIn(root, runRel+"/"+RunFileName, run); err != nil {
		return err
	}
	res.RunRecordPath = runRel + "/" + RunFileName

	// The commit marker is down, so the run HAPPENED. A stage that will not clear
	// is reported as what it is — a leftover the next invocation sweeps — rather
	// than as a failed ingest, because an operator told the run failed retries it,
	// and a retry of a committed run is refused (refuseARerun) or, worse, would
	// duplicate its records.
	if err := root.RemoveAll(stageRel); err != nil {
		res.Degraded = strings.TrimSpace(res.Degraded + " " + fmt.Sprintf(
			"the run committed but its stage at %s could not be cleared (%v); the bare `reading` verb "+
				"reports it, and the next ingest of a DIFFERENT run that validates sweeps it — a rerun of "+
				"this one is refused, because the run already has an outcome",
			stageRel, err))
	}
	return nil
}

// refuseARerun holds "a rerun is a NEW run with a new run id, never an
// amendment" as a check rather than as a sentence.
//
// The run id is payload-chosen, so without this a second ingest of one run
// overwrites its metadata while the first run's records stay in the ledger named
// by nothing — and unreachable from any later sweep, because the rollback bails
// whenever a commit marker exists. A refusal could likewise land beside a commit
// marker, leaving one directory asserting both that the run committed and that
// it was refused.
//
// It runs before the refusal path for that second reason: a rerun must not
// overwrite the refusal record of the run it is repeating either.
func refuseARerun(root *os.Root, runID string) error {
	for _, name := range []string{RunFileName, RefusalFileName} {
		rel := ReadingsRecordDir + "/" + runID + "/" + name
		_, err := root.Lstat(rel)
		switch {
		case err == nil:
			return fmt.Errorf("reading: run %s already has an outcome at %s; a rerun is a new run with a "+
				"new run id, never an amendment — assemble again, and ingest the run that assembly parked",
				runID, rel)
		case os.IsNotExist(err):
			continue
		default:
			return fmt.Errorf("reading: probing the outcome of run %s: %w", runID, err)
		}
	}
	return nil
}

// refuse records a list-level refusal and returns it. It is the ONE writer of a
// refusal record: every list-level refusal past the identity point routes
// through here, and a refusal that returns bare instead is the defect
// iss-2608311518250688 names.
//
// The record is durable because the event is: a refused run is a run that
// happened, and a rerun is a NEW run with a new run id, never an amendment. It
// carries the run metadata and the named reason and no items.
//
// OTHER runs' orphans, found at the start, stay exactly where they were: a run
// being refused for a reason of its own destroys nobody else's records. Its own
// are the exception — a previous ingest of this run id can have died between
// its ledger write and its commit marker, and ac-10 says a refused run leaves
// no reading records — so the one delete a refusal performs is rollbackThisRun
// on the id it was already proven to carry no outcome for.
func refuse(root *os.Root, res *IngestResult, out Output, m Manifest, def Definition, cause error) error {
	// ac-10's other half: a refused run leaves a refusal record and NO reading
	// records. Usually there are none to leave, because nothing has been staged
	// yet — but a PREVIOUS ingest of this run id can have died between its
	// ledger write and its commit marker, and refuseARerun has already proven
	// the id carries no outcome, so any records under it belong to a run that
	// never committed and they are this run's own. The general sweep is held
	// back to the commit path because it destroys OTHER runs' records; this is
	// the one rollback a refusal owes, and it is owed on every refusal path.
	if err := rollbackThisRun(root, res, out.RunID); err != nil {
		return fmt.Errorf("reading: %w (and the earlier attempt at run %s could not be rolled back: %v)",
			cause, out.RunID, err)
	}
	// The reason is carried WHOLE. Every payload-derived substring inside it
	// was already cleaned where it was interpolated, so a second cap here
	// would only cut the repository's own prose — and it did: a 338-rune
	// refusal reached the record as 123 runes, ending mid-word, and an
	// every-item-refused run lost its per-item refusals entirely. A record
	// whose stated purpose is to carry the named reason has to carry it.
	//
	// The one thing stripped is this package's own "reading: " prefix, which
	// the locator's errors carry and the checks' do not: the message below
	// adds it once, and a record is not the place for a stutter.
	reason := strings.TrimPrefix(cause.Error(), "reading: ")
	rec := RefusalRecord{
		Type: RefusalType, SchemaVersion: SchemaVersion, RunID: out.RunID,
		Position: def.Position, Regime: def.Regime, TargetCommit: m.TargetCommit,
		ManifestSHA256: out.ManifestSHA256, Instrument: sanitizeInstrument(out.Instrument),
		Reason: reason,
	}
	rel := ReadingsRecordDir + "/" + out.RunID + "/" + RefusalFileName
	if err := writeJSONIn(root, rel, rec); err != nil {
		return fmt.Errorf("reading: %s (and the refusal record could not be written: %v)", reason, err)
	}
	res.RefusalPath = rel
	return fmt.Errorf("reading: %s; the refusal is recorded at %s", reason, rel)
}

// rollbackThisRun removes one run's half-landed records and its stage, and says
// on the result what it removed. It is the sweep's rollback applied to a single
// named run rather than to whatever the stage directory happens to hold.
//
// It also drops the stage it cleared from PendingStages, which is the one place
// that subtraction can live for the refusal path. The orphan probe runs before
// anything is refused, so a run whose OWN earlier attempt left a stage finds
// itself in that list — and the refusal below then clears it. Left unsubtracted
// the disclosure named a stage that no longer existed, wrong by one entry in
// exactly the case it exists to describe (iss-2609020848468450). Pending means
// STILL STANDING: it is read against the tree, so the two have to agree.
func rollbackThisRun(root *os.Root, res *IngestResult, runID string) error {
	removed, err := rollbackRun(root, runID)
	res.RolledBack = append(res.RolledBack, removed...)
	if err != nil {
		return err
	}
	stageRel := IngestStageDir + "/" + runID
	if _, err := root.Lstat(stageRel); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading: probing the stage of run %s: %w", runID, err)
	}
	if err := root.RemoveAll(stageRel); err != nil {
		return fmt.Errorf("reading: clearing the stage of refused run %s: %w", runID, err)
	}
	res.ClearedStages = append(res.ClearedStages, runID)
	res.PendingStages = leftPending(res.PendingStages, []string{runID})
	return nil
}

// findOrphanStages lists, without touching anything, every stage a previous
// invocation left. It is the sweep's first half split off so that a refusal can
// report the orphans it is leaving in place, and it reads only the stage root —
// never the ledger, which the rollback alone walks.
func findOrphanStages(root *os.Root) ([]string, error) {
	entries, err := readDirIn(root, IngestStageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading: listing the ingest stage: %w", err)
	}
	var found []string
	for _, e := range entries {
		// Only a well-formed run id is an orphan, and only a real directory: a
		// DirEntry for a symlink reports IsDir false, so a planted link is
		// skipped rather than followed. A directory named anything else was not
		// written by this verb, and a sweep that removed it would be deleting
		// somebody else's file on a guess.
		if !e.IsDir() || !recordid.ValidReadingRunID(e.Name()) {
			continue
		}
		found = append(found, e.Name())
	}
	sort.Strings(found)
	return found, nil
}

// sweepOrphanStages reports and clears the orphaned stages findOrphanStages
// named. It runs only once the whole payload has validated: it is the verb's
// first delete in the committed tier, and a refused run must not reach it.
//
// An orphan means an ingest reached the stage and never reached its commit
// marker, so the run never happened — and a run that never happened must leave
// no reading records behind. The rollback therefore removes what the stage says
// the ingest wrote, then the run's own directory when the commit marker is
// absent. A stage whose run DID commit is a leftover and only the stage goes.
//
// On an error the two lists carry what was done before it: a caller reports a
// sweep that stopped short as far as it got, never as nothing.
func sweepOrphanStages(root *os.Root, stages []string) (cleared, rolledBack []string, err error) {
	for _, runID := range stages {
		removed, err := rollbackRun(root, runID)
		if err != nil {
			return cleared, rolledBack, err
		}
		rolledBack = append(rolledBack, removed...)
		if err := root.RemoveAll(IngestStageDir + "/" + runID); err != nil {
			return cleared, rolledBack, fmt.Errorf("reading: clearing the orphaned stage of run %s: %w",
				runID, err)
		}
		cleared = append(cleared, runID)
	}
	sort.Strings(cleared)
	sort.Strings(rolledBack)
	return cleared, rolledBack, nil
}

// rollbackRun removes the durable half of a run whose commit marker never
// landed. A run WITH a commit marker is left entirely alone: its stage is a
// leftover from a crash after the marker, and the run is complete.
func rollbackRun(root *os.Root, runID string) ([]string, error) {
	runRel := ReadingsRecordDir + "/" + runID
	_, err := root.Lstat(runRel + "/" + RunFileName)
	switch {
	case err == nil:
		return nil, nil
	case os.IsNotExist(err):
	default:
		return nil, fmt.Errorf("reading: probing the commit marker of run %s: %w", runID, err)
	}

	var removed []string
	// The stage lock serialises ingest against ingest. It says nothing about
	// core/capture, whose own verbs read these records — so the unlink below
	// takes the LEDGER lock as well, and a concurrent disposition or promote
	// waits rather than reading a record as it disappears.
	//
	// It is taken here and not around the whole ingest, because
	// capture.IngestReading re-takes it internally: an flock is not reentrant,
	// and holding it across that call would deadlock the verb against itself.
	// The two locks are always acquired stage-then-ledger and never the other
	// way, so no cycle exists.
	ledgerRel := capture.LedgerRelPath + "/" + issueschema.ReadingsDir + "/" + runID
	entries, err := readDirIn(root, ledgerRel)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading: listing the records of orphaned run %s: %w", runID, err)
	}
	unlink := func() error {
		if err := fireFault(faultDuringRollback); err != nil {
			return err
		}
		for _, e := range entries {
			// Bounded by the item-id grammar: a rollback removes reading records
			// and nothing else, so a file a person put in the directory survives.
			if e.IsDir() || !isReadingItemFile(e.Name()) {
				continue
			}
			if err := root.Remove(ledgerRel + "/" + e.Name()); err != nil {
				return fmt.Errorf("reading: rolling back record %s of run %s: %w", e.Name(), runID, err)
			}
			removed = append(removed, strings.TrimSuffix(e.Name(), ".md"))
		}
		// Remove on a directory succeeds only when it is empty, which is the
		// guard wanted here: a directory still holding something is left standing.
		_ = root.Remove(ledgerRel)
		return nil
	}
	if len(entries) > 0 {
		if err := underLedgerLock(root.Name(), unlink); err != nil {
			return nil, err
		}
	}
	_ = root.Remove(runRel + "/" + ManifestFileName)
	_ = root.Remove(runRel)
	sort.Strings(removed)
	return removed, nil
}

// underLedgerLock runs fn while holding core/capture's ledger lock, so a delete
// in the committed ledger cannot race that package's own readers.
//
// The lock file is capture's, and its name is capture's to state — this asks for
// it through the exported helper rather than restating the path, so the two
// cannot come to disagree about which file is the lock.
func underLedgerLock(repoRoot string, fn func() error) error {
	return capture.WithLedgerLock(repoRoot, fn)
}

// readDirIn lists a directory INSIDE the containment root, refusing a symlinked
// leaf before it is walked.
//
// The root already stops a link leaving the repository. This refuses one
// pointing at another directory INSIDE it, which is the stance core/capture
// takes on the same ledger directory (refuseSymlinkedDir) and the right one for
// a path a rollback deletes from: a run's directory is a run's directory, never
// a pointer at somebody else's.
func readDirIn(root *os.Root, rel string) ([]os.DirEntry, error) {
	fi, err := root.Lstat(rel)
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("reading: %s is a symlink; a directory this verb walks or removes from is "+
			"a real directory, never a pointer at another one", rel)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("reading: %s is not a directory", rel)
	}
	d, err := root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	return d.ReadDir(-1)
}

// writeJSONIn renders one artefact through the package's canonical encoder and
// writes it atomically INSIDE the containment root, so a reader never opens a
// half-written record and a symlinked ancestor cannot land it outside the
// repository. The missing parents are created through the root too.
func writeJSONIn(root *os.Root, rel string, v any) error {
	data, err := encode(v)
	if err != nil {
		return err
	}
	if err := fsutil.WriteFileAtomicInRoot(root, rel, data, 0o644); err != nil {
		return fmt.Errorf("reading: writing %s: %w", rel, err)
	}
	return nil
}

// recordIDs lists the ids of the records a ledger write returned.
func recordIDs(refs []capture.ReadingRecordRef) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.ID)
	}
	return out
}

// echo renders a payload-derived string for a terminal or a durable record. It
// is termsafe.CleanProseLine under this package's cap, NOT a fourth
// sanitise-and-cap: that package declares itself the canonical home for the
// untrusted-prose cleaner every host-delegated ingest boundary needs, and
// lifeboat, release and ideate already route through it. Using it here also
// picks up the HTML-opener neutralisation a hand-rolled mask would have missed.
//
// The line form is the right one: every value this wraps is interpolated into a
// one-line message or a single JSON field, so a newline in it would forge a line
// the reader did not get from this binary.
func echo(s string) string {
	return termsafe.CleanProseLine(s, maxEchoedBytes)
}

// echoAll is echo over a list of payload-chosen strings.
func echoAll(names []string) []string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		out = append(out, echo(n))
	}
	return out
}

// boundedNames caps how many payload-chosen names one refusal quotes. Each name
// is capped by echo; the LIST is not, and a payload carrying ten thousand
// unknown keys would otherwise put a megabyte of model-chosen text into a
// committed record.
func boundedNames(names []string) []string {
	if len(names) <= maxQuotedNames {
		return names
	}
	out := append([]string{}, names[:maxQuotedNames]...)
	return append(out, fmt.Sprintf("and %d more", len(names)-maxQuotedNames))
}

// sanitizeInstrument is echo applied to the one payload-supplied identity that
// lands in a durable record. No item body text ever does: bodies belong in the
// ledger records, which capture.IngestReading redacts on the way in.
func sanitizeInstrument(i Instrument) Instrument {
	return Instrument{
		Model:            echo(i.Model),
		DefinitionSHA256: echo(i.DefinitionSHA256),
		AssemblerVersion: echo(i.AssemblerVersion),
	}
}
