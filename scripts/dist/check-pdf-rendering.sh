#!/usr/bin/env bash

# check-pdf-rendering.sh determines whether a container image can rasterize a PDF through
# ImageMagick's Ghostscript delegate, and which layer blocks it when it cannot.
#
# It answers two questions that cannot be settled by reading the Dockerfiles:
#
#   1. Does Ubuntu's AppArmor "gs" profile attach inside the container? That decides whether the
#      unconfined /usr/local/bin/gs copy in the resolute-slim base removes a real confinement layer
#      or is a no-op that can be dropped.
#   2. Does the effective ImageMagick policy allow the PDF coder? An image without our own policy.xml
#      inherits the distro default. Debian and Ubuntu disabled PS/PDF there after the 2018 Ghostscript
#      CVEs and have since dropped that block again, so the answer depends on the base image release.
#
# Run it on the DOCKER HOST (it needs the Docker CLI and a host that loads AppArmor policy).
# It does not modify any image; it only runs containers and reads their output.
#
# Usage:
#   scripts/dist/check-pdf-rendering.sh [image ...]
#
# With no arguments it checks the current production images. Example:
#   scripts/dist/check-pdf-rendering.sh photoprism/photoprism:ce photoprism/photoprism:preview

set -euo pipefail

IMAGES=("$@")

if [[ ${#IMAGES[@]} -eq 0 ]]; then
  IMAGES=("photoprism/photoprism:ce" "photoprism/photoprism:preview")
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "FATAL: docker CLI not found — run this on the Docker host, not inside the dev container." >&2
  exit 1
fi

echo "== Host AppArmor state =="
if [[ -d /sys/kernel/security/apparmor ]]; then
  echo "apparmor: enabled"
  # A loaded gs profile is the precondition for the confinement this checks.
  if [[ -r /sys/kernel/security/apparmor/profiles ]]; then
    if grep -qE '(^|/)gs( |\()' /sys/kernel/security/apparmor/profiles 2>/dev/null; then
      echo "gs profile: LOADED on host"
      grep -E '(^|/)gs( |\()' /sys/kernel/security/apparmor/profiles 2>/dev/null | sed 's/^/  /'
    else
      echo "gs profile: not loaded on host — the confinement under test cannot apply here"
    fi
  else
    echo "gs profile: cannot read /sys/kernel/security/apparmor/profiles (need root?)"
  fi
else
  echo "apparmor: NOT enabled on this host — the confinement under test cannot apply here"
fi
echo

# A minimal one-page PDF, written from the host so the container reads it over a bind mount.
# The bind mount is the point: the reported denial was for reads outside the container's own
# filesystem, so an in-image temp file would not reproduce it.
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

cat > "$WORKDIR/probe.pdf" <<'PDF'
%PDF-1.4
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj
3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 96 96]/Resources<</Font<</F1 4 0 R>>>>/Contents 5 0 R>>endobj
4 0 obj<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>endobj
5 0 obj<</Length 44>>stream
BT /F1 12 Tf 10 40 Td (PhotoPrism) Tj ET
endstream
endobj
trailer<</Root 1 0 R>>
PDF

chmod 0644 "$WORKDIR/probe.pdf"

# probe runs one command in the image with the work directory bind-mounted, and reports the
# exit status plus any output. Kept read-only except for the output file the delegate writes.
probe() {
  local image="$1"
  shift
  docker run --rm \
    -v "$WORKDIR:/probe" \
    -w /probe \
    --entrypoint /bin/bash \
    "$image" -lc "$*" 2>&1 || true
}

for image in "${IMAGES[@]}"; do
  echo "======================================================================"
  echo "Image: $image"
  echo "======================================================================"

  if ! docker image inspect "$image" >/dev/null 2>&1; then
    echo "SKIP: image not present locally — pull or build it first."
    echo
    continue
  fi

  echo "-- Ghostscript binaries on PATH --"
  probe "$image" 'command -v gs; ls -l /usr/bin/gs /usr/local/bin/gs 2>/dev/null; gs --version 2>/dev/null'
  echo

  echo "-- ImageMagick policy in effect --"
  # shellcheck disable=SC2016  # single quotes are deliberate: expand in the container, not here
  probe "$image" 'for f in /etc/ImageMagick-7/policy.xml /etc/ImageMagick-6/policy.xml; do
      [ -f "$f" ] && { echo "== $f =="; grep -E "PDF|PS|ghostscript|delegate|coder" "$f" | grep -v "^ *<!--" | sed "s/^/  /"; }
    done; echo "(none listed above means no policy file present)"'
  echo

  echo "-- Direct Ghostscript render (bypasses ImageMagick) --"
  probe "$image" 'gs -dNOPAUSE -dBATCH -sDEVICE=jpeg -r72 -sOutputFile=/probe/gs-direct.jpg /probe/probe.pdf >/dev/null 2>&1 && echo "gs: OK" || echo "gs: FAILED"'
  echo

  echo "-- ImageMagick delegate render (the path PhotoPrism actually uses) --"
  # Mirrors convertImageJpeg's document branch in internal/photoprism/convert_image_jpeg.go.
  # shellcheck disable=SC2016  # single quotes are deliberate: expand in the container, not here
  probe "$image" 'CONVERT=$(command -v convert || command -v magick);
    "$CONVERT" -colorspace sRGB -density 300 /probe/probe.pdf[0] -background white -alpha remove -alpha off -resize "720x720>" -quality 92 /probe/im-out.jpg && echo "convert: OK" || echo "convert: FAILED"'
  echo

  echo "-- Results --"
  probe "$image" 'ls -l /probe/gs-direct.jpg /probe/im-out.jpg 2>/dev/null || true'
  rm -f "$WORKDIR/gs-direct.jpg" "$WORKDIR/im-out.jpg"
  echo
done

cat <<'NOTE'
======================================================================
How to read this
======================================================================
  gs OK  + convert OK       PDF indexing works. If /usr/local/bin/gs exists, re-run after
                            renaming it to see whether /usr/bin/gs also succeeds — if it
                            does, the unconfined copy is a no-op and can be dropped.
  gs FAILED                 Ghostscript itself is blocked. Check `dmesg | grep -i apparmor`
                            on the host for DENIED lines naming the bind-mount path; that
                            confirms the profile attaches inside the container.
  gs OK  + convert FAILED   The ImageMagick policy blocks the PDF coder, not AppArmor — so
                            PDF indexing is broken by policy rather than by confinement.
  no gs binary              The image ships no Ghostscript; PDF covers cannot be generated.

Check the host kernel log alongside the run:
  sudo dmesg -T | grep -i 'apparmor.*DENIED' | tail -20
NOTE
