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

// subKind names what opened a command-substitution frame, so a close is matched
// against the right opener: `)` closes a `$( … )`, a backtick closes a backtick
// span. Both delimiters of a backtick span are the SAME character, which is why
// the open/close decision is taken from the frame stack rather than the byte.
type subKind int

const (
	kindParen subKind = iota
	kindBacktick
)

// subFrame is one level of command substitution. A substitution RUNS what is
// inside it, so its content is command position in its own right — but the
// command CONTAINING it does not end at the boundary: `rm $(true) -rf *` is
// still a recursive force delete, and a force push written with a backticked
// substitution in front of `--force` is still a force push.
// Flushing the enclosing segment at the boundary (what the tokenizer did before
// frames existed) threw away every flag and operand that followed, so the
// registry entry looked armed and never fired.
//
// The frame therefore parks the enclosing command's tokens while the
// substitution accumulates its own, and hands the finished segments back when it
// closes, so the enclosing segment resumes where it left off.
type subFrame struct {
	kind subKind
	// toks is the segment currently accumulating in this frame.
	toks []string
	// out is the segments this frame has finished, in source order.
	out []segment
	// pend is segments produced by substitutions sitting INSIDE the segment that
	// is still accumulating. They are held until that segment flushes, so a
	// substitution never lands ahead of the command it is written inside — the
	// order `precededByCD` reads to recognise a cd chain.
	pend []segment
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
		cur     []byte
		hasCur  bool
		chain   int
		pending []heredoc
		// lastList records that the previous operator was a list operator
		// (`&&`, `||`, `|`), whose newline is a line continuation rather than a
		// new command — `cd scratch &&\nrm -rf *` is one chain, not two.
		lastList bool
	)
	// frames is the substitution stack. frames[0] is the command line itself; a
	// `$( … )` or a backtick span pushes one, and its close pops it back.
	frames := []*subFrame{{}}
	top := func() *subFrame { return frames[len(frames)-1] }
	flushToken := func() {
		if hasCur {
			f := top()
			f.toks = append(f.toks, string(cur))
			cur = nil
			hasCur = false
		}
	}
	flushSegment := func() {
		flushToken()
		f := top()
		if len(f.toks) > 0 {
			f.out = append(f.out, segment{tokens: f.toks, chain: chain})
			f.toks = nil
		}
		if len(f.pend) > 0 {
			f.out = append(f.out, f.pend...)
			f.pend = nil
		}
	}
	// openFrame returns the index of the innermost open frame of kind k, or -1.
	openFrame := func(k subKind) int {
		for i := len(frames) - 1; i > 0; i-- {
			if frames[i].kind == k {
				return i
			}
		}
		return -1
	}
	openSub := func(k subKind) {
		// The token in progress belongs to the ENCLOSING command (the `$` of a
		// `$(`, or whatever the substitution is glued onto), so it is finalised
		// there before the substitution starts accumulating its own.
		flushToken()
		frames = append(frames, &subFrame{kind: k})
	}
	// closeSub finishes the innermost frame of kind k — emitting the
	// substitution's own content as its own segments, which is what makes a
	// hazard hidden inside a substitution matchable — and hands those segments to
	// the frame enclosing it, which then resumes accumulating.
	//
	// Imbalance degrades, it never panics: a close with no matching open falls
	// back to the plain segment flush the tokenizer did before frames existed,
	// and a close that skips over frames of the other kind unwinds them rather
	// than stranding their content. Unbalanced parens and backticks are NOT a
	// parse error here — a real shell would reject them itself, and this guard's
	// job is to refuse to misparse into an allow, not to validate shell grammar.
	// Only genuinely unterminated quotes and heredocs, which change what the rest
	// of the input MEANS, return ErrUnparsableCommand.
	closeSub := func(k subKind) {
		depth := openFrame(k)
		if depth < 0 {
			flushSegment()
			return
		}
		for len(frames) > depth {
			flushSegment()
			f := frames[len(frames)-1]
			frames = frames[:len(frames)-1]
			p := top()
			p.pend = append(p.pend, f.out...)
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
				i = skipHeredocBodies(line, i, pending)
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
			if !hd.quoted && !isDelimStart(hd.delim) {
				cur = append(cur, '<', '<')
				hasCur = true
				lastList = false
				i += 2
				continue
			}
			flushToken()
			pending = append(pending, hd)
			i = next
		case c == '`':
			// A backtick command substitution RUNS what is inside it, so what is
			// inside it is command position: a backticked find in an rm's argument
			// list is a find that executes. Before iss-148 a backtick was an
			// operator nowhere in this switch and fell through to the default as
			// ordinary token text, so nothing inside one was ever seen.
			//
			// Both delimiters are the SAME character, so which one this is comes
			// from the stack, not from the byte: a backtick with a backtick frame
			// already open closes it, any other backtick opens one.
			//
			// The scope is parity with `$( … )`, not POSIX completeness: neither
			// form is followed inside double quotes, where the quoting branch
			// above consumes the whole span as one token.
			if openFrame(kindBacktick) > 0 {
				closeSub(kindBacktick)
			} else {
				openSub(kindBacktick)
			}
			lastList = false
			i++
		case c == '(':
			// `$( … )` is a substitution INSIDE a command, so the command survives
			// it. A BARE `(` is a subshell group, which is a command of its own —
			// and it genuinely ends the token run before it, which is what keeps
			// the body of a function definition (`f() { rm -rf *; }`) in command
			// position instead of gluing it onto the function's name.
			if hasCur && cur[len(cur)-1] == '$' {
				openSub(kindParen)
			} else {
				flushSegment()
			}
			lastList = false
			i++
		case c == ')':
			closeSub(kindParen)
			lastList = false
			i++
		case c == '&' || c == '|' || c == ';':
			// A list operator ends a command for real, so it gets no frame
			// treatment: what follows it is a different command, not the rest of
			// this one.
			flushSegment()
			if i+1 < len(line) && line[i+1] == c {
				lastList = true
				i += 2
				continue
			}
			// A single pipe continues the list across a newline; `;` and `&` do
			// not.
			lastList = c == '|'
			i++
		default:
			cur = append(cur, c)
			hasCur = true
			lastList = false
			i++
		}
	}
	// A substitution whose close never arrived is unwound rather than dropped: a
	// hazard inside an unbalanced span still executes in the shell that accepts
	// it, so it must still reach command position here.
	for len(frames) > 1 {
		closeSub(top().kind)
	}
	flushSegment()
	return frames[0].out, nil
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
// returns the position just past the last body. An unterminated body swallows
// the rest of the input — exactly as a shell would treat it.
func skipHeredocBodies(line string, pos int, pending []heredoc) int {
	for _, hd := range pending {
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
				break
			}
		}
	}
	return pos
}
