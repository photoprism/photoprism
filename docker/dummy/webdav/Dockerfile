##################################################### BUILD STAGE ######################################################
FROM golang:alpine AS build

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

ARG GOPROXY
ARG GODEBUG

# Set up project directory
WORKDIR "/webdav"

RUN go install github.com/hacdias/webdav@latest

ENV WEBDAV_USERNAME admin
ENV WEBDAV_PASSWORD photoprism

# Expose HTTP port
EXPOSE 80

COPY /docker/dummy/webdav/config.yml /webdav/config.yml
COPY /docker/dummy/webdav/files /webdav/files

CMD ["/go/bin/webdav", "-c", "/webdav/config.yml"]
