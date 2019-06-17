FROM ubuntu:18.04

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
    nano \
    wget \
    git \
    mysql-client \
    libgtk-3-bin \
    tzdata \
    gconf-service \
    chromium-browser \
    firefox \
    libheif-examples \
    exiftool

# Install RAW to JPEG converter
RUN add-apt-repository ppa:pmjdebruijn/darktable-release && \
    apt-get update && \
    apt-get install darktable && \
    apt-get upgrade && \
    apt-get dist-upgrade

# Install TensorFlow C library
RUN curl -L \
   "https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.13.1.tar.gz" | \
   tar -C "/usr/local" -xz
RUN ldconfig

# Show TensorFlow debug log
ENV TF_CPP_MIN_LOG_LEVEL 0

# Install NodeJS
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get update && \
    apt-get install nodejs && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install and configure NodeJS Package Manager (npm)
ENV NODE_ENV production
RUN npm install --unsafe-perm=true --allow-root -g npm testcafe chromedriver
RUN npm config set cache ~/.cache/npm

# Install Go
ENV GOLANG_VERSION 1.12.6
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "dbcf71a3c1ea53b8d54ef1b48c85a39a6c9a935d01fc8291ff2b92028e59913c *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

# Configure Go environment
ENV GOPATH /go
ENV GOBIN $GOPATH/bin
ENV PATH $GOBIN:/usr/local/go/bin:$PATH
ENV GO111MODULE on
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Download TensorFlow model and test files
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism
RUN wget "https://dl.photoprism.org/tensorflow/nasnet.zip?${BUILD_TAG}" -O /tmp/photoprism/nasnet.zip
RUN wget "https://dl.photoprism.org/fixtures/testdata.zip?${BUILD_TAG}" -O /tmp/photoprism/testdata.zip

# Install goimports and richgo (colorizes "go test" output)
RUN env GO111MODULE=off /usr/local/go/bin/go get -u golang.org/x/tools/cmd/goimports
RUN env GO111MODULE=off /usr/local/go/bin/go get -u github.com/kyoh86/richgo
RUN echo "alias go=richgo" > /root/.bash_aliases

# Configure broadwayd (HTML5 display server)
# Command: broadwayd -p 8080 -a 0.0.0.0 :5
ENV GDK_BACKEND broadway
ENV BROADWAY_DISPLAY :5

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Expose HTTP port plus 4000 for TiDB, 8080 for broadwayd and 9515 for chromedriver
EXPOSE 80 2342 4000 8080 9515

# Keep container running (services can be started manually using a terminal)
CMD tail -f /dev/null
