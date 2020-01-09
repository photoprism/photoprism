package classify

type LabelRule struct {
	Label      string
	Threshold  float32
	Categories []string
	Priority   int
}

type LabelRules map[string]LabelRule

func (rules LabelRules) Find(label string) LabelRule {
	if rule, ok := rules[label]; ok {
		return rule
	}

	return LabelRule{Threshold: 0.1}
}
