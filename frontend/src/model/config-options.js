/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

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

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

import Api from "common/api";
import Model from "./model";
import { config } from "../session";

export class ConfigOptions extends Model {
  getDefaults() {
    return {
      Debug: config.values.debug,
      ReadOnly: config.values.readonly,
      Experimental: config.values.experimental,
      OriginalsLimit: 0,
      Workers: 0,
      WakeupInterval: 0,
      DisableBackups: config.values.disable.backups,
      DisableWebDAV: config.values.disable.webdav,
      DisableSettings: config.values.disable.settings,
      DisablePlaces: config.values.disable.places,
      DisableExifTool: config.values.disable.exiftool,
      DisableDarktable: config.values.disable.darktable,
      DisableRawtherapee: config.values.disable.rawtherapee,
      DisableSips: config.values.disable.sips,
      DisableHeifConvert: config.values.disable.heifconvert,
      DisableFFmpeg: config.values.disable.ffmpeg,
      DisableTensorFlow: config.values.disable.tensorflow,
      DetectNSFW: false,
      UploadNSFW: config.values.uploadNSFW,
      RawPresets: false,
      ThumbUncached: true,
      ThumbFilter: "",
      ThumbSize: 0,
      ThumbSizeUncached: 0,
      JpegSize: 0,
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
    return Api.post("config/options", this.getValues(true)).then((response) =>
      Promise.resolve(this.setValues(response.data))
    );
  }
}

export default ConfigOptions;
