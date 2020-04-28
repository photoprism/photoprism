PhotoPrism for Raspberry Pi (64bit)
===================================

Big thank you to [Guy Sheffer](https://github.com/guysoft) for 
[building](https://github.com/photoprism/photoprism/issues/109) this!

Simply download our [`docker-compose.yml`](https://dl.photoprism.org/docker/arm64/docker-compose.yml) (edit to 
change config) and run `docker-compose up` to start PhotoPrism:

```
wget https://dl.photoprism.org/docker/arm64/docker-compose.yml
sudo docker-compose up
```

Image name on Docker Hub: [`photoprism/photoprism-arm64`](https://hub.docker.com/repository/docker/photoprism/photoprism-arm64)

## Operating System and Hardware Requirements ##

You need to boot your Raspberry Pi 3/4 with the parameter `arm_64bit=1` in `config.txt`
to be able to use this image.
A fast SD card and 4 GB of RAM are recommended, in addition you might want to add swap for large photo collections.

Make sure your docker compose configuration contains the following setting:

```
  photoprism:
    security_opt:
      - seccomp:unconfined
```

Alternatively, you can run the image on [UbuntuDockerPi](https://github.com/guysoft/UbuntuDockerPi). It's a 64bit Ubuntu Server with Docker pre-installed.

See also:
https://www.raspberrypi.org/documentation/installation/installing-images/README.md

