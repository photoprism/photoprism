PhotoPrism: Browse Your Life in Pictures
========================================

[![License: AGPL](https://img.shields.io/badge/license-AGPL-blue.svg)][license]
[![Code Quality](https://goreportcard.com/badge/github.com/photoprism/photoprism)][goreport]
[![Build Status](https://travis-ci.org/photoprism/photoprism.png?branch=develop)][ci]
[![GitHub contributors](https://img.shields.io/github/contributors/photoprism/photoprism.svg)](https://github.com/photoprism/photoprism/graphs/contributors/)
[![Documentation](https://img.shields.io/badge/read-the%20docs-4aa087.svg)][docs]
[![Community Chat](https://img.shields.io/badge/chat-on%20gitter-4aa087.svg)][chat]
[![Twitter](https://img.shields.io/badge/follow-@browseyourlife-00acee.svg)][twitter]

PhotoPrism™ is a server-based application for browsing, organizing and sharing your personal photo collection.
It makes use of the latest technologies to automatically tag and find pictures without getting in your way.
Say goodbye to solutions that force you to upload your visual memories to the public cloud.

![](https://dl.photoprism.org/assets/img/preview.jpg)

### What to expect... ###

* Clearly structured Web interface for browsing, organizing and sharing your personal photo collection
* Import everything without worrying about duplicates or [RAW to JPEG conversion][wiki:raw]
* [Geocoding][wiki:geocoding], [XMP support][wiki:xmp] and [automated tagging][wiki:classification] 
  based on Google TensorFlow

## For the Early Birds ##

You're welcome to play with our demo at [demo.photoprism.org](https://demo.photoprism.org).
Leave your email to get a [release notification](https://goo.gl/forms/KBPVGl9PCsOKrAv33).

Step-by-step [installation instructions](https://docs.photoprism.org/getting-started/) can be found
in our [User Guide](https://docs.photoprism.org/).
Developers can skip this and move on to the [Developer Guide](https://docs.photoprism.org/developer-guide/).

All you need is a Web browser and [Docker](https://store.docker.com/search?type=edition&offering=community)
to run the server. It is available for Mac, Linux and Windows.

There is also a [pre-installed Raspberry Pi image here](https://github.com/guysoft/PhotoPrismPi).

Note that this is work in progress. We do our best to provide a complete, stable version.
If you have a question, don't hesitate to ask in our [help forum][help] 
or [contact us via email](mailto:hello@photoprism.org).

## Why This Has to Be Free Software ##

The development of every commercial product is focused on monetization.
We've [built similar apps more than once](https://github.com/photoprism/photoprism/wiki/Mediencenter) 
and every single time the constraints of working
in a profit-oriented corporate environment were an impediment.

We are sure we can do better with only a fraction of the budget. Simplicity - the art of maximizing the 
amount of work not done - can be very powerful.
Go itself is a [great example](https://talks.golang.org/2015/simplicity-is-complicated.slide).

Our long-term goal is to become an open platform for machine 
learning [research](https://github.com/photoprism/photoprism/wiki/Research) based on real-world photo collections.
We're already in contact with data scientists who like our idea.

## Contributions ##

We welcome contributions of any kind. If you have a bug or an idea, read our 
[guide](https://docs.photoprism.org/developer-guide/) before opening an issue.
Issues labeled [help wanted](https://github.com/photoprism/photoprism/labels/help%20wanted) / 
[easy](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aeasy) can be
good (first) contributions. 

Please follow us on [Twitter][twitter] and join our [developers mailing list](https://groups.google.com/a/photoprism.org/forum/#!forum/developers) 
to receive regular project updates and discuss development related topics. Don't be afraid to ask stupid questions.

## No Free Beer ##

This project is about freedom and privacy but not necessarily about free beer. We feel like it
would be a mistake to state there will be no costs, because clearly we have huge expenses, your server hardware
will have a price tag and then maybe you'd like to have some extra features that need to be developed.

It's fair to say that users with basic needs will have no monthly costs. We were also way more effective 
per dollar than commercial projects and learned a lot on top of it.
An earlier version of this document contained a rough number, but at the end of the day it doesn't matter.

Most established OSS companies make the bulk of their revenue with enterprise customers, that's why private users 
and single developers typically get everything for free. Obviously that doesn't work if you have only private users
that refuse to share their data on top of it.

Looking forward, specific solutions for funding development and maintenance could be to...

  - sell a tested & supported version in the app store while our contributors and other developers can 
    continue to use Docker or build from source
  - provide additional features to users who support us financially, maybe with a different license similar to GitLab
  - develop a one-click solution for private cloud hosting together with selected providers we trust
  - offer a geodata, public events and maps subscription since OpenStreetMap doesn't want us to use their development 
    API for production, which is perfectly fine

## Donations ##

You're welcome to support us via [GitHub Sponsors](https://github.com/sponsors/lastzero), 
especially if you have feature requests or need help with using our software.
They will match every donation in the first year.
In addition, you can find us on [Patreon](https://www.patreon.com/photoprism) and [PayPal](https://www.paypal.me/photoprism). 
Our [sponsors](https://github.com/photoprism/photoprism/blob/develop/SPONSORS.md) and 
[contributors](https://github.com/photoprism/photoprism/graphs/contributors/) will get for free whatever we might 
have to charge for a geodata subscription and/or additional features later (see [tiers](https://github.com/sponsors/lastzero/dashboard/tiers)).

Also, please [leave a star](https://github.com/photoprism/photoprism/stargazers) on GitHub if you like this project. 
It provides additional motivation to keep going.

Ideas backed by a sponsor are marked with a golden [sponsor](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor) label.
Let us know if we mistakenly label an idea as [unfunded](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded).

Thank you very much! <3

## Trademarks ##

PhotoPrism™ is a registered trademark of Michael Mayer. You may use it as required to describe 
our software, run your own server, for educational purposes, but not for offering commercial 
goods, products, or services without prior written permission. In other words, please ask.

## Don't Panic ##

We'd like to remind everyone that we are not full-time marketing specialists but developers 
who work a lot and enjoy a bit of sarcasm from time to time. Please let us know when there is 
an issue with our "nuance and tone", and we'll find a solution.

[wiki:classification]: https://github.com/photoprism/photoprism/wiki/Image-Classification
[wiki:xmp]: https://github.com/photoprism/photoprism/wiki/XMP
[wiki:geocoding]: https://github.com/photoprism/photoprism/wiki/Geocoding
[wiki:raw]: https://github.com/photoprism/photoprism/wiki/Converting-RAW-to-JPEG
[help]: https://groups.google.com/a/photoprism.org/forum/#!forum/help
[license]: https://github.com/photoprism/photoprism/blob/develop/LICENSE
[patreon]: https://www.patreon.com/photoprism
[paypal]: https://www.paypal.me/photoprism
[goreport]: https://goreportcard.com/report/github.com/photoprism/photoprism
[coverage]: https://codecov.io/gh/photoprism/photoprism
[ci]: https://travis-ci.org/photoprism/photoprism
[docs]: https://docs.photoprism.org/
[issuehunt]: https://issuehunt.io/repos/119160553
[chat]: https://gitter.im/browseyourlife/community
[twitter]: https://twitter.com/browseyourlife
[unfunded issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded
[sponsored issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor
