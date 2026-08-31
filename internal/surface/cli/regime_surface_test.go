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
// The enumeration is PROGRAMMATIC (itd-195), on both halves.
//
// The command tree is walked through commandSurface, the repository's one
// canonical cobra walk — the same one the release compatibility gate reads —
// rather than through a second walk written here.
//
// The configuration side walks every key of every committed configuration file,
// found by DIRECTORY rather than by a list of schemas: a configuration file
// added tomorrow is walked tomorrow, without anyone remembering to name it here.
// The two largest schema types are additionally walked by reflection over their
// json tags, which reaches a key the schema declares that no committed file
// happens to carry. A hand-written list of either kind would be a prose claim
// about how the surface behaves, and it would fall behind the surface the day
// someone added a command or a config.
//
// Disclosed residue, as itd-184 states it: what the walk cannot see is a channel
// that was never registered — an environment variable read ad hoc, say. Nothing
// here adds a mechanism for one, and the walk sees every channel that is. The
// one narrower edge: a key declared by a schema type OUTSIDE the two walked here
// and written into no committed configuration file is unregistered in both
// senses until a file carries it, at which point the file walk reaches it.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/lint"
	"github.com/intentdriven/abcd/internal/core/rules"
	"github.com/intentdriven/abcd/internal/fsutil"
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

// configFileGlobs are the DIRECTORIES abcd's committed configuration lives in.
// They are directories, not a list of schemas: everything matching is walked, so
// a configuration file added later is covered without an edit here.
var configFileGlobs = []string{
	".abcd/*.json",
	".abcd/config/*.json",
}

// configKey is one configuration key as the walk found it, with the path through
// the file or schema that reaches it.
type configKey struct {
	schema string
	path   string
}

// walkConfigFiles returns every key of every committed configuration file, keys
// nested inside objects and arrays included. This is the enumeration of the
// registered configuration surface: a key an operator can actually write.
func walkConfigFiles(t *testing.T, repoRoot string) []configKey {
	t.Helper()
	var out []configKey
	var files int
	var walk func(file string, v any, prefix string)
	walk = func(file string, v any, prefix string) {
		switch node := v.(type) {
		case map[string]any:
			for k, child := range node {
				key := k
				if prefix != "" {
					key = prefix + "." + k
				}
				out = append(out, configKey{schema: file, path: key})
				walk(file, child, key)
			}
		case []any:
			for _, child := range node {
				walk(file, child, prefix)
			}
		}
	}
	for _, glob := range configFileGlobs {
		matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash(glob)))
		if err != nil {
			t.Fatalf("glob %s: %v", glob, err)
		}
		for _, path := range matches {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			var doc any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("decode %s: %v", path, err)
			}
			files++
			walk(filepath.ToSlash(fsutil.RepoRel(repoRoot, path)), doc, "")
		}
	}
	if files < 8 {
		t.Fatalf("the configuration-file walk found %d files under %v; the repository carries more, "+
			"so the walk is broken", files, configFileGlobs)
	}
	return out
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

	keys := walkConfigFiles(t, repoRootFromTest(t))
	keys = append(keys, walkConfigKeys("record-lint schema", reflect.TypeOf(lint.Config{}))...)
	keys = append(keys, walkConfigKeys("rules schema", reflect.TypeOf(rules.RuleSet{}))...)
	if len(keys) < 100 {
		t.Fatalf("the configuration walk found %d keys; the committed configuration is larger than that, "+
			"so the walk is broken", len(keys))
	}
	for _, k := range keys {
		if strings.Contains(strings.ToLower(k.path), regimeToken) {
			t.Errorf("%s carries the key %q: a configuration file is an operator surface too, and the "+
				"regime is not settable from one", k.schema, k.path)
		}
	}
}
