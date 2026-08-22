package site

// `site-src/ui.json` — the complete list of words the generator may add.
//
// adr-47 decision 2 says every sentence the site renders is a span of a
// repository file, and that the only words the generator may add are the
// interface strings in this file plus numbers, dates, file names and asset
// names. That makes ui.json a CLOSED allowlist, and the way it is kept closed is
// this struct: the file is decoded with unknown fields refused, so a string
// added to the JSON that no field here reads fails the build. Adding a word to
// the site therefore costs a code change and a review, which is the price the
// single-source rule is meant to charge.

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// maxUIBytes bounds the allowlist read.
const maxUIBytes = 64 * 1024

// ErrUIInvalid is returned for an unusable interface-string file.
var ErrUIInvalid = errors.New("site: interface strings are invalid")

// UI is the interface-string allowlist.
type UI struct {
	Purpose       string   `json:"_purpose"`
	NavStory      string   `json:"nav_story"`
	NavInstall    string   `json:"nav_install"`
	NavDocs       string   `json:"nav_docs"`
	NavRecord     string   `json:"nav_record"`
	NavReferences string   `json:"nav_references"`
	CTARoles      string   `json:"cta_roles"`
	CTAInstall    string   `json:"cta_install"`
	CTADocs       string   `json:"cta_docs"`
	RecordLink    string   `json:"record_link"`
	FromTheRecord string   `json:"from_the_record"`
	LatestRelease string   `json:"latest_release"`
	AllReleases   string   `json:"all_releases"`
	Copy          string   `json:"copy"`
	Copied        string   `json:"copied"`
	SearchDocs    string   `json:"search_docs"`
	Platform      Platform `json:"platform"`
	Tiles         Tiles    `json:"tiles"`
	More          string   `json:"more"`
	Standby       string   `json:"standby"`
	CLIGroup      string   `json:"cli_group"`
	MatchesSystem string   `json:"matches_system"`
	Beta          string   `json:"beta"`
}

// Platform names each released binary in the reader's own words.
type Platform struct {
	DarwinARM64 string `json:"darwin-arm64"`
	DarwinAMD64 string `json:"darwin-amd64"`
	LinuxARM64  string `json:"linux-arm64"`
	LinuxAMD64  string `json:"linux-amd64"`
}

// Tiles captions the record dashboard's counts.
type Tiles struct {
	Releases string `json:"releases"`
	ADR      string `json:"adr"`
	Intent   string `json:"intent"`
	Issue    string `json:"issue"`
	Commits  string `json:"commits"`
}

// LoadUI reads the interface-string allowlist named by the manifest.
func LoadUI(repoRoot, rel string) (UI, error) {
	data, err := fsutil.ReadGuarded(joinRepo(repoRoot, rel), maxUIBytes)
	if err != nil {
		return UI{}, err
	}
	var ui UI
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ui); err != nil {
		return UI{}, fmt.Errorf("%w: %s: %v", ErrUIInvalid, rel, err)
	}
	if missing := ui.missing(); len(missing) > 0 {
		return UI{}, fmt.Errorf("%w: %s: no text for %s — every interface string the site renders is declared here",
			ErrUIInvalid, rel, strings.Join(missing, ", "))
	}
	return ui, nil
}

// missing names every interface string the file leaves empty. An empty string
// renders as a blank button or an unlabelled tab, which reads as a rendering
// bug rather than as the missing declaration it is.
func (ui UI) missing() []string {
	var out []string
	var walk func(v reflect.Value, t reflect.Type)
	walk = func(v reflect.Value, t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			switch f.Type.Kind() {
			case reflect.Struct:
				walk(v.Field(i), f.Type)
			case reflect.String:
				// `_purpose` documents the file for its human readers and is
				// never rendered, so it is the one field with nothing to say.
				if name != "_purpose" && strings.TrimSpace(v.Field(i).String()) == "" {
					out = append(out, name)
				}
			}
		}
	}
	walk(reflect.ValueOf(ui), reflect.TypeOf(ui))
	return out
}
