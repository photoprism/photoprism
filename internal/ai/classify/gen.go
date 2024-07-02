//go:build ignore
// +build ignore

// This generates stopwords.go by running "go generate"
package main

import (
	"log"
	"os"
	"path/filepath"
	"text/template"
	"unicode"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"gopkg.in/yaml.v2"
)

// LabelRule defines the rule for a given Label
type LabelRule struct {
	Label      string
	See        string
	Threshold  float32
	Categories []string
	Priority   int
}

type LabelRules map[string]LabelRule

// This function generates the rules.go file containing rule extracted from rules.yml file
func main() {
	rules := make(LabelRules)

	fileName := "rules.yml"

	if !fs.FileExists(fileName) {
		log.Panicf("classify: found no label rules in %s", clean.Log(filepath.Base(fileName)))
	}

	yamlConfig, err := os.ReadFile(fileName)

	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlConfig, rules)

	if err != nil {
		panic(err)
	}

	for label, rule := range rules {
		for _, char := range label {
			if unicode.IsUpper(char) {
				log.Panicf("classify: %s must be lowercase", label)
			}
		}

		if rule.See != "" {
			rule, ok := rules[rule.See]

			if !ok {
				log.Panicf("missing label: %s", rule.See)
			}

			rules[label] = rule
		}
	}

	f, err := os.Create("rules.go")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	packageTemplate.Execute(f, struct {
		Rules LabelRules
	}{
		Rules: rules,
	})
}

var packageTemplate = template.Must(template.New("").Parse(`
package classify

// Generated code, do not edit.

var Rules = LabelRules{
{{- range $key, $value := .Rules }}
	{{ printf "%q" $key }}:  {
		Label:      {{ printf "%q" $value.Label }},
		Threshold:  {{ printf "%f" $value.Threshold }},
		Priority:   {{ $value.Priority }},
		Categories: []string{ {{- range $value.Categories }} {{ printf "%q" . }}, {{- end }} },
	},
{{- end }}
}`))
