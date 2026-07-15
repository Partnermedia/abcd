package capture

import (
	"fmt"
	"strconv"
	"strings"
)

// parseFrontmatterAndBody splits text into a frontmatter map and a body,
// mirroring _issue_lib._parse_text_with_body. The text MUST start with an
// opening --- line; the next un-indented --- closes the block. At most one
// leading blank line after the closing delimiter is stripped from the body.
//
// The frontmatter is parsed with a restricted YAML subset matching what
// buildIssueText/setScalarField emit: top-level `key: value` scalars, inline
// lists (`[]`, `[itd-4, fn-12]`, `["a", "b"]`), and a single level of nested
// object (used only by the optional resolved_by field). Values decode to
// string, int, []string, or map[string]any.
func parseFrontmatterAndBody(text string) (map[string]any, string, error) {
	if !strings.HasPrefix(text, "---\n") && !strings.HasPrefix(text, "---\r\n") {
		return nil, "", fmt.Errorf("%w: frontmatter must start with '---' on the first line", ErrMalformedFrontmatter)
	}
	lines := splitKeepEnds(text)
	closeIdx := -1
	for i := 1; i < len(lines); i++ {
		ln := lines[i]
		if strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t") {
			continue
		}
		if strings.TrimRight(ln, "\r\n") == "---" {
			closeIdx = i
			break
		}
	}
	if closeIdx == -1 {
		return nil, "", fmt.Errorf("%w: frontmatter not terminated: missing closing '---'", ErrMalformedFrontmatter)
	}

	body := strings.Join(lines[closeIdx+1:], "")
	switch {
	case strings.HasPrefix(body, "\r\n"):
		body = body[2:]
	case strings.HasPrefix(body, "\n"):
		body = body[1:]
	}

	fm, err := parseFrontmatterBlock(lines[1:closeIdx])
	if err != nil {
		return nil, "", err
	}
	return fm, body, nil
}

// parseFrontmatterBlock parses the interior lines of a frontmatter block.
func parseFrontmatterBlock(lines []string) (map[string]any, error) {
	fm := map[string]any{}
	i := 0
	for i < len(lines) {
		raw := strings.TrimRight(lines[i], "\r\n")
		if strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			i++
			continue
		}
		if strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t") {
			return nil, fmt.Errorf("%w: unexpected indented line %q", ErrMalformedFrontmatter, raw)
		}
		idx := strings.Index(raw, ":")
		if idx < 0 {
			return nil, fmt.Errorf("%w: line is not key: value %q", ErrMalformedFrontmatter, raw)
		}
		key := strings.TrimSpace(raw[:idx])
		rest := strings.TrimSpace(raw[idx+1:])
		if key == "" {
			return nil, fmt.Errorf("%w: empty key in %q", ErrMalformedFrontmatter, raw)
		}
		if rest == "" {
			// Nested one-level object: consume following indented lines.
			sub := map[string]any{}
			i++
			for i < len(lines) {
				subRaw := strings.TrimRight(lines[i], "\r\n")
				if !strings.HasPrefix(subRaw, " ") && !strings.HasPrefix(subRaw, "\t") {
					break
				}
				subTrim := strings.TrimSpace(subRaw)
				if subTrim == "" {
					i++
					continue
				}
				sidx := strings.Index(subTrim, ":")
				if sidx < 0 {
					return nil, fmt.Errorf("%w: nested line is not key: value %q", ErrMalformedFrontmatter, subRaw)
				}
				sval, err := parseScalarOrList(strings.TrimSpace(subTrim[sidx+1:]))
				if err != nil {
					return nil, err
				}
				sub[strings.TrimSpace(subTrim[:sidx])] = sval
				i++
			}
			fm[key] = sub
			continue
		}
		val, err := parseScalarOrList(rest)
		if err != nil {
			return nil, err
		}
		fm[key] = val
		i++
	}
	return fm, nil
}

// parseScalarOrList decodes one YAML value into string, int, or []string.
func parseScalarOrList(s string) (any, error) {
	if s == "[]" {
		return []string{}, nil
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := strings.TrimSpace(s[1 : len(s)-1])
		if inner == "" {
			return []string{}, nil
		}
		var out []string
		for _, part := range splitInlineListItems(inner) {
			item := strings.TrimSpace(part)
			dec, err := decodeScalar(item)
			if err != nil {
				return nil, err
			}
			str, ok := dec.(string)
			if !ok {
				str = fmt.Sprint(dec)
			}
			out = append(out, str)
		}
		return out, nil
	}
	return decodeScalar(s)
}

// splitInlineListItems splits the interior of an inline list on top-level
// commas, honouring the double-quoting and backslash escaping that yamlScalar
// emits: a comma inside a quoted item (or an escaped comma/quote) is not a
// separator. Keeping the tokenizer symmetric with the serializer is what lets a
// quoted item containing a comma — e.g. synthesis_clusters: ["design review,
// session 3"] — round-trip faithfully instead of being split mid-item with
// stray quote characters left behind.
func splitInlineListItems(inner string) []string {
	var items []string
	var cur strings.Builder
	inQuote := false
	esc := false
	for _, r := range inner {
		switch {
		case esc:
			cur.WriteRune(r)
			esc = false
		case r == '\\':
			cur.WriteRune(r)
			esc = true
		case r == '"':
			cur.WriteRune(r)
			inQuote = !inQuote
		case r == ',' && !inQuote:
			items = append(items, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	items = append(items, cur.String())
	return items
}

// decodeScalar decodes a single non-list scalar token.
func decodeScalar(s string) (any, error) {
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) && len(s) >= 2 {
		return unquote(s[1 : len(s)-1]), nil
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n, nil
	}
	// Bare token (unquoted string, e.g. an abcd id or a legacy value).
	return s, nil
}

// unquote reverses yamlScalar's backslash + dquote escaping.
func unquote(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if esc {
			b.WriteRune(r)
			esc = false
			continue
		}
		if r == '\\' {
			esc = true
			continue
		}
		b.WriteRune(r)
	}
	if esc {
		b.WriteRune('\\')
	}
	return b.String()
}
