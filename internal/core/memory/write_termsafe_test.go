package memory

import (
	"strings"
	"testing"
)

// write_termsafe_test.go — iss-2608270655495573. The memory READ path was masked
// in #250/#262, but the WRITE-time string emitters still put raw control bytes
// into the committed page: yaml.go dumpString escapes only \ " \n \t \r and
// writes ESC/C1/bidi/zero-width runes verbatim, and the markdown renderers
// (index.md, contradictions.md, log.md) interpolate page content unmasked. A
// `cat`/`git diff`/pager over the committed store then replays the escapes. Every
// write-time value emitter must route CONTENT through termsafe.CleanProse while
// leaving the structural YAML/markdown intact. Attack runes are numeric so this
// file carries none of them.
var writeAttackRunes = map[string]rune{
	"ESC":          0x1b,
	"C1 CSI":       0x9b,
	"RLO override": 0x202e,
	"zero-width":   0x200b,
}

func poisoned(prefix, suffix string) string {
	var b strings.Builder
	b.WriteString(prefix)
	for _, r := range writeAttackRunes {
		b.WriteRune(r)
	}
	b.WriteString(suffix)
	return b.String()
}

func assertNoWriteAttackRunes(t *testing.T, where, out string) {
	t.Helper()
	for name, r := range writeAttackRunes {
		if strings.ContainsRune(out, r) {
			t.Errorf("iss-2608270655495573: %s emits a raw %s (U+%04X) into the committed store; route content through termsafe.CleanProse\noutput: %q", where, name, r, out)
		}
	}
}

// TestDumpFrontmatterMasksAndRoundTrips proves a poisoned scalar VALUE is masked
// in the emitted frontmatter, the structural keys survive, and the page still
// parses on the way back (structure not corrupted).
func TestDumpFrontmatterMasksAndRoundTrips(t *testing.T) {
	fm := map[string]any{
		"title": poisoned("Hello", "World"),
		"class": "external_article",
	}
	dumped, err := dumpFrontmatter(fm)
	if err != nil {
		t.Fatalf("dumpFrontmatter: %v", err)
	}
	assertNoWriteAttackRunes(t, "dumpString", dumped)

	parsed, err := parseFrontmatter("---\n" + dumped + "---\n")
	if err != nil {
		t.Fatalf("round-trip parse failed — structural output corrupted: %v\ndumped:\n%s", err, dumped)
	}
	title, _ := parsed["title"].(string)
	if !strings.Contains(title, "Hello") || !strings.Contains(title, "World") {
		t.Fatalf("masking dropped legitimate title content: %q", title)
	}
	if parsed["class"] != "external_article" {
		t.Fatalf("structural key corrupted: class = %v", parsed["class"])
	}
}

func TestRenderIndexMasksContent(t *testing.T) {
	pages := []PageInfo{{
		Filename: "topic_auth_tokens.md",
		Classes:  []string{"external_article"},
		Domain:   poisoned("au", "th"),
		Summary:  poisoned("Rotate ", " tokens"),
	}}
	out := RenderIndex(pages)
	assertNoWriteAttackRunes(t, "RenderIndex", out)
	if !strings.Contains(out, "`topic_auth_tokens.md`") {
		t.Fatalf("RenderIndex dropped the page line structure:\n%s", out)
	}
}

func TestRenderContradictionsMasksContent(t *testing.T) {
	pages := []PageInfo{{
		Filename:    "topic_auth_a.md",
		Contradicts: []string{poisoned("topic_auth_", "b")},
	}}
	out := RenderContradictions(pages)
	assertNoWriteAttackRunes(t, "RenderContradictions", out)
	if !strings.Contains(out, "`topic_auth_a.md` contradicts") {
		t.Fatalf("RenderContradictions dropped the entry structure:\n%s", out)
	}
}

func TestRenderLogEventMasksContent(t *testing.T) {
	out := renderLogEvent("2026-01-01 00:00", poisoned("external_", "article"), poisoned("sl", "ug"), poisoned("Rotate ", " tokens"))
	assertNoWriteAttackRunes(t, "renderLogEvent", out)
	if !strings.Contains(out, "## [2026-01-01 00:00]") {
		t.Fatalf("renderLogEvent dropped the header structure:\n%s", out)
	}
}
