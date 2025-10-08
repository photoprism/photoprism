import $api from "common/api";
import Model from "./model";
import { $config } from "app/session";

// ConfigOptions represents the editable server configuration that powers the UI toggles.
export class ConfigOptions extends Model {
  getDefaults() {
    return {
      Debug: $config.values.debug,
      ReadOnly: $config.values.readonly,
      Experimental: $config.values.experimental,
      OriginalsLimit: 0,
      Workers: 0,
      WakeupInterval: 0,
      BackupAlbums: true,
      BackupDatabase: true,
      BackupSchedule: "",
      BackupRetain: 3,
      SidecarYaml: true,
      DisableRestart: true,
      DisableWebDAV: $config.values.disable.webdav,
      DisableSettings: $config.values.disable.settings,
      DisablePlaces: $config.values.disable.places,
      DisableBackups: $config.values.disable.backups,
      DisableTensorFlow: $config.values.disable.tensorflow,
      DisableSips: $config.values.disable.sips,
      DisableFFmpeg: $config.values.disable.ffmpeg,
      DisableExifTool: $config.values.disable.exiftool,
      DisableDarktable: $config.values.disable.darktable,
      DisableRawTherapee: $config.values.disable.rawtherapee,
      DisableImageMagick: $config.values.disable.imagemagick,
      DisableHeifConvert: $config.values.disable.heifconvert,
      DisableVectors: $config.values.disable.vectors,
      DisableJpegXL: $config.values.disable.jpegxl,
      DisableRaw: $config.values.disable.raw,
      DetectNSFW: false,
      UploadNSFW: $config.values.uploadNSFW,
      RawPresets: false,
      ThumbUncached: true,
      ThumbLibrary: "",
      ThumbColor: "",
      ThumbFilter: "",
      ThumbSize: 0,
      ThumbSizeUncached: 0,
      JpegSize: 0,
      PngSize: 0,
      JpegQuality: 0,
      SiteUrl: $config.values.siteUrl,
      SitePreview: $config.values.siteUrl,
      SiteTitle: $config.values.siteTitle,
      SiteCaption: $config.values.siteCaption,
      SiteDescription: $config.values.siteDescription,
      SiteAuthor: $config.values.siteAuthor,
    };
  }

  changed(area, key) {
    if (typeof this.__originalValues[area] === "undefined") {
      return false;
    }

    return this[area][key] !== this.__originalValues[area][key];
  }

  load() {
    return $api.get("config/options").then((response) => {
      return Promise.resolve(this.setValues(response.data));
    });
  }

  save() {
    return $api
      .post("config/options", this.getValues(true))
      .then((response) => Promise.resolve(this.setValues(response.data)));
  }
}

export default ConfigOptions;
