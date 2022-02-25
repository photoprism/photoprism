FROM photoprism/develop:220225-bullseye

## alternative base images
# FROM photoprism/develop:buster   # Debian 10, Codename "Buster"
# FROM photoprism/develop:impish   # Ubuntu 21.10, Codename "Impish Indri"

# copy entrypoint script to container
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh

# define working directory in container
WORKDIR "/go/src/github.com/photoprism/photoprism"

# copy project source code to container
COPY . .