FROM photoprism/development:20190617 as build

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build PhotoPrism
RUN make all install

# Base base image as photoprism/development
FROM ubuntu:18.04

# Set environment variables
ENV DEBIAN_FRONTEND noninteractive
ENV TF_CPP_MIN_LOG_LEVEL 2
ENV PHOTOPRISM_CONFIG_FILE /srv/photoprism/config/photoprism.yml

WORKDIR /srv/photoprism

# Install additional distribution packages
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        unzip \
        nano \
        wget \
        ca-certificates \
        tzdata \
        libheif-examples \
        exiftool && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy built binaries and assets to this image
COPY --from=build /etc/apt/sources.list.d/pmjdebruijn-ubuntu-darktable-release-bionic.list /etc/apt/sources.list.d/pmjdebruijn-ubuntu-darktable-release-bionic.list
COPY --from=build /etc/apt/trusted.gpg.d/pmjdebruijn_ubuntu_darktable-release.gpg /etc/apt/trusted.gpg.d/pmjdebruijn_ubuntu_darktable-release.gpg
COPY --from=build /usr/local/bin/photoprism /usr/local/bin/photoprism
COPY --from=build /usr/local/lib/libtensorflow.so /usr/local/lib/libtensorflow.so
COPY --from=build /usr/local/lib/libtensorflow_framework.so /usr/local/lib/libtensorflow_framework.so
COPY --from=build /srv/photoprism /srv/photoprism

# Configure dynamic linker run-time bindings
RUN ldconfig

# Install darktable (RAW to JPEG converter)
RUN apt-get update && \
    apt-get install -y --no-install-recommends darktable && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Show photoprism version
RUN photoprism -v

# Expose HTTP & TiDB port
EXPOSE 80
EXPOSE 4000

# Start PhotoPrism server
CMD photoprism start
