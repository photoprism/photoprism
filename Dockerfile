FROM tensorflow/tensorflow

# Install TensorFlow C library
RUN curl -L \
   "https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.3.0.tar.gz" | \
   tar -C "/usr/local" -xz
RUN ldconfig
# Hide some warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

RUN curl -L "https://download.opensuse.org/repositories/graphics:darktable:stable/xUbuntu_16.04/Release.key" | apt-key add -
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/graphics:/darktable:/stable/xUbuntu_16.04/ /' > /etc/apt/sources.list.d/darktable.list"

# Install Go
RUN apt-get update && apt-get install -y --no-install-recommends \
    g++ \
    gcc \
    libc6-dev \
    make \
    pkg-config \
    nano \
    build-essential \
    wget \
    darktable \
    git \
	&& rm -rf /var/lib/apt/lists/*

RUN apt-get upgrade -y

# Install NPM (NodeJS)
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install -y nodejs

# Install YARN (Package Manager)
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get update && apt-get install yarn

ENV GOLANG_VERSION 1.10
RUN set -eux; \
	\
	dpkgArch="$(dpkg --print-architecture)"; \
	case "${dpkgArch##*-}" in \
		amd64) goRelArch='linux-amd64'; goRelSha256='b5a64335f1490277b585832d1f6c7f8c6c11206cba5cd3f771dcb87b98ad1a33' ;; \
		armhf) goRelArch='linux-armv6l'; goRelSha256='6ff665a9ab61240cf9f11a07e03e6819e452a618a32ea05bbb2c80182f838f4f' ;; \
		arm64) goRelArch='linux-arm64'; goRelSha256='efb47e5c0e020b180291379ab625c6ec1c2e9e9b289336bc7169e6aa1da43fd8' ;; \
		i386) goRelArch='linux-386'; goRelSha256='2d26a9f41fd80eeb445cc454c2ba6b3d0db2fc732c53d7d0427a9f605bfc55a1' ;; \
		ppc64el) goRelArch='linux-ppc64le'; goRelSha256='a1e22e2fbcb3e551e0bf59d0f8aeb4b3f2df86714f09d2acd260c6597c43beee' ;; \
		s390x) goRelArch='linux-s390x'; goRelSha256='71cde197e50afe17f097f81153edb450f880267699f22453272d184e0f4681d7' ;; \
		*) goRelArch='src'; goRelSha256='f3de49289405fda5fd1483a8fe6bd2fa5469e005fd567df64485c4fa000c7f24'; \
			echo >&2; echo >&2 "warning: current architecture ($dpkgArch) does not have a corresponding Go binary release; will be building from source"; echo >&2 ;; \
	esac; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	if [ "$goRelArch" = 'src' ]; then \
		echo >&2; \
		echo >&2 'error: UNIMPLEMENTED'; \
		echo >&2 'TODO install golang-any from jessie-backports for GOROOT_BOOTSTRAP (and uninstall after build)'; \
		echo >&2; \
		exit 1; \
	fi; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install dependencies
RUN go get github.com/tensorflow/tensorflow/tensorflow/go \
  github.com/tensorflow/tensorflow/tensorflow/go/op \
  github.com/julienschmidt/httprouter

# Download InceptionV3 model
RUN mkdir -p /model && \
  wget "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip" -O /model/inception.zip && \
  unzip /model/inception.zip -d /model && \
  chmod -R 777 /model

# Doesn't work properly at the moment (wait for stable release)
# RUN go get -u github.com/kardianos/govendor

# Using dep for the moment...
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN mkdir -m 777 /go/pkg/dep

# Create user
# RUN adduser --disabled-password --gecos '' photoprism
# USER photoprism

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

RUN cp config.example.yml ~/.photoprism

# RUN govendor sync
RUN dep ensure

# Build
# RUN govendor install +local
RUN go build cmd/photoprism/photoprism.go
