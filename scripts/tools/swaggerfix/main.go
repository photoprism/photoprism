package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: swaggerfix <swagger.json>")
		os.Exit(2)
	}
	path := os.Args[1]
	b, err := os.ReadFile(path) //nolint:gosec // G304: path comes from the command line
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(1)
	}
	// Traverse to definitions.time.Duration
	defs, _ := doc["definitions"].(map[string]any)
	if defs == nil {
		fmt.Fprintln(os.Stderr, "no definitions in swagger file")
		os.Exit(1)
	}
	td, _ := defs["time.Duration"].(map[string]any)
	if td == nil {
		fmt.Fprintln(os.Stderr, "no time.Duration schema found; nothing to do")
		os.Exit(0)
	}
	// Remove unstable enums and varnames to ensure deterministic output.
	delete(td, "enum")
	delete(td, "x-enum-varnames")
	defs["time.Duration"] = td
	doc["definitions"] = defs
	out, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, out, 0644); err != nil { //nolint:gosec // G306: generated file checked into the repository
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
}
