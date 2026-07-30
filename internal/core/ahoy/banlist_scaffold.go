package ahoy

import (
	"bytes"
	_ "embed"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/banlist"
	"github.com/REPPL/abcd-cli/internal/fsutil"
	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// The committed guard's two halves, repo-relative and slash-separated (they are
// os.Root paths as well as report text).
//
// They are COMMITTED hooks, not .git/hooks copies: .git/hooks is per-clone and
// untracked, so a guard installed there protects one working copy and no one
// else's. A clone points git at this directory once
// (`git config core.hooksPath .githooks`) and every checkout inherits the check.
//
// pre-merge-commit exists because git runs NO pre-commit hook for a merge commit.
// Without it every name the private layer bans walks into history the moment a
// branch carrying it is merged, and the guard reports nothing because it was never
// asked.
const (
	GuardHookRelPath      = ".githooks/pre-commit"
	GuardMergeHookRelPath = ".githooks/pre-merge-commit"
	guardHooksDirRelPath  = ".githooks"
)

// guardHookMarker identifies a hook abcd wrote. A repo may already have its own
// pre-commit hook, and reporting a foreign one as "the abcd guard is installed" is
// the worst of both states: nothing checks the banlist and the status board says
// something does. Presence is not identity, so identity is stamped.
const guardHookMarker = "# abcd-name-guard: v1"

// guardHookTemplate and guardMergeHookTemplate are the canonical scaffolded guard:
// the generalised form of the prototype this repo runs on itself, with the
// repo-specific gates dropped. They are embedded so the binary is self-contained
// (the marker block's precedent) and so scaffolding cannot depend on a plugin-root
// file that a broken install left behind.
//
//go:embed defaults/pre-commit
var guardHookTemplate []byte

//go:embed defaults/pre-merge-commit
var guardMergeHookTemplate []byte

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
# The committed pre-commit and pre-merge-commit guards (.githooks/) refuse any
# commit whose staged content matches an entry below, naming the entry's KEY alone.
# CI cannot enforce this layer, and neither guard sees a rebase, a ` + "`git am`" + `, a
# cherry-pick, or a commit made with --no-verify. That is the design, not a gap: a
# pattern written into public CI config is a published pattern. Names a repo may
# state out loud belong in the public family instead (.abcd/docs-lint.json), where
# the lint gates them for everyone.
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

// ---------------------------------------------------------------------------
// classification
// ---------------------------------------------------------------------------

// HookState is what occupies one of the guard's two hook paths.
type HookState string

const (
	HookAbsent    HookState = "absent"
	HookInstalled HookState = "installed"
	// HookForeign is a hook abcd did not write. It is NOT a failure of the repo —
	// a maintainer's own pre-commit hook is legitimate — but it is a state in which
	// nothing checks the banlist, so it can never be reported as installed.
	HookForeign HookState = "foreign"
	// HookUnreadable is a path that exists and cannot be read as a hook (a symlink,
	// a directory, an oversize file). Like GuardHealth's unknown state it asserts
	// nothing it did not check.
	HookUnreadable HookState = "unreadable"
)

// classifyGuardHook reports what occupies rel. Identity comes from the marker line,
// never from mere presence: a repo's own pre-commit hook is legitimate, and calling
// it "the abcd guard" would mean nothing checks the banlist while the status board
// says something does.
func classifyGuardHook(cwd, rel string) HookState {
	p := filepath.Join(cwd, filepath.FromSlash(rel))
	fi, err := os.Lstat(p)
	switch {
	case err != nil && os.IsNotExist(err):
		return HookAbsent
	case err != nil:
		return HookUnreadable
	case fi.Mode()&os.ModeSymlink != 0 || !fi.Mode().IsRegular():
		return HookUnreadable
	}
	data, err := fsutil.ReadGuarded(p, maxAhoyFileBytes)
	if err != nil {
		return HookUnreadable
	}
	if bytes.Contains(data, []byte(guardHookMarker)) {
		return HookInstalled
	}
	return HookForeign
}

// PublicFamilyState classifies the public layer's config for detection.
type PublicFamilyState string

const (
	PublicFamilyPresent PublicFamilyState = "present"
	// PublicFamilyAbsent is no docs-lint config at all — scaffolding writes one.
	PublicFamilyAbsent PublicFamilyState = "absent"
	// PublicFamilyUnusable is a config that parses to no usable banned_tokens array.
	// abcd must not rewrite it: the config gates CI and a contributor owns it.
	PublicFamilyUnusable PublicFamilyState = "unusable"
	// PublicFamilyUnreadable is an I/O fault — an oversize file, a symlink, a
	// permission error. It is reported apart from "unusable" because "add a
	// banned_tokens array" is not the fix for a file nobody can open.
	PublicFamilyUnreadable PublicFamilyState = "unreadable"
	// PublicFamilyIgnored is the contradiction: a readable family in a file git
	// ignores. The layer's whole claim is that it is committed and CI-enforced, and
	// an ignored config is enforced by nobody — most sharply under
	// `visibility: public`, where the abcd fence ignores the whole `.abcd/` namespace
	// and public exposure is the risk the layer exists for.
	PublicFamilyIgnored PublicFamilyState = "ignored"
)

// classifyPublicFamily reports whether the repo's docs-lint config carries a
// banned-names family CI can actually enforce. It asks the banlist package about the
// family rather than re-reading the file, so detection and the verbs cannot disagree
// about what "the family is present" means (one canonical primitive), and it asks
// git whether the file is enforceable at all rather than inferring it from a
// configured visibility.
func classifyPublicFamily(cwd string) PublicFamilyState {
	p := filepath.Join(cwd, filepath.FromSlash(banlist.PublicConfigRelPath))
	if _, err := fsutil.ReadGuarded(p, maxAhoyFileBytes); err != nil {
		if os.IsNotExist(err) {
			return PublicFamilyAbsent
		}
		return PublicFamilyUnreadable
	}
	if _, err := banlist.ListPublic(cwd); err != nil {
		return PublicFamilyUnusable
	}
	// git's own verdict, and only inside a repo: check-ignore consults the index, so
	// a config that is already committed reads as not-ignored whatever the fence
	// says — which is the right answer, because it IS in the repo and CI sees it.
	if gitutil.InRepo(cwd) && gitutil.IsIgnored(cwd, banlist.PublicConfigRelPath) {
		return PublicFamilyIgnored
	}
	return PublicFamilyPresent
}

// privateStoreIsIgnored reports git's OWN verdict on whether the store's path would
// enter history — the only authority on the question. A text comparison of the
// abcd-managed .gitignore block answers a different one: a repo can carry a
// byte-perfect block and still track the store (a negation after it, a tracked
// tier), and a repo that never reached the visibility step has no block at all while
// its config may be otherwise complete.
//
// Outside a git repo the check is SKIPPED and the path is treated as safe, exactly
// as banlist's own requireIgnoredStore does: a scaffold cannot demand proof no one
// can supply, and a directory that is not a repository cannot commit anything.
func privateStoreIsIgnored(cwd string) bool {
	if !gitutil.InRepo(cwd) {
		return true
	}
	return gitutil.IsIgnored(cwd, banlist.PrivateRelPath)
}

// ---------------------------------------------------------------------------
// health
// ---------------------------------------------------------------------------

// BanlistHealth is the two-layer name guard's state in one repo: which of the
// scaffolded artefacts are there, what shape the private layer is in, and — always —
// how far it reaches (spc-20 AC7).
//
// Reach is carried in the struct rather than added by a renderer because this report
// is read on two surfaces at once. A human sees a line; a machine consumer reads the
// fields. "hook: installed" beside a populated private store reads as "covered" to
// both of them, and it is not: the guard runs on machines that opted in, for the
// commits git asks a hook about, and nowhere else. The caveat travels with the state
// it qualifies.
type BanlistHealth struct {
	// Hook and MergeHook are the committed guard's two halves. A merge commit runs
	// NO pre-commit hook, so a repo with only the first half is uncovered for exactly
	// the commit that merges someone else's branch.
	Hook      HookState `json:"hook"`
	MergeHook HookState `json:"merge_hook"`
	// HooksPathArmed reports whether THIS clone's local git config points at the
	// committed hooks directory. False is not a fault — the hooks path can be armed
	// by a user-level dispatcher this deliberately does not claim to see — so the
	// surfaces phrase it as an instruction, never as an accusation.
	HooksPathArmed bool `json:"hooks_path_armed"`
	// PublicFamily is the committed layer's state, including the case where it is
	// present and readable but git ignores the file, so CI never sees it.
	PublicFamily PublicFamilyState `json:"public_family"`
	// PrivateStore reports whether THIS MACHINE has opted in. False is the honest
	// "inactive here" state, never silence: an absent store checks nothing, and the
	// guard says so at commit time for the same reason.
	PrivateStore bool `json:"private_store"`
	// PrivateStoreIgnored is git's verdict on the store's path. A present store git
	// would track is one `git add -A` from committing the patterns it exists to keep
	// out of history.
	PrivateStoreIgnored bool `json:"private_store_ignored"`
	// PrivateKeyed, PrivateEntries and PrivateUnparsed are the store's SHAPE, never
	// its content: a format flag and two counts, taken from the shared parser. A line
	// that does not parse stops every commit until it is fixed.
	PrivateKeyed    bool `json:"private_keyed"`
	PrivateEntries  int  `json:"private_entries"`
	PrivateUnparsed int  `json:"private_unparsed"`
	// PrivateUnreadable reports a store that exists and cannot be read for what it
	// is — a damaged format declaration, an unreadable file. The guard refuses every
	// commit in that state, so a status board must not render it as healthy.
	PrivateUnreadable bool `json:"private_unreadable"`
	// Reach is PrivateReachNote, unconditionally.
	Reach string `json:"reach"`
}

// detectBanlistHealth answers the artefact questions for one repo. It reads the
// store's SHAPE and never its content: the patterns are the secret, and a status
// board is exactly the surface that must not hold them.
func detectBanlistHealth(cwd string) BanlistHealth {
	h := BanlistHealth{
		Hook:                classifyGuardHook(cwd, GuardHookRelPath),
		MergeHook:           classifyGuardHook(cwd, GuardMergeHookRelPath),
		HooksPathArmed:      hooksPathArmed(cwd),
		PublicFamily:        classifyPublicFamily(cwd),
		PrivateStoreIgnored: true,
		Reach:               banlist.PrivateReachNote,
	}
	sum, err := banlist.SummarisePrivate(cwd)
	if err != nil {
		// A store that exists and cannot be read for what it is: present, and health
		// says so rather than rendering the absence of a readable store as "inactive".
		h.PrivateStore, h.PrivateUnreadable = true, true
		h.PrivateStoreIgnored = privateStoreIsIgnored(cwd)
		return h
	}
	h.PrivateStore = sum.Present
	h.PrivateStoreIgnored = sum.Ignored
	h.PrivateKeyed = sum.Keyed
	h.PrivateEntries = sum.Entries
	h.PrivateUnparsed = sum.Unparsed
	if !sum.Present {
		// An absent store's safety question is about the PATH, not the file: it is what
		// decides whether scaffolding may create one here.
		h.PrivateStoreIgnored = privateStoreIsIgnored(cwd)
	}
	return h
}

// hooksPathArmed reports whether this clone's LOCAL git config points core.hooksPath
// at the committed hooks directory. `--local` reads only .git/config, so the probe's
// own `-c core.hooksPath=…` isolation cannot answer for it.
//
// A false answer is deliberately weak evidence: the hooks path can also be armed by
// a user-level dispatcher this never sees, so no surface may turn it into "the guard
// is not running" — only into "arm it like this".
func hooksPathArmed(cwd string) bool {
	out, err := gitutil.Run(cwd, "config", "--local", "--get", "core.hooksPath")
	if err != nil || strings.TrimSpace(out) == "" {
		return false
	}
	return path.Clean(filepath.ToSlash(strings.TrimSpace(out))) == guardHooksDirRelPath
}

// ---------------------------------------------------------------------------
// detection
// ---------------------------------------------------------------------------

// detectBanlistScaffold reports the name guard's missing or occupied artefacts
// (spc-20 AC5). Each is state-keyed: the gap is the artefact's state, never a
// version stamp, so a hand-deleted hook comes back and a present one is left alone.
// It consumes the health pass rather than re-reading the repo, so a gap and the
// status line beside it can never disagree.
func detectBanlistScaffold(h BanlistHealth) []Gap {
	var gaps []Gap
	gaps = append(gaps, hookGaps("banlist.hook", GuardHookRelPath, "pre-commit", h.Hook)...)
	gaps = append(gaps, hookGaps("banlist.merge_hook", GuardMergeHookRelPath, "pre-merge-commit", h.MergeHook)...)
	gaps = append(gaps, publicFamilyGaps(h.PublicFamily)...)
	gaps = append(gaps, privateStoreGaps(h)...)
	return gaps
}

// publicFamilyGaps turns the public layer's state into gaps. Only a genuinely absent
// config is abcd's to write; every other fault names its own remedy, because "add a
// banned_tokens array" is the wrong instruction for a file nobody can open and for
// one git will never show CI.
func publicFamilyGaps(state PublicFamilyState) []Gap {
	switch state {
	case PublicFamilyAbsent:
		return []Gap{{
			ID: "banlist.public_family_missing", Category: SafeAutocreate, Scope: "repo",
			Title:      "public banned-names family absent",
			Detail:     banlist.PublicConfigRelPath + " is absent, so no banned name is gated in CI.",
			FixHint:    "ahoy install writes the docs-lint config with an empty banned-names family.",
			Required:   true,
			Resolvable: true,
		}}
	case PublicFamilyUnusable:
		// Diagnostic only: the config gates CI and a contributor owns it, so abcd
		// reports the fault and never rewrites a file it cannot read (the same posture
		// stepConfigValues takes towards a malformed config.json).
		return []Gap{{
			ID: "banlist.public_family_unusable", Category: ConfigChange, Scope: "repo",
			Title:      "public banned-names family is not readable",
			Detail:     banlist.PublicConfigRelPath + " carries no usable top-level banned_tokens array, so `abcd banlist add --public` cannot write to it.",
			FixHint:    "add a top-level \"banned_tokens\": [] array to " + banlist.PublicConfigRelPath + " (`abcd banlist list --public` names the fault).",
			Required:   false,
			Resolvable: false,
		}}
	case PublicFamilyUnreadable:
		return []Gap{{
			ID: "banlist.public_family_unreadable", Category: ConfigChange, Scope: "repo",
			Title:      "public banned-names config cannot be read",
			Detail:     banlist.PublicConfigRelPath + " exists but cannot be read (not a regular file, oversize, or unreadable).",
			FixHint:    "restore " + banlist.PublicConfigRelPath + " as a regular file abcd can read.",
			Required:   false,
			Resolvable: false,
		}}
	case PublicFamilyIgnored:
		return []Gap{{
			ID: "banlist.public_family_ignored", Category: ConfigChange, Scope: "repo",
			Title:  "public banned-names family is not enforceable",
			Detail: "git ignores " + banlist.PublicConfigRelPath + ", so the family it carries never reaches CI — and the public layer's whole claim is that it is committed and enforced for everyone.",
			FixHint: "commit " + banlist.PublicConfigRelPath + " (`git add -f`), or ban the name on the private layer instead; " +
				"under `visibility: public` the abcd fence ignores the whole .abcd/ namespace, which is a placement question a maintainer must settle.",
			Required:   false,
			Resolvable: false,
		}}
	}
	return nil
}

// privateStoreGaps turns the private layer's state into gaps.
func privateStoreGaps(h BanlistHealth) []Gap {
	var gaps []Gap
	if !h.PrivateStore {
		// Resolvable ONLY when git itself reports the store's path as ignored. Writing
		// the stub into a repo that would track it is the exact hazard the layer exists
		// to prevent, and advertising a resolvable gap apply refuses to close would
		// leave the repo permanently "partial" (the markerSymlink precedent).
		fix := "ahoy install writes the documented stub into the gitignored local tier."
		if !h.PrivateStoreIgnored {
			fix = "git does not ignore " + banlist.PrivateRelPath + " — add `" + banlist.PrivateDirRelPath +
				"/` to .gitignore (ahoy install writes that fence once this repo has a configured visibility), then re-run."
		}
		gaps = append(gaps, Gap{
			ID: "banlist.private_stub_missing", Category: SafeAutocreate, Scope: "repo",
			Title:      "private banlist stub absent",
			Detail:     banlist.PrivateRelPath + " is absent, so the private layer is inactive on this machine.",
			FixHint:    fix,
			Required:   true,
			Resolvable: h.PrivateStoreIgnored,
		})
		return gaps
	}
	if h.PrivateUnreadable {
		gaps = append(gaps, Gap{
			ID: "banlist.private_store_unreadable", Category: ConfigChange, Scope: "repo",
			Title:      "private banlist cannot be read",
			Detail:     banlist.PrivateRelPath + " exists but cannot be read for what it is, so the guard refuses every commit.",
			FixHint:    "`abcd banlist list --private` names the fault; the store's content is withheld by design.",
			Required:   false,
			Resolvable: false,
		})
	}
	if !h.PrivateStoreIgnored {
		gaps = append(gaps, Gap{
			ID: "banlist.private_store_tracked", Category: ConfigChange, Scope: "repo",
			Title:      "private banlist is not gitignored",
			Detail:     "git does not ignore " + banlist.PrivateRelPath + ", so it is one `git add -A` from committing the patterns it exists to keep out of history.",
			FixHint:    "add `" + banlist.PrivateDirRelPath + "/` to .gitignore, and check the path is not already tracked.",
			Required:   false,
			Resolvable: false,
		})
	}
	return gaps
}

// hookGaps turns one hook's state into gaps. An absent hook is abcd's to write; a
// foreign one is the maintainer's and is reported, never replaced — mirroring
// GuardHealth's posture, where wiring abcd does not own is a diagnostic rather than
// something apply silently takes over.
func hookGaps(idPrefix, rel, name string, state HookState) []Gap {
	switch state {
	case HookAbsent:
		return []Gap{{
			ID: idPrefix + "_missing", Category: SafeAutocreate, Scope: "repo",
			Title:  "private name guard's " + name + " hook not committed",
			Detail: rel + " is absent, so no " + name + " on any clone is checked against a private banlist.",
			FixHint: "ahoy install writes the guard hook; point git at it with " +
				"`git config core.hooksPath " + guardHooksDirRelPath + "`.",
			Required: true, Resolvable: true,
		}}
	case HookForeign:
		return []Gap{{
			ID: idPrefix + "_foreign", Category: PluginOwned, Scope: "repo",
			Title:      "a foreign " + name + " hook occupies " + rel,
			Detail:     rel + " carries no `" + guardHookMarker + "` line, so abcd did not write it and the private banlist is not checked on this path.",
			FixHint:    "chain the abcd guard from it by hand, or move it aside and re-run `abcd ahoy install`.",
			Required:   false,
			Resolvable: false,
		}}
	case HookUnreadable:
		return []Gap{{
			ID: idPrefix + "_unreadable", Category: PluginOwned, Scope: "repo",
			Title:      rel + " cannot be read as a hook",
			Detail:     rel + " exists but is a symlink, a directory, or otherwise unreadable, so abcd can neither identify nor replace it.",
			FixHint:    "restore " + rel + " as a regular file, or remove it and re-run `abcd ahoy install`.",
			Required:   false,
			Resolvable: false,
		}}
	}
	return nil
}

// ---------------------------------------------------------------------------
// apply
// ---------------------------------------------------------------------------

// stepBanlist scaffolds the name guard: the two committed hooks, the public family,
// and the gitignored private stub (spc-20 AC5). Every write is create-if-absent — a
// hook, a CI-gating config, and above all a populated private store are the
// maintainer's, and re-seeding one would delete work abcd cannot see.
//
// Every write is CONTAINED: it resolves through an os.Root opened at the repo, so a
// symlink committed at `.githooks` or at the local tier cannot land a 0755 hook or a
// 0600 stub outside the repo while every surface reports the in-repo path.
func (a *applyCtx) stepBanlist() {
	if !a.approved[SafeAutocreate] {
		return
	}
	root, err := os.OpenRoot(a.cwd)
	if err != nil {
		return
	}
	defer root.Close()

	if a.has("banlist.hook_missing") {
		a.createContained(root, GuardHookRelPath, guardHookTemplate, 0o755, 0o755)
	}
	if a.has("banlist.merge_hook_missing") {
		a.createContained(root, GuardMergeHookRelPath, guardMergeHookTemplate, 0o755, 0o755)
	}
	if a.has("banlist.public_family_missing") {
		a.createContained(root, banlist.PublicConfigRelPath, []byte(publicFamilySeed), 0o644, 0o755)
	}
	// Re-asked HERE, after stepVisibility has written the fence, and answered by git
	// rather than by a comparison of .gitignore text: what matters is whether git
	// would track this path now, on disk.
	if a.has("banlist.private_stub_missing") && privateStoreIsIgnored(a.cwd) {
		// 0700/0600: the directory holding private patterns is no more readable than
		// the patterns are, matching the store the banlist verbs write.
		a.createContained(root, banlist.PrivateRelPath, []byte(privateStubContent()), 0o600, 0o700)
	}
}

// createContained writes data at rel inside root when nothing is there, noting the
// write. rel is slash-separated and repo-relative.
//
// It uses CreateExclusiveIn, not WriteFileAtomic: the exclusive create IS both
// guarantees at once — the file cannot already exist, and every component is resolved
// through the Root, so a symlinked ancestor is refused rather than followed. An
// atomic rename would clobber whatever was there and would depend on the caller
// having resolved the directory safely beforehand, which is precisely the
// check-then-write window a scaffold must not have.
//
// Any failure — the file already exists (the ordinary idempotent no-op), an escaping
// symlink, a permission fault — leaves the artefact unwritten and unnoted. Detection
// reports it on the next pass rather than this run claiming a file it did not create.
func (a *applyCtx) createContained(root *os.Root, rel string, data []byte, perm, dirPerm os.FileMode) {
	if dir := path.Dir(rel); dir != "." {
		// MkdirAll on the PARENT, and it may create more than one level: every artefact
		// here sits at most two deep under the repo root (.githooks/, .abcd/.work.local/),
		// both of which are abcd's own namespaces, so there is no third party's directory
		// for it to bring into being. dirPerm applies only to levels it CREATES — an
		// existing directory keeps its mode, which is why the private tier's mode is
		// asserted separately by the banlist package rather than assumed from here.
		if err := root.MkdirAll(dir, dirPerm); err != nil {
			return
		}
	}
	if err := fsutil.CreateExclusiveIn(root, rel, data, perm); err != nil {
		return
	}
	a.note(filepath.Join(a.cwd, filepath.FromSlash(rel)))
}
