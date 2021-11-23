/*

Copyright (c) 2018 - 2021 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

import RestModel from "model/rest";
import File from "model/file";
import Marker from "model/marker";
import Api from "common/api";
import { DateTime } from "luxon";
import Util from "common/util";
import { config } from "../session";
import countries from "options/countries.json";
import { $gettext } from "common/vm";
import Clipboard from "common/clipboard";
import download from "common/download";
import * as src from "common/src";

export const CodecAvc1 = "avc1";
export const FormatMp4 = "mp4";
export const FormatAvc = "avc";
export const FormatJpeg = "jpg";
export const TypeImage = "image";
export const TypeVideo = "video";
export const TypeLive = "live";
export const TypeRaw = "raw";
export const YearUnknown = -1;
export const MonthUnknown = -1;
export const DayUnknown = -1;
export const TimeZoneUTC = "UTC";

const num = "numeric";
const short = "short";
const long = "long";

const DATE_FULL = {
  year: num,
  month: long,
  day: num,
  weekday: long,
  hour: num,
  minute: num,
};

const DATE_FULL_TZ = {
  year: num,
  month: long,
  day: num,
  weekday: long,
  hour: num,
  minute: num,
  timeZoneName: short,
};

export class Photo extends RestModel {
  constructor(values) {
    super(values);
  }

  getDefaults() {
    return {
      ID: "",
      UID: "",
      DocumentID: "",
      Type: TypeImage,
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
      Description: "",
      DescriptionSrc: "",
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
      CameraSerial: "",
      CameraSrc: "",
      Lens: {},
      LensID: 0,
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
    let classes = ["is-photo", "uid-" + this.UID, "type-" + this.Type];

    if (this.isPlayable()) classes.push("is-playable");
    if (Clipboard.has(this)) classes.push("is-selected");
    if (this.Portrait) classes.push("is-portrait");
    if (this.Favorite) classes.push("is-favorite");
    if (this.Private) classes.push("is-private");
    if (this.Files.length > 1) classes.push("is-stack");

    return classes;
  }

  localDayString() {
    if (!this.TakenAtLocal) {
      return new Date().getDate().toString().padStart(2, "0");
    }

    if (!this.Day || this.Day <= 0) {
      return this.TakenAtLocal.substr(8, 2);
    }

    return this.Day.toString().padStart(2, "0");
  }

  localMonthString() {
    if (!this.TakenAtLocal) {
      return (new Date().getMonth() + 1).toString().padStart(2, "0");
    }

    if (!this.Month || this.Month <= 0) {
      return this.TakenAtLocal.substr(5, 2);
    }

    return this.Month.toString().padStart(2, "0");
  }

  localYearString() {
    if (!this.TakenAtLocal) {
      return new Date().getFullYear().toString().padStart(4, "0");
    }

    if (!this.Year || this.Year <= 1000) {
      return this.TakenAtLocal.substr(0, 4);
    }

    return this.Year.toString();
  }

  localDateString(time) {
    if (!this.localYearString()) {
      return this.TakenAtLocal;
    }

    let date = this.localYearString() + "-" + this.localMonthString() + "-" + this.localDayString();

    if (!time) {
      time = this.TakenAtLocal.substr(11, 8);
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

    if (this.getTimeZone() === "") {
      zone = "UTC";
    }

    return DateTime.fromISO(iso, { zone });
  }

  utcDate() {
    return DateTime.fromISO(this.TakenAt).toUTC();
  }

  baseName(truncate) {
    let result = this.fileBase(this.FileName ? this.FileName : this.mainFile().Name);

    if (truncate) {
      result = Util.truncate(result, truncate, "…");
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
    const file = this.mainFile();

    if (!file || !file.Hash) {
      return;
    }

    this.Hash = file.Hash;
    this.Width = file.Width;
    this.Height = file.Height;
  }

  isPlayable() {
    if (!this.Files) {
      return false;
    }

    return this.Files.findIndex((f) => f.Video) !== -1;
  }

  videoParams() {
    const uri = this.videoUrl();

    if (!uri) {
      return { error: "no video selected" };
    }

    let main = this.mainFile();
    let file = this.videoFile();

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

    if (vw < width + 80) {
      let newWidth = vw - 90;
      height = Math.round(newWidth * (actualHeight / actualWidth));
      width = newWidth;
    }

    if (vh < height + 100) {
      let newHeight = vh - 160;
      width = Math.round(newHeight * (actualWidth / actualHeight));
      height = newHeight;
    }

    const loop = file.Duration >= 0 && file.Duration <= 5000000000;
    const poster = this.thumbnailUrl("fit_720");
    const error = false;

    return { width, height, loop, poster, uri, error };
  }

  videoFile() {
    if (!this.Files) {
      return false;
    }

    let file = this.Files.find((f) => f.Codec === CodecAvc1);

    if (!file) {
      file = this.Files.find((f) => f.Type === FormatMp4);
    }

    if (!file) {
      file = this.Files.find((f) => !!f.Video);
    }

    return file;
  }

  videoUrl() {
    const file = this.videoFile();

    if (file) {
      return `${config.contentUri}/videos/${file.Hash}/${config.previewToken()}/${FormatAvc}`;
    }

    return `${config.contentUri}/videos/${this.Hash}/${config.previewToken()}/${FormatAvc}`;
  }

  mainFile() {
    if (!this.Files) {
      return this;
    }

    let file = this.Files.find((f) => !!f.Primary);

    if (file) {
      return file;
    }

    return this.Files.find((f) => f.Type === FormatJpeg);
  }

  jpegFiles() {
    if (!this.Files) {
      return [this];
    }

    return this.Files.filter((f) => f.Type === FormatJpeg);
  }

  mainFileHash() {
    if (this.Files) {
      let file = this.mainFile();

      if (file && file.Hash) {
        return file.Hash;
      }
    } else if (this.Hash) {
      return this.Hash;
    }

    return "";
  }

  fileModels() {
    let result = [];

    if (!this.Files) {
      return result;
    }

    this.Files.forEach((f) => {
      result.push(new File(f));
    });

    result.sort((a, b) => {
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
    let hash = this.mainFileHash();

    if (!hash) {
      let video = this.videoFile();

      if (video && video.Hash) {
        return `${config.contentUri}/t/${video.Hash}/${config.previewToken()}/${size}`;
      }

      return `${config.contentUri}/svg/photo`;
    }

    return `${config.contentUri}/t/${hash}/${config.previewToken()}/${size}`;
  }

  getDownloadUrl() {
    return `${config.apiUri}/dl/${this.mainFileHash()}?t=${config.downloadToken()}`;
  }

  downloadAll() {
    const token = config.downloadToken();

    if (!this.Files) {
      const hash = this.mainFileHash();

      if (hash) {
        download(`/${config.apiUri}/dl/${hash}?t=${token}`, this.baseName(false));
      } else if (config.debug) {
        console.log("download: failed, empty file hash", this);
      }

      return;
    }

    this.Files.forEach((file) => {
      if (!file || !file.Hash || file.Sidecar) {
        // Don't download broken files and sidecars.
        if (config.debug) console.log("download: skipped file", file);
        return;
      }

      // Skip related images if video.
      // see https://github.com/photoprism/photoprism/issues/1436
      if (this.Type === TypeVideo && !file.Video) {
        if (config.debug) console.log("download: skipped image", file);
        return;
      }

      download(`${config.apiUri}/dl/${file.Hash}?t=${token}`, this.fileBase(file.Name));
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
      newH = Math.round(newW / srcAspectRatio);
    } else {
      newH = height;
      newW = Math.round(newH * srcAspectRatio);
    }

    return { width: newW, height: newH };
  }

  getDateString(showTimeZone) {
    if (!this.TakenAt || this.Year === YearUnknown) {
      return $gettext("Unknown");
    } else if (this.Month === MonthUnknown) {
      return this.localYearString();
    } else if (this.Day === DayUnknown) {
      return this.localDate().toLocaleString({
        month: long,
        year: num,
      });
    } else if (this.TimeZone) {
      return this.localDate().toLocaleString(showTimeZone ? DATE_FULL_TZ : DATE_FULL);
    }

    return this.localDate().toLocaleString(DateTime.DATE_HUGE);
  }

  shortDateString() {
    if (!this.TakenAt || this.Year === YearUnknown) {
      return $gettext("Unknown");
    } else if (this.Month === MonthUnknown) {
      return this.localYearString();
    } else if (this.Day === DayUnknown) {
      return this.localDate().toLocaleString({ month: "long", year: "numeric" });
    }

    return this.localDate().toLocaleString(DateTime.DATE_MED);
  }

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

  locationInfo() {
    if (this.PlaceID === "zz" && this.Country !== "zz") {
      const country = countries.find((c) => c.Code === this.Country);

      if (country) {
        return country.Name;
      }
    } else if (this.Place && this.Place.Label) {
      return this.Place.Label;
    }

    return this.PlaceLabel ? this.PlaceLabel : $gettext("Unknown");
  }

  addSizeInfo(file, info) {
    if (!file) {
      return;
    }

    if (file.Width && file.Height) {
      info.push(file.Width + " × " + file.Height);
    } else if (!file.Primary) {
      let main = this.mainFile();
      if (main && main.Width && main.Height) {
        info.push(main.Width + " × " + main.Height);
      }
    }

    if (file.Size > 102400) {
      const size = Number.parseFloat(file.Size) / 1048576;

      info.push(size.toFixed(1) + " MB");
    } else if (file.Size) {
      const size = Number.parseFloat(file.Size) / 1024;

      info.push(size.toFixed(1) + " KB");
    }
  }

  getVideoInfo() {
    let info = [];
    let file = this.videoFile();

    if (!file) {
      file = this.mainFile();
    }

    if (!file) {
      return $gettext("Video");
    }

    if (file.Duration > 0) {
      info.push(Util.duration(file.Duration));
    }

    if (file.Codec) {
      info.push(file.Codec.toUpperCase());
    }

    this.addSizeInfo(file, info);

    if (!info.length) {
      return $gettext("Video");
    }

    return info.join(", ");
  }

  getPhotoInfo() {
    let info = [];

    if (this.Camera) {
      info.push(this.Camera.Make + " " + this.Camera.Model);
    } else if (this.CameraModel && this.CameraMake) {
      info.push(this.CameraMake + " " + this.CameraModel);
    }

    let file = this.videoFile();

    if (!file || !file.Width) {
      file = this.mainFile();
    } else if (file.Codec) {
      info.push(file.Codec.toUpperCase());
    }

    this.addSizeInfo(file, info);

    if (!info.length) {
      return $gettext("Unknown");
    }

    return info.join(", ");
  }

  getCamera() {
    if (this.Camera) {
      return this.Camera.Make + " " + this.Camera.Model;
    } else if (this.CameraModel) {
      return this.CameraMake + " " + this.CameraModel;
    }

    return $gettext("Unknown");
  }

  archive() {
    return Api.post("batch/photos/archive", { photos: [this.getId()] });
  }

  approve() {
    return Api.post(this.getEntityResource() + "/approve");
  }

  toggleLike() {
    const favorite = !this.Favorite;
    const elements = document.querySelectorAll(`.uid-${this.UID}`);

    if (favorite) {
      elements.forEach((el) => el.classList.add("is-favorite"));
      return Api.post(this.getEntityResource() + "/like");
    } else {
      elements.forEach((el) => el.classList.remove("is-favorite"));
      return Api.delete(this.getEntityResource() + "/like");
    }
  }

  togglePrivate() {
    this.Private = !this.Private;

    return Api.put(this.getEntityResource(), { Private: this.Private });
  }

  primaryFile(fileUID) {
    return Api.post(`${this.getEntityResource()}/files/${fileUID}/primary`).then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
  }

  unstackFile(fileUID) {
    return Api.post(`${this.getEntityResource()}/files/${fileUID}/unstack`).then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
  }

  deleteFile(fileUID) {
    return Api.delete(`${this.getEntityResource()}/files/${fileUID}`).then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
  }

  like() {
    this.Favorite = true;
    return Api.post(this.getEntityResource() + "/like");
  }

  unlike() {
    this.Favorite = false;
    return Api.delete(this.getEntityResource() + "/like");
  }

  addLabel(name) {
    return Api.post(this.getEntityResource() + "/label", { Name: name, Priority: 10 }).then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
  }

  activateLabel(id) {
    return Api.put(this.getEntityResource() + "/label/" + id, { Uncertainty: 0 }).then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
  }

  renameLabel(id, name) {
    return Api.put(this.getEntityResource() + "/label/" + id, { Label: { Name: name } }).then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
  }

  removeLabel(id) {
    return Api.delete(this.getEntityResource() + "/label/" + id).then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
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

    if (values.Title) {
      values.TitleSrc = src.Manual;
    }

    if (values.Type) {
      values.TypeSrc = src.Manual;
    }

    if (values.Description) {
      values.DescriptionSrc = src.Manual;
    }

    if (values.Lat || values.Lng || values.Country) {
      values.PlaceSrc = src.Manual;
    }

    if (values.TakenAt || values.TimeZone || values.Day || values.Month || values.Year) {
      values.TakenSrc = src.Manual;
    }

    if (
      values.CameraID ||
      values.LensID ||
      values.FocalLength ||
      values.FNumber ||
      values.Iso ||
      values.Exposure
    ) {
      values.CameraSrc = src.Manual;
    }

    // Update details source if needed.
    if (values.Details) {
      if (values.Details.Keywords) {
        values.Details.KeywordsSrc = src.Manual;
      }

      if (values.Details.Notes) {
        values.Details.NotesSrc = src.Manual;
      }

      if (values.Details.Subject) {
        values.Details.SubjectSrc = src.Manual;
      }

      if (values.Details.Artist) {
        values.Details.ArtistSrc = src.Manual;
      }

      if (values.Details.Copyright) {
        values.Details.CopyrightSrc = src.Manual;
      }

      if (values.Details.License) {
        values.Details.LicenseSrc = src.Manual;
      }
    }

    return Api.put(this.getEntityResource(), values).then((resp) => {
      if (values.Type || values.Lat) {
        config.update();
      }

      return Promise.resolve(this.setValues(resp.data));
    });
  }

  static batchSize() {
    return 60;
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
