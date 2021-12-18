##################################################### BUILD STAGE ######################################################
FROM golang:alpine AS build

LABEL maintainer="Michael Mayer <hello@photoprism.app>"

ARG GOPROXY
ARG GODEBUG

# Move to working directory /app
WORKDIR /app

# Copy files and download dependency using go mod
COPY /docker/dummy/oidc/app/. .
RUN go mod download

# Build the application
RUN go build -o server .

# Allow HTTP scheme
ENV CAOS_OIDC_DEV=true

# Expose HTTP port
EXPOSE 9998

CMD ["/app/server"]