#!/usr/bin/env bash

FILENAME="/tmp/photoprism/inception.zip"

if [[ ! -e assets/tensorflow/inception/tensorflow_inception_graph.pb ]]; then
  if [[ ! -e ${FILENAME} ]]; then
      mkdir -p /tmp/photoprism
      wget "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip" -O ${FILENAME}
  fi

  mkdir -p assets/tensorflow/inception
  unzip ${FILENAME} -d assets/tensorflow/inception
else
  echo "TensorFlow InceptionV3 model already downloaded."
fi