//go:build smoke || coldreading

package evals

// The read-block eval: the only component in this workstream capable of
// falsifying the cold reading's blindfold rather than restating it.
//
// Every other component asserts the blindfold. This one plants sentinel warm
// content across every warm location class in a fixture repository state, runs
// the assembler over it, and asserts that no sentinel and no excluded field
// reaches the assembled input — against an exclusion table transcribed from the
// record rather than read from the assembler.

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestReadBlockBaselineIsClean is ac-1: with every plant in its canonical home,
// the assembly passes no planted warm content and no excluded field.
//
// Watched red against an assembler whose exclusion of the local ledger tier is
// removed by a one-line local patch, with ABCD-EVAL-SENTINEL-LEDGER-FRAMING the
// sentinel that must be caught.
func TestReadBlockBaselineIsClean(t *testing.T) {
	f := materialise(t, variantBaseline)
	for _, position := range fullyAsserted {
		t.Run(position, func(t *testing.T) {
			vs := checkReadBlock(assemble(t, f, position))
			if len(vs) > 0 {
				t.Fatalf("the read-block leaked %d time(s) at the %s position:\n%s",
					len(vs), position, reportViolations(vs))
			}
		})
	}
}

// TestReadBlockCatchesAHoledFirewall is ac-2: a plant relocated into positively
// included material fails, naming the leaked sentinel's class token and the
// position at which it leaked.
//
// It is also the permanent proof that the assertion CAN fail. It demands
// exactly the declared holes, naming their classes, so an assertion that stops
// detecting anything fails here rather than passing quietly.
func TestReadBlockCatchesAHoledFirewall(t *testing.T) {
	if len(holes) == 0 {
		t.Fatal("the holes table is empty, so the negative control controls nothing; " +
			"this test is the permanent proof that the assertion can fail, and it cannot " +
			"be that with no hole in it")
	}
	f := materialise(t, variantHoled)
	want := make([]string, 0, len(holes))
	for _, h := range holes {
		want = append(want, h.Class)
	}
	sort.Strings(want)

	for _, position := range fullyAsserted {
		t.Run(position, func(t *testing.T) {
			vs := checkReadBlock(assemble(t, f, position))
			got := []string{}
			for _, v := range vs {
				if v.Rule != ruleSentinel {
					t.Errorf("the holed variant produced an unexpected finding: %s", v)
					continue
				}
				got = append(got, v.Class)
			}
			sort.Strings(got)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("the holed firewall reported classes %v at the %s position, want %v\n%s",
					got, position, want, reportViolations(vs))
			}
			// The message an operator reads has to carry both, or a failure names
			// no way back to the plant that caused it.
			for _, v := range vs {
				if v.Rule != ruleSentinel {
					continue // already reported above; it carries no class token
				}
				msg := v.String()
				if !strings.Contains(msg, sentinelPrefix+v.Class) || !strings.Contains(msg, position) {
					t.Errorf("the failure message %q names neither the class token nor the position", msg)
				}
			}
		})
	}
}

// TestReadBlockCatchesWarmFieldsOnIncludedTypes is ac-3: a warm field on a
// record type already on the include list fails.
//
// The origin and production-mode plants sit on included record types, and the
// Audit Notes, Why This Matters and scope-condition-disposition plants sit
// inside an included file. Each is a real leak path the moment the key filter
// or the projection is absent — which is how this test is watched red, by
// removing the single exclusion it covers.
func TestReadBlockCatchesWarmFieldsOnIncludedTypes(t *testing.T) {
	f := materialise(t, variantBaseline)
	warm := map[string]bool{"WARM-KEY": true, "WARM-FIELD": true, "DRAFT-ORIGIN": true}
	for _, position := range everyPosition {
		t.Run(position, func(t *testing.T) {
			a := assemble(t, f, position)
			var vs []violation
			for _, v := range checkSentinelAbsence(a) {
				if warm[v.Class] {
					vs = append(vs, v)
				}
			}
			vs = append(vs, checkFieldAbsence(a)...)
			if len(vs) > 0 {
				t.Fatalf("a warm field on an included record type reached the %s position %d time(s):\n%s",
					position, len(vs), reportViolations(vs))
			}
		})
	}
}

// TestPriorRunExhaustNeverReaches is ac-4: no prior-run manifest, reading record
// or disposition reaches the assembled input.
//
// The instrument's own exhaust is what a comparative reading is most likely to
// be handed by accident, and nothing else in the cycle tests it.
func TestPriorRunExhaustNeverReaches(t *testing.T) {
	f := materialise(t, variantBaseline)
	for _, position := range everyPosition {
		t.Run(position, func(t *testing.T) {
			a := assemble(t, f, position)
			var vs []violation
			for _, v := range checkSentinelAbsence(a) {
				if v.Class == "EXHAUST" {
					vs = append(vs, v)
				}
			}
			for _, v := range checkFamilyAbsence(a) {
				if strings.Contains(v.Detail, ".abcd/development/readings") ||
					strings.Contains(v.Detail, ".abcd/work/issues") {
					vs = append(vs, v)
				}
			}
			if len(vs) > 0 {
				t.Fatalf("the instrument's own exhaust reached the %s position %d time(s):\n%s",
					position, len(vs), reportViolations(vs))
			}
		})
	}
}

// bannedImports are the import paths that would make this oracle derived rather
// than transcribed: the assembler's own package, and the include list its
// structural deny is modelled on.
var bannedImports = []string{"internal/core/reading", "internal/core/launch"}

// TestOracleImportsNothingFromTheAssembler is ac-5: no file under evals/ names
// the assembler's package or its include list in an import path, so the
// exclusion table above is transcribed rather than derived.
//
// The check is over DIRECT import paths, which is the independence the record
// asks for. The dedicated lane is stronger than that and can be checked by
// hand: `go list -tags coldreading -deps -test ./evals/` reports this module's
// gitutil and gittest and nothing else, so the assembler is not linked into the
// cold-reading test binary at all, transitively or otherwise. The smoke lane's
// command-tree walk imports the CLI surface, which does reach the assembler;
// that is disclosed rather than caught, and it is harmless because no
// cold-reading assertion reads a Go symbol at all — every one of them reads
// bytes the built binary wrote.
func TestOracleImportsNothingFromTheAssembler(t *testing.T) {
	var files []string
	err := filepath.WalkDir(".", func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".go") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking evals/ for Go files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("found no Go files under evals/ — the import guard would pass vacuously")
	}

	fset := token.NewFileSet()
	for _, p := range files {
		f, perr := parser.ParseFile(fset, p, nil, parser.ImportsOnly)
		if perr != nil {
			t.Errorf("parsing %s: %v", p, perr)
			continue
		}
		for _, imp := range f.Imports {
			pathValue := strings.Trim(imp.Path.Value, `"`)
			for _, banned := range bannedImports {
				if strings.Contains(pathValue, banned) {
					t.Errorf("%s imports %s; the oracle's exclusion table is transcribed from the "+
						"record, and an eval that read the assembler's own table could only confirm "+
						"it, never falsify it", p, pathValue)
				}
			}
		}
	}
}

// TestEverySentinelIsPlanted is ac-6: every declared sentinel class is present
// in the materialised fixture its declared number of times.
//
// It is the anti-vacuity guard, and it is a criterion rather than a nicety: an
// absence assertion cannot see a corpus that lost its plants, so a corpus that
// silently dropped one would turn every assertion above green for the wrong
// reason. It also checks that every planted repository file is TRACKED, because
// the assembler intersects its walk with the tracked set — a plant git refused
// to track is a plant the assembler could never have leaked.
func TestEverySentinelIsPlanted(t *testing.T) {
	for _, variant := range []string{variantBaseline, variantHoled} {
		t.Run(variant, func(t *testing.T) {
			f := materialise(t, variant)
			tracked := trackedFiles(t, f.Root)
			counts, wheres := countSentinels(t, f)

			for _, c := range sentinelClasses {
				if counts[c.Name] != c.Count {
					t.Errorf("%s is planted %d time(s), want %d; found in %v",
						c.Token(), counts[c.Name], c.Count, wheres[c.Name])
				}
			}
			if variant != variantBaseline {
				return
			}
			for _, c := range sentinelClasses {
				want := append([]string{}, c.Homes...)
				sort.Strings(want)
				got := append([]string{}, wheres[c.Name]...)
				sort.Strings(got)
				if strings.Join(got, ",") != strings.Join(want, ",") {
					t.Errorf("%s is planted in %v, want %v", c.Token(), got, want)
				}
				for _, home := range c.Homes {
					rel, ok := strings.CutPrefix(home, "repo:")
					if !ok {
						continue
					}
					if !tracked[rel] {
						t.Errorf("%s is planted in %s, which the fixture repository does not track; "+
							"the assembler walks the tracked set, so an untracked plant tests nothing",
							c.Token(), rel)
					}
				}
			}
		})
	}
}

// countSentinels walks the materialised fixture — the repository and the
// fixture home both — and reports how many times each class's token occurs and
// which files carry it.
func countSentinels(t *testing.T, f fixture) (map[string]int, map[string][]string) {
	t.Helper()
	counts := map[string]int{}
	wheres := map[string][]string{}
	for _, tree := range []struct{ root, prefix string }{
		{f.Root, "repo:"},
		{f.Home, "home:"},
	} {
		err := filepath.WalkDir(tree.root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// The object database is compressed, not planted material, and a
				// packed blob would count a plant twice.
				if d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			data, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			rel, rerr := filepath.Rel(tree.root, p)
			if rerr != nil {
				return rerr
			}
			rel = filepath.ToSlash(rel)
			// The fixture home is keyed on a sha only the run knows, so it is put
			// back to the placeholder the corpus declares its homes with.
			if tree.prefix == "home:" && f.RootSHA != "" {
				rel = strings.ReplaceAll(rel, f.RootSHA, rootSHAPlaceholder)
			}
			for _, c := range sentinelClasses {
				n := bytes.Count(data, []byte(c.Token()))
				if n == 0 {
					continue
				}
				counts[c.Name] += n
				wheres[c.Name] = append(wheres[c.Name], tree.prefix+rel)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking the materialised fixture at %s: %v", tree.root, err)
		}
	}
	return counts, wheres
}
