PhotoPrism for Raspberry Pi (64bit)
===================================

Big thank you to [Guy Sheffer](https://github.com/guysoft) for 
[building](https://github.com/photoprism/photoprism/issues/109) this!

Simply download `docker-compose.yml` (edit to change directories) and run `docker-compose up` to start PhotoPrism:

```
wget https://raw.githubusercontent.com/photoprism/photoprism/develop/docker/photoprism-arm64/docker-compose.yml
sudo docker-compose up
```

*Note: This is work in progress and we'll shorten the URL as soon as possible.
Please read the following OS and hardware requirements before starting the app.*

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

You can also use [UbuntuDockerPi](https://github.com/guysoft/UbuntuDockerPi) to run the image.


