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
)

const ruleRecordSchema = "record_schema"

var (
	// A record handle as it is written in a cross-reference field: adr-6, ADR-6,
	// itd-47, iss-39. The number is compared numerically, so the zero-padded and
	// bare spellings of one id are the same handle.
	recordHandleRe = regexp.MustCompile(`(?i)\b(adr|itd|iss)-(\d+)\b`)
	// Filename → id number for each store. An ADR filename is the zero-padded
	// sequential form (0012-<slug>.md); the other three carry the prose handle.
	adrFileNumRe    = regexp.MustCompile(`^(\d+)-.*\.md$`)
	intentFileNumRe = regexp.MustCompile(`^itd-(\d+).*\.md$`)
	specFileNumRe   = regexp.MustCompile(`^spc-(\d+).*\.md$`)
	issueFileNumRe  = regexp.MustCompile(`^iss-(\d+).*\.md$`)
	// The cross-reference frontmatter fields whose targets must resolve. They are
	// the record's machine-readable claims that another record exists and is a
	// live input — as distinct from prose, where naming a released or retired id
	// ("itd-38's id was released, not reserved") is legitimate narration.
	recordRefFields = []string{"supersedes", "related_adrs", "related_intents", "builds_on", "blocked_by"}
)

// recordStore describes one identified record store: where the buckets are, and
// how a well-formed filename spells the id.
type recordStore struct {
	// prefix is the id prefix (adr, itd, spc, iss) and the record_stores key.
	prefix string
	// noun names the record kind in a finding message.
	noun string
	// buckets are the lifecycle directories the store declares. A nil buckets is
	// a FLAT store (the ADR store): records sit directly in it.
	buckets []string
	// fileNumRe extracts the id number from a filename (submatch 1).
	fileNumRe *regexp.Regexp
	// filename describes the convention a finding message quotes.
	filename string
}

// recordStores is the closed set of identified record stores. It is code, not
// config: which lifecycle states exist is the record's schema, and a config that
// could add a bucket could also hide one.
var recordStores = []recordStore{
	{prefix: "adr", noun: "ADR", buckets: nil, fileNumRe: adrFileNumRe, filename: "<NNNN>-<slug>.md"},
	{prefix: "itd", noun: "intent", buckets: intentBucketNames, fileNumRe: intentFileNumRe, filename: "itd-<N>-<slug>.md"},
	{prefix: "spc", noun: "spec", buckets: specBucketNames, fileNumRe: specFileNumRe, filename: "spc-<N>-<slug>.md"},
	{prefix: "iss", noun: "issue", buckets: issueStatusDirs, fileNumRe: issueFileNumRe, filename: "iss-<N>-<slug>.md"},
}

// schemaRecord is one record file as the schema rule sees it: which store and
// bucket hold it, the id number its FILENAME claims, and its frontmatter.
type schemaRecord struct {
	rel    string
	store  recordStore
	num    int
	bucket string
	fields map[string]fmField
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
// store once and asserts four invariants across them: bucket coverage, filename ↔
// id agreement, cross-reference resolution, and bidirectional supersession. A
// store that is not configured, or whose directory is absent, contributes nothing
// and is not an error — an unpopulated repository is a state, not a fault.
func checkRecordSchema(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	records, out, err := scanRecordStores(repoRoot, cfg)
	if err != nil {
		return nil, err
	}

	// index: what the corpus HAS. retired: ids that a record declares it replaced.
	// The ADR lifecycle prunes a superseded record once its successor lands and
	// keeps the trace in the successor's `supersedes` (decisions/adrs/README.md),
	// so a handle naming a pruned id resolves to that declaration rather than to a
	// file — the record has not lost track of it.
	index := map[recordRef]schemaRecord{}
	retired := map[recordRef]bool{}
	for _, r := range records {
		index[recordRef{r.store.prefix, r.num}] = r
	}
	for _, r := range records {
		for _, h := range recordRefsIn(r.fields["supersedes"].value) {
			retired[h] = true
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

		// Cross-references: a named record must be in the corpus, or declared
		// retired by the record that replaced it.
		for _, field := range recordRefFields {
			f := r.fields[field]
			if isNull(f.value) {
				continue
			}
			for _, h := range recordRefsIn(f.value) {
				if _, ok := index[h]; ok || retired[h] {
					continue
				}
				add(r.rel, f.line, field+" names '"+h.String()+"', which is not a record in the corpus and no record declares it superseded; a cross-reference is a claim that the record exists")
			}
		}

		// Supersession, direction A→B: the successor must exist (a live decision is
		// never a pruned one) and must name this record back.
		sb := r.fields["superseded_by"]
		if isNull(sb.value) {
			continue
		}
		targets := recordRefsIn(sb.value)
		if len(targets) == 0 {
			add(r.rel, sb.line, "superseded_by '"+sb.value+"' is not a record handle (want adr-N, itd-N, or iss-N)")
			continue
		}
		for _, h := range targets {
			target, ok := index[h]
			if !ok {
				add(r.rel, sb.line, "superseded_by names '"+h.String()+"', which is not a record in the corpus; a successor decision must be present")
				continue
			}
			if !recordRefsContain(target.fields["supersedes"].value, recordRef{r.store.prefix, r.num}) {
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
		if isNull(sup.value) {
			continue
		}
		for _, h := range recordRefsIn(sup.value) {
			target, ok := index[h]
			if !ok {
				continue
			}
			if !recordRefsContain(target.fields["superseded_by"].value, recordRef{r.store.prefix, r.num}) {
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
	if isNull(f.value) {
		return nil
	}
	want := r.handle()
	got := strings.Trim(strings.TrimSpace(f.value), `"'`)
	if strings.EqualFold(got, want) {
		return nil
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

		declared := make(map[string]bool, len(store.buckets))
		for _, b := range store.buckets {
			declared[b] = true
		}

		// A flat store holds its records directly; a bucketed store holds only its
		// declared lifecycle directories (plus the store README).
		readBucket := func(bucketAbs, bucketRel, bucket string) error {
			es, err := os.ReadDir(bucketAbs)
			if err != nil {
				return err
			}
			for _, e := range es {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				rel := filepath.Join(bucketRel, e.Name())
				if strings.EqualFold(e.Name(), "README.md") {
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
				records = append(records, schemaRecord{
					rel:    rel,
					store:  store,
					num:    num,
					bucket: bucket,
					fields: frontmatterFields(strings.Split(string(content), "\n")),
				})
			}
			return nil
		}

		storeRel := repoRel(repoRoot, storeAbs)
		if store.buckets == nil {
			if err := readBucket(storeAbs, storeRel, ""); err != nil {
				return nil, nil, err
			}
			continue
		}

		for _, e := range entries {
			rel := filepath.Join(storeRel, e.Name())
			if e.IsDir() {
				if !declared[e.Name()] {
					add(rel, "lifecycle directory '"+e.Name()+"' is not a declared "+store.noun+
						" bucket ("+strings.Join(store.buckets, ", ")+"); an undeclared bucket is a lifecycle state no rule reads")
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
					strings.Join(store.buckets, ", ")+"); the directory IS the lifecycle state")
			}
		}
	}

	sort.SliceStable(records, func(i, j int) bool { return records[i].rel < records[j].rel })
	return records, out, nil
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

// recordRefsContain reports whether a frontmatter value names the given handle.
func recordRefsContain(value string, want recordRef) bool {
	for _, h := range recordRefsIn(value) {
		if h == want {
			return true
		}
	}
	return false
}
