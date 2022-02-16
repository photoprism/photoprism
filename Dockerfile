FROM photoprism/develop:20220216-bullseye

## other base images to choose from...
# FROM photoprism/develop:buster   # Debian 10 (Buster)
# FROM photoprism/develop:impish   # Ubuntu 21.10 (Impish Indri)

# update NPM JS package manager
RUN npm install -g npm

# copy scripts to test changes
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile

# set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

COPY . .