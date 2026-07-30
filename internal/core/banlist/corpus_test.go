package banlist

// The shared fixture corpus and its two probe halves. testdata/parse-corpus.txt is
// read by BOTH readers of the private-banlist format — the Go parser (parse_test.go)
// and the committed shell hook (hook_test.go) — and both drive the SAME probe
// tables below. That is what makes "the hook and the Go parser agree on the format"
// a checked claim rather than a hope: a divergence in key derivation, in comment
// handling, or in case folding fails one side against the other's expectation.
//
// Every value is reserved for documentation (RFC 5737 / 3849 / 2606 / 7042) or
// derived from the persona registry. Nothing here names a real machine, network,
// or organisation.

// corpusMustBlock is the must-block half: content that matches exactly one corpus
// entry, and the key that entry resolves to.
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
	{"case-insensitive", "lab-host", "ALICE-LAPTOP.EXAMPLE.COM\n"},
	{"legacy bare pattern", "entry-19", "partnerco.example signed\n"},
}

// corpusMustPass is the must-pass half (guards-prove-themselves: a guard proven
// only against forbidden input may simply refuse everything). Each case is one
// distinct facet: a near miss on a name, a neighbouring address, an unbanned
// persona host, a comment line's text, and reserved-range prose.
var corpusMustPass = []struct {
	name string
	text string
}{
	{"near miss on a name", "the widgetwork prototype ships\n"},
	{"neighbouring address", "the box answers on 192.0.2.18\n"},
	{"unbanned persona host", "bob-desktop.example.org is idle\n"},
	{"comment text is not a pattern", "# fixture corpus mentions nothing real\n"},
	{"reserved-range prose", "use 198.51.100.0/24 in examples\n"},
}

// corpusKeys is the key sequence both readers must derive from the corpus, in
// file order. The last entry is the legacy bare-pattern line: no key column, so
// it carries the synthetic key naming its line number.
var corpusKeys = []string{
	"widget-partner",
	"lab-host",
	"lab-ip",
	"lab-cidr",
	"lab-mac",
	"lab-v6",
	"indented-key",
	"entry-19",
}
