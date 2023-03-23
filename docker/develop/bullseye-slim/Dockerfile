#### Base Image: Debian 11 (Bullseye)
FROM debian:bullseye-slim

# Copyright © 2018 - 2023 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

# Add Open Container Initiative (OCI) annotations.
# See: https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.title="PhotoPrism® Base Image (Debian 11)"
LABEL org.opencontainers.image.description="Debian 11 (Bullseye)"
LABEL org.opencontainers.image.url="https://hub.docker.com/repository/docker/photoprism/develop"
LABEL org.opencontainers.image.source="https://github.com/photoprism/photoprism"
LABEL org.opencontainers.image.documentation="https://docs.photoprism.app/developer-guide/setup/"
LABEL org.opencontainers.image.authors="Michael Mayer <hello@photoprism.app>"
LABEL org.opencontainers.image.vendor="PhotoPrism UG"

# Declare build parameters.
ARG TARGETARCH
ARG BUILD_TAG

# Set environment variables, see https://docs.photoprism.app/getting-started/config-options/.
ENV PHOTOPRISM_ARCH=$TARGETARCH \
    DOCKER_TAG=$BUILD_TAG \
    DOCKER_ENV="prod" \
    PS1="\u@$DOCKER_TAG:\w\$ " \
    PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/opt/photoprism/bin" \
    LD_LIBRARY_PATH="/usr/local/lib:/usr/lib" \
    TMPDIR="/tmp" \
    DEBIAN_FRONTEND="noninteractive" \
    TF_CPP_MIN_LOG_LEVEL="2" \
    PROG="photoprism"

# copy scripts and debian backports sources list
COPY --chown=root:root --chmod=755 /scripts/dist/ /scripts/
COPY --chown=root:root --chmod=644 /docker/develop/bullseye/sources.list /etc/apt/sources.list.d/bullseye.list

# install additional distribution packages
RUN echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing && \
    echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/force-confold && \
    apt-get update && apt-get -qq dist-upgrade && apt-get -qq install --no-install-recommends \
      ca-certificates \
      jq \
      zip \
      gpg \
      lshw \
      wget \
      curl \
      make \
      sudo \
      bash \
      sqlite3 \
      tzdata \
      libc6 \
      libatomic1 \
      libheif-examples \
      librsvg2-bin \
      exiftool \
      rawtherapee \
      imagemagick \
      libvips-dev \
      ffmpeg \
      libavcodec-extra \
      x264 \
      x265 \
      libvpx6 \
      libwebm1 \
    && \
    /scripts/install-mariadb.sh mariadb-client && \
    /scripts/install-darktable.sh && \
    echo 'alias ll="ls -alh"' >> /etc/skel/.bashrc && \
    echo 'export PS1="\u@$DOCKER_TAG:\w\$ "' >> /etc/skel/.bashrc && \
    echo "ALL ALL=(ALL) NOPASSWD:SETENV: /scripts/entrypoint-init.sh" >> /etc/sudoers.d/init && \
    cp /etc/skel/.bashrc /root/.bashrc && \
    /scripts/create-users.sh && \
    cp /scripts/heif-convert.sh /usr/local/bin/heif-convert && \
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

# define default directory and user
WORKDIR /photoprism

# Expose HTTP(S) ports.
EXPOSE 2342 2443

# keep container running
CMD ["tail", "-f", "/dev/null"]