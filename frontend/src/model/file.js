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

import RestModel from "model/rest";
import { DateTime } from "luxon";
import { $config } from "app/session";
import { $gettext } from "common/gettext";
import download from "common/download";
import * as media from "common/media";
import $api from "common/api";
import $util from "common/util";

export class File extends RestModel {
  getDefaults() {
    return {
      UID: "",
      PhotoUID: "",
      InstanceID: "",
      MediaID: "",
      MediaUTC: 0,
      TakenAt: "",
      Root: "/",
      Name: "",
      OriginalName: "",
      Hash: "",
      Size: 0,
      ModTime: 0,
      Codec: "",
      FileType: "",
      MediaType: "",
      Mime: "",
      Primary: false,
      Sidecar: false,
      Missing: false,
      Portrait: false,
      Video: false,
      Duration: 0,
      FPS: 0.0,
      Frames: 0,
      Pages: 0,
      Width: 0,
      Height: 0,
      Orientation: 0,
      OrientationSrc: "",
      Projection: "",
      AspectRatio: 1.0,
      HDR: false,
      Watermark: false,
      ColorProfile: "",
      MainColor: "",
      Colors: "",
      Luminance: "",
      Diff: 0,
      Chroma: 0,
      Software: "",
      Error: "",
      Markers: [],
      CreatedAt: "",
      CreatedIn: 0,
      UpdatedAt: "",
      UpdatedIn: 0,
      DeletedAt: "",
    };
  }

  classes(selected) {
    let classes = ["is-file", "uid-" + this.UID];

    if (this.Primary) classes.push("is-primary");
    if (this.Sidecar) classes.push("is-sidecar");
    if (this.Video) classes.push("is-video");
    if (selected) classes.push("is-selected");

    return classes;
  }

  baseName(truncate) {
    let result = this.Name;
    const slash = result.lastIndexOf("/");

    if (slash >= 0) {
      result = this.Name.substring(slash + 1);
    }

    if (truncate) {
      result = $util.truncate(result, truncate, "…");
    }

    return result;
  }

  isFile() {
    return true;
  }

  getEntityName() {
    return this.Root + "/" + this.Name;
  }

  thumbnailUrl(size) {
    if (this.Error || this.Missing) {
      return `${$config.contentUri}/svg/broken`;
    } else if (this.Sidecar) {
      return `${$config.contentUri}/svg/file`;
    }

    return `${$config.contentUri}/t/${this.Hash}/${$config.previewToken}/${size}`;
  }

  getDownloadUrl() {
    return `${$config.apiUri}/dl/${this.Hash}?t=${$config.downloadToken}`;
  }

  download() {
    if (!this.Hash) {
      return;
    }

    download(this.getDownloadUrl(), this.baseName(this.Name));
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

  getDateString() {
    return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
  }

  getInfo() {
    let info = [];

    if (this.FileType) {
      info.push(this.FileType.toUpperCase());
    }

    if (this.Duration > 0) {
      info.push($util.formatDuration(this.Duration));
    }

    if (this.FPS > 0) {
      info.push($util.formatFPS(this.FPS));
    }

    this.addSizeInfo(info);

    return info.join(", ");
  }

  storageInfo() {
    if (!this.Root || this.Root === "") {
      return "";
    }

    if (this.Root.length === 1) {
      return $gettext("Originals");
    } else {
      return $util.capitalize(this.Root);
    }
  }

  isAnimated() {
    return (
      this.MediaType &&
      this.MediaType === media.Image &&
      ((this.Frames && this.Frames > 1) || (this.Duration && this.Duration > 1))
    );
  }

  typeInfo() {
    let info = [];

    if (this.Sidecar) {
      info.push($gettext("Sidecar"));
    }

    if (this.Primary && !this.MediaType) {
      info.push($gettext("Image"));
      return info.join(" ");
    } else if (this.Video && !this.MediaType) {
      info.push($gettext("Video"));
      return info.join(" ");
    } else if (this.MediaType === "vector") {
      info.push($util.fileType(this.FileType));
      return info.join(" ");
    } else {
      const format = $util.fileType(this.FileType);
      if (format) {
        info.push(format);
      }

      if (this.MediaType && this.MediaType !== this.FileType) {
        const media = $util.capitalize(this.MediaType);
        if (media) {
          info.push(media);
        }
      }

      return info.join(" ");
    }
  }

  sizeInfo() {
    let info = [];

    this.addSizeInfo(info);

    return info.join(", ");
  }

  addSizeInfo(info) {
    if (this.Width && this.Height) {
      info.push(this.Width + " × " + this.Height);
    }

    if (this.Size) {
      info.push($util.formatBytes(this.Size));
    }
  }

  toggleLike() {
    this.Favorite = !this.Favorite;

    if (this.Favorite) {
      return $api.post(this.getPhotoResource() + "/like");
    } else {
      return $api.delete(this.getPhotoResource() + "/like");
    }
  }

  getPhotoResource() {
    return "photos/" + this.PhotoUID;
  }

  like() {
    this.Favorite = true;
    return $api.post(this.getPhotoResource() + "/like");
  }

  unlike() {
    this.Favorite = false;
    return $api.delete(this.getPhotoResource() + "/like");
  }

  static getCollectionResource() {
    return "files";
  }

  static getModelName() {
    return $gettext("File");
  }
}

export default File;
