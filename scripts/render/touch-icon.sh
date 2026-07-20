#!/usr/bin/env bash

# Renders full-bleed "touch" icon variants used by the apple-touch-icon links, so round theme
# icons get a squircle that fills the iOS home-screen mask shape instead of showing filled
# transparent corners. Round icons provide a "<name>.touch.svg" source (the background circle
# replaced by a rounded square); the default "logo" reuses the pre-rendered "app" squircle.

if [[ -n $1 ]] && [[ $1 == "-h" || $1 == "--help" ]]; then
  echo "Usage: (1) ${0##*/}                 (renders touch icons for all assets/static/icons/*.touch.svg sources + logo)" 1>&2
  echo "       (2) ${0##*/} [name]          (renders touch icons for assets/static/icons/[name].touch.svg only)" 1>&2
  exit 1
fi

set -e

# Sizes must match the apple-touch-icon ladder in assets/templates/favicons.gohtml
# (the sizes iOS actually requests for home-screen web clips: iPhone 180, iPad Pro 167,
# iPad 152, older iPhone 120).
sizes=(120 152 167 180)
icons_dir="assets/static/icons"

# render_touch_svg renders a "<name>.touch.svg" source into "<name>/touch/<size>.png".
render_touch_svg() {
  local svg="$1"
  local name
  name="$(basename "$svg" .touch.svg)"
  local dest="${icons_dir}/${name}/touch"

  echo "Creating touch icons from ${svg}..."
  mkdir -p "$dest"

  for size in "${sizes[@]}"; do
    rsvg-convert -a -w "$size" -h "$size" "$svg" > "$dest/$size.png"
    echo "$dest/$size.png"
  done
}

# copy_app_touch reuses the "app" squircle as the "logo" default's touch variant.
copy_app_touch() {
  local dest="${icons_dir}/logo/touch"

  echo "Copying touch icons for logo from ${icons_dir}/app..."
  mkdir -p "$dest"

  for size in "${sizes[@]}"; do
    cp "${icons_dir}/app/${size}.png" "$dest/${size}.png"
    echo "$dest/${size}.png"
  done
}

if [[ -n $1 ]]; then
  if [[ $1 == "logo" ]]; then
    copy_app_touch
  elif [ -f "${icons_dir}/${1}.touch.svg" ]; then
    render_touch_svg "${icons_dir}/${1}.touch.svg"
  else
    echo "${icons_dir}/${1}.touch.svg not found" 1>&2
    exit 1
  fi
else
  copy_app_touch
  for svg in "${icons_dir}"/*.touch.svg; do
    [ -e "$svg" ] || continue
    render_touch_svg "$svg"
  done
fi

echo "Done."
