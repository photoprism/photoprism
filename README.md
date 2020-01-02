PhotoPrism: Browse your life in pictures
========================================

[![License: GPL](https://img.shields.io/badge/license-GPL-blue.svg)][license]
[![Code Quality](https://goreportcard.com/badge/github.com/photoprism/photoprism)][goreport]
[![Build Status](https://travis-ci.org/photoprism/photoprism.png?branch=develop)][ci]
[![Documentation](https://readthedocs.org/projects/photoprism-docs/badge/?version=latest&style=flat)][docs]
[![GitHub contributors](https://img.shields.io/github/contributors/photoprism/photoprism.svg)](https://github.com/photoprism/photoprism/graphs/contributors/)
[![Community Chat](https://img.shields.io/badge/chat-on%20gitter-4aa087.svg)][chat]
[![Twitter](https://img.shields.io/badge/follow-@browseyourlife-00acee.svg)][twitter]

PhotoPrism is a server-based application for browsing, organizing and sharing your personal photo collection.
It makes use of the latest technologies to automatically tag and find pictures without getting in your way.
Say goodbye to solutions that force you to upload your visual memories to the cloud or pay monthly fees.

![](https://dl.photoprism.org/assets/img/preview.jpg)

More screenshots: https://photoprism.org/#screenshots

## What to expect ##

* Clearly structured Web interface for browsing, organizing and sharing your personal photo collection.
* Import everything without worrying about duplicates or RAW to JPEG conversion.
* Reverse geocoding and automated tagging based on Google TensorFlow.
* No monthly costs. No proprietary formats. No privacy concerns.

## For the early birds ##

You're welcome to play with our demo at [demo.photoprism.org](https://demo.photoprism.org).
Leave your email to get a [release notification](https://goo.gl/forms/KBPVGl9PCsOKrAv33).

Step-by-step [installation instructions](https://docs.photoprism.org/en/latest/setup/) can be found
in our [User Guide](https://docs.photoprism.org/en/latest/).
Developers can skip this and move on to the [Developer Guide](https://github.com/photoprism/photoprism/wiki).

All you need is a Web browser and [Docker](https://store.docker.com/search?type=edition&offering=community)
to run the server. It is available for Mac, Linux and Windows.

Note that this is work in progress. We do our best to provide a complete, stable version. 
Financial [support](SPONSORS.md) makes a huge difference and enables us to spend more time with this project.

If you have a question, don't hesitate to ask in our [help forum][help] 
or [contact us via email](mailto:hello@photoprism.org).

## Why this has to be free software ##

The development of every commercial product is focused on monetization.
We've [built similar apps more than once](https://github.com/photoprism/photoprism/wiki/Mediencenter) 
and every single time the constraints of working
in a profit-oriented corporate environment were an impediment.

We are sure we can do better with only a fraction of the budget. Simplicity - the art of maximizing the 
amount of work not done - can be very powerful.
Go itself is a [great example](https://talks.golang.org/2015/simplicity-is-complicated.slide).

Our long-term goal is to become an open platform for machine 
learning [research](https://github.com/photoprism/photoprism/wiki/Research) based on real-world photo collections.

## How to contribute ##

We welcome contributions of any kind. If you have a bug or an idea, read our 
[guide](https://docs.photoprism.org/en/latest/contribute/) before opening an issue.
Issues labeled [help wanted](https://github.com/photoprism/photoprism/labels/help%20wanted) / 
[easy](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aeasy) can be
good (first) contributions. 

You'll get a small reward for working on [funded issues](https://github.com/photoprism/photoprism/labels/funded), 
see [issuehunt.io](https://issuehunt.io/repos/119160553) for details. 
Note that issue descriptions may be outdated on their site. Rewards are paid out when 
all [acceptance criteria](https://github.com/photoprism/photoprism/wiki/Issues#user-stories) prioritized as 
MUST are met and your [pull request](https://github.com/photoprism/photoprism/wiki/Pull-Requests) 
was successfully merged.

Please follow us on [Twitter][twitter] and join our [developers mailing list](https://groups.google.com/a/photoprism.org/forum/#!forum/developers) 
to receive regular project updates and discuss development related topics. Don't be afraid to ask stupid questions.

## Donations ##

PhotoPrism is a non-profit project run entirely by volunteers. We need your funds to pay for 
[organizing meetups](https://github.com/photoprism/photoprism/wiki/Meetups),
[running our servers](https://github.com/photoprism/photoprism/wiki/Infrastructure),
visiting conferences, buying test devices, offering rewards for contributions, and covering our cost of living.

You're most welcome to support us via [GitHub Sponsors](https://github.com/sponsors/lastzero), 
especially if you need help with using our software. They will match every donation in the first year.
In addition, you can find us on [Patreon][patreon] and [PayPal][paypal].
Our sponsors and contributors will get for free whatever we have to charge for a geodata, 
public events and maps subscription later.

Also please leave a star if you like this project, it provides enough motivation to keep going. Thank you very much! <3

Ideas backed by a sponsor are marked with a golden [sponsor][sponsored issues] label.
Let us know if we mistakenly label an idea as [unfunded][unfunded issues].

## Public and corporate sponsorship ##

Our software is now almost done after two years of hard work, some days 16 to 20 hours on top of other projects we
did to finance this. We spent weeks asking organizations like [The Prototype Fund](https://prototypefund.de/en/) 
for help and also tried to cooperate with companies like Mapbox and Cewe. We have been ignored and even given
the advice that what we do "already exists in America".

If any of those organizations changes their mind, they are welcome to [reach out to us](mailto:hello@photoprism.org).

[help]: https://groups.google.com/a/photoprism.org/forum/#!forum/help
[license]: https://github.com/photoprism/photoprism/blob/develop/LICENSE
[patreon]: https://www.patreon.com/photoprism
[paypal]: https://www.paypal.me/photoprism
[goreport]: https://goreportcard.com/report/github.com/photoprism/photoprism
[coverage]: https://codecov.io/gh/photoprism/photoprism
[ci]: https://travis-ci.org/photoprism/photoprism
[docs]: https://docs.photoprism.org/en/latest/
[issuehunt]: https://issuehunt.io/repos/119160553
[chat]: https://gitter.im/browseyourlife/community
[twitter]: https://twitter.com/browseyourlife
[unfunded issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Aunfunded
[sponsored issues]: https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor
