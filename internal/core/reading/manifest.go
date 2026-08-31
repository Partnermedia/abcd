package reading

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SchemaVersion is the shape version of both artefacts an assembly writes.
//
// It is ONE constant for two shapes, so a change to either restamps both. At
// version 2 the manifest item gained a kind and the bundle's shape did not
// move; the bundle was restamped anyway. At version 3 BOTH shapes moved, the
// bundle gaining the scope a reading was given and the manifest gaining the
// effective scope, its hash and the override stamp — so at this version the
// shared constant costs nothing. That is a known consequence of the
// shared constant, accepted rather than fixed inside a change that needed only
// one half of it — splitting the two is a larger change, and making the split
// silently is how a shape version stops meaning anything (spc-68).
const SchemaVersion = 3

// The two artefact type tags. They are carried in the documents themselves so a
// reader of a loose file can tell the two apart without its filename.
const (
	BundleType   = "abcd.reading.bundle"
	ManifestType = "abcd.reading.manifest"
)

// BundleItem is one passed item as the reading receives it: a key, a material
// class, and the text. It carries NO repository path, by construction — the key
// is an ordinal and the kind names a class, never a location (invariant 15).
type BundleItem struct {
	ItemKey string `json:"item_key"`
	Kind    Kind   `json:"kind"`
	Text    string `json:"text"`
}

// Bundle is the assembled input: the reading's entire working set.
//
// It carries no run identifier and no timestamp, so two assemblies of one
// repository state at one commit are byte-identical — the property itd-187's
// eval falsifies independently, and the reason the run identifier lives on the
// manifest alone.
type Bundle struct {
	Type          string   `json:"_type"`
	SchemaVersion int      `json:"schema_version"`
	Position      Position `json:"position"`
	// Scope is what THIS run was given, and it is the reading's own fact
	// rather than the auditor's. A reader told its object is the shipped tree
	// and handed a tenth of it will report the missing nine tenths as a
	// tension against the claim record, with every gate green.
	//
	// It carries NO repository path under any scope, and no provenance. See
	// BundleScope: it is a projection rather than the resolved scope precisely
	// because the obvious implementation carried a path, and which token the
	// operator typed — and whether that departed from the presets — is the
	// auditor's business and lives on the manifest.
	Scope BundleScope  `json:"scope"`
	Items []BundleItem `json:"items"`
}

// BundleScope is the scope as a READING sees it, and it is deliberately NOT
// the Scope the manifest carries.
//
// The manifest may name repository paths; the bundle may not, by brief
// invariant 15 — the assembled input is the reading's entire working set and
// no repository path enters its context. A scope's Path selectors ARE
// repository paths, so writing one Scope type into both artefacts put a path
// into the reading's own working set. That is what this split exists to
// prevent, and it was a live breach before it was caught
// (iss-2608312058244357).
//
// A reading still has to know it was handed a subset: told its object is the
// shipped tree and given a tenth of it, it reports the missing nine tenths as
// a finding. So it is told the kinds and the records it was scoped to, and
// that a narrowing by LOCATION applied — never where. That is enough to know
// the bundle is not the whole object, and it carries no location.
type BundleScope struct {
	Kinds   []Kind   `json:"kinds,omitempty"`
	Records []string `json:"records,omitempty"`
	// LocationNarrowings counts the location-based narrowings applied. It is a
	// count and never a list, because the list would be the paths.
	LocationNarrowings int `json:"location_narrowings,omitempty"`
}

// bundleScope projects a resolved scope down to what a reading may see.
func bundleScope(s Scope) BundleScope {
	var out BundleScope
	for _, sel := range s.Selectors {
		switch {
		case sel.Kind != "":
			out.Kinds = append(out.Kinds, sel.Kind)
		case sel.Record != "":
			out.Records = append(out.Records, sel.Record)
		case sel.Path != "":
			out.LocationNarrowings++
		}
	}
	return out
}

// ManifestItem maps one bundle item back to the file and field it came from,
// with the hash of the passed text. This mapping is the auditor's, and only the
// auditor's: it is the reason the bundle can be pathless and still checkable.
type ManifestItem struct {
	ItemKey string `json:"item_key"`
	Path    string `json:"path"`
	Field   string `json:"field,omitempty"`
	// Kind is the item's material class, carried so a size report is checkable
	// against the manifest rather than asserted beside it. It is deliberately
	// NOT omitempty: an item without a kind is a defect, and a shape that can
	// omit the field cannot tell that defect from a well-formed item (spc-68).
	Kind   Kind   `json:"kind"`
	SHA256 string `json:"sha256"`
}

// Manifest enumerates what an assembly passed, by path, by field and by hash,
// and asserts what it refused. It carries no item content, so committing it
// needs no redaction.
//
// It carries no timestamp FIELD, but it is not timestamp-free and must not be
// described as such: RunID embeds a mint stamp by construction (adr-45). So two
// assemblies of one repository state at one commit produce manifests that differ
// in RunID and in nothing else — Items and Exclusions are identical, and the
// bundle beside them is byte-identical. That, not manifest byte-identity, is the
// determinism a re-run can be checked against, and it is why the manifest sits
// outside the amnesia eval's comparison rather than inside it.
type Manifest struct {
	Type             string   `json:"_type"`
	SchemaVersion    int      `json:"schema_version"`
	RunID            string   `json:"run_id"`
	Position         Position `json:"position"`
	TargetCommit     string   `json:"target_commit"`
	AssemblerVersion string   `json:"assembler_version"`
	// Scope, ScopeHash and ScopeOverridden are the auditor's account of what
	// this run was about. The hash lets a reader tell two runs apart by their
	// scope rather than by re-deriving it, and it means a preset edited later
	// can never make a past run unreadable. ScopeOverridden is false when the
	// operator named a committed preset — running as reviewed — and true when
	// they named a record or a kind directly, so drift between what is
	// committed and what people actually run is countable rather than
	// invisible.
	Scope           Scope          `json:"scope"`
	ScopeHash       string         `json:"scope_hash"`
	ScopeOverridden bool           `json:"scope_overridden"`
	Items           []ManifestItem `json:"items"`
	Exclusions      []Exclusion    `json:"exclusions"`
}

// encode is the one definition of canonical bytes for both artefacts: fixed
// field order from the struct, two-space indent, HTML escaping off, exactly one
// trailing newline. A document that encoded differently depending on how it was
// built would turn the determinism gate into a coin flip.
func encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encoding reading artefact: %w", err)
	}
	return buf.Bytes(), nil
}

// EncodeBundle renders the assembled input as canonical JSON.
func EncodeBundle(b Bundle) ([]byte, error) { return encode(b) }

// EncodeManifest renders the manifest as canonical JSON.
func EncodeManifest(m Manifest) ([]byte, error) { return encode(m) }

// DecodeManifest reads a manifest strictly: unknown fields, trailing content
// and a schema-version mismatch are all refused. All three are fail-closed on
// purpose, because a manifest is the evidence a reader judges contamination by.
//
// It has NO front door yet, and that is deliberate rather than an oversight: the
// verb that reads a manifest back is the ingest, which spc-63 owns and which is
// not in this delivery. It is exported and tested here because the writer and
// the reader of one format belong together — a decoder written later, against
// the file rather than against the encoder, is how the two drift.
func DecodeManifest(data []byte) (Manifest, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return Manifest{}, fmt.Errorf("decoding reading manifest: %w", err)
	}
	if dec.More() {
		return Manifest{}, fmt.Errorf("decoding reading manifest: trailing content after the document")
	}
	if m.Type != ManifestType {
		return Manifest{}, fmt.Errorf("decoding reading manifest: _type is %q, want %q", m.Type, ManifestType)
	}
	if m.SchemaVersion != SchemaVersion {
		return Manifest{}, fmt.Errorf("decoding reading manifest: schema_version is %d, want %d",
			m.SchemaVersion, SchemaVersion)
	}
	// An item's kind is REFUSED when absent rather than defaulted, so the
	// not-omitempty decision on the write side is a property of the format
	// rather than a habit of this writer. Without this, an item carrying no
	// kind decodes clean and the strictness the type's own comment claims is
	// true of what the binary writes and false of what it reads — an
	// attestation asserting more than its examination establishes, which brief
	// invariant 16 forbids.
	known := make(map[Kind]bool, len(Kinds()))
	for _, k := range Kinds() {
		known[k] = true
	}
	for i, it := range m.Items {
		if it.Kind == "" {
			return Manifest{}, fmt.Errorf("decoding reading manifest: item %d (%s) carries no kind",
				i, it.ItemKey)
		}
		if !known[it.Kind] {
			return Manifest{}, fmt.Errorf("decoding reading manifest: item %d (%s) carries the "+
				"unknown kind %q; the vocabulary is closed", i, it.ItemKey, it.Kind)
		}
	}
	return m, nil
}

// ManifestHash is the manifest's own content hash over its canonical bytes. It
// is the reference an ingest cites back.
func ManifestHash(m Manifest) (string, error) {
	data, err := EncodeManifest(m)
	if err != nil {
		return "", err
	}
	return sha256Hex(data), nil
}

// sha256Hex is the one hash this package computes: item text and manifest bytes.
func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
