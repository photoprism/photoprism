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

import RestModel from "model/rest";
import Api from "common/api";
import { DateTime } from "luxon";
import { config } from "app/session";
import { $gettext } from "common/vm";

export let BatchSize = 24;

export class Label extends RestModel {
  getDefaults() {
    return {
      ID: 0,
      UID: "",
      Slug: "",
      CustomSlug: "",
      Name: "",
      Priority: 0,
      Favorite: false,
      Description: "",
      Notes: "",
      PhotoCount: 0,
      CreatedAt: "",
      UpdatedAt: "",
      DeletedAt: "",
    };
  }

  route(view) {
    return { name: view, query: { q: "label:" + (this.CustomSlug ? this.CustomSlug : this.Slug) } };
  }

  classes(selected) {
    let classes = ["is-label", "uid-" + this.UID];

    if (this.Favorite) classes.push("is-favorite");
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
    if (this.Thumb) {
      return `${config.contentUri}/t/${this.Thumb}/${config.previewToken}/${size}`;
    } else if (this.UID) {
      return `${config.contentUri}/labels/${this.UID}/t/${config.previewToken}/${size}`;
    } else {
      return `${config.contentUri}/svg/label`;
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
    return BatchSize;
  }

  static setBatchSize(count) {
    const s = parseInt(count);
    if (!isNaN(s) && s >= 24) {
      BatchSize = s;
    }
  }

  static getCollectionResource() {
    return "labels";
  }

  static getModelName() {
    return $gettext("Label");
  }
}

export default Label;
