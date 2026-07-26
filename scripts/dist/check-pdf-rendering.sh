#!/usr/bin/env bash

# check-pdf-rendering.sh determines whether a container image can rasterize a PDF through
# ImageMagick's Ghostscript delegate, and reports which layer blocks it when it cannot.
#
# The result cannot be settled by reading the Dockerfiles: a policy rule whose pattern matches no
# registered coder, module, or delegate name is silently ignored while still printing in
# "-list policy", so an inert rule looks exactly like an enforced one. The checks below therefore
# probe the running image rather than read its configuration. An image without our own policy.xml
# inherits the distro default, which varies by release.
#
# Run it on the DOCKER HOST (it needs the Docker CLI). It does not modify any image; it only runs
# containers and reads their output.
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
  # A loaded gs profile on the host can affect whether Ghostscript runs inside the container.
  if [[ -r /sys/kernel/security/apparmor/profiles ]]; then
    if grep -qE '(^|/)gs( |\()' /sys/kernel/security/apparmor/profiles 2>/dev/null; then
      echo "gs profile: LOADED on host"
      grep -E '(^|/)gs( |\()' /sys/kernel/security/apparmor/profiles 2>/dev/null | sed 's/^/  /'
    else
      echo "gs profile: not loaded on host — host policy cannot affect the result here"
    fi
  else
    echo "gs profile: cannot read /sys/kernel/security/apparmor/profiles (need root?)"
  fi
else
  echo "apparmor: NOT enabled on this host — host policy cannot affect the result here"
fi
echo

# A minimal one-page PDF, written from the host so the container reads it over a bind mount.
# The bind mount is the point: host policy applies to reads outside the container's own
# filesystem, so an in-image temp file would not exercise the same path.
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

# Fixtures for the policy probes. A 1x1 PNG gives the MSL and indirect-read probes something
# real to read, so "no output file" means the policy refused rather than that the input was bad.
base64 -d > "$WORKDIR/policy-probe.png" <<'PNG'
iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==
PNG

cat > "$WORKDIR/policy-probe.msl" <<'MSL'
<?xml version="1.0" encoding="UTF-8"?>
<image>
  <read filename="/probe/policy-probe.png"/>
  <write filename="/probe/policy-msl-out.png"/>
</image>
MSL

echo "/probe/policy-probe.png" > "$WORKDIR/policy-probe-list.txt"

chmod 0644 "$WORKDIR"/probe.pdf "$WORKDIR"/policy-probe.*  "$WORKDIR"/policy-probe-list.txt

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

  echo "-- ImageMagick policy file present --"
  # shellcheck disable=SC2016  # single quotes are deliberate: expand in the container, not here
  probe "$image" 'found=no
    for f in /etc/ImageMagick-7/policy.xml /etc/ImageMagick-6/policy.xml; do
      [ -f "$f" ] || continue
      found=yes
      if grep -q PhotoPrism "$f"; then echo "  $f (PhotoPrism policy)"; else echo "  $f (distribution default)"; fi
    done
    if [ "$found" = no ]; then echo "  none — the image ships no ImageMagick policy at all"; fi
    CONVERT=$(command -v convert || command -v magick)
    if [ -n "$CONVERT" ]; then echo "  rules loaded: $("$CONVERT" -list policy 2>/dev/null | grep -c "Policy:")"; fi'
  echo

  echo "-- ImageMagick policy enforcement (probed, not read) --"
  # Each denial is confirmed by attempting the operation. Reading the file or counting rules in
  # "-list policy" cannot distinguish an enforced rule from one whose pattern matches nothing.
  # shellcheck disable=SC2016  # single quotes are deliberate: expand in the container, not here
  probe "$image" 'CONVERT=$(command -v convert || command -v magick)
    verdict() { if [ -s "$2" ]; then echo "  $1 ALLOWED"; else echo "  $1 blocked"; fi; }
    rm -f /probe/policy-text-out.png /probe/policy-msl-out.png /probe/policy-at-out.png /probe/policy-http-out.png
    "$CONVERT" text:/etc/hostname /probe/policy-text-out.png >/dev/null 2>&1
    verdict "text: coder ............" /probe/policy-text-out.png
    "$CONVERT" msl:/probe/policy-probe.msl null: >/dev/null 2>&1
    verdict "msl: coder ............." /probe/policy-msl-out.png
    "$CONVERT" @/probe/policy-probe-list.txt /probe/policy-at-out.png >/dev/null 2>&1
    verdict "@file indirect read ...." /probe/policy-at-out.png
    ERR=$("$CONVERT" http://127.0.0.1:9/probe.png /probe/policy-http-out.png 2>&1 | head -1)
    if [ -s /probe/policy-http-out.png ]; then
      echo "  http: coder ............ ALLOWED"
    elif echo "$ERR" | grep -qiE "security policy|NotAuthorized"; then
      echo "  http: coder ............ blocked"
    elif echo "$ERR" | grep -qiE "no decode delegate|unable to open image"; then
      echo "  http: coder ............ blocked (not resolved as a URL)"
    elif echo "$ERR" | grep -qiE "delegate|curl"; then
      echo "  http: coder ............ ALLOWED (reached the fetch, then failed to connect)"
    else
      echo "  http: coder ............ INCONCLUSIVE: $ERR"
    fi'
  echo

  echo "-- ImageMagick resource limits in effect --"
  # shellcheck disable=SC2016  # single quotes are deliberate: expand in the container, not here
  probe "$image" 'CONVERT=$(command -v convert || command -v magick); "$CONVERT" -list resource 2>/dev/null | sed "s/^/  /"'
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
  rm -f "$WORKDIR/gs-direct.jpg" "$WORKDIR/im-out.jpg" "$WORKDIR"/policy-*-out.png
  echo
done

cat <<'NOTE'
======================================================================
How to read this
======================================================================
  gs OK  + convert OK       PDF indexing works. Ghostscript installation in the base images
                            is deliberate — see the related specs before changing it.
  gs FAILED                 Ghostscript itself is blocked. Check `dmesg | grep -i apparmor`
                            on the host for DENIED lines naming the bind-mount path.
  gs OK  + convert FAILED   The ImageMagick policy blocks the PDF coder — PDF indexing is
                            broken by policy rather than by the host.
  no gs binary              The image ships no Ghostscript; PDF covers cannot be generated.

  text/msl/@file ALLOWED    Our policy is not in force. Either the image carries the distribution
                            default, or a rule is present but its pattern matches nothing.
  http: ALLOWED             ImageMagick can still make outbound requests while processing media.
                            "reached the fetch" means the coder was permitted and only the
                            connection failed, which is an allowed result, not a block.
  http: INCONCLUSIVE        Unrecognized error text — read the line and classify it by hand rather
                            than assuming the block held.

  Resource limits matter as much as the denials: Disk below roughly 10GiB or Width/Height below
  the PHOTOPRISM_JPEG_SIZE ceiling rejects valid media with "cache resources exhausted" or
  "width or height exceeds limit".

Check the host kernel log alongside the run:
  sudo dmesg -T | grep -i 'apparmor.*DENIED' | tail -20
NOTE
