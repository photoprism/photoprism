#!/usr/bin/env bash

if [[ -z $1 ]] && [[ -z $2 ]]; then
    echo "USAGE: heif-convert <filename> <output>" 1>&2
    exit 1
fi


# USAGE: heif-convert [-q quality 0..100] <filename> <output>

/usr/bin/heif-convert -q 92 "$1" "$2"