FROM photoprism/develop:20220116

# Copy latest entrypoint script
COPY --chown=root:root /docker/develop/entrypoint.sh /entrypoint.sh
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

COPY . .