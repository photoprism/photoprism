FROM golang:bullseye

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILD_TAG
ARG GOPROXY
ARG GODEBUG

# set environment variables
ENV DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp" \
    LD_LIBRARY_PATH="/root/.local/lib:/usr/local/lib:/usr/lib:/lib" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    NODE_ENV="production" \
    GOPATH="/go" \
    GOBIN="/go/bin" \
    PATH="/go/bin:/usr/local/go/bin:~/.local/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr"

# configure apt
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing

# copy scripts to /root/.local/bin
COPY --chown=root:root --chmod=755 /docker/scripts/*.sh /root/.local/bin/

# update image and install build dependencies
RUN apt-get update && apt-get -qq dist-upgrade && apt-get -qq install --no-install-recommends \
    apt-utils \
    gpg-agent \
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
    wget \
    curl \
    rsync \
    unzip \
    zip \
    git \
    gettext \
    chromium \
    chromium \
    chromium-driver \
    mariadb-client \
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
    darktable \
    rawtherapee \
    ffmpeg \
    ffmpegthumbnailer \
    libavcodec-extra \
    davfs2 \
    chrpath \
    lsof \
    apache2-utils && \
    curl -sL https://deb.nodesource.com/setup_14.x | bash - && \
    apt-get update && apt-get -qq install nodejs && \
    npm install --unsafe-perm=true --allow-root -g npm && \
    npm config set cache ~/.cache/npm && \
    apt-get -y autoremove && apt-get -y autoclean && apt-get -y clean && rm -rf /var/lib/apt/lists/* && \
    /root/.local/bin/install-tensorflow.sh ${TARGETARCH} && \
    mkdir -p "/go/src" "/go/bin" && \
    chmod -R 777 "/go"

# download TensorFlow models & example files for testing
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism && \
    wget "https://dl.photoprism.app/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.app/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.app/tensorflow/facenet.zip?${BUILD_TAG}" -O /tmp/photoprism/facenet.zip && \
    wget "https://dl.photoprism.app/qa/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# copy additional scripts to image
COPY --chown=root:root /docker/scripts/heif-convert.sh /usr/local/bin/heif-convert
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh

# install Go tools
RUN /usr/local/go/bin/go install github.com/tianon/gosu@latest && \
    /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@latest && \
    /usr/local/go/bin/go install github.com/kyoh86/richgo@latest && \
    /usr/local/go/bin/go install github.com/psampaz/go-mod-outdated@latest && \
    /usr/local/go/bin/go install github.com/dsoprea/go-exif/v3/command/exif-read-tool@latest; \
    echo "alias go=richgo" > /root/.bash_aliases && \
    cp /go/bin/gosu /bin/gosu

# create photoprism user and directory for deployment
RUN useradd -m -U -u 1000 -d /photoprism photoprism && chmod a+rwx /photoprism && \
    mkdir -m 777 -p /var/lib/photoprism /tmp/photoprism && \
    echo "alias go=richgo" > /photoprism/.bash_aliases && \
    echo "photoprism ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers && \
    chown -Rf photoprism:photoprism /photoprism /var/lib/photoprism /tmp/photoprism && \
    chmod -Rf a+rw /photoprism /var/lib/photoprism /tmp/photoprism /go && \
    chmod 755 /usr/local/bin/heif-convert /entrypoint.sh && \
    find /go -type d -print0 | xargs -0 chmod 777

# copy mysql client config for develop
COPY --chown=root:root /docker/develop/.my.cnf /root/.my.cnf
COPY --chown=photoprism:photoprism /docker/develop/.my.cnf /photoprism/.my.cnf
RUN chmod 644 /root/.my.cnf /photoprism/.my.cnf

# set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# expose HTTP ports: 2342 (HTTP), 2343 (Acceptance Tests), 9515 (Chromedriver), 40000 (Go)
EXPOSE 2342 2343 9515 40000

# configure entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# keep container running
CMD ["tail", "-f", "/dev/null"]