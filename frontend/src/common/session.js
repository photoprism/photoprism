/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

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

	PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
	to describe our software, run your own server, for educational purposes, but not for
	offering commercial goods, products, or services without prior written permission.
	In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

import Api from "./api";
import Event from "pubsub-js";
import User from "../model/user";
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

        if (this.applyToken(this.storage.getItem("session_token"))) {
            const userJson = this.storage.getItem("user");
            this.user = userJson !== "undefined" ? new User(JSON.parse(userJson)) : null;
        }

        if (this.isUser()) {
            this.auth = true;
        }

        Event.subscribe("session.logout", () => {
            this.onLogout();
        });

        Event.subscribe("websocket.connected", () => {
            this.sendClientInfo();
        });

        this.sendClientInfo();
    }

    useSessionStorage() {
        this.deleteToken();
        this.storage.setItem("session_storage", "true");
        this.storage = window.sessionStorage;
    }

    useLocalStorage() {
        this.storage.setItem("session_storage", "false");
        this.storage = window.localStorage;
    }

    applyToken(token) {
        if (!token) {
            this.deleteToken();
            return false;
        }

        this.session_token = token;
        Api.defaults.headers.common["X-Session-Token"] = token;

        return true;
    }

    setToken(token) {
        this.storage.setItem("session_token", token);
        return this.applyToken(token);
    }

    setConfig(values) {
        this.config.setValues(values);
    }

    getToken() {
        return this.session_token;
    }

    deleteToken() {
        this.session_token = null;
        this.storage.removeItem("session_token");
        delete Api.defaults.headers.common["X-Session-Token"];
        this.deleteUser();
    }

    setUser(user) {
        this.user = user;
        this.storage.setItem("user", JSON.stringify(user.getValues()));
        this.auth = true;
    }

    getUser() {
        return this.user;
    }

    getEmail() {
        if (this.isUser()) {
            return this.user.Email;
        }

        return "";
    }

    getFirstName() {
        if (this.isUser()) {
            return this.user.FirstName;
        }

        return "";
    }

    getFullName() {
        if (this.isUser()) {
            return this.user.FirstName + " " + this.user.LastName;
        }

        return "";
    }

    isUser() {
        return this.user && this.user.hasId();
    }

    isAdmin() {
        return this.user && this.user.hasId() && this.user.Role === "admin";
    }

    isAnonymous() {
        return !this.user || !this.user.hasId();
    }

    deleteUser() {
        this.auth = false;
        this.user = null;
        this.storage.removeItem("user");
    }

    sendClientInfo() {
        const clientInfo = {
            "session": this.getToken(),
            "js": window.__CONFIG__.jsHash,
            "css": window.__CONFIG__.cssHash,
            "version": window.__CONFIG__.version,
        };

        try {
            Socket.send(JSON.stringify(clientInfo));
        } catch(e) {
            console.log("can't send client info, websocket not connected (yet)");
        }
    }

    login(email, password) {
        this.deleteToken();

        return Api.post("session", {email: email, password: password}).then(
            (result) => {
                this.setConfig(result.data.config);
                this.setToken(result.data.token);
                this.setUser(new User(result.data.user));
                this.sendClientInfo();
            }
        );
    }

    onLogout() {
        console.log("ON LOGOUT");
        this.deleteToken();
        window.location = "/";
    }

    logout() {
        const token = this.getToken();

        this.deleteToken();

        Api.delete("session/" + token).then(
            () => {
                window.location = "/";
            }
        );
    }
}
