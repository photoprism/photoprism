#### Base Image: Debian 12, Codename 'Bookworm'
FROM debian:bookworm-slim

# Open Container Initiative (OCI) Annotations
# see https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.title="PhotoPrism Build Environment"
LABEL org.opencontainers.image.description="Debian 12, Codename 'Bookworm'"
LABEL org.opencontainers.image.url="https://hub.docker.com/repository/docker/photoprism/develop"
LABEL org.opencontainers.image.source="https://github.com/photoprism/photoprism"
LABEL org.opencontainers.image.documentation="https://docs.photoprism.app/developer-guide/setup/"
LABEL org.opencontainers.image.authors="Michael Mayer <hello@photoprism.app>"
LABEL org.opencontainers.image.vendor="PhotoPrism UG"

ARG TARGETARCH
ARG BUILD_TAG

# set environment variables
ENV PHOTOPRISM_ARCH=$TARGETARCH \
    DOCKER_TAG=$BUILD_TAG \
    DOCKER_ENV="develop" \
    NODE_ENV="production" \
    PS1="\u@$DOCKER_TAG:\w\$ " \
    PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:/opt/photoprism/bin" \
    LD_LIBRARY_PATH="/usr/local/lib:/usr/lib" \
    DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    GOPATH="/go" \
    GOBIN="/go/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr" \
    PROG="photoprism"

# copy scripts and debian backports sources list
COPY --chown=root:root --chmod=755 /scripts/dist/* /scripts/
COPY --chown=root:root --chmod=644 /docker/develop/bookworm/sources.list /etc/apt/sources.list.d/bookworm.list
COPY --chown=root:root --chmod=644 /.my.cnf /etc/my.cnf

# update image and install build dependencies
RUN echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing && \
    apt-get update && apt-get -qq dist-upgrade && \
    apt-get -qq install \
      libc6 ca-certificates sudo bash tzdata \
      gpg zip unzip wget curl rsync make nano \
      jq lsof lshw sqlite3 mariadb-client \
      exiftool darktable rawtherapee libheif-examples librsvg2-bin \
      ffmpeg ffmpegthumbnailer libavcodec-extra \
      libmatroska7 libdvdread8 libebml5 libgav1-0 libatomic1 \
      x264 x265 libvpx7 libwebm1 \
    && \
    apt-get -qq install \
      apt-utils pkg-config software-properties-common \
      build-essential gcc g++ git gettext davfs2 chrpath apache2-utils \
      chromium chromium-driver chromium-sandbox firefox-esr \
      libx264-dev libx265-dev libvpx-dev libwebm-dev libpng-dev libxft-dev \
      libc6-dev libhdf5-serial-dev libzmq3-dev libssl-dev libnss3 \
      libfreetype6 libfreetype6-dev libfontconfig1 libfontconfig1-dev fonts-roboto \
    && \
    /scripts/install-nodejs.sh && \
    /scripts/install-tensorflow.sh && \
    /scripts/install-go.sh && \
    /scripts/install-go-tools.sh && \
    echo 'alias go=richgo ll="ls -alh"' >> /etc/skel/.bashrc && \
    echo 'export PS1="\u@$DOCKER_TAG:\w\$ "' >> /etc/skel/.bashrc && \
    echo "ALL ALL=(ALL) NOPASSWD:SETENV: ALL" >> /etc/sudoers.d/all && \
    cp /etc/skel/.bashrc /root/.bashrc && \
    /scripts/create-users.sh && \
    /scripts/cleanup.sh && \
    cp /scripts/heif-convert.sh /usr/local/bin/heif-convert && \
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
      /photoprism/storage/cache

# download models and testdata
RUN wget "https://dl.photoprism.app/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
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
