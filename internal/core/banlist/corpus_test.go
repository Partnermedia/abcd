package banlist

// The shared fixture corpora and their probe tables. Every file under testdata/ is
// read by BOTH readers of the private-banlist format — the Go parser
// (private_test.go) and the committed shell hook (hook_test.go) — and both drive
// the SAME tables below. That is what makes "the hook and the Go parser agree on
// the format" a checked claim rather than a hope: a divergence in key derivation,
// in whitespace handling, in comment handling, or in case folding fails one side
// against the other's expectation.
//
// There are three corpora because the format has three states a reader can be
// wrong about independently:
//
//   - parse-corpus.txt          a KEYED store (line 1 declares the format)
//   - parse-corpus-legacy.txt   a LEGACY store (no declaration: whole-line patterns)
//   - parse-corpus-malformed.txt a keyed store with one of every unusable line class
//
// Every value is reserved for documentation (RFC 5737 / 3849 / 2606 / 7042) or
// derived from the persona registry. Nothing here names a real machine, network,
// or organisation.

// corpusMustBlock is the keyed corpus's must-block half: content that matches
// exactly one entry, and the key that entry resolves to.
var corpusMustBlock = []struct {
	name string
	key  string
	text string
}{
	{"name", "widget-partner", "the widgetworks deal closes friday\n"},
	{"hostname", "lab-host", "reached alice-laptop.example.com at noon\n"},
	{"ipv4", "lab-ip", "the box answers on 192.0.2.17\n"},
	{"cidr", "lab-cidr", "route 203.0.113.0/24 over the tunnel\n"},
	{"mac", "lab-mac", "nic 00:00:5E:00:53:1A came up\n"},
	{"ipv6", "lab-v6", "and 2001:db8::5 replied\n"},
	{"indented entry", "indented-key", "carol-server.test is the build box\n"},
	{"tab-separated entry", "tab-key", "widgetco.example renewed\n"},
	{"case-insensitive", "lab-host", "ALICE-LAPTOP.EXAMPLE.COM\n"}, // abcd-audit:allow
}

// corpusMustPass is the keyed corpus's must-pass half (guards-prove-themselves: a
// guard proven only against forbidden input may simply refuse everything). Each
// case is one distinct facet: a near miss on a name, a neighbouring address, an
// unbanned persona host, a comment line's text, and reserved-range prose.
var corpusMustPass = []struct {
	name string
	text string
}{
	{"near miss on a name", "the widgetwork prototype ships\n"},
	{"neighbouring address", "the box answers on 192.0.2.18\n"},
	{"unbanned persona host", "bob-desktop.example.org is idle\n"},
	{"comment text is not a pattern", "# fixture corpus mentions nothing real\n"},
	{"reserved-range prose", "use 198.51.100.0/24 in examples\n"},
	{"the format declaration is not a pattern", "abcd-banlist keyed stores\n"},
}

// corpusKeys is the key sequence both readers must derive from the keyed corpus,
// in file order. Every key is declared in the file: a keyed store has no synthetic
// keys, because a line that does not parse has no key to name.
var corpusKeys = []string{
	"widget-partner",
	"lab-host",
	"lab-ip",
	"lab-cidr",
	"lab-mac",
	"lab-v6",
	"indented-key",
	"tab-key",
}

// legacyMustBlock is the legacy corpus's must-block half. A legacy line is a
// WHOLE-LINE pattern, so the text that matches line 9 is the whole two-field line
// — and the key named is always synthetic, never the line's first field.
var legacyMustBlock = []struct {
	name string
	key  string
	text string
}{
	{"bare pattern", "entry-8", "partnerco.example signed\n"},
	{"multi-word line is one pattern", "entry-9", "log: widget-partner   widgetworks appears\n"},
	{"single-space multi-word line", "entry-10", "lab-host alice-laptop.example.com was reached\n"},
}

// legacyMustPass is the legacy corpus's must-pass half, and it is the detector for
// the key-splitting leak: if a reader splits a legacy line into key + pattern, the
// first two cases below start matching (the remainder-only semantics) and the
// refusal starts printing part of the line as a "key".
var legacyMustPass = []struct {
	name string
	text string
}{
	{"key-shaped first word is not a key", "widget-partner is a handle\n"},
	{"the remainder alone does not match", "the widgetworks deal closes friday\n"},
	{"the remainder of a one-space line alone does not match", "alice-laptop.example.com answered\n"}, // abcd-audit:allow
	{"unrelated prose", "nothing sensitive here\n"},
}

// legacyKeys is the key sequence both readers must derive from the legacy corpus.
// Every key is synthetic: no part of a legacy line is ever read as a key, so no
// part of one can ever be printed.
var legacyKeys = []string{"entry-8", "entry-9", "entry-10"}

// legacyFirstFields are the first whitespace-delimited fields of the legacy
// corpus's multi-field lines. They are key-shaped on purpose: no reader may ever
// print one, because on a legacy line that text is part of the pattern — which on
// a real store is the secret.
var legacyFirstFields = []string{"widget-partner", "lab-host"}

// malformedUnusableLines are the 1-based line numbers of the malformed corpus that
// no reader can use: two that do not parse in the keyed format (an entry indented
// with U+00A0 and one separated by U+000B — neither is an ASCII space or tab), one
// with no separator at all, and one that parses but whose pattern the enforcing
// engine refuses.
var malformedUnusableLines = []int{12, 13, 14, 15}

// malformedKeys are the keys the malformed corpus still yields: a line that does
// not parse has no key, so only the two parseable lines are listed.
var malformedKeys = []string{"widget-partner", "bad-pattern"}
