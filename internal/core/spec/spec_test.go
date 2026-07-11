package spec

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile writes content to root/rel, creating parent directories. Shared by
// both test files in this package.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{"good", Spec{ID: "spc-1", Slug: "thing", Intent: "itd-9"}, false},
		{"bad id", Spec{ID: "spec-1", Intent: "itd-9"}, true},
		{"empty id", Spec{ID: "", Intent: "itd-9"}, true},
		{"bad intent", Spec{ID: "spc-1", Intent: "itd-x"}, true},
		{"empty intent", Spec{ID: "spc-1", Intent: ""}, true},
		{"traversal id", Spec{ID: "spc-../../etc", Intent: "itd-9"}, true},
		{"traversal intent", Spec{ID: "spc-1", Intent: "itd-../../etc"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate(%+v) err = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
		})
	}
}

func TestStoreLookupAndByIntent(t *testing.T) {
	store := Store{Specs: []Spec{
		{ID: "spc-1", Slug: "a", Intent: "itd-9", Status: StatusOpen},
		{ID: "spc-2", Slug: "b", Intent: "itd-12", Status: StatusClosed},
	}}

	if s, ok := store.Lookup("spc-2"); !ok || s.Intent != "itd-12" {
		t.Fatalf("Lookup(spc-2) = %+v, %v", s, ok)
	}
	if _, ok := store.Lookup("spc-99"); ok {
		t.Fatal("Lookup(spc-99) unexpectedly found")
	}
	if s, ok := store.ByIntent("itd-9"); !ok || s.ID != "spc-1" {
		t.Fatalf("ByIntent(itd-9) = %+v, %v", s, ok)
	}
	if _, ok := store.ByIntent("itd-77"); ok {
		t.Fatal("ByIntent(itd-77) unexpectedly found")
	}
}
