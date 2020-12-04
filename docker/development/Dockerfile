FROM ubuntu:20.04

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

ARG BUILD_TAG

ENV DEBIAN_FRONTEND noninteractive

# Configure apt-get
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin

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
    libx264-dev \
    libx265-dev \
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
    firefox \
    libheif-examples \
    exiftool \
    rawtherapee \
    ffmpeg \
    lsof \
    apache2-utils \
    fonts-roboto \
    sudo

# Install Chromium Browser via .deb
RUN add-apt-repository -y ppa:xalt7x/chromium-deb-vaapi
COPY /docker/development/pin-xalt7x-chromium-deb-vaapi /etc/apt/preferences.d/pin-xalt7x-chromium-deb-vaapi

# Install RAW to JPEG converter
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/graphics:/darktable/xUbuntu_20.04/ /' > /etc/apt/sources.list.d/graphics:darktable.list" && \
    wget -qO - https://download.opensuse.org/repositories/graphics:darktable/xUbuntu_20.04/Release.key | apt-key add - && \
    apt-get update && \
    apt-get install chromium-browser chromium-codecs-ffmpeg-extra darktable && \
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

# Set default https://en.wikipedia.org/wiki/Umask
RUN umask 0000

# Install and configure NodeJS Package Manager (npm)
ENV NODE_ENV production
RUN npm install --unsafe-perm=true --allow-root -g npm testcafe chromedriver
RUN npm config set cache ~/.cache/npm

# Install Go
ENV GOLANG_VERSION 1.15.5
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "9a58494e8da722c3aef248c9227b0e9c528c7318309827780f16220998180a0d *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

# Configure Go environment
ENV GOPATH /go
ENV GOBIN $GOPATH/bin
ENV PATH $GOBIN:/usr/local/go/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV TMPDIR /tmp
ENV GO111MODULE on
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Download TensorFlow model and test files
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism && \
    wget "https://dl.photoprism.org/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.org/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.org/fixtures/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# Install additional tools for development
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/psampaz/go-mod-outdated && \
    env GO111MODULE=off /usr/local/go/bin/go get -u golang.org/x/tools/cmd/goimports && \
    env GO111MODULE=off /usr/local/go/bin/go get -u github.com/tsliwowicz/go-wrk && \
    env GO111MODULE=off /usr/local/go/bin/go get -u github.com/kyoh86/richgo && \
    env GO111MODULE=off /usr/local/go/bin/go get -u github.com/tianon/gosu && \
    echo "alias go=richgo" > /root/.bash_aliases

# Create photoprism user and directory for deployment
RUN useradd photoprism -m -d /photoprism && \
    mkdir -p /var/lib/photoprism && \
    echo "alias go=richgo" > /photoprism/.bash_aliases && \
    echo "photoprism ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers && \
    chown -R photoprism:photoprism /photoprism /var/lib/photoprism && \
    chmod -R a+rw /photoprism /var/lib/photoprism /go && \
    find /go -type d -print0 | xargs -0 chmod 777

# Copy mysql client config for development
COPY --chown=root:root /docker/development/.my.cnf /root/.my.cnf
COPY --chown=photoprism:photoprism /docker/development/.my.cnf /photoprism/.my.cnf
RUN chmod 644 /root/.my.cnf /photoprism/.my.cnf

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Expose HTTP port 2342, 2343 plus 9515 for chromedriver
EXPOSE 2342 2343 9515
VOLUME /var/lib/photoprism

# Keep container running (services can be started manually using a terminal)
CMD tail -f /dev/null
