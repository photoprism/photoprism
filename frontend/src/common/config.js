/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

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

import $api from "common/api";
import $event from "common/event";
import * as themes from "options/themes";
import * as options from "options/options";
import { Photo } from "model/photo";
import { onInit, onSetTheme } from "common/hooks";
import { ref, reactive } from "vue";

onInit();

export default class Config {
  /**
   * @param {Storage} storage
   * @param {object} values
   */
  constructor(storage, values) {
    this.disconnected = ref(false);
    this.storage = storage;
    this.storageKey = "config";
    this.previewToken = "";
    this.downloadToken = "";
    this.updating = false;

    this.$vuetify = null;
    this.translations = {};

    if (!values || !values.siteTitle) {
      // Omit warning in unit tests.
      if (navigator && navigator.userAgent && !navigator.userAgent.includes("HeadlessChrome")) {
        console.warn("config: values missing");
      }

      const now = new Date();

      this.debug = true;
      this.trace = false;
      this.test = true;
      this.demo = false;
      this.version = `${now.getUTCFullYear().toString().substr(-2)}${now.getMonth() + 1}${now.getDate()}-TEST`;
      this.theme = themes.Get("default");
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

    this.page = reactive({
      title: values.siteTitle,
      caption: values.siteCaption,
    });

    this.values = reactive(values);
    this.debug = !!values.debug;
    this.trace = !!values.trace;
    this.test = !!values.test;
    this.demo = !!values.demo;
    this.version = values.version;

    this.updateTokens();

    $event.subscribe("config.updated", (ev, data) => this.setValues(data.config));
    $event.subscribe("config.tokens", (ev, data) => this.setTokens(data));
    $event.subscribe("count", (ev, data) => this.onCount(ev, data));
    $event.subscribe("people", (ev, data) => this.onPeople(ev, data));

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

    this.updating = $api
      .get("config")
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

  async setValues(values) {
    if (!values || typeof values !== "object") {
      return;
    }

    if (values.jsUri && values.mode === "user" && this.values.jsUri && this.values.jsUri !== values.jsUri) {
      $event.publish("dialog.update", { values });
    }

    if (values.DefaultLocale && options.DefaultLocale !== values.DefaultLocale) {
      options.SetDefaultLocale(values.DefaultLocale);
    }

    for (let key in values) {
      if (values.hasOwnProperty(key) && values[key] != null) {
        this.set(key, values[key]);
      }
    }

    this.updateTokens();

    if (values.settings) {
      this.setBatchSize(values.settings);
      await this.setLanguage(this.getLanguageLocale(), true);
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

    if (settings?.search?.batchSize > 0) {
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

  // getPerson returns the details of a person by name
  // (case-insensitive), or null if it does not exist.
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

  // onCount updates the media type, location, people and other
  // counters used e.g. in the expanded sidebar navigation.
  onCount(ev, data) {
    const type = ev.split(".")[1];

    switch (type) {
      case "photos":
        this.values.count.all += data.count;
        this.values.count.photos += data.count;
        break;
      case "animated":
        this.values.count.all += data.count;
        this.values.count.media += data.count;
        this.values.count.animated += data.count;
        break;
      case "videos":
        this.values.count.all += data.count;
        this.values.count.media += data.count;
        this.values.count.videos += data.count;
        break;
      case "live":
        this.values.count.all += data.count;
        this.values.count.media += data.count;
        this.values.count.live += data.count;
        break;
      case "audio":
        this.values.count.all += data.count;
        this.values.count.media += data.count;
        this.values.count.audio += data.count;
        break;
      case "documents":
        this.values.count.all += data.count;
        this.values.count.documents += data.count;
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
      case "hidden":
        this.values.count.hidden += data.count;
        break;
      case "archived":
        this.values.count.archived += data.count;
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

  // setVuetify sets a reference to the current Vuetify instance.
  setVuetify(instance) {
    this.$vuetify = instance;
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

  // loadTranslation asynchronously loads the specified locale file.
  async loadTranslation(locale) {
    if (!locale || (this.translations && this.translations[locale])) {
      return;
    }

    try {
      // Dynamically import the translation JSON file.
      await import(
        /* webpackChunkName: "[request]" */
        /* webpackMode: "lazy" */
        `../locales/json/${locale}.json`
      ).then((module) => {
        Object.assign(this.translations, module.default);
      });
    } catch (error) {
      console.error(`failed to load translations for locale ${locale}:`, error);
    }
  }

  // setLanguage sets the ISO/IEC 15897 locale,
  // e.g. "en" or "zh_TW" (minimum 2 letters).
  async setLanguage(locale, apply) {
    // Skip setting language if no locale is specified.
    if (!locale) {
      return this;
    }

    // Apply locale to browser window?
    if (apply) {
      await this.loadTranslation(locale);

      // Update the Accept-Language header for XHR requests.
      if ($api) {
        $api.defaults.headers.common["Accept-Language"] = locale;
      }

      // Update the language-specific attributes of the <html> and <body> elements.
      if (document && document.body) {
        const isRtl = this.isRtl(locale);

        // Update <html> lang attribute and dir attribute to match the current locale.
        document.documentElement.setAttribute("lang", locale);
        document.documentElement.setAttribute("dir", this.dir(isRtl));

        // Set body.is-rtl or .is-ltr, depending on the writing direction of the current locale.
        if (isRtl) {
          document.body.classList.add("is-rtl");
          document.body.classList.remove("is-ltr");
        } else {
          document.body.classList.remove("is-rtl");
          document.body.classList.add("is-ltr");
        }
      }
    }

    // Don't update the configuration settings if they haven't been loaded yet.
    if (this.loading()) {
      return this;
    }

    // Update the configuration settings and save them to window.localStorage.
    if (this.values.settings && this.values.settings.ui) {
      this.values.settings.ui.language = locale;
      this.storage.setItem(this.storageKey + ".locale", locale);
    }

    return this;
  }

  // getLanguageLocale returns the ISO/IEC 15897 locale,
  // e.g. "en" or "zh_TW" (minimum 2 letters).
  getLanguageLocale() {
    // Get default locale from web browser.
    let locale = navigator?.language;

    // Override language locale with query parameter?
    if (window.location?.search) {
      const query = new URLSearchParams(window.location.search);
      const queryLocale = query.get("locale");
      if (queryLocale && queryLocale.length > 1 && queryLocale.length < 6) {
        // Change the locale settings.
        locale = options.FindLocale(queryLocale);
        this.storage.setItem(this.storageKey + ".locale", locale);
        if (this.values?.settings?.ui) {
          this.values.settings.ui.language = locale;
        }
      }
    }

    // Get user locale from localStorage if settings have not yet been loaded from backend.
    if (this.loading()) {
      const stored = this.storage.getItem(this.storageKey + ".locale");
      if (stored) {
        locale = stored;

        if (this.values?.settings?.ui) {
          this.values.settings.ui.language = locale;
        }
      }
    } else if (this.values?.settings?.ui?.language) {
      locale = this.values.settings.ui.language;
    }

    // Find and return the best matching language locale that exists.
    return options.FindLocale(locale);
  }

  // getLanguageCode returns the ISO 639-1 language code (2 letters),
  // see https://www.loc.gov/standards/iso639-2/php/code_list.php.
  getLanguageCode() {
    return this.getLanguageLocale().substring(0, 2);
  }

  // getTimeZone returns user time zone.
  getTimeZone() {
    if (this.values?.settings?.ui?.timeZone) {
      return this.values?.settings?.ui?.timeZone;
    }

    return "Local";
  }

  // isRtl returns true if a right-to-left language is currently used.
  isRtl(locale) {
    if (!locale) {
      locale = this.getLanguageLocale();
    }

    return options.Languages().some((l) => l.value === locale && l.rtl);
  }

  // dir returns the user interface direction (for the current locale if no argument is given).
  dir(isRtl) {
    if (typeof isRtl === "undefined") {
      isRtl = this.isRtl();
    }

    return isRtl ? "rtl" : "ltr";
  }

  // setTheme set the current UI theme based on the specified name.
  setTheme(name) {
    let theme = onSetTheme(name, this);

    if (!theme) {
      theme = themes.Get(name);
      this.themeName = theme?.name;
    }

    if (this.values.settings && this.values.settings.ui) {
      this.values.settings.ui.theme = this.themeName;
    }

    $event.publish("view.refresh", this);

    this.theme = theme;

    this.setBodyTheme(this.themeName);

    if (this.theme.dark) {
      this.setColorMode("dark");
    } else {
      this.setColorMode("light");
    }

    if (this.themeName && this.$vuetify) {
      this.$vuetify.theme.name = this.themeName;
    }

    return this;
  }

  // setBodyTheme updates the classes of the <body> element based on the specified theme name.
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

  // setColorMode updates the dark/light mode attributes of the <html> and <body> elements.
  setColorMode(value) {
    if (!document || !document.body) {
      return;
    }

    if (document.documentElement) {
      document.documentElement.setAttribute("data-color-mode", value);
    }

    if (value === "dark") {
      document.body.classList.add("dark-theme");
    } else {
      document.body.classList.remove("dark-theme");
    }
  }

  // getSettings returns the current user's configuration settings.
  getSettings() {
    return this.values.settings;
  }

  // setSettings updates the current user's configuration settings
  // and then changes the UI language and theme as needed.
  setSettings(settings) {
    if (!settings) {
      return;
    }

    if (this.debug) {
      console.log("config: new settings", settings);
    }

    this.values.settings = settings;

    this.setBatchSize(settings);
    this.setLanguage(settings.ui.language, false);
    this.setTheme(settings.ui.theme);

    return this;
  }

  // getDefaultRoute returns the default route to use after login or in case of routing errors.
  getDefaultRoute() {
    if (this.isPortal()) {
      return "cluster";
    }

    const albumsRoute = "albums";
    const browseRoute = "browse";
    const defaultRoute = this.deny("photos", "access_library") ? albumsRoute : browseRoute;

    if (this.allow("settings", "update")) {
      const features = this.getSettings()?.features;
      const startPage = this.getSettings()?.ui?.startPage;

      if (features && typeof features === "object" && startPage && typeof startPage === "string") {
        switch (startPage) {
          case "browse":
            return defaultRoute;
          case "albums":
            return features.albums ? startPage : defaultRoute;
          case "photos":
            return features.albums ? startPage : defaultRoute;
          case "videos":
            return features.library ? startPage : defaultRoute;
          case "people":
            return features.people && features.edit ? startPage : defaultRoute;
          case "favorites":
            return features.favorites ? startPage : defaultRoute;
          case "places":
            return features.places ? startPage : defaultRoute;
          case "calendar":
            return features.calendar ? startPage : defaultRoute;
          case "moments":
            return features.moments ? startPage : defaultRoute;
          case "labels":
            return features.labels ? startPage : defaultRoute;
          case "folders":
            return features.folders ? startPage : defaultRoute;
          default:
            return defaultRoute;
        }
      }
    }

    return defaultRoute;
  }

  // getValues returns all client configuration values as exposed by the backend.
  getValues() {
    return this.values;
  }

  // storeValues saves the current configuration values in window.localStorage.
  storeValues() {
    this.storage.setItem(this.storageKey, JSON.stringify(this.getValues()));
    return this;
  }

  // restoreValues restores the configuration values from window.localStorage.
  restoreValues() {
    const json = this.storage.getItem(this.storageKey);
    if (json !== "undefined") {
      this.setValues(JSON.parse(json));
    }
    return this;
  }

  // set updates a top-level config value.
  set(key, value) {
    this.values[key] = value;
    return this;
  }

  // has checks if the specified top-level config value exists.
  has(key) {
    return !!this.values[key];
  }

  // get returns a top-level config value.
  get(key) {
    return this.values[key];
  }

  // featDevelop checks if features under development should be enabled.
  featDevelop() {
    return this.values?.develop === true;
  }

  // featExperimental checks if new features that may be incomplete or unstable should be enabled.
  featExperimental() {
    return this.values?.experimental === true;
  }

  // featPreview checks if features available for preview should be enabled.
  featPreview() {
    return this.featDevelop() || this.featExperimental();
  }

  // feature checks a single feature flag by name and returns true if it is set.
  feature(name) {
    return this.values.settings.features[name] === true;
  }

  // filesQuotaReached returns true if the filesystem quota is reached or exceeded.
  filesQuotaReached() {
    return this.values?.usage?.filesUsedPct >= 100;
  }

  // setTokens sets the security tokens required to load thumbnails and download files from the server.
  setTokens(tokens) {
    if (!tokens || typeof tokens !== "object") {
      return;
    }

    if (tokens.previewToken && this.values?.previewToken !== tokens.previewToken) {
      this.values.previewToken = tokens.previewToken;
    }

    if (tokens.downloadToken && this.values?.downloadToken !== tokens.downloadToken) {
      this.values.downloadToken = tokens.downloadToken;
    }

    this.updateTokens();
  }

  // updateTokens updates the security tokens required to load thumbnails and download files from the server.
  updateTokens() {
    if (this.values?.previewToken && this.previewToken !== this.values.previewToken) {
      this.previewToken = this.values.previewToken;
    }

    if (this.values?.downloadToken && this.downloadToken !== this.values.downloadToken) {
      this.downloadToken = this.values.downloadToken;
    }
  }

  // albumCategories returns an array containing the categories
  // assigned to albums, or an empty array if there are none.
  albumCategories() {
    if (this.values["albumCategories"]) {
      return this.values["albumCategories"];
    }

    return [];
  }

  // isPublic returns true if the instance is running in public mode, i.e. without authentication.
  isPublic() {
    return this.values && this.values.public;
  }

  // isDemo returns true if the instance is running in demo mode for public or private testing.
  isDemo() {
    return this.values && this.values.demo;
  }

  // isPortal returns true if this is a cluster portal server.
  isPortal() {
    return this.values && this.values.portal;
  }

  // isPro returns true if this is team version.
  isPro() {
    return !!this.values?.ext["pro"];
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

  getTier() {
    const tier = this.get("tier");

    if (!tier) {
      return 0;
    }

    return tier;
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
    if (this.theme?.variables?.icon) {
      return this.theme.variables.icon;
    }

    switch (this.get("appIcon")) {
      case "crisp":
      case "mint":
      case "bold":
        return `${this.staticUri}/icons/${this.get("appIcon")}.svg`;
      default:
        return `${this.staticUri}/icons/logo.svg`;
    }
  }

  getLoginIcon() {
    const loginTheme = themes.Get("login", false);
    if (loginTheme?.variables?.icon) {
      return loginTheme?.variables?.icon;
    }

    return this.getIcon();
  }

  getVersion() {
    return this.version;
  }

  getServerVersion() {
    return this.values?.version;
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
