# Dockerfiles for Development and Production

[**Dockerfiles**](https://docs.docker.com/engine/reference/builder/) are text documents that contain all commands a user could call in a terminal to assemble an application image.

[**Docker Compose**](https://docs.docker.com/compose/) uses [human-friendly YAML files](https://docs.photoprism.app/developer-guide/technologies/yaml/) to configure all application services so you can easily start them with a single command.

See our [Getting Started FAQ](https://docs.photoprism.app/getting-started/faq/#how-can-i-install-photoprism-without-docker) for alternative installation methods, for example using the [*tar.gz* packages](https://github.com/photoprism/photoprism/blob/develop/setup/pkg/linux/README.md) we provide for download at [dl.photoprism.app/pkg/linux/](https://dl.photoprism.app/pkg/linux/README.html).

## What are the benefits of using Docker? ##

**(1) Docker uses standard features of the Linux kernel.** Containers are nothing new; [Solaris Zones](https://en.wikipedia.org/wiki/Solaris_Containers) were released about 20 years ago and the chroot system call was introduced during [development of Version 7 Unix in 1979](https://en.wikipedia.org/wiki/Chroot). It is used ever since for hosting applications exposed to the public Internet. Modern Linux containers are an incremental improvement of this, based on standard functionality that is part of the kernel.

**(2) Docker saves time through simplified deployment and testing.** A main advantage of Docker is that application images can be easily made available to users via Internet. It provides a common standard across most operating systems and devices, which saves our team a lot of time that we can then spend [more effectively](https://docs.photoprism.app/developer-guide/code-quality/#effectiveness-efficiency), for example, providing support and developing one of the many features that users are waiting for.

**(3) Versioned Dockerfiles can be included in the source code repository.** Human-readable and [versioned Dockerfiles as part of our public source code](https://github.com/photoprism/photoprism/tree/develop/docker) help avoid "works for me" moments and other unwelcome surprises by enabling teams to have the exact same environment everywhere in [development](https://github.com/photoprism/photoprism/blob/develop/docker/develop/), staging, and [production](https://github.com/photoprism/photoprism/blob/develop/docker/photoprism/).

**(4) Running applications in containers is more secure.** Last but not least, virtually all file format parsers have vulnerabilities that just haven't been discovered yet. This is a known risk that can affect you even if your computer is not directly connected to the Internet. Running apps in a container with limited host access is an easy way to improve security without compromising performance and usability.

## Why not use virtual machines instead? ##

A virtual machine with a dedicated operating system environment provides even more security, but usually has side effects such as lower performance and more difficult handling. Using a VM, however, doesn't prevent you from running containerized apps to get the best of both worlds. This is essentially what happens when you install Docker on [virtual cloud servers](https://docs.photoprism.app/getting-started/cloud/digitalocean/) and operating systems other than Linux.
