#!/usr/bin/env bash

CPU_DETECTED=$(/usr/bin/lshw -c processor -json 2>/dev/null)

if [[ $(echo "${CPU_DETECTED}" | /usr/bin/jq -r '.[].capabilities.avx2') == "true" ]]; then
  echo "avx2"
elif [[ $(echo "${CPU_DETECTED}" | /usr/bin/jq -r '.[].capabilities.avx') == "true" ]]; then
  echo "avx"
fi