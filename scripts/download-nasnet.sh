#!/usr/bin/env bash

TODAY=`date -u +%Y%m%d`

MODEL_NAME="NASNet Mobile"
MODEL_URL="https://dl.photoprism.org/tensorflow/nasnet.zip?$TODAY"
MODEL_PATH="assets/resources/nasnet"
MODEL_ZIP="/tmp/photoprism/nasnet.zip"
MODEL_HASH="f18b801354e95cade497b4f12e8d2537d04c04f6  $MODEL_ZIP"
MODEL_VERSION="$MODEL_PATH/version.txt"
MODEL_BACKUP="assets/backups/nasnet-$TODAY"

echo "Installing $MODEL_NAME for TensorFlow..."

# Create directories
mkdir -p /tmp/photoprism
mkdir -p assets/backups

# Check for update
if [[ -f ${MODEL_ZIP} ]] && [[ `sha1sum ${MODEL_ZIP}` == ${MODEL_HASH} ]]; then
  if [[ -f ${MODEL_VERSION} ]]; then
    echo "Already up to date."
    exit
  fi
else
  # Download model
  echo "Downloading latest model from $MODEL_URL..."
  wget ${MODEL_URL} -O ${MODEL_ZIP}

  TMP_HASH=`sha1sum ${MODEL_ZIP}`

  echo ${TMP_HASH}
fi

# Create backup
if [[ -e ${MODEL_PATH} ]]; then
  echo "Creating backup of existing directory: $MODEL_BACKUP"
  rm -rf ${MODEL_BACKUP}
  mv ${MODEL_PATH} ${MODEL_BACKUP}
fi

# Unzip model
unzip ${MODEL_ZIP} -d assets/resources
echo "$MODEL_NAME $TODAY $MODEL_HASH" > ${MODEL_VERSION}

echo "Latest $MODEL_NAME installed."
