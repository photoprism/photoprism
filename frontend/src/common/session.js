import Api from "./api";
import User from "../model/user";

export default class Session {
    /**
     * @param {Storage} storage
     */
    constructor(storage) {
        this.auth = false;

        if(storage.getItem("session_storage") === "true") {
            this.storage = window.sessionStorage;
        } else {
            this.storage = storage;
        }

        this.session_token = this.storage.getItem("session_token");

        const userJson = this.storage.getItem("user");

        this.user = userJson !== "undefined" ? new User(JSON.parse(userJson)) : null;

        if(this.isUser()) {
            this.auth = true;
        }
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

    setToken(token) {
        this.session_token = token;
        this.storage.setItem("session_token", token);
        Api.defaults.headers.common["X-Session-Token"] = token;
    }

    getToken() {
        return this.session_token;
    }

    deleteToken() {
        this.session_token = null;
        this.storage.removeItem("session_token");
        Api.defaults.headers.common["X-Session-Token"] = "";
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

    login(email, password) {
        this.deleteToken();

        return Api.post("session", { email: email, password: password }).then(
            (result) => {
                this.setToken(result.data.token);
                this.setUser(new User(result.data.user));
            }
        );
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
