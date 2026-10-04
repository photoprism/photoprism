/*
Package main generates label classification rules from YAML.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package main

import (
	"log"
	"os"
	"path/filepath"
	"text/template"
	"unicode"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// LabelRule defines a label mapping or a direct rule alias.
type LabelRule struct {
	Label      string
	See        string
	Threshold  float32
	Categories []string
	Priority   int
}

// LabelRules maps raw class names to their label rules.
type LabelRules map[string]LabelRule

// main generates classification rules from a strict YAML source.
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

	err = yaml.UnmarshalStrict(yamlConfig, rules)

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
			target := rule.See
			targetRule, ok := rules[target]

			if !ok {
				log.Panicf("missing label: %s", target)
			} else if targetRule.See != "" {
				log.Panicf("classify: alias %s must reference a direct rule", label)
			}
		}
	}

	for label, rule := range rules {
		if rule.See != "" {
			rules[label] = rules[rule.See]
		}
	}

	f, err := os.Create("rules.go")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	err = packageTemplate.Execute(f, struct {
		Rules LabelRules
	}{
		Rules: rules,
	})
	if err != nil {
		panic(err)
	}
}

var packageTemplate = template.Must(template.New("").Parse(`
package classify

// Generated code, do not edit.

// Rules contains the generated label classification rules from rules.yml.
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
