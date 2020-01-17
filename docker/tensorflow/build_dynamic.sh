#!/bin/bash
time bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so

