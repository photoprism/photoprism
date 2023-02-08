# PhotoPrism Setup (RedHat, CentOS, Fedora, and AlmaLinux)


[Podman](https://podman.io/) is supported as a replacement for Docker on Red Hat Enterprise Linux® and compatible Linux distributions such as CentOS, Fedora, AlmaLinux, and Rocky Linux. The following installs the `podman` and `podman-compose` commands if they are not already installed:

```
sudo dnf upgrade -y
sudo dnf install epel-release -y
sudo dnf install netavark aardvark-dns podman podman-docker podman-compose -y
sudo systemctl start podman
sudo systemctl enable podman
podman --version
```

We also provide a setup script that conveniently installs Podman and downloads the default configuration to a directory of your choice:

```
mkdir -p /opt/photoprism
cd /opt/photoprism
curl -sSf https://dl.photoprism.app/podman/install.sh | bash
```

Please keep in mind to replace the `docker` and `docker compose` commands with `podman` and `podman-compose` when following the examples in our documentation.

## Documentation

### Getting Started

↪ https://docs.photoprism.app/getting-started/

### Knowledge Base

↪ https://www.photoprism.app/kb

### Compliance FAQ

↪ https://www.photoprism.app/kb/compliance-faq

### Firewall Settings

↪ https://docs.photoprism.app/getting-started/troubleshooting/firewall/

----

*PhotoPrism® is a [registered trademark](https://www.photoprism.app/trademark). By using the software and services we provide, you agree to our [Terms of Service](https://www.photoprism.app/terms), [Privacy Policy](https://www.photoprism.app/privacy), and [Code of Conduct](https://www.photoprism.app/code-of-conduct). Docs are [available](https://link.photoprism.app/github-docs) under the [CC BY-NC-SA 4.0 License](https://creativecommons.org/licenses/by-nc-sa/4.0/); [additional terms](https://github.com/photoprism/photoprism/blob/develop/assets/README.md) may apply.*