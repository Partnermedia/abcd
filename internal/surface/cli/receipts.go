package cli

// receiptPreflight measures the semantic-pass receipts for the release preview
// (iss-2608231226342272), the same way citationPreflight measures the citation
// baseline: the front door takes the measurement and hands it to
// internal/core/launch as data, because the core does not reach back out to git
// or to a directory outside its own inputs.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/intentdriven/abcd/internal/core/launch"
	"github.com/intentdriven/abcd/internal/fsutil"
	"github.com/intentdriven/abcd/internal/gitutil"
)

// maxReceiptBytes caps one receipt read. A receipt is a small JSON verdict; the
// cap is generous for that and bounded so a pathological file in the reviews
// directory cannot stall a preview.
const maxReceiptBytes = 1 << 20 // 1 MiB

// receiptPreflight reports which detectors have a receipt recorded for HEAD.
//
// HEAD is indicative, not authoritative: release.yml arms the gate against the
// content commit it derives from the merge (`<merge>^2^`), which does not exist
// until the release branch is merged. Reporting HEAD is still the useful signal
// — on a correctly shaped release branch the receipts commit is HEAD and the
// commit it names is HEAD^ — and the rendered detail says which commit it looked
// at, so the number is never mistaken for a verdict.
//
// It reports PRESENCE only. Whether a receipt is valid — PROMOTE, a pinned judge
// model, the right detector binding, a matching manifest hash — is receipt_gate's
// judgement, and re-deciding it here would be the second trust root the preview
// exists to avoid.
func receiptPreflight(repoRoot string) *launch.ReceiptPreflight {
	head, err := gitutil.Run(repoRoot, "rev-parse", "HEAD")
	if err != nil || head == "" {
		return &launch.ReceiptPreflight{Unreadable: "cannot resolve HEAD"}
	}
	// The receipts naming a commit live in a LATER commit, so on a correctly
	// shaped release branch they sit in HEAD's tree naming HEAD^. Look for both:
	// receipts for HEAD (this commit has already been reviewed) and receipts for
	// HEAD^ (this commit IS the receipts commit). Either answers the question the
	// preview is asked, which is whether the semantic passes have been run.
	for _, rev := range []string{"HEAD^", "HEAD"} {
		sha, err := gitutil.Run(repoRoot, "rev-parse", rev)
		if err != nil || sha == "" {
			continue
		}
		found, readErr := recordedReceipts(repoRoot, sha)
		if readErr != "" {
			return &launch.ReceiptPreflight{Commit: sha, Unreadable: readErr}
		}
		if len(found) > 0 {
			return &launch.ReceiptPreflight{Commit: sha, Recorded: found}
		}
	}
	return &launch.ReceiptPreflight{Commit: head}
}

// recordedReceipts lists the detectors named by the receipts under
// .abcd/work/reviews/<sha>/. The detector comes from the receipt's
// policy.detector field, not its filename: the gate binds on content, so a
// filename-derived answer here would report a binding the gate does not honour.
func recordedReceipts(repoRoot, sha string) ([]string, string) {
	dir := filepath.Join(repoRoot, ".abcd", "work", "reviews", sha)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "" // no receipts yet is the normal case, not an error
		}
		return nil, "receipts directory unreadable"
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := fsutil.ReadGuarded(filepath.Join(dir, e.Name()), maxReceiptBytes)
		if err != nil {
			continue // an unreadable receipt is one the gate will refuse anyway
		}
		var r struct {
			Policy struct {
				Detector string `json:"detector"`
			} `json:"policy"`
		}
		if json.Unmarshal(raw, &r) != nil || r.Policy.Detector == "" {
			continue
		}
		out = append(out, r.Policy.Detector)
	}
	sort.Strings(out)
	return out, ""
}
