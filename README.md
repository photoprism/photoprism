PhotoPrism: Browse Your Life in Pictures
========================================

[![License: AGPL](https://img.shields.io/badge/license-AGPL-blue.svg)][license]
![Build Status](https://dl.photoprism.org/img/badges/badge-build-success.svg)
[![GitHub contributors](https://img.shields.io/github/contributors/photoprism/photoprism.svg)](https://github.com/photoprism/photoprism/graphs/contributors/)
[![Documentation](https://img.shields.io/badge/read-the%20docs-4aa087.svg)][docs]
[![Community Chat](https://img.shields.io/badge/chat-on%20gitter-4aa087.svg)][chat]
[![GitHub Discussions](https://img.shields.io/badge/ask-%20on%20github-4d6a91.svg)][ask]
[![Twitter](https://img.shields.io/badge/follow-@photoprism_app-00acee.svg)][twitter]

PhotoPrism® is an AI-powered app for browsing, organizing & sharing your photo collection.
It makes use of the latest technologies to tag and find pictures automatically without getting in your way.
You can run it at home, on a private server, or in the cloud.

![](https://dl.photoprism.org/img/ui/desktop-1000px.jpg)

To get a first impression, you're welcome to play with our public demo at [demo.photoprism.org](https://demo.photoprism.org/).

## Feature Overview ##

* Browse all your photos and [videos](https://demo.photoprism.org/videos) without worrying about RAW conversion, duplicates or video formats
* Easily find specific pictures using powerful [search filters](https://demo.photoprism.org/browse?view=cards&q=flower%20color%3Ared)
* Privacy-friendly: No data is ever sent to Google, Amazon, Facebook, or Apple unless you explicitly upload files to one of their services 🔐
* Recognizes the faces of your [family and friends](https://demo.photoprism.org/people) 💎
* [Automatic classification](https://demo.photoprism.org/labels) of pictures based on their content and location 🌈
* Play Live Photos™ by hovering over them in [albums](https://demo.photoprism.org/albums) and [search results](https://demo.photoprism.org/browse?view=cards&q=type%3Alive)
* Since the [user interface](https://demo.photoprism.org/) is a [Progressive Web App](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps),
  it provides a native app-like experience, and you can conveniently install it on the home screen of all major operating systems and mobile devices
* Includes four high-resolution [world maps](https://demo.photoprism.org/places) to bring back the memories of your favorite trips 🌍
* Metadata is extracted and merged from Exif, XMP, and other sources such as Google Photos
* Many more image properties like [colors](https://demo.photoprism.org/browse?view=cards&q=color:red), [chroma](https://demo.photoprism.org/browse?view=cards&q=mono%3Atrue), and [quality](https://demo.photoprism.org/review) can be searched as well
* Use [PhotoSync](https://www.photosync-app.com/) to securely backup iOS and Android phones in the background 📡
* WebDAV clients such as Microsoft's Windows Explorer and Apple's Finder can [connect](https://docs.photoprism.org/user-guide/sync/webdav/) directly to PhotoPrism, allowing you to open, edit, and delete files from your computer as if they were local

## Getting Started ##

Step-by-step installation instructions for our self-hosted [community edition](https://photoprism.app/get) can be found 
on [docs.photoprism.org](https://docs.photoprism.org/getting-started/) -
all you need is a Web browser and [Docker](https://docs.docker.com/get-docker/) to run the server. 
It is available for Mac, Linux, and Windows.

Our [latest release](https://docs.photoprism.org/release-notes/) not only includes 
**facial recognition**, it also comes as a 
**single [multi-arch image](https://hub.docker.com/r/photoprism/photoprism) for AMD64, ARM64, and ARMv7**. 
That means you don't need to pull from different Docker repositories anymore.
We recommend updating your existing `docker-compose.yml` config based on 
[our examples](https://dl.photoprism.org/docker/).

## Back us on [Patreon](https://www.patreon.com/photoprism) or [GitHub Sponsors](https://github.com/sponsors/photoprism) ##

Your continued support helps us provide services like satellite maps and develop new features.
GitHub now also offers [one-time donations](https://github.com/sponsors/photoprism?frequency=one-time).
You're welcome to [contact us](https://photoprism.app/contact) for other options.

Also, please [leave a star](https://github.com/photoprism/photoprism/stargazers) on GitHub if you like this project. 
It provides additional motivation to keep going.

Thank you very much! 💜

## Roadmap ##

Our vision is to provide the most user- and privacy-friendly solution to keep your pictures organized and accessible.
The [roadmap](https://github.com/photoprism/photoprism/projects/5) shows what tasks are in progress, 
what needs testing, and which features are going to be implemented next.

Please give ideas you like a thumbs-up 👍  , so that we know what is most popular.
You are welcome to submit specific feature requests via [GitHub Issues](https://github.com/photoprism/photoprism/issues)
if you have verified that no similar [idea](https://github.com/photoprism/photoprism/labels/idea) or
[todo](https://github.com/photoprism/photoprism/labels/todo) already exists.

Ideas endorsed by silver, [gold and platinum sponsors](SPONSORS.md) receive a [golden label](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor) and will be prioritized.

## Questions?

Follow us on [Twitter](https://twitter.com/photoprism_app) or join our [Community Chat](https://gitter.im/browseyourlife/community)
to get regular updates, connect with other users, and discuss your ideas. Don't be afraid to ask silly questions.

## Getting Support ##

Please use [GitHub Issues](https://github.com/photoprism/photoprism/issues) only to report clearly identified bugs to us. If you are not sure, first use [GitHub Discussions](https://github.com/photoprism/photoprism/discussions) 
or ask in our [Community Chat](https://gitter.im/browseyourlife/community). [Sponsors](https://docs.photoprism.org/funding/) 
receive direct [technical support](https://photoprism.app/contact) via email.

When reporting a problem, please include the version you are using and information
about your environment such as browser, operating system, installed memory, and
processor type.

## Contributions ##

We welcome contributions of any kind, including bug reports, testing, writing documentation, 
tutorials, blog posts, and pull requests. Issues labeled 
[help wanted](https://github.com/photoprism/photoprism/labels/help%20wanted) / 
[easy](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aeasy) can be
good (first) contributions. Our step-by-step guide explains how to submit 
[pull requests](https://docs.photoprism.org/developer-guide/pull-requests/).

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
[ask]: https://github.com/photoprism/photoprism/discussions
[twitter]: https://twitter.com/photoprism_app
[unfunded issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded
[sponsored issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor
