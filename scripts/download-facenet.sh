#!/usr/bin/env bash

TODAY=$(date -u +%Y%m%d)

MODEL_NAME="Facenet"
MODEL_URL="https://dl.photoprism.app/tensorflow/facenet.zip?$TODAY"
MODEL_PATH="assets/facenet"
MODEL_ZIP="/tmp/photoprism/facenet.zip"
MODEL_HASH="0492eb1d67789108b7eefb274e26633504b059be  $MODEL_ZIP"
MODEL_VERSION="$MODEL_PATH/version.txt"
MODEL_BACKUP="storage/backup/facenet-$TODAY"

echo "Installing $MODEL_NAME model for TensorFlow..."

# Create directories
mkdir -p /tmp/photoprism
mkdir -p storage/backup

# Check for update
if [[ -f ${MODEL_ZIP} ]] && [[ $(sha1sum ${MODEL_ZIP}) == ${MODEL_HASH} ]]; then
  if [[ -f ${MODEL_VERSION} ]]; then
    echo "Already up to date."
    exit
  fi
else
  # Download model
  echo "Downloading latest model from $MODEL_URL..."
  wget --inet4-only -c "${MODEL_URL}" -O ${MODEL_ZIP}

  TMP_HASH=$(sha1sum ${MODEL_ZIP})

  echo "${TMP_HASH}"
fi

# Create backup
if [[ -e ${MODEL_PATH} ]]; then
  echo "Creating backup of existing directory: $MODEL_BACKUP"
  rm -rf "${MODEL_BACKUP}"
  mv ${MODEL_PATH} "${MODEL_BACKUP}"
fi

# Unzip model
unzip ${MODEL_ZIP} -d assets
echo "$MODEL_NAME $TODAY $MODEL_HASH" > ${MODEL_VERSION}

echo "Latest $MODEL_NAME installed."
