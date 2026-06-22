## Commit Messages

Use concise, imperative subjects with a one-word prefix indicating the scope or topic:

- `Config: Add tests for "darktable-cli" path detection`

If the commit relates to specific issues or pull requests, reference their IDs in the message:

- `Docker: Use two stage build to reduce image size #123 #5632`

Commit messages must not exceed 80 characters in length.

Do not add `Co-Authored-By: Claude …` trailers (or any other AI-authorship trailer) to commit messages.

## GitHub Issues

Issue titles MUST be concise, use the imperative mood, and start with a single capitalized prefix followed by a colon and a space, e.g. `Search: Add filter for RAW image formats`.

Issue descriptions MUST begin with a one-sentence **User Story** in the format: `**As a <role>, I want <goal>, so that <outcome>.**`
Use level-3 Markdown headings for sections within issue descriptions, for example `### Acceptance Criteria`.
Follow the User Story with a clear summary of the expected behavior, rationale, technical considerations, and constraints.

Descriptions MUST conclude with a checklist of **Acceptance Criteria**:
- Use GitHub checklist formatting: `- [ ]`
- Criteria MUST be clear, testable, and unambiguous.
- Each item MUST use one of the following requirement-level keywords:
  - `MUST`   — required for the issue to be considered complete
  - `SHOULD` — strongly recommended but not strictly required
  - `MAY`    — optional enhancement
- Keep the checklist current: once the work for a criterion is implemented **and verified**, mark it done (`- [x]`).
- Leave items that are unverified, not yet implemented, or skipped optional (`MAY`) enhancements unchecked.
- An issue is complete only when every `MUST` is checked; never tick a box on the strength of a plan alone or an unrun test.
- When referencing an issue from a commit that fulfills some of its criteria, update the matching boxes first.

> Agents MUST create, edit, close, reopen, relabel, or otherwise modify GitHub issues only when explicitly requested by the user.

The repo's issue templates use the new GitHub `type:` property (`Bug`, `Feature`) instead of `bug`/`idea` labels. `gh issue create` does not yet accept a `--type` flag, so when filing issues programmatically use `--label` only and tell the user to set the issue type via the web UI.

## Specifications & Documentation

- Document headings use a **Chicago-style title case**, with additional code- and path-aware normalization rules (see below). Always spell the product name as `PhotoPrism`.
- When writing CLI examples or scripts, place option flags before positional arguments unless the command requires a different order.
- Use RFC 3339 UTC timestamps in request and response examples, and valid ID, UID and UUID examples in docs and tests.
- Technical specifications in the nested `specs/` subrepository may not be present in every clone or environment. Do not add `Makefile` targets in the main project that depend on `specs/` paths.
  - Auto-generated configuration and command references live under `specs/generated/`. Agents MUST NOT read, analyze, or modify anything in this directory.
  - Nested Git repositories may appear to be ignored; if so, change directories before staging or committing updates.
- **Never reference `specs/` paths from public artifacts** — issue bodies, PR descriptions, package READMEs (`frontend/README.md`, `internal/*/README.md`, etc.), top-level `CODEMAP.md`/`GLOSSARY.md`, code comments outside `specs/`. External readers see a 404 and the private subrepo's existence is leaked. Hints in `AGENTS.md` and `CLAUDE.md` files are the documented exception. Quick check: `grep -n "specs/" <file>` should return no matches before saving any public-facing file.

> **Title Case** rules (Chicago-style headline capitalization, with code- and path-aware normalization):
> - Capitalize the first word, the first word after a colon, dash, or end punctuation, and all major words, including the second part of a hyphenated major word.
> - Lowercase only articles, short conjunctions, and short prepositions of three letters or fewer when they are not in one of those positions.
> - Preserve known acronyms (for example, API, CLI, HTTP, JSON) and slash-separated acronym groups (for example, CSV/TSV) as uppercase.
> - Preserve RFC 2119 / RFC 8174 normative keywords (MUST, SHOULD, MAY, SHALL, REQUIRED, RECOMMENDED, OPTIONAL) as uppercase when used in their normative sense.
> - Preserve inline code spans (`` `foo` ``), file paths (e.g. `docs/foo-bar.md`), and slash commands (e.g. `/grill-me`) verbatim; do not recase their contents.
> - Use `&` instead of `And`/`Or` in headings.

> Refresh the `**Last Updated:**` date at the top of documents whenever you make changes to their contents, using the format `January 20, 2026` (without time); leave it as-is for simple formatting or whitespace-only edits.
