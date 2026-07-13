package audit

// DefaultRules returns the bundled, in-binary v1 rule set in a stable order.
// Rules are data the evaluator ranges over (the rule-loader seam), so a later
// phase can add repo-level overrides without touching Evaluate.
//
// The five v1 rules land in itd-85 M3; this loader is the seam they plug into.
func DefaultRules() []Rule {
	return []Rule{}
}
