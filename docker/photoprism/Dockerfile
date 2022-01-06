##################################################### BUILD STAGE ######################################################
FROM photoprism/develop:20220106 as build

ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILD_TAG
ARG GOPROXY
ARG GODEBUG

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build frontend and backend
RUN make dep build-js install

################################################## PRODUCTION STAGE ####################################################
FROM ubuntu:21.10

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

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
    libavcodec-extra  && \
    [ "$TARGETARCH" = "arm" ] || apt-get install darktable; \
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
