FROM ubuntu:18.04

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

ARG BUILD_TAG

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
    build-essential \
    curl \
    chrpath \
    libssl-dev \
    libxft-dev \
    libfreetype6 \
    libfreetype6-dev \
    libfontconfig1 \
    libfontconfig1-dev \
    libhdf5-serial-dev \
    libpng-dev \
    libzmq3-dev \
    pkg-config \
    software-properties-common \
    rsync \
    unzip \
    zip \
    g++ \
    gcc \
    libc6-dev \
    gpg-agent \
    apt-utils \
    make \
    nano \
    wget \
    git \
    gettext \
    mysql-client \
    tzdata \
    gconf-service \
    chromium-browser \
    firefox \
    libheif-examples \
    exiftool \
    ffmpeg \
    lsof

# Install RAW to JPEG converter
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/graphics:/darktable/xUbuntu_18.04/ /' > /etc/apt/sources.list.d/graphics:darktable.list" && \
    wget -qO - https://download.opensuse.org/repositories/graphics:darktable/xUbuntu_18.04/Release.key | apt-key add - && \
    apt-get update && \
    apt-get install darktable && \
    apt-get upgrade && \
    apt-get dist-upgrade

# Install & configure TensorFlow for C,
# see https://www.tensorflow.org/install/lang_c
#
# We also have custom builds available:
# https://dl.photoprism.org/tensorflow/linux/
#
ENV LD_LIBRARY_PATH /root/.local/lib:/usr/local/lib:/usr/lib:/lib
ENV TF_CPP_MIN_LOG_LEVEL 0
RUN curl -L \
   "https://dl.photoprism.org/tensorflow/linux/libtensorflow-linux-cpu-1.15.2.tar.gz" | \
   tar -C "/usr" -xz
RUN ldconfig

# Install NodeJS
RUN curl -sL https://deb.nodesource.com/setup_14.x | bash -
RUN apt-get update && \
    apt-get install nodejs && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install and configure NodeJS Package Manager (npm)
ENV NODE_ENV production
RUN npm install --unsafe-perm=true --allow-root -g npm testcafe chromedriver
RUN npm config set cache ~/.cache/npm

# Install Go
ENV GOLANG_VERSION 1.14.4
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "aed845e4185a0b2a3c3d5e1d0a35491702c55889192bb9c30e67a3de6849c067 *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

# Configure Go environment
ENV GOPATH /go
ENV GOBIN $GOPATH/bin
ENV PATH $GOBIN:/usr/local/go/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV GO111MODULE on
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Download TensorFlow model and test files
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism
RUN wget "https://dl.photoprism.org/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip
RUN wget "https://dl.photoprism.org/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip
RUN wget "https://dl.photoprism.org/fixtures/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# Install additional tools for development
RUN env GO111MODULE=off /usr/local/go/bin/go get -u golang.org/x/tools/cmd/goimports
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/miku/zek/cmd/...
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/tsliwowicz/go-wrk
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/kyoh86/richgo
RUN echo "alias go=richgo" > /root/.bash_aliases

# MariaDB test database settings
COPY /docker/development/.my.cnf /root/.my.cnf

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Expose HTTP port 2342 plus 9515 for chromedriver
EXPOSE 2342 9515

# Keep container running (services can be started manually using a terminal)
CMD tail -f /dev/null
