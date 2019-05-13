FROM photoprism/photoprism:latest as build

# Hide TensorFlow warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

RUN mkdir -p /srv/photoprism/photos/import && \
    wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C /srv/photoprism/photos/import

# Import example photos
RUN photoprism import

# Start PhotoPrism server
CMD photoprism start
