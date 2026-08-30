package reading

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SchemaVersion is the shape version of both artefacts an assembly writes.
const SchemaVersion = 1

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
// eval falsifies independently.
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
	SHA256  string `json:"sha256"`
}

// Manifest enumerates what an assembly passed, by path, by field and by hash,
// and asserts what it refused. It carries no item content and no timestamp of
// any kind, so committing it needs no redaction and one repository state
// produces one manifest.
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
