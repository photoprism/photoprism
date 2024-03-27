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
import Notify from "common/notify";
import { $gettext } from "vm.js";
import { config } from "app/session";
export const MaxItems = 999;

export class Clipboard {
  /**
   * @param {Storage} storage
   * @param {string} key
   */
  constructor(storage, key) {
    this.storageKey = key ? key : "clipboard";

    this.storage = storage;
    this.selectionMap = {};
    this.selection = [];
    this.lastId = "";
    this.maxItems = MaxItems;

    this.loadFromStorage();
  }

  isModel(model) {
    if (!model) {
      if (config.debug) {
        console.warn("Clipboard::isModel() - empty model", model);
      }
      return false;
    }

    if (typeof model.getId !== "function") {
      if (config.debug) {
        console.warn("Clipboard::isModel() - model.getId() is not a function", model);
      }
      return false;
    }

    return true;
  }

  loadFromStorage() {
    const photosJson = this.storage.getItem(this.storageKey);

    if (photosJson !== null && typeof photosJson !== "undefined") {
      this.setIds(JSON.parse(photosJson));
    }
  }

  saveToStorage() {
    this.storage.setItem(this.storageKey, JSON.stringify(this.selection));
  }

  toggle(model) {
    if (!this.isModel(model)) {
      return;
    }

    const id = model.getId();

    const result = this.toggleId(id);

    this.updateDom(id, result);

    return result;
  }

  toggleId(id) {
    const index = this.selection.indexOf(id);

    let result = false;

    if (index === -1) {
      if (this.selection.length >= this.maxItems) {
        Notify.warn($gettext("Can't select more items"));
        return;
      }

      this.selection.push(id);
      this.selectionMap["id:" + id] = true;
      this.lastId = id;
      result = true;
    } else {
      this.selection.splice(index, 1);
      delete this.selectionMap["id:" + id];
      this.lastId = "";
    }

    this.saveToStorage();

    return result;
  }

  add(model) {
    if (!this.isModel(model)) {
      return;
    }

    const id = model.getId();

    this.addId(id);
  }

  addId(id) {
    this.updateDom(id, true);

    if (this.hasId(id)) {
      return true;
    }

    if (this.selection.length >= this.maxItems) {
      Notify.warn($gettext("Can't select more items"));
      return false;
    }

    this.selection.push(id);
    this.selectionMap["id:" + id] = true;
    this.lastId = id;

    this.saveToStorage();

    return true;
  }

  addRange(rangeEnd, models) {
    if (!models || !models[rangeEnd] || !(models[rangeEnd] instanceof RestModel)) {
      if (config.debug) {
        console.warn("Clipboard::addRange() - invalid arguments:", rangeEnd, models);
      }
      return;
    }

    let rangeStart = models.findIndex((photo) => photo.UID === this.lastId);

    if (rangeStart === -1) {
      this.toggle(models[rangeEnd]);
      return 1;
    }

    if (rangeStart > rangeEnd) {
      const newEnd = rangeStart;
      rangeStart = rangeEnd;
      rangeEnd = newEnd;
    }

    for (let i = rangeStart; i <= rangeEnd; i++) {
      this.add(models[i], false);
      this.updateDom(models[i].getId(), true);
    }

    return rangeEnd - rangeStart + 1;
  }

  has(model) {
    if (!this.isModel(model)) {
      return;
    }

    return this.hasId(model.getId());
  }

  hasId(id) {
    return typeof this.selectionMap["id:" + id] !== "undefined";
  }

  remove(model, publish) {
    if (!this.isModel(model)) {
      return;
    }

    this.removeId(model.getId(), publish);
  }

  removeId(id) {
    this.updateDom(id, false);

    if (!this.hasId(id)) {
      return false;
    }

    const index = this.selection.indexOf(id);

    this.selection.splice(index, 1);
    this.lastId = "";
    delete this.selectionMap["id:" + id];

    this.saveToStorage();

    return false;
  }

  getIds() {
    return this.selection;
  }

  setIds(ids) {
    if (!Array.isArray(ids)) return;

    this.selection = ids;
    this.selectionMap = {};
    this.lastId = "";

    for (let i = 0; i < this.selection.length; i++) {
      this.selectionMap["id:" + this.selection[i]] = true;
    }
  }

  clear() {
    this.selection.forEach((id) => this.updateDom(id, false));
    this.lastId = "";
    this.selectionMap = {};
    this.selection.splice(0, this.selection.length);
    this.storage.removeItem(this.storageKey);
  }

  updateDom(uid, selected) {
    if (typeof uid === "object") {
      if (typeof uid.getId !== "function") {
        return;
      }

      uid = uid.getId();
    }

    document.querySelectorAll(`.uid-${uid}`).forEach((el) => {
      if (selected) {
        el.classList.add("is-selected");
      } else {
        el.classList.remove("is-selected");
      }
    });
  }
}

const PhotoClipboard = new Clipboard(window.localStorage, "photo_clipboard");

export default PhotoClipboard;
