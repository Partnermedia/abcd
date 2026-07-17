package memory

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// schema.go — the page-side schema of the memory store (07-memory.md §1–§3):
// the closed source.class enum, typed source: frontmatter builders/validators,
// page-filename parsing, the DistilledPage distiller contract + topic hash, the
// pure dedup helper, and derived-sibling rendering.

// memorySourceClasses is the closed, PR-to-extend page-class enum.
var memorySourceClasses = map[string]bool{
	"session_memory":            true,
	"external_pdf":              true,
	"external_transcript":       true,
	"external_article":          true,
	"oracle_review":             true,
	"work_notes":                true,
	"issue_ledger":              true,
	"dredge_synthesis":          true,
	"spec_modification_grammar": true,
	"modification_grammar":      true,
}

var memorySourceClassList = []string{
	"session_memory", "external_pdf", "external_transcript", "external_article",
	"oracle_review", "work_notes", "issue_ledger", "dredge_synthesis",
	"spec_modification_grammar", "modification_grammar",
}

// backfillSourceClass is the default class backfilled onto pre-itd-36 flat pages.
const backfillSourceClass = "session_memory"

// externalClassPrefix marks a licence-bearing source.class value.
const externalClassPrefix = "external_"

func isExternalClass(cls any) bool {
	s, ok := cls.(string)
	return ok && strings.HasPrefix(s, externalClassPrefix)
}

var siblingFiles = map[string]bool{
	"README.md": true, "index.md": true, "log.md": true, "contradictions.md": true,
}

var pageNameRe = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9-]*)_([A-Za-z0-9][A-Za-z0-9-]*)_([A-Za-z0-9][A-Za-z0-9_-]*)\.md$`)
var typeDomainRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)
var slugRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// ---------------------------------------------------------------------------
// Page filenames
// ---------------------------------------------------------------------------

// ParsePageFilename parses <type>_<domain>_<slug>.md, reporting ok=false on a
// non-conforming name.
func ParsePageFilename(filename string) (typ, domain, slug string, ok bool) {
	m := pageNameRe.FindStringSubmatch(filename)
	if m == nil {
		return "", "", "", false
	}
	return m[1], m[2], m[3], true
}

// IsMemoryPageName reports whether filename is an individual memory page — a .md
// directly in the store that is not a sibling and not a dotfile.
func IsMemoryPageName(filename string) bool {
	return strings.HasSuffix(filename, ".md") &&
		!siblingFiles[filename] &&
		!strings.HasPrefix(filename, ".") &&
		!strings.ContainsAny(filename, `/\`)
}

// ---------------------------------------------------------------------------
// source: block accessors + validation
// ---------------------------------------------------------------------------

// SourceClasses returns the class list of either source shape (scalar class to
// a one-element list).
func SourceClasses(source map[string]any) []string {
	if source == nil {
		return nil
	}
	if s, ok := source["class"].(string); ok {
		return []string{s}
	}
	if classes, ok := source["classes"].([]any); ok {
		var out []string
		for _, c := range classes {
			if s, ok := c.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// SourceHashes returns the page's full contributing source-hash set:
// source.source_hash (single) and every source.sources[].source_hash (multi).
func SourceHashes(source map[string]any) []string {
	if source == nil {
		return nil
	}
	var hashes []string
	if sh, ok := source["source_hash"].(string); ok && sh != "" {
		hashes = append(hashes, sh)
	}
	if entries, ok := source["sources"].([]any); ok {
		for _, e := range entries {
			if em, ok := e.(map[string]any); ok {
				if eh, ok := em["source_hash"].(string); ok && eh != "" && !contains(hashes, eh) {
					hashes = append(hashes, eh)
				}
			}
		}
	}
	return hashes
}

func deriveClasses(sources []any) []string {
	var seen []string
	for _, e := range sources {
		if em, ok := e.(map[string]any); ok {
			if cls, ok := em["class"].(string); ok && !contains(seen, cls) {
				seen = append(seen, cls)
			}
		}
	}
	return seen
}

func requireClass(value any, where string) (string, error) {
	s, ok := value.(string)
	if !ok || !memorySourceClasses[s] {
		return "", newSchemaError("%s: source class must be one of the closed enum, got %v", where, value)
	}
	return s, nil
}

func requireCitation(value any, where string) error {
	m, ok := value.(map[string]any)
	if !ok || len(m) == 0 {
		return newSchemaError("%s: citation must be a non-empty mapping (the object shape from 09-provenance-substrate.md §2)", where)
	}
	return nil
}

func requireDate(value any, where string) error {
	s, ok := value.(string)
	if !ok || !dateRe.MatchString(s) {
		return newSchemaError("%s: ingested_at must be YYYY-MM-DD, got %v", where, value)
	}
	return nil
}

var sourceEntryKeys = []string{"class", "citation", "licence", "source_hash", "ingested_at"}

func validateSourceEntry(entry any, where string) (string, error) {
	em, ok := entry.(map[string]any)
	if !ok {
		return "", newSchemaError("%s: each sources[] entry must be a mapping", where)
	}
	for _, k := range sourceEntryKeys {
		if _, present := em[k]; !present {
			return "", newSchemaError("%s: sources[] entry missing required key %q", where, k)
		}
	}
	cls, err := requireClass(em["class"], where)
	if err != nil {
		return "", err
	}
	if err := requireCitation(em["citation"], where); err != nil {
		return "", err
	}
	if err := requireDate(em["ingested_at"], where); err != nil {
		return "", err
	}
	lic, ok := em["licence"].(string)
	if !ok || lic == "" {
		return "", newSchemaError("%s: licence must be a non-empty string (explicit 'unknown' is acceptable)", where)
	}
	sh, ok := em["source_hash"].(string)
	if !ok || !hex64Re.MatchString(sh) {
		return "", newSchemaError("%s: source_hash must be a sha256 hex digest", where)
	}
	return cls, nil
}

// validateSourceBlock validates a source: frontmatter block (either shape).
func validateSourceBlock(source any) error {
	sm, ok := source.(map[string]any)
	if !ok {
		return newSchemaError("source: block must be a mapping")
	}
	_, isSingle := sm["class"]
	_, hasClasses := sm["classes"]
	_, hasSources := sm["sources"]
	isMulti := hasClasses || hasSources
	if isSingle && isMulti {
		return newSchemaError("source: block mixes the single-source and multi-source shapes")
	}
	if !isSingle && !isMulti {
		return newSchemaError("source: block carries neither a scalar `class` nor a `classes` + `sources` pair")
	}
	if isSingle {
		if _, err := requireClass(sm["class"], "source"); err != nil {
			return err
		}
		if _, ok := sm["weighting_note"]; ok {
			return newSchemaError("source.weighting_note never appears on a single-source page")
		}
		if _, ok := sm["citation"]; ok {
			if err := requireCitation(sm["citation"], "source"); err != nil {
				return err
			}
		}
		if _, ok := sm["ingested_at"]; ok {
			if err := requireDate(sm["ingested_at"], "source"); err != nil {
				return err
			}
		}
		return nil
	}
	if !hasClasses || !hasSources {
		return newSchemaError("multi-source block requires BOTH source.classes and source.sources")
	}
	sources, ok := sm["sources"].([]any)
	if !ok || len(sources) == 0 {
		return newSchemaError("source.sources must be a non-empty list of per-source entries")
	}
	for i, e := range sources {
		if _, err := validateSourceEntry(e, fmt.Sprintf("source.sources[%d]", i)); err != nil {
			return err
		}
	}
	derived := deriveClasses(sources)
	declared, ok := sm["classes"].([]any)
	declaredStr := make([]string, 0, len(declared))
	for _, c := range declared {
		if s, ok := c.(string); ok {
			declaredStr = append(declaredStr, s)
		}
	}
	if !ok || !equalStringSets(declaredStr, derived) {
		return newSchemaError("source.classes must equal the set derived from each sources[].class (expected %v, got %v)", derived, declaredStr)
	}
	if len(derived) >= 2 {
		note, ok := sm["weighting_note"].(string)
		if !ok || strings.TrimSpace(note) == "" {
			return newSchemaError("source.weighting_note is required when a page mixes >=2 source classes")
		}
	}
	return nil
}

func buildSingleSource(sourceClass string, citation map[string]any, licence, sourceHash, ingestedAt string) (map[string]any, error) {
	block := map[string]any{
		"class":       sourceClass,
		"citation":    deepCopyMap(citation),
		"licence":     licence,
		"source_hash": sourceHash,
		"ingested_at": ingestedAt,
	}
	if err := validateSourceBlock(block); err != nil {
		return nil, err
	}
	return block, nil
}

func buildMultiSource(sources []map[string]any, weightingNote string) (map[string]any, error) {
	entries := make([]any, len(sources))
	for i, e := range sources {
		entries[i] = deepCopyMap(e)
	}
	classes := deriveClasses(entries)
	block := map[string]any{"classes": toAnySlice(classes)}
	if weightingNote != "" {
		block["weighting_note"] = weightingNote
	}
	block["sources"] = entries
	if err := validateSourceBlock(block); err != nil {
		return nil, err
	}
	return block, nil
}

// ---------------------------------------------------------------------------
// DistilledPage (distiller boundary) + topic hash
// ---------------------------------------------------------------------------

var distilledPageKeys = map[string]bool{
	"type": true, "domain": true, "slug": true, "body": true, "source": true,
	"topic_hash": true, "contradicts": true, "recall": true,
}
var distilledPageRequired = []string{"type", "domain", "slug", "body", "source"}

// DistilledPage is one validated distiller output page. TopicHash is computed by
// ValidateDistilledPage (a supplied one is rejected) and persisted into the
// written page's frontmatter.
type DistilledPage struct {
	Type        string
	Domain      string
	Slug        string
	Body        string
	TopicHash   string
	Source      map[string]any
	Contradicts []string
	Recall      []string
}

// Filename returns <type>_<domain>_<slug>.md.
func (p DistilledPage) Filename() string {
	return fmt.Sprintf("%s_%s_%s.md", p.Type, p.Domain, p.Slug)
}

func pageSubject(body string) string {
	for _, line := range strings.Split(body, "\n") {
		s := strings.TrimSpace(line)
		if s != "" {
			s = strings.TrimSpace(strings.TrimLeft(s, "#"))
			return strings.Join(strings.Fields(strings.ToLower(s)), " ")
		}
	}
	return ""
}

func computeTopicHash(typ, domain, slug, subject string) string {
	normalised := strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(subject))), " ")
	h := sha256.New()
	for _, field := range []string{typ, domain, slug, normalised} {
		data := []byte(field)
		var lenBuf [8]byte
		binary.BigEndian.PutUint64(lenBuf[:], uint64(len(data)))
		h.Write(lenBuf[:])
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func requireStrList(value any, key string) ([]string, error) {
	list, ok := value.([]any)
	if !ok {
		return nil, newSchemaError("DistilledPage.%s must be a list of non-empty strings", key)
	}
	var out []string
	for _, item := range list {
		s, ok := item.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return nil, newSchemaError("DistilledPage.%s entries must be non-empty strings", key)
		}
		out = append(out, s)
	}
	return out, nil
}

// ValidateDistilledPage validates one raw page dict against the DistilledPage
// schema, computing topic_hash (a supplied one is rejected). The ingest path
// runs this before any write.
func ValidateDistilledPage(data map[string]any) (DistilledPage, error) {
	if data == nil {
		return DistilledPage{}, newSchemaError("DistilledPage must be a mapping")
	}
	var unknown []string
	for k := range data {
		if !distilledPageKeys[k] {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return DistilledPage{}, newSchemaError("DistilledPage carries unknown key(s) %v — the boundary fails closed on keys outside the schema", unknown)
	}
	for _, k := range distilledPageRequired {
		if _, ok := data[k]; !ok {
			return DistilledPage{}, newSchemaError("DistilledPage missing required key %q", k)
		}
	}
	typ, _ := data["type"].(string)
	domain, _ := data["domain"].(string)
	slug, _ := data["slug"].(string)
	if !typeDomainRe.MatchString(typ) {
		return DistilledPage{}, newSchemaError("DistilledPage.type must be a filename-safe token, got %v", data["type"])
	}
	if !typeDomainRe.MatchString(domain) {
		return DistilledPage{}, newSchemaError("DistilledPage.domain must be a filename-safe token, got %v", data["domain"])
	}
	if !slugRe.MatchString(slug) {
		return DistilledPage{}, newSchemaError("DistilledPage.slug must be a filename-safe token, got %v", data["slug"])
	}
	body, ok := data["body"].(string)
	if !ok || strings.TrimSpace(body) == "" {
		return DistilledPage{}, newSchemaError("DistilledPage.body must be non-empty text")
	}
	source, ok := data["source"].(map[string]any)
	if !ok {
		return DistilledPage{}, newSchemaError("DistilledPage.source must be a mapping")
	}
	if err := validateSourceBlock(source); err != nil {
		return DistilledPage{}, err
	}
	var contradicts []string
	if raw, ok := data["contradicts"]; ok && raw != nil {
		list, err := requireStrList(raw, "contradicts")
		if err != nil {
			return DistilledPage{}, err
		}
		for _, c := range list {
			c = strings.TrimSuffix(c, ".md")
			if c == "" {
				return DistilledPage{}, newSchemaError("DistilledPage.contradicts entries must be page ids (non-empty after stripping a trailing .md)")
			}
			contradicts = append(contradicts, c)
		}
	}
	var recall []string
	if raw, ok := data["recall"]; ok && raw != nil {
		list, err := requireStrList(raw, "recall")
		if err != nil {
			return DistilledPage{}, err
		}
		recall = list
	}
	if _, ok := data["topic_hash"]; ok {
		return DistilledPage{}, newSchemaError("DistilledPage.topic_hash must not be supplied — the boundary computes it deterministically")
	}
	page := DistilledPage{
		Type:        typ,
		Domain:      domain,
		Slug:        slug,
		Body:        body,
		TopicHash:   computeTopicHash(typ, domain, slug, pageSubject(body)),
		Source:      deepCopyMap(source),
		Contradicts: contradicts,
		Recall:      recall,
	}
	if _, _, _, ok := ParsePageFilename(page.Filename()); !ok || !IsMemoryPageName(page.Filename()) {
		return DistilledPage{}, newSchemaError("DistilledPage assembles an unwritable filename: %q", page.Filename())
	}
	return page, nil
}

// ---------------------------------------------------------------------------
// Pure dedup — resolve_distilled_pages
// ---------------------------------------------------------------------------

// PlannedWrite is one page the write plan materialises (field-compatible with
// PageWrite).
type PlannedWrite struct {
	Filename    string
	Frontmatter map[string]any
	Body        string
}

// WritePlan is the output of ResolveDistilledPages.
type WritePlan struct {
	Writes         []PlannedWrite
	RegistryPages  map[string][]string
	Linked         [][2]string
	Contradictions [][2]string
}

type knownPage struct {
	topic  string
	hashes map[string]bool
}

func uniqueFilename(page DistilledPage, taken map[string]bool) (string, error) {
	if !taken[page.Filename()] {
		return page.Filename(), nil
	}
	for n := 2; n < 10000; n++ {
		candidate := fmt.Sprintf("%s_%s_%s-%d.md", page.Type, page.Domain, page.Slug, n)
		if !taken[candidate] {
			return candidate, nil
		}
	}
	return "", newSchemaError("cannot derive a free fork filename for %q", page.Filename())
}

// ResolveDistilledPages is the pure cross-ref dedup owner (link / fork+contradiction
// / new rule). existing maps each on-disk page filename to its parsed
// frontmatter (unparseable pages present-but-{} so names stay taken).
func ResolveDistilledPages(existing map[string]map[string]any, pages []DistilledPage) (WritePlan, error) {
	known := map[string]knownPage{}
	for fname, fm := range existing {
		var topic string
		if t, ok := fm["topic_hash"].(string); ok {
			topic = t
		}
		hashes := map[string]bool{}
		if src, ok := fm["source"].(map[string]any); ok {
			for _, h := range SourceHashes(src) {
				hashes[h] = true
			}
		}
		known[fname] = knownPage{topic: topic, hashes: hashes}
	}
	taken := map[string]bool{}
	for f := range known {
		taken[f] = true
	}
	var writes []PlannedWrite
	registry := map[string][]string{}
	var linked [][2]string
	var contradictions [][2]string

	register := func(hashes map[string]bool, filename string) {
		for _, h := range sortedSet(hashes) {
			if !contains(registry[h], filename) {
				registry[h] = append(registry[h], filename)
			}
		}
	}

	for _, page := range pages {
		pageHashes := map[string]bool{}
		for _, h := range SourceHashes(page.Source) {
			pageHashes[h] = true
		}
		var sameTopic []string
		for f, kp := range known {
			if kp.topic != "" && kp.topic == page.TopicHash {
				sameTopic = append(sameTopic, f)
			}
		}
		sort.Strings(sameTopic)

		var linkTargets []string
		for _, f := range sameTopic {
			if intersects(known[f].hashes, pageHashes) {
				linkTargets = append(linkTargets, f)
			}
		}
		if len(linkTargets) > 0 {
			target := linkTargets[0]
			linked = append(linked, [2]string{page.Filename(), target})
			register(pageHashes, target)
			merged := known[target]
			for h := range pageHashes {
				merged.hashes[h] = true
			}
			known[target] = merged
			continue
		}

		filename, err := uniqueFilename(page, taken)
		if err != nil {
			return WritePlan{}, err
		}
		contradicts := append([]string(nil), page.Contradicts...)
		for _, existingF := range sameTopic {
			pageID := strings.TrimSuffix(existingF, ".md")
			if !contains(contradicts, pageID) {
				contradicts = append(contradicts, pageID)
			}
			pair := [2]string{filename, pageID}
			if !containsPair(contradictions, pair) {
				contradictions = append(contradictions, pair)
			}
		}
		if err := validateSourceBlock(page.Source); err != nil {
			return WritePlan{}, err
		}
		frontmatter := map[string]any{
			"source":     deepCopyMap(page.Source),
			"topic_hash": page.TopicHash,
		}
		if len(page.Recall) > 0 {
			frontmatter["recall"] = toAnySlice(page.Recall)
		}
		if len(contradicts) > 0 {
			sorted := append([]string(nil), contradicts...)
			sort.Strings(sorted)
			frontmatter["contradicts"] = toAnySlice(sorted)
		}
		writes = append(writes, PlannedWrite{Filename: filename, Frontmatter: frontmatter, Body: page.Body})
		taken[filename] = true
		known[filename] = knownPage{topic: page.TopicHash, hashes: pageHashes}
		register(pageHashes, filename)
	}

	return WritePlan{
		Writes:         writes,
		RegistryPages:  registry,
		Linked:         linked,
		Contradictions: contradictions,
	}, nil
}

// ---------------------------------------------------------------------------
// PageInfo + derived-sibling rendering
// ---------------------------------------------------------------------------

// PageInfo carries the per-page facts the derived siblings render from.
type PageInfo struct {
	Filename    string
	Classes     []string
	Domain      string
	Summary     string
	Contradicts []string
}

func pageInfoFrom(filename, text string) PageInfo {
	var fm map[string]any
	body := text
	if strings.HasPrefix(text, "---") {
		if region, b, err := splitFileFrontmatter(text); err == nil {
			if parsed, err := parseFrontmatter("---\n" + region + "---\n"); err == nil {
				fm = parsed
				body = b
			}
		}
	}
	var classes []string
	if fm != nil {
		if src, ok := fm["source"].(map[string]any); ok {
			classes = SourceClasses(src)
		}
	}
	domain := ""
	if _, d, _, ok := ParsePageFilename(filename); ok {
		domain = d
	}
	summary := ""
	for _, line := range strings.Split(body, "\n") {
		s := strings.TrimSpace(line)
		if s != "" {
			summary = strings.TrimSpace(strings.TrimLeft(s, "#"))
			break
		}
	}
	var contradicts []string
	if fm != nil {
		if raw, ok := fm["contradicts"].([]any); ok {
			for _, c := range raw {
				if s, ok := c.(string); ok && s != "" {
					contradicts = append(contradicts, s)
				}
			}
		}
	}
	return PageInfo{Filename: filename, Classes: classes, Domain: domain, Summary: summary, Contradicts: contradicts}
}

const indexHeader = "# Memory index\n\nGenerated catalog — one line per page (class | domain | summary).\nRegenerated from the per-page files on every memory write; do not\nhand-edit (the per-page files are the source of truth).\n"

const contradictionsHeader = "# Contradictions\n\nCurator-surfaced contradictions register. Rendered ONE-WAY from page\n`contradicts:` frontmatter (only the newer page carries the key; the\nexisting immutable page is never touched). The page frontmatter is the\nsource of truth — record contradictions there, not here.\n"

// RenderIndex renders index.md deterministically (sorted by filename).
func RenderIndex(pages []PageInfo) string {
	sorted := append([]PageInfo(nil), pages...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Filename < sorted[j].Filename })
	lines := []string{strings.TrimRight(indexHeader, "\n"), ""}
	if len(sorted) == 0 {
		lines = append(lines, "(no pages)")
	} else {
		for _, p := range sorted {
			cls := "(unclassified)"
			if len(p.Classes) > 0 {
				cls = strings.Join(p.Classes, "+")
			}
			domain := p.Domain
			if domain == "" {
				domain = "(no domain)"
			}
			summary := p.Summary
			if summary == "" {
				summary = "(no summary)"
			}
			lines = append(lines, fmt.Sprintf("- `%s` — %s | %s | %s", p.Filename, cls, domain, summary))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// RenderContradictions renders contradictions.md deterministically from page facts.
func RenderContradictions(pages []PageInfo) string {
	sorted := append([]PageInfo(nil), pages...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Filename < sorted[j].Filename })
	lines := []string{strings.TrimRight(contradictionsHeader, "\n"), ""}
	var entries []string
	for _, p := range sorted {
		targets := append([]string(nil), p.Contradicts...)
		sort.Strings(targets)
		for _, t := range targets {
			entries = append(entries, fmt.Sprintf("- `%s` contradicts `%s`", p.Filename, t))
		}
	}
	if len(entries) == 0 {
		lines = append(lines, "(none recorded)")
	} else {
		lines = append(lines, entries...)
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderLogEvent(timestamp, classLabel, slug, summary string) string {
	return fmt.Sprintf("\n## [%s] %s | %s — %s\n", timestamp, classLabel, slug, summary)
}

// ---------------------------------------------------------------------------
// Skeletons (cold-start creation)
// ---------------------------------------------------------------------------

func skeletonReadme() string {
	var enum strings.Builder
	for i, cls := range memorySourceClassList {
		if i > 0 {
			enum.WriteByte('\n')
		}
		enum.WriteString("- `" + cls + "`")
	}
	return "# `.abcd/memory/`\n\n" +
		"abcd's compounding-curated knowledge substrate (itd-36). Individual\n" +
		"pages use the flat naming `<type>_<domain>_<slug>.md` and carry typed\n" +
		"`source:` frontmatter; `citation` is an object (mapping), and a\n" +
		"multi-source page lists each contributing source — with its own\n" +
		"`class` — under `source.sources`.\n\n" +
		"Siblings:\n\n" +
		"- `index.md` — generated catalog, one line per page; regenerated from\n" +
		"  the per-page files on every write (do not hand-edit).\n" +
		"- `log.md` — append-only page-write event record.\n" +
		"- `contradictions.md` — contradictions register, rendered one-way\n" +
		"  from page `contradicts:` frontmatter.\n\n" +
		"`source.class` enum (closed, PR-to-extend via\n" +
		"`02-constraints/04-naming.md`):\n\n" +
		enum.String() + "\n\n" +
		"Full schema + lint contract:\n" +
		"`.abcd/development/brief/05-internals/07-memory.md`.\n"
}

func skeletonIndex() string { return RenderIndex(nil) }

func skeletonLog() string {
	return "# Memory log\n\nAppend-only page-write event record:\n`## [YYYY-MM-DD HH:MM] <upstream_class> | <slug> — <summary>`.\n"
}

func skeletonContradictions() string { return RenderContradictions(nil) }

// ---------------------------------------------------------------------------
// small helpers
// ---------------------------------------------------------------------------

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// equalStringSets reports multiset equality, order-independent. source.classes is
// specified as the SET derived from sources[].class, so a caller listing the same
// classes in a different order must validate — comparing order-sensitively (as
// equalStrings did) rejected a correct page and contradicted the error message's
// own "set" wording.
func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, x := range a {
		seen[x]++
	}
	for _, y := range b {
		if seen[y] == 0 {
			return false
		}
		seen[y]--
	}
	return true
}

func sortedSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func intersects(a, b map[string]bool) bool {
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}

func containsPair(pairs [][2]string, p [2]string) bool {
	for _, x := range pairs {
		if x == p {
			return true
		}
	}
	return false
}
