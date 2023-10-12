Hybrid Photo/Video File Support
===============================

## Apple iPhone and iPad

[iOS Live Photos](https://developer.apple.com/live-photos/) consist of a JPEG/HEIC image and a QuickTime AVC/HEVC video, which are both required for viewing.

We recommend [using an app like PhotoSync](https://docs.photoprism.app/user-guide/sync/mobile-devices/#photosync) to upload Live Photos to PhotoPrism, since the iOS web upload usually only submits the HEIC image file without the video.

## Android Devices

Some Samsung and Google Android devices support taking "Motion Photos" with the included Camera app. Motion Photos are JPEG/HEIC image with a short MP4 video embedded after the image data.

The image part of these files can be opened in any image viewer that supports JPEG/HEIC, but the video part cannot. However, since the MP4 video is simply appended at the end of the image file, it can be easily read by our software and streamed through the API as needed.

## Introductory Tutorials

| Title                                                   | Date     | URL                                                                                |
|---------------------------------------------------------|----------|------------------------------------------------------------------------------------|
| How to detect Android motion photos in Flutter          | May 2023 | https://ente.io/blog/tech/android-motion-photos-flutter/                           |
| Stripping Embedded MP4s out of Android 12 Motion Photos | Oct 2021 | https://mjanja.ch/2021/10/stripping-embedded-mp4s-out-of-android-12-motion-photos/ |
| Google Pixel "Motion Photo" Howto                       | Mar 2021 | https://linuxreviews.org/Google_Pixel_%22Motion_Photo%22                           |
| go-mp4: Golang Library and CLI Tool for MP4             | Jul 2020 | https://dev.to/sunfishshogi/go-mp4-golang-library-and-cli-tool-for-mp4-52o1        |
| Working with Motion Photos                              | Jan 2019 | https://medium.com/android-news/working-with-motion-photos-da0aa49b50c             |
| Google: Behind the Motion Photos Technology in Pixel 2  | Mar 2018 | https://blog.research.google/2018/03/behind-motion-photos-technology-in.html       |

## Software Libraries and References

| Title                                                | URL                                                                     |
|------------------------------------------------------|-------------------------------------------------------------------------|
| Web Video Codec Guide                                | https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Video_codecs |
| Media Container Formats                              | https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Containers   |
| MP4 Signature Format                                 | https://www.file-recovery.com/mp4-signature-format.htm                  |
| List of file signatures (Wikipedia)                  | https://en.wikipedia.org/wiki/List_of_file_signatures                   |
| Go library for reading and writing MP4 files         | https://github.com/abema/go-mp4                                         |
| Go library for buffered I/O with io.Seeker interface | https://github.com/sunfish-shogi/bufseekio                              |
| How to use the io.Reader interface                   | https://yourbasic.org/golang/io-reader-interface-explained/             |
| AV1 Codec ISO Media File Format                      | https://aomediacodec.github.io/av1-isobmff                              |

## Related GitHub Issues

- https://github.com/photoprism/photoprism/issues/439 (Samsung: Initial support for Motion Photos)
- https://github.com/photoprism/photoprism/issues/1739 (Google: Initial support for Motion Photos)
- https://github.com/photoprism/photoprism/issues/2788 (Metadata: Flag Samsung/Google Motion Photos as Live Photos)
- https://github.com/cliveontoast/GoMoPho/issues/23 (Google Motion Photos Video Extractor: Add Android 12 Support)

## Related Pull Requests

- https://github.com/photoprism/photoprism/pull/3709 (Google: Initial support for Motion Photos)
- https://github.com/photoprism/photoprism/pull/3722 (Google: Add support for Motion Photos)
- https://github.com/photoprism/photoprism/pull/3660 (Samsung: Improved support for Motion Photos)

----

*PhotoPrism® is a [registered trademark](https://www.photoprism.app/trademark). By using the software and services we provide, you agree to our [Terms of Service](https://www.photoprism.app/terms), [Privacy Policy](https://www.photoprism.app/privacy), and [Code of Conduct](https://www.photoprism.app/code-of-conduct). Docs are [available](https://link.photoprism.app/github-docs) under the [CC BY-NC-SA 4.0 License](https://creativecommons.org/licenses/by-nc-sa/4.0/); [additional terms](https://github.com/photoprism/photoprism/blob/develop/assets/README.md) may apply.*
