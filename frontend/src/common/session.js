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
import { createNamespacedStorage } from "common/storage";
import { $view } from "common/view";
import User from "model/user";
import Socket from "websocket.js";

const RequestHeader = "X-Auth-Token";
const PublicSessionID = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";
const PublicAuthToken = "234200000000000000000000000000000000000000000000";
const LoginPage = "login";

const resolveStorageNamespace = (config) => {
  if (typeof config?.storageNamespace === "string" && config.storageNamespace !== "") {
    return config.storageNamespace;
  }

  if (typeof config?.get === "function") {
    const storageNamespace = config.get("storageNamespace");

    if (typeof storageNamespace === "string" && storageNamespace !== "") {
      return storageNamespace;
    }
  }

  if (typeof config?.values?.storageNamespace === "string" && config.values.storageNamespace !== "") {
    return config.values.storageNamespace;
  }

  return "";
};

export default class Session {
  /**
   * @param {Storage} storage
   * @param {Config} config
   * @param {object} shared
   */
  constructor(storage, config, shared) {
    this.storageKey = "session";
    this.loginRedirect = false;
    this.config = config;
    this.baseUri = config?.baseUri;
    this.storageNamespace = resolveStorageNamespace(config);
    this.provider = "";
    this.user = new User(false);
    this.scope = "";
    this.data = null;
    this.localStorage = storage;
    const sessionStorage = typeof window === "undefined" ? undefined : window.sessionStorage;
    this.sessionStorage = createNamespacedStorage(sessionStorage, this.storageNamespace);

    // Set session storage.
    if (storage.getItem(this.storageKey) === "true") {
      this.storage = this.sessionStorage;
    } else {
      this.storage = this.localStorage;
    }

    // Restore authentication data stored under previously used keys.
    if (
      !this.storage.getItem(this.storageKey + ".token") &&
      this.storage.getItem("authToken") &&
      !this.storage.getItem(this.storageKey + ".id") &&
      this.storage.getItem("sessionId")
    ) {
      this.storage.setItem(this.storageKey + ".token", this.storage.getItem("authToken"));
      this.storage.removeItem("authToken");

      this.storage.setItem(this.storageKey + ".id", this.storage.getItem("sessionId"));
      this.storage.removeItem("sessionId");

      const dataJson = this.storage.getItem("sessionData");
      if (dataJson && dataJson !== "undefined") {
        this.storage.setItem(this.storageKey + ".data", dataJson);
        this.storage.removeItem("sessionData");
      }

      const userJson = this.storage.getItem("user");
      if (userJson && userJson !== "undefined") {
        this.storage.setItem(this.storageKey + ".user", userJson);
        this.storage.removeItem("user");
      }

      const provider = this.storage.getItem("provider");
      if (provider !== null && provider !== "undefined") {
        this.storage.setItem(this.storageKey + ".provider", provider);
        this.storage.removeItem("provider");
      }
    }

    // Restore authentication from session storage.
    if (this.applyAuthToken(this.storage.getItem(this.storageKey + ".token")) && this.applyId(this.storage.getItem(this.storageKey + ".id"))) {
      const data = this.getStoredJSON(this.storageKey + ".data");
      if (data) {
        this.data = data;
      }

      const user = this.getStoredJSON(this.storageKey + ".user");
      if (user) {
        this.user = new User(user);
      }

      const provider = this.storage.getItem(this.storageKey + ".provider");
      if (provider !== null && provider !== "undefined") {
        this.provider = provider;
      }

      const scope = this.storage.getItem(this.storageKey + ".scope");
      if (scope !== null && scope !== "undefined") {
        this.scope = scope;
      }
    }

    // Authenticated?
    this.auth = this.isUser();

    // Subscribe to session events.
    $event.subscribe("session.logout", () => {
      return this.onLogout();
    });

    $event.subscribe("websocket.connected", () => {
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
        this.followRedirect(location, 1000);
      });
    } else {
      this.config.progress(80);
      this.refresh().then(() => {
        this.config.progress(90);
        this.sendClientInfo();
      });
    }
  }

  getStoredJSON(key) {
    const value = this.storage.getItem(key);

    if (!value || value === "undefined") {
      return null;
    }

    try {
      return JSON.parse(value);
    } catch (_error) {
      this.storage.removeItem(key);
      return null;
    }
  }

  setStoragePreference(useSessionStorage) {
    this.localStorage.setItem(this.storageKey, useSessionStorage ? "true" : "false");
    this.sessionStorage.removeItem(this.storageKey);
  }

  useSessionStorage() {
    this.reset();
    this.setStoragePreference(true);
    this.storage = this.sessionStorage;
  }

  useLocalStorage() {
    this.setStoragePreference(false);
    this.storage = this.localStorage;
  }

  usesSessionStorage() {
    return this.storage === this.sessionStorage;
  }

  clearNamespacedKey(key) {
    [this.localStorage, this.sessionStorage].forEach((storage) => {
      if (!storage || typeof storage.removeItem !== "function") {
        return;
      }

      if (typeof storage.getLegacyItem === "function") {
        storage.removeItem(key, { legacy: false });
      } else {
        storage.removeItem(key);
      }
    });
  }

  clearLegacyKey(key) {
    if (!key) {
      return;
    }

    [this.localStorage, this.sessionStorage].forEach((storage) => {
      if (!storage) {
        return;
      }

      if (typeof storage.removeLegacyItem === "function") {
        storage.removeLegacyItem(key);
      } else if (typeof storage.removeItem === "function") {
        storage.removeItem(key);
      }
    });
  }

  setConfig(values) {
    this.config.setValues(values);
  }

  setAuthToken(authToken) {
    if (authToken) {
      this.storage.setItem(this.storageKey + ".token", authToken);
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

    $api.defaults.headers.common[RequestHeader] = authToken;

    return true;
  }

  setId(id) {
    this.storage.setItem(this.storageKey + ".id", id);
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
    this.scope = "";

    // "session.id" is the SHA256 hash of the auth token.
    this.clearNamespacedKey(this.storageKey + ".id");
    this.clearNamespacedKey(this.storageKey + ".token");
    this.clearNamespacedKey(this.storageKey + ".provider");
    this.clearNamespacedKey(this.storageKey + ".scope");

    // Remove deprecated raw keys from both storage backends. Supported
    // instances use namespaced keys, so clearing legacy globals is safe.
    this.clearLegacyKey("session.token");
    this.clearLegacyKey("authToken");
    this.clearLegacyKey("session.id");
    this.clearLegacyKey("sessionId");
    this.clearLegacyKey("session_id");
    this.clearLegacyKey("session.provider");
    this.clearLegacyKey("provider");
    this.clearLegacyKey("session.scope");
    this.clearLegacyKey("authError");
    this.clearLegacyKey("session.error");

    delete $api.defaults.headers.common[RequestHeader];
  }

  setProvider(provider) {
    this.storage.setItem(this.storageKey + ".provider", provider);
    this.provider = provider;
  }

  getProvider() {
    if (!this.provider) {
      return "";
    }

    return this.provider;
  }

  hasPassword() {
    switch (this.getProvider()) {
      case "local":
      case "ldap":
        return true;
      default:
        return false;
    }
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

    if (resp.data.scope) {
      this.setScope(resp.data.scope);
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
    this.storage.setItem(this.storageKey + ".data", JSON.stringify(data));

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
    this.storage.setItem(this.storageKey + ".user", JSON.stringify(user));
    this.auth = this.isUser();
  }

  getUser() {
    return this.user;
  }

  setScope(scope) {
    this.scope = scope;
    this.storage.setItem(this.storageKey + ".scope", scope);
  }

  hasScope() {
    return Boolean(this.scope) && this.scope !== "*";
  }

  getScope() {
    if (this.hasScope()) {
      return this.scope;
    }

    return "*";
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

  followLoginRedirectUrl(defaultUrl) {
    const url = this.getLoginRedirectUrl(defaultUrl);
    this.clearLoginRedirectUrl();
    this.followRedirect(url);
    return this;
  }

  getLoginRedirectUrl(defaultUrl) {
    if (!defaultUrl) {
      defaultUrl = "/";
    }

    return this.loginRedirect ? this.loginRedirect : defaultUrl;
  }

  clearLoginRedirectUrl() {
    this.loginRedirect = false;

    return this;
  }

  setLoginRedirectUrl(url) {
    if (!url) {
      return this.clearLoginRedirectUrl();
    }

    this.loginRedirect = url;

    return this;
  }

  isUser() {
    return this.user && this.user.hasId();
  }

  getDefaultRoute() {
    if (this.loginRequired()) {
      return LoginPage;
    }

    return this.config.getDefaultRoute();
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
    this.clearNamespacedKey(this.storageKey + ".data");
    this.clearLegacyKey("sessionData");
    this.clearLegacyKey("session.data");
  }

  deleteUser() {
    this.auth = false;
    this.user = new User(false);
    this.clearNamespacedKey(this.storageKey + ".user");
    this.clearLegacyKey("user");
    this.clearLegacyKey("session.user");
  }

  deleteClipboard() {
    this.clearNamespacedKey("clipboard");
    this.clearNamespacedKey("clipboard.photos");
    this.clearNamespacedKey("clipboard.albums");
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

    return $api.post("session", { username, password, code, token }).then((resp) => {
      const reload = this.config.getLanguageLocale() !== resp.data?.config?.settings?.ui?.language;
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
      return $api.get("session").then((resp) => {
        this.setResp(resp);
        return Promise.resolve();
      });
    } else if (this.isAuthenticated()) {
      // Check the auth token by fetching the client session data from the API.
      return $api
        .get("session")
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

    return $api.post("session", { token }).then((resp) => {
      this.setResp(resp);
      this.sendClientInfo();
    });
  }

  createApp(client_name, scope, expires_in, password) {
    if (!this.isUser() || !this.user.Name) {
      return Promise.reject();
    }

    if (!scope) {
      scope = "*";
    }

    return $api
      .post("oauth/token", {
        grant_type: password ? "password" : "session",
        client_name: client_name,
        scope: scope,
        expires_in: expires_in,
        username: this.user.Name,
        password: password,
      })
      .then((response) => Promise.resolve(response.data));
  }

  deleteApp(token) {
    return $api
      .post("oauth/revoke", {
        token: token,
      })
      .then((response) => Promise.resolve(response.data));
  }

  followRedirect(url, delay) {
    if (!url) {
      return;
    }

    // Default redirect delay in milliseconds.
    if (!delay) {
      delay = 100;
    }

    // Redirect to URL with the specified delay.
    $view.redirect(url, delay, true);
  }

  onLogout(noRedirect) {
    // Delete all authentication and session data.
    this.reset();

    // Perform redirect?
    if (noRedirect !== true && !this.isLogin()) {
      this.followRedirect(this.config.loginUri);
    }

    return Promise.resolve();
  }

  logout(noRedirect) {
    if (this.isAuthenticated()) {
      return $api
        .delete("session")
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
