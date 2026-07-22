package lint

import (
	"os"
	"path/filepath"
	"strings"
)

// IntentLink is one intent record as the link index sees it: which bucket holds
// it (the lifecycle state, because directory-as-truth) and which spec it names.
type IntentLink struct {
	// ID is the intent id (itd-N), taken from the frontmatter when well-formed
	// and from the filename otherwise.
	ID string
	// Bucket is the lifecycle directory holding it (drafts, planned, shipped,
	// disciplines, superseded), or "" for a file outside a known bucket.
	Bucket string
	// Path is the record's repo-relative path.
	Path string
	// SpecID is the raw spec_id frontmatter value (which may be a YAML null).
	SpecID string
}

// SpecLink is one spec record as the link index sees it. The unexported fields
// carry what the spec-lifecycle lint needs from the same read — the frontmatter
// with its line numbers, and whether the file is content-exempt — so the tree is
// walked once for both consumers.
type SpecLink struct {
	// ID is the spec id (spc-N) from frontmatter, "" when malformed or absent.
	ID string
	// Bucket is the lifecycle directory holding it: "open" or "closed".
	Bucket string
	// Path is the spec's repo-relative path.
	Path string
	// IntentID is the raw intent frontmatter value — the back-link.
	IntentID string

	fields map[string]fmField
	exempt bool
}

// SpecLinkIndex is ONE traversal of the intent buckets and the spec store,
// shared by every consumer that must reason about the intent↔spec link.
//
// It exists because two consumers ask opposite questions of the same two trees.
// spec_lifecycle walks the specs and asks whether each names an intent that
// agrees with it; the release cut walks the intents and asks whether any still
// sitting in planned/ has a spec that has already CLOSED — a merged feature whose
// record never moved, which is invisible to the shipped/-tree diff and would
// silently under-bump the release. Two walks would be two answers about one tree,
// so there is one scan and two readings of it.
type SpecLinkIndex struct {
	Intents []IntentLink
	Specs   []SpecLink
}

// IntentSpecID maps each known intent id to its raw spec_id value.
func (x SpecLinkIndex) IntentSpecID() map[string]string {
	out := make(map[string]string, len(x.Intents))
	for _, i := range x.Intents {
		out[i.ID] = i.SpecID
	}
	return out
}

// KnownIntents is the set of intent ids present anywhere in the corpus.
func (x SpecLinkIndex) KnownIntents() map[string]bool {
	out := make(map[string]bool, len(x.Intents))
	for _, i := range x.Intents {
		out[i.ID] = true
	}
	return out
}

// SpecBucket resolves a spec_id value to the lifecycle bucket holding that spec.
//
// Matching is on the spec NUMBER, not the literal string, because a spec_id is
// written both bare (`spc-9`) and with its slug (`spc-9-thing`) across the
// record; comparing literals would silently fail to resolve half the corpus and
// make a fail-closed gate fail open.
func (x SpecLinkIndex) SpecBucket(specID string) (string, bool) {
	n := specNum(specID)
	if n < 0 {
		return "", false
	}
	for _, s := range x.Specs {
		if specNum(s.ID) == n {
			return s.Bucket, true
		}
	}
	return "", false
}

// ScanSpecLinks reads the intent buckets and the spec store once, both relative
// to repoRoot. A missing tree contributes nothing and is not an error, mirroring
// the rest of the record lint: an unpopulated repository is a state, not a fault.
//
// top supplies the content exemptions. They are recorded per spec rather than
// applied here, because they exempt a file from the lint's CONTENT checks — they
// do not mean the record stopped existing, which is the only thing the release
// cut asks about.
func ScanSpecLinks(repoRoot, intentsDir, specsDir string, top Config) (SpecLinkIndex, error) {
	var idx SpecLinkIndex

	intentsRoot := filepath.Join(repoRoot, filepath.FromSlash(intentsDir))
	_ = filepath.WalkDir(intentsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !intentFileRe.MatchString(d.Name()) {
			return nil
		}
		content, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		fields := frontmatterFields(strings.Split(string(content), "\n"))
		id := fields["id"].value
		if !intentIDFullRe.MatchString(id) {
			id = intentIDRe.FindString(d.Name())
		}
		if id == "" {
			return nil
		}
		bucket := filepath.Base(filepath.Dir(path))
		if !intentBuckets[bucket] {
			bucket = ""
		}
		idx.Intents = append(idx.Intents, IntentLink{
			ID:     id,
			Bucket: bucket,
			Path:   repoRel(repoRoot, path),
			SpecID: fields["spec_id"].value,
		})
		return nil
	})

	specsRoot := filepath.Join(repoRoot, filepath.FromSlash(specsDir))
	for _, bucket := range []string{"open", "closed"} {
		entries, err := os.ReadDir(filepath.Join(specsRoot, bucket))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return SpecLinkIndex{}, err
		}
		for _, e := range entries {
			if e.IsDir() || !specFileRe.MatchString(e.Name()) {
				continue
			}
			fileAbs := filepath.Join(specsRoot, bucket, e.Name())
			content, err := os.ReadFile(fileAbs)
			if err != nil {
				return SpecLinkIndex{}, err
			}
			rel := repoRel(repoRoot, fileAbs)
			fields := frontmatterFields(strings.Split(string(content), "\n"))
			idx.Specs = append(idx.Specs, SpecLink{
				ID:       fields["id"].value,
				Bucket:   bucket,
				Path:     rel,
				IntentID: fields["intent"].value,
				fields:   fields,
				exempt:   contentExempt(rel, fields, top),
			})
		}
	}
	return idx, nil
}
