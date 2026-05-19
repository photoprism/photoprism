import Model from "model.js";
import Photo from "model/photo";
import $api from "common/api";
import $util from "common/util";
import { $config } from "app/session.js";
import { $gettext } from "common/gettext";

const thumbs = window.__CONFIG__.thumbs;

// Thumb represents a lightweight slide/photo preview record used by the lightbox.
export class Thumb extends Model {
  // Returns the default field shape for a Thumb. These fields are
  // all reactive once the instance is wrapped by Vue's data() proxy
  // and define the snapshot served when no server data is available.
  // `Archived` and `Removed` are intentionally NOT declared here so
  // the lightbox's tri-state visibility checks (e.g. the explicit
  // `this.model?.Archived === false` at lightbox.vue:1437) can
  // distinguish "never set" from "explicitly not archived".
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

  // Returns the canonical identifier for this slide, preferring UID
  // over numeric ID. Returns `false` when neither is set so callers
  // can branch on truthiness.
  getId() {
    if (this.UID) {
      return this.UID;
    }

    return this.ID ? this.ID : false;
  }

  // Convenience predicate around getId() — true when this Thumb
  // represents a real photo (vs. a notFound() placeholder).
  hasId() {
    return !!this.getId();
  }

  // Toggles the favorite flag and posts/deletes the like to the
  // backend. Flips `Favorite` synchronously first so the heart icon
  // re-renders immediately; the API call is fire-and-forget. No
  // rollback on failure — matches the existing optimistic-toggle
  // pattern used elsewhere in the frontend.
  toggleLike() {
    this.Favorite = !this.Favorite;

    if (this.Favorite) {
      return $api.post("photos/" + this.UID + "/like");
    } else {
      return $api.delete("photos/" + this.UID + "/like");
    }
  }

  // loadPhoto resolves to the full Photo entity for this slide via the LRU
  // cache, or an empty Photo placeholder when this thumb has no UID.
  loadPhoto() {
    if (!this.UID) {
      return Promise.resolve(new Photo());
    }
    return Photo.findCached(this.UID);
  }

  // evictPhoto drops the cached Photo for this slide so the next loadPhoto()
  // rehydrates from /photos/:uid. Use for mutations without a photos.* WS event.
  evictPhoto() {
    if (this.UID) {
      Photo.evictCache(this.UID);
    }
  }

  // archive moves the photo to the archive (soft delete) and optimistically
  // flips Archived so menu buttons re-render before the round-trip; rollback
  // restores the captured prior value (not a literal false) on rejection.
  archive() {
    const prev = this.Archived;
    this.Archived = true;
    return $api.post("batch/photos/archive", { photos: [this.UID] }).catch((err) => {
      this.Archived = prev;
      throw err;
    });
  }

  // Restores this photo from the archive. Captures the pre-call
  // Archived value and restores it on rejection (mirroring
  // archive()) so a no-op restore on an already-restored photo
  // doesn't leave Archived === true. Resolves on success; the
  // backend publishes photos.restored for automatic cache eviction.
  restore() {
    const prev = this.Archived;
    this.Archived = false;
    return $api.post("batch/photos/restore", { photos: [this.UID] }).catch((err) => {
      this.Archived = prev;
      throw err;
    });
  }

  // Removes this photo from the given album. Optimistic flip on
  // Removed (drives menu visibility) with previous-value rollback
  // on rejection. Backend publishes only albums.updated (not a
  // photos event), so callers that mutate the sidebar's cached
  // Photo.Albums list MUST also call evictPhoto() — see
  // lightbox.vue onRemoveFromAlbum for the pattern.
  removeFromAlbum(albumUID) {
    const prev = this.Removed;
    this.Removed = true;
    return $api.delete(`albums/${albumUID}/photos`, { data: { photos: [this.UID] } }).catch((err) => {
      this.Removed = prev;
      throw err;
    });
  }

  // getLatLng formats Lat/Lng as EXIF-style coordinates (en-space separator),
  // returning a 0/0 placeholder when coordinates are missing so the row holds.
  getLatLng() {
    if (!this.Lat || !this.Lng) {
      return `0°N\u20030°E`;
    }

    return `${this.Lat.toFixed(5)}°N\u2002${this.Lng.toFixed(5)}°E`;
  }

  // getLatLngShort formats Lat/Lng rounded to ~11 m precision for the sidebar
  // Location row; returns an empty string when coordinates are missing.
  getLatLngShort() {
    if (!this.Lat || !this.Lng) {
      return "";
    }

    return `${this.Lat.toFixed(4)}°N\u2002${this.Lng.toFixed(4)}°E`;
  }

  // copyLatLng copies coordinates to the clipboard as `lat,lng` decimals;
  // no-op when coordinates are missing (avoid pasting a misleading "0,0").
  copyLatLng() {
    if (!this.Lat || !this.Lng) {
      return;
    }
    $util.copyText(`${this.Lat.toString()},${this.Lng.toString()}`);
  }

  // getMegaPixels returns a rounded megapixel string (e.g. "12.0MP"); returns
  // the literal "0.0MP" when dimensions are unknown — getTypeInfo skips on that.
  getMegaPixels() {
    if (!this.Width || !this.Height) {
      return "0.0MP";
    }

    return `${((this.Width * this.Height) / 1000000).toFixed(1)}MP`;
  }

  // getTypeIcon returns the Material Design icon name for the type chip; falls back to mdi-image.
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

  // getTypeInfo builds the codec/megapixels/dimensions summary next to the
  // type chip. Segment order varies by media type so the most useful field
  // leads (duration for video, codec for raw); may return an empty string.
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

  // notFound returns a placeholder Thumb-shaped object for slides that can't
  // be rendered (missing hash, deleted file); each Thumbs entry uses the 404 image.
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

  // fromPhotos builds a Thumb array from a Photos search response via fromPhoto.
  static fromPhotos(photos) {
    let result = [];
    const n = photos.length;

    for (let i = 0; i < n; i++) {
      result.push(this.fromPhoto(photos[i]));
    }

    return result;
  }

  // fromPhoto builds a Thumb from a Photo entity using originalFile() (RAW/Live
  // preferred over JPEG) and falling back to top-level fields when Files is empty.
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

  // fromFile builds a Thumb from a specific File of a Photo for the file-list view;
  // the Photo provides metadata, the File provides hash/dimensions/codec.
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

  // wrap turns plain Thumb-shaped values into Thumb instances, bypassing the
  // fromPhoto / fromFile mappers for endpoints that already return Thumb shape.
  static wrap(data) {
    return data.map((values) => new this(values));
  }

  // fromFiles is like fromPhotos but expands each photo's Files[] into one
  // Thumb per jpg/png file; used by stack views. Other file types are skipped.
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

  // calculateSize scales a file's dimensions to fit within a (width, height)
  // box, preserving aspect ratio; never upscales (returns native size when smaller).
  static calculateSize(file, width, height) {
    if (width >= file.Width && height >= file.Height) {
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

  // thumbnailUrl builds the cached-thumbnail URL for the given file and size;
  // returns the static 404 image when the file has no hash.
  static thumbnailUrl(file, size) {
    if (!file.Hash) {
      return `${$config.staticUri}/img/404.jpg`;
    }

    return `${$config.contentUri}/t/${file.Hash}/${$config.previewToken}/${size}`;
  }

  // downloadUrl builds the original-file download URL; returns "" when the
  // file has no hash so consumers can guard with truthiness.
  static downloadUrl(file) {
    if (!file || !file.Hash) {
      return "";
    }

    return `${$config.apiUri}/dl/${file.Hash}?t=${$config.downloadToken}`;
  }
}

export default Thumb;
