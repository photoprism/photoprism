#!/usr/bin/env bash

if [[ ! -e assets/tensorflow/tensorflow_inception_graph.pb ]]; then
  wget "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip" -O assets/tensorflow/inception.zip &&
  unzip assets/tensorflow/inception.zip -d assets/tensorflow
else
  echo "TensorFlow InceptionV3 model already downloaded."
fi