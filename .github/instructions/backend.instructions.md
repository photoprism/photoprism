---
applyTo:
- "internal/**"
- "pkg/**"
- "cmd/**"
---

For backend changes:

- Maintain Go style; keep functions small, errors wrapped with context.
- Add/update unit tests; ensure `make test-go` passes.
- Touching schema? Mention migrations and verify with `photoprism migrations ls`/`migrate`.
- For CLI/config, confirm flags/envs via `photoprism --help` / `photoprism show config-options` before suggesting.
