#!/usr/bin/env bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

######################################## TEST STORAGE FOLDER PERMISSIONS ########################################

STORAGE_PATH=${PHOTOPRISM_STORAGE_PATH:-/photoprism/storage}

DOC_URL="https://docs.photoprism.app/getting-started/troubleshooting/docker/#file-permissions"

set -e

# create directory if not exists
mkdir -p "${STORAGE_PATH}" || (echo "Failed creating storage folder \"$STORAGE_PATH\", see $DOC_URL" 1>&2; exit 1)

# create and delete test file
(touch "${STORAGE_PATH}/is-writable" 2>/dev/null && rm "${STORAGE_PATH}/is-writable") || \
  (echo "Failed creating test file in storage folder, see $DOC_URL" 1>&2; exit 1)
