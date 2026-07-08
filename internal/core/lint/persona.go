package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// personaAttrRe matches a press-release quote attribution: `said <Name>,`.
// The trailing comma anchors the persona-attribution form ("said Kira, a
// maintainer") and keeps ordinary prose ("as we said above") out of scope.
// The name class is Unicode-wide (letters, marks, apostrophes, hyphens) so
// compound and non-ASCII names (O'Brien, Anne-Marie, Zoë) cannot slip past
// as silent non-matches.
var personaAttrRe = regexp.MustCompile(`\bsaid (\p{Lu}[\p{L}\p{M}'’-]*),`)

// loadPersonaRoster reads the personas registry and returns the set of
// registered names. The registry is the single source of truth for persona
// names (selection is by role; the role's registered name is used).
func loadPersonaRoster(repoRoot, rel string) (map[string]bool, error) {
	if rel == "" {
		return nil, fmt.Errorf("persona_registry: rule enabled but \"registry\" is not set")
	}
	data, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		return nil, fmt.Errorf("persona_registry: reading roster %s: %w", rel, err)
	}
	var reg struct {
		Personas []struct {
			Name string `json:"name"`
		} `json:"personas"`
	}
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("persona_registry: parsing roster %s: %w", rel, err)
	}
	if len(reg.Personas) == 0 {
		return nil, fmt.Errorf("persona_registry: roster %s has no personas — misconfigured registry, refusing to flag the whole record", rel)
	}
	roster := make(map[string]bool, len(reg.Personas))
	for _, p := range reg.Personas {
		roster[p.Name] = true
	}
	return roster, nil
}

// checkPersonaRegistry flags quote attributions whose persona name is not in
// the registry roster. Fenced-code lines are skipped via the caller's mask;
// content-exempt files (historical record) are the caller's concern.
func checkPersonaRegistry(rel string, lines []string, mask []bool, roster map[string]bool, cfg RuleConfig) []Finding {
	var out []Finding
	for i, line := range lines {
		if mask[i] {
			continue
		}
		for _, m := range personaAttrRe.FindAllStringSubmatch(line, -1) {
			if roster[m[1]] {
				continue
			}
			out = append(out, Finding{
				File:     rel,
				Line:     i + 1,
				RuleID:   "persona_registry",
				Severity: cfg.Severity,
				Message:  fmt.Sprintf("persona %q is not in the registry (%s); personas are selected by role and use the role's registered name", m[1], cfg.Registry),
			})
		}
	}
	return out
}
