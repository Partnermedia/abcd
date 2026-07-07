package launch

import (
	"errors"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
)

// ShipRequest is the input to a ship run.
type ShipRequest struct {
	RepoRoot        string
	VersionOverride string
	AllowDirty      bool
	ExistingTags    []Semver
}

// ShipReport is the outcome of a ship run. It stops at WouldPublish before any
// network/publish (no GitHub Release, SLSA, tag push, or retention execution).
type ShipReport struct {
	Version      string             `json:"version"`
	Bundle       Bundle             `json:"bundle"`
	Scan         scanner.ScanResult `json:"scan"`
	Lockstep     LockstepResult     `json:"lockstep"`
	Retention    RetentionPlan      `json:"retention"`
	Blocked      bool               `json:"blocked"`
	BlockReasons []string           `json:"block_reasons,omitempty"`
	WouldPublish bool               `json:"would_publish"` // true iff all gates pass
}

// ErrShipBlocked is returned when any gate hard-fails.
var ErrShipBlocked = errors.New("ship blocked by a launch gate")

// Ship runs the SAME gates as DryRun but HARD-FAILS: a scanner hard-fail, any
// bundle rejected[] entry, a lockstep drift/unreadable contract, or a retention
// refusal sets Blocked=true and returns ErrShipBlocked. If every gate passes it
// stops HERE and returns WouldPublish=true with NO network call — the real
// GitHub Release + SLSA + tag push + retention prune are a later phase.
// --allow-dirty must NOT bypass lockstep (adr-20); it is carried but never
// consulted by the lockstep gate.
func Ship(req ShipRequest) (ShipReport, error) {
	var report ShipReport

	bundle, err := ResolveBundle(req.RepoRoot, nil)
	if err != nil {
		return ShipReport{}, err // preflight fault
	}
	report.Bundle = bundle

	scan := scanBundle(req.RepoRoot, bundle)
	report.Scan = scan

	vlPath := filepath.Join(req.RepoRoot, versionLocationRelPath)
	lockstep := CheckLockstep(TreePublic, req.RepoRoot, vlPath)
	report.Lockstep = lockstep

	version := resolveVersion(req.VersionOverride, req.RepoRoot, lockstep)
	report.Version = version
	report.Retention = computeRetentionForReport(version, DryRunRequest{
		RepoRoot: req.RepoRoot, VersionOverride: req.VersionOverride, ExistingTags: req.ExistingTags,
	})

	report.BlockReasons = wouldRefuseOn(bundle, scan, lockstep, report.Retention)
	if len(report.BlockReasons) > 0 {
		report.Blocked = true
		report.WouldPublish = false
		return report, ErrShipBlocked
	}
	report.WouldPublish = true
	return report, nil
}
