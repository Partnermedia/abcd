package frontmatter

import "testing"

func TestFieldsReadsLeadingBlock(t *testing.T) {
	lines := []string{
		"---",
		"id: itd-9",
		"slug: my-thing",
		"spec_id: null",
		"---",
		"# Title",
		"key: not-frontmatter",
	}
	fields := Fields(lines)
	if got := fields["id"]; got.Value != "itd-9" || got.Line != 2 {
		t.Fatalf("id = %+v, want {itd-9 2}", got)
	}
	if got := fields["slug"]; got.Value != "my-thing" || got.Line != 3 {
		t.Fatalf("slug = %+v, want {my-thing 3}", got)
	}
	if _, ok := fields["key"]; ok {
		t.Fatal("a key past the closing --- must not be read")
	}
}

func TestFieldsFirstKeyWins(t *testing.T) {
	lines := []string{"---", "kind: standalone", "kind: discipline", "---"}
	if got := Fields(lines)["kind"]; got.Value != "standalone" || got.Line != 2 {
		t.Fatalf("kind = %+v, want first-key-wins {standalone 2}", got)
	}
}

func TestFieldsNoFrontmatter(t *testing.T) {
	if got := Fields([]string{"# Title", "id: itd-9"}); len(got) != 0 {
		t.Fatalf("no leading --- must yield no fields, got %+v", got)
	}
	if got := Fields(nil); len(got) != 0 {
		t.Fatalf("empty input must yield no fields, got %+v", got)
	}
}

func TestFieldsIgnoresNested(t *testing.T) {
	lines := []string{"---", "top: v", "  nested: v", "- item", "---"}
	fields := Fields(lines)
	if _, ok := fields["nested"]; ok {
		t.Fatal("indented key must be ignored (top-level only)")
	}
	if got := fields["top"]; got.Value != "v" {
		t.Fatalf("top = %+v, want v", got)
	}
}

func TestFieldsTrimsCarriageReturn(t *testing.T) {
	lines := []string{"---\r", "id: itd-9\r", "---\r"}
	if got := Fields(lines)["id"]; got.Value != "itd-9" {
		t.Fatalf("CRLF id = %+v, want itd-9", got)
	}
}

func TestIsNull(t *testing.T) {
	for _, v := range []string{"", "null", "~"} {
		if !IsNull(v) {
			t.Errorf("IsNull(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"itd-9", "spc-1", "standalone"} {
		if IsNull(v) {
			t.Errorf("IsNull(%q) = true, want false", v)
		}
	}
}
