FROM ubuntu:18.04

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

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
    libfreetype6-dev \
    libhdf5-serial-dev \
    libpng-dev \
    libzmq3-dev \
    pkg-config \
    software-properties-common \
    rsync \
    unzip \
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
    tzdata

# Install darktable (RAW to JPEG converter)
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

# Hide some warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

# Install NPM (NodeJS)
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install nodejs

# Install YARN (Package Manager)
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get update && \
    apt-get install yarn && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Go
ENV GOLANG_VERSION 1.12.4
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "d7d1f1f88ddfe55840712dc1747f37a790cbcaa448f6c9cf51bbe10aa65442f5 *go.tgz" | sha256sum -c -; \
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
RUN wget "https://dl.photoprism.org/tensorflow/nasnet-20190501.zip" -O /tmp/photoprism/nasnet.zip
RUN wget "https://dl.photoprism.org/fixtures/testdata-20190501.zip" -O /tmp/photoprism/testdata.zip

# Install goimports
RUN env GO111MODULE=off /usr/local/go/bin/go get golang.org/x/tools/cmd/goimports

# Configure JS environment for building
ENV NODE_ENV production
ENV YARN_CACHE_FOLDER /root/.cache/yarn

# Configure broadwayd (HTML5 display server)
# Command: broadwayd -p 8080 -a 0.0.0.0 :5
ENV GDK_BACKEND broadway
ENV BROADWAY_DISPLAY :5

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Expose HTTP port plus 4000 for TiDB and 8080 for broadwayd (not running by default)
EXPOSE 80
EXPOSE 4000
EXPOSE 8080

# Keep container running (services can be started manually using a terminal)
CMD tail -f /dev/null
