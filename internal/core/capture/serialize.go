package capture

import (
	"fmt"
	"strconv"
	"strings"
)

// kv is one ordered frontmatter entry. val is a string, int, or []string.
type kv struct {
	key string
	val any
}

// yamlScalar encodes a scalar value as a safe YAML literal, mirroring
// _issue_lib._yaml_scalar. Strings are double-quoted with backslash/dquote
// escaping and reject any ASCII control char (< 0x20); ints render bare.
func yamlScalar(value any) (string, error) {
	switch v := value.(type) {
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case int:
		return strconv.Itoa(v), nil
	case string:
		for _, r := range v {
			if r < 0x20 {
				return "", fmt.Errorf("%w: control char U+%04X in scalar string %q",
					ErrMalformedFrontmatter, r, v)
			}
		}
		esc := strings.ReplaceAll(v, `\`, `\\`)
		esc = strings.ReplaceAll(esc, `"`, `\"`)
		return `"` + esc + `"`, nil
	default:
		return "", fmt.Errorf("%w: unsupported scalar type %T", ErrMalformedFrontmatter, value)
	}
}

// buildIssueText serialises an ordered field list + body into on-disk text,
// mirroring _issue_lib.build_issue_text: opening ---, one key: value line per
// entry, closing ---, exactly one blank separator, then body. No `...`
// end-marker. Empty list -> `key: []`; a list whose items are all abcd ids
// (itd-N/fn-N/iss-N) -> unquoted inline `[itd-4, fn-12]`; any other list ->
// per-item quoted inline. List items must all be strings (they are, by type).
func buildIssueText(fields []kv, body string) (string, error) {
	lines := []string{"---"}
	for _, f := range fields {
		switch v := f.val.(type) {
		case []string:
			switch {
			case len(v) == 0:
				lines = append(lines, f.key+": []")
			case allAbcdIDs(v):
				lines = append(lines, f.key+": ["+strings.Join(v, ", ")+"]")
			default:
				parts := make([]string, len(v))
				for i, item := range v {
					enc, err := yamlScalar(item)
					if err != nil {
						return "", err
					}
					parts[i] = enc
				}
				lines = append(lines, f.key+": ["+strings.Join(parts, ", ")+"]")
			}
		default:
			enc, err := yamlScalar(v)
			if err != nil {
				return "", err
			}
			lines = append(lines, f.key+": "+enc)
		}
	}
	lines = append(lines, "---")
	return strings.Join(lines, "\n") + "\n\n" + body, nil
}

func allAbcdIDs(items []string) bool {
	for _, it := range items {
		if !reAbcdListID.MatchString(it) {
			return false
		}
	}
	return len(items) > 0
}

// setScalarField sets `key: <yamlScalar(value)>` inside content's frontmatter
// block, mirroring _issue_lib.set_scalar_field. If the key already exists at
// the top level it is replaced in place (order preserved); otherwise the new
// line is inserted before the closing ---. Only string/int values are accepted.
func setScalarField(content, key string, value any) (string, error) {
	if !reScalarKey.MatchString(key) {
		return "", fmt.Errorf("%w: invalid key %q", ErrMalformedFrontmatter, key)
	}
	encoded, err := yamlScalar(value)
	if err != nil {
		return "", err
	}

	lines := splitKeepEnds(content)

	// Locate the opening delimiter — must be the first non-empty line.
	openIdx := -1
	for i, ln := range lines {
		stripped := strings.TrimRight(ln, "\r\n")
		if stripped == "" {
			continue
		}
		if stripped == "---" {
			openIdx = i
			break
		}
		return "", fmt.Errorf("%w: content has no frontmatter block", ErrMalformedFrontmatter)
	}
	if openIdx == -1 {
		return "", fmt.Errorf("%w: content has no frontmatter block", ErrMalformedFrontmatter)
	}

	// Locate the closing delimiter.
	closeIdx := -1
	for j := openIdx + 1; j < len(lines); j++ {
		ln := lines[j]
		if strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t") {
			continue
		}
		if strings.TrimRight(ln, "\r\n") == "---" {
			closeIdx = j
			break
		}
	}
	if closeIdx == -1 {
		return "", fmt.Errorf("%w: frontmatter not terminated", ErrMalformedFrontmatter)
	}

	eol := "\n"
	if strings.HasSuffix(lines[openIdx], "\r\n") {
		eol = "\r\n"
	}
	newLine := key + ": " + encoded + eol

	replaced := false
	for k := openIdx + 1; k < closeIdx; k++ {
		ln := lines[k]
		if strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t") {
			continue
		}
		bodyLn := strings.TrimRight(ln, "\r\n")
		if bodyLn == "" || strings.HasPrefix(bodyLn, "#") {
			continue
		}
		idx := strings.Index(bodyLn, ":")
		if idx < 0 {
			continue
		}
		if strings.TrimSpace(bodyLn[:idx]) == key {
			lines[k] = newLine
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines[:closeIdx], append([]string{newLine}, lines[closeIdx:]...)...)
	}
	return strings.Join(lines, ""), nil
}

// splitKeepEnds splits s into lines preserving their trailing newline(s),
// mirroring Python's str.splitlines(keepends=True) for \n and \r\n.
func splitKeepEnds(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
