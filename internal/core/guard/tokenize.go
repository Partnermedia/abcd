package guard

import (
	"fmt"
	"strings"
	"unicode/utf8"
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
// The tokenizer stops at the token boundary: a command string carried as a DATA
// argument — `sh -c '<payload>'`, `eval '<payload>'`, `bash -lc "<payload>"`,
// `env -S<value>` — stays one opaque token here. Descending into that payload is
// the execute-a-string family's job, done ONCE in Check via expandPayloads
// (payload.go), never in this splitter — so a hazard hidden there is matched
// (iss-200), while an uninspectable payload takes the family's posture.
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
		case c == '>' || (c == '<' && !strings.HasPrefix(line[i:], "<<")):
			// A redirection operator (`>`, `>>`, `>|`, `>&`, `<`, `<>`, `<&`),
			// optionally prefixed by an fd digit that sits in cur. `<<`/`<<<`
			// are recognised above; process substitution `>(...)`/`<(...)` is
			// not a redirection and keeps its prior handling. A redirection
			// terminates the current word and its target is a filename or fd,
			// never a command in command position — so both the operator and
			// the target are dropped. Without this, gluing a redirection onto a
			// token (`git push --force>/dev/null`) mutated the flag token so
			// every blocker missed and the verdict was a silent allow, and a
			// leading redirection (`>/dev/null git push --force`) displaced the
			// command out of position and degraded a Tier-1 block to a warn.
			if i+1 < len(line) && line[i+1] == '(' {
				cur = append(cur, c)
				hasCur = true
				lastList = false
				i++
				break
			}
			opEnd := i + 1
			if c == '>' {
				if opEnd < len(line) && (line[opEnd] == '>' || line[opEnd] == '|' || line[opEnd] == '&') {
					opEnd++
				}
			} else if opEnd < len(line) && (line[opEnd] == '>' || line[opEnd] == '&') {
				opEnd++
			}
			// A pure-digit cur immediately before the operator is the fd prefix
			// (`2>`, `1>&2`), part of the redirection rather than a token; drop
			// it. Otherwise flush the real word the operator terminates.
			if hasCur && isAllDigits(cur) {
				cur = nil
				hasCur = false
			} else {
				flushToken()
			}
			i = skipRedirectTarget(line, opEnd)
			lastList = false
		case c == '&' && i+1 < len(line) && line[i+1] == '>':
			// bash's `&>` / `&>>`: redirect both stdout and stderr. It has to be
			// recognised BEFORE the list-operator case below, which would read
			// the leading `&` as a background/`&&` operator and call
			// flushSegment -- splitting one simple command into two segments and
			// dropping the command's own dangerous flags out of command
			// position, so every blocker missed and the verdict was a silent
			// allow (`git push &>/dev/null --force origin main`). Unlike `>`/`<`,
			// a digit before `&>` is NOT an fd prefix in bash (`f 2&>x` passes
			// `2` to `f` and still redirects both streams), so the preceding word
			// is a real token and is flushed, never dropped.
			opEnd := i + 2
			if opEnd < len(line) && line[opEnd] == '>' {
				opEnd++
			}
			flushToken()
			i = skipRedirectTarget(line, opEnd)
			lastList = false
		case c == '$' && i+1 < len(line) && line[i+1] == '\'':
			// bash ANSI-C quoting: $'...' contributes its escape-decoded body to
			// the SAME word, exactly as '...' contributes its raw body. Without
			// this the leading `$` fell through to the default word branch and
			// prefixed the token (`$'--force'` tokenised to `$--force`), so a
			// blocker naming `--force` missed — a silent allow of the very argv
			// bash hands the child. Quoting must not change argument semantics
			// (doc.go): `git push $'--force'` fires, like `git push '--force'`.
			decoded, next, err := readAnsiCQuote(line, i+2)
			if err != nil {
				return nil, err
			}
			cur = append(cur, decoded...)
			hasCur = true
			lastList = false
			i = next
		case c == '$' && i+1 < len(line) && line[i+1] == '"':
			// bash locale quoting: $"..." is a plain double-quoted word with the
			// `$` stripped (the translation is identity for a tokenizer). Skip the
			// `$` so the double-quote branch reads the string; the same silent
			// allow as $'...' otherwise.
			i++
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

// readAnsiCQuote decodes a bash ANSI-C `$'...'` body that begins at start (the
// byte just after the opening quote) and returns the decoded bytes together with
// the index just past the closing quote. Inside `$'...'` a backslash introduces
// an escape — so `\'` does not end the string — and the common escapes are
// resolved so an encoded spelling of a hazard (`$'\x2d\x2dforce'`) tokenises to
// the same bytes bash would hand the child (`--force`).
func readAnsiCQuote(line string, start int) ([]byte, int, error) {
	var out []byte
	for i := start; i < len(line); {
		switch {
		case line[i] == '\'':
			return out, i + 1, nil
		case line[i] == '\\' && i+1 < len(line):
			decoded, next := decodeAnsiCEscape(line, i+1)
			out = append(out, decoded...)
			i = next
		default:
			out = append(out, line[i])
			i++
		}
	}
	return nil, 0, fmt.Errorf("%w: unterminated $'' quote", ErrUnparsableCommand)
}

// decodeAnsiCEscape resolves one ANSI-C escape whose leading backslash has
// already been consumed; p indexes the character after the backslash. It returns
// the decoded bytes and the index just past the escape. An unrecognised escape
// keeps the backslash and the character, matching bash.
func decodeAnsiCEscape(line string, p int) ([]byte, int) {
	switch c := line[p]; c {
	case 'n':
		return []byte{'\n'}, p + 1
	case 't':
		return []byte{'\t'}, p + 1
	case 'r':
		return []byte{'\r'}, p + 1
	case 'a':
		return []byte{'\a'}, p + 1
	case 'b':
		return []byte{'\b'}, p + 1
	case 'f':
		return []byte{'\f'}, p + 1
	case 'v':
		return []byte{'\v'}, p + 1
	case 'e', 'E':
		return []byte{0x1b}, p + 1
	case '\\', '\'', '"', '?':
		return []byte{c}, p + 1
	case 'x':
		val, n := 0, 0
		for j := p + 1; j < len(line) && n < 2 && isHexDigit(line[j]); j++ {
			val = val*16 + hexValue(line[j])
			n++
		}
		if n == 0 {
			return []byte{c}, p + 1
		}
		return []byte{byte(val)}, p + 1 + n
	case '0', '1', '2', '3', '4', '5', '6', '7':
		val, n := 0, 0
		for j := p; j < len(line) && n < 3 && line[j] >= '0' && line[j] <= '7'; j++ {
			val = val*8 + int(line[j]-'0')
			n++
		}
		return []byte{byte(val)}, p + n
	case 'u', 'U':
		// bash \uHHHH (up to 4 hex) and \UHHHHHHHH (up to 8 hex): a Unicode code
		// point, UTF-8 encoded. Without these an ASCII hazard spelled as
		// `$'--force'` decoded to --force in the shell but not here, so a
		// Tier-1 blocker missed on the same bytes the \x and octal forms already
		// close.
		width := 4
		if c == 'U' {
			width = 8
		}
		val, n := 0, 0
		for j := p + 1; j < len(line) && n < width && isHexDigit(line[j]); j++ {
			val = val*16 + hexValue(line[j])
			n++
		}
		if n == 0 {
			return []byte{c}, p + 1
		}
		r := rune(val)
		if !utf8.ValidRune(r) {
			r = utf8.RuneError
		}
		return utf8.AppendRune(nil, r), p + 1 + n
	case 'c':
		// bash \cX: the control character for X (X with bit 6 cleared, uppercased).
		// `\c` at end of string is left literal.
		if p+1 >= len(line) {
			return []byte{c}, p + 1
		}
		x := line[p+1]
		if x >= 'a' && x <= 'z' {
			x -= 'a' - 'A'
		}
		return []byte{x & 0x1f}, p + 2
	default:
		return []byte{'\\', c}, p + 1
	}
}

// isHexDigit reports whether b is an ASCII hexadecimal digit.
func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// hexValue returns the value of an ASCII hexadecimal digit; the caller guarantees
// isHexDigit(b).
func hexValue(b byte) int {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0')
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10
	default:
		return int(b-'A') + 10
	}
}

// isAllDigits reports whether b is a non-empty run of ASCII digits — the shape
// of a file-descriptor prefix on a redirection (`2>`, `1>&2`).
func isAllDigits(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// skipRedirectTarget consumes optional whitespace and a single redirection
// target word starting at pos, returning the position just past it. The target
// of a redirection is a filename or fd, never a command in command position, so
// the guard drops it. Quotes and backslashes inside the target are honoured so
// an embedded space or operator does not end the target early.
func skipRedirectTarget(line string, pos int) int {
	for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t') {
		pos++
	}
	for pos < len(line) {
		c := line[pos]
		switch {
		case c == '\\':
			if pos+1 >= len(line) {
				return pos + 1
			}
			pos += 2
		case c == '\'':
			j := pos + 1
			for j < len(line) && line[j] != '\'' {
				j++
			}
			if j >= len(line) {
				return j
			}
			pos = j + 1
		case c == '"':
			j := pos + 1
			for j < len(line) {
				if line[j] == '\\' && j+1 < len(line) {
					j += 2
					continue
				}
				if line[j] == '"' {
					break
				}
				j++
			}
			if j >= len(line) {
				return j
			}
			pos = j + 1
		case c == ' ' || c == '\t' || c == '\n' || c == '&' || c == '|' ||
			c == ';' || c == '(' || c == ')' || c == '<' || c == '>':
			return pos
		default:
			pos++
		}
	}
	return pos
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
