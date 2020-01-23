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
            "js": window.clientConfig.jsHash,
            "css": window.clientConfig.cssHash,
            "version": window.clientConfig.version,
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
