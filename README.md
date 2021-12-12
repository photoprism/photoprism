PhotoPrism: Browse Your Life in Pictures
========================================

[![License: AGPL](https://img.shields.io/badge/license-AGPL-blue.svg)][license]
![Build Status](https://dl.photoprism.app/img/badges/badge-build-success.svg)
[![GitHub contributors](https://img.shields.io/github/contributors/photoprism/photoprism.svg)](https://github.com/photoprism/photoprism/graphs/contributors/)
[![Documentation](https://img.shields.io/badge/read-the%20docs-4aa087.svg)][docs]
[![Community Chat](https://img.shields.io/badge/chat-on%20gitter-4aa087.svg)][chat]
[![GitHub Discussions](https://img.shields.io/badge/ask-%20on%20github-4d6a91.svg)][ask]
[![Twitter](https://img.shields.io/badge/follow-@photoprism_app-00acee.svg)][twitter]

PhotoPrism® is an AI-powered app for browsing, organizing & sharing your photo collection.
It makes use of the latest technologies to tag and find pictures automatically without getting in your way.
You can run it at home, on a private server, or in the cloud.

![](https://dl.photoprism.app/img/ui/desktop-1000px.jpg)

To get a first impression, you're welcome to play with our public demo at [demo.photoprism.app](https://demo.photoprism.app/) 
(also available in [汉语](https://demo-zh.photoprism.app/), [Deutsch](https://demo-de.photoprism.app/), and [Français](https://demo-fr.photoprism.app/)).

## Feature Overview ##

* Browse [all your photos](https://docs.photoprism.app/user-guide/organize/browse/) and [videos](https://demo.photoprism.app/videos) without worrying about [RAW conversion, duplicates or video formats](https://docs.photoprism.app/user-guide/settings/library/)
* Easily find specific pictures using [powerful search filters](https://demo.photoprism.app/browse?view=cards&q=flower%20color%3Ared)
* Privacy-friendly: No data is ever sent to Google, Amazon, Facebook, or Apple unless you explicitly upload files to one of their services 🔐
* Recognizes [the faces of your family and friends](https://demo.photoprism.app/people)
* [Automatic classification](https://demo.photoprism.app/labels) of pictures based on their content and location
* [Play Live Photos](https://demo.photoprism.app/live) by hovering over them in [albums](https://demo.photoprism.app/albums) and [search results](https://demo.photoprism.app/browse?view=cards&q=type%3Alive)
* Since the [User Interface](https://demo.photoprism.app/) is a [Progressive Web App](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps),
  it provides a native app-like experience, and you can conveniently install it on the home screen of all major operating systems and mobile devices
* Includes four high-resolution [World Maps](https://demo.photoprism.app/places) to bring back the memories of your favorite trips
* Metadata is extracted and merged from Exif, XMP, and other sources such as Google Photos
* Many more image properties like [Colors](https://demo.photoprism.app/browse?view=cards&q=color:red), [Chroma](https://demo.photoprism.app/browse?view=cards&q=mono%3Atrue), and [Quality](https://demo.photoprism.app/review) can be searched as well
* Use [PhotoSync](https://www.photosync-app.com/) to securely backup iOS and Android phones in the background
* WebDAV clients such as Microsoft's Windows Explorer and Apple's Finder [can connect directly](https://docs.photoprism.app/user-guide/sync/webdav/) to PhotoPrism, allowing you to open, edit, and delete files from your computer as if they were local

## Getting Started ##
<img align="right" width="25%" src="https://photoprism.app/user/pages/01.home/03._screenshots/iphone-maps-hybrid-540px.png">

Step-by-step installation instructions for our self-hosted [community edition](https://photoprism.app/get) can be found 
on [docs.photoprism.app](https://docs.photoprism.app/getting-started/) -
all you need is a Web browser and [Docker](https://docs.docker.com/get-docker/) to run the server. 
It is available for Mac, Linux, and Windows.

The [stable version](https://docs.photoprism.app/release-notes/) and development 
preview have been built into a single [multi-arch image](https://hub.docker.com/r/photoprism/photoprism) for 64-bit AMD, Intel,
and ARM processors. That means, [Raspberry Pi](https://docs.photoprism.app/getting-started/raspberry-pi/) 3 / 4 owners can pull 
from the same repository, enjoy the exact same functionality, and can follow the regular 
[installation instructions](https://docs.photoprism.app/getting-started/docker-compose/) 
after going through a short list of [requirements](https://docs.photoprism.app/getting-started/raspberry-pi/).

Existing users are advised to update their `docker-compose.yml` config based on our examples
available at [dl.photoprism.app/docker](https://dl.photoprism.app/docker/).

## Back us on [Patreon](https://www.patreon.com/photoprism) or [GitHub Sponsors](https://github.com/sponsors/photoprism) ##

Your continued support helps us provide services like world maps and develop new features.
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
or ask in our [Community Chat](https://gitter.im/browseyourlife/community). [Sponsors](https://docs.photoprism.app/funding/) 
receive direct [technical support](https://photoprism.app/contact) via email.

When reporting a problem, please include the version you are using and information
about your environment such as browser, operating system, installed memory, and
processor type.

## Contributions ##

We welcome contributions of any kind, including bug reports, testing, writing documentation, 
tutorials, blog posts, and pull requests.

The [Developer Guide](https://docs.photoprism.app/developer-guide/) contains all information
necessary to get you started.
Issues labeled [help wanted](https://github.com/photoprism/photoprism/labels/help%20wanted) /
[easy](https://github.com/photoprism/photoprism/labels/easy) can be good (first) contributions.

Because we want to create the best possible product for our users,
we have a set of [guidelines](https://docs.photoprism.app/developer-guide/pull-requests)
to ensure that all source code submissions are acceptable.
They include step-by-step instructions for submitting new features,
bug fixes, and documentation improvements.

**Reviewing, testing and finally merging pull requests requires significant resources
on our side. If it's not just a small fix, it can take several months.**

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
[docs]: https://docs.photoprism.app/
[issuehunt]: https://issuehunt.io/repos/119160553
[chat]: https://gitter.im/browseyourlife/community
[ask]: https://github.com/photoprism/photoprism/discussions
[twitter]: https://twitter.com/photoprism_app
[unfunded issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded
[sponsored issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor
