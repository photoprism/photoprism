##################################################### BUILD STAGE ######################################################
FROM ubuntu:21.10 as build

LABEL maintainer="Michael Mayer <hello@photoprism.org>"

ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILD_TAG

# Set environment variables
ENV DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp" \
    LD_LIBRARY_PATH="/root/.local/lib:/usr/local/lib:/usr/lib:/lib" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    NODE_ENV="production" \
    GOPATH="/go" \
    GOBIN="/go/bin" \
    PATH="/go/bin:/usr/local/go/bin:~/.local/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr"

# Configure apt
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing

# Copy scripts to /root/.local/bin
COPY --chown=root:root --chmod=755 /docker/scripts/*.sh /root/.local/bin/

# 1. Install distribution packages
# 2. Install TensorFlow for C
# 3. Install Chrome, NodeJS, NPM, Puppeteer, TestCafe & ChromeDriver
# 4. Install Go
RUN apt-get update && apt-get -qq dist-upgrade && apt dist-upgrade 2>/dev/null && apt-get -qq install --no-install-recommends \
    build-essential \
    ca-certificates \
    wget \
    curl \
    davfs2 \
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
    libx264-dev \
    libx265-dev \
    libnss3 \
    libxtst6 \
    librsvg2-bin \
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
    nano \
    git \
    gettext \
    sqlite3 \
    tzdata \
    gconf-service \
    libheif-examples \
    exiftool \
    ffmpeg \
    ffmpegthumbnailer \
    libavcodec-extra \
    sudo && \
    /root/.local/bin/install-tensorflow.sh ${TARGETARCH} && \
    /root/.local/bin/install-devtools.sh ${TARGETARCH} && \
    /root/.local/bin/install-go.sh ${TARGETARCH} && \
    mkdir -p "/go/src" "/go/bin" && \
    chmod -R 777 "/go"

# Download TensorFlow models & example files for testing
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism && \
    wget "https://dl.photoprism.app/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.app/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.app/tensorflow/facenet.zip?${BUILD_TAG}" -O /tmp/photoprism/facenet.zip && \
    wget "https://dl.photoprism.app/qa/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# Copy additional scripts to image
COPY --chown=root:root /docker/scripts/heif-convert.sh /usr/local/bin/heif-convert
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh

# Install Go tools
RUN /usr/local/go/bin/go install github.com/tianon/gosu@latest && \
    /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@latest && \
    /usr/local/go/bin/go install github.com/kyoh86/richgo@latest && \
    echo "alias go=richgo" > /root/.bash_aliases && \
    cp /go/bin/gosu /bin/gosu

# Create photoprism user and directory for deployment
RUN useradd -m -U -u 1000 -d /photoprism photoprism && chmod a+rwx /photoprism && \
    mkdir -m 777 -p /var/lib/photoprism /tmp/photoprism && \
    echo "alias go=richgo" > /photoprism/.bash_aliases && \
    echo "photoprism ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers && \
    chown -Rf photoprism:photoprism /photoprism /var/lib/photoprism /tmp/photoprism && \
    chmod -Rf a+rw /photoprism /var/lib/photoprism /tmp/photoprism /go && \
    chmod 755 /usr/local/bin/heif-convert /entrypoint.sh && \
    find /go -type d -print0 | xargs -0 chmod 777

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Expose HTTP port 2342, 2343 plus 9515 for chromedriver
EXPOSE 2342 2343 9515
VOLUME /var/lib/photoprism

# Configure entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# Run server
CMD ["tail", "-f", "/dev/null"]