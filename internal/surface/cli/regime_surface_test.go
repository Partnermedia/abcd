package cli

// regime_surface_test.go is itd-184's ac-3, and it has its own file so the
// ingest verb's surface tests and this guard never contend for one.
//
// The criterion: no registered flag and no registered configuration key sets or
// overrides a run's supply regime. The regime is the definition's property —
// stated in the definition file's `regime:` key and resolved from the position by
// construction (internal/core/reading) — and the whole point of putting it there
// is that an operator cannot choose the licence a reading reads under.
//
// The enumeration is PROGRAMMATIC (itd-195). The command tree is walked through
// commandSurface, the repository's one canonical cobra walk — the same one the
// release compatibility gate reads — rather than through a second walk written
// here, and the configuration schemas are walked by reflection over their json
// tags. A hand-written list of flag names would be a prose claim about how the
// surface behaves, and it would fall behind the surface the day someone added a
// command.
//
// Disclosed residue, as itd-184 states it: what the walk cannot see is a channel
// that was never registered — an environment variable read ad hoc, say. Nothing
// here adds a mechanism for one, and the walk sees every channel that is.

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/lint"
	"github.com/intentdriven/abcd/internal/core/rules"
)

// regimeToken is the thing no operator surface may carry. It is matched
// case-insensitively against flag names, flag shorthands and configuration keys.
const regimeToken = "regime"

// readingOperands is the reading verb's closed operand set, as the verb's own
// refusal messages state it: a position and a target state and nothing else,
// plus the two write-side operands and the root's inherited render switch.
//
// It is pinned deliberately. The name check below catches a flag CALLED regime;
// this catches a flag that would set one under any other name, by making any
// addition to the one verb that runs at a position fail until somebody says what
// the new operand does. A new operand on this verb is a decision, not an
// accident.
var readingOperands = map[string][]string{
	"abcd reading":          {},
	"abcd reading assemble": {"dry-run", "out", "position", "target"},
}

// configKey is one configuration key as reflection found it, with the path
// through the schema that reaches it.
type configKey struct {
	schema string
	path   string
}

// walkConfigKeys returns every json key reachable in a configuration schema,
// following struct fields, slices, maps and pointers. Reflection over the type
// rather than a written list, for ac-3's reason: a key added tomorrow is walked
// tomorrow, without anyone remembering to add it here.
func walkConfigKeys(schema string, root reflect.Type) []configKey {
	seen := map[reflect.Type]bool{}
	var out []configKey
	var walk func(t reflect.Type, prefix string)
	walk = func(t reflect.Type, prefix string) {
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice ||
			t.Kind() == reflect.Array || t.Kind() == reflect.Map {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || seen[t] {
			return
		}
		seen[t] = true
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				name = f.Name
			}
			key := name
			if prefix != "" {
				key = prefix + "." + name
			}
			out = append(out, configKey{schema: schema, path: key})
			walk(f.Type, key)
		}
	}
	walk(root, "")
	return out
}

// TestNoOperatorSurfaceSetsARegime is itd-184's ac-3.
//
// It is proved capable of failing by MUTATION rather than by passing:
// registering a `--regime` flag on `reading assemble` turns it red, and removing
// it turns it green again. A guard nobody has watched fail is a guard that
// asserts nothing.
func TestNoOperatorSurfaceSetsARegime(t *testing.T) {
	commands := commandSurface(NewRootCommand())

	// The walk must actually have walked. Without this, a tree that failed to
	// build, or a visitor that silently visited nothing, would report the
	// criterion satisfied by seeing no surface at all — which is exactly the
	// vacuity this guard exists to avoid.
	flagCount := 0
	for _, cmd := range commands {
		flagCount += len(cmd.Flags)
	}
	if len(commands) < 10 {
		t.Fatalf("the command walk found %d commands; the tree is larger than that, so the walk is broken",
			len(commands))
	}
	if flagCount < 20 {
		t.Fatalf("the flag walk found %d flags across %d commands; the walk is broken", flagCount, len(commands))
	}

	seenReadingCommands := map[string]bool{}
	for _, cmd := range commands {
		for _, f := range cmd.Flags {
			if strings.Contains(strings.ToLower(f.Name), regimeToken) {
				t.Errorf("%q registers the flag --%s: a reading's supply regime is its definition's "+
					"property, stated in the definition file, and no operand may set or override it",
					cmd.Path, f.Name)
			}
			if strings.Contains(strings.ToLower(f.Shorthand), regimeToken) {
				t.Errorf("%q registers the shorthand -%s on --%s, which names a regime",
					cmd.Path, f.Shorthand, f.Name)
			}
		}

		want, pinned := readingOperands[cmd.Path]
		if !pinned {
			continue
		}
		seenReadingCommands[cmd.Path] = true
		got := make([]string, 0, len(cmd.Flags))
		for _, f := range cmd.Flags {
			got = append(got, f.Name)
		}
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%q declares operands %v, want exactly %v; a new operand on the verb that runs at a "+
				"position has to say what it does before it can ship, because the regime is not one of them",
				cmd.Path, got, want)
		}
	}
	for path := range readingOperands {
		if !seenReadingCommands[path] {
			t.Errorf("the command walk never reached %q, so its operand set was never checked", path)
		}
	}

	var keys []configKey
	keys = append(keys, walkConfigKeys("record-lint config", reflect.TypeOf(lint.Config{}))...)
	keys = append(keys, walkConfigKeys("rules config", reflect.TypeOf(rules.RuleSet{}))...)
	if len(keys) < 20 {
		t.Fatalf("the configuration walk found %d keys; the schemas are larger than that, so the walk is broken",
			len(keys))
	}
	for _, k := range keys {
		if strings.Contains(strings.ToLower(k.path), regimeToken) {
			t.Errorf("the %s carries the key %q: a configuration file is an operator surface too, and the "+
				"regime is not settable from one", k.schema, k.path)
		}
	}
}
