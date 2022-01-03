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

const SubjPerson = "person";

export class Subject extends RestModel {
  getDefaults() {
    return {
      UID: "",
      Type: "",
      Src: "",
      Slug: "",
      Name: "",
      Alias: "",
      Bio: "",
      Notes: "",
      Favorite: false,
      Hidden: false,
      Private: false,
      Excluded: false,
      FileCount: 0,
      PhotoCount: 0,
      Thumb: "",
      ThumbSrc: "",
      Metadata: {},
      CreatedAt: "",
      UpdatedAt: "",
      DeletedAt: "",
    };
  }

  route(view) {
    if (this.Slug && (!this.Type || this.Type === SubjPerson)) {
      return { name: view, query: { q: `person:${this.Slug}` } };
    }

    return { name: view, query: { q: "subject:" + this.UID } };
  }

  classes(selected) {
    let classes = ["is-subject", "uid-" + this.UID];

    if (this.Favorite) classes.push("is-favorite");
    if (this.Hidden) classes.push("is-hidden");
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
    if (!this.Thumb) {
      return `${config.contentUri}/svg/portrait`;
    }

    if (!size) {
      size = "tile_160";
    }

    return `${config.contentUri}/t/${this.Thumb}/${config.previewToken()}/${size}`;
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

  toggleHidden() {
    this.Hidden = !this.Hidden;

    return Api.put(this.getEntityResource(), { Hidden: this.Hidden });
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
