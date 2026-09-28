# Repository Checks

**Last Updated:** September 26, 2026

Standalone repository checks live here. Run them through the root Makefile so callers do not depend on script paths:

- `make check-api-request-limits` checks request-body limit coverage.
- `make check-libheif-install` exercises installer selection in an isolated user namespace.
- `make check-cuda-install` exercises installation and recovery with synthetic packages and no GPU.
- `make check-make-help` checks that advertised Makefile targets exist.
- `make check-scripts-copy-mode` checks container script ownership and modes.

All are included in `make lint`. Two Go-based checks that `make lint` also runs live under `scripts/tools/`: `check-api-failure-codes` reports API handlers that can answer 413 without listing it in their `@Failure` annotation, and `check-audit-events` checks audit-event formatting against its baseline. `check-log-rendering` lives there too but is not part of `make lint`.

Invoke commands from the repository root. The installer checks use isolated destinations and stubbed downloads; they do not install libraries into the host system.
