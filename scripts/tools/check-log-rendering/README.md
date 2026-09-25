# Log Rendering Review Tool

**Last Updated:** September 21, 2026

## Purpose & Scope

`check-log-rendering` inventories ordinary logger arguments that need a rendering decision. It is an
**advisory source review aid**, not a general data-flow analysis or a verdict about a log entry.
It is independent of `check-audit-events` and is not wired into `make lint`.

Run from the repository root:

```sh
go run ./scripts/tools/check-log-rendering -list
```

The default roots are `internal`, `pkg`, `cmd`, and the three optional edition `internal` directories.
Missing editions are skipped; missing explicit roots and parse failures are errors. Explicit symlink
roots are rejected rather than silently omitted. Selected `.go` inputs must be regular files; child
source symlinks and special files are rejected before parsing. Positional roots
can select a smaller source tree. Flags precede roots. Explicit root directories are inspected even when their
names match an exclusion. Below each root, tests, `testdata`, dot directories, `vendor`
and `zz*` scratch directories are excluded. Build tags are not evaluated.

## Recognized Calls

- Direct `event.Log` calls, resolving the event import path even with an import alias.
- Package variables initialized directly to `event.Log`, including declarations in another file in
  the same directory/package. Locally shadowed names are not assumed to be the package logger.
- Print, Info, Warn/Warning, Error, Fatal, Panic and generic Log methods, including `f`/`ln` variants.
  Explicit Debug/Trace calls are excluded. Generic Log calls are inspected at all levels.
- Each argument is considered independently, including a dynamic format expression. Literal-only
  expressions are accepted; named values (including shadowable `true`, `false` and `nil`) remain
  candidates unless rendered. No type inference or printf-verb inference is performed.

Explicit `event.System*` and `event.SystemLog` calls, and package loggers bound to `event.SystemLog`,
are outside the ordinary-log inventory. This separation is a routing distinction, not permission to
include credentials in operator output. A stack sent through the ordinary logger remains a review
candidate, not a demonstrated defect or an instruction to scrub away useful diagnostics.

## Recognized Renderers

Whole-value `clean.Log`, `LogQuote`, `LogLower`, `LogNames`, `LogUri`, `Error` and `status.Error`
calls are recognized by import path. `clean.LogUri` is the composed form for a URL, and counts for
both decisions. `clean.UriRedacted`, `clean.UriRedactedText` and `clean.FileNameRedacted` are
redactors rather than text renderers: a bare call remains a rendering candidate, while satisfying
the credential decision. `clean.Log(clean.UriRedacted(uri))` composes the two explicitly, subject to
the input-shape limitation below.
Parentheses, composition of recognized values, `strings.ToLower`/`ToUpper`/`TrimSpace`, and
`fmt.Sprintf`/`Sprint` over entirely recognized arguments retain recognition. An arbitrary wrapper or
an expression with an additional unrendered operand does not.

A separate `redaction` candidate label highlights expressions with credential/address-shaped
identifiers: password, passwd, secret, token, authorization, cookie, credential, acckey, dsn, proxy,
uri or url (case-insensitive substrings). A text renderer alone does not settle that review. Explicit
whole-value redactors stop this syntactic name inspection. These names are a heuristic: for example,
`tokenCount` can be harmless, and a credential held in `value` will not receive the special label.

`clean.ErrorFull` is not an ordinary-log renderer in this inventory. Recognizing `clean.Error` does
not prove that a producer preserved its typed error chain, or that an error has no embedded query
credential. Recognizing `clean.Log` does not establish that a path is logical rather than absolute.
Those remain semantic review obligations, as does choosing a redactor appropriate to the input shape:
passing an arbitrary secret to a URI redactor does not establish that it will be masked.

## Optional Per-Call Accounting

No baseline is shipped or silently created. To capture a private inventory for review:

```sh
go run ./scripts/tools/check-log-rendering -update -baseline .local/log-rendering.json
go run ./scripts/tools/check-log-rendering -baseline .local/log-rendering.json
```

`-update` requires an explicit baseline path. Baseline/gate adoption requires a maintainer decision;
recording an inventory does not mean its entries have been approved. Do not automatically update a
baseline to clear a new candidate.

A versioned JSON entry records file, enclosing function/receiver, full rendered call, argument index,
rule and occurrence count. Line movement does not change identity. An unrelated removal cannot pay
for an addition; repeated identical calls retain their multiplicity. Identical calls moved within one
function are indistinguishable. Changing the selected roots or their path spelling changes the review
scope, so compare like with like. Reductions are allowed; explicitly refresh the inventory after review
so stale allowances do not survive a resolved entry.

Baselines contain source expressions, including literal strings; keep inventories private until
reviewed for publication. Baseline writes publish a private (`0600`) staged file, including replacement
of an existing baseline; they do not follow a destination-file symlink. Normal listing output prints only source location, function, argument index
and review label, never the argument's source expression.

Exit codes: `0` for advisory completion or no per-identity increase, `1` for new baseline candidates,
and `2` for command, input or baseline errors. `-list` prints every candidate in either mode.

## Limits & Validation

This is an AST-level filter. It does not follow assignments, interprocedural flows, return values,
logger factories, interface fields, method values, structured logger fields or aliases of sanitizers.
Only declared function/method bodies (including their nested closures) are inspected; package-variable
initializers and handler closures registered directly in those initializers are not inspected.
It does not inspect event slice helpers, direct logrus/stdlog calls, JavaScript, stderr or access-log
writers. Dot imports are not resolved. Package bindings are collected from all parsed files, not a
particular build configuration; later rebinding is not analyzed. A dynamic scalar or trusted constant
can need no sanitization but still appear in this inventory. A clean run establishes only that this
specific syntax filter found no candidates in the selected sources.

The fixtures cover ordinary folder/import/transcode arguments, errors and credential summaries,
rendered controls, separated operator diagnostics, import aliases and shadowing, mixed expressions,
per-call replacement and multiplicity, baseline validation and CLI outcomes:

```sh
go test ./scripts/tools/check-log-rendering -count=1
```
