#### Base Image: Debian 11, Codename "Bullseye"
FROM golang:1.18-bullseye

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

ARG TARGETARCH
ARG BUILD_TAG

# set environment variables
ENV PHOTOPRISM_ARCH=$TARGETARCH \
    DOCKER_TAG=$BUILD_TAG \
    DOCKER_ENV="develop" \
    NODE_ENV="production" \
    DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp" \
    LD_LIBRARY_PATH="/root/.local/lib:/usr/local/lib:/usr/lib:/lib" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    GOPATH="/go" \
    GOBIN="/go/bin" \
    PATH="/usr/local/sbin:/usr/sbin:/sbin:/bin:/scripts:/usr/local/go/bin:/go/bin:/usr/local/bin:/usr/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr"

# copy scripts and debian backports sources list
COPY --chown=root:root --chmod=755 /scripts/dist/* /scripts/
COPY --chown=root:root --chmod=644 /docker/develop/bullseye/sources.list /etc/apt/sources.list.d/bullseye.list
COPY --chown=root:root --chmod=644 /.my.cnf /etc/my.cnf

# update image and install build dependencies
RUN echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing && \
    groupadd -f -r -g 44 video && \
    groupadd -f -r -g 109 render && \
    groupadd -f -g 1000 photoprism && \
    useradd -m -g 1000 -u 1000 -d /photoprism -G video,render photoprism && \
    chmod 777 /photoprism && \
    apt-get update && apt-get -qq dist-upgrade && apt-get -qq install --no-install-recommends \
      apt-utils \
      gpg \
      pkg-config \
      software-properties-common \
      ca-certificates \
      build-essential \
      gcc \
      g++ \
      sudo \
      bash \
      make \
      nano \
      lsof \
      lshw \
      wget \
      curl \
      rsync \
      jq \
      git \
      zip \
      unzip \
      gettext \
      chromium \
      chromium-driver \
      chromium-sandbox \
      firefox-esr \
      sqlite3 \
      libc6-dev \
      libssl-dev \
      libxft-dev \
      libhdf5-serial-dev \
      libpng-dev \
      libheif-examples \
      librsvg2-bin \
      libzmq3-dev \
      libx264-dev \
      libx265-dev \
      libnss3 \
      libfreetype6 \
      libfreetype6-dev \
      libfontconfig1 \
      libfontconfig1-dev \
      fonts-roboto \
      tzdata \
      exiftool \
      rawtherapee \
      ffmpeg \
      ffmpegthumbnailer \
      libavcodec-extra \
      davfs2 \
      chrpath \
      apache2-utils \
    && \
    /scripts/install-nodejs.sh && \
    /scripts/install-mariadb.sh mariadb-client && \
    /scripts/install-tensorflow.sh && \
    /scripts/install-darktable.sh && \
    /scripts/cleanup.sh && \
    mkdir -p "/go/src" "/go/bin" && \
    chmod -R 777 "/go" && \
    install -d -m 0777 -o 1000 -g 1000 \
      /var/lib/photoprism \
      /tmp/photoprism \
      /photoprism/originals \
      /photoprism/import \
      /photoprism/storage \
      /photoprism/storage/sidecar \
      /photoprism/storage/albums \
      /photoprism/storage/backups \
      /photoprism/storage/config \
      /photoprism/storage/cache \
    && \
    wget "https://dl.photoprism.app/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.app/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.app/tensorflow/facenet.zip?${BUILD_TAG}" -O /tmp/photoprism/facenet.zip && \
    wget "https://dl.photoprism.app/qa/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# install Go tools
RUN /usr/local/go/bin/go install github.com/tianon/gosu@latest && \
    /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@latest && \
    /usr/local/go/bin/go install github.com/kyoh86/richgo@latest && \
    /usr/local/go/bin/go install github.com/psampaz/go-mod-outdated@latest && \
    /usr/local/go/bin/go install github.com/dsoprea/go-exif/v3/command/exif-read-tool@latest; \
    cp /go/bin/gosu /bin/gosu && \
    echo "alias go=richgo ll='ls -alh'" > /photoprism/.bash_aliases && \
    echo "alias go=richgo ll='ls -alh'" > /root/.bash_aliases && \
    echo "ALL ALL=(ALL) NOPASSWD:SETENV: ALL" >> /etc/sudoers.d/all && \
    cp /scripts/heif-convert.sh /usr/local/bin/heif-convert && \
    chmod -R a+rwX /go

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
