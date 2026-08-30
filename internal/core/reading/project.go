package reading

import (
	"fmt"
	"path"
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
	for i, s := range sections {
		if s.Level == 0 || !headings[s.Title] {
			continue
		}
		end := len(lines)
		for _, next := range sections[i+1:] {
			if next.Level > 0 && next.Level <= s.Level {
				end = next.Line - 1
				break
			}
		}
		for j := s.Line - 1; j < end && j < len(lines); j++ {
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
	return strings.Join(kept, "\n"), nil
}

// projectField extracts one named field from a record's text. Only a record is
// ever projected, and a record is markdown, so the same scope holds here.
func projectField(rel, doc, field string) (string, bool, error) {
	body, offset := site.StripFrontmatter(doc)
	sections, err := site.Sections(rel, body, offset)
	if err != nil {
		return "", false, fmt.Errorf("reading: projecting %s from %s: %w", field, rel, err)
	}
	for _, s := range sections {
		if s.Level >= 2 && s.Title == field {
			return s.Body, true, nil
		}
	}
	fields := frontmatter.Fields(strings.Split(doc, "\n"))
	if f, ok := fields[field]; ok && !frontmatter.IsNull(f.Value) {
		return f.Value, true, nil
	}
	return "", false, nil
}
