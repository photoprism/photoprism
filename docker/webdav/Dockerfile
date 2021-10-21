FROM golang:1

LABEL maintainer="Michael Mayer <hello@photoprism.org>"

# Set up project directory
WORKDIR "/webdav"

RUN go install github.com/hacdias/webdav@latest

# Expose HTTP port
EXPOSE 80

COPY /docker/webdav/config.yml /webdav/config.yml
COPY /docker/webdav/files /webdav/files

CMD webdav -c /webdav/config.yml
