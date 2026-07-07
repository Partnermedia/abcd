package launch

import (
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
)

// versionLocationRelPath is the committed version-location decision artefact.
const versionLocationRelPath = ".abcd/config/version-location.json"

// DryRunRequest is the input to a dry-run.
type DryRunRequest struct {
	RepoRoot        string
	VersionOverride string   // --version; empty → read from the resolved manifests
	ExistingTags    []Semver // injected; nil → default `git tag -l v*` provider
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

	vlPath := filepath.Join(req.RepoRoot, versionLocationRelPath)
	lockstep := CheckLockstep(TreePublic, req.RepoRoot, vlPath)
	report.Lockstep = lockstep

	version := resolveVersion(req.VersionOverride, req.RepoRoot, lockstep)
	report.Version = version

	report.Retention = computeRetentionForReport(version, req)

	report.Gates = []GateSummary{
		{Name: "secret+pii-scan", Status: "ran", Detail: scanDetail(scan)},
		{Name: "marker-block", Status: "not_implemented", Detail: "Phase-5 deferred"},
		{Name: "plugin.json-parse", Status: "not_implemented", Detail: "Phase-5 deferred"},
		{Name: "documentation-auditor", Status: "not_implemented", Detail: "Phase-5 deferred"},
	}

	report.WouldRefuseOn = wouldRefuseOn(bundle, scan, lockstep, report.Retention)
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

// resolveVersion picks the version input: an explicit override, else the
// lockstep primary version when readable.
func resolveVersion(override, repoRoot string, lockstep LockstepResult) string {
	if override != "" {
		return override
	}
	if lockstep.OK {
		if v := primaryVersion(repoRoot); v != "" {
			return v
		}
	}
	return ""
}

// primaryVersion reads the primary version string via the version-location
// contract (best-effort; empty when unreadable).
func primaryVersion(repoRoot string) string {
	vlPath := filepath.Join(repoRoot, versionLocationRelPath)
	decision, err := loadJSON(vlPath)
	if err != nil {
		return ""
	}
	mp, ptr, verr := validateVersionLocation(decision)
	if verr != "" {
		return ""
	}
	doc, err := loadJSON(filepath.Join(repoRoot, mp))
	if err != nil {
		return ""
	}
	v, present := resolvePointer(doc, ptr)
	if !present {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
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
// bundle rejections, lockstep drift/unreadable, and retention refusal.
func wouldRefuseOn(bundle Bundle, scan scanner.ScanResult, lockstep LockstepResult, retention RetentionPlan) []string {
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
