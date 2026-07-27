package lint

// The committed citation baseline: the record the zero-network gate enforces.
//
// The split this file sits on is the whole point of the citation gate. Fetching
// a URL is non-deterministic and needs the network; deciding whether a commit is
// allowed must be neither. So the refresh verb does the live checking ONCE and
// writes what it learned here, and the lint reads only this file. A commit is
// judged against a committed artefact a reviewer can read in a diff, never
// against whatever the network happened to answer that minute.
//
// The schema is deliberately, structurally incapable of recording HOW a link was
// verified. A human who confirms a paywalled source records THAT they confirmed
// it and WHEN — never the browser, the credential, the workaround, or the
// transcript. That is honest labelling without turning the record into a log of
// a maintainer's private working. The refusal is enforced twice: BaselineEntry
// declares no such field, and the decoder runs with DisallowUnknownFields so a
// hand-added one is a load-time refusal rather than a silently-dropped key.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// BaselineSchemaVersion is the only baseline schema this build understands. A
// document declaring anything else is refused rather than best-effort parsed:
// the gate must never enforce a record whose meaning it is guessing at.
const BaselineSchemaVersion = 1

// DefaultBaselinePath is the committed baseline's repo-relative home. It sits
// beside the other committed machine records under .abcd/ (docs-lint.json,
// record-lint.json, rules.json) rather than inside the development record,
// because it is configuration the gate reads on every commit, not design
// material.
const DefaultBaselinePath = ".abcd/citations-baseline.json"

// The outcomes a checked citation can have. Two values, not a free string: the
// gate branches on this, so an unrecognised outcome is a refusal at load.
const (
	// OutcomeAlive means the URL resolved to a live document.
	OutcomeAlive = "alive"
	// OutcomeBroken means it did not, and the gate must block on it.
	OutcomeBroken = "broken"
)

// The verification labels. This is the "automatic or manual" axis of AC 2 — WHO
// established the outcome, never how.
const (
	// VerificationAutomatic means the refresh verb's fetcher established it.
	VerificationAutomatic = "automatic"
	// VerificationManual means a human confirmed it after the fetcher could not.
	VerificationManual = "manual"
)

// baselineDateLayout is the date format every timestamp in the baseline uses.
// Date precision, not instant precision: the staleness policy is measured in
// days (180/365), and a wall-clock instant would add churn to a committed file
// for no gain in the only arithmetic that reads it.
const baselineDateLayout = "2006-01-02"

// baselineSizeLimit caps the baseline read. It is a committed file, but the
// guarded read is what keeps a hostile branch from handing the gate a decompression
// bomb in place of a record.
const baselineSizeLimit = 8 << 20 // 8 MiB

// Baseline is the committed citation record: one entry per cited URL.
type Baseline struct {
	// SchemaVersion pins the document's shape. Required.
	SchemaVersion int `json:"schema_version"`
	// Entries is keyed by the cited URL exactly as it appears in the docs.
	Entries map[string]BaselineEntry `json:"entries"`
}

// BaselineEntry is one URL's record.
//
// Every field here answers "what is true about this link", and none answers
// "how did we find out". That omission is the schema's contract (AC 2), pinned
// by TestBaselineHasNoMethodFieldByConstruction.
type BaselineEntry struct {
	// URL optionally restates the map key. It is redundant by design — a
	// committed record reads better in a diff when each entry names itself — and
	// is validated to AGREE with its key rather than being a second source of
	// truth. Empty is fine; disagreeing is a refusal.
	URL string `json:"url,omitempty"`
	// FinalURL is the address the citation actually resolved to after redirects.
	// When it differs from the cited URL, the docs are citing a stale address and
	// the gate says so.
	FinalURL string `json:"final_url"`
	// LastChecked is when the outcome was established. This is THE staleness
	// clock — the same clock for automatic and manual entries alike, so a human
	// verification buys no exemption from ageing (AC 3).
	LastChecked string `json:"last_checked"`
	// Outcome is OutcomeAlive or OutcomeBroken.
	Outcome string `json:"outcome"`
	// Verification is VerificationAutomatic or VerificationManual — who
	// established the outcome.
	Verification string `json:"verification"`
	// VerifiedOn is the date that verification was recorded.
	VerifiedOn string `json:"verified_on"`
}

// BaselineError is a malformed-baseline refusal. It is a distinct type so a
// caller can tell a bad record from an I/O fault without string matching.
type BaselineError struct{ msg string }

func (e *BaselineError) Error() string { return e.msg }

// Checked parses the entry's staleness clock.
func (e BaselineEntry) Checked() (time.Time, error) {
	return time.Parse(baselineDateLayout, e.LastChecked)
}

// LoadBaseline reads, decodes, and validates a committed baseline.
//
// A missing file is returned as the underlying os.ErrNotExist, so the caller can
// distinguish "this repo has no baseline yet" (a state the gate reports as one
// finding) from "this repo has a corrupt baseline" (a refusal).
func LoadBaseline(path string) (Baseline, error) {
	data, err := fsutil.ReadGuarded(path, baselineSizeLimit)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Baseline{}, err
		}
		if errors.Is(err, fsutil.ErrNotRegular) || errors.Is(err, fsutil.ErrTooBig) {
			return Baseline{}, &BaselineError{"citation baseline " + path + ": " + err.Error()}
		}
		return Baseline{}, err
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	// The AC 2 teeth: a smuggled method/how/transcript key is a refusal, not a
	// silently-ignored field. Unknown keys anywhere fail the load.
	dec.DisallowUnknownFields()
	var b Baseline
	if err := dec.Decode(&b); err != nil {
		return Baseline{}, &BaselineError{"citation baseline is not valid: " + err.Error()}
	}
	if err := b.validate(); err != nil {
		return Baseline{}, err
	}
	return b, nil
}

// validate enforces every invariant the gate then relies on, so no check
// downstream has to re-ask whether a field is well-formed.
func (b Baseline) validate() error {
	if b.SchemaVersion != BaselineSchemaVersion {
		return &BaselineError{
			"citation baseline schema_version is " + strconv.Itoa(b.SchemaVersion) +
				"; this build understands only " + strconv.Itoa(BaselineSchemaVersion)}
	}
	// Sorted so a document with several faults always reports the same one, and
	// a message never depends on Go's map iteration order.
	keys := make([]string, 0, len(b.Entries))
	for k := range b.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		e := b.Entries[k]
		if strings.TrimSpace(k) == "" {
			return &BaselineError{"citation baseline has an empty url key"}
		}
		who := "citation baseline entry " + k
		// Declared-key tolerance: restating the key is fine, disagreeing is not.
		if e.URL != "" && e.URL != k {
			return &BaselineError{who + " declares url " + e.URL + ", which disagrees with its key"}
		}
		if strings.TrimSpace(e.FinalURL) == "" {
			return &BaselineError{who + " has no final_url; the resolved address is required"}
		}
		if _, err := time.Parse(baselineDateLayout, e.LastChecked); err != nil {
			return &BaselineError{who + " has an unparsable last_checked " + quote(e.LastChecked) +
				"; expected " + baselineDateLayout}
		}
		switch e.Outcome {
		case OutcomeAlive, OutcomeBroken:
		default:
			return &BaselineError{who + " has outcome " + quote(e.Outcome) +
				"; expected " + OutcomeAlive + " or " + OutcomeBroken}
		}
		switch e.Verification {
		case VerificationAutomatic, VerificationManual:
		default:
			return &BaselineError{who + " has verification " + quote(e.Verification) +
				"; expected " + VerificationAutomatic + " or " + VerificationManual}
		}
		if _, err := time.Parse(baselineDateLayout, e.VerifiedOn); err != nil {
			return &BaselineError{who + " has an unparsable verified_on " + quote(e.VerifiedOn) +
				"; expected " + baselineDateLayout}
		}
	}
	return nil
}

// SaveBaseline validates and writes a baseline atomically. Validation runs on
// the way OUT as well as in, so a producer bug can never commit a record the
// gate would then refuse to load.
func SaveBaseline(path string, b Baseline) error {
	if b.SchemaVersion == 0 {
		b.SchemaVersion = BaselineSchemaVersion
	}
	if err := b.validate(); err != nil {
		return err
	}
	// Indented and newline-terminated: this file is committed and reviewed in
	// diffs, so it is written to be read. Go's encoder sorts map keys, which is
	// what keeps the diff stable across runs.
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(path, append(data, '\n'), 0o644)
}

// baselineEntryJSONTags lists BaselineEntry's json field names. It exists for
// the construction test that asserts no how-shaped field can creep in.
func baselineEntryJSONTags() []string {
	t := reflect.TypeOf(BaselineEntry{})
	tags := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		tags = append(tags, name)
	}
	return tags
}

// quote renders a value inside the message without pulling in fmt for one verb.
func quote(s string) string { return `"` + s + `"` }
