PhotoPrism for Raspberry Pi (64bit)
===================================

Big thank you to [Guy Sheffer](https://github.com/guysoft) for 
[building](https://github.com/photoprism/photoprism/issues/109) this!

Download our [docker-compose.yml](https://dl.photoprism.org/docker/arm64/docker-compose.yml) file
(right click and *Save Link As...* or use `wget`) to a folder of your choice,
change the configuration as needed, and run `sudo docker-compose up` to start PhotoPrism:

```
wget https://dl.photoprism.org/docker/arm64/docker-compose.yml
sudo docker-compose up
```

The default port 2342 and other configuration values may be changed as needed,
see [Setup Using Docker Compose](https://docs.photoprism.org/getting-started/docker-compose/)
and [Config Options](https://docs.photoprism.org/getting-started/config-options/) for details.

Our repository on Docker Hub: https://hub.docker.com/r/photoprism/photoprism-arm64

!!! attention
    Please change PHOTOPRISM_ADMIN_PASSWORD so that PhotoPrism starts with a secure initial password. 
    Never use "photoprism", or other easy-to-guess passwords, on a public server.

## Docker Compose Command Reference ##

All commands may have to be prefixed with `sudo` when not running as root.
Note that this will change the home directory `~` to `/root` in your configuration.

| Action           | Command                                                   |
|------------------|-----------------------------------------------------------|
| Start            | `docker-compose up -d photoprism`                         |
| Stop             | `docker-compose stop photoprism`                          |
| Update           | `docker-compose pull photoprism`                          |
| View Logs        | `docker-compose logs --tail=25 -f`                        |
| Open Terminal    | `docker-compose exec photoprism bash`                     |
| Show Help        | `docker-compose exec photoprism photoprism help`          |
| Show Config      | `docker-compose exec photoprism photoprism config`        |
| Reset Database   | `docker-compose exec photoprism photoprism reset`         |
| Backup Database  | `docker-compose exec photoprism photoprism backup -a -i`  |
| Restore Database | `docker-compose exec photoprism photoprism restore -a -i` |
| Index Library    | `docker-compose exec photoprism photoprism index`         |
| Complete Rescan  | `docker-compose exec photoprism photoprism index --all`   |
| Import Files     | `docker-compose exec photoprism photoprism import`        |

!!! info "Complete Rescan"
    `photoprism index --all` will re-index all originals, including already indexed and unchanged files. This may be
    necessary after upgrading, especially to new major versions.

## System Requirements ##

You need to boot your Raspberry Pi 3 / 4 with the parameter `arm_64bit=1` in `config.txt` in order to use this image.
Alternatively, you may run the image on [UbuntuDockerPi](https://github.com/guysoft/UbuntuDockerPi).
It's a 64bit Ubuntu Server with Docker pre-installed.

Indexing large photo and video collections significantly benefits from fast, local SSD storage and enough memory for caching.

!!! tip "Reducing Server Load"
    If you're running out of memory - or other system resources - while indexing, please limit the
    [number of workers](https://docs.photoprism.org/getting-started/config-options/) by setting
    `PHOTOPRISM_WORKERS` to a value less than the number of logical CPU cores in `docker-compose.yml`.
    Also make sure your server has [swap](https://opensource.com/article/18/9/swap-space-linux-systems)
    configured so that indexing doesn't cause restarts when there are memory usage spikes.
    As a measure of last resort, you may additionally disable image classification using TensorFlow.

To avoid permission issues, docker-compose.yml should include the following security options:

```
  photoprism:
    security_opt:
      - seccomp:unconfined
      - apparmor:unconfined
```

## Additional Documentation ##

- https://docs.photoprism.org/getting-started/raspberry-pi/
- https://docs.photoprism.org/getting-started/faq/#why-is-photoprism-getting-stuck-in-a-restart-loop
- https://www.raspberrypi.org/documentation/installation/installing-images/README.md



