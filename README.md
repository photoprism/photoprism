# PhotoPrism Fork — raphaelmatto

Personal fork of [PhotoPrism](https://github.com/photoprism/photoprism) with display quality enhancements for retina/HiDPI screens.

## Features Added

### Display Settings (Settings > Content > Display)

Three new settings that work independently of each other:

- **Original Images** — Serve original image files in the lightbox viewer instead of generated thumbnails. Preserves per-area sharpening and full resolution from tools like Photoshop.
- **Retina Lightbox** — Scale lightbox images by `1/devicePixelRatio` for pixel-perfect quality on HiDPI displays (e.g. Apple Retina). Works with both originals and thumbnails.
- **Retina Thumbnails** — Use higher-resolution tiles in grid views. Mosaic and list views use `tile_500` instead of `tile_224`, cards view uses `tile_1080` instead of `tile_500`.

All settings take effect immediately without a server restart. They are stored in `settings.yml` under the `display` key.

### ARM64 Docker Build

Custom ARM64 Dockerfile (`docker/photoprism/arm64/Dockerfile`) that fixes incomplete TensorFlow C headers in the upstream `photoprism/develop:bookworm` base image. See [upstream issue #5444](https://github.com/photoprism/photoprism/issues/5444).

## Branch Structure

| Branch | Purpose |
|--------|---------|
| `develop` | Active development. Synced with `upstream/develop`. |
| `production` | Stable deploy branch. Pushes trigger CI to build and push the Docker image. |
| `feature/*` | Feature branches for PRs to upstream. |
| `release` | Upstream release branch (untouched). |

## Remotes

| Remote | URL |
|--------|-----|
| `origin` | `git@github.com:raphaelmatto/photoprism.git` (this fork) |
| `upstream` | `https://github.com/photoprism/photoprism.git` (official repo) |

## CI/CD — Automated Docker Builds

A GitHub Actions workflow (`.github/workflows/build-production.yml`) builds and pushes an ARM64 Docker image on every push to the `production` branch.

- **Runner**: `ubuntu-24.04-arm` (native ARM64, no emulation)
- **Registry**: `ghcr.io/raphaelmatto/photoprism:latest`
- **Dockerfile**: `docker/photoprism/arm64/Dockerfile`
- **Build time**: ~5 minutes

The workflow uses `GITHUB_TOKEN` for registry auth — no secrets to configure.

## How to Deploy

### Server Setup (one-time)

1. Create a GitHub Personal Access Token with `read:packages` scope at https://github.com/settings/tokens/new

2. Authenticate Docker on the server:
   ```bash
   echo "<token>" | docker login ghcr.io -u raphaelmatto --password-stdin
   ```

3. Update the compose file — change one line:
   ```yaml
   # Was:
   image: photoprism/photoprism:latest
   # Now:
   image: ghcr.io/raphaelmatto/photoprism:latest
   ```

### Deploy Workflow

After testing changes locally:

```bash
# 1. Merge to production
git checkout production
git merge develop
git push origin production

# 2. Wait ~5 minutes for GitHub Actions to build

# 3. On the server
docker compose pull
docker compose up -d
```

### Reverting to Official Image

Change the compose file back:
```yaml
image: photoprism/photoprism:latest
```
Then `docker compose pull && docker compose up -d`.

## How to Sync with Upstream

```bash
git checkout develop
git fetch upstream
git merge upstream/develop
# Resolve any conflicts, then:
git push origin develop
```

Then merge to `production` and deploy when ready.

## Recommended Server Settings

For retina display support with minimal disk usage:

| Setting | Value | Reason |
|---------|-------|--------|
| Static Size Limit | 1084+ | Pre-generates `tile_1080` for retina cards view |
| Dynamic Size Limit | 720 | No need for large dynamic thumbnails when serving originals |
| Dynamic Previews | Off | All needed sizes are pre-generated |

After changing the Static Size Limit, regenerate thumbnails:
```bash
# Delete existing thumbnail cache
rm -rf /path/to/photoprism/storage/cache/thumbnails

# Rescan to regenerate
# (resize EC2 instance temporarily if needed for CPU)
```

## Local Development

```bash
docker compose up --build
# In another terminal:
make terminal
make fix-permissions   # first time only
make dep
make build-js
make build-go
./photoprism start
```

Frontend hot reload: `make watch-js` (in a separate terminal inside the container).

Go changes require `make build-go` and restarting PhotoPrism.

## Upstream Contributions

- **PR [#5442](https://github.com/photoprism/photoprism/pull/5442)** — Display settings (Original Images, Retina Lightbox, Retina Thumbnails)
- **Issue [#5444](https://github.com/photoprism/photoprism/issues/5444)** — ARM64 TensorFlow header bug in base image

### TODO
- [ ] Validate retina thumbnails in: Labels, photo preview, edit details, edit labels, edit files, batch edit, account avatar

### Future Ideas (not yet implemented)
- Configurable lightbox background color (e.g. `#303030` instead of black)
- Configurable metadata templates
- Non-square image packing in grid views
