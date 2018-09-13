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
        python \
        python-dev \
        rsync \
        software-properties-common \
        unzip \
        g++ \
        gcc \
        libc6-dev \
        gpg-agent \
        apt-utils \
        make \
        nano \
        wget \
        darktable \
        git \
        python3 \
        python-setuptools \
        python3-dev \
        mysql-client \
        && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN add-apt-repository ppa:pmjdebruijn/darktable-release

RUN apt-get update && apt-get install -y --no-install-recommends \
        darktable \
        && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN apt-get upgrade -y

RUN curl -O https://bootstrap.pypa.io/get-pip.py && \
    python get-pip.py && \
    rm get-pip.py

# Install TensorFlow CPU version from central repo
RUN pip --no-cache-dir install \
    http://storage.googleapis.com/tensorflow/linux/cpu/tensorflow-1.10.1-cp27-none-linux_x86_64.whl

RUN pip --no-cache-dir install --upgrade \
        requests \
        Pillow \
        h5py \
        ipykernel \
        jupyter \
        keras_applications \
        keras_preprocessing \
        matplotlib \
        numpy \
        pandas \
        scipy \
        sklearn \
        && \
    python -m ipykernel.kernelspec

# Install TensorFlow C library
RUN curl -L \
   "https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.10.1.tar.gz" | \
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
RUN apt-get update && apt-get install yarn

ENV GOLANG_VERSION 1.11
RUN set -eux; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "b3fcf280ff86558e0559e185b601c9eade0fd24c900b4c63cd14d1d38613e499 *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
ENV GOBIN $GOPATH/bin
ENV PATH $GOBIN:/usr/local/go/bin:$PATH
ENV GO111MODULE on
ENV NODE_ENV production

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" /etc/photoprism /var/photos && chmod -R 777 "$GOPATH"

# Download InceptionV3 model
RUN mkdir -p /model && \
  wget "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip" -O /model/inception.zip && \
  unzip /model/inception.zip -d /model && \
  chmod -R 777 /model

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

RUN cp config.prod.yml /etc/photoprism/config.yml

# Build PhotoPrism
RUN make dep js install

RUN cp -r server/assets /etc/photoprism

# Expose HTTP port
EXPOSE 80

# Start PhotoPrism server
CMD photoprism start