FROM photoprism/photoprism:latest as build

# Hide TensorFlow warnings
ENV TF_CPP_MIN_LOG_LEVEL 2

# Copy assets
COPY /docker/demo/index.tmpl /photoprism/assets/templates

# Download example photos
RUN wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C /photoprism/originals

# Configure PhotoPrism
ENV PHOTOPRISM_STORAGE_PATH /photoprism/storage
ENV PHOTOPRISM_DEBUG false
ENV PHOTOPRISM_READONLY false
ENV PHOTOPRISM_PUBLIC true
ENV PHOTOPRISM_EXPERIMENTAL true
ENV PHOTOPRISM_UPLOAD_NSFW false
ENV PHOTOPRISM_DETECT_NSFW false
ENV PHOTOPRISM_SIDECAR_JSON true
ENV PHOTOPRISM_SIDECAR_YAML false
ENV PHOTOPRISM_GEOCODING_API places
ENV PHOTOPRISM_THUMB_FILTER lanczos
ENV PHOTOPRISM_THUMB_UNCACHED true
ENV PHOTOPRISM_THUMB_SIZE 4096
ENV PHOTOPRISM_THUMB_SIZE_UNCACHED 4096
ENV PHOTOPRISM_JPEG_SIZE 4096
ENV PHOTOPRISM_JPEG_QUALITY 95
ENV PHOTOPRISM_SITE_CAPTION "Try our demo"

# Import example photos
RUN photoprism index
RUN photoprism moments

# Start PhotoPrism server
CMD photoprism --public start
