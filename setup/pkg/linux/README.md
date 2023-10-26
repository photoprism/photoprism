# PhotoPrism® Installation Packages

As an alternative to our [Docker images](https://docs.photoprism.app/getting-started/docker-compose/), you can [use the following tar.gz archives](https://dl.photoprism.app/pkg/linux/) to install PhotoPrism on compatible Linux distributions without building it from source:

- <https://dl.photoprism.app/pkg/linux/amd64.tar.gz>
- <https://dl.photoprism.app/pkg/linux/arm64.tar.gz>

Since these packages need to be set up manually and do not include the system dependencies required to make use of all the features, we recommend that only advanced users choose this installation method.

Also note that the minimum required glibc version is 2.35, so for example Ubuntu 22.04 (or later) and Debian Bookworm will work with these binaries, but older Linux distributions may not be compatible.

## Usage

You can download and install PhotoPrism in `/opt/photoprism` by running the following commands:

```
sudo mkdir -p /opt/photoprism
cd /opt/photoprism
wget -c https://dl.photoprism.app/pkg/linux/amd64.tar.gz -O - | sudo tar -xz
sudo ln -sf /opt/photoprism/bin/photoprism /usr/local/bin/photoprism
photoprism --version
```

If your server has an ARM-based CPU, make sure to install `arm64.tar.gz` instead of `amd64.tar.gz` when using the commands above. Both are linked to the latest available build.

## Updates

To update your installation, please stop all running PhotoPrism instances and replace the contents of the installation directory, e.g. `/opt/photoprism`, with the new version.

## Dependencies

In order to use all PhotoPrism features and have full file format support, additional distribution packages must be installed manually as they are not included in the tar.gz archive, for example exiftool, darktable, rawtherapee, imagemagick, libheif, ffmpeg, libavcodec-extra, mariadb, sqlite3, and tzdata.

For details on the packages installed in our official Docker images, see <https://github.com/photoprism/photoprism/tree/develop/docker/develop>.

## Configuration

Run `photoprism --help` in a terminal to get an overview of the command flags and environment variables available for configuration. Their current values can be displayed with the `photoprism config` command.

Please keep in mind that any changes you make to the global [config options](https://docs.photoprism.app/getting-started/config-options/) will require a restart to take effect.

### `defaults.yml`

Default values, including the config path to use, may optionally be specified in a `defaults.yml` file located in `/etc/photoprism`. A custom default config filename can be set with the `PHOTOPRISM_DEFAULTS_YAML` variable or the `--defaults-yaml` command flag.

Since you only need to specify the values for which you want to set a default, a `defaults.yml` file does not need to contain a value for all the available options and can thus be kept to a minimum, e.g.:

```
Debug: false
ReadOnly: true
JpegQuality: 85
```

### `options.yml`

Config defaults set in a `defaults.yml` file can be overridden by values in an [`options.yml`](https://docs.photoprism.app/getting-started/config-files/) file, command flags, and environment variables. The config path from which the `options.yml` file is loaded, if it exists, can be specified with the `PHOTOPRISM_CONFIG_PATH` variable or the `--config-path` command flag.

For a list of supported options and their names, see <https://docs.photoprism.app/getting-started/config-files/>.

## Documentation

For detailed information on specific features and related resources, see our [Knowledge Base](https://www.photoprism.app/kb), or check the [User Guide](https://docs.photoprism.app/user-guide/) for help [navigating the user interface](https://docs.photoprism.app/user-guide/navigate/), a [complete list of config options](https://docs.photoprism.app/getting-started/config-options/), and [additional installation methods](https://docs.photoprism.app/getting-started/):

- [PhotoPrism® User Guide](https://docs.photoprism.app/user-guide/)
- [PhotoPrism® Developer Guide](https://docs.photoprism.app/developer-guide/)
- [PhotoPrism® Knowledge Base](https://www.photoprism.app/kb)

## Getting Support

If you need help installing our software at home, you are welcome to post your question in [GitHub Discussions](https://link.photoprism.app/discussions) or ask in our [Community Chat](https://link.photoprism.app/chat). Common problems can be quickly diagnosed and solved using our [Troubleshooting Checklists](https://docs.photoprism.app/getting-started/troubleshooting/). [Silver, Gold, and Platinum](https://link.photoprism.app/membership) members are also welcome to email us for technical support and advice.

[View Support Options ›](https://www.photoprism.app/kb/getting-support)
