FROM ubuntu:18.04

LABEL maintainer="Michael Mayer <michael@liquidbytes.net>"

WORKDIR "/src"

ENV DARKTABLE_VERSION 2.5.0

ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV LC_MESSAGES C.UTF-8
ENV LANGUAGE C.UTF-8
ENV GCC_VER=8
ENV LLVM_VER=7
ENV DEBIAN_FRONTEND noninteractive

# Paper over occasional network flakiness of some mirrors.
RUN echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry

# Do not install recommended packages
RUN echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends

# Do not install suggested packages
RUN echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests

# Assume yes
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes

# Fix broken packages
RUN echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissin

# Install general build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        gpg-agent \
        apt-utils \
        gpg	\
        gpgconf	\
        gpgv \
        pkg-config \
        nano \
        wget

# Add llvm repo
RUN echo "deb http://apt.llvm.org/bionic/ llvm-toolchain-bionic-7 main" | tee /etc/apt/sources.list.d/llvm.list
COPY /docker/darktable/llvm.gpg.key /tmp/llvm.gpg.key
RUN apt-key add /tmp/llvm.gpg.key

# Install darktable build depenencies
RUN apt-get update && apt-get install appstream-util clang-$LLVM_VER cmake desktop-file-utils \
    g++-$GCC_VER gcc-$GCC_VER gettext git intltool libatk1.0-dev libcairo2-dev \
    libcolord-dev libcolord-gtk-dev libcups2-dev libcurl4-gnutls-dev \
    libexiv2-dev libflickcurl-dev libgdk-pixbuf2.0-dev libglib2.0-dev \
    libgphoto2-dev libgraphicsmagick1-dev libgtk-3-dev libjpeg-dev \
    libjson-glib-dev liblcms2-dev liblensfun-dev liblua5.2-dev liblua5.3-dev \
    libopenexr-dev libopenjp2-7-dev libosmgpsmap-1.0-dev libpango1.0-dev \
    libpng-dev libpugixml-dev librsvg2-dev libsaxon-java libsecret-1-dev \
    libsoup2.4-dev libsqlite3-dev libtiff5-dev libwebp-dev libx11-dev \
    libxml2-dev libxml2-utils make ninja-build perl po4a python3-jsonschema \
    xsltproc zlib1g-dev libxslt1-dev gtk+-3.0 libsoup2.4 libtiff-dev libgtk-3-bin && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN dpkg-divert --add --rename --divert /usr/bin/ld.original /usr/bin/ld && \
    ln -s /usr/bin/ld.gold /usr/bin/ld

RUN rm -rf /var/lib/apt/lists/* && apt-get update && \
    apt-get install clang-$LLVM_VER libclang-common-$LLVM_VER-dev \
    llvm-$LLVM_VER-dev && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN rm -rf /var/lib/apt/lists/* && apt-get update && \
    apt-get install default-jdk-headless default-jre-headless docbook \
    docbook-xml docbook-xsl docbook-xsl-saxon fop gnome-doc-utils imagemagick \
    libsaxon-java xsltproc && apt-get clean && rm -rf /var/lib/apt/lists/*

# Build and install darktable
RUN git clone https://github.com/darktable-org/darktable.git
WORKDIR "/src/darktable"
RUN git fetch --all --tags --prune && git checkout tags/release-$DARKTABLE_VERSION
RUN git submodule init && git submodule update
RUN ./build.sh --prefix /opt/darktable --build-type Release --install

# Copy darktable header files
WORKDIR "/src/darktable/src"
RUN mkdir -p /opt/darktable/include/darktable
RUN find . -name '*.h' -exec cp --parents \{\} /opt/darktable/include/darktable \;

# Configure shell environment
WORKDIR "/opt/darktable"
ENV PATH /opt/darktable/bin:$PATH

# Configure broadwayd
ENV GDK_BACKEND broadway
ENV BROADWAY_DISPLAY :5
EXPOSE 8080
CMD broadwayd -p 8080 -a 0.0.0.0 :5