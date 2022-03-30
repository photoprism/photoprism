/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import Event from "pubsub-js";
import themes from "options/themes.json";
import translations from "locales/translations.json";
import Api from "api.js";
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
        caption: "AI-Powered Photos App",
      };
      return;
    } else {
      this.baseUri = values.baseUri ? values.baseUri : "";
      this.staticUri = values.staticUri ? values.staticUri : this.baseUri + "/static";
      this.apiUri = values.apiUri ? values.apiUri : this.baseUri + "/api/v1";
      this.contentUri = values.contentUri ? values.contentUri : this.apiUri;
    }

    if (document && document.body) {
      document.body.classList.remove("nojs");

      // Set body class for browser optimizations.
      if (navigator.appVersion.indexOf("Chrome/") !== -1) {
        document.body.classList.add("chrome");
      } else if (navigator.appVersion.indexOf("Safari/") !== -1) {
        document.body.classList.add("safari");
      }
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

    if (values.jsUri && this.values.jsUri !== values.jsUri) {
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

  setColorMode(value) {
    if (!document || !document.body) {
      return;
    }

    const tags = document.getElementsByTagName("html");

    if (tags && tags.length > 0) {
      tags[0].setAttribute("data-color-mode", value);
    }

    if (value === "dark") {
      document.body.classList.add("dark-theme");
    } else {
      document.body.classList.remove("dark-theme");
    }
  }

  setTheme(name) {
    this.themeName = name;

    Event.publish("view.refresh", this);

    this.theme = themes[name] ? themes[name] : themes["default"];

    if (this.theme.dark) {
      this.setColorMode("dark");
    } else {
      this.setColorMode("light");
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

  searchBatchSize() {
    if (!this.values || !this.values.settings || !this.values.settings.search.batchSize) {
      return 80;
    }

    return this.values.settings.search.batchSize;
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

  isSponsor() {
    if (!this.values || !this.values.sponsor) {
      return false;
    }

    return !this.values.demo && !this.values.test;
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
