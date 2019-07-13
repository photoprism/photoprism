#!/usr/bin/env bash
# Copyright 2016 The TensorFlow Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
# Downloads and builds all of TensorFlow's dependencies for Linux, and compiles
# the TensorFlow library itself. Supported on Ubuntu 14.04 and 16.04.

set -e

SCRIPT_DIR="tensorflow/contrib/makefile"

source "${SCRIPT_DIR}/build_helper.subr"

# Note: Limit the number of jobs to 1 on ARM64 to prevent the build from being killed
JOB_COUNT="${JOB_COUNT:-$(get_job_count)}"

# Set CPU architecture e.g. core-avx-i, haswell or armv8-a
# See https://gcc.gnu.org/onlinedocs/gcc/x86-Options.html
# and https://gcc.gnu.org/onlinedocs/gcc/AArch64-Options.html
ARCH=${ARCH:-core-avx-i}

# Remove any old files first.
make -f tensorflow/contrib/makefile/Makefile clean
rm -rf tensorflow/contrib/makefile/downloads

# Pull down the required versions of the frameworks we need.
tensorflow/contrib/makefile/download_dependencies.sh

# Compile nsync.
# Don't use  export var=`something` syntax; it swallows the exit status.
HOST_NSYNC_LIB=`tensorflow/contrib/makefile/compile_nsync.sh`
TARGET_NSYNC_LIB="$HOST_NSYNC_LIB"
export HOST_NSYNC_LIB TARGET_NSYNC_LIB

# Compile protobuf.
tensorflow/contrib/makefile/compile_linux_protobuf.sh

# Build TensorFlow for ARCH.
make -j"${JOB_COUNT}" -f tensorflow/contrib/makefile/Makefile \
  OPTFLAGS="-O3 -march=${ARCH}" \
  HOST_CXXFLAGS="--std=c++11 -march=${ARCH}" \
  MAKEFILE_DIR="${SCRIPT_DIR}"
