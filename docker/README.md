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

Modern Linux containers are an incremental improvement. A main advantage of Docker is that application images
can be easily made available to users via Internet. It provides a common standard across most operating
systems and devices, which saves our team a lot of time that we can then spend more effectively, for example,
providing support and developing one of the many features that users are waiting for.

Human-readable and versioned Dockerfiles as part of our public source code also help avoid surprises and
"works for me" moments by enabling us to have the exact same environment everywhere in [development](develop/Dockerfile)
and [production](photoprism/Dockerfile).

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
make all install
```

Missing build dependencies must be installed manually as shown in our human-readable and versioned
[Dockerfile](https://github.com/photoprism/photoprism/blob/develop/docker/develop/Dockerfile). You often don't
need to use the exact same versions, so it's possible to replace packages with what is available in your environment.

Note we don't have the resources to provide private users with dependencies and TensorFlow libraries for their personal
environments. We therefore recommend learning Docker if your operating system supports it. Docker vastly simplifies
installation and upgrades. It saves our team a lot of time that we can then spend more effectively.

### Installation Packages ###

Everyone is invited to [contribute by building and testing standalone packages](https://docs.photoprism.app/developer-guide/)
for Linux distributions and other operating systems. We currently don't have the resources to do this,
as new versions are released several times a month and the long list of dependencies would lead to an enormous
testing effort.

### BSD Ports 😈 ###

An [unofficial port](https://docs.photoprism.app/getting-started/freebsd/) is available for FreeBSD / FreeNAS users. 
