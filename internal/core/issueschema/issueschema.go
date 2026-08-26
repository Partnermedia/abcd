// Package issueschema is the issue record's frontmatter schema as DATA: the one
// list of properties issue.schema.json marks required, held where every gate that
// asks the question can read it.
//
// It exists for the same reason core/changelog holds the impact enum: the writer
// and reader of the ledger (core/capture) and the lint that GATES the committed
// ledger (core/lint) must agree about what a well-formed record carries, and two
// hand-kept copies of a required-property list drift the moment one side gains a
// field. It is a leaf — no imports, no filesystem, no transport — because
// core/capture's own tests import core/lint, so a lint that imported capture back
// would be an import cycle in test.
package issueschema

// Required is every property the issue schema marks required, in the order a
// record writes them. A record missing one is not a lax record: the ledger reader
// refuses it and skips it, so it goes silently invisible to every capture surface
// while still sitting in the ledger — which is what record_schema reports.
var Required = []string{
	"schema_version", "id", "slug", "severity", "category", "source", "found_during",
}

// RequiredStrings is Required minus schema_version — the required properties
// whose value is a string. schema_version carries the integer version and is
// checked against the readers this repository ships, so it is validated on its
// own before the string-typed properties are walked. It is a slice OF Required
// rather than a second literal, so the two can never disagree about the set.
var RequiredStrings = Required[1:]
