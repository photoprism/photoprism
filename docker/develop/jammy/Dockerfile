#### Base Image: Ubuntu 22.04 LTS (Jammy Jellyfish)
FROM ubuntu:jammy

# Copyright © 2018 - 2024 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

# Add Open Container Initiative (OCI) annotations.
# See: https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.title="PhotoPrism® Build Image (Ubuntu 22.04 LTS)"
LABEL org.opencontainers.image.description="Ubuntu 22.04 LTS (Jammy Jellyfish)"
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
    NODE_ENV="production" \
    PS1="\u@$BUILD_TAG:\w\$ " \
    PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:/opt/photoprism/bin" \
    LD_LIBRARY_PATH="/usr/local/lib:/usr/lib" \
    DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    MALLOC_ARENA_MAX="4" \
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
    apt-get update && apt-get -qq dist-upgrade && \
    apt-get -qq install \
        libc6 ca-certificates bash sudo nano avahi-utils jq lsof lshw \
        exiftool sqlite3 tzdata gpg make zip unzip wget curl rsync \
        imagemagick libvips-dev rawtherapee ffmpeg libavcodec-extra x264 x265 libde265-dev \
        libaom-dev libvpx-dev libwebm-dev libjpeg-dev libmatroska-dev libdvdread-dev libebml5 libgav1-bin libatomic1 \
    && \
    apt-get -qq install \
        apt-utils pkg-config software-properties-common \
        build-essential gcc g++ git gettext davfs2 chrpath apache2-utils \
        autoconf automake cmake libtool libjpeg-dev libpng-dev libwebp-dev libavcodec-dev \
        libx264-dev libx265-dev libaom-dev libvpx-dev libwebm-dev libxft-dev \
        libc6-dev libhdf5-serial-dev libzmq3-dev libssl-dev libnss3 \
        libfreetype6 libfreetype6-dev libfontconfig1 libfontconfig1-dev fonts-roboto \
        librsvg2-bin ghostscript gsfonts pdf2svg ps2eps \
    && \
    /scripts/install-nodejs.sh && \
    /scripts/install-mariadb.sh mariadb-client && \
    /scripts/install-tensorflow.sh && \
    /scripts/install-darktable.sh && \
    /scripts/install-libheif.sh && \
    /scripts/install-chrome.sh && \
    echo "ALL ALL=(ALL) NOPASSWD:SETENV: ALL" >> /etc/sudoers.d/all && \
    mkdir -p /etc/skel/.config/go && \
    echo 'off 2024-08-27' > '/etc/skel/.config/go/telemetry' && \
    cp -r /etc/skel/.config /root/.config && \
    GOLANG_VERSION=go1.22.6 /scripts/install-go.sh && \
    /scripts/install-go-tools.sh && \
    echo 'alias go=richgo ll="ls -alh"' >> /etc/skel/.bashrc && \
    echo 'export PS1="\u@$DOCKER_TAG:\w\$ "' >> /etc/skel/.bashrc && \
    cp /etc/skel/.bashrc /root/.bashrc && \
    cp /scripts/convert/policy.xml /etc/ImageMagick-6/policy.xml && \
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
