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

import Marker from "model/marker";
import RestModel from "model/rest";
import { DateTime } from "luxon";
import { config } from "app/session";
import { $gettext } from "common/vm";
import * as src from "common/src";
import Api from "common/api";

export let BatchSize = 24;

export class Face extends RestModel {
  constructor(values) {
    super(values);
  }

  getDefaults() {
    return {
      ID: "",
      Src: "",
      SubjUID: "",
      SubjSrc: "",
      FileUID: "",
      MarkerUID: "",
      Samples: 0,
      SampleRadius: 0.0,
      Collisions: 0,
      CollisionRadius: 0.0,
      Hidden: false,
      MatchedAt: "",
      CreatedAt: "",
      UpdatedAt: "",
      Name: "",
      FaceDist: 0.0,
      Size: 0,
      Score: 0,
      Review: false,
      Invalid: false,
      Thumb: "",
    };
  }

  route(view) {
    return { name: view, query: { q: "face:" + this.ID } };
  }

  classes(selected) {
    let classes = ["is-face", "uid-" + this.ID];

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

  setName(newName = this.Name) {
    if (!newName || newName.trim() === "") {
      // Can't save an empty name.
      return Promise.resolve(this);
    }

    this.SubjSrc = src.Manual;

    const payload = { SubjSrc: this.SubjSrc, Name: newName };

    return Api.put(Marker.getCollectionResource() + "/" + this.MarkerUID, payload).then((resp) => {
      if (resp && resp.data && resp.data.Name) {
        const data = resp.data;
        this.setValues({
          Name: data.Name,
          SubjSrc: data.SubjSrc,
          SubjUID: data.SubjUID,
          Review: data.Review,
          Invalid: data.Invalid,
          Thumb: data.Thumb,
        });
      }

      return Promise.resolve(this);
    });
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
    return "faces";
  }

  static getModelName() {
    return $gettext("Face");
  }
}

export default Face;
