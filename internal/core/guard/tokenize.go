package guard

import (
	"fmt"
	"strings"
)

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
		segs    []segment
		toks    []string
		cur     []byte
		hasCur  bool
		chain   int
		pending []heredoc
		// lastList records that the previous operator was a list operator
		// (`&&`, `||`, `|`), whose newline is a line continuation rather than a
		// new command — `cd scratch &&\nrm -rf *` is one chain, not two.
		lastList bool
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
			lastList = false
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
			lastList = false
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
			lastList = false
			i = j + 1
		case c == ' ' || c == '\t' || c == '\r':
			flushToken()
			i++
		case c == '\n':
			flushSegment()
			i++
			// A pending heredoc body starts once the command LINE is complete,
			// and is DATA, not commands: writing a document that names a hazard
			// must never read as running one. A line ending in a list operator
			// is not complete, so the body waits — what follows it is still
			// command text.
			if len(pending) > 0 && !lastList {
				next, ok := skipHeredocBodies(line, i, pending)
				if !ok {
					return nil, fmt.Errorf("%w: unterminated here-document body", ErrUnparsableCommand)
				}
				i = next
				pending = nil
			}
			// lastList is NOT cleared here: a blank or comment-only line after a
			// list operator does not end the list, and every token-producing
			// branch clears the flag as soon as real content arrives.
			if !lastList {
				chain++
			}
		case c == '#' && !hasCur:
			// A comment starts only at a word boundary (POSIX): `url/#frag` is
			// part of the token, a bare `#` runs to the end of the line.
			for i < len(line) && line[i] != '\n' {
				i++
			}
		case c == '<' && strings.HasPrefix(line[i:], "<<<"):
			// A herestring, not a heredoc: its payload is an ordinary argument
			// token, so the operator is kept as plain token text.
			cur = append(cur, '<', '<', '<')
			hasCur = true
			lastList = false
			i += 3
		case c == '<' && strings.HasPrefix(line[i:], "<<"):
			// A heredoc redirection (`<<`, `<<-`) — but only when a delimiter
			// word follows. `$((1<<20))` is an arithmetic shift, and taking it
			// for a heredoc would swallow every later line as body text and
			// silently unguard them.
			hd, next, err := readHeredocDelim(line, i+2)
			if err != nil {
				return nil, err
			}
			// A real heredoc delimiter is never immediately followed by a bare
			// paren: its body and terminator line have to come first, so nothing
			// legitimate places `(` or `)` directly against the delimiter word
			// with no separator. `$((expr<<ident))` produces exactly this shape
			// -- readHeredocDelim reads the shift's right-hand identifier as if
			// it were a delimiter and stops at the arithmetic expression's own
			// closing paren. Catching this at classification time (rather than
			// only an unterminated body, below) matters: without it, an
			// attacker can supply a later line that happens to equal the
			// misread "delimiter" and have it swallow real commands with no
			// error at all.
			followedByParen := next < len(line) && (line[next] == '(' || line[next] == ')')
			if !hd.quoted && (!isDelimStart(hd.delim) || followedByParen) {
				cur = append(cur, '<', '<')
				hasCur = true
				lastList = false
				i += 2
				continue
			}
			flushToken()
			pending = append(pending, hd)
			i = next
		case c == '&' || c == '|' || c == ';' || c == '(' || c == ')':
			flushSegment()
			if (c == '&' || c == '|') && i+1 < len(line) && line[i+1] == c {
				lastList = true
				i += 2
				continue
			}
			// A single pipe continues the list across a newline; `;`, `&`, and
			// the grouping parens do not.
			lastList = c == '|'
			i++
		default:
			cur = append(cur, c)
			hasCur = true
			lastList = false
			i++
		}
	}
	flushSegment()
	return segs, nil
}

// heredoc is one pending here-document: the delimiter word that ends its body,
// and whether the `<<-` form allows the delimiter line to be tab-indented.
type heredoc struct {
	delim     string
	stripTabs bool
	// quoted records that the delimiter word carried quotes, which makes it a
	// delimiter beyond doubt however exotic it looks (`<<'---'`).
	quoted bool
}

// isDelimStart reports whether a word looks like a here-document delimiter: an
// unquoted one starts with a letter or an underscore, which is what separates
// `cat <<EOF` from the `<<` of an arithmetic shift.
func isDelimStart(delim string) bool {
	if delim == "" {
		return false
	}
	c := delim[0]
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// readHeredocDelim reads the delimiter word after a `<<` at pos, honouring the
// `<<-` variant and any quoting of the word itself (`<<'EOF'`, `<<"EOF"`), and
// returns the position just past it.
func readHeredocDelim(line string, pos int) (heredoc, int, error) {
	hd := heredoc{}
	if pos < len(line) && line[pos] == '-' {
		hd.stripTabs = true
		pos++
	}
	for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t') {
		pos++
	}
	var w []byte
	for pos < len(line) {
		c := line[pos]
		switch c {
		case '\'', '"':
			j := pos + 1
			for j < len(line) && line[j] != c {
				j++
			}
			if j >= len(line) {
				return heredoc{}, 0, fmt.Errorf("%w: unterminated quote in heredoc delimiter", ErrUnparsableCommand)
			}
			w = append(w, line[pos+1:j]...)
			hd.quoted = true
			pos = j + 1
			continue
		case '\\':
			if pos+1 >= len(line) {
				return heredoc{}, 0, fmt.Errorf("%w: trailing backslash", ErrUnparsableCommand)
			}
			w = append(w, line[pos+1])
			pos += 2
			continue
		case ' ', '\t', '\n', ';', '&', '|', '(', ')', '<', '>':
			hd.delim = string(w)
			return hd, pos, nil
		}
		w = append(w, c)
		pos++
	}
	hd.delim = string(w)
	return hd, pos, nil
}

// skipHeredocBodies consumes the body of every pending here-document, starting
// at pos (the first byte after the newline that ended the command line), and
// returns the position just past the last body. The second return is false if
// any body never finds its terminating delimiter line before the input ends —
// which is either a genuinely truncated heredoc, or a `<<` that isDelimStart
// mistook for one (an identifier-operand arithmetic shift, `$((1<<shift))`,
// reads as a delimiter word). Either way, silently consuming the remainder of
// the line as unchecked "body" text would swallow real commands with no
// signal; the caller turns a false result into ErrUnparsableCommand so the
// guard fails open LOUDLY on it instead of silently.
func skipHeredocBodies(line string, pos int, pending []heredoc) (int, bool) {
	for _, hd := range pending {
		found := false
		for pos < len(line) {
			end := pos
			for end < len(line) && line[end] != '\n' {
				end++
			}
			text := line[pos:end]
			if hd.stripTabs {
				text = strings.TrimLeft(text, "\t")
			}
			if end < len(line) {
				pos = end + 1
			} else {
				pos = end
			}
			if text == hd.delim {
				found = true
				break
			}
		}
		if !found {
			return pos, false
		}
	}
	return pos, true
}
