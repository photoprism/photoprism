##################################################### BUILD STAGE ######################################################
FROM ubuntu:21.10 as build

LABEL maintainer="Michael Mayer <hello@photoprism.org>"

ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILD_TAG

# Set environment variables
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

# Configure apt
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing

# Copy scripts to /root/.local/bin
COPY --chown=root:root --chmod=755 /docker/scripts/*.sh /root/.local/bin/

# 1. Install distribution packages
# 2. Install TensorFlow for C
# 3. Install Chrome, NodeJS, NPM, Puppeteer, TestCafe & ChromeDriver
# 4. Install Go
RUN apt-get update && apt-get -qq dist-upgrade && apt dist-upgrade 2>/dev/null && apt-get -qq install --no-install-recommends \
    build-essential \
    ca-certificates \
    wget \
    curl \
    davfs2 \
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
    libnss3 \
    libxtst6 \
    librsvg2-bin \
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
    git \
    gettext \
    sqlite3 \
    tzdata \
    gconf-service \
    libheif-examples \
    exiftool \
    ffmpeg \
    ffmpegthumbnailer \
    libavcodec-extra \
    sudo && \
    /root/.local/bin/install-tensorflow.sh ${TARGETARCH} && \
    /root/.local/bin/install-devtools.sh ${TARGETARCH} && \
    /root/.local/bin/install-go.sh ${TARGETARCH} && \
    mkdir -p "/go/src" "/go/bin" && \
    chmod -R 777 "/go"

# Download TensorFlow models & example files for testing
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism && \
    wget "https://dl.photoprism.app/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.app/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.app/tensorflow/facenet.zip?${BUILD_TAG}" -O /tmp/photoprism/facenet.zip && \
    wget "https://dl.photoprism.app/qa/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# Copy additional scripts to image
COPY --chown=root:root /docker/scripts/heif-convert.sh /usr/local/bin/heif-convert
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile
COPY --chown=root:root /docker/development/entrypoint.sh /entrypoint.sh

# Install Go tools
RUN /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@latest && \
    /usr/local/go/bin/go install github.com/tianon/gosu@latest && \
    cp /go/bin/gosu /bin/gosu

# Create photoprism user and directory for deployment
RUN useradd -m -U -u 1000 -d /photoprism photoprism && chmod a+rwx /photoprism && \
    mkdir -m 777 -p /var/lib/photoprism /tmp/photoprism && \
    echo "alias go=richgo" > /photoprism/.bash_aliases && \
    echo "photoprism ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers && \
    chown -Rf photoprism:photoprism /photoprism /var/lib/photoprism /tmp/photoprism && \
    chmod -Rf a+rw /photoprism /var/lib/photoprism /tmp/photoprism /go && \
    chmod 755 /usr/local/bin/heif-convert /entrypoint.sh && \
    find /go -type d -print0 | xargs -0 chmod 777

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build frontend and backend
RUN make dep build-js install

################################################## PRODUCTION STAGE ####################################################
FROM ubuntu:21.10

LABEL maintainer="Michael Mayer <hello@photoprism.org>"

ARG TARGETARCH
ARG TARGETPLATFORM

# Set environment variables
ENV DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp"

# Configure apt
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing

# Install additional distribution packages
RUN apt-get update && apt-get -qq dist-upgrade && apt dist-upgrade 2>/dev/null && apt-get -qq install --no-install-recommends \
    gpgv \
    wget \
    curl \
    make \
    davfs2 \
    ca-certificates \
    mariadb-client \
    sqlite3 \
    tzdata \
    libc6 \
    libatomic1 \
    libheif-examples \
    librsvg2-bin \
    gnupg \
    gpg-agent \
    apt-utils \
    add-apt-key \
    exiftool \
    rawtherapee \
    ffmpeg \
    ffmpegthumbnailer \
    libavcodec-extra && \
    apt-get -y autoremove && apt-get -y autoclean && apt-get clean && rm -rf /var/lib/apt/lists/*

# Set environment variables, see https://docs.photoprism.app/getting-started/config-options/
ENV TF_CPP_MIN_LOG_LEVEL="2" \
    PATH="/photoprism/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
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
    PHOTOPRISM_HTTP_PORT=2342 \
    PHOTOPRISM_DATABASE_DRIVER="sqlite" \
    PHOTOPRISM_DATABASE_SERVER="" \
    PHOTOPRISM_DATABASE_NAME="photoprism" \
    PHOTOPRISM_DATABASE_USER="photoprism" \
    PHOTOPRISM_DATABASE_PASSWORD="" \
    PHOTOPRISM_DISABLE_CHOWN="false" \
    PHOTOPRISM_DISABLE_WEBDAV="false" \
    PHOTOPRISM_DISABLE_SETTINGS="false" \
    PHOTOPRISM_DISABLE_BACKUPS="false" \
    PHOTOPRISM_DISABLE_EXIFTOOL="false" \
    PHOTOPRISM_DISABLE_PLACES="false" \
    PHOTOPRISM_DISABLE_TENSORFLOW="false" \
    PHOTOPRISM_DISABLE_FACES="false" \
    PHOTOPRISM_DISABLE_CLASSIFICATION="false" \
    PHOTOPRISM_DARKTABLE_PRESETS="false" \
    PHOTOPRISM_THUMB_FILTER="lanczos" \
    PHOTOPRISM_THUMB_UNCACHED="false" \
    PHOTOPRISM_THUMB_SIZE=2048 \
    PHOTOPRISM_THUMB_SIZE_UNCACHED=7680 \
    PHOTOPRISM_JPEG_SIZE=7680 \
    PHOTOPRISM_JPEG_QUALITY=92 \
    PHOTOPRISM_WORKERS=0 \
    PHOTOPRISM_WAKEUP_INTERVAL=900 \
    PHOTOPRISM_AUTO_INDEX=300 \
    PHOTOPRISM_AUTO_IMPORT=300

# Copy dependencies
COPY --from=build /go/bin/gosu /bin/gosu
COPY --from=build /usr/lib/libtensorflow.so /usr/lib/libtensorflow.so
COPY --from=build /usr/lib/libtensorflow_framework.so /usr/lib/libtensorflow_framework.so
RUN ldconfig

# Set default umask and create photoprism user
RUN umask 0000 && useradd -m -U -u 1000 -d /photoprism photoprism && chmod a+rwx /photoprism
WORKDIR /photoprism

# Copy additional files to image
COPY --from=build /root/.local/bin/photoprism /photoprism/bin/photoprism
COPY --from=build /root/.photoprism/assets /photoprism/assets
COPY --chown=root:root /docker/scripts/heif-convert.sh /usr/local/bin/heif-convert
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile
COPY --chown=root:root /docker/photoprism/entrypoint.sh /entrypoint.sh

# Create directories
RUN mkdir -m 777 -p \
    /var/lib/photoprism \
    /tmp/photoprism \
    /photoprism/originals \
    /photoprism/import \
    /photoprism/storage/config \
    /photoprism/storage/cache && \
    chown -Rf photoprism:photoprism /photoprism /var/lib/photoprism /tmp/photoprism && \
    chmod -Rf a+rwx /photoprism /var/lib/photoprism /tmp/photoprism && \
    chmod 755 /usr/local/bin/heif-convert /entrypoint.sh

# Show photoprism version
RUN photoprism -v

# Expose http port
EXPOSE 2342

# Configure entrypoint
ENTRYPOINT ["/entrypoint.sh"]
VOLUME /var/lib/photoprism

# Run server
CMD ["photoprism", "start"]
