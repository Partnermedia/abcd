package audit

import "encoding/json"

// Serializer turns a Result into bytes for a machine consumer. It is the
// output-serializer seam: the JSON serializer ships now; a SARIF serializer is a
// later, additive implementation of the same interface (itd-85 P3), not a change
// to the engine.
type Serializer interface {
	Serialize(res Result) ([]byte, error)
}

// JSONSerializer emits abcd's native compact-but-indented JSON with stable rule
// ids. The top-level shape is { "findings": [...] } — always a present array,
// never null, so a clean repo emits { "findings": [] } exactly.
type JSONSerializer struct{}

func (JSONSerializer) Serialize(res Result) ([]byte, error) {
	// Guarantee present empty arrays rather than JSON null: the command doc
	// promises { findings, skipped }, so neither key may be null or absent.
	if res.Findings == nil {
		res.Findings = []Finding{}
	}
	if res.Skipped == nil {
		res.Skipped = []string{}
	}
	return json.MarshalIndent(res, "", "  ")
}
