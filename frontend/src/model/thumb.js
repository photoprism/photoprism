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

import Model from "model.js";
import Api from "common/api";
import { config } from "app/session.js";
import { $gettext } from "common/vm";

const thumbs = window.__CONFIG__.thumbs;

export class Thumb extends Model {
  getDefaults() {
    return {
      UID: "",
      Title: "",
      TakenAtLocal: "",
      Description: "",
      Favorite: false,
      Playable: false,
      DownloadUrl: "",
      Width: 0,
      Height: 0,
      Thumbs: {},
    };
  }

  getId() {
    if (this.UID) {
      return this.UID;
    }

    return this.ID ? this.ID : false;
  }

  hasId() {
    return !!this.getId();
  }

  toggleLike() {
    this.Favorite = !this.Favorite;

    if (this.Favorite) {
      return Api.post("photos/" + this.UID + "/like");
    } else {
      return Api.delete("photos/" + this.UID + "/like");
    }
  }

  static notFound() {
    const result = {
      UID: "",
      Title: $gettext("Invalid photo selected"),
      TakenAtLocal: "",
      Description: "",
      Favorite: false,
      Playable: false,
      DownloadUrl: "",
      Width: 0,
      Height: 0,
      Thumbs: {},
    };

    for (let i = 0; i < thumbs.length; i++) {
      let t = thumbs[i];

      result.Thumbs[t.size] = {
        src: `${config.staticUri}/img/404.jpg`,
        w: t.w,
        h: t.h,
      };
    }

    return result;
  }

  static fromPhotos(photos) {
    let result = [];
    const n = photos.length;

    for (let i = 0; i < n; i++) {
      result.push(this.fromPhoto(photos[i]));
    }

    return result;
  }

  static fromPhoto(photo) {
    if (photo.Files) {
      return this.fromFile(photo, photo.mainFile());
    }

    if (!photo || !photo.Hash) {
      return this.notFound();
    }

    const result = {
      UID: photo.UID,
      Title: photo.Title,
      TakenAtLocal: photo.getDateString(),
      Description: photo.Description,
      Favorite: photo.Favorite,
      Playable: photo.isPlayable(),
      DownloadUrl: this.downloadUrl(photo),
      Width: photo.Width,
      Height: photo.Height,
      Thumbs: {},
    };

    for (let i = 0; i < thumbs.length; i++) {
      let t = thumbs[i];
      let size = photo.calculateSize(t.w, t.h);

      result.Thumbs[t.size] = {
        src: photo.thumbnailUrl(t.size),
        w: size.width,
        h: size.height,
      };
    }

    return new this(result);
  }

  static fromFile(photo, file) {
    if (!photo || !file || !file.Hash) {
      return this.notFound();
    }

    const result = {
      UID: photo.UID,
      Title: photo.Title,
      TakenAtLocal: photo.getDateString(),
      Description: photo.Description,
      Favorite: photo.Favorite,
      Playable: photo.isPlayable(),
      DownloadUrl: this.downloadUrl(file),
      Width: file.Width,
      Height: file.Height,
      Thumbs: {},
    };

    for (let i = 0; i < thumbs.length; i++) {
      let t = thumbs[i];
      let size = this.calculateSize(file, t.w, t.h);

      result.Thumbs[t.size] = {
        src: this.thumbnailUrl(file, t.size),
        w: size.width,
        h: size.height,
      };
    }

    return new this(result);
  }

  static wrap(data) {
    return data.map((values) => new this(values));
  }

  static fromFiles(photos) {
    let result = [];

    if (!photos || !photos.length) {
      return result;
    }

    const n = photos.length;

    for (let i = 0; i < n; i++) {
      let p = photos[i];

      if (!p.Files || !p.Files.length) {
        continue;
      }

      for (let j = 0; j < p.Files.length; j++) {
        let f = p.Files[j];

        if (!f || (f.FileType !== "jpg" && f.FileType !== "png")) {
          continue;
        }

        let thumb = this.fromFile(p, f);

        if (thumb) {
          result.push(thumb);
        }
      }
    }

    return result;
  }

  static calculateSize(file, width, height) {
    if (width >= file.Width && height >= file.Height) {
      // Smaller
      return { width: file.Width, height: file.Height };
    }

    const srcAspectRatio = file.Width / file.Height;
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

  static thumbnailUrl(file, size) {
    if (!file.Hash) {
      return `${config.staticUri}/img/404.jpg`;
    }

    return `${config.contentUri}/t/${file.Hash}/${config.previewToken}/${size}`;
  }

  static downloadUrl(file) {
    if (!file || !file.Hash) {
      return "";
    }

    return `${config.apiUri}/dl/${file.Hash}?t=${config.downloadToken}`;
  }
}

export default Thumb;
