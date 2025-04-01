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

import memoizeOne from "memoize-one";
import RestModel from "model/rest";
import File from "model/file";
import Marker from "model/marker";
import { DateTime } from "luxon";
import { $config } from "app/session";
import $api from "common/api";
import $util from "common/util";
import countries from "options/countries.json";
import { $gettext } from "common/gettext";
import { PhotoClipboard } from "common/clipboard";
import download from "common/download";
import * as src from "common/src";
import * as media from "common/media";
import * as formats from "options/formats";

export const YearUnknown = -1;
export const MonthUnknown = -1;
export const DayUnknown = -1;
export const TimeZoneUTC = "UTC";

export let BatchSize = 156;

export class Photo extends RestModel {
  constructor(values) {
    super(values);
  }

  getDefaults() {
    return {
      ID: "",
      UID: "",
      DocumentID: "",
      Type: media.Image,
      TypeSrc: "",
      Stack: 0,
      Favorite: false,
      Private: false,
      Scan: false,
      Panorama: false,
      Portrait: false,
      TakenAt: "",
      TakenAtLocal: "",
      TakenSrc: "",
      TimeZone: "",
      Path: "",
      Color: 0,
      Name: "",
      OriginalName: "",
      Title: "",
      TitleSrc: "",
      Caption: "",
      CaptionSrc: "",
      Resolution: 0,
      Quality: 0,
      Faces: 0,
      Lat: 0.0,
      Lng: 0.0,
      Altitude: 0,
      Iso: 0,
      FocalLength: 0,
      FNumber: 0.0,
      Exposure: "",
      Views: 0,
      Camera: {},
      CameraID: 0,
      CameraMake: "",
      CameraModel: "",
      CameraType: "",
      CameraSerial: "",
      CameraSrc: "",
      Lens: {},
      LensID: 0,
      LensMake: "",
      LensModel: "",
      Country: "",
      Year: YearUnknown,
      Month: MonthUnknown,
      Day: DayUnknown,
      Details: {
        Keywords: "",
        KeywordsSrc: "",
        Notes: "",
        NotesSrc: "",
        Subject: "",
        SubjectSrc: "",
        Artist: "",
        ArtistSrc: "",
        Copyright: "",
        CopyrightSrc: "",
        License: "",
        LicenseSrc: "",
        Software: "",
        SoftwareSrc: "",
      },
      Files: [],
      Labels: [],
      Keywords: [],
      Albums: [],
      Cell: {},
      CellID: "",
      CellAccuracy: 0,
      Place: {},
      PlaceID: "",
      PlaceSrc: "",
      // Additional data in result lists.
      PlaceLabel: "",
      PlaceCity: "",
      PlaceState: "",
      PlaceCountry: "",
      FileUID: "",
      FileRoot: "",
      FileName: "",
      FileType: "",
      MediaType: "",
      FPS: 0.0,
      Frames: 0,
      Hash: "",
      Width: "",
      Height: "",
      // Date fields.
      CreatedAt: "",
      UpdatedAt: "",
      EditedAt: null,
      CheckedAt: null,
      DeletedAt: null,
    };
  }

  classes() {
    return this.generateClasses(
      this.isPlayable(),
      PhotoClipboard.has(this),
      this.Portrait,
      this.Favorite,
      this.Private,
      this.isStack()
    );
  }

  generateClasses = memoizeOne((isPlayable, isInClipboard, portrait, favorite, isPrivate, isStack) => {
    let classes = ["is-photo", "uid-" + this.UID, "type-" + this.Type];

    if (isPlayable) classes.push("is-playable");
    if (isInClipboard) classes.push("is-selected");
    if (portrait) classes.push("is-portrait");
    if (favorite) classes.push("is-favorite");
    if (isPrivate) classes.push("is-private");
    if (isStack) classes.push("is-stack");

    return classes;
  });

  localDayString() {
    if (!this.TakenAtLocal) {
      return new Date().getDate().toString().padStart(2, "0");
    }

    if (!this.Day || this.Day <= 0) {
      return this.TakenAtLocal.substring(8, 10);
    }

    return this.Day.toString().padStart(2, "0");
  }

  localMonthString() {
    if (!this.TakenAtLocal) {
      return (new Date().getMonth() + 1).toString().padStart(2, "0");
    }

    if (!this.Month || this.Month <= 0) {
      return this.TakenAtLocal.substring(5, 7);
    }

    return this.Month.toString().padStart(2, "0");
  }

  localYearString() {
    if (!this.TakenAtLocal) {
      return new Date().getFullYear().toString().padStart(4, "0");
    }

    if (!this.Year || this.Year <= 1000) {
      return this.TakenAtLocal.substring(0, 4);
    }

    return this.Year.toString();
  }

  localDateString(time) {
    if (!this.localYearString()) {
      return this.TakenAtLocal;
    }

    let date = this.localYearString() + "-" + this.localMonthString() + "-" + this.localDayString();

    if (!time) {
      time = this.TakenAtLocal.substring(11, 19);
    }

    let iso = `${date}T${time}`;

    if (this.originalTimeZoneUTC()) {
      iso += "Z";
    }

    return iso;
  }

  getTimeZone() {
    if (this.TimeZone) {
      return this.TimeZone;
    }

    return "";
  }

  timeIsUTC() {
    return this.originalTimeZoneUTC() || this.currentTimeZoneUTC();
  }

  getDateTime() {
    if (this.timeIsUTC()) {
      return DateTime.fromISO(this.TakenAt).toUTC();
    } else {
      return DateTime.fromISO(this.TakenAtLocal).toUTC();
    }
  }

  currentTimeZoneUTC() {
    const tz = this.getTimeZone();

    if (tz) {
      return tz.toLowerCase() === TimeZoneUTC.toLowerCase();
    }

    return false;
  }

  originalTimeZoneUTC() {
    const tz = this.originalValue("TimeZone");

    if (tz) {
      return tz.toLowerCase() === TimeZoneUTC.toLowerCase();
    }

    return false;
  }

  localDate(time) {
    if (!this.TakenAtLocal) {
      return this.utcDate();
    }

    let iso = this.localDateString(time);
    let zone = this.getTimeZone();

    if (zone === "") {
      zone = "UTC";
    }

    return DateTime.fromISO(iso, { zone });
  }

  utcDate() {
    return this.generateUtcDate(this.TakenAt);
  }

  generateUtcDate = memoizeOne((takenAt) => {
    return DateTime.fromISO(takenAt).toUTC();
  });

  getOriginalName() {
    const file = this.originalFile();
    return this.generateOriginalName(file);
  }

  generateOriginalName = memoizeOne((file) => {
    let name = "";

    if (file) {
      if (file.OriginalName) {
        name = file.OriginalName;
      } else if (file.Name) {
        name = file.Name;
      }
    }

    if (!name) {
      if (this.OriginalName) {
        name = this.OriginalName;
      } else if (this.FileName) {
        name = this.FileName;
      } else if (this.Name) {
        name = this.Name;
      } else {
        return $gettext("Unknown");
      }
    }

    return this.fileBase(name);
  });

  baseName(truncate) {
    let result = this.fileBase(this.FileName ? this.FileName : this.primaryFile().Name);

    if (truncate) {
      result = $util.truncate(result, truncate, "…");
    }

    return result;
  }

  fileBase(name) {
    let result = name;
    const slash = result.lastIndexOf("/");

    if (slash >= 0) {
      result = name.substring(slash + 1);
    }

    return result;
  }

  getEntityName() {
    return this.Title;
  }

  getTitle() {
    return this.Title;
  }

  getGoogleMapsLink() {
    return "https://www.google.com/maps/place/" + this.Lat + "," + this.Lng;
  }

  refreshFileAttr() {
    const file = this.primaryFile();

    if (!file || !file.Hash) {
      return;
    }

    this.Hash = file.Hash;
    this.Width = file.Width;
    this.Height = file.Height;
  }

  isPlayable() {
    return this.generateIsPlayable(this.Type, this.Files);
  }

  generateIsPlayable = memoizeOne((type, files) => {
    if (type === media.Animated || type === media.Audio) {
      return true;
    } else if (!files) {
      return false;
    }

    return files.some((f) => f.Video);
  });

  isStack() {
    return this.generateIsStack(this.Type, this.Files);
  }

  generateIsStack = memoizeOne((type, files) => {
    if (type !== media.Image) {
      return false;
    } else if (!files) {
      return false;
    } else if (files.length < 2) {
      return false;
    }

    let jpegs = 0;

    this.Files.forEach((f) => {
      if (f && f.FileType === media.FormatJpeg) {
        jpegs++;
      }
    });

    return jpegs > 1;
  });

  videoParams() {
    const uri = this.videoUrl();

    if (!uri) {
      return { error: "no video selected" };
    }

    let main = this.primaryFile();
    let file = this.videoFile();

    if (!file) {
      file = main;
    }

    const vw = Math.max(document.documentElement.clientWidth, window.innerWidth || 0);
    const vh = Math.max(document.documentElement.clientHeight, window.innerHeight || 0);

    let actualWidth = 640;
    let actualHeight = 480;

    if (file.Width > 0) {
      actualWidth = file.Width;
    } else if (main && main.Width > 0) {
      actualWidth = main.Width;
    }

    if (file.Height > 0) {
      actualHeight = file.Height;
    } else if (main && main.Height > 0) {
      actualHeight = main.Height;
    }

    let width = actualWidth;
    let height = actualHeight;

    if (vw < width + 90) {
      let newWidth = vw - 100;
      height = Math.ceil(newWidth * (actualHeight / actualWidth));
      width = newWidth;
    }

    if (vh < height + 90) {
      let newHeight = vh - 100;
      width = Math.ceil(newHeight * (actualWidth / actualHeight));
      height = newHeight;
    }

    const loop = this.Type === media.Animated || (file.Duration >= 0 && file.Duration <= 5000000000);
    const poster = this.thumbnailUrl("fit_720");
    const error = false;

    return { width, height, loop, poster, uri, error };
  }

  videoFile() {
    return this.getVideoFileFromFiles(this.Files);
  }

  getVideoFileFromFiles = memoizeOne((files) => {
    if (!files) {
      return false;
    }

    let file = files.find((f) => f.Codec === media.CodecAvc1);

    if (!file) {
      file = files.find((f) => f.FileType === media.FormatMp4);
    }

    if (!file) {
      file = files.find((f) => !!f.Video);
    }

    if (!file) {
      file = this.animatedFile();
    }

    return file;
  });

  animatedFile() {
    if (!this.Files) {
      return false;
    }

    return this.Files.find((f) => f.FileType === media.FormatGif || !!f.Frames || !!f.Duration);
  }

  videoContentType() {
    const file = this.videoFile();

    if (file) {
      return $util.videoContentType(file?.Codec, file?.Mime);
    } else {
      return media.ContentTypeMp4AvcMain;
    }
  }

  videoCodec() {
    const file = this.videoFile();

    if (file) {
      return file?.Codec;
    } else {
      return "";
    }
  }

  videoUrl() {
    const file = this.videoFile();

    return $util.videoUrl(file ? file.Hash : this.Hash, file?.Codec, file?.Mime);
  }

  primaryFile() {
    return this.generatePrimaryFile(this.Files);
  }

  generatePrimaryFile = memoizeOne((files) => {
    if (!files) {
      return this;
    }

    // Return the primary image, if found.
    let file = files.find((f) => !!f.Primary);

    // Found?
    if (file) {
      return file;
    }

    // Find and return the first JPEG or PNG image otherwise.
    file = files.find((f) => f.FileType === media.FormatJpeg || f.FileType === media.FormatPng);

    // Found?
    if (file) {
      return file;
    }

    return files.find((f) => !f.Sidecar);
  });

  originalFile() {
    // Default to main file if there is only one.
    if (this.Files?.length < 2) {
      return this.primaryFile();
    }

    // If there are multiple files, find the first one with
    // a format other than JPEG, e.g. RAW or Live.
    return this.generateOriginalFile(this.Files);
  }

  generateOriginalFile = memoizeOne((files) => {
    if (!files) {
      return this;
    }

    let file;

    // Find file with matching media type.
    switch (this.Type) {
      case media.Animated:
        file = files.find((f) => f.MediaType === media.Image && f.Root === "/");
        break;
      case media.Live:
        file = files.find(
          (f) => (f.MediaType === media.Video || f.MediaType === media.Live || f.Video) && f.Root === "/"
        );
        break;
      case media.Video:
        file = files.find((f) => (f.MediaType === media.Video || f.Video) && f.Root === "/");
        break;
      case media.Raw:
      case media.Vector:
        file = files.find((f) => f.MediaType === this.Type && f.Root === "/");
        break;
    }

    // Found?
    if (file) {
      return file;
    }

    // Find first original media file with a format other than JPEG.
    file = files.find((f) => !f.Sidecar && f.FileType !== media.FormatJpeg && f.Root === "/");

    // Found?
    if (file) {
      return file;
    }

    // Find and return the primary JPEG or PNG otherwise.
    return this.generatePrimaryFile(files);
  });

  jpegFiles() {
    if (!this.Files) {
      return [this];
    }

    return this.Files.filter((f) => f.FileType === media.FormatJpeg || f.FileType === media.FormatPng);
  }

  primaryFileHash() {
    return this.generatePrimaryFileHash(this.primaryFile(), this.Hash);
  }

  generatePrimaryFileHash = memoizeOne((primary, hash) => {
    if (this.Files) {
      if (primary && primary.Hash) {
        return primary.Hash;
      }
    } else if (hash) {
      return hash;
    }

    return "";
  });

  fileModels() {
    let result = [];

    if (!this.Files) {
      return result;
    }

    this.Files.forEach((f) => {
      result.push(new File(f));
    });

    // Get main file UID so it can be sorted first.
    const mainUID = this.originalFile()?.UID;

    result.sort((a, b) => {
      if (mainUID) {
        // Ensure that the main file is sorted first.
        if (mainUID === a.UID) {
          return -1;
        } else if (mainUID === b.UID) {
          return 1;
        }
      }

      if (a.Primary > b.Primary) {
        return -1;
      } else if (a.Primary < b.Primary) {
        return 1;
      }

      return a.Name.localeCompare(b.Name);
    });

    return result;
  }

  thumbnailUrl(size) {
    return this.generateThumbnailUrl(
      this.primaryFileHash(),
      this.videoFile(),
      $config.staticUri,
      $config.contentUri,
      $config.previewToken,
      size
    );
  }

  generateThumbnailUrl = memoizeOne((mainFileHash, videoFile, staticUri, contentUri, previewToken, size) => {
    let hash = mainFileHash;

    if (!hash) {
      if (videoFile && videoFile.Hash) {
        return `${contentUri}/t/${videoFile.Hash}/${previewToken}/${size}`;
      }

      return `${staticUri}/img/404.jpg`;
    }

    return `${contentUri}/t/${hash}/${previewToken}/${size}`;
  });

  getDownloadUrl() {
    return `${$config.apiUri}/dl/${this.primaryFileHash()}?t=${$config.downloadToken}`;
  }

  downloadAll() {
    const s = $config.getSettings();

    if (!s || !s.features || !s.download || !s.features.download || s.download.disabled) {
      console.log("download: disabled in settings", s.features, s.download);
      return;
    }

    const token = $config.downloadToken;

    if (!this.Files) {
      const hash = this.primaryFileHash();

      if (hash) {
        download(`/${$config.apiUri}/dl/${hash}?t=${token}`, this.baseName(false));
      } else if ($config.debug) {
        console.log("download: failed, empty file hash", this);
      }

      return;
    }

    this.Files.forEach((file) => {
      if (!file || !file.Hash) {
        return;
      }

      // Originals only?
      if (s.download.originals && file.Root.length > 1) {
        // Don't download broken files and sidecars.
        if ($config.debug) console.log(`download: skipped ${file.Root} file ${file.Name}`);
        return;
      }

      // Skip metadata sidecar files?
      if (!s.download.mediaSidecar && (file.MediaType === media.Sidecar || file.Sidecar)) {
        // Don't download broken files and sidecars.
        if ($config.debug) console.log(`download: skipped sidecar file ${file.Name}`);
        return;
      }

      // Skip RAW images?
      if (!s.download.mediaRaw && (file.MediaType === media.Raw || file.FileType === media.Raw)) {
        if ($config.debug) console.log(`download: skipped raw file ${file.Name}`);
        return;
      }

      // If this is a video, always skip stacked images...
      // see https://github.com/photoprism/photoprism/issues/1436
      if (this.Type === media.Video && !(file.MediaType === media.Video || file.Video)) {
        if ($config.debug) console.log(`download: skipped video sidecar ${file.Name}`);
        return;
      }

      download(`${$config.apiUri}/dl/${file.Hash}?t=${token}`, this.fileBase(file.Name));
    });
  }

  calculateSize(width, height) {
    if (width >= this.Width && height >= this.Height) {
      // Smaller
      return { width: this.Width, height: this.Height };
    }

    const srcAspectRatio = this.Width / this.Height;
    const maxAspectRatio = width / height;

    let newW, newH;

    if (srcAspectRatio > maxAspectRatio) {
      newW = width;
      newH = Math.ceil(newW / srcAspectRatio);
    } else {
      newH = height;
      newW = Math.ceil(newH * srcAspectRatio);
    }

    return { width: newW, height: newH };
  }

  getDateString(showTimeZone) {
    return this.generateDateString(showTimeZone, this.TakenAt, this.Year, this.Month, this.Day, this.TimeZone);
  }

  generateDateString = memoizeOne((showTimeZone, takenAt, year, month, day, timeZone) => {
    if (!takenAt || year === YearUnknown) {
      return $gettext("Unknown");
    } else if (month === MonthUnknown) {
      return this.localYearString();
    } else if (day === DayUnknown) {
      return this.localDate().toLocaleString({
        month: formats.long,
        year: formats.num,
      });
    } else if (timeZone) {
      return this.localDate().toLocaleString(showTimeZone ? formats.DATE_FULL_TZ : formats.DATE_FULL);
    }

    return this.localDate().toLocaleString(DateTime.DATE_HUGE);
  });

  shortDateString = () => {
    return this.generateShortDateString(this.TakenAt, this.Year, this.Month, this.Day);
  };

  generateShortDateString = memoizeOne((takenAt, year, month, day) => {
    if (!takenAt || year === YearUnknown) {
      return $gettext("Unknown");
    } else if (month === MonthUnknown) {
      return this.localYearString();
    } else if (day === DayUnknown) {
      return this.localDate().toLocaleString({ month: "long", year: "numeric" });
    }

    return this.localDate().toLocaleString({ day: "numeric", month: "numeric", year: "numeric" });
  });

  hasLocation() {
    return this.Lat !== 0 || this.Lng !== 0;
  }

  countryName() {
    if (this.Country !== "zz") {
      const country = countries.find((c) => c.Code === this.Country);

      if (country) {
        return country.Name;
      }
    }

    return $gettext("Unknown");
  }

  locationInfo = () => {
    return this.generateLocationInfo(this.PlaceID, this.Country, this.Place, this.PlaceLabel);
  };

  generateLocationInfo = memoizeOne((placeId, countryCode, place, placeLabel) => {
    if (placeId === "zz" && countryCode !== "zz") {
      const country = countries.find((c) => c.Code === countryCode);

      if (country) {
        return country.Name;
      }
    } else if (place && place.Label) {
      return place.Label;
    }

    return placeLabel ? placeLabel : $gettext("Unknown");
  });

  addSizeInfo(file, info) {
    if (!file) {
      return;
    }

    if (file?.Pages > 0) {
      info.push(file.Pages + " " + $gettext("Pages"));
    }

    if (file?.MediaType !== media.Document) {
      if (file.Width && file.Height) {
        info.push(file.Width + " × " + file.Height);
      } else if (!file.Primary) {
        let primary = this.primaryFile();
        if (primary && primary.Width && primary.Height) {
          info.push(primary.Width + " × " + primary.Height);
        }
      }
    }

    if (!file.Size) {
      return;
    }

    info.push($util.formatBytes(file.Size));
  }

  vectorFile() {
    if (!this.Files) {
      return this;
    }

    return this.Files.find(
      (f) => f.MediaType === media.Document || f.MediaType === media.Vector || f.FileType === media.FormatSVG
    );
  }

  getVectorInfo = () => {
    let file = this.vectorFile() || this.primaryFile();
    return this.generateVectorInfo(file);
  };

  generateVectorInfo = memoizeOne((file) => {
    if (!file) {
      return $gettext("Unknown");
    }

    const info = [];

    if (file.MediaType === media.Vector || file.MediaType === media.Document) {
      info.push($util.fileType(file.FileType));
    } else {
      info.push($gettext("Unknown"));
    }

    this.addSizeInfo(file, info);

    return info.join(", ");
  });

  // Example: 1:03:46, HEVC, 1440 × 1920, 4.2 MB
  getVideoInfo = () => {
    let file = this.videoFile() || this.primaryFile();
    return this.generateVideoInfo(this.Camera, this.CameraID, this.CameraMake, this.CameraModel, file);
  };

  generateVideoInfo = memoizeOne((camera, cameraId, cameraMake, cameraModel, file) => {
    if (!file) {
      return $gettext("Video");
    }

    const info = [];

    if (file.Duration > 0) {
      info.push($util.formatDuration(file.Duration));
    }

    if (file.Codec) {
      info.push($util.formatCodec(file.Codec));
    } else if (file.FileType) {
      info.push($util.formatCodec(file.FileType));
    }

    this.addSizeInfo(file, info);

    if (!info.length) {
      return $gettext("Video");
    }

    return info.join(", ");
  });

  // Example: 1:03:46
  getDurationInfo = () => {
    let file = this.videoFile() || this.primaryFile();
    return this.generateDurationInfo(file);
  };

  generateDurationInfo = memoizeOne((file) => {
    if (!file) {
      return "▶";
    } else if (file.Duration && file.Duration > 0) {
      return $util.formatDuration(file.Duration);
    }

    if (file.Codec) {
      return $util.formatCodec(file.Codec);
    } else if (file.FileType) {
      return $util.formatCodec(file.FileType);
    }

    return "▶";
  });

  // Example: Apple iPhone 12 Pro Max, DNG, 4032 × 3024, 32.9 MB
  getCameraInfo = () => {
    return this.generateCameraInfo(
      this.Camera,
      this.CameraID,
      this.CameraMake,
      this.CameraModel,
      this.Iso,
      this.Exposure
    );
  };

  generateCameraInfo = memoizeOne((camera, cameraId, cameraMake, cameraModel, iso, exposure) => {
    let info = [];

    // Return only the complete camera name if the original is or contains a video.
    info.push($util.formatCamera(camera, cameraId, cameraMake, cameraModel, true));

    if (iso) {
      info.push("ISO " + iso);
    }

    if (exposure) {
      info.push(exposure);
    }

    if (!info.length) {
      return $gettext("Unknown");
    }

    return info.join(", ");
  });

  // Example: DNG, 4032 × 3024, 32.9 MB
  getImageInfo = () => {
    let file = this.originalFile() || this.videoFile();
    return this.generateImageInfo(file);
  };

  generateImageInfo = memoizeOne((file) => {
    let info = [];

    if (file && file.Width && file.Codec) {
      info.push($util.formatCodec(file.Codec));
    }

    this.addSizeInfo(file, info);

    if (!info.length) {
      return $gettext("Unknown");
    }

    return info.join(", ");
  });

  // Example: iPhone 12 Pro Max 5.1mm ƒ/1.6, 26mm, ISO32, 1/4525
  getLensInfo = () => {
    return this.generateLensInfo(
      this.Lens,
      this.LensID,
      this.LensMake,
      this.LensModel,
      this.CameraModel,
      this.FNumber,
      this.FocalLength
    );
  };

  generateLensInfo = memoizeOne((lens, lensId, lensMake, lensModel, cameraModel, fNumber, focalLength) => {
    let info = [];
    const id = lensId ? lensId : lens && lens.ID ? lens.ID : 1;
    const make = lensMake ? lensMake : lens && lens.Make ? lens.Make : "";
    const model = (lensModel ? lensModel : lens && lens.Model ? lens.Model : "").replace("f/", "ƒ/");

    // Example: EF-S18-55mm f/3.5-5.6 IS STM
    if (id > 1) {
      if (!model && !!make) {
        info.push(make);
      } else if (model.length > 45) {
        return model;
      } else if (model) {
        info.push(model);
      }
    }

    if (focalLength) {
      info.push(focalLength + "mm");
    }

    if (fNumber && (!model || !model.endsWith(fNumber.toString()))) {
      info.push("ƒ/" + fNumber);
    }

    if (!info.length) {
      return $gettext("Unknown");
    }

    return info.join(", ");
  });

  getCamera() {
    if (this.Camera) {
      return this.Camera.Make + " " + this.Camera.Model;
    } else if (this.CameraModel) {
      return this.CameraMake + " " + this.CameraModel;
    }

    return $gettext("Unknown");
  }

  archive() {
    return $api.post("batch/photos/archive", { photos: [this.getId()] });
  }

  approve() {
    return $api.post(this.getEntityResource() + "/approve");
  }

  toggleLike() {
    const favorite = !this.Favorite;
    const elements = document.querySelectorAll(`.uid-${this.UID}`);

    if (favorite) {
      elements.forEach((el) => el.classList.add("is-favorite"));
      return $api.post(this.getEntityResource() + "/like");
    } else {
      elements.forEach((el) => el.classList.remove("is-favorite"));
      return $api.delete(this.getEntityResource() + "/like");
    }
  }

  togglePrivate() {
    this.Private = !this.Private;

    return $api.put(this.getEntityResource(), { Private: this.Private });
  }

  setPrimaryFile(fileUID) {
    return $api
      .post(`${this.getEntityResource()}/files/${fileUID}/primary`)
      .then((r) => Promise.resolve(this.setValues(r.data)));
  }

  unstackFile(fileUID) {
    return $api
      .post(`${this.getEntityResource()}/files/${fileUID}/unstack`)
      .then((r) => Promise.resolve(this.setValues(r.data)));
  }

  deleteFile(fileUID) {
    return $api
      .delete(`${this.getEntityResource()}/files/${fileUID}`)
      .then((r) => Promise.resolve(this.setValues(r.data)));
  }

  changeFileOrientation(file) {
    // Return if no file was provided.
    if (!file) {
      return Promise.resolve(this);
    }

    // Get updated values.
    const values = file.getValues(true);

    // Return if no values were changed.
    if (Object.keys(values).length === 0) {
      return Promise.resolve(this);
    }

    // Change file orientation.
    return $api
      .put(`${this.getEntityResource()}/files/${file.UID}/orientation`, values)
      .then((r) => Promise.resolve(this.setValues(r.data)));
  }

  like() {
    this.Favorite = true;
    return $api.post(this.getEntityResource() + "/like");
  }

  unlike() {
    this.Favorite = false;
    return $api.delete(this.getEntityResource() + "/like");
  }

  addLabel(name) {
    return $api
      .post(this.getEntityResource() + "/label", { Name: name, Priority: 10 })
      .then((r) => Promise.resolve(this.setValues(r.data)));
  }

  activateLabel(id) {
    return $api
      .put(this.getEntityResource() + "/label/" + id, { Uncertainty: 0 })
      .then((r) => Promise.resolve(this.setValues(r.data)));
  }

  renameLabel(id, name) {
    return $api
      .put(this.getEntityResource() + "/label/" + id, { Label: { Name: name } })
      .then((r) => Promise.resolve(this.setValues(r.data)));
  }

  removeLabel(id) {
    return $api.delete(this.getEntityResource() + "/label/" + id).then((r) => Promise.resolve(this.setValues(r.data)));
  }

  getMarkers(valid) {
    let result = [];

    let file = this.Files.find((f) => !!f.Primary);

    if (!file || !file.Markers) {
      return result;
    }

    file.Markers.forEach((m) => {
      if (valid && m.Invalid) {
        return;
      }

      result.push(new Marker(m));
    });

    return result;
  }

  update() {
    const values = this.getValues(true);

    if (typeof values.Title === "string") {
      values.TitleSrc = src.Manual;
    }

    if (values.Type) {
      values.TypeSrc = src.Manual;
    }

    if (typeof values.Caption === "string") {
      values.CaptionSrc = src.Manual;
    }

    if (values.Lat || values.Lng || values.Country) {
      values.PlaceSrc = src.Manual;
    }

    if (values.TakenAt || values.TakenAtLocal || values.TimeZone || values.Day || values.Month || values.Year) {
      values.TakenSrc = src.Manual;
    }

    if (values.CameraID || values.LensID || values.FocalLength || values.FNumber || values.Iso || values.Exposure) {
      values.CameraSrc = src.Manual;
    }

    // Update details source if needed.
    if (values.Details) {
      if (values.Details.Keywords !== this.__originalValues.Details.Keywords) {
        values.Details.KeywordsSrc = src.Manual;
      }

      if (values.Details.Notes !== this.__originalValues.Details.Notes) {
        values.Details.NotesSrc = src.Manual;
      }

      if (values.Details.Subject !== this.__originalValues.Details.Subject) {
        values.Details.SubjectSrc = src.Manual;
      }

      if (values.Details.Artist !== this.__originalValues.Details.Artist) {
        values.Details.ArtistSrc = src.Manual;
      }

      if (values.Details.Copyright !== this.__originalValues.Details.Copyright) {
        values.Details.CopyrightSrc = src.Manual;
      }

      if (values.Details.License !== this.__originalValues.Details.License) {
        values.Details.LicenseSrc = src.Manual;
      }
    }

    return $api.put(this.getEntityResource(), values).then((resp) => {
      if (values.Type || values.Lat) {
        $config.update();
      }

      return Promise.resolve(this.setValues(resp.data));
    });
  }

  static batchSize() {
    return BatchSize;
  }

  static setBatchSize(count) {
    const s = parseInt(count);
    if (!isNaN(s) && s >= 24) {
      BatchSize = s;
    }
  }

  static getCollectionResource() {
    return "photos";
  }

  static getModelName() {
    return $gettext("Photo");
  }

  static mergeResponse(results, response) {
    if (response.offset === 0 || results.length === 0) {
      return response.models;
    }

    if (response.models.length > 0) {
      let i = results.length - 1;

      if (results[i].UID === response.models[0].UID) {
        const first = response.models.shift();
        results[i].Files = results[i].Files.concat(first.Files);
      }
    }

    return results.concat(response.models);
  }
}

export default Photo;
