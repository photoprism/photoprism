========================================================================
PhotoPrism for Cloud Servers
Based on Ubuntu 22.04 LTS (Jammy Jellyfish)
========================================================================

SOFTWARE INCLUDED
------------------------------------------------------------------------

PhotoPrism latest, AGPL 3
Docker CE 23.0.1, Apache 2
Traefik 2.9, MIT
MariaDB 10.10, GPL 2
Ofelia 0.3.7, MIT
Watchtower 1.5.3, Apache 2

GETTING STARTED
------------------------------------------------------------------------

Run this script as root to install PhotoPrism on a cloud server e.g.
at DigitalOcean:

  bash <(curl -s https://dl.photoprism.app/docker/cloud/setup.sh)

This may take a while to complete, depending on the performance of
your server and its internet connection.

When done - and you see no errors - please open

  https://<YOUR SERVER IP>/

in a Web browser and log in using the initial admin password shown
by the script. You may also see the initial password by running

  cat /root/.initial-password.txt

as root on your server. To open a terminal:

  ssh root@<YOUR SERVER IP>

Data and all config files related to PhotoPrism can be found in

  /opt/photoprism

The main docker compose config file for changing config options is

  /opt/photoprism/docker-compose.yml

The server is running as "photoprism" (UID 1000) by default. There's no need
to change defaults unless you experience conflicts with other services running
on the same server. For example, you may need to disable the Traefik reverse
proxy as the ports 80 and 443 can only be used by a single web server / proxy.

Configuring multiple apps on the same server is beyond the scope of this base
config and for advanced users only.

SYSTEM REQUIREMENTS
------------------------------------------------------------------------

We recommend hosting PhotoPrism on a server with at least 2 cores and
4 GB of memory. Beyond these minimum requirements, the amount of RAM
should match the number of cores. Indexing large photo and video
collections significantly benefits from fast, local SSD storage.

RAW image conversion and automatic image classification using TensorFlow
will be disabled on servers with 1 GB or less memory.

USING LET'S ENCRYPT HTTPS
------------------------------------------------------------------------

By default, a self-signed certificate will be used for HTTPS connections.
Browsers are going to show a security warning because of that. Depending
on your settings, they may also refuse connecting at all.

To get an official, free HTTPS certificate from Let's Encrypt, your server
needs a fully qualified public domain name, e.g. "photos.yourdomain.com".

You may add a static DNS entry (on DigitalOcean go to Networking > Domains)
for this, or use a Dynamic DNS service of your choice.

Once your server has a public domain name, please disable the self-signed
certificate and enable domain based routing in docker-compose.yml and
traefik.yaml (see inline instructions in !! UPPERCASE !!):

  ssh root@<YOUR SERVER IP>
  cd /opt/photoprism
  nano docker-compose.yml
  nano traefik.yaml

Then restart services in a terminal for the changes to take effect:

  docker compose stop
  docker compose up -d

To check logs for errors:

  docker compose logs -f

If you see a "letsencrypt.json" file permission error:

  chmod 600 /opt/photoprism/traefik/letsencrypt.json
  docker compose stop
  docker compose up -d

You should now be able to access your instance without security warnings:

  https://photos.yourdomain.com/

Note the first request may still fail while Traefik gets and installs the
new certificate. Try again after 30 seconds.