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
import * as src from "common/src";

export let BatchSize = 48;

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
      return `${config.contentUri}/t/${this.Thumb}/${config.previewToken}/${size}`;
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
    return BatchSize;
  }

  static setBatchSize(count) {
    const s = parseInt(count);
    if (!isNaN(s) && s >= 24) {
      BatchSize = s;
    }
  }

  static getCollectionResource() {
    return "markers";
  }

  static getModelName() {
    return $gettext("Marker");
  }
}

export default Marker;
