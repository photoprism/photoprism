#!/usr/bin/env bash

# Renders safe-zone-padded "maskable" PWA icons from the theme icon SVGs, so Android fills the
# home-screen adaptive-icon mask shape instead of letterboxing the default icon. The logo is
# rendered at 80% of the target size (10% safe-zone padding per side) and centered on an opaque
# background. See https://web.dev/articles/maskable-icon for the safe-zone specification.

if [[ -n $1 ]] && [[ $1 == "-h" || $1 == "--help" ]]; then
  echo "Usage: (1) ${0##*/}                 (renders maskable icons for all assets/static/icons/*.svg themes)" 1>&2
  echo "       (2) ${0##*/} [name]          (renders maskable icons for assets/static/icons/[name].svg only)" 1>&2
  exit 1
fi

set -e

sizes=(192 512)
safe_zone=80   # percentage of the icon that holds the visible logo (the rest is mask-safe padding)
background="white"

# Collect the theme SVG sources to render.
sources=()
if [[ -n $1 ]]; then
  sources=("assets/static/icons/${1}.svg")
else
  for svg in assets/static/icons/*.svg; do
    sources+=("$svg")
  done
fi

for svg in "${sources[@]}"; do
  if [ ! -f "$svg" ]; then
    echo "$svg not found" 1>&2
    exit 1
  fi

  name="$(basename "$svg" .svg)"
  dest="assets/static/icons/${name}/maskable"
  mkdir -p "$dest"

  echo "Creating maskable icons from ${svg}..."

  for size in "${sizes[@]}"; do
    inner=$(( size * safe_zone / 100 ))
    tmp="$(mktemp --suffix=.png)"
    rsvg-convert -a -w "$inner" -h "$inner" "$svg" > "$tmp"
    magick -size "${size}x${size}" "xc:${background}" "$tmp" -gravity center -composite -depth 8 -strip "$dest/$size.png"
    rm -f "$tmp"
    echo "$dest/$size.png"
  done
done

echo "Done."
