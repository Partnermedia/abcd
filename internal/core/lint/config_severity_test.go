package lint

import (
	"strings"
	"testing"
)

// A rule whose severity is off-enum would emit findings that serialize yet
// count toward no exit code — a clean exit beside a non-empty findings list,
// the shape the sibling engines (repolint.Evaluate, guard.Validate) fail
// closed on. The loader is the one place that can refuse it before a gate
// runs vacuously green.

func TestLoadConfigRefusesOffEnumRuleSeverity(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "rules": {"links_resolve": {"enabled": true, "severity": "blocking"}}
	}`)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("LoadConfig accepted an enabled rule with severity \"blocking\"; want rejection")
	}
	if !strings.Contains(err.Error(), "links_resolve") {
		t.Fatalf("rejection must name the offending rule, got: %v", err)
	}
}

func TestLoadConfigRefusesEnabledRuleWithNoSeverity(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "rules": {"links_resolve": {"enabled": true}}
	}`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig accepted an enabled rule with no severity; want rejection")
	}
}

func TestLoadConfigRefusesOffEnumTokenSeverity(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "banned_tokens": [
	    {"id":"t1","pattern":"foo","message":"no foo","severity":"Blocker","successor":"bar","allow_context":["ok"]}
	  ]
	}`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig accepted a banned token with severity \"Blocker\"; want rejection")
	}
}

func TestLoadConfigRefusesUnknownKeys(t *testing.T) {
	// A misspelt key silently zero-values the field it missed: "enabld" leaves
	// Enabled false (a rule disarmed), "severty" leaves Severity "" (findings
	// that count toward no exit). Strict decoding turns both into a refusal.
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "rules": {"links_resolve": {"enabld": true, "severity": "blocker"}}
	}`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig accepted an unknown config key (\"enabld\"); want rejection")
	}
}

func TestLoadConfigAcceptsDisabledRuleWithoutSeverity(t *testing.T) {
	// A disabled rule is inert; its severity is not consulted, so absence there
	// is not a fault.
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "rules": {"links_resolve": {"enabled": false}}
	}`)
	if _, err := LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig refused a disabled rule with no severity: %v", err)
	}
}

func TestLoadConfigAcceptsBothLiveSeverities(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "rules": {
	    "links_resolve": {"enabled": true, "severity": "blocker"},
	    "no_brittle_line_refs": {"enabled": true, "severity": "warn"}
	  }
	}`)
	if _, err := LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig refused the live severity vocabulary: %v", err)
	}
}
