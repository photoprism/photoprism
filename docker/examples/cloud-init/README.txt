PhotoPrism for Cloud Servers running Ubuntu 20.04 LTS (Focal Fossa)
===================================================================

Attention: This config example is under development, and not stable yet!

Run this script as root to install PhotoPrism on a cloud server e.g. at DigitalOcean:

  bash <(curl -s https://dl.photoprism.org/docker/cloud-init/setup.sh)

This may take a while to complete, depending on the performance of your
server and its internet connection.

When done - and you see no errors - please open

  https://<YOUR SERVER IP>/

in a Web browser and log in using the initial admin password shown
by the script. You may also see it by running

  cat /root/.initial-password.txt

in a terminal.

Data and all config files related to PhotoPrism can be found in /photoprism.
The main docker-compose config file for changing config options is

  /photoprism/docker-compose.yml

The server is running as "photoprism" (UID 1000) by default. There's no need
to change defaults unless you experience conflicts with other services running
on the same server. For example, you may need to disable the Traefik reverse
proxy as the ports 80 and 443 can only be used by a single web server / proxy.

Configuring multiple apps on the same server is beyond the scope of this and for
advanced users only.

## System Requirements ##

We recommend hosting PhotoPrism on a server with at least 2 cores and
4 GB of memory. Beyond these minimum requirements, the amount of RAM
should match the number of cores. Indexing large photo and video
collections significantly benefits from fast, local SSD storage.


