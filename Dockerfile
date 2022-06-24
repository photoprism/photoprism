# Debian 12, Codename 'Bookworm'
FROM photoprism/develop:220620-bookworm

## alternative base images
# FROM photoprism/develop:bullseye # Debian 11, Codename 'Bullseye'
# FROM photoprism/develop:buster   # Debian 10, Codename 'Buster'
# FROM photoprism/develop:impish   # Ubuntu 21.10, Codename 'Impish Indri'

# define working directory in container
WORKDIR "/go/src/github.com/photoprism/photoprism"

# copy project source code to container
COPY . .
COPY --chown=root:root /scripts/dist/* /scripts/