FROM ubuntu:21.04

LABEL maintainer="Michael Mayer <hello@photoprism.org>"

ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILD_TAG

# Set environment variables
ENV DEBIAN_FRONTEND="noninteractive" \
    TMPDIR="/tmp"

# Configure apt
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
    echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
    echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
    echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
    echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing

# Copy scripts to /root/.local/bin
COPY --chown=root:root --chmod=755 /docker/scripts/*.sh /root/.local/bin/

# Install dev / build dependencies
RUN apt-get update && apt-get -qq dist-upgrade && \
    apt-get -qq install -y --no-install-recommends \
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
    mariadb-client \
    sqlite3 \
    tzdata \
    gconf-service \
    firefox \
    libheif-examples \
    exiftool \
    rawtherapee \
    ffmpeg \
    ffmpegthumbnailer \
    libavcodec-extra \
    lsof \
    apache2-utils \
    fonts-roboto \
    sudo && \
    [ "$TARGETARCH" = "arm" ] || apt-get -qq install darktable; \
    apt-get -y autoremove && apt-get -y autoclean && apt-get clean

# Set environment variables
ENV LD_LIBRARY_PATH="/root/.local/lib:/usr/local/lib:/usr/lib:/lib" \
    TF_CPP_MIN_LOG_LEVEL="0" \
    NODE_ENV="production" \
    GOPATH="/go" \
    GOBIN="/go/bin" \
    PATH="/go/bin:/usr/local/go/bin:~/.local/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
    GO111MODULE="on" \
    CGO_CFLAGS="-g -O2 -Wno-return-local-addr"

# Install & configure TensorFlow for C,
# see https://www.tensorflow.org/install/lang_c
#
# We also have custom builds available:
# https://dl.photoprism.org/tensorflow/linux/
#
RUN /root/.local/bin/install-tensorflow.sh $TARGETARCH

# Install NodeJS
RUN curl -sL https://deb.nodesource.com/setup_14.x | bash -
RUN add-apt-repository -y ppa:saiarcot895/chromium-beta && \
    apt-get update && \
    apt-get -qq dist-upgrade && \
    apt-get -qq install chromium-browser chromium-codecs-ffmpeg-extra nodejs && \
    apt-get -y autoremove && \
    apt-get -y autoclean && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install NodeJS & NPM
RUN umask 0000 && npm install --unsafe-perm=true --allow-root -g npm && npm config set cache ~/.cache/npm && \
    [ "$TARGETARCH" = "arm" ] || npm install --unsafe-perm=true --allow-root -g testcafe chromedriver;


# Install Go
RUN /root/.local/bin/install-go.sh ${TARGETARCH}

# Download TensorFlow model and test files
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism && \
    wget "https://dl.photoprism.org/tensorflow/nsfw.zip?${BUILD_TAG}" -O /tmp/photoprism/nsfw.zip && \
    wget "https://dl.photoprism.org/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip && \
    wget "https://dl.photoprism.org/tensorflow/facenet.zip?${BUILD_TAG}" -O /tmp/photoprism/facenet.zip && \
    wget "https://dl.photoprism.org/qa/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# Install additional tools
COPY --chown=root:root --chmod=755 /docker/scripts/heif-convert.sh /usr/local/bin/heif-convert
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/tianon/gosu \
    golang.org/x/tools/cmd/goimports github.com/kyoh86/richgo && \
    [ "$TARGETARCH" = "arm" ] || \
    env GO111MODULE=off /usr/local/go/bin/go get -u github.com/psampaz/go-mod-outdated \
    github.com/tsliwowicz/go-wrk github.com/dsoprea/go-exif/v3/command/exif-read-tool; \
    echo "alias go=richgo" > /root/.bash_aliases && \
    cp /go/bin/gosu /bin/gosu

# Create photoprism user and directory for deployment
RUN useradd photoprism -m -d /photoprism && chmod a+rwx /photoprism && \
    mkdir -m 777 -p /var/lib/photoprism /tmp/photoprism && \
    echo "alias go=richgo" > /photoprism/.bash_aliases && \
    echo "photoprism ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers && \
    chown -Rf photoprism:photoprism /photoprism /var/lib/photoprism /tmp/photoprism && \
    chmod -Rf a+rw /photoprism /var/lib/photoprism /tmp/photoprism /go && \
    find /go -type d -print0 | xargs -0 chmod 777

# Copy mysql client config for development
COPY --chown=root:root /docker/development/.my.cnf /root/.my.cnf
COPY --chown=photoprism:photoprism /docker/development/.my.cnf /photoprism/.my.cnf
RUN chmod 644 /root/.my.cnf /photoprism/.my.cnf

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Expose HTTP port 2342, 2343 plus 9515 for chromedriver
EXPOSE 2342 2343 9515
VOLUME /var/lib/photoprism

# Configure entrypoint
COPY --chown=root:root --chmod=755 /docker/development/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

# Run server
CMD ["tail", "-f", "/dev/null"]