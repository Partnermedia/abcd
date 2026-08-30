package reading

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/intentdriven/abcd/internal/core/frontmatter"
	"github.com/intentdriven/abcd/internal/core/site"
)

// Field projection: the heading-scoped extractor this package owns.
//
// Enumeration of the records comes from the record graph, which reports a
// record's id, store, bucket and path and no body at all. Projection is
// therefore a read of the file the graph named, not a second reading of the
// record's shape — which is the distinction that keeps one parser of the
// record in this binary.
//
// A field resolves as a heading section where the file carries a heading of
// that name, and otherwise as a frontmatter key. Nothing else resolves: a field
// the file does not carry contributes no item, which is what lets one
// projection describe an intent whose sections the record is still growing.

// trimBlankEdges joins a section body, dropping the blank lines at either end so
// a projected field is the text and not the whitespace around it.
func trimBlankEdges(lines []string) string {
	start, end := 0, len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

// redactExcluded removes the exclusion floor's key-signalled and
// heading-signalled material from a document before anything is taken out of
// it.
//
// It runs on every admitted file, projected or whole. Positive inclusion at
// file granularity is not enough on its own: a brief chapter and a spec travel
// whole, and a frontmatter key stamped on every record by the command that mints
// it would ride along inside them. The floor names those keys and headings
// precisely so the signal is mechanical, and this is where the signal is read.
func redactExcluded(rel, doc string, exclusions []Exclusion) (string, error) {
	// Both signals are RECORD shapes, and only a markdown file can carry one. A
	// Go file has no frontmatter and no Audit Notes heading; what it can have is
	// a raw string literal opening a fence at the left margin, which the section
	// scan rightly refuses as an unterminated block — and refusing it there would
	// let one unrelated source file stop every assembly the repository can run.
	// The scope of the signal is the scope of the parse.
	if !strings.EqualFold(path.Ext(rel), ".md") {
		return doc, nil
	}
	keys := map[string]bool{}
	headings := map[string]bool{}
	for _, e := range exclusions {
		switch e.Signal {
		case "frontmatter key":
			keys[e.Detail] = true
		case "heading":
			headings[e.Detail] = true
		}
	}
	lines := strings.Split(doc, "\n")
	drop := make([]bool, len(lines))

	// The key signal. A key's continuation lines are indented, so dropping the
	// indented run below it removes a block value whole rather than leaving its
	// body behind as orphaned prose.
	for key, field := range frontmatter.Fields(lines) {
		if !keys[key] {
			continue
		}
		i := field.Line - 1
		if i < 0 || i >= len(lines) {
			continue
		}
		drop[i] = true
		for j := i + 1; j < len(lines); j++ {
			if lines[j] == "" || (lines[j][0] != ' ' && lines[j][0] != '\t') {
				break
			}
			drop[j] = true
		}
	}

	// The heading signal, over the fence-aware section scan so a `#` inside a
	// code block is not mistaken for one.
	body, offset := site.StripFrontmatter(doc)
	sections, err := site.Sections(rel, body, offset)
	if err != nil {
		return "", fmt.Errorf("reading: reading the sections of %s: %w", rel, err)
	}
	for i, sec := range sections {
		if sec.Level == 0 || !headings[sec.Title] {
			continue
		}
		start, end := sectionSpan(sections, i, len(lines))
		for j := start; j < end; j++ {
			if j >= 0 {
				drop[j] = true
			}
		}
	}

	kept := make([]string, 0, len(lines))
	for i, line := range lines {
		if !drop[i] {
			kept = append(kept, line)
		}
	}
	out := strings.Join(kept, "\n")
	if err := verifyRedaction(rel, doc, out, keys, headings); err != nil {
		return "", err
	}
	return out, nil
}

// excludedKeyLineRe matches an excluded key still at column 0. The key set is
// small and fixed, so the pattern is composed from it rather than spelled twice.
var excludedKeyLineRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)\s*:`)

// headingLineRe matches a markdown ATX heading and captures its title.
var headingLineRe = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)

// verifyRedaction is the key-and-heading half of the exclusion floor made
// fail-closed, the way the path half already is.
//
// Redaction is a positive act over what a parser reported, and three shapes slip
// past a parser that reports one value per key and matches a title exactly. A
// DUPLICATED key keeps its second copy, because Fields keeps the first
// occurrence and drops the rest silently. A frontmatter block closed with four
// dashes is cut from the body by StripFrontmatter, which closes on a `---`
// PREFIX, while Fields wants the delimiter exactly and so reads no fields at all.
// And a heading spelled in another case is not the title the redactor looked for.
//
// In each case the field travels and the manifest still asserts it was refused,
// which is the one thing the manifest exists not to do. A floor a file can
// quietly walk through is a disclosure, not a gate — so a file that still
// carries an excluded shape after redaction refuses the run and names the shape.
func verifyRedaction(rel, original, redacted string, keys, headings map[string]bool) error {
	if len(keys) > 0 {
		for _, dup := range frontmatter.Duplicates(strings.Split(original, "\n")) {
			if keys[dup.Key] {
				return fmt.Errorf("reading: %s declares the excluded key %q more than once (line %d); "+
					"only the first occurrence is redactable, so the rest would travel", rel, dup.Key, dup.Line)
			}
		}
		if line, key, ok := excludedKeyInFirstBlock(redacted, keys); ok {
			return fmt.Errorf("reading: %s still carries the excluded key %q at line %d after redaction; "+
				"the frontmatter block is not closed the way the field reader expects it", rel, key, line)
		}
	}
	if len(headings) == 0 {
		return nil
	}
	for i, line := range strings.Split(redacted, "\n") {
		m := headingLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		for title := range headings {
			if strings.EqualFold(m[2], title) {
				return fmt.Errorf("reading: %s still carries the excluded heading %q at line %d after "+
					"redaction; the floor names %q, and a heading is excluded however it is spelled",
					rel, m[2], i+1, title)
			}
		}
	}
	return nil
}

// excludedKeyInFirstBlock reports an excluded key sitting at column 0 inside the
// document's first delimiter-fenced region.
//
// Two loosenesses are deliberate, and each closes a shape the strict reading
// misses. The region is bounded by any line OPENING with three dashes, not by an
// exact `---`, because that is the rule the frontmatter stripper applies and the
// gap between the two rules is where a key survives. And the region is found
// wherever it starts rather than at line 0, because a preamble ahead of the
// block makes the field reader report nothing at all while the keys sit there
// in plain sight.
func excludedKeyInFirstBlock(doc string, keys map[string]bool) (int, string, bool) {
	lines := strings.Split(doc, "\n")
	open := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 {
			trimmed = strings.TrimSpace(frontmatter.TrimBOM(line))
		}
		if strings.HasPrefix(trimmed, "---") {
			open = i
			break
		}
	}
	if open < 0 {
		return 0, "", false
	}
	for i := open + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "---") {
			return 0, "", false
		}
		m := excludedKeyLineRe.FindStringSubmatch(lines[i])
		if m != nil && keys[m[1]] {
			return i + 1, m[1], true
		}
	}
	return 0, "", false
}

// sectionSpan is the half-open line range one heading OWNS: the heading itself
// through everything under it, ending at the next heading of the same level or
// shallower. Indices are 0-based over the whole document, frontmatter included.
//
// It is shared by the redactor and the projector because the two must agree
// about where a section ends. The section scan's own Body ends at the next
// heading of ANY level, which is right for rendering a page and wrong here: a
// `###` under a projected `##` would be dropped from the item while the redactor
// treated it as part of the section it belongs to, so a field would travel short
// and the manifest would hash the short version.
func sectionSpan(sections []site.Section, i, total int) (int, int) {
	start := sections[i].Line - 1
	for _, next := range sections[i+1:] {
		if next.Level > 0 && next.Level <= sections[i].Level {
			return start, next.Line - 1
		}
	}
	return start, total
}

// projectField extracts one named field from a record's text. Only a record is
// ever projected, and a record is markdown, so the same scope holds here.
func projectField(rel, doc, field string) (string, bool, error) {
	body, offset := site.StripFrontmatter(doc)
	sections, err := site.Sections(rel, body, offset)
	if err != nil {
		return "", false, fmt.Errorf("reading: projecting %s from %s: %w", field, rel, err)
	}
	lines := strings.Split(doc, "\n")
	for i, sec := range sections {
		if sec.Level < 2 || sec.Title != field {
			continue
		}
		start, end := sectionSpan(sections, i, len(lines))
		return trimBlankEdges(lines[min(start+1, len(lines)):min(end, len(lines))]), true, nil
	}
	fields := frontmatter.Fields(strings.Split(doc, "\n"))
	if f, ok := fields[field]; ok && !frontmatter.IsNull(f.Value) {
		return f.Value, true, nil
	}
	return "", false, nil
}
