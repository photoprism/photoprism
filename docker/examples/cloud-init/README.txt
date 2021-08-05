PhotoPrism for Cloud Servers running Ubuntu 20.04 LTS (Focal Fossa)
===================================================================

Attention: This config example is under development, and not stable yet!

Run this script to install PhotoPrism on a cloud server e.g. at DigitalOcean:

  bash <(curl -s https://dl.photoprism.org/docker/cloud-init/setup.sh)

This may take a while to complete. Then open

  http://<YOUR SERVER IP>:2342/

in a Web browser to log in using the initial admin password "insecure".
Please immediately change it in Settings to protect your new PhotoPrism
installation.

All files related to PhotoPrism can be found in /photoprism. The server
will be running as "photoprism" (UID 1000). There should be no need to
change defaults unless other services are running on the same server
and there are conflicts.

## System Requirements ##

We recommend hosting PhotoPrism on a server with at least 2 cores and
4 GB of memory. Beyond these minimum requirements, the amount of RAM
should match the number of cores. Indexing large photo and video
collections significantly benefits from fast, local SSD storage.


