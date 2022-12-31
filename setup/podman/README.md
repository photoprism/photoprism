# PhotoPrism Setup (RedHat, CentOS, Fedora, and AlmaLinux)

Running this command will install the required dependencies and download the configuration files:

```
mkdir -p /opt/photoprism
cd /opt/photoprism
curl -sSf https://dl.photoprism.app/podman/install.sh | bash
```

## Docker

Users of RedHat-based Linux distributions can substitute the `docker` and `docker compose` commands with `podman` and `podman-compose` as [drop-in replacements](https://docs.photoprism.app/getting-started/troubleshooting/docker/#redhat-linux).

## Firewall Settings

### Incoming Requests

By default, the application is accessible via port 2342 on all network devices. If you use a firewall, please make sure that this port is reachable from other computers on your network.

### Outgoing Connections

For the installation script and app to work as expected, we recommend whitelisting requests to the prsm.app, [photoprism.app](https://photoprism.app), and photoprism.xyz domains and their subdomains, e.g.:

- prsm.app
- dl.photoprism.app
- my.photoprism.app
- api.photoprism.app
- cdn.photoprism.app
- hub.photoprism.app
- setup.photoprism.app
- places.photoprism.app
- places.photoprism.xyz

Visit https://docs.photoprism.app/getting-started/#maps-places to learn more.

In addition, the following domains should be whitelisted so that Docker can pull public images, e.g. for MariaDB:

- auth.docker.io
- registry-1.docker.io
- index.docker.io
- dseasb33srnrn.cloudfront.net
- production.cloudflare.docker.com

----

*PhotoPrism® is a [registered trademark](https://photoprism.app/trademark). By using the software and services we provide, you agree to our [Terms of Service](https://photoprism.app/terms), [Privacy Policy](https://photoprism.app/privacy), and [Code of Conduct](https://photoprism.app/code-of-conduct). Docs are [available](https://link.photoprism.app/github-docs) under the [CC BY-NC-SA 4.0 License](https://creativecommons.org/licenses/by-nc-sa/4.0/); [additional terms](https://github.com/photoprism/photoprism/blob/develop/assets/README.md) may apply.*