package guard

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// SchemaVersion is the guard config schema this build implements. It is
// independent of the rules loader's schema: the two features keep separate
// switches so a rules kill switch can never silently disable a safety guard.
const SchemaVersion = 1

// Tier values, mirroring the record-lint severity family: a blocker refuses the
// command, a warn lets it through with the warning attached.
const (
	TierBlocker = "blocker"
	TierWarn    = "warn"
)

// Verdict is the decision a front door acts on.
type Verdict string

const (
	// VerdictAllow means no entry matched: run the command.
	VerdictAllow Verdict = "allow"
	// VerdictWarn means a warn-tier entry matched: run the command with the
	// warning surfaced.
	VerdictWarn Verdict = "warn"
	// VerdictBlock means a blocker-tier entry matched: refuse, and reply with
	// the successor as the block message.
	VerdictBlock Verdict = "block"
)

// Pattern is a command-position match expressed over shell tokens, never raw
// text. All declared constraints must hold for the pattern to fire.
type Pattern struct {
	// Command is the executable that must stand in command position. It is
	// compared against the segment's argv[0] by basename, so /bin/rm matches rm.
	Command string `json:"command"`
	// Subcommand, when set, must be the first non-flag argument (git push,
	// git reset). ValueFlags lets a global option that consumes the next token
	// be skipped while looking for it.
	Subcommand string `json:"subcommand,omitempty"`
	// ValueFlags are flags of Command that consume the FOLLOWING token, so the
	// scan for Subcommand steps over the value rather than reading it as the
	// subcommand (`git -C /repo push`).
	ValueFlags []string `json:"value_flags,omitempty"`
	// Flags are argument constraints. Each element is one alternation group
	// written with "|" ("--force|-f"); every group must be satisfied by some
	// argument token. A long alternative also matches its --flag=value form; a
	// single-letter short alternative also matches inside a bundled cluster
	// (-rf satisfies -r and -f).
	Flags []string `json:"flags,omitempty"`
	// AfterCD, when true, additionally requires that some EARLIER command in the
	// same chain is a `cd` — the cd-chain structure (`cd scratch && rm -rf *`)
	// whose hazard is that a failed cd silently redirects the command. A nil
	// pointer means false; it is a pointer so a per-repo override can turn the
	// requirement off as well as on.
	AfterCD *bool `json:"after_cd,omitempty"`
}

// Fixtures is an entry's proof corpus: command lines it must fire on, and
// command lines it must not. The admission gate in fixtures_test.go enforces
// both sides plus the 40% known-good floor.
type Fixtures struct {
	KnownBad  []string `json:"known_bad,omitempty"`
	KnownGood []string `json:"known_good,omitempty"`
}

// Entry is one hazard. The map key in Registry.Entries is its id; ID is filled
// from that key on load, so the key is always authoritative.
type Entry struct {
	ID        string   `json:"-"`
	Pattern   Pattern  `json:"pattern"`
	Tier      string   `json:"tier"`
	Successor string   `json:"successor"`
	Why       string   `json:"why"`
	Fixtures  Fixtures `json:"fixtures,omitempty"`
}

// Registry is the merged, validated hazard model: the bundled defaults overlaid
// with the per-repo override.
type Registry struct {
	SchemaVersion int              `json:"schema_version"`
	Disabled      bool             `json:"disabled"`
	Entries       map[string]Entry `json:"entries"`
}

// Decision is what core returns for one candidate command. It carries no
// formatting and no exit status — the front door decides how to render it, and
// the hook shim decides what to do with a broken guard.
type Decision struct {
	Verdict Verdict `json:"verdict"`
	// EntryID, Tier, Successor, and Why describe the winning entry; they are
	// empty on an allow.
	EntryID   string `json:"entry_id,omitempty"`
	Tier      string `json:"tier,omitempty"`
	Successor string `json:"successor,omitempty"`
	Why       string `json:"why,omitempty"`
	// Message is the ready-to-surface sentence: the why plus the successor, so
	// the refusal itself teaches the safe form.
	Message string `json:"message,omitempty"`
	// Matches lists every entry the command tripped (blocker ids first, then
	// warn, each in id order), so a warn plane can surface all of them.
	Matches []string `json:"matches,omitempty"`
}

//go:embed defaults/guard.json
var defaultsJSON []byte

// defaultRegistry is parsed once at init; a malformed embedded asset is a build
// error surfaced as a panic (it can never happen at runtime).
var defaultRegistry = mustParseDefaults()

func mustParseDefaults() Registry {
	r, err := parse(defaultsJSON)
	if err != nil {
		panic("guard: bundled defaults are malformed: " + err.Error())
	}
	if err := Validate(r); err != nil {
		panic("guard: bundled defaults fail validation: " + err.Error())
	}
	return r
}

// Defaults returns a deep copy of the binary-bundled registry, safe for the
// caller to mutate.
func Defaults() Registry { return cloneRegistry(defaultRegistry) }

// parse decodes registry JSON and stamps each entry's ID from its map key.
func parse(data []byte) (Registry, error) {
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return Registry{}, fmt.Errorf("%w: %v", ErrMalformedConfig, err)
	}
	for id, e := range r.Entries {
		e.ID = id
		r.Entries[id] = e
	}
	return r, nil
}

// Validate checks the registry schema: the schema version, every entry id, and
// every entry's tier, pattern command, successor, and why. A per-repo override
// is validated AFTER merging, so an override can never produce an entry the
// bundled defaults would not have been allowed to ship.
func Validate(r Registry) error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: schema_version must be %d, got %d", ErrSchemaVersion, SchemaVersion, r.SchemaVersion)
	}
	ids := make([]string, 0, len(r.Entries))
	for id := range r.Entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if !validEntryID(id) {
			return fmt.Errorf("%w: entry id %q must match [a-z0-9][a-z0-9-]*", ErrInvalidEntry, id)
		}
		e := r.Entries[id]
		switch e.Tier {
		case TierBlocker, TierWarn:
		default:
			return fmt.Errorf("%w: entry %s has tier %q, want %q or %q", ErrUnknownTier, id, e.Tier, TierBlocker, TierWarn)
		}
		if strings.TrimSpace(e.Pattern.Command) == "" {
			return fmt.Errorf("%w: entry %s has no pattern command", ErrInvalidEntry, id)
		}
		// A refusal with no successor leaves its replacement in prose only, and
		// one with no why cannot teach — both are load-time rejections, as in
		// the record-lint banned_tokens family (iss-51).
		if strings.TrimSpace(e.Successor) == "" {
			return fmt.Errorf("%w: entry %s has no successor", ErrInvalidEntry, id)
		}
		if strings.TrimSpace(e.Why) == "" {
			return fmt.Errorf("%w: entry %s has no why", ErrInvalidEntry, id)
		}
	}
	return nil
}

func validEntryID(id string) bool {
	if id == "" {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '-' && i > 0:
		default:
			return false
		}
	}
	return true
}

// Check evaluates one candidate command line against the registry. A disabled
// registry allows everything (the escape is committed and reviewable, never an
// in-session override). A command line the tokenizer cannot split is an error,
// not an allow — core names the failure and the caller decides what it means.
func (r Registry) Check(command string) (Decision, error) {
	if r.Disabled {
		return Decision{Verdict: VerdictAllow}, nil
	}
	segs, err := tokenize(command)
	if err != nil {
		return Decision{}, err
	}
	ids := make([]string, 0, len(r.Entries))
	for id := range r.Entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var blockers, warns []string
	for _, id := range ids {
		if matchesAny(r.Entries[id].Pattern, segs) {
			if r.Entries[id].Tier == TierBlocker {
				blockers = append(blockers, id)
			} else {
				warns = append(warns, id)
			}
		}
	}
	matches := append(append([]string(nil), blockers...), warns...)
	if len(matches) == 0 {
		return Decision{Verdict: VerdictAllow}, nil
	}
	// A blocker outranks a warn; within a tier the id order is the tiebreak, so
	// the decision is deterministic for the same registry and command.
	win := r.Entries[matches[0]]
	verdict := VerdictWarn
	if win.Tier == TierBlocker {
		verdict = VerdictBlock
	}
	return Decision{
		Verdict:   verdict,
		EntryID:   win.ID,
		Tier:      win.Tier,
		Successor: win.Successor,
		Why:       win.Why,
		Message:   message(verdict, win),
		Matches:   matches,
	}, nil
}

// message renders the sentence a front door surfaces: what happened, why in
// plain language, and what to run instead. The refusal is the lesson, so the
// successor is always part of it.
func message(v Verdict, e Entry) string {
	lead := "Blocked"
	if v == VerdictWarn {
		lead = "Warning"
	}
	return fmt.Sprintf("%s by the abcd guard (%s): %s Run instead: %s", lead, e.ID, e.Why, e.Successor)
}

func cloneRegistry(r Registry) Registry {
	out := Registry{SchemaVersion: r.SchemaVersion, Disabled: r.Disabled}
	if r.Entries != nil {
		out.Entries = make(map[string]Entry, len(r.Entries))
		for id, e := range r.Entries {
			out.Entries[id] = cloneEntry(e)
		}
	}
	return out
}

func cloneEntry(e Entry) Entry {
	out := e
	out.Pattern.ValueFlags = append([]string(nil), e.Pattern.ValueFlags...)
	out.Pattern.Flags = append([]string(nil), e.Pattern.Flags...)
	if e.Pattern.AfterCD != nil {
		v := *e.Pattern.AfterCD
		out.Pattern.AfterCD = &v
	}
	out.Fixtures.KnownBad = append([]string(nil), e.Fixtures.KnownBad...)
	out.Fixtures.KnownGood = append([]string(nil), e.Fixtures.KnownGood...)
	return out
}
