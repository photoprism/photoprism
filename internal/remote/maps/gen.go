//go:build ignore
// +build ignore

// This generates countries.go by running "go generate"
package main

import (
	"encoding/json"
	"os"
	"strings"
	"text/template"
)

type Country struct {
	Code string
	Name string
}

var countries []Country

func main() {
	rawData, err := os.ReadFile("./countries.json")

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(rawData, &countries)

	if err != nil {
		panic(err)
	}

	f, err := os.Create("countries.go")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	for i, v := range countries {
		countries[i].Code = strings.ToLower(v.Code)
	}

	packageTemplate.Execute(f, struct {
		Countries []Country
	}{
		Countries: countries,
	})
}

var packageTemplate = template.Must(template.New("").Parse(`
package maps

// Generated code, do not edit.

var CountryNames = map[string]string{
{{- range .Countries }}
	{{ printf "%q" .Code }}: {{ printf "%q" .Name }},
{{- end }}
}`))
