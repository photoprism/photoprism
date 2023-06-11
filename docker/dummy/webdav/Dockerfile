FROM golang:1

LABEL maintainer="PhotoPrism UG <hello@photoprism.app>"

# Set up project directory
WORKDIR "/webdav"

RUN go install github.com/hacdias/webdav@latest

ENV WEBDAV_USERNAME admin
ENV WEBDAV_PASSWORD photoprism

# Expose HTTP port
EXPOSE 80

COPY /docker/dummy/webdav/config.yml /webdav/config.yml
COPY /docker/dummy/webdav/files /webdav/files

CMD webdav -c /webdav/config.yml
