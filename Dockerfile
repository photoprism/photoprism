FROM photoprism/develop:20220213

# update NPM JS package manager
RUN npm update -g npm

# copy scripts
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile

# set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

COPY . .