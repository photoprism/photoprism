FROM photoprism/develop:20220215

## experimental base images
# FROM photoprism/develop:buster
# FROM photoprism/develop:bullseye

# update NPM JS package manager
RUN npm install -g npm

# copy scripts to test changes
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile

# set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

COPY . .