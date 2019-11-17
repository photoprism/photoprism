FROM photoprism/development:20191105 as build

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build PhotoPrism
RUN make dep build-js install

# Base base image as photoprism/development
FROM ubuntu:18.04

# Set environment variables
ENV DEBIAN_FRONTEND noninteractive

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

# Copy dependencies
COPY --from=build /etc/apt/sources.list.d/pmjdebruijn-ubuntu-darktable-release-bionic.list /etc/apt/sources.list.d/pmjdebruijn-ubuntu-darktable-release-bionic.list
COPY --from=build /etc/apt/trusted.gpg.d/pmjdebruijn_ubuntu_darktable-release.gpg /etc/apt/trusted.gpg.d/pmjdebruijn_ubuntu_darktable-release.gpg
COPY --from=build /usr/lib/libtensorflow.so /usr/lib/libtensorflow.so
COPY --from=build /usr/lib/libtensorflow_framework.so /usr/lib/libtensorflow_framework.so
RUN ldconfig

# Install darktable (RAW to JPEG converter)
RUN apt-get update && \
    apt-get install -y --no-install-recommends darktable && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create and configure photoprism user
RUN useradd -ms /bin/bash photoprism
USER photoprism

WORKDIR /home/photoprism
ENV PATH /home/photoprism/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV TF_CPP_MIN_LOG_LEVEL 2
ENV PHOTOPRISM_CONFIG_FILE /home/photoprism/.config/photoprism/photoprism.yml

# Copy assets to /home/photoprism
COPY --chown=photoprism --from=build /root/.local/bin/photoprism /home/photoprism/.local/bin/photoprism
COPY --chown=photoprism --from=build /root/.local/share/photoprism /home/photoprism/.local/share/photoprism
COPY --chown=photoprism --from=build /root/.config/photoprism /home/photoprism/.config/photoprism

# Create directories
RUN mkdir -p /home/photoprism/.cache/photoprism \
    /home/photoprism/Pictures/Originals \
    /home/photoprism/Pictures/Import \
    /home/photoprism/Pictures/Export

# Fix permissions
USER root
RUN chown -R photoprism:photoprism /home/photoprism
USER photoprism

# Show photoprism version
RUN photoprism -v

# Expose http and database ports
EXPOSE 2342 4000

# Run server
CMD photoprism start
