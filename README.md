PhotoPrism: Browse Your Life in Pictures
========================================

[![License: AGPL](https://img.shields.io/badge/license-AGPL-blue.svg)][license]
[![Code Quality](https://goreportcard.com/badge/github.com/photoprism/photoprism)][goreport]
[![Build Status](https://drone.photoprism.app/api/badges/photoprism/photoprism/status.svg?ref=refs/heads/develop)][ci]
[![GitHub contributors](https://img.shields.io/github/contributors/photoprism/photoprism.svg)](https://github.com/photoprism/photoprism/graphs/contributors/)
[![Documentation](https://img.shields.io/badge/read-the%20docs-4aa087.svg)][docs]
[![Community Chat](https://img.shields.io/badge/chat-on%20gitter-4aa087.svg)][chat]
[![Twitter](https://img.shields.io/badge/follow-@browseyourlife-00acee.svg)][twitter]

PhotoPrism® is a server-based application for browsing, organizing and sharing your personal photo collection.
It makes use of the latest technologies to automatically tag and find pictures without getting in your way.
Say goodbye to solutions that force you to upload your visual memories to the cloud!

![](https://dl.photoprism.org/assets/img/preview.jpg)

## Key Features ##

* Our clearly structured [user interface](https://demo.photoprism.org/) makes organizing and sharing your personal photo collection as easy as 
  it should be — whether it’s on a phone, tablet, or desktop computer.
* Index everything without worrying about duplicates or [RAW to JPEG conversion](https://docs.photoprism.org/developer-guide/library/converting/).
* Automatic [image classification](https://docs.photoprism.org/developer-guide/metadata/classification/) 
  based on Google TensorFlow. In addition, our indexer detects _colors_, _chroma_, _luminance_, _projection_, 
  _location type_, _quality_, and many other properties.
* Includes four high-resolution [world maps](https://demo.photoprism.org/places) to see where you've been,
  and for rediscovering long-forgotten shots.
* WebDAV clients, like Microsoft’s Windows Explorer or Apple's Finder, may 
  [connect](https://docs.photoprism.org/user-guide/backup/webdav/) directly to PhotoPrism so that you to can open, 
  edit, and delete files from your computer or phone as if they were local. 
  You may easily sync your pictures with [Nextcloud](https://nextcloud.com/) as well.  
* PhotoPrism feels [like a native app](https://en.wikipedia.org/wiki/Progressive_web_application) and, 
  of course, you can also add it to your home screen.

## Getting Started ##

Step-by-step installation instructions for our self-hosted [community edition](https://photoprism.app/get) can be found 
on [docs.photoprism.org](https://docs.photoprism.org/getting-started/) -
all you need is a Web browser and Docker to run the server. It is available for Mac, Linux and Windows.

We do our best to provide a complete, stable version very soon. Check the 
[roadmap](https://github.com/photoprism/photoprism/projects/5) for open issues.
Leave your email to get a [release notification](https://goo.gl/forms/KBPVGl9PCsOKrAv33).

## Try PhotoPrism ##

You're welcome to play with our public demo at [demo.photoprism.org](https://demo.photoprism.org/).

## Our Vision ##

We strive to provide the most user-friendly solution for browsing, organizing and sharing your personal photo collection.

## Feature Requests ##

Our [roadmap](https://github.com/photoprism/photoprism/projects/5) shows what tasks are in progress, 
what needs testing, and which feature requests are going to be implemented next.

Please give ideas you like a thumbs-up 👍  , so that we know what is most popular.
Ideas backed by one or more [sponsors](SPONSORS.md) will be prioritized as well.

## Contributions ##

We welcome contributions of any kind. If you have a bug or an idea, read our 
[guide](https://docs.photoprism.org/developer-guide/) before opening an issue.
Issues labeled [help wanted](https://github.com/photoprism/photoprism/labels/help%20wanted) / 
[easy](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aeasy) can be
good (first) contributions. 

Follow us on [Twitter][twitter] to receive regular project updates and discuss development 
related topics. Don't be afraid to ask stupid questions.

## Funding ##

This project is about freedom and privacy but not necessarily about free beer. We feel like it
would be a mistake to state there will be no costs, because clearly we have huge expenses, your server hardware
will have a price tag and then maybe you'd like to have some extra features that need to be developed.

It's fair to say that users with basic needs will have no monthly costs. We were also way more effective 
per dollar than investor-backed startups and learned a lot on top of it.
An earlier version of this document contained a rough number, but at the end of the day it doesn't matter.

Most established OSS companies make the bulk of their revenue with enterprise customers, that's why private users 
and single developers typically get everything for free. Obviously that doesn't work if you have only private users
that refuse to share their data on top of it.

Looking forward, specific solutions for funding development and maintenance could be to...

  - provide additional benefits to our sponsors
  - sell a tested & supported version in the app store while our contributors and other developers can 
    continue to use Docker or build from source
  - develop a one-click solution for private cloud hosting together with selected providers we trust
  - offer a geodata, public events, and maps subscription since OpenStreetMap doesn't want us to use their development 
    API for production, which is perfectly fine

## Donations ##

You're most welcome to support us via [GitHub Sponsors](https://github.com/sponsors/lastzero), 
especially if you have feature requests or need help with using our software.
In addition, you can find us on [Patreon](https://www.patreon.com/photoprism) 
and [PayPal](https://www.paypal.me/browseyourlife). 

Your continuous support helps...

* to pay for hosting, hardware and external services like satellite maps
* developing new features, and keeping them free for everyone 🌈

Also, please [leave a star](https://github.com/photoprism/photoprism/stargazers) on GitHub if you like this project. 
It provides additional motivation to keep going.

Ideas backed by one or more eligible sponsors are marked with a golden [sponsor](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor) label.
Let us know if we mistakenly label an idea as [unfunded](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded).

Thank you very much! <3

## Trademarks ##

PhotoPrism® is a registered trademark of Michael Mayer. You may use it as required to describe 
our software, run your own server, for educational purposes, but not for offering commercial 
goods, products, or services without prior written permission. In other words, please ask.

[wiki:classification]: https://github.com/photoprism/photoprism/wiki/Image-Classification
[wiki:xmp]: https://github.com/photoprism/photoprism/wiki/XMP
[wiki:geocoding]: https://github.com/photoprism/photoprism/wiki/Geocoding
[wiki:raw]: https://github.com/photoprism/photoprism/wiki/Converting-RAW-to-JPEG
[license]: https://github.com/photoprism/photoprism/blob/develop/LICENSE
[patreon]: https://www.patreon.com/photoprism
[paypal]: https://www.paypal.me/browseyourlife
[goreport]: https://goreportcard.com/report/github.com/photoprism/photoprism
[coverage]: https://codecov.io/gh/photoprism/photoprism
[ci]: https://drone.photoprism.app/photoprism/photoprism
[docs]: https://docs.photoprism.org/
[issuehunt]: https://issuehunt.io/repos/119160553
[chat]: https://gitter.im/browseyourlife/community
[twitter]: https://twitter.com/browseyourlife
[unfunded issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded
[sponsored issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor
