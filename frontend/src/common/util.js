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

const Nanosecond = 1;
const Microsecond = 1000 * Nanosecond;
const Millisecond = 1000 * Microsecond;
const Second = 1000 * Millisecond;
const Minute = 60 * Second;
const Hour = 60 * Minute;
let start = new Date();

export default class Util {
  static fps(fps) {
    return `${fps.toFixed(1)} FPS`;
  }

  static duration(d) {
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

    result.push(h.toString().padStart(2, "0"));
    result.push(min.toString().padStart(2, "0"));
    result.push(sec.toString().padStart(2, "0"));

    // return `${h}h${min}m${sec}s`

    return result.join(":");
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

  static encodeHTML(text) {
    return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#x27;");
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

  static generateToken() {
    return (Math.random() + 1).toString(36).substring(6);
  }

  static fileType(value) {
    if (!value || typeof value !== "string") {
      return "";
    }

    switch (value) {
      case "jpg":
        return "JPEG";
      case "jxl":
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
      case "avc":
      case "avc1":
        return "Advanced Video Coding (AVC) / H.264";
      case "avif":
        return "AOMedia Video 1 (AV1)";
      case "avifs":
        return "AVIF Image Sequence";
      case "hvc":
      case "hevc":
      case "hev1":
      case "hvc1":
        return "High Efficiency Video Coding (HEVC) / H.265";
      case "m4v":
        return "Apple iTunes Multimedia Container";
      case "mkv":
        return "Matroska Multimedia Container";
      case "mts":
        return "Advanced Video Coding High Definition (AVCHD)";
      case "m2ts":
        return "Blu-ray MPEG-2 Transport Stream";
      case "webp":
        return "Google WebP";
      case "webm":
        return "Google WebM";
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
        return "Scalable Vector Graphics";
      case "pdf":
        return "Portable Document Format";
      case "ai":
        return "Adobe Illustrator";
      case "ps":
        return "Adobe PostScript";
      case "eps":
        return "Encapsulated PostScript";
      default:
        return value.toUpperCase();
    }
  }

  static formatCodec(codec) {
    if (!codec) {
      return "";
    }

    switch (codec) {
      case "webp":
      case "extended webp":
        return "WebP";
      case "webm":
        return "WebM";
      case "av1c":
      case "av01":
        return "AV1";
      case "avc1":
        return "AVC";
      case "hvc":
      case "hev1":
      case "hvc1":
        return "HEVC";
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
      case "avc1":
        return "Advanced Video Coding (AVC) / H.264";
      case "hvc":
      case "hevc":
      case "hev1":
      case "hvc1":
        return "High Efficiency Video Coding (HEVC) / H.265";
      case "vvc":
        return "Versatile Video Coding (VVC) / H.266";
      case "av1c":
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

  static async copyToMachineClipboard(text) {
    if (window.navigator.clipboard) {
      await window.navigator.clipboard.writeText(text);
    } else if (document.execCommand) {
      // Clipboard is available only in HTTPS pages. see https://web.dev/async-clipboard/
      // So if the the official 'clipboard' doesn't supported and the 'document.execCommand' is supported.
      // copy by a work-around by creating a textarea in the DOM and execute copy command from him.

      // Create the text area element (to copy from)
      const clipboardElement = document.createElement("textarea");

      // Set the text content to copy
      clipboardElement.value = text;

      // Avoid scrolling to bottom
      clipboardElement.style.top = "0";
      clipboardElement.style.left = "0";
      clipboardElement.style.position = "fixed";

      // Add element to DOM
      document.body.appendChild(clipboardElement);

      // "Select" the new textarea
      clipboardElement.focus();
      clipboardElement.select();

      // Copy the selected textarea content
      const succeed = document.execCommand("copy");

      // Remove the textarea from DOM
      document.body.removeChild(clipboardElement);

      // Validate operation succeed
      if (!succeed) {
        throw new Error("Failed copying to clipboard");
      }
    } else {
      throw new Error("Copy to clipboard does not support in your browser");
    }
  }
}
