package surface

import (
	"testing"
)

// The fixture constructors keep the taxonomy table readable: a row should show
// what changed, not how a Command literal is spelt.
func cmdOf(path string, flags ...Flag) Command { return Command{Path: path, Flags: flags} }
func optFlag(name string) Flag                 { return Flag{Name: name, Type: "bool"} }
func reqFlag(name string) Flag                 { return Flag{Name: name, Type: "bool", Required: true} }
func entry(key string) ManifestEntry {
	return ManifestEntry{File: ".claude-plugin/plugin.json", Key: key}
}

// describe renders a break set as the lines a failure message would name, so a
// table row can state the expected surface rather than a struct.
func describe(breaks []Break) []string {
	out := make([]string, 0, len(breaks))
	for _, b := range breaks {
		out = append(out, b.String())
	}
	return out
}

func sameLines(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestBreakKindWireValues pins the four identifiers a machine-readable front door
// emits. The constants are the only part of the taxonomy a consumer switches on,
// and every other test in this file asserts through String(), which would keep
// passing if a value were retyped — so without this, changing the contract is a
// silent edit. Renaming a constant stays safe; changing a value fails here, which
// is the point.
func TestBreakKindWireValues(t *testing.T) {
	tests := []struct {
		kind BreakKind
		want string
	}{
		{BreakCommandRemoved, "command_removed"},
		{BreakFlagRemoved, "flag_removed"},
		{BreakFlagRequired, "flag_required"},
		{BreakManifestRemoved, "manifest_entry_removed"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if string(tc.kind) != tc.want {
				t.Errorf("BreakKind = %q, want %q: the wire value is a contract a consumer switches on", tc.kind, tc.want)
			}
		})
	}
}

// TestDiffReportsTheRightKind pins that each taxonomy row is emitted under its own
// Kind, not merely under the right prose. Every other case here reads String(),
// which cannot tell a mislabelled Kind from a correct one once the renderer is
// changed to switch on Kind instead.
func TestDiffReportsTheRightKind(t *testing.T) {
	tests := []struct {
		name    string
		base    Snapshot
		current Snapshot
		want    BreakKind
	}{
		{
			name:    "removed command",
			base:    NewSnapshot([]Command{cmdOf("abcd"), cmdOf("abcd ghost")}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd")}, nil),
			want:    BreakCommandRemoved,
		},
		{
			name:    "removed flag",
			base:    NewSnapshot([]Command{cmdOf("abcd", optFlag("json"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd")}, nil),
			want:    BreakFlagRemoved,
		},
		{
			name:    "flag becoming required",
			base:    NewSnapshot([]Command{cmdOf("abcd launch", optFlag("dry-run"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd launch", reqFlag("dry-run"))}, nil),
			want:    BreakFlagRequired,
		},
		{
			name:    "removed manifest entry",
			base:    NewSnapshot(nil, []ManifestEntry{entry("description")}),
			current: NewSnapshot(nil, nil),
			want:    BreakManifestRemoved,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Diff(tc.base, tc.current)
			if len(got) != 1 {
				t.Fatalf("Diff = %v, want exactly one break", describe(got))
			}
			if got[0].Kind != tc.want {
				t.Errorf("Kind = %q, want %q", got[0].Kind, tc.want)
			}
		})
	}
}

// TestDiffTaxonomy walks every row of the break taxonomy (spc-10, plan outcome
// 5) in BOTH directions: what must be reported as a break, and what must not.
// The table is the specification — a row removed here is a hole in the gate.
func TestDiffTaxonomy(t *testing.T) {
	tests := []struct {
		name    string
		base    Snapshot
		current Snapshot
		want    []string
	}{
		{
			name:    "removed command is a break",
			base:    NewSnapshot([]Command{cmdOf("abcd"), cmdOf("abcd ghost")}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd")}, nil),
			want:    []string{"command removed or renamed: abcd ghost"},
		},
		{
			name:    "renamed command is a break at its old path",
			base:    NewSnapshot([]Command{cmdOf("abcd"), cmdOf("abcd ghost")}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd"), cmdOf("abcd spirit")}, nil),
			want:    []string{"command removed or renamed: abcd ghost"},
		},
		{
			name:    "removed flag is a break",
			base:    NewSnapshot([]Command{cmdOf("abcd", optFlag("json"), optFlag("quiet"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd", optFlag("json"))}, nil),
			want:    []string{"flag removed or renamed: abcd --quiet"},
		},
		{
			name:    "renamed flag is a break at its old name",
			base:    NewSnapshot([]Command{cmdOf("abcd", optFlag("json"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd", optFlag("jsonl"))}, nil),
			want:    []string{"flag removed or renamed: abcd --json"},
		},
		{
			name:    "optional flag becoming required is a break",
			base:    NewSnapshot([]Command{cmdOf("abcd launch", optFlag("dry-run"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd launch", reqFlag("dry-run"))}, nil),
			want:    []string{"flag is now required: abcd launch --dry-run"},
		},
		{
			name:    "absent flag arriving required on an existing command is a break",
			base:    NewSnapshot([]Command{cmdOf("abcd launch")}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd launch", reqFlag("token"))}, nil),
			want:    []string{"flag is now required: abcd launch --token"},
		},
		{
			name:    "removed manifest entry is a break",
			base:    NewSnapshot(nil, []ManifestEntry{entry("author.name"), entry("description")}),
			current: NewSnapshot(nil, []ManifestEntry{entry("author.name")}),
			want:    []string{"manifest entry removed: .claude-plugin/plugin.json:description"},
		},
		{
			name:    "new command is not a break",
			base:    NewSnapshot([]Command{cmdOf("abcd")}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd"), cmdOf("abcd brand-new")}, nil),
			want:    nil,
		},
		{
			name:    "new optional flag is not a break",
			base:    NewSnapshot([]Command{cmdOf("abcd", optFlag("json"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd", optFlag("json"), optFlag("verbose"))}, nil),
			want:    nil,
		},
		{
			name:    "new manifest entry is not a break",
			base:    NewSnapshot(nil, []ManifestEntry{entry("author.name")}),
			current: NewSnapshot(nil, []ManifestEntry{entry("author.name"), entry("version")}),
			want:    nil,
		},
		{
			name: "reordering is not a break",
			base: Snapshot{SchemaVersion: SchemaVersion,
				Commands: []Command{cmdOf("abcd ghost", optFlag("quiet"), optFlag("json")), cmdOf("abcd")},
				Manifest: []ManifestEntry{entry("description"), entry("author.name")}},
			current: Snapshot{SchemaVersion: SchemaVersion,
				Commands: []Command{cmdOf("abcd"), cmdOf("abcd ghost", optFlag("json"), optFlag("quiet"))},
				Manifest: []ManifestEntry{entry("author.name"), entry("description")}},
			want: nil,
		},
		{
			name:    "a flag staying required is not a fresh break",
			base:    NewSnapshot([]Command{cmdOf("abcd launch", reqFlag("token"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd launch", reqFlag("token"))}, nil),
			want:    nil,
		},
		{
			name:    "a required flag becoming optional is not a break",
			base:    NewSnapshot([]Command{cmdOf("abcd launch", reqFlag("token"))}, nil),
			current: NewSnapshot([]Command{cmdOf("abcd launch", optFlag("token"))}, nil),
			want:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := describe(Diff(tc.base, tc.current))
			if !sameLines(got, tc.want) {
				t.Errorf("Diff = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestDiffIgnoresRequirednessOnNewCommands is the edge the taxonomy's two rows
// meet at: "a new command is not a break" and "an absent flag becoming required
// is a break". A brand-new command's flags were all absent at the baseline, so a
// naive requiredness check would report every new command carrying a required
// flag as a break and make adding one impossible without a `breaking` record.
// Nobody can depend on surface that did not exist, so requiredness is only
// judged for commands present in BOTH snapshots.
func TestDiffIgnoresRequirednessOnNewCommands(t *testing.T) {
	base := NewSnapshot([]Command{cmdOf("abcd")}, nil)
	current := NewSnapshot([]Command{cmdOf("abcd"), cmdOf("abcd ship", reqFlag("changelog-json"))}, nil)

	if got := describe(Diff(base, current)); len(got) != 0 {
		t.Errorf("Diff = %v, want no breaks: a new command's required flag is new surface, not a narrowed one", got)
	}
}

// TestDiffReportsRemovedCommandOnce pins that removing a command reports the
// command, not also every flag it carried. A five-flag command would otherwise
// produce six lines for one break and bury the surface that actually matters.
func TestDiffReportsRemovedCommandOnce(t *testing.T) {
	base := NewSnapshot([]Command{cmdOf("abcd"), cmdOf("abcd ghost", optFlag("json"), optFlag("quiet"))}, nil)
	current := NewSnapshot([]Command{cmdOf("abcd")}, nil)

	want := []string{"command removed or renamed: abcd ghost"}
	if got := describe(Diff(base, current)); !sameLines(got, want) {
		t.Errorf("Diff = %v, want %v", got, want)
	}
}

// TestDiffIsOrderedAndCumulative pins that a multi-break cut reports every break
// in one deterministic pass: commands before manifest entries, each in canonical
// order. A gate that stopped at the first break would make the operator fix and
// re-run once per break.
func TestDiffIsOrderedAndCumulative(t *testing.T) {
	base := NewSnapshot(
		[]Command{cmdOf("abcd", optFlag("json")), cmdOf("abcd zeta"), cmdOf("abcd alpha")},
		[]ManifestEntry{entry("author.name"), entry("description")},
	)
	current := NewSnapshot([]Command{cmdOf("abcd")}, nil)

	want := []string{
		"flag removed or renamed: abcd --json",
		"command removed or renamed: abcd alpha",
		"command removed or renamed: abcd zeta",
		"manifest entry removed: .claude-plugin/plugin.json:author.name",
		"manifest entry removed: .claude-plugin/plugin.json:description",
	}
	if got := describe(Diff(base, current)); !sameLines(got, want) {
		t.Errorf("Diff = %v, want %v", got, want)
	}
}

// TestDiffDistinguishesManifestFiles pins that an entry is identified by its file
// AND its key: the same key path in the other manifest is a different entry, so
// moving a key between manifests is a removal from the one that lost it.
func TestDiffDistinguishesManifestFiles(t *testing.T) {
	base := NewSnapshot(nil, []ManifestEntry{{File: ".claude-plugin/plugin.json", Key: "name"}})
	current := NewSnapshot(nil, []ManifestEntry{{File: ".claude-plugin/marketplace.json", Key: "name"}})

	want := []string{"manifest entry removed: .claude-plugin/plugin.json:name"}
	if got := describe(Diff(base, current)); !sameLines(got, want) {
		t.Errorf("Diff = %v, want %v", got, want)
	}
}
