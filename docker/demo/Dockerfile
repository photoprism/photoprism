FROM photoprism/photoprism:latest as build

# Hide TensorFlow warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

# Copy assets
COPY /docker/demo/index.tmpl /photoprism/assets/resources/templates

# Download example photos
RUN wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C /photoprism/import

# Configure PhotoPrism
ENV PHOTOPRISM_DEBUG false
ENV PHOTOPRISM_READONLY false
ENV PHOTOPRISM_PUBLIC true
ENV PHOTOPRISM_EXPERIMENTAL true
ENV PHOTOPRISM_UPLOAD_NSFW false
ENV PHOTOPRISM_DETECT_NSFW true
ENV PHOTOPRISM_JPEG_QUALITY 95
ENV PHOTOPRISM_THUMB_FILTER lanczos
ENV PHOTOPRISM_THUMB_UNCACHED true
ENV PHOTOPRISM_THUMB_SIZE 3840
ENV PHOTOPRISM_THUMB_LIMIT 3840
ENV PHOTOPRISM_GEOCODING_API places
ENV PHOTOPRISM_SIDECAR_JSON true

# Import example photos
RUN photoprism import

# Start PhotoPrism server
CMD photoprism --public start
