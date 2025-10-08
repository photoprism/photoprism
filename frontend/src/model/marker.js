import RestModel from "model/rest";
import $api from "common/api";
import { DateTime } from "luxon";
import { $config } from "app/session";
import { $gettext } from "common/gettext";
import * as src from "common/src";

export let BatchSize = 60;

// Marker represents detected regions (faces or objects) within photos.
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
      return `${$config.contentUri}/t/${this.Thumb}/${$config.previewToken}/${size}`;
    } else {
      return `${$config.contentUri}/svg/portrait`;
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

  setName() {
    if (!this.Name || this.Name.trim() === "") {
      // Can't save an empty name.
      return Promise.resolve(this);
    }

    this.SubjSrc = src.Manual;

    const payload = { SubjSrc: this.SubjSrc, Name: this.Name };

    return $api.put(this.getEntityResource(), payload).then((resp) => Promise.resolve(this.setValues(resp.data)));
  }

  clearSubject() {
    return $api
      .delete(this.getEntityResource(this.getId()) + "/subject")
      .then((r) => Promise.resolve(this.setValues(r.data)));
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
