// Package frontmatter is abcd's shared markdown-frontmatter line scanner. It is
// deliberately a line scanner, not a YAML parser: it reads only the top-level
// keys of the leading `---`…`---` block, first key wins, and pulls in zero
// dependencies. It exists so its consumers (internal/core/spec,
// internal/core/intent, and record-lint's top-level frontmatter checks) share ONE
// copy of this primitive rather than each keeping a private replica.
//
// It is transport-agnostic: no stdout, no os.Exit, no filesystem access — the
// caller supplies the file's lines and decides what the fields mean.
package frontmatter

import (
	"regexp"
	"strings"
)

// keyRe matches a top-level frontmatter key (column 0, no indentation).
var keyRe = regexp.MustCompile(`^([A-Za-z0-9_]+):(.*)$`)

// Field is a frontmatter key's value and its 1-based source line.
type Field struct {
	Value string
	Line  int
}

// Fields returns the top-level keys of the leading frontmatter block (the block
// between the first two `---` lines). Nested keys and list items are ignored,
// and the first occurrence of a key wins. An input whose first line is not `---`
// (or is empty) yields no fields.
func Fields(lines []string) map[string]Field {
	fields := map[string]Field{}
	// A delimiter line may carry trailing whitespace ("--- "); trim spaces/tabs/CR
	// before comparing, so a trailing-space closing delimiter is still seen as the
	// close and body lines after it do not leak in as fields.
	if len(lines) == 0 || strings.TrimRight(lines[0], " \t\r") != "---" {
		return fields
	}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.TrimRight(line, " \t") == "---" {
			break
		}
		m := keyRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := m[1]
		if _, exists := fields[key]; !exists {
			fields[key] = Field{Value: strings.TrimSpace(m[2]), Line: i + 1}
		}
	}
	return fields
}

// IsNull treats an empty value and the YAML nulls ""/"null"/"~" as null.
func IsNull(v string) bool {
	return v == "" || v == "null" || v == "~"
}
