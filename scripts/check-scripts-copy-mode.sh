#!/usr/bin/env bash

# Verifies that every Dockerfile copying the dist scripts into the image sets an explicit owner
# and mode, and that a later step in the same stage normalizes the result. The copy states the
# intent and this check guards it; "cleanup.sh" is what every image build ends with.
#
# Usage: scripts/check-scripts-copy-mode.sh [dockerfile ...]
#
# Without arguments, all Dockerfiles in this working copy are checked, including the ones in
# optional private subdirectories when they are present.

set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

# A copy instruction, ignoring comments, and allowing tabs and either instruction keyword.
COPY_PATTERN='^[[:space:]]*(COPY|ADD)[[:space:]].*scripts/dist'

# Accept any spelling of root and of mode 755, rather than one literal form of each.
OWNER_PATTERN='--chown=(root|0):(root|0)([[:space:]]|$)'
MODE_PATTERN='--chmod=0?755([[:space:]]|$)'

# A step that normalizes the mode after the copy: the shared cleanup script, which every image
# build ends with, or the equivalent command spelled out. The path is anchored so that a
# neighboring directory whose name merely starts with "/scripts" does not satisfy it.
NORMALIZE_PATTERN='(/scripts/cleanup\.sh([[:space:]]|$)|chmod[[:space:]]+-R[[:space:]]+go-w[[:space:]]+/scripts([[:space:]]|/|$))'

# A step that records what the image is, which the startup scripts read instead of taking it
# from the caller: either the installer, which validates the name itself, or the file written
# out. The values are matched here so that the images writing it directly are held to the same
# set the installer enforces - a typo would otherwise resolve to "unknown" at runtime, which
# skips the init script and the privilege drop with it.
IDENTITY_PATTERN='(/scripts/install-sudoers\.sh([[:space:]]|$)|DOCKER_ENV=(prod|develop)\\nDOCKER_IMG=(develop|ce|plus|pro|portal)\\n.*>[[:space:]]*/scripts/\.env([[:space:]]|$))'

if [[ $# -gt 0 ]]; then
  DOCKERFILES=("$@")

  for dockerfile in "${DOCKERFILES[@]}"; do
    if [[ ! -f ${dockerfile} ]]; then
      echo "${dockerfile}: no such file, relative paths resolve from the repository root" 1>&2
      exit 1
    fi
  done
else
  mapfile -t DOCKERFILES < <(
    find . -name Dockerfile -not -path "./.git/*" -not -path "*/node_modules/*" \
      -not -path "./.local/*" -not -path "./storage/*" -not -path "./.gocache/*" | sort
  )

  if [[ ${#DOCKERFILES[@]} -eq 0 ]]; then
    echo "No Dockerfile found, so nothing was checked." 1>&2
    exit 1
  fi
fi

FOUND=0
FAILED=0

for dockerfile in "${DOCKERFILES[@]}"; do
  matched=0

  while IFS=: read -r line_no line; do
    matched=$((matched + 1))
    FOUND=$((FOUND + 1))

    if [[ ! ${line} =~ ${OWNER_PATTERN} ]]; then
      echo "${dockerfile}:${line_no}: copies the dist scripts without --chown=root:root" 1>&2
      FAILED=$((FAILED + 1))
    fi

    if [[ ! ${line} =~ ${MODE_PATTERN} ]]; then
      echo "${dockerfile}:${line_no}: copies the dist scripts without --chmod=755" 1>&2
      FAILED=$((FAILED + 1))
    fi

    # The stated mode is not sufficient on its own, so require a step that normalizes it after
    # the copy. Stop at the next stage: a later one writes a different filesystem, and a step
    # before the copy would be overwritten by it.
    # Comments are stripped first: a guard that a comment naming the script can satisfy would
    # report green on a build that never runs it.
    stage_tail=$(tail -n "+$((line_no + 1))" "${dockerfile}" |
      sed -n '/^[[:space:]]*FROM[[:space:]]/q;p' | grep -vE '^[[:space:]]*#')

    if ! grep -qE "${NORMALIZE_PATTERN}" <<< "${stage_tail}"; then
      echo "${dockerfile}:${line_no}: copies the dist scripts without normalizing their mode afterwards" 1>&2
      echo "${dockerfile}:${line_no}: run /scripts/cleanup.sh, or 'chmod -R go-w /scripts', later in the same stage" 1>&2
      FAILED=$((FAILED + 1))
    fi

    if ! grep -qE "${IDENTITY_PATTERN}" <<< "${stage_tail}"; then
      echo "${dockerfile}:${line_no}: copies the dist scripts without recording the image identity" 1>&2
      echo "${dockerfile}:${line_no}: run /scripts/install-sudoers.sh, or write /scripts/.env with" 1>&2
      echo "${dockerfile}:${line_no}: DOCKER_ENV=prod|develop and DOCKER_IMG=develop|ce|plus|pro|portal" 1>&2
      FAILED=$((FAILED + 1))
    fi
  done < <(grep -nEi "${COPY_PATTERN}" "${dockerfile}" || true)

  # A file that names the scripts without a recognized instruction has been reshaped into a
  # form this check no longer reads, which would otherwise disable it silently.
  if [[ ${matched} -eq 0 ]] && grep -qE '^[[:space:]]*[^#]*scripts/dist' "${dockerfile}"; then
    echo "${dockerfile}: names the dist scripts in an instruction this check cannot read" 1>&2
    FAILED=$((FAILED + 1))
  fi
done

if [[ ${FAILED} -gt 0 ]]; then
  echo "Found ${FAILED} problem(s) with how the dist scripts are copied." 1>&2
  exit 1
fi

echo "Script copy modes checked (${FOUND} instruction(s) in ${#DOCKERFILES[@]} Dockerfile(s))."
