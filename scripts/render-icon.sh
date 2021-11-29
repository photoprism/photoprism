#!/usr/bin/env bash

set -e

sizes=(16 20 29 32 40 48 50 55 56 60 64 72 76 80 100 114 120 128 144 152 160 167 172 175 180 192 196 200 216 256 267 400 512 1024)

if [[ -z $1 ]] && [[ -z $2 ]]; then
  echo "Please provide a source SVG and target PNG name" 1>&2
  exit 1
elif [[ $1 ]] && [[ -z $2 ]]; then
  mkdir -p "assets/static/icons/$1"

  if [ -f "assets/static/icons/$1.svg" ]; then
    echo "creating png icons from assets/static/icons/$1.svg..."
  else
    echo "assets/static/icons/$1.svg not found"
  fi

  for i in "${sizes[@]}"
  do
   rsvg-convert -a -w $i -h $i "assets/static/icons/$1.svg" > "assets/static/icons/$1/$i.png"
   echo "assets/static/icons/$1/$i.png"
  done
else
  if [ -f "$1" ]; then
    echo "creating png icons from $1..."
  else
    echo "$1 not found"
  fi

  for i in "${sizes[@]}"
  do
   rsvg-convert -a -w $i -h $i $1 > "$2-$i.png"
   echo "$2-$i.png"
  done
fi

echo "Done"