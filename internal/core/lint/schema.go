package lint

// The record-schema family (record_schema): the mechanical shape of the record's
// identified stores — ADRs, intents, specs, issues.
//
// Every other record rule asks a question about ONE store (is this intent's
// bucket schema right, does this spec agree with its intent). None of them asks
// the questions that only make sense ACROSS the stores, which is where the record
// actually drifts (iss-39):
//
//   - a cross-reference names a record that is not in the corpus;
//   - a supersession is declared from one side only, so the record contradicts
//     itself about which decision is in force;
//   - a filename and the id inside it disagree, so the same record answers to two
//     handles;
//   - a lifecycle directory nobody declared holds records no rule ever reads.
//
// The last one is why this rule enumerates the buckets rather than allowlisting
// them per store: an undeclared directory is not "clean", it is a lifecycle state
// that silently escapes every other rule.
//
// It is a STRUCTURAL rule, so it never consults contentExempt — the historical
// part of the record is excused from content-authoring rules, not from being
// well-formed (iss-39; a superseded intent that carries no supersession is the
// exact defect the broad exemption used to hide).

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/core/recordid"
)

const ruleRecordSchema = "record_schema"

var (
	// A record handle as it is written in a cross-reference field: adr-6, ADR-6,
	// itd-47, iss-39, spc-5. The number is compared numerically, so the zero-padded
	// and bare spellings of one id are the same handle. The alternation covers
	// every store the rule INDEXES — a prefix indexed but not matched here reads as
	// "no handle at all", which turns a well-formed link into a false blocker and
	// leaves its reverse direction unchecked.
	recordHandleRe = regexp.MustCompile(`(?i)\b(adr|itd|iss|spc)-(\d+)\b`)
	// The same handle, anchored: a whole frontmatter id value and nothing else.
	recordHandleFullRe = regexp.MustCompile(`(?i)^(adr|itd|iss|spc)-(\d+)$`)
	// A frontmatter id of ANY store, parsed by shape rather than against the
	// cross-reference vocabulary above. The two are deliberately different sets:
	// the vocabulary says which prefixes may appear in a cross-reference FIELD,
	// while this says what an id looks like at all. The filename ↔ id agreement is
	// a question every store has to answer, including the ones whose records are
	// never cited — and checking it against the citation vocabulary would report a
	// perfectly good rdi-N id as disagreeing with itself.
	anyHandleFullRe = regexp.MustCompile(`(?i)^([a-z]+)-(\d+)$`)
	// recordHandleKinds renders the legal prefixes for a message, composed from the
	// stores rather than spelled as a literal so it cannot drift from the pattern.
	recordHandleKinds = "adr-N, itd-N, spc-N, or iss-N"
	// Filename → id number for each store. An ADR filename is the zero-padded
	// sequential form (0012-<slug>.md); the other three carry the prose handle and
	// borrow the recordid resolver's OWN grammar, not a looser local copy.
	//
	// The prose-handle stores match recordid.FilenameNumRe exactly, so lint reads a
	// filename as a record iff the resolver does: the old `^iss-(\d+).*\.md$` and
	// its siblings accepted an arbitrary tail (`iss-5_bad.md`), which capture's
	// scanLedger silently dropped and the resolver hard-errored on when the record
	// was cited — a record the gate passed but no consumer could read
	// (iss-2608270908346617). The ADR store keeps its own pattern (it has no
	// prose-handle prefix) but is tightened to a kebab slug tail for the same
	// reason: no arbitrary bytes after the ordinal.
	adrFileNumRe    = regexp.MustCompile(`^([0-9]+)-[a-z0-9]+(?:-[a-z0-9]+)*\.md$`)
	intentFileNumRe = recordid.FilenameNumRe("itd")
	specFileNumRe   = recordid.FilenameNumRe("spc")
	issueFileNumRe  = recordid.FilenameNumRe("iss")
	// The reading families' filename and bucket grammars. A run directory is
	// named for the run that minted it; a disposition directory is named for the
	// ITEM it answers, which is what makes the status signal one directory probe
	// rather than a folder-membership question.
	readingItemFileNumRe = recordid.FilenameNumRe(issueschema.ReadingItemFamily)
	readingRunFileNumRe  = recordid.FilenameNumRe(issueschema.ReadingRunFamily)
	dispositionFileNumRe = recordid.FilenameNumRe(issueschema.DispositionFamily)
	readingRunBucketRe   = regexp.MustCompile(`^` + issueschema.ReadingRunFamily + `-[0-9]+$`)
	dispositionBucketRe  = regexp.MustCompile(`^` + issueschema.ReadingItemFamily + `-[0-9]+$`)
	// The cross-reference frontmatter fields whose targets must resolve. They are
	// the record's machine-readable claims that another record exists and is a
	// live input — as distinct from prose, where naming a released or retired id
	// ("itd-38's id was released, not reserved") is legitimate narration.
	//
	// `supersedes` is deliberately NOT here. It is the field that DECLARES a
	// pruned id, so it can never fail its own resolution check; listing it would
	// read as coverage it does not provide. Its targets are checked separately,
	// against the store's allocation high-water mark.
	recordRefFields = []string{"related_adrs", "related_intents", "builds_on", "blocked_by"}
	// Every field the rule reads handles out of, so the scan parses each once.
	recordHandleFields = append([]string{"supersedes", "superseded_by"}, recordRefFields...)
	// recordGraphFields are cross-reference fields the record carries that this
	// rule does not judge — the intent↔spec pair, whose resolution is the spec
	// store's own rule — but which the exported record graph (graph.go) must
	// carry, because they are half the record's typed links. They are parsed by
	// the same scan so the graph never needs a second parser: reading them here
	// costs one regexp pass per record and adds no finding.
	recordGraphFields = []string{"spec_id", "intent"}
	// recordParsedFields is every field the scan reads handles out of, in a fixed
	// order so the graph export is deterministic.
	recordParsedFields = append(append([]string{}, recordHandleFields...), recordGraphFields...)
)

// recordStore describes one identified record store: where the buckets are, and
// how a well-formed filename spells the id.
type recordStore struct {
	// prefix is the id prefix (adr, itd, spc, iss) and the record_stores key.
	prefix string
	// noun names the record kind in a finding message.
	noun string
	// nodeType is the store's name in the exported record graph (graph.go). It
	// is the store's identity spelled for a consumer rather than for a message,
	// and it lives here so the graph never re-declares which prefixes exist.
	nodeType string
	// buckets are the lifecycle directories the store declares. A store with
	// neither buckets nor bucketRe is FLAT (the ADR store): records sit directly
	// in it.
	buckets []string
	// bucketRe declares the store's buckets by GRAMMAR instead of by list, for a
	// store whose buckets are MINTED rather than enumerated — a reading run's
	// directory and an item-keyed disposition directory are both that shape.
	// Nobody can list them ahead of time, so a store that could not declare a
	// grammar would have to leave its whole tree undeclared, which is exactly the
	// escape the undeclared-bucket check exists to close.
	bucketRe *regexp.Regexp
	// fileNumRe extracts the id number from a filename (submatch 1).
	fileNumRe *regexp.Regexp
	// fileFamily is the family prefix a filename in this store carries, spelled
	// without its hyphen — the argument recordid.SplitRecordFilename takes. It is
	// NOT always the store's prefix: an ADR filename is the bare zero-padded
	// number (0022-<slug>.md), so its family is empty. Held as data rather than
	// derived from prefix with a special case, because a store that spells its
	// filenames differently is a property of the store.
	fileFamily string
	// filename describes the convention a finding message quotes.
	filename string
	// requiredFields are the frontmatter properties every record in this store
	// must carry. A nil requiredFields means the store declares none HERE and is
	// left alone — the four stores have different schemas, and the ones whose
	// frontmatter other rules already judge (intent_lifecycle, spec_id_unique)
	// must not be re-judged against the issue's shape.
	requiredFields []string
}

// bucketed reports whether the store holds its records in lifecycle
// directories, by list or by grammar, rather than directly in its root.
func (s recordStore) bucketed() bool { return s.buckets != nil || s.bucketRe != nil }

// declaresBucket reports whether name is one of the store's buckets.
func (s recordStore) declaresBucket(name string) bool {
	for _, b := range s.buckets {
		if b == name {
			return true
		}
	}
	return s.bucketRe != nil && s.bucketRe.MatchString(name)
}

// bucketDesc renders the store's declared buckets for a finding message: the
// list where there is one, the grammar where the buckets are minted.
func (s recordStore) bucketDesc() string {
	if len(s.buckets) > 0 {
		return strings.Join(s.buckets, ", ")
	}
	if s.bucketRe != nil {
		return "names matching " + s.bucketRe.String()
	}
	return ""
}

// recordStores is the closed set of identified record stores. It is code, not
// config: which lifecycle states exist is the record's schema, and a config that
// could add a bucket could also hide one.
var recordStores = []recordStore{
	// `id` is required for an ADR because the record dispatcher (record.describeADR)
	// routes by the filename ordinal but CONFIRMS the frontmatter id before it will
	// render the record — so an id-less ADR reads as "not found" though its file
	// plainly sits in the store, while the lint that never asked for the id stayed
	// green (iss-2608270908344426). This is parity with the prose-handle stores,
	// whose loaders (intent.Load) fail closed on a missing id: the id is a required
	// property, and its absence must be a finding, not silent invisibility.
	{prefix: "adr", noun: "ADR", nodeType: "adr", buckets: nil, fileNumRe: adrFileNumRe, fileFamily: "", filename: "<NNNN>-<slug>.md",
		requiredFields: []string{"id"}},
	{prefix: "itd", noun: "intent", nodeType: "intent", buckets: intentBucketNames, fileNumRe: intentFileNumRe, fileFamily: "itd", filename: "itd-<N>-<slug>.md"},
	{prefix: "spc", noun: "spec", nodeType: "spec", buckets: specBucketNames, fileNumRe: specFileNumRe, fileFamily: "spc", filename: "spc-<N>-<slug>.md"},
	// The issue store's required properties come from the schema's ONE definition
	// (core/issueschema), the same list the ledger reader validates against — a
	// hand-copied list here would drift the moment the schema gains a field, and
	// the drift would show up as a silently unread record, which is the defect
	// this invariant exists to catch.
	{prefix: "iss", noun: "issue", nodeType: "issue", buckets: issueStatusDirs, fileNumRe: issueFileNumRe, fileFamily: "iss", filename: "iss-<N>-<slug>.md",
		requiredFields: issueschema.Required},
	// The three reading families (spc-58). Each buckets by GRAMMAR because its
	// buckets are minted: a reading item and a run record live under the run that
	// produced them, and a disposition lives under the item it answers.
	//
	// Their required-field sets are absent, and that is a STATED GAP rather than a
	// delegation. What this rule gives them is structural — the bucket grammar,
	// the filename ↔ id agreement, no undeclared lifecycle directory — and their
	// CONTENT is judged by the writer that refuses a malformed record at the
	// boundary, and by review. So the guarantee the issue store has, that a record
	// the reader would refuse is not lint-green, does not yet hold here: a record
	// hand-written into these trees can carry a body no reader reads and pass this
	// gate. Declaring their required fields is what closes it, and until that
	// lands the gap belongs in writing rather than in the difference between two
	// store entries.
	{prefix: "rdi", noun: "reading item", nodeType: "reading", bucketRe: readingRunBucketRe,
		fileNumRe: readingItemFileNumRe, fileFamily: "rdi", filename: "rdi-<N>.md"},
	{prefix: "rdg", noun: "reading run", nodeType: "reading-run", bucketRe: readingRunBucketRe,
		fileNumRe: readingRunFileNumRe, fileFamily: "rdg", filename: "rdg-<N>.md"},
	{prefix: "dsp", noun: "disposition", nodeType: "disposition", bucketRe: dispositionBucketRe,
		fileNumRe: dispositionFileNumRe, fileFamily: "dsp", filename: "dsp-<N>.md"},
}

// recordStorePrefixes is the set of store prefixes the scanner knows, derived
// from recordStores so the configuration validator and the walk cannot disagree
// about what a store is.
func recordStorePrefixes() map[string]bool {
	out := make(map[string]bool, len(recordStores))
	for _, s := range recordStores {
		out[s.prefix] = true
	}
	return out
}

// schemaRecord is one record file as the schema rule sees it: which store and
// bucket hold it, the id number its FILENAME claims, and its frontmatter.
type schemaRecord struct {
	rel    string
	store  recordStore
	num    int
	bucket string
	// title is the record's H1, or — for a store whose records carry none, the
	// issue ledger — its first body line. The schema rule never reads it; it is
	// read here because the scan already holds the file's lines, and the record
	// graph (graph.go) would otherwise have to open every record a second time.
	title  string
	fields map[string]fmField
	// refs holds the handles read out of each cross-reference field, parsed once
	// at scan time so the block-sequence spelling is read in ONE place. Reading
	// only fields[...].value would miss it: the shared frontmatter scanner is a
	// same-line scanner, so `supersedes:` followed by indented `- adr-12` lines
	// reads as empty — and an empty read here would make the bidirectional check
	// assert, confidently and falsely, that another file omits a link it carries.
	refs map[string][]recordRef
}

// handle renders the record's prose handle (adr-12, itd-47).
func (r schemaRecord) handle() string {
	return r.store.prefix + "-" + strconv.Itoa(r.num)
}

// recordRef is one handle read out of a cross-reference field.
type recordRef struct {
	prefix string
	num    int
}

func (h recordRef) String() string { return h.prefix + "-" + strconv.Itoa(h.num) }

// checkRecordSchema implements the record_schema family. It reads each configured
// store once and asserts five invariants across them: bucket coverage, filename ↔
// id agreement, filename ↔ slug agreement, cross-reference resolution, and
// bidirectional supersession. A
// store that is not configured, or whose directory is absent, contributes nothing
// and is not an error — an unpopulated repository is a state, not a fault.
func checkRecordSchema(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	records, out, err := scanRecordStores(repoRoot, cfg)
	if err != nil {
		return nil, err
	}

	// index: what the corpus HAS. highWater: the highest id each store has ever
	// issued, as far as the corpus can show.
	//
	// records is sorted by path, so on an ordinal collision the FIRST record wins
	// the index slot and every later claimant is reported: two records sharing a
	// (prefix, ordinal) handle are one handle that resolves to only the first, so
	// the second is unreachable to every cross-reference and index that keys on it —
	// silently, before this check, because the map assignment just overwrote it
	// (iss-2608270908346940). For the prose-handle stores the id-unique rules
	// (issue_id_unique, intent_lifecycle, spec_id_unique) catch the frontmatter-id
	// collision; the ADR store has no such rule, so this is its only guard.
	index := map[recordRef]schemaRecord{}
	highWater := map[string]int{}
	for _, r := range records {
		ref := recordRef{r.store.prefix, r.num}
		if first, dup := index[ref]; dup {
			out = append(out, Finding{
				File: r.rel, Line: 1, RuleID: ruleRecordSchema, Severity: cfg.Severity,
				Message: "filename ordinal '" + r.handle() + "' collides with " + first.rel +
					"; two " + r.store.noun + " records sharing an id are one handle that resolves to only the first, so this one is unreachable to every cross-reference and index that keys on it",
			})
		} else {
			index[ref] = r
		}
		if r.num > highWater[r.store.prefix] {
			highWater[r.store.prefix] = r.num
		}
	}

	// retired: ids a record declares it replaced. The ADR lifecycle prunes a
	// superseded record once its successor lands and keeps the trace in the
	// successor's `supersedes` (decisions/adrs/README.md), so a handle naming a
	// pruned id resolves to that declaration rather than to a file — the record
	// has not lost track of it.
	//
	// The declaration is BOUNDED by the store's allocation high-water mark,
	// because it is otherwise self-attesting: `supersedes: [adr-9999]` would mint
	// a phantom the rule then accepts in every other record's related_adrs,
	// silently reopening the very class of reference this rule exists to close. An
	// id above the high-water mark was never issued, so nothing can have pruned
	// it, and the declaration itself is reported below.
	retired := map[recordRef]bool{}
	for _, r := range records {
		for _, h := range r.refs["supersedes"] {
			if h.num >= 1 && h.num <= highWater[h.prefix] {
				retired[h] = true
			}
		}
	}

	add := func(rel string, line int, msg string) {
		if line == 0 {
			line = 1
		}
		out = append(out, Finding{
			File: rel, Line: line, RuleID: ruleRecordSchema, Severity: cfg.Severity, Message: msg,
		})
	}

	for _, r := range records {
		out = append(out, checkRecordFilename(r, cfg.Severity)...)
		out = append(out, checkRecordFilenameSlug(r, cfg.Severity)...)
		out = append(out, checkRecordRequiredFields(r, cfg.Severity)...)
		out = append(out, checkIssueRecordShape(r, cfg.Severity)...)

		// Cross-references: a named record must be in the corpus, or declared
		// retired by the record that replaced it.
		for _, field := range recordRefFields {
			f := r.fields[field]
			for _, h := range r.refs[field] {
				if _, ok := index[h]; ok || retired[h] {
					continue
				}
				add(r.rel, f.line, field+" names '"+h.String()+"', which is not a record in the corpus and no record declares it superseded; a cross-reference is a claim that the record exists")
			}
		}

		// A pruned id must be one the store actually issued. Without this, the
		// retirement declaration is a blank cheque: any id written here becomes
		// resolvable everywhere else in the corpus.
		sup := r.fields["supersedes"]
		for _, h := range r.refs["supersedes"] {
			if _, ok := index[h]; ok {
				continue
			}
			if h.num >= 1 && h.num <= highWater[h.prefix] {
				continue // a plausible pruned id: below the store's high-water mark
			}
			add(r.rel, sup.line, "supersedes names '"+h.String()+"', which is neither a record in the corpus nor an id the "+
				h.prefix+" store has issued (its highest is "+h.prefix+"-"+strconv.Itoa(highWater[h.prefix])+
				"); a pruned record's id must be one that was allocated")
		}

		// Supersession, direction A→B: the successor must exist (a live decision is
		// never a pruned one) and must name this record back.
		sb := r.fields["superseded_by"]
		targets := r.refs["superseded_by"]
		if len(targets) == 0 {
			if !isAbsentValue(sb.value) {
				add(r.rel, sb.line, "superseded_by '"+sb.value+"' is not a record handle (want "+recordHandleKinds+")")
			}
			continue
		}
		for _, h := range targets {
			target, ok := index[h]
			if !ok {
				add(r.rel, sb.line, "superseded_by names '"+h.String()+"', which is not a record in the corpus; a successor decision must be present")
				continue
			}
			if !refsContain(target.refs["supersedes"], recordRef{r.store.prefix, r.num}) {
				add(r.rel, sb.line, "one-way supersession: this record declares superseded_by '"+h.String()+
					"' but "+target.rel+" does not list '"+r.handle()+"' in supersedes; both directions must be present")
			}
		}
	}

	// Supersession, direction B→A: a record that claims to replace another must be
	// named by it. A target that is not in the corpus was pruned with the record's
	// blessing (see above) and has nothing left to answer with.
	for _, r := range records {
		sup := r.fields["supersedes"]
		for _, h := range r.refs["supersedes"] {
			target, ok := index[h]
			if !ok {
				continue
			}
			if !refsContain(target.refs["superseded_by"], recordRef{r.store.prefix, r.num}) {
				add(r.rel, sup.line, "one-way supersession: this record declares supersedes '"+h.String()+
					"' but "+target.rel+" does not carry 'superseded_by: "+r.handle()+"'; both directions must be present")
			}
		}
	}

	return out, nil
}

// checkRecordFilename asserts that a record's frontmatter id agrees with the id
// its filename claims. The filename is the handle every cross-reference and every
// index row resolves through, so a disagreement means one record answers to two
// ids — and the rules that key on the filename and the rules that key on the
// field then lint two different records. An absent id is a different (and larger)
// schema question than this rule's, so only a present one is compared.
func checkRecordFilename(r schemaRecord, severity string) []Finding {
	f := r.fields["id"]
	if isAbsentValue(f.value) {
		return nil
	}
	want := r.handle()
	got := strings.Trim(strings.TrimSpace(f.value), `"'`)
	// Compared as a PARSED handle, not as a string: `adr-0012` and `adr-12` are one
	// id written two ways (the rest of the rule already compares numerically), and
	// a string comparison would report the record's own zero-padded spelling as a
	// disagreement with itself.
	if m := anyHandleFullRe.FindStringSubmatch(got); m != nil {
		if n, err := strconv.Atoi(m[2]); err == nil &&
			strings.EqualFold(m[1], r.store.prefix) && n == r.num {
			return nil
		}
	}
	line := f.line
	if line == 0 {
		line = 1
	}
	return []Finding{{
		File: r.rel, Line: line, RuleID: ruleRecordSchema, Severity: severity,
		Message: "filename claims id '" + want + "' but frontmatter declares '" + got +
			"'; a " + r.noun() + " filename is " + r.store.filename,
	}}
}

// checkRecordFilenameSlug asserts that a record's frontmatter slug agrees with
// the slug its filename carries — the other half of the agreement
// checkRecordFilename pins for the id. Both halves are one value the store wrote
// twice, and a record renamed by hand keeps a stale handle in its name while its
// field says something else: readers that locate a record by filename and readers
// that trust the field then describe two different records, and neither says so.
//
// Until this landed the question was asked only by the LEDGER READER
// (core/capture's validateInvariants), which means a drifted record passed
// record-lint and every CI gate and then dropped to a Skipped line at read time —
// invisible to `capture list`, `capture status` and every other surface. Turning
// that silent skip into a red gate is why the check belongs in the structural
// rule as well as in the reader, exactly as checkRecordRequiredFields does.
//
// The filename is split by recordid.SplitRecordFilename, the SAME call the reader
// makes, so the gate and the reader cannot reach different verdicts on one record.
// The comparison is exact, with no prefix tolerance: every store applies its slug
// length cap while DERIVING the slug, before that one value forks into the
// filename and the frontmatter, so a filename is never a truncated form of a
// longer field and tolerating a prefix would license the drift this catches.
//
// Two cases are silent here, for different reasons. A record carrying no slug is
// a different (and larger) schema question, owned by checkRecordRequiredFields
// for the stores that declare the property — absence is not disagreement.
//
// A filename this splitter cannot read is silent too, and that one is a KNOWN
// GAP, not a delegation. The store's own filename rule in scanRecordStores does
// not cover it: that rule matches store.fileNumRe, whose issue pattern
// (`^iss-(\d+).*\.md$`) accepts an ARBITRARY tail, so a non-kebab name like
// iss-2-another_finding.md clears it, produces no id finding either (the id is
// compared numerically), and is skipped here — while capture's own reader holds
// it to the strict grammar. That divergence between the gate's filename grammar
// and the reader's is recorded as iss-2608270908346617 and owns the fix; the
// reader half of it (such a file reaching neither Issues nor Skipped) is closed,
// so the ledger at least reports the name it cannot read. Tightening this rule's
// filename grammar to match belongs on that record, because it changes what the
// gate refuses across all four stores.
func checkRecordFilenameSlug(r schemaRecord, severity string) []Finding {
	f := r.fields["slug"]
	if isAbsentValue(f.value) {
		return nil
	}
	_, fnSlug, ok := recordid.SplitRecordFilename(r.store.fileFamily, filepath.Base(r.rel))
	if !ok {
		return nil
	}
	got := strings.Trim(strings.TrimSpace(f.value), `"'`)
	if fnSlug == got {
		return nil
	}
	line := f.line
	if line == 0 {
		line = 1
	}
	return []Finding{{
		File: r.rel, Line: line, RuleID: ruleRecordSchema, Severity: severity,
		Message: "filename carries slug '" + fnSlug + "' but frontmatter declares '" + got +
			"'; a " + r.noun() + " filename is " + r.store.filename +
			", and the two spellings must be the same value",
	}}
}

// checkRecordRequiredFields asserts that a record carries every frontmatter
// property its store declares required (iss-2608261437041050). Only a store that
// declares a list is judged; the rest are left to the rules that know their own
// schemas.
//
// The defect it closes is a SILENT one, which is why it belongs in the structural
// rule rather than in the reader alone: the issue ledger's reader validates each
// record and skips the ones that fail, so a committed record missing a required
// property disappears from `capture list`, `capture status` and every other
// surface — while sitting in the ledger, counted by nothing, reported by nothing.
// A record nobody can read is not a lax record, it is a lost one.
//
// A property present but empty (`schema_version:`) counts as missing for the same
// reason: the reader cannot make a value out of it either.
func checkRecordRequiredFields(r schemaRecord, severity string) []Finding {
	var out []Finding
	for _, field := range r.store.requiredFields {
		f, present := r.fields[field]
		if present && !isAbsentValue(f.value) {
			continue
		}
		line := f.line
		if line == 0 {
			line = 1
		}
		out = append(out, Finding{
			File: r.rel, Line: line, RuleID: ruleRecordSchema, Severity: severity,
			Message: "frontmatter is missing required property '" + field + "'; the " + r.store.noun +
				" reader validates before it reads, so a record without it is skipped — invisible to every " +
				r.store.noun + " surface while it still sits in the store",
		})
	}
	return out
}

// checkIssueRecordShape mirrors capture's validateStrict shape checks for the
// ISSUE store: the additionalProperties:false unknown-key check, enum membership
// (severity/category/source), and the kebab-slug check. Each reads the ONE shared
// schema data in core/issueschema — the same allow-list and value sets capture
// validates against — so a record capture would REFUSE (and therefore skip,
// making it invisible to every capture surface) is not lint-green
// (iss-2608261447039180, iss-2608270908342889).
//
// It is scoped to the issue store: the other three stores carry different
// schemas, judged by their own rules. It deliberately does NOT yet mirror
// capture's folder<->field invariants (resolution in resolved/, wontfix_reason in
// wontfix/) or its optional-field type checks — the enum, slug and unknown-key
// checks are the determinate, highest-value half; the folder invariants are a
// follow-up.
//
// An ABSENT required value is the required-fields check's business, not this
// one's, so each value check skips a missing/null value rather than double-report.
func checkIssueRecordShape(r schemaRecord, severity string) []Finding {
	if r.store.prefix != "iss" {
		return nil
	}
	var out []Finding
	add := func(line int, msg string) {
		if line == 0 {
			line = 1
		}
		out = append(out, Finding{
			File: r.rel, Line: line, RuleID: ruleRecordSchema, Severity: severity, Message: msg,
		})
	}

	// Unknown property: capture's reader refuses any key outside the allow-list and
	// skips the whole record.
	for key, f := range r.fields {
		if issueschema.Known[key] {
			continue
		}
		add(f.line, "unknown frontmatter property '"+key+"'; capture's reader refuses a key outside the issue schema and skips the record, so it is invisible to every capture surface while it still sits in the ledger")
	}

	// Enum membership: a present but out-of-enum value is refused by capture.
	enums := []struct {
		field string
		set   []string
	}{
		{"severity", issueschema.Severities},
		{"category", issueschema.Categories},
		{"source", issueschema.Sources},
	}
	for _, e := range enums {
		f, present := r.fields[e.field]
		if !present || isAbsentValue(f.value) {
			continue
		}
		v := issueScalar(f.value)
		if !inSet(v, e.set) {
			add(f.line, "invalid "+e.field+" '"+v+"'; capture refuses a value outside {"+strings.Join(e.set, ", ")+"} and skips the record")
		}
	}

	// Kebab-slug: the slug becomes a filename, and capture refuses any other shape.
	if f, present := r.fields["slug"]; present && !isAbsentValue(f.value) {
		v := issueScalar(f.value)
		if !issueschema.SlugRe.MatchString(v) {
			add(f.line, "invalid slug '"+v+"'; a slug is kebab-case (lower-case alphanumerics joined by single hyphens) and capture refuses any other shape")
		}
	}
	return out
}

// issueScalar strips the surrounding whitespace and quotes off a frontmatter
// value so a quoted enum (`severity: "minor"`) compares as capture's parser reads
// it — unquoted.
func issueScalar(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

// inSet reports membership in a small value list.
func inSet(v string, set []string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}

// noun renders the record kind for a message.
func (r schemaRecord) noun() string { return r.store.noun }

// scanRecordStores reads every configured store once, returning its records plus
// the findings the walk itself produces — an undeclared lifecycle directory, a
// record sitting outside every bucket, and a markdown file whose name is not the
// store's id pattern. Those three are the "no lifecycle state escapes" half of the
// rule: they are about where a file IS, so they are found by the walk, not by a
// later pass over frontmatter.
func scanRecordStores(repoRoot string, cfg RuleConfig) ([]schemaRecord, []Finding, error) {
	var records []schemaRecord
	var out []Finding
	add := func(rel string, msg string) {
		out = append(out, Finding{
			File: rel, Line: 0, RuleID: ruleRecordSchema, Severity: cfg.Severity, Message: msg,
		})
	}

	for _, store := range recordStores {
		dir := cfg.RecordStores[store.prefix]
		if dir == "" {
			continue
		}
		storeAbs := filepath.Join(repoRoot, filepath.FromSlash(dir))
		entries, err := os.ReadDir(storeAbs)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, err
		}

		// A directory that is itself a CONFIGURED store root is not an undeclared
		// bucket of its parent: the reading families live inside the issue store's
		// own root, and without this the day they appear is the day the gate calls
		// each of them an undeclared issue bucket — a blocker over a directory the
		// configuration itself declares. The set is derived from cfg.RecordStores,
		// so the config stays the single declaration of what a store is.
		nestedRoots := nestedStoreRoots(cfg.RecordStores, dir)

		// A flat store holds its records directly; a bucketed store holds only its
		// declared lifecycle directories (plus the store README).
		readBucket := func(bucketAbs, bucketRel, bucket string) error {
			es, err := os.ReadDir(bucketAbs)
			if err != nil {
				return err
			}
			for _, e := range es {
				rel := filepath.Join(bucketRel, e.Name())
				if e.IsDir() {
					// A bucket (and a flat store) holds its records DIRECTLY. A
					// directory inside one is a lifecycle nobody declared, and every
					// check in this rule stops at it — which is the same escape the
					// undeclared-bucket check closes one level up, so it is closed
					// here too rather than only for the store roots. A dot-directory
					// is tooling state, not a lifecycle the record authored.
					if !strings.HasPrefix(e.Name(), ".") {
						add(rel, undeclaredSubdirMessage(store, bucket, e.Name()))
					}
					continue
				}
				if !hasMarkdownExt(e.Name()) || strings.EqualFold(e.Name(), "README.md") {
					continue
				}
				m := store.fileNumRe.FindStringSubmatch(e.Name())
				if m == nil {
					add(rel, "filename is not a well-formed "+store.noun+" filename ("+store.filename+
						"); the filename is the handle every cross-reference resolves through")
					continue
				}
				num, err := strconv.Atoi(m[1])
				if err != nil {
					continue
				}
				content, err := os.ReadFile(filepath.Join(bucketAbs, e.Name()))
				if err != nil {
					return err
				}
				lines := strings.Split(string(content), "\n")
				// A duplicated top-level key is malformed to every record consumer:
				// the strict ledger parser refuses the file (dropping it out of the
				// corpus its surfaces read), while the lenient scanner keeps only the
				// first value — so a second line can hide the value a blocker is armed
				// to reject. The lenient read below is what the rest of this rule
				// needs; the duplicate is reported alongside it so the gate refuses
				// what its consumers refuse (GitHub #357).
				for _, dup := range frontmatterDuplicates(lines) {
					out = append(out, Finding{
						File: rel, Line: dup.Line, RuleID: ruleRecordSchema, Severity: cfg.Severity,
						Message: "frontmatter has a duplicate top-level key '" + dup.Key +
							"'; the record reader refuses a duplicated key, so the file is skipped by every " +
							store.noun + " surface while the lenient scanner keeps only the first value — a second line can silence a blocker armed on the value the first hides",
					})
				}
				fields := frontmatterFields(lines)
				records = append(records, schemaRecord{
					rel:    rel,
					store:  store,
					num:    num,
					bucket: bucket,
					title:  recordTitle(lines),
					fields: fields,
					refs:   recordRefsOf(lines, fields),
				})
			}
			return nil
		}

		storeRel := repoRel(repoRoot, storeAbs)
		if !store.bucketed() {
			// A flat store declares NO buckets, so it holds its records directly and
			// readBucket reports any subdirectory of it, exactly as it does for a
			// declared bucket.
			if err := readBucket(storeAbs, storeRel, ""); err != nil {
				return nil, nil, err
			}
			continue
		}

		for _, e := range entries {
			rel := filepath.Join(storeRel, e.Name())
			if e.IsDir() {
				// A dot-directory is tooling state (an editor's, a scanner's), never
				// a lifecycle the record authored — the record's own buckets are all
				// named, so refusing one would be a blocker over somebody's .vscode.
				if strings.HasPrefix(e.Name(), ".") {
					continue
				}
				if nestedRoots[e.Name()] {
					continue
				}
				if !store.declaresBucket(e.Name()) {
					add(rel, "lifecycle directory '"+e.Name()+"' is not a declared "+store.noun+
						" bucket ("+store.bucketDesc()+"); an undeclared bucket is a lifecycle state no rule reads")
					continue
				}
				if err := readBucket(filepath.Join(storeAbs, e.Name()), rel, e.Name()); err != nil {
					return nil, nil, err
				}
				continue
			}
			if !strings.HasSuffix(e.Name(), ".md") || strings.EqualFold(e.Name(), "README.md") {
				continue
			}
			if store.fileNumRe.MatchString(e.Name()) {
				add(rel, "record sits in the store root rather than a lifecycle bucket ("+
					store.bucketDesc()+"); the directory IS the lifecycle state")
			}
		}
	}

	sort.SliceStable(records, func(i, j int) bool { return records[i].rel < records[j].rel })
	return records, out, nil
}

// nestedStoreRoots names the immediate children of dir that are themselves
// configured record-store roots. Both arguments are repo-relative, slash-joined
// store paths as the configuration spells them.
//
// It walks the CODE's store list and looks each prefix up in the configuration,
// never the configuration's own values, and the difference is the whole point.
// The exemption says "this directory is not an undeclared bucket because
// something else scans it" — so it may only be granted to a directory the scanner
// actually visits. Deriving it from every value in the map would let a committed
// line naming no store at all (a prefix this code has never heard of, pointed
// inside a real store) exempt a directory that nothing scans: a lifecycle state
// no rule reads, which is precisely the escape this rule exists to close. Which
// lifecycle states exist is code, not config, and a config that could add a
// bucket could also hide one.
func nestedStoreRoots(stores map[string]string, dir string) map[string]bool {
	parent := strings.Trim(filepath.ToSlash(dir), "/")
	out := map[string]bool{}
	for _, store := range recordStores {
		other := stores[store.prefix]
		if other == "" {
			continue
		}
		child := strings.Trim(filepath.ToSlash(other), "/")
		if child == parent || !strings.HasPrefix(child, parent+"/") {
			continue
		}
		// Only the IMMEDIATE child is a bucket-shaped name from this store's
		// point of view; a deeper store root is a child of something else.
		if rest := child[len(parent)+1:]; !strings.Contains(rest, "/") {
			out[rest] = true
		}
	}
	return out
}

// undeclaredSubdirMessage names a directory that sits where records should. A
// flat store and a declared bucket both hold records directly, so the defect is
// the same and only the wording differs.
func undeclaredSubdirMessage(store recordStore, bucket, name string) string {
	if bucket == "" {
		return "the " + store.noun + " store is flat, so subdirectory '" + name +
			"' is undeclared; records inside it are read by no rule"
	}
	return "lifecycle bucket '" + bucket + "' holds records directly, so subdirectory '" + name +
		"' is undeclared; records inside it are read by no rule"
}

// isAbsentValue reports whether a frontmatter value says "nothing here". It is
// isNull widened by the empty flow sequence, which is this record's house
// spelling for an empty list (`related_rfcs: []`) and therefore an absence, not a
// malformed value. It is local rather than folded into the shared isNull because
// isNull also judges SCALAR fields (kind, impact, slug), where a list literal is a
// wrong value rather than an unset one, and should keep saying so.
func isAbsentValue(value string) bool {
	v := strings.TrimSpace(value)
	if isNull(v) {
		return true
	}
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		return strings.TrimSpace(v[1:len(v)-1]) == ""
	}
	return false
}

// recordRefsOf reads the handles of every cross-reference field once per record.
//
// YAML spells a list two ways and the record uses both: the inline flow sequence
// (`supersedes: [adr-14, adr-15]`) and the indented block sequence. The shared
// frontmatter scanner reads same-line values only — correct for its own job, and
// the reason a block sequence reaches this rule as an empty string. An empty read
// is not a harmless degradation HERE: the bidirectional check would report that
// some OTHER file omits a link that file plainly carries, which is a confident
// false statement about a record the reader then has to go and disprove. So the
// block form is folded in at the one place the handles are parsed.
func recordRefsOf(lines []string, fields map[string]fmField) map[string][]recordRef {
	refs := make(map[string][]recordRef, len(recordParsedFields))
	for _, field := range recordParsedFields {
		f, ok := fields[field]
		if !ok {
			continue
		}
		value := f.value
		if strings.TrimSpace(value) == "" {
			value = blockSequenceAt(lines, f.line)
		}
		if hs := recordRefsIn(value); len(hs) > 0 {
			refs[field] = hs
		}
	}
	return refs
}

// blockSequenceAt returns the `- item` lines that continue a frontmatter key whose
// own line carries no value, joined so the handle scanner reads them as one value.
// line is the key's 1-based source line, so the scan starts at the line after it
// and stops at the first line that is not a list item — the next key, the closing
// delimiter, or the end of the file. Indentation is not required: YAML nests a
// block sequence under a mapping key at column 0 too, which the record uses.
func blockSequenceAt(lines []string, line int) string {
	var items []string
	for i := line; i >= 1 && i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		// A blank line or a comment INTERRUPTS a sequence without ending it: YAML
		// reads `- a`, blank, `# why`, `- b` as one two-item list. Stopping at the
		// interruption would drop the tail — and a dropped item is not a quiet
		// under-read here, it makes the bidirectional check assert that ANOTHER
		// file omits a link that file carries, the exact false claim this parse
		// exists to prevent. The closing `---` is neither, so the scan still ends
		// at the frontmatter boundary.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// A block sequence nested under a mapping key needs NO extra indentation:
		// `key:\n- item` is the same list as `key:\n  - item`, and the record writes
		// both. Only a line that is not a list item at all ends the scan.
		if !strings.HasPrefix(trimmed, "- ") {
			break
		}
		items = append(items, strings.TrimPrefix(trimmed, "- "))
	}
	return strings.Join(items, ", ")
}

// recordRefsIn reads every record handle out of a frontmatter value, tolerating
// the inline-list, quoted, and bare spellings the record uses interchangeably
// (`[adr-14, adr-15]`, `["adr-8", "adr-27"]`, `adr-4`).
func recordRefsIn(value string) []recordRef {
	if isNull(value) {
		return nil
	}
	var out []recordRef
	for _, m := range recordHandleRe.FindAllStringSubmatch(value, -1) {
		n, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		out = append(out, recordRef{prefix: strings.ToLower(m[1]), num: n})
	}
	return out
}

// recordTitle reads a record's human title out of its lines: the first H1, or —
// for a store whose records carry none, the issue ledger — the first non-blank
// body line, whitespace-collapsed, which is the same one-line summary the
// ledger's own promotion path derives. The frontmatter block (and a leading
// attribution comment before it) is skipped first, so a frontmatter value can
// never be mistaken for a body line. An empty record yields "", and the caller
// substitutes the handle.
func recordTitle(lines []string) string {
	i := recordBodyStart(lines)
	first := ""
	for ; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "# ") {
			return strings.TrimSpace(strings.TrimPrefix(lines[i], "# "))
		}
		if first == "" {
			if f := strings.Fields(lines[i]); len(f) > 0 {
				first = strings.Join(f, " ")
			}
		}
	}
	return first
}

// recordBodyStart returns the index of the first BODY line: past a leading
// attribution comment and blank lines, and past the frontmatter block if the
// document opens one. It is the one place that skip is expressed, so recordTitle
// and recordH1 can never disagree about where a document's prose begins.
func recordBodyStart(lines []string) int {
	i := 0
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if t == "" || strings.HasPrefix(t, "<!--") {
			i++
			continue
		}
		break
	}
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		i++
		for i < len(lines) && strings.TrimSpace(lines[i]) != "---" {
			i++
		}
		if i < len(lines) {
			i++
		}
	}
	return i
}

// recordH1 returns a document's first H1 and its 1-based line, or ("", 0) when it
// has none.
//
// recordTitle's first-body-line fallback is deliberately NOT applied here. That
// fallback exists for the issue ledger, whose records carry no heading; applied
// to arbitrary markdown it would read an ordinary opening sentence as a title,
// and a sentence that happens to open with a record handle is a mention, not a
// claim on the id.
func recordH1(lines []string) (string, int) {
	for i := recordBodyStart(lines); i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "# ") {
			return strings.TrimSpace(strings.TrimPrefix(lines[i], "# ")), i + 1
		}
	}
	return "", 0
}

// refsContain reports whether a parsed handle set names the given handle.
func refsContain(refs []recordRef, want recordRef) bool {
	for _, h := range refs {
		if h == want {
			return true
		}
	}
	return false
}
