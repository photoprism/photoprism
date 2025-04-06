ARG BUILDER_SHA2=06040763c500bd2ebaaa4585d4729c88d2c8ccec94baa7fbe9bbe3dc2827d79d
FROM gcr.io/tensorflow-testing/ml-devinfra-linux-aarch64-cross-compile:infrastructure-public-image-${BUILDER_SHA2}

# see https://docs.docker.com/build/building/variables/#env-usage-example
ARG TF_VERSION=2.18.0
ENV TF_VERSION=$TF_VERSION

COPY ./create_archive.sh .
COPY ./Makefile Makefile
