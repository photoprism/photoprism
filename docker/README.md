## Dockerfiles for Development & Production

[**Dockerfiles**](https://docs.docker.com/engine/reference/builder/) are text documents that contain all commands a user could call in a terminal to assemble an application image.

[**Docker Compose**](https://docs.docker.com/compose/) uses [human-friendly YAML files](https://docs.photoprism.app/developer-guide/technologies/yaml/) to configure all application services so you can easily start them with a single command.

See our [Getting Started FAQ](https://docs.photoprism.app/getting-started/faq/#how-can-i-install-photoprism-without-docker) for alternative installation methods, for example using the [*tar.gz* packages](https://github.com/photoprism/photoprism/blob/develop/setup/pkg/linux/README.md) we provide for download at [dl.photoprism.app/pkg/linux/](https://dl.photoprism.app/pkg/linux/README.html).

### What Are the Benefits of Using Docker?

**(1) Docker uses standard features of the Linux kernel.** Containers are nothing new; [Solaris Zones](https://en.wikipedia.org/wiki/Solaris_Containers) were released about 20 years ago and the chroot system call was introduced during [development of Version 7 Unix in 1979](https://en.wikipedia.org/wiki/Chroot). It is used ever since for hosting applications exposed to the public Internet. Modern Linux containers are an incremental improvement of this, based on standard functionality that is part of the kernel.

**(2) Docker saves time through simplified deployment and testing.** A main advantage of Docker is that application images can be [easily made available](https://hub.docker.com/r/photoprism/photoprism) to users via Internet. It provides a common standard across most operating systems and devices, which saves our team a lot of time that we can then spend [more effectively](https://docs.photoprism.app/developer-guide/code-quality/#effectiveness-efficiency), for example, providing support and developing one of the many features that users are waiting for.

**(3) Dockerfiles are part of the source code repository.** [Human-readable](https://docs.docker.com/engine/reference/builder/) and [versioned Dockerfiles](https://github.com/photoprism/photoprism/tree/develop/docker) that are part of our public source code help avoid "works for me" moments and other unwelcome surprises by enabling us to have the exact [same environment](https://docs.photoprism.app/developer-guide/setup/) everywhere in [development](https://github.com/photoprism/photoprism/tree/develop/docker/develop), [staging, and production](https://github.com/photoprism/photoprism/tree/develop/docker/photoprism).

**(4) Running applications in containers is more secure.** Last but not least, virtually all file format parsers have vulnerabilities that just haven't been discovered yet. This is a known risk that can affect you even if your computer is not directly connected to the Internet. Running apps in a container with limited host access is an easy way to improve security without compromising performance and usability.

### Why Not Use Virtual Machines Instead?

A virtual machine with a dedicated operating system environment provides even more security, but usually has side effects such as lower performance and more difficult handling. Using a VM, however, doesn't prevent you from running containerized apps to get the best of both worlds. This is essentially what happens when you install Docker on [virtual cloud servers](https://docs.photoprism.app/getting-started/cloud/digitalocean/) and operating systems other than Linux.

### Directory Overview

| Directory     | Purpose                                                                                                                                      |
|---------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `develop/`    | Development images used as the base for the production image and by contributors to build and test PhotoPrism in a reproducible environment. |
| `photoprism/` | Production (runtime) images published as [`photoprism/photoprism`](https://hub.docker.com/r/photoprism/photoprism) on Docker Hub.            |
| `tensorflow/` | Helper image used to build and package the TensorFlow C library shipped with PhotoPrism.                                                     |
| `demo/`       | Build context for the public [demo site](https://demo.photoprism.app/), including Traefik and job configuration.                             |
| `ddns/`       | Minimal image for the Dynamic DNS updater used by our hosted services.                                                                       |
| `dummy/`      | Test doubles for OIDC and WebDAV used by acceptance tests and local development.                                                             |
| `goproxy/`    | Caching Go module proxy used to speed up CI and local builds.                                                                                |

### Which Image Should I Use?

> **Use the latest image whenever possible.** The base image used by the top-level [`Dockerfile`](../Dockerfile) is recommended for development. Currently, this is Ubuntu 25.10 ("Questing Quokka"), i.e. `photoprism/develop:questing` and `photoprism/photoprism:questing`.

New container deployments should pull the `:latest` tag (or an explicit current release tag) from Docker Hub. Contributors setting up a local development environment should follow the [Developer Guide](https://docs.photoprism.app/developer-guide/setup/), which is kept in sync with the top-level `Dockerfile` and `compose.yaml`.

### Legacy Images ("Use at Your Own Risk")

The `develop/` and `photoprism/` directories also contain Dockerfiles for a number of **older Linux distributions**, for example Ubuntu *Jammy*, *Noble*, *Oracular*, *Plucky*, *Impish*, *Lunar*, *Mantic* and Debian *Buster*, *Bullseye*, *Bookworm*, as well as an `armv7` variant.

These files are **kept for documentation and reference purposes only**:

- Their base images, system packages, and bundled dependencies (for example `libvips`, `libheif`, `ffmpeg`, `exiftool`, `darktable`) are typically outdated and may contain unpatched security vulnerabilities.
- We **no longer build, publish, or test** these images as part of our regular release process.
- Features and bug fixes added to the current image are not backported.
- Compatibility with current Go dependencies (notably C bindings such as [`govips`](https://github.com/davidbyttow/govips), [`onnxruntime_go`](https://github.com/yalue/onnxruntime_go), and the TensorFlow C API) is **not guaranteed** — a successful build on an older base does not imply a working runtime.

If you still need to build one of these images, consider it **"use at your own risk"**: you may need to upgrade individual system packages (for example pulling `libvips` from a backport PPA, as `scripts/dist/install-libvips.sh` does for Ubuntu Jammy), adjust pinned Go dependencies, or patch the Dockerfile to match the current upstream. Please do not open issues for problems that only affect legacy images.

For questions about supported environments and current system requirements, see the [System Requirements](https://docs.photoprism.app/getting-started/#system-requirements) documentation.
