//go:build ignore
// +build ignore

// This generates countries.go by running "go generate"
package main

import (
	"bufio"
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
	file, err := os.Open("./resources/countries.txt")
	defer file.Close()

	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")

		if len(parts) < 2 {
			continue
		}

		countries = append(countries, Country{Code: strings.ToLower(parts[0]), Name: strings.ToLower(parts[1])})
	}

	f, err := os.Create("countries.go")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	packageTemplate.Execute(f, struct {
		Countries []Country
	}{
		Countries: countries,
	})
}

var packageTemplate = template.Must(template.New("").Parse(`
package txt

// Generated code, do not edit.

var Countries = map[string]string{
{{- range .Countries }}
	{{ printf "%q" .Name }}: {{ printf "%q" .Code }},
{{- end }}
}`))
