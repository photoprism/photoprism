# Build stage 
FROM golang:1.17 AS build-env
RUN go install github.com/go-delve/delve/cmd/dlv@latest \
  && cp $GOPATH/bin/dlv /usr/local/bin/

# Ubuntu 23.04 (Lunar Lobster)
FROM photoprism/develop:230715-lunar

## Alternative Environments:
# FROM photoprism/develop:jammy    # Ubuntu 22.04 LTS (Jammy Jellyfish)
# FROM photoprism/develop:impish   # Ubuntu 21.10 (Impish Indri)
# FROM photoprism/develop:bookworm # Debian 12 (Bookworm)
# FROM photoprism/develop:bullseye # Debian 11 (Bullseye)
# FROM photoprism/develop:buster   # Debian 10 (Buster)

# Set default working directory.
WORKDIR "/go/src/github.com/photoprism/photoprism"

# Copy source to image.
COPY . .
COPY --chown=root:root /scripts/dist/ /scripts/
COPY --from=build-env /usr/local/bin/dlv /