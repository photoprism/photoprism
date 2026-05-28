# Ubuntu 26.04 LTS (Resolute Raccoon)
FROM photoprism/develop:260527-resolute

# Harden npm usage by default (applies to npm ci / install in dev container)
ENV NPM_CONFIG_IGNORE_SCRIPTS=true

## Alternative Environments:
# FROM photoprism/develop:questing # Ubuntu 25.10 (Questing Quokka)
# FROM photoprism/develop:plucky   # Ubuntu 25.04 (Plucky Puffin)
# FROM photoprism/develop:armv7    # ARMv7 (32bit)
# FROM photoprism/develop:oracular # Ubuntu 24.10 (Oracular Oriole)
# FROM photoprism/develop:noble    # Ubuntu 24.04 LTS (Noble Numbat)
# FROM photoprism/develop:mantic   # Ubuntu 23.10 (Mantic Minotaur)
# FROM photoprism/develop:lunar    # Ubuntu 23.04 (Lunar Lobster)
# FROM photoprism/develop:jammy    # Ubuntu 22.04 LTS (Jammy Jellyfish)
# FROM photoprism/develop:impish   # Ubuntu 21.10 (Impish Indri)
# FROM photoprism/develop:bookworm # Debian 12 (Bookworm)
# FROM photoprism/develop:bullseye # Debian 11 (Bullseye)
# FROM photoprism/develop:buster   # Debian 10 (Buster)

# Set default working directory. Override WORKING_DIR (e.g. via compose build
# args populated from .env) to match a custom host-side clone layout; the value
# must stay aligned with compose's working_dir and the source bind mount target.
ARG WORKING_DIR=/go/src/github.com/photoprism/photoprism
WORKDIR "${WORKING_DIR}"

# Copy source to image.
COPY . .

# Update scripts in image.
COPY --chown=root:root ./scripts/dist/ /scripts/

# Re-install the dev "mariadb" client config so a custom MARIADB_PORT in .env
# is honored even when the base image was built before the port=<n> line was
# removed (no-op once the next dated base image picks up the new .my.cnf).
COPY --chown=root:root --chmod=644 ./.my.cnf /etc/my.cnf

RUN sudo /scripts/install-yt-dlp.sh
