#!/usr/bin/env bash
#
# tag-release.sh — create an annotated PhotoPrism release tag with cross-repo
# provenance (main/pro/plus/portal commit SHAs + the pushed image digest in the
# tag message). One mechanism for every edition so release tags never drift.
#
# Tag convention:
#
#   edition      repo      tag
#   ----------   -------   --------------------------------
#   pro          pro       pro/<semver>-<abbrev>
#   enterprise   pro       enterprise/<semver>-<abbrev>
#   portal       portal    portal/<semver>-<abbrev>
#   plus | ce    (main)    <YYMMDD>-<abbrev>          (existing build-tag format)
#
#   <semver> = 1.YYMM.DD (zero-padded day), override with --version
#   <abbrev> = `git describe --always` in the MAIN repo — the build's
#              BUILD_VERSION / BUILD_GIT. core.abbrev is "auto", so the length is
#              dynamic; computing it the same way the build does keeps the tag
#              suffix equal to the image version string ("1.YYMM.DD-<abbrev>").
#
# The Docker tag and chart appVersion stay the clean <semver>; the -<abbrev>
# suffix is git-tag-only and prevents same-day collisions.
#
# Usage (run from the main repo root, with pro/ plus/ portal/ sub-repos present):
#   scripts/tag-release.sh <edition> [options]
#     --version <1.YYMM.DD>   override the semver (default: today, UTC)
#     --image <ref>           image to read the digest from (default: per edition)
#     --digest <sha256:...>   use this digest instead of inspecting the image
#     --ref <commit>          tag this commit (default: the edition repo's HEAD)
#     --remote <name>         push remote (default: origin)
#     --push                  push the tag after creating it (default: local only)
#     --dry-run               print the tag name + message, create nothing
#
set -euo pipefail

die() { echo "tag-release: $*" >&2; exit 1; }

EDITION="${1:-}"; [ -n "$EDITION" ] || die "usage: tag-release.sh <pro|enterprise|portal|plus|ce> [options]"
shift || true

VERSION=""; IMAGE=""; DIGEST=""; REF=""; REMOTE="origin"; PUSH=0; DRY=0; SPECS_SHA=""
while [ $# -gt 0 ]; do
  case "$1" in
    --version)    VERSION="${2:?}";   shift 2 ;;
    --image)      IMAGE="${2:?}";     shift 2 ;;
    --digest)     DIGEST="${2:?}";    shift 2 ;;
    --ref)        REF="${2:?}";       shift 2 ;;
    --remote)     REMOTE="${2:?}";    shift 2 ;;
    --specs-sha)  SPECS_SHA="${2:?}"; shift 2 ;;   # for hosts where specs/ is absent (e.g. ci)
    --push)    PUSH=1; shift ;;
    --dry-run) DRY=1;  shift ;;
    *) die "unknown option: $1" ;;
  esac
done

# Resolve the main repo root from this script's location (scripts/ is at the root).
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[ -d "$ROOT/.git" ] || [ -f "$ROOT/.git" ] || die "not a git repo: $ROOT"
git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "main repo not found at $ROOT"

# <abbrev>: exactly what the build embeds (BUILD_VERSION/BUILD_GIT), computed in main.
ABBREV="$(git -C "$ROOT" describe --always)"
[ -n "$ABBREV" ] || die "git describe --always returned empty in $ROOT"

# <semver>: 1.YYMM.DD (zero-padded day, UTC) unless overridden.
SEMVER="${VERSION:-1.$(date -u +%y%m).$(date -u +%d)}"
DATE_TAG="$(date -u +%y%m%d)"

# Capture cross-repo SHAs (sub-repos may be absent on some hosts).
sha() { local d="$1"; [ -d "$ROOT/$d/.git" ] || [ -e "$ROOT/$d/.git" ] && git -C "$ROOT/$d" rev-parse HEAD 2>/dev/null || echo "(absent)"; }
MAIN_SHA="$(git -C "$ROOT" rev-parse HEAD)"
PRO_SHA="$(sha pro)"; PLUS_SHA="$(sha plus)"; PORTAL_SHA="$(sha portal)"
# specs/ is docs-only (no code dependency) and absent on build hosts like ci;
# record it when present, or accept it via --specs-sha.
[ -n "$SPECS_SHA" ] || SPECS_SHA="$(sha specs)"

# Map edition -> repo dir, tag name, default image, human label, image-version string.
case "$EDITION" in
  pro)        REPO="pro";    TAG="pro/${SEMVER}-${ABBREV}";        DEF_IMAGE="photoprism/pro:${SEMVER}";        LABEL="Pro ${SEMVER}";        IMGVER="${SEMVER}-${ABBREV}" ;;
  enterprise) REPO="pro";    TAG="enterprise/${SEMVER}-${ABBREV}"; DEF_IMAGE="photoprism/enterprise:${SEMVER}"; LABEL="Enterprise ${SEMVER} (SDD)"; IMGVER="${SEMVER}-${ABBREV}" ;;
  portal)     REPO="portal"; TAG="portal/${SEMVER}-${ABBREV}";     DEF_IMAGE="photoprism/portal:${SEMVER}";     LABEL="Portal ${SEMVER}";     IMGVER="${SEMVER}-${ABBREV}" ;;
  plus|ce)    REPO=".";      TAG="${DATE_TAG}-${ABBREV}";          DEF_IMAGE="photoprism/photoprism:${DATE_TAG}"; LABEL="${EDITION^^} ${DATE_TAG}"; IMGVER="${DATE_TAG}-${ABBREV}" ;;
  *) die "unknown edition: $EDITION (expected pro|enterprise|portal|plus|ce)" ;;
esac
[ -d "$ROOT/$REPO/.git" ] || [ -e "$ROOT/$REPO/.git" ] || [ "$REPO" = "." ] || die "target repo '$REPO' not present under $ROOT"

# Resolve the image digest (best-effort) unless one was provided.
IMAGE="${IMAGE:-$DEF_IMAGE}"
if [ -z "$DIGEST" ]; then
  if command -v docker >/dev/null 2>&1; then
    DIGEST="$(docker buildx imagetools inspect "$IMAGE" --format '{{.Manifest.Digest}}' 2>/dev/null || true)"
  fi
fi
[ -n "$DIGEST" ] || DIGEST="(not resolved — pass --digest)"

TARGET_REF="${REF:-$(git -C "$ROOT/$REPO" rev-parse HEAD)}"

read -r -d '' MESSAGE <<EOF || true
PhotoPrism ${LABEL}

image:  ${IMAGE}@${DIGEST}
main=${MAIN_SHA}  (image version: ${IMGVER})
pro=${PRO_SHA}
plus=${PLUS_SHA}
portal=${PORTAL_SHA}
specs=${SPECS_SHA}
EOF

echo "edition: ${EDITION}"
echo "repo:    ${REPO}    ref: ${TARGET_REF}"
echo "tag:     ${TAG}"
echo "---- message ----"
echo "${MESSAGE}"
echo "-----------------"

if [ "$DRY" -eq 1 ]; then
  echo "(dry-run: nothing created)"
  exit 0
fi

git -C "$ROOT/$REPO" tag -a "$TAG" "$TARGET_REF" -m "${MESSAGE}"
echo "created annotated tag ${TAG} in ${REPO}"

if [ "$PUSH" -eq 1 ]; then
  git -C "$ROOT/$REPO" push "$REMOTE" "$TAG"
  echo "pushed ${TAG} to ${REMOTE}"
else
  echo "(local only — re-run with --push, or: git -C ${REPO} push ${REMOTE} ${TAG})"
fi
