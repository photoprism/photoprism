#!/usr/bin/env bash

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Usage: $0 [platform] [tf-version]" 1>&2
    exit 1
fi

echo "Creating 'libtensorflow-$1-$2.tar.gz'...";

rm -rf tmp
mkdir -p tmp/lib/
mkdir -p tmp/include/tensorflow/c/eager/
mkdir -p tmp/include/tensorflow/core/platform
mkdir -p tmp/include/tsl/platform
mkdir -p tmp/include/xla/tsl/c

cp bazel-bin/tensorflow/libtensorflow* tmp/lib/
cp tensorflow/c/eager/*.h tmp/include/tensorflow/c/eager/
cp tensorflow/c/*.h LICENSE tmp/include/tensorflow/c/
cp tensorflow/core/platform/*.h tmp/include/tensorflow/core/platform
cp third_party/xla/third_party/tsl/tsl/platform/*.h tmp/include/tsl/platform
cp third_party/xla/xla/tsl/c/*.h tmp/include/xla/tsl/c
(cd tmp && tar -czf ../libtensorflow-$1-$2.tar.gz .)
du -h libtensorflow-$1-$2.tar.gz

echo "Done"
