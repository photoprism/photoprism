##################################################### BUILD STAGE ######################################################
FROM photoprism/develop:buster as build

ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILD_TAG
ARG GOPROXY
ARG GODEBUG

# set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# build and install dist files for prod env
RUN make all install DESTDIR=/opt/photoprism

################################################## PRODUCTION STAGE ####################################################
#### Base Image: Debian 10, Codename "Buster"
FROM debian:buster-slim

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

ARG TARGETARCH
ARG TARGETPLATFORM

# set environment variables, see https://docs.photoprism.app/getting-started/config-options/
ENV PATH="/opt/photoprism/bin:/opt/photoprism/scripts:/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin" \
    TMPDIR="/tmp" \
    DEBIAN_FRONTEND="noninteractive" \
    TF_CPP_MIN_LOG_LEVEL="2" \
    PHOTOPRISM_IMPORT_PATH="/photoprism/import" \
    PHOTOPRISM_ORIGINALS_PATH="/photoprism/originals" \
    PHOTOPRISM_ASSETS_PATH="/opt/photoprism/assets" \
    PHOTOPRISM_STORAGE_PATH="/photoprism/storage" \
    PHOTOPRISM_BACKUP_PATH="/photoprism/storage/backups" \
    PHOTOPRISM_LOG_FILENAME="/photoprism/storage/photoprism.log" \
    PHOTOPRISM_PID_FILENAME="/photoprism/storage/photoprism.pid" \
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

# copy app dist files and debian backports sources list
COPY --from=build /opt/photoprism/ /opt/photoprism
COPY --chown=root:root --chmod=644 /docker/develop/buster/backports.list /etc/apt/sources.list.d/backports.list

# install additional distribution packages
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing && \
    cp /opt/photoprism/bin/gosu /bin/gosu && \
    chown root:root /bin/gosu && \
    useradd -m -U -u 1000 -d /photoprism photoprism && \
    apt-get update && apt-get -qq dist-upgrade && apt-get -qq install --no-install-recommends \
      ca-certificates \
      gpg \
      wget \
      curl \
      make \
      mariadb-client \
      sqlite3 \
      tzdata \
      libc6 \
      libatomic1 \
      libheif-examples \
      librsvg2-bin \
      exiftool \
      rawtherapee \
      ffmpeg \
      ffmpegthumbnailer \
      libavcodec-extra && \
    install-darktable.sh && \
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
      /photoprism/storage/cache && \
    cleanup.sh

# define default directory and user
WORKDIR /photoprism

# expose default http port 2342
EXPOSE 2342

# define container entrypoint script
ENTRYPOINT ["/opt/photoprism/scripts/entrypoint.sh"]

# start app server
CMD ["/opt/photoprism/bin/photoprism", "start"]
