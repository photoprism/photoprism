/*

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

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
import { $gettext } from "common/vm";
import File from "model/file";
import Photo from "model/photo";
import {
  CodecAv1,
  CodecHvc1,
  CodecOGV,
  CodecVP8,
  CodecVP9,
  FormatAv1,
  FormatAvc,
  FormatHevc,
  FormatWebM,
  MediaAnimated,
} from "model/photo";
import { config } from "app/session";
import { canUseOGV, canUseVP8, canUseVP9, canUseAv1, canUseWebM, canUseHevc } from "common/caniuse";

export class Video extends Model {
  getDefaults() {
    return {
      UID: "",
      PhotoUID: "",
      Hash: "",
      PosterHash: "",
      Title: "",
      Description: "",
      TakenAt: "",
      Favorite: false,
      Playable: false,
      HDR: false,
      Mime: "",
      Type: "",
      Codec: "",
      Width: 640,
      Height: 480,
      Duration: 0,
      FPS: 0,
      Frames: 0,
      Projection: "",
      ColorProfile: "",
      Error: "",
    };
  }

  getId() {
    return this.UID ? this.UID : this.PhotoUID;
  }

  hasId() {
    return !!this.getId();
  }

  toggleLike() {
    this.Favorite = !this.Favorite;

    if (this.Favorite) {
      return Api.post("photos/" + this.PhotoUID + "/like");
    } else {
      return Api.delete("photos/" + this.PhotoUID + "/like");
    }
  }

  url() {
    let hash = this.Hash ? this.Hash : this.PosterHash;

    if (!hash) {
      return `${config.staticUri}/video/404.mp4`;
    }

    if (this.Hash && (this.Codec || this.FileType)) {
      let videoFormat = FormatAvc;

      if (canUseHevc && this.Codec === CodecHvc1) {
        videoFormat = FormatHevc;
      } else if (canUseOGV && this.Codec === CodecOGV) {
        videoFormat = CodecOGV;
      } else if (canUseVP8 && this.Codec === CodecVP8) {
        videoFormat = CodecVP8;
      } else if (canUseVP9 && this.Codec === CodecVP9) {
        videoFormat = CodecVP9;
      } else if (canUseAv1 && this.Codec === CodecAv1) {
        videoFormat = FormatAv1;
      } else if (canUseWebM && this.FileType === FormatWebM) {
        videoFormat = FormatWebM;
      }

      return `${config.videoUri}/videos/${hash}/${config.previewToken}/${videoFormat}`;
    }

    return `${config.videoUri}/videos/${hash}/${config.previewToken}/${FormatAvc}`;
  }

  downloadUrl() {
    return `${config.apiUri}/dl/${this.Hash}?t=${config.downloadToken}`;
  }

  posterUrl(size) {
    let hash = this.PosterHash ? this.PosterHash : this.Hash;

    if (!size) {
      size = "fit_720";
    }

    if (!hash) {
      return `${config.contentUri}/svg/video`;
    }

    return `${config.contentUri}/t/${hash}/${config.previewToken}/${size}`;
  }

  loop() {
    return this.Type === MediaAnimated || (this.Duration >= 0 && this.Duration <= 5000000000);
  }

  playerSize() {
    const vw = Math.max(document.documentElement.clientWidth, window.innerWidth || 0);
    const vh = Math.max(document.documentElement.clientHeight, window.innerHeight || 0);

    let actualWidth = this.Width;
    let actualHeight = this.Height;

    let width = actualWidth;
    let height = actualHeight;

    if (vw < width + 70) {
      let newWidth = vw - 80;
      height = Math.round(newWidth * (actualHeight / actualWidth));
      width = newWidth;
    }

    if (vh < height + 100) {
      let newHeight = vh - 140;
      width = Math.round(newHeight * (actualWidth / actualHeight));
      height = newHeight;
    }

    if (!width || !height) {
      width = 640;
      height = 480;
    }

    return { width, height };
  }

  static notFound() {
    const result = new this();
    result.Title = $gettext("Not Found");
    result.Error = "not found";
    return result;
  }

  static fromPhotos(photos, photosIndex) {
    let videos = [];
    let index = 0;

    if (!photosIndex) {
      photosIndex = 0;
    }

    const n = photos.length;

    for (let i = 0; i < n; i++) {
      const video = this.fromPhoto(photos[i]);
      if (video && !video.Error) {
        videos.push(video);
        if (photosIndex > i) {
          index++;
        }
      }
    }

    return { videos, index };
  }

  static fromPhoto(photo) {
    if (!photo || !photo.Hash) {
      return this.notFound();
    }

    if (!(photo instanceof Photo)) {
      photo = new Photo(photo);
    }

    return photo.video();
  }

  static fromFile(photo, file) {
    if (!file || !file.Hash) {
      return false;
    }

    if (!(file instanceof File)) {
      file = new File(file);
    }

    if (!file.isPlayable()) {
      return false;
    }

    const video = file.video();

    if (photo) {
      video.Title = photo.Title;
      video.Description = photo.Description;
      video.Favorite = photo.Favorite;
    }

    return video;
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

        let video = this.fromFile(p, f);

        if (video && !video.Error) {
          result.push(video);
        }
      }
    }

    return result;
  }
}

export default Video;
