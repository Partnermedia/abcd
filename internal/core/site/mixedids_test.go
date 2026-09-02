package site

import (
	"reflect"
	"testing"
)

// TestDevelopmentDeckOrdersMixedIdVintagesByRecency proves the development
// deck's newest-first order survives the ledger's two id vintages (adr-45): on
// one date a timestamp-numeric id is the most recent mint and leads, the
// ordinals follow in descending numeric order, and the date still dominates the
// id. Nothing here parses the id as anything but a number, so a 16-digit id
// and a 1-digit one are judged on the same axis.
func TestDevelopmentDeckOrdersMixedIdVintagesByRecency(t *testing.T) {
	nodes := []ExportNode{
		{ID: "itd-3", Date: "2026-08-22"},
		{ID: "itd-2608221126066632", Date: "2026-08-22"},
		{ID: "itd-200", Date: "2026-08-22"},
		{ID: "itd-10", Date: "2026-08-22"},
		{ID: "itd-1", Date: "2026-09-01"},
	}
	sortDevelopmentNodes(nodes)
	var got []string
	for _, n := range nodes {
		got = append(got, n.ID)
	}
	want := []string{"itd-1", "itd-2608221126066632", "itd-200", "itd-10", "itd-3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("development deck order:\n got %v\nwant %v", got, want)
	}
}
