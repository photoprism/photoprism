import Model from "model.js";
import $api from "common/api";
import $util from "common/util";
import { $config } from "app/session.js";
import { $gettext } from "common/gettext";

const thumbs = window.__CONFIG__.thumbs;

// Thumb represents a lightweight slide/photo preview record used by the lightbox.
export class Thumb extends Model {
  getDefaults() {
    return {
      UID: "",
      Type: "image",
      Title: "",
      Caption: "",
      Lat: 0.0,
      Lng: 0.0,
      TakenAtLocal: "",
      TimeZone: "",
      Favorite: false,
      Playable: false,
      Duration: 0,
      Width: 0,
      Height: 0,
      Hash: "",
      Codec: "",
      Mime: "",
      Thumbs: {},
      DownloadUrl: "",
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
      return $api.post("photos/" + this.UID + "/like");
    } else {
      return $api.delete("photos/" + this.UID + "/like");
    }
  }

  getLatLng() {
    if (!this.Lat || !this.Lng) {
      return `0°N\u20030°E`;
    }

    return `${this.Lat.toFixed(5)}°N\u2003${this.Lng.toFixed(5)}°E`;
  }

  copyLatLng() {
    // Abort if latitude or longitude are not set.
    if (!this.Lat || !this.Lng) {
      return;
    }

    // Use the browser API to copy the coordinates to the clipboard.
    $util.copyText(`${this.Lat.toString()},${this.Lng.toString()}`);
  }

  getMegaPixels() {
    if (!this.Width || !this.Height) {
      return "0.0MP";
    }

    return `${((this.Width * this.Height) / 1000000).toFixed(1)}MP`;
  }

  getTypeIcon() {
    switch (this.Type) {
      case "raw":
        return "mdi-raw";
      case "video":
        return "mdi-video";
      case "animated":
        return "mdi-file-gif-box";
      case "vector":
        return "mdi-vector-polyline";
      case "document":
        return "mdi-file-pdf-box";
      case "live":
        return "mdi-play-circle-outline";
      default:
        return "mdi-image";
    }
  }

  getTypeInfo() {
    let info = [];
    const mp = this.getMegaPixels();

    switch (this.Type) {
      case "image":
        if (this.Codec) {
          info.push($util.formatCodec(this.Codec));
        }

        if (mp !== "0.0MP") {
          info.push(mp);
        }

        if (this.Width) {
          info.push(`${this.Width}×${this.Height}`);
        }
        break;
      case "raw":
      case "vector":
        if (this.Codec && this.Codec !== "jpeg") {
          info.push($util.formatCodec(this.Codec));
        }

        if (mp !== "0.0MP") {
          info.push(mp);
        }

        if (this.Width) {
          info.push(`${this.Width}×${this.Height}`);
        }
        break;
      case "live":
      case "video":
      case "animated":
        if (this.Duration) {
          info.push($util.formatDuration(this.Duration));
        }

        if (mp !== "0.0MP") {
          info.push(mp);
        } else if (this.Codec && this.Codec !== "jpeg") {
          info.push($util.formatCodec(this.Codec));
        }

        if (this.Width) {
          info.push(`${this.Width}×${this.Height}`);
        }

        break;
      case "document":
        info.push($gettext("Document"));
        break;
      default:
        if (this.Codec && this.Codec !== "jpeg") {
          info.push($util.formatCodec(this.Codec));
        }

        if (mp !== "0.0MP") {
          info.push(mp);
        }

        if (this.Width) {
          info.push(`${this.Width}×${this.Height}`);
        }
    }

    return info.join("\u2003");
  }

  static notFound() {
    const result = {
      UID: "",
      Type: "image",
      Title: $gettext("Invalid photo selected"),
      Caption: "",
      Lat: 0.0,
      Lng: 0.0,
      TakenAtLocal: "",
      TimeZone: "",
      Favorite: false,
      Playable: false,
      Duration: 0,
      Width: 0,
      Height: 0,
      Hash: "",
      Codec: "",
      Mime: "",
      Thumbs: {},
      DownloadUrl: "",
    };

    for (let i = 0; i < thumbs.length; i++) {
      let t = thumbs[i];

      result.Thumbs[t.size] = {
        w: t.w,
        h: t.h,
        src: `${$config.staticUri}/img/404.jpg`,
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
    if (!photo || (!photo.Hash && !photo.Files?.length)) {
      return this.notFound();
    }

    let file, width, height, hash, codec, mime;

    if (photo.Files?.length) {
      file = photo.originalFile();
    }

    if (file) {
      width = file.Width ? file.Width : photo.Width;
      height = file.Height ? file.Height : photo.Height;
      hash = file.Hash ? file.Hash : photo.Hash;
      codec = file.Codec ? file.Codec : photo.videoCodec();
      mime = file.Mime ? file.Mime : photo.videoContentType();
    } else {
      width = photo.Width;
      height = photo.Height;
      hash = photo.Hash;
      codec = photo.videoCodec();
      mime = photo.videoContentType();
    }

    const result = {
      UID: photo.UID,
      Type: photo.Type,
      Title: photo.Title,
      Caption: photo.Caption,
      Lat: photo.Lat,
      Lng: photo.Lng,
      TakenAtLocal: photo.TakenAtLocal,
      TimeZone: photo.TimeZone,
      Favorite: photo.Favorite,
      Playable: photo.isPlayable(),
      Duration: photo.Duration,
      Width: width,
      Height: height,
      Hash: hash,
      Codec: codec,
      Mime: mime,
      Thumbs: {},
      DownloadUrl: this.downloadUrl(photo),
    };

    for (let i = 0; i < thumbs.length; i++) {
      let t = thumbs[i];
      let size = photo.calculateSize(t.w, t.h);

      result.Thumbs[t.size] = {
        w: size.width,
        h: size.height,
        src: photo.thumbnailUrl(t.size),
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
      Type: file.MediaType ? file.MediaType : photo.Type,
      Title: photo.Title,
      Caption: photo.Caption,
      Lat: photo.Lat,
      Lng: photo.Lng,
      TakenAtLocal: photo.TakenAtLocal,
      TimeZone: photo.TimeZone,
      Favorite: photo.Favorite,
      Playable: photo.isPlayable(),
      Duration: photo.Duration,
      Width: file.Width,
      Height: file.Height,
      Hash: file.Hash,
      Codec: file.Codec,
      Mime: file.Mime,
      Thumbs: {},
      DownloadUrl: this.downloadUrl(file),
    };

    for (let i = 0; i < thumbs.length; i++) {
      let t = thumbs[i];
      let size = this.calculateSize(file, t.w, t.h);

      result.Thumbs[t.size] = {
        w: size.width,
        h: size.height,
        src: this.thumbnailUrl(file, t.size),
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
      newH = Math.ceil(newW / srcAspectRatio);
    } else {
      newH = height;
      newW = Math.ceil(newH * srcAspectRatio);
    }

    return { width: newW, height: newH };
  }

  static thumbnailUrl(file, size) {
    if (!file.Hash) {
      return `${$config.staticUri}/img/404.jpg`;
    }

    return `${$config.contentUri}/t/${file.Hash}/${$config.previewToken}/${size}`;
  }

  static downloadUrl(file) {
    if (!file || !file.Hash) {
      return "";
    }

    return `${$config.apiUri}/dl/${file.Hash}?t=${$config.downloadToken}`;
  }
}

export default Thumb;
