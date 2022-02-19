FROM photoprism/develop:220219-bullseye

## other base images to choose from...
# FROM photoprism/develop:buster   # Debian 10 (Buster)
# FROM photoprism/develop:impish   # Ubuntu 21.10 (Impish Indri)

# copy entrypoint script to container
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh

# define working directory in container
WORKDIR "/go/src/github.com/photoprism/photoprism"

# copy project source code to container
COPY . .