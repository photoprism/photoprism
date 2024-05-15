#### Base Image: Ubuntu 24.04 LTS (Noble Numbat)
FROM ubuntu:noble

# Copyright © 2018 - 2024 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

# Add Open Container Initiative (OCI) annotations.
# See: https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.title="PhotoPrism® Build Image (ARMv7)"
LABEL org.opencontainers.image.description="Ubuntu 24.04 LTS (Noble Numbat)"
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
    PS1="\u@$DOCKER_TAG:\w\$ " \
    PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:/opt/photoprism/bin" \
    LD_LIBRARY_PATH="/usr/local/lib:/usr/lib" \
    NODE_ENV="production" \
    DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    GOPATH="/go" \
    GOBIN="/usr/local/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr" \
    PROG="photoprism"

# Copy scripts and package sources config.
COPY --chown=root:root --chmod=755 /scripts/dist/ /scripts/
COPY --chown=root:root --chmod=644 /.my.cnf /etc/my.cnf

# Update base image and add dependencies.
RUN echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing && \
    echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/force-confold && \
    apt-get update && apt-get -qq dist-upgrade && apt-get -qq install --no-install-recommends \
        apt-utils gpg pkg-config software-properties-common ca-certificates avahi-utils \
        build-essential gcc g++ sudo bash make nano lsof lshw git jq \
        autoconf automake cmake libtool libjpeg-dev libpng-dev libwebp-dev \
        zip unzip wget curl rsync sqlite3 chrpath gettext libc6-dev \
        libssl-dev libxft-dev libfreetype6 libfreetype6-dev libfontconfig1 \
        libfontconfig1-dev libhdf5-serial-dev libzmq3-dev \
        libx264-dev libx265-dev libde265-dev libaom-dev libnss3 libxtst6 librsvg2-bin tzdata \
        exiftool ffmpeg libavcodec-extra libdav1d-dev \
    && \
    /scripts/install-nodejs.sh && \
    /scripts/install-libheif.sh && \
    /scripts/install-tensorflow.sh && \
    /scripts/install-go.sh && \
    /scripts/install-go-tools.sh && \
    echo 'alias ll="ls -alh"' >> /etc/skel/.bashrc && \
    echo 'export PS1="\u@$DOCKER_TAG:\w\$ "' >> /etc/skel/.bashrc && \
    echo "ALL ALL=(ALL) NOPASSWD:SETENV: ALL" >> /etc/sudoers.d/all && \
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

# Download models and testdata.
RUN mkdir /tmp/photoprism && \
    wget "https://dl.photoprism.app/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.app/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.app/tensorflow/facenet.zip?${BUILD_TAG}" -O /tmp/photoprism/facenet.zip && \
    wget "https://dl.photoprism.app/qa/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# Default working directory.
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Expose the following container ports:
# - 2342 (HTTP)
# - 2343 (Acceptance Tests)
# - 2442 (HTTP)
# - 2443 (HTTPS)
# - 9515 (Chromedriver)
# - 40000 (Go Debugger)
EXPOSE 2342 2343 2442 2443 9515 40000

# Declare container entrypoint script.
ENTRYPOINT ["/scripts/entrypoint.sh"]

# Keep container running.
CMD ["tail", "-f", "/dev/null"]

