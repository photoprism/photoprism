FROM photoprism/photoprism:latest as build

RUN mkdir -p /srv/photoprism/photos/import && \
    wget -qO- https://dl.photoprism.org/fixtures/demo.tar.gz | tar xvz -C /srv/photoprism/photos/import

RUN photoprism import

# Start PhotoPrism server
CMD photoprism start