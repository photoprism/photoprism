FROM ubuntu:21.04 as build

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

ARG BUILD_TAG

ENV DEBIAN_FRONTEND noninteractive

# Configure apt
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
    wget \
    git \
    gettext \
    tzdata \
    gconf-service

# Install & configure TensorFlow for C
#
# Big thank you to Qengineering for building this!
# https://github.com/Qengineering/TensorFlow-Raspberry-Pi
#
ENV LD_LIBRARY_PATH="/root/.local/lib:/usr/local/lib:/usr/lib:/lib" \
    TF_CPP_MIN_LOG_LEVEL="0"
RUN curl -L \
   "https://github.com/Qengineering/TensorFlow-Raspberry-Pi/blob/master/libtensorflow_1_15_2.tar.gz?raw=true" | \
   tar -C "/usr" -xz
RUN ldconfig

# Install NodeJS
RUN curl -sL https://deb.nodesource.com/setup_14.x | bash -
RUN apt-get update && \
    apt-get install nodejs && \
    apt-get dist-upgrade && \
    apt-get autoremove && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install and configure NodeJS Package Manager (npm)
ENV NODE_ENV production
RUN npm install --unsafe-perm=true --allow-root -g npm
RUN npm config set cache ~/.cache/npm

# Install Go
ENV GOPATH="/go" \
    GOBIN="/go/bin" \
    PATH="/go/bin:/usr/local/go/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr" \
    GOLANG_VERSION=1.16.4
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"; \
	curl -o go.tgz -L "$url"; \
	echo "a53391a800ddec749ee90d38992babb27b95cfb864027350c737b9aa8e069494 *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	go version
RUN mkdir -p "/go/src" "/go/bin" && chmod -R 777 "/go"

# Download TensorFlow model and test files
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism
RUN curl "https://dl.photoprism.org/tensorflow/nsfw.zip?${BUILD_TAG}" -o /tmp/photoprism/nsfw.zip
RUN curl "https://dl.photoprism.org/tensorflow/nasnet.zip?${BUILD_TAG}" -o /tmp/photoprism/nasnet.zip

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build PhotoPrism
ENV NODE_OPTIONS "$NODE_OPTIONS --max-old-space-size=2048"
RUN make dep build-js install

# Build gosu
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/tianon/gosu

# Same base image as photoprism/development
FROM ubuntu:21.04

# Set environment variables
# see https://docs.photoprism.org/getting-started/config-options/
ENV DEBIAN_FRONTEND="noninteractive" \
    TF_CPP_MIN_LOG_LEVEL="2" \
    PATH="/photoprism/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
    TMPDIR="/tmp" \
    PHOTOPRISM_ASSETS_PATH="/photoprism/assets" \
    PHOTOPRISM_STORAGE_PATH="/photoprism/storage" \
    PHOTOPRISM_BACKUP_PATH="/var/lib/photoprism" \
    PHOTOPRISM_ORIGINALS_PATH="/photoprism/originals" \
    PHOTOPRISM_IMPORT_PATH="/photoprism/import" \
    PHOTOPRISM_LOG_FILENAME="/photoprism/photoprism.log" \
    PHOTOPRISM_PID_FILENAME="/photoprism/photoprism.pid" \
    PHOTOPRISM_DEBUG="false" \
    PHOTOPRISM_PUBLIC="false" \
    PHOTOPRISM_READONLY="false" \
    PHOTOPRISM_UPLOAD_NSFW="true" \
    PHOTOPRISM_DETECT_NSFW="false" \
    PHOTOPRISM_EXPERIMENTAL="false" \
    PHOTOPRISM_SITE_URL="http://localhost:2342/" \
    PHOTOPRISM_SITE_TITLE="PhotoPrism" \
    PHOTOPRISM_SITE_CAPTION="Browse Your Life" \
    PHOTOPRISM_SITE_DESCRIPTION="" \
    PHOTOPRISM_SITE_AUTHOR="" \
    PHOTOPRISM_HTTP_HOST="0.0.0.0" \
    PHOTOPRISM_HTTP_PORT="2342" \
    PHOTOPRISM_DATABASE_DRIVER="sqlite" \
    PHOTOPRISM_DATABASE_SERVER="" \
    PHOTOPRISM_DATABASE_NAME="photoprism" \
    PHOTOPRISM_DATABASE_USER="photoprism" \
    PHOTOPRISM_DATABASE_PASSWORD="" \
    PHOTOPRISM_DISABLE_WEBDAV="false" \
    PHOTOPRISM_DISABLE_SETTINGS="false" \
    PHOTOPRISM_DISABLE_BACKUPS="false" \
    PHOTOPRISM_DISABLE_EXIFTOOL="false" \
    PHOTOPRISM_DISABLE_PLACES="false" \
    PHOTOPRISM_DISABLE_TENSORFLOW="false" \
    PHOTOPRISM_DARKTABLE_PRESETS="false" \
    PHOTOPRISM_THUMB_FILTER="lanczos" \
    PHOTOPRISM_THUMB_UNCACHED="false" \
    PHOTOPRISM_THUMB_SIZE="2048" \
    PHOTOPRISM_THUMB_SIZE_UNCACHED="7680" \
    PHOTOPRISM_JPEG_SIZE="7680" \
    PHOTOPRISM_JPEG_QUALITY="92" \
    PHOTOPRISM_WORKERS="0" \
    PHOTOPRISM_WAKEUP_INTERVAL="900" \
    PHOTOPRISM_AUTO_INDEX="300" \
    PHOTOPRISM_AUTO_IMPORT="300"

# Configure apt
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry
RUN echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends
RUN echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes
RUN echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin

# Install additional distribution packages
RUN apt-get update && apt-get upgrade && apt-get install -y --no-install-recommends \
        wget \
        curl \
        davfs2 \
        ca-certificates \
        mariadb-client \
        sqlite3 \
        tzdata \
        libheif-examples \
        libatomic1 \
        rawtherapee \
        exiftool \
        ffmpeg \
        ffmpegthumbnailer \
        libavcodec-extra && \
    apt-get dist-upgrade && \
    apt-get autoremove && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy dependencies
COPY --from=build /go/bin/gosu /bin/gosu
COPY --from=build /usr/lib/libtensorflow.so /usr/lib/libtensorflow.so
COPY --from=build /usr/lib/libtensorflow_framework.so /usr/lib/libtensorflow_framework.so
COPY --chown=root:root --chmod=755 /docker/scripts/heif-convert.sh /usr/local/bin/heif-convert

RUN ldconfig

# Set default umask and create photoprism user
RUN umask 0000 && useradd photoprism -m -d /photoprism
WORKDIR /photoprism

# Copy files to /photoprism
COPY --from=build /root/.local/bin/photoprism /photoprism/bin/photoprism
COPY --from=build /root/.photoprism/assets /photoprism/assets

# Create directories
RUN mkdir -p \
    /var/lib/photoprism \
    /tmp/photoprism \
    /photoprism/originals \
    /photoprism/import \
    /photoprism/storage/config \
    /photoprism/storage/cache && \
    chown -Rf photoprism:photoprism /photoprism /var/lib/photoprism /tmp/photoprism && \
    chmod -Rf a+rw /photoprism /var/lib/photoprism /tmp/photoprism

# Show photoprism version
RUN photoprism -v

# Expose http port
EXPOSE 2342

# Configure entrypoint
COPY --chown=root:root /docker/scripts/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
VOLUME /var/lib/photoprism

# Run server
CMD ["photoprism", "start"]
