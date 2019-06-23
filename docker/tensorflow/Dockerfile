FROM ubuntu:18.04

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

ENV DEBIAN_FRONTEND noninteractive

# Configure apt-get
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry
RUN echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends
RUN echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes
RUN echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin

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
    git

# Use GCC 4.8 as default compiler
# See https://www.tensorflow.org/install/source#tested_build_configurations
RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 10 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-4.8 10

# Set up project directory
WORKDIR "/home/tensorflow"

# Download TensorFlow
RUN wget https://github.com/tensorflow/tensorflow/archive/v1.13.1.tar.gz
RUN tar -xzf v1.13.1.tar.gz

# Build TensorFlow
WORKDIR "/home/tensorflow/tensorflow-1.13.1"
COPY /docker/tensorflow/build_all_linux_arch.sh tensorflow/contrib/makefile/build_all_linux_arch.sh
RUN env JOB_COUNT=1 ARCH=core-avx-i tensorflow/contrib/makefile/build_all_linux_arch.sh

# Keep container running
CMD tail -f /dev/null
