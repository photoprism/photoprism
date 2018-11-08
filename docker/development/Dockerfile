FROM ubuntu:18.04

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

RUN apt-get update && apt-get install -y --no-install-recommends \
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
        mysql-client

# Install darktable (RAW to JPEG converter)
RUN add-apt-repository ppa:pmjdebruijn/darktable-release && \
    apt-get update && \
    apt-get install -y --no-install-recommends darktable && \
    apt-get upgrade -y

# Install TensorFlow C library
RUN curl -L \
   "https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.12.0.tar.gz" | \
   tar -C "/usr/local" -xz
RUN ldconfig

# Hide some warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

# Install NPM (NodeJS)
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install -y nodejs

# Install YARN (Package Manager)
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get update && \
    apt-get install yarn && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.11.2
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "1dfe664fa3d8ad714bbd15a36627992effd150ddabd7523931f077b3926d736d *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
ENV GOBIN $GOPATH/bin
ENV PATH $GOBIN:/usr/local/go/bin:$PATH
ENV GO111MODULE on
ENV NODE_ENV production

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Download TensorFlow model and test files
RUN rm -rf /tmp/* && mkdir -p /tmp/photoprism
RUN wget "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip" -O /tmp/photoprism/inception.zip
RUN wget "https://www.dropbox.com/s/na9p9wwt98l7m5b/import.zip?dl=1" -O /tmp/photoprism/testdata.zip

# Install goimports
RUN env GO111MODULE=off /usr/local/go/bin/go get golang.org/x/tools/cmd/goimports

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Get PhotoPrism dependencies
COPY . .
RUN make dep
RUN rm -rf /go/src/github.com/photoprism/photoprism

# Expose HTTP port
EXPOSE 80

# Start PhotoPrism server
CMD tail -f /dev/null