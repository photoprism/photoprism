/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.org>

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
import Notify from "common/notify";
import { $gettext } from "./vm";
import { config } from "../session";
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
