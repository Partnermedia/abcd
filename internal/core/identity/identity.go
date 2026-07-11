// Package identity checks that the git author identity a commit would use in a
// managed repo matches the identity pinned in .abcd/config/identity.json.
//
// It is the single source of truth for the iss-62 managed-repo identity gate:
// `ahoy doctor` surfaces a divergence as a detection gap, and the installed
// pre-commit hook calls Check to fail closed before a mis-attributed commit can
// land — so a stray repo-local override (e.g. a sandbox "Test User") is caught
// up front rather than discovered later.
package identity

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PinRelPath is the committed identity pin, relative to the repo root.
const PinRelPath = ".abcd/config/identity.json"

// Pin is the expected commit identity, committed so every checkout enforces the
// same value regardless of local git config.
type Pin struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Effective is the identity git would actually stamp on a commit in the repo,
// resolved through git's normal local > global > system layering.
type Effective struct {
	Name  string
	Email string
}

// Status is the outcome of comparing the effective identity to the pin.
type Status int

const (
	// StatusOK: a pin exists and the effective identity matches it.
	StatusOK Status = iota
	// StatusNoPin: no identity.json — the repo has not opted into the gate.
	StatusNoPin
	// StatusMismatch: a pin exists and the effective identity differs.
	StatusMismatch
	// StatusUnset: a pin exists but git has no author identity configured.
	StatusUnset
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusNoPin:
		return "no-pin"
	case StatusMismatch:
		return "mismatch"
	case StatusUnset:
		return "unset"
	default:
		return "unknown"
	}
}

// Result carries the comparison outcome and both identities for reporting.
type Result struct {
	Status    Status
	Pin       Pin
	Effective Effective
	Reason    string
}

// Blocks reports whether a pre-commit hook should refuse the commit. A mismatch
// or an unset identity blocks; a match, or an un-pinned (opted-out) repo, does
// not — an absent pin must never break commits in a repo that has not adopted
// the gate.
func (r Result) Blocks() bool {
	return r.Status == StatusMismatch || r.Status == StatusUnset
}

// LoadPin reads .abcd/config/identity.json. It returns (pin, true, nil) when the
// pin is present and well formed, (Pin{}, false, nil) when the file is absent,
// and an error when it is malformed or missing a field — validating this
// external input rather than trusting it.
func LoadPin(root string) (Pin, bool, error) {
	path := filepath.Join(root, PinRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Pin{}, false, nil
		}
		return Pin{}, false, fmt.Errorf("reading %s: %w", PinRelPath, err)
	}
	var p Pin
	if err := json.Unmarshal(data, &p); err != nil {
		return Pin{}, false, fmt.Errorf("malformed %s: %w", PinRelPath, err)
	}
	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	if p.Name == "" || p.Email == "" {
		return Pin{}, false, fmt.Errorf("%s must set both name and email", PinRelPath)
	}
	return p, true, nil
}

// WritePin writes the pin to .abcd/config/identity.json (creating the config
// directory), pretty-printed with a trailing newline. It is how a repo adopts
// the identity gate. Both fields are required.
func WritePin(root string, p Pin) error {
	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	if p.Name == "" || p.Email == "" {
		return fmt.Errorf("identity pin requires both name and email")
	}
	path := filepath.Join(root, PinRelPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// EffectiveIdentity returns the author identity git would use for a commit in
// root, via `git config` (which honours the local > global > system layering).
// Unset name or email yields an empty field, not an error.
func EffectiveIdentity(root string) (Effective, error) {
	name, err := gitConfig(root, "user.name")
	if err != nil {
		return Effective{}, err
	}
	email, err := gitConfig(root, "user.email")
	if err != nil {
		return Effective{}, err
	}
	return Effective{Name: name, Email: email}, nil
}

// gitConfig returns the trimmed value of a git config key, or "" when the key is
// unset. Git exits 1 for an unset key; that is not an error here. Any other
// failure (git absent, not a repo) is returned.
func gitConfig(root, key string) (string, error) {
	cmd := exec.Command("git", "-C", root, "config", "--get", key)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return "", nil // unset key
		}
		return "", fmt.Errorf("git config %s: %w", key, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Check resolves the effective identity, loads the pin, and compares them.
func Check(root string) (Result, error) {
	pin, pinned, err := LoadPin(root)
	if err != nil {
		return Result{}, err
	}
	eff, err := EffectiveIdentity(root)
	if err != nil {
		return Result{}, err
	}
	if !pinned {
		return Result{Status: StatusNoPin, Effective: eff, Reason: "no " + PinRelPath + "; repo has not adopted the identity gate"}, nil
	}
	if eff.Name == "" || eff.Email == "" {
		return Result{Status: StatusUnset, Pin: pin, Effective: eff, Reason: "git author identity is not configured (user.name/user.email)"}, nil
	}
	if eff.Name != pin.Name || eff.Email != pin.Email {
		return Result{
			Status: StatusMismatch, Pin: pin, Effective: eff,
			Reason: fmt.Sprintf("commit identity %q <%s> does not match the pin %q <%s>", eff.Name, eff.Email, pin.Name, pin.Email),
		}, nil
	}
	return Result{Status: StatusOK, Pin: pin, Effective: eff}, nil
}
