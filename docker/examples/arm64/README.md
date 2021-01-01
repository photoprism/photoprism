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
    Please change `PHOTOPRISM_ADMIN_PASSWORD` so that PhotoPrism starts with a secure **initial password**.
    Never use `photoprism` or `insecure` as password if you're running it on a public server.

## Docker Compose Command Reference ##

Please prefix with `sudo` if your current user doesn't have permission
to run Docker commands:

| Action   | Command                                                   |
|----------|-----------------------------------------------------------|
| Update   | `docker-compose pull photoprism`                          |
| Stop     | `docker-compose stop photoprism`                          |
| Start    | `docker-compose up -d photoprism`                         |
| Logs     | `docker-compose logs --tail=25 -f`                        |
| Terminal | `docker-compose exec photoprism bash`                     |
| Help     | `docker-compose exec photoprism photoprism help`          |
| Config   | `docker-compose exec photoprism photoprism config`        |
| Reset    | `docker-compose exec photoprism photoprism reset`         |
| Backup   | `docker-compose exec photoprism photoprism backup -a -i`  |
| Restore  | `docker-compose exec photoprism photoprism restore -a -i` |
| Import   | `docker-compose exec photoprism photoprism import`        |
| Index    | `docker-compose exec photoprism photoprism index`         |
| Reindex  | `docker-compose exec photoprism photoprism index -a`      |

## System Requirements ##

You need to boot your Raspberry Pi 3/4 with the parameter `arm_64bit=1` in `config.txt` in order to use this image.
Alternatively, you can run the image on [UbuntuDockerPi](https://github.com/guysoft/UbuntuDockerPi).
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



