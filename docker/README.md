# Dockerfiles and Docker Compose Examples

[**Dockerfiles**](https://docs.docker.com/engine/reference/builder/) are text documents that contain all commands a user
could call in a terminal to assemble an application image.

[**Docker Compose**](https://docs.docker.com/compose/) uses [human-friendly YAML files](https://docs.photoprism.app/developer-guide/technologies/yaml/)
to configure all application services so you can easily start them with a single command.

## Why are we using Docker? ##

Containers are nothing new; [Solaris Zones](https://en.wikipedia.org/wiki/Solaris_Containers) have been around for
about 15 years, first released publicly in 2004. The chroot system call was introduced during
[development of Version 7 Unix in 1979](https://en.wikipedia.org/wiki/Chroot). It is used ever since for hosting
applications exposed to the public Internet.

Modern Linux containers are an incremental enhancement. A main advantage of Docker is that application images
can be easily made available to users via Internet. It provides a common standard across most operating
systems and devices, which saves our team a lot of time that we can then spend [more effectively](https://docs.photoprism.app/developer-guide/issues/#effectiveness-efficiency), for example,
providing support and developing one of the many features that users are waiting for.

Human-readable and versioned Dockerfiles as part of our public source code also help avoid "works for me" moments and
other unwelcome surprises by enabling teams to have the exact same environment everywhere in
[development](https://github.com/photoprism/photoprism/blob/develop/docker/develop/bookworm/Dockerfile), staging,
and [production](https://github.com/photoprism/photoprism/blob/develop/docker/photoprism/bookworm/Dockerfile).

Last but not least, virtually all file format parsers have vulnerabilities that just haven't been discovered yet.
This is a known risk that can affect you even if your computer is not directly connected to the Internet.
Running apps in a container with limited host access is an easy way to improve security without
compromising performance and usability.

## What about Virtual Machines? ##

A virtual machine running its own operating system provides more security, but typically has side effects
such as lower performance and more difficult handling. You can also run Docker in a VM to get the best of
both worlds. It's essentially what happens when you run dockerized applications on [virtual cloud servers](https://docs.photoprism.app/getting-started/cloud/digitalocean/)
and operating systems other than Linux.

## Alternatives ##

### Building From Source ###

You can build and install PhotoPrism from the publicly available [source code](https://docs.photoprism.app/developer-guide/setup/):

```bash
git clone https://github.com/photoprism/photoprism.git
cd photoprism
make all install DESTDIR=/opt/photoprism
```

Missing build dependencies must be installed manually as shown in our human-readable and versioned
[Dockerfile](https://github.com/photoprism/photoprism/blob/develop/docker/develop/Dockerfile). You often don't
need to use the exact same versions, so it's possible to replace packages with what is available in your environment.

Please note that we do not have the resources to provide private users with dependencies and
[TensorFlow libraries](https://dl.photoprism.app/tensorflow/) for their personal environments. We recommend giving
Docker a try if you use Linux as it saves developers a lot of time when building, testing, and deploying complex
applications like PhotoPrism. It also effectively helps avoid "works for me" moments and missing dependencies.

### Installation Packages ###

An [unofficial port](https://docs.photoprism.app/getting-started/freebsd/) is available for FreeBSD / FreeNAS users.
You are invited to contribute by [building and testing standalone packages](https://docs.photoprism.app/developer-guide/) for Linux distributions and other operating systems.

Updates are [released several times a month](https://docs.photoprism.app/release-notes/), so maintaining the long list of dependencies for additional environments would currently consume too many of [our resources](https://link.photoprism.app/membership).
