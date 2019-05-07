FROM photoprism/photoprism:latest as build

RUN mkdir -p /srv/photoprism/photos/import && \
    wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C /srv/photoprism/photos/import

RUN photoprism import

# Hide TensorFlow warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

# Start PhotoPrism server
CMD photoprism start
