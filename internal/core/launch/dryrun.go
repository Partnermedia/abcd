package launch

import (
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
)

// versionLocationRelPath is the committed version-location decision artefact.
const versionLocationRelPath = ".abcd/config/version-location.json"

// DryRunRequest is the input to a dry-run.
type DryRunRequest struct {
	RepoRoot string
	// Version is the release version this launch would publish, SUPPLIED by the
	// caller. adr-19 leaves no version key in the source tree, so there is
	// nothing here for the core to read: the version is a fact about the release
	// cut, and the front door that knows the cut injects it. Empty is honest —
	// it means the caller could not name one, and retention says so.
	Version      string
	ExistingTags []Semver // injected; nil → default `git tag -l v*` provider
}

// GateSummary records one gate's dry-run disposition.
type GateSummary struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "ran" | "not_implemented"
	Detail string `json:"detail"`
}

// DryRunReport is the full dry-run preview. No artefact is written.
type DryRunReport struct {
	Version       string             `json:"version"`
	Bundle        Bundle             `json:"bundle"`
	Scan          scanner.ScanResult `json:"scan"`
	Lockstep      LockstepResult     `json:"lockstep"`
	Retention     RetentionPlan      `json:"retention"`
	Smoke         SmokeReport        `json:"smoke"`
	Gates         []GateSummary      `json:"gates"`
	WouldPublish  bool               `json:"would_publish"` // always false in dry-run
	WouldRefuseOn []string           `json:"would_refuse_on,omitempty"`
}

// DryRun assembles the bundle, scans it, checks lockstep and previews retention,
// then reports what a real ship WOULD refuse on. It ALWAYS returns exit-0
// semantics: an error is returned only for a preflight fault (bad include config)
// that makes a report impossible — never on a finding.
func DryRun(req DryRunRequest) (DryRunReport, error) {
	var report DryRunReport

	bundle, err := ResolveBundle(req.RepoRoot, nil)
	if err != nil {
		return DryRunReport{}, err // preflight fault only
	}
	report.Bundle = bundle

	scan := scanBundle(req.RepoRoot, bundle)
	report.Scan = scan

	// The DEV polarity is the one the SOURCE TREE must satisfy: adr-19 keeps the
	// version keys out of the committed manifests, so a public check here would
	// accuse a correct repository of drift and prescribe the exact key the ADR
	// forbids. The public polarity belongs over the rendered payload, where
	// RenderPayload applies it to its own output.
	vlPath := filepath.Join(req.RepoRoot, versionLocationRelPath)
	lockstep := CheckLockstep(TreeDev, req.RepoRoot, vlPath)
	report.Lockstep = lockstep

	report.Version = req.Version
	report.Retention = computeRetentionForReport(req.Version, req)

	// The smoke reads the RESOLVED BUNDLE, not the working tree: a file present
	// in the tree but excluded from the payload is exactly the break it exists
	// to catch. It subsumes itd-65's placeholder `plugin.json-parse` gate, which
	// asserted a strict subset of what the light tier asserts.
	smoke := SmokeLight(NewBundleTree(bundle))
	report.Smoke = smoke

	report.Gates = []GateSummary{
		{Name: "secret+pii-scan", Status: "ran", Detail: scanDetail(scan)},
		{Name: "marker-block", Status: "not_implemented", Detail: "Phase-5 deferred"},
		{Name: "installability-smoke", Status: "ran", Detail: smokeDetail(smoke)},
		{Name: "documentation-auditor", Status: "not_implemented", Detail: "Phase-5 deferred"},
	}

	report.WouldRefuseOn = wouldRefuseOn(bundle, scan, lockstep, report.Retention, smoke)
	report.WouldPublish = false
	return report, nil
}

// scanBundle adapts the bundle's Included files to the scanner and runs it.
func scanBundle(repoRoot string, bundle Bundle) scanner.ScanResult {
	sc, err := scanner.New(repoRoot)
	if err != nil {
		return scanner.ScanResult{Unavailable: true, UnavailableReason: err.Error()}
	}
	files := make([]scanner.BundleFile, 0, len(bundle.Included))
	for _, f := range bundle.Included {
		files = append(files, scanner.BundleFile{LogicalPath: f.LogicalPath, ResolvedPath: f.ResolvedPath})
	}
	res, _ := sc.ScanBundle(files)
	return res
}

// computeRetentionForReport builds the retention preview, resolving the existing
// tag list from the injected slice or the default git provider.
func computeRetentionForReport(version string, req DryRunRequest) RetentionPlan {
	pub, err := ParseSemver(version)
	if err != nil {
		return RetentionPlan{
			Published: "v" + version, Refused: true,
			RefusalReason: "published version is not strict SemVer: " + version,
		}
	}
	existing := req.ExistingTags
	if existing == nil {
		existing, _ = GitExistingTags(req.RepoRoot)
	}
	return ComputeRetention(pub, existing)
}

// wouldRefuseOn collects everything a real ship WOULD block on: scan hard-fails,
// bundle rejections, lockstep drift/unreadable, retention refusal, and an
// uninstallable declared surface.
func wouldRefuseOn(bundle Bundle, scan scanner.ScanResult, lockstep LockstepResult, retention RetentionPlan, smoke SmokeReport) []string {
	var reasons []string
	if scan.Unavailable {
		reasons = append(reasons, "scanner unavailable: "+scan.UnavailableReason)
	}
	if scan.HardFails > 0 {
		reasons = append(reasons, hardFailReason(scan))
	}
	for _, r := range bundle.Rejected {
		reasons = append(reasons, "bundle rejected: "+r.LogicalPath+" ("+string(r.Reason)+")")
	}
	if lockstep.Unreadable {
		reasons = append(reasons, "lockstep contract unreadable: "+lockstep.Detail)
	}
	for _, d := range lockstep.Drifts {
		reasons = append(reasons, "lockstep drift: "+d)
	}
	if retention.Refused {
		reasons = append(reasons, "retention refused: "+retention.RefusalReason)
	}
	reasons = append(reasons, smokeRefusals(smoke)...)
	return reasons
}

func hardFailReason(scan scanner.ScanResult) string {
	n := 0
	for _, f := range scan.Findings {
		if f.Severity == scanner.SeverityHardFail {
			n++
		}
	}
	return "secret/PII hard-fail findings: " + itoa(n)
}

func scanDetail(scan scanner.ScanResult) string {
	if scan.Unavailable {
		return "unavailable: " + scan.UnavailableReason
	}
	return "scanned " + itoa(scan.FilesScanned) + " files, " + itoa(scan.HardFails) + " hard-fails"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
