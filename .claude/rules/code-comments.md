## JS/Go Code Comment Rules

A doc comment is **required** for every function (including unexported helpers), as well as for every non-trivial Vue `methods:` / `computed:` / watcher:
- Keep comments **compact** and default to one line for "what" in the format `// Name does X.`. Skip trivial getters (`isOpen: () => this.open`).
- Add 1-2 follow-up lines (`// …`) **only** if the "why" is non-obvious: a hidden invariant, a workaround that would otherwise be undone by a future cleanup, a contract a reader can't infer from the code. If readers can infer the "why" from the function body or a nearby line, then omit it.
- Multi-paragraph explanations belong in `specs/`, package `README.md` files, or GitHub issues — never in the source itself.

Doc comments for packages and exported identifiers must be complete sentences that begin with the name of the thing being described and end with a period. For short examples in comments, indent code instead of using backticks.

Use US English spelling in all code comments (`parameterized`, `behavior`, `color`, `serialize`, `normalize`, `optimize`, …) — not the British `-ised`/`-our`/`-re` variants.

> **Don't include in code comments:** Issue / PR numbers, "previously…" history, alternatives considered, what the function used to do, references to old commits, names of subsequent reviewers, or any narrative that names the change rather than the steady-state behavior. That context belongs in commit messages, specs, or handover notes.
