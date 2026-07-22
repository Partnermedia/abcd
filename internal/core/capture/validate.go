package capture

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/changelog"
	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
)

// knownFields is the additionalProperties:false allow-list from
// issue.schema.json.
var knownFields = map[string]bool{
	"schema_version": true, "id": true, "slug": true, "severity": true,
	"category": true, "source": true, "found_during": true, "found_at": true,
	"details": true, "suggested_fix": true, "related_intents": true,
	"promoted_to": true, "related_specs": true, "related_issues": true,
	"synthesis_clusters": true, "wontfix_reason": true, "resolution": true,
	"resolved_by": true, "blocked_by": true,
	// impact is the product judgement the derived version and the generated
	// changelog are computed from (spc-10). It is optional here — an open issue
	// has not been judged yet, and the record-lint blocker issue_impact_valid is
	// what gates the move into resolved/ — but it must be a KNOWN property, or
	// the reader drops every judged record as malformed.
	"impact": true,
	// created/updated are no longer written, but legacy ledgers still carry
	// them. Tolerate (accept, then drop) them on read so an existing committed
	// ledger is not rejected as an unknown property; the reader ignores their
	// values entirely.
	"created": true, "updated": true,
}

// uniqueItemsFields are the array properties issue.schema.json flags
// uniqueItems:true.
var uniqueItemsFields = []string{"related_intents", "related_specs", "related_issues", "synthesis_clusters", "blocked_by"}

// validateStrict validates a frontmatter map against the issue schema. It
// special-cases schema_version first (mirrors _validate_strict) and rejects
// unknown keys (additionalProperties:false).
func validateStrict(fm map[string]any) error {
	sv, ok := fm["schema_version"]
	if !ok {
		return fmt.Errorf("%w: missing required property 'schema_version'", ErrMissingRequiredField)
	}
	if n, isInt := sv.(int); !isInt || n != 1 {
		return fmt.Errorf("%w: unsupported schema_version %v (this reader only handles 1)", ErrMissingRequiredField, sv)
	}

	for k := range fm {
		if !knownFields[k] {
			return fmt.Errorf("%w: unknown property %q", ErrMalformedFrontmatter, k)
		}
	}

	// Required strings.
	for _, req := range []string{"id", "slug", "severity", "category", "source", "found_during"} {
		v, present := fm[req]
		if !present {
			return fmt.Errorf("%w: missing required property %q", ErrMissingRequiredField, req)
		}
		if _, isStr := v.(string); !isStr {
			return fmt.Errorf("%w: %q must be a string", ErrMalformedFrontmatter, req)
		}
	}

	id := fm["id"].(string)
	if !reIssID.MatchString(id) {
		return fmt.Errorf("%w: id %q does not match ^iss-[0-9]+$", ErrMalformedFrontmatter, id)
	}
	if !reSlug.MatchString(fm["slug"].(string)) {
		return fmt.Errorf("%w: slug %q is not kebab-case", ErrMalformedFrontmatter, fm["slug"])
	}
	if !validSeverities[Severity(fm["severity"].(string))] {
		return fmt.Errorf("%w: invalid severity %q", ErrMalformedFrontmatter, fm["severity"])
	}
	if !validCategories[Category(fm["category"].(string))] {
		return fmt.Errorf("%w: invalid category %q", ErrMalformedFrontmatter, fm["category"])
	}
	if !validSources[Source(fm["source"].(string))] {
		return fmt.Errorf("%w: invalid source %q", ErrMalformedFrontmatter, fm["source"])
	}
	if strings.TrimSpace(fm["found_during"].(string)) == "" {
		return fmt.Errorf("%w: found_during must be non-empty", ErrMalformedFrontmatter)
	}

	// impact is optional but, when written, is checked against the ONE shared
	// enum (internal/core/changelog) rather than a private copy — the same enum
	// the record lint gates on and the release derivation consumes, so all three
	// can never disagree about what a legal judgement is.
	//
	// A YAML null reads as ABSENT rather than as a value, via the repo's one null
	// test (frontmatter.IsNull) — the same test the record-lint blocker
	// issue_impact_valid applies. Both gates must reach the same verdict on one
	// value: an open issue written `impact: null` ("not judged yet", the schema's
	// own convention) would otherwise pass the lint and then be refused here,
	// failing `abcd capture resolve` on a record nothing is wrong with.
	if v, present := fm["impact"]; present {
		s, isStr := v.(string)
		if !isStr {
			return fmt.Errorf("%w: %q must be a string", ErrMalformedFrontmatter, "impact")
		}
		if !frontmatter.IsNull(s) {
			if _, err := changelog.ParseImpact(s); err != nil {
				return fmt.Errorf("%w: %v", ErrMalformedFrontmatter, err)
			}
		}
	}

	// Optional scalar strings.
	for _, opt := range []string{"found_at", "details", "suggested_fix", "wontfix_reason", "resolution", "promoted_to"} {
		if v, present := fm[opt]; present {
			if _, isStr := v.(string); !isStr {
				return fmt.Errorf("%w: %q must be a string", ErrMalformedFrontmatter, opt)
			}
		}
	}
	if v, present := fm["promoted_to"]; present {
		if !reItdID.MatchString(v.(string)) {
			return fmt.Errorf("%w: promoted_to %q does not match ^itd-[0-9]+$", ErrMalformedFrontmatter, v)
		}
	}

	// Optional id-list fields.
	idListFields := []struct {
		field string
		re    *regexp.Regexp
		desc  string
	}{
		{"related_intents", reItdID, "itd-N"},
		{"related_specs", reFnID, "fn-N"},
		{"related_issues", reIssID, "iss-N"},
		{"blocked_by", reIssID, "iss-N"},
	}
	for _, f := range idListFields {
		v, present := fm[f.field]
		if !present {
			continue
		}
		items, isList := v.([]string)
		if !isList {
			return fmt.Errorf("%w: %q must be a list", ErrMalformedFrontmatter, f.field)
		}
		for _, it := range items {
			if !f.re.MatchString(it) {
				return fmt.Errorf("%w: %q item %q does not match %s", ErrMalformedFrontmatter, f.field, it, f.desc)
			}
		}
	}
	if v, present := fm["synthesis_clusters"]; present {
		if _, isList := v.([]string); !isList {
			return fmt.Errorf("%w: synthesis_clusters must be a list", ErrMalformedFrontmatter)
		}
	}
	if v, present := fm["resolved_by"]; present {
		m, isMap := v.(map[string]any)
		if !isMap {
			return fmt.Errorf("%w: resolved_by must be an object", ErrMalformedFrontmatter)
		}
		for k, sv := range m {
			if k != "intent" && k != "spec" && k != "commit" {
				return fmt.Errorf("%w: resolved_by has unknown key %q", ErrMalformedFrontmatter, k)
			}
			// Type-check the sub-value: issueFromFrontmatter reads it via asString,
			// which coerces a non-string (a number, a list) to "". Without this
			// check a malformed resolved_by validates cleanly and then silently
			// loses its value on read, so the round-trip is lossy and undetected.
			if _, isStr := sv.(string); !isStr {
				return fmt.Errorf("%w: resolved_by.%s must be a string", ErrMalformedFrontmatter, k)
			}
		}
	}
	return nil
}

// validateInvariants enforces folder<->field invariants and filename<->id
// match, mirroring _issue_lib._validate_invariants. Assumes validateStrict ran.
func validateInvariants(fm map[string]any, status State, path string) error {
	id, _ := fm["id"].(string)
	if !reIssID.MatchString(id) {
		return fmt.Errorf("%w: id %q does not match ^iss-[0-9]+$", ErrInvariantViolation, id)
	}
	name := filepath.Base(path)
	m := reFilenameID.FindStringSubmatch(name)
	if m == nil {
		return fmt.Errorf("%w: filename %q does not match iss-N[-slug].md", ErrInvariantViolation, name)
	}
	if m[1] != id {
		return fmt.Errorf("%w: filename id %q does not match frontmatter id %q", ErrInvariantViolation, m[1], id)
	}

	_, hasResolution := fm["resolution"]
	_, hasWontfix := fm["wontfix_reason"]
	switch status {
	case StateOpen:
		if hasResolution {
			return fmt.Errorf("%w: resolution must not appear in open/", ErrInvariantViolation)
		}
		if hasWontfix {
			return fmt.Errorf("%w: wontfix_reason must not appear in open/", ErrInvariantViolation)
		}
	case StateResolved:
		if !hasResolution {
			return fmt.Errorf("%w: resolution required in resolved/", ErrMissingRequiredField)
		}
		if isBlank(fm["resolution"]) {
			return fmt.Errorf("%w: resolution required non-empty in resolved/", ErrInvariantViolation)
		}
		if hasWontfix {
			return fmt.Errorf("%w: wontfix_reason must not appear in resolved/", ErrInvariantViolation)
		}
	case StateWontfix:
		if !hasWontfix {
			return fmt.Errorf("%w: wontfix_reason required in wontfix/", ErrMissingRequiredField)
		}
		if isBlank(fm["wontfix_reason"]) {
			return fmt.Errorf("%w: wontfix_reason required non-empty in wontfix/", ErrInvariantViolation)
		}
		if hasResolution {
			return fmt.Errorf("%w: resolution must not appear in wontfix/", ErrInvariantViolation)
		}
	default:
		return fmt.Errorf("%w: unknown status directory %q", ErrInvariantViolation, status)
	}

	for _, field := range uniqueItemsFields {
		v, present := fm[field]
		if !present {
			continue
		}
		items, ok := v.([]string)
		if !ok {
			continue
		}
		seen := map[string]bool{}
		for _, it := range items {
			if seen[it] {
				return fmt.Errorf("%w: %s contains duplicate items", ErrInvariantViolation, field)
			}
			seen[it] = true
		}
	}
	return nil
}

func isBlank(v any) bool {
	s, ok := v.(string)
	return ok && strings.TrimSpace(s) == ""
}

// issueFromFrontmatter builds a typed Issue from a validated frontmatter map.
func issueFromFrontmatter(fm map[string]any, status State, path, body string) Issue {
	iss := Issue{
		SchemaVersion: fm["schema_version"].(int),
		ID:            asString(fm["id"]),
		Slug:          asString(fm["slug"]),
		Severity:      Severity(asString(fm["severity"])),
		Category:      Category(asString(fm["category"])),
		Source:        Source(asString(fm["source"])),
		FoundDuring:   asString(fm["found_during"]),
		FoundAt:       asString(fm["found_at"]),
		PromotedTo:    asString(fm["promoted_to"]),
		Resolution:    asString(fm["resolution"]),
		WontfixReason: asString(fm["wontfix_reason"]),
		Status:        status,
		Path:          path,
		Body:          body,
	}
	iss.RelatedIntents = asStrList(fm["related_intents"])
	iss.RelatedSpecs = asStrList(fm["related_specs"])
	iss.RelatedIssues = asStrList(fm["related_issues"])
	iss.BlockedBy = asStrList(fm["blocked_by"])
	if rb, ok := fm["resolved_by"].(map[string]any); ok {
		iss.ResolvedBy = &ResolvedBy{
			Intent: asString(rb["intent"]),
			Spec:   asString(rb["spec"]),
			Commit: asString(rb["commit"]),
		}
	}
	return iss
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asStrList(v any) []string {
	l, _ := v.([]string)
	if len(l) == 0 {
		return nil
	}
	return l
}
