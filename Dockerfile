ARG ARCH
ARG BUILD_TAG
FROM photoprism/development${ARCH:+-$ARCH}:${BUILD_TAG:-20210520}

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .