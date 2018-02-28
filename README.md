PhotoPrism: Long-Term Digital Photo Archiving
=============================================

[![Build Status](https://travis-ci.org/photoprism/photoprism.png?branch=master)][ci]
[![Code Quality](https://goreportcard.com/badge/github.com/photoprism/photoprism)][code quality]
[![GitHub issues](https://img.shields.io/github/issues/photoprism/photoprism.svg)][issues]
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)][license]

[ci]: https://travis-ci.org/photoprism/photoprism
[code quality]: https://travis-ci.org/photoprism/photoprism
[issues]: https://github.com/photoprism/photoprism/issues
[license]: https://github.com/photoprism/photoprism/blob/master/LICENSE

PhotoPrism is a free tool for importing, filtering and archiving large amounts of
JPEG and RAW files. Originals, thumbnails and metadata are stored in the file system for easy
backup and reliable long-term accessibility.

**Note: This software is still pre-alpha and under active development.
You're welcome to join our team.**

Our goal is to provide the following features (tested as a proof-of-concept):

- High-performance command line tool
- Web frontend
- No proprietary or binary data formats
- Duplicate detection
- Automated tagging using Google TensorFlow
- Image search with powerful filters
- Easy backup and export

Unit Tests
----------

Tests are currently not running on Travis CI as they require
docker container configuration that is still on the todo list. We
will provide a testing guide once everything is up and running.

Dependencies
------------

We are using [dep](https://github.com/golang/dep) for dependency management:

```
dep ensure
go test
```

In addition, PhotoPrism requires [darktable](https://www.darktable.org/) to convert RAW images to JPEG.
We are working on a docker container that contains it so that you don't have to install it locally.