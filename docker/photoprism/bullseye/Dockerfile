##################################################### BUILD STAGE ######################################################
FROM photoprism/develop:bullseye as build

ARG TARGETARCH
ARG BUILD_TAG

# set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# build and install dist files for prod env
RUN make all install DESTDIR=/opt/photoprism

################################################## PRODUCTION STAGE ####################################################
#### Base Image: Debian 11, Codename "Bullseye"
FROM photoprism/develop:bullseye-slim

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

ARG TARGETARCH
ARG BUILD_TAG

# set environment variables, see https://docs.photoprism.app/getting-started/config-options/
ENV DOCKER_ARCH=$TARGETARCH \
    DOCKER_TAG=$BUILD_TAG \
    DOCKER_ENV="prod" \
    PATH="/opt/photoprism/bin:/opt/photoprism/scripts:/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin" \
    TMPDIR="/tmp" \
    DEBIAN_FRONTEND="noninteractive" \
    TF_CPP_MIN_LOG_LEVEL="2" \
    PHOTOPRISM_ASSETS_PATH="/opt/photoprism/assets" \
    PHOTOPRISM_IMPORT_PATH="/photoprism/import" \
    PHOTOPRISM_ORIGINALS_PATH="/photoprism/originals" \
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
    PHOTOPRISM_SITE_CAPTION="AI-Powered Photos App" \
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

# update pre-installed packages if needed
RUN apt-get update && apt-get -qq dist-upgrade && cleanup.sh

# define default directory and user
WORKDIR /photoprism

# expose default http port 2342
EXPOSE 2342

# copy app dist files and debian backports sources list
COPY --from=build /opt/photoprism/ /opt/photoprism
RUN cp /opt/photoprism/scripts/entrypoint.sh /entrypoint.sh

# set container entrypoint script
ENTRYPOINT ["/entrypoint.sh"]

# start app server
CMD ["/opt/photoprism/bin/photoprism", "start"]
