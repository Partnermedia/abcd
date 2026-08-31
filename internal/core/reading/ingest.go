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

// The three filenames under a run's durable directory.
const (
	RunFileName     = "run.json"
	RefusalFileName = "refusal.json"
	stageFileName   = "stage.json"
)

// PatternField is the envelope's provenance field: the pattern the reading read
// under. It is the reading RECORD's own field name (issueschema.ReadingRequired)
// and the key the four definitions instruct, so the payload, the record and the
// instruction all say one word. itd-185 and spc-63 call the same field the
// pattern named; there is one field, and this is its wire name.
const PatternField = "pattern"

// maxEchoedRunes caps a payload-derived string echoed into a message or a
// durable record. A refusal that quoted a four-megabyte field would be a denial
// of service against the reader of the refusal.
const maxEchoedRunes = 120

// readingItemFileRe is the grammar the orphan rollback removes by. A rollback
// deletes what an ingest wrote and nothing else, so it matches file names rather
// than clearing a directory.
var readingItemFileRe = regexp.MustCompile(`^` + issueschema.ReadingItemFamily + `-[0-9]+\.md$`)

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
type ItemRefusal struct {
	Ordinal int    `json:"ordinal"`
	Rule    string `json:"rule"`
	Field   string `json:"field,omitempty"`
	Detail  string `json:"detail"`
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
	ReviewFlags    []ReviewFlag               `json:"review_flags"`
}

// RefusalRecord is what a list-level refusal leaves behind: the run metadata and
// the named reason, and no items. The event is durable, and a rerun is a new run
// with a new run id, never an amendment.
type RefusalRecord struct {
	Type           string     `json:"_type"`
	SchemaVersion  int        `json:"schema_version"`
	RunID          string     `json:"run_id"`
	Position       Position   `json:"position"`
	Regime         string     `json:"regime"`
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
	RunID         string                     `json:"run_id"`
	Position      Position                   `json:"position"`
	Regime        string                     `json:"regime"`
	Records       []capture.ReadingRecordRef `json:"records"`
	RefusedItems  []ItemRefusal              `json:"refused_items,omitempty"`
	ReviewFlags   []ReviewFlag               `json:"review_flags,omitempty"`
	RunRecordPath string                     `json:"run_record,omitempty"`
	RefusalPath   string                     `json:"refusal_record,omitempty"`
	ClearedStages []string                   `json:"cleared_stages,omitempty"`
	Redacted      int                        `json:"redacted,omitempty"`
	Degraded      string                     `json:"redaction_degraded,omitempty"`
}

// The two points in the staged-write protocol a fault can be injected at.
const (
	faultAfterStage  = "after-stage"
	faultAfterLedger = "after-ledger"
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
// The order is the protocol, and it is the order for a reason: nothing durable
// is written until the whole payload validates, the ledger records land as one
// batch, and the run metadata lands last as the commit marker.
//
//  1. Sweep any orphaned stage a previous invocation left, naming and clearing it.
//  2. Read and decode the payload strictly.
//  3. Check the envelope, and resolve the parked run's manifest by content hash.
//     Until this passes, the run has no proven identity, so nothing — not even a
//     refusal — can be recorded against it.
//  4. Resolve the definition and check the regime and the instrument against it.
//     A refusal from here on writes a refusal record.
//  5. Validate every item; an item-level violation refuses that item and lands
//     the rest.
//  6. Stage, write the ledger records, promote the manifest, and write the run
//     metadata last.
func Ingest(req IngestRequest) (IngestResult, error) {
	res := IngestResult{}
	repoRoot := strings.TrimSpace(req.RepoRoot)
	if repoRoot == "" {
		return res, errors.New("reading: ingesting an output needs a repository root")
	}
	if strings.TrimSpace(req.OutputPath) == "" {
		return res, errors.New("reading: ingest needs the reading's output JSON")
	}

	cleared, err := sweepOrphanStages(repoRoot)
	res.ClearedStages = cleared
	if err != nil {
		return res, err
	}

	raw, err := readOutputFile(req.OutputPath)
	if err != nil {
		return res, err
	}
	out, err := decodeOutput(raw)
	if err != nil {
		return res, err
	}

	pos, err := checkEnvelope(out)
	if err != nil {
		return res, err
	}
	res.RunID = out.RunID
	res.Position = pos

	manifest, err := resolveParkedManifest(repoRoot, out)
	if err != nil {
		return res, err
	}

	// The run's identity is proven from here: the payload names a parked run and
	// cites its manifest by the manifest's own content hash. A refusal below is
	// therefore recordable against that run.
	def, err := LoadDefinition(repoRoot, pos)
	if err != nil {
		return res, err
	}
	res.Regime = def.Regime

	if err := checkRegime(out, def); err != nil {
		return refuse(repoRoot, &res, out, manifest, def, err)
	}
	if err := checkInstrument(out, def, manifest); err != nil {
		return refuse(repoRoot, &res, out, manifest, def, err)
	}

	items, refusals, flags, err := validateItems(out, def)
	res.RefusedItems = refusals
	res.ReviewFlags = flags
	if err != nil {
		return refuse(repoRoot, &res, out, manifest, def, err)
	}

	if err := write(repoRoot, &res, out, manifest, def, items); err != nil {
		return res, err
	}
	return res, nil
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
	if strings.TrimSpace(out.Instrument.Model) == "" ||
		strings.TrimSpace(out.Instrument.DefinitionSHA256) == "" ||
		strings.TrimSpace(out.Instrument.AssemblerVersion) == "" {
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
func resolveParkedManifest(repoRoot string, out Output) (Manifest, error) {
	// out.RunID has already been matched against the run-id grammar, which is
	// what makes this join safe: it holds no separator and no dot, so it cannot
	// escape the run directory.
	rel := DefaultRunDir + "/" + out.RunID + "/" + ManifestFileName
	raw, err := fsutil.ReadGuarded(filepath.Join(repoRoot, filepath.FromSlash(rel)), MaxFileBytes)
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
func write(repoRoot string, res *IngestResult, out Output, m Manifest, def Definition, items []capture.ReadingItem) error {
	stageDir := filepath.Join(repoRoot, filepath.FromSlash(IngestStageDir), out.RunID)
	marker := stageMarker{Type: StageType, RunID: out.RunID, Records: []string{}}
	if err := writeJSON(filepath.Join(stageDir, stageFileName), marker); err != nil {
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

	// The marker now names what landed, so a rollback removes exactly the files
	// this ingest wrote rather than clearing a directory it did not fill.
	marker.Records = recordIDs(written.Records)
	if err := writeJSON(filepath.Join(stageDir, stageFileName), marker); err != nil {
		return err
	}
	if err := fireFault(faultAfterLedger); err != nil {
		return err
	}

	runDir := filepath.Join(repoRoot, filepath.FromSlash(ReadingsRecordDir), out.RunID)
	if err := writeJSON(filepath.Join(runDir, ManifestFileName), m); err != nil {
		return err
	}
	run := RunRecord{
		Type: RunType, SchemaVersion: SchemaVersion, RunID: out.RunID,
		Position: def.Position, Regime: def.Regime, TargetCommit: m.TargetCommit,
		ManifestSHA256: out.ManifestSHA256, Instrument: sanitizeInstrument(out.Instrument),
		Records: written.Records, RefusedItems: res.RefusedItems, ReviewFlags: res.ReviewFlags,
	}
	if run.RefusedItems == nil {
		run.RefusedItems = []ItemRefusal{}
	}
	if run.ReviewFlags == nil {
		run.ReviewFlags = []ReviewFlag{}
	}
	if err := writeJSON(filepath.Join(runDir, RunFileName), run); err != nil {
		return err
	}
	res.RunRecordPath = ReadingsRecordDir + "/" + out.RunID + "/" + RunFileName

	// The commit marker is down. The stage has nothing left to be evidence of.
	return os.RemoveAll(stageDir)
}

// refuse records a list-level refusal and returns it.
//
// The record is durable because the event is: a refused run is a run that
// happened, and a rerun is a NEW run with a new run id, never an amendment. It
// carries the run metadata and the named reason and no items, and nothing was
// ever moved out of the stage — so there are no reading records to leave behind.
func refuse(repoRoot string, res *IngestResult, out Output, m Manifest, def Definition, cause error) (IngestResult, error) {
	rec := RefusalRecord{
		Type: RefusalType, SchemaVersion: SchemaVersion, RunID: out.RunID,
		Position: def.Position, Regime: def.Regime, TargetCommit: m.TargetCommit,
		ManifestSHA256: out.ManifestSHA256, Instrument: sanitizeInstrument(out.Instrument),
		Reason: echo(cause.Error()),
	}
	rel := ReadingsRecordDir + "/" + out.RunID + "/" + RefusalFileName
	if err := writeJSON(filepath.Join(repoRoot, filepath.FromSlash(rel)), rec); err != nil {
		return *res, fmt.Errorf("reading: %w (and the refusal record could not be written: %v)", cause, err)
	}
	res.RefusalPath = rel
	return *res, fmt.Errorf("reading: %w; the refusal is recorded at %s", cause, rel)
}

// sweepOrphanStages reports and clears every stage a previous invocation left.
//
// An orphan means an ingest reached the stage and never reached its commit
// marker, so the run never happened — and a run that never happened must leave
// no reading records behind. The rollback therefore removes what the stage says
// the ingest wrote, then the run's own directory when the commit marker is
// absent. A stage whose run DID commit is a leftover and only the stage goes.
func sweepOrphanStages(repoRoot string) ([]string, error) {
	root := filepath.Join(repoRoot, filepath.FromSlash(IngestStageDir))
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading: listing the ingest stage: %w", err)
	}
	var cleared []string
	for _, e := range entries {
		// Only a well-formed run id is swept. A directory named anything else was
		// not written by this verb, and a sweep that removed it would be deleting
		// somebody else's file on a guess.
		if !e.IsDir() || !recordid.ValidReadingRunID(e.Name()) {
			continue
		}
		if err := rollbackRun(repoRoot, e.Name()); err != nil {
			return cleared, err
		}
		if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
			return cleared, fmt.Errorf("reading: clearing the orphaned stage of run %s: %w", e.Name(), err)
		}
		cleared = append(cleared, e.Name())
	}
	sort.Strings(cleared)
	return cleared, nil
}

// rollbackRun removes the durable half of a run whose commit marker never
// landed. A run WITH a commit marker is left entirely alone: its stage is a
// leftover from a crash after the marker, and the run is complete.
func rollbackRun(repoRoot, runID string) error {
	runDir := filepath.Join(repoRoot, filepath.FromSlash(ReadingsRecordDir), runID)
	if _, err := os.Stat(filepath.Join(runDir, RunFileName)); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("reading: probing the commit marker of run %s: %w", runID, err)
	}

	ledgerDir := filepath.Join(repoRoot, filepath.FromSlash(capture.LedgerRelPath),
		issueschema.ReadingsDir, runID)
	entries, err := os.ReadDir(ledgerDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading: listing the records of orphaned run %s: %w", runID, err)
	}
	for _, e := range entries {
		// Bounded by the item-id grammar: a rollback removes reading records and
		// nothing else, so a file a person put in the directory survives it.
		if e.IsDir() || !readingItemFileRe.MatchString(e.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(ledgerDir, e.Name())); err != nil {
			return fmt.Errorf("reading: rolling back record %s of run %s: %w", e.Name(), runID, err)
		}
	}
	// os.Remove on a directory succeeds only when it is empty, which is the
	// guard wanted here: a directory still holding something is left standing.
	_ = os.Remove(ledgerDir)
	_ = os.Remove(filepath.Join(runDir, ManifestFileName))
	_ = os.Remove(runDir)
	return nil
}

// writeJSON renders one artefact through the package's canonical encoder and
// writes it atomically, so a reader never opens a half-written record.
func writeJSON(path string, v any) error {
	data, err := encode(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("reading: %w", err)
	}
	if err := fsutil.WriteFileAtomic(path, data, 0o644); err != nil {
		return fmt.Errorf("reading: writing %s: %w", filepath.Base(path), err)
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

// echo renders a payload-derived string for a terminal or a durable record:
// terminal-display attack runes masked, and the length capped. Every message in
// this file that quotes the payload goes through it.
func echo(s string) string {
	s = termsafe.Sanitize(s)
	r := []rune(s)
	if len(r) <= maxEchoedRunes {
		return s
	}
	return string(r[:maxEchoedRunes]) + "..."
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
