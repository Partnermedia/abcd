package site

import (
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/lint"
)

// TestBuildRecordExportRefusesIDCollision proves a frontmatter-free store file
// whose derived id collides with a typed record's id is refused loudly, rather
// than silently overwriting that record in the export (index/derived are keyed by
// id, last write wins).
func TestBuildRecordExportRefusesIDCollision(t *testing.T) {
	graph := lint.RecordGraph{Nodes: []lint.RecordNode{
		{ID: "adr-1", Type: "adr", Title: "The first decision", Path: ".abcd/development/decisions/adrs/0001-x.md"},
	}}
	extra := []lint.RecordNode{
		{ID: "adr-1", Type: "principle", Title: "A colliding principle", Path: ".abcd/development/principles/adr-1.md"},
	}
	_, err := BuildRecordExport(t.TempDir(), "", graph, extra, History{Files: map[string]FileDates{}}, BuildStamp{}, RecordOpts{})
	if err == nil {
		t.Fatal("BuildRecordExport accepted a store file colliding with a typed record id")
	}
	if !strings.Contains(err.Error(), "adr-1") {
		t.Errorf("refusal does not name the colliding id: %v", err)
	}
}
