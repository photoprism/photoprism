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

import { $config } from "app/session";
import sanitizeHtml from "sanitize-html";
import { DateTime } from "luxon";
import { $gettext } from "common/gettext";
import $notify from "common/notify";
import * as media from "common/media";
import * as can from "common/can";
import * as formats from "options/formats";

const Nanosecond = 1;
const Microsecond = 1000 * Nanosecond;
const Millisecond = 1000 * Microsecond;
const Second = 1000 * Millisecond;
const Minute = 60 * Second;
const Hour = 60 * Minute;
let start = new Date();

// List of characters used in the values returned by generateToken.
const tokenAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789";
export const tokenRegexp = /^[a-z0-9]{7}$/;
export const tokenLength = 7;

// True if debug logs should be created.
const debug = window.__CONFIG__?.debug || window.__CONFIG__?.trace;

export default class $util {
  static formatBytes(b) {
    if (!b) {
      return "0 KB";
    }

    if (typeof b === "string") {
      b = Number.parseFloat(b);
    }

    if (b >= 1073741824) {
      const gb = b / 1073741824;
      return gb.toFixed(1) + " GB";
    } else if (b >= 1048576) {
      const mb = b / 1048576;
      return mb.toFixed(1) + " MB";
    }

    return Math.ceil(b / 1024) + " KB";
  }

  static gigaBytes(b) {
    if (!b) {
      return 0;
    }

    if (typeof b === "string") {
      b = Number.parseFloat(b);
    }

    return Math.round(b / 1073741824);
  }

  static formatDate(s, format, zone) {
    if (!s || !s.length) {
      return s;
    }

    const l = s.length;

    if (l !== 20 || s[l - 1] !== "Z") {
      return s;
    }

    let options;

    switch (format) {
      case "date_med":
      case "DATE_MED":
        options = formats.DATE_MED;
        break;
      case "datetime_med":
      case "DATETIME_MED":
        options = formats.DATETIME_MED;
        break;
      case "date_med_tz":
      case "DATE_MED_TZ":
      case "datetime_med_tz":
      case "DATETIME_MED_TZ":
        options = formats.DATETIME_MED_TZ;
        break;
      case "date_full":
      case "DATE_FULL":
      case "datetime_full":
      case "DATETIME_FULL":
        options = formats.DATETIME_LONG;
        break;
      case "date_full_tz":
      case "datetime_full_tz":
      case "DATE_FULL_TZ":
      case "DATETIME_FULL_TZ":
        options = formats.DATETIME_LONG_TZ;
        break;
      default:
        options = formats.DATETIME_LONG;
        break;
    }

    if (!zone) {
      zone = "UTC";
    }

    return DateTime.fromISO(s, { zone }).toLocaleString(options);
  }

  static formatDuration(d) {
    let u = d;

    let neg = d < 0;

    if (neg) {
      u = -u;
    }

    if (u < Second) {
      // Special case: if duration is smaller than a second,
      // use smaller units, like 1.2ms
      if (!u) {
        return "0s";
      }

      if (u < Microsecond) {
        return u + "ns";
      }

      if (u < Millisecond) {
        return Math.round(u / Microsecond) + "µs";
      }

      return Math.round(u / Millisecond) + "ms";
    }

    let result = [];

    let h = Math.floor(u / Hour);
    let min = Math.floor(u / Minute) % 60;
    let sec = Math.ceil(u / Second) % 60;

    if (h && h > 0) {
      result.push(h.toString());
      result.push(min.toString().padStart(2, "0"));
    } else {
      result.push(min.toString());
    }

    result.push(sec.toString().padStart(2, "0"));

    // return `${h}h${min}m${sec}s`

    return result.join(":");
  }

  static formatSeconds(time) {
    if (!time || time < 0) {
      return "0:00";
    }

    let sec = time % 60;
    let min = Math.floor((time - sec) / 60);

    return `${min.toString()}:${sec.toString().padStart(2, "0")}`;
  }

  static formatRemainingSeconds(time, duration) {
    if (!duration || (time && time >= duration - 0.00001)) {
      return "0:00";
    } else if (!time || time < 0) {
      return this.formatSeconds(Math.ceil(duration));
    }

    return this.formatSeconds(Math.ceil(duration - Math.floor(time)));
  }

  static formatNs(d) {
    if (!d || Number.isNaN(d)) {
      return "";
    }

    const ms = Math.round(d / 1000000).toLocaleString();

    return `${ms} ms`;
  }

  static formatFPS(fps) {
    return `${fps.toFixed(1)} FPS`;
  }

  static arabicToRoman(number) {
    let roman = "";
    const romanNumList = {
      M: 1000,
      CM: 900,
      D: 500,
      CD: 400,
      C: 100,
      XC: 90,
      L: 50,
      XV: 40,
      X: 10,
      IX: 9,
      V: 5,
      IV: 4,
      I: 1,
    };
    let a;
    if (number < 1 || number > 3999) return "";
    else {
      for (let key in romanNumList) {
        a = Math.floor(number / romanNumList[key]);
        if (a >= 0) {
          for (let i = 0; i < a; i++) {
            roman += key;
          }
        }
        number = number % romanNumList[key];
      }
    }

    return roman;
  }

  static truncate(str, length, ending) {
    if (length == null) {
      length = 100;
    }
    if (ending == null) {
      ending = "…";
    }
    if (str.length > length) {
      return str.substring(0, length - ending.length) + ending;
    } else {
      return str;
    }
  }

  static sanitizeHtml(html) {
    if (!html) {
      return "";
    }

    return sanitizeHtml(html);
  }

  static encodeHTML(text) {
    const linkRegex = /(https?:\/\/)[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&;/=]*)/g;

    function linkFunc(matched) {
      if (!matched) {
        return "";
      }

      // Strip query parameters for added security and shorter links.
      matched = matched.split("?")[0];

      // Ampersand characters (&) should generally be ok in the link URL (though it should already be stripped as it may only be part of the query).
      let url = matched.replace(/&amp;/g, "&");

      // Make sure the URL starts with "http://" or "https://".
      if (!url.startsWith("https")) {
        url = "https://" + matched;
      }

      // Return HTML link markup.
      return `<a href="${url}" target="_blank">${matched}</a>`;
    }

    // Escape HTML control characters.
    text = text
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&apos;");

    // Make URLs clickable.
    text = text.replace(linkRegex, linkFunc);

    return text;
  }

  static resetTimer() {
    start = new Date();
  }

  static logTime(label) {
    const now = new Date();
    console.log(`${label}: ${now.getTime() - start.getTime()}ms`);
    start = now;
  }

  static capitalize(s) {
    if (!s || s === "") {
      return "";
    }

    return s.replace(/\w\S*/g, (w) => w.replace(/^\w/, (c) => c.toUpperCase()));
  }

  static ucFirst(s) {
    if (!s || s === "") {
      return "";
    }

    return s.charAt(0).toUpperCase() + s.slice(1);
  }

  // Generates a random token that makes some effort to be relatively
  // unique each time. This isn't suitable for use in
  // security-critical locations where predictability or collisions
  // would cause a serious problem.
  static generateToken() {
    let result = "";
    for (let i = 0; i < tokenLength; i++) {
      result += tokenAlphabet.charAt(Math.floor(Math.random() * tokenAlphabet.length));
    }
    return result;
  }

  static hasTouch() {
    if (!navigator.maxTouchPoints) {
      return false;
    }

    return navigator.maxTouchPoints > 0;
  }

  static isMobile() {
    return (
      /Android|webOS|iPhone|iPad|iPod|BlackBerry|Mobile|IEMobile|Opera Mini/i.test(navigator.userAgent) ||
      (navigator.maxTouchPoints && navigator.maxTouchPoints > 2)
    );
  }

  static isHttps() {
    return window.location.protocol === "https:";
  }

  static fileType(value) {
    if (!value || typeof value !== "string") {
      return "";
    }

    switch (value) {
      case "pdf":
        return "PDF";
      case "jpg":
        return "JPEG";
      case media.FormatJpegXL:
        return "JPEG XL";
      case "raw":
        return "Unprocessed Sensor Data (RAW)";
      case "mov":
      case "qt":
      case "qt  ":
        return "Apple QuickTime";
      case "bmp":
        return "Bitmap";
      case "png":
        return "Portable Network Graphics";
      case "apng":
        return "Animated PNG";
      case "tiff":
        return "TIFF";
      case "psd":
        return "Adobe Photoshop";
      case "gif":
        return "GIF";
      case "dng":
        return "Adobe Digital Negative";
      case media.CodecAvc1:
      case media.FormatAvc:
        return "Advanced Video Coding (AVC) / H.264";
      case media.CodecAvc3:
        return "Advanced Video Coding (AVC) Bitstream";
      case "avif":
        return "AOMedia Video 1 (AV1)";
      case "avifs":
        return "AVIF Image Sequence";
      case "hev":
      case "hvc":
      case media.CodecHvc1:
      case media.FormatHvc:
        return "High Efficiency Video Coding (HEVC) / H.265";
      case media.CodecHev1:
      case media.FormatHev:
        return "High Efficiency Video Coding (HEVC) Bitstream";
      case media.FormatEvc:
      case media.CodecEvc1:
        return "Essential Video Coding (MPEG-5 Part 1)";
      case "m4v":
        return "Apple iTunes Multimedia Container";
      case "mkv":
        return "Matroska Multimedia Container";
      case "mts":
        return "Advanced Video Coding High Definition (AVCHD)";
      case "m2t":
      case "m2ts":
        return "MPEG-2 Transport Stream (M2TS)";
      case "webp":
        return "Google WebP";
      case media.FormatWebm:
        return "Google WebM";
      case media.CodecVp08:
      case media.FormatVp8:
        return "Google VP8";
      case media.CodecVp09:
      case media.FormatVp9:
        return "Google VP9";
      case "flv":
        return "Flash";
      case "mpg":
        return "MPEG";
      case "mjpg":
        return "Motion JPEG";
      case "ogg":
      case "ogv":
        return "Ogg Media";
      case "wmv":
        return "Windows Media";
      case "svg":
        return "SVG";
      case "ai":
        return "Adobe Illustrator";
      case "ps":
        return "Adobe PostScript";
      case "eps":
        return "EPS";
      default:
        return value.toUpperCase();
    }
  }

  static formatCamera(camera, cameraID, cameraMake, cameraModel, long) {
    if (camera) {
      if (!long && camera.Model.length > 7) {
        // Return only the model name if it is longer than 7 characters.
        return camera.Model;
      } else {
        // Return the full camera name with make and model.
        return camera.Make + " " + camera.Model;
      }
    } else if (cameraMake && cameraModel) {
      if (!long && cameraModel.length > 7) {
        // Return only the model name if it is longer than 7 characters.
        return cameraModel;
      } else {
        // Return the full camera name with make and model.
        return cameraMake + " " + cameraModel;
      }
    } else if (cameraID > 1 && cameraModel) {
      return cameraModel;
    } else if (cameraID > 1 && cameraMake) {
      return cameraMake;
    }

    // Return a placeholder string for unknown cameras.
    if (long) {
      return $gettext("Unknown");
    }

    return "";
  }

  static formatCodec(codec) {
    if (!codec) {
      return "";
    }

    switch (codec) {
      case media.CodecAv1C:
      case media.CodecAv1:
        return "AV1";
      case media.CodecAvc1:
      case media.CodecAvc3:
      case media.CodecAvc4:
      case media.FormatAvc:
        return "AVC";
      case "hvc":
      case media.CodecHev1:
      case media.FormatHev:
      case media.CodecHvc1:
      case media.FormatHvc:
        return "HEVC";
      case media.CodecVvc1:
      case media.FormatVvc:
        return "VVC";
      case media.CodecEvc1:
      case media.FormatEvc:
        return "EVC";
      case media.FormatWebm:
        return "WebM";
      case media.CodecVp08:
      case media.FormatVp8:
        return "VP8";
      case media.CodecVp09:
      case media.FormatVp9:
        return "VP9";
      case media.FormatM2TS:
        return "M2TS";
      case "extended webp":
      case media.FormatWebp:
        return "WebP";
      default:
        return codec.toUpperCase();
    }
  }

  static codecName(value) {
    if (!value || typeof value !== "string") {
      return "";
    }

    switch (value) {
      case "raw":
        return "Unprocessed Sensor Data (RAW)";
      case "mov":
      case "qt":
      case "qt  ":
        return "Apple QuickTime (MOV)";
      case "avc":
      case media.CodecAvc1:
        return "Advanced Video Coding (AVC) / H.264";
      case media.CodecAvc3:
        return "Advanced Video Coding (AVC) Bitstream";
      case "hvc":
      case "hev":
      case media.CodecHvc1:
      case media.FormatHvc:
        return "High Efficiency Video Coding (HEVC) / H.265";
      case media.CodecHev1:
      case media.FormatHev:
        return "High Efficiency Video Coding (HEVC) Bitstream";
      case media.FormatVvc:
      case media.CodecVvc1:
        return "Versatile Video Coding (VVC) / H.266";
      case media.FormatEvc:
      case media.CodecEvc1:
        return "Essential Video Coding (MPEG-5 Part 1)";
      case "av1":
      case "av1c":
      case "av1C":
      case "av01":
        return "AOMedia Video 1 (AV1)";
      case "gif":
        return "Graphics Interchange Format (GIF)";
      case "mkv":
        return "Matroska Multimedia Container (MKV)";
      case "webp":
        return "Google WebP";
      case "extended webp":
        return "Extended WebP";
      case "webm":
        return "Google WebM";
      case "m2t":
      case "m2ts":
        return "MPEG-2 Transport Stream (M2TS)";
      case "mpeg":
        return "Moving Picture Experts Group (MPEG)";
      case "mjpg":
        return "Motion JPEG (M-JPEG)";
      case "avif":
        return "AV1 Image File Format (AVIF)";
      case "avifs":
        return "AVIF Image Sequence";
      case "heif":
        return "High Efficiency Image File Format (HEIF)";
      case "heic":
        return "High Efficiency Image Container (HEIC)";
      case "heics":
        return "HEIC Image Sequence";
      case media.FormatJpegXL:
        return "JPEG XL";
      case "1":
        return "Uncompressed";
      case "2":
        return "CCITT 1D";
      case "3":
        return "T4/Group 3 Fax";
      case "4":
        return "T6/Group 4 Fax";
      case "5":
        return "LZW";
      case "jpg":
      case "jpeg":
      case "6":
      case "7":
      case "99":
        return "JPEG";
      case "8":
        return "Adobe Deflate";
      case "9":
        return "JBIG B&W";
      case "10":
        return "JBIG Color";
      case "262":
        return "Kodak 262";
      case "32766":
        return "Next";
      case "32767":
        return "Sony ARW";
      case "32769":
        return "Packed RAW";
      case "32770":
        return "Samsung SRW";
      case "32771":
        return "CCIRLEW";
      case "32772":
        return "Samsung SRW 2";
      case "32773":
        return "PackBits";
      case "32809":
        return "Thunderscan";
      case "32867":
        return "Kodak KDC";
      case "32895":
        return "IT8CTPAD";
      case "32896":
        return "IT8LW";
      case "32897":
        return "IT8MP";
      case "32898":
        return "IT8BL";
      case "32908":
        return "PixarFilm";
      case "32909":
        return "PixarLog";
      case "32946":
        return "Deflate";
      case "32947":
        return "DCS";
      case "33003":
        return "Aperio JPEG 2000 YCbCr";
      case "33005":
        return "Aperio JPEG 2000 RGB";
      case "34661":
        return "JBIG";
      case "34676":
        return "SGILog";
      case "34677":
        return "SGILog24";
      case "34712":
        return "JPEG 2000";
      case "34713":
        return "Nikon NEF";
      case "34715":
        return "JBIG2 TIFF FX";
      case "34718":
        return "Microsoft DI Binary";
      case "34719":
        return "Microsoft DI Progressive";
      case "34720":
        return "Microsoft DI Vector";
      case "34887":
        return "ESRI Lerc";
      case "34892":
        return "Lossy JPEG";
      case "34925":
        return "LZMA2";
      case "34926":
        return "Zstd";
      case "34927":
        return "WebP";
      case "34933":
        return "PNG";
      case "34934":
        return "JPEG XR";
      case "65000":
        return "Kodak DCR";
      case "65535":
        return "Pentax PEF";
      default:
        return value.toUpperCase();
    }
  }

  // Returns the translated source name string to be displayed in the user interface.
  static sourceName(src, defaultValue) {
    switch (src) {
      case null:
      case false:
      case undefined:
      case "":
      case "auto":
        return defaultValue ? defaultValue : $gettext("Auto");
      case "default":
        return $gettext("Default");
      case "estimate":
        return $gettext("Estimate");
      case "file":
        return $gettext("File");
      case "name":
        return $gettext("Name");
      case "image":
        return $gettext("Image");
      case "location":
        return $gettext("Location");
      case "marker":
        return $gettext("Marker");
      case "caption":
        return $gettext("Caption");
      case "keyword":
        return $gettext("Keyword");
      case "meta":
        return $gettext("Metadata");
      case "subject":
        return $gettext("Subject");
      case "title":
        return $gettext("Title");
      case "xmp":
        return "XMP";
      case "batch":
        return $gettext("Batch");
      case "manual":
        return $gettext("Manual");
      case "vision":
        return $gettext("Vision");
      case "admin":
        return $gettext("Admin");
      default:
        return this.ucFirst(src);
    }
  }

  // Returns the best matching thumbnail based on the provided list of available images,
  // as well as the viewport width and height.
  static thumb(thumbs, viewportWidth, viewportHeight) {
    const sizes = $config.values.thumbs;

    if (!sizes || !thumbs || typeof thumbs !== "object") {
      return {
        src: `${$config.staticUri}/img/404.jpg`,
        w: 1280,
        h: 720,
        size: "fit_1280",
      };
    }

    for (let i = 0; i < sizes.length; i++) {
      const t = thumbs[sizes[i].size];

      if (t && (t.w >= viewportWidth || t.h >= viewportHeight)) {
        return Object.assign({}, sizes[i], t);
      }
    }

    let fallback = sizes[sizes.length - 1];

    if (!fallback?.size || !thumbs[fallback.size]) {
      fallback = sizes[0];
    }

    return Object.assign({}, fallback, thumbs[fallback.size]);
  }

  // Returns the approximate best thumbnail size based on the maximum image dimensions,
  // viewport width, and viewport height.
  static thumbSize(viewportWidth, viewportHeight) {
    const sizes = $config.values.thumbs;

    for (let i = 0; i < sizes.length; i++) {
      let t = sizes[i];

      if (t.w >= viewportWidth || t.h >= viewportHeight) {
        return t.size;
      }
    }

    const largest = sizes[sizes.length - 1];

    if (largest?.size) {
      return largest?.size;
    }

    return "fit_720";
  }

  static videoFormat(codec, mime) {
    if ((!codec && !mime) || mime?.startsWith('video/mp4; codecs="avc')) {
      return media.FormatAvc;
    } else if (can.useMp4Hvc && (codec === media.CodecHvc1 || mime?.startsWith('video/mp4; codecs="hvc'))) {
      return media.FormatHvc; // HEVC video with parameter sets not in the Samples
    } else if (can.useMp4Hev && (codec === media.CodecHev1 || mime?.startsWith('video/mp4; codecs="hev'))) {
      return media.FormatHev; // HEVC video with parameter sets also in the Samples, won't play on macOS
    } else if (can.useMp4Vvc && (codec === media.CodecVvc1 || mime?.startsWith('video/mp4; codecs="vvc'))) {
      return media.FormatVvc;
    } else if (can.useMp4Evc && (codec === media.CodecEvc1 || mime?.startsWith('video/mp4; codecs="evc'))) {
      return media.FormatEvc;
    } else if (can.useVP8 && (codec === media.CodecVp08 || mime?.startsWith('video/mp4; codecs="vp8'))) {
      return media.FormatVp8;
    } else if (can.useVP9 && (codec === media.CodecVp09 || mime?.startsWith('video/mp4; codecs="vp09'))) {
      return media.FormatVp9;
    } else if (can.useMp4Av1 && (mime?.startsWith('video/mp4; codecs="av01') || mime?.startsWith("video/AV1"))) {
      return media.FormatAv1;
    } else if (can.useWebmAv1 && mime?.startsWith('video/webm; codecs="av01')) {
      return media.FormatWebmAv1;
    } else if (can.useMkvAv1 && mime?.startsWith('video/matroska; codecs="av01')) {
      return media.FormatMkvAv1;
    } else if (can.useWebM && (codec === media.FormatWebm || mime === media.ContentTypeWebm)) {
      return media.FormatWebm;
    } else if (can.useTheora && (codec === media.CodecTheora || mime === media.ContentTypeOgg)) {
      return media.FormatTheora;
    }

    return media.FormatAvc;
  }

  static videoFormatUrl(hash, format) {
    if (!hash) {
      return "";
    }

    if (!format) {
      format = media.FormatAvc;
    }

    return `${$config.videoUri}/videos/${hash}/${$config.previewToken}/${format}`;
  }

  static videoUrl(hash, codec, mime) {
    return this.videoFormatUrl(hash, this.videoFormat(codec, mime));
  }

  static videoContentType(codec, mime) {
    switch (this.videoFormat(codec, mime)) {
      case media.FormatAvc:
        return media.ContentTypeMp4AvcMain;
      case media.FormatHvc:
        return media.ContentTypeMp4HvcMain;
      case media.FormatHev:
        return media.ContentTypeMp4HevMain;
      case media.FormatVvc:
        return media.ContentTypeMp4Vvc;
      case media.FormatVp8:
        return media.ContentTypeWebmVp8;
      case media.FormatVp9:
        return media.ContentTypeWebmVp9;
      case media.FormatWebmAv1:
        return media.ContentTypeWebmAv1Main10;
      case media.FormatMkvAv1:
        return media.ContentTypeMkvAv1Main10;
      case media.FormatWebm:
        return media.ContentTypeWebm;
      case media.FormatTheora:
        return media.ContentTypeOgg;
      default:
        return "video/mp4";
    }
  }

  static copyText(text) {
    if (!text) {
      if (debug) {
        console.warn("clipboard: missing text");
      }

      return false;
    }

    // Join additional text arguments, if any.
    for (let i = 1; i < arguments.length; i++) {
      if (typeof arguments[i] === "string" && arguments[i].length > 0) {
        text += " " + arguments[i];
      }
    }

    return this.writeToClipboard(text);
  }

  static writeToClipboard(text) {
    if (window.navigator?.clipboard && window.navigator.clipboard instanceof EventTarget) {
      window.navigator.clipboard
        .writeText(text)
        .then(() => {
          $notify.success($gettext("Copied to clipboard"));
        })
        .catch((err) => {
          if (debug && err) {
            console.error("clipboard:", err);
          }

          $notify.error($gettext("Cannot copy to clipboard"));
        });
      return true;
    } else if (debug) {
      console.warn("clipboard: window.navigator.clipboard is not an instance of EventTarget");
    }

    $notify.warn($gettext("Cannot copy to clipboard"));

    return false;
  }
}
