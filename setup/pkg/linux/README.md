# PhotoPrism® Installation Packages

As an alternative to our [Docker images](https://docs.photoprism.app/getting-started/docker-compose/), you can use the packages available at [**dl.photoprism.app/pkg/linux/**](https://dl.photoprism.app/pkg/linux/) to install PhotoPrism on compatible Linux distributions without [building it from source](https://docs.photoprism.app/getting-started/faq/#building-from-source).

These [binary installation packages](https://dl.photoprism.app/pkg/linux/) are intended for **experienced users** and **maintainers of third-party integrations** only, as they [require manual configuration](#configuration) and [do not include tested system dependencies](#dependencies). Since we are unable to [provide support](https://www.photoprism.app/kb/getting-support) for custom installations, we recommend using [one of our Docker images](https://docs.photoprism.app/getting-started/docker-compose/) to run PhotoPrism on a private server or NAS device.

Also note that the minimum required glibc version is 2.35, so for example Ubuntu 22.04 and Debian Bookworm will work, but older Linux distributions may not be compatible.

## Usage

### Installation Using *tar.gz* Archives

You can download and install PhotoPrism in `/opt/photoprism` by running the following commands:

```
sudo mkdir -p /opt/photoprism
cd /opt/photoprism
wget -c https://dl.photoprism.app/pkg/linux/amd64.tar.gz -O - | sudo tar -xz
sudo ln -sf /opt/photoprism/bin/photoprism /usr/local/bin/photoprism
photoprism --version
```

If your server has an **ARM-based CPU**, please make sure to install `arm64.tar.gz` instead of `amd64.tar.gz` when using the commands above. Both are linked to the [latest stable release](https://github.com/photoprism/photoprism/releases).

Since the packages currently do not include a default configuration, we recommend that you create a [`defaults.yml`](https://docs.photoprism.app/getting-started/config-files/defaults/) in `/etc/photoprism` next, in which you configure the paths and other settings that you want to use for your instance.

### *.deb* Packages for Ubuntu / Debian Linux

As an alternative to the plain *tar.gz* archives, that you need to unpack manually, we also offer *.deb* packages for Debian-based distributions such as Ubuntu Linux. They install PhotoPrism under `/opt/photoprism`, add a `/usr/local/bin/photoprism` symlink, create `/etc/photoprism/defaults.yml`, and pull in the runtime libraries listed in the [Dependencies](#dependencies) section.

On servers with a **64-bit Intel or AMD CPU**, our [latest stable release](https://github.com/photoprism/photoprism/releases) can be installed as follows:

```
curl -sLO https://dl.photoprism.app/pkg/linux/deb/amd64.deb
sudo apt install --no-install-recommends ./amd64.deb
```

If your server has an **ARM-based CPU**, such as a [Raspberry Pi](https://docs.photoprism.app/getting-started/raspberry-pi/), use the following commands instead:

```
curl -sLO https://dl.photoprism.app/pkg/linux/deb/arm64.deb
sudo apt install --no-install-recommends ./arm64.deb
```

Omit `--no-install-recommends` if you want APT to install MariaDB, Darktable, RawTherapee, ImageMagick, and other recommended extras automatically.

### *.rpm* Packages for Fedora / RHEL / openSUSE

For RPM-based distributions we publish *.rpm* packages that mirror the layout described above. Install the latest release on **x86_64** hardware with:

```
sudo dnf install https://dl.photoprism.app/pkg/linux/rpm/x86_64.rpm
```

and on **aarch64** (ARM64):

```
sudo dnf install https://dl.photoprism.app/pkg/linux/rpm/aarch64.rpm
```

Replace `dnf` with `zypper` on openSUSE (use `--allow-unsigned-rpm` when required). On distributions that do not ship FFmpeg in the base repositories, enable the appropriate multimedia repository (EPEL, RPM Fusion, Packman, …) before installing the dependencies below.

### AUR Packages for Arch Linux

Thomas Eizinger additionally maintains [AUR packages for installation on Arch Linux](https://aur.archlinux.org/packages/photoprism-bin). They are based on our *tar.gz* packages and have a systemd integration so that PhotoPrism can be started and restarted automatically.

[Learn more ›](https://aur.archlinux.org/packages/photoprism-bin)

## Updates

To update your installation, please stop all running PhotoPrism instances and make sure that there are [no media, database, or custom config files](#configuration) in the `/opt/photoprism` directory. You can then delete its contents with the command `sudo rm -rf /opt/photoprism/*` and install a new version as shown above.

If you have used a *.deb* package for installation, you may need to remove the currently installed `photoprism` package by running `sudo dpkg -r photoprism` before you can install a new version with `sudo apt install ./package.deb` or `sudo dpkg -i package.deb`.

## Dependencies

PhotoPrism packages bundle TensorFlow 2.18.0 and, starting with the October 2025 builds, ONNX Runtime 1.22.0 as described in [`ai/face/README.md`](https://github.com/photoprism/photoprism/blob/develop/internal/ai/face/README.md). The shared libraries for both frameworks are shipped inside `/opt/photoprism/lib`, so no additional system packages are needed to switch `PHOTOPRISM_FACE_ENGINE` to `onnx`. The binaries still rely on glibc ≥ 2.35 and the standard C/C++ runtime libraries (`libstdc++6`, `libgcc_s1`, `libgomp1`, …) provided by your distribution.

### Required Runtime Packages

Install the following packages **before** running PhotoPrism so that thumbnailing, metadata extraction, and the SQLite fallback database work out of the box:

| Distribution family          | Command                                                                                                                                              |
|------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| Debian / Ubuntu              | `sudo apt install libvips42t64 libimage-exiftool-perl ffmpeg sqlite3 tzdata`<br/>Use `libvips42` or, as a fallback, `libvips-dev` on older releases. |
| Fedora / RHEL / Alma / Rocky | `sudo dnf install vips perl-Image-ExifTool ffmpeg sqlite tzdata`                                                                                     |
| openSUSE                     | `sudo zypper install vips perl-Image-ExifTool ffmpeg sqlite3 tzdata`                                                                                 |

These packages pull in the full libvips stack (GLib, libjpeg/libtiff/libwebp, archive/zstd, etc.) that the PhotoPrism binary links against. Run `ldd /opt/photoprism/bin/photoprism` if you need to diagnose missing libraries on custom distributions.

### Recommended Extras

For extended RAW processing, HEIF/HEIC support, and database scalability we recommend installing:

- MariaDB or MariaDB Server (external database)
- Darktable and/or RawTherapee (RAW converters)
- ImageMagick (CLI utilities)
- libheif (prefer the up-to-date binaries from [dl.photoprism.app/dist/libheif/](https://dl.photoprism.app/dist/libheif/); install with `bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-libheif.sh)` when distro packages are outdated)
- librsvg2-bin or librsvg2-tools (SVG conversion helpers)

Use `sudo apt install`, `sudo dnf install`, or `sudo zypper install` with the package names above to pull them in as needed.

We publish the same libheif builds that ship in our Docker images. They include fixes for rotation metadata and newer iOS HEIC variants that are often missing from distribution packages. Advanced users can regenerate them via `make build-libheif-*`, which calls `scripts/dist/build-libheif.sh` for each supported base image and architecture before uploading the archives to `dl.photoprism.app`.

Keep in mind that even if all dependencies are installed, it is possible that you are using a version that is not fully compatible with your pictures, phone, or camera. Our team cannot [provide support](https://www.photoprism.app/kb/getting-support) in these cases if the same issue does not occur with our [official Docker images](https://docs.photoprism.app/getting-started/docker-compose/). Details on the packages and package versions we use can be found in the Dockerfiles available in our [public project repository](https://github.com/photoprism/photoprism/tree/develop/docker).

## Configuration

After unpacking the binaries you only need a writable configuration and storage location. The typical workflow is:

1. Inspect the current settings: `photoprism config`
2. Create dedicated directories for runtime data, for example (replace `photoprism:photoprism` with the user/group that should own the data):
   ```
   sudo mkdir -p /var/lib/photoprism/{config,storage}
   sudo chown -R photoprism:photoprism /var/lib/photoprism
   ```
3. Update `/etc/photoprism/defaults.yml` so `ConfigPath`, `StoragePath`, `OriginalsPath`, and `ImportPath` point outside the installation directory.
4. Optionally place per-instance overrides in `<config-path>/options.yml`.
5. Restart the PhotoPrism service (or rerun the CLI command) so the changes take effect.

Run `photoprism --help` in a terminal to get an [overview of the command flags and environment variables](https://docs.photoprism.app/getting-started/config-options/) available for configuration. Their current values can always be displayed with the `photoprism config` command.

### Precedence

PhotoPrism reads settings in the following order (later entries override earlier ones):

| Order | Source                            | Notes                                                                      |
|-------|-----------------------------------|----------------------------------------------------------------------------|
| 1     | Built-in defaults                 | Hard-coded, fall back when nothing else is set.                            |
| 2     | `/etc/photoprism/defaults.yml`    | Global defaults for all users on the host.                                 |
| 3     | Environment variables / CLI flags | Combine with service managers or wrappers.                                 |
| 4     | `<config-path>/options.yml`       | Instance-specific overrides (per user or per deployment).                  |
| 5     | Runtime changes in the UI         | Persisted to `options.yml`; require a restart when running outside Docker. |

If no explicit *originals*, *import* and/or *assets* path has been configured, a list of [default directory paths](https://github.com/photoprism/photoprism/blob/develop/pkg/fs/directories.go) will be searched and the first existing directory will be used for the respective path. To simplify [updates](#updates), we recommend **not** storing media, database files, or custom configs in the installation directory itself (for example `/opt/photoprism`); use another base such as `/var/lib/photoprism` or a path under the user’s home directory.

All configuration changes—whether made [via UI](https://docs.photoprism.app/user-guide/settings/advanced/), [config files](https://docs.photoprism.app/getting-started/config-files/), or [environment variables](https://docs.photoprism.app/getting-started/config-options/)—**require a restart** to take effect when PhotoPrism runs as a standalone process.

### `defaults.yml`

Packages install a starter `/etc/photoprism/defaults.yml`. Adjust it with root privileges to set global defaults such as filesystem locations, database options, and network ports. When specifying strings you can use `~` as the current user’s home directory and relative paths starting with `./`:

```yaml
ConfigPath: "~/.config/photoprism"
StoragePath: "~/.photoprism"
OriginalsPath: "~/Pictures"
ImportPath: "/media"
AdminUser: "admin"
AdminPassword: "insecure"
AuthMode: "password"
DatabaseDriver: "sqlite"
JpegQuality: 85
DetectNSFW: false
UploadNSFW: true
```

For a list of supported options and their names, see <https://docs.photoprism.app/getting-started/config-files/#config-options>.

When specifying values, make sure that the data type is the [same as in the documentation](https://docs.photoprism.app/getting-started/config-files/#config-options), e.g. *bool* values must be either `true` or `false` and *int* values must be whole numbers without any quotes like in the example above.

### `options.yml`

Default config values can be overridden by values [specified in an `options.yml` file](https://docs.photoprism.app/getting-started/config-files/) as well as with command flags and environment variables. To load values from an existing `options.yml` file, you can specify its storage path (excluding the filename) by setting the `ConfigPath` option in your `defaults.yml` file, using the `--config-path` command flag, or with the `PHOTOPRISM_CONFIG_PATH` environment variable.

The values in an `options.yml` file are not global and can be used to customize individual instances e.g. based on the default values in a `defaults.yml` file. Both files allow you to set any of the [supported options](https://docs.photoprism.app/getting-started/config-files/#config-options).

Tip: when running PhotoPrism as a systemd service, export environment variables in the service unit or in `/etc/default/photoprism`. For interactive shells, specify the corresponding flags or prefix commands with variables (for example `PHOTOPRISM_DEBUG=true photoprism index`). Use the smallest scope that fits your deployment so updates stay manageable.

## Documentation

For detailed information on specific features and related resources, see our [Knowledge Base](https://www.photoprism.app/kb), or check the [User Guide](https://docs.photoprism.app/user-guide/) for help [navigating the user interface](https://docs.photoprism.app/user-guide/navigate/), a [complete list of config options](https://docs.photoprism.app/getting-started/config-options/), and [other installation methods](https://docs.photoprism.app/getting-started/):

- [PhotoPrism® User Guide](https://docs.photoprism.app/user-guide/)
- [PhotoPrism® Developer Guide](https://docs.photoprism.app/developer-guide/)
- [PhotoPrism® Knowledge Base](https://www.photoprism.app/kb)

## Getting Support

If you need help installing our software at home, you are welcome to post your question in [GitHub Discussions](https://link.photoprism.app/discussions) or ask in our [Community Chat](https://link.photoprism.app/chat). Common problems can be quickly diagnosed and solved using our [Troubleshooting Checklists](https://docs.photoprism.app/getting-started/troubleshooting/). [Silver, Gold, and Platinum](https://link.photoprism.app/membership) members are also welcome to email us for technical support and advice.

[View Support Options ›](https://www.photoprism.app/kb/getting-support)
