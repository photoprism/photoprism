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

import RestModel from "model/rest";
import Api from "common/api";
import { DateTime } from "luxon";
import { config } from "../session";
import { $gettext } from "common/vm";

export class Subject extends RestModel {
  getDefaults() {
    return {
      UID: "",
      MarkerUID: "",
      MarkerSrc: "",
      Type: "",
      Src: "",
      Slug: "",
      Name: "",
      Alias: "",
      Bio: "",
      Notes: "",
      Favorite: false,
      Private: false,
      Excluded: false,
      FileCount: 0,
      FileHash: "",
      CropArea: "",
      Metadata: {},
      CreatedAt: "",
      UpdatedAt: "",
      DeletedAt: "",
    };
  }

  route(view) {
    return { name: view, query: { q: "subject:" + this.UID } };
  }

  classes(selected) {
    let classes = ["is-subject", "uid-" + this.UID];

    if (this.Favorite) classes.push("is-favorite");
    if (this.Private) classes.push("is-private");
    if (this.Excluded) classes.push("is-excluded");
    if (selected) classes.push("is-selected");

    return classes;
  }

  getEntityName() {
    return this.Slug;
  }

  getTitle() {
    return this.Name;
  }

  thumbnailUrl(size) {
    if (!this.FileHash) {
      return `${config.contentUri}/svg/portrait`;
    }

    if (!size) {
      size = "tile_160";
    }

    if (this.CropArea && (size === "tile_160" || size === "tile_320")) {
      return `${config.contentUri}/t/${this.FileHash}/${config.previewToken()}/${size}/${
        this.CropArea
      }`;
    } else {
      return `${config.contentUri}/t/${this.FileHash}/${config.previewToken()}/${size}`;
    }
  }

  getDateString() {
    return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
  }

  toggleLike() {
    this.Favorite = !this.Favorite;

    if (this.Favorite) {
      return Api.post(this.getEntityResource() + "/like");
    } else {
      return Api.delete(this.getEntityResource() + "/like");
    }
  }

  like() {
    this.Favorite = true;
    return Api.post(this.getEntityResource() + "/like");
  }

  unlike() {
    this.Favorite = false;
    return Api.delete(this.getEntityResource() + "/like");
  }

  static batchSize() {
    return 60;
  }

  static getCollectionResource() {
    return "subjects";
  }

  static getModelName() {
    return $gettext("Subject");
  }
}

export default Subject;
