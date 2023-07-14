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
import User from "model/user";
import Socket from "websocket.js";

const SessionHeader = "X-Session-ID";
const PublicID = "234200000000000000000000000000000000000000000000";
const LoginPage = "login";

export default class Session {
  /**
   * @param {Storage} storage
   * @param {Config} config
   * @param {object} shared
   */
  constructor(storage, config, shared) {
    this.storage_key = "session_storage";
    this.auth = false;
    this.config = config;
    this.user = new User(false);
    this.data = null;

    // Set session storage.
    if (storage.getItem(this.storage_key) === "true") {
      this.storage = window.sessionStorage;
    } else {
      this.storage = storage;
    }

    // Restore from session storage.
    if (this.applyId(this.storage.getItem("session_id"))) {
      const dataJson = this.storage.getItem("data");
      if (dataJson !== "undefined") {
        this.data = JSON.parse(dataJson);
      }

      const userJson = this.storage.getItem("user");
      if (userJson !== "undefined") {
        this.user = new User(JSON.parse(userJson));
      }
    }

    // Authenticated?
    if (this.isUser()) {
      this.auth = true;
    }

    // Subscribe to session events.
    Event.subscribe("session.logout", () => {
      return this.onLogout();
    });

    Event.subscribe("websocket.connected", () => {
      this.sendClientInfo();
    });

    // Say hello.
    if (shared && shared.token) {
      this.config.progress(80);
      this.redeemToken(shared.token).finally(() => {
        this.config.progress(99);
        if (shared.uri) {
          window.location = shared.uri;
        } else {
          window.location = this.config.baseUri + "/";
        }
      });
    } else {
      this.config.progress(80);
      this.refresh().then(() => {
        this.config.progress(90);
        this.sendClientInfo();
      });
    }
  }

  useSessionStorage() {
    this.reset();
    this.storage.setItem(this.storage_key, "true");
    this.storage = window.sessionStorage;
  }

  useLocalStorage() {
    this.storage.setItem(this.storage_key, "false");
    this.storage = window.localStorage;
  }

  applyId(id) {
    if (!id) {
      this.reset();
      return false;
    }

    this.session_id = id;

    Api.defaults.headers.common[SessionHeader] = id;

    return true;
  }

  setId(id) {
    this.storage.setItem("session_id", id);
    return this.applyId(id);
  }

  setConfig(values) {
    this.config.setValues(values);
  }

  getId() {
    return this.session_id;
  }

  hasId() {
    return !!this.session_id;
  }

  deleteId() {
    this.session_id = null;
    this.provider = "";
    this.storage.removeItem("session_id");

    delete Api.defaults.headers.common[SessionHeader];
  }

  setResp(resp) {
    if (!resp || !resp.data) {
      return;
    }

    if (resp.data.id) {
      this.setId(resp.data.id);
    }
    if (resp.data.provider) {
      this.provider = resp.data.provider;
    }
    if (resp.data.config) {
      this.setConfig(resp.data.config);
    }
    if (resp.data.user) {
      this.setUser(resp.data.user);
    }
    if (resp.data.data) {
      this.setData(resp.data.data);
    }
  }

  setData(data) {
    if (!data) {
      return;
    }

    this.data = data;
    this.storage.setItem("data", JSON.stringify(data));

    if (data.user) {
      this.setUser(data.user);
    }
  }

  getEmail() {
    if (this.isUser()) {
      return this.user.Email;
    }

    return "";
  }

  getDisplayName() {
    if (this.isUser()) {
      return this.user.getEntityName();
    }

    return "";
  }

  setUser(user) {
    if (!user) {
      return;
    }

    this.user = new User(user);
    this.storage.setItem("user", JSON.stringify(user));
    this.auth = this.isUser();
  }

  getUser() {
    return this.user;
  }

  getUserUID() {
    if (this.user && this.user.UID) {
      return this.user.UID;
    } else {
      return "u000000000000001"; // Unknown.
    }
  }

  loginRequired() {
    return !this.config.isPublic() && !this.isUser();
  }

  isUser() {
    return this.user && this.user.hasId();
  }

  getHome() {
    if (this.loginRequired()) {
      return LoginPage;
    } else if (this.config.allow("photos", "access_library")) {
      return "browse";
    } else {
      return "albums";
    }
  }

  isAdmin() {
    return this.user && this.user.hasId() && (this.user.Role === "admin" || this.user.SuperAdmin);
  }

  isSuperAdmin() {
    return this.user && this.user.hasId() && this.user.SuperAdmin;
  }

  isAnonymous() {
    return !this.user || !this.user.hasId();
  }

  hasToken(token) {
    if (!this.data || !this.data.tokens) {
      return false;
    }

    return this.data.tokens.indexOf(token) >= 0;
  }

  deleteData() {
    this.data = null;
    this.storage.removeItem("data");
  }

  deleteUser() {
    this.auth = false;
    this.user = new User(false);
    this.storage.removeItem("user");
  }

  deleteClipboard() {
    this.storage.removeItem("clipboard");
    this.storage.removeItem("photo_clipboard");
    this.storage.removeItem("album_clipboard");
  }

  reset() {
    this.deleteId();
    this.deleteData();
    this.deleteUser();
    this.deleteClipboard();
  }

  sendClientInfo() {
    const hasConfig = !!window.__CONFIG__;
    const clientInfo = {
      session: this.getId(),
      cssUri: hasConfig ? window.__CONFIG__.cssUri : "",
      jsUri: hasConfig ? window.__CONFIG__.jsUri : "",
      version: hasConfig ? window.__CONFIG__.version : "",
    };

    try {
      Socket.send(JSON.stringify(clientInfo));
    } catch (e) {
      if (this.config.debug) {
        console.log("session: can't use websocket, not connected (yet)");
      }
    }
  }

  isLogin() {
    if (!window || !window.location) {
      return true;
    }
    return LoginPage === window.location.href.substring(window.location.href.lastIndexOf("/") + 1);
  }

  login(username, password, token) {
    this.reset();

    return Api.post("session", { username, password, token }).then((resp) => {
      const reload = this.config.getLanguage() !== resp.data?.config?.settings?.ui?.language;
      this.setResp(resp);
      this.onLogin();
      return Promise.resolve(reload);
    });
  }

  onLogin() {
    this.sendClientInfo();
  }

  refresh() {
    // Refresh session information.
    if (this.config.isPublic()) {
      // No authentication in public mode.
      this.setId(PublicID);
      return Api.get("session/" + this.getId()).then((resp) => {
        this.setResp(resp);
        return Promise.resolve();
      });
    } else if (this.hasId()) {
      // Verify authentication.
      return Api.get("session/" + this.getId())
        .then((resp) => {
          this.setResp(resp);
          return Promise.resolve();
        })
        .catch(() => {
          this.reset();
          if (!this.isLogin()) {
            window.location.reload();
          }
          return Promise.reject();
        });
    } else {
      // No authentication yet.
      return Promise.resolve();
    }
  }

  redeemToken(token) {
    if (!token) {
      return Promise.reject();
    }

    return Api.post("session", { token }).then((resp) => {
      this.setResp(resp);
      this.sendClientInfo();
    });
  }

  onLogout(noRedirect) {
    this.reset();
    if (noRedirect !== true && !this.isLogin()) {
      window.location = this.config.baseUri + "/";
    }
    return Promise.resolve();
  }

  logout(noRedirect) {
    if (this.hasId()) {
      return Api.delete("session/" + this.getId())
        .then(() => {
          return this.onLogout(noRedirect);
        })
        .catch(() => {
          return this.onLogout(noRedirect);
        });
    } else {
      return this.onLogout(noRedirect);
    }
  }
}
