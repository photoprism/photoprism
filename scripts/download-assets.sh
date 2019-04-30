#!/usr/bin/env bash

FILENAME="/tmp/photoprism/nasnet.zip"

if [[ ! -e assets/tensorflow/nasnet/saved_model.pb ]]; then
  if [[ ! -e ${FILENAME} ]]; then
      mkdir -p /tmp/photoprism
      wget "https://dl.photoprism.org/tensorflow/nasnet.zip" -O ${FILENAME}
  fi

  mkdir -p assets/tensorflow
  unzip ${FILENAME} -d assets/tensorflow
else
  echo "TensorFlow model already downloaded."
fi
