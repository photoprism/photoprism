# Contributing to PhotoPrism

We welcome contributions to PhotoPrism of any kind including documentation,
organization, tutorials, blog posts, bug reports, issues, feature requests,
feature implementations, pull requests, answering questions on the forum,
helping to manage issues, etc.

*Note that this repository only contains the actual source code of PhotoPrism. For **only** documentation-related pull requests / issues please refer to the [photoprism-docs](https://github.com/photoprism/photoprism-docs) repository.*

## Table of Contents

* [Asking Support Questions](#asking-support-questions)
* [Reporting Issues](#reporting-issues)
* [Submitting Patches](#submitting-patches)
  * [Code Contribution Guidelines](#code-contribution-guidelines)
  * [Git Commit Message Guidelines](#git-commit-message-guidelines)
  * [Fetching the Sources From GitHub](#fetching-the-sources-from-github)

## Asking Support Questions

The best way to get in touch is to write an email to hello@photoprism.org or join our [Telegram](https://t.me/joinchat/B8AmeBAUEugGszzuklsj5w) group.

You can also find us on GitHub, Twitter, Instagram, and LinkedIn.

Please don't use the GitHub issue tracker to ask questions.

## Reporting Issues

If you believe you have found a defect in PhotoPrism or its documentation, use
the GitHub [issue tracker](https://github.com/photoprism/photoprism/issues) to report
the problem to the maintainers. If you're not sure if it's a bug or not,
start by asking via email.
When reporting the issue, please provide the version of PhotoPrism in use (`photoprism -v`) and information about your environment.

## Submitting Patches

The PhotoPrism project welcomes all contributors and contributions regardless of skill or experience level. If you are interested in helping with the project, we will help you with your contribution.

### Code Contribution Guidelines

Because we want to create the best possible product for our users and the best contribution experience for our developers, we have a set of guidelines which ensure that all contributions are acceptable. The guidelines are not intended as a filter or barrier to participation. If you are unfamiliar with the contribution process, the PhotoPrism team will help you and teach you how to bring your contribution in accordance with the guidelines.

To make the contribution process as seamless as possible, we ask for the following:

* Go ahead and fork the project and make your changes. We encourage pull requests to allow for review and discussion of code changes.
* When you’re ready to create a pull request, be sure to:
    * Sign the [CLA](https://cla-assistant.io/photoprism/photoprism).
    * Have test cases for the new code. If you have questions about how to do this, please ask in your pull request.
    * Run `go fmt`.
    * Add documentation if you are adding new features or changing functionality.  The docs site lives at `docs.photoprism.org`.
    * Follow the **Git Commit Message Guidelines** below.

### Git Commit Message Guidelines

If your commit references one or more GitHub issues, always end your commit message body with *See #1234* or *Fixes #1234*.
Replace *1234* with the GitHub issue ID. The last example will close the issue when the commit is merged into *master*.

Please use a short and descriptive branch name, e.g. **NOT** "patch-1". It's very common but creates a naming conflict each time when a submission is pulled for a review.

###  Fetching the Sources From GitHub

PhotoPrism uses the Go Modules support built into Go 1.11 to build. The easiest is to build it inside the Docker container. See [Developer Guide](https://github.com/photoprism/photoprism/wiki/Developer-Guide).