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
Say goodbye to solutions that force you to upload your visual memories to the cloud.

![](https://dl.photoprism.org/assets/img/preview.jpg)

More screenshots: https://photoprism.org/#screenshots

## What to expect ##

* Clearly structured Web interface for browsing, organizing and sharing your personal photo collection
* Import everything without worrying about duplicates or RAW to JPEG conversion
* Reverse geocoding, XMP support and automated tagging based on Google TensorFlow

## For the early birds ##

You're welcome to play with our demo at [demo.photoprism.org](https://demo.photoprism.org).
Leave your email to get a [release notification](https://goo.gl/forms/KBPVGl9PCsOKrAv33).

Step-by-step [installation instructions](https://docs.photoprism.org/en/latest/setup/) can be found
in our [User Guide](https://docs.photoprism.org/en/latest/).
Developers can skip this and move on to the [Developer Guide](https://github.com/photoprism/photoprism/wiki).

All you need is a Web browser and [Docker](https://store.docker.com/search?type=edition&offering=community)
to run the server. It is available for Mac, Linux and Windows.

Note that this is work in progress. We do our best to provide a complete, stable version.
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

## Funding ##

It's clear many users are waiting for a stable release while only very few donate or help with development. 
This project is about freedom but not necessarily about free beer.

We are not Google and don't have billions of dollars on our bank accounts to give away to our fans in exchange for their data.
It's also somewhat disappointing how little support we get by companies and especially public organizations. 
Not a single dollar. 

Every politician wants to support Open Source and warns social media is bad for your privacy, but only very 
few are willing to help those that actually do something.

The consequence is that we are forced to think about monetization. We honestly didn't expect this will be an issue 
and didn't ask for anything the first year. Thank you very much to our few [sponsors](SPONSORS.md)! 
We still love each and everyone of you, even those that send multiple pages of written requirements and then ask 
every week when it is done.

Specific solutions could be to...

  - sell a tested & supported version in the app store while our contributors and other developers can 
    continue to use Docker or build from source
  - offer a geodata, public events and maps subscription since OpenStreetMap doesn't want us to use 
    their development API for production

## Donations ##

You're most welcome to support us via [GitHub Sponsors](https://github.com/sponsors/lastzero), 
especially if you need help with using our software. They will match every donation in the first year.
In addition, you can find us on [Patreon][patreon] and [PayPal][paypal].
Our [sponsors](SPONSORS.md) and [contributors](https://github.com/photoprism/photoprism/graphs/contributors/)
will get for free whatever we might have to charge for a geodata subscription later.

Also please [leave a star](https://github.com/photoprism/photoprism/stargazers) here on GitHub if you like this project, 
it provides additional motivation to keep going.

Financial support makes a huge difference and enables us to spend more time with the features you care about.
Ideas backed by a sponsor are marked with a golden [sponsor][sponsored issues] label.
Let us know if we mistakenly label an idea as [unfunded][unfunded issues].

Thank you very much! <3

## Lessons learned ##

Having done mostly commercial projects in the last 10+ years, it is important for us to explore various forms of funding
and communication for independent Open Source projects. Note that many of today's popular projects are backed by 
corporations like Google, Facebook, Microsoft or Intel.

That's a good thing and we profit from it, but doesn't mean independent developers should not do this full-time or pay 
everything themselves without asking the community for support. In fact, crowdfunding is a pretty common way to cover 
development expenses if you look at [Indiegogo](https://www.indiegogo.com/) or [Kickstarter](https://www.kickstarter.com/).

In no way do we spurn other OSS projects like [OpenStreetMap](https://www.openstreetmap.org/), 
as a Twitter user suggested. We just state the fact that even a non-commercial app can't use their API 
for production, which is perfectly fine.
On the other hand, it also doesn't mean we have to provide this service for free to our users. 

We understand money is a very sensitive topic most of our users don't engage with.
For this reason, we'll exclude this topic from our social media communication from now on.

## Public and corporate sponsorship ##

We spent weeks asking organizations like [The Prototype Fund](https://prototypefund.de/en/) for help 
and also tried to cooperate with companies like Mapbox and Cewe.

Some conversations were good without leading to a sponsorship yet. Others have given us the advice that what we 
do "already exists in America". You would think it's easier to get a few dollars with 
[our background](http://docs.photoprism.org/en/latest/team/) and a [working demo](https://demo.photoprism.org/).

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
