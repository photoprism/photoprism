# Repository Checks

**Last Updated:** October 9, 2026

Standalone repository checks live here. Run them through the root Makefile so callers do not depend on script paths:

- `make check-api-request-limits` checks request-body limit coverage.
- `make check-libheif-install` exercises installer selection in an isolated user namespace.
- `make check-cuda-install` exercises installation and recovery with synthetic packages and no GPU.
- `make check-buildignore` checks that `assets/.buildignore` lets `make install` bundle exactly the models in `BUNDLED_MODELS`; `make dep-models` runs it first.
- `make check-make-help` checks that advertised Makefile targets exist.
- `make check-scripts-copy-mode` checks container script ownership and modes.
- `make check-sql-dumps` checks that SQL scripts, database dumps and database files are only in `scripts/sql/`, `internal/entity/` or a `testdata/` directory. It also recognizes a dump by the header in its first lines, so a file without an `.sql` extension is caught as well, and it checks the `plus`, `pro` and `portal` repositories when they are present.

All are included in `make lint`. Three Go-based checks that `make lint` also runs live under `scripts/tools/`: `check-api-failure-codes` reports API handlers that can answer 413 without listing it in their `@Failure` annotation, `check-audit-events` checks audit-event formatting against its baseline, and `check-gorm-v1` counts code that depends on GORM v1 against its baseline, so the sites to change for GORM v2 do not grow; the entries of the edition repositories are kept in a `baseline.txt` at the same path in each of them. `check-log-rendering` lives there too but is not part of `make lint`. `make check-migrations` runs `TestGeneratedDialects` in `internal/entity/migrate`, which fails when the generated migration files differ from their SQL sources.

Invoke commands from the repository root. The installer checks use isolated destinations and stubbed downloads; they do not install libraries into the host system.
