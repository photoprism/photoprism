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

import RestModel from "model/rest";
import Api from "common/api";
import { DateTime } from "luxon";
import { config } from "../session";
import { $gettext } from "common/vm";
import * as src from "../common/src";

export class Marker extends RestModel {
  getDefaults() {
    return {
      UID: "",
      FileUID: "",
      Thumb: "",
      Type: "",
      Src: "",
      Name: "",
      Invalid: false,
      Review: false,
      X: 0.0,
      Y: 0.0,
      W: 0.0,
      H: 0.0,
      CropID: "",
      FaceID: "",
      SubjSrc: "",
      SubjUID: "",
      Score: 0,
      Size: 0,
    };
  }

  route(view) {
    return { name: view, query: { q: "marker:" + this.getId() } };
  }

  classes(selected) {
    let classes = ["is-marker", "uid-" + this.getId()];

    if (this.Invalid) classes.push("is-invalid");
    if (this.Review) classes.push("is-review");
    if (selected) classes.push("is-selected");

    return classes;
  }

  getEntityName() {
    return this.Name;
  }

  getTitle() {
    return this.Name;
  }

  thumbnailUrl(size) {
    if (!size) {
      size = "tile_160";
    }

    if (this.Thumb) {
      return `${config.contentUri}/t/${this.Thumb}/${config.previewToken()}/${size}`;
    } else {
      return `${config.contentUri}/svg/portrait`;
    }
  }

  getDateString() {
    return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
  }

  approve() {
    this.Review = false;
    this.Invalid = false;
    return this.update();
  }

  reject() {
    this.Review = false;
    this.Invalid = true;
    return this.update();
  }

  rename() {
    if (!this.Name || this.Name.trim() === "") {
      // Can't save an empty name.
      return Promise.resolve(this);
    }

    this.SubjSrc = src.Manual;

    const payload = { SubjSrc: this.SubjSrc, Name: this.Name };

    return Api.put(this.getEntityResource(), payload).then((resp) =>
      Promise.resolve(this.setValues(resp.data))
    );
  }

  clearSubject() {
    return Api.delete(this.getEntityResource(this.getId()) + "/subject").then((r) =>
      Promise.resolve(this.setValues(r.data))
    );
  }

  static batchSize() {
    return 48;
  }

  static getCollectionResource() {
    return "markers";
  }

  static getModelName() {
    return $gettext("Marker");
  }
}

export default Marker;
