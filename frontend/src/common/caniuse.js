/*

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

export const canUseVideo = !!document.createElement("video").canPlayType;
export const canUseAvc = canUseVideo // AVC
  ? !!document.createElement("video").canPlayType('video/mp4; codecs="avc1"')
  : false;
export const canUseOGV = canUseVideo // Ogg Theora
  ? !!document.createElement("video").canPlayType("video/ogg")
  : false;
export const canUseVP8 = canUseVideo // Google WebM, VP8
  ? !!document.createElement("video").canPlayType('video/webm; codecs="vp8"')
  : false;
export const canUseVP9 = canUseVideo // Google WebM, VP9
  ? !!document.createElement("video").canPlayType('video/webm; codecs="vp9"')
  : false;
export const canUseAv1 = canUseVideo // AV1, Main Profile
  ? !!document.createElement("video").canPlayType('video/webm; codecs="av01.0.08M.08"')
  : false;
export const canUseWebM = canUseVideo // Google WebM
  ? !!document.createElement("video").canPlayType("video/webm")
  : false;
export const canUseHevc = canUseVideo // HVC, Basic Support
  ? !!document.createElement("video").canPlayType('video/mp4; codecs="hvc1.1.6.L93.90"')
  : false;
