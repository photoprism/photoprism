#!/usr/bin/env bash

STORAGE_PATH=${PHOTOPRISM_STORAGE_PATH:-/photoprism/storage}

URL="https://docs.photoprism.app/getting-started/troubleshooting/docker/#file-permissions"

# shellcheck disable=SC2028
(touch "${STORAGE_PATH}/.writable" 2>/dev/null && rm "${STORAGE_PATH}/.writable") || \
  (printf "The storage folder \"%s\" is not writable, please fix filesystem permissions: %s\n" "$STORAGE_PATH" "$URL"; exit 1)
