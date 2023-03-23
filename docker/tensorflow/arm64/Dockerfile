FROM ubuntu:18.04

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

ENV DEBIAN_FRONTEND noninteractive
ENV TMP /tmp
ENV EXTRA_BAZEL_ARGS "--host_javabase=@local_jdk//:jdk"

# apt default settings
RUN echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin && \
    echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/force-confold

# Install dev / build dependencies
RUN apt-get update && apt-get upgrade && \
    apt-get install \
    ca-certificates \
    build-essential \
    autoconf \
    automake \
    libtool \
    g++-4.8 \
    gcc-4.8 \
    libc6-dev \
    zlib1g-dev \
    libssl-dev \
    curl \
    chrpath \
    pkg-config \
    unzip \
    zip \
    make \
    nano \
    wget \
    git \
    libtool \
    python3 \
    python3-git \
    openjdk-8-jdk

# Use GCC 4.8 and Python 3 as default
# See https://www.tensorflow.org/install/source#tested_build_configurations
RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 10 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-4.8 10 && \
    update-alternatives --install /usr/bin/python python /usr/bin/python3.6 10

# Download Bazel & TensorFlow
WORKDIR "/home/tensorflow"
RUN wget https://github.com/tensorflow/tensorflow/archive/v1.15.2.tar.gz
RUN tar -xzf v1.15.2.tar.gz

# Install Bazel
# RUN wget https://github.com/bazelbuild/bazel/releases/download/0.24.1/bazel-0.24.1-linux-x86_64
RUN wget https://github.com/guysoft/bazel-bin/raw/master/bazel-0.24.1-aarch64
RUN mv bazel-0.24.1-aarch64 /usr/local/bin/bazel && chmod 755 /usr/local/bin/bazel

# Configure TensorFlow
WORKDIR "/home/tensorflow/tensorflow-1.15.2"
COPY ./*.sh ./
COPY ./.tf_configure.bazelrc .tf_configure.bazelrc
COPY ./Makefile Makefile