package ahoy

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/core/banlist"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// GuardHookRelPath is the repo-relative home of the committed name guard. It is a
// COMMITTED hook, not a .git/hooks copy: .git/hooks is per-clone and untracked, so
// a guard installed there protects one working copy and no one else's. A clone
// points git at this directory once (`git config core.hooksPath .githooks`) and
// every future checkout inherits the check.
const GuardHookRelPath = ".githooks/pre-commit"

// guardHookTemplate is the canonical scaffolded guard: the generalised form of the
// prototype this repo runs on itself, with the repo-specific gates dropped. It is
// embedded so the binary is self-contained (the marker block's precedent) and so
// scaffolding cannot depend on a plugin-root file that a broken install left behind.
//
//go:embed defaults/pre-commit
var guardHookTemplate []byte

// publicFamilySeed is the docs-lint config a repo with none inherits: the roots the
// lint walks and an EMPTY banned-names family. Empty is the point — abcd cannot know
// which names a repo may not publish, and seeding a ban nobody declared would fail a
// build over a word the maintainer never chose. The array is what `abcd banlist add
// --public` writes into, so its presence is what makes the public layer usable.
const publicFamilySeed = `{
  "roots": ["docs", "README.md"],
  "banned_tokens": [],
  "rules": {},
  "exempt_paths": [],
  "exempt_if_status": []
}
`

// privateStubBody is the scaffolded private banlist: the format declaration, the
// format's documentation, and worked examples that are ALL COMMENTED OUT.
//
// Commented, because a scaffolded live entry would be a ban the user never declared;
// and because the guard must report this machine as unprotected until someone opts
// in, which is exactly what a zero-entry store makes it do (its loud NO ENTRIES
// warning). Every illustrative value is a reserved documentation value or a
// persona-derived fixture host, per examples-use-reserved-identifiers: this file is
// scaffolded into every managed repo, so an example that looked like a real host or
// address would teach the shape of the leak the layer exists to prevent.
const privateStubBody = `# abcd-banlist: keyed
# abcd private banlist — LOCAL TO THIS MACHINE, never committed.
#
# The committed pre-commit guard (.githooks/pre-commit) refuses any commit whose
# staged content matches an entry below, naming the entry's KEY alone. CI cannot
# enforce this layer — it protects only machines that have opted in. That is the
# design: a pattern written into public CI config is a published pattern. Names a
# repo may state out loud belong in the public family instead
# (.abcd/docs-lint.json), where the lint gates them for everyone.
#
# The first line above declares the KEYED format: every entry line is
#     KEY<space-or-tab>PATTERN
#   KEY      a stable, non-sensitive handle ([A-Za-z0-9][A-Za-z0-9._/-]*). It is
#            the only part of an entry any output ever names.
#   PATTERN  a POSIX extended regular expression, matched case-insensitively by
#            grep. Inline flag groups and Perl escapes are NOT supported there:
#            no (?i) — matching is already case-insensitive — and no \d, \w, \b.
# Blank lines and lines starting with '#' are ignored; leading and trailing ASCII
# spaces and tabs are stripped. A line that does not parse is refused by number,
# never skipped. Machine identifiers — hostnames, IP addresses, CIDR prefixes, MAC
# addresses, device names — are ordinary entries.
#
# Add entries with ` + "`abcd banlist add --private <key> <pattern>`" + `, which keeps the
# pattern out of argv and shell history when you pipe it:
#     printf %s 'PATTERN' | abcd banlist add --private lab-host -
#
# Worked examples, commented out. Uncomment nothing: replace them with your own
# values. Every value here is a reserved documentation value (RFC 5737, RFC 3849,
# RFC 2606, RFC 7042) or a persona-derived fixture host, so nothing below names a
# real machine:
#
#   lab-host      alice-laptop\.example\.com
#   lab-device    bob-desktop
#   lab-ipv4      192\.0\.2\.17
#   lab-subnet    198\.51\.100\.
#   lab-ipv6      2001:db8:
#   lab-mac       00:00:5e:00:53:
#   lab-domain    carol-server\.example\.org
#   lab-project   a-private-project-name
#
# Remove the first line and every line below becomes a whole-line pattern again.
`

// privateStubContent is the stub's bytes as scaffolding writes them. It is a
// function rather than a bare constant so a test can assert against exactly what
// lands on disk.
func privateStubContent() string { return privateStubBody }

// BanlistHealth is the two-layer name guard's state in one repo: which of the
// scaffolded artefacts are there, and — always — how far the private layer reaches
// (spc-20 AC7).
//
// Reach is carried in the struct rather than added by a renderer because this
// report is read on two surfaces at once. A human sees a line; a machine consumer
// reads the booleans. HookInstalled=true beside a populated private store reads as
// "covered" to both of them, and it is not: the guard runs on machines that opted
// in and nowhere else. The caveat travels with the state it qualifies.
type BanlistHealth struct {
	// HookInstalled reports the committed guard hook's presence. It is a fact about
	// the repo, not about this clone's hooks path: a clone that has not pointed git
	// at the hooks directory still carries the hook for everyone else.
	HookInstalled bool `json:"hook_installed"`
	// PublicFamily reports a docs-lint config whose banned-names family is readable,
	// so a public entry can be added and CI can enforce one.
	PublicFamily bool `json:"public_family"`
	// PrivateStore reports whether THIS MACHINE has the private store at all. False
	// is the honest "inactive here" state, never silence: an absent store checks
	// nothing, and the guard says so at commit time for the same reason.
	PrivateStore bool `json:"private_store"`
	// Reach is PrivateReachNote, unconditionally.
	Reach string `json:"reach"`
}

// detectBanlistHealth answers the three artefact questions for one repo. It reads
// no entry and spawns no subprocess: the store's CONTENT is the secret, and a
// status pass that opened it would be one formatting mistake from printing it.
func detectBanlistHealth(cwd string) BanlistHealth {
	return BanlistHealth{
		HookInstalled: fileExists(filepath.Join(cwd, filepath.FromSlash(GuardHookRelPath))),
		PublicFamily:  classifyPublicFamily(cwd) == publicFamilyPresent,
		PrivateStore:  fileExists(filepath.Join(cwd, filepath.FromSlash(banlist.PrivateRelPath))),
		Reach:         banlist.PrivateReachNote,
	}
}

// publicFamilyState classifies the public layer's config for detection.
type publicFamilyState int

const (
	publicFamilyPresent  publicFamilyState = iota
	publicFamilyAbsent                     // no docs-lint config at all — scaffolding writes one
	publicFamilyUnusable                   // a config abcd must not rewrite: no array, or unparseable
)

// classifyPublicFamily reports whether the repo's docs-lint config carries the
// banned-names family. It asks the banlist package rather than re-reading the file,
// so detection and the verbs cannot disagree about what "the family is present"
// means (one canonical primitive).
func classifyPublicFamily(cwd string) publicFamilyState {
	if !fileExists(filepath.Join(cwd, filepath.FromSlash(banlist.PublicConfigRelPath))) {
		return publicFamilyAbsent
	}
	if _, err := banlist.ListPublic(cwd); err != nil {
		return publicFamilyUnusable
	}
	return publicFamilyPresent
}

// detectBanlistScaffold reports the two-layer name guard's missing artefacts
// (spc-20 AC5). Each is state-keyed: the gap is the file's absence, never a version
// stamp, so a hand-deleted hook comes back and a present one is left alone.
func detectBanlistScaffold(cwd string) []Gap {
	var gaps []Gap
	if !fileExists(filepath.Join(cwd, filepath.FromSlash(GuardHookRelPath))) {
		gaps = append(gaps, Gap{
			ID: "banlist.hook_missing", Category: SafeAutocreate, Scope: "repo",
			Title:  "private name guard not committed",
			Detail: GuardHookRelPath + " is absent, so no commit on any clone is checked against a private banlist.",
			FixHint: "ahoy install writes the guard hook; point git at it with " +
				"`git config core.hooksPath " + filepath.Dir(GuardHookRelPath) + "`.",
			Required: true, Resolvable: true,
		})
	}
	switch classifyPublicFamily(cwd) {
	case publicFamilyAbsent:
		gaps = append(gaps, Gap{
			ID: "banlist.public_family_missing", Category: SafeAutocreate, Scope: "repo",
			Title:      "public banned-names family absent",
			Detail:     banlist.PublicConfigRelPath + " is absent, so no banned name is gated in CI.",
			FixHint:    "ahoy install writes the docs-lint config with an empty banned-names family.",
			Required:   true,
			Resolvable: true,
		})
	case publicFamilyUnusable:
		// Diagnostic only: the config gates CI and a contributor owns it, so abcd
		// reports the fault and never rewrites a file it cannot read (the same posture
		// stepConfigValues takes towards a malformed config.json).
		gaps = append(gaps, Gap{
			ID: "banlist.public_family_unusable", Category: ConfigChange, Scope: "repo",
			Title:      "public banned-names family is not readable",
			Detail:     banlist.PublicConfigRelPath + " carries no usable top-level banned_tokens array, so `abcd banlist add --public` cannot write to it.",
			FixHint:    "add a top-level \"banned_tokens\": [] array to " + banlist.PublicConfigRelPath + " (`abcd banlist list --public` names the fault).",
			Required:   false,
			Resolvable: false,
		})
	}
	if !fileExists(filepath.Join(cwd, filepath.FromSlash(banlist.PrivateRelPath))) {
		// Resolvable ONLY once the abcd-managed .gitignore block covers the local tier.
		// Writing the stub into a repo that would track it is the exact hazard the layer
		// exists to prevent, and advertising a resolvable gap apply refuses to close
		// would leave the repo permanently "partial" (the markerSymlink precedent).
		gaps = append(gaps, Gap{
			ID: "banlist.private_stub_missing", Category: SafeAutocreate, Scope: "repo",
			Title:      "private banlist stub absent",
			Detail:     banlist.PrivateRelPath + " is absent, so the private layer is inactive on this machine.",
			FixHint:    "ahoy install writes the documented stub into the gitignored local tier.",
			Required:   true,
			Resolvable: localTierIgnored(cwd),
		})
	}
	return gaps
}

// localTierIgnored reports whether the abcd-managed .gitignore block for this repo's
// persisted visibility is in place — the fence entry that keeps the private store
// out of `git add -A`. An unset or invalid visibility answers true: the first
// install writes the block before the stub, in step order, so a fresh repo must not
// be told its stub is unresolvable.
func localTierIgnored(cwd string) bool {
	cfg, err := readConfig(cwd)
	if err != nil {
		return true
	}
	visibility, ok := stringVal(subMap(cfg, "repo"), "visibility")
	if !ok || !inSet(visibility, visibilityChoices) {
		return true
	}
	return !gitignoreBlockDrifts(cwd, visibility)
}

// stepBanlist scaffolds the two-layer name guard: the committed hook, the public
// family, and the gitignored private stub (spc-20 AC5). Every write is
// create-if-absent — a hook, a docs-lint config, and above all a populated private
// store are the maintainer's, and re-seeding one would delete work abcd cannot see.
func (a *applyCtx) stepBanlist(cfg *InstallConfig) {
	if !a.approved[SafeAutocreate] {
		return
	}
	if a.has("banlist.hook_missing") {
		path := filepath.Join(a.cwd, filepath.FromSlash(GuardHookRelPath))
		if wrote := createIfAbsent(path, guardHookTemplate, 0o755, 0o755); wrote {
			a.note(path)
		}
	}
	if a.has("banlist.public_family_missing") {
		path := filepath.Join(a.cwd, filepath.FromSlash(banlist.PublicConfigRelPath))
		if wrote := createIfAbsent(path, []byte(publicFamilySeed), 0o644, 0o755); wrote {
			a.note(path)
		}
	}
	if a.has("banlist.private_stub_missing") && banlistFenceReady(a.cwd, cfg) {
		path := filepath.Join(a.cwd, filepath.FromSlash(banlist.PrivateRelPath))
		// 0700/0600: the directory holding private patterns is no more readable than
		// the patterns are, matching the store the banlist verbs write.
		if wrote := createIfAbsent(path, []byte(privateStubContent()), 0o600, 0o700); wrote {
			a.note(path)
		}
	}
}

// banlistFenceReady reports whether the .gitignore fence covering the local tier is
// in place for the visibility this install settled on. It reads the freshly-applied
// visibility from cfg rather than re-reading config.json, so a first install — which
// writes the fence one step earlier — sees the value it just persisted. A run with
// no settled visibility writes no stub: fail closed, because a stub git would track
// is the hazard, not the remedy.
func banlistFenceReady(cwd string, cfg *InstallConfig) bool {
	if cfg == nil || cfg.Visibility == "" {
		return false
	}
	return !gitignoreBlockDrifts(cwd, cfg.Visibility)
}

// createIfAbsent writes data at path when nothing is there, creating the parent
// directory with dirPerm. It reports whether it wrote. An existing path of ANY kind
// — regular file, directory, symlink — is left untouched and reported as not
// written: this is a scaffold, and refusing to touch what is already there is the
// whole of its contract.
func createIfAbsent(path string, data []byte, perm, dirPerm os.FileMode) bool {
	if _, err := os.Lstat(path); err == nil {
		return false
	}
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return false
	}
	// WriteFileAtomic renames over the target, so a path that appeared between the
	// Lstat and here would be replaced. Re-check on the same beat: the loser of that
	// race must not clobber the winner's file.
	if _, err := os.Lstat(path); err == nil {
		return false
	}
	return fsutil.WriteFileAtomic(path, data, perm) == nil
}
