## Agent Runtime (Host vs Container)

Agents MAY run either inside the Development Environment container (recommended) or on the host.

### Detecting the Environment

- **Inside container if** `/.dockerenv` exists (authoritative signal).
- Path hint: when the project path is `/go/src/github.com/photoprism/photoprism` *and* `/.dockerenv` is absent, assume you are on the host with a bind mount.

### Host Mode

- Build local dev image (once): `make docker-build`
- Start services: `docker compose up` (add `-d` for background)
- Follow live app logs: `docker compose logs -f --tail=100 photoprism`
- Execute a single command in the app container: `docker compose exec photoprism <command>`
  - Run as non-root to avoid root-owned files: `docker compose exec -u "$(id -u):$(id -g)" photoprism <command>`
- Open a terminal session: `make terminal`
- Stop everything: `docker compose --profile=all down --remove-orphans` (`make down`)

### Container Mode

- Install deps: `make dep`
- Build frontend/backend: `make build-js` and `make build-go`
- Watch frontend changes: `make watch-js`
- Start PhotoPrism server: `./photoprism start`
  - HTTP: http://localhost:2342/
  - HTTPS: https://app.localssl.dev/ (via Traefik, if running)
- Admin Login: Default credentials are `admin` / `photoprism`; check `compose.yaml` for `PHOTOPRISM_ADMIN_USER` and `PHOTOPRISM_ADMIN_PASSWORD` if they differ.
- Do not use the Docker CLI inside the container; starting/stopping services requires host Docker access.

### Operating Systems & Architectures

- Guides and command examples assume Linux/Unix shell on 64-bit AMD64 or ARM64.
- For Windows-specifics, see the Developer Guide FAQ: https://docs.photoprism.app/developer-guide/faq/#can-your-development-environment-be-used-under-windows

### CLI Binary Names

The CLI name is `photoprism` in production and public docs. Other binary names (`photoprism-plus`, `photoprism-pro`, `photoprism-portal`) are only used in development builds for side-by-side comparisons.
