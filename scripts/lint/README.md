# Repository Checks

**Last Updated:** September 21, 2026

Standalone repository checks live here. Run them through the root Makefile so callers do not depend on script paths:

- `make check-api-request-limits` checks request-body limit coverage.
- `make check-libheif-install` exercises installer selection in an isolated user namespace.
- `make check-make-help` checks that advertised Makefile targets exist.
- `make check-scripts-copy-mode` checks container script ownership and modes.

All are included in `make lint`. The Go-based `check-audit-events` and log-rendering tools remain under `scripts/tools/`.

Invoke commands from the repository root. The installer checks use isolated destinations and stubbed downloads; they do not install libraries into the host system.
