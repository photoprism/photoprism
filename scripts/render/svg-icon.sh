#!/usr/bin/env bash

if [[ -z $1 ]]; then
  echo "Usage: (1) ${0##*/} assets/static/icons/[name].svg (icons are rendered as assets/static/icons/[name]/{size}.png)" 1>&2
  echo "       (2) ${0##*/} [source.svg] [dest]/{size}.png" 1>&2
  exit 1
fi

set -e

sizes=(16 20 29 32 40 48 50 55 56 60 64 72 76 80 100 114 120 128 144 152 160 167 172 175 180 192 196 200 216 256 267 400 512 1024)

if [[ -z $2 ]]; then
  # Check if source file exists.
  if [ -f "assets/static/icons/$1.svg" ]; then
    echo "Creating icons from assets/static/icons/$1.svg..."
  else
    echo "assets/static/icons/$1.svg not found"
    exit 1
  fi

  # Create dest folder.
  mkdir -p "assets/static/icons/$1"

  # Create icons in all sizes.
  for i in "${sizes[@]}"
  do
   rsvg-convert -a -w $i -h $i "assets/static/icons/$1.svg" > "assets/static/icons/$1/$i.png"
   echo "assets/static/icons/$1/$i.png"
  done
else
  # Check if source file exists.
  if [ -f "$1" ]; then
    echo "Creating icons from $1..."
  else
    echo "$1 not found"
    exit 1
  fi

  # Create dest folder.
  mkdir -p "$2"

  # Create icons in all sizes.
  for i in "${sizes[@]}"
  do
   rsvg-convert -a -w $i -h $i $1 > "$2/$i.png"
   echo "$2/$i.png"
  done
fi

echo "Done."