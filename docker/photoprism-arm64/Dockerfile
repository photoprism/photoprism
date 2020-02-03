FROM ubuntu:18.04 as build

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

ARG BUILD_TAG

ENV DEBIAN_FRONTEND noninteractive

# Configure apt-get
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry
RUN echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends
RUN echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes
RUN echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin

# Install dev / build dependencies
RUN apt-get update && apt-get upgrade && \
    apt-get install \
    build-essential \
    curl \
    chrpath \
    libssl-dev \
    libxft-dev \
    libfreetype6 \
    libfreetype6-dev \
    libfontconfig1 \
    libfontconfig1-dev \
    libhdf5-serial-dev \
    libpng-dev \
    libzmq3-dev \
    pkg-config \
    software-properties-common \
    rsync \
    unzip \
    zip \
    g++ \
    gcc \
    libc6-dev \
    gpg-agent \
    apt-utils \
    make \
    wget \
    git \
    tzdata \
    gconf-service

# Install & configure TensorFlow for C
#
# Big thank you to Guy Sheffer for building this!
# https://github.com/guysoft
#
ENV LD_LIBRARY_PATH /root/.local/lib:/usr/local/lib:/usr/lib:/lib
ENV TF_CPP_MIN_LOG_LEVEL 0
RUN curl -L \
   "https://dl.photoprism.org/tensorflow/arm64/libtensorflow-arm64-1.14.0.tar.gz" | \
   tar -C "/usr" -xz
RUN ldconfig

# Install NodeJS
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash -
RUN apt-get update && \
    apt-get install nodejs && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install and configure NodeJS Package Manager (npm)
ENV NODE_ENV production
RUN npm install --unsafe-perm=true --allow-root -g npm
RUN npm config set cache ~/.cache/npm

# Install Go
ENV GOLANG_VERSION 1.13.7
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "8717de6c662ada01b7bf318f5025c046b57f8c10cd39a88268bdc171cc7e4eab *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

# Configure Go environment
ENV GOPATH /go
ENV GOBIN $GOPATH/bin
ENV PATH $GOBIN:/usr/local/go/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV GO111MODULE on
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Download TensorFlow model and test files
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism
RUN wget "https://dl.photoprism.org/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip
RUN wget "https://dl.photoprism.org/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip

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
        ca-certificates \
        tzdata \
        libheif-examples \
        darktable \
        exiftool \
        ffmpeg && \
    apt-get update && \
    apt-get upgrade && \
    apt-get dist-upgrade && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy dependencies
COPY --from=build /usr/lib/libtensorflow.so /usr/lib/libtensorflow.so
COPY --from=build /usr/lib/libtensorflow_framework.so /usr/lib/libtensorflow_framework.so

RUN ldconfig

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
