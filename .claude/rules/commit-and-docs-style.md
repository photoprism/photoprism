## Commit Messages

Use concise, imperative subjects with a one-word prefix indicating the scope or topic:

- `Config: Add tests for "darktable-cli" path detection`

If the commit relates to specific issues or pull requests, reference their IDs in the message:

- `Docker: Use two stage build to reduce image size #123 #5632`

Commit messages must not exceed 80 characters in length.

## GitHub Issues

Issue titles MUST be concise, use the imperative mood, and start with a single capitalized prefix followed by a colon and a space, e.g. `Search: Add filter for RAW image formats`.

Issue descriptions MUST begin with a one-sentence **User Story** in the format: `**As a <role>, I want <goal>, so that <outcome>.**`
Follow the User Story with a clear summary of the expected behavior, rationale, technical considerations, and constraints.

Descriptions MUST conclude with a checklist of **Acceptance Criteria**:
- Use GitHub checklist formatting: `- [ ]`
- Criteria MUST be clear, testable, and unambiguous.
- Each item MUST use one of the following priority keywords:
  - `MUST`   — required for the issue to be considered complete
  - `SHOULD` — strongly recommended but not strictly required
  - `MAY`    — optional enhancement

> Agents MUST create, edit, close, reopen, relabel, or otherwise modify GitHub issues only when explicitly requested by the user.

## Specifications & Documentation

- Document headings must use **Title Case** (APA/AP style) across Markdown files. Always spell the product name as `PhotoPrism`.
- When writing CLI examples or scripts, place option flags before positional arguments unless the command requires a different order.
- Use RFC 3339 UTC timestamps in request and response examples, and valid ID, UID and UUID examples in docs and tests.
- Technical specifications in the nested `specs/` subrepository may not be present in every clone or environment. Do not add `Makefile` targets in the main project that depend on `specs/` paths.
  - Auto-generated configuration and command references live under `specs/generated/`. Agents MUST NOT read, analyze, or modify anything in this directory.
  - Nested Git repositories may appear to be ignored; if so, change directories before staging or committing updates.

> **Title Case** rules (APA/AP implementation):
> - Capitalize the first word of a title/heading and the first word of a subtitle.
> - Capitalize the first word after a colon, an em dash, or end punctuation.
> - Capitalize major words, including the second part of hyphenated major words.
> - Capitalize all words of four letters or more.
> - Lowercase only minor words of three letters or fewer (articles, short conjunctions, short prepositions), except when they are in one of the positions above.
> - In headings, prefer `&` where needed; do not use `And` or `Or` in titles.

> Refresh the `**Last Updated:**` date at the top of documents whenever you make changes to their contents, using the format `January 20, 2026` (without time); leave it as-is for simple formatting or whitespace-only edits.
