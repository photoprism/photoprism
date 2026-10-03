FROM golang:1

LABEL maintainer="PhotoPrism UG <hello@photoprism.app>"

# Set up project directory
WORKDIR "/webdav"

# Pinned so that rebuilding this fixture produces the server the acceptance tests were written
# against. This is the version the unversioned module path resolves to; the maintained line is
# the /v5 module, whose configuration format differs and needs its own compatibility pass.
RUN go install github.com/hacdias/webdav@v1.4.1

ENV DO_NOT_TRACK=true \
    WEBDAV_USERNAME="admin" \
    WEBDAV_PASSWORD="photoprism"

# Expose HTTP port
EXPOSE 80

COPY /docker/dummy/webdav/config.yml /webdav/config.yml
COPY /docker/dummy/webdav/files /webdav/files

CMD ["webdav", "-c", "/webdav/config.yml"]
