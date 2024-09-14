#### Base Image: Ubuntu 21.10 (Impish Indri)
FROM ubuntu:impish

# Copyright © 2018 - 2024 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

# Add Open Container Initiative (OCI) annotations.
# See: https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.title="PhotoPrism® Build Image (Ubuntu 21.10)"
LABEL org.opencontainers.image.description="Ubuntu 21.10 (Impish Indri)"
LABEL org.opencontainers.image.url="https://hub.docker.com/repository/docker/photoprism/develop"
LABEL org.opencontainers.image.source="https://github.com/photoprism/photoprism"
LABEL org.opencontainers.image.documentation="https://docs.photoprism.app/developer-guide/setup/"
LABEL org.opencontainers.image.authors="PhotoPrism UG <hello@photoprism.app>"
LABEL org.opencontainers.image.vendor="PhotoPrism UG"

# Declare build parameters.
ARG TARGETARCH
ARG BUILD_TAG

# Set environment variables, see https://docs.photoprism.app/getting-started/config-options/.
ENV PHOTOPRISM_ARCH=$TARGETARCH \
    DOCKER_TAG=$BUILD_TAG \
    DOCKER_ENV="develop" \
    PS1="\u@$BUILD_TAG:\w\$ " \
    PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:/opt/photoprism/bin" \
    LD_LIBRARY_PATH="/usr/local/lib:/usr/lib" \
    NODE_ENV="production" \
    DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    MALLOC_ARENA_MAX="4" \
    GOPATH="/go" \
    GOBIN="/usr/local/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr" \
    PROG="photoprism"

# copy scripts and debian backports sources list
COPY --chown=root:root --chmod=755 /scripts/dist/ /scripts/
COPY --chown=root:root --chmod=644 /.my.cnf /etc/my.cnf

# update image and install build dependencies
RUN echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing && \
    echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/force-confold && \
    apt-get update && apt-get -qq dist-upgrade && apt-get -qq install --no-install-recommends \
      apt-utils \
      gpg \
      gpg-agent \
      pkg-config \
      software-properties-common \
      ca-certificates \
      build-essential \
      g++ \
      gcc \
      sudo \
      bash \
      make \
      nano \
      lsof \
      lshw \
      wget \
      curl \
      jq \
      git \
      zip \
      unzip \
      gettext \
      firefox \
      mariadb-client \
      davfs2 \
      chrpath \
      libc6-dev \
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
      libnss3 \
      libxtst6 \
      librsvg2-bin \
      rsync \
      sqlite3 \
      tzdata \
      libheif-examples \
      exiftool \
      rawtherapee \
      ffmpeg \
      libavcodec-extra \
      apache2-utils \
      fonts-roboto \
    && \
    /scripts/install-nodejs.sh && \
    /scripts/install-tensorflow.sh && \
    /scripts/install-darktable.sh && \
    /scripts/install-chrome.sh && \
    /scripts/install-go.sh && \
    /scripts/install-go-tools.sh && \
    echo 'alias go=richgo ll="ls -alh"' >> /etc/skel/.bashrc && \
    echo "ALL ALL=(ALL) NOPASSWD:SETENV: ALL" >> /etc/sudoers.d/all && \
    echo 'export PS1="\u@$DOCKER_TAG:\w\$ "' >> /etc/skel/.bashrc && \
    cp /etc/skel/.bashrc /root/.bashrc && \
    /scripts/create-users.sh && \
    install -d -m 0777 -o 1000 -g 1000 \
      /photoprism/originals \
      /photoprism/import \
      /photoprism/storage \
      /photoprism/storage/sidecar \
      /photoprism/storage/albums \
      /photoprism/storage/backups \
      /photoprism/storage/config \
      /photoprism/storage/cache && \
    /scripts/cleanup.sh

# download models and testdata
RUN mkdir /tmp/photoprism && \
    wget "https://dl.photoprism.app/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.app/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.app/tensorflow/facenet.zip?${BUILD_TAG}" -O /tmp/photoprism/facenet.zip && \
    wget "https://dl.photoprism.app/qa/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# expose the following container ports:
# - 2342 (HTTP)
# - 2343 (Acceptance Tests)
# - 9515 (Chromedriver)
# - 40000 (Go Debugger)
EXPOSE 2342 2343 9515 40000

# set container entrypoint script
ENTRYPOINT ["/scripts/entrypoint.sh"]

# keep container running
CMD ["tail", "-f", "/dev/null"]
