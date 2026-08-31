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
// move; the bundle is restamped anyway. That is a known consequence of the
// shared constant, accepted rather than fixed inside a change that needed only
// one half of it — splitting the two is a larger change, and making the split
// silently is how a shape version stops meaning anything (spc-68).
const SchemaVersion = 2

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
	Type          string       `json:"_type"`
	SchemaVersion int          `json:"schema_version"`
	Position      Position     `json:"position"`
	Items         []BundleItem `json:"items"`
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
	Type             string         `json:"_type"`
	SchemaVersion    int            `json:"schema_version"`
	RunID            string         `json:"run_id"`
	Position         Position       `json:"position"`
	TargetCommit     string         `json:"target_commit"`
	AssemblerVersion string         `json:"assembler_version"`
	Items            []ManifestItem `json:"items"`
	Exclusions       []Exclusion    `json:"exclusions"`
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
