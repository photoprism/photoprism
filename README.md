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

More screenshots: https://photoprism.org/#screenshots

Features
--------

We focus on what really matters for photographers:

* Clearly structured Web interface for browsing, organizing and sharing your personal photo collection.
* Import everything without worrying about duplicates or RAW to JPEG conversion.
* Reverse geocoding and automated tagging based on Google TensorFlow.
* No monthly costs. No proprietary formats. No privacy concerns.

*Note: This is not a photo editor. All images are stored in the file system, so you can continue using your favorite tools like Photoshop or Lightroom. No upload or download needed, if you run it at home. Easy, isn't it?*

Installation
------------

Before you start, make sure you got [Docker](https://store.docker.com/search?type=edition&offering=community) installed on your system. It is available for Mac, Linux and Windows.
Developers can skip this and move on to the [Developer Guide](https://github.com/photoprism/photoprism/wiki/Developer-Guide) in our [Wiki](https://github.com/photoprism/photoprism/wiki).

**Step 1:** Download [docker-compose.prod.yml](https://github.com/photoprism/photoprism/blob/master/docker-compose.prod.yml), rename it to `docker-compose.yml` and set the default photo path `~/Photos` to whatever directory you want to use on your local computer:

```yaml
    volumes:
        - ~/Photos:/Photos
```

PhotoPrism will create the following sub-directories in your photo path: `Import`, `Export` and `Originals`. Copy existing photos to `Import`, not directly to `Originals` as they need to be renamed and indexed in order to remove duplicates.
Files that can not be imported - like videos - will stay in the `Import` directory, nothing gets lost.

**Step 2:** Start PhotoPrism using `docker-compose` in the same directory:

```bash
docker-compose up -d
```

The Web frontend is now available at http://localhost/. The port can be changed in `docker-compose.yml` if needed. Remember to run `docker-compose restart` every time you change the config.

**Step 3:** Open a terminal to import photos:

```bash
docker-compose exec photoprism bash
photoprism import
```

*Note: This is the official way to test our development snapshot. We just started working on the UI and features are neither complete or stable. Feedback early in development helps saving a lot of time. We're a small team and need to move fast.*

Contribute
----------

We are currently setting up the infrastructure necessary to coordinate a remote team and keep everyone up-to-date.

At the moment, the best way to get in touch is to write an email to hello@photoprism.org. We'd love to hear from you!

Donations
---------

To continue working full-time and build a community, we are looking for public funding or a private sponsor who shares our vision. Any help and advice is much appreciated.

Since the software is not released yet, we don't want to ask for small donations from individuals. Please leave a star if you like this project, it provides enough motivation to keep going.

Sponsors
--------

Support this project by becoming a sponsor. Your logo will show up here with a link to your website and we can help you or your development team getting started with any of the technologies we use, either on-site or remote.
