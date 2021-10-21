FROM photoprism/development:20211021

# Copy latest entrypoint script
COPY --chown=root:root /docker/development/entrypoint.sh /entrypoint.sh
COPY --chown=root:root /docker/scripts/Makefile /root/Makefile

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"

COPY . .