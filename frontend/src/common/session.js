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

import Api from "api.js";
import Event from "pubsub-js";
import User from "model/user";
import Socket from "websocket.js";

const RequestHeader = "X-Auth-Token";
const PublicSessionID = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";
const PublicAuthToken = "234200000000000000000000000000000000000000000000";
const LoginPage = "login";

export default class Session {
  /**
   * @param {Storage} storage
   * @param {Config} config
   * @param {object} shared
   */
  constructor(storage, config, shared) {
    this.storage_key = "sessionStorage";
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

    // Restore authentication from session storage.
    if (this.applyAuthToken(this.storage.getItem("authToken")) && this.applyId(this.storage.getItem("sessionId"))) {
      const dataJson = this.storage.getItem("sessionData");
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

        // Redirect URL.
        const location = shared.uri ? shared.uri : this.config.baseUri + "/";

        // Redirect to URL after one second.
        setTimeout(() => {
          window.location = location;
        }, 1000);
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

  setConfig(values) {
    this.config.setValues(values);
  }

  setAuthToken(authToken) {
    if (authToken) {
      this.storage.setItem("authToken", authToken);
      if (authToken === PublicAuthToken) {
        this.setId(PublicSessionID);
      }
    }

    return this.applyAuthToken(authToken);
  }

  getAuthToken() {
    return this.authToken;
  }

  hasAuthToken() {
    return !!this.authToken;
  }

  applyAuthToken(authToken) {
    if (!authToken) {
      this.reset();
      return false;
    }

    this.authToken = authToken;

    Api.defaults.headers.common[RequestHeader] = authToken;

    return true;
  }

  setId(id) {
    this.storage.setItem("sessionId", id);
    this.id = id;
  }

  getId() {
    return this.id;
  }

  hasId() {
    return !!this.id;
  }

  applyId(id) {
    if (!id) {
      return false;
    }

    this.setId(id);

    return true;
  }

  isAuthenticated() {
    return this.hasId() && this.hasAuthToken();
  }

  deleteAuthentication() {
    this.id = null;
    this.authToken = null;
    this.provider = "";

    // "sessionId" is the SHA256 hash of the auth token.
    this.storage.removeItem("sessionId");
    this.storage.removeItem("authToken");
    this.storage.removeItem("provider");

    // The "session_id" storage key is deprecated in favor of "authToken",
    // but should continue to be removed when logging out:
    this.storage.removeItem("session_id");

    delete Api.defaults.headers.common[RequestHeader];
  }

  setProvider(provider) {
    this.storage.setItem("provider", provider);
    this.provider = provider;
  }

  getProvider() {
    return this.provider;
  }

  hasProvider() {
    return !!this.provider;
  }

  setResp(resp) {
    if (!resp || !resp.data) {
      return;
    }

    if (resp.data.session_id) {
      this.setId(resp.data.session_id);
    }

    if (resp.data.access_token) {
      this.setAuthToken(resp.data.access_token);
    } else if (resp.data.id) {
      // TODO: "id" field is deprecated! Clients should now use "access_token" instead.
      // see https://github.com/photoprism/photoprism/commit/0d2f8be522dbf0a051ae6ef78abfc9efded0082d
      this.setAuthToken(resp.data.id);
    }

    if (resp.data.provider) {
      this.setProvider(resp.data.provider);
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
    this.storage.setItem("sessionData", JSON.stringify(data));

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
    this.storage.removeItem("sessionData");
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
    this.deleteAuthentication();
    this.deleteData();
    this.deleteUser();
    this.deleteClipboard();
  }

  sendClientInfo() {
    const hasConfig = !!window.__CONFIG__;
    const clientInfo = {
      session: this.getAuthToken(),
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

  login(username, password, code, token) {
    this.reset();

    return Api.post("session", { username, password, code, token }).then((resp) => {
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
    // Check if the authentication is still valid and update the client session data.
    if (this.config.isPublic()) {
      // Use a static auth token in public mode, as no additional authentication is required.
      this.setAuthToken(PublicAuthToken);
      this.setId(PublicSessionID);
      return Api.get("session").then((resp) => {
        this.setResp(resp);
        return Promise.resolve();
      });
    } else if (this.isAuthenticated()) {
      // Check the auth token by fetching the client session data from the API.
      return Api.get("session")
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
      // Skip updating session data if client is not authenticated.
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
    // Delete all authentication and session data.
    this.reset();

    // Perform redirect?
    if (noRedirect !== true && !this.isLogin()) {
      window.location = this.config.baseUri + "/";
    }

    return Promise.resolve();
  }

  logout(noRedirect) {
    if (this.isAuthenticated()) {
      return Api.delete("session")
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
