package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// maxPromptBytes bounds how much prompt text is scanned for recall, so a huge
// pasted prompt cannot blow up matching (trust boundary).
const maxPromptBytes = 64 * 1024

// maxStateFileBytes caps a session-state ledger file on read.
const maxStateFileBytes = 256 * 1024

// DefaultRefreshBackstop is the fixed-N full-refresh backstop (D1): the primary
// refresh is event-driven (a SessionStart/PreCompact reset clears the ledger),
// and this large counter only catches always-relevant domains that never
// recall-match. It is deliberately larger than CARL's every-5.
const DefaultRefreshBackstop = 15

// InjectResult is the outcome of one prompt-router evaluation. Text is empty
// when nothing new is injected (a healthy no-match renders zero model-facing
// tokens, per D3).
type InjectResult struct {
	Text     string
	Injected []string
	State    SessionState
}

// SessionState is the per-session dedup ledger plus the prompt counter that
// drives the fixed-N refresh backstop.
type SessionState struct {
	Count  int               `json:"count"`
	Ledger map[string]string `json:"ledger"` // domain name -> last-injected signature
}

// Inject is the pure heart of the prompt-router hook: it recall-matches prompt,
// drops any matched domain already injected this session with an unchanged
// signature, renders the remainder, and returns the updated session state. It
// performs no I/O and never reflects prompt bytes into the output — the rendered
// text is abcd's own rule content only.
//
// backstop is the fixed-N full-refresh interval; <= 0 uses DefaultRefreshBackstop.
// When the (incremented) prompt count is a multiple of the backstop, the ledger
// is cleared first so always-relevant domains re-inject.
func Inject(rs RuleSet, prompt string, prev SessionState, backstop int) InjectResult {
	if backstop <= 0 {
		backstop = DefaultRefreshBackstop
	}
	if len(prompt) > maxPromptBytes {
		prompt = prompt[:maxPromptBytes]
	}

	ledger := map[string]string{}
	for k, v := range prev.Ledger {
		ledger[k] = v
	}
	count := prev.Count + 1
	if count%backstop == 0 {
		ledger = map[string]string{}
	}

	var fresh []ResolvedDomain
	var injected []string
	for _, d := range rs.Match(prompt) {
		sig := Signature(d)
		if ledger[d.Name] == sig {
			continue // already injected this session, unchanged
		}
		ledger[d.Name] = sig
		fresh = append(fresh, d)
		injected = append(injected, d.Name)
	}
	sort.Strings(injected)

	return InjectResult{
		Text:     Render(fresh),
		Injected: injected,
		State:    SessionState{Count: count, Ledger: ledger},
	}
}

// stateDir is the machine-local directory holding per-session ledgers. It is
// overridable via ABCD_RULES_STATE_DIR (used by tests and by operators who want
// the state off the default temp dir).
func stateDir() string {
	if d := os.Getenv("ABCD_RULES_STATE_DIR"); d != "" {
		return d
	}
	return filepath.Join(os.TempDir(), "abcd-rules-state")
}

// sessionFile maps a session id to a state file. The id is hashed, so an
// attacker-supplied session id can never traverse out of the state dir.
func sessionFile(session string) string {
	sum := sha256.Sum256([]byte(session))
	return filepath.Join(stateDir(), hex.EncodeToString(sum[:])+".json")
}

// LoadState reads the ledger for a session. A missing, oversized, or malformed
// file yields the zero state (a fresh session), never an error — dedup is a
// best-effort optimisation, not a correctness gate.
func LoadState(session string) SessionState {
	fi, err := os.Lstat(sessionFile(session))
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > maxStateFileBytes {
		return SessionState{}
	}
	data, err := os.ReadFile(sessionFile(session))
	if err != nil {
		return SessionState{}
	}
	var st SessionState
	if err := json.Unmarshal(data, &st); err != nil {
		return SessionState{}
	}
	if st.Ledger == nil {
		st.Ledger = map[string]string{}
	}
	return st
}

// SaveState persists the ledger for a session (0700 dir, 0600 file).
func SaveState(session string, st SessionState) error {
	if err := os.MkdirAll(stateDir(), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	tmp := sessionFile(session) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, sessionFile(session))
}

// ResetState clears a session's ledger (the event-driven refresh: the next
// prompt re-injects every matching domain). A missing file is not an error.
func ResetState(session string) error {
	err := os.Remove(sessionFile(session))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
