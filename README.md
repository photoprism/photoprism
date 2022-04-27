PhotoPrism: Browse Your Life in Pictures
========================================

[![License: AGPL](https://img.shields.io/badge/license-AGPL-blue.svg)](https://docs.photoprism.app/license/agpl/)
[![GitHub contributors](https://img.shields.io/github/contributors/photoprism/photoprism.svg)](https://photoprism.app/team)
[![Documentation](https://img.shields.io/badge/read-the%20docs-4aa087.svg)](https://docs.photoprism.app/)
[![Community Chat](https://img.shields.io/badge/chat-on%20gitter-4aa087.svg)](https://link.photoprism.app/chat)
[![GitHub Discussions](https://img.shields.io/badge/ask-%20on%20github-4d6a91.svg)](https://link.photoprism.app/discussions)
[![Twitter](https://img.shields.io/badge/follow-@photoprism_app-00acee.svg)](https://link.photoprism.app/twitter)
[![Reddit](https://img.shields.io/badge/join-/r/photoprism-EC5800.svg)](https://link.photoprism.app/reddit)

PhotoPrism® is an AI-powered app for browsing, organizing & sharing your photo collection.
It makes use of the latest technologies to tag and find pictures automatically without getting in your way.
You can run it at home, on a private server, or in the cloud.

![](https://dl.photoprism.app/img/ui/desktop-1000px.jpg)

To get a first impression, you are welcome to play with our [public demo](https://try.photoprism.app/). Be careful not to upload any private pictures.

## Feature Overview ##

* Browse [all your photos](https://docs.photoprism.app/user-guide/organize/browse/) and [videos](https://try.photoprism.app/videos) without worrying about [RAW conversion, duplicates or video formats](https://docs.photoprism.app/user-guide/settings/library/)
* Easily find specific pictures using [powerful search filters](https://try.photoprism.app/browse?view=cards&q=flower%20color%3Ared)
* Privacy-friendly: No data is ever sent to Google, Amazon, Facebook, or Apple unless you explicitly upload files to one of their services 🔐
* Recognizes [the faces of your family and friends](https://try.photoprism.app/people)
* [Automatic classification](https://try.photoprism.app/labels) of pictures based on their content and location
* [Play Live Photos](https://try.photoprism.app/live) by hovering over them in [albums](https://try.photoprism.app/albums) and [search results](https://try.photoprism.app/browse?view=cards&q=type%3Alive)
* Since the [User Interface](https://try.photoprism.app/) is a [Progressive Web App](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps),
  it provides a native app-like experience, and you can conveniently install it on the home screen of all major operating systems and mobile devices
* Includes four high-resolution [World Maps](https://try.photoprism.app/places) to bring back the memories of your favorite trips
* Metadata is extracted and merged from Exif, XMP, and other sources such as Google Photos
* Many more image properties like [Colors](https://try.photoprism.app/browse?view=cards&q=color:red), [Chroma](https://try.photoprism.app/browse?view=cards&q=mono%3Atrue), and [Quality](https://try.photoprism.app/review) can be searched as well
* Use [PhotoSync](https://link.photoprism.app/photosync) to securely backup iOS and Android phones in the background
* WebDAV clients such as Microsoft's Windows Explorer and Apple's Finder [can connect directly](https://docs.photoprism.app/user-guide/sync/webdav/) to PhotoPrism, allowing you to open, edit, and delete files from your computer as if they were local

## Getting Started ##
<img align="right" width="25%" src="https://photoprism.app/user/pages/01.home/03._screenshots/iphone-maps-hybrid-540px.png">

Step-by-step installation instructions for our self-hosted [community edition](https://photoprism.app/get) can be found 
on [docs.photoprism.app](https://docs.photoprism.app/getting-started/) -
all you need is a Web browser and [Docker](https://docs.docker.com/get-docker/) to run the server. 
It is available for Mac, Linux, and Windows.

The [stable version](https://docs.photoprism.app/release-notes/) and development 
preview have been built into a single [multi-arch image](https://link.photoprism.app/docker-hub) for 64-bit AMD, Intel,
and ARM processors. That means, [Raspberry Pi](https://docs.photoprism.app/getting-started/raspberry-pi/) 3 / 4 owners can pull 
from the same repository, enjoy the exact same functionality, and can follow the regular 
[installation instructions](https://docs.photoprism.app/getting-started/docker-compose/) 
after going through a short list of [requirements](https://docs.photoprism.app/getting-started/raspberry-pi/).

Existing users are advised to update their `docker-compose.yml` config based on our examples
available at [dl.photoprism.app/docker](https://dl.photoprism.app/docker/).

## Back us on [Patreon](https://link.photoprism.app/patreon) or [GitHub Sponsors](https://link.photoprism.app/sponsor) 💎 ##

**PhotoPrism is 100% self-funded and independent.** Your continued support helps us provide [regular updates](https://docs.photoprism.app/release-notes/)
and services like [world maps](https://try.photoprism.app/places).
Sponsors get access to [additional features](https://github.com/photoprism/photoprism/issues?q=label%3Asponsor-feature),
receive direct [technical support](https://photoprism.app/contact) via email, and can join our private chat room 
on [matrix.org](https://matrix.org/).

We currently have the following sponsorship options:

- [GitHub Sponsors](https://link.photoprism.app/sponsor) is priced in USD and also offers [one-time donations](https://link.photoprism.app/donate)
- [Patreon](https://link.photoprism.app/patreon) is priced in Euro and also offers yearly payments
- Stripe will be available in early 2022, so you can sign up directly in the app without having a Patreon or GitHub account
- You are welcome to [contact us](https://photoprism.app/contact) for [crypto donations](https://photoprism.app/crypto-donations), bank account details, and business partnerships

Also, please [leave a star](https://github.com/photoprism/photoprism/stargazers) on GitHub if you like this project.
It provides additional motivation to keep going.

## Upcoming Features and Improvements ##

**Our vision is to provide the most user- and privacy-friendly solution to keep your pictures organized and accessible.**
The [roadmap](https://link.photoprism.app/roadmap) shows what tasks are in progress, what needs testing, and which features are going to be implemented next.
Please give ideas you like a thumbs-up 👍  , so that we know what is most popular.

Ideas endorsed by [silver, gold, and platinum sponsors](https://link.photoprism.app/sponsors) receive a [golden label](https://github.com/photoprism/photoprism/issues?q=is%3Aissue+is%3Aopen+label%3Asponsor) and will be prioritized on the [roadmap](https://link.photoprism.app/roadmap).
Note that we have a zero bug policy and do our best to help users when they need support or have other questions.
This comes at a price, as we can't give exact deadlines for new features.
Our team will consider all requests, but is not obligated to implement the features, improvements, or other changes you request.

Having said that, funding really has the highest impact. So users can do their part and
[become a sponsor](https://docs.photoprism.app/funding/) to get their favorite features as soon as possible.

## Getting Support ##

If you need help installing our software at home, you can [join us on Reddit](https://link.photoprism.app/reddit), ask in our [Community Chat](https://link.photoprism.app/chat), or post your question in [GitHub Discussions](https://link.photoprism.app/discussions). Common problems can be quickly diagnosed and solved using the [Troubleshooting Checklists](https://docs.photoprism.app/getting-started/troubleshooting/) in [Getting Started](https://docs.photoprism.app/getting-started/).

We'll do our best to answer all your questions. In return, we ask you to [back us](https://docs.photoprism.app/funding/) on [Patreon](https://link.photoprism.app/patreon) or [GitHub Sponsors](https://link.photoprism.app/sponsors).
Think of "free software" as in "free speech," not as in "free beer". Thank you! 💜

In exchange for their continued support, [Sponsors](https://docs.photoprism.app/funding/) are also welcome to request direct technical [support via email](mailto:sponsors@photoprism.app). Please bear with us if we are unable to get back to you immediately due to the high volume of emails and contact requests we receive.

### GitHub Issues ###

We kindly ask you not to report bugs via GitHub Issues **unless you are certain to have found a fully reproducible and previously unreported issue** that must be fixed directly in the app. Thank you for your careful consideration!

- When reporting a problem, always include the software versions you are using and other information about your environment such as [browser, browser plugins](https://docs.photoprism.app/getting-started/troubleshooting/browsers/), operating system, [storage type](https://docs.photoprism.app/getting-started/troubleshooting/performance/#storage), [memory size](https://docs.photoprism.app/getting-started/troubleshooting/performance/#memory), and [processor](https://docs.photoprism.app/getting-started/troubleshooting/performance/#server-cpu)
- Note that all issue **subscribers receive an email notification** from GitHub for each new comment, so these should only be used for sharing important information and not for personal discussions/questions
- [Contact us](https://photoprism.app/contact) or [a community member](https://link.photoprism.app/discussions) if you need help, it could be a local configuration problem, or a misunderstanding in how the software works
- This gives our team the opportunity to [improve the docs](https://docs.photoprism.app/getting-started/troubleshooting/) and provide best-in-class support to you, instead of handling unclear/duplicate bug reports or triggering a flood of notifications by responding to comments

## Join the Community ##

Follow us on [Twitter](https://link.photoprism.app/twitter) and join the [Community Chat](https://link.photoprism.app/chat)
to get regular updates, connect with other users, and discuss your ideas.
Our [Code of Conduct](https://photoprism.app/code-of-conduct) explains the "dos and don’ts."

An important part of our journey is to explore new directions in product development and build better software through consistent use of feedback.

Feel free to [contact us](https://photoprism.app/contact) with anything that is on your mind, even if you just want to say hello. It is our goal to answer all questions we receive from our users, customers, partners, and the developer community. Our team also welcomes feedback if you believe something can or should be improved.

## Every Contribution Makes a Difference ##

We welcome [contributions](CONTRIBUTING.md) of any kind, including blog posts, tutorials, testing, writing documentation, and pull requests. Our [Developer Guide](https://docs.photoprism.app/developer-guide/) contains all the information necessary for you to get started.

----

*PhotoPrism® is a [registered trademark](https://photoprism.app/trademark). Docs are [available](https://link.photoprism.app/github-docs) under the [CC BY-NC-SA 4.0 License](https://creativecommons.org/licenses/by-nc-sa/4.0/); [additional terms](https://github.com/photoprism/photoprism/blob/develop/assets/README.md) may apply. By using the software and services we provide, you agree to our [Terms of Service](https://photoprism.app/terms), [Privacy Policy](https://photoprism.app/privacy), and [Code of Conduct](https://photoprism.app/code-of-conduct).*
