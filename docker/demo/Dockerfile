FROM photoprism/photoprism:latest as build

# Hide TensorFlow warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

RUN mkdir -p /srv/photoprism/photos/import && \
    wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C /srv/photoprism/photos/import

# Import example photos
RUN photoprism import

# Thumbnail cache warmup
RUN photoprism thumbnails -q -s 50 -s 100 -s 320 -s 500
RUN photoprism thumbnails -s 720 -s 1280 -s 1920 -s 2560 -s 3840

# Start PhotoPrism server
CMD photoprism start
