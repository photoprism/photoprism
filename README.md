PhotoPrism: Browse Your Life in Pictures
========================================

[![License: AGPL](https://img.shields.io/badge/license-AGPL-blue.svg)][license]
[![Code Quality](https://goreportcard.com/badge/github.com/photoprism/photoprism)][goreport]
[![Build Status](https://drone.photoprism.app/api/badges/photoprism/photoprism/status.svg?ref=refs/heads/develop)][ci]
[![GitHub contributors](https://img.shields.io/github/contributors/photoprism/photoprism.svg)](https://github.com/photoprism/photoprism/graphs/contributors/)
[![Documentation](https://img.shields.io/badge/read-the%20docs-4aa087.svg)][docs]
[![Community Chat](https://img.shields.io/badge/chat-on%20gitter-4aa087.svg)][chat]
[![Twitter](https://img.shields.io/badge/follow-@photoprism_app-00acee.svg)][twitter]

PhotoPrism® is a privately hosted app for browsing, organizing, and sharing your photo collection.
It makes use of the latest technologies to tag and find pictures automatically without getting in your way.
Say goodbye to solutions that force you to upload your visual memories to the cloud!

![](https://dl.photoprism.org/assets/img/preview.jpg)

To get a first impression, you're welcome to play with our public demo at [demo.photoprism.org](https://demo.photoprism.org/).

## Key Features ##

* Our intuitive [user interface](https://demo.photoprism.org/) makes browsing and organizing your photo collection as easy as 
  it should be — whether it’s on a phone, tablet, or desktop computer.
* Index everything without worrying about duplicates or [RAW to JPEG conversion](https://docs.photoprism.org/developer-guide/library/converting/).
* Automatic [image classification](https://docs.photoprism.org/developer-guide/metadata/classification/) 
  based on Google TensorFlow. In addition, our indexer detects _colors_, _chroma_, _luminance_, _quality_, _panoramic projection_, 
  _location type_, and many other properties.
* Includes four high-resolution [world maps](https://demo.photoprism.org/places) to see where you've been,
  and for rediscovering long-forgotten shots.
* WebDAV clients, like Microsoft’s Windows Explorer or Apple's Finder, may 
  [connect](https://docs.photoprism.org/user-guide/sync/webdav/) directly to PhotoPrism so that you to can open, 
  edit, and delete files from your computer or phone as if they were local. 
  You may easily sync your pictures with [PhotoSync](https://www.photosync-app.com/) as well.  
* Because PhotoPrism is built as a [progressive web app](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps),
  it provides a native app-like experience, and you may install it on your home screen.
  There's also a [community-maintained native app in development](https://github.com/photoprism/photoprism-mobile).

## Getting Started ##

Step-by-step installation instructions for our self-hosted [community edition](https://photoprism.app/get) can be found 
on [docs.photoprism.org](https://docs.photoprism.org/getting-started/) -
all you need is a Web browser and Docker to run the server. It is available for Mac, Linux, and Windows.

We recommend hosting PhotoPrism on a server with **at least 2 cores** and **4 GB of memory**.
Beyond these minimum requirements, the amount of RAM should match the number of cores.
Indexing large photo and video collections significantly benefits from fast, local SSD storage.

## Roadmap ##

Our vision is to provide the most user-friendly solution for browsing, organizing, and sharing your photo collection.
The [roadmap](https://github.com/photoprism/photoprism/projects/5) shows what tasks are in progress, 
what needs testing, and which feature requests are going to be implemented next.

Please give ideas you like a thumbs-up 👍  , so that we know what is most popular.
Ideas backed by one or more eligible [sponsors](SPONSORS.md) will be prioritized as well.

## Contributions ##

We welcome contributions of any kind. If you have a bug or an idea, read our 
[guide](https://docs.photoprism.org/developer-guide/) before opening an issue.
Issues labeled [help wanted](https://github.com/photoprism/photoprism/labels/help%20wanted) / 
[easy](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aeasy) can be
good (first) contributions. 

Follow us on [Twitter][twitter] to receive regular project updates and discuss development related topics. Don't be afraid to ask stupid questions.

## Funding ##

You're welcome to support us via [GitHub Sponsors](https://github.com/sponsors/photoprism), 
especially if you have feature requests<sup>1</sup> or need help with using our software.
In addition, you can find us on [Patreon](https://www.patreon.com/photoprism) and 
[PayPal](https://www.paypal.me/photoprism). We've also set up [crypto wallets](SPONSORS.md).

Your continuous support helps...

* pay for operating expenses and external services like satellite maps
* developing new features and keeping them free for everyone 🌈

Also, please [leave a star](https://github.com/photoprism/photoprism/stargazers) on GitHub if you like this project. 
It provides additional motivation to keep going.

Thank you very much! <3

<sup>1</sup> Ideas backed by one or more [eligible sponsors](SPONSORS.md) are marked with a [golden label](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor).
Let us know if we mistakenly [label an idea as unfunded](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded).

## Trademarks ##

PhotoPrism® is a registered trademark of Michael Mayer. You may use it as required to describe 
our software, run your server, for educational purposes, but not for offering commercial 
goods, products, or services without prior written permission. In other words, please ask.

[wiki:classification]: https://github.com/photoprism/photoprism/wiki/Image-Classification
[wiki:xmp]: https://github.com/photoprism/photoprism/wiki/XMP
[wiki:geocoding]: https://github.com/photoprism/photoprism/wiki/Geocoding
[wiki:raw]: https://github.com/photoprism/photoprism/wiki/Converting-RAW-to-JPEG
[license]: https://github.com/photoprism/photoprism/blob/develop/LICENSE
[patreon]: https://www.patreon.com/photoprism
[paypal]: https://www.paypal.me/photoprism
[goreport]: https://goreportcard.com/report/github.com/photoprism/photoprism
[coverage]: https://codecov.io/gh/photoprism/photoprism
[ci]: https://drone.photoprism.app/photoprism/photoprism
[docs]: https://docs.photoprism.org/
[issuehunt]: https://issuehunt.io/repos/119160553
[chat]: https://gitter.im/browseyourlife/community
[twitter]: https://twitter.com/photoprism_app
[unfunded issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded
[sponsored issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor
