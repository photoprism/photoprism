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

import Event from "pubsub-js";
import themes from "options/themes.json";
import translations from "locales/translations.json";
import Api from "./api";
import { Languages } from "options/options";

export default class Config {
  /**
   * @param {Storage} storage
   * @param {object} values
   */
  constructor(storage, values) {
    this.disconnected = false;
    this.storage = storage;
    this.storage_key = "config";

    this.$vuetify = null;
    this.translations = translations;

    if (!values || !values.siteTitle) {
      // Omit warning in unit tests.
      if (navigator && navigator.userAgent && !navigator.userAgent.includes("HeadlessChrome")) {
        console.warn("config: values missing");
      }

      this.debug = true;
      this.test = true;
      this.demo = false;
      this.themeName = "";
      this.baseUri = "";
      this.staticUri = "/static";
      this.apiUri = "/api/v1";
      this.contentUri = this.apiUri;
      this.values = {};
      this.page = {
        title: "PhotoPrism",
        caption: "Browse Your Life",
      };
      return;
    } else {
      this.baseUri = values.baseUri ? values.baseUri : "";
      this.staticUri = values.staticUri ? values.staticUri : this.baseUri + "/static";
      this.apiUri = values.apiUri ? values.apiUri : this.baseUri + "/api/v1";
      this.contentUri = values.contentUri ? values.contentUri : this.apiUri;
    }

    this.page = {
      title: values.siteTitle,
      caption: values.siteCaption,
    };

    this.values = values;
    this.debug = !!values.debug;
    this.test = !!values.test;
    this.demo = !!values.demo;

    Event.subscribe("config.updated", (ev, data) => this.setValues(data.config));
    Event.subscribe("count", (ev, data) => this.onCount(ev, data));
    Event.subscribe("people", (ev, data) => this.onPeople(ev, data));

    if (this.has("settings")) {
      this.setTheme(this.get("settings").ui.theme);
    } else {
      this.setTheme("default");
    }
  }

  loading() {
    return !this.values.mode || this.values.mode === "public";
  }

  load() {
    if (this.loading()) {
      return this.update();
    }

    return Promise.resolve();
  }

  update() {
    return Api.get("config")
      .then(
        (response) => this.setValues(response.data),
        () => console.warn("failed pulling updated client config")
      )
      .finally(() => Promise.resolve());
  }

  setValues(values) {
    if (!values) return;

    if (this.debug) {
      console.log("config: new values", values);
    }

    if (values.jsHash && this.values.jsHash !== values.jsHash) {
      Event.publish("dialog.reload", { values });
    }

    for (let key in values) {
      if (values.hasOwnProperty(key)) {
        this.set(key, values[key]);
      }
    }

    if (values.settings) {
      this.setTheme(values.settings.ui.theme);
    }

    return this;
  }

  onPeople(ev, data) {
    const type = ev.split(".")[1];

    if (this.debug) {
      console.log(ev, data);
    }

    if (!this.values.people) {
      this.values.people = [];
    }

    if (!data || !data.entities || !Array.isArray(data.entities)) {
      return;
    }

    switch (type) {
      case "created":
        this.values.people.unshift(...data.entities);
        break;
      case "updated":
        for (let i = 0; i < data.entities.length; i++) {
          const values = data.entities[i];

          this.values.people
            .filter((m) => m.UID === values.UID)
            .forEach((m) => {
              for (let key in values) {
                if (
                  key !== "UID" &&
                  values.hasOwnProperty(key) &&
                  values[key] != null &&
                  typeof values[key] !== "object"
                ) {
                  m[key] = values[key];
                }
              }
            });
        }
        break;
      case "deleted":
        for (let i = 0; i < data.entities.length; i++) {
          const index = this.values.people.findIndex((m) => m.UID === data.entities[i]);

          if (index >= 0) {
            this.values.people.splice(index, 1);
          }
        }
        break;
    }
  }

  getPerson(name) {
    name = name.toLowerCase();

    const result = this.values.people.filter((m) => m.Name.toLowerCase() === name);
    const l = result ? result.length : 0;

    if (l === 0) {
      return null;
    } else if (l === 1) {
      return result[0];
    } else {
      if (this.debug) {
        console.warn("more than one person matching the same name", result);
      }
      return result[0];
    }
  }

  onCount(ev, data) {
    const type = ev.split(".")[1];

    switch (type) {
      case "photos":
        this.values.count.all += data.count;
        this.values.count.photos += data.count;
        break;
      case "live":
        this.values.count.all += data.count;
        this.values.count.live += data.count;
        break;
      case "videos":
        this.values.count.all += data.count;
        this.values.count.videos += data.count;
        break;
      case "cameras":
        this.values.count.cameras += data.count;
        this.update();
        break;
      case "lenses":
        this.values.count.lenses += data.count;
        break;
      case "countries":
        this.values.count.countries += data.count;
        this.update();
        break;
      case "states":
        this.values.count.states += data.count;
        break;
      case "people":
        this.values.count.people += data.count;
        break;
      case "places":
        this.values.count.places += data.count;
        break;
      case "labels":
        this.values.count.labels += data.count;
        break;
      case "albums":
        this.values.count.albums += data.count;
        break;
      case "moments":
        this.values.count.moments += data.count;
        break;
      case "months":
        this.values.count.months += data.count;
        break;
      case "folders":
        this.values.count.folders += data.count;
        break;
      case "files":
        this.values.count.files += data.count;
        break;
      case "favorites":
        this.values.count.favorites += data.count;
        break;
      case "review":
        this.values.count.review += data.count;
        break;
      case "private":
        this.values.count.private += data.count;
        break;
      default:
        console.warn("unknown count type", ev, data);
    }

    this.values.count;
  }

  setVuetify(instance) {
    this.$vuetify = instance;
  }

  setTheme(name) {
    this.themeName = name;

    const el = document.getElementById("photoprism");

    if (el) {
      el.className = "theme-" + name;
    }

    this.theme = themes[name] ? themes[name] : themes["default"];

    if (this.theme.dark) {
      document.body.classList.add("dark-theme");
    } else {
      document.body.classList.remove("dark-theme");
    }

    if (this.$vuetify) {
      this.$vuetify.theme = this.theme.colors;
    }

    return this;
  }

  getValues() {
    return this.values;
  }

  storeValues() {
    this.storage.setItem(this.storage_key, JSON.stringify(this.getValues()));
    return this;
  }

  set(key, value) {
    this.values[key] = value;
    return this;
  }

  has(key) {
    return !!this.values[key];
  }

  get(key) {
    return this.values[key];
  }

  feature(name) {
    return this.values.settings.features[name];
  }

  settings() {
    return this.values.settings;
  }

  rtl() {
    if (!this.values || !this.values.settings || !this.values.settings.ui.language) {
      return false;
    }

    return Languages().some((lang) => lang.value === this.values.settings.ui.language && lang.rtl);
  }

  downloadToken() {
    return this.values["downloadToken"];
  }

  previewToken() {
    return this.values["previewToken"];
  }

  albumCategories() {
    if (this.values["albumCategories"]) {
      return this.values["albumCategories"];
    }

    return [];
  }

  appIcon() {
    switch (this.get("appIcon")) {
      case "crisp":
      case "mint":
      case "bold":
        return `${this.staticUri}/icons/${this.get("appIcon")}.svg`;
      default:
        return `${this.staticUri}/icons/logo.svg`;
    }
  }
}
