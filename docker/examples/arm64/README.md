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

Image name on Docker Hub: [`photoprism/photoprism-arm64`](https://hub.docker.com/r/photoprism/photoprism-arm64)

## System Requirements ##

You need to boot your Raspberry Pi 3/4 with the parameter `arm_64bit=1` in `config.txt` in order to use this image.
Alternatively, you can run the image on [UbuntuDockerPi](https://github.com/guysoft/UbuntuDockerPi).
It's a 64bit Ubuntu Server with Docker pre-installed.

Indexing large photo and video collections significantly benefits from fast, local SSD storage and enough memory for caching.

If you're running out of memory - or other system resources - while indexing, please reduce the number of workers to a
value less than the number of logical CPU cores. Also make sure your server has swap configured, so that indexing
doesn't cause restarts when there are memory usage spikes. As a measure of last resort, you may additionally disable
image classification using TensorFlow.

To prevent permission issues, your docker-compose.yaml config must include the following security options:

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



