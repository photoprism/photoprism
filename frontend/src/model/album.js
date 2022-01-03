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

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

import RestModel from "model/rest";
import Api from "common/api";
import countries from "options/countries.json";
import { DateTime } from "luxon";
import { config } from "../session";
import { $gettext } from "common/vm";

export class Album extends RestModel {
  getDefaults() {
    return {
      UID: "",
      ParentUID: "",
      Thumb: "",
      ThumbSrc: "",
      Path: "",
      Slug: "",
      Type: "",
      Title: "",
      Location: "",
      Caption: "",
      Category: "",
      Description: "",
      Notes: "",
      Filter: "",
      Order: "",
      Template: "",
      State: "",
      Country: "",
      Day: -1,
      Year: -1,
      Month: -1,
      Favorite: false,
      Private: false,
      PhotoCount: 0,
      LinkCount: 0,
      CreatedAt: "",
      UpdatedAt: "",
    };
  }

  route(view) {
    return {
      name: view,
      params: { uid: this.UID, slug: this.Slug, year: this.Year, month: this.Month },
    };
  }

  classes(selected) {
    let classes = ["is-album", "uid-" + this.UID, "type-" + this.Type];

    if (this.Favorite) classes.push("is-favorite");
    if (this.Private) classes.push("is-private");
    if (selected) classes.push("is-selected");

    return classes;
  }

  getEntityName() {
    return this.Slug;
  }

  getTitle() {
    return this.Title;
  }

  hasCountry() {
    return this.Country !== "" && this.Country !== "zz";
  }

  getCountry() {
    if (!this.hasCountry()) {
      return "";
    }

    const country = countries.find((c) => c.Code === this.Country);

    if (country) {
      return country.Name;
    }

    return "";
  }

  hasLocation() {
    return this.Location !== "" || this.State !== "" || this.hasCountry();
  }

  getLocation() {
    if (this.Location !== "" && this.Location !== this.Title) {
      return this.Location;
    }

    const country = this.getCountry();

    if (country !== "" && !this.Title.includes(country)) {
      if (this.State !== "" && this.State !== this.Title) {
        return `${this.State}, ${country}`;
      }

      return country;
    }

    if (this.State !== "" && !this.Title.includes(this.State)) {
      return this.State;
    }

    return "";
  }

  thumbnailUrl(size) {
    if (this.Thumb) {
      return `${config.contentUri}/t/${this.Thumb}/${config.previewToken()}/${size}`;
    } else if (this.UID) {
      return `${config.contentUri}/albums/${this.UID}/t/${config.previewToken()}/${size}`;
    } else {
      return `${config.contentUri}/svg/album`;
    }
  }

  dayString() {
    if (!this.Day || this.Day <= 0) {
      return "01";
    }

    return this.Day.toString().padStart(2, "0");
  }

  monthString() {
    if (!this.Month || this.Month <= 0) {
      return "01";
    }

    return this.Month.toString().padStart(2, "0");
  }

  yearString() {
    if (!this.Year || this.Year <= 1000) {
      return new Date().getFullYear().toString().padStart(4, "0");
    }

    return this.Year.toString();
  }

  getDate() {
    let date = this.yearString() + "-" + this.monthString() + "-" + this.dayString();

    return DateTime.fromISO(`${date}T12:00:00Z`).toUTC();
  }

  getDateString() {
    if (!this.Year || this.Year <= 1000) {
      return $gettext("Unknown");
    } else if (!this.Month || this.Month <= 0) {
      return this.Year.toString();
    } else if (!this.Day || this.Day <= 0) {
      return this.getDate().toLocaleString({ month: "long", year: "numeric" });
    }

    return this.getDate().toLocaleString(DateTime.DATE_HUGE);
  }

  getCreatedString() {
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
    return 24;
  }

  static getCollectionResource() {
    return "albums";
  }

  static getModelName() {
    return $gettext("Album");
  }
}

export default Album;
