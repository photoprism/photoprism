#!/usr/bin/env bash

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Usage: $0 [platform] [tf-version]" 1>&2
    exit 1
fi

echo "Creating 'libtensorflow-$1-$2.tar.gz'...";

rm -rf tmp
mkdir -p tmp/lib/
mkdir -p tmp/include/tensorflow/c/eager/
cp bazel-bin/tensorflow/libtensorflow.so.$2 tmp/lib/libtensorflow.so
cp bazel-bin/tensorflow/libtensorflow_framework.so.$2 tmp/lib/libtensorflow_framework.so
cp tensorflow/c/eager/c_api.h tmp/include/tensorflow/c/eager/
cp tensorflow/c/c_api.h tensorflow/c/c_api_experimental.h LICENSE tmp/include/tensorflow/c/
(cd tmp && tar -czf ../libtensorflow-$1-$2.tar.gz .)
du -h libtensorflow-$1-$2.tar.gz

echo "Done"
