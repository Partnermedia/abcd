package lifeboat

import (
	"fmt"
	"sort"
	"strings"
)

// conventionSources returns the Tier-1 adapters: brief sections derivable from
// conventional project files (README, docs/, CHANGELOG, LICENSE, CONTRIBUTING,
// build manifests, CI workflows, and ADRs wherever they live). Every adapter
// reads only through the SourceContext file surface and never touches git, so it
// grounds the same sections against any directory — a git working tree or a bare
// snapshot.
func conventionSources() []Source {
	return []Source{
		convContextSource{},
		convPressReleaseSource{},
		convScopeSource{},
		convPlatformSource{},
		convDependenciesSource{},
		convADRsSource{},
		convSurfacesSource{},
		convOutOfScopeSource{},
		convGlossarySource{},
		convIssuesSource{},
		convWhatWorkedSource{},
	}
}

// convReadmeNames are the conventional README spellings, in preference order.
var convReadmeNames = []string{
	"README.md", "README", "README.rst", "README.txt", "readme.md",
}

// convGroundedProseBytes is the body-prose threshold above which a README is
// treated as real documentation (grounded) rather than a near-empty stub
// (partial). Body prose excludes heading lines, code fences, and blank lines.
const convGroundedProseBytes = 120

// convReadme returns the first README that exists, with its contents. ok is
// false when no README is present.
func convReadme(ctx *SourceContext) (path string, data []byte, ok bool) {
	p := ctx.FindFirst(convReadmeNames...)
	if p == "" {
		return "", nil, false
	}
	d, read := ctx.ReadFile(p)
	if !read {
		return "", nil, false
	}
	return p, d, true
}

// convIsHeading reports whether a line is a Markdown heading.
func convIsHeading(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// convProseBytes counts the body-prose characters of a README: non-blank,
// non-heading, non-fence lines. It is the measure that separates a documented
// project from a stub.
func convProseBytes(data []byte) int {
	total := 0
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if t == "" || convIsHeading(line) || strings.HasPrefix(t, "```") {
			continue
		}
		total += len(t)
	}
	return total
}

// convLede extracts a README's lede: the first heading as a title and the first
// paragraph beneath it. ok is false when there is no non-blank content at all.
func convLede(data []byte) (title, para string, ok bool) {
	lines := strings.Split(string(data), "\n")
	i := 0
	// Title: first heading, else first non-blank line.
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			continue
		}
		if convIsHeading(lines[i]) {
			title = strings.TrimSpace(strings.TrimLeft(t, "#"))
		} else {
			title = t
		}
		i++
		break
	}
	if title == "" {
		return "", "", false
	}
	// Paragraph: first block of consecutive non-blank, non-heading lines.
	var buf []string
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			if len(buf) > 0 {
				break
			}
			continue
		}
		if convIsHeading(lines[i]) {
			if len(buf) > 0 {
				break
			}
			continue
		}
		buf = append(buf, t)
	}
	return title, strings.Join(buf, " "), true
}

// convHasHeadingLike reports whether any Markdown heading in data contains one
// of keywords (case-insensitive).
func convHasHeadingLike(data []byte, keywords ...string) bool {
	for _, line := range strings.Split(string(data), "\n") {
		if !convIsHeading(line) {
			continue
		}
		low := strings.ToLower(line)
		for _, k := range keywords {
			if strings.Contains(low, k) {
				return true
			}
		}
	}
	return false
}

// convHasCodeFence reports whether data contains a fenced code block, a strong
// signal of usage/CLI examples.
func convHasCodeFence(data []byte) bool {
	return strings.Contains(string(data), "\n```") || strings.HasPrefix(string(data), "```")
}

// convADRDirs are the conventional homes for ADRs, sorted so a citation is
// deterministic.
var convADRDirs = []string{
	"adr", "architecture/decisions", "docs/adr", "docs/adrs", "docs/decisions",
}

// convListADRs returns the repo-relative paths of ADR documents found under the
// conventional ADR directories, unique and sorted.
func convListADRs(ctx *SourceContext) []string {
	var out []string
	for _, dir := range convADRDirs {
		if !ctx.IsDir(dir) {
			continue
		}
		for _, name := range ctx.ListDir(dir) {
			low := strings.ToLower(name)
			if strings.HasSuffix(low, ".md") || strings.HasSuffix(low, ".markdown") {
				out = append(out, dir+"/"+name)
			}
		}
	}
	return dedupeSorted(out)
}

// convContextSource grounds "product/context" from a README carrying real prose.
// A near-empty README is partial; no README is a blank a human must answer.
type convContextSource struct{}

func (convContextSource) Section() Section { return "product/context" }
func (convContextSource) Tier() Tier       { return TierConventions }

func (convContextSource) Probe(ctx *SourceContext) Evidence {
	path, data, ok := convReadme(ctx)
	if !ok {
		return blank(
			[]string{"README (" + strings.Join(convReadmeNames, ", ") + ")"},
			"What does this project do? No README to describe it.",
		)
	}
	if convProseBytes(data) < convGroundedProseBytes {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceLow,
			Sources:    []string{path + " (near-empty)"},
		}
	}
	return Evidence{
		Status:     StatusGrounded,
		Confidence: ConfidenceHigh,
		Sources:    []string{path},
	}
}

// convPressReleaseSource partially grounds "product/press-release" from the
// README lede — its first heading and paragraph read as the project's pitch.
type convPressReleaseSource struct{}

func (convPressReleaseSource) Section() Section { return "product/press-release" }
func (convPressReleaseSource) Tier() Tier       { return TierConventions }

func (convPressReleaseSource) Probe(ctx *SourceContext) Evidence {
	path, data, ok := convReadme(ctx)
	if !ok {
		return blank(
			[]string{"README lede (first heading + paragraph)"},
			"What is the one-line pitch for this project? No README lede to draw it from.",
		)
	}
	title, _, hasLede := convLede(data)
	if !hasLede {
		return blank(
			[]string{"README lede (first heading + paragraph)"},
			"What is the one-line pitch for this project? The README carries no lede.",
		)
	}
	return Evidence{
		Status:     StatusPartial,
		Confidence: ConfidenceLow,
		Sources:    []string{fmt.Sprintf("%s lede (%q)", path, title)},
	}
}

// convScopeSource partially grounds "product/scope" from a README's features or
// usage sections. Blank when no README is present.
type convScopeSource struct{}

func (convScopeSource) Section() Section { return "product/scope" }
func (convScopeSource) Tier() Tier       { return TierConventions }

func (convScopeSource) Probe(ctx *SourceContext) Evidence {
	path, data, ok := convReadme(ctx)
	if !ok {
		return blank(
			[]string{"README features/usage sections"},
			"What is in scope for this project? No README to read features from.",
		)
	}
	if convHasHeadingLike(data, "feature", "usage", "install", "getting started", "what") {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{path + " (features/usage sections)"},
		}
	}
	return Evidence{
		Status:     StatusPartial,
		Confidence: ConfidenceLow,
		Sources:    []string{path + " (no explicit features section)"},
	}
}

// convPlatformFiles are the build/CI signals that ground "constraints/platform",
// sorted so a citation is deterministic.
var convPlatformFiles = []string{
	"Dockerfile", "Makefile", "go.mod", "package.json",
}

// convPlatformSource grounds "constraints/platform" from build manifests and CI
// workflows. Blank when the project carries neither.
type convPlatformSource struct{}

func (convPlatformSource) Section() Section { return "constraints/platform" }
func (convPlatformSource) Tier() Tier       { return TierConventions }

func (convPlatformSource) Probe(ctx *SourceContext) Evidence {
	var found []string
	for _, f := range convPlatformFiles {
		if ctx.Exists(f) {
			found = append(found, f)
		}
	}
	if ctx.IsDir(".github/workflows") {
		found = append(found, ".github/workflows")
	}
	if len(found) == 0 {
		return blank(
			[]string{"build manifests (" + strings.Join(convPlatformFiles, ", ") + ")", "CI workflows (.github/workflows)"},
			"What platform and toolchain does this project require? No build manifest or CI found.",
		)
	}
	sort.Strings(found)
	return Evidence{
		Status:     StatusGrounded,
		Confidence: ConfidenceHigh,
		Sources:    found,
	}
}

// convManifestLock pairs a dependency manifest with the lockfile that pins it. A
// present lockfile makes the dependency list authoritative (grounded); a missing
// one leaves it partial.
type convManifestLock struct {
	manifest string
	locks    []string // any one present counts as the lock
}

// convManifestLocks are the recognised manifest/lock pairs, in a fixed order so
// the first match is deterministic. It spans the ecosystems the probe is likely
// to meet — Go, Node, Rust, Python (pip/poetry/pdm/uv/pipenv), Ruby, and PHP —
// so a real project is not reported as having no dependencies merely because the
// probe did not know its packaging tool.
var convManifestLocks = []convManifestLock{
	{"go.mod", []string{"go.sum"}},
	{"package.json", []string{"package-lock.json", "yarn.lock", "pnpm-lock.yaml"}},
	{"Cargo.toml", []string{"Cargo.lock"}},
	{"pyproject.toml", []string{"uv.lock", "poetry.lock", "pdm.lock"}},
	{"Pipfile", []string{"Pipfile.lock"}},
	{"requirements.txt", []string{"requirements.txt"}}, // pinned requirements is its own lock
	{"Gemfile", []string{"Gemfile.lock"}},
	{"composer.json", []string{"composer.lock"}},
}

// convDependenciesSource grounds "constraints/dependencies" from a manifest and
// its lockfile; a manifest without a lockfile is partial. Blank when no manifest
// is present.
type convDependenciesSource struct{}

func (convDependenciesSource) Section() Section { return "constraints/dependencies" }
func (convDependenciesSource) Tier() Tier       { return TierConventions }

func (convDependenciesSource) Probe(ctx *SourceContext) Evidence {
	for _, ml := range convManifestLocks {
		if !ctx.Exists(ml.manifest) {
			continue
		}
		var lock string
		for _, l := range ml.locks {
			if l == ml.manifest {
				lock = l // requirements.txt pins itself
				break
			}
			if ctx.Exists(l) {
				lock = l
				break
			}
		}
		if lock != "" && lock != ml.manifest {
			return Evidence{
				Status:     StatusGrounded,
				Confidence: ConfidenceHigh,
				Sources:    []string{ml.manifest, lock},
			}
		}
		if lock == ml.manifest {
			return Evidence{
				Status:     StatusGrounded,
				Confidence: ConfidenceMedium,
				Sources:    []string{ml.manifest},
			}
		}
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{ml.manifest + " (no lockfile)"},
		}
	}
	return blank(
		[]string{"dependency manifest + lockfile (go.mod, package.json, Cargo.toml, pyproject.toml, Pipfile, requirements.txt, Gemfile, composer.json)"},
		"What does this project depend on? No dependency manifest found.",
	)
}

// convADRsSource grounds "docs/adrs" when ADRs live under a conventional ADR
// directory, listing the documents found. Blank when none exist.
type convADRsSource struct{}

func (convADRsSource) Section() Section { return "docs/adrs" }
func (convADRsSource) Tier() Tier       { return TierConventions }

func (convADRsSource) Probe(ctx *SourceContext) Evidence {
	adrs := convListADRs(ctx)
	if len(adrs) == 0 {
		return blank(
			[]string{"ADRs under " + strings.Join(convADRDirs, ", ")},
			"What architectural decisions has this project recorded? No ADR directory found.",
		)
	}
	sources := []string{fmt.Sprintf("%d ADR(s) under conventional dirs", len(adrs))}
	sources = append(sources, adrs...)
	return Evidence{
		Status:     StatusGrounded,
		Confidence: ConfidenceHigh,
		Sources:    dedupeSorted(sources),
	}
}

// convSurfacesSource partially grounds "surfaces" from a README's usage or CLI
// sections, or its fenced usage examples. Blank when no README is present.
type convSurfacesSource struct{}

func (convSurfacesSource) Section() Section { return "surfaces" }
func (convSurfacesSource) Tier() Tier       { return TierConventions }

func (convSurfacesSource) Probe(ctx *SourceContext) Evidence {
	path, data, ok := convReadme(ctx)
	if !ok {
		return blank(
			[]string{"README usage/CLI sections", "CLI help text"},
			"What are this project's surfaces (CLI, API)? No README usage to read them from.",
		)
	}
	if convHasHeadingLike(data, "usage", "cli", "command", "api", "getting started") || convHasCodeFence(data) {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{path + " (usage/CLI sections)"},
		}
	}
	return Evidence{
		Status:     StatusPartial,
		Confidence: ConfidenceLow,
		Sources:    []string{path + " (no explicit usage section)"},
	}
}

// convOutOfScopeSource partially grounds "delivery/out-of-scope" from a README's
// non-goals or out-of-scope section. Blank when that section is absent.
type convOutOfScopeSource struct{}

func (convOutOfScopeSource) Section() Section { return "delivery/out-of-scope" }
func (convOutOfScopeSource) Tier() Tier       { return TierConventions }

func (convOutOfScopeSource) Probe(ctx *SourceContext) Evidence {
	path, data, ok := convReadme(ctx)
	if ok && convHasHeadingLike(data, "non-goal", "non goal", "out of scope", "out-of-scope", "not a goal") {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{path + " (non-goals / out-of-scope section)"},
		}
	}
	return blank(
		[]string{"README non-goals / out-of-scope section"},
		"What did this project deliberately leave out? No out-of-scope section found.",
	)
}

// convGlossaryDocNames are the glossary spellings checked outside docs/.
var convGlossaryDocNames = []string{"GLOSSARY.md", "GLOSSARY", "glossary.md"}

// convGlossarySource partially grounds "glossary" from a glossary document under
// docs/ or at the repo root. Blank when none exists.
type convGlossarySource struct{}

func (convGlossarySource) Section() Section { return "glossary" }
func (convGlossarySource) Tier() Tier       { return TierConventions }

func (convGlossarySource) Probe(ctx *SourceContext) Evidence {
	if p := ctx.FindFirst(convGlossaryDocNames...); p != "" {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{p},
		}
	}
	for _, name := range ctx.ListDir("docs") {
		if strings.HasPrefix(strings.ToLower(name), "glossary") {
			return Evidence{
				Status:     StatusPartial,
				Confidence: ConfidenceMedium,
				Sources:    []string{"docs/" + name},
			}
		}
	}
	return blank(
		[]string{"GLOSSARY.md", "docs/glossary*"},
		"What terms does this project define? No glossary document found.",
	)
}

// convIssuesSource partially grounds "activity/issues" from a checked-in issues
// ledger, an issues directory, or issue templates. Blank when none is present.
type convIssuesSource struct{}

func (convIssuesSource) Section() Section { return "activity/issues" }
func (convIssuesSource) Tier() Tier       { return TierConventions }

func (convIssuesSource) Probe(ctx *SourceContext) Evidence {
	if p := ctx.FindFirst("ISSUES.md", "ISSUES"); p != "" {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{p},
		}
	}
	if ctx.IsDir("issues") {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{"issues/"},
		}
	}
	if ctx.IsDir(".github/ISSUE_TEMPLATE") {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceLow,
			Sources:    []string{".github/ISSUE_TEMPLATE"},
		}
	}
	return blank(
		[]string{"ISSUES.md", "issues/", ".github/ISSUE_TEMPLATE"},
		"What issues has this project tracked? No issue ledger or templates found.",
	)
}

// convChangelogNames are the conventional changelog spellings.
var convChangelogNames = []string{
	"CHANGELOG.md", "CHANGELOG", "CHANGELOG.rst", "CHANGES.md", "HISTORY.md",
}

// convWhatWorkedSource partially grounds "evidence/what-worked" from a CHANGELOG
// — what shipped and survived. Blank when no changelog exists.
type convWhatWorkedSource struct{}

func (convWhatWorkedSource) Section() Section { return "evidence/what-worked" }
func (convWhatWorkedSource) Tier() Tier       { return TierConventions }

func (convWhatWorkedSource) Probe(ctx *SourceContext) Evidence {
	if p := ctx.FindFirst(convChangelogNames...); p != "" {
		return Evidence{
			Status:     StatusPartial,
			Confidence: ConfidenceMedium,
			Sources:    []string{p},
		}
	}
	return blank(
		[]string{"CHANGELOG (" + strings.Join(convChangelogNames, ", ") + ")"},
		"What has this project shipped that worked? No changelog found.",
	)
}
