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

  // Resolves to the full Photo entity for this slide, fetched via
  // the shared LRU cache. Each call returns a fresh hydrated
  // instance so consumers can mutate locally without aliasing the
  // cached snapshot. Resolves to an empty Photo placeholder when
  // this thumb has no UID. Lightbox uses this to load sidebar
  // metadata for the current slide and to warm neighbours via
  // Photo.prefetchAround. Rejections (e.g. ModelCacheStaleFetchError
  // after a logout-clear) propagate to callers' .catch handlers.
  loadPhoto() {
    if (!this.UID) {
      return Promise.resolve(new Photo());
    }
    return Photo.findCached(this.UID);
  }

  // Drops the cached Photo entity for this slide so the next
  // loadPhoto() rehydrates from GET /photos/:uid. Used by flows
  // that mutate the photo without firing a photos.* WS event the
  // cache subscribes to (currently: album-membership changes,
  // which only publish albums.updated). archive / restore / delete
  // already auto-evict via the photo.js WS subscribers, so call
  // sites for those don't need this.
  evictPhoto() {
    if (this.UID) {
      Photo.evictCache(this.UID);
    }
  }

  // Moves this photo to the archive (soft delete) and flips the
  // local Archived flag immediately so menu buttons re-render
  // without waiting for the API round-trip. The pre-call value of
  // Archived is captured in a closure and restored on rejection —
  // a literal `false` rollback would be wrong if the field was
  // undefined (defaults aren't declared) or already true (no-op
  // re-archive). Resolves on success; the backend publishes
  // photos.archived which the photo cache subscribes to for
  // automatic eviction. Mirrors the batch-photos endpoint that
  // view/cards.vue already targets via Photo.prototype.archive —
  // kept on Thumb for the lightbox flow which holds a Thumb (not a
  // Photo) for the current slide. Rapid-fire prevention is the
  // caller's responsibility; if a UI flow needs it, it can add
  // `blockUI("busy")` / `unblockUI()` or a similar guard at the
  // handler layer without making the model stateful.
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

  // Formats Lat/Lng as an EXIF-style coordinate pair separated by
  // an em-space (U+2003) — see the literal in the body. Returns a
  // 0/0 placeholder when coordinates are missing so the sidebar
  // EXIF row doesn't collapse.
  getLatLng() {
    if (!this.Lat || !this.Lng) {
      return `0°N\u20030°E`;
    }

    return `${this.Lat.toFixed(5)}°N\u2003${this.Lng.toFixed(5)}°E`;
  }

  // Copies Lat/Lng to the system clipboard as `lat,lng` decimals so
  // they paste cleanly into mapping tools. No-ops when coordinates
  // are missing rather than copying a misleading "0,0".
  copyLatLng() {
    // Abort if latitude or longitude are not set.
    if (!this.Lat || !this.Lng) {
      return;
    }

    // Use the browser API to copy the coordinates to the clipboard.
    $util.copyText(`${this.Lat.toString()},${this.Lng.toString()}`);
  }

  // Returns a rounded megapixel string (e.g. "12.0MP") for the
  // type-info row. Returns the literal "0.0MP" — not 0 or empty —
  // when dimensions are unknown; getTypeInfo() uses that sentinel
  // to decide whether to skip the MP segment.
  getMegaPixels() {
    if (!this.Width || !this.Height) {
      return "0.0MP";
    }

    return `${((this.Width * this.Height) / 1000000).toFixed(1)}MP`;
  }

  // Returns the Material Design icon name for this slide's media
  // type, used by the lightbox type chip. Falls back to a generic
  // image icon for unknown types.
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

  // Builds the EXIF-summary string shown next to the type chip
  // (codec / megapixels / dimensions, joined by em-spaces). The
  // segment order varies by media type so the most useful field
  // leads — duration first for video, codec for raw, etc. Returns
  // localized "Document" for documents and may return an empty
  // string when nothing useful is known.
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

  // Returns a placeholder Thumb-shaped object for slides that can't
  // be rendered (missing hash, deleted file, etc.). Every Thumbs
  // entry points at the static 404 image so the lightbox grid stays
  // visually consistent without throwing on missing assets. Returns
  // a plain object (not a Thumb instance) — callers that need an
  // instance wrap the result themselves.
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

  // Builds a Thumb array from a Photos search response. Each photo
  // flows through fromPhoto(), which picks the best available file
  // for thumbnail rendering (RAW/Live preferred over JPEG).
  static fromPhotos(photos) {
    let result = [];
    const n = photos.length;

    for (let i = 0; i < n; i++) {
      result.push(this.fromPhoto(photos[i]));
    }

    return result;
  }

  // Constructs a Thumb from a Photo entity, choosing the original
  // file (RAW/Live preferred over JPEG via Photo.originalFile()) to
  // source hash/dimensions/codec. Falls back to the Photo's own
  // top-level fields when no Files are available. Returns a
  // notFound() placeholder when neither hash nor files exist.
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

  // Constructs a Thumb from a specific File belonging to a Photo —
  // used by the file-list view to surface each file as its own
  // slide. The Photo provides title / location / EXIF metadata; the
  // File provides hash / dimensions / codec. Returns notFound()
  // when any required input or hash is missing.
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

  // Wraps an array of plain Thumb-shaped values as Thumb instances.
  // Used by endpoints that already return Thumb-shaped JSON (e.g.
  // /photos/view), bypassing the fromPhoto / fromFile mappers.
  static wrap(data) {
    return data.map((values) => new this(values));
  }

  // Like fromPhotos but expands each photo's Files[] into one Thumb
  // per JPEG/PNG file — used by stack views where every variant
  // should be its own slide. Skips photos without files and any
  // file types other than jpg/png.
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

  // Scales a file's actual dimensions to fit within a (width,
  // height) box, preserving aspect ratio. Returns the file's
  // native size unchanged when the box is already big enough so we
  // don't upscale low-resolution thumbnails. Mirrors
  // Photo.calculateSize but operates on a File rather than a Photo
  // so file/edit views can size individual files.
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

  // Builds the cached-thumbnail URL for a file at the given size
  // (one of the keys from window.__CONFIG__.thumbs). Returns the
  // static 404 image when the file has no hash so the lightbox grid
  // still renders without throwing on broken assets.
  static thumbnailUrl(file, size) {
    if (!file.Hash) {
      return `${$config.staticUri}/img/404.jpg`;
    }

    return `${$config.contentUri}/t/${file.Hash}/${$config.previewToken}/${size}`;
  }

  // Builds the original-file download URL for a file. Returns an
  // empty string (not a "broken" URL) when the file has no hash so
  // consumers can guard with truthiness.
  static downloadUrl(file) {
    if (!file || !file.Hash) {
      return "";
    }

    return `${$config.apiUri}/dl/${file.Hash}?t=${$config.downloadToken}`;
  }
}

export default Thumb;
