PhotoPrism: Browse your life in pictures
========================================

[![Powered By](https://img.shields.io/badge/powered%20by-Go,%20TensorFlow%20%26%20Vuetify-blue.svg)][powered by]
[![GPL License](https://img.shields.io/badge/license-GPL-blue.svg)][license]
[![Code Quality](https://goreportcard.com/badge/github.com/photoprism/photoprism)][code quality]
[![GitHub issues](https://img.shields.io/github/issues/photoprism/photoprism.svg)][issues]
[![Build Status](https://travis-ci.org/photoprism/photoprism.png?branch=master)][ci]

[powered by]: https://www.tensorflow.org/install/install_go
[license]: https://github.com/photoprism/photoprism/blob/master/LICENSE
[code quality]: https://goreportcard.com/report/github.com/photoprism/photoprism
[issues]: https://github.com/photoprism/photoprism/issues
[ci]: https://travis-ci.org/photoprism/photoprism

We love taking photos and they belong to our most valuable (and storage consuming) assets. Privacy concerns - and the wish to properly
archive them for the next generation - brought us to the conclusion that existing cloud solutions are not the right tool to keep them organized.
At the same time, traditional desktop software like Adobe Lightroom lacks many features and can only be used on a single computer.
That's why we started working on an easy-to-use application that can be hosted at home or on a private server.

![](assets/docs/img/screenshot-detailview.jpg)

Features
-------

* Clearly structured Web interface for browsing, organizing and sharing your personal photo collection.
* Import everything without worrying about duplicates or RAW to JPEG conversion.
* Reverse geocoding and automated tagging based on Google TensorFlow.
* No monthly costs. No proprietary formats. No privacy concerns.

For screenshots see https://photoprism.org/#screenshots

Contribute
----------

We are currently setting up the infrastructure necessary to coordinate a remote team and keep everyone up-to-date.

At the moment, the best way to get in touch is to write an email to hello@photoprism.org. We'd love to hear from you!

Donations
---------

To continue working full-time and build a community, we are looking for public funding or a private sponsor who shares our vision. Any help and advice is much appreciated.

Since the software is not released yet, we don't want to ask for small donations from individuals. Please leave a star if you like this project, it provides enough motivation to keep going.

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