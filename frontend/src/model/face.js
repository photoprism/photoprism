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

import Marker from "model/marker";
import RestModel from "model/rest";
import { DateTime } from "luxon";
import { config } from "../session";
import { $gettext } from "common/vm";

export class Face extends RestModel {
  constructor(values) {
    if (values && values.Marker) {
      values.Marker = new Marker(values.Marker);
    }

    super(values);
  }

  getDefaults() {
    return {
      ID: "",
      Src: "",
      SubjUID: "",
      Samples: 0,
      SampleRadius: 0.0,
      Collisions: 0,
      CollisionRadius: 0.0,
      Marker: new Marker(),
      Hidden: false,
      MatchedAt: "",
      CreatedAt: "",
      UpdatedAt: "",
    };
  }

  route(view) {
    return { name: view, query: { q: "face:" + this.ID } };
  }

  classes(selected) {
    let classes = ["is-face", "uid-" + this.UID];

    if (this.Hidden) classes.push("is-hidden");
    if (selected) classes.push("is-selected");

    return classes;
  }

  getEntityName() {
    return this.ID;
  }

  getTitle() {
    return this.Name;
  }

  thumbnailUrl(size) {
    if (!this.Marker || !this.Marker.FileHash) {
      return `${config.contentUri}/svg/portrait`;
    }

    if (!size) {
      size = "tile_160";
    }

    if (this.Marker.CropArea && (size === "tile_160" || size === "tile_320")) {
      return `${config.contentUri}/t/${this.Marker.FileHash}/${config.previewToken()}/${size}/${
        this.Marker.CropArea
      }`;
    } else {
      return `${config.contentUri}/t/${this.Marker.FileHash}/${config.previewToken()}/${size}`;
    }
  }

  getDateString() {
    return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
  }

  show() {
    this.Hidden = false;
    return this.update();
  }

  hide() {
    this.Hidden = true;
    return this.update();
  }

  static batchSize() {
    return 60;
  }

  static getCollectionResource() {
    return "faces";
  }

  static getModelName() {
    return $gettext("Face");
  }
}

export default Face;
