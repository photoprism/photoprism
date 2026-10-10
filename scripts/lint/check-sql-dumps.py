#!/usr/bin/env python3
"""Check that database dumps and files are only committed where the repository expects them."""

import re
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]

# ALLOWED_PREFIXES lists the directories that hold the reviewed SQL scripts, migrations and schemas.
ALLOWED_PREFIXES = ("scripts/sql/", "internal/entity/")

# DUMP_EXTENSIONS lists the file name endings of SQL scripts, dumps and database files.
DUMP_EXTENSIONS = (
    ".sql", ".sql.gz", ".sql.bz2", ".sql.xz", ".sql.zst", ".sql.zip",
    ".dump", ".db", ".sqlite", ".sqlite3",
)

# DUMP_HEADERS matches the headers that database clients write at the start of a dump.
DUMP_HEADERS = re.compile(
    rb"^-- (?:MariaDB|MySQL) dump \d"
    rb"|^-- PostgreSQL database dump\r?$"
    rb"|^PRAGMA foreign_keys=OFF;\r?\nBEGIN TRANSACTION;",
    re.MULTILINE,
)

# HEADER_LINES is how many lines at the start of a file are checked for a dump header.
HEADER_LINES = 5

# EDITIONS lists the edition repositories that are checked when they are present.
EDITIONS = ("plus", "pro", "portal")

# BINARY_HEADERS lists the magic bytes of SQLite databases and PostgreSQL custom-format dumps.
BINARY_HEADERS = (b"SQLite format 3\x00", b"PGDMP")

HEAD_SIZE = 4096


def allowed(path):
    """Returns True if the path is in a directory that may hold SQL files."""
    return path.startswith(ALLOWED_PREFIXES) or path.startswith("testdata/") or "/testdata/" in path


def reason(path, head):
    """Returns why a file looks like a database dump or file, or an empty string if it does not."""
    if path.lower().endswith(DUMP_EXTENSIONS):
        return "SQL or database file name"
    if head.startswith(BINARY_HEADERS):
        return "database file header"
    if DUMP_HEADERS.search(b"\n".join(head.split(b"\n")[:HEADER_LINES])):
        return "database dump header"
    return ""


def candidates(repo=ROOT):
    """Returns the tracked and untracked, not ignored, files of a working tree."""
    out = subprocess.run(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"],
        cwd=repo, check=True, capture_output=True,
    ).stdout
    return sorted({p for p in out.decode("utf-8", "surrogateescape").split("\0") if p})


def repositories():
    """Returns the main repository and the edition repositories that are present, with their path prefix."""
    repos = [(ROOT, "")]
    for edition in EDITIONS:
        if (ROOT / edition / ".git").exists():
            repos.append((ROOT / edition, edition + "/"))
    return repos


def read_head(full):
    """Returns the first bytes of a regular file, or empty bytes for links, directories and missing files."""
    if full.is_symlink() or not full.is_file():
        return b""
    with open(full, "rb") as f:
        return f.read(HEAD_SIZE)


def self_check():
    """Verifies the classification with synthetic inputs, so that a broken pattern cannot pass silently."""
    mariadb = b"/*M!999999\\- enable the sandbox mode */\n-- MariaDB dump 10.19-11.8.6-MariaDB\n"
    cases = [
        ("backup", mariadb, True),
        ("notes.txt", b"PRAGMA foreign_keys=OFF;\nBEGIN TRANSACTION;\nCREATE TABLE t(x);\n", True),
        ("x", b"-- PostgreSQL database dump\n", True),
        ("index", b"SQLite format 3\x00" + b"\x00" * 16, True),
        ("backup.SQL.gz", b"\x1f\x8b", True),
        ("README.md", b"Run mysqldump to create a backup.\n-- MariaDB dumps are not committed.\n", False),
        ("main.go", b'const header = "-- MariaDB dump"\n', False),
        ("notes.md", b"# Notes\n\n\n\n\nExample:\n-- MariaDB dump 10.19\n", False),
        ("link.sql", b"", True),
    ]
    for path, head, expected in cases:
        if bool(reason(path, head)) != expected:
            sys.exit(f"check-sql-dumps: self-check failed for {path!r}")
    for path, expected in [
        ("scripts/sql/reset-local.sql", True),
        ("internal/entity/migrate/testdata/migrate_mysql.sql", True),
        ("internal/config/testdata/importtest.sql", True),
        ("testdata/dump.sql", True),
        ("backup", False),
        ("scripts/sqlite/x.sql", False),
        ("internal/entityx/x.sql", False),
    ]:
        if allowed(path) != expected:
            sys.exit(f"check-sql-dumps: self-check failed for allowed({path!r})")


def main():
    self_check()

    problems, checked = [], 0
    for repo, prefix in repositories():
        for path in candidates(repo):
            if allowed(path):
                continue
            checked += 1
            found = reason(path, read_head(repo / path))
            if found:
                problems.append(f"{prefix}{path}: {found}")

    if problems:
        print("SQL scripts, dumps and database files belong under scripts/sql/, internal/entity/ or a testdata/ directory:", file=sys.stderr)
        for problem in problems:
            print(f"  {problem}", file=sys.stderr)
        sys.exit(1)

    print(f"OK: No database dumps or files outside the allowed directories ({checked} files checked).")


if __name__ == "__main__":
    main()
