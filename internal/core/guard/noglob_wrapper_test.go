package guard

import "testing"

// TestZshPrecommandModifiersAreWrappers — iss-2608270655497992. zsh's `noglob`
// and `nocorrect` are precommand modifiers: each runs the command that follows
// it with globbing / spelling-correction turned off, exactly the way `command`,
// `nohup` or `exec` run the following command. They take NO options of their
// own, so the token right after them is command position. Missing from the
// wrappers map, they left the launched command out of command position, so a
// Tier-1 blocker behind `noglob rm -rf *` was a silent allow while the bare
// spelling blocked.
func TestZshPrecommandModifiersAreWrappers(t *testing.T) {
	stepped := []struct {
		name string
		line string
		want string
	}{
		{"noglob steps to the command", "noglob rm -rf *", "rm"},
		{"nocorrect steps to the command", "nocorrect rm -rf *", "rm"},
		{"noglob nests under sudo", "sudo noglob rm -rf *", "rm"},
		{"noglob then nocorrect nest", "noglob nocorrect rm -rf *", "rm"},
	}
	for _, tc := range stepped {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := commandOf(firstSegment(t, tc.line)); got != tc.want {
				t.Fatalf("commandOf(%q) = %q, want %q", tc.line, got, tc.want)
			}
		})
	}

	// Through the full registry: a blocker behind the modifier must reach the
	// SAME verdict as the bare hazard.
	blocked := []struct {
		name string
		line string
	}{
		{"noglob force push", "noglob git push --force origin main"},
		{"nocorrect gh repo delete", "nocorrect gh repo delete owner/repo"},
		{"noglob rm after cd", "cd scratch && noglob rm -rf *"},
	}
	for _, tc := range blocked {
		t.Run(tc.name, func(t *testing.T) {
			if got := guardVerdict(t, tc.line).Verdict; got != VerdictBlock {
				t.Fatalf("Check(%q).Verdict = %q, want %q", tc.line, got, VerdictBlock)
			}
		})
	}
}
