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

## Container Image Builds

- **Pin the dev `Dockerfile` `FROM` to a dated `photoprism/develop` tag, never the floating codename.** Use `photoprism/develop:<YYMMDD>-<codename>` (e.g. `photoprism/develop:260520-resolute`), not `photoprism/develop:resolute` / `:latest` / `:ubuntu`. The dated tag pins to a known-good build so a drive-by `docker pull` doesn't silently swap in a fresh base mid-session, and bumping the date in a commit gives developers + reviewers a visible signal that a new base image is available. The `buildx-multi.sh` script publishes both `:<codename>` and `:<YYMMDD>-<codename>` on every build — the dated one is the durable pointer. When you bump the base image, also update `Dockerfile`, the host `compose.yaml` (if it references the tag explicitly), and any docs that quote the current dev image.
- **Ghostscript AppArmor shim (`install -m 0755 /usr/bin/gs /usr/local/bin/gs`).** The develop `Dockerfile`s install an unconfined `gs` copy at `/usr/local/bin/gs`. Ubuntu's AppArmor profile attaches to `/usr/bin/gs` and permits reads only under standard locations (home dirs, `/tmp`, media mounts). The dev tree is an XFS bind mount at `/go/src/github.com/photoprism/photoprism`, which the profile doesn't cover, so a confined `gs` gets `EACCES` reading originals and PDF cover generation fails (`/undefinedfilename … Permission denied`) while PostScript rendering and `/tmp` reads still work. AppArmor attaches by path, so the copy runs unconfined; ImageMagick's PDF/PS delegate calls bare `gs` and `/usr/local/bin` precedes `/usr/bin` on `PATH`, so the copy shadows the confined binary for `convert`. Dev-only — it bypasses the gs sandbox and doesn't belong in runtime images that process untrusted PDFs. Keep it inline (a wrapper script would re-exec `/usr/bin/gs` and re-enter the profile).
- **Never mix Debian and Ubuntu `apt` repositories in the same image:**
  - Don't add a Debian source to an Ubuntu base (or vice versa) to install a single missing package — the transitive deps drift, apt's solver pulls newer libraries from the foreign distro, and other build steps in the same `RUN` (e.g. `install-libheif.sh` running `apt-get install libavcodec-dev`) silently link against the wrong soname.
  - Symptoms surface much later as `dlopen: libfoo.so.N: cannot open shared object file` at image runtime, with the binary referencing a soname that exists only in the foreign distro.
  - If a package isn't available in the host distro's repos, prefer (a) a same-distro PPA / backports source, (b) a vendor-supplied .deb (e.g. Google Chrome from `dl.google.com`), or (c) a from-source build pinned to a known version.
