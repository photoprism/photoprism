package classify

// LabelRule defines the rule for a given Label
type LabelRule struct {
	Label      string
	Threshold  float32
	Categories []string
	Priority   int
}

// LabelRules is a map of rules with label name as index
type LabelRules map[string]LabelRule

// Find is a getter for LabelRules that give a default rule with a non-zero threshold for missing keys
func (rules LabelRules) Find(label string) LabelRule {
	if rule, ok := rules[label]; ok {
		return rule
	}

	return LabelRule{Threshold: 0.1}
}
