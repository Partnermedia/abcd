package guard

import "fmt"

// segment is one command in command position: the tokens of a single simple
// command, plus the index of the logical line (chain) it belongs to. Segments in
// the same chain are separated by `&&`, `;`, `||`, `|`, `&`, or a subshell
// parenthesis; a newline starts a new chain. The chain is what lets an entry
// require the cd-chain structure (`cd scratch && rm -rf *`) without matching an
// unrelated `rm` on the next line.
type segment struct {
	tokens []string
	chain  int
}

// tokenize splits a candidate command line into command-position segments,
// honouring shell quoting: single quotes are literal, double quotes take the
// POSIX backslash escapes, and a backslash outside quotes escapes the next
// character. Operators and comments are recognised OUTSIDE quotes only, which is
// exactly what keeps a hazard named inside a quoted argument from ever reaching
// command position.
//
// v1 GAP (documented, not silent): the tokenizer stops at the token boundary. A
// command string carried as a DATA argument — `sh -c '<payload>'`, `eval
// '<payload>'`, `bash -lc "<payload>"` — stays one opaque token and its payload
// is never parsed, so a hazard hidden there is not matched. Descending into
// those payloads is out of scope for v1 (spc-16 "Out of scope"); it is stated in
// the verb's reference doc and the registry README rather than silently assumed.
func tokenize(line string) ([]segment, error) {
	var (
		segs   []segment
		toks   []string
		cur    []byte
		hasCur bool
		chain  int
	)
	flushToken := func() {
		if hasCur {
			toks = append(toks, string(cur))
			cur = nil
			hasCur = false
		}
	}
	flushSegment := func() {
		flushToken()
		if len(toks) > 0 {
			segs = append(segs, segment{tokens: toks, chain: chain})
			toks = nil
		}
	}

	for i := 0; i < len(line); {
		c := line[i]
		switch {
		case c == '\\':
			if i+1 >= len(line) {
				return nil, fmt.Errorf("%w: trailing backslash", ErrUnparsableCommand)
			}
			if line[i+1] == '\n' {
				// Line continuation: the newline is removed, the chain continues.
				i += 2
				continue
			}
			cur = append(cur, line[i+1])
			hasCur = true
			i += 2
		case c == '\'':
			j := i + 1
			for j < len(line) && line[j] != '\'' {
				j++
			}
			if j >= len(line) {
				return nil, fmt.Errorf("%w: unterminated single quote", ErrUnparsableCommand)
			}
			cur = append(cur, line[i+1:j]...)
			hasCur = true
			i = j + 1
		case c == '"':
			j := i + 1
			closed := false
			for j < len(line) {
				if line[j] == '\\' && j+1 < len(line) {
					switch line[j+1] {
					case '"', '\\', '$', '`':
						cur = append(cur, line[j+1])
					case '\n':
						// Line continuation inside double quotes: both dropped.
					default:
						// Backslash is literal before any other character.
						cur = append(cur, '\\', line[j+1])
					}
					j += 2
					continue
				}
				if line[j] == '"' {
					closed = true
					break
				}
				cur = append(cur, line[j])
				j++
			}
			if !closed {
				return nil, fmt.Errorf("%w: unterminated double quote", ErrUnparsableCommand)
			}
			hasCur = true
			i = j + 1
		case c == ' ' || c == '\t' || c == '\r':
			flushToken()
			i++
		case c == '\n':
			flushSegment()
			chain++
			i++
		case c == '#' && !hasCur:
			// A comment starts only at a word boundary (POSIX): `url/#frag` is
			// part of the token, a bare `#` runs to the end of the line.
			for i < len(line) && line[i] != '\n' {
				i++
			}
		case c == '&' || c == '|' || c == ';' || c == '(' || c == ')':
			flushSegment()
			if (c == '&' || c == '|') && i+1 < len(line) && line[i+1] == c {
				i += 2
				continue
			}
			i++
		default:
			cur = append(cur, c)
			hasCur = true
			i++
		}
	}
	flushSegment()
	return segs, nil
}
