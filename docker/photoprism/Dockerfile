FROM photoprism/development:20200715 as build

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build PhotoPrism
RUN make dep build-js install

# Same base image as photoprism/development
FROM ubuntu:18.04

# Set environment variables
ENV DEBIAN_FRONTEND noninteractive

# Configure apt-get
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry
RUN echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends
RUN echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes
RUN echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin

# Install additional distribution packages
RUN apt-get update && apt-get install -y --no-install-recommends \
        wget \
        ca-certificates \
        tzdata \
        libheif-examples \
        gnupg \
        gpg-agent \
        apt-utils \
        add-apt-key \
        exiftool \
        ffmpeg

# Copy dependencies
COPY --from=build /usr/lib/libtensorflow.so /usr/lib/libtensorflow.so
COPY --from=build /usr/lib/libtensorflow_framework.so /usr/lib/libtensorflow_framework.so
RUN ldconfig

# Install RAW to JPEG converter
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/graphics:/darktable/xUbuntu_18.04/ /' > /etc/apt/sources.list.d/graphics:darktable.list" && \
    wget -qO - https://download.opensuse.org/repositories/graphics:darktable/xUbuntu_18.04/Release.key | apt-key add - && \
    apt-get update && \
    apt-get install darktable && \
    apt-get upgrade && \
    apt-get dist-upgrade && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /photoprism

ENV TF_CPP_MIN_LOG_LEVEL 2

ENV PATH /photoprism/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

ENV PHOTOPRISM_ASSETS_PATH /photoprism/assets
ENV PHOTOPRISM_ORIGINALS_PATH /photoprism/originals
ENV PHOTOPRISM_IMPORT_PATH /photoprism/import
ENV PHOTOPRISM_LOG_FILENAME /photoprism/photoprism.log
ENV PHOTOPRISM_PID_FILENAME /photoprism/photoprism.pid

# Copy files to /photoprism
COPY --from=build /root/.local/bin/photoprism /photoprism/bin/photoprism
COPY --from=build /root/.photoprism/assets /photoprism/assets

# Create directories
RUN mkdir -p \
    /photoprism/originals \
    /photoprism/import \
    /photoprism/storage/settings \
    /photoprism/storage/cache

RUN chmod -R 777 /photoprism

# Show photoprism version
RUN photoprism -v

# Expose http port
EXPOSE 2342

# Run server
CMD photoprism start
