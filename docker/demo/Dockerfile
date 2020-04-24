FROM photoprism/photoprism:latest as build

# Hide TensorFlow warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

RUN wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C /photoprism/import

# Configure PhotoPrism
ENV PHOTOPRISM_DEBUG false
ENV PHOTOPRISM_READONLY false
ENV PHOTOPRISM_PUBLIC true
ENV PHOTOPRISM_EXPERIMENTAL true
ENV PHOTOPRISM_UPLOAD_NSFW false
ENV PHOTOPRISM_DETECT_NSFW false
ENV PHOTOPRISM_THUMB_QUALITY 95
ENV PHOTOPRISM_THUMB_SIZE 3840
ENV PHOTOPRISM_THUMB_LIMIT 3840
ENV PHOTOPRISM_THUMB_FILTER lanczos
ENV PHOTOPRISM_GEOCODING_API places

# Import example photos
RUN photoprism import

# Start PhotoPrism server
CMD photoprism --public start
