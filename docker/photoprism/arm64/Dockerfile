FROM ubuntu:20.04 as build

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
# Big thank you to Guy Sheffer for building this!
# https://github.com/guysoft
#
ENV LD_LIBRARY_PATH /root/.local/lib:/usr/local/lib:/usr/lib:/lib
ENV TF_CPP_MIN_LOG_LEVEL 0
RUN curl -L \
   "https://dl.photoprism.org/tensorflow/arm64/libtensorflow-arm64-1.15.2.tar.gz" | \
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
RUN npm install --unsafe-perm=true --allow-root -g npm
RUN npm config set cache ~/.cache/npm

# Install Go
ENV GOLANG_VERSION 1.15.6
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "f87515b9744154ffe31182da9341d0a61eb0795551173d242c8cad209239e492 *go.tgz" | sha256sum -c -; \
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

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build PhotoPrism
RUN make dep build-js install

# Build gosu
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/tianon/gosu

# Same base image as photoprism/development
FROM ubuntu:20.04

# Set environment variables
ENV DEBIAN_FRONTEND noninteractive

# Configure apt-get
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry
RUN echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends
RUN echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes
RUN echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin

# Install additional distribution packages
RUN apt-get update && apt-get upgrade && apt-get install -y --no-install-recommends \
        ca-certificates \
        mariadb-client \
        sqlite3 \
        tzdata \
        libheif-examples \
        darktable \
        rawtherapee \
        exiftool \
        ffmpeg && \
    apt-get dist-upgrade && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy dependencies
COPY --from=build /go/bin/gosu /bin/gosu
COPY --from=build /usr/lib/libtensorflow.so /usr/lib/libtensorflow.so
COPY --from=build /usr/lib/libtensorflow_framework.so /usr/lib/libtensorflow_framework.so

RUN ldconfig

# Set default umask and create photoprism user
RUN umask 0000 && useradd photoprism -m -d /photoprism
WORKDIR /photoprism

ENV TF_CPP_MIN_LOG_LEVEL 2

ENV PATH /photoprism/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV TMPDIR /tmp

# Storage path names
ENV PHOTOPRISM_ASSETS_PATH /photoprism/assets
ENV PHOTOPRISM_STORAGE_PATH /photoprism/storage
ENV PHOTOPRISM_BACKUP_PATH /var/lib/photoprism
ENV PHOTOPRISM_ORIGINALS_PATH /photoprism/originals
ENV PHOTOPRISM_IMPORT_PATH /photoprism/import
ENV PHOTOPRISM_LOG_FILENAME /photoprism/photoprism.log
ENV PHOTOPRISM_PID_FILENAME /photoprism/photoprism.pid

# Defaults for common config values
# See https://docs.photoprism.org/getting-started/config-options/
ENV PHOTOPRISM_DEBUG "false"
ENV PHOTOPRISM_PUBLIC "false"
ENV PHOTOPRISM_READONLY "false"
ENV PHOTOPRISM_UPLOAD_NSFW "true"
ENV PHOTOPRISM_DETECT_NSFW "false"
ENV PHOTOPRISM_EXPERIMENTAL "false"
ENV PHOTOPRISM_SITE_URL "http://localhost:2342/"
ENV PHOTOPRISM_SITE_TITLE "PhotoPrism"
ENV PHOTOPRISM_SITE_CAPTION "Browse Your Life"
ENV PHOTOPRISM_SITE_DESCRIPTION ""
ENV PHOTOPRISM_SITE_AUTHOR ""
ENV PHOTOPRISM_HTTP_HOST "0.0.0.0"
ENV PHOTOPRISM_HTTP_PORT 2342
ENV PHOTOPRISM_DATABASE_DRIVER "sqlite"
ENV PHOTOPRISM_DATABASE_SERVER ""
ENV PHOTOPRISM_DATABASE_NAME "photoprism"
ENV PHOTOPRISM_DATABASE_USER "photoprism"
ENV PHOTOPRISM_DATABASE_PASSWORD ""
ENV PHOTOPRISM_DISABLE_WEBDAV "false"
ENV PHOTOPRISM_DISABLE_SETTINGS "false"
ENV PHOTOPRISM_DISABLE_BACKUPS "false"
ENV PHOTOPRISM_DISABLE_EXIFTOOL "false"
ENV PHOTOPRISM_DISABLE_PLACES "false"
ENV PHOTOPRISM_DISABLE_TENSORFLOW "false"
ENV PHOTOPRISM_DARKTABLE_PRESETS "false"
ENV PHOTOPRISM_THUMB_FILTER "lanczos"
ENV PHOTOPRISM_THUMB_UNCACHED "false"
ENV PHOTOPRISM_THUMB_SIZE 2048
ENV PHOTOPRISM_THUMB_SIZE_UNCACHED 7680
ENV PHOTOPRISM_JPEG_SIZE 7680
ENV PHOTOPRISM_JPEG_QUALITY 92
ENV PHOTOPRISM_WORKERS 0
ENV PHOTOPRISM_WAKEUP_INTERVAL 900
ENV PHOTOPRISM_AUTO_INDEX 300
ENV PHOTOPRISM_AUTO_IMPORT 300

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
COPY --chown=root:root /docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
VOLUME /var/lib/photoprism

# Run server
CMD ["photoprism", "start"]
