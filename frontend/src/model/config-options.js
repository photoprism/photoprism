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

import Api from "common/api";
import Model from "./model";
import { config } from "app/session";

export class ConfigOptions extends Model {
  getDefaults() {
    return {
      Debug: config.values.debug,
      ReadOnly: config.values.readonly,
      Experimental: config.values.experimental,
      OriginalsLimit: 0,
      Workers: 0,
      WakeupInterval: 0,
      BackupDatabase: false,
      DisableWebDAV: config.values.disable.webdav,
      DisableSettings: config.values.disable.settings,
      DisablePlaces: config.values.disable.places,
      DisableBackups: config.values.disable.backups,
      DisableTensorFlow: config.values.disable.tensorflow,
      DisableSips: config.values.disable.sips,
      DisableFFmpeg: config.values.disable.ffmpeg,
      DisableExifTool: config.values.disable.exiftool,
      DisableDarktable: config.values.disable.darktable,
      DisableRawTherapee: config.values.disable.rawtherapee,
      DisableImageMagick: config.values.disable.imagemagick,
      DisableHeifConvert: config.values.disable.heifconvert,
      DisableVectors: config.values.disable.vectors,
      DisableJpegXL: config.values.disable.jpegxl,
      DisableRaw: config.values.disable.raw,
      DetectNSFW: false,
      UploadNSFW: config.values.uploadNSFW,
      RawPresets: false,
      ThumbUncached: true,
      ThumbFilter: "",
      ThumbSize: 0,
      ThumbSizeUncached: 0,
      JpegSize: 0,
      PngSize: 0,
      JpegQuality: 0,
      SiteUrl: config.values.siteUrl,
      SitePreview: config.values.siteUrl,
      SiteTitle: config.values.siteTitle,
      SiteCaption: config.values.siteCaption,
      SiteDescription: config.values.siteDescription,
      SiteAuthor: config.values.siteAuthor,
    };
  }

  changed(area, key) {
    if (typeof this.__originalValues[area] === "undefined") {
      return false;
    }

    return this[area][key] !== this.__originalValues[area][key];
  }

  load() {
    return Api.get("config/options").then((response) => {
      return Promise.resolve(this.setValues(response.data));
    });
  }

  save() {
    return Api.post("config/options", this.getValues(true)).then((response) => Promise.resolve(this.setValues(response.data)));
  }
}

export default ConfigOptions;
