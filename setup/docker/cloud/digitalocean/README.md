PhotoPrism 1-Click App for DigitalOcean
=======================================

Privately browse, organize, and share your photo collection.

DESCRIPTION
---------------------------------------

PhotoPrism® is a privately hosted app for browsing, organizing, and sharing your photo collection. It makes use of the latest technologies to tag and find pictures automatically without getting in your way. Say goodbye to uploading your visual memories to the cloud!

To learn more, visit https://www.photoprism.app/ or try our [demo](https://try.photoprism.app/).

SOFTWARE INCLUDED
---------------------------------------

- [PhotoPrism latest](https://docs.photoprism.app/release-notes/), AGPL 3
- [Docker CE 23.0](https://docs.docker.com/engine/release-notes/), Apache 2
- [Traefik 2.9](https://github.com/traefik/traefik/releases), MIT
- [MariaDB 10.10](https://mariadb.com/kb/en/release-notes/), GPL 2
- [Ofelia 0.3](https://github.com/mcuadros/ofelia/releases), MIT
- [Watchtower 1.5](https://github.com/containrrr/watchtower/releases), Apache 2

GETTING STARTED
---------------------------------------

It may take a few minutes until your Droplet is provisioned, and all services have been initialized.

The initial admin password is stored on your Droplet, you'll see it when running these commands:

```
ssh root@YOUR-SERVER-IP
cat /root/.initial-password.txt
```

You can then access your instance by opening the following URL in a Web browser (see "Using Let's Encrypt HTTPS" for how to get a valid certificate):

```
https://YOUR-SERVER-IP/
```

All files related to PhotoPrism can be found in `/opt/photoprism`. It is running as "photoprism" (UID 1000) by default.

To edit the main config file containing services, storage paths, and basic settings (save changes by pressing *Ctrl+O*, then *Ctrl+X* to exit):

```
cd /opt/photoprism
nano compose.yaml
```

Remember to restart services for changes to take effect:

```
docker compose stop
docker compose up -d
```

## Using Let's Encrypt HTTPS ##

By default, a self-signed certificate will be used for HTTPS connections. Browsers are going to show a security warning because of that. Depending on your settings, they may also refuse connecting at all.

To get an official, free HTTPS certificate from Let's Encrypt, your server needs a fully qualified public domain name, e.g. "photos.yourdomain.com".

You may add a static DNS entry (on DigitalOcean go to Networking > Domains) for this, or use a Dynamic DNS service of your choice.

Once your server has a public domain name, please disable the self-signed certificate and enable domain based routing in compose.yaml and traefik.yaml (see inline instructions in !! UPPERCASE !!):

```
ssh root@YOUR-SERVER-IP
cd /opt/photoprism
nano compose.yaml
nano traefik.yaml
```

Then restart services for the changes to take effect:

```
docker compose stop
docker compose up -d
```

You should now be able to access your instance without security warnings:

```
https://photos.yourdomain.com/
```

Note the first request may still fail while Traefik gets and installs the new certificate. Try again after 30 seconds.

## System Requirements ##

We recommend hosting PhotoPrism on a server with at least 2 cores and 4 GB of memory. Indexing and searching may be slow on smaller Droplets, depending on how many and what types of files you upload.
