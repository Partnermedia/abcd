package reading

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefinitionsDir is where the four reading definitions live. The assembler never
// reads one — a definition is the reader's, and this package denies the whole
// directory to every assembly — but the status render reports which are present,
// because a missing definition is the difference between an instrument that can
// be dispatched and one that cannot.
const DefinitionsDir = "agents"

// definitionPrefix is the filename prefix a cold-reading definition carries.
const definitionPrefix = "cold-reading-"

// Status is the read-only render behind the bare verb: what this assembler is,
// what it would admit, which definitions are present, and which runs are staged
// in the local tier without having been ingested.
type Status struct {
	AssemblerVersion string     `json:"assembler_version"`
	SchemaVersion    int        `json:"schema_version"`
	CharterPath      string     `json:"charter_path"`
	Positions        []Position `json:"positions"`
	IncludeRows      int        `json:"include_rows"`
	ExclusionRows    int        `json:"exclusion_rows"`
	Definitions      []string   `json:"definitions"`
	StagedRuns       []string   `json:"staged_runs"`
}

// Describe reports the assembler's state over a repository. It writes nothing.
func Describe(repoRoot string) (Status, error) {
	s := Status{
		AssemblerVersion: AssemblerVersion(),
		SchemaVersion:    SchemaVersion,
		CharterPath:      CharterPath,
		Positions:        Positions(),
		IncludeRows:      len(Table),
		ExclusionRows:    len(Exclusions),
		Definitions:      []string{},
		StagedRuns:       []string{},
	}
	if repoRoot == "" {
		return s, nil
	}

	// Resolved, not listed. The render reports the definitions the ingest verb
	// would actually resolve: a file the locator refuses — silent about its
	// position or its regime — is a fault reported here rather than an
	// instrument reported present, and a `cold-reading-*.md` naming no position
	// is not an instrument at all, because the position set is closed.
	defs, err := LoadDefinitions(repoRoot)
	if err != nil {
		return Status{}, err
	}
	for _, d := range defs {
		s.Definitions = append(s.Definitions, definitionPrefix+string(d.Position))
	}
	sort.Strings(s.Definitions)

	runs, err := os.ReadDir(filepath.Join(repoRoot, filepath.FromSlash(DefaultRunDir)))
	if err != nil && !os.IsNotExist(err) {
		return Status{}, fmt.Errorf("reading: listing the staged runs: %w", err)
	}
	for _, e := range runs {
		if e.IsDir() && strings.HasPrefix(e.Name(), RunIDFamily+"-") {
			s.StagedRuns = append(s.StagedRuns, e.Name())
		}
	}
	sort.Strings(s.StagedRuns)
	return s, nil
}
