/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

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

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

import Api from "./api";
import Event from "pubsub-js";
import User from "model/user";
import Socket from "./websocket";

export default class Session {
  /**
   * @param {Storage} storage
   * @param {Config} config
   */
  constructor(storage, config) {
    this.auth = false;
    this.config = config;

    if (storage.getItem("session_storage") === "true") {
      this.storage = window.sessionStorage;
    } else {
      this.storage = storage;
    }

    if (this.applyId(this.storage.getItem("session_id"))) {
      const dataJson = this.storage.getItem("data");
      this.data = dataJson !== "undefined" ? JSON.parse(dataJson) : null;
    }

    if (this.data && this.data.user) {
      this.user = new User(this.data.user);
    }

    if (this.isUser()) {
      this.auth = true;
    }

    Event.subscribe("session.logout", () => {
      return this.onLogout();
    });

    Event.subscribe("websocket.connected", () => {
      this.sendClientInfo();
    });

    this.sendClientInfo();
  }

  useSessionStorage() {
    this.deleteId();
    this.storage.setItem("session_storage", "true");
    this.storage = window.sessionStorage;
  }

  useLocalStorage() {
    this.storage.setItem("session_storage", "false");
    this.storage = window.localStorage;
  }

  applyId(id) {
    if (!id) {
      this.deleteId();
      return false;
    }

    this.session_id = id;
    Api.defaults.headers.common["X-Session-ID"] = id;

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
    this.storage.removeItem("session_id");
    delete Api.defaults.headers.common["X-Session-ID"];
    this.deleteData();
  }

  setData(data) {
    if (!data) {
      return;
    }

    this.data = data;
    this.user = new User(this.data.user);
    this.storage.setItem("data", JSON.stringify(data));
    this.auth = true;
  }

  getUser() {
    return this.user;
  }

  getEmail() {
    if (this.isUser()) {
      return this.user.PrimaryEmail;
    }

    return "";
  }

  getNickName() {
    if (this.isUser()) {
      return this.user.NickName;
    }

    return "";
  }

  getFullName() {
    if (this.isUser()) {
      return this.user.FullName;
    }

    return "";
  }

  isUser() {
    return this.user && this.user.hasId();
  }

  isAdmin() {
    return this.user && this.user.hasId() && this.user.RoleAdmin;
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
    this.auth = false;
    this.user = new User();
    this.data = null;
    this.storage.removeItem("data");
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

  login(username, password, token) {
    this.deleteId();

    return Api.post("session", { username, password, token }).then((resp) => {
      this.setConfig(resp.data.config);
      this.setId(resp.data.id);
      this.setData(resp.data.data);
      this.sendClientInfo();
    });
  }

  redeemToken(token) {
    return Api.post("session", { token }).then((resp) => {
      this.setConfig(resp.data.config);
      this.setId(resp.data.id);
      this.setData(resp.data.data);
      this.sendClientInfo();
    });
  }

  onLogout(noRedirect) {
    this.deleteId();
    if (noRedirect !== true) {
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
