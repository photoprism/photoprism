/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

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

import * as media from "common/media";

// see https://tools.woolyss.com/html5-canplaytype-tester/
export const useVideo = !!document.createElement("video").canPlayType;
export const useMp4Avc = useVideo // AVC
  ? !!document.createElement("video").canPlayType(media.ContentTypeMp4AvcMain)
  : false;
export const useMp4Hvc = useVideo // HEVC, Basic Support
  ? !!document.createElement("video").canPlayType(media.ContentTypeMp4HvcMain10)
  : false;
export const useMp4Hev = useVideo // HEV1, Basic Support
  ? !!document.createElement("video").canPlayType(media.ContentTypeMp4HevMain10)
  : false;
export const useMp4Vvc = useVideo // VVC, Basic Support
  ? !!document.createElement("video").canPlayType(media.ContentTypeMp4Vvc)
  : false;
export const useMp4Evc = useVideo // EVC, Basic Support
  ? !!document.createElement("video").canPlayType(media.ContentTypeMp4Evc)
  : false;
export const useWebM = useVideo // Google WebM
  ? !!document.createElement("video").canPlayType(media.ContentTypeWebm)
  : false;
export const useVP8 = useVideo // Google WebM, VP8
  ? !!document.createElement("video").canPlayType(media.ContentTypeWebmVp8)
  : false;
export const useVP9 = useVideo // Google WebM, VP9
  ? !!document.createElement("video").canPlayType(media.ContentTypeWebmVp9)
  : false;
export const useMp4Av1 = useVideo // AV1 in MP4, Main Profile 10-bit HDR
  ? !!document.createElement("video").canPlayType(media.ContentTypeMp4Av1Main10)
  : false;
export const useWebmAv1 = useVideo // AV1 in WebM, Main Profile 10-bit HDR
  ? !!document.createElement("video").canPlayType(media.ContentTypeWebmAv1Main10)
  : false;
export const useMkvAv1 = useVideo // AV1 in MKV, Main Profile 10-bit HDR
  ? !!document.createElement("video").canPlayType(media.ContentTypeMkvAv1Main10)
  : false;
export const useTheora = useVideo // Ogg Theora
  ? !!document.createElement("video").canPlayType(media.ContentTypeOgg)
  : false;
