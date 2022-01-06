#!/usr/bin/env bash

if [[ -z $1 ]] && [[ -z $2 ]]; then
    echo "USAGE: heif-convert <filename> <output>" 1>&2
    exit 1
fi


# USAGE: heif-convert [-q quality 0..100] <filename> <output>

/usr/bin/heif-convert -q 92 "$1" "$2"

# Reset Exif orientation flag if output image was rotated based on "QuickTime:Rotation"

if [[ $(exiftool -n -QuickTime:Rotation "$1") ]]; then
    /usr/bin/exiftool -overwrite_original -P -n '-ModifyDate<FileModifyDate' -Orientation=1 "$2"
else
    /usr/bin/exiftool -overwrite_original -P -n '-ModifyDate<FileModifyDate' "$2"
fi
