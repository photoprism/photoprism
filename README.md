PhotoPrism
==========

[![Powered By](https://img.shields.io/badge/powered%20by-Go,%20TensorFlow%20%26%20Vuetify-blue.svg)][powered by]
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)][license]
[![Code Quality](https://goreportcard.com/badge/github.com/photoprism/photoprism)][code quality]
[![GitHub issues](https://img.shields.io/github/issues/photoprism/photoprism.svg)][issues]
[![Build Status](https://travis-ci.org/photoprism/photoprism.png?branch=master)][ci]

[powered by]: https://www.tensorflow.org/install/install_go
[license]: https://github.com/photoprism/photoprism/blob/master/LICENSE
[code quality]: https://goreportcard.com/report/github.com/photoprism/photoprism
[issues]: https://github.com/photoprism/photoprism/issues
[ci]: https://travis-ci.org/photoprism/photoprism

PhotoPrism is a server-based application for automatically tagging, searching and organizing digital photo collections.
It is functionally similar to popular cloud services such as [Flickr](https://www.flickr.com/) or [Google Photos](https://photos.google.com/).
Originals are stored in the file system in a structured way for easy backup and reliable long-term accessibility.

![](assets/docs/img/screenshot-detailview.jpg)

Overview
--------

Our goal is to provide the following features for the first release (tested as a proof-of-concept and partly implemented):

- Easy-to-use Web interface based on [Material Design](https://material.io/) (20% implemented)
- High-performance command line tool (80% implemented)
- No proprietary or binary data formats
- Automatic RAW to JPEG conversion (implemented using [Darktable](https://www.darktable.org/))
- Duplicate detection (RAW plus multiple JPEG files can be used simultaneously) (implemented)
- Automated tagging using [Google TensorFlow](https://www.tensorflow.org/install/install_go) (90% implemented)
- [Reverse geocoding](https://wiki.openstreetmap.org/wiki/Nominatim#Reverse_Geocoding) based on latitude and longitude (implemented using OpenStreetMap; support for Google Maps on the todo list)
- Image search with powerful filters (40% implemented)
- Albums to organize your photos (0% implemented)
- Easy backup and export (10% implemented)

Please ask if you have any questions and leave a star if you like this project. A more detailed documentation - also for non-developers - will follow.

Setup
-----
Before you start, make sure you got Git and Docker installed on your system.
Instead of using Docker, you can also setup your own runtime environment
based on the existing Docker configuration (not recommended).

**Step 1:** Run [Git](https://getcomposer.org/) to clone this project:

```
git clone git@github.com:photoprism/photoprism.git
```

**Step 2:** Start [Docker](https://www.docker.com/) containers:

```
cd photoprism
docker-compose up
```

*Note: This docker-compose configuration is for testing and development purposes only.*

**Step 3:** Open a terminal to run tests and commands:

```
docker-compose exec photoprism bash
make
make test
make install
go run cmd/photoprism/photoprism.go start
```

See [Quick and easy guide for migrating to Go 1.11 modules](https://blog.liquidbytes.net/2018/09/quick-and-easy-guide-for-migrating-to-go-1-11-modules/) for an introduction to Go Modules and Makefiles.

Directory Layout
----------------

The directory layout is loosely based on https://github.com/golang-standards/project-layout.

Assets like photos, built JavaScript/CSS files and HTML templates are located in `assets/` by default. You can configure individual paths in the config file, using environment variables or command flags.

Example configuration files can be found in `configs/`.

The frontend code is located in `frontend/`. Developers run `npm run dev` to watch files and automatically re-build them when changed.

All other paths contain Go source code and scripts used for building the application.

Web Frontend
------------
Open a terminal an type `photoprism start` to start the built-in server. It will listen on port 80 by default.
The UI is based on [Vuetify](https://vuetifyjs.com/en/), a Material Design component framework for Vue.js 2.

Command-line Interface
----------------------

Running `photoprism` without arguments displays usage hints:

```
NAME:
   PhotoPrism - Digital Photo Archive

USAGE:
   photoprism [global options] command [command options] [arguments...]

COMMANDS:
     config      Displays global configuration values
     start       Starts web server
     migrate     Automatically migrates / initializes database
     import      Imports photos
     index       Re-indexes all originals
     convert     Converts RAW originals to JPEG
     thumbnails  Creates thumbnails
     export      Exports photos as JPEG
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                              run in debug mode [$PHOTOPRISM_DEBUG]
   --config-file FILENAME, -c FILENAME  load configuration from FILENAME (default: "/etc/photoprism/photoprism.yml") [$PHOTOPRISM_CONFIG_FILE]
   --darktable-cli FILENAME             darktable command-line executable FILENAME (default: "/usr/bin/darktable-cli") [$PHOTOPRISM_DARKTABLE_CLI]
   --originals-path PATH                originals PATH (default: "/var/photoprism/photos/originals") [$PHOTOPRISM_ORIGINALS_PATH]
   --thumbnails-path PATH               thumbnails PATH (default: "/var/photoprism/photos/thumbnails") [$PHOTOPRISM_THUMBNAILS_PATH]
   --import-path PATH                   import PATH (default: "/var/photoprism/photos/import") [$PHOTOPRISM_IMPORT_PATH]
   --export-path PATH                   export PATH (default: "/var/photoprism/photos/export") [$PHOTOPRISM_EXPORT_PATH]
   --assets-path PATH                   assets PATH (default: "/var/photoprism") [$PHOTOPRISM_ASSETS_PATH]
   --database-driver DRIVER             database DRIVER (mysql, mssql, postgres or sqlite) (default: "mysql") [$PHOTOPRISM_DATABASE_DRIVER]
   --database-dsn DSN                   database data source name (DSN) (default: "photoprism:photoprism@tcp(localhost:3306)/photoprism") [$PHOTOPRISM_DATABASE_DSN]
   --help, -h                           show help
   --version, -v                        print the version
```

### Example

```
# ./photoprism import
Importing photos from /photos/import...
Moving main raw file "Canon 6D South Africa 2018/IMG_2171.CR2" to "/photos/originals/2017/12/20171226_093058_8BA53355C9BF.cr2"
Converting "/photos/originals/2017/12/20171226_093058_8BA53355C9BF.cr2" to "/photos/originals/2017/12/20171226_093058_8BA53355C9BF.jpg"
Added main raw file "2017/12/20171226_093058_8BA53355C9BF.cr2"
Added related jpg file "2017/12/20171226_093058_8BA53355C9BF.jpg"
Moving main raw file "Canon 6D South Africa 2018/IMG_2172.CR2" to "/photos/originals/2017/12/20171226_093107_B522D1D35DD7.cr2"
Converting "/photos/originals/2017/12/20171226_093107_B522D1D35DD7.cr2" to "/photos/originals/2017/12/20171226_093107_B522D1D35DD7.jpg"
Added main raw file "2017/12/20171226_093107_B522D1D35DD7.cr2"
Added related jpg file "2017/12/20171226_093107_B522D1D35DD7.jpg"
Moving main raw file "Canon 6D South Africa 2018/IMG_2173.CR2" to "/photos/originals/2017/12/20171226_093117_E1EEE95F488F.cr2"
Converting "/photos/originals/2017/12/20171226_093117_E1EEE95F488F.cr2" to "/photos/originals/2017/12/20171226_093117_E1EEE95F488F.jpg"
Added main raw file "2017/12/20171226_093117_E1EEE95F488F.cr2"
Added related jpg file "2017/12/20171226_093117_E1EEE95F488F.jpg"
Moving main raw file "Canon 6D South Africa 2018/IMG_2174.CR2" to "/photos/originals/2017/12/20171226_093120_9D205FF627B3.cr2"
Converting "/photos/originals/2017/12/20171226_093120_9D205FF627B3.cr2" to "/photos/originals/2017/12/20171226_093120_9D205FF627B3.jpg"
Added main raw file "2017/12/20171226_093120_9D205FF627B3.cr2"
Added related jpg file "2017/12/20171226_093120_9D205FF627B3.jpg"
Deleted empty directory "/photos/import/Canon 6D South Africa 2018"
Done.
```

Concept
-------

![](assets/docs/img/concept.jpg)
