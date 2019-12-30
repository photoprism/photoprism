FROM photoprism/photoprism:latest as build

# Hide TensorFlow warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

RUN wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C Pictures/Import

# Configure PhotoPrism
ENV PHOTOPRISM_DEBUG false
ENV PHOTOPRISM_READONLY false
ENV PHOTOPRISM_PUBLIC true
ENV PHOTOPRISM_EXPERIMENTAL true
ENV PHOTOPRISM_UPLOAD_NSFW false
ENV PHOTOPRISM_HIDE_NSFW false

# Import example photos
RUN photoprism import

# Start PhotoPrism server
CMD photoprism --public start
