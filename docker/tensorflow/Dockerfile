FROM ubuntu:22.04

LABEL maintainer="PhotoPrism UG <hello@photoprism.app>"

ENV DEBIAN_FRONTEND=noninteractive
ENV TMP=/tmp

ENV TF_VERSION=2.18.0

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
    build-essential \
    python3 \
    ca-certificates \
    curl \
    wget \
    git \
    lsb-release \
    software-properties-common \
    gnupg \
    jq 

# Install bazelisk
RUN wget https://apt.llvm.org/llvm.sh && chmod u+x llvm.sh && \
    ./llvm.sh 17 && rm llvm.sh && \
    ln -s /usr/bin/python3 /usr/bin/python && \
    ln -s /usr/bin/clang-17 /usr/bin/clang && \
    ln -s /usr/bin/clang++-17 /usr/bin/clang++ && \
    ln -s /usr/bin/clang-cpp /usr/bin/clang-cpp && \
    export BAZELISK_VERSION=$(curl -L https://api.github.com/repos/bazelbuild/bazelisk/releases/latest | jq -r .tag_name) && \
    curl -L https://github.com/bazelbuild/bazelisk/releases/download/${BAZELISK_VERSION}/bazelisk-linux-amd64 -o /usr/bin/bazel && \
    chmod +x /usr/bin/bazel && \
    mkdir -p /home/tensorflow

WORKDIR "/home/tensorflow"

# Configure TensorFlow
COPY ./*.sh ./
COPY ./Makefile Makefile
