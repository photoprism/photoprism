#!/usr/bin/env bash

if [[ -z $1 ]] && [[ -z $2 ]]; then
  echo "Usage: ${0##*/} [source.png] [dest]/{size}.png" 1>&2
  exit 1
fi

set -e

sizes=(16 20 29 32 40 48 50 55 56 60 64 72 76 80 100 114 120 128 144 152 160 167 172 175 180 192 196 200 216 224 256 267 400 500 512 1024)
app_sizes=(16 20 29 32 40 48 50 55 56 60 64 72 76 80 100 114 120 128 144 152 160 167 172 175 180 192 196 200 216 224 256 267 400 500 512)
gloss_sizes=(16 20 29 32 40 48 50 55 56 60 64 72 76 80 100 114 120 128 144 152 160 167 172 175 180 192 196 200 216 224 256 267 400)

# Check if source file exists.
if [ -f "$1" ]; then
  echo "Creating icons from $1..."
else
  echo "$1 not found"
  exit 1
fi

# Create dest folder.
mkdir -p "$2"
mkdir -p "$2/app"
mkdir -p "$2/gloss"
mkdir -p "$2/round"

# Create 1024x1024 icon with rounded corners.

# convert -size 1024x1024 xc:none -fill white -draw \
#    'roundRectangle 0,0 1024,1024 100,100' in.png \
#    -compose SrcIn -composite rounded.png

# convert -size "1024x1024" xc:none -fill white \
#  -draw 'roundRectangle 0,0 1024x1024 100,100' "$1" \
#  -compose SrcIn -composite "$2/1024.png"
# echo "$2/$i.png"

# Created square icons in all sizes.
for i in "${sizes[@]}"
do
  convert "$1" -resize "${i}x${i}^" -gravity center -extent "${i}x${i}" "$2/$i.png"
  echo "$2/$i.png"
done

# Create rounded app icons.
convert -size "1024x1024" xc:none -fill white -draw "roundRectangle 0,0 1024,1024 179,179" "$2/1024.png" -compose SrcIn -composite "$2/app/1024.png"
for i in "${app_sizes[@]}"
do
  convert "$2/app/1024.png" -resize "${i}x${i}" "$2/app/$i.png"
  echo "$2/app/$i.png"
done

# Create glossy app icons.
SCRIPT_DIR=$(dirname "$0")
GLOSS_PNG="$SCRIPT_DIR/gloss-512.png"
convert -size "512x512" xc:none -fill white -draw "roundRectangle 0,0 512,512 50,50" "$2/512.png" -compose SrcIn -composite "$2/gloss/512.png"
convert -draw "image Screen 0,0 0,0 '${GLOSS_PNG}'" "$2/gloss/512.png" "$2/gloss/512.png"
for i in "${gloss_sizes[@]}"
do
  convert "$2/gloss/512.png" -resize "${i}x${i}" "$2/gloss/$i.png"
  echo "$2/gloss/$i.png"
done

# Create round icons.
convert -size "1024x1024" xc:none -fill white -draw "circle 512,512 512,0" "$2/1024.png" -compose SrcIn -composite "$2/round/1024.png"
for i in "${app_sizes[@]}"
do
  convert "$2/round/1024.png" -resize "${i}x${i}" "$2/round/$i.png"
  echo "$2/round/$i.png"
done

echo "Done."