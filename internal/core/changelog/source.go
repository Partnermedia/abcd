package changelog

import (
	"regexp"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
)

// maxSummaryRunes bounds the prose a record contributes to a cut. The cut is
// rendered to a terminal and serialised to a host as JSON, so one record whose
// first paragraph is a wall of text must not be able to swamp either. Runes, not
// bytes, because the bound exists for a reader.
const maxSummaryRunes = 400

// h1Re matches a level-one ATX heading — the record's title line. Deeper
// headings are section labels ("## Press Release"), never the record's name, so
// only `# ` counts.
var h1Re = regexp.MustCompile(`^#\s+(.+?)\s*$`)

// summarise extracts the source material a changelog composer writes prose
// FROM: what the record is called, and its opening paragraph.
//
// It lives beside the record reader rather than in the composer because both the
// preview and the ship read the SAME blob once; extracting this later would mean
// a second read of every record out of git, and two readers that could disagree
// about what a record says.
//
// The two record families are shaped differently and both must yield a title. An
// intent opens with an `# ` heading; an issue carries no heading at all, so its
// frontmatter slug names it. The final fallback is the record id, so a title is
// never empty — a changelog line with nothing to name its record by is worse
// than an ugly one.
//
// The summary is the first body paragraph that is not a heading, with blockquote
// markers stripped (the press-release convention wraps the opening paragraph in
// `>`), wrapped lines joined, and whitespace collapsed. Markdown emphasis is
// deliberately left intact: this is source material for a writer, not display
// text, and stripping it would lose the author's own emphasis.
func summarise(blob string, id string) (title, summary string) {
	lines := strings.Split(blob, "\n")
	fields := frontmatter.Fields(lines)
	body := lines[bodyStart(lines):]

	title = firstHeading(body)
	if title == "" {
		title = strings.Trim(fields["slug"].Value, `"'`)
	}
	if title == "" {
		title = id
	}
	return title, firstParagraph(body)
}

// bodyStart returns the index of the first line after the frontmatter block, or
// 0 when the document has none. The block is delimited by the first TWO `---`
// lines, exactly as internal/core/frontmatter reads it, so the two never
// disagree about where the body begins.
func bodyStart(lines []string) int {
	if len(lines) == 0 || strings.TrimRight(lines[0], " \t\r") != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t\r") == "---" {
			return i + 1
		}
	}
	return 0
}

// firstHeading returns the text of the first level-one heading in body, or "".
func firstHeading(body []string) string {
	for _, line := range body {
		if m := h1Re.FindStringSubmatch(strings.TrimRight(line, "\r")); m != nil {
			return m[1]
		}
	}
	return ""
}

// firstParagraph returns the first run of non-blank, non-heading body lines,
// joined into one collapsed and capped line.
func firstParagraph(body []string) string {
	var para []string
	for _, raw := range body {
		line := strings.TrimSpace(strings.TrimRight(raw, "\r"))
		line = strings.TrimSpace(strings.TrimLeft(line, ">"))
		if line == "" || strings.HasPrefix(line, "#") {
			if len(para) > 0 {
				break
			}
			continue
		}
		para = append(para, line)
	}
	return capRunes(strings.Join(strings.Fields(strings.Join(para, " ")), " "))
}

// capRunes truncates to maxSummaryRunes, marking the cut with an ellipsis so a reader
// can tell a bounded summary from a short one.
func capRunes(s string) string {
	runes := []rune(s)
	if len(runes) <= maxSummaryRunes {
		return s
	}
	return strings.TrimSpace(string(runes[:maxSummaryRunes-1])) + "…"
}
