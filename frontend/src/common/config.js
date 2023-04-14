/*

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

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

import Api from "api.js";
import Event from "pubsub-js";
import * as themes from "options/themes";
import translations from "locales/translations.json";
import { Languages } from "options/options";
import { Photo } from "model/photo";
import { onInit, onSetTheme } from "common/hooks";

onInit();

export default class Config {
  /**
   * @param {Storage} storage
   * @param {object} values
   */
  constructor(storage, values) {
    this.disconnected = false;
    this.storage = storage;
    this.storage_key = "config";
    this.previewToken = "";
    this.downloadToken = "";
    this.updating = false;

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
      this.loginUri = "/library/login";
      this.apiUri = "/api/v1";
      this.contentUri = this.apiUri;
      this.videoUri = this.apiUri;
      this.values = {
        mode: "test",
        name: "Test",
      };
      this.page = {
        title: "PhotoPrism",
        caption: "AI-Powered Photos App",
      };
      return;
    } else {
      this.baseUri = values.baseUri ? values.baseUri : "";
      this.staticUri = values.staticUri ? values.staticUri : this.baseUri + "/static";
      this.loginUri = values.loginUri ? values.loginUri : this.baseUri + "/library/login";
      this.apiUri = values.apiUri ? values.apiUri : this.baseUri + "/api/v1";
      this.contentUri = values.contentUri ? values.contentUri : this.apiUri;
      this.videoUri = values.videoUri ? values.videoUri : this.apiUri;
    }

    if (document && document.body) {
      document.body.classList.remove("nojs");

      // Set body class for browser optimizations.
      if (navigator.userAgent.indexOf("Chrome/") !== -1) {
        document.body.classList.add("chrome");
      } else if (navigator.userAgent.indexOf("Safari/") !== -1) {
        document.body.classList.add("safari");
        document.body.classList.add("not-chrome");
      } else if (navigator.userAgent.indexOf("Firefox/") !== -1) {
        document.body.classList.add("firefox");
        document.body.classList.add("not-chrome");
      } else {
        document.body.classList.add("other-browser");
        document.body.classList.add("not-chrome");
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

    this.updateTokens();

    Event.subscribe("config.updated", (ev, data) => this.setValues(data.config));
    Event.subscribe("config.tokens", (ev, data) => this.setTokens(data));
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

    return Promise.resolve(this);
  }

  update() {
    if (this.updating !== false) {
      return this.updating;
    }

    this.updating = Api.get("config")
      .then(
        (resp) => {
          return this.setValues(resp.data);
        },
        () => console.warn("config update failed")
      )
      .finally(() => {
        this.updating = false;
        return Promise.resolve(this);
      });

    return this.updating;
  }

  setValues(values) {
    if (!values) return;

    if (this.debug) {
      console.log("config: updated", values);
    }

    if (values.jsUri && this.values.jsUri !== values.jsUri) {
      Event.publish("dialog.reload", { values });
    }

    for (let key in values) {
      if (values.hasOwnProperty(key) && values[key] != null) {
        this.set(key, values[key]);
      }
    }

    this.updateTokens();

    if (values.settings) {
      this.setBatchSize(values.settings);
      this.setLanguage(values.settings.ui.language);
      this.setTheme(values.settings.ui.theme);
    }

    // Adjust album counts by access level.
    if (values.count && this.deny("photos", "access_private")) {
      this.values.count.albums -= values.count.private_albums;
      this.values.count.folders -= values.count.private_folders;
      this.values.count.moments -= values.count.private_moments;
      this.values.count.months -= values.count.private_months;
      this.values.count.states -= values.count.private_states;
    }

    return this;
  }

  setBatchSize(settings) {
    if (!settings || !settings.search) {
      return;
    }

    if (settings.search.batchSize) {
      Photo.setBatchSize(settings.search.batchSize);
    }
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
        console.warn("more than one person having the same name", result);
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
        this.values.count.all -= data.count;
        this.values.count.photos -= data.count;
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

  setBodyTheme(name) {
    if (!document || !document.body) {
      return;
    }
    document.body.classList.forEach((c) => {
      if (c.startsWith("theme-")) {
        document.body.classList.remove(c);
      }
    });

    document.body.classList.add("theme-" + name);
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

  aclClasses(resource) {
    let result = [];
    const perms = ["update", "search", "manage", "share", "delete"];

    perms.forEach((perm) => {
      if (this.deny(resource, perm)) result.push(`disable-${perm}`);
    });

    return result;
  }

  // allow checks whether the current user is granted permission for the specified resource.
  allow(resource, perm) {
    if (this.values["acl"] && this.values["acl"][resource]) {
      if (this.values["acl"][resource]["full_access"]) {
        return true;
      } else if (this.values["acl"][resource][perm]) {
        return true;
      }
    }

    return false;
  }

  // allowAny checks whether the current user is granted any of the permissions for the specified resource.
  allowAny(resource, perms) {
    if (this.values["acl"] && this.values["acl"][resource]) {
      if (this.values["acl"][resource]["full_access"]) {
        return true;
      }
      for (const perm of perms) {
        if (this.values["acl"][resource][perm]) {
          return true;
        }
      }
    }

    return false;
  }

  // deny checks whether the current user must be denied access to the specified resource.
  deny(resource, perm) {
    return !this.allow(resource, perm);
  }

  // denyAll checks whether the current user is granted none of the permissions for the specified resource.
  denyAll(resource, perm) {
    return !this.allowAny(resource, perm);
  }

  settings() {
    return this.values.settings;
  }

  setSettings(settings) {
    if (!settings) return;

    if (this.debug) {
      console.log("config: new settings", settings);
    }

    this.values.settings = settings;

    this.setBatchSize(settings);
    this.setLanguage(settings.ui.language);
    this.setTheme(settings.ui.theme);

    return this;
  }

  setLanguage(locale) {
    if (!locale || this.loading()) {
      return;
    }

    if (this.values.settings && this.values.settings.ui) {
      this.values.settings.ui.language = locale;
      this.storage.setItem(this.storage_key + ".locale", locale);
      Api.defaults.headers.common["Accept-Language"] = locale;
    }

    return this;
  }

  getLanguage() {
    let locale = "en";

    if (this.loading()) {
      const stored = this.storage.getItem(this.storage_key + ".locale");
      if (stored) {
        locale = stored;
      }
    } else if (
      this.values.settings &&
      this.values.settings.ui &&
      this.values.settings.ui.language
    ) {
      locale = this.values.settings.ui.language;
    }

    return locale;
  }

  setTheme(name) {
    let theme = onSetTheme(name, this);

    if (!theme) {
      theme = themes.Get(name);
      this.themeName = theme.name;
    }

    if (this.values.settings && this.values.settings.ui) {
      this.values.settings.ui.theme = this.themeName;
    }

    Event.publish("view.refresh", this);

    this.theme = theme;

    this.setBodyTheme(this.themeName);

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

  restoreValues() {
    const json = this.storage.getItem(this.storage_key);
    if (json !== "undefined") {
      this.setValues(JSON.parse(json));
    }
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
    return this.values.settings.features[name] === true;
  }

  rtl() {
    if (!this.values || !this.values.settings || !this.values.settings.ui.language) {
      return false;
    }

    return Languages().some((lang) => lang.value === this.values.settings.ui.language && lang.rtl);
  }

  setTokens(tokens) {
    if (!tokens || !tokens.previewToken || !tokens.downloadToken) {
      return;
    }
    this.previewToken = tokens.previewToken;
    this.values.previewToken = tokens.previewToken;
    this.downloadToken = tokens.downloadToken;
    this.values.downloadToken = tokens.downloadToken;
  }

  updateTokens() {
    if (this.values["previewToken"]) {
      this.previewToken = this.values.previewToken;
    }
    if (this.values["downloadToken"]) {
      this.downloadToken = this.values.downloadToken;
    }
  }

  albumCategories() {
    if (this.values["albumCategories"]) {
      return this.values["albumCategories"];
    }

    return [];
  }

  isPublic() {
    return this.values && this.values.public;
  }

  isDemo() {
    return this.values && this.values.demo;
  }

  isSponsor() {
    if (!this.values || !this.values.sponsor) {
      return false;
    }

    return !this.values.demo && !this.values.test;
  }

  getName() {
    const s = this.get("name");

    if (!s) {
      return "PhotoPrism";
    }

    return s;
  }

  getAbout() {
    const s = this.get("about");

    if (!s) {
      return "PhotoPrism®";
    }

    return s;
  }

  getEdition() {
    const s = this.get("edition");

    if (!s) {
      return "ce";
    }

    return s;
  }

  ce() {
    return this.getEdition() === "ce";
  }

  getMembership() {
    const s = this.get("membership");

    if (!s) {
      return "ce";
    } else if (s === "ce" && this.isSponsor()) {
      return "essentials";
    }

    return s;
  }

  getCustomer() {
    const s = this.get("customer");

    if (!s) {
      return "";
    }

    return s;
  }

  getIcon() {
    switch (this.get("appIcon")) {
      case "crisp":
      case "mint":
      case "bold":
        return `${this.staticUri}/icons/${this.get("appIcon")}.svg`;
      default:
        return `${this.staticUri}/icons/logo.svg`;
    }
  }

  getVersion() {
    return this.get("version");
  }

  getSiteDescription() {
    return this.values.siteDescription ? this.values.siteDescription : this.values.siteCaption;
  }

  progress(p) {
    const el = document.getElementById("progress");
    if (el) {
      el.value = p;
    }
  }
}
